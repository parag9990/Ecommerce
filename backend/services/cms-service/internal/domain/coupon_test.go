package domain

import (
	"errors"
	"testing"
	"time"
)

func TestCalculateCouponDiscountCapsFixedAndPercentage(t *testing.T) {
	maxCap := int64(5000)
	tests := []struct {
		name     string
		coupon   Coupon
		subtotal int64
		want     int64
	}{
		{
			name: "fixed discount capped at eligible subtotal",
			coupon: Coupon{
				DiscountType:  DiscountTypeFixed,
				DiscountValue: 25000,
				Currency:      "INR",
				Status:        CouponStatusActive,
			},
			subtotal: 10000,
			want:     10000,
		},
		{
			name: "percentage discount capped by max discount amount",
			coupon: Coupon{
				DiscountType:      DiscountTypePercentage,
				DiscountValue:     10,
				MaxDiscountAmount: &maxCap,
				Currency:          "INR",
				Status:            CouponStatusActive,
			},
			subtotal: 100000,
			want:     5000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateCouponDiscount(tt.coupon, tt.subtotal); got != tt.want {
				t.Fatalf("CalculateCouponDiscount() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCalculateEligibleSubtotalAppliesSellerAndCategoryScopes(t *testing.T) {
	sellerID := "seller_1"
	coupon := Coupon{
		SellerID: &sellerID,
		Currency: "INR",
		Status:   CouponStatusActive,
	}
	rules := []CouponRule{
		{Type: CouponRuleTypeCategoryScope, CategoryIDs: []string{"cat_shoes"}},
	}
	items := []CouponCartItem{
		{ProductID: "prod_1", SellerID: "seller_1", CategoryIDs: []string{"cat_shoes"}, LineSubtotalAmount: 150000},
		{ProductID: "prod_2", SellerID: "seller_1", CategoryIDs: []string{"cat_bags"}, LineSubtotalAmount: 50000},
		{ProductID: "prod_3", SellerID: "seller_2", CategoryIDs: []string{"cat_shoes"}, LineSubtotalAmount: 75000},
	}

	got := CalculateEligibleSubtotal(275000, items, coupon, rules)
	if got != 150000 {
		t.Fatalf("CalculateEligibleSubtotal() = %d, want 150000", got)
	}
}

func TestValidateCouponDefinitionRejectsInvalidPercentage(t *testing.T) {
	err := ValidateCouponDefinition(Coupon{
		CouponID:      "coupon_1",
		Code:          "SAVE150",
		DiscountType:  DiscountTypePercentage,
		DiscountValue: 150,
		Currency:      "INR",
		Status:        CouponStatusActive,
	}, nil)
	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestValidateCouponEligibilityChecksWindowCurrencyAndMinimumCart(t *testing.T) {
	start := time.Unix(2000, 0).UTC()
	end := time.Unix(3000, 0).UTC()
	coupon := Coupon{
		Code:          "SAVE10",
		DiscountType:  DiscountTypePercentage,
		DiscountValue: 10,
		MinCartAmount: 10000,
		Currency:      "INR",
		Status:        CouponStatusActive,
		StartsAt:      &start,
		EndsAt:        &end,
	}
	if reason := ValidateCouponEligibility(coupon, "USD", 20000, time.Unix(2500, 0).UTC()); reason != CouponInvalidReasonCurrencyMismatch {
		t.Fatalf("expected currency mismatch, got %q", reason)
	}
	if reason := ValidateCouponEligibility(coupon, "INR", 5000, time.Unix(2500, 0).UTC()); reason != CouponInvalidReasonMinCartNotMet {
		t.Fatalf("expected min cart reason, got %q", reason)
	}
	if reason := ValidateCouponEligibility(coupon, "INR", 20000, time.Unix(3500, 0).UTC()); reason != CouponInvalidReasonExpired {
		t.Fatalf("expected expired reason, got %q", reason)
	}
}
