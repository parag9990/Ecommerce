package usecase

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"product-service/internal/client"
	"product-service/internal/domain"
	"product-service/internal/repository"
)

func TestCreateProductCreatesSellerDraft(t *testing.T) {
	repo := newMemoryProductRepository()
	service := newTestSellerProductService(t, repo, client.ModerationAutoPublish)

	product, err := service.CreateProduct(context.Background(), CreateProductRequest{
		Actor:   sellerActor(),
		Product: validProductInput(),
	})
	if err != nil {
		t.Fatalf("CreateProduct returned error: %v", err)
	}
	if product.Status != domain.ProductStatusDraft {
		t.Fatalf("status = %s, want draft", product.Status)
	}
	if product.SellerID != "seller_456" {
		t.Fatalf("seller id = %s, want seller_456", product.SellerID)
	}
	if product.ID == "" || product.Variants[0].ID == "" || product.Images[0].ID == "" {
		t.Fatalf("expected generated product, variant, and image ids: %+v", product)
	}
	if product.Variants[0].ReservedQuantity != 0 {
		t.Fatalf("reserved quantity = %d, want 0", product.Variants[0].ReservedQuantity)
	}
}

func TestUpdateProductRejectsOwnershipMismatch(t *testing.T) {
	repo := newMemoryProductRepository()
	service := newTestSellerProductService(t, repo, client.ModerationAutoPublish)
	repo.save(validDraftProduct())

	_, err := service.UpdateProduct(context.Background(), UpdateProductRequest{
		Actor: domain.ActorContext{
			UserID:   "user_999",
			SellerID: "seller_999",
			Roles:    []string{domain.RoleSeller},
		},
		ProductID: "prod_123",
		Product:   validProductInput(),
	})
	assertServiceError(t, err, ErrorCodeProductOwnership)
}

func TestUpdateProductRejectsPublishedProduct(t *testing.T) {
	repo := newMemoryProductRepository()
	service := newTestSellerProductService(t, repo, client.ModerationAutoPublish)
	product := validDraftProduct()
	product.Status = domain.ProductStatusPublished
	repo.save(product)

	_, err := service.UpdateProduct(context.Background(), UpdateProductRequest{
		Actor:     sellerActor(),
		ProductID: "prod_123",
		Product:   validProductInput(),
	})
	assertServiceError(t, err, ErrorCodeProductNotEditable)
}

func TestPublishProductAutoPublishesWhenCMSAllows(t *testing.T) {
	repo := newMemoryProductRepository()
	service := newTestSellerProductService(t, repo, client.ModerationAutoPublish)
	repo.save(validDraftProduct())

	product, err := service.PublishProduct(context.Background(), ProductLifecycleRequest{
		Actor:     sellerActor(),
		ProductID: "prod_123",
	})
	if err != nil {
		t.Fatalf("PublishProduct returned error: %v", err)
	}
	if product.Status != domain.ProductStatusPublished {
		t.Fatalf("status = %s, want published", product.Status)
	}
	if product.PublishedAt == nil {
		t.Fatal("expected published_at to be set")
	}
}

func TestPublishProductSubmitsWhenCMSRequiresReview(t *testing.T) {
	repo := newMemoryProductRepository()
	service := newTestSellerProductService(t, repo, client.ModerationReviewRequired)
	repo.save(validDraftProduct())

	product, err := service.PublishProduct(context.Background(), ProductLifecycleRequest{
		Actor:     sellerActor(),
		ProductID: "prod_123",
	})
	if err != nil {
		t.Fatalf("PublishProduct returned error: %v", err)
	}
	if product.Status != domain.ProductStatusSubmitted {
		t.Fatalf("status = %s, want submitted", product.Status)
	}
	if product.PublishedAt != nil {
		t.Fatalf("published_at = %v, want nil", product.PublishedAt)
	}
}

func TestUnpublishProductMovesPublishedProductToUnpublished(t *testing.T) {
	repo := newMemoryProductRepository()
	service := newTestSellerProductService(t, repo, client.ModerationAutoPublish)
	product := validDraftProduct()
	product.Status = domain.ProductStatusPublished
	now := testNow()
	product.PublishedAt = &now
	repo.save(product)

	got, err := service.UnpublishProduct(context.Background(), ProductLifecycleRequest{
		Actor:     sellerActor(),
		ProductID: "prod_123",
	})
	if err != nil {
		t.Fatalf("UnpublishProduct returned error: %v", err)
	}
	if got.Status != domain.ProductStatusUnpublished {
		t.Fatalf("status = %s, want unpublished", got.Status)
	}
}

func newTestSellerProductService(t *testing.T, repo *memoryProductRepository, decision client.ModerationDecision) *SellerProductService {
	t.Helper()
	validator := NewCatalogModelService(repo, repo, slog.Default(), domain.DefaultValidationOptions())
	cmsClient, err := client.NewStaticCMSPolicyClient(client.StaticCMSPolicy{
		CatalogManagementAllowed: true,
		ModerationDecision:       decision,
	})
	if err != nil {
		t.Fatalf("new cms client: %v", err)
	}
	service, err := NewSellerProductService(
		repo,
		validator,
		cmsClient,
		sequenceIDs{},
		fixedClock{},
		slog.Default(),
		SellerProductServiceOptions{},
	)
	if err != nil {
		t.Fatalf("new seller product service: %v", err)
	}
	return service
}

type memoryProductRepository struct {
	products map[string]domain.Product
	skus     map[string]string
	category domain.Category
}

func newMemoryProductRepository() *memoryProductRepository {
	return &memoryProductRepository{
		products: make(map[string]domain.Product),
		skus:     make(map[string]string),
		category: validCategory(),
	}
}

func (r *memoryProductRepository) InsertProduct(ctx context.Context, product *domain.Product) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, exists := r.products[product.ID]; exists {
		return repository.ErrDuplicateKey
	}
	for _, variant := range product.Variants {
		if owner, exists := r.skus[variant.SKU]; exists && owner != product.ID {
			return repository.ErrDuplicateKey
		}
	}
	r.save(*product)
	return nil
}

func (r *memoryProductRepository) FindProductByID(ctx context.Context, productID string) (*domain.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	product, exists := r.products[productID]
	if !exists {
		return nil, repository.ErrNotFound
	}
	return cloneProduct(product), nil
}

func (r *memoryProductRepository) UpdateProduct(ctx context.Context, product *domain.Product, expectedStatuses ...domain.ProductStatus) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	current, exists := r.products[product.ID]
	if !exists {
		return repository.ErrWriteConflict
	}
	if current.SellerID != product.SellerID {
		return repository.ErrWriteConflict
	}
	if len(expectedStatuses) > 0 {
		matched := false
		for _, status := range expectedStatuses {
			if current.Status == status {
				matched = true
				break
			}
		}
		if !matched {
			return repository.ErrWriteConflict
		}
	}
	for _, variant := range product.Variants {
		if owner, exists := r.skus[variant.SKU]; exists && owner != product.ID {
			return repository.ErrDuplicateKey
		}
	}
	r.save(*product)
	return nil
}

func (r *memoryProductRepository) GetCategoryByID(ctx context.Context, categoryID string) (*domain.Category, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if categoryID != r.category.ID {
		return nil, repository.ErrNotFound
	}
	category := r.category
	return &category, nil
}

func (r *memoryProductRepository) IsSKUUnique(ctx context.Context, sku string, excludeProductID string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	owner, exists := r.skus[sku]
	return !exists || owner == excludeProductID, nil
}

func (r *memoryProductRepository) save(product domain.Product) {
	for sku, owner := range r.skus {
		if owner == product.ID {
			delete(r.skus, sku)
		}
	}
	r.products[product.ID] = product
	for _, variant := range product.Variants {
		r.skus[variant.SKU] = product.ID
	}
}

type sequenceIDs struct{}

func (sequenceIDs) NewProductID() string { return "prod_generated" }
func (sequenceIDs) NewVariantID() string { return "var_generated" }
func (sequenceIDs) NewImageID() string   { return "img_generated" }

type fixedClock struct{}

func (fixedClock) Now() time.Time { return testNow() }

func testNow() time.Time {
	return time.Date(2026, 5, 23, 12, 0, 0, 0, time.UTC)
}

func sellerActor() domain.ActorContext {
	return domain.ActorContext{
		UserID:      "user_123",
		SellerID:    "seller_456",
		Roles:       []string{domain.RoleSellerCatalogEditor},
		Permissions: []string{domain.PermissionProductWriteOwnSeller},
	}
}

func validProductInput() ProductInput {
	return ProductInput{
		Title:       "Running Shoes",
		Description: "Lightweight running shoes",
		Brand:       "Acme",
		CategoryID:  "cat_shoes_running",
		Attributes: domain.Attributes{
			"material": "mesh",
			"gender":   "men",
		},
		Images: []domain.ProductImage{
			{
				URL:       "https://cdn.example.com/products/prod_123/main.jpg",
				AltText:   "Black running shoes",
				Position:  1,
				IsPrimary: true,
				Status:    domain.ImageStatusActive,
			},
		},
		Variants: []domain.Variant{
			{
				SKU: "ACME-RUN-BLK-9",
				Attributes: domain.Attributes{
					"size":  "9",
					"color": "black",
				},
				Price:         domain.NewMoney(299900, "INR"),
				StockQuantity: 120,
				Status:        domain.VariantStatusActive,
			},
		},
	}
}

func validDraftProduct() domain.Product {
	now := testNow()
	input := validProductInput()
	return domain.Product{
		ID:          "prod_123",
		SellerID:    "seller_456",
		Title:       input.Title,
		Description: input.Description,
		Brand:       input.Brand,
		CategoryID:  input.CategoryID,
		Status:      domain.ProductStatusDraft,
		Attributes:  input.Attributes,
		Images: []domain.ProductImage{
			{
				ID:        "img_123",
				URL:       input.Images[0].URL,
				AltText:   input.Images[0].AltText,
				Position:  1,
				IsPrimary: true,
				Status:    domain.ImageStatusActive,
			},
		},
		Variants: []domain.Variant{
			{
				ID:               "var_123",
				SKU:              input.Variants[0].SKU,
				Attributes:       input.Variants[0].Attributes,
				Price:            input.Variants[0].Price,
				StockQuantity:    input.Variants[0].StockQuantity,
				ReservedQuantity: 0,
				Status:           domain.VariantStatusActive,
			},
		},
		CreatedBy: "user_123",
		UpdatedBy: "user_123",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func validCategory() domain.Category {
	return domain.Category{
		ID:        "cat_shoes_running",
		Name:      "Running Shoes",
		Slug:      "running-shoes",
		Path:      []string{"cat_fashion", "cat_shoes_running"},
		Level:     1,
		IsActive:  true,
		SortOrder: 1,
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
				Key:      "gender",
				Label:    "Gender",
				Type:     domain.AttributeTypeEnum,
				Scope:    domain.AttributeScopeProduct,
				Required: true,
				Values:   []string{"men", "women"},
			},
			{
				Key:      "size",
				Label:    "Size",
				Type:     domain.AttributeTypeEnum,
				Scope:    domain.AttributeScopeVariant,
				Required: true,
				Values:   []string{"9"},
			},
			{
				Key:      "color",
				Label:    "Color",
				Type:     domain.AttributeTypeEnum,
				Scope:    domain.AttributeScopeVariant,
				Required: true,
				Values:   []string{"black"},
			},
		},
	}
}

func cloneProduct(product domain.Product) *domain.Product {
	clone := product
	clone.Images = append([]domain.ProductImage(nil), product.Images...)
	clone.Variants = append([]domain.Variant(nil), product.Variants...)
	return &clone
}

func assertServiceError(t *testing.T, err error, code string) {
	t.Helper()
	var serviceErr *ServiceError
	if !errors.As(err, &serviceErr) {
		t.Fatalf("expected service error %s, got %v", code, err)
	}
	if serviceErr.Code != code {
		t.Fatalf("error code = %s, want %s", serviceErr.Code, code)
	}
}
