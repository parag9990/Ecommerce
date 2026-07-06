package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
)

func TestUpdateItemSetsQuantityAndRefreshesSnapshot(t *testing.T) {
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
	repo := &fakeCartRepository{cart: cart}
	cache := &fakeCartCache{}
	uc := newTestUpdateItemUsecase(t, repo, cache, now)

	updated, err := uc.UpdateItem(context.Background(), UpdateItemCommand{
		UserID:   userID,
		ItemID:   "item_1",
		Quantity: 3,
	})
	if err != nil {
		t.Fatalf("UpdateItem() error = %v", err)
	}
	if updated.Items[0].Quantity != 3 || updated.Items[0].LineSubtotal.Amount != 3000 {
		t.Fatalf("unexpected updated item: %#v", updated.Items[0])
	}
	if updated.Totals.Subtotal.Amount != 3000 || updated.Totals.Total.Amount != 3000 {
		t.Fatalf("unexpected totals: %#v", updated.Totals)
	}
	if updated.CouponCode != nil || updated.CouponPreview != nil {
		t.Fatalf("coupon fields were not cleared: code=%v preview=%v", updated.CouponCode, updated.CouponPreview)
	}
	if repo.savedExpectedVersion != 4 || updated.Version != 5 {
		t.Fatalf("version = %d expectedSave = %d, want 5/4", updated.Version, repo.savedExpectedVersion)
	}
	if !cache.activeSet || !cache.summarySet {
		t.Fatal("cache was not refreshed")
	}
}

func TestUpdateItemMatchesExistingLineBySKU(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	cart := validUsecaseCart(userID, now)
	cart.Items[0].VariantID = "RUN-BLK-9"
	repo := &fakeCartRepository{cart: cart}
	uc := newTestUpdateItemUsecase(t, repo, &fakeCartCache{}, now)

	updated, err := uc.UpdateItem(context.Background(), UpdateItemCommand{
		UserID:   userID,
		ItemID:   "item_1",
		Quantity: 1,
	})
	if err != nil {
		t.Fatalf("UpdateItem() error = %v", err)
	}
	if got := updated.Items[0]; got.VariantID != "var_1" || got.SKUSnapshot == nil || *got.SKUSnapshot != "RUN-BLK-9" {
		t.Fatalf("unexpected SKU-normalized item: %#v", got)
	}
}

func TestUpdateItemRejectsMissingItem(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	uc := newTestUpdateItemUsecase(t, &fakeCartRepository{cart: validUsecaseCart("user_123", now)}, &fakeCartCache{}, now)

	_, err := uc.UpdateItem(context.Background(), UpdateItemCommand{
		UserID:   "user_123",
		ItemID:   "missing_item",
		Quantity: 1,
	})
	if !errors.Is(err, domain.ErrCartItemNotFound) {
		t.Fatalf("UpdateItem() error = %v, want ErrCartItemNotFound", err)
	}
}

func TestUpdateItemRetriesVersionConflict(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	repo := &fakeCartRepository{
		cart:         validUsecaseCart(userID, now),
		conflictOnce: true,
	}
	uc := newTestUpdateItemUsecase(t, repo, &fakeCartCache{}, now)

	updated, err := uc.UpdateItem(context.Background(), UpdateItemCommand{
		UserID:   userID,
		ItemID:   "item_1",
		Quantity: 1,
	})
	if err != nil {
		t.Fatalf("UpdateItem() error = %v", err)
	}
	if repo.saveCalls != 2 {
		t.Fatalf("saveCalls = %d, want 2", repo.saveCalls)
	}
	if updated.Version != 5 {
		t.Fatalf("version = %d, want 5", updated.Version)
	}
}

func TestUpdateItemRequiresQuantityWithinLimits(t *testing.T) {
	err := ValidateUpdateItemCommand(UpdateItemCommand{UserID: "user_123", ItemID: "item_1"})
	if !errors.Is(err, domain.ErrQuantityTooSmall) {
		t.Fatalf("ValidateUpdateItemCommand() error = %v, want ErrQuantityTooSmall", err)
	}
}

func newTestUpdateItemUsecase(t *testing.T, repo *fakeCartRepository, cache *fakeCartCache, now time.Time) *UpdateItemUsecase {
	t.Helper()
	uc, err := NewUpdateItemUsecase(UpdateItemDependencies{
		Repository:      repo,
		Cache:           cache,
		ProductClient:   &fakeProductClient{product: validProductForCart()},
		Clock:           fakeClock{now: now},
		CartTTL:         90 * 24 * time.Hour,
		DefaultCurrency: domain.CurrencyINR,
		MaxSaveAttempts: 3,
	})
	if err != nil {
		t.Fatalf("NewUpdateItemUsecase() error = %v", err)
	}
	return uc
}
