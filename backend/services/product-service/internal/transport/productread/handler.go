package productread

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"product-service/internal/transport/dto"
	"product-service/internal/usecase"
)

type Handler struct {
	useCase usecase.ProductReadUseCase
	logger  *slog.Logger
}

func NewHandler(useCase usecase.ProductReadUseCase, logger *slog.Logger) (*Handler, error) {
	if useCase == nil {
		return nil, fmt.Errorf("product read usecase is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{useCase: useCase, logger: logger}, nil
}

func (h *Handler) ListProducts(ctx context.Context, request dto.ListProductsRequestDTO) (*dto.ProductListResponseDTO, error) {
	result, err := h.useCase.ListPublicProducts(ctx, request.ToUseCase())
	if err != nil {
		h.logger.Error("list public products failed", "category_id", request.CategoryID, "seller_id", request.SellerID, "error", err)
		return nil, err
	}
	response := dto.PublicProductListFromUseCase(result)
	return &response, nil
}

func (h *Handler) GetProduct(ctx context.Context, request dto.GetProductRequestDTO) (*dto.ProductReadDTO, error) {
	productID := strings.TrimSpace(request.ProductID)
	product, err := h.useCase.GetPublicProduct(ctx, productID)
	if err != nil {
		h.logger.Error("get public product failed", "product_id", productID, "error", err)
		return nil, err
	}
	response := dto.PublicProductFromDomain(*product)
	return &response, nil
}

func (h *Handler) BatchGetProducts(ctx context.Context, request dto.BatchGetProductsRequestDTO) (*dto.ProductListResponseDTO, error) {
	products, err := h.useCase.BatchGetProducts(ctx, request.ToUseCase())
	if err != nil {
		h.logger.Error("batch get products failed", "requested_count", len(request.ProductIDs), "error", err)
		return nil, err
	}
	response := dto.ProductListResponseDTO{
		Products: dto.PublicProductsFromDomain(products),
		Total:    int64(len(products)),
	}
	return &response, nil
}

func (h *Handler) ExportSearchProducts(ctx context.Context, request dto.SearchProductExportRequestDTO) (*dto.SearchProductExportResponseDTO, error) {
	result, err := h.useCase.ExportSearchProducts(ctx, request.ToUseCase())
	if err != nil {
		h.logger.Error("search product export failed", "limit", request.Limit, "error", err)
		return nil, err
	}
	response := dto.SearchProductExportFromUseCase(result)
	return &response, nil
}

func (h *Handler) ListCategories(ctx context.Context, request dto.ListCategoriesRequestDTO) (*dto.CategoryListResponseDTO, error) {
	categories, err := h.useCase.ListCategories(ctx, request.ToUseCase())
	if err != nil {
		h.logger.Error("list categories failed", "parent_id", request.ParentID, "error", err)
		return nil, err
	}
	response := dto.CategoryListResponseDTO{Categories: dto.CategoriesFromDomain(categories)}
	return &response, nil
}

func (h *Handler) ListSellerProducts(ctx context.Context, request dto.ListSellerProductsRequestDTO) (*dto.ProductListResponseDTO, error) {
	result, err := h.useCase.ListSellerProducts(ctx, request.ToUseCase())
	if err != nil {
		h.logger.Error("list seller products failed", "seller_id", request.Actor.SellerID, "category_id", request.CategoryID, "error", err)
		return nil, err
	}
	response := dto.SellerProductListFromUseCase(result)
	return &response, nil
}

func (h *Handler) GetSellerProduct(ctx context.Context, request dto.GetSellerProductRequestDTO) (*dto.ProductReadDTO, error) {
	product, err := h.useCase.GetSellerProduct(ctx, request.ToUseCase())
	if err != nil {
		h.logger.Error("get seller product failed", "seller_id", request.Actor.SellerID, "product_id", request.ProductID, "error", err)
		return nil, err
	}
	response := dto.SellerProductFromDomain(*product)
	return &response, nil
}
