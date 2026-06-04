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
	listSellerOrders  usecase.ListSellerOrdersExecutor
	updateFulfillment usecase.UpdateFulfillmentExecutor
	cancelOrder       usecase.CancelOrderExecutor
	updateSeller      usecase.UpdateSellerFulfillmentExecutor
	logger            *slog.Logger
	now               func() time.Time
}

func NewServer(
	createOrder usecase.CreateOrderExecutor,
	getOrder usecase.GetOrderExecutor,
	listOrders usecase.ListOrdersExecutor,
	listSellerOrders usecase.ListSellerOrdersExecutor,
	updateFulfillment usecase.UpdateFulfillmentExecutor,
	cancelOrder usecase.CancelOrderExecutor,
	updateSeller usecase.UpdateSellerFulfillmentExecutor,
	logger *slog.Logger,
) (*Server, error) {
	if createOrder == nil || getOrder == nil || listOrders == nil || listSellerOrders == nil ||
		updateFulfillment == nil || cancelOrder == nil || updateSeller == nil {
		return nil, errors.New("all order RPC usecases are required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{
		createOrder:       createOrder,
		getOrder:          getOrder,
		listOrders:        listOrders,
		listSellerOrders:  listSellerOrders,
		updateFulfillment: updateFulfillment,
		cancelOrder:       cancelOrder,
		updateSeller:      updateSeller,
		logger:            logger,
		now:               func() time.Time { return time.Now().UTC() },
	}, nil
}
