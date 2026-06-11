package usecase

import (
	"context"
	"log/slog"
	"testing"

	"product-service/internal/domain"
	"product-service/internal/repository"
)

func TestValidateProductLoadsCategoryAndChecksSKUUniqueness(t *testing.T) {
	product := testProduct()
	service := NewCatalogModelService(
		staticCategoryReader{category: testCategory()},
		staticSKUChecker{unique: false},
		slog.Default(),
		domain.DefaultValidationOptions(),
	)

	report, err := service.ValidateProduct(context.Background(), ValidateProductRequest{
		Product:            product,
		CheckSKUUniqueness: true,
	})
	if err != nil {
		t.Fatalf("ValidateProduct returned error: %v", err)
	}
	if !hasIssue(report, domain.CodeDuplicateSKU) {
		t.Fatalf("expected duplicate SKU issue, got %+v", report.Issues)
	}
}

func TestValidateProductReportsMissingCategoryAsValidationIssue(t *testing.T) {
	service := NewCatalogModelService(
		staticCategoryReader{err: repository.ErrNotFound},
		nil,
		slog.Default(),
		domain.DefaultValidationOptions(),
	)

	report, err := service.ValidateProduct(context.Background(), ValidateProductRequest{Product: testProduct()})
	if err != nil {
		t.Fatalf("ValidateProduct returned error: %v", err)
	}
	if !hasIssue(report, domain.CodeCategoryNotFound) {
		t.Fatalf("expected category not found issue, got %+v", report.Issues)
	}
}

func TestValidateProductPropagatesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service := NewCatalogModelService(nil, nil, slog.Default(), domain.DefaultValidationOptions())

	_, err := service.ValidateProduct(ctx, ValidateProductRequest{Product: testProduct()})
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}

type staticCategoryReader struct {
	category *domain.Category
	err      error
}

func (r staticCategoryReader) GetCategoryByID(context.Context, string) (*domain.Category, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.category, nil
}

type staticSKUChecker struct {
	unique bool
	err    error
}

func (c staticSKUChecker) IsSKUUnique(context.Context, string, string) (bool, error) {
	return c.unique, c.err
}

func testProduct() domain.Product {
	return domain.Product{
		ID:         "prod_123",
		SellerID:   "seller_456",
		Title:      "Running Shoes",
		CategoryID: "cat_shoes_running",
		Status:     domain.ProductStatusDraft,
		Attributes: domain.Attributes{"material": "mesh"},
		Variants: []domain.Variant{
			{
				ID:               "var_black_9",
				SKU:              "ACME-RUN-BLK-9",
				Attributes:       domain.Attributes{"size": "9"},
				Price:            domain.NewMoney(299900, "INR"),
				StockQuantity:    10,
				ReservedQuantity: 0,
				SafetyStock:      0,
				Status:           domain.VariantStatusActive,
			},
		},
	}
}

func testCategory() *domain.Category {
	return &domain.Category{
		ID:       "cat_shoes_running",
		Name:     "Running Shoes",
		Slug:     "running-shoes",
		Path:     []string{"cat_fashion", "cat_shoes_running"},
		Level:    1,
		IsActive: true,
		AttributeSchema: []domain.AttributeDefinition{
			{
				Key:      "material",
				Label:    "Material",
				Type:     domain.AttributeTypeEnum,
				Scope:    domain.AttributeScopeProduct,
				Required: true,
				Values:   []string{"mesh"},
			},
			{
				Key:      "size",
				Label:    "Size",
				Type:     domain.AttributeTypeEnum,
				Scope:    domain.AttributeScopeVariant,
				Required: true,
				Values:   []string{"9"},
			},
		},
	}
}

func hasIssue(report domain.ValidationReport, code string) bool {
	for _, issue := range report.Issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}
