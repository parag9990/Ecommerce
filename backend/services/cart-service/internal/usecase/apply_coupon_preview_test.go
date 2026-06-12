package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
)

func TestApplyCouponPreviewAppliesValidDiscount(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	repo := &fakeCartRepository{cart: validUsecaseCart(userID, now)}
	cache := &fakeCartCache{}
	validator := &fakeCouponValidator{
		result: &CouponValidationResult{
			Valid:    true,
			CouponID: "coupon_save10",
			Code:     "SAVE10",
			Discount: domain.NewMoney(200, domain.CurrencyINR),
		},
	}
	uc := newTestApplyCouponPreviewUsecase(t, repo, cache, validator, now)

	preview, err := uc.ApplyCouponPreview(context.Background(), ApplyCouponPreviewCommand{
		UserID:     userID,
		CouponCode: " save10 ",
	})
	if err != nil {
		t.Fatalf("ApplyCouponPreview() error = %v", err)
	}
	if !preview.Valid || preview.CouponID != "coupon_save10" || preview.Discount.Amount != 200 {
		t.Fatalf("preview = %#v, want valid discount", preview)
	}
	if validator.request.Code != "SAVE10" || validator.request.CartID != "cart_123" || len(validator.request.Items) != 1 {
		t.Fatalf("validator request = %#v, want normalized cart context", validator.request)
	}
	if repo.savedExpectedVersion != 4 || repo.cart.Version != 5 {
		t.Fatalf("version = %d expectedSave = %d, want 5/4", repo.cart.Version, repo.savedExpectedVersion)
	}
	if repo.cart.CouponCode == nil || *repo.cart.CouponCode != "SAVE10" || repo.cart.Totals.Total.Amount != 1800 {
		t.Fatalf("stored cart = %#v, want coupon preview applied", repo.cart)
	}
	if !cache.activeSet || !cache.summarySet {
		t.Fatal("cache was not refreshed")
	}
}

func TestApplyCouponPreviewClearsDiscountWhenCouponInvalid(t *testing.T) {
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
	repo := &fakeCartRepository{cart: cart}
	validator := &fakeCouponValidator{
		result: &CouponValidationResult{
			Valid:  false,
			Reason: "MIN_CART_AMOUNT_NOT_MET",
		},
	}
	uc := newTestApplyCouponPreviewUsecase(t, repo, &fakeCartCache{}, validator, now)

	preview, err := uc.ApplyCouponPreview(context.Background(), ApplyCouponPreviewCommand{
		UserID:     userID,
		CouponCode: "BIGSAVE",
	})
	if err != nil {
		t.Fatalf("ApplyCouponPreview() error = %v", err)
	}
	if preview.Valid || preview.Reason != "MIN_CART_AMOUNT_NOT_MET" || preview.Discount.Amount != 0 {
		t.Fatalf("preview = %#v, want invalid zero discount", preview)
	}
	if repo.cart.CouponPreview == nil || repo.cart.CouponPreview.Valid || repo.cart.Totals.Total.Amount != 2000 {
		t.Fatalf("stored cart = %#v, want previous discount cleared", repo.cart)
	}
}

func TestApplyCouponPreviewRejectsEmptyCouponCode(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	uc := newTestApplyCouponPreviewUsecase(t, &fakeCartRepository{}, &fakeCartCache{}, &fakeCouponValidator{}, now)

	_, err := uc.ApplyCouponPreview(context.Background(), ApplyCouponPreviewCommand{
		UserID:     "user_123",
		CouponCode: " ",
	})
	if !errors.Is(err, domain.ErrCouponCodeRequired) {
		t.Fatalf("ApplyCouponPreview() error = %v, want ErrCouponCodeRequired", err)
	}
}

func TestApplyCouponPreviewRejectsEmptyCart(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	repo := &fakeCartRepository{cart: &domain.Cart{
		ID:        "cart_123",
		UserID:    &userID,
		Status:    domain.CartStatusActive,
		Items:     []domain.CartItem{},
		Totals:    emptyUsecaseTotals(),
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(90 * 24 * time.Hour),
	}}
	uc := newTestApplyCouponPreviewUsecase(t, repo, &fakeCartCache{}, &fakeCouponValidator{}, now)

	_, err := uc.ApplyCouponPreview(context.Background(), ApplyCouponPreviewCommand{
		UserID:     userID,
		CouponCode: "SAVE10",
	})
	if !errors.Is(err, domain.ErrEmptyCart) {
		t.Fatalf("ApplyCouponPreview() error = %v, want ErrEmptyCart", err)
	}
	if repo.saveCalls != 0 {
		t.Fatalf("saveCalls = %d, want 0", repo.saveCalls)
	}
}

func TestApplyCouponPreviewRejectsCartIDMismatch(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	uc := newTestApplyCouponPreviewUsecase(t, &fakeCartRepository{cart: validUsecaseCart(userID, now)}, &fakeCartCache{}, &fakeCouponValidator{}, now)

	_, err := uc.ApplyCouponPreview(context.Background(), ApplyCouponPreviewCommand{
		UserID:     userID,
		CartID:     "cart_other",
		CouponCode: "SAVE10",
	})
	if !errors.Is(err, domain.ErrCartOwnershipMismatch) {
		t.Fatalf("ApplyCouponPreview() error = %v, want ErrCartOwnershipMismatch", err)
	}
}

func TestApplyCouponPreviewRejectsCurrencyMismatch(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	repo := &fakeCartRepository{cart: validUsecaseCart(userID, now)}
	uc := newTestApplyCouponPreviewUsecase(t, repo, &fakeCartCache{}, &fakeCouponValidator{
		result: &CouponValidationResult{
			Valid:    true,
			CouponID: "coupon_usd",
			Code:     "SAVE10",
			Discount: domain.NewMoney(100, "USD"),
		},
	}, now)

	_, err := uc.ApplyCouponPreview(context.Background(), ApplyCouponPreviewCommand{
		UserID:     userID,
		CouponCode: "SAVE10",
	})
	if !errors.Is(err, domain.ErrCouponCurrencyMismatch) {
		t.Fatalf("ApplyCouponPreview() error = %v, want ErrCouponCurrencyMismatch", err)
	}
	if repo.saveCalls != 0 {
		t.Fatalf("saveCalls = %d, want 0", repo.saveCalls)
	}
}

func TestApplyCouponPreviewDoesNotMutateWhenCMSUnavailable(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	repo := &fakeCartRepository{cart: validUsecaseCart(userID, now)}
	uc := newTestApplyCouponPreviewUsecase(t, repo, &fakeCartCache{}, &fakeCouponValidator{err: domain.ErrCouponPreviewUnavailable}, now)

	_, err := uc.ApplyCouponPreview(context.Background(), ApplyCouponPreviewCommand{
		UserID:     userID,
		CouponCode: "SAVE10",
	})
	if !errors.Is(err, domain.ErrCouponPreviewUnavailable) {
		t.Fatalf("ApplyCouponPreview() error = %v, want ErrCouponPreviewUnavailable", err)
	}
	if repo.saveCalls != 0 {
		t.Fatalf("saveCalls = %d, want 0", repo.saveCalls)
	}
}

func TestApplyCouponPreviewRetriesVersionConflict(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	repo := &fakeCartRepository{
		cart:         validUsecaseCart(userID, now),
		conflictOnce: true,
	}
	uc := newTestApplyCouponPreviewUsecase(t, repo, &fakeCartCache{}, &fakeCouponValidator{
		result: &CouponValidationResult{
			Valid:    true,
			CouponID: "coupon_save10",
			Code:     "SAVE10",
			Discount: domain.NewMoney(200, domain.CurrencyINR),
		},
	}, now)

	preview, err := uc.ApplyCouponPreview(context.Background(), ApplyCouponPreviewCommand{
		UserID:     userID,
		CouponCode: "SAVE10",
	})
	if err != nil {
		t.Fatalf("ApplyCouponPreview() error = %v", err)
	}
	if repo.saveCalls != 2 {
		t.Fatalf("saveCalls = %d, want 2", repo.saveCalls)
	}
	if preview == nil || !preview.Valid {
		t.Fatalf("preview = %#v, want valid", preview)
	}
}

func TestApplyCouponPreviewReturnsSuccessWhenCacheFails(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	uc := newTestApplyCouponPreviewUsecase(t, &fakeCartRepository{cart: validUsecaseCart(userID, now)}, &fakeCartCache{err: domain.ErrCacheUnavailable}, &fakeCouponValidator{
		result: &CouponValidationResult{
			Valid:    true,
			CouponID: "coupon_save10",
			Code:     "SAVE10",
			Discount: domain.NewMoney(200, domain.CurrencyINR),
		},
	}, now)

	preview, err := uc.ApplyCouponPreview(context.Background(), ApplyCouponPreviewCommand{
		UserID:     userID,
		CouponCode: "SAVE10",
	})
	if err != nil {
		t.Fatalf("ApplyCouponPreview() error = %v", err)
	}
	if preview == nil || !preview.Valid {
		t.Fatalf("preview = %#v, want valid", preview)
	}
}

func TestApplyCouponPreviewRejectsExpiredActiveCart(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	cart := validUsecaseCart(userID, now)
	cart.ExpiresAt = now.Add(-time.Second)
	repo := &fakeCartRepository{cart: cart}
	cache := &fakeCartCache{}
	uc := newTestApplyCouponPreviewUsecase(t, repo, cache, &fakeCouponValidator{
		result: &CouponValidationResult{
			Valid:    true,
			CouponID: "coupon_save10",
			Code:     "SAVE10",
			Discount: domain.NewMoney(200, domain.CurrencyINR),
		},
	}, now)

	_, err := uc.ApplyCouponPreview(context.Background(), ApplyCouponPreviewCommand{
		UserID:     userID,
		CouponCode: "SAVE10",
	})
	if !errors.Is(err, domain.ErrCartExpired) {
		t.Fatalf("ApplyCouponPreview() error = %v, want ErrCartExpired", err)
	}
	if repo.expireCalls != 1 || repo.saveCalls != 0 || repo.cart.Status != domain.CartStatusExpired {
		t.Fatalf("repo state = %#v expireCalls=%d saveCalls=%d, want expired without coupon save", repo.cart, repo.expireCalls, repo.saveCalls)
	}
	if !cache.activeDeleted || !cache.summaryDeleted {
		t.Fatalf("cache flags = %#v, want expired cache deleted", cache)
	}
}

func newTestApplyCouponPreviewUsecase(t *testing.T, repo *fakeCartRepository, cache *fakeCartCache, validator *fakeCouponValidator, now time.Time) *ApplyCouponPreviewUsecase {
	t.Helper()
	uc, err := NewApplyCouponPreviewUsecase(ApplyCouponPreviewDependencies{
		Repository:      repo,
		Cache:           cache,
		CouponValidator: validator,
		Clock:           fakeClock{now: now},
		CartTTL:         90 * 24 * time.Hour,
		RequestTimeout:  time.Second,
		MaxSaveAttempts: 3,
	})
	if err != nil {
		t.Fatalf("NewApplyCouponPreviewUsecase() error = %v", err)
	}
	return uc
}

type fakeCouponValidator struct {
	request CouponValidationRequest
	result  *CouponValidationResult
	err     error
}

func (f *fakeCouponValidator) ValidateCoupon(ctx context.Context, req CouponValidationRequest) (*CouponValidationResult, error) {
	f.request = req
	if f.err != nil {
		return nil, f.err
	}
	if f.result == nil {
		return &CouponValidationResult{Valid: false, Reason: "COUPON_NOT_FOUND"}, nil
	}
	result := *f.result
	return &result, nil
}
