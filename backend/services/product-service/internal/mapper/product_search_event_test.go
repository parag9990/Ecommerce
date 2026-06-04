package mapper

import (
	"testing"
	"time"

	"product-service/internal/domain"
)

func TestToProductSearchEventPayloadMapsPublishedProductToUpsert(t *testing.T) {
	now := time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
	product := domain.Product{
		ID:           "prod_123",
		SellerID:     "seller_456",
		Title:        "Running Shoes",
		Description:  "Lightweight running shoes",
		Brand:        "Acme",
		CategoryID:   "cat_running",
		CategoryPath: []string{"cat_fashion", "cat_running"},
		Status:       domain.ProductStatusPublished,
		Attributes: domain.Attributes{
			"material": "mesh",
		},
		Images: []domain.ProductImage{
			{
				URL:       "https://cdn.example.com/products/prod_123/main.jpg",
				Position:  1,
				IsPrimary: true,
				Status:    domain.ImageStatusActive,
			},
		},
		Variants: []domain.Variant{
			{
				ID:               "var_1",
				SKU:              "SKU-1",
				Attributes:       domain.Attributes{"size": "9", "color": "black"},
				Price:            domain.NewMoney(249900, "INR"),
				StockQuantity:    5,
				ReservedQuantity: 1,
				Status:           domain.VariantStatusActive,
			},
		},
		RatingSummary: domain.RatingSummary{Average: 4.5, Count: 25},
		UpdatedAt:     now,
	}

	payload := ToProductSearchEventPayload(product)
	if payload.SearchAction != domain.SearchActionUpsert {
		t.Fatalf("search action = %s, want UPSERT", payload.SearchAction)
	}
	if !payload.InStock {
		t.Fatal("expected product to be in stock")
	}
	if payload.Price.Amount != 249900 || payload.Price.Currency != "INR" {
		t.Fatalf("price = %+v", payload.Price)
	}
	if payload.ImageURL == "" {
		t.Fatal("expected image url")
	}
	if payload.Attributes["size"] != "9" || payload.Attributes["material"] != "mesh" {
		t.Fatalf("attributes = %+v", payload.Attributes)
	}
	if payload.PopularityScore <= 0 {
		t.Fatalf("popularity score = %d, want positive", payload.PopularityScore)
	}
}

func TestToProductSearchEventPayloadMapsDraftProductToDelete(t *testing.T) {
	payload := ToProductSearchEventPayload(domain.Product{
		ID:        "prod_123",
		SellerID:  "seller_456",
		Status:    domain.ProductStatusDraft,
		UpdatedAt: time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC),
	})
	if payload.SearchAction != domain.SearchActionDelete {
		t.Fatalf("search action = %s, want DELETE", payload.SearchAction)
	}
}
