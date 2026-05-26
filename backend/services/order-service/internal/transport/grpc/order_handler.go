package ordergrpc

import (
	"context"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/authctx"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/usecase"
	orderv1 "github.com/example/ecommerce-platform/backend/shared/gen/go/ecommerce/order/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) CreateOrder(ctx context.Context, request *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	actor, err := authctx.ActorFromContext(ctx)
	if err != nil {
		return nil, toStatusError(err)
	}
	if !authctx.HasRole(actor, "buyer") {
		return nil, status.Error(codes.PermissionDenied, "buyer role required")
	}
	if err := validateCreateOrderRequest(request); err != nil {
		return nil, toStatusError(err)
	}
	result, err := s.createOrder.Execute(ctx, usecase.CreateOrderCommand{
		UserID:         actor.UserID,
		SessionID:      actor.SessionID,
		TraceID:        actor.RequestID,
		CartID:         request.GetCartId(),
		IdempotencyKey: request.GetIdempotencyKey(),
		ShippingAddress: domain.AddressSnapshot{
			RecipientName: request.GetShippingAddress().GetRecipientName(),
			Phone:         request.GetShippingAddress().GetPhone(),
			Line1:         request.GetShippingAddress().GetLine1(),
			Line2:         request.GetShippingAddress().GetLine2(),
			City:          request.GetShippingAddress().GetCity(),
			State:         request.GetShippingAddress().GetState(),
			PostalCode:    request.GetShippingAddress().GetPostalCode(),
			Country:       request.GetShippingAddress().GetCountryCode(),
		},
	})
	if err != nil {
		return nil, toStatusError(err)
	}
	return &orderv1.CreateOrderResponse{
		Order:         mapOrder(result.Order),
		PaymentAction: mapPaymentAction(result.PaymentAction),
	}, nil
}

func (s *Server) GetOrder(ctx context.Context, request *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	actor, err := authctx.ActorFromContext(ctx)
	if err != nil {
		return nil, toStatusError(err)
	}
	if request == nil || request.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}
	order, err := s.getOrder.Execute(ctx, usecase.GetOrderQuery{
		ActorID: actor.UserID,
		Roles:   actor.Roles,
		OrderID: request.GetOrderId(),
	})
	if err != nil {
		return nil, toStatusError(err)
	}
	return &orderv1.GetOrderResponse{Order: mapOrder(order)}, nil
}

func (s *Server) ListOrders(ctx context.Context, request *orderv1.ListOrdersRequest) (*orderv1.ListOrdersResponse, error) {
	actor, err := authctx.ActorFromContext(ctx)
	if err != nil {
		return nil, toStatusError(err)
	}
	if !authctx.HasRole(actor, "buyer") {
		return nil, status.Error(codes.PermissionDenied, "buyer role required")
	}
	pageSize, filter, token, err := validateListOrdersRequest(request)
	if err != nil {
		return nil, toStatusError(err)
	}
	page, err := s.listOrders.Execute(ctx, usecase.ListOrdersQuery{
		UserID:       actor.UserID,
		PageSize:     pageSize,
		PageToken:    token,
		StatusFilter: filter,
	})
	if err != nil {
		return nil, toStatusError(err)
	}
	return &orderv1.ListOrdersResponse{
		Orders:        mapOrders(page.Orders),
		NextPageToken: page.NextPageToken,
	}, nil
}

func (s *Server) UpdateFulfillment(ctx context.Context, request *orderv1.UpdateFulfillmentRequest) (*orderv1.UpdateFulfillmentResponse, error) {
	actor, err := authctx.ActorFromContext(ctx)
	if err != nil {
		return nil, toStatusError(err)
	}
	if !authctx.HasRole(actor, "seller", "order_manager", "admin", "logistics") {
		return nil, status.Error(codes.PermissionDenied, "fulfillment role required")
	}
	targetStatus, err := validateFulfillmentRequest(request)
	if err != nil {
		return nil, toStatusError(err)
	}
	order, err := s.updateFulfillment.Execute(ctx, usecase.UpdateFulfillmentCommand{
		ActorID:        actor.UserID,
		Roles:          actor.Roles,
		OrderID:        request.GetOrderId(),
		TraceID:        actor.RequestID,
		TargetStatus:   targetStatus,
		TrackingNumber: request.GetTrackingNumber(),
		Carrier:        request.GetCarrier(),
		Note:           request.GetNote(),
		OccurredAt:     s.now().UTC(),
	})
	if err != nil {
		return nil, toStatusError(err)
	}
	return &orderv1.UpdateFulfillmentResponse{Order: mapOrder(order)}, nil
}
