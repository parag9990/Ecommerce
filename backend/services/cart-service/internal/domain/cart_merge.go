package domain

import (
	"fmt"
	"strings"
	"time"
)

const (
	CartMergeWarningQuantityCapped         = "QUANTITY_CAPPED"
	CartMergeWarningUniqueItemLimitReached = "UNIQUE_ITEM_LIMIT_REACHED"
)

type CartMergeWarning struct {
	Code      string
	ProductID string
	VariantID string
	Message   string
}

type NewMergeItemIDFunc func() (string, error)

func MergeGuestItemsIntoUserCart(userCart *Cart, guestCart *Cart, now time.Time, newItemID NewMergeItemIDFunc) ([]CartMergeWarning, error) {
	if userCart == nil || guestCart == nil {
		return nil, fmt.Errorf("%w: merge carts are required", ErrInvalidCart)
	}
	if userCart.Status != CartStatusActive || guestCart.Status != CartStatusActive {
		return nil, ErrCartNotActive
	}
	if newItemID == nil {
		return nil, fmt.Errorf("%w: item id generator is required", ErrInvalidCartItem)
	}
	if !cartCurrenciesCompatible(userCart, guestCart) {
		return nil, ErrMixedCurrencyCart
	}
	now = now.UTC()
	if now.IsZero() {
		return nil, fmt.Errorf("%w: merge timestamp is required", ErrInvalidCart)
	}

	warnings := make([]CartMergeWarning, 0)
	index := make(map[string]int, len(userCart.Items))
	for idx, item := range userCart.Items {
		index[cartItemMergeKey(item.ProductID, item.VariantID)] = idx
	}

	for _, guestItem := range guestCart.Items {
		key := cartItemMergeKey(guestItem.ProductID, guestItem.VariantID)
		if userIdx, exists := index[key]; exists {
			userItem := &userCart.Items[userIdx]
			combinedQty := userItem.Quantity + guestItem.Quantity
			if combinedQty > MaxItemQuantity {
				combinedQty = MaxItemQuantity
				warnings = append(warnings, CartMergeWarning{
					Code:      CartMergeWarningQuantityCapped,
					ProductID: guestItem.ProductID,
					VariantID: guestItem.VariantID,
					Message:   "quantity adjusted to max allowed limit",
				})
			}
			userItem.Quantity = combinedQty
			userItem.LineSubtotal = NewMoney(userItem.UnitPrice.Amount*int64(combinedQty), userItem.UnitPrice.Currency)
			userItem.UpdatedAt = now
			continue
		}

		if len(userCart.Items) >= MaxUniqueItemsPerCart {
			warnings = append(warnings, CartMergeWarning{
				Code:      CartMergeWarningUniqueItemLimitReached,
				ProductID: guestItem.ProductID,
				VariantID: guestItem.VariantID,
				Message:   "item skipped because cart item limit reached",
			})
			continue
		}
		itemID, err := newItemID()
		if err != nil {
			return nil, err
		}
		itemID = strings.TrimSpace(itemID)
		if itemID == "" {
			return nil, fmt.Errorf("%w: item id is required", ErrInvalidCartItem)
		}
		copied := cloneCartItemForMerge(guestItem, itemID, now)
		userCart.Items = append(userCart.Items, copied)
		index[key] = len(userCart.Items) - 1
	}

	return warnings, nil
}

func (c *Cart) MarkMerged(targetCartID string, now time.Time) error {
	if c == nil {
		return fmt.Errorf("%w: cart is nil", ErrInvalidCart)
	}
	if c.Status != CartStatusActive {
		return ErrCartNotActive
	}
	targetCartID = strings.TrimSpace(targetCartID)
	if targetCartID == "" {
		return fmt.Errorf("%w: merged target cart id is required", ErrInvalidCart)
	}
	now = now.UTC()
	if now.IsZero() {
		return fmt.Errorf("%w: merge timestamp is required", ErrInvalidCart)
	}
	c.Status = CartStatusMerged
	c.MergedIntoCartID = &targetCartID
	c.UpdatedAt = now
	return nil
}

func cartItemMergeKey(productID string, variantID string) string {
	return strings.TrimSpace(productID) + "\x00" + strings.TrimSpace(variantID)
}

func cartCurrenciesCompatible(userCart *Cart, guestCart *Cart) bool {
	userCurrency := strings.ToUpper(strings.TrimSpace(userCart.Totals.Currency))
	guestCurrency := strings.ToUpper(strings.TrimSpace(guestCart.Totals.Currency))
	if userCurrency == "" || guestCurrency == "" {
		return true
	}
	return userCurrency == guestCurrency
}

func cloneCartItemForMerge(item CartItem, itemID string, now time.Time) CartItem {
	copied := item
	copied.ItemID = itemID
	copied.VariantSnapshot = cloneStringMap(item.VariantSnapshot)
	copied.SKUSnapshot = cloneOptionalString(item.SKUSnapshot)
	copied.ImageURLSnapshot = cloneOptionalString(item.ImageURLSnapshot)
	copied.AddedAt = now
	copied.UpdatedAt = now
	copied.LineSubtotal = NewMoney(copied.UnitPrice.Amount*int64(copied.Quantity), copied.UnitPrice.Currency)
	return copied
}

func cloneOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}
