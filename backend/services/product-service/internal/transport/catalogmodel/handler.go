package catalogmodel

import (
	"context"
	"fmt"
	"log/slog"

	"product-service/internal/transport/dto"
	"product-service/internal/usecase"
)

type Handler struct {
	useCase usecase.CatalogModelUseCase
	logger  *slog.Logger
}

func NewHandler(useCase usecase.CatalogModelUseCase, logger *slog.Logger) (*Handler, error) {
	if useCase == nil {
		return nil, fmt.Errorf("catalog model usecase is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{useCase: useCase, logger: logger}, nil
}

func (h *Handler) ValidateProduct(ctx context.Context, request dto.ValidateProductRequestDTO) (*dto.ValidateProductResponseDTO, error) {
	report, err := h.useCase.ValidateProduct(ctx, request.ToUseCase())
	if err != nil {
		h.logger.Error("validate product model failed", "error", err)
		return nil, err
	}
	return &dto.ValidateProductResponseDTO{Report: dto.ValidationReportFromDomain(report)}, nil
}

func (h *Handler) CheckPublishReadiness(ctx context.Context, request dto.ValidateProductRequestDTO) (*dto.ValidateProductResponseDTO, error) {
	report, err := h.useCase.CheckPublishReadiness(ctx, request.ToUseCase())
	if err != nil {
		h.logger.Error("check product publish readiness failed", "error", err)
		return nil, err
	}
	return &dto.ValidateProductResponseDTO{Report: dto.ValidationReportFromDomain(report)}, nil
}
