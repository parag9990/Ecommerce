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

type handlerFulfillmentExecutor struct{}

func (handlerFulfillmentExecutor) Execute(context.Context, usecase.UpdateFulfillmentCommand) (domain.Order, error) {
	return domain.Order{}, nil
}

func newTestHandler(t *testing.T, create usecase.CreateOrderExecutor) *Server {
	t.Helper()
	handler, err := NewServer(create, handlerGetExecutor{}, handlerListExecutor{}, handlerFulfillmentExecutor{}, nil)
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
