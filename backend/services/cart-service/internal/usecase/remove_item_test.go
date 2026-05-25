package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
)

func TestRemoveItemDeletesExistingItem(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	existing := validUsecaseCart(userID, now)
	second := validUsecaseItem(now)
	second.ItemID = "item_2"
	second.ProductID = "prod_456"
	second.VariantID = "var_2"
	existing.Items = append(existing.Items, second)
	existing.Totals = domain.CartTotals{
		Subtotal:        domain.NewMoney(4000, domain.CurrencyINR),
		Discount:        domain.NewMoney(0, domain.CurrencyINR),
		Total:           domain.NewMoney(4000, domain.CurrencyINR),
		Currency:        domain.CurrencyINR,
		ItemCount:       4,
		UniqueItemCount: 2,
	}
	repo := &fakeCartRepository{cart: existing}
	cache := &fakeCartCache{}
	uc := newTestRemoveItemUsecase(t, repo, cache, now)

	cart, err := uc.RemoveItem(context.Background(), RemoveItemCommand{
		UserID: userID,
		ItemID: "item_1",
	})
	if err != nil {
		t.Fatalf("RemoveItem() error = %v", err)
	}
	if len(cart.Items) != 1 || cart.Items[0].ItemID != "item_2" {
		t.Fatalf("items = %#v, want only item_2", cart.Items)
	}
	if cart.Totals.Subtotal.Amount != 2000 || cart.Totals.ItemCount != 2 {
		t.Fatalf("unexpected totals: %#v", cart.Totals)
	}
	if repo.savedExpectedVersion != 4 || cart.Version != 5 {
		t.Fatalf("version = %d expectedSave = %d, want 5/4", cart.Version, repo.savedExpectedVersion)
	}
	if !cache.activeSet || !cache.summarySet {
		t.Fatal("cache was not refreshed")
	}
}

func TestRemoveItemReturnsEmptyCartAfterLastItem(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	repo := &fakeCartRepository{cart: validUsecaseCart(userID, now)}
	uc := newTestRemoveItemUsecase(t, repo, &fakeCartCache{}, now)

	cart, err := uc.RemoveItem(context.Background(), RemoveItemCommand{
		UserID: userID,
		ItemID: "item_1",
	})
	if err != nil {
		t.Fatalf("RemoveItem() error = %v", err)
	}
	if len(cart.Items) != 0 {
		t.Fatalf("item count = %d, want 0", len(cart.Items))
	}
	if cart.Totals.Total.Amount != 0 || cart.Totals.ItemCount != 0 || cart.Totals.UniqueItemCount != 0 {
		t.Fatalf("unexpected empty totals: %#v", cart.Totals)
	}
}

func TestRemoveItemKeepsRemainingItemCurrency(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	item1 := validUsecaseItem(now)
	item1.UnitPrice = domain.NewMoney(1000, "USD")
	item1.LineSubtotal = domain.NewMoney(2000, "USD")
	item2 := validUsecaseItem(now)
	item2.ItemID = "item_2"
	item2.ProductID = "prod_456"
	item2.VariantID = "var_2"
	item2.UnitPrice = domain.NewMoney(1000, "USD")
	item2.LineSubtotal = domain.NewMoney(2000, "USD")
	cart := validUsecaseCart(userID, now)
	cart.Items = []domain.CartItem{item1, item2}
	cart.Totals = domain.CartTotals{
		Subtotal:        domain.NewMoney(4000, "USD"),
		Discount:        domain.NewMoney(0, "USD"),
		Total:           domain.NewMoney(4000, "USD"),
		Currency:        "USD",
		ItemCount:       4,
		UniqueItemCount: 2,
	}
	uc := newTestRemoveItemUsecase(t, &fakeCartRepository{cart: cart}, &fakeCartCache{}, now)

	updated, err := uc.RemoveItem(context.Background(), RemoveItemCommand{
		UserID: userID,
		ItemID: "item_1",
	})
	if err != nil {
		t.Fatalf("RemoveItem() error = %v", err)
	}
	if updated.Totals.Currency != "USD" || updated.Totals.Total.Currency != "USD" {
		t.Fatalf("totals currency = %#v, want USD", updated.Totals)
	}
}

func TestRemoveItemCleansZeroQuantityItems(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	cart := validUsecaseCart(userID, now)
	invalid := validUsecaseItem(now)
	invalid.ItemID = "item_zero"
	invalid.ProductID = "prod_zero"
	invalid.VariantID = "var_zero"
	invalid.Quantity = 0
	invalid.LineSubtotal = domain.NewMoney(0, domain.CurrencyINR)
	cart.Items = append(cart.Items, invalid)
	cart.Totals = domain.CartTotals{
		Subtotal:        domain.NewMoney(2000, domain.CurrencyINR),
		Discount:        domain.NewMoney(0, domain.CurrencyINR),
		Total:           domain.NewMoney(2000, domain.CurrencyINR),
		Currency:        domain.CurrencyINR,
		ItemCount:       2,
		UniqueItemCount: 2,
	}
	repo := &fakeCartRepository{cart: cart}
	uc := newTestRemoveItemUsecase(t, repo, &fakeCartCache{}, now)

	updated, err := uc.RemoveItem(context.Background(), RemoveItemCommand{
		UserID: userID,
		ItemID: "missing_item",
	})
	if err != nil {
		t.Fatalf("RemoveItem() error = %v", err)
	}
	if len(updated.Items) != 1 || updated.Items[0].ItemID != "item_1" {
		t.Fatalf("items = %#v, want zero quantity item cleaned", updated.Items)
	}
	if repo.saveCalls != 1 {
		t.Fatalf("saveCalls = %d, want 1", repo.saveCalls)
	}
}

func TestRemoveItemIsIdempotentWhenItemMissing(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	repo := &fakeCartRepository{cart: validUsecaseCart(userID, now)}
	cache := &fakeCartCache{}
	uc := newTestRemoveItemUsecase(t, repo, cache, now)

	cart, err := uc.RemoveItem(context.Background(), RemoveItemCommand{
		UserID: userID,
		ItemID: "missing_item",
	})
	if err != nil {
		t.Fatalf("RemoveItem() error = %v", err)
	}
	if len(cart.Items) != 1 {
		t.Fatalf("item count = %d, want 1", len(cart.Items))
	}
	if repo.saveCalls != 0 {
		t.Fatalf("saveCalls = %d, want 0", repo.saveCalls)
	}
	if !cache.activeSet || !cache.summarySet {
		t.Fatal("cache was not refreshed for idempotent delete")
	}
}

func TestRemoveItemClearsCouponPreviewOnMutation(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	cart := validUsecaseCart(userID, now)
	code := "SAVE10"
	cart.CouponCode = &code
	cart.CouponPreview = &domain.CouponView{
		CouponID: "coupon_1",
		Code:     code,
		Valid:    true,
		Discount: domain.NewMoney(100, domain.CurrencyINR),
	}
	cart.Totals = domain.CartTotals{
		Subtotal:        domain.NewMoney(2000, domain.CurrencyINR),
		Discount:        domain.NewMoney(100, domain.CurrencyINR),
		Total:           domain.NewMoney(1900, domain.CurrencyINR),
		Currency:        domain.CurrencyINR,
		ItemCount:       2,
		UniqueItemCount: 1,
	}
	uc := newTestRemoveItemUsecase(t, &fakeCartRepository{cart: cart}, &fakeCartCache{}, now)

	updated, err := uc.RemoveItem(context.Background(), RemoveItemCommand{
		UserID: userID,
		ItemID: "item_1",
	})
	if err != nil {
		t.Fatalf("RemoveItem() error = %v", err)
	}
	if updated.CouponCode != nil || updated.CouponPreview != nil {
		t.Fatalf("coupon fields were not cleared: code=%v preview=%v", updated.CouponCode, updated.CouponPreview)
	}
	if updated.Totals.Discount.Amount != 0 || updated.Totals.Total.Amount != 0 {
		t.Fatalf("unexpected totals after coupon clear: %#v", updated.Totals)
	}
}

func TestRemoveItemRetriesVersionConflict(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	repo := &fakeCartRepository{
		cart:         validUsecaseCart(userID, now),
		conflictOnce: true,
	}
	uc := newTestRemoveItemUsecase(t, repo, &fakeCartCache{}, now)

	cart, err := uc.RemoveItem(context.Background(), RemoveItemCommand{
		UserID: userID,
		ItemID: "item_1",
	})
	if err != nil {
		t.Fatalf("RemoveItem() error = %v", err)
	}
	if repo.saveCalls != 2 {
		t.Fatalf("saveCalls = %d, want 2", repo.saveCalls)
	}
	if cart.Version != 5 {
		t.Fatalf("version = %d, want 5", cart.Version)
	}
}

func TestRemoveItemReturnsCartWhenCacheFails(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	uc := newTestRemoveItemUsecase(t, &fakeCartRepository{cart: validUsecaseCart(userID, now)}, &fakeCartCache{err: domain.ErrCacheUnavailable}, now)

	cart, err := uc.RemoveItem(context.Background(), RemoveItemCommand{
		UserID: userID,
		ItemID: "item_1",
	})
	if err != nil {
		t.Fatalf("RemoveItem() error = %v", err)
	}
	if cart == nil || len(cart.Items) != 0 {
		t.Fatalf("cart = %#v, want updated cart", cart)
	}
}

func TestRemoveItemRejectsExpiredActiveCart(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	cart := validUsecaseCart(userID, now)
	cart.ExpiresAt = now.Add(-time.Second)
	repo := &fakeCartRepository{cart: cart}
	cache := &fakeCartCache{}
	uc := newTestRemoveItemUsecase(t, repo, cache, now)

	_, err := uc.RemoveItem(context.Background(), RemoveItemCommand{
		UserID: userID,
		ItemID: "item_1",
	})
	if !errors.Is(err, domain.ErrCartExpired) {
		t.Fatalf("RemoveItem() error = %v, want ErrCartExpired", err)
	}
	if repo.expireCalls != 1 || repo.cart.Status != domain.CartStatusExpired {
		t.Fatalf("repo cart = %#v expireCalls=%d, want expired", repo.cart, repo.expireCalls)
	}
	if !cache.activeDeleted || !cache.summaryDeleted {
		t.Fatalf("cache flags = %#v, want expired cache deleted", cache)
	}
}

func TestRemoveItemRequiresItemID(t *testing.T) {
	err := ValidateRemoveItemCommand(RemoveItemCommand{UserID: "user_123"})
	if !errors.Is(err, domain.ErrItemIDRequired) {
		t.Fatalf("ValidateRemoveItemCommand() error = %v, want ErrItemIDRequired", err)
	}
}

func TestRemoveItemReturnsCartNotFound(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	uc := newTestRemoveItemUsecase(t, &fakeCartRepository{}, &fakeCartCache{}, now)

	_, err := uc.RemoveItem(context.Background(), RemoveItemCommand{
		UserID: "user_123",
		ItemID: "item_1",
	})
	if !errors.Is(err, domain.ErrCartNotFound) {
		t.Fatalf("RemoveItem() error = %v, want ErrCartNotFound", err)
	}
}

func newTestRemoveItemUsecase(t *testing.T, repo *fakeCartRepository, cache *fakeCartCache, now time.Time) *RemoveItemUsecase {
	t.Helper()
	uc, err := NewRemoveItemUsecase(RemoveItemDependencies{
		Repository:      repo,
		Cache:           cache,
		Clock:           fakeClock{now: now},
		CartTTL:         90 * 24 * time.Hour,
		DefaultCurrency: domain.CurrencyINR,
		MaxSaveAttempts: 3,
	})
	if err != nil {
		t.Fatalf("NewRemoveItemUsecase() error = %v", err)
	}
	return uc
}

func validUsecaseCart(userID string, now time.Time) *domain.Cart {
	return &domain.Cart{
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
}
