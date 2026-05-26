package orderevents

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
)

const (
	ProducerOrderService = "order-service"
	OrderEventsTopic     = "order.events"
	EventVersion         = 1
	AggregateTypeOrder   = "order"

	EventTypeOrderCreated   = "OrderCreated"
	EventTypeOrderPaid      = "OrderPaid"
	EventTypeOrderCancelled = "OrderCancelled"
	EventTypeOrderDelivered = "OrderDelivered"

	RoutingKeyOrderCreated   = "order.created"
	RoutingKeyOrderPaid      = "order.paid"
	RoutingKeyOrderCancelled = "order.cancelled"
	RoutingKeyOrderDelivered = "order.delivered"
)

var ErrInvalidEvent = errors.New("invalid order event")

type Envelope struct {
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	Version       int             `json:"version"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Producer      string          `json:"producer"`
	TraceID       string          `json:"trace_id,omitempty"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   string          `json:"aggregate_id"`
	Payload       json.RawMessage `json:"payload"`
}

type OrderCreatedPayload struct {
	OrderID     string                    `json:"order_id"`
	UserID      string                    `json:"user_id"`
	Status      string                    `json:"status"`
	Currency    string                    `json:"currency"`
	TotalAmount int64                     `json:"total_amount"`
	Items       []OrderCreatedItemPayload `json:"items"`
}

type OrderCreatedItemPayload struct {
	OrderItemID string `json:"order_item_id"`
	SellerID    string `json:"seller_id"`
	ProductID   string `json:"product_id"`
	VariantID   string `json:"variant_id"`
	Quantity    int32  `json:"quantity"`
	TotalAmount int64  `json:"total_amount"`
}

type OrderPaidPayload struct {
	OrderID     string `json:"order_id"`
	UserID      string `json:"user_id"`
	PaymentID   string `json:"payment_id"`
	Status      string `json:"status"`
	Currency    string `json:"currency"`
	TotalAmount int64  `json:"total_amount"`
}

type OrderCancelledPayload struct {
	OrderID        string `json:"order_id"`
	UserID         string `json:"user_id"`
	PreviousStatus string `json:"previous_status"`
	Status         string `json:"status"`
	ReasonCode     string `json:"reason_code"`
	ActorType      string `json:"actor_type"`
}

type OrderDeliveredPayload struct {
	OrderID        string    `json:"order_id"`
	UserID         string    `json:"user_id"`
	PreviousStatus string    `json:"previous_status"`
	Status         string    `json:"status"`
	DeliveredAt    time.Time `json:"delivered_at"`
}

func BuildOrderCreatedEvent(eventID string, traceID string, occurredAt time.Time, order domain.Order) (domain.OutboxEvent, error) {
	payload := OrderCreatedPayload{
		OrderID:     strings.TrimSpace(order.OrderID),
		UserID:      strings.TrimSpace(order.UserID),
		Status:      domain.OrderStatusCreated.String(),
		Currency:    strings.TrimSpace(order.Currency),
		TotalAmount: order.TotalAmount,
		Items:       make([]OrderCreatedItemPayload, 0, len(order.Items)),
	}
	for _, item := range order.Items {
		payload.Items = append(payload.Items, OrderCreatedItemPayload{
			OrderItemID: strings.TrimSpace(item.OrderItemID),
			SellerID:    strings.TrimSpace(item.SellerID),
			ProductID:   strings.TrimSpace(item.ProductID),
			VariantID:   strings.TrimSpace(item.VariantID),
			Quantity:    item.Quantity,
			TotalAmount: item.TotalAmount,
		})
	}
	if payload.Status != order.Status.String() || strings.TrimSpace(order.InitialHistory.ID) == "" {
		return domain.OutboxEvent{}, ErrInvalidEvent
	}
	if len(payload.Items) == 0 || payload.TotalAmount < 0 {
		return domain.OutboxEvent{}, ErrInvalidEvent
	}
	for _, item := range payload.Items {
		if item.OrderItemID == "" || item.SellerID == "" || item.ProductID == "" ||
			item.VariantID == "" || item.Quantity <= 0 || item.TotalAmount < 0 {
			return domain.OutboxEvent{}, ErrInvalidEvent
		}
	}
	return buildOutboxEvent(
		eventID,
		EventTypeOrderCreated,
		RoutingKeyOrderCreated,
		traceID,
		occurredAt,
		payload.OrderID,
		order.InitialHistory.ID,
		payload.OrderID+":created",
		payload,
	)
}

func BuildOrderPaidEvent(eventID string, traceID string, occurredAt time.Time, order domain.Order, sourceHistoryID string) (domain.OutboxEvent, error) {
	payload := OrderPaidPayload{
		OrderID:     strings.TrimSpace(order.OrderID),
		UserID:      strings.TrimSpace(order.UserID),
		PaymentID:   strings.TrimSpace(order.PaymentID),
		Status:      domain.OrderStatusPaid.String(),
		Currency:    strings.TrimSpace(order.Currency),
		TotalAmount: order.TotalAmount,
	}
	if payload.PaymentID == "" || payload.TotalAmount <= 0 {
		return domain.OutboxEvent{}, ErrInvalidEvent
	}
	return buildOutboxEvent(
		eventID,
		EventTypeOrderPaid,
		RoutingKeyOrderPaid,
		traceID,
		occurredAt,
		payload.OrderID,
		sourceHistoryID,
		strings.TrimSpace(sourceHistoryID)+":"+EventTypeOrderPaid,
		payload,
	)
}

func BuildOrderCancelledEvent(
	eventID string,
	traceID string,
	occurredAt time.Time,
	order domain.Order,
	previousStatus domain.OrderStatus,
	sourceHistoryID string,
	reasonCode string,
	actorType domain.OrderStatusActorType,
) (domain.OutboxEvent, error) {
	payload := OrderCancelledPayload{
		OrderID:        strings.TrimSpace(order.OrderID),
		UserID:         strings.TrimSpace(order.UserID),
		PreviousStatus: previousStatus.String(),
		Status:         domain.OrderStatusCancelled.String(),
		ReasonCode:     normalizeReasonCode(reasonCode),
		ActorType:      actorType.String(),
	}
	if !domain.IsKnownOrderStatus(previousStatus) || !domain.IsKnownOrderStatusActorType(actorType) {
		return domain.OutboxEvent{}, ErrInvalidEvent
	}
	if payload.ReasonCode == "" || payload.ActorType == "" {
		return domain.OutboxEvent{}, ErrInvalidEvent
	}
	return buildOutboxEvent(
		eventID,
		EventTypeOrderCancelled,
		RoutingKeyOrderCancelled,
		traceID,
		occurredAt,
		payload.OrderID,
		sourceHistoryID,
		strings.TrimSpace(sourceHistoryID)+":"+EventTypeOrderCancelled,
		payload,
	)
}

func BuildOrderDeliveredEvent(
	eventID string,
	traceID string,
	occurredAt time.Time,
	order domain.Order,
	previousStatus domain.OrderStatus,
	sourceHistoryID string,
	deliveredAt time.Time,
) (domain.OutboxEvent, error) {
	payload := OrderDeliveredPayload{
		OrderID:        strings.TrimSpace(order.OrderID),
		UserID:         strings.TrimSpace(order.UserID),
		PreviousStatus: previousStatus.String(),
		Status:         domain.OrderStatusDelivered.String(),
		DeliveredAt:    deliveredAt.UTC(),
	}
	if !domain.IsKnownOrderStatus(previousStatus) || deliveredAt.IsZero() {
		return domain.OutboxEvent{}, ErrInvalidEvent
	}
	return buildOutboxEvent(
		eventID,
		EventTypeOrderDelivered,
		RoutingKeyOrderDelivered,
		traceID,
		occurredAt,
		payload.OrderID,
		sourceHistoryID,
		strings.TrimSpace(sourceHistoryID)+":"+EventTypeOrderDelivered,
		payload,
	)
}

func EventTypeForStatus(status domain.OrderStatus) (string, bool) {
	switch status {
	case domain.OrderStatusPaid:
		return EventTypeOrderPaid, true
	case domain.OrderStatusCancelled:
		return EventTypeOrderCancelled, true
	case domain.OrderStatusDelivered:
		return EventTypeOrderDelivered, true
	default:
		return "", false
	}
}

func buildOutboxEvent(
	eventID string,
	eventType string,
	routingKey string,
	traceID string,
	occurredAt time.Time,
	orderID string,
	sourceHistoryID string,
	deduplicationKey string,
	payload any,
) (domain.OutboxEvent, error) {
	eventID = strings.TrimSpace(eventID)
	orderID = strings.TrimSpace(orderID)
	sourceHistoryID = strings.TrimSpace(sourceHistoryID)
	deduplicationKey = strings.TrimSpace(deduplicationKey)
	traceID = strings.TrimSpace(traceID)
	if eventID == "" || orderID == "" || sourceHistoryID == "" || deduplicationKey == "" || occurredAt.IsZero() {
		return domain.OutboxEvent{}, ErrInvalidEvent
	}
	if !isRequiredEventType(eventType) || routingKeyForEventType(eventType) != routingKey {
		return domain.OutboxEvent{}, ErrInvalidEvent
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return domain.OutboxEvent{}, fmt.Errorf("marshal order event payload: %w", err)
	}
	if !json.Valid(payloadBytes) || !payloadHasRequiredBusinessFields(payloadBytes) {
		return domain.OutboxEvent{}, ErrInvalidEvent
	}

	occurredAt = occurredAt.UTC()
	envelopeBytes, err := json.Marshal(Envelope{
		EventID:       eventID,
		EventType:     eventType,
		Version:       EventVersion,
		OccurredAt:    occurredAt,
		Producer:      ProducerOrderService,
		TraceID:       traceID,
		AggregateType: AggregateTypeOrder,
		AggregateID:   orderID,
		Payload:       payloadBytes,
	})
	if err != nil {
		return domain.OutboxEvent{}, fmt.Errorf("marshal order event envelope: %w", err)
	}

	event := domain.OutboxEvent{
		EventID:          eventID,
		EventType:        eventType,
		Version:          EventVersion,
		AggregateType:    AggregateTypeOrder,
		AggregateID:      orderID,
		RoutingKey:       routingKey,
		DeduplicationKey: deduplicationKey,
		SourceHistoryID:  sourceHistoryID,
		Payload:          envelopeBytes,
		TraceID:          traceID,
		Status:           domain.OutboxStatusPending,
		CreatedAt:        occurredAt,
		UpdatedAt:        occurredAt,
	}
	if err := event.ValidateForInsert(); err != nil {
		return domain.OutboxEvent{}, err
	}
	return event, nil
}

func payloadHasRequiredBusinessFields(payload []byte) bool {
	var fields map[string]any
	if err := json.Unmarshal(payload, &fields); err != nil {
		return false
	}
	orderID, _ := fields["order_id"].(string)
	userID, _ := fields["user_id"].(string)
	status, _ := fields["status"].(string)
	if strings.TrimSpace(orderID) == "" || strings.TrimSpace(userID) == "" || strings.TrimSpace(status) == "" {
		return false
	}
	if amount, exists := fields["total_amount"]; exists {
		value, ok := amount.(float64)
		if !ok || value < 0 {
			return false
		}
	}
	return true
}

func isRequiredEventType(eventType string) bool {
	switch eventType {
	case EventTypeOrderCreated,
		EventTypeOrderPaid,
		EventTypeOrderCancelled,
		EventTypeOrderDelivered:
		return true
	default:
		return false
	}
}

func routingKeyForEventType(eventType string) string {
	switch eventType {
	case EventTypeOrderCreated:
		return RoutingKeyOrderCreated
	case EventTypeOrderPaid:
		return RoutingKeyOrderPaid
	case EventTypeOrderCancelled:
		return RoutingKeyOrderCancelled
	case EventTypeOrderDelivered:
		return RoutingKeyOrderDelivered
	default:
		return ""
	}
}

func normalizeReasonCode(reasonCode string) string {
	reasonCode = strings.ToLower(strings.TrimSpace(reasonCode))
	if reasonCode == "" {
		return "unspecified"
	}
	reasonCode = strings.ReplaceAll(reasonCode, " ", "_")
	if len(reasonCode) > 80 {
		return reasonCode[:80]
	}
	return reasonCode
}
