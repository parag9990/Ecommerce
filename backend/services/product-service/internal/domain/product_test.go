package domain

import (
	"testing"
	"time"
)

func TestProductValidateEnforcesCatalogRules(t *testing.T) {
	product := validProduct()
	product.Variants = append(product.Variants, product.Variants[0])
	product.Images = append(product.Images, ProductImage{
		ID:        "img_2",
		URL:       "https://cdn.example.com/products/prod_123/alt.jpg",
		Position:  1,
		IsPrimary: true,
		Status:    ImageStatusActive,
	})

	report := product.Validate(DefaultValidationOptions(), runningShoesCategory())
	assertIssue(t, report, CodeDuplicateSKU)
	assertIssue(t, report, CodeDuplicateImagePosition)
	assertIssue(t, report, CodeMultiplePrimaryImages)
	assertIssue(t, report, CodeDuplicateVariantAttributes)
}

func TestValidateForPublishAllowsZeroStockButRequiresActiveVariant(t *testing.T) {
	product := validProduct()
	product.Variants[0].StockQuantity = 0
	product.Variants[0].ReservedQuantity = 0
	product.Variants[0].SafetyStock = 0

	report := product.ValidateForPublish(DefaultValidationOptions(), runningShoesCategory())
	if report.HasErrors() {
		t.Fatalf("expected zero stock to be publishable, got issues: %+v", report.Issues)
	}

	product.Variants[0].Status = VariantStatusInactive
	report = product.ValidateForPublish(DefaultValidationOptions(), runningShoesCategory())
	assertIssue(t, report, CodeActiveVariantRequired)
}

func TestStrictAttributeSchemaRejectsUnknownAttributes(t *testing.T) {
	product := validProduct()
	product.Attributes["unknown"] = "value"

	report := product.Validate(DefaultValidationOptions(), runningShoesCategory())
	assertIssue(t, report, CodeUnknownAttributeKey)
}

func TestInventoryAvailableQuantityFormula(t *testing.T) {
	variant := validProduct().Variants[0]
	variant.StockQuantity = 10
	variant.ReservedQuantity = 3
	variant.SafetyStock = 2

	if got := variant.AvailableQuantity(); got != 5 {
		t.Fatalf("available quantity = %d, want 5", got)
	}
	if !variant.InventoryState().CanFulfill(5) {
		t.Fatal("expected inventory to fulfill available quantity")
	}
	if variant.InventoryState().CanFulfill(6) {
		t.Fatal("expected inventory not to fulfill more than available quantity")
	}
}

func TestProductLifecycleTransitions(t *testing.T) {
	product := validProduct()
	now := time.Date(2026, 5, 23, 0, 0, 0, 0, time.UTC)

	if err := product.TransitionTo(ProductStatusPublished, now); err != nil {
		t.Fatalf("publish transition failed: %v", err)
	}
	if product.PublishedAt == nil || !product.PublishedAt.Equal(now) {
		t.Fatalf("published_at not set correctly: %+v", product.PublishedAt)
	}
	if err := product.TransitionTo(ProductStatusDraft, now); err == nil {
		t.Fatal("expected published -> draft transition to fail")
	}
}

func validProduct() Product {
	now := time.Date(2026, 5, 23, 0, 0, 0, 0, time.UTC)
	return Product{
		ID:           "prod_123",
		SellerID:     "seller_456",
		Title:        "Running Shoes",
		Slug:         "running-shoes",
		Description:  "Lightweight running shoes",
		Brand:        "Acme",
		CategoryID:   "cat_shoes_running",
		CategoryPath: []string{"cat_fashion", "cat_footwear", "cat_shoes_running"},
		Status:       ProductStatusDraft,
		Attributes: Attributes{
			"material": "mesh",
			"gender":   "men",
		},
		Images: []ProductImage{
			{
				ID:        "img_1",
				URL:       "https://cdn.example.com/products/prod_123/main.jpg",
				AltText:   "Black running shoes side view",
				Position:  1,
				IsPrimary: true,
				Status:    ImageStatusActive,
				Width:     1200,
				Height:    1200,
			},
		},
		Variants: []Variant{
			{
				ID:    "var_black_9",
				SKU:   "ACME-RUN-BLK-9",
				Title: "Black / Size 9",
				Attributes: Attributes{
					"color": "black",
					"size":  "9",
				},
				Price:            NewMoney(299900, "INR"),
				MRP:              moneyPtr(NewMoney(399900, "INR")),
				StockQuantity:    120,
				ReservedQuantity: 5,
				SafetyStock:      2,
				Status:           VariantStatusActive,
				Barcode:          "8900000000012",
			},
		},
		RatingSummary: RatingSummary{},
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func runningShoesCategory() *Category {
	parentID := "cat_footwear"
	return &Category{
		ID:        "cat_shoes_running",
		Name:      "Running Shoes",
		Slug:      "running-shoes",
		ParentID:  &parentID,
		Path:      []string{"cat_fashion", "cat_footwear", "cat_shoes_running"},
		Level:     2,
		SortOrder: 10,
		IsActive:  true,
		AttributeSchema: []AttributeDefinition{
			{
				Key:        "material",
				Label:      "Material",
				Type:       AttributeTypeEnum,
				Scope:      AttributeScopeProduct,
				Required:   true,
				Filterable: true,
				Values:     []string{"mesh", "leather", "synthetic"},
			},
			{
				Key:        "gender",
				Label:      "Gender",
				Type:       AttributeTypeEnum,
				Scope:      AttributeScopeProduct,
				Required:   true,
				Filterable: true,
				Values:     []string{"men", "women", "unisex"},
			},
			{
				Key:        "size",
				Label:      "Size",
				Type:       AttributeTypeEnum,
				Scope:      AttributeScopeVariant,
				Required:   true,
				Filterable: true,
				Values:     []string{"7", "8", "9", "10", "11"},
			},
			{
				Key:        "color",
				Label:      "Color",
				Type:       AttributeTypeEnum,
				Scope:      AttributeScopeVariant,
				Required:   true,
				Filterable: true,
				Values:     []string{"black", "white"},
			},
		},
	}
}

func moneyPtr(value Money) *Money {
	return &value
}

func assertIssue(t *testing.T, report ValidationReport, code string) {
	t.Helper()
	for _, issue := range report.Issues {
		if issue.Code == code {
			return
		}
	}
	t.Fatalf("expected issue %s, got %+v", code, report.Issues)
}
