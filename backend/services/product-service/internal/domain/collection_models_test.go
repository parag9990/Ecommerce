package domain

import (
	"testing"
	"time"
)

func TestBrandValidate(t *testing.T) {
	now := time.Date(2026, 5, 23, 0, 0, 0, 0, time.UTC)
	brand := Brand{
		ID:        "brand_acme",
		Name:      "Acme",
		Slug:      "acme",
		LogoURL:   "https://cdn.example.com/brands/acme/logo.png",
		Status:    BrandStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	report := brand.Validate()
	if report.HasErrors() {
		t.Fatalf("expected valid brand, got %+v", report.Issues)
	}

	brand.Status = "unknown"
	report = brand.Validate()
	assertIssue(t, report, CodeInvalidBrandStatus)
}

func TestInventorySnapshotValidateUsesAvailableQuantityFormula(t *testing.T) {
	now := time.Date(2026, 5, 23, 0, 0, 0, 0, time.UTC)
	snapshot := NewInventorySnapshot(
		"inv_snap_123",
		"prod_123",
		Variant{
			ID:               "var_1",
			SKU:              "ACME-SHOE-9-BLK",
			StockQuantity:    120,
			ReservedQuantity: 5,
			SafetyStock:      2,
		},
		"seller_456",
		InventorySnapshotTypeManualAdjustment,
		"Seller updated warehouse stock",
		&InventoryReference{Type: "seller_action", ID: "action_789"},
		now,
	)

	if snapshot.AvailableQuantity != 113 {
		t.Fatalf("available quantity = %d, want 113", snapshot.AvailableQuantity)
	}
	if report := snapshot.Validate(); report.HasErrors() {
		t.Fatalf("expected valid inventory snapshot, got %+v", report.Issues)
	}

	snapshot.AvailableQuantity = 114
	report := snapshot.Validate()
	assertIssue(t, report, CodeInvalidAvailableQuantity)
}

func TestPriceBookValidate(t *testing.T) {
	now := time.Date(2026, 5, 23, 0, 0, 0, 0, time.UTC)
	startsAt := now.Add(24 * time.Hour)
	endsAt := startsAt.Add(14 * 24 * time.Hour)
	priceBook := PriceBook{
		ID:        "pb_summer_sale_2026",
		SellerID:  "seller_456",
		Name:      "Summer Sale 2026",
		Currency:  "INR",
		Status:    PriceBookStatusActive,
		Priority:  10,
		StartsAt:  &startsAt,
		EndsAt:    &endsAt,
		CreatedBy: "seller_456",
		UpdatedBy: "seller_456",
		CreatedAt: now,
		UpdatedAt: now,
		Entries: []PriceBookEntry{
			{
				ID:          "pbe_1",
				ProductID:   "prod_123",
				VariantID:   "var_1",
				SKU:         "ACME-SHOE-9-BLK",
				Price:       NewMoney(249900, "INR"),
				MRP:         moneyPtr(NewMoney(399900, "INR")),
				MinQuantity: 1,
			},
		},
	}

	if report := priceBook.Validate(); report.HasErrors() {
		t.Fatalf("expected valid price book, got %+v", report.Issues)
	}

	priceBook.EndsAt = &now
	report := priceBook.Validate()
	assertIssue(t, report, CodeInvalidPriceBookWindow)
}
