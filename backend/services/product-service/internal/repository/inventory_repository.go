package repository

import (
	"context"
	"time"

	"product-service/internal/domain"
)

type InventoryStockMutation struct {
	ProductID                 string
	VariantID                 string
	SKU                       string
	SellerID                  string
	StockQuantity             int64
	ReservedQuantity          int64
	SafetyStock               int64
	AvailableQuantity         int64
	PreviousAvailableQuantity int64
	PreviousInStock           bool
	HasPreviousAvailability   bool
}

func (m InventoryStockMutation) InStockChanged() bool {
	if !m.HasPreviousAvailability {
		return true
	}
	return m.PreviousInStock != (m.AvailableQuantity > 0)
}

type InventoryStockRepository interface {
	ReserveVariant(ctx context.Context, productID string, variantID string, quantity int64, now time.Time) (InventoryStockMutation, error)
	ReleaseVariant(ctx context.Context, productID string, variantID string, quantity int64, now time.Time) (InventoryStockMutation, error)
	CommitVariant(ctx context.Context, productID string, variantID string, quantity int64, now time.Time) (InventoryStockMutation, error)
}

type InventoryReservationRepository interface {
	FindReservationByID(ctx context.Context, reservationID string) (*domain.InventoryReservation, error)
	FindReservationByOrderID(ctx context.Context, orderID string) (*domain.InventoryReservation, error)
	CreateInventoryReservation(ctx context.Context, reservation *domain.InventoryReservation) error
	MarkReservationReleased(ctx context.Context, reservationID string, reason string, releasedAt time.Time) error
	MarkReservationCommitted(ctx context.Context, reservationID string, committedAt time.Time) error
	MarkReservationExpired(ctx context.Context, reservationID string, reason string, expiredAt time.Time) error
	ListExpiredReservations(ctx context.Context, now time.Time, limit int64) ([]domain.InventoryReservation, error)
}

type InventorySnapshotRepository interface {
	CreateInventorySnapshot(ctx context.Context, snapshot domain.InventorySnapshot) error
}
