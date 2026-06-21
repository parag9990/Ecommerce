package usecase

import (
	"context"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
)

type GetCartRequest struct {
	UserID string
	CartID string
}

type ProductLookupItem struct {
	ProductID string
	VariantID string
}

type BatchGetProductsRequest struct {
	Items []ProductLookupItem
}

type BatchGetProductsResponse struct {
	Items []domain.ProductSnapshot
}

type ReserveInventoryItem struct {
	ProductID string
	VariantID string
	Quantity  int32
}

type ReserveInventoryRequest struct {
	OrderID        string
	UserID         string
	CartID         string
	IdempotencyKey string
	TTL            time.Duration
	Items          []ReserveInventoryItem
}

type ReserveInventoryResponse struct {
	ReservationID string
	ExpiresAt     time.Time
}

type ReleaseInventoryRequest struct {
	ReservationID  string
	IdempotencyKey string
	Reason         string
}

type CommitInventoryRequest struct {
	ReservationID  string
	IdempotencyKey string
}

type CreatePaymentIntentRequest struct {
	OrderID        string
	UserID         string
	Amount         int64
	Currency       string
	IdempotencyKey string
	ReturnURL      string
}

type CreatePaymentIntentResponse struct {
	PaymentID         string
	Status            string
	Provider          string
	ProviderIntentRef string
	ClientActionToken string
	ExpiresAt         time.Time
}

// CartClient and ProductClient are downstream ports. Concrete gRPC adapters
// attach when the shared service contracts and Task 5 transport wiring exist.
type CartClient interface {
	GetCart(ctx context.Context, req GetCartRequest) (*domain.CartSnapshot, error)
}

type ProductClient interface {
	BatchGetProducts(ctx context.Context, req BatchGetProductsRequest) (*BatchGetProductsResponse, error)
	ReserveInventory(ctx context.Context, req ReserveInventoryRequest) (*ReserveInventoryResponse, error)
	ReleaseInventory(ctx context.Context, req ReleaseInventoryRequest) error
}

type InventoryClient interface {
	CommitInventory(ctx context.Context, req CommitInventoryRequest) error
	ReleaseInventory(ctx context.Context, req ReleaseInventoryRequest) error
}

type InventoryReleaser interface {
	ReleaseInventory(ctx context.Context, req ReleaseInventoryRequest) error
}

type PaymentClient interface {
	// Implementations must wrap domain.ErrPaymentIntentPendingResolution when
	// a timeout or transport failure leaves intent creation outcome unknown.
	CreatePaymentIntent(ctx context.Context, req CreatePaymentIntentRequest) (*CreatePaymentIntentResponse, error)
}

type OrderRepository interface {
	CreateOrderAndAttachClaim(ctx context.Context, order domain.Order, claim domain.IdempotencyClaim, event domain.OutboxEvent) error
}

type IdempotencyRepository interface {
	Claim(ctx context.Context, userID string, key string, requestHash string, expiresAt time.Time) (domain.IdempotencyClaim, error)
	Complete(ctx context.Context, userID string, key string, orderID string) error
	MarkFailed(ctx context.Context, userID string, key string) error
}

type PaymentOrderRepository interface {
	GetForPayment(ctx context.Context, orderID string) (domain.Order, error)
	AttachPaymentIntent(ctx context.Context, orderID string, paymentID string, history domain.OrderStatusHistoryEntry) (bool, error)
	MarkPaymentFailedFromCreated(ctx context.Context, orderID string, history domain.OrderStatusHistoryEntry) (bool, error)
	TransitionPaymentResult(ctx context.Context, orderID string, paymentID string, from domain.OrderStatus, to domain.OrderStatus, history domain.OrderStatusHistoryEntry, event *domain.OutboxEvent) (bool, error)
}

type IDGenerator interface {
	NewID(prefix string) (string, error)
}

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now().UTC()
}
