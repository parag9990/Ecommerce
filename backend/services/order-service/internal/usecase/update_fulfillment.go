package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/orderevents"
)

type UpdateFulfillmentUsecase struct {
	orders FulfillmentRepository
	reader OrderReadRepository
	ids    IDGenerator
	clock  Clock
	logger *slog.Logger
}

func NewUpdateFulfillmentUsecase(
	orders FulfillmentRepository,
	reader OrderReadRepository,
	ids IDGenerator,
	logger *slog.Logger,
) (*UpdateFulfillmentUsecase, error) {
	if orders == nil || reader == nil {
		return nil, errors.New("fulfillment and order read repositories are required")
	}
	if ids == nil {
		return nil, errors.New("id generator is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &UpdateFulfillmentUsecase{orders: orders, reader: reader, ids: ids, clock: realClock{}, logger: logger}, nil
}

func (u *UpdateFulfillmentUsecase) WithClock(clock Clock) {
	if clock != nil {
		u.clock = clock
	}
}

func (u *UpdateFulfillmentUsecase) Execute(ctx context.Context, command UpdateFulfillmentCommand) (domain.Order, error) {
	command.ActorID = strings.TrimSpace(command.ActorID)
	command.OrderID = strings.TrimSpace(command.OrderID)
	command.TraceID = strings.TrimSpace(command.TraceID)
	command.Carrier = strings.TrimSpace(command.Carrier)
	command.TrackingNumber = strings.TrimSpace(command.TrackingNumber)
	command.Note = strings.TrimSpace(command.Note)
	if command.ActorID == "" || command.OrderID == "" || command.OccurredAt.IsZero() ||
		len(command.ActorID) > maxRequestIDLength || len(command.OrderID) > maxRequestIDLength ||
		len(command.TraceID) > maxSessionIDLength {
		return domain.Order{}, domain.ErrInvalidRequest
	}
	actorType, requiredSellerID, err := fulfillmentActor(command.ActorID, command.Roles)
	if err != nil {
		return domain.Order{}, err
	}
	current, err := u.reader.GetOrder(ctx, command.OrderID)
	if err != nil {
		return domain.Order{}, err
	}
	historyID, err := u.ids.NewID("osh")
	if err != nil {
		return domain.Order{}, err
	}
	shipmentID, err := u.ids.NewID("shp")
	if err != nil {
		return domain.Order{}, err
	}
	fromStatus := current.Status
	transition := domain.FulfillmentTransition{
		OrderID:          command.OrderID,
		ShipmentID:       shipmentID,
		TargetStatus:     command.TargetStatus,
		TrackingNumber:   command.TrackingNumber,
		Carrier:          command.Carrier,
		Note:             command.Note,
		ActorID:          command.ActorID,
		ActorType:        actorType,
		RequiredSellerID: requiredSellerID,
		OccurredAt:       command.OccurredAt.UTC(),
		History: domain.OrderStatusHistoryEntry{
			ID:         historyID,
			OrderID:    command.OrderID,
			FromStatus: &fromStatus,
			ToStatus:   command.TargetStatus,
			Reason:     "fulfillment_" + command.TargetStatus.String(),
			ActorType:  actorType,
			ActorID:    command.ActorID,
			Metadata: map[string]any{
				"carrier":         command.Carrier,
				"tracking_number": command.TrackingNumber,
				"note":            command.Note,
			},
			CreatedAt: u.clock.Now().UTC(),
		},
	}
	if err := transition.Validate(current.Status); err != nil {
		return domain.Order{}, err
	}
	var event *domain.OutboxEvent
	if eventType, ok := orderevents.EventTypeForStatus(command.TargetStatus); ok {
		if eventType != orderevents.EventTypeOrderDelivered {
			return domain.Order{}, fmt.Errorf("%w: unsupported fulfillment event type", domain.ErrInvalidRequest)
		}
		eventID, err := u.ids.NewID("evt")
		if err != nil {
			return domain.Order{}, fmt.Errorf("generate order fulfillment event id: %w", err)
		}
		outboxEvent, err := orderevents.BuildOrderDeliveredEvent(
			eventID,
			command.TraceID,
			transition.History.CreatedAt,
			current,
			fromStatus,
			historyID,
			command.OccurredAt,
		)
		if err != nil {
			return domain.Order{}, err
		}
		event = &outboxEvent
	}
	order, err := u.orders.UpdateFulfillment(ctx, transition, event)
	if err != nil {
		return domain.Order{}, err
	}
	u.logger.Info("order.fulfillment.updated",
		slog.String("order_id", order.OrderID),
		slog.String("actor_id", command.ActorID),
		slog.String("status", order.Status.String()),
	)
	if event != nil {
		u.logger.Info("order.outbox.created",
			slog.String("event_id", event.EventID),
			slog.String("event_type", event.EventType),
			slog.String("order_id", order.OrderID),
			slog.String("trace_id", event.TraceID),
		)
	}
	return order, nil
}

func fulfillmentActor(actorID string, roles []string) (domain.OrderStatusActorType, string, error) {
	switch {
	case hasAnyRole(roles, "admin", "order_manager"):
		return domain.OrderStatusActorAdmin, "", nil
	case hasAnyRole(roles, "logistics"):
		return domain.OrderStatusActorLogistics, "", nil
	case hasAnyRole(roles, "seller"):
		return domain.OrderStatusActorSeller, actorID, nil
	default:
		return "", "", domain.ErrForbidden
	}
}
