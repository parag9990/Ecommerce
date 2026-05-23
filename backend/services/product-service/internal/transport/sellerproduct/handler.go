package sellerproduct

import (
	"context"
	"fmt"
	"log/slog"

	"product-service/internal/transport/dto"
	"product-service/internal/usecase"
)

type Handler struct {
	useCase usecase.SellerProductUseCase
	logger  *slog.Logger
}

func NewHandler(useCase usecase.SellerProductUseCase, logger *slog.Logger) (*Handler, error) {
	if useCase == nil {
		return nil, fmt.Errorf("seller product usecase is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{useCase: useCase, logger: logger}, nil
}

func (h *Handler) CreateProduct(ctx context.Context, request dto.CreateSellerProductRequestDTO) (*dto.ProductDTO, error) {
	product, err := h.useCase.CreateProduct(ctx, request.ToUseCase())
	if err != nil {
		h.logger.Error("create seller product failed", "error", err)
		return nil, err
	}
	response := dto.ProductFromDomain(*product)
	return &response, nil
}

func (h *Handler) UpdateProduct(ctx context.Context, request dto.UpdateSellerProductRequestDTO) (*dto.ProductDTO, error) {
	product, err := h.useCase.UpdateProduct(ctx, request.ToUseCase())
	if err != nil {
		h.logger.Error("update seller product failed", "product_id", request.ProductID, "error", err)
		return nil, err
	}
	response := dto.ProductFromDomain(*product)
	return &response, nil
}

func (h *Handler) PublishProduct(ctx context.Context, request dto.ProductLifecycleRequestDTO) (*dto.ProductDTO, error) {
	product, err := h.useCase.PublishProduct(ctx, request.ToUseCase())
	if err != nil {
		h.logger.Error("publish seller product failed", "product_id", request.ProductID, "error", err)
		return nil, err
	}
	response := dto.ProductFromDomain(*product)
	return &response, nil
}

func (h *Handler) UnpublishProduct(ctx context.Context, request dto.ProductLifecycleRequestDTO) (*dto.ProductDTO, error) {
	product, err := h.useCase.UnpublishProduct(ctx, request.ToUseCase())
	if err != nil {
		h.logger.Error("unpublish seller product failed", "product_id", request.ProductID, "error", err)
		return nil, err
	}
	response := dto.ProductFromDomain(*product)
	return &response, nil
}
