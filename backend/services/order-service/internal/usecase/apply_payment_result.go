package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/orderevents"
)

const maxProviderEventIDLength = 128

type ApplyPaymentResultCommand struct {
	PaymentID       string
	OrderID         string
	TraceID         string
	Result          string
	Amount          int64
	Currency        string
	ProviderEventID string
	OccurredAt      time.Time
}

type ApplyPaymentResultConfig struct {
	InventoryActionTimeout time.Duration
}

type ApplyPaymentResultUsecase struct {
	orders        PaymentOrderRepository
	inventory     InventoryClient
	ids           IDGenerator
	clock         Clock
	actionTimeout time.Duration
	logger        *slog.Logger
}

func NewApplyPaymentResultUsecase(
	orders PaymentOrderRepository,
	inventory InventoryClient,
	ids IDGenerator,
	config ApplyPaymentResultConfig,
	logger *slog.Logger,
) (*ApplyPaymentResultUsecase, error) {
	if orders == nil {
		return nil, errors.New("payment order repository is required")
	}
	if inventory == nil {
		return nil, errors.New("inventory client is required")
	}
	if ids == nil {
		return nil, errors.New("id generator is required")
	}
	if config.InventoryActionTimeout <= 0 {
		return nil, errors.New("payment inventory action timeout must be greater than zero")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &ApplyPaymentResultUsecase{
		orders:        orders,
		inventory:     inventory,
		ids:           ids,
		clock:         realClock{},
		actionTimeout: config.InventoryActionTimeout,
		logger:        logger,
	}, nil
}

func (u *ApplyPaymentResultUsecase) WithClock(clock Clock) {
	if clock != nil {
		u.clock = clock
	}
}

func (u *ApplyPaymentResultUsecase) Execute(ctx context.Context, command ApplyPaymentResultCommand) error {
	result, err := validatePaymentResultCommand(&command)
	if err != nil {
		return err
	}
	order, err := u.orders.GetForPayment(ctx, command.OrderID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(order.PaymentID) != command.PaymentID {
		u.logger.Error("order.payment.result_binding_mismatch",
			slog.String("order_id", command.OrderID),
			slog.String("payment_id", command.PaymentID),
			slog.String("provider_event_id", command.ProviderEventID),
		)
		return domain.ErrPaymentOrderMismatch
	}
	if order.TotalAmount != command.Amount || order.Currency != command.Currency {
		u.logger.Error("order.payment.result_amount_mismatch",
			slog.String("order_id", command.OrderID),
			slog.String("payment_id", command.PaymentID),
			slog.String("provider_event_id", command.ProviderEventID),
		)
		return domain.ErrPaymentAmountMismatch
	}
	if strings.TrimSpace(order.InventoryReservationID) == "" ||
		len(strings.TrimSpace(order.InventoryReservationID)) > maxRequestIDLength {
		return domain.ErrInvalidReservation
	}

	toStatus, reason := paymentResultTransition(result)
	if order.Status == toStatus {
		u.logger.Info("order.payment.result_duplicate",
			slog.String("order_id", order.OrderID),
			slog.String("payment_id", order.PaymentID),
			slog.String("result", string(result)),
		)
		return nil
	}
	if order.Status != domain.OrderStatusPendingPayment {
		return domain.ErrInvalidPaymentTransition
	}
	history, err := newPaymentHistory(
		u.ids, u.clock, order.OrderID, domain.OrderStatusPendingPayment, toStatus, reason,
		domain.OrderStatusActorPaymentService, "payment-service",
		map[string]any{
			"provider_event_id": command.ProviderEventID,
			"occurred_at":       command.OccurredAt.UTC().Format(time.RFC3339Nano),
		},
	)
	if err != nil {
		return err
	}
	var event *domain.OutboxEvent
	if eventType, ok := orderevents.EventTypeForStatus(toStatus); ok {
		if eventType != orderevents.EventTypeOrderPaid {
			return fmt.Errorf("%w: unsupported payment event type", domain.ErrInvalidPaymentTransition)
		}
		eventID, err := u.ids.NewID("evt")
		if err != nil {
			return fmt.Errorf("generate order payment event id: %w", err)
		}
		outboxEvent, err := orderevents.BuildOrderPaidEvent(
			eventID,
			command.TraceID,
			history.CreatedAt,
			order,
			history.ID,
		)
		if err != nil {
			return err
		}
		event = &outboxEvent
	}
	changed, err := u.orders.TransitionPaymentResult(
		ctx, order.OrderID, order.PaymentID, domain.OrderStatusPendingPayment, toStatus, history, event,
	)
	if err != nil {
		return fmt.Errorf("apply payment result transition: %w", err)
	}
	if !changed {
		latest, getErr := u.orders.GetForPayment(ctx, order.OrderID)
		if getErr == nil && latest.Status == toStatus && latest.PaymentID == order.PaymentID {
			return nil
		}
		return domain.ErrInvalidPaymentTransition
	}
	u.logger.Info("order.payment.result_applied",
		slog.String("order_id", order.OrderID),
		slog.String("payment_id", order.PaymentID),
		slog.String("result", string(result)),
		slog.String("status", toStatus.String()),
	)
	if event != nil {
		u.logger.Info("order.outbox.created",
			slog.String("event_id", event.EventID),
			slog.String("event_type", event.EventType),
			slog.String("order_id", order.OrderID),
			slog.String("trace_id", event.TraceID),
		)
	}
	return u.finalizeInventory(ctx, order, result)
}

func validatePaymentResultCommand(command *ApplyPaymentResultCommand) (domain.PaymentResult, error) {
	command.PaymentID = strings.TrimSpace(command.PaymentID)
	command.OrderID = strings.TrimSpace(command.OrderID)
	command.TraceID = strings.TrimSpace(command.TraceID)
	command.Result = strings.TrimSpace(command.Result)
	command.Currency = strings.ToUpper(strings.TrimSpace(command.Currency))
	command.ProviderEventID = strings.TrimSpace(command.ProviderEventID)
	if command.PaymentID == "" || command.OrderID == "" ||
		len(command.PaymentID) > maxRequestIDLength || len(command.OrderID) > maxRequestIDLength ||
		len(command.TraceID) > maxSessionIDLength ||
		command.Amount <= 0 || command.ProviderEventID == "" ||
		len(command.ProviderEventID) > maxProviderEventIDLength || command.OccurredAt.IsZero() {
		return "", domain.ErrInvalidPaymentResult
	}
	if currency, valid := normalizeCurrencyCode(command.Currency); !valid || currency != command.Currency {
		return "", domain.ErrInvalidPaymentResult
	}
	return domain.ParsePaymentResult(command.Result)
}

func paymentResultTransition(result domain.PaymentResult) (domain.OrderStatus, string) {
	if result == domain.PaymentResultCaptured {
		return domain.OrderStatusPaid, "payment_captured_verified"
	}
	return domain.OrderStatusPaymentFailed, "provider_payment_failed"
}

func (u *ApplyPaymentResultUsecase) finalizeInventory(ctx context.Context, order domain.Order, result domain.PaymentResult) error {
	actionCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), u.actionTimeout)
	defer cancel()
	var err error
	if result == domain.PaymentResultCaptured {
		err = u.inventory.CommitInventory(actionCtx, CommitInventoryRequest{
			ReservationID:  order.InventoryReservationID,
			IdempotencyKey: "inventory_commit:" + order.OrderID,
		})
	} else {
		err = u.inventory.ReleaseInventory(actionCtx, ReleaseInventoryRequest{
			ReservationID:  order.InventoryReservationID,
			IdempotencyKey: "inventory_release:" + order.OrderID + ":payment_failed",
			Reason:         "provider_payment_failed",
		})
	}
	if err != nil {
		u.logger.Error("order.payment.inventory_finalization_failed",
			slog.String("order_id", order.OrderID),
			slog.String("payment_id", order.PaymentID),
			slog.String("result", string(result)),
			slog.String("reservation_id", order.InventoryReservationID),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("%w: %v", domain.ErrInventoryFinalizationFailed, err)
	}
	return nil
}
