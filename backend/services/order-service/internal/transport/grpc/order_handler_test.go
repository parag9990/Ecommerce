package ordergrpc

import (
	"context"
	"io"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/authctx"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/identifier"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/usecase"
	orderv1 "github.com/example/ecommerce-platform/backend/shared/gen/go/ecommerce/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestCreateOrderMapsAuthenticatedBuyerAndSnapshotResponse(t *testing.T) {
	create := &handlerCreateExecutor{result: &usecase.CreateOrderResult{
		Order: domain.Order{
			OrderID: "ord_1", UserID: "user_1", Status: domain.OrderStatusPendingPayment,
			Currency: "INR", TotalAmount: 19900,
		},
		PaymentAction: &usecase.PaymentActionView{PaymentID: "pay_1", ClientActionToken: "action"},
	}}
	handler := newTestHandler(t, create)
	ctx := authctx.WithActor(context.Background(), authctx.Actor{UserID: "user_1", Roles: []string{"buyer"}, SessionID: "session_1"})
	response, err := handler.CreateOrder(ctx, &orderv1.CreateOrderRequest{
		CartId: "cart_1", IdempotencyKey: "attempt_1",
		ShippingAddress: &orderv1.AddressSnapshot{
			RecipientName: "Buyer", Line1: "Street", City: "Delhi", PostalCode: "110001", CountryCode: "IN",
		},
	})
	if err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}
	if create.command.UserID != "user_1" || create.command.SessionID != "session_1" ||
		create.command.ShippingAddress.Country != "IN" || response.GetOrder().GetOrderId() != "ord_1" ||
		response.GetOrder().GetStatus() != orderv1.OrderStatus_ORDER_STATUS_PENDING_PAYMENT ||
		response.GetPaymentAction().GetPaymentId() != "pay_1" {
		t.Fatalf("command/response = %+v/%+v, want mapped checkout", create.command, response)
	}
}

func TestCreateOrderMapsIdempotencyOutcomes(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code codes.Code
	}{
		{name: "payload conflict", err: domain.ErrIdempotencyConflict, code: codes.AlreadyExists},
		{name: "already processing", err: domain.ErrCheckoutInProgress, code: codes.Aborted},
		{name: "terminal failure", err: domain.ErrCheckoutFailed, code: codes.FailedPrecondition},
	}
	ctx := authctx.WithActor(context.Background(), authctx.Actor{UserID: "user_1", Roles: []string{"buyer"}})
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := newTestHandler(t, &handlerCreateExecutor{err: test.err})
			_, err := handler.CreateOrder(ctx, validCreateRequest())
			if status.Code(err) != test.code {
				t.Fatalf("CreateOrder() code = %v, want %v", status.Code(err), test.code)
			}
		})
	}
}

func TestCreateOrderRejectsMissingOrOversizedIdempotencyKey(t *testing.T) {
	handler := newTestHandler(t, &handlerCreateExecutor{})
	ctx := authctx.WithActor(context.Background(), authctx.Actor{UserID: "user_1", Roles: []string{"buyer"}})
	for _, key := range []string{"", strings.Repeat("x", 129)} {
		request := validCreateRequest()
		request.IdempotencyKey = key
		_, err := handler.CreateOrder(ctx, request)
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("CreateOrder(key length=%d) code = %v, want InvalidArgument", len(key), status.Code(err))
		}
	}
}

func TestUpdateFulfillmentRejectsBuyerAndInvalidStatus(t *testing.T) {
	handler := newTestHandler(t, &handlerCreateExecutor{})
	buyer := authctx.WithActor(context.Background(), authctx.Actor{UserID: "user_1", Roles: []string{"buyer"}})
	_, err := handler.UpdateFulfillment(buyer, &orderv1.UpdateFulfillmentRequest{
		OrderId: "ord_1", TargetStatus: orderv1.OrderStatus_ORDER_STATUS_PACKED,
	})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("UpdateFulfillment(buyer) code = %v, want PermissionDenied", status.Code(err))
	}
	admin := authctx.WithActor(context.Background(), authctx.Actor{UserID: "admin_1", Roles: []string{"admin"}})
	_, err = handler.UpdateFulfillment(admin, &orderv1.UpdateFulfillmentRequest{
		OrderId: "ord_1", TargetStatus: orderv1.OrderStatus_ORDER_STATUS_PAID,
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("UpdateFulfillment(invalid target) code = %v, want InvalidArgument", status.Code(err))
	}
}

func TestListSellerOrdersUsesTrustedSellerMetadata(t *testing.T) {
	sellerList := &handlerSellerListExecutor{page: domain.SellerOrderPage{
		Orders: []domain.SellerOrderView{{
			OrderID: "ord_1", ParentOrderStatus: domain.OrderStatusPaid,
			SellerFulfillmentStatus: domain.SellerFulfillmentStatusPacked,
			Currency:                "INR", SellerItemsTotal: 1500,
			Items: []domain.SellerOrderItemView{{
				OrderItemID: "oi_1", OrderID: "ord_1", ProductID: "product_1",
				TitleSnapshot: "Snapshot", Quantity: 1, Currency: "INR", UnitAmount: 1500,
				TotalAmount: 1500, FulfillmentStatus: domain.FulfillmentStatusPacked,
			}},
		}},
	}}
	handler := newTestHandlerWithSeller(t, &handlerCreateExecutor{}, sellerList, &handlerSellerFulfillmentExecutor{})
	ctx := authctx.WithActor(context.Background(), authctx.Actor{
		UserID: "user_1", SellerID: "seller_1", Roles: []string{"seller_order_manager"},
	})
	response, err := handler.ListSellerOrders(ctx, &orderv1.ListSellerOrdersRequest{
		PageSize:          10,
		FulfillmentFilter: orderv1.SellerFulfillmentStatus_SELLER_FULFILLMENT_STATUS_PACKED,
	})
	if err != nil {
		t.Fatalf("ListSellerOrders() error = %v", err)
	}
	if sellerList.query.SellerID != "seller_1" || sellerList.query.ActorUserID != "user_1" ||
		sellerList.query.PageSize != 10 || sellerList.query.FulfillmentFilter == nil ||
		*sellerList.query.FulfillmentFilter != domain.SellerFulfillmentStatusPacked {
		t.Fatalf("query = %+v, want trusted seller metadata and filter", sellerList.query)
	}
	if len(response.GetOrders()) != 1 || response.GetOrders()[0].GetSellerItemsTotal().GetMinorUnits() != 1500 ||
		len(response.GetOrders()[0].GetItems()) != 1 {
		t.Fatalf("response = %+v, want seller projection", response)
	}
}

func TestSellerFulfillmentReturnsSellerProjectionOnly(t *testing.T) {
	sellerUpdate := &handlerSellerFulfillmentExecutor{view: domain.SellerOrderView{
		OrderID: "ord_1", ParentOrderStatus: domain.OrderStatusPaid,
		SellerFulfillmentStatus: domain.SellerFulfillmentStatusPacked,
		Currency:                "INR", SellerItemsTotal: 1000,
		Items: []domain.SellerOrderItemView{{
			OrderItemID: "oi_1", OrderID: "ord_1", ProductID: "product_1",
			TitleSnapshot: "Snapshot", Quantity: 1, Currency: "INR",
			UnitAmount: 1000, TotalAmount: 1000, FulfillmentStatus: domain.FulfillmentStatusPacked,
		}},
	}}
	handler := newTestHandlerWithSeller(t, &handlerCreateExecutor{}, &handlerSellerListExecutor{}, sellerUpdate)
	ctx := authctx.WithActor(context.Background(), authctx.Actor{
		UserID: "user_1", SellerID: "seller_1", Roles: []string{"seller"},
	})
	response, err := handler.UpdateFulfillment(ctx, &orderv1.UpdateFulfillmentRequest{
		OrderId: "ord_1", TargetStatus: orderv1.OrderStatus_ORDER_STATUS_PACKED,
	})
	if err != nil {
		t.Fatalf("UpdateFulfillment() error = %v", err)
	}
	if sellerUpdate.command.ActorUserID != "user_1" || sellerUpdate.command.SellerID != "seller_1" ||
		sellerUpdate.command.TargetStatus != domain.FulfillmentStatusPacked {
		t.Fatalf("command = %+v, want trusted seller update", sellerUpdate.command)
	}
	if response.GetOrder() != nil || response.GetSellerOrder().GetOrderId() != "ord_1" ||
		len(response.GetSellerOrder().GetItems()) != 1 {
		t.Fatalf("response = %+v, want seller projection only", response)
	}
}

func TestSellerRoutesRequireSellerIDMetadata(t *testing.T) {
	handler := newTestHandler(t, &handlerCreateExecutor{})
	ctx := authctx.WithActor(context.Background(), authctx.Actor{UserID: "user_1", Roles: []string{"seller"}})
	_, err := handler.ListSellerOrders(ctx, &orderv1.ListSellerOrdersRequest{})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("ListSellerOrders(no seller id) code = %v, want Unauthenticated", status.Code(err))
	}
	_, err = handler.UpdateFulfillment(ctx, &orderv1.UpdateFulfillmentRequest{
		OrderId: "ord_1", TargetStatus: orderv1.OrderStatus_ORDER_STATUS_PACKED,
	})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("UpdateFulfillment(no seller id) code = %v, want Unauthenticated", status.Code(err))
	}
}

func TestCancelOrderMapsAuthorizedActorAndReason(t *testing.T) {
	cancel := &handlerCancelExecutor{order: domain.Order{
		OrderID: "ord_1", UserID: "user_1", Status: domain.OrderStatusCancelled,
	}}
	handler, err := NewServer(
		&handlerCreateExecutor{},
		handlerGetExecutor{},
		handlerListExecutor{},
		&handlerSellerListExecutor{},
		handlerFulfillmentExecutor{},
		cancel,
		&handlerSellerFulfillmentExecutor{},
		nil,
	)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	ctx := authctx.WithActor(context.Background(), authctx.Actor{
		UserID: "user_1", Roles: []string{"buyer"}, RequestID: "req_1",
	})
	response, err := handler.CancelOrder(ctx, &orderv1.CancelOrderRequest{
		OrderId: "ord_1", ReasonCode: "Buyer Requested",
	})
	if err != nil {
		t.Fatalf("CancelOrder() error = %v", err)
	}
	if cancel.command.ActorID != "user_1" || cancel.command.OrderID != "ord_1" ||
		cancel.command.TraceID != "req_1" || cancel.command.ReasonCode != "Buyer Requested" ||
		response.GetOrder().GetStatus() != orderv1.OrderStatus_ORDER_STATUS_CANCELLED {
		t.Fatalf("command/response = %+v/%+v, want mapped cancellation", cancel.command, response)
	}
}

func TestCancelOrderRejectsUnprivilegedActor(t *testing.T) {
	handler := newTestHandler(t, &handlerCreateExecutor{})
	ctx := authctx.WithActor(context.Background(), authctx.Actor{UserID: "seller_1", Roles: []string{"seller"}})
	_, err := handler.CancelOrder(ctx, &orderv1.CancelOrderRequest{OrderId: "ord_1"})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("CancelOrder(seller) code = %v, want PermissionDenied", status.Code(err))
	}
}

func TestRegisteredServiceRequiresTrustedCallerMetadata(t *testing.T) {
	handler := newTestHandler(t, &handlerCreateExecutor{})
	server, err := NewGRPCServer(
		handler,
		identifier.NewCryptoGenerator(),
		RuntimeConfig{TrustedCallerToken: "a-trusted-internal-caller-token-with-length"},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	if err != nil {
		t.Fatalf("NewGRPCServer() error = %v", err)
	}
	listener := bufconn.Listen(1024 * 1024)
	go func() {
		_ = server.Serve(listener)
	}()
	defer server.Stop()

	clientConn, err := grpc.NewClient(
		"passthrough:///order-test",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("grpc.NewClient() error = %v", err)
	}
	defer clientConn.Close()
	client := orderv1.NewOrderServiceClient(clientConn)
	healthClient := grpc_health_v1.NewHealthClient(clientConn)
	healthResponse, err := healthClient.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{
		Service: orderv1.OrderService_ServiceDesc.ServiceName,
	})
	if err != nil || healthResponse.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Fatalf("Health.Check() response/error = %+v/%v, want SERVING without business credentials", healthResponse, err)
	}

	_, err = client.GetOrder(context.Background(), &orderv1.GetOrderRequest{OrderId: "ord_1"})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("GetOrder(untrusted) code = %v, want Unauthenticated", status.Code(err))
	}
	trusted := metadata.AppendToOutgoingContext(
		context.Background(),
		"x-internal-token", "a-trusted-internal-caller-token-with-length",
		"x-user-id", "user_1",
		"x-roles", "buyer",
	)
	response, err := client.GetOrder(trusted, &orderv1.GetOrderRequest{OrderId: "ord_1"})
	if err != nil || response.GetOrder().GetOrderId() != "ord_1" {
		t.Fatalf("GetOrder(trusted) response/error = %+v/%v", response, err)
	}
}

type handlerCreateExecutor struct {
	command usecase.CreateOrderCommand
	result  *usecase.CreateOrderResult
	err     error
}

func (f *handlerCreateExecutor) Execute(_ context.Context, command usecase.CreateOrderCommand) (*usecase.CreateOrderResult, error) {
	f.command = command
	return f.result, f.err
}

type handlerGetExecutor struct{}

func (handlerGetExecutor) Execute(context.Context, usecase.GetOrderQuery) (domain.Order, error) {
	return domain.Order{OrderID: "ord_1", UserID: "user_1", Status: domain.OrderStatusCreated, CreatedAt: time.Now()}, nil
}

type handlerListExecutor struct{}

func (handlerListExecutor) Execute(context.Context, usecase.ListOrdersQuery) (domain.OrderPage, error) {
	return domain.OrderPage{}, nil
}

type handlerSellerListExecutor struct {
	query usecase.ListSellerOrdersQuery
	page  domain.SellerOrderPage
	err   error
}

func (f *handlerSellerListExecutor) Execute(_ context.Context, query usecase.ListSellerOrdersQuery) (domain.SellerOrderPage, error) {
	f.query = query
	return f.page, f.err
}

type handlerFulfillmentExecutor struct{}

func (handlerFulfillmentExecutor) Execute(context.Context, usecase.UpdateFulfillmentCommand) (domain.Order, error) {
	return domain.Order{}, nil
}

type handlerCancelExecutor struct {
	command usecase.CancelOrderCommand
	order   domain.Order
	err     error
}

func (f *handlerCancelExecutor) Execute(_ context.Context, command usecase.CancelOrderCommand) (domain.Order, error) {
	f.command = command
	if f.order.OrderID == "" {
		f.order = domain.Order{OrderID: command.OrderID, UserID: command.ActorID, Status: domain.OrderStatusCancelled}
	}
	return f.order, f.err
}

type handlerSellerFulfillmentExecutor struct {
	command usecase.UpdateSellerFulfillmentCommand
	view    domain.SellerOrderView
	err     error
}

func (f *handlerSellerFulfillmentExecutor) Execute(_ context.Context, command usecase.UpdateSellerFulfillmentCommand) (domain.SellerOrderView, error) {
	f.command = command
	return f.view, f.err
}

func newTestHandler(t *testing.T, create usecase.CreateOrderExecutor) *Server {
	t.Helper()
	return newTestHandlerWithSeller(t, create, &handlerSellerListExecutor{}, &handlerSellerFulfillmentExecutor{})
}

func newTestHandlerWithSeller(
	t *testing.T,
	create usecase.CreateOrderExecutor,
	listSeller usecase.ListSellerOrdersExecutor,
	updateSeller usecase.UpdateSellerFulfillmentExecutor,
) *Server {
	t.Helper()
	handler, err := NewServer(
		create,
		handlerGetExecutor{},
		handlerListExecutor{},
		listSeller,
		handlerFulfillmentExecutor{},
		&handlerCancelExecutor{},
		updateSeller,
		nil,
	)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	return handler
}

func validCreateRequest() *orderv1.CreateOrderRequest {
	return &orderv1.CreateOrderRequest{
		CartId: "cart_1", IdempotencyKey: "attempt_1",
		ShippingAddress: &orderv1.AddressSnapshot{
			RecipientName: "Buyer", Line1: "Street", City: "Delhi", PostalCode: "110001", CountryCode: "IN",
		},
	}
}
