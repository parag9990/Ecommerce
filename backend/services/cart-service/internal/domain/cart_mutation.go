package domain

import (
	"fmt"
	"strings"
	"time"
)

type ProductVariantSnapshot struct {
	ProductID     string
	VariantID     string
	SellerID      string
	SKU           string
	Title         string
	ImageURL      string
	Attributes    map[string]string
	UnitPrice     Money
	StockQuantity int
}

type CartSummary struct {
	ItemCount       int   `json:"item_count"`
	UniqueItemCount int   `json:"unique_item_count"`
	Total           Money `json:"total"`
}

type StockError struct {
	Requested int
	Available int
}

func (e StockError) Error() string {
	return fmt.Sprintf("%v: requested=%d available=%d", ErrInsufficientStock, e.Requested, e.Available)
}

func (e StockError) Unwrap() error {
	return ErrInsufficientStock
}

func NewActiveCart(id string, owner CartOwner, currency string, now time.Time, expiresAt time.Time) (*Cart, error) {
	if err := owner.Validate(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("%w: cart id is required", ErrInvalidCart)
	}
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if !validCurrency(currency) {
		return nil, fmt.Errorf("%w: cart currency must be a 3-letter ISO code", ErrInvalidCart)
	}
	if now.IsZero() || expiresAt.IsZero() || !expiresAt.After(now) {
		return nil, fmt.Errorf("%w: valid created and expiry timestamps are required", ErrInvalidCart)
	}

	cart := &Cart{
		ID:        strings.TrimSpace(id),
		Status:    CartStatusActive,
		Items:     []CartItem{},
		Totals:    emptyTotals(currency),
		Version:   1,
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC(),
		ExpiresAt: expiresAt.UTC(),
	}
	if owner.IsUser() {
		userID := owner.UserID
		cart.UserID = &userID
	} else {
		guestSessionID := owner.GuestSessionID
		cart.GuestSessionID = &guestSessionID
	}
	return cart, nil
}

func (c *Cart) AddOrIncrementItem(snapshot ProductVariantSnapshot, addQty int, now time.Time, newItemID string) error {
	if c == nil {
		return fmt.Errorf("%w: cart is nil", ErrInvalidCart)
	}
	if c.Status != CartStatusActive {
		return ErrCartNotActive
	}
	if err := validateSnapshot(snapshot); err != nil {
		return err
	}
	if addQty < MinItemQuantity {
		return ErrQuantityTooSmall
	}
	if addQty > MaxItemQuantity {
		return ErrQuantityTooLarge
	}
	now = now.UTC()
	if now.IsZero() {
		return fmt.Errorf("%w: mutation timestamp is required", ErrInvalidCart)
	}

	for idx := range c.Items {
		item := &c.Items[idx]
		if item.ProductID != snapshot.ProductID || !variantMatchesSnapshot(item.VariantID, snapshot) {
			continue
		}
		finalQty := item.Quantity + addQty
		if finalQty > MaxItemQuantity {
			return ErrQuantityTooLarge
		}
		if finalQty > snapshot.StockQuantity {
			return StockError{Requested: finalQty, Available: snapshot.StockQuantity}
		}
		item.VariantID = snapshot.VariantID
		item.Quantity = finalQty
		item.SellerID = snapshot.SellerID
		item.SKUSnapshot = optionalString(snapshot.SKU)
		item.TitleSnapshot = strings.TrimSpace(snapshot.Title)
		item.ImageURLSnapshot = optionalString(snapshot.ImageURL)
		item.VariantSnapshot = cloneStringMap(snapshot.Attributes)
		item.UnitPrice = snapshot.UnitPrice
		item.LineSubtotal = NewMoney(snapshot.UnitPrice.Amount*int64(finalQty), snapshot.UnitPrice.Currency)
		item.PriceSnapshotAt = now
		item.UpdatedAt = now
		return nil
	}

	if len(c.Items) >= MaxUniqueItemsPerCart {
		return ErrCartItemLimitReached
	}
	if addQty > snapshot.StockQuantity {
		return StockError{Requested: addQty, Available: snapshot.StockQuantity}
	}
	if strings.TrimSpace(newItemID) == "" {
		return fmt.Errorf("%w: item id is required", ErrInvalidCartItem)
	}

	c.Items = append(c.Items, CartItem{
		ItemID:           strings.TrimSpace(newItemID),
		ProductID:        snapshot.ProductID,
		VariantID:        snapshot.VariantID,
		SellerID:         snapshot.SellerID,
		SKUSnapshot:      optionalString(snapshot.SKU),
		TitleSnapshot:    strings.TrimSpace(snapshot.Title),
		ImageURLSnapshot: optionalString(snapshot.ImageURL),
		VariantSnapshot:  cloneStringMap(snapshot.Attributes),
		UnitPrice:        snapshot.UnitPrice,
		Quantity:         addQty,
		LineSubtotal:     NewMoney(snapshot.UnitPrice.Amount*int64(addQty), snapshot.UnitPrice.Currency),
		PriceSnapshotAt:  now,
		AddedAt:          now,
		UpdatedAt:        now,
	})
	return nil
}

func (c *Cart) RemoveItem(itemID string, now time.Time) (bool, error) {
	if c == nil {
		return false, fmt.Errorf("%w: cart is nil", ErrInvalidCart)
	}
	if c.Status != CartStatusActive {
		return false, ErrCartNotActive
	}
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return false, ErrItemIDRequired
	}
	now = now.UTC()
	if now.IsZero() {
		return false, fmt.Errorf("%w: mutation timestamp is required", ErrInvalidCart)
	}

	originalLen := len(c.Items)
	filtered := c.Items[:0]
	for _, item := range c.Items {
		if strings.TrimSpace(item.ItemID) == itemID {
			continue
		}
		filtered = append(filtered, item)
	}
	c.Items = filtered

	removed := len(c.Items) != originalLen
	if removed {
		c.UpdatedAt = now
	}
	return removed, nil
}

func (c *Cart) SetItemQuantity(itemID string, quantity int, snapshot ProductVariantSnapshot, now time.Time) error {
	if c == nil {
		return fmt.Errorf("%w: cart is nil", ErrInvalidCart)
	}
	if c.Status != CartStatusActive {
		return ErrCartNotActive
	}
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return ErrItemIDRequired
	}
	if quantity < MinItemQuantity {
		return ErrQuantityTooSmall
	}
	if quantity > MaxItemQuantity {
		return ErrQuantityTooLarge
	}
	if err := validateSnapshot(snapshot); err != nil {
		return err
	}
	if quantity > snapshot.StockQuantity {
		return StockError{Requested: quantity, Available: snapshot.StockQuantity}
	}
	now = now.UTC()
	if now.IsZero() {
		return fmt.Errorf("%w: mutation timestamp is required", ErrInvalidCart)
	}

	for idx := range c.Items {
		item := &c.Items[idx]
		if strings.TrimSpace(item.ItemID) != itemID {
			continue
		}
		if item.ProductID != snapshot.ProductID || !variantMatchesSnapshot(item.VariantID, snapshot) {
			return ErrCartItemNotFound
		}
		item.VariantID = snapshot.VariantID
		item.SellerID = snapshot.SellerID
		item.SKUSnapshot = optionalString(snapshot.SKU)
		item.TitleSnapshot = strings.TrimSpace(snapshot.Title)
		item.ImageURLSnapshot = optionalString(snapshot.ImageURL)
		item.VariantSnapshot = cloneStringMap(snapshot.Attributes)
		item.UnitPrice = snapshot.UnitPrice
		item.Quantity = quantity
		item.LineSubtotal = NewMoney(snapshot.UnitPrice.Amount*int64(quantity), snapshot.UnitPrice.Currency)
		item.PriceSnapshotAt = now
		item.UpdatedAt = now
		c.UpdatedAt = now
		return nil
	}
	return ErrCartItemNotFound
}

func variantMatchesSnapshot(variantID string, snapshot ProductVariantSnapshot) bool {
	variantID = strings.TrimSpace(variantID)
	return variantID != "" && (variantID == snapshot.VariantID || variantID == snapshot.SKU)
}

func (c *Cart) CleanupZeroQuantityItems(now time.Time) int {
	if c == nil || len(c.Items) == 0 {
		return 0
	}
	now = now.UTC()

	cleaned := 0
	filtered := c.Items[:0]
	for _, item := range c.Items {
		if item.Quantity <= 0 {
			cleaned++
			continue
		}
		filtered = append(filtered, item)
	}
	c.Items = filtered
	if cleaned > 0 && !now.IsZero() {
		c.UpdatedAt = now
	}
	return cleaned
}

func (c *Cart) ClearCouponPreview() {
	if c == nil {
		return
	}
	c.CouponCode = nil
	c.CouponPreview = nil
}

func (c *Cart) RecalculateTotals(defaultCurrency string) error {
	if c == nil {
		return fmt.Errorf("%w: cart is nil", ErrInvalidCart)
	}
	defaultCurrency = strings.ToUpper(strings.TrimSpace(defaultCurrency))
	if defaultCurrency == "" && len(c.Items) > 0 {
		defaultCurrency = c.Items[0].UnitPrice.Currency
	}
	if !validCurrency(defaultCurrency) {
		return fmt.Errorf("%w: totals currency must be a 3-letter ISO code", ErrInvalidCartTotals)
	}
	for _, item := range c.Items {
		if !strings.EqualFold(item.UnitPrice.Currency, defaultCurrency) {
			return ErrMixedCurrencyCart
		}
	}
	totals, err := CalculateTotals(c.Items, defaultCurrency, NewMoney(0, defaultCurrency))
	if err != nil {
		return err
	}
	c.Totals = totals
	return nil
}

func (c *Cart) Touch(now time.Time, expiresAt time.Time) {
	if c == nil {
		return
	}
	c.UpdatedAt = now.UTC()
	if !expiresAt.IsZero() {
		c.ExpiresAt = expiresAt.UTC()
	}
}

func BuildCartSummary(cart *Cart) CartSummary {
	if cart == nil {
		return CartSummary{}
	}
	return CartSummary{
		ItemCount:       cart.Totals.ItemCount,
		UniqueItemCount: cart.Totals.UniqueItemCount,
		Total:           cart.Totals.Total,
	}
}

func validateSnapshot(snapshot ProductVariantSnapshot) error {
	required := map[string]string{
		"product_id": snapshot.ProductID,
		"variant_id": snapshot.VariantID,
		"seller_id":  snapshot.SellerID,
		"title":      snapshot.Title,
	}
	for field, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: %s is required", ErrInvalidCartItem, field)
		}
	}
	if err := snapshot.UnitPrice.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrProductPriceInvalid, err)
	}
	if snapshot.StockQuantity < 0 {
		return fmt.Errorf("%w: stock cannot be negative", ErrVariantNotSellable)
	}
	return nil
}

func emptyTotals(currency string) CartTotals {
	return CartTotals{
		Subtotal:        NewMoney(0, currency),
		Discount:        NewMoney(0, currency),
		Total:           NewMoney(0, currency),
		Currency:        currency,
		ItemCount:       0,
		UniqueItemCount: 0,
	}
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		output[key] = strings.TrimSpace(value)
	}
	if len(output) == 0 {
		return nil
	}
	return output
}
