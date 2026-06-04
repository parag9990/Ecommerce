package domain

import (
	"fmt"
	"strings"
	"time"
)

type SellerFulfillmentStatus string

const (
	SellerFulfillmentStatusPending            SellerFulfillmentStatus = "pending"
	SellerFulfillmentStatusPartiallyPacked    SellerFulfillmentStatus = "partially_packed"
	SellerFulfillmentStatusPacked             SellerFulfillmentStatus = "packed"
	SellerFulfillmentStatusPartiallyShipped   SellerFulfillmentStatus = "partially_shipped"
	SellerFulfillmentStatusShipped            SellerFulfillmentStatus = "shipped"
	SellerFulfillmentStatusPartiallyDelivered SellerFulfillmentStatus = "partially_delivered"
	SellerFulfillmentStatusDelivered          SellerFulfillmentStatus = "delivered"
)

type ShipmentStatus string

const (
	ShipmentStatusPending   ShipmentStatus = "pending"
	ShipmentStatusPacked    ShipmentStatus = "packed"
	ShipmentStatusShipped   ShipmentStatus = "shipped"
	ShipmentStatusDelivered ShipmentStatus = "delivered"
	ShipmentStatusFailed    ShipmentStatus = "failed"
)

type SellerOrderView struct {
	OrderID                 string
	ParentOrderStatus       OrderStatus
	SellerFulfillmentStatus SellerFulfillmentStatus
	Currency                string
	SellerItemsTotal        int64
	Items                   []SellerOrderItemView
	Shipments               []SellerShipmentView
	CreatedAt               time.Time
}

type SellerOrderItemView struct {
	OrderItemID       string
	OrderID           string
	ProductID         string
	VariantID         string
	SKU               string
	TitleSnapshot     string
	ImageURLSnapshot  string
	Quantity          int32
	Currency          string
	UnitAmount        int64
	TotalAmount       int64
	FulfillmentStatus FulfillmentStatus
}

type SellerShipmentView struct {
	ShipmentID     string
	OrderID        string
	Status         ShipmentStatus
	Carrier        string
	TrackingNumber string
	ShippedAt      *time.Time
	DeliveredAt    *time.Time
}

type ListSellerOrdersFilter struct {
	SellerID          string
	PageSize          int
	Cursor            *OrderCursor
	FulfillmentFilter *SellerFulfillmentStatus
}

type SellerOrderPage struct {
	Orders        []SellerOrderView
	NextCursor    *OrderCursor
	NextPageToken string
}

type SellerFulfillmentTransition struct {
	OrderID        string
	SellerID       string
	ShipmentID     string
	TargetStatus   FulfillmentStatus
	TrackingNumber string
	Carrier        string
	Note           string
	ActorID        string
	OccurredAt     time.Time
	HistoryID      string
}

func (s SellerFulfillmentStatus) String() string {
	return string(s)
}

func IsKnownSellerFulfillmentStatus(status SellerFulfillmentStatus) bool {
	switch status {
	case SellerFulfillmentStatusPending,
		SellerFulfillmentStatusPartiallyPacked,
		SellerFulfillmentStatusPacked,
		SellerFulfillmentStatusPartiallyShipped,
		SellerFulfillmentStatusShipped,
		SellerFulfillmentStatusPartiallyDelivered,
		SellerFulfillmentStatusDelivered:
		return true
	default:
		return false
	}
}

func IsKnownFulfillmentStatus(status FulfillmentStatus) bool {
	switch status {
	case FulfillmentStatusPending,
		FulfillmentStatusPacked,
		FulfillmentStatusShipped,
		FulfillmentStatusDelivered,
		FulfillmentStatusCancelled,
		FulfillmentStatusReturned:
		return true
	default:
		return false
	}
}

func IsKnownShipmentStatus(status ShipmentStatus) bool {
	switch status {
	case ShipmentStatusPending,
		ShipmentStatusPacked,
		ShipmentStatusShipped,
		ShipmentStatusDelivered,
		ShipmentStatusFailed:
		return true
	default:
		return false
	}
}

func DeriveSellerFulfillmentStatus(items []SellerOrderItemView) SellerFulfillmentStatus {
	var active []FulfillmentStatus
	for _, item := range items {
		if isActiveFulfillmentStatus(item.FulfillmentStatus) {
			active = append(active, item.FulfillmentStatus)
		}
	}
	if len(active) == 0 {
		return SellerFulfillmentStatusPending
	}
	switch {
	case allFulfillmentAtLeast(active, FulfillmentStatusDelivered):
		return SellerFulfillmentStatusDelivered
	case anyFulfillmentAtLeast(active, FulfillmentStatusDelivered):
		return SellerFulfillmentStatusPartiallyDelivered
	case allFulfillmentAtLeast(active, FulfillmentStatusShipped):
		return SellerFulfillmentStatusShipped
	case anyFulfillmentAtLeast(active, FulfillmentStatusShipped):
		return SellerFulfillmentStatusPartiallyShipped
	case allFulfillmentAtLeast(active, FulfillmentStatusPacked):
		return SellerFulfillmentStatusPacked
	case anyFulfillmentAtLeast(active, FulfillmentStatusPacked):
		return SellerFulfillmentStatusPartiallyPacked
	default:
		return SellerFulfillmentStatusPending
	}
}

func (t SellerFulfillmentTransition) Validate() error {
	if err := validateRequiredID("order_id", t.OrderID); err != nil {
		return err
	}
	if err := validateRequiredID("seller_id", t.SellerID); err != nil {
		return err
	}
	if err := validateRequiredID("shipment_id", t.ShipmentID); err != nil {
		return err
	}
	if err := validateRequiredID("history_id", t.HistoryID); err != nil {
		return err
	}
	if strings.TrimSpace(t.ActorID) == "" || len(strings.TrimSpace(t.ActorID)) > maxPublicIDLength {
		return ErrInvalidRequest
	}
	if !IsSellerFulfillmentTargetStatus(t.TargetStatus) {
		return fmt.Errorf("%w: fulfillment target status %q", ErrInvalidRequest, t.TargetStatus)
	}
	if len(strings.TrimSpace(t.Carrier)) > maxCarrierLength ||
		len(strings.TrimSpace(t.TrackingNumber)) > maxTrackingNumberLength ||
		len(strings.TrimSpace(t.Note)) > maxFulfillmentNoteLength {
		return ErrInvalidRequest
	}
	if t.TargetStatus == FulfillmentStatusShipped &&
		(strings.TrimSpace(t.Carrier) == "" || strings.TrimSpace(t.TrackingNumber) == "") {
		return ErrTrackingRequired
	}
	if t.OccurredAt.IsZero() {
		return ErrInvalidRequest
	}
	return nil
}

func IsSellerFulfillmentTargetStatus(status FulfillmentStatus) bool {
	switch status {
	case FulfillmentStatusPacked, FulfillmentStatusShipped, FulfillmentStatusDelivered:
		return true
	default:
		return false
	}
}

func ValidateSellerItemTransition(from FulfillmentStatus, to FulfillmentStatus) error {
	if !IsKnownFulfillmentStatus(from) || !IsSellerFulfillmentTargetStatus(to) {
		return ErrInvalidFulfillmentTransition
	}
	next := map[FulfillmentStatus]FulfillmentStatus{
		FulfillmentStatusPending: FulfillmentStatusPacked,
		FulfillmentStatusPacked:  FulfillmentStatusShipped,
		FulfillmentStatusShipped: FulfillmentStatusDelivered,
	}
	if next[from] != to {
		return ErrInvalidFulfillmentTransition
	}
	return nil
}

func CanSellerFulfillParentStatus(status OrderStatus) bool {
	switch status {
	case OrderStatusPaid, OrderStatusPacked, OrderStatusShipped:
		return true
	default:
		return false
	}
}

func AggregateParentFulfillmentStatus(current OrderStatus, itemStatuses []FulfillmentStatus) (OrderStatus, bool) {
	active := make([]FulfillmentStatus, 0, len(itemStatuses))
	for _, status := range itemStatuses {
		if isActiveFulfillmentStatus(status) {
			active = append(active, status)
		}
	}
	if len(active) == 0 {
		return current, false
	}
	switch {
	case current == OrderStatusShipped && allFulfillmentAtLeast(active, FulfillmentStatusDelivered):
		return OrderStatusDelivered, true
	case current == OrderStatusPacked && allFulfillmentAtLeast(active, FulfillmentStatusShipped):
		return OrderStatusShipped, true
	case current == OrderStatusPaid && allFulfillmentAtLeast(active, FulfillmentStatusPacked):
		return OrderStatusPacked, true
	default:
		return current, false
	}
}

func isActiveFulfillmentStatus(status FulfillmentStatus) bool {
	return status != FulfillmentStatusCancelled && status != FulfillmentStatusReturned
}

func anyFulfillmentAtLeast(statuses []FulfillmentStatus, minimum FulfillmentStatus) bool {
	for _, status := range statuses {
		if fulfillmentRank(status) >= fulfillmentRank(minimum) {
			return true
		}
	}
	return false
}

func allFulfillmentAtLeast(statuses []FulfillmentStatus, minimum FulfillmentStatus) bool {
	for _, status := range statuses {
		if fulfillmentRank(status) < fulfillmentRank(minimum) {
			return false
		}
	}
	return len(statuses) > 0
}

func fulfillmentRank(status FulfillmentStatus) int {
	switch status {
	case FulfillmentStatusDelivered:
		return 3
	case FulfillmentStatusShipped:
		return 2
	case FulfillmentStatusPacked:
		return 1
	default:
		return 0
	}
}
