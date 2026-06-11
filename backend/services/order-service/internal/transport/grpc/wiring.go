package ordergrpc

import (
	"errors"
	"log/slog"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/usecase"
	"google.golang.org/grpc"
)

type OrderApplicationRepository interface {
	usecase.OrderReadRepository
	usecase.FulfillmentRepository
	usecase.SellerOrderRepository
	usecase.OrderCancellationRepository
}

type ApplicationDependencies struct {
	Checkout            usecase.CreateFromCartExecutor
	Payment             usecase.PaymentInitiationExecutor
	Orders              OrderApplicationRepository
	Idempotency         usecase.IdempotencyRepository
	CreateOrderConfig   usecase.CreateOrderConfig
	IDs                 usecase.IDGenerator
	PageTokenSigningKey []byte
	TrustedCallerToken  string
	Logger              *slog.Logger
}

// NewApplicationServer wires the Task 5 RPC application usecases to the gRPC
// boundary. Cart, product, and payment adapters remain injected ports owned by
// their existing checkout/payment workflows.
func NewApplicationServer(dependencies ApplicationDependencies) (*grpc.Server, error) {
	if dependencies.Orders == nil {
		return nil, errors.New("order application repository is required")
	}
	create, err := usecase.NewCreateOrderUsecase(
		dependencies.Checkout,
		dependencies.Payment,
		dependencies.Orders,
		dependencies.Idempotency,
		dependencies.CreateOrderConfig,
		dependencies.Logger,
	)
	if err != nil {
		return nil, err
	}
	get, err := usecase.NewGetOrderUsecase(dependencies.Orders)
	if err != nil {
		return nil, err
	}
	list, err := usecase.NewListOrdersUsecase(dependencies.Orders, dependencies.PageTokenSigningKey)
	if err != nil {
		return nil, err
	}
	listSeller, err := usecase.NewListSellerOrdersUsecase(
		dependencies.Orders,
		dependencies.PageTokenSigningKey,
		dependencies.Logger,
	)
	if err != nil {
		return nil, err
	}
	update, err := usecase.NewUpdateFulfillmentUsecase(
		dependencies.Orders,
		dependencies.Orders,
		dependencies.IDs,
		dependencies.Logger,
	)
	if err != nil {
		return nil, err
	}
	cancel, err := usecase.NewCancelOrderUsecase(
		dependencies.Orders,
		dependencies.Orders,
		dependencies.IDs,
		dependencies.Logger,
	)
	if err != nil {
		return nil, err
	}
	updateSeller, err := usecase.NewUpdateSellerFulfillmentUsecase(
		dependencies.Orders,
		dependencies.Orders,
		dependencies.IDs,
		dependencies.Logger,
	)
	if err != nil {
		return nil, err
	}
	handler, err := NewServer(create, get, list, listSeller, update, cancel, updateSeller, dependencies.Logger)
	if err != nil {
		return nil, err
	}
	return NewGRPCServer(handler, dependencies.IDs, RuntimeConfig{
		TrustedCallerToken: dependencies.TrustedCallerToken,
	}, dependencies.Logger)
}
