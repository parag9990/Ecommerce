package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
)

func TestAddItemCreatesActiveCartWhenMissing(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	repo := &fakeCartRepository{}
	cache := &fakeCartCache{}
	uc := newTestAddItemUsecase(t, repo, cache, now)

	cart, err := uc.AddItem(context.Background(), AddItemCommand{
		UserID:    "user_123",
		ProductID: "prod_123",
		VariantID: "var_1",
		Quantity:  2,
	})
	if err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}
	if cart.ID != "cart_1" || len(cart.Items) != 1 {
		t.Fatalf("unexpected cart: %#v", cart)
	}
	if cart.Items[0].Quantity != 2 || cart.Totals.Subtotal.Amount != 2000 {
		t.Fatalf("unexpected cart totals/items: %#v", cart)
	}
	if !cache.activeSet || !cache.summarySet {
		t.Fatal("cache was not refreshed")
	}
}

func TestAddItemIncrementsExistingVariant(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	existing := &domain.Cart{
		ID:        "cart_123",
		UserID:    &userID,
		Status:    domain.CartStatusActive,
		Items:     []domain.CartItem{validUsecaseItem(now)},
		Totals:    validUsecaseTotals(),
		Version:   4,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(90 * 24 * time.Hour),
	}
	repo := &fakeCartRepository{cart: existing}
	uc := newTestAddItemUsecase(t, repo, &fakeCartCache{}, now)

	cart, err := uc.AddItem(context.Background(), AddItemCommand{
		UserID:    userID,
		ProductID: "prod_123",
		VariantID: "var_1",
		Quantity:  1,
	})
	if err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}
	if cart.Items[0].Quantity != 3 {
		t.Fatalf("quantity = %d, want 3", cart.Items[0].Quantity)
	}
	if repo.savedExpectedVersion != 4 || cart.Version != 5 {
		t.Fatalf("version = %d expectedSave = %d, want 5/4", cart.Version, repo.savedExpectedVersion)
	}
}

func TestAddItemRejectsOutOfStockVariant(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	product := validProductForCart()
	product.Variants[0].StockQuantity = 1
	uc := newTestAddItemUsecaseWithProduct(t, &fakeCartRepository{}, &fakeCartCache{}, now, product)

	_, err := uc.AddItem(context.Background(), AddItemCommand{
		UserID:    "user_123",
		ProductID: "prod_123",
		VariantID: "var_1",
		Quantity:  2,
	})
	if !errors.Is(err, domain.ErrInsufficientStock) {
		t.Fatalf("AddItem() error = %v, want ErrInsufficientStock", err)
	}
}

func TestAddItemReturnsCartWhenCacheFails(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	uc := newTestAddItemUsecase(t, &fakeCartRepository{}, &fakeCartCache{err: domain.ErrCacheUnavailable}, now)

	cart, err := uc.AddItem(context.Background(), AddItemCommand{
		UserID:    "user_123",
		ProductID: "prod_123",
		VariantID: "var_1",
		Quantity:  1,
	})
	if err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}
	if cart == nil || len(cart.Items) != 1 {
		t.Fatalf("cart = %#v, want updated cart", cart)
	}
}

func TestAddItemRetriesVersionConflict(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	repo := &fakeCartRepository{
		cart: &domain.Cart{
			ID:        "cart_123",
			UserID:    &userID,
			Status:    domain.CartStatusActive,
			Items:     []domain.CartItem{},
			Totals:    emptyUsecaseTotals(),
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
			ExpiresAt: now.Add(90 * 24 * time.Hour),
		},
		conflictOnce: true,
	}
	uc := newTestAddItemUsecase(t, repo, &fakeCartCache{}, now)

	cart, err := uc.AddItem(context.Background(), AddItemCommand{
		UserID:    userID,
		ProductID: "prod_123",
		VariantID: "var_1",
		Quantity:  1,
	})
	if err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}
	if repo.saveCalls != 2 {
		t.Fatalf("saveCalls = %d, want 2", repo.saveCalls)
	}
	if cart.Version != 2 {
		t.Fatalf("version = %d, want 2", cart.Version)
	}
}

func TestAddItemClearsStaleCouponPreview(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	cart := validUsecaseCart(userID, now)
	code := "SAVE10"
	cart.CouponCode = &code
	cart.CouponPreview = &domain.CouponView{
		CouponID: "coupon_save10",
		Code:     code,
		Valid:    true,
		Discount: domain.NewMoney(200, domain.CurrencyINR),
	}
	cart.Totals = domain.CartTotals{
		Subtotal:        domain.NewMoney(2000, domain.CurrencyINR),
		Discount:        domain.NewMoney(200, domain.CurrencyINR),
		Total:           domain.NewMoney(1800, domain.CurrencyINR),
		Currency:        domain.CurrencyINR,
		ItemCount:       2,
		UniqueItemCount: 1,
	}
	uc := newTestAddItemUsecase(t, &fakeCartRepository{cart: cart}, &fakeCartCache{}, now)

	updated, err := uc.AddItem(context.Background(), AddItemCommand{
		UserID:    userID,
		ProductID: "prod_123",
		VariantID: "var_1",
		Quantity:  1,
	})
	if err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}
	if updated.CouponCode != nil || updated.CouponPreview != nil {
		t.Fatalf("coupon fields were not cleared: code=%v preview=%v", updated.CouponCode, updated.CouponPreview)
	}
	if updated.Totals.Discount.Amount != 0 || updated.Totals.Total.Amount != 3000 {
		t.Fatalf("unexpected totals after coupon clear: %#v", updated.Totals)
	}
}

func TestAddItemReplacesExpiredActiveCart(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	expired := validUsecaseCart(userID, now)
	expired.ID = "cart_expired"
	expired.ExpiresAt = now.Add(-time.Second)
	repo := &fakeCartRepository{cart: expired}
	cache := &fakeCartCache{}
	uc := newTestAddItemUsecase(t, repo, cache, now)

	cart, err := uc.AddItem(context.Background(), AddItemCommand{
		UserID:    userID,
		ProductID: "prod_123",
		VariantID: "var_1",
		Quantity:  1,
	})
	if err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}
	if cart.ID != "cart_1" || cart.ExpiresAt.Sub(now) != 90*24*time.Hour {
		t.Fatalf("cart = %#v, want fresh active cart", cart)
	}
	if repo.expireCalls != 1 {
		t.Fatalf("expireCalls = %d, want 1", repo.expireCalls)
	}
	if !cache.activeDeleted || !cache.summaryDeleted || !cache.activeSet || !cache.summarySet {
		t.Fatalf("cache flags = %#v, want expired invalidation and fresh refresh", cache)
	}
}

func TestAddItemUsesGuestExpiryTTLWhenConfigured(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	repo := &fakeCartRepository{}
	cache := &fakeCartCache{}
	uc, err := NewAddItemUsecase(AddItemDependencies{
		Repository:    repo,
		Cache:         cache,
		ProductClient: &fakeProductClient{product: validProductForCart()},
		IDGenerator:   &fakeIDGenerator{},
		Clock:         fakeClock{now: now},
		UserCartTTL:   90 * 24 * time.Hour,
		GuestCartTTL:  30 * 24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("NewAddItemUsecase() error = %v", err)
	}

	cart, err := uc.AddItem(context.Background(), AddItemCommand{
		GuestSessionID: "sess_123",
		ProductID:      "prod_123",
		VariantID:      "var_1",
		Quantity:       1,
	})
	if err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}
	if got := cart.ExpiresAt.Sub(now); got != 30*24*time.Hour {
		t.Fatalf("guest cart ttl = %s, want 720h", got)
	}
}

func TestValidateAddItemCommandRejectsMissingProduct(t *testing.T) {
	err := ValidateAddItemCommand(AddItemCommand{VariantID: "var_1", Quantity: 1})
	if !errors.Is(err, domain.ErrProductIDRequired) {
		t.Fatalf("ValidateAddItemCommand() error = %v, want ErrProductIDRequired", err)
	}
}

func newTestAddItemUsecase(t *testing.T, repo *fakeCartRepository, cache *fakeCartCache, now time.Time) *AddItemUsecase {
	t.Helper()
	return newTestAddItemUsecaseWithProduct(t, repo, cache, now, validProductForCart())
}

func newTestAddItemUsecaseWithProduct(t *testing.T, repo *fakeCartRepository, cache *fakeCartCache, now time.Time, product *ProductForCart) *AddItemUsecase {
	t.Helper()
	uc, err := NewAddItemUsecase(AddItemDependencies{
		Repository:      repo,
		Cache:           cache,
		ProductClient:   &fakeProductClient{product: product},
		IDGenerator:     &fakeIDGenerator{},
		Clock:           fakeClock{now: now},
		CartTTL:         90 * 24 * time.Hour,
		MaxSaveAttempts: 3,
	})
	if err != nil {
		t.Fatalf("NewAddItemUsecase() error = %v", err)
	}
	return uc
}

type fakeCartRepository struct {
	cart                 *domain.Cart
	savedExpectedVersion int64
	saveCalls            int
	expireCalls          int
	conflictOnce         bool
}

func (f *fakeCartRepository) FindActiveByOwner(ctx context.Context, owner domain.CartOwner) (*domain.Cart, error) {
	if f.cart == nil {
		return nil, domain.ErrCartNotFound
	}
	return cloneCart(f.cart), nil
}

func (f *fakeCartRepository) CreateActiveCart(ctx context.Context, cart *domain.Cart) error {
	f.cart = cloneCart(cart)
	return nil
}

func (f *fakeCartRepository) SaveWithVersion(ctx context.Context, cart *domain.Cart, expectedVersion int64) error {
	f.saveCalls++
	f.savedExpectedVersion = expectedVersion
	if f.conflictOnce {
		f.conflictOnce = false
		return domain.ErrCartVersionConflict
	}
	cart.Version = expectedVersion + 1
	f.cart = cloneCart(cart)
	return nil
}

func (f *fakeCartRepository) ExpireActiveCart(ctx context.Context, cartID string, expectedVersion int64, now time.Time) (bool, error) {
	f.expireCalls++
	if f.cart == nil || f.cart.ID != cartID || f.cart.Version != expectedVersion || f.cart.Status != domain.CartStatusActive {
		return false, nil
	}
	if !domain.IsActiveCartExpired(f.cart, now) {
		return false, nil
	}
	f.cart.Status = domain.CartStatusExpired
	f.cart.UpdatedAt = now
	f.cart.Version++
	return true, nil
}

type fakeCartCache struct {
	err            error
	activeSet      bool
	summarySet     bool
	activeDeleted  bool
	summaryDeleted bool
}

func (f *fakeCartCache) SetActiveCart(ctx context.Context, owner domain.CartOwner, cart *domain.Cart) error {
	f.activeSet = true
	return f.err
}

func (f *fakeCartCache) SetCartSummary(ctx context.Context, owner domain.CartOwner, summary domain.CartSummary) error {
	f.summarySet = true
	return f.err
}

func (f *fakeCartCache) DeleteActiveCart(ctx context.Context, owner domain.CartOwner) error {
	f.activeDeleted = true
	return f.err
}

func (f *fakeCartCache) DeleteCartSummary(ctx context.Context, owner domain.CartOwner) error {
	f.summaryDeleted = true
	return f.err
}

type fakeProductClient struct {
	product *ProductForCart
	err     error
}

func (f *fakeProductClient) GetProductForCart(ctx context.Context, productID string) (*ProductForCart, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.product, nil
}

type fakeIDGenerator struct{}

func (*fakeIDGenerator) NewCartID() (string, error) {
	return "cart_1", nil
}

func (*fakeIDGenerator) NewItemID() (string, error) {
	return "item_1", nil
}

type fakeClock struct {
	now time.Time
}

func (f fakeClock) Now() time.Time {
	return f.now
}

func validProductForCart() *ProductForCart {
	return &ProductForCart{
		ID:       "prod_123",
		SellerID: "seller_456",
		Title:    "Running Shoes",
		ImageURL: "https://cdn.example.com/prod_123/main.jpg",
		Status:   "published",
		Variants: []ProductVariantForCart{
			{
				ID:            "var_1",
				SKU:           "RUN-BLK-9",
				Attributes:    map[string]string{"size": "9", "color": "black"},
				PriceAmount:   1000,
				Currency:      domain.CurrencyINR,
				StockQuantity: 10,
				Available:     true,
				Status:        "active",
			},
		},
	}
}

func validUsecaseItem(now time.Time) domain.CartItem {
	return domain.CartItem{
		ItemID:          "item_1",
		ProductID:       "prod_123",
		VariantID:       "var_1",
		SellerID:        "seller_456",
		TitleSnapshot:   "Running Shoes",
		UnitPrice:       domain.NewMoney(1000, domain.CurrencyINR),
		Quantity:        2,
		LineSubtotal:    domain.NewMoney(2000, domain.CurrencyINR),
		PriceSnapshotAt: now,
		AddedAt:         now,
		UpdatedAt:       now,
	}
}

func validUsecaseTotals() domain.CartTotals {
	return domain.CartTotals{
		Subtotal:        domain.NewMoney(2000, domain.CurrencyINR),
		Discount:        domain.NewMoney(0, domain.CurrencyINR),
		Total:           domain.NewMoney(2000, domain.CurrencyINR),
		Currency:        domain.CurrencyINR,
		ItemCount:       2,
		UniqueItemCount: 1,
	}
}

func emptyUsecaseTotals() domain.CartTotals {
	return domain.CartTotals{
		Subtotal:        domain.NewMoney(0, domain.CurrencyINR),
		Discount:        domain.NewMoney(0, domain.CurrencyINR),
		Total:           domain.NewMoney(0, domain.CurrencyINR),
		Currency:        domain.CurrencyINR,
		ItemCount:       0,
		UniqueItemCount: 0,
	}
}

func cloneCart(cart *domain.Cart) *domain.Cart {
	if cart == nil {
		return nil
	}
	clone := *cart
	clone.Items = append([]domain.CartItem(nil), cart.Items...)
	return &clone
}
