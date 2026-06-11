package domain

import (
	"fmt"
	"time"
)

type OrderStatus string

const (
	OrderStatusCreated        OrderStatus = "created"
	OrderStatusPendingPayment OrderStatus = "pending_payment"
	OrderStatusPaid           OrderStatus = "paid"
	OrderStatusPaymentFailed  OrderStatus = "payment_failed"
	OrderStatusPacked         OrderStatus = "packed"
	OrderStatusShipped        OrderStatus = "shipped"
	OrderStatusDelivered      OrderStatus = "delivered"
	OrderStatusCancelled      OrderStatus = "cancelled"
	OrderStatusRefunded       OrderStatus = "refunded"
)

type OrderStatusActorType string

const (
	OrderStatusActorSystem         OrderStatusActorType = "system"
	OrderStatusActorBuyer          OrderStatusActorType = "buyer"
	OrderStatusActorSeller         OrderStatusActorType = "seller"
	OrderStatusActorAdmin          OrderStatusActorType = "admin"
	OrderStatusActorPaymentService OrderStatusActorType = "payment_service"
	OrderStatusActorLogistics      OrderStatusActorType = "logistics"
)

// OrderStatusHistoryEntry is the persistence-independent audit shape for a
// completed status change. Persistence is introduced with the order schema.
type OrderStatusHistoryEntry struct {
	ID         string
	OrderID    string
	FromStatus *OrderStatus
	ToStatus   OrderStatus
	Reason     string
	ActorType  OrderStatusActorType
	ActorID    string
	Metadata   map[string]any
	CreatedAt  time.Time
}

func (s OrderStatus) String() string {
	return string(s)
}

func ParseOrderStatus(value string) (OrderStatus, error) {
	status := OrderStatus(value)
	if !IsKnownOrderStatus(status) {
		return "", fmt.Errorf("%w: %q", ErrUnknownOrderStatus, value)
	}

	return status, nil
}

func IsKnownOrderStatus(status OrderStatus) bool {
	switch status {
	case OrderStatusCreated,
		OrderStatusPendingPayment,
		OrderStatusPaid,
		OrderStatusPaymentFailed,
		OrderStatusPacked,
		OrderStatusShipped,
		OrderStatusDelivered,
		OrderStatusCancelled,
		OrderStatusRefunded:
		return true
	default:
		return false
	}
}

// AllowedOrderStatusTransitions returns a new slice so callers cannot mutate
// the lifecycle definition shared by other requests.
func AllowedOrderStatusTransitions(from OrderStatus) []OrderStatus {
	switch from {
	case OrderStatusCreated:
		return []OrderStatus{OrderStatusPendingPayment, OrderStatusPaymentFailed, OrderStatusCancelled}
	case OrderStatusPendingPayment:
		return []OrderStatus{OrderStatusPaid, OrderStatusPaymentFailed, OrderStatusCancelled}
	case OrderStatusPaymentFailed:
		return []OrderStatus{}
	case OrderStatusPaid:
		return []OrderStatus{OrderStatusPacked, OrderStatusCancelled, OrderStatusRefunded}
	case OrderStatusPacked:
		return []OrderStatus{OrderStatusShipped, OrderStatusCancelled}
	case OrderStatusShipped:
		return []OrderStatus{OrderStatusDelivered}
	case OrderStatusDelivered:
		return []OrderStatus{OrderStatusRefunded}
	case OrderStatusCancelled:
		return []OrderStatus{OrderStatusRefunded}
	case OrderStatusRefunded:
		return []OrderStatus{}
	default:
		return nil
	}
}

func CanTransitionOrderStatus(from OrderStatus, to OrderStatus) bool {
	if !IsKnownOrderStatus(from) || !IsKnownOrderStatus(to) {
		return false
	}

	for _, next := range AllowedOrderStatusTransitions(from) {
		if next == to {
			return true
		}
	}

	return false
}

func ValidateOrderStatusTransition(from OrderStatus, to OrderStatus) error {
	if !IsKnownOrderStatus(from) {
		return fmt.Errorf("%w: from=%q", ErrUnknownOrderStatus, from)
	}
	if !IsKnownOrderStatus(to) {
		return fmt.Errorf("%w: to=%q", ErrUnknownOrderStatus, to)
	}
	if from == to {
		return fmt.Errorf("%w: %s", ErrOrderStatusUnchanged, to)
	}
	if !CanTransitionOrderStatus(from, to) {
		return fmt.Errorf("%w: %s to %s", ErrInvalidOrderStatusTransition, from, to)
	}

	return nil
}

func (a OrderStatusActorType) String() string {
	return string(a)
}

func ParseOrderStatusActorType(value string) (OrderStatusActorType, error) {
	actorType := OrderStatusActorType(value)
	if !IsKnownOrderStatusActorType(actorType) {
		return "", fmt.Errorf("%w: %q", ErrUnknownOrderStatusActorType, value)
	}

	return actorType, nil
}

func IsKnownOrderStatusActorType(actorType OrderStatusActorType) bool {
	switch actorType {
	case OrderStatusActorSystem,
		OrderStatusActorBuyer,
		OrderStatusActorSeller,
		OrderStatusActorAdmin,
		OrderStatusActorPaymentService,
		OrderStatusActorLogistics:
		return true
	default:
		return false
	}
}
