package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewWishlistDefaultsToSinglePrivateEmptyWishlist(t *testing.T) {
	createdAt := time.Date(2026, 5, 18, 10, 30, 0, 0, time.FixedZone("IST", 5*60*60+30*60))

	wishlist, err := NewWishlist(" wish_123 ", " user_123 ", createdAt)
	if err != nil {
		t.Fatalf("NewWishlist returned error: %v", err)
	}

	if wishlist.ID != "wish_123" {
		t.Fatalf("wishlist.ID = %q, want wish_123", wishlist.ID)
	}
	if wishlist.UserID != "user_123" {
		t.Fatalf("wishlist.UserID = %q, want user_123", wishlist.UserID)
	}
	if wishlist.Visibility != VisibilityPrivate {
		t.Fatalf("wishlist.Visibility = %q, want %q", wishlist.Visibility, VisibilityPrivate)
	}
	if len(wishlist.Items) != 0 {
		t.Fatalf("len(wishlist.Items) = %d, want 0", len(wishlist.Items))
	}
	if wishlist.CreatedAt.Location() != time.UTC {
		t.Fatalf("CreatedAt location = %v, want UTC", wishlist.CreatedAt.Location())
	}
}

func TestWishlistAddItemStoresSnapshotsAndBlocksDuplicateProduct(t *testing.T) {
	wishlist := mustWishlist(t)
	addedAt := time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC)

	err := wishlist.AddItem(AddWishlistItemInput{
		ProductID: " prod_123 ",
		VariantID: " var_1 ",
		LastKnownPrice: &Money{
			Amount:   299900,
			Currency: " inr ",
		},
		Availability: AvailabilityInStock,
	}, addedAt)
	if err != nil {
		t.Fatalf("AddItem returned error: %v", err)
	}

	item, ok := wishlist.FindItem("prod_123")
	if !ok {
		t.Fatal("FindItem(prod_123) = false, want true")
	}
	if item.VariantID != "var_1" {
		t.Fatalf("item.VariantID = %q, want var_1", item.VariantID)
	}
	if item.LastKnownPrice == nil {
		t.Fatal("item.LastKnownPrice is nil")
	}
	if item.LastKnownPrice.Currency != "INR" {
		t.Fatalf("currency = %q, want INR", item.LastKnownPrice.Currency)
	}
	if wishlist.UpdatedAt != addedAt {
		t.Fatalf("UpdatedAt = %v, want %v", wishlist.UpdatedAt, addedAt)
	}

	err = wishlist.AddItem(AddWishlistItemInput{
		ProductID: "prod_123",
		VariantID: "var_2",
	}, addedAt.Add(time.Minute))
	if !errors.Is(err, ErrDuplicateProduct) {
		t.Fatalf("duplicate AddItem error = %v, want ErrDuplicateProduct", err)
	}
}

func TestWishlistRemoveProduct(t *testing.T) {
	wishlist := mustWishlist(t)
	addedAt := time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC)
	removedAt := addedAt.Add(time.Hour)

	if err := wishlist.AddItem(AddWishlistItemInput{ProductID: "prod_123"}, addedAt); err != nil {
		t.Fatalf("AddItem returned error: %v", err)
	}
	if err := wishlist.RemoveProduct("prod_123", removedAt); err != nil {
		t.Fatalf("RemoveProduct returned error: %v", err)
	}
	if wishlist.HasProduct("prod_123") {
		t.Fatal("HasProduct(prod_123) = true, want false")
	}
	if wishlist.UpdatedAt != removedAt {
		t.Fatalf("UpdatedAt = %v, want %v", wishlist.UpdatedAt, removedAt)
	}

	err := wishlist.RemoveProduct("prod_404", removedAt.Add(time.Minute))
	if !errors.Is(err, ErrWishlistItemNotFound) {
		t.Fatalf("missing RemoveProduct error = %v, want ErrWishlistItemNotFound", err)
	}
}

func TestRehydrateWishlistRejectsUnsupportedVisibility(t *testing.T) {
	now := time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC)

	_, err := RehydrateWishlist(Wishlist{
		ID:         "wish_123",
		UserID:     "user_123",
		Visibility: "public",
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	if !errors.Is(err, ErrInvalidWishlist) {
		t.Fatalf("RehydrateWishlist error = %v, want ErrInvalidWishlist", err)
	}
}

func TestWishlistValidationRejectsInvalidItemSnapshot(t *testing.T) {
	wishlist := mustWishlist(t)

	err := wishlist.AddItem(AddWishlistItemInput{
		ProductID: "prod_123",
		LastKnownPrice: &Money{
			Amount:   -1,
			Currency: "INR",
		},
	}, time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC))
	if !errors.Is(err, ErrInvalidWishlist) {
		t.Fatalf("AddItem error = %v, want ErrInvalidWishlist", err)
	}
}

func TestWishlistItemAllowsDeletedAvailabilitySnapshot(t *testing.T) {
	item, err := NewWishlistItem(AddWishlistItemInput{
		ProductID:    "prod_123",
		Availability: AvailabilityDeleted,
	}, time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewWishlistItem returned error: %v", err)
	}
	if item.Availability != AvailabilityDeleted {
		t.Fatalf("Availability = %q, want %q", item.Availability, AvailabilityDeleted)
	}
}

func mustWishlist(t *testing.T) *Wishlist {
	t.Helper()

	wishlist, err := NewWishlist("wish_123", "user_123", time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewWishlist returned error: %v", err)
	}
	return wishlist
}
