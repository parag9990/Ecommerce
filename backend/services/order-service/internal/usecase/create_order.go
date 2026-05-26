package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
)

type CreateOrderConfig struct {
	IdempotencyTTL time.Duration
}

// CreateOrderUsecase owns the checkout idempotency boundary around order
// creation and payment coordination.
type CreateOrderUsecase struct {
	checkout    CreateFromCartExecutor
	payment     PaymentInitiationExecutor
	orders      OrderReadRepository
	idempotency IdempotencyRepository
	clock       Clock
	config      CreateOrderConfig
	logger      *slog.Logger
}

func NewCreateOrderUsecase(
	checkout CreateFromCartExecutor,
	payment PaymentInitiationExecutor,
	orders OrderReadRepository,
	idempotency IdempotencyRepository,
	config CreateOrderConfig,
	logger *slog.Logger,
) (*CreateOrderUsecase, error) {
	if checkout == nil {
		return nil, errors.New("cart checkout usecase is required")
	}
	if payment == nil {
		return nil, errors.New("payment initiation usecase is required")
	}
	if orders == nil {
		return nil, errors.New("order read repository is required")
	}
	if idempotency == nil {
		return nil, errors.New("idempotency repository is required")
	}
	if config.IdempotencyTTL <= 0 {
		return nil, errors.New("idempotency ttl must be greater than zero")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &CreateOrderUsecase{
		checkout:    checkout,
		payment:     payment,
		orders:      orders,
		idempotency: idempotency,
		clock:       realClock{},
		config:      config,
		logger:      logger,
	}, nil
}

func (u *CreateOrderUsecase) WithClock(clock Clock) {
	if clock != nil {
		u.clock = clock
	}
}

func (u *CreateOrderUsecase) Execute(ctx context.Context, command CreateOrderCommand) (*CreateOrderResult, error) {
	command = normalizeCreateOrderCommand(command)
	if command.UserID == "" || command.CartID == "" ||
		len(command.UserID) > maxRequestIDLength || len(command.CartID) > maxRequestIDLength ||
		len(command.SessionID) > maxSessionIDLength || len(command.TraceID) > maxSessionIDLength {
		return nil, domain.ErrInvalidCheckoutCommand
	}
	if err := domain.ValidateIdempotencyKey(command.IdempotencyKey); err != nil {
		return nil, err
	}
	if err := command.ShippingAddress.Validate(); err != nil {
		return nil, err
	}
	requestHash, err := checkoutRequestHash(command)
	if err != nil {
		return nil, fmt.Errorf("hash checkout request: %w", err)
	}
	claim, err := u.idempotency.Claim(
		ctx,
		command.UserID,
		command.IdempotencyKey,
		requestHash,
		u.clock.Now().Add(u.config.IdempotencyTTL),
	)
	if err != nil {
		if errors.Is(err, domain.ErrIdempotencyConflict) {
			u.logger.Warn("order.checkout.idempotency_conflict",
				slog.String("key_reference", idempotencyKeyReference(command.UserID, command.IdempotencyKey)),
			)
		}
		return nil, err
	}
	u.logger.Info("order.checkout.idempotency_claim",
		slog.String("key_reference", idempotencyKeyReference(command.UserID, command.IdempotencyKey)),
		slog.String("decision", string(claim.Decision)),
		slog.String("order_id", claim.OrderID),
	)

	switch claim.Decision {
	case domain.ClaimAcquired:
		return u.executeClaimedCheckout(ctx, command, claim)
	case domain.ClaimReplay:
		return u.resultForExistingOrder(ctx, claim, false)
	case domain.ClaimResume:
		return u.resultForExistingOrder(ctx, claim, true)
	case domain.ClaimInProgress:
		return nil, domain.ErrCheckoutInProgress
	case domain.ClaimFailed:
		return nil, domain.ErrCheckoutFailed
	default:
		return nil, fmt.Errorf("%w: unsupported idempotency decision", domain.ErrTemporarilyUnavailable)
	}
}

func (u *CreateOrderUsecase) executeClaimedCheckout(
	ctx context.Context,
	command CreateOrderCommand,
	claim domain.IdempotencyClaim,
) (*CreateOrderResult, error) {
	created, err := u.checkout.Execute(ctx, CreateOrderFromCartCommand{
		UserID:          command.UserID,
		SessionID:       command.SessionID,
		TraceID:         command.TraceID,
		CartID:          command.CartID,
		IdempotencyKey:  command.IdempotencyKey,
		ShippingAddress: command.ShippingAddress,
		CouponCode:      command.CouponCode,
		Claim:           claim,
	})
	if err != nil {
		if isConfirmedTerminalBeforeOrder(err) {
			if markErr := u.idempotency.MarkFailed(ctx, claim.UserID, claim.Key); markErr != nil {
				u.logger.Error("order.checkout.idempotency_fail_update_failed",
					slog.String("key_reference", idempotencyKeyReference(claim.UserID, claim.Key)),
					slog.String("error", markErr.Error()),
				)
			}
		}
		return nil, err
	}
	return u.resultForExistingOrder(ctx, domain.IdempotencyClaim{
		UserID:      claim.UserID,
		Key:         claim.Key,
		RequestHash: claim.RequestHash,
		OrderID:     created.OrderID,
		Status:      domain.IdempotencyStatusProcessing,
		Decision:    domain.ClaimResume,
		ExpiresAt:   claim.ExpiresAt,
	}, true)
}

func (u *CreateOrderUsecase) resultForExistingOrder(
	ctx context.Context,
	claim domain.IdempotencyClaim,
	complete bool,
) (*CreateOrderResult, error) {
	if strings.TrimSpace(claim.OrderID) == "" {
		return nil, domain.ErrCheckoutInProgress
	}
	order, err := u.orders.GetOrder(ctx, claim.OrderID)
	if err != nil {
		return nil, err
	}
	if order.UserID != claim.UserID {
		return nil, fmt.Errorf("%w: idempotency order owner mismatch", domain.ErrTemporarilyUnavailable)
	}
	if order.Status == domain.OrderStatusPaymentFailed {
		return nil, domain.ErrCheckoutFailed
	}

	var action *PaymentActionView
	if order.Status == domain.OrderStatusCreated || order.Status == domain.OrderStatusPendingPayment {
		payment, paymentErr := u.payment.Execute(ctx, InitiateOrderPaymentCommand{
			OrderID: order.OrderID,
			UserID:  claim.UserID,
		})
		if paymentErr != nil {
			return nil, paymentErr
		}
		action = &PaymentActionView{
			PaymentID:         payment.PaymentID,
			ClientActionToken: payment.ClientActionToken,
			ExpiresAt:         payment.ExpiresAt,
		}
		order, err = u.orders.GetOrder(ctx, claim.OrderID)
		if err != nil {
			return nil, err
		}
	}
	if complete {
		if err := u.idempotency.Complete(ctx, claim.UserID, claim.Key, claim.OrderID); err != nil {
			return nil, err
		}
	}
	return &CreateOrderResult{Order: order, PaymentAction: action}, nil
}

func checkoutRequestHash(command CreateOrderCommand) (string, error) {
	type checkoutFingerprint struct {
		CartID          string                 `json:"cart_id"`
		ShippingAddress domain.AddressSnapshot `json:"shipping_address"`
		CouponCode      string                 `json:"coupon_code"`
	}
	payload, err := json.Marshal(checkoutFingerprint{
		CartID:          command.CartID,
		ShippingAddress: command.ShippingAddress,
		CouponCode:      command.CouponCode,
	})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}

func normalizeCreateOrderCommand(command CreateOrderCommand) CreateOrderCommand {
	command.UserID = strings.TrimSpace(command.UserID)
	command.SessionID = strings.TrimSpace(command.SessionID)
	command.TraceID = strings.TrimSpace(command.TraceID)
	command.CartID = strings.TrimSpace(command.CartID)
	command.IdempotencyKey = strings.TrimSpace(command.IdempotencyKey)
	command.CouponCode = strings.TrimSpace(command.CouponCode)
	command.ShippingAddress.RecipientName = strings.TrimSpace(command.ShippingAddress.RecipientName)
	command.ShippingAddress.Phone = strings.TrimSpace(command.ShippingAddress.Phone)
	command.ShippingAddress.Line1 = strings.TrimSpace(command.ShippingAddress.Line1)
	command.ShippingAddress.Line2 = strings.TrimSpace(command.ShippingAddress.Line2)
	command.ShippingAddress.City = strings.TrimSpace(command.ShippingAddress.City)
	command.ShippingAddress.State = strings.TrimSpace(command.ShippingAddress.State)
	command.ShippingAddress.PostalCode = strings.TrimSpace(command.ShippingAddress.PostalCode)
	command.ShippingAddress.Country = strings.ToUpper(strings.TrimSpace(command.ShippingAddress.Country))
	return command
}

func isConfirmedTerminalBeforeOrder(err error) bool {
	return errors.Is(err, domain.ErrCartNotFound) ||
		errors.Is(err, domain.ErrCartForbidden) ||
		errors.Is(err, domain.ErrCartEmpty) ||
		errors.Is(err, domain.ErrInvalidCartItem) ||
		errors.Is(err, domain.ErrInvalidQuantity) ||
		errors.Is(err, domain.ErrDuplicateCartItem) ||
		errors.Is(err, domain.ErrProductUnavailable) ||
		errors.Is(err, domain.ErrVariantUnavailable) ||
		errors.Is(err, domain.ErrCurrencyMismatch) ||
		errors.Is(err, domain.ErrInvalidPrice) ||
		errors.Is(err, domain.ErrMissingSeller) ||
		errors.Is(err, domain.ErrInvalidOrderTotal) ||
		errors.Is(err, domain.ErrInventoryUnavailable) ||
		errors.Is(err, domain.ErrCheckoutFailed)
}

func idempotencyKeyReference(userID string, key string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(userID) + ":" + strings.TrimSpace(key)))
	return hex.EncodeToString(sum[:8])
}
