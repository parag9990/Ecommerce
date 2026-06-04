package inventory

import (
	"context"
	"fmt"
	"log/slog"

	"product-service/internal/transport/dto"
	"product-service/internal/usecase"
)

type Handler struct {
	useCase usecase.InventoryUseCase
	logger  *slog.Logger
}

func NewHandler(useCase usecase.InventoryUseCase, logger *slog.Logger) (*Handler, error) {
	if useCase == nil {
		return nil, fmt.Errorf("inventory usecase is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{useCase: useCase, logger: logger}, nil
}

func (h *Handler) ReserveInventory(
	ctx context.Context,
	request dto.InventoryReservationRequestDTO,
) (*dto.InventoryReservationResponseDTO, error) {
	result, err := h.useCase.ReserveInventory(ctx, request.ToUseCase())
	if err != nil {
		h.logger.Error("reserve inventory failed", "order_id", request.OrderID, "error", err)
		return nil, err
	}
	response := dto.InventoryReservationResponseFromUseCase(*result)
	return &response, nil
}

func (h *Handler) ReleaseInventory(
	ctx context.Context,
	request dto.InventoryReservationActionRequestDTO,
) (*dto.SuccessResponseDTO, error) {
	if err := h.useCase.ReleaseInventory(ctx, request.ToUseCase()); err != nil {
		h.logger.Error("release inventory failed", "reservation_id", request.ReservationID, "error", err)
		return nil, err
	}
	return &dto.SuccessResponseDTO{Success: true}, nil
}

func (h *Handler) CommitInventory(
	ctx context.Context,
	request dto.InventoryReservationActionRequestDTO,
) (*dto.SuccessResponseDTO, error) {
	if err := h.useCase.CommitInventory(ctx, request.ToUseCase()); err != nil {
		h.logger.Error("commit inventory failed", "reservation_id", request.ReservationID, "error", err)
		return nil, err
	}
	return &dto.SuccessResponseDTO{Success: true}, nil
}
