package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/domain"
)

var (
	ErrUnauthenticated           = errors.New("authenticated buyer is required")
	ErrPermissionDenied          = errors.New("buyer role is required")
	ErrInvalidInput              = errors.New("invalid wishlist request")
	ErrProductNotFound           = errors.New("product not found")
	ErrProductNotAvailable       = errors.New("product is not available")
	ErrVariantNotFound           = errors.New("product variant not found")
	ErrProductServiceUnavailable = errors.New("product service unavailable")
	ErrVariantRequiredForCart    = errors.New("variant is required to move wishlist item to cart")
	ErrCartServiceUnavailable    = errors.New("cart service unavailable")
	ErrCartValidation            = errors.New("cart validation failed")
	ErrCartUnauthenticated       = errors.New("cart service authentication failed")
	ErrCartForbidden             = errors.New("cart service authorization failed")
	ErrCartProductNotFound       = errors.New("cart product not found")
	ErrCartItemUnavailable       = errors.New("cart item unavailable")
	ErrMoveToCartPartialFailure  = errors.New("move to cart partially failed")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (e ValidationError) Unwrap() error {
	return ErrInvalidInput
}

type AddWishlistItemInput struct {
	UserID    string
	ProductID string
	VariantID string
	TraceID   string
}

type GetWishlistInput struct {
	UserID string
}

type RemoveWishlistItemInput struct {
	UserID    string
	ProductID string
	TraceID   string
}

type MoveToCartInput struct {
	UserID         string
	ProductID      string
	RequestID      string
	IdempotencyKey string
}

type SyncProductAvailabilityInput struct {
	EventID      string
	EventType    string
	ProductID    string
	VariantID    string
	Availability domain.Availability
	OccurredAt   time.Time
	TraceID      string
}

type AvailabilitySyncResult struct {
	ProductID     string
	VariantID     string
	Availability  domain.Availability
	MatchedCount  int64
	ModifiedCount int64
}

type ProductSnapshot struct {
	ProductID      string
	VariantID      string
	LastKnownPrice *domain.Money
	Availability   domain.Availability
}

type CartAddItemInput struct {
	UserID         string
	ProductID      string
	VariantID      string
	Quantity       int
	RequestID      string
	IdempotencyKey string
}

type Cart struct {
	CartID   string
	UserID   string
	Items    []CartItem
	Subtotal *domain.Money
	Discount *domain.Money
	Total    *domain.Money
}

type CartItem struct {
	ItemID    string
	ProductID string
	VariantID string
	Quantity  int
}

type ProductValidator interface {
	ValidateWishlistProduct(ctx context.Context, productID, variantID string) (ProductSnapshot, error)
}

type CartClient interface {
	AddItem(ctx context.Context, input CartAddItemInput) (*Cart, error)
}

type WishlistRepository interface {
	EnsureWishlist(ctx context.Context, userID string, now time.Time) error
	AddItemIfNotExists(ctx context.Context, userID string, item domain.WishlistItem, now time.Time) error
	RemoveItem(ctx context.Context, userID, productID string, now time.Time) error
	FindByUserID(ctx context.Context, userID string) (*domain.Wishlist, error)
	UpdateProductAvailability(ctx context.Context, productID, variantID string, availability domain.Availability, updatedAt time.Time) (domain.AvailabilityUpdateResult, error)
}

type WishlistEventRepository interface {
	Enqueue(ctx context.Context, event domain.WishlistAnalyticsEvent) error
	ClaimPending(ctx context.Context, limit int, now time.Time) ([]domain.WishlistAnalyticsEvent, error)
	MarkPublished(ctx context.Context, eventID string, publishedAt time.Time) error
	MarkRetry(ctx context.Context, eventID string, attempts int, nextRetryAt time.Time, lastError string) error
	MarkFailed(ctx context.Context, eventID string, attempts int, lastError string, failedAt time.Time) error
}

type WishlistService struct {
	repository       WishlistRepository
	eventRepository  WishlistEventRepository
	productValidator ProductValidator
	cartClient       CartClient
	eventIDFactory   func() string
	eventTopic       string
	clock            func() time.Time
	logger           *slog.Logger
}

type WishlistServiceOption func(*WishlistService)

func WithClock(clock func() time.Time) WishlistServiceOption {
	return func(s *WishlistService) {
		if clock != nil {
			s.clock = clock
		}
	}
}

func WithWishlistEventRepository(repository WishlistEventRepository) WishlistServiceOption {
	return func(s *WishlistService) {
		s.eventRepository = repository
	}
}

func WithWishlistEventTopic(topic string) WishlistServiceOption {
	return func(s *WishlistService) {
		if strings.TrimSpace(topic) != "" {
			s.eventTopic = strings.TrimSpace(topic)
		}
	}
}

func WithWishlistEventIDFactory(factory func() string) WishlistServiceOption {
	return func(s *WishlistService) {
		if factory != nil {
			s.eventIDFactory = factory
		}
	}
}

func NewWishlistService(repository WishlistRepository, productValidator ProductValidator, cartClient CartClient, logger *slog.Logger, options ...WishlistServiceOption) (*WishlistService, error) {
	if repository == nil {
		return nil, errors.New("wishlist repository is required")
	}
	if productValidator == nil {
		return nil, errors.New("product validator is required")
	}
	if cartClient == nil {
		return nil, errors.New("cart client is required")
	}
	if logger == nil {
		logger = slog.Default()
	}

	service := &WishlistService{
		repository:       repository,
		productValidator: productValidator,
		cartClient:       cartClient,
		eventIDFactory:   defaultWishlistEventID,
		eventTopic:       DefaultRecommendationEventsTopic,
		clock: func() time.Time {
			return time.Now().UTC()
		},
		logger: logger,
	}
	for _, option := range options {
		option(service)
	}
	return service, nil
}

func (s *WishlistService) GetWishlist(ctx context.Context, input GetWishlistInput) (*domain.Wishlist, error) {
	if s == nil || s.repository == nil {
		return nil, errors.New("wishlist service is not initialized")
	}
	ctx = contextOrBackground(ctx)

	userID := normalizeID(input.UserID)
	if userID == "" {
		return nil, ErrUnauthenticated
	}

	now := s.clock().UTC()
	if err := s.repository.EnsureWishlist(ctx, userID, now); err != nil {
		return nil, err
	}
	return s.repository.FindByUserID(ctx, userID)
}

func (s *WishlistService) AddItem(ctx context.Context, input AddWishlistItemInput) (*domain.Wishlist, error) {
	if s == nil || s.repository == nil || s.productValidator == nil {
		return nil, errors.New("wishlist service is not initialized")
	}
	ctx = contextOrBackground(ctx)

	userID := normalizeID(input.UserID)
	productID := normalizeID(input.ProductID)
	variantID := normalizeID(input.VariantID)
	if userID == "" {
		return nil, ErrUnauthenticated
	}
	if productID == "" {
		return nil, ValidationError{Field: "product_id", Message: "is required"}
	}

	product, err := s.productValidator.ValidateWishlistProduct(ctx, productID, variantID)
	if err != nil {
		s.logger.Warn("wishlist product validation failed", "user_id", userID, "product_id", productID, "error", err)
		return nil, err
	}

	now := s.clock().UTC()
	itemInput := domain.AddWishlistItemInput{
		ProductID:      firstNonEmpty(product.ProductID, productID),
		VariantID:      firstNonEmpty(product.VariantID, variantID),
		LastKnownPrice: validMoneyOrNil(product.LastKnownPrice),
		Availability:   product.Availability.Normalized(),
	}
	item, err := domain.NewWishlistItem(itemInput, now)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}

	if err := s.repository.EnsureWishlist(ctx, userID, now); err != nil {
		return nil, err
	}
	if err := s.repository.AddItemIfNotExists(ctx, userID, item, now); err != nil {
		if errors.Is(err, domain.ErrDuplicateProduct) {
			s.logger.Info("wishlist duplicate add blocked", "user_id", userID, "product_id", productID)
		}
		return nil, err
	}

	s.enqueueWishlistAnalyticsEvent(ctx, domain.WishlistEventItemAdded, userID, item, normalizeID(input.TraceID), now)

	wishlist, err := s.repository.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	s.logger.Info("wishlist item added", "user_id", userID, "product_id", productID)
	return wishlist, nil
}

func (s *WishlistService) RemoveItem(ctx context.Context, input RemoveWishlistItemInput) (*domain.Wishlist, error) {
	if s == nil || s.repository == nil {
		return nil, errors.New("wishlist service is not initialized")
	}
	ctx = contextOrBackground(ctx)

	userID := normalizeID(input.UserID)
	productID := normalizeID(input.ProductID)
	if userID == "" {
		return nil, ErrUnauthenticated
	}
	if productID == "" {
		return nil, ValidationError{Field: "product_id", Message: "is required"}
	}

	now := s.clock().UTC()
	if err := s.repository.EnsureWishlist(ctx, userID, now); err != nil {
		return nil, err
	}
	wishlistBeforeRemove, err := s.repository.FindByUserID(ctx, userID)
	if err != nil && !errors.Is(err, domain.ErrWishlistNotFound) {
		return nil, err
	}

	removedItem := domain.WishlistItem{}
	itemExisted := false
	if wishlistBeforeRemove != nil {
		removedItem, itemExisted = wishlistBeforeRemove.FindItem(productID)
	}

	if err := s.repository.RemoveItem(ctx, userID, productID, now); err != nil {
		return nil, err
	}
	if itemExisted {
		s.enqueueWishlistAnalyticsEvent(ctx, domain.WishlistEventItemRemoved, userID, removedItem, normalizeID(input.TraceID), now)
	}

	wishlist, err := s.repository.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	s.logger.Info("wishlist item removed", "user_id", userID, "product_id", productID)
	return wishlist, nil
}

func (s *WishlistService) MoveToCart(ctx context.Context, input MoveToCartInput) (*Cart, error) {
	if s == nil || s.repository == nil || s.cartClient == nil {
		return nil, errors.New("wishlist service is not initialized")
	}
	ctx = contextOrBackground(ctx)

	userID := normalizeID(input.UserID)
	productID := normalizeID(input.ProductID)
	if userID == "" {
		return nil, ErrUnauthenticated
	}
	if productID == "" {
		return nil, ValidationError{Field: "product_id", Message: "is required"}
	}

	wishlist, err := s.repository.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrWishlistNotFound) {
			return nil, fmt.Errorf("%w: %s", domain.ErrWishlistItemNotFound, productID)
		}
		return nil, err
	}

	item, found := wishlist.FindItem(productID)
	if !found {
		return nil, fmt.Errorf("%w: %s", domain.ErrWishlistItemNotFound, productID)
	}
	item.VariantID = normalizeID(item.VariantID)
	if item.VariantID == "" {
		return nil, fmt.Errorf("%w: %w", ErrVariantRequiredForCart, ValidationError{
			Field:   "variant_id",
			Message: "is required to move wishlist item to cart",
		})
	}

	cart, err := s.cartClient.AddItem(ctx, CartAddItemInput{
		UserID:         userID,
		ProductID:      item.ProductID,
		VariantID:      item.VariantID,
		Quantity:       1,
		RequestID:      normalizeID(input.RequestID),
		IdempotencyKey: moveToCartIdempotencyKey(input, item),
	})
	if err != nil {
		s.logger.Warn("wishlist move to cart failed",
			"user_id", userID,
			"product_id", productID,
			"request_id", normalizeID(input.RequestID),
			"error", err,
		)
		return nil, err
	}

	now := s.clock().UTC()
	if err := s.repository.RemoveItem(ctx, userID, productID, now); err != nil {
		s.logger.Error("wishlist cleanup failed after cart add",
			"user_id", userID,
			"product_id", productID,
			"request_id", normalizeID(input.RequestID),
			"error", err,
		)
		return nil, fmt.Errorf("%w: cart add succeeded but wishlist cleanup failed: %w", ErrMoveToCartPartialFailure, err)
	}

	s.logger.Info("wishlist item moved to cart",
		"user_id", userID,
		"product_id", productID,
		"request_id", normalizeID(input.RequestID),
	)
	return cart, nil
}

func (s *WishlistService) SyncProductAvailability(ctx context.Context, input SyncProductAvailabilityInput) (AvailabilitySyncResult, error) {
	if s == nil || s.repository == nil {
		return AvailabilitySyncResult{}, errors.New("wishlist service is not initialized")
	}
	ctx = contextOrBackground(ctx)

	productID := normalizeID(input.ProductID)
	variantID := normalizeID(input.VariantID)
	if productID == "" {
		return AvailabilitySyncResult{}, ValidationError{Field: "product_id", Message: "is required"}
	}

	availability := input.Availability.Normalized()
	if !availability.IsValid() {
		return AvailabilitySyncResult{}, ValidationError{
			Field:   "availability",
			Message: "must be in_stock, out_of_stock, unknown, or deleted",
		}
	}

	occurredAt := input.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = s.clock().UTC()
		s.logger.Warn("product event occurred_at missing; using processing time",
			"event_id", normalizeID(input.EventID),
			"event_type", normalizeID(input.EventType),
			"product_id", productID,
			"trace_id", normalizeID(input.TraceID),
		)
	} else {
		occurredAt = occurredAt.UTC()
	}

	result, err := s.repository.UpdateProductAvailability(ctx, productID, variantID, availability, occurredAt)
	if err != nil {
		s.logger.Error("wishlist availability sync failed",
			"event_id", normalizeID(input.EventID),
			"event_type", normalizeID(input.EventType),
			"product_id", productID,
			"variant_id", variantID,
			"availability", availability,
			"trace_id", normalizeID(input.TraceID),
			"error", err,
		)
		return AvailabilitySyncResult{}, err
	}

	s.logger.Info("wishlist availability synced",
		"event_id", normalizeID(input.EventID),
		"event_type", normalizeID(input.EventType),
		"product_id", productID,
		"variant_id", variantID,
		"availability", availability,
		"matched_count", result.MatchedCount,
		"modified_count", result.ModifiedCount,
		"trace_id", normalizeID(input.TraceID),
	)
	return AvailabilitySyncResult{
		ProductID:     productID,
		VariantID:     variantID,
		Availability:  availability,
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
	}, nil
}

func (s *WishlistService) enqueueWishlistAnalyticsEvent(ctx context.Context, eventType domain.WishlistEventType, userID string, item domain.WishlistItem, traceID string, occurredAt time.Time) {
	if s == nil || s.eventRepository == nil {
		return
	}
	event := newWishlistAnalyticsEvent(
		s.eventIDFactory(),
		s.eventTopic,
		eventType,
		userID,
		item,
		traceID,
		occurredAt,
	)
	if err := s.eventRepository.Enqueue(ctx, event); err != nil {
		s.logger.Warn("wishlist analytics event enqueue failed",
			"event_id", event.EventID,
			"event_type", event.EventType,
			"user_id", userID,
			"product_id", item.ProductID,
			"trace_id", traceID,
			"error", err,
		)
		return
	}
	s.logger.Info("wishlist analytics event enqueued",
		"event_id", event.EventID,
		"event_type", event.EventType,
		"user_id", userID,
		"product_id", item.ProductID,
		"trace_id", traceID,
	)
}

func moveToCartIdempotencyKey(input MoveToCartInput, item domain.WishlistItem) string {
	if key := normalizeID(input.IdempotencyKey); key != "" {
		return key
	}
	return "wishlist-move:" + normalizeID(input.UserID) + ":" + normalizeID(item.ProductID) + ":" + normalizeID(item.VariantID)
}

func normalizeID(value string) string {
	return strings.TrimSpace(value)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = normalizeID(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func validMoneyOrNil(money *domain.Money) *domain.Money {
	if money == nil {
		return nil
	}
	copied := &domain.Money{
		Amount:   money.Amount,
		Currency: strings.ToUpper(strings.TrimSpace(money.Currency)),
	}
	if err := copied.Validate(); err != nil {
		return nil
	}
	return copied
}

func contextOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
