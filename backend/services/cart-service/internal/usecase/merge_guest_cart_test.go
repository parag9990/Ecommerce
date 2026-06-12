package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
)

func TestValidateMergeGuestCartCommandRequiresAuthenticatedUser(t *testing.T) {
	err := ValidateMergeGuestCartCommand(MergeGuestCartCommand{GuestCartID: "cart_guest"})
	if !errors.Is(err, domain.ErrUserIDRequired) {
		t.Fatalf("ValidateMergeGuestCartCommand() error = %v, want ErrUserIDRequired", err)
	}
}

func TestValidateMergeGuestCartCommandRequiresGuestCartID(t *testing.T) {
	err := ValidateMergeGuestCartCommand(MergeGuestCartCommand{UserID: "user_123"})
	if !errors.Is(err, domain.ErrGuestCartIDRequired) {
		t.Fatalf("ValidateMergeGuestCartCommand() error = %v, want ErrGuestCartIDRequired", err)
	}
}

func TestMergeGuestCartCreatesUserCartWhenMissing(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	guest := validMergeGuestCart("cart_guest", "sess_123", now)
	repo := newFakeMergeRepository(guest)
	cache := &fakeMergeCartCache{}
	uc := newTestMergeGuestCartUsecase(t, repo, cache, now)

	cart, err := uc.MergeGuestCart(context.Background(), MergeGuestCartCommand{
		UserID:         "user_123",
		GuestSessionID: "sess_123",
		GuestCartID:    "cart_guest",
	})
	if err != nil {
		t.Fatalf("MergeGuestCart() error = %v", err)
	}
	if cart.ID != "cart_merge_new" || cart.UserID == nil || *cart.UserID != "user_123" {
		t.Fatalf("cart = %#v, want new user cart", cart)
	}
	if len(cart.Items) != 1 || cart.Items[0].ItemID != "item_merge_new" {
		t.Fatalf("items = %#v, want copied guest item with new id", cart.Items)
	}
	storedGuest := repo.carts["cart_guest"]
	if storedGuest.Status != domain.CartStatusMerged || storedGuest.MergedIntoCartID == nil || *storedGuest.MergedIntoCartID != cart.ID {
		t.Fatalf("guest cart = %#v, want merged into user cart", storedGuest)
	}
	if !cache.activeSet || !cache.summarySet || !cache.activeDeleted || !cache.summaryDeleted {
		t.Fatalf("cache flags = %#v, want user refresh and guest invalidation", cache)
	}
}

func TestMergeGuestCartMergesIntoExistingUserCartAndClearsCoupon(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	user := validMergeUserCart("cart_user", "user_123", now)
	code := "SAVE10"
	user.CouponCode = &code
	user.CouponPreview = &domain.CouponView{
		CouponID: "coupon_123",
		Code:     code,
		Valid:    true,
		Discount: domain.NewMoney(100, domain.CurrencyINR),
	}
	user.Totals = domain.CartTotals{
		Subtotal:        domain.NewMoney(2000, domain.CurrencyINR),
		Discount:        domain.NewMoney(100, domain.CurrencyINR),
		Total:           domain.NewMoney(1900, domain.CurrencyINR),
		Currency:        domain.CurrencyINR,
		ItemCount:       2,
		UniqueItemCount: 1,
	}
	guest := validMergeGuestCart("cart_guest", "sess_123", now)
	guest.Items[0].Quantity = 3
	guest.Items[0].LineSubtotal = domain.NewMoney(3000, domain.CurrencyINR)
	guest.Totals = domain.CartTotals{
		Subtotal:        domain.NewMoney(3000, domain.CurrencyINR),
		Discount:        domain.NewMoney(0, domain.CurrencyINR),
		Total:           domain.NewMoney(3000, domain.CurrencyINR),
		Currency:        domain.CurrencyINR,
		ItemCount:       3,
		UniqueItemCount: 1,
	}
	repo := newFakeMergeRepository(user, guest)
	uc := newTestMergeGuestCartUsecase(t, repo, &fakeMergeCartCache{}, now)

	cart, err := uc.MergeGuestCart(context.Background(), MergeGuestCartCommand{
		UserID:         "user_123",
		GuestSessionID: "sess_123",
		GuestCartID:    "cart_guest",
	})
	if err != nil {
		t.Fatalf("MergeGuestCart() error = %v", err)
	}
	if cart.Items[0].Quantity != 5 || cart.Totals.Total.Amount != 5000 {
		t.Fatalf("cart = %#v, want combined quantity and recalculated totals", cart)
	}
	if cart.CouponCode != nil || cart.CouponPreview != nil || cart.Totals.Discount.Amount != 0 {
		t.Fatalf("coupon fields/totals were not cleared: %#v", cart)
	}
	if repo.expectedUserVersion != 4 || repo.expectedGuestVersion != 2 {
		t.Fatalf("versions = user %d guest %d, want 4/2", repo.expectedUserVersion, repo.expectedGuestVersion)
	}
}

func TestMergeGuestCartRejectsGuestSessionMismatch(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	repo := newFakeMergeRepository(validMergeGuestCart("cart_guest", "sess_123", now))
	uc := newTestMergeGuestCartUsecase(t, repo, &fakeMergeCartCache{}, now)

	_, err := uc.MergeGuestCart(context.Background(), MergeGuestCartCommand{
		UserID:         "user_123",
		GuestSessionID: "sess_other",
		GuestCartID:    "cart_guest",
	})
	if !errors.Is(err, domain.ErrGuestCartAccessDenied) {
		t.Fatalf("MergeGuestCart() error = %v, want ErrGuestCartAccessDenied", err)
	}
}

func TestMergeGuestCartIsIdempotentForAlreadyMergedGuestCart(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	user := validMergeUserCart("cart_user", "user_123", now)
	guest := validMergeGuestCart("cart_guest", "sess_123", now)
	targetID := "cart_user"
	guest.Status = domain.CartStatusMerged
	guest.MergedIntoCartID = &targetID
	repo := newFakeMergeRepository(user, guest)
	cache := &fakeMergeCartCache{}
	uc := newTestMergeGuestCartUsecase(t, repo, cache, now)

	cart, err := uc.MergeGuestCart(context.Background(), MergeGuestCartCommand{
		UserID:         "user_123",
		GuestSessionID: "sess_123",
		GuestCartID:    "cart_guest",
	})
	if err != nil {
		t.Fatalf("MergeGuestCart() error = %v", err)
	}
	if cart.ID != "cart_user" {
		t.Fatalf("cart.ID = %q, want existing merged target", cart.ID)
	}
	if repo.mergeCalls != 0 {
		t.Fatalf("mergeCalls = %d, want no write for idempotent response", repo.mergeCalls)
	}
	if !cache.activeSet || !cache.activeDeleted {
		t.Fatal("cache was not refreshed/invalidated for idempotent merge")
	}
}

func TestMergeGuestCartRetriesVersionConflict(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	repo := newFakeMergeRepository(validMergeUserCart("cart_user", "user_123", now), validMergeGuestCart("cart_guest", "sess_123", now))
	repo.conflictOnce = true
	uc := newTestMergeGuestCartUsecase(t, repo, &fakeMergeCartCache{}, now)

	cart, err := uc.MergeGuestCart(context.Background(), MergeGuestCartCommand{
		UserID:         "user_123",
		GuestSessionID: "sess_123",
		GuestCartID:    "cart_guest",
	})
	if err != nil {
		t.Fatalf("MergeGuestCart() error = %v", err)
	}
	if repo.mergeCalls != 2 {
		t.Fatalf("mergeCalls = %d, want 2", repo.mergeCalls)
	}
	if cart.Version != 5 {
		t.Fatalf("cart.Version = %d, want 5", cart.Version)
	}
}

func TestMergeGuestCartReturnsCartWhenCacheFails(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	repo := newFakeMergeRepository(validMergeUserCart("cart_user", "user_123", now), validMergeGuestCart("cart_guest", "sess_123", now))
	uc := newTestMergeGuestCartUsecase(t, repo, &fakeMergeCartCache{err: domain.ErrCacheUnavailable}, now)

	cart, err := uc.MergeGuestCart(context.Background(), MergeGuestCartCommand{
		UserID:         "user_123",
		GuestSessionID: "sess_123",
		GuestCartID:    "cart_guest",
	})
	if err != nil {
		t.Fatalf("MergeGuestCart() error = %v", err)
	}
	if cart == nil || cart.ID != "cart_user" {
		t.Fatalf("cart = %#v, want merged user cart", cart)
	}
}

func TestMergeGuestCartRejectsExpiredGuestCart(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	guest := validMergeGuestCart("cart_guest", "sess_123", now)
	guest.ExpiresAt = now.Add(-time.Second)
	repo := newFakeMergeRepository(guest)
	cache := &fakeMergeCartCache{}
	uc := newTestMergeGuestCartUsecase(t, repo, cache, now)

	_, err := uc.MergeGuestCart(context.Background(), MergeGuestCartCommand{
		UserID:         "user_123",
		GuestSessionID: "sess_123",
		GuestCartID:    "cart_guest",
	})
	if !errors.Is(err, domain.ErrCartExpired) {
		t.Fatalf("MergeGuestCart() error = %v, want ErrCartExpired", err)
	}
	if repo.expireCalls != 1 || repo.carts["cart_guest"].Status != domain.CartStatusExpired {
		t.Fatalf("guest cart = %#v expireCalls=%d, want expired", repo.carts["cart_guest"], repo.expireCalls)
	}
	if !cache.activeDeleted || !cache.summaryDeleted {
		t.Fatalf("cache flags = %#v, want expired guest cache deleted", cache)
	}
}

func newTestMergeGuestCartUsecase(t *testing.T, repo *fakeMergeCartRepository, cache *fakeMergeCartCache, now time.Time) *MergeGuestCartUsecase {
	t.Helper()
	uc, err := NewMergeGuestCartUsecase(MergeGuestCartDependencies{
		Repository:      repo,
		Cache:           cache,
		IDGenerator:     &fakeMergeIDGenerator{},
		Clock:           fakeClock{now: now},
		CartTTL:         90 * 24 * time.Hour,
		DefaultCurrency: domain.CurrencyINR,
		MaxSaveAttempts: 3,
	})
	if err != nil {
		t.Fatalf("NewMergeGuestCartUsecase() error = %v", err)
	}
	return uc
}

type fakeMergeCartRepository struct {
	carts                map[string]*domain.Cart
	mergeCalls           int
	expireCalls          int
	conflictOnce         bool
	expectedUserVersion  int64
	expectedGuestVersion int64
}

func newFakeMergeRepository(carts ...*domain.Cart) *fakeMergeCartRepository {
	repo := &fakeMergeCartRepository{carts: map[string]*domain.Cart{}}
	for _, cart := range carts {
		repo.carts[cart.ID] = cloneCart(cart)
	}
	return repo
}

func (f *fakeMergeCartRepository) FindCartByID(ctx context.Context, cartID string) (*domain.Cart, error) {
	cart := f.carts[cartID]
	if cart == nil {
		return nil, domain.ErrCartNotFound
	}
	return cloneCart(cart), nil
}

func (f *fakeMergeCartRepository) FindActiveByOwner(ctx context.Context, owner domain.CartOwner) (*domain.Cart, error) {
	for _, cart := range f.carts {
		if cart.Status != domain.CartStatusActive {
			continue
		}
		if owner.IsUser() && cart.UserID != nil && *cart.UserID == owner.UserID {
			return cloneCart(cart), nil
		}
		if owner.IsGuest() && cart.GuestSessionID != nil && *cart.GuestSessionID == owner.GuestSessionID {
			return cloneCart(cart), nil
		}
	}
	return nil, domain.ErrCartNotFound
}

func (f *fakeMergeCartRepository) MergeGuestIntoUser(ctx context.Context, userCart *domain.Cart, guestCart *domain.Cart, expectedUserVersion int64, expectedGuestVersion int64) error {
	f.mergeCalls++
	f.expectedUserVersion = expectedUserVersion
	f.expectedGuestVersion = expectedGuestVersion
	if f.conflictOnce {
		f.conflictOnce = false
		return domain.ErrCartVersionConflict
	}
	if expectedUserVersion == 0 {
		for _, cart := range f.carts {
			if cart.Status == domain.CartStatusActive && cart.UserID != nil && userCart.UserID != nil && *cart.UserID == *userCart.UserID {
				return domain.ErrActiveCartExists
			}
		}
	} else {
		storedUser := f.carts[userCart.ID]
		if storedUser == nil || storedUser.Version != expectedUserVersion || storedUser.Status != domain.CartStatusActive {
			return domain.ErrCartVersionConflict
		}
		userCart.Version = expectedUserVersion + 1
	}
	storedGuest := f.carts[guestCart.ID]
	if storedGuest == nil || storedGuest.Version != expectedGuestVersion || storedGuest.Status != domain.CartStatusActive {
		return domain.ErrCartVersionConflict
	}
	guestCart.Version = expectedGuestVersion + 1
	f.carts[userCart.ID] = cloneCart(userCart)
	f.carts[guestCart.ID] = cloneCart(guestCart)
	return nil
}

func (f *fakeMergeCartRepository) ExpireActiveCart(ctx context.Context, cartID string, expectedVersion int64, now time.Time) (bool, error) {
	f.expireCalls++
	stored := f.carts[cartID]
	if stored == nil || stored.Version != expectedVersion || stored.Status != domain.CartStatusActive {
		return false, nil
	}
	if !domain.IsActiveCartExpired(stored, now) {
		return false, nil
	}
	stored.Status = domain.CartStatusExpired
	stored.UpdatedAt = now
	stored.Version++
	return true, nil
}

type fakeMergeCartCache struct {
	err            error
	activeSet      bool
	summarySet     bool
	activeDeleted  bool
	summaryDeleted bool
}

func (f *fakeMergeCartCache) SetActiveCart(ctx context.Context, owner domain.CartOwner, cart *domain.Cart) error {
	f.activeSet = true
	return f.err
}

func (f *fakeMergeCartCache) SetCartSummary(ctx context.Context, owner domain.CartOwner, summary domain.CartSummary) error {
	f.summarySet = true
	return f.err
}

func (f *fakeMergeCartCache) DeleteActiveCart(ctx context.Context, owner domain.CartOwner) error {
	f.activeDeleted = true
	return f.err
}

func (f *fakeMergeCartCache) DeleteCartSummary(ctx context.Context, owner domain.CartOwner) error {
	f.summaryDeleted = true
	return f.err
}

type fakeMergeIDGenerator struct{}

func (*fakeMergeIDGenerator) NewCartID() (string, error) {
	return "cart_merge_new", nil
}

func (*fakeMergeIDGenerator) NewItemID() (string, error) {
	return "item_merge_new", nil
}

func validMergeUserCart(cartID string, userID string, now time.Time) *domain.Cart {
	return &domain.Cart{
		ID:        cartID,
		UserID:    &userID,
		Status:    domain.CartStatusActive,
		Items:     []domain.CartItem{validUsecaseItem(now)},
		Totals:    validUsecaseTotals(),
		Version:   4,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(90 * 24 * time.Hour),
	}
}

func validMergeGuestCart(cartID string, guestSessionID string, now time.Time) *domain.Cart {
	return &domain.Cart{
		ID:             cartID,
		GuestSessionID: &guestSessionID,
		Status:         domain.CartStatusActive,
		Items:          []domain.CartItem{validUsecaseItem(now)},
		Totals:         validUsecaseTotals(),
		Version:        2,
		CreatedAt:      now,
		UpdatedAt:      now,
		ExpiresAt:      now.Add(90 * 24 * time.Hour),
	}
}
