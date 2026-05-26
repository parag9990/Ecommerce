package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
)

const paymentIntentAttemptNumber = "1"

type InitiateOrderPaymentCommand struct {
	OrderID string
	UserID  string
}

type InitiateOrderPaymentResult struct {
	OrderID           string
	OrderStatus       domain.OrderStatus
	PaymentID         string
	PaymentStatus     string
	Provider          string
	ClientActionToken string
	ExpiresAt         time.Time
}

type InitiateOrderPaymentConfig struct {
	ReturnURL               string
	AllowedCurrencies       []string
	InventoryReleaseTimeout time.Duration
}

type InitiateOrderPaymentUsecase struct {
	orders            PaymentOrderRepository
	payments          PaymentClient
	inventory         InventoryReleaser
	ids               IDGenerator
	clock             Clock
	returnURL         string
	allowedCurrencies map[string]struct{}
	releaseTimeout    time.Duration
	logger            *slog.Logger
}

func NewInitiateOrderPaymentUsecase(
	orders PaymentOrderRepository,
	payments PaymentClient,
	inventory InventoryReleaser,
	ids IDGenerator,
	config InitiateOrderPaymentConfig,
	logger *slog.Logger,
) (*InitiateOrderPaymentUsecase, error) {
	if orders == nil {
		return nil, errors.New("payment order repository is required")
	}
	if payments == nil {
		return nil, errors.New("payment client is required")
	}
	if inventory == nil {
		return nil, errors.New("inventory client is required")
	}
	if ids == nil {
		return nil, errors.New("id generator is required")
	}
	returnURL := strings.TrimSpace(config.ReturnURL)
	parsedReturnURL, err := url.ParseRequestURI(returnURL)
	if err != nil || parsedReturnURL.Scheme != "https" || parsedReturnURL.Host == "" {
		return nil, errors.New("payment return URL must be an absolute HTTPS URL")
	}
	if config.InventoryReleaseTimeout <= 0 {
		return nil, errors.New("payment inventory release timeout must be greater than zero")
	}
	allowedCurrencies := make(map[string]struct{}, len(config.AllowedCurrencies))
	for _, currency := range config.AllowedCurrencies {
		normalized, valid := normalizeCurrencyCode(currency)
		if !valid {
			return nil, fmt.Errorf("invalid allowed payment currency %q", currency)
		}
		allowedCurrencies[normalized] = struct{}{}
	}
	if len(allowedCurrencies) == 0 {
		return nil, errors.New("at least one allowed payment currency is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &InitiateOrderPaymentUsecase{
		orders:            orders,
		payments:          payments,
		inventory:         inventory,
		ids:               ids,
		clock:             realClock{},
		returnURL:         returnURL,
		allowedCurrencies: allowedCurrencies,
		releaseTimeout:    config.InventoryReleaseTimeout,
		logger:            logger,
	}, nil
}

func (u *InitiateOrderPaymentUsecase) WithClock(clock Clock) {
	if clock != nil {
		u.clock = clock
	}
}

func (u *InitiateOrderPaymentUsecase) Execute(ctx context.Context, command InitiateOrderPaymentCommand) (*InitiateOrderPaymentResult, error) {
	command.OrderID = strings.TrimSpace(command.OrderID)
	command.UserID = strings.TrimSpace(command.UserID)
	if command.OrderID == "" || command.UserID == "" ||
		len(command.OrderID) > maxRequestIDLength || len(command.UserID) > maxRequestIDLength {
		return nil, domain.ErrOrderNotPayable
	}
	order, err := u.orders.GetForPayment(ctx, command.OrderID)
	if err != nil {
		return nil, err
	}
	if err := u.validatePayableOrder(order, command.UserID); err != nil {
		return nil, err
	}

	intent, err := u.payments.CreatePaymentIntent(ctx, CreatePaymentIntentRequest{
		OrderID:        order.OrderID,
		UserID:         order.UserID,
		Amount:         order.TotalAmount,
		Currency:       order.Currency,
		IdempotencyKey: paymentIntentKey(order.OrderID),
		ReturnURL:      u.returnURL,
	})
	if err != nil {
		return nil, u.handleIntentFailure(ctx, order, err)
	}
	if err := u.validateIntentResponse(intent); err != nil {
		u.logger.Error("order.payment.intent_invalid_response",
			slog.String("order_id", order.OrderID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	history, err := newPaymentHistory(
		u.ids, u.clock, order.OrderID, domain.OrderStatusCreated, domain.OrderStatusPendingPayment,
		"payment_intent_created", domain.OrderStatusActorSystem, "order-service", nil,
	)
	if err != nil {
		return nil, err
	}
	changed, err := u.orders.AttachPaymentIntent(ctx, order.OrderID, intent.PaymentID, history)
	if err != nil {
		return nil, fmt.Errorf("attach payment intent: %w", err)
	}
	if !changed {
		latest, getErr := u.orders.GetForPayment(ctx, order.OrderID)
		if getErr == nil && latest.Status == domain.OrderStatusPendingPayment && latest.PaymentID == intent.PaymentID {
			return paymentIntentResult(latest.OrderID, intent), nil
		}
		return nil, domain.ErrInvalidPaymentTransition
	}
	u.logger.Info("order.payment.intent_attached",
		slog.String("order_id", order.OrderID),
		slog.String("payment_id", intent.PaymentID),
		slog.String("status", domain.OrderStatusPendingPayment.String()),
	)
	return paymentIntentResult(order.OrderID, intent), nil
}

func (u *InitiateOrderPaymentUsecase) validatePayableOrder(order domain.Order, userID string) error {
	if strings.TrimSpace(order.OrderID) == "" {
		return domain.ErrOrderNotFound
	}
	if strings.TrimSpace(order.UserID) != userID {
		return domain.ErrOrderForbidden
	}
	isFreshOrder := order.Status == domain.OrderStatusCreated && strings.TrimSpace(order.PaymentID) == ""
	isExistingIntent := order.Status == domain.OrderStatusPendingPayment && strings.TrimSpace(order.PaymentID) != ""
	if !isFreshOrder && !isExistingIntent {
		return domain.ErrOrderNotPayable
	}
	if order.TotalAmount <= 0 {
		return domain.ErrInvalidOrderAmount
	}
	currency, valid := normalizeCurrencyCode(order.Currency)
	if !valid || currency != order.Currency {
		return domain.ErrUnsupportedCurrency
	}
	if _, permitted := u.allowedCurrencies[currency]; !permitted {
		return domain.ErrUnsupportedCurrency
	}
	if strings.TrimSpace(order.InventoryReservationID) == "" ||
		len(strings.TrimSpace(order.InventoryReservationID)) > maxRequestIDLength ||
		order.InventoryReservedUntil.IsZero() ||
		!order.InventoryReservedUntil.After(u.clock.Now()) {
		return domain.ErrInventoryReservationExpired
	}
	return nil
}

func (u *InitiateOrderPaymentUsecase) validateIntentResponse(intent *CreatePaymentIntentResponse) error {
	if intent == nil || strings.TrimSpace(intent.PaymentID) == "" ||
		len(strings.TrimSpace(intent.PaymentID)) > maxRequestIDLength ||
		strings.TrimSpace(intent.Provider) == "" || !domain.IsPayableIntentStatus(strings.TrimSpace(intent.Status)) ||
		intent.ExpiresAt.IsZero() || !intent.ExpiresAt.After(u.clock.Now()) {
		return domain.ErrInvalidPaymentIntentResponse
	}
	if strings.TrimSpace(intent.Status) == string(domain.PaymentIntentStatusRequiresAction) &&
		strings.TrimSpace(intent.ClientActionToken) == "" {
		return domain.ErrInvalidPaymentIntentResponse
	}
	return nil
}

func (u *InitiateOrderPaymentUsecase) handleIntentFailure(ctx context.Context, order domain.Order, intentErr error) error {
	if errors.Is(intentErr, domain.ErrPaymentIntentPendingResolution) {
		u.logger.Warn("order.payment.intent_outcome_unknown",
			slog.String("order_id", order.OrderID),
		)
		return domain.ErrPaymentIntentPendingResolution
	}
	history, err := newPaymentHistory(
		u.ids, u.clock, order.OrderID, domain.OrderStatusCreated, domain.OrderStatusPaymentFailed,
		"intent_creation_failed", domain.OrderStatusActorSystem, "order-service", nil,
	)
	if err != nil {
		return err
	}
	changed, err := u.orders.MarkPaymentFailedFromCreated(ctx, order.OrderID, history)
	if err != nil {
		return fmt.Errorf("record terminal intent failure: %w", err)
	}
	if !changed {
		latest, getErr := u.orders.GetForPayment(ctx, order.OrderID)
		if getErr == nil && latest.Status == domain.OrderStatusPaymentFailed {
			return domain.ErrPaymentIntentCreationFailed
		}
		return domain.ErrInvalidPaymentTransition
	}
	u.logger.Error("order.payment.intent_creation_failed",
		slog.String("order_id", order.OrderID),
	)
	releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), u.releaseTimeout)
	defer cancel()
	if err := u.inventory.ReleaseInventory(releaseCtx, ReleaseInventoryRequest{
		ReservationID:  order.InventoryReservationID,
		IdempotencyKey: "inventory_release:" + order.OrderID + ":payment_intent_failed",
		Reason:         "intent_creation_failed",
	}); err != nil {
		u.logger.Error("order.payment.intent_failure_release_failed",
			slog.String("order_id", order.OrderID),
			slog.String("reservation_id", order.InventoryReservationID),
			slog.String("error", err.Error()),
		)
	}
	return domain.ErrPaymentIntentCreationFailed
}

func paymentIntentKey(orderID string) string {
	return "payment_intent:" + strings.TrimSpace(orderID) + ":" + paymentIntentAttemptNumber
}

func paymentIntentResult(orderID string, intent *CreatePaymentIntentResponse) *InitiateOrderPaymentResult {
	return &InitiateOrderPaymentResult{
		OrderID:           orderID,
		OrderStatus:       domain.OrderStatusPendingPayment,
		PaymentID:         strings.TrimSpace(intent.PaymentID),
		PaymentStatus:     strings.TrimSpace(intent.Status),
		Provider:          strings.TrimSpace(intent.Provider),
		ClientActionToken: intent.ClientActionToken,
		ExpiresAt:         intent.ExpiresAt.UTC(),
	}
}
