package usecase

import (
	"context"
	"log/slog"
	"sort"
	"testing"

	"product-service/internal/domain"
	"product-service/internal/repository"
)

func TestListPublicProductsForcesPublishedVisibility(t *testing.T) {
	repo := newReadMemoryRepository()
	repo.saveProduct(readProduct("prod_published", "seller_456", domain.ProductStatusPublished, "cat_shoes_running"))
	repo.saveProduct(readProduct("prod_draft", "seller_456", domain.ProductStatusDraft, "cat_shoes_running"))
	service := newTestProductReadService(t, repo)

	result, err := service.ListPublicProducts(context.Background(), ListProductsRequest{
		Status:   string(domain.ProductStatusDraft),
		Page:     0,
		PageSize: 500,
	})
	if err != nil {
		t.Fatalf("ListPublicProducts returned error: %v", err)
	}
	if result.Page != 1 || result.PageSize != 2 {
		t.Fatalf("pagination = page %d size %d, want page 1 size 2", result.Page, result.PageSize)
	}
	if result.Total != 1 || len(result.Products) != 1 || result.Products[0].ID != "prod_published" {
		t.Fatalf("public products = %+v total=%d, want only published product", result.Products, result.Total)
	}
	if len(repo.lastProductFilter.Statuses) != 1 || repo.lastProductFilter.Statuses[0] != domain.ProductStatusPublished {
		t.Fatalf("statuses = %+v, want forced published", repo.lastProductFilter.Statuses)
	}
}

func TestGetPublicProductHidesUnpublishedProducts(t *testing.T) {
	repo := newReadMemoryRepository()
	repo.saveProduct(readProduct("prod_draft", "seller_456", domain.ProductStatusDraft, "cat_shoes_running"))
	service := newTestProductReadService(t, repo)

	_, err := service.GetPublicProduct(context.Background(), "prod_draft")
	assertServiceError(t, err, ErrorCodeProductNotFound)
}

func TestListSellerProductsUsesAuthenticatedSellerAndStatus(t *testing.T) {
	repo := newReadMemoryRepository()
	repo.saveProduct(readProduct("prod_own_draft", "seller_456", domain.ProductStatusDraft, "cat_shoes_running"))
	repo.saveProduct(readProduct("prod_own_rejected", "seller_456", domain.ProductStatusRejected, "cat_shoes_running"))
	repo.saveProduct(readProduct("prod_other_draft", "seller_999", domain.ProductStatusDraft, "cat_shoes_running"))
	service := newTestProductReadService(t, repo)

	result, err := service.ListSellerProducts(context.Background(), SellerListProductsRequest{
		Actor:  sellerActor(),
		Status: string(domain.ProductStatusDraft),
	})
	if err != nil {
		t.Fatalf("ListSellerProducts returned error: %v", err)
	}
	if result.Total != 1 || len(result.Products) != 1 || result.Products[0].ID != "prod_own_draft" {
		t.Fatalf("seller products = %+v total=%d, want only own draft", result.Products, result.Total)
	}
	if repo.lastProductFilter.SellerID != "seller_456" {
		t.Fatalf("seller filter = %q, want seller_456", repo.lastProductFilter.SellerID)
	}
}

func TestListSellerProductsRequiresSellerContext(t *testing.T) {
	service := newTestProductReadService(t, newReadMemoryRepository())

	_, err := service.ListSellerProducts(context.Background(), SellerListProductsRequest{
		Actor: domain.ActorContext{UserID: "user_123", Roles: []string{domain.RoleSeller}},
	})
	assertServiceError(t, err, ErrorCodePermissionDenied)
}

func TestBatchGetProductsReturnsPublishedProductsInRequestOrder(t *testing.T) {
	repo := newReadMemoryRepository()
	repo.saveProduct(readProduct("prod_1", "seller_456", domain.ProductStatusPublished, "cat_shoes_running"))
	repo.saveProduct(readProduct("prod_2", "seller_456", domain.ProductStatusDraft, "cat_shoes_running"))
	repo.saveProduct(readProduct("prod_3", "seller_456", domain.ProductStatusPublished, "cat_shoes_running"))
	service := newTestProductReadService(t, repo)

	products, err := service.BatchGetProducts(context.Background(), BatchGetProductsRequest{
		ProductIDs: []string{"prod_3", "prod_2", "prod_1", "prod_1"},
	})
	if err != nil {
		t.Fatalf("BatchGetProducts returned error: %v", err)
	}
	if len(products) != 2 || products[0].ID != "prod_3" || products[1].ID != "prod_1" {
		t.Fatalf("products = %+v, want published products in request order", products)
	}
}

func TestListCategoriesReturnsActiveCategoriesOnly(t *testing.T) {
	parentID := "cat_fashion"
	repo := newReadMemoryRepository()
	repo.categories = []domain.Category{
		{ID: "cat_inactive", Name: "Inactive", Slug: "inactive", ParentID: &parentID, Path: []string{"cat_fashion", "cat_inactive"}, IsActive: false, SortOrder: 1},
		{ID: "cat_second", Name: "Second", Slug: "second", ParentID: &parentID, Path: []string{"cat_fashion", "cat_second"}, IsActive: true, SortOrder: 20},
		{ID: "cat_first", Name: "First", Slug: "first", ParentID: &parentID, Path: []string{"cat_fashion", "cat_first"}, IsActive: true, SortOrder: 10},
	}
	service := newTestProductReadService(t, repo)

	categories, err := service.ListCategories(context.Background(), ListCategoriesRequest{ParentID: parentID})
	if err != nil {
		t.Fatalf("ListCategories returned error: %v", err)
	}
	if len(categories) != 2 || categories[0].ID != "cat_first" || categories[1].ID != "cat_second" {
		t.Fatalf("categories = %+v, want active sorted children", categories)
	}
}

func newTestProductReadService(t *testing.T, repo *readMemoryRepository) *ProductReadService {
	t.Helper()
	service, err := NewProductReadService(repo, repo, slog.Default(), ProductReadServiceOptions{
		DefaultPageSize: 2,
		MaxPageSize:     2,
		MaxBatchSize:    10,
	})
	if err != nil {
		t.Fatalf("new product read service: %v", err)
	}
	return service
}

type readMemoryRepository struct {
	products          map[string]domain.Product
	categories        []domain.Category
	lastProductFilter repository.ProductReadFilter
}

func newReadMemoryRepository() *readMemoryRepository {
	return &readMemoryRepository{
		products:   make(map[string]domain.Product),
		categories: []domain.Category{validCategory()},
	}
}

func (r *readMemoryRepository) FindProduct(ctx context.Context, filter repository.ProductReadFilter) (*domain.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.lastProductFilter = filter
	product, ok := r.products[filter.ProductID]
	if !ok || !readProductMatches(product, filter) {
		return nil, repository.ErrNotFound
	}
	return cloneProduct(product), nil
}

func (r *readMemoryRepository) ListProducts(ctx context.Context, filter repository.ProductReadFilter) ([]domain.Product, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	r.lastProductFilter = filter
	matches := make([]domain.Product, 0, len(r.products))
	for _, product := range r.products {
		if readProductMatches(product, filter) {
			matches = append(matches, *cloneProduct(product))
		}
	}
	sort.Slice(matches, func(i, j int) bool { return matches[i].ID < matches[j].ID })
	total := int64(len(matches))
	start := (filter.Page - 1) * filter.PageSize
	if start >= len(matches) {
		return []domain.Product{}, total, nil
	}
	end := start + filter.PageSize
	if end > len(matches) {
		end = len(matches)
	}
	return matches[start:end], total, nil
}

func (r *readMemoryRepository) BatchGetProducts(ctx context.Context, filter repository.ProductReadFilter) ([]domain.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.lastProductFilter = filter
	products := make([]domain.Product, 0, len(filter.ProductIDs))
	for _, id := range filter.ProductIDs {
		product, ok := r.products[id]
		if ok && readProductMatches(product, filter) {
			products = append(products, *cloneProduct(product))
		}
	}
	return products, nil
}

func (r *readMemoryRepository) ListCategories(ctx context.Context, filter repository.CategoryReadFilter) ([]domain.Category, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	categories := make([]domain.Category, 0, len(r.categories))
	for _, category := range r.categories {
		if filter.ActiveOnly && !category.IsActive {
			continue
		}
		if filter.ParentID != nil {
			if category.ParentID == nil || *category.ParentID != *filter.ParentID {
				continue
			}
		}
		categories = append(categories, category)
	}
	sort.Slice(categories, func(i, j int) bool {
		if categories[i].SortOrder == categories[j].SortOrder {
			return categories[i].Name < categories[j].Name
		}
		return categories[i].SortOrder < categories[j].SortOrder
	})
	return categories, nil
}

func (r *readMemoryRepository) saveProduct(product domain.Product) {
	r.products[product.ID] = product
}

func readProductMatches(product domain.Product, filter repository.ProductReadFilter) bool {
	if filter.SellerID != "" && product.SellerID != filter.SellerID {
		return false
	}
	if filter.CategoryID != "" && product.CategoryID != filter.CategoryID && !containsString(product.CategoryPath, filter.CategoryID) {
		return false
	}
	if len(filter.Statuses) > 0 && !containsStatus(filter.Statuses, product.Status) {
		return false
	}
	return true
}

func readProduct(id, sellerID string, status domain.ProductStatus, categoryID string) domain.Product {
	product := validDraftProduct()
	product.ID = id
	product.SellerID = sellerID
	product.Status = status
	product.CategoryID = categoryID
	product.CategoryPath = []string{"cat_fashion", categoryID}
	return product
}

func containsStatus(statuses []domain.ProductStatus, status domain.ProductStatus) bool {
	for _, candidate := range statuses {
		if candidate == status {
			return true
		}
	}
	return false
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
