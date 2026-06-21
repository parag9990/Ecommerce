package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"product-service/internal/domain"
	"product-service/internal/repository"
)

const (
	defaultInventoryReservationTTL = 15 * time.Minute
	minInventoryReservationTTL     = 30 * time.Second
	maxInventoryReservationTTL     = time.Hour
	defaultExpiryBatchLimit        = 100
)

type InventoryUseCase interface {
	ReserveInventory(ctx context.Context, request ReserveInventoryRequest) (*InventoryReservationResult, error)
	ReleaseInventory(ctx context.Context, request InventoryReservationActionRequest) error
	CommitInventory(ctx context.Context, request InventoryReservationActionRequest) error
	ExpireReservations(ctx context.Context, request ExpireInventoryReservationsRequest) (ExpireInventoryReservationsResult, error)
}

type InventoryIDGenerator interface {
	NewInventoryReservationID() string
	NewInventorySnapshotID() string
}

type InventoryServiceOptions struct {
	DefaultReservationTTL time.Duration
	MinReservationTTL     time.Duration
	MaxReservationTTL     time.Duration
	ExpiryBatchLimit      int64
}

type ReserveInventoryRequest struct {
	OrderID        string
	Items          []ReserveInventoryItem
	TTLSeconds     int
	IdempotencyKey string
}

type ReserveInventoryItem struct {
	ProductID string
	VariantID string
	Quantity  int64
}

type InventoryReservationResult struct {
	ReservationID string
	ExpiresAt     time.Time
}

type InventoryReservationActionRequest struct {
	ReservationID string
	Reason        string
}

type ExpireInventoryReservationsRequest struct {
	Limit int64
}

type ExpireInventoryReservationsResult struct {
	ExpiredCount int
	FailedCount  int
}

type InventoryService struct {
	stock         repository.InventoryStockRepository
	reservations  repository.InventoryReservationRepository
	snapshots     repository.InventorySnapshotRepository
	ids           InventoryIDGenerator
	clock         Clock
	logger        *slog.Logger
	options       InventoryServiceOptions
	productEvents ProductEventRecorder
	productReader repository.ProductFinder
	atomic        repository.AtomicInventoryRepository
}

type reservedInventoryItem struct {
	item     domain.InventoryReservationItem
	mutation repository.InventoryStockMutation
}

func NewInventoryService(
	stock repository.InventoryStockRepository,
	reservations repository.InventoryReservationRepository,
	snapshots repository.InventorySnapshotRepository,
	ids InventoryIDGenerator,
	clock Clock,
	logger *slog.Logger,
	options InventoryServiceOptions,
) (*InventoryService, error) {
	if stock == nil {
		return nil, fmt.Errorf("inventory stock repository is required")
	}
	if reservations == nil {
		return nil, fmt.Errorf("inventory reservation repository is required")
	}
	if snapshots == nil {
		return nil, fmt.Errorf("inventory snapshot repository is required")
	}
	if ids == nil {
		ids = NewRandomIDGenerator()
	}
	if clock == nil {
		clock = SystemClock{}
	}
	if logger == nil {
		logger = slog.Default()
	}
	options = normalizeInventoryOptions(options)
	service := &InventoryService{
		stock:        stock,
		reservations: reservations,
		snapshots:    snapshots,
		ids:          ids,
		clock:        clock,
		logger:       logger,
		options:      options,
	}
	if atomic, ok := stock.(repository.AtomicInventoryRepository); ok {
		service.atomic = atomic
	}
	return service, nil
}

func (s *InventoryService) EnableProductEventRecording(
	events ProductEventRecorder,
	productReader repository.ProductFinder,
) {
	s.productEvents = events
	s.productReader = productReader
}

func (s *InventoryService) ReserveInventory(
	ctx context.Context,
	request ReserveInventoryRequest,
) (*InventoryReservationResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	request, ttl, err := s.normalizeReserveRequest(request)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now().UTC()

	existing, err := s.reservations.FindReservationByOrderID(ctx, request.OrderID)
	if err == nil {
		return s.reserveResultForExisting(ctx, *existing, now)
	}
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("find existing inventory reservation: %w", err)
	}

	reservation := domain.InventoryReservation{
		ID:             s.ids.NewInventoryReservationID(),
		OrderID:        request.OrderID,
		IdempotencyKey: request.IdempotencyKey,
		Status:         domain.ReservationStatusReserved,
		Items:          inventoryReservationItems(request.Items),
		ExpiresAt:      now.Add(ttl),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	reservedItems, err := s.reserveItems(ctx, &reservation, now)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateKey) {
			existing, findErr := s.reservations.FindReservationByOrderID(ctx, request.OrderID)
			if findErr == nil {
				return s.reserveResultForExisting(ctx, *existing, now)
			}
		}
		return nil, s.mapStockMutationError(err)
	}

	for _, item := range reservedItems {
		s.writeInventorySnapshot(ctx, reservation, item.mutation, domain.InventorySnapshotTypeReservation, "checkout_reservation", now)
	}
	if err := s.queueInventoryChangedEvents(ctx, inventoryMutations(reservedItems), now); err != nil {
		return nil, err
	}

	s.logger.Info(
		"inventory reserved",
		"order_id", reservation.OrderID,
		"reservation_id", reservation.ID,
		"item_count", len(reservation.Items),
		"expires_at", reservation.ExpiresAt,
	)
	return &InventoryReservationResult{ReservationID: reservation.ID, ExpiresAt: reservation.ExpiresAt}, nil
}

func (s *InventoryService) ReleaseInventory(ctx context.Context, request InventoryReservationActionRequest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	reservationID := strings.TrimSpace(request.ReservationID)
	if reservationID == "" {
		return serviceError(ErrorKindInvalidArgument, ErrorCodeReservationIDRequired, "reservation id is required", nil)
	}

	reservation, err := s.reservations.FindReservationByID(ctx, reservationID)
	if errors.Is(err, repository.ErrNotFound) {
		return serviceError(ErrorKindNotFound, ErrorCodeReservationNotFound, "inventory reservation not found", err)
	}
	if err != nil {
		return fmt.Errorf("find inventory reservation for release: %w", err)
	}

	switch reservation.Status {
	case domain.ReservationStatusReleased, domain.ReservationStatusExpired:
		return nil
	case domain.ReservationStatusCommitted:
		return serviceError(ErrorKindFailedPrecondition, ErrorCodeReservationAlreadyCommitted, "committed reservation cannot be released", nil)
	case domain.ReservationStatusReserved:
		return s.releaseReservation(ctx, *reservation, normalizeReason(request.Reason, "manual_release"), s.clock.Now().UTC())
	default:
		return serviceError(ErrorKindFailedPrecondition, ErrorCodeReservationNotActive, "reservation is not active", nil)
	}
}

func (s *InventoryService) CommitInventory(ctx context.Context, request InventoryReservationActionRequest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	reservationID := strings.TrimSpace(request.ReservationID)
	if reservationID == "" {
		return serviceError(ErrorKindInvalidArgument, ErrorCodeReservationIDRequired, "reservation id is required", nil)
	}

	reservation, err := s.reservations.FindReservationByID(ctx, reservationID)
	if errors.Is(err, repository.ErrNotFound) {
		return serviceError(ErrorKindNotFound, ErrorCodeReservationNotFound, "inventory reservation not found", err)
	}
	if err != nil {
		return fmt.Errorf("find inventory reservation for commit: %w", err)
	}

	switch reservation.Status {
	case domain.ReservationStatusCommitted:
		return nil
	case domain.ReservationStatusReleased, domain.ReservationStatusExpired:
		return serviceError(ErrorKindFailedPrecondition, ErrorCodeReservationNotActive, "reservation is not active", nil)
	case domain.ReservationStatusReserved:
	default:
		return serviceError(ErrorKindFailedPrecondition, ErrorCodeReservationNotActive, "reservation is not active", nil)
	}

	now := s.clock.Now().UTC()
	if now.After(reservation.ExpiresAt) {
		if err := s.expireReservation(ctx, *reservation, now); err != nil {
			s.logger.Error("expire reservation during commit failed", "reservation_id", reservation.ID, "error", err)
		}
		return serviceError(ErrorKindFailedPrecondition, ErrorCodeReservationExpired, "reservation has expired", nil)
	}

	mutations, err := s.finalizeReservation(ctx, *reservation, domain.ReservationStatusCommitted, normalizeReason(request.Reason, "payment_success"), now)
	if err != nil {
		if s.terminalStatusMatches(ctx, reservation.ID, domain.ReservationStatusCommitted) {
			return nil
		}
		return s.mapStockMutationError(err)
	}

	for _, mutation := range mutations {
		s.writeInventorySnapshot(ctx, *reservation, mutation, domain.InventorySnapshotTypeOrderCommit, normalizeReason(request.Reason, "payment_success"), now)
	}
	if err := s.queueInventoryChangedEvents(ctx, mutations, now); err != nil {
		return err
	}
	s.logger.Info("inventory committed", "reservation_id", reservation.ID, "order_id", reservation.OrderID, "item_count", len(reservation.Items))
	return nil
}

func (s *InventoryService) ExpireReservations(
	ctx context.Context,
	request ExpireInventoryReservationsRequest,
) (ExpireInventoryReservationsResult, error) {
	if err := ctx.Err(); err != nil {
		return ExpireInventoryReservationsResult{}, err
	}
	limit := request.Limit
	if limit <= 0 || limit > s.options.ExpiryBatchLimit {
		limit = s.options.ExpiryBatchLimit
	}
	now := s.clock.Now().UTC()
	reservations, err := s.reservations.ListExpiredReservations(ctx, now, limit)
	if err != nil {
		return ExpireInventoryReservationsResult{}, fmt.Errorf("list expired inventory reservations: %w", err)
	}

	var result ExpireInventoryReservationsResult
	for _, reservation := range reservations {
		if err := s.expireReservation(ctx, reservation, now); err != nil {
			result.FailedCount++
			s.logger.Error("expire inventory reservation failed", "reservation_id", reservation.ID, "order_id", reservation.OrderID, "error", err)
			continue
		}
		result.ExpiredCount++
	}
	if result.ExpiredCount > 0 || result.FailedCount > 0 {
		s.logger.Info("inventory reservation expiry batch processed", "expired_count", result.ExpiredCount, "failed_count", result.FailedCount)
	}
	return result, nil
}

func (s *InventoryService) normalizeReserveRequest(request ReserveInventoryRequest) (ReserveInventoryRequest, time.Duration, error) {
	orderID := strings.TrimSpace(request.OrderID)
	if orderID == "" {
		return ReserveInventoryRequest{}, 0, serviceError(ErrorKindInvalidArgument, ErrorCodeOrderIDRequired, "order id is required", nil)
	}
	if len(request.Items) == 0 {
		return ReserveInventoryRequest{}, 0, serviceError(ErrorKindInvalidArgument, ErrorCodeReservationItemsRequired, "reservation items are required", nil)
	}
	items := make([]ReserveInventoryItem, 0, len(request.Items))
	for _, item := range request.Items {
		productID := strings.TrimSpace(item.ProductID)
		variantID := strings.TrimSpace(item.VariantID)
		if productID == "" {
			return ReserveInventoryRequest{}, 0, serviceError(ErrorKindInvalidArgument, ErrorCodeProductNotAvailable, "product id is required", nil)
		}
		if variantID == "" {
			return ReserveInventoryRequest{}, 0, serviceError(ErrorKindInvalidArgument, ErrorCodeVariantNotAvailable, "variant id is required", nil)
		}
		if item.Quantity <= 0 {
			return ReserveInventoryRequest{}, 0, serviceError(ErrorKindInvalidArgument, ErrorCodeInvalidQuantity, "quantity must be greater than zero", nil)
		}
		items = append(items, ReserveInventoryItem{
			ProductID: productID,
			VariantID: variantID,
			Quantity:  item.Quantity,
		})
	}

	ttl := s.options.DefaultReservationTTL
	if request.TTLSeconds > 0 {
		ttl = time.Duration(request.TTLSeconds) * time.Second
	}
	if ttl < s.options.MinReservationTTL || ttl > s.options.MaxReservationTTL {
		return ReserveInventoryRequest{}, 0, serviceError(ErrorKindInvalidArgument, ErrorCodeInvalidReservationTTL, "reservation ttl is outside allowed range", nil)
	}

	idempotencyKey := strings.TrimSpace(request.IdempotencyKey)
	if idempotencyKey == "" {
		idempotencyKey = orderID
	}
	return ReserveInventoryRequest{
		OrderID:        orderID,
		Items:          items,
		TTLSeconds:     request.TTLSeconds,
		IdempotencyKey: idempotencyKey,
	}, ttl, nil
}

func (s *InventoryService) reserveResultForExisting(
	ctx context.Context,
	reservation domain.InventoryReservation,
	now time.Time,
) (*InventoryReservationResult, error) {
	switch reservation.Status {
	case domain.ReservationStatusReserved:
		if now.After(reservation.ExpiresAt) {
			if err := s.expireReservation(ctx, reservation, now); err != nil {
				s.logger.Error("expire stale reservation during idempotent reserve failed", "reservation_id", reservation.ID, "error", err)
			}
			return nil, serviceError(ErrorKindFailedPrecondition, ErrorCodeReservationExpired, "reservation has expired", nil)
		}
		return &InventoryReservationResult{ReservationID: reservation.ID, ExpiresAt: reservation.ExpiresAt}, nil
	case domain.ReservationStatusCommitted:
		return &InventoryReservationResult{ReservationID: reservation.ID, ExpiresAt: reservation.ExpiresAt}, nil
	case domain.ReservationStatusReleased, domain.ReservationStatusExpired:
		return nil, serviceError(ErrorKindFailedPrecondition, ErrorCodeReservationNotActive, "reservation is not active", nil)
	default:
		return nil, serviceError(ErrorKindFailedPrecondition, ErrorCodeReservationNotActive, "reservation is not active", nil)
	}
}

func (s *InventoryService) releaseReservation(
	ctx context.Context,
	reservation domain.InventoryReservation,
	reason string,
	now time.Time,
) error {
	mutations, err := s.finalizeReservation(ctx, reservation, domain.ReservationStatusReleased, reason, now)
	if err != nil {
		if s.terminalStatusMatches(ctx, reservation.ID, domain.ReservationStatusReleased) ||
			s.terminalStatusMatches(ctx, reservation.ID, domain.ReservationStatusExpired) {
			return nil
		}
		if s.terminalStatusMatches(ctx, reservation.ID, domain.ReservationStatusCommitted) {
			return serviceError(ErrorKindFailedPrecondition, ErrorCodeReservationAlreadyCommitted, "committed reservation cannot be released", nil)
		}
		return s.mapStockMutationError(err)
	}

	for _, mutation := range mutations {
		s.writeInventorySnapshot(ctx, reservation, mutation, domain.InventorySnapshotTypeRelease, reason, now)
	}
	if err := s.queueInventoryChangedEvents(ctx, mutations, now); err != nil {
		return err
	}
	s.logger.Info("inventory released", "reservation_id", reservation.ID, "order_id", reservation.OrderID, "reason", reason)
	return nil
}

func (s *InventoryService) expireReservation(ctx context.Context, reservation domain.InventoryReservation, now time.Time) error {
	if reservation.Status != domain.ReservationStatusReserved {
		return nil
	}
	reason := "reservation_expired"
	mutations, err := s.finalizeReservation(ctx, reservation, domain.ReservationStatusExpired, reason, now)
	if err != nil {
		if s.terminalStatusMatches(ctx, reservation.ID, domain.ReservationStatusExpired) ||
			s.terminalStatusMatches(ctx, reservation.ID, domain.ReservationStatusReleased) {
			return nil
		}
		return s.mapStockMutationError(err)
	}
	for _, mutation := range mutations {
		s.writeInventorySnapshot(ctx, reservation, mutation, domain.InventorySnapshotTypeRelease, reason, now)
	}
	if err := s.queueInventoryChangedEvents(ctx, mutations, now); err != nil {
		return err
	}
	return nil
}

func (s *InventoryService) reserveItems(
	ctx context.Context,
	reservation *domain.InventoryReservation,
	now time.Time,
) ([]reservedInventoryItem, error) {
	if s.atomic != nil {
		mutations, err := s.atomic.ReserveInventoryAtomic(ctx, reservation)
		if err != nil {
			return nil, err
		}
		items := make([]reservedInventoryItem, 0, len(mutations))
		for index, mutation := range mutations {
			items = append(items, reservedInventoryItem{item: reservation.Items[index], mutation: mutation})
		}
		return items, nil
	}

	items := make([]reservedInventoryItem, 0, len(reservation.Items))
	for _, item := range reservation.Items {
		mutation, err := s.stock.ReserveVariant(ctx, item.ProductID, item.VariantID, item.Quantity, now)
		if err != nil {
			s.rollbackReservedItems(ctx, items, now, "reserve_failed")
			return nil, err
		}
		items = append(items, reservedInventoryItem{
			item: domain.InventoryReservationItem{
				ProductID: mutation.ProductID, VariantID: mutation.VariantID, SKU: mutation.SKU,
				SellerID: mutation.SellerID, Quantity: item.Quantity,
			},
			mutation: mutation,
		})
	}
	reservation.Items = reservationItems(items)
	if report := reservation.Validate(); report.HasErrors() {
		s.rollbackReservedItems(ctx, items, now, "reservation_validation_failed")
		return nil, validationFailed(report)
	}
	if err := s.reservations.CreateInventoryReservation(ctx, reservation); err != nil {
		s.rollbackReservedItems(ctx, items, now, "reservation_create_failed")
		return nil, err
	}
	return items, nil
}

func (s *InventoryService) finalizeReservation(
	ctx context.Context,
	reservation domain.InventoryReservation,
	terminalStatus domain.InventoryReservationStatus,
	reason string,
	now time.Time,
) ([]repository.InventoryStockMutation, error) {
	if s.atomic != nil {
		return s.atomic.FinalizeInventoryReservationAtomic(ctx, reservation, terminalStatus, reason, now)
	}
	mutations := make([]repository.InventoryStockMutation, 0, len(reservation.Items))
	for _, item := range reservation.Items {
		var mutation repository.InventoryStockMutation
		var err error
		if terminalStatus == domain.ReservationStatusCommitted {
			mutation, err = s.stock.CommitVariant(ctx, item.ProductID, item.VariantID, item.Quantity, now)
		} else {
			mutation, err = s.stock.ReleaseVariant(ctx, item.ProductID, item.VariantID, item.Quantity, now)
		}
		if err != nil {
			return nil, err
		}
		mutations = append(mutations, mutation)
	}
	var err error
	switch terminalStatus {
	case domain.ReservationStatusCommitted:
		err = s.reservations.MarkReservationCommitted(ctx, reservation.ID, now)
	case domain.ReservationStatusReleased:
		err = s.reservations.MarkReservationReleased(ctx, reservation.ID, reason, now)
	case domain.ReservationStatusExpired:
		err = s.reservations.MarkReservationExpired(ctx, reservation.ID, reason, now)
	default:
		err = fmt.Errorf("unsupported terminal reservation status %q", terminalStatus)
	}
	if err != nil {
		return nil, err
	}
	return mutations, nil
}

func (s *InventoryService) rollbackReservedItems(
	ctx context.Context,
	items []reservedInventoryItem,
	now time.Time,
	reason string,
) {
	for index := len(items) - 1; index >= 0; index-- {
		item := items[index].item
		if _, err := s.stock.ReleaseVariant(ctx, item.ProductID, item.VariantID, item.Quantity, now); err != nil {
			s.logger.Error(
				"inventory reserve rollback failed",
				"product_id", item.ProductID,
				"variant_id", item.VariantID,
				"quantity", item.Quantity,
				"reason", reason,
				"error", err,
			)
		}
	}
}

func (s *InventoryService) writeInventorySnapshot(
	ctx context.Context,
	reservation domain.InventoryReservation,
	mutation repository.InventoryStockMutation,
	snapshotType domain.InventorySnapshotType,
	reason string,
	now time.Time,
) {
	snapshot := domain.InventorySnapshot{
		ID:                s.ids.NewInventorySnapshotID(),
		ProductID:         mutation.ProductID,
		VariantID:         mutation.VariantID,
		SKU:               mutation.SKU,
		SellerID:          mutation.SellerID,
		SnapshotType:      snapshotType,
		StockQuantity:     mutation.StockQuantity,
		ReservedQuantity:  mutation.ReservedQuantity,
		SafetyStock:       mutation.SafetyStock,
		AvailableQuantity: mutation.AvailableQuantity,
		Reason:            reason,
		Reference: &domain.InventoryReference{
			Type:    "reservation",
			ID:      reservation.ID,
			OrderID: reservation.OrderID,
		},
		CreatedAt: now,
	}
	if report := snapshot.Validate(); report.HasErrors() {
		s.logger.Error("inventory snapshot validation failed", "reservation_id", reservation.ID, "issues", report.Issues)
		return
	}
	if err := s.snapshots.CreateInventorySnapshot(ctx, snapshot); err != nil {
		s.logger.Error("inventory snapshot write failed", "reservation_id", reservation.ID, "snapshot_type", snapshotType, "error", err)
	}
}

func (s *InventoryService) mapStockMutationError(err error) error {
	switch {
	case errors.Is(err, repository.ErrInsufficientStock):
		return serviceError(ErrorKindFailedPrecondition, ErrorCodeOutOfStock, "insufficient inventory available", err)
	case errors.Is(err, repository.ErrInventoryUnavailable):
		return serviceError(ErrorKindFailedPrecondition, ErrorCodeProductNotAvailable, "product or variant is not available", err)
	case errors.Is(err, repository.ErrNotFound):
		return serviceError(ErrorKindFailedPrecondition, ErrorCodeVariantNotAvailable, "product variant is not available", err)
	case errors.Is(err, repository.ErrWriteConflict):
		return serviceError(ErrorKindConflict, ErrorCodeInventoryWriteConflict, "inventory write conflict", err)
	default:
		return err
	}
}

func (s *InventoryService) mapReservationWriteError(err error, operation string) error {
	if errors.Is(err, repository.ErrWriteConflict) {
		return serviceError(ErrorKindConflict, ErrorCodeInventoryWriteConflict, operation+" conflicted with another inventory operation", err)
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func (s *InventoryService) terminalStatusMatches(
	ctx context.Context,
	reservationID string,
	status domain.InventoryReservationStatus,
) bool {
	reservation, err := s.reservations.FindReservationByID(ctx, reservationID)
	if err != nil {
		return false
	}
	return reservation.Status == status
}

func reservationItems(items []reservedInventoryItem) []domain.InventoryReservationItem {
	result := make([]domain.InventoryReservationItem, 0, len(items))
	for _, item := range items {
		result = append(result, item.item)
	}
	return result
}

func inventoryReservationItems(items []ReserveInventoryItem) []domain.InventoryReservationItem {
	result := make([]domain.InventoryReservationItem, 0, len(items))
	for _, item := range items {
		result = append(result, domain.InventoryReservationItem{
			ProductID: item.ProductID,
			VariantID: item.VariantID,
			Quantity:  item.Quantity,
		})
	}
	return result
}

func inventoryMutations(items []reservedInventoryItem) []repository.InventoryStockMutation {
	result := make([]repository.InventoryStockMutation, 0, len(items))
	for _, item := range items {
		result = append(result, item.mutation)
	}
	return result
}

func (s *InventoryService) queueInventoryChangedEvents(
	ctx context.Context,
	mutations []repository.InventoryStockMutation,
	now time.Time,
) error {
	if s.productEvents == nil || s.productReader == nil {
		return nil
	}
	productIDs := make([]string, 0, len(mutations))
	seen := make(map[string]struct{}, len(mutations))
	for _, mutation := range mutations {
		if !mutation.InStockChanged() {
			continue
		}
		if _, exists := seen[mutation.ProductID]; exists {
			continue
		}
		seen[mutation.ProductID] = struct{}{}
		productIDs = append(productIDs, mutation.ProductID)
	}
	for _, productID := range productIDs {
		product, err := s.productReader.FindProductByID(ctx, productID)
		if err != nil {
			return fmt.Errorf("find product for inventory change event %q: %w", productID, err)
		}
		event, err := s.productEvents.BuildProductEvent(ctx, domain.ProductEventInventoryChanged, *product, now)
		if err != nil {
			return err
		}
		if event == nil {
			continue
		}
		if err := s.productEvents.QueueProductEvent(ctx, *event); err != nil {
			return err
		}
	}
	return nil
}

func normalizeInventoryOptions(options InventoryServiceOptions) InventoryServiceOptions {
	if options.DefaultReservationTTL <= 0 {
		options.DefaultReservationTTL = defaultInventoryReservationTTL
	}
	if options.MinReservationTTL <= 0 {
		options.MinReservationTTL = minInventoryReservationTTL
	}
	if options.MaxReservationTTL <= 0 {
		options.MaxReservationTTL = maxInventoryReservationTTL
	}
	if options.MaxReservationTTL < options.MinReservationTTL {
		options.MaxReservationTTL = options.MinReservationTTL
	}
	if options.DefaultReservationTTL < options.MinReservationTTL {
		options.DefaultReservationTTL = options.MinReservationTTL
	}
	if options.DefaultReservationTTL > options.MaxReservationTTL {
		options.DefaultReservationTTL = options.MaxReservationTTL
	}
	if options.ExpiryBatchLimit <= 0 {
		options.ExpiryBatchLimit = defaultExpiryBatchLimit
	}
	return options
}

func normalizeReason(reason string, fallback string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return fallback
	}
	return reason
}
