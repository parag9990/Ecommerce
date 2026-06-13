package domain

import (
	"errors"
	"testing"
)

func TestProductDocumentValidate(t *testing.T) {
	rating := 4.5
	doc := ProductDocument{
		ID:              "prod_123",
		Title:           "Nike Running Shoes",
		CategoryIDs:     []string{"cat_shoes", "cat_running"},
		SellerID:        "seller_456",
		Price:           2499,
		Rating:          &rating,
		PopularityScore: 982,
		InStock:         true,
		CreatedAt:       1735689600,
	}

	if err := doc.Validate(); err != nil {
		t.Fatalf("expected valid document, got %v", err)
	}
}

func TestProductDocumentValidateRejectsRequiredFields(t *testing.T) {
	doc := ProductDocument{
		ID:              "prod_123",
		CategoryIDs:     []string{"cat_shoes"},
		SellerID:        "seller_456",
		Price:           10,
		PopularityScore: 1,
		CreatedAt:       1735689600,
	}

	err := doc.Validate()
	if !errors.Is(err, ErrInvalidProductDocument) {
		t.Fatalf("expected ErrInvalidProductDocument, got %v", err)
	}
}

func TestProductDocumentValidateRejectsInvalidRating(t *testing.T) {
	rating := 6.0
	doc := ProductDocument{
		ID:              "prod_123",
		Title:           "Nike Running Shoes",
		CategoryIDs:     []string{"cat_shoes"},
		SellerID:        "seller_456",
		Price:           10,
		Rating:          &rating,
		PopularityScore: 1,
		CreatedAt:       1735689600,
	}

	err := doc.Validate()
	if !errors.Is(err, ErrInvalidProductDocument) {
		t.Fatalf("expected ErrInvalidProductDocument, got %v", err)
	}
}
