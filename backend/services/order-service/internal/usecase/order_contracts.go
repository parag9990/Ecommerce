package usecase

import (
	"context"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
)

type CreateOrderCommand struct {
	UserID          string
	SessionID       string
	TraceID         string
	CartID          string
	IdempotencyKey  string
	ShippingAddress domain.AddressSnapshot
	CouponCode      string
}

type PaymentActionView struct {
	PaymentID         string
	ClientActionToken string
	ExpiresAt         time.Time
}

type CreateOrderResult struct {
	Order         domain.Order
	PaymentAction *PaymentActionView
}

type GetOrderQuery struct {
	ActorID string
	Roles   []string
	OrderID string
}

type ListOrdersQuery struct {
	UserID       string
	PageSize     int
	PageToken    string
	StatusFilter *domain.OrderStatus
}

type UpdateFulfillmentCommand struct {
	ActorID        string
	Roles          []string
	OrderID        string
	TraceID        string
	TargetStatus   domain.OrderStatus
	TrackingNumber string
	Carrier        string
	Note           string
	OccurredAt     time.Time
}

type OrderCursor = domain.OrderCursor
type ListOrdersFilter = domain.ListOrdersFilter
type OrderPage = domain.OrderPage

type OrderReadRepository interface {
	GetOrder(ctx context.Context, orderID string) (domain.Order, error)
	ListOrders(ctx context.Context, filter domain.ListOrdersFilter) (domain.OrderPage, error)
}

type FulfillmentRepository interface {
	UpdateFulfillment(ctx context.Context, transition domain.FulfillmentTransition, event *domain.OutboxEvent) (domain.Order, error)
}

type CreateFromCartExecutor interface {
	Execute(ctx context.Context, command CreateOrderFromCartCommand) (*CreateOrderFromCartResult, error)
}

type PaymentInitiationExecutor interface {
	Execute(ctx context.Context, command InitiateOrderPaymentCommand) (*InitiateOrderPaymentResult, error)
}

type CreateOrderExecutor interface {
	Execute(ctx context.Context, command CreateOrderCommand) (*CreateOrderResult, error)
}

type GetOrderExecutor interface {
	Execute(ctx context.Context, query GetOrderQuery) (domain.Order, error)
}

type ListOrdersExecutor interface {
	Execute(ctx context.Context, query ListOrdersQuery) (OrderPage, error)
}

type UpdateFulfillmentExecutor interface {
	Execute(ctx context.Context, command UpdateFulfillmentCommand) (domain.Order, error)
}
