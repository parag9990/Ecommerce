package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/domain"
)

func TestWishlistServiceGetWishlistCreatesEmptyWishlistForBuyer(t *testing.T) {
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	repository := newFakeWishlistRepository(t)
	service := mustWishlistService(t, repository, &fakeProductValidator{}, now)

	wishlist, err := service.GetWishlist(context.Background(), GetWishlistInput{UserID: " user_123 "})
	if err != nil {
		t.Fatalf("GetWishlist returned error: %v", err)
	}
	if wishlist.UserID != "user_123" || len(wishlist.Items) != 0 {
		t.Fatalf("wishlist = %#v, want empty wishlist for user_123", wishlist)
	}
}

func TestWishlistServiceGetWishlistRequiresAuthentication(t *testing.T) {
	service := mustWishlistService(t, newFakeWishlistRepository(t), &fakeProductValidator{}, time.Now().UTC())
	if _, err := service.GetWishlist(context.Background(), GetWishlistInput{}); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("GetWishlist error = %v, want ErrUnauthenticated", err)
	}
}

func TestWishlistServiceAddItemValidatesProductAndBlocksDuplicate(t *testing.T) {
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	repository := newFakeWishlistRepository(t)
	eventRepository := &fakeWishlistEventRepository{}
	validator := &fakeProductValidator{
		snapshot: ProductSnapshot{
			ProductID: "prod_123",
			VariantID: "var_1",
			LastKnownPrice: &domain.Money{
				Amount:   299900,
				Currency: "INR",
			},
			Availability: domain.AvailabilityInStock,
		},
	}
	service := mustWishlistServiceWithEvents(t, repository, validator, eventRepository, now)

	wishlist, err := service.AddItem(context.Background(), AddWishlistItemInput{
		UserID:    " user_123 ",
		ProductID: " prod_123 ",
		VariantID: " var_1 ",
		TraceID:   " req_add ",
	})
	if err != nil {
		t.Fatalf("AddItem returned error: %v", err)
	}
	if wishlist.UserID != "user_123" {
		t.Fatalf("wishlist.UserID = %q, want user_123", wishlist.UserID)
	}
	item, ok := wishlist.FindItem("prod_123")
	if !ok {
		t.Fatal("wishlist missing prod_123")
	}
	if item.VariantID != "var_1" {
		t.Fatalf("item.VariantID = %q, want var_1", item.VariantID)
	}
	if item.LastKnownPrice == nil || item.LastKnownPrice.Amount != 299900 {
		t.Fatalf("item.LastKnownPrice = %#v, want amount 299900", item.LastKnownPrice)
	}
	if len(eventRepository.enqueued) != 1 {
		t.Fatalf("analytics events = %d, want 1", len(eventRepository.enqueued))
	}
	event := eventRepository.enqueued[0]
	if event.EventID != "evt_test" || event.EventType != domain.WishlistEventItemAdded {
		t.Fatalf("event id/type = %q/%q, want evt_test/%q", event.EventID, event.EventType, domain.WishlistEventItemAdded)
	}
	if event.Topic != "recommendation.events.test" || event.TraceID != "req_add" {
		t.Fatalf("event topic/trace = %q/%q", event.Topic, event.TraceID)
	}
	if event.Payload.UserID != "user_123" || event.Payload.ProductID != "prod_123" || event.Payload.VariantID != "var_1" {
		t.Fatalf("event payload = %#v, want user_123/prod_123/var_1", event.Payload)
	}
	if event.Payload.Action != domain.WishlistEventActionAdd || event.Payload.Source != domain.WishlistEventSource {
		t.Fatalf("event action/source = %q/%q", event.Payload.Action, event.Payload.Source)
	}

	_, err = service.AddItem(context.Background(), AddWishlistItemInput{
		UserID:    "user_123",
		ProductID: "prod_123",
		VariantID: "var_2",
	})
	if !errors.Is(err, domain.ErrDuplicateProduct) {
		t.Fatalf("duplicate AddItem error = %v, want ErrDuplicateProduct", err)
	}
	if len(eventRepository.enqueued) != 1 {
		t.Fatalf("analytics events after duplicate = %d, want 1", len(eventRepository.enqueued))
	}
}

func TestWishlistServiceAddItemDoesNotFailWhenAnalyticsEnqueueFails(t *testing.T) {
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	eventRepository := &fakeWishlistEventRepository{err: errors.New("outbox unavailable")}
	service := mustWishlistServiceWithEvents(t, newFakeWishlistRepository(t), &fakeProductValidator{}, eventRepository, now)

	wishlist, err := service.AddItem(context.Background(), AddWishlistItemInput{
		UserID:    "user_123",
		ProductID: "prod_123",
	})
	if err != nil {
		t.Fatalf("AddItem returned error: %v", err)
	}
	if !wishlist.HasProduct("prod_123") {
		t.Fatal("wishlist mutation did not succeed after analytics enqueue failure")
	}
	if len(eventRepository.enqueued) != 0 {
		t.Fatalf("analytics events = %d, want 0 on enqueue error", len(eventRepository.enqueued))
	}
}

func TestWishlistServiceAddItemRequiresAuthenticatedUser(t *testing.T) {
	service := mustWishlistService(t, newFakeWishlistRepository(t), &fakeProductValidator{}, time.Now())

	_, err := service.AddItem(context.Background(), AddWishlistItemInput{ProductID: "prod_123"})
	if !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("AddItem error = %v, want ErrUnauthenticated", err)
	}
}

func TestWishlistServiceAddItemPropagatesProductValidationFailure(t *testing.T) {
	validator := &fakeProductValidator{err: ErrProductNotFound}
	service := mustWishlistService(t, newFakeWishlistRepository(t), validator, time.Now())

	_, err := service.AddItem(context.Background(), AddWishlistItemInput{
		UserID:    "user_123",
		ProductID: "prod_missing",
	})
	if !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("AddItem error = %v, want ErrProductNotFound", err)
	}
}

func TestWishlistServiceRemoveItemIsIdempotentAndSkipsProductValidation(t *testing.T) {
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	validator := &fakeProductValidator{}
	eventRepository := &fakeWishlistEventRepository{}
	service := mustWishlistServiceWithEvents(t, newFakeWishlistRepository(t), validator, eventRepository, now)

	wishlist, err := service.RemoveItem(context.Background(), RemoveWishlistItemInput{
		UserID:    "user_123",
		ProductID: "prod_absent",
	})
	if err != nil {
		t.Fatalf("RemoveItem returned error: %v", err)
	}
	if len(wishlist.Items) != 0 {
		t.Fatalf("len(wishlist.Items) = %d, want 0", len(wishlist.Items))
	}
	if validator.calls != 0 {
		t.Fatalf("product validator calls = %d, want 0", validator.calls)
	}
	if len(eventRepository.enqueued) != 0 {
		t.Fatalf("analytics events = %d, want 0 for absent remove", len(eventRepository.enqueued))
	}
}

func TestWishlistServiceRemoveItemEnqueuesAnalyticsEventForExistingItem(t *testing.T) {
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	repository := newFakeWishlistRepository(t)
	if err := repository.EnsureWishlist(context.Background(), "user_123", now); err != nil {
		t.Fatalf("EnsureWishlist returned error: %v", err)
	}
	if err := repository.AddItemIfNotExists(context.Background(), "user_123", mustWishlistItem(t, "prod_123", "var_1", now), now); err != nil {
		t.Fatalf("AddItemIfNotExists returned error: %v", err)
	}
	eventRepository := &fakeWishlistEventRepository{}
	service := mustWishlistServiceWithEvents(t, repository, &fakeProductValidator{}, eventRepository, now.Add(time.Minute))

	wishlist, err := service.RemoveItem(context.Background(), RemoveWishlistItemInput{
		UserID:    "user_123",
		ProductID: "prod_123",
		TraceID:   "req_remove",
	})
	if err != nil {
		t.Fatalf("RemoveItem returned error: %v", err)
	}
	if wishlist.HasProduct("prod_123") {
		t.Fatal("prod_123 still exists after remove")
	}
	if len(eventRepository.enqueued) != 1 {
		t.Fatalf("analytics events = %d, want 1", len(eventRepository.enqueued))
	}
	event := eventRepository.enqueued[0]
	if event.EventType != domain.WishlistEventItemRemoved || event.Payload.Action != domain.WishlistEventActionRemove {
		t.Fatalf("event type/action = %q/%q, want removed/remove", event.EventType, event.Payload.Action)
	}
	if event.Payload.UserID != "user_123" || event.Payload.ProductID != "prod_123" || event.Payload.VariantID != "var_1" {
		t.Fatalf("event payload = %#v, want removed item snapshot", event.Payload)
	}
	if event.TraceID != "req_remove" {
		t.Fatalf("event trace_id = %q, want req_remove", event.TraceID)
	}
}

func TestWishlistServiceMoveToCartAddsCartThenRemovesWishlistItem(t *testing.T) {
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	repository := newFakeWishlistRepository(t)
	item := mustWishlistItem(t, "prod_123", "var_1", now)
	if err := repository.EnsureWishlist(context.Background(), "user_123", now); err != nil {
		t.Fatalf("EnsureWishlist returned error: %v", err)
	}
	if err := repository.AddItemIfNotExists(context.Background(), "user_123", item, now); err != nil {
		t.Fatalf("AddItemIfNotExists returned error: %v", err)
	}
	cartClient := &fakeCartClient{
		cart: &Cart{
			CartID: "cart_123",
			UserID: "user_123",
			Items:  []CartItem{{ItemID: "item_1", ProductID: "prod_123", VariantID: "var_1", Quantity: 1}},
		},
	}
	service := mustWishlistServiceWithCart(t, repository, &fakeProductValidator{}, cartClient, now)

	cart, err := service.MoveToCart(context.Background(), MoveToCartInput{
		UserID:         " user_123 ",
		ProductID:      " prod_123 ",
		RequestID:      " req_test ",
		IdempotencyKey: " idem_test ",
	})
	if err != nil {
		t.Fatalf("MoveToCart returned error: %v", err)
	}
	if cart.CartID != "cart_123" {
		t.Fatalf("cart.CartID = %q, want cart_123", cart.CartID)
	}
	if cartClient.calls != 1 {
		t.Fatalf("cart calls = %d, want 1", cartClient.calls)
	}
	if cartClient.input.ProductID != "prod_123" || cartClient.input.VariantID != "var_1" || cartClient.input.Quantity != 1 {
		t.Fatalf("cart input = %#v, want prod_123 var_1 quantity 1", cartClient.input)
	}
	if cartClient.input.IdempotencyKey != "idem_test" {
		t.Fatalf("idempotency key = %q, want idem_test", cartClient.input.IdempotencyKey)
	}
	wishlist, err := repository.FindByUserID(context.Background(), "user_123")
	if err != nil {
		t.Fatalf("FindByUserID returned error: %v", err)
	}
	if wishlist.HasProduct("prod_123") {
		t.Fatal("prod_123 still exists in wishlist after move")
	}
}

func TestWishlistServiceMoveToCartRequiresWishlistItemAndVariantBeforeCartCall(t *testing.T) {
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	repository := newFakeWishlistRepository(t)
	cartClient := &fakeCartClient{}
	service := mustWishlistServiceWithCart(t, repository, &fakeProductValidator{}, cartClient, now)

	_, err := service.MoveToCart(context.Background(), MoveToCartInput{
		UserID:    "user_123",
		ProductID: "prod_missing",
	})
	if !errors.Is(err, domain.ErrWishlistItemNotFound) {
		t.Fatalf("MoveToCart missing item error = %v, want ErrWishlistItemNotFound", err)
	}
	if cartClient.calls != 0 {
		t.Fatalf("cart calls = %d, want 0", cartClient.calls)
	}

	if err := repository.EnsureWishlist(context.Background(), "user_123", now); err != nil {
		t.Fatalf("EnsureWishlist returned error: %v", err)
	}
	if err := repository.AddItemIfNotExists(context.Background(), "user_123", mustWishlistItem(t, "prod_123", "", now), now); err != nil {
		t.Fatalf("AddItemIfNotExists returned error: %v", err)
	}
	_, err = service.MoveToCart(context.Background(), MoveToCartInput{
		UserID:    "user_123",
		ProductID: "prod_123",
	})
	if !errors.Is(err, ErrVariantRequiredForCart) {
		t.Fatalf("MoveToCart missing variant error = %v, want ErrVariantRequiredForCart", err)
	}
	if cartClient.calls != 0 {
		t.Fatalf("cart calls = %d, want 0", cartClient.calls)
	}
}

func TestWishlistServiceMoveToCartKeepsWishlistItemWhenCartFails(t *testing.T) {
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	repository := newFakeWishlistRepository(t)
	if err := repository.EnsureWishlist(context.Background(), "user_123", now); err != nil {
		t.Fatalf("EnsureWishlist returned error: %v", err)
	}
	if err := repository.AddItemIfNotExists(context.Background(), "user_123", mustWishlistItem(t, "prod_123", "var_1", now), now); err != nil {
		t.Fatalf("AddItemIfNotExists returned error: %v", err)
	}
	service := mustWishlistServiceWithCart(t, repository, &fakeProductValidator{}, &fakeCartClient{err: ErrCartServiceUnavailable}, now)

	_, err := service.MoveToCart(context.Background(), MoveToCartInput{
		UserID:    "user_123",
		ProductID: "prod_123",
	})
	if !errors.Is(err, ErrCartServiceUnavailable) {
		t.Fatalf("MoveToCart error = %v, want ErrCartServiceUnavailable", err)
	}
	wishlist, err := repository.FindByUserID(context.Background(), "user_123")
	if err != nil {
		t.Fatalf("FindByUserID returned error: %v", err)
	}
	if !wishlist.HasProduct("prod_123") {
		t.Fatal("prod_123 was removed even though cart add failed")
	}
}

func TestWishlistServiceMoveToCartReturnsPartialFailureWhenCleanupFails(t *testing.T) {
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	repository := newFakeWishlistRepository(t)
	repository.removeErr = errors.New("mongo write failed")
	if err := repository.EnsureWishlist(context.Background(), "user_123", now); err != nil {
		t.Fatalf("EnsureWishlist returned error: %v", err)
	}
	if err := repository.AddItemIfNotExists(context.Background(), "user_123", mustWishlistItem(t, "prod_123", "var_1", now), now); err != nil {
		t.Fatalf("AddItemIfNotExists returned error: %v", err)
	}
	service := mustWishlistServiceWithCart(t, repository, &fakeProductValidator{}, &fakeCartClient{cart: &Cart{CartID: "cart_123"}}, now)

	_, err := service.MoveToCart(context.Background(), MoveToCartInput{
		UserID:    "user_123",
		ProductID: "prod_123",
	})
	if !errors.Is(err, ErrMoveToCartPartialFailure) {
		t.Fatalf("MoveToCart error = %v, want ErrMoveToCartPartialFailure", err)
	}
}

func TestWishlistServiceSyncProductAvailabilityUpdatesProductLevelItems(t *testing.T) {
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	eventTime := now.Add(time.Hour)
	repository := newFakeWishlistRepository(t)
	for _, userID := range []string{"user_1", "user_2"} {
		if err := repository.EnsureWishlist(context.Background(), userID, now); err != nil {
			t.Fatalf("EnsureWishlist returned error: %v", err)
		}
		if err := repository.AddItemIfNotExists(context.Background(), userID, mustWishlistItem(t, "prod_123", "var_1", now), now); err != nil {
			t.Fatalf("AddItemIfNotExists returned error: %v", err)
		}
	}
	service := mustWishlistService(t, repository, &fakeProductValidator{}, now)

	result, err := service.SyncProductAvailability(context.Background(), SyncProductAvailabilityInput{
		EventID:      " evt_123 ",
		EventType:    "ProductDeleted",
		ProductID:    " prod_123 ",
		Availability: domain.AvailabilityDeleted,
		OccurredAt:   eventTime,
		TraceID:      " trace_123 ",
	})
	if err != nil {
		t.Fatalf("SyncProductAvailability returned error: %v", err)
	}
	if result.ProductID != "prod_123" || result.Availability != domain.AvailabilityDeleted {
		t.Fatalf("result = %#v, want prod_123 deleted", result)
	}
	if result.MatchedCount != 2 || result.ModifiedCount != 2 {
		t.Fatalf("result counts = %d/%d, want 2/2", result.MatchedCount, result.ModifiedCount)
	}
	for _, userID := range []string{"user_1", "user_2"} {
		wishlist, err := repository.FindByUserID(context.Background(), userID)
		if err != nil {
			t.Fatalf("FindByUserID returned error: %v", err)
		}
		item, ok := wishlist.FindItem("prod_123")
		if !ok {
			t.Fatalf("%s missing prod_123", userID)
		}
		if item.Availability != domain.AvailabilityDeleted {
			t.Fatalf("%s availability = %q, want deleted", userID, item.Availability)
		}
	}
}

func TestWishlistServiceSyncProductAvailabilityUpdatesOnlyMatchingVariant(t *testing.T) {
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	repository := newFakeWishlistRepository(t)
	if err := repository.EnsureWishlist(context.Background(), "user_1", now); err != nil {
		t.Fatalf("EnsureWishlist returned error: %v", err)
	}
	if err := repository.AddItemIfNotExists(context.Background(), "user_1", mustWishlistItem(t, "prod_123", "var_1", now), now); err != nil {
		t.Fatalf("AddItemIfNotExists returned error: %v", err)
	}
	if err := repository.EnsureWishlist(context.Background(), "user_2", now); err != nil {
		t.Fatalf("EnsureWishlist returned error: %v", err)
	}
	if err := repository.AddItemIfNotExists(context.Background(), "user_2", mustWishlistItem(t, "prod_123", "var_2", now), now); err != nil {
		t.Fatalf("AddItemIfNotExists returned error: %v", err)
	}
	service := mustWishlistService(t, repository, &fakeProductValidator{}, now)

	result, err := service.SyncProductAvailability(context.Background(), SyncProductAvailabilityInput{
		EventID:      "evt_456",
		EventType:    "ProductOutOfStock",
		ProductID:    "prod_123",
		VariantID:    "var_2",
		Availability: domain.AvailabilityOutOfStock,
	})
	if err != nil {
		t.Fatalf("SyncProductAvailability returned error: %v", err)
	}
	if result.MatchedCount != 1 || result.ModifiedCount != 1 {
		t.Fatalf("result counts = %d/%d, want 1/1", result.MatchedCount, result.ModifiedCount)
	}

	user1, _ := repository.FindByUserID(context.Background(), "user_1")
	item1, _ := user1.FindItem("prod_123")
	if item1.Availability != domain.AvailabilityInStock {
		t.Fatalf("user_1 availability = %q, want in_stock", item1.Availability)
	}
	user2, _ := repository.FindByUserID(context.Background(), "user_2")
	item2, _ := user2.FindItem("prod_123")
	if item2.Availability != domain.AvailabilityOutOfStock {
		t.Fatalf("user_2 availability = %q, want out_of_stock", item2.Availability)
	}
}

func TestWishlistServiceSyncProductAvailabilityValidatesInput(t *testing.T) {
	service := mustWishlistService(t, newFakeWishlistRepository(t), &fakeProductValidator{}, time.Now())

	_, err := service.SyncProductAvailability(context.Background(), SyncProductAvailabilityInput{
		Availability: domain.AvailabilityDeleted,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("empty product id error = %v, want ErrInvalidInput", err)
	}

	_, err = service.SyncProductAvailability(context.Background(), SyncProductAvailabilityInput{
		ProductID:    "prod_123",
		Availability: domain.Availability("invalid"),
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("invalid availability error = %v, want ErrInvalidInput", err)
	}
}

func mustWishlistService(t *testing.T, repository WishlistRepository, validator ProductValidator, now time.Time) *WishlistService {
	t.Helper()
	return mustWishlistServiceWithCart(t, repository, validator, &fakeCartClient{}, now)
}

func mustWishlistServiceWithCart(t *testing.T, repository WishlistRepository, validator ProductValidator, cartClient CartClient, now time.Time) *WishlistService {
	t.Helper()
	service, err := NewWishlistService(repository, validator, cartClient, nil, WithClock(func() time.Time {
		return now
	}))
	if err != nil {
		t.Fatalf("NewWishlistService returned error: %v", err)
	}
	return service
}

func mustWishlistServiceWithEvents(t *testing.T, repository WishlistRepository, validator ProductValidator, eventRepository WishlistEventRepository, now time.Time) *WishlistService {
	t.Helper()
	service, err := NewWishlistService(
		repository,
		validator,
		&fakeCartClient{},
		nil,
		WithClock(func() time.Time { return now }),
		WithWishlistEventRepository(eventRepository),
		WithWishlistEventTopic("recommendation.events.test"),
		WithWishlistEventIDFactory(func() string { return "evt_test" }),
	)
	if err != nil {
		t.Fatalf("NewWishlistService returned error: %v", err)
	}
	return service
}

func mustWishlistItem(t *testing.T, productID, variantID string, addedAt time.Time) domain.WishlistItem {
	t.Helper()
	item, err := domain.NewWishlistItem(domain.AddWishlistItemInput{
		ProductID:    productID,
		VariantID:    variantID,
		Availability: domain.AvailabilityInStock,
	}, addedAt)
	if err != nil {
		t.Fatalf("NewWishlistItem returned error: %v", err)
	}
	return item
}

type fakeProductValidator struct {
	snapshot ProductSnapshot
	err      error
	calls    int
}

func (f *fakeProductValidator) ValidateWishlistProduct(ctx context.Context, productID, variantID string) (ProductSnapshot, error) {
	f.calls++
	if f.err != nil {
		return ProductSnapshot{}, f.err
	}
	if f.snapshot.ProductID == "" {
		f.snapshot.ProductID = productID
	}
	return f.snapshot, nil
}

type fakeCartClient struct {
	cart  *Cart
	err   error
	input CartAddItemInput
	calls int
}

func (f *fakeCartClient) AddItem(ctx context.Context, input CartAddItemInput) (*Cart, error) {
	f.calls++
	f.input = input
	if f.err != nil {
		return nil, f.err
	}
	if f.cart != nil {
		return f.cart, nil
	}
	return &Cart{CartID: "cart_test", UserID: input.UserID}, nil
}

type fakeWishlistRepository struct {
	t               *testing.T
	wishlists       map[string]*domain.Wishlist
	removeErr       error
	availabilityErr error
}

type fakeWishlistEventRepository struct {
	enqueued []domain.WishlistAnalyticsEvent
	err      error
}

func (f *fakeWishlistEventRepository) Enqueue(ctx context.Context, event domain.WishlistAnalyticsEvent) error {
	if f.err != nil {
		return f.err
	}
	f.enqueued = append(f.enqueued, event)
	return nil
}

func (f *fakeWishlistEventRepository) ClaimPending(ctx context.Context, limit int, now time.Time) ([]domain.WishlistAnalyticsEvent, error) {
	return nil, nil
}

func (f *fakeWishlistEventRepository) MarkPublished(ctx context.Context, eventID string, publishedAt time.Time) error {
	return nil
}

func (f *fakeWishlistEventRepository) MarkRetry(ctx context.Context, eventID string, attempts int, nextRetryAt time.Time, lastError string) error {
	return nil
}

func (f *fakeWishlistEventRepository) MarkFailed(ctx context.Context, eventID string, attempts int, lastError string, failedAt time.Time) error {
	return nil
}

func newFakeWishlistRepository(t *testing.T) *fakeWishlistRepository {
	t.Helper()
	return &fakeWishlistRepository{
		t:         t,
		wishlists: make(map[string]*domain.Wishlist),
	}
}

func (f *fakeWishlistRepository) EnsureWishlist(ctx context.Context, userID string, now time.Time) error {
	if _, exists := f.wishlists[userID]; exists {
		return nil
	}
	wishlist, err := domain.NewWishlist("wish_"+userID, userID, now)
	if err != nil {
		return err
	}
	f.wishlists[userID] = wishlist
	return nil
}

func (f *fakeWishlistRepository) AddItemIfNotExists(ctx context.Context, userID string, item domain.WishlistItem, now time.Time) error {
	wishlist, ok := f.wishlists[userID]
	if !ok {
		return fmt.Errorf("%w: %s", domain.ErrWishlistNotFound, userID)
	}
	return wishlist.AddItem(domain.AddWishlistItemInput{
		ProductID:      item.ProductID,
		VariantID:      item.VariantID,
		LastKnownPrice: item.LastKnownPrice,
		Availability:   item.Availability,
	}, item.AddedAt)
}

func (f *fakeWishlistRepository) RemoveItem(ctx context.Context, userID, productID string, now time.Time) error {
	if f.removeErr != nil {
		return f.removeErr
	}
	wishlist, ok := f.wishlists[userID]
	if !ok {
		return nil
	}
	if err := wishlist.RemoveProduct(productID, now); err != nil && !errors.Is(err, domain.ErrWishlistItemNotFound) {
		return err
	}
	return nil
}

func (f *fakeWishlistRepository) FindByUserID(ctx context.Context, userID string) (*domain.Wishlist, error) {
	wishlist, ok := f.wishlists[userID]
	if !ok {
		return nil, fmt.Errorf("%w: %s", domain.ErrWishlistNotFound, userID)
	}
	snapshot := wishlist.Snapshot()
	return domain.RehydrateWishlist(snapshot)
}

func (f *fakeWishlistRepository) UpdateProductAvailability(ctx context.Context, productID, variantID string, availability domain.Availability, updatedAt time.Time) (domain.AvailabilityUpdateResult, error) {
	if f.availabilityErr != nil {
		return domain.AvailabilityUpdateResult{}, f.availabilityErr
	}
	var matched int64
	var modified int64
	for _, wishlist := range f.wishlists {
		wishlistMatched := false
		for idx := range wishlist.Items {
			item := &wishlist.Items[idx]
			if item.ProductID != productID {
				continue
			}
			if variantID != "" && item.VariantID != variantID {
				continue
			}
			if item.AddedAt.After(updatedAt) {
				continue
			}
			item.Availability = availability
			wishlistMatched = true
		}
		if wishlistMatched {
			wishlist.UpdatedAt = updatedAt
			matched++
			modified++
		}
	}
	return domain.AvailabilityUpdateResult{MatchedCount: matched, ModifiedCount: modified}, nil
}
