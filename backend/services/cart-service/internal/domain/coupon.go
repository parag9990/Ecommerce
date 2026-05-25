package domain

import (
	"fmt"
	"strings"
	"time"
)

const (
	MaxCouponCodeLength        = 64
	DefaultInvalidCouponReason = "COUPON_INVALID"
)

type CouponPreviewDecision struct {
	CouponID string
	Code     string
	Valid    bool
	Discount Money
	Reason   string
}

func NormalizeCouponCode(code string) (string, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return "", ErrCouponCodeRequired
	}
	if len(code) > MaxCouponCodeLength {
		return "", fmt.Errorf("%w: max %d characters allowed", ErrCouponCodeTooLong, MaxCouponCodeLength)
	}
	return code, nil
}

func EnsureCouponPreviewAllowed(cart *Cart) error {
	if cart == nil {
		return fmt.Errorf("%w: cart is nil", ErrInvalidCart)
	}
	if cart.Status != CartStatusActive {
		return ErrCartNotActive
	}
	if len(cart.Items) == 0 || cart.Totals.Subtotal.Amount == 0 {
		return ErrEmptyCart
	}
	return nil
}

func ValidatePreviewDiscount(subtotal Money, discount Money) error {
	if err := subtotal.Validate(); err != nil {
		return fmt.Errorf("subtotal: %w", err)
	}
	discount = NewMoney(discount.Amount, discount.Currency)
	if err := discount.Validate(); err != nil {
		return fmt.Errorf("discount: %w", err)
	}
	if !subtotal.SameCurrency(discount) {
		return ErrCouponCurrencyMismatch
	}
	if discount.Amount > subtotal.Amount {
		return ErrCouponDiscountTooLarge
	}
	return nil
}

func (c *Cart) ApplyValidCouponPreview(decision CouponPreviewDecision, now time.Time) error {
	if err := EnsureCouponPreviewAllowed(c); err != nil {
		return err
	}
	now = now.UTC()
	if now.IsZero() {
		return fmt.Errorf("%w: coupon preview timestamp is required", ErrInvalidCart)
	}
	code, err := NormalizeCouponCode(decision.Code)
	if err != nil {
		return err
	}
	couponID := strings.TrimSpace(decision.CouponID)
	if couponID == "" {
		return fmt.Errorf("%w: coupon_id is required for a valid preview", ErrInvalidCouponPreview)
	}
	discount := NewMoney(decision.Discount.Amount, decision.Discount.Currency)
	if err := ValidatePreviewDiscount(c.Totals.Subtotal, discount); err != nil {
		return err
	}

	totals, err := CalculateTotals(c.Items, c.Totals.Currency, discount)
	if err != nil {
		return err
	}

	c.CouponCode = &code
	c.CouponPreview = &CouponView{
		CouponID: couponID,
		Code:     code,
		Valid:    true,
		Discount: discount,
		Reason:   "",
	}
	c.Totals = totals
	c.UpdatedAt = now
	return nil
}

func (c *Cart) ApplyInvalidCouponPreview(code string, reason string, now time.Time) error {
	if err := EnsureCouponPreviewAllowed(c); err != nil {
		return err
	}
	now = now.UTC()
	if now.IsZero() {
		return fmt.Errorf("%w: coupon preview timestamp is required", ErrInvalidCart)
	}
	normalizedCode, err := NormalizeCouponCode(code)
	if err != nil {
		return err
	}
	reason = strings.ToUpper(strings.TrimSpace(reason))
	if reason == "" {
		reason = DefaultInvalidCouponReason
	}
	zero := NewMoney(0, c.Totals.Currency)
	totals, err := CalculateTotals(c.Items, c.Totals.Currency, zero)
	if err != nil {
		return err
	}

	c.CouponCode = &normalizedCode
	c.CouponPreview = &CouponView{
		CouponID: "",
		Code:     normalizedCode,
		Valid:    false,
		Discount: zero,
		Reason:   reason,
	}
	c.Totals = totals
	c.UpdatedAt = now
	return nil
}
