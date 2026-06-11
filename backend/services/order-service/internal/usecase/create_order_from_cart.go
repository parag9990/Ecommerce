package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/orderevents"
)

const inventoryReleaseReasonOrderPersistenceFailed = "order_persistence_failed"

const (
	maxRequestIDLength      = 64
	maxSessionIDLength      = 128
	maxIdempotencyKeyLength = 128
)

type CreateOrderFromCartCommand struct {
	UserID          string
	CartID          string
	SessionID       string
	TraceID         string
	IdempotencyKey  string
	ShippingAddress domain.AddressSnapshot
	CouponCode      string
	Claim           domain.IdempotencyClaim
}

type CreateOrderFromCartResult struct {
	OrderID       string
	Status        domain.OrderStatus
	Currency      string
	TotalAmount   int64
	ReservationID string
	ReservedUntil time.Time
}

type CreateOrderFromCartConfig struct {
	ReservationTTL time.Duration
	ReleaseTimeout time.Duration
}

type CreateOrderFromCartUsecase struct {
	carts    CartClient
	products ProductClient
	orders   OrderRepository
	ids      IDGenerator
	clock    Clock
	config   CreateOrderFromCartConfig
	logger   *slog.Logger
}

func NewCreateOrderFromCartUsecase(
	carts CartClient,
	products ProductClient,
	orders OrderRepository,
	ids IDGenerator,
	config CreateOrderFromCartConfig,
	logger *slog.Logger,
) (*CreateOrderFromCartUsecase, error) {
	if carts == nil {
		return nil, errors.New("cart client is required")
	}
	if products == nil {
		return nil, errors.New("product client is required")
	}
	if orders == nil {
		return nil, errors.New("order repository is required")
	}
	if ids == nil {
		return nil, errors.New("id generator is required")
	}
	if config.ReservationTTL <= 0 {
		return nil, errors.New("inventory reservation ttl must be greater than zero")
	}
	if config.ReleaseTimeout <= 0 {
		return nil, errors.New("inventory release timeout must be greater than zero")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &CreateOrderFromCartUsecase{
		carts:    carts,
		products: products,
		orders:   orders,
		ids:      ids,
		clock:    realClock{},
		config:   config,
		logger:   logger,
	}, nil
}

func (u *CreateOrderFromCartUsecase) WithClock(clock Clock) {
	if clock != nil {
		u.clock = clock
	}
}

func (u *CreateOrderFromCartUsecase) Execute(ctx context.Context, command CreateOrderFromCartCommand) (*CreateOrderFromCartResult, error) {
	command.UserID = strings.TrimSpace(command.UserID)
	command.CartID = strings.TrimSpace(command.CartID)
	command.SessionID = strings.TrimSpace(command.SessionID)
	command.TraceID = strings.TrimSpace(command.TraceID)
	command.IdempotencyKey = strings.TrimSpace(command.IdempotencyKey)
	if command.UserID == "" || command.CartID == "" ||
		len(command.UserID) > maxRequestIDLength || len(command.CartID) > maxRequestIDLength ||
		len(command.SessionID) > maxSessionIDLength || len(command.TraceID) > maxSessionIDLength ||
		len(command.IdempotencyKey) > maxIdempotencyKeyLength {
		return nil, domain.ErrInvalidCheckoutCommand
	}
	if err := domain.ValidateIdempotencyKey(command.IdempotencyKey); err != nil {
		return nil, err
	}
	if err := command.Claim.ValidateForOrderBinding(command.UserID, command.IdempotencyKey); err != nil {
		return nil, err
	}
	if err := command.ShippingAddress.Validate(); err != nil {
		return nil, err
	}

	cart, err := u.carts.GetCart(ctx, GetCartRequest{UserID: command.UserID, CartID: command.CartID})
	if err != nil {
		return nil, fmt.Errorf("get cart: %w", err)
	}
	if cart == nil {
		return nil, domain.ErrCartNotFound
	}
	if strings.TrimSpace(cart.CartID) != command.CartID {
		return nil, domain.ErrCartNotFound
	}
	if err := domain.ValidateCartForCheckout(*cart, command.UserID); err != nil {
		return nil, err
	}

	products, err := u.products.BatchGetProducts(ctx, BatchGetProductsRequest{Items: productLookupItems(cart.Items)})
	if err != nil {
		return nil, fmt.Errorf("get checkout product snapshots: %w", err)
	}
	if products == nil {
		return nil, domain.ErrProductUnavailable
	}
	order, err := u.buildCreatedOrder(command, *cart, products.Items)
	if err != nil {
		return nil, err
	}
	eventID, err := u.ids.NewID("evt")
	if err != nil {
		return nil, fmt.Errorf("generate order event id: %w", err)
	}

	reservation, err := u.products.ReserveInventory(ctx, ReserveInventoryRequest{
		UserID:         command.UserID,
		CartID:         command.CartID,
		IdempotencyKey: childOperationKey("order-reservation", command.UserID, command.IdempotencyKey),
		TTL:            u.config.ReservationTTL,
		Items:          reservationItems(cart.Items),
	})
	if err != nil {
		return nil, fmt.Errorf("reserve inventory: %w", err)
	}
	if err := u.validateReservation(reservation); err != nil {
		if reservation != nil && strings.TrimSpace(reservation.ReservationID) != "" {
			_ = u.releaseReservation(ctx, command, reservation.ReservationID, "invalid_reservation_response")
		}
		return nil, err
	}
	order.InitialHistory.CreatedAt = u.clock.Now().UTC()
	if err := order.AttachInventoryReservation(reservation.ReservationID, reservation.ExpiresAt, u.clock.Now()); err != nil {
		_ = u.releaseReservation(ctx, command, reservation.ReservationID, "invalid_reservation_response")
		return nil, err
	}
	event, err := orderevents.BuildOrderCreatedEvent(eventID, command.TraceID, order.InitialHistory.CreatedAt, order)
	if err != nil {
		_ = u.releaseReservation(ctx, command, reservation.ReservationID, "invalid_order_event")
		return nil, err
	}

	if err := u.orders.CreateOrderAndAttachClaim(ctx, order, command.Claim, event); err != nil {
		u.logger.Error("order.create.persist_failed",
			slog.String("user_id", command.UserID),
			slog.String("cart_id", command.CartID),
			slog.String("order_id", order.OrderID),
			slog.String("error", err.Error()),
		)
		if errors.Is(err, domain.ErrOrderCreateOutcomeUnknown) {
			return nil, err
		}
		if releaseErr := u.releaseReservation(ctx, command, reservation.ReservationID, inventoryReleaseReasonOrderPersistenceFailed); releaseErr == nil {
			return nil, fmt.Errorf("%w: %w", domain.ErrOrderCreateFailed, domain.ErrCheckoutFailed)
		}
		return nil, domain.ErrOrderCreateFailed
	}

	u.logger.Info("order.create.from_cart_succeeded",
		slog.String("user_id", command.UserID),
		slog.String("cart_id", command.CartID),
		slog.String("session_id", command.SessionID),
		slog.String("order_id", order.OrderID),
		slog.String("status", order.Status.String()),
	)
	u.logger.Info("order.outbox.created",
		slog.String("event_id", event.EventID),
		slog.String("event_type", event.EventType),
		slog.String("order_id", order.OrderID),
		slog.String("trace_id", event.TraceID),
	)
	return &CreateOrderFromCartResult{
		OrderID:       order.OrderID,
		Status:        order.Status,
		Currency:      order.Currency,
		TotalAmount:   order.TotalAmount,
		ReservationID: reservation.ReservationID,
		ReservedUntil: reservation.ExpiresAt.UTC(),
	}, nil
}

func (u *CreateOrderFromCartUsecase) buildCreatedOrder(command CreateOrderFromCartCommand, cart domain.CartSnapshot, products []domain.ProductSnapshot) (domain.Order, error) {
	orderID, err := u.ids.NewID("ord")
	if err != nil {
		return domain.Order{}, fmt.Errorf("generate order id: %w", err)
	}
	historyID, err := u.ids.NewID("osh")
	if err != nil {
		return domain.Order{}, fmt.Errorf("generate order history id: %w", err)
	}
	itemIDs := make([]string, len(cart.Items))
	for index := range cart.Items {
		itemIDs[index], err = u.ids.NewID("oi")
		if err != nil {
			return domain.Order{}, fmt.Errorf("generate order item id: %w", err)
		}
	}
	return domain.BuildCreatedOrder(domain.BuildCreatedOrderInput{
		OrderID:      orderID,
		HistoryID:    historyID,
		OrderItemIDs: itemIDs,
		UserID:       command.UserID,
		Cart:         cart,
		Products:     products,
		Address:      command.ShippingAddress,
		CouponCode:   command.CouponCode,
		CreatedAt:    u.clock.Now(),
	})
}

func (u *CreateOrderFromCartUsecase) validateReservation(reservation *ReserveInventoryResponse) error {
	if reservation == nil || strings.TrimSpace(reservation.ReservationID) == "" ||
		reservation.ExpiresAt.IsZero() || !reservation.ExpiresAt.After(u.clock.Now()) {
		return domain.ErrInvalidReservation
	}
	return nil
}

func (u *CreateOrderFromCartUsecase) releaseReservation(ctx context.Context, command CreateOrderFromCartCommand, reservationID string, reason string) error {
	releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), u.config.ReleaseTimeout)
	defer cancel()
	if err := u.products.ReleaseInventory(releaseCtx, ReleaseInventoryRequest{
		ReservationID:  reservationID,
		IdempotencyKey: childOperationKey("order-reservation-release", command.UserID, command.IdempotencyKey),
		Reason:         reason,
	}); err != nil {
		u.logger.Error("order.inventory.release_failed",
			slog.String("reservation_id", reservationID),
			slog.String("reason", reason),
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}

func childOperationKey(prefix string, userID string, checkoutKey string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(userID) + ":" + strings.TrimSpace(checkoutKey)))
	return prefix + ":" + hex.EncodeToString(sum[:])
}

func productLookupItems(items []domain.CartItemSnapshot) []ProductLookupItem {
	result := make([]ProductLookupItem, 0, len(items))
	for _, item := range items {
		result = append(result, ProductLookupItem{ProductID: item.ProductID, VariantID: item.VariantID})
	}
	return result
}

func reservationItems(items []domain.CartItemSnapshot) []ReserveInventoryItem {
	result := make([]ReserveInventoryItem, 0, len(items))
	for _, item := range items {
		result = append(result, ReserveInventoryItem{
			ProductID: item.ProductID,
			VariantID: item.VariantID,
			Quantity:  item.Quantity,
		})
	}
	return result
}
