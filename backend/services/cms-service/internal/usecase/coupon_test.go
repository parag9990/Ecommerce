package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
)

func TestValidateCouponCalculatesScopedPercentageDiscount(t *testing.T) {
	repo := newMemoryCouponRepository()
	sellerID := "seller_1"
	repo.coupons["coupon_1"] = domain.Coupon{
		CouponID:      "coupon_1",
		SellerID:      &sellerID,
		Code:          "SHOE10",
		DiscountType:  domain.DiscountTypePercentage,
		DiscountValue: 10,
		MinCartAmount: 100000,
		Currency:      "INR",
		Status:        domain.CouponStatusActive,
	}
	repo.rules["coupon_1"] = []domain.CouponRule{
		{RuleID: "rule_1", CouponID: "coupon_1", Type: domain.CouponRuleTypeCategoryScope, CategoryIDs: []string{"cat_shoes"}},
	}

	uc := newTestCouponUsecase(t, repo)
	result, err := uc.ValidateCoupon(context.Background(), ValidateCouponInput{
		CouponCode:     "shoe10",
		UserID:         "user_1",
		Currency:       "INR",
		SubtotalAmount: 200000,
		Items: []domain.CouponCartItem{
			{ProductID: "prod_shoe", SellerID: "seller_1", CategoryIDs: []string{"cat_shoes"}, LineSubtotalAmount: 150000},
			{ProductID: "prod_bag", SellerID: "seller_1", CategoryIDs: []string{"cat_bags"}, LineSubtotalAmount: 50000},
		},
	})
	if err != nil {
		t.Fatalf("ValidateCoupon: %v", err)
	}
	if !result.Valid || result.Discount.Amount != 15000 || result.EligibleSubtotal != 150000 {
		t.Fatalf("unexpected validation result: %+v", result)
	}
}

func TestValidateCouponReturnsPerUserLimitReason(t *testing.T) {
	repo := newMemoryCouponRepository()
	limit := int64(1)
	repo.coupons["coupon_1"] = domain.Coupon{
		CouponID:      "coupon_1",
		Code:          "SAVE10",
		DiscountType:  domain.DiscountTypePercentage,
		DiscountValue: 10,
		Currency:      "INR",
		Status:        domain.CouponStatusActive,
		PerUserLimit:  &limit,
	}
	repo.redemptions["coupon_1:order_1"] = domain.CouponRedemption{
		RedemptionID: "redemption_1",
		CouponID:     "coupon_1",
		OrderID:      "order_1",
		UserID:       "user_1",
		Discount:     domain.Money{Amount: 1000, Currency: "INR"},
	}

	uc := newTestCouponUsecase(t, repo)
	result, err := uc.ValidateCoupon(context.Background(), ValidateCouponInput{
		CouponCode:     "SAVE10",
		UserID:         "user_1",
		Currency:       "INR",
		SubtotalAmount: 100000,
	})
	if err != nil {
		t.Fatalf("ValidateCoupon: %v", err)
	}
	if result.Valid || result.Reason != domain.CouponInvalidReasonPerUserLimitReached {
		t.Fatalf("expected per user limit reason, got %+v", result)
	}
}

func TestRecordRedemptionIsIdempotentForSameOrder(t *testing.T) {
	repo := newMemoryCouponRepository()
	sellerID := "seller_1"
	repo.coupons["coupon_1"] = domain.Coupon{
		CouponID:      "coupon_1",
		SellerID:      &sellerID,
		Code:          "SAVE10",
		DiscountType:  domain.DiscountTypeFixed,
		DiscountValue: 1000,
		Currency:      "INR",
		Status:        domain.CouponStatusActive,
	}

	uc := newTestCouponUsecase(t, repo)
	first, err := uc.RecordRedemption(context.Background(), RecordCouponRedemptionInput{
		CouponID:       "coupon_1",
		OrderID:        "order_1",
		UserID:         "user_1",
		DiscountAmount: 1000,
		Currency:       "INR",
	})
	if err != nil {
		t.Fatalf("RecordRedemption first call: %v", err)
	}
	second, err := uc.RecordRedemption(context.Background(), RecordCouponRedemptionInput{
		CouponID:       "coupon_1",
		OrderID:        "order_1",
		UserID:         "user_1",
		DiscountAmount: 1000,
		Currency:       "INR",
	})
	if err != nil {
		t.Fatalf("RecordRedemption retry: %v", err)
	}
	if second.RedemptionID != first.RedemptionID {
		t.Fatalf("expected idempotent existing redemption, got first=%+v second=%+v", first, second)
	}
}

func TestCreateCouponRejectsCrossSellerScopeRule(t *testing.T) {
	repo := newMemoryCouponRepository()
	uc := newTestCouponUsecase(t, repo)

	_, err := uc.CreateCoupon(context.Background(), CreateCouponInput{
		Actor:         activeActor(domain.RoleSellerCatalogEditor),
		Code:          "SAVE10",
		DiscountType:  domain.DiscountTypePercentage,
		DiscountValue: 10,
		Currency:      "INR",
		Status:        domain.CouponStatusActive,
		Rules: []domain.CouponRule{
			{Type: domain.CouponRuleTypeSellerScope, SellerIDs: []string{"seller_2"}},
		},
	})
	var validationErr domain.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func newTestCouponUsecase(t *testing.T, repo CouponRepository) *CouponUsecase {
	t.Helper()
	uc, err := NewCouponUsecase(
		newTestAuthorizer(t, nil, nil),
		repo,
		auditRecorderFunc(func(context.Context, domain.AuditEvent) error { return nil }),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	if err != nil {
		t.Fatalf("NewCouponUsecase: %v", err)
	}
	uc.now = func() time.Time { return time.Unix(1700000000, 0).UTC() }
	uc.newID = func(prefix string) string { return prefix + "_test" }
	return uc
}

type memoryCouponRepository struct {
	coupons     map[string]domain.Coupon
	rules       map[string][]domain.CouponRule
	redemptions map[string]domain.CouponRedemption
}

func newMemoryCouponRepository() *memoryCouponRepository {
	return &memoryCouponRepository{
		coupons:     make(map[string]domain.Coupon),
		rules:       make(map[string][]domain.CouponRule),
		redemptions: make(map[string]domain.CouponRedemption),
	}
}

func (r *memoryCouponRepository) CreateCoupon(_ context.Context, coupon domain.Coupon, rules []domain.CouponRule) (domain.Coupon, error) {
	for _, existing := range r.coupons {
		if domain.NormalizeCouponCode(existing.Code) == domain.NormalizeCouponCode(coupon.Code) {
			return domain.Coupon{}, domain.ErrCouponCodeAlreadyExists
		}
	}
	r.coupons[coupon.CouponID] = coupon.Normalized()
	r.rules[coupon.CouponID] = domain.NormalizeCouponRules(rules)
	return r.GetCouponByID(context.Background(), coupon.CouponID)
}

func (r *memoryCouponRepository) UpdateCoupon(_ context.Context, coupon domain.Coupon, rules []domain.CouponRule) (domain.Coupon, error) {
	if _, ok := r.coupons[coupon.CouponID]; !ok {
		return domain.Coupon{}, domain.ErrCouponNotFound
	}
	r.coupons[coupon.CouponID] = coupon.Normalized()
	r.rules[coupon.CouponID] = domain.NormalizeCouponRules(rules)
	return r.GetCouponByID(context.Background(), coupon.CouponID)
}

func (r *memoryCouponRepository) UpdateCouponStatus(_ context.Context, couponID string, status domain.CouponStatus, updatedAt time.Time) (domain.Coupon, error) {
	coupon, ok := r.coupons[couponID]
	if !ok {
		return domain.Coupon{}, domain.ErrCouponNotFound
	}
	coupon.Status = status
	coupon.UpdatedAt = updatedAt
	r.coupons[couponID] = coupon
	return coupon, nil
}

func (r *memoryCouponRepository) GetCouponByID(_ context.Context, couponID string) (domain.Coupon, error) {
	coupon, ok := r.coupons[couponID]
	if !ok {
		return domain.Coupon{}, domain.ErrCouponNotFound
	}
	coupon.Rules = r.rules[couponID]
	return coupon.Normalized(), nil
}

func (r *memoryCouponRepository) FindCouponByCode(_ context.Context, code string) (domain.Coupon, error) {
	code = domain.NormalizeCouponCode(code)
	for _, coupon := range r.coupons {
		if domain.NormalizeCouponCode(coupon.Code) == code {
			coupon.Rules = r.rules[coupon.CouponID]
			return coupon.Normalized(), nil
		}
	}
	return domain.Coupon{}, domain.ErrCouponNotFound
}

func (r *memoryCouponRepository) ListCoupons(_ context.Context, filter CouponFilter) ([]domain.Coupon, error) {
	out := make([]domain.Coupon, 0, len(r.coupons))
	for _, coupon := range r.coupons {
		coupon = coupon.Normalized()
		if filter.SellerID != "" && (coupon.SellerID == nil || *coupon.SellerID != filter.SellerID) {
			continue
		}
		if filter.Status.Normalized() != "" && coupon.Status != filter.Status.Normalized() {
			continue
		}
		coupon.Rules = r.rules[coupon.CouponID]
		out = append(out, coupon)
	}
	return out, nil
}

func (r *memoryCouponRepository) ListCouponRules(_ context.Context, couponID string) ([]domain.CouponRule, error) {
	return append([]domain.CouponRule(nil), r.rules[couponID]...), nil
}

func (r *memoryCouponRepository) CountRedemptions(_ context.Context, couponID string) (int64, error) {
	var count int64
	for _, redemption := range r.redemptions {
		if redemption.CouponID == couponID {
			count++
		}
	}
	return count, nil
}

func (r *memoryCouponRepository) CountUserRedemptions(_ context.Context, couponID string, userID string) (int64, error) {
	var count int64
	for _, redemption := range r.redemptions {
		if redemption.CouponID == couponID && redemption.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (r *memoryCouponRepository) InsertRedemption(_ context.Context, redemption domain.CouponRedemption, coupon domain.Coupon) error {
	key := redemption.CouponID + ":" + redemption.OrderID
	if _, ok := r.redemptions[key]; ok {
		return domain.ErrCouponRedemptionDuplicate
	}
	if coupon.UsageLimit != nil {
		count, _ := r.CountRedemptions(context.Background(), redemption.CouponID)
		if count+1 > *coupon.UsageLimit {
			return domain.ErrCouponUsageLimitReached
		}
	}
	if coupon.PerUserLimit != nil {
		count, _ := r.CountUserRedemptions(context.Background(), redemption.CouponID, redemption.UserID)
		if count+1 > *coupon.PerUserLimit {
			return domain.ErrCouponPerUserLimitReached
		}
	}
	r.redemptions[key] = redemption
	return nil
}

func (r *memoryCouponRepository) GetRedemption(_ context.Context, couponID string, orderID string) (domain.CouponRedemption, error) {
	redemption, ok := r.redemptions[couponID+":"+orderID]
	if !ok {
		return domain.CouponRedemption{}, domain.ErrCouponRedemptionNotFound
	}
	return redemption, nil
}
