package dto

import (
	"strings"
	"time"

	"product-service/internal/usecase"
)

type InventoryReservationRequestDTO struct {
	OrderID        string                        `json:"order_id"`
	Items          []InventoryReservationItemDTO `json:"items"`
	TTLSeconds     int                           `json:"ttl_seconds,omitempty"`
	IdempotencyKey string                        `json:"idempotency_key,omitempty"`
}

type InventoryReservationItemDTO struct {
	ProductID string `json:"product_id"`
	VariantID string `json:"variant_id"`
	Quantity  int64  `json:"quantity"`
}

type InventoryReservationResponseDTO struct {
	ReservationID string    `json:"reservation_id"`
	ExpiresAt     time.Time `json:"expires_at"`
}

type InventoryReservationActionRequestDTO struct {
	ReservationID string `json:"reservation_id"`
	Reason        string `json:"reason,omitempty"`
}

type SuccessResponseDTO struct {
	Success bool `json:"success"`
}

func (d InventoryReservationRequestDTO) ToUseCase() usecase.ReserveInventoryRequest {
	items := make([]usecase.ReserveInventoryItem, 0, len(d.Items))
	for _, item := range d.Items {
		items = append(items, usecase.ReserveInventoryItem{
			ProductID: strings.TrimSpace(item.ProductID),
			VariantID: strings.TrimSpace(item.VariantID),
			Quantity:  item.Quantity,
		})
	}
	return usecase.ReserveInventoryRequest{
		OrderID:        strings.TrimSpace(d.OrderID),
		Items:          items,
		TTLSeconds:     d.TTLSeconds,
		IdempotencyKey: strings.TrimSpace(d.IdempotencyKey),
	}
}

func (d InventoryReservationActionRequestDTO) ToUseCase() usecase.InventoryReservationActionRequest {
	return usecase.InventoryReservationActionRequest{
		ReservationID: strings.TrimSpace(d.ReservationID),
		Reason:        strings.TrimSpace(d.Reason),
	}
}

func InventoryReservationResponseFromUseCase(result usecase.InventoryReservationResult) InventoryReservationResponseDTO {
	return InventoryReservationResponseDTO{
		ReservationID: result.ReservationID,
		ExpiresAt:     result.ExpiresAt,
	}
}
