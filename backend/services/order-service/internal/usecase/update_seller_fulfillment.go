package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/orderevents"
)

type UpdateSellerFulfillmentUsecase struct {
	orders SellerOrderRepository
	reader OrderReadRepository
	ids    IDGenerator
	clock  Clock
	logger *slog.Logger
}

func NewUpdateSellerFulfillmentUsecase(
	orders SellerOrderRepository,
	reader OrderReadRepository,
	ids IDGenerator,
	logger *slog.Logger,
) (*UpdateSellerFulfillmentUsecase, error) {
	if orders == nil {
		return nil, errors.New("seller order repository is required")
	}
	if reader == nil {
		return nil, errors.New("order read repository is required")
	}
	if ids == nil {
		return nil, errors.New("id generator is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &UpdateSellerFulfillmentUsecase{orders: orders, reader: reader, ids: ids, clock: realClock{}, logger: logger}, nil
}

func (u *UpdateSellerFulfillmentUsecase) WithClock(clock Clock) {
	if clock != nil {
		u.clock = clock
	}
}

func (u *UpdateSellerFulfillmentUsecase) Execute(ctx context.Context, command UpdateSellerFulfillmentCommand) (domain.SellerOrderView, error) {
	command.ActorUserID = strings.TrimSpace(command.ActorUserID)
	command.SellerID = strings.TrimSpace(command.SellerID)
	command.OrderID = strings.TrimSpace(command.OrderID)
	command.RequestID = strings.TrimSpace(command.RequestID)
	command.Carrier = strings.TrimSpace(command.Carrier)
	command.TrackingNumber = strings.TrimSpace(command.TrackingNumber)
	command.Note = strings.TrimSpace(command.Note)
	if command.ActorUserID == "" || command.SellerID == "" || command.OrderID == "" ||
		command.OccurredAt.IsZero() || len(command.ActorUserID) > maxRequestIDLength ||
		len(command.SellerID) > maxRequestIDLength || len(command.OrderID) > maxRequestIDLength ||
		len(command.RequestID) > maxSessionIDLength {
		return domain.SellerOrderView{}, domain.ErrInvalidRequest
	}
	if !hasAnyRole(command.Roles, "seller", "seller_manager", "seller_order_manager") {
		return domain.SellerOrderView{}, domain.ErrForbidden
	}
	historyID, err := u.ids.NewID("osh")
	if err != nil {
		return domain.SellerOrderView{}, err
	}
	shipmentID, err := u.ids.NewID("shp")
	if err != nil {
		return domain.SellerOrderView{}, err
	}
	transition := domain.SellerFulfillmentTransition{
		OrderID:        command.OrderID,
		SellerID:       command.SellerID,
		ShipmentID:     shipmentID,
		TargetStatus:   command.TargetStatus,
		TrackingNumber: command.TrackingNumber,
		Carrier:        command.Carrier,
		Note:           command.Note,
		ActorID:        command.ActorUserID,
		OccurredAt:     command.OccurredAt.UTC(),
		HistoryID:      historyID,
	}
	if err := transition.Validate(); err != nil {
		return domain.SellerOrderView{}, err
	}
	event, err := u.buildDeliveredEventCandidate(ctx, command, transition)
	if err != nil {
		return domain.SellerOrderView{}, err
	}
	view, err := u.orders.UpdateSellerFulfillment(ctx, transition, event)
	if err != nil {
		return domain.SellerOrderView{}, err
	}
	u.logger.Info("seller.fulfillment.updated",
		slog.String("seller_id", command.SellerID),
		slog.String("actor_user_id", command.ActorUserID),
		slog.String("order_id", view.OrderID),
		slog.String("target_status", string(command.TargetStatus)),
		slog.String("request_id", command.RequestID),
	)
	if event != nil && view.ParentOrderStatus == domain.OrderStatusDelivered {
		u.logger.Info("order.outbox.created",
			slog.String("event_id", event.EventID),
			slog.String("event_type", event.EventType),
			slog.String("order_id", view.OrderID),
			slog.String("trace_id", event.TraceID),
		)
	}
	return view, nil
}

func (u *UpdateSellerFulfillmentUsecase) buildDeliveredEventCandidate(
	ctx context.Context,
	command UpdateSellerFulfillmentCommand,
	transition domain.SellerFulfillmentTransition,
) (*domain.OutboxEvent, error) {
	if command.TargetStatus != domain.FulfillmentStatusDelivered {
		return nil, nil
	}
	order, err := u.reader.GetOrder(ctx, command.OrderID)
	if err != nil {
		return nil, err
	}
	if order.Status != domain.OrderStatusShipped {
		return nil, nil
	}
	eventID, err := u.ids.NewID("evt")
	if err != nil {
		return nil, err
	}
	event, err := orderevents.BuildOrderDeliveredEvent(
		eventID,
		command.RequestID,
		transition.OccurredAt,
		order,
		order.Status,
		transition.HistoryID,
		command.OccurredAt,
	)
	if err != nil {
		return nil, err
	}
	return &event, nil
}
