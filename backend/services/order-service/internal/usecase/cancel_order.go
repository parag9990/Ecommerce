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

const maxCancellationReasonCodeLength = 80

type CancelOrderUsecase struct {
	reader        OrderReadRepository
	cancellations OrderCancellationRepository
	ids           IDGenerator
	clock         Clock
	logger        *slog.Logger
}

func NewCancelOrderUsecase(
	reader OrderReadRepository,
	cancellations OrderCancellationRepository,
	ids IDGenerator,
	logger *slog.Logger,
) (*CancelOrderUsecase, error) {
	if reader == nil {
		return nil, errors.New("order read repository is required")
	}
	if cancellations == nil {
		return nil, errors.New("order cancellation repository is required")
	}
	if ids == nil {
		return nil, errors.New("id generator is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &CancelOrderUsecase{
		reader:        reader,
		cancellations: cancellations,
		ids:           ids,
		clock:         realClock{},
		logger:        logger,
	}, nil
}

func (u *CancelOrderUsecase) WithClock(clock Clock) {
	if clock != nil {
		u.clock = clock
	}
}

func (u *CancelOrderUsecase) Execute(ctx context.Context, command CancelOrderCommand) (domain.Order, error) {
	command = normalizeCancelOrderCommand(command)
	if command.ActorID == "" || command.OrderID == "" || len(command.ActorID) > maxRequestIDLength ||
		len(command.OrderID) > maxRequestIDLength || len(command.TraceID) > maxSessionIDLength ||
		len(command.ReasonCode) > maxCancellationReasonCodeLength {
		return domain.Order{}, domain.ErrInvalidRequest
	}
	actorType, err := cancellationActorType(command.Roles)
	if err != nil {
		return domain.Order{}, err
	}

	order, err := u.reader.GetOrder(ctx, command.OrderID)
	if err != nil {
		return domain.Order{}, err
	}
	if actorType == domain.OrderStatusActorBuyer && order.UserID != command.ActorID {
		return domain.Order{}, domain.ErrOrderForbidden
	}
	if order.Status == domain.OrderStatusCancelled {
		return order, nil
	}
	if err := domain.ValidateOrderStatusTransition(order.Status, domain.OrderStatusCancelled); err != nil {
		return domain.Order{}, err
	}

	historyID, err := u.ids.NewID("osh")
	if err != nil {
		return domain.Order{}, fmt.Errorf("generate cancellation history id: %w", err)
	}
	eventID, err := u.ids.NewID("evt")
	if err != nil {
		return domain.Order{}, fmt.Errorf("generate cancellation event id: %w", err)
	}
	occurredAt := command.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = u.clock.Now()
	}
	occurredAt = occurredAt.UTC()
	reasonCode := cancellationReasonCode(command.ReasonCode, actorType)
	fromStatus := order.Status
	history := domain.OrderStatusHistoryEntry{
		ID:         historyID,
		OrderID:    order.OrderID,
		FromStatus: &fromStatus,
		ToStatus:   domain.OrderStatusCancelled,
		Reason:     reasonCode,
		ActorType:  actorType,
		ActorID:    command.ActorID,
		Metadata: map[string]any{
			"reason_code": reasonCode,
		},
		CreatedAt: occurredAt,
	}
	event, err := orderevents.BuildOrderCancelledEvent(
		eventID,
		command.TraceID,
		occurredAt,
		order,
		fromStatus,
		historyID,
		reasonCode,
		actorType,
	)
	if err != nil {
		return domain.Order{}, err
	}

	cancelled, err := u.cancellations.CancelOrder(ctx, order.OrderID, fromStatus, history, event)
	if err != nil {
		return domain.Order{}, err
	}
	u.logger.Info("order.cancelled",
		slog.String("order_id", cancelled.OrderID),
		slog.String("actor_id", command.ActorID),
		slog.String("actor_type", actorType.String()),
		slog.String("previous_status", fromStatus.String()),
	)
	u.logger.Info("order.outbox.created",
		slog.String("event_id", event.EventID),
		slog.String("event_type", event.EventType),
		slog.String("order_id", order.OrderID),
		slog.String("trace_id", event.TraceID),
	)
	return cancelled, nil
}

func normalizeCancelOrderCommand(command CancelOrderCommand) CancelOrderCommand {
	command.ActorID = strings.TrimSpace(command.ActorID)
	command.OrderID = strings.TrimSpace(command.OrderID)
	command.TraceID = strings.TrimSpace(command.TraceID)
	command.ReasonCode = normalizeCommandReasonCode(command.ReasonCode)
	return command
}

func cancellationActorType(roles []string) (domain.OrderStatusActorType, error) {
	if hasAnyRole(roles, "admin", "order_manager") {
		return domain.OrderStatusActorAdmin, nil
	}
	if hasAnyRole(roles, "buyer") {
		return domain.OrderStatusActorBuyer, nil
	}
	return "", domain.ErrForbidden
}

func cancellationReasonCode(reasonCode string, actorType domain.OrderStatusActorType) string {
	reasonCode = normalizeCommandReasonCode(reasonCode)
	if reasonCode != "" {
		return reasonCode
	}
	if actorType == domain.OrderStatusActorBuyer {
		return "buyer_requested"
	}
	return "admin_cancelled"
}

func normalizeCommandReasonCode(reasonCode string) string {
	reasonCode = strings.ToLower(strings.TrimSpace(reasonCode))
	reasonCode = strings.ReplaceAll(reasonCode, " ", "_")
	return strings.Trim(reasonCode, "_")
}
