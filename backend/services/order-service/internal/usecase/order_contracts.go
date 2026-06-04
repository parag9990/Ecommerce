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

type ListSellerOrdersQuery struct {
	SellerID          string
	ActorUserID       string
	PageSize          int
	PageToken         string
	FulfillmentFilter *domain.SellerFulfillmentStatus
}

type UpdateFulfillmentCommand struct {
	ActorID        string
	SellerID       string
	Roles          []string
	OrderID        string
	TraceID        string
	TargetStatus   domain.OrderStatus
	TrackingNumber string
	Carrier        string
	Note           string
	OccurredAt     time.Time
}

type CancelOrderCommand struct {
	ActorID    string
	Roles      []string
	OrderID    string
	TraceID    string
	ReasonCode string
	OccurredAt time.Time
}

type UpdateSellerFulfillmentCommand struct {
	ActorUserID    string
	SellerID       string
	Roles          []string
	OrderID        string
	TargetStatus   domain.FulfillmentStatus
	TrackingNumber string
	Carrier        string
	Note           string
	RequestID      string
	OccurredAt     time.Time
}

type OrderCursor = domain.OrderCursor
type ListOrdersFilter = domain.ListOrdersFilter
type OrderPage = domain.OrderPage
type SellerOrderPage = domain.SellerOrderPage

type OrderReadRepository interface {
	GetOrder(ctx context.Context, orderID string) (domain.Order, error)
	ListOrders(ctx context.Context, filter domain.ListOrdersFilter) (domain.OrderPage, error)
}

type SellerOrderRepository interface {
	ListSellerOrders(ctx context.Context, filter domain.ListSellerOrdersFilter) (domain.SellerOrderPage, error)
	UpdateSellerFulfillment(ctx context.Context, transition domain.SellerFulfillmentTransition, event *domain.OutboxEvent) (domain.SellerOrderView, error)
}

type FulfillmentRepository interface {
	UpdateFulfillment(ctx context.Context, transition domain.FulfillmentTransition, event *domain.OutboxEvent) (domain.Order, error)
}

type OrderCancellationRepository interface {
	CancelOrder(ctx context.Context, orderID string, from domain.OrderStatus, history domain.OrderStatusHistoryEntry, event domain.OutboxEvent) (domain.Order, error)
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

type ListSellerOrdersExecutor interface {
	Execute(ctx context.Context, query ListSellerOrdersQuery) (SellerOrderPage, error)
}

type UpdateFulfillmentExecutor interface {
	Execute(ctx context.Context, command UpdateFulfillmentCommand) (domain.Order, error)
}

type CancelOrderExecutor interface {
	Execute(ctx context.Context, command CancelOrderCommand) (domain.Order, error)
}

type UpdateSellerFulfillmentExecutor interface {
	Execute(ctx context.Context, command UpdateSellerFulfillmentCommand) (domain.SellerOrderView, error)
}
