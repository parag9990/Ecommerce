package repository

import (
	"context"
	"errors"

	"product-service/internal/domain"
)

var ErrNotFound = errors.New("repository: not found")
var ErrDuplicateKey = errors.New("repository: duplicate key")
var ErrWriteConflict = errors.New("repository: write conflict")
var ErrInsufficientStock = errors.New("repository: insufficient stock")
var ErrInventoryUnavailable = errors.New("repository: inventory unavailable")

type CategoryReader interface {
	GetCategoryByID(ctx context.Context, categoryID string) (*domain.Category, error)
}

type SKUUniquenessChecker interface {
	IsSKUUnique(ctx context.Context, sku string, excludeProductID string) (bool, error)
}

type CatalogModelRepository interface {
	CategoryReader
	SKUUniquenessChecker
}

type ProductRepository interface {
	CategoryReader
	SKUUniquenessChecker
	InsertProduct(ctx context.Context, product *domain.Product) error
	FindProductByID(ctx context.Context, productID string) (*domain.Product, error)
	UpdateProduct(ctx context.Context, product *domain.Product, expectedStatuses ...domain.ProductStatus) error
}
