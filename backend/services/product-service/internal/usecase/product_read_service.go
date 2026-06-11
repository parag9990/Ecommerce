package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"product-service/internal/domain"
	"product-service/internal/repository"
)

const (
	defaultReadPageSize = 20
	defaultMaxPageSize  = 100
	defaultMaxBatchSize = 100
)

type ProductReadUseCase interface {
	ListPublicProducts(ctx context.Context, request ListProductsRequest) (ProductListResult, error)
	GetPublicProduct(ctx context.Context, productID string) (*domain.Product, error)
	BatchGetProducts(ctx context.Context, request BatchGetProductsRequest) ([]domain.Product, error)
	ListCategories(ctx context.Context, request ListCategoriesRequest) ([]domain.Category, error)
	ListSellerProducts(ctx context.Context, request SellerListProductsRequest) (ProductListResult, error)
	GetSellerProduct(ctx context.Context, request SellerGetProductRequest) (*domain.Product, error)
}

type ProductReadServiceOptions struct {
	DefaultPageSize int
	MaxPageSize     int
	MaxBatchSize    int
}

type ListProductsRequest struct {
	CategoryID string
	SellerID   string
	Status     string
	Page       int
	PageSize   int
	Sort       string
}

type SellerListProductsRequest struct {
	Actor      domain.ActorContext
	CategoryID string
	Status     string
	Page       int
	PageSize   int
	Sort       string
}

type SellerGetProductRequest struct {
	Actor     domain.ActorContext
	ProductID string
}

type BatchGetProductsRequest struct {
	ProductIDs []string
}

type ListCategoriesRequest struct {
	ParentID string
}

type ProductListResult struct {
	Products []domain.Product
	Total    int64
	Page     int
	PageSize int
}

type ProductReadService struct {
	products   repository.ProductReadRepository
	categories repository.CategoryReadRepository
	logger     *slog.Logger
	options    ProductReadServiceOptions
}

func NewProductReadService(
	products repository.ProductReadRepository,
	categories repository.CategoryReadRepository,
	logger *slog.Logger,
	options ProductReadServiceOptions,
) (*ProductReadService, error) {
	if products == nil {
		return nil, fmt.Errorf("product read repository is required")
	}
	if categories == nil {
		return nil, fmt.Errorf("category read repository is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	options = normalizeProductReadOptions(options)
	return &ProductReadService{
		products:   products,
		categories: categories,
		logger:     logger,
		options:    options,
	}, nil
}

func (s *ProductReadService) ListPublicProducts(ctx context.Context, request ListProductsRequest) (ProductListResult, error) {
	if err := ctx.Err(); err != nil {
		return ProductListResult{}, err
	}

	page, pageSize := s.normalizePagination(request.Page, request.PageSize)
	sort, err := normalizePublicReadSort(request.Sort)
	if err != nil {
		return ProductListResult{}, err
	}

	products, total, err := s.products.ListProducts(ctx, repository.ProductReadFilter{
		SellerID:   strings.TrimSpace(request.SellerID),
		CategoryID: strings.TrimSpace(request.CategoryID),
		Statuses:   []domain.ProductStatus{domain.ProductStatusPublished},
		Page:       page,
		PageSize:   pageSize,
		Sort:       sort,
	})
	if err != nil {
		return ProductListResult{}, fmt.Errorf("list public products: %w", err)
	}

	s.logger.Debug(
		"public product list read",
		"viewer_type", "public",
		"seller_id", strings.TrimSpace(request.SellerID),
		"category_id", strings.TrimSpace(request.CategoryID),
		"page_size", pageSize,
		"result_count", len(products),
		"total", total,
		"sort", sort,
	)
	return ProductListResult{Products: products, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *ProductReadService) GetPublicProduct(ctx context.Context, productID string) (*domain.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	productID = strings.TrimSpace(productID)
	if productID == "" {
		return nil, serviceError(ErrorKindInvalidArgument, ErrorCodeValidation, "product id is required", nil)
	}

	product, err := s.products.FindProduct(ctx, repository.ProductReadFilter{
		ProductID: productID,
		Statuses:  []domain.ProductStatus{domain.ProductStatusPublished},
	})
	if errors.Is(err, repository.ErrNotFound) {
		return nil, serviceError(ErrorKindNotFound, ErrorCodeProductNotFound, "product not found", err)
	}
	if err != nil {
		return nil, fmt.Errorf("get public product: %w", err)
	}
	return product, nil
}

func (s *ProductReadService) BatchGetProducts(ctx context.Context, request BatchGetProductsRequest) ([]domain.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	ids, err := s.normalizeBatchIDs(request.ProductIDs)
	if err != nil {
		return nil, err
	}

	products, err := s.products.BatchGetProducts(ctx, repository.ProductReadFilter{
		ProductIDs: ids,
		Statuses:   []domain.ProductStatus{domain.ProductStatusPublished},
	})
	if err != nil {
		return nil, fmt.Errorf("batch get public products: %w", err)
	}
	s.logger.Debug("batch product read", "viewer_type", "internal", "requested_count", len(ids), "result_count", len(products))
	return products, nil
}

func (s *ProductReadService) ListCategories(ctx context.Context, request ListCategoriesRequest) ([]domain.Category, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var parentID *string
	if trimmed := strings.TrimSpace(request.ParentID); trimmed != "" {
		parentID = &trimmed
	}
	categories, err := s.categories.ListCategories(ctx, repository.CategoryReadFilter{
		ParentID:   parentID,
		ActiveOnly: true,
	})
	if err != nil {
		return nil, fmt.Errorf("list public categories: %w", err)
	}
	s.logger.Debug("category browse read", "viewer_type", "public", "parent_id", strings.TrimSpace(request.ParentID), "result_count", len(categories))
	return categories, nil
}

func (s *ProductReadService) ListSellerProducts(ctx context.Context, request SellerListProductsRequest) (ProductListResult, error) {
	if err := ctx.Err(); err != nil {
		return ProductListResult{}, err
	}
	if err := authorizeSellerRead(request.Actor); err != nil {
		return ProductListResult{}, err
	}

	statuses, err := sellerReadableStatuses(request.Status)
	if err != nil {
		return ProductListResult{}, err
	}
	page, pageSize := s.normalizePagination(request.Page, request.PageSize)
	sort, err := normalizeSellerReadSort(request.Sort)
	if err != nil {
		return ProductListResult{}, err
	}

	sellerID := strings.TrimSpace(request.Actor.SellerID)
	products, total, err := s.products.ListProducts(ctx, repository.ProductReadFilter{
		SellerID:   sellerID,
		CategoryID: strings.TrimSpace(request.CategoryID),
		Statuses:   statuses,
		Page:       page,
		PageSize:   pageSize,
		Sort:       sort,
	})
	if err != nil {
		return ProductListResult{}, fmt.Errorf("list seller products: %w", err)
	}

	s.logger.Debug(
		"seller product list read",
		"viewer_type", "seller",
		"seller_id", sellerID,
		"category_id", strings.TrimSpace(request.CategoryID),
		"page_size", pageSize,
		"result_count", len(products),
		"total", total,
		"sort", sort,
	)
	return ProductListResult{Products: products, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *ProductReadService) GetSellerProduct(ctx context.Context, request SellerGetProductRequest) (*domain.Product, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := authorizeSellerRead(request.Actor); err != nil {
		return nil, err
	}
	productID := strings.TrimSpace(request.ProductID)
	if productID == "" {
		return nil, serviceError(ErrorKindInvalidArgument, ErrorCodeValidation, "product id is required", nil)
	}

	product, err := s.products.FindProduct(ctx, repository.ProductReadFilter{
		ProductID: productID,
		SellerID:  strings.TrimSpace(request.Actor.SellerID),
	})
	if errors.Is(err, repository.ErrNotFound) {
		return nil, serviceError(ErrorKindNotFound, ErrorCodeProductNotFound, "product not found", err)
	}
	if err != nil {
		return nil, fmt.Errorf("get seller product: %w", err)
	}
	return product, nil
}

func normalizeProductReadOptions(options ProductReadServiceOptions) ProductReadServiceOptions {
	if options.DefaultPageSize <= 0 {
		options.DefaultPageSize = defaultReadPageSize
	}
	if options.MaxPageSize <= 0 {
		options.MaxPageSize = defaultMaxPageSize
	}
	if options.DefaultPageSize > options.MaxPageSize {
		options.DefaultPageSize = options.MaxPageSize
	}
	if options.MaxBatchSize <= 0 {
		options.MaxBatchSize = defaultMaxBatchSize
	}
	return options
}

func (s *ProductReadService) normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = s.options.DefaultPageSize
	}
	if pageSize > s.options.MaxPageSize {
		pageSize = s.options.MaxPageSize
	}
	return page, pageSize
}

func (s *ProductReadService) normalizeBatchIDs(input []string) ([]string, error) {
	if len(input) == 0 {
		return nil, serviceError(ErrorKindInvalidArgument, ErrorCodeValidation, "at least one product id is required", nil)
	}
	if len(input) > s.options.MaxBatchSize {
		return nil, serviceError(ErrorKindInvalidArgument, ErrorCodeValidation, "too many product ids requested", nil)
	}

	ids := make([]string, 0, len(input))
	seen := make(map[string]struct{}, len(input))
	for _, id := range input {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, serviceError(ErrorKindInvalidArgument, ErrorCodeValidation, "at least one product id is required", nil)
	}
	return ids, nil
}

func normalizePublicReadSort(value string) (repository.ProductReadSort, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", string(repository.ProductReadSortNewest):
		return repository.ProductReadSortNewest, nil
	case string(repository.ProductReadSortUpdated):
		return repository.ProductReadSortUpdated, nil
	case string(repository.ProductReadSortPriceAsc):
		return repository.ProductReadSortPriceAsc, nil
	case string(repository.ProductReadSortPriceDesc):
		return repository.ProductReadSortPriceDesc, nil
	case string(repository.ProductReadSortRatingDesc):
		return repository.ProductReadSortRatingDesc, nil
	default:
		return "", serviceError(ErrorKindInvalidArgument, ErrorCodeValidation, "unsupported product sort", nil)
	}
}

func normalizeSellerReadSort(value string) (repository.ProductReadSort, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", string(repository.ProductReadSortUpdated):
		return repository.ProductReadSortUpdated, nil
	case string(repository.ProductReadSortNewest):
		return repository.ProductReadSortNewest, nil
	case string(repository.ProductReadSortPriceAsc):
		return repository.ProductReadSortPriceAsc, nil
	case string(repository.ProductReadSortPriceDesc):
		return repository.ProductReadSortPriceDesc, nil
	case string(repository.ProductReadSortRatingDesc):
		return repository.ProductReadSortRatingDesc, nil
	default:
		return "", serviceError(ErrorKindInvalidArgument, ErrorCodeValidation, "unsupported product sort", nil)
	}
}

func sellerReadableStatuses(requested string) ([]domain.ProductStatus, error) {
	if strings.TrimSpace(requested) == "" {
		return []domain.ProductStatus{
			domain.ProductStatusDraft,
			domain.ProductStatusSubmitted,
			domain.ProductStatusPublished,
			domain.ProductStatusUnpublished,
			domain.ProductStatusRejected,
		}, nil
	}
	status := domain.ProductStatus(strings.ToLower(strings.TrimSpace(requested)))
	if !status.Valid() {
		return nil, serviceError(ErrorKindInvalidArgument, ErrorCodeValidation, "unsupported product status", nil)
	}
	return []domain.ProductStatus{status}, nil
}

func authorizeSellerRead(actor domain.ActorContext) error {
	if actor.ActorID() == "" {
		return serviceError(ErrorKindUnauthenticated, ErrorCodeUnauthenticated, "authenticated actor is required", nil)
	}
	if strings.TrimSpace(actor.SellerID) == "" {
		return serviceError(ErrorKindPermissionDenied, ErrorCodePermissionDenied, "seller id is required for seller product reads", nil)
	}
	if !actor.CanReadSellerProducts() {
		return serviceError(ErrorKindPermissionDenied, ErrorCodePermissionDenied, "actor cannot read seller products", nil)
	}
	return nil
}
