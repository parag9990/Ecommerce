package domain

import (
	"fmt"
	"strings"
	"time"
)

const (
	MaxUniqueItemsPerCart = 100
	MinItemQuantity       = 1
	MaxItemQuantity       = 10
)

type CartStatus string

const (
	CartStatusActive     CartStatus = "active"
	CartStatusMerged     CartStatus = "merged"
	CartStatusCheckedOut CartStatus = "checked_out"
	CartStatusExpired    CartStatus = "expired"
	CartStatusAbandoned  CartStatus = "abandoned"
)

func (s CartStatus) Valid() bool {
	switch s {
	case CartStatusActive, CartStatusMerged, CartStatusCheckedOut, CartStatusExpired, CartStatusAbandoned:
		return true
	default:
		return false
	}
}

type Cart struct {
	ID                string      `bson:"_id" json:"cart_id"`
	UserID            *string     `bson:"user_id,omitempty" json:"user_id,omitempty"`
	GuestSessionID    *string     `bson:"guest_session_id,omitempty" json:"guest_session_id,omitempty"`
	Status            CartStatus  `bson:"status" json:"status"`
	Items             []CartItem  `bson:"items" json:"items"`
	CouponCode        *string     `bson:"coupon_code,omitempty" json:"coupon_code,omitempty"`
	CouponPreview     *CouponView `bson:"coupon_preview,omitempty" json:"coupon_preview,omitempty"`
	Totals            CartTotals  `bson:"totals" json:"totals"`
	Version           int64       `bson:"version" json:"version"`
	CreatedAt         time.Time   `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time   `bson:"updated_at" json:"updated_at"`
	ExpiresAt         time.Time   `bson:"expires_at" json:"expires_at"`
	MergedIntoCartID  *string     `bson:"merged_into_cart_id,omitempty" json:"merged_into_cart_id,omitempty"`
	CheckedOutOrderID *string     `bson:"checked_out_order_id,omitempty" json:"checked_out_order_id,omitempty"`
}

type CartItem struct {
	ItemID           string            `bson:"item_id" json:"item_id"`
	ProductID        string            `bson:"product_id" json:"product_id"`
	VariantID        string            `bson:"variant_id" json:"variant_id"`
	SellerID         string            `bson:"seller_id" json:"seller_id"`
	SKUSnapshot      *string           `bson:"sku_snapshot,omitempty" json:"sku_snapshot,omitempty"`
	TitleSnapshot    string            `bson:"title_snapshot" json:"title_snapshot"`
	ImageURLSnapshot *string           `bson:"image_url_snapshot,omitempty" json:"image_url_snapshot,omitempty"`
	VariantSnapshot  map[string]string `bson:"variant_snapshot,omitempty" json:"variant_snapshot,omitempty"`
	UnitPrice        Money             `bson:"unit_price" json:"unit_price"`
	Quantity         int               `bson:"quantity" json:"quantity"`
	LineSubtotal     Money             `bson:"line_subtotal" json:"line_subtotal"`
	PriceSnapshotAt  time.Time         `bson:"price_snapshot_at" json:"price_snapshot_at"`
	AddedAt          time.Time         `bson:"added_at" json:"added_at"`
	UpdatedAt        time.Time         `bson:"updated_at" json:"updated_at"`
}

type CartTotals struct {
	Subtotal        Money  `bson:"subtotal" json:"subtotal"`
	Discount        Money  `bson:"discount" json:"discount"`
	Total           Money  `bson:"total" json:"total"`
	Currency        string `bson:"currency" json:"currency"`
	ItemCount       int    `bson:"item_count" json:"item_count"`
	UniqueItemCount int    `bson:"unique_item_count" json:"unique_item_count"`
}

type CouponView struct {
	CouponID string `bson:"coupon_id" json:"coupon_id"`
	Code     string `bson:"code" json:"code"`
	Valid    bool   `bson:"valid" json:"valid"`
	Discount Money  `bson:"discount" json:"discount"`
	Reason   string `bson:"reason,omitempty" json:"reason,omitempty"`
}

func (c Cart) Validate() error {
	if strings.TrimSpace(c.ID) == "" {
		return fmt.Errorf("%w: cart id is required", ErrInvalidCart)
	}
	if !c.Status.Valid() {
		return fmt.Errorf("%w: %q", ErrInvalidCartStatus, c.Status)
	}
	if err := validateOwner(c.Status, c.UserID, c.GuestSessionID); err != nil {
		return err
	}
	if len(c.Items) > MaxUniqueItemsPerCart {
		return fmt.Errorf("%w: max %d unique items allowed", ErrCartItemLimitReached, MaxUniqueItemsPerCart)
	}
	if c.Version < 1 {
		return fmt.Errorf("%w: version must be greater than zero", ErrInvalidCart)
	}
	if c.CreatedAt.IsZero() || c.UpdatedAt.IsZero() || c.ExpiresAt.IsZero() {
		return fmt.Errorf("%w: timestamps are required", ErrInvalidCart)
	}
	if c.UpdatedAt.Before(c.CreatedAt) {
		return fmt.Errorf("%w: updated_at cannot be before created_at", ErrInvalidCart)
	}
	if c.CouponCode != nil {
		if _, err := NormalizeCouponCode(*c.CouponCode); err != nil {
			return fmt.Errorf("coupon_code: %w", err)
		}
	}

	seen := make(map[string]struct{}, len(c.Items))
	for i, item := range c.Items {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("items[%d]: %w", i, err)
		}
		key := item.ProductID + "\x00" + item.VariantID
		if _, ok := seen[key]; ok {
			return fmt.Errorf("%w: product_id=%s variant_id=%s", ErrDuplicateCartItem, item.ProductID, item.VariantID)
		}
		seen[key] = struct{}{}
	}

	expectedTotals, err := CalculateTotals(c.Items, c.Totals.Currency, c.Totals.Discount)
	if err != nil {
		return err
	}
	if err := c.Totals.ValidateAgainst(expectedTotals); err != nil {
		return err
	}
	if c.CouponPreview != nil {
		previewCode, err := NormalizeCouponCode(c.CouponPreview.Code)
		if err != nil {
			return fmt.Errorf("coupon_preview.code: %w", err)
		}
		if c.CouponCode == nil {
			return fmt.Errorf("%w: coupon_code is required when coupon_preview is present", ErrInvalidCouponPreview)
		}
		couponCode, err := NormalizeCouponCode(*c.CouponCode)
		if err != nil {
			return fmt.Errorf("coupon_code: %w", err)
		}
		if couponCode != previewCode {
			return fmt.Errorf("%w: coupon_code must match coupon_preview.code", ErrInvalidCouponPreview)
		}
		if err := c.CouponPreview.Discount.Validate(); err != nil {
			return fmt.Errorf("coupon_preview.discount: %w", err)
		}
		if !c.CouponPreview.Discount.SameCurrency(c.Totals.Discount) ||
			c.CouponPreview.Discount.Amount != c.Totals.Discount.Amount {
			return fmt.Errorf("%w: coupon preview discount must match cart totals discount", ErrInvalidCouponPreview)
		}
		if c.CouponPreview.Valid && strings.TrimSpace(c.CouponPreview.CouponID) == "" {
			return fmt.Errorf("%w: coupon_id is required for valid preview", ErrInvalidCouponPreview)
		}
		if !c.CouponPreview.Valid && c.CouponPreview.Discount.Amount != 0 {
			return fmt.Errorf("%w: invalid coupon preview discount must be zero", ErrInvalidCouponPreview)
		}
	}
	return nil
}

func (i CartItem) Validate() error {
	required := map[string]string{
		"item_id":        i.ItemID,
		"product_id":     i.ProductID,
		"variant_id":     i.VariantID,
		"seller_id":      i.SellerID,
		"title_snapshot": i.TitleSnapshot,
	}
	for field, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: %s is required", ErrInvalidCartItem, field)
		}
	}
	if i.Quantity < MinItemQuantity || i.Quantity > MaxItemQuantity {
		return fmt.Errorf("%w: quantity must be between %d and %d", ErrInvalidCartItem, MinItemQuantity, MaxItemQuantity)
	}
	if err := i.UnitPrice.Validate(); err != nil {
		return fmt.Errorf("unit_price: %w", err)
	}
	if err := i.LineSubtotal.Validate(); err != nil {
		return fmt.Errorf("line_subtotal: %w", err)
	}
	if !i.UnitPrice.SameCurrency(i.LineSubtotal) {
		return fmt.Errorf("%w: line_subtotal currency must match unit_price", ErrInvalidMoney)
	}
	if i.LineSubtotal.Amount != i.UnitPrice.Amount*int64(i.Quantity) {
		return fmt.Errorf("%w: line_subtotal must equal unit_price * quantity", ErrInvalidCartItem)
	}
	if i.PriceSnapshotAt.IsZero() || i.AddedAt.IsZero() || i.UpdatedAt.IsZero() {
		return fmt.Errorf("%w: item timestamps are required", ErrInvalidCartItem)
	}
	if i.UpdatedAt.Before(i.AddedAt) {
		return fmt.Errorf("%w: updated_at cannot be before added_at", ErrInvalidCartItem)
	}
	return nil
}

func CalculateTotals(items []CartItem, currency string, discount Money) (CartTotals, error) {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if !validCurrency(currency) {
		return CartTotals{}, fmt.Errorf("%w: totals currency must be a 3-letter ISO code", ErrInvalidCartTotals)
	}
	if err := discount.Validate(); err != nil {
		return CartTotals{}, fmt.Errorf("discount: %w", err)
	}
	if !strings.EqualFold(discount.Currency, currency) {
		return CartTotals{}, fmt.Errorf("%w: discount currency must match totals currency", ErrInvalidCartTotals)
	}

	var subtotal int64
	var itemCount int
	for _, item := range items {
		if err := item.Validate(); err != nil {
			return CartTotals{}, err
		}
		if !strings.EqualFold(item.UnitPrice.Currency, currency) {
			return CartTotals{}, fmt.Errorf("%w: item currency must match totals currency", ErrInvalidCartTotals)
		}
		subtotal += item.LineSubtotal.Amount
		itemCount += item.Quantity
	}

	total := subtotal - discount.Amount
	if total < 0 {
		total = 0
	}
	return CartTotals{
		Subtotal:        NewMoney(subtotal, currency),
		Discount:        NewMoney(discount.Amount, currency),
		Total:           NewMoney(total, currency),
		Currency:        currency,
		ItemCount:       itemCount,
		UniqueItemCount: len(items),
	}, nil
}

func (t CartTotals) ValidateAgainst(expected CartTotals) error {
	if err := t.Subtotal.Validate(); err != nil {
		return fmt.Errorf("subtotal: %w", err)
	}
	if err := t.Discount.Validate(); err != nil {
		return fmt.Errorf("discount: %w", err)
	}
	if err := t.Total.Validate(); err != nil {
		return fmt.Errorf("total: %w", err)
	}
	if !validCurrency(t.Currency) {
		return fmt.Errorf("%w: currency must be a 3-letter ISO code", ErrInvalidCartTotals)
	}
	if !strings.EqualFold(t.Subtotal.Currency, t.Currency) ||
		!strings.EqualFold(t.Discount.Currency, t.Currency) ||
		!strings.EqualFold(t.Total.Currency, t.Currency) {
		return fmt.Errorf("%w: all money currencies must match totals currency", ErrInvalidCartTotals)
	}
	if t.Subtotal.Amount != expected.Subtotal.Amount ||
		t.Discount.Amount != expected.Discount.Amount ||
		t.Total.Amount != expected.Total.Amount ||
		t.ItemCount != expected.ItemCount ||
		t.UniqueItemCount != expected.UniqueItemCount ||
		!strings.EqualFold(t.Currency, expected.Currency) {
		return fmt.Errorf("%w: totals snapshot does not match cart items", ErrInvalidCartTotals)
	}
	return nil
}

func validateOwner(status CartStatus, userID *string, guestSessionID *string) error {
	hasUser := userID != nil && strings.TrimSpace(*userID) != ""
	hasGuest := guestSessionID != nil && strings.TrimSpace(*guestSessionID) != ""
	if !hasUser && !hasGuest {
		return fmt.Errorf("%w: user_id or guest_session_id is required", ErrInvalidCartOwner)
	}
	if status == CartStatusActive && hasUser == hasGuest {
		return fmt.Errorf("%w: active cart must have exactly one owner", ErrInvalidCartOwner)
	}
	return nil
}
