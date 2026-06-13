package domain

import (
	"errors"
	"testing"
)

func TestProductStatusTransitions(t *testing.T) {
	tests := []struct {
		name    string
		from    ProductStatus
		to      ProductStatus
		allowed bool
	}{
		{name: "draft to submitted", from: ProductStatusDraft, to: ProductStatusSubmitted, allowed: true},
		{name: "submitted to approved", from: ProductStatusSubmitted, to: ProductStatusApproved, allowed: true},
		{name: "submitted to rejected", from: ProductStatusSubmitted, to: ProductStatusRejected, allowed: true},
		{name: "approved to published", from: ProductStatusApproved, to: ProductStatusPublished, allowed: true},
		{name: "published to unpublished", from: ProductStatusPublished, to: ProductStatusUnpublished, allowed: true},
		{name: "draft cannot publish", from: ProductStatusDraft, to: ProductStatusPublished, allowed: false},
		{name: "rejected cannot publish", from: ProductStatusRejected, to: ProductStatusPublished, allowed: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanTransitionProduct(tt.from, tt.to); got != tt.allowed {
				t.Fatalf("CanTransitionProduct(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.allowed)
			}
		})
	}
}

func TestValidateProductReadyForReview(t *testing.T) {
	product := validReviewProduct()

	if err := ValidateProductReadyForReview(product); err != nil {
		t.Fatalf("expected product to be review ready: %v", err)
	}
}

func TestValidateProductReadyForReviewReportsFieldErrors(t *testing.T) {
	product := validReviewProduct()
	product.Images = nil
	product.Variants = []ProductVariant{
		{SKU: "SKU-1", Price: Money{Amount: 1000, Currency: "INR"}},
		{SKU: "sku-1", Price: Money{Amount: 0, Currency: "INR"}},
	}

	err := ValidateProductReadyForReview(product)
	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if len(validationErr.Fields) != 3 {
		t.Fatalf("expected 3 field violations, got %+v", validationErr.Fields)
	}
}

func TestMajorEditRequiresReview(t *testing.T) {
	if !MajorEditRequiresReview([]string{"stock_quantity", "images"}) {
		t.Fatal("expected image edits to require review")
	}
	if MajorEditRequiresReview([]string{"stock_quantity", "price.amount"}) {
		t.Fatal("expected inventory and price edits to remain minor")
	}
}

func validReviewProduct() Product {
	return Product{
		ID:          "prod_1",
		SellerID:    "seller_1",
		Title:       "Running Shoes",
		Description: "Comfortable running shoes for daily training.",
		CategoryID:  "cat_1",
		Brand:       "Acme",
		Status:      ProductStatusDraft,
		Images: []ProductImage{
			{URL: "https://cdn.example.com/prod_1/main.jpg"},
		},
		Variants: []ProductVariant{
			{SKU: "SKU-1", Price: Money{Amount: 120000, Currency: "INR"}},
		},
	}
}
