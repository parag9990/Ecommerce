package repository

import (
	"context"

	"product-service/internal/domain"
)

type ProductReadSort string

const (
	ProductReadSortNewest     ProductReadSort = "newest"
	ProductReadSortUpdated    ProductReadSort = "updated_at_desc"
	ProductReadSortPriceAsc   ProductReadSort = "price_asc"
	ProductReadSortPriceDesc  ProductReadSort = "price_desc"
	ProductReadSortRatingDesc ProductReadSort = "rating_desc"
)

type ProductReadFilter struct {
	ProductID  string
	ProductIDs []string
	SellerID   string
	CategoryID string
	Statuses   []domain.ProductStatus
	Page       int
	PageSize   int
	Sort       ProductReadSort
}

type CategoryReadFilter struct {
	ParentID   *string
	ActiveOnly bool
}

type ProductReadRepository interface {
	FindProduct(ctx context.Context, filter ProductReadFilter) (*domain.Product, error)
	ListProducts(ctx context.Context, filter ProductReadFilter) ([]domain.Product, int64, error)
	BatchGetProducts(ctx context.Context, filter ProductReadFilter) ([]domain.Product, error)
}

type CategoryReadRepository interface {
	ListCategories(ctx context.Context, filter CategoryReadFilter) ([]domain.Category, error)
}
