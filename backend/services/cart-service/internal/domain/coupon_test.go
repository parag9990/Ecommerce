package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNormalizeCouponCode(t *testing.T) {
	code, err := NormalizeCouponCode(" save10 ")
	if err != nil {
		t.Fatalf("NormalizeCouponCode() error = %v", err)
	}
	if code != "SAVE10" {
		t.Fatalf("code = %q, want SAVE10", code)
	}
}

func TestNormalizeCouponCodeRejectsTooLong(t *testing.T) {
	_, err := NormalizeCouponCode(strings.Repeat("A", MaxCouponCodeLength+1))
	if !errors.Is(err, ErrCouponCodeTooLong) {
		t.Fatalf("NormalizeCouponCode() error = %v, want ErrCouponCodeTooLong", err)
	}
}

func TestApplyValidCouponPreviewUpdatesTotals(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	cart := validDomainCart(userID, now)

	err := cart.ApplyValidCouponPreview(CouponPreviewDecision{
		CouponID: "coupon_save10",
		Code:     "save10",
		Valid:    true,
		Discount: NewMoney(200, CurrencyINR),
	}, now)
	if err != nil {
		t.Fatalf("ApplyValidCouponPreview() error = %v", err)
	}
	if cart.CouponCode == nil || *cart.CouponCode != "SAVE10" {
		t.Fatalf("CouponCode = %v, want SAVE10", cart.CouponCode)
	}
	if cart.CouponPreview == nil || !cart.CouponPreview.Valid || cart.CouponPreview.Discount.Amount != 200 {
		t.Fatalf("CouponPreview = %#v, want valid discount", cart.CouponPreview)
	}
	if cart.Totals.Total.Amount != 1800 || cart.Totals.Discount.Amount != 200 {
		t.Fatalf("Totals = %#v, want discount reflected", cart.Totals)
	}
}

func TestApplyInvalidCouponPreviewClearsDiscount(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	cart := validDomainCart(userID, now)
	code := "SAVE10"
	cart.CouponCode = &code
	cart.CouponPreview = &CouponView{
		CouponID: "coupon_save10",
		Code:     code,
		Valid:    true,
		Discount: NewMoney(200, CurrencyINR),
	}
	cart.Totals = CartTotals{
		Subtotal:        NewMoney(2000, CurrencyINR),
		Discount:        NewMoney(200, CurrencyINR),
		Total:           NewMoney(1800, CurrencyINR),
		Currency:        CurrencyINR,
		ItemCount:       2,
		UniqueItemCount: 1,
	}

	err := cart.ApplyInvalidCouponPreview("bigsale", "min_cart_amount_not_met", now)
	if err != nil {
		t.Fatalf("ApplyInvalidCouponPreview() error = %v", err)
	}
	if cart.CouponPreview == nil || cart.CouponPreview.Valid || cart.CouponPreview.Reason != "MIN_CART_AMOUNT_NOT_MET" {
		t.Fatalf("CouponPreview = %#v, want invalid reason", cart.CouponPreview)
	}
	if cart.Totals.Discount.Amount != 0 || cart.Totals.Total.Amount != 2000 {
		t.Fatalf("Totals = %#v, want discount cleared", cart.Totals)
	}
}

func TestValidatePreviewDiscountRejectsTooLargeDiscount(t *testing.T) {
	err := ValidatePreviewDiscount(NewMoney(100, CurrencyINR), NewMoney(101, CurrencyINR))
	if !errors.Is(err, ErrCouponDiscountTooLarge) {
		t.Fatalf("ValidatePreviewDiscount() error = %v, want ErrCouponDiscountTooLarge", err)
	}
}

func validDomainCart(userID string, now time.Time) *Cart {
	return &Cart{
		ID:     "cart_123",
		UserID: &userID,
		Status: CartStatusActive,
		Items: []CartItem{
			{
				ItemID:          "item_1",
				ProductID:       "prod_123",
				VariantID:       "var_1",
				SellerID:        "seller_456",
				TitleSnapshot:   "Running Shoes",
				UnitPrice:       NewMoney(1000, CurrencyINR),
				Quantity:        2,
				LineSubtotal:    NewMoney(2000, CurrencyINR),
				PriceSnapshotAt: now,
				AddedAt:         now,
				UpdatedAt:       now,
			},
		},
		Totals: CartTotals{
			Subtotal:        NewMoney(2000, CurrencyINR),
			Discount:        NewMoney(0, CurrencyINR),
			Total:           NewMoney(2000, CurrencyINR),
			Currency:        CurrencyINR,
			ItemCount:       2,
			UniqueItemCount: 1,
		},
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(90 * 24 * time.Hour),
	}
}
