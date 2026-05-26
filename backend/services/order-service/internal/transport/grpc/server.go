package ordergrpc

import (
	"errors"
	"log/slog"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/usecase"
	orderv1 "github.com/example/ecommerce-platform/backend/shared/gen/go/ecommerce/order/v1"
)

type Server struct {
	orderv1.UnimplementedOrderServiceServer
	createOrder       usecase.CreateOrderExecutor
	getOrder          usecase.GetOrderExecutor
	listOrders        usecase.ListOrdersExecutor
	updateFulfillment usecase.UpdateFulfillmentExecutor
	logger            *slog.Logger
	now               func() time.Time
}

func NewServer(
	createOrder usecase.CreateOrderExecutor,
	getOrder usecase.GetOrderExecutor,
	listOrders usecase.ListOrdersExecutor,
	updateFulfillment usecase.UpdateFulfillmentExecutor,
	logger *slog.Logger,
) (*Server, error) {
	if createOrder == nil || getOrder == nil || listOrders == nil || updateFulfillment == nil {
		return nil, errors.New("all order RPC usecases are required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{
		createOrder:       createOrder,
		getOrder:          getOrder,
		listOrders:        listOrders,
		updateFulfillment: updateFulfillment,
		logger:            logger,
		now:               func() time.Time { return time.Now().UTC() },
	}, nil
}
