package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"testing"
	"time"

	"product-service/internal/domain"
	"product-service/internal/repository"
)

func TestReserveInventoryCreatesReservationAndIsIdempotent(t *testing.T) {
	repo := newInventoryMemoryRepository()
	repo.saveProduct(inventoryProduct("prod_1", "var_1", 5, 0, 1))
	service := newTestInventoryService(t, repo)

	result, err := service.ReserveInventory(context.Background(), ReserveInventoryRequest{
		OrderID:    "ord_1",
		TTLSeconds: 60,
		Items: []ReserveInventoryItem{
			{ProductID: "prod_1", VariantID: "var_1", Quantity: 2},
		},
	})
	if err != nil {
		t.Fatalf("ReserveInventory returned error: %v", err)
	}
	if result.ReservationID != "res_1" {
		t.Fatalf("reservation id = %s, want res_1", result.ReservationID)
	}
	if got := repo.variant("prod_1", "var_1").ReservedQuantity; got != 2 {
		t.Fatalf("reserved quantity = %d, want 2", got)
	}
	if len(repo.snapshots) != 1 || repo.snapshots[0].SnapshotType != domain.InventorySnapshotTypeReservation {
		t.Fatalf("snapshots = %+v, want one reservation snapshot", repo.snapshots)
	}

	retry, err := service.ReserveInventory(context.Background(), ReserveInventoryRequest{
		OrderID:    "ord_1",
		TTLSeconds: 60,
		Items: []ReserveInventoryItem{
			{ProductID: "prod_1", VariantID: "var_1", Quantity: 2},
		},
	})
	if err != nil {
		t.Fatalf("idempotent ReserveInventory returned error: %v", err)
	}
	if retry.ReservationID != result.ReservationID {
		t.Fatalf("retry reservation id = %s, want %s", retry.ReservationID, result.ReservationID)
	}
	if got := repo.variant("prod_1", "var_1").ReservedQuantity; got != 2 {
		t.Fatalf("reserved quantity after retry = %d, want 2", got)
	}
}

func TestReserveInventoryRollsBackPartialFailure(t *testing.T) {
	repo := newInventoryMemoryRepository()
	repo.saveProduct(inventoryProduct("prod_1", "var_1", 5, 0, 0))
	repo.saveProduct(inventoryProduct("prod_2", "var_2", 0, 0, 0))
	service := newTestInventoryService(t, repo)

	_, err := service.ReserveInventory(context.Background(), ReserveInventoryRequest{
		OrderID:    "ord_rollback",
		TTLSeconds: 60,
		Items: []ReserveInventoryItem{
			{ProductID: "prod_1", VariantID: "var_1", Quantity: 2},
			{ProductID: "prod_2", VariantID: "var_2", Quantity: 1},
		},
	})
	assertServiceError(t, err, ErrorCodeOutOfStock)
	if got := repo.variant("prod_1", "var_1").ReservedQuantity; got != 0 {
		t.Fatalf("rolled back reserved quantity = %d, want 0", got)
	}
	if len(repo.reservations) != 0 {
		t.Fatalf("reservation count = %d, want 0", len(repo.reservations))
	}
}

func TestReleaseAndCommitInventoryAreIdempotent(t *testing.T) {
	repo := newInventoryMemoryRepository()
	repo.saveProduct(inventoryProduct("prod_1", "var_1", 5, 0, 0))
	service := newTestInventoryService(t, repo)

	released := reserveForTest(t, service, "ord_release")
	if err := service.ReleaseInventory(context.Background(), InventoryReservationActionRequest{ReservationID: released.ReservationID, Reason: "payment_failed"}); err != nil {
		t.Fatalf("ReleaseInventory returned error: %v", err)
	}
	if err := service.ReleaseInventory(context.Background(), InventoryReservationActionRequest{ReservationID: released.ReservationID}); err != nil {
		t.Fatalf("idempotent ReleaseInventory returned error: %v", err)
	}
	if got := repo.variant("prod_1", "var_1").ReservedQuantity; got != 0 {
		t.Fatalf("reserved quantity after release = %d, want 0", got)
	}
	assertServiceError(t, service.CommitInventory(context.Background(), InventoryReservationActionRequest{ReservationID: released.ReservationID}), ErrorCodeReservationNotActive)

	committed := reserveForTest(t, service, "ord_commit")
	if err := service.CommitInventory(context.Background(), InventoryReservationActionRequest{ReservationID: committed.ReservationID, Reason: "payment_success"}); err != nil {
		t.Fatalf("CommitInventory returned error: %v", err)
	}
	if err := service.CommitInventory(context.Background(), InventoryReservationActionRequest{ReservationID: committed.ReservationID}); err != nil {
		t.Fatalf("idempotent CommitInventory returned error: %v", err)
	}
	variant := repo.variant("prod_1", "var_1")
	if variant.StockQuantity != 3 || variant.ReservedQuantity != 0 {
		t.Fatalf("variant after commit = %+v, want stock 3 reserved 0", variant)
	}
	assertServiceError(t, service.ReleaseInventory(context.Background(), InventoryReservationActionRequest{ReservationID: committed.ReservationID}), ErrorCodeReservationAlreadyCommitted)
}

func TestExpireReservationsReleasesReservedStock(t *testing.T) {
	repo := newInventoryMemoryRepository()
	repo.saveProduct(inventoryProduct("prod_1", "var_1", 5, 0, 0))
	clock := &inventoryTestClock{now: testNow()}
	service := newTestInventoryServiceWithClock(t, repo, clock)

	result := reserveForTest(t, service, "ord_expire")
	clock.now = result.ExpiresAt.Add(time.Second)

	expired, err := service.ExpireReservations(context.Background(), ExpireInventoryReservationsRequest{})
	if err != nil {
		t.Fatalf("ExpireReservations returned error: %v", err)
	}
	if expired.ExpiredCount != 1 || expired.FailedCount != 0 {
		t.Fatalf("expiry result = %+v, want one expired", expired)
	}
	if got := repo.variant("prod_1", "var_1").ReservedQuantity; got != 0 {
		t.Fatalf("reserved quantity after expiry = %d, want 0", got)
	}
	reservation := repo.reservations[result.ReservationID]
	if reservation.Status != domain.ReservationStatusExpired {
		t.Fatalf("reservation status = %s, want expired", reservation.Status)
	}
}

func TestReserveInventoryQueuesInventoryChangedEventWhenInStockChanges(t *testing.T) {
	repo := newInventoryMemoryRepository()
	repo.saveProduct(inventoryProduct("prod_1", "var_1", 1, 0, 0))
	service := newTestInventoryService(t, repo)
	outbox := &productEventMemoryOutbox{}
	eventService, err := NewProductEventService(
		outbox,
		&eventSequenceIDs{},
		&inventoryTestClock{now: testNow()},
		slog.Default(),
		ProductEventServiceOptions{Topic: "product.events", Source: "product-service"},
	)
	if err != nil {
		t.Fatalf("NewProductEventService returned error: %v", err)
	}
	service.EnableProductEventRecording(eventService, repo)

	if _, err := service.ReserveInventory(context.Background(), ReserveInventoryRequest{
		OrderID: "order_event",
		Items: []ReserveInventoryItem{
			{ProductID: "prod_1", VariantID: "var_1", Quantity: 1},
		},
	}); err != nil {
		t.Fatalf("ReserveInventory returned error: %v", err)
	}

	if len(outbox.events) != 1 {
		t.Fatalf("queued event count = %d, want 1", len(outbox.events))
	}
	event := outbox.events[0]
	if event.EventType != string(domain.ProductEventInventoryChanged) {
		t.Fatalf("event type = %s, want ProductInventoryChanged", event.EventType)
	}
	if event.Payload["in_stock"] != false {
		t.Fatalf("in_stock payload = %v, want false", event.Payload["in_stock"])
	}
}

func newTestInventoryService(t *testing.T, repo *inventoryMemoryRepository) *InventoryService {
	t.Helper()
	return newTestInventoryServiceWithClock(t, repo, &inventoryTestClock{now: testNow()})
}

func newTestInventoryServiceWithClock(t *testing.T, repo *inventoryMemoryRepository, clock *inventoryTestClock) *InventoryService {
	t.Helper()
	service, err := NewInventoryService(
		repo,
		repo,
		repo,
		&inventorySequenceIDs{},
		clock,
		slog.Default(),
		InventoryServiceOptions{},
	)
	if err != nil {
		t.Fatalf("new inventory service: %v", err)
	}
	return service
}

func reserveForTest(t *testing.T, service *InventoryService, orderID string) *InventoryReservationResult {
	t.Helper()
	result, err := service.ReserveInventory(context.Background(), ReserveInventoryRequest{
		OrderID:    orderID,
		TTLSeconds: 60,
		Items: []ReserveInventoryItem{
			{ProductID: "prod_1", VariantID: "var_1", Quantity: 2},
		},
	})
	if err != nil {
		t.Fatalf("ReserveInventory returned error: %v", err)
	}
	return result
}

type inventorySequenceIDs struct {
	reservation int
	snapshot    int
}

func (g *inventorySequenceIDs) NewInventoryReservationID() string {
	g.reservation++
	return fmt.Sprintf("res_%d", g.reservation)
}

func (g *inventorySequenceIDs) NewInventorySnapshotID() string {
	g.snapshot++
	return fmt.Sprintf("inv_snap_%d", g.snapshot)
}

type inventoryTestClock struct {
	now time.Time
}

func (c *inventoryTestClock) Now() time.Time {
	return c.now
}

type inventoryMemoryRepository struct {
	products     map[string]domain.Product
	reservations map[string]domain.InventoryReservation
	snapshots    []domain.InventorySnapshot
}

func newInventoryMemoryRepository() *inventoryMemoryRepository {
	return &inventoryMemoryRepository{
		products:     make(map[string]domain.Product),
		reservations: make(map[string]domain.InventoryReservation),
	}
}

func (r *inventoryMemoryRepository) ReserveVariant(
	ctx context.Context,
	productID string,
	variantID string,
	quantity int64,
	now time.Time,
) (repository.InventoryStockMutation, error) {
	if err := ctx.Err(); err != nil {
		return repository.InventoryStockMutation{}, err
	}
	product, index, err := r.findVariant(productID, variantID)
	if err != nil {
		return repository.InventoryStockMutation{}, err
	}
	if product.Status != domain.ProductStatusPublished || !product.Variants[index].Active() {
		return repository.InventoryStockMutation{}, repository.ErrInventoryUnavailable
	}
	if !product.Variants[index].InventoryState().CanFulfill(quantity) {
		return repository.InventoryStockMutation{}, repository.ErrInsufficientStock
	}
	previous := product.Variants[index].InventoryState()
	product.Variants[index].ReservedQuantity += quantity
	product.UpdatedAt = now
	r.products[product.ID] = product
	return inventoryMutationWithPrevious(product, product.Variants[index], previous), nil
}

func (r *inventoryMemoryRepository) ReleaseVariant(
	ctx context.Context,
	productID string,
	variantID string,
	quantity int64,
	now time.Time,
) (repository.InventoryStockMutation, error) {
	if err := ctx.Err(); err != nil {
		return repository.InventoryStockMutation{}, err
	}
	product, index, err := r.findVariant(productID, variantID)
	if err != nil {
		return repository.InventoryStockMutation{}, err
	}
	if product.Variants[index].ReservedQuantity < quantity {
		return repository.InventoryStockMutation{}, repository.ErrWriteConflict
	}
	previous := product.Variants[index].InventoryState()
	product.Variants[index].ReservedQuantity -= quantity
	product.UpdatedAt = now
	r.products[product.ID] = product
	return inventoryMutationWithPrevious(product, product.Variants[index], previous), nil
}

func (r *inventoryMemoryRepository) CommitVariant(
	ctx context.Context,
	productID string,
	variantID string,
	quantity int64,
	now time.Time,
) (repository.InventoryStockMutation, error) {
	if err := ctx.Err(); err != nil {
		return repository.InventoryStockMutation{}, err
	}
	product, index, err := r.findVariant(productID, variantID)
	if err != nil {
		return repository.InventoryStockMutation{}, err
	}
	if product.Variants[index].StockQuantity < quantity || product.Variants[index].ReservedQuantity < quantity {
		return repository.InventoryStockMutation{}, repository.ErrWriteConflict
	}
	previous := product.Variants[index].InventoryState()
	product.Variants[index].StockQuantity -= quantity
	product.Variants[index].ReservedQuantity -= quantity
	product.UpdatedAt = now
	r.products[product.ID] = product
	return inventoryMutationWithPrevious(product, product.Variants[index], previous), nil
}

func (r *inventoryMemoryRepository) FindProductByID(ctx context.Context, productID string) (*domain.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	product, ok := r.products[productID]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return cloneProduct(product), nil
}

func (r *inventoryMemoryRepository) FindReservationByID(ctx context.Context, reservationID string) (*domain.InventoryReservation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	reservation, ok := r.reservations[reservationID]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return cloneReservation(reservation), nil
}

func (r *inventoryMemoryRepository) FindReservationByOrderID(ctx context.Context, orderID string) (*domain.InventoryReservation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	for _, reservation := range r.reservations {
		if reservation.OrderID == orderID {
			return cloneReservation(reservation), nil
		}
	}
	return nil, repository.ErrNotFound
}

func (r *inventoryMemoryRepository) CreateInventoryReservation(ctx context.Context, reservation *domain.InventoryReservation) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, exists := r.reservations[reservation.ID]; exists {
		return repository.ErrDuplicateKey
	}
	for _, existing := range r.reservations {
		if existing.OrderID == reservation.OrderID ||
			(reservation.IdempotencyKey != "" && existing.IdempotencyKey == reservation.IdempotencyKey) {
			return repository.ErrDuplicateKey
		}
	}
	r.reservations[reservation.ID] = *cloneReservation(*reservation)
	return nil
}

func (r *inventoryMemoryRepository) MarkReservationReleased(ctx context.Context, reservationID string, reason string, releasedAt time.Time) error {
	return r.markReservation(ctx, reservationID, domain.ReservationStatusReleased, reason, releasedAt)
}

func (r *inventoryMemoryRepository) MarkReservationCommitted(ctx context.Context, reservationID string, committedAt time.Time) error {
	return r.markReservation(ctx, reservationID, domain.ReservationStatusCommitted, "", committedAt)
}

func (r *inventoryMemoryRepository) MarkReservationExpired(ctx context.Context, reservationID string, reason string, expiredAt time.Time) error {
	return r.markReservation(ctx, reservationID, domain.ReservationStatusExpired, reason, expiredAt)
}

func (r *inventoryMemoryRepository) ListExpiredReservations(ctx context.Context, now time.Time, limit int64) ([]domain.InventoryReservation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	reservations := make([]domain.InventoryReservation, 0)
	for _, reservation := range r.reservations {
		if reservation.Status == domain.ReservationStatusReserved && !reservation.ExpiresAt.After(now) {
			reservations = append(reservations, *cloneReservation(reservation))
		}
	}
	sort.Slice(reservations, func(i, j int) bool {
		return reservations[i].ExpiresAt.Before(reservations[j].ExpiresAt)
	})
	if limit > 0 && int64(len(reservations)) > limit {
		reservations = reservations[:limit]
	}
	return reservations, nil
}

func (r *inventoryMemoryRepository) CreateInventorySnapshot(ctx context.Context, snapshot domain.InventorySnapshot) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.snapshots = append(r.snapshots, snapshot)
	return nil
}

func (r *inventoryMemoryRepository) markReservation(
	ctx context.Context,
	reservationID string,
	status domain.InventoryReservationStatus,
	reason string,
	now time.Time,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	reservation, ok := r.reservations[reservationID]
	if !ok {
		return repository.ErrNotFound
	}
	if reservation.Status != domain.ReservationStatusReserved {
		return repository.ErrWriteConflict
	}
	reservation.Status = status
	reservation.Reason = reason
	reservation.UpdatedAt = now
	switch status {
	case domain.ReservationStatusReleased:
		reservation.ReleasedAt = &now
	case domain.ReservationStatusCommitted:
		reservation.CommittedAt = &now
	case domain.ReservationStatusExpired:
		reservation.ExpiredAt = &now
	}
	r.reservations[reservationID] = reservation
	return nil
}

func (r *inventoryMemoryRepository) saveProduct(product domain.Product) {
	r.products[product.ID] = product
}

func (r *inventoryMemoryRepository) findVariant(productID string, variantID string) (domain.Product, int, error) {
	product, ok := r.products[productID]
	if !ok {
		return domain.Product{}, 0, repository.ErrNotFound
	}
	for index := range product.Variants {
		if product.Variants[index].ID == variantID {
			return product, index, nil
		}
	}
	return domain.Product{}, 0, repository.ErrNotFound
}

func (r *inventoryMemoryRepository) variant(productID string, variantID string) domain.Variant {
	product, index, err := r.findVariant(productID, variantID)
	if err != nil {
		panic(err)
	}
	return product.Variants[index]
}

func inventoryProduct(productID string, variantID string, stock int64, reserved int64, safetyStock int64) domain.Product {
	product := validDraftProduct()
	product.ID = productID
	product.Status = domain.ProductStatusPublished
	product.Variants[0].ID = variantID
	product.Variants[0].StockQuantity = stock
	product.Variants[0].ReservedQuantity = reserved
	product.Variants[0].SafetyStock = safetyStock
	product.Variants[0].Status = domain.VariantStatusActive
	return product
}

func inventoryMutation(product domain.Product, variant domain.Variant) repository.InventoryStockMutation {
	state := variant.InventoryState()
	return repository.InventoryStockMutation{
		ProductID:         product.ID,
		VariantID:         variant.ID,
		SKU:               variant.SKU,
		SellerID:          product.SellerID,
		StockQuantity:     state.StockQuantity,
		ReservedQuantity:  state.ReservedQuantity,
		SafetyStock:       state.SafetyStock,
		AvailableQuantity: state.AvailableQuantity(),
	}
}

func inventoryMutationWithPrevious(
	product domain.Product,
	variant domain.Variant,
	previous domain.InventoryState,
) repository.InventoryStockMutation {
	mutation := inventoryMutation(product, variant)
	mutation.PreviousAvailableQuantity = previous.AvailableQuantity()
	mutation.PreviousInStock = previous.AvailableQuantity() > 0
	mutation.HasPreviousAvailability = true
	return mutation
}

func cloneReservation(reservation domain.InventoryReservation) *domain.InventoryReservation {
	clone := reservation
	clone.Items = append([]domain.InventoryReservationItem(nil), reservation.Items...)
	return &clone
}

func assertNotServiceError(t *testing.T, err error, code string) {
	t.Helper()
	var serviceErr *ServiceError
	if errors.As(err, &serviceErr) && serviceErr.Code == code {
		t.Fatalf("did not expect service error %s", code)
	}
}
