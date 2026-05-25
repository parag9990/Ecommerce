package domain

import (
	"errors"
	"testing"
	"time"
)

func TestCartValidateAcceptsActiveUserCart(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	cart := Cart{
		ID:        "cart_123",
		UserID:    &userID,
		Status:    CartStatusActive,
		Items:     []CartItem{validItem(now)},
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(90 * 24 * time.Hour),
	}
	cart.Totals = CartTotals{
		Subtotal:        NewMoney(2000, CurrencyINR),
		Discount:        NewMoney(100, CurrencyINR),
		Total:           NewMoney(1900, CurrencyINR),
		Currency:        CurrencyINR,
		ItemCount:       2,
		UniqueItemCount: 1,
	}

	if err := cart.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestCartValidateRejectsActiveCartWithBothOwners(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	guestSessionID := "sess_123"
	cart := Cart{
		ID:             "cart_123",
		UserID:         &userID,
		GuestSessionID: &guestSessionID,
		Status:         CartStatusActive,
		Items:          []CartItem{},
		Totals: CartTotals{
			Subtotal:        NewMoney(0, CurrencyINR),
			Discount:        NewMoney(0, CurrencyINR),
			Total:           NewMoney(0, CurrencyINR),
			Currency:        CurrencyINR,
			ItemCount:       0,
			UniqueItemCount: 0,
		},
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(30 * 24 * time.Hour),
	}

	err := cart.Validate()
	if !errors.Is(err, ErrInvalidCartOwner) {
		t.Fatalf("Validate() error = %v, want ErrInvalidCartOwner", err)
	}
}

func TestCartValidateRejectsDuplicateProductVariant(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	item1 := validItem(now)
	item2 := validItem(now)
	item2.ItemID = "item_2"
	cart := Cart{
		ID:        "cart_123",
		UserID:    &userID,
		Status:    CartStatusActive,
		Items:     []CartItem{item1, item2},
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(90 * 24 * time.Hour),
		Totals: CartTotals{
			Subtotal:        NewMoney(4000, CurrencyINR),
			Discount:        NewMoney(0, CurrencyINR),
			Total:           NewMoney(4000, CurrencyINR),
			Currency:        CurrencyINR,
			ItemCount:       4,
			UniqueItemCount: 2,
		},
	}

	err := cart.Validate()
	if !errors.Is(err, ErrDuplicateCartItem) {
		t.Fatalf("Validate() error = %v, want ErrDuplicateCartItem", err)
	}
}

func TestCartValidateRejectsInconsistentTotals(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	cart := Cart{
		ID:        "cart_123",
		UserID:    &userID,
		Status:    CartStatusActive,
		Items:     []CartItem{validItem(now)},
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(90 * 24 * time.Hour),
		Totals: CartTotals{
			Subtotal:        NewMoney(9999, CurrencyINR),
			Discount:        NewMoney(0, CurrencyINR),
			Total:           NewMoney(9999, CurrencyINR),
			Currency:        CurrencyINR,
			ItemCount:       2,
			UniqueItemCount: 1,
		},
	}

	err := cart.Validate()
	if !errors.Is(err, ErrInvalidCartTotals) {
		t.Fatalf("Validate() error = %v, want ErrInvalidCartTotals", err)
	}
}

func TestCartAddOrIncrementItemCreatesNewLine(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	owner := CartOwner{Type: CartOwnerTypeUser, UserID: "user_123"}
	cart, err := NewActiveCart("cart_123", owner, CurrencyINR, now, now.Add(90*24*time.Hour))
	if err != nil {
		t.Fatalf("NewActiveCart() error = %v", err)
	}

	if err := cart.AddOrIncrementItem(validSnapshot(), 2, now, "item_1"); err != nil {
		t.Fatalf("AddOrIncrementItem() error = %v", err)
	}
	if err := cart.RecalculateTotals(CurrencyINR); err != nil {
		t.Fatalf("RecalculateTotals() error = %v", err)
	}

	if len(cart.Items) != 1 {
		t.Fatalf("item count = %d, want 1", len(cart.Items))
	}
	if cart.Items[0].Quantity != 2 || cart.Totals.Subtotal.Amount != 2000 {
		t.Fatalf("unexpected cart after add: %#v", cart)
	}
}

func TestCartAddOrIncrementItemIncrementsExistingLine(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	cart := Cart{
		ID:        "cart_123",
		UserID:    &userID,
		Status:    CartStatusActive,
		Items:     []CartItem{validItem(now)},
		Totals:    validTotals(),
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(90 * 24 * time.Hour),
	}

	if err := cart.AddOrIncrementItem(validSnapshot(), 1, now.Add(time.Minute), "unused"); err != nil {
		t.Fatalf("AddOrIncrementItem() error = %v", err)
	}
	if cart.Items[0].Quantity != 3 {
		t.Fatalf("quantity = %d, want 3", cart.Items[0].Quantity)
	}
}

func TestCartAddOrIncrementItemRejectsStockShortage(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	owner := CartOwner{Type: CartOwnerTypeGuest, GuestSessionID: "guest_123"}
	cart, err := NewActiveCart("cart_123", owner, CurrencyINR, now, now.Add(90*24*time.Hour))
	if err != nil {
		t.Fatalf("NewActiveCart() error = %v", err)
	}
	snapshot := validSnapshot()
	snapshot.StockQuantity = 1

	err = cart.AddOrIncrementItem(snapshot, 2, now, "item_1")
	if !errors.Is(err, ErrInsufficientStock) {
		t.Fatalf("AddOrIncrementItem() error = %v, want ErrInsufficientStock", err)
	}
}

func TestCartRemoveItemDeletesExistingItem(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	item1 := validItem(now)
	item2 := validItem(now)
	item2.ItemID = "item_2"
	item2.ProductID = "prod_456"
	item2.VariantID = "var_2"
	cart := Cart{
		ID:        "cart_123",
		UserID:    &userID,
		Status:    CartStatusActive,
		Items:     []CartItem{item1, item2},
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(90 * 24 * time.Hour),
	}

	removed, err := cart.RemoveItem("item_1", now.Add(time.Minute))
	if err != nil {
		t.Fatalf("RemoveItem() error = %v", err)
	}
	if !removed {
		t.Fatal("RemoveItem() removed = false, want true")
	}
	if len(cart.Items) != 1 || cart.Items[0].ItemID != "item_2" {
		t.Fatalf("items = %#v, want only item_2", cart.Items)
	}
}

func TestCartRemoveItemIsIdempotentWhenMissing(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	cart := Cart{
		ID:        "cart_123",
		UserID:    &userID,
		Status:    CartStatusActive,
		Items:     []CartItem{validItem(now)},
		Totals:    validTotals(),
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(90 * 24 * time.Hour),
	}

	removed, err := cart.RemoveItem("missing_item", now.Add(time.Minute))
	if err != nil {
		t.Fatalf("RemoveItem() error = %v", err)
	}
	if removed {
		t.Fatal("RemoveItem() removed = true, want false")
	}
	if len(cart.Items) != 1 {
		t.Fatalf("item count = %d, want 1", len(cart.Items))
	}
}

func TestCartCleanupZeroQuantityItems(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	zero := validItem(now)
	zero.ItemID = "item_zero"
	zero.ProductID = "prod_zero"
	zero.VariantID = "var_zero"
	zero.Quantity = 0
	zero.LineSubtotal = NewMoney(0, CurrencyINR)
	negative := validItem(now)
	negative.ItemID = "item_negative"
	negative.ProductID = "prod_negative"
	negative.VariantID = "var_negative"
	negative.Quantity = -1
	negative.LineSubtotal = NewMoney(0, CurrencyINR)
	valid := validItem(now)
	valid.ItemID = "item_valid"
	valid.ProductID = "prod_valid"
	valid.VariantID = "var_valid"
	cart := Cart{
		ID:        "cart_123",
		UserID:    &userID,
		Status:    CartStatusActive,
		Items:     []CartItem{zero, negative, valid},
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(90 * 24 * time.Hour),
	}

	cleaned := cart.CleanupZeroQuantityItems(now.Add(time.Minute))
	if cleaned != 2 {
		t.Fatalf("cleaned = %d, want 2", cleaned)
	}
	if len(cart.Items) != 1 || cart.Items[0].ItemID != "item_valid" {
		t.Fatalf("items = %#v, want only item_valid", cart.Items)
	}
}

func TestCartRemoveItemRecalculatesEmptyTotals(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	cart := Cart{
		ID:        "cart_123",
		UserID:    &userID,
		Status:    CartStatusActive,
		Items:     []CartItem{validItem(now)},
		Totals:    validTotals(),
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(90 * 24 * time.Hour),
	}

	if _, err := cart.RemoveItem("item_1", now.Add(time.Minute)); err != nil {
		t.Fatalf("RemoveItem() error = %v", err)
	}
	if err := cart.RecalculateTotals(CurrencyINR); err != nil {
		t.Fatalf("RecalculateTotals() error = %v", err)
	}
	if len(cart.Items) != 0 || cart.Totals.Total.Amount != 0 || cart.Totals.ItemCount != 0 {
		t.Fatalf("unexpected empty cart totals: %#v", cart)
	}
}

func TestCartRemoveItemRejectsInactiveCart(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	cart := Cart{
		ID:        "cart_123",
		UserID:    &userID,
		Status:    CartStatusCheckedOut,
		Items:     []CartItem{validItem(now)},
		Totals:    validTotals(),
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(90 * 24 * time.Hour),
	}

	_, err := cart.RemoveItem("item_1", now.Add(time.Minute))
	if !errors.Is(err, ErrCartNotActive) {
		t.Fatalf("RemoveItem() error = %v, want ErrCartNotActive", err)
	}
}

func validItem(now time.Time) CartItem {
	return CartItem{
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
	}
}

func validSnapshot() ProductVariantSnapshot {
	return ProductVariantSnapshot{
		ProductID:     "prod_123",
		VariantID:     "var_1",
		SellerID:      "seller_456",
		Title:         "Running Shoes",
		UnitPrice:     NewMoney(1000, CurrencyINR),
		StockQuantity: 10,
	}
}

func validTotals() CartTotals {
	return CartTotals{
		Subtotal:        NewMoney(2000, CurrencyINR),
		Discount:        NewMoney(0, CurrencyINR),
		Total:           NewMoney(2000, CurrencyINR),
		Currency:        CurrencyINR,
		ItemCount:       2,
		UniqueItemCount: 1,
	}
}
