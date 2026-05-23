package catalogschema

import (
	"context"
	"fmt"
	"log/slog"

	"product-service/internal/transport/dto"
	"product-service/internal/usecase"
)

type Handler struct {
	useCase usecase.CollectionSchemaUseCase
	logger  *slog.Logger
}

func NewHandler(useCase usecase.CollectionSchemaUseCase, logger *slog.Logger) (*Handler, error) {
	if useCase == nil {
		return nil, fmt.Errorf("collection schema usecase is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{useCase: useCase, logger: logger}, nil
}

func (h *Handler) EnsureCollections(ctx context.Context) (*dto.CollectionSetupResponseDTO, error) {
	result, err := h.useCase.EnsureProductCollections(ctx)
	if err != nil {
		h.logger.Error("ensure product collections failed", "error", err)
		return nil, err
	}
	response := dto.CollectionSetupResponseFromUseCase(result)
	return &response, nil
}

func (h *Handler) DescribeCollections(ctx context.Context) (*dto.CollectionDescriptionResponseDTO, error) {
	descriptions, err := h.useCase.DescribeProductCollections(ctx)
	if err != nil {
		h.logger.Error("describe product collections failed", "error", err)
		return nil, err
	}
	response := dto.CollectionDescriptionResponseFromUseCase(descriptions)
	return &response, nil
}
