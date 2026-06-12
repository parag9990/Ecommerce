package domain

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestMergeGuestItemsCopiesNewItemsWithNewItemID(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	guestSessionID := "sess_123"
	userCart := &Cart{
		ID:        "cart_user",
		UserID:    &userID,
		Status:    CartStatusActive,
		Items:     []CartItem{},
		Totals:    emptyTotals(CurrencyINR),
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
	}
	guestCart := &Cart{
		ID:             "cart_guest",
		GuestSessionID: &guestSessionID,
		Status:         CartStatusActive,
		Items:          []CartItem{validItem(now)},
		Totals:         validTotals(),
		Version:        1,
		CreatedAt:      now,
		UpdatedAt:      now,
		ExpiresAt:      now.Add(24 * time.Hour),
	}

	warnings, err := MergeGuestItemsIntoUserCart(userCart, guestCart, now.Add(time.Minute), staticItemID("item_user_copy"))
	if err != nil {
		t.Fatalf("MergeGuestItemsIntoUserCart() error = %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %#v, want none", warnings)
	}
	if len(userCart.Items) != 1 || userCart.Items[0].ItemID != "item_user_copy" {
		t.Fatalf("items = %#v, want copied item with new item id", userCart.Items)
	}
}

func TestMergeGuestItemsCombinesDuplicateVariantQuantity(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	guestSessionID := "sess_123"
	userCart := mergeUserCart(userID, now)
	guestItem := validItem(now)
	guestItem.ItemID = "guest_item_1"
	guestItem.Quantity = 3
	guestItem.LineSubtotal = NewMoney(3000, CurrencyINR)
	guestCart := mergeGuestCart(guestSessionID, now, []CartItem{guestItem})

	warnings, err := MergeGuestItemsIntoUserCart(userCart, guestCart, now.Add(time.Minute), staticItemID("unused"))
	if err != nil {
		t.Fatalf("MergeGuestItemsIntoUserCart() error = %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %#v, want none", warnings)
	}
	if userCart.Items[0].Quantity != 5 || userCart.Items[0].LineSubtotal.Amount != 5000 {
		t.Fatalf("item = %#v, want combined quantity and subtotal", userCart.Items[0])
	}
}

func TestMergeGuestItemsCapsQuantityAtMax(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	guestSessionID := "sess_123"
	userCart := mergeUserCart(userID, now)
	userCart.Items[0].Quantity = 7
	userCart.Items[0].LineSubtotal = NewMoney(7000, CurrencyINR)
	userCart.Totals = CartTotals{
		Subtotal:        NewMoney(7000, CurrencyINR),
		Discount:        NewMoney(0, CurrencyINR),
		Total:           NewMoney(7000, CurrencyINR),
		Currency:        CurrencyINR,
		ItemCount:       7,
		UniqueItemCount: 1,
	}
	guestItem := validItem(now)
	guestItem.ItemID = "guest_item_1"
	guestItem.Quantity = 6
	guestItem.LineSubtotal = NewMoney(6000, CurrencyINR)
	guestCart := mergeGuestCart(guestSessionID, now, []CartItem{guestItem})

	warnings, err := MergeGuestItemsIntoUserCart(userCart, guestCart, now.Add(time.Minute), staticItemID("unused"))
	if err != nil {
		t.Fatalf("MergeGuestItemsIntoUserCart() error = %v", err)
	}
	if userCart.Items[0].Quantity != MaxItemQuantity {
		t.Fatalf("quantity = %d, want %d", userCart.Items[0].Quantity, MaxItemQuantity)
	}
	if len(warnings) != 1 || warnings[0].Code != CartMergeWarningQuantityCapped {
		t.Fatalf("warnings = %#v, want quantity capped warning", warnings)
	}
}

func TestMergeGuestItemsRejectsMixedCurrencyCart(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	guestSessionID := "sess_123"
	userCart := mergeUserCart(userID, now)
	guestCart := mergeGuestCart(guestSessionID, now, []CartItem{validItem(now)})
	guestCart.Totals.Currency = "USD"
	guestCart.Totals.Subtotal.Currency = "USD"
	guestCart.Totals.Discount.Currency = "USD"
	guestCart.Totals.Total.Currency = "USD"

	_, err := MergeGuestItemsIntoUserCart(userCart, guestCart, now.Add(time.Minute), staticItemID("unused"))
	if !errors.Is(err, ErrMixedCurrencyCart) {
		t.Fatalf("MergeGuestItemsIntoUserCart() error = %v, want ErrMixedCurrencyCart", err)
	}
}

func TestCartMarkMergedSetsStatusAndTarget(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	guestSessionID := "sess_123"
	cart := mergeGuestCart(guestSessionID, now, []CartItem{validItem(now)})

	if err := cart.MarkMerged("cart_user", now.Add(time.Minute)); err != nil {
		t.Fatalf("MarkMerged() error = %v", err)
	}
	if cart.Status != CartStatusMerged || cart.MergedIntoCartID == nil || *cart.MergedIntoCartID != "cart_user" {
		t.Fatalf("cart = %#v, want merged into cart_user", cart)
	}
}

func mergeUserCart(userID string, now time.Time) *Cart {
	return &Cart{
		ID:        "cart_user",
		UserID:    &userID,
		Status:    CartStatusActive,
		Items:     []CartItem{validItem(now)},
		Totals:    validTotals(),
		Version:   2,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
	}
}

func mergeGuestCart(guestSessionID string, now time.Time, items []CartItem) *Cart {
	return &Cart{
		ID:             "cart_guest",
		GuestSessionID: &guestSessionID,
		Status:         CartStatusActive,
		Items:          items,
		Totals: CartTotals{
			Subtotal:        NewMoney(sumItemSubtotals(items), CurrencyINR),
			Discount:        NewMoney(0, CurrencyINR),
			Total:           NewMoney(sumItemSubtotals(items), CurrencyINR),
			Currency:        CurrencyINR,
			ItemCount:       sumItemQuantities(items),
			UniqueItemCount: len(items),
		},
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
	}
}

func staticItemID(itemID string) NewMergeItemIDFunc {
	return func() (string, error) {
		return itemID, nil
	}
}

func sumItemSubtotals(items []CartItem) int64 {
	var total int64
	for _, item := range items {
		total += item.LineSubtotal.Amount
	}
	return total
}

func sumItemQuantities(items []CartItem) int {
	total := 0
	for _, item := range items {
		total += item.Quantity
	}
	return total
}

func TestMergeGuestItemsSkipsWhenUniqueItemLimitReached(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	guestSessionID := "sess_123"
	userCart := mergeUserCart(userID, now)
	userCart.Items = userCart.Items[:0]
	for i := 0; i < MaxUniqueItemsPerCart; i++ {
		item := validItem(now)
		item.ItemID = fmt.Sprintf("item_%d", i)
		item.ProductID = fmt.Sprintf("prod_%d", i)
		item.VariantID = fmt.Sprintf("var_%d", i)
		userCart.Items = append(userCart.Items, item)
	}
	if err := userCart.RecalculateTotals(CurrencyINR); err != nil {
		t.Fatalf("RecalculateTotals() error = %v", err)
	}
	guestItem := validItem(now)
	guestItem.ItemID = "guest_item"
	guestItem.ProductID = "guest_prod"
	guestItem.VariantID = "guest_var"
	guestCart := mergeGuestCart(guestSessionID, now, []CartItem{guestItem})

	warnings, err := MergeGuestItemsIntoUserCart(userCart, guestCart, now.Add(time.Minute), staticItemID("item_new"))
	if err != nil {
		t.Fatalf("MergeGuestItemsIntoUserCart() error = %v", err)
	}
	if len(userCart.Items) != MaxUniqueItemsPerCart {
		t.Fatalf("item count = %d, want cap %d", len(userCart.Items), MaxUniqueItemsPerCart)
	}
	if len(warnings) != 1 || warnings[0].Code != CartMergeWarningUniqueItemLimitReached {
		t.Fatalf("warnings = %#v, want unique item limit warning", warnings)
	}
}
