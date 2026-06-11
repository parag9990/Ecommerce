package domain

import (
	"fmt"
	"strings"
	"time"
)

const (
	maxCarrierLength         = 128
	maxTrackingNumberLength  = 128
	maxFulfillmentNoteLength = 512
)

// FulfillmentTransition contains an order-level shipment/status update. The
// service accepts it only for orders representable by one shipment; split
// seller fulfillment belongs to the seller-view workflow.
type FulfillmentTransition struct {
	OrderID          string
	ShipmentID       string
	TargetStatus     OrderStatus
	TrackingNumber   string
	Carrier          string
	Note             string
	ActorID          string
	ActorType        OrderStatusActorType
	RequiredSellerID string
	OccurredAt       time.Time
	History          OrderStatusHistoryEntry
}

func IsFulfillmentTargetStatus(status OrderStatus) bool {
	switch status {
	case OrderStatusPacked, OrderStatusShipped, OrderStatusDelivered:
		return true
	default:
		return false
	}
}

func (t FulfillmentTransition) Validate(current OrderStatus) error {
	if err := validateRequiredID("order_id", t.OrderID); err != nil {
		return err
	}
	if err := validateRequiredID("shipment_id", t.ShipmentID); err != nil {
		return err
	}
	if !IsFulfillmentTargetStatus(t.TargetStatus) {
		return fmt.Errorf("%w: fulfillment target status %q", ErrInvalidRequest, t.TargetStatus)
	}
	if err := ValidateOrderStatusTransition(current, t.TargetStatus); err != nil {
		return err
	}
	if !IsKnownOrderStatusActorType(t.ActorType) || strings.TrimSpace(t.ActorID) == "" {
		return ErrForbidden
	}
	if strings.TrimSpace(t.RequiredSellerID) != "" && t.ActorType != OrderStatusActorSeller {
		return ErrForbidden
	}
	if len(strings.TrimSpace(t.Carrier)) > maxCarrierLength ||
		len(strings.TrimSpace(t.TrackingNumber)) > maxTrackingNumberLength ||
		len(strings.TrimSpace(t.Note)) > maxFulfillmentNoteLength {
		return ErrInvalidRequest
	}
	if t.TargetStatus == OrderStatusShipped &&
		(strings.TrimSpace(t.Carrier) == "" || strings.TrimSpace(t.TrackingNumber) == "") {
		return ErrTrackingRequired
	}
	if t.OccurredAt.IsZero() {
		return ErrInvalidRequest
	}
	if err := validateTransitionHistory(t.History, t.OrderID, current, t.TargetStatus); err != nil {
		return err
	}
	return nil
}

func validateTransitionHistory(history OrderStatusHistoryEntry, orderID string, from OrderStatus, to OrderStatus) error {
	if strings.TrimSpace(history.ID) == "" || history.OrderID != strings.TrimSpace(orderID) ||
		history.FromStatus == nil || *history.FromStatus != from || history.ToStatus != to ||
		strings.TrimSpace(history.Reason) == "" || !IsKnownOrderStatusActorType(history.ActorType) ||
		history.CreatedAt.IsZero() {
		return ErrInvalidOrder
	}
	return nil
}
