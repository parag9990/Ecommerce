package orderevents

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
)

func TestBuildOrderCreatedEventUsesStableEnvelopeAndSafePayload(t *testing.T) {
	order := validEventOrder()
	event, err := BuildOrderCreatedEvent("evt_created", "trace_1", order.InitialHistory.CreatedAt, order)
	if err != nil {
		t.Fatalf("BuildOrderCreatedEvent() error = %v", err)
	}
	if event.EventType != EventTypeOrderCreated || event.RoutingKey != RoutingKeyOrderCreated ||
		event.AggregateID != "ord_1" || event.DeduplicationKey != "ord_1:created" ||
		event.SourceHistoryID != "osh_created" {
		t.Fatalf("outbox event = %+v, want created event metadata", event)
	}

	var envelope Envelope
	if err := json.Unmarshal(event.Payload, &envelope); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if envelope.EventID != "evt_created" || envelope.EventType != EventTypeOrderCreated ||
		envelope.Version != EventVersion || envelope.Producer != ProducerOrderService ||
		envelope.AggregateType != AggregateTypeOrder || envelope.AggregateID != "ord_1" ||
		envelope.TraceID != "trace_1" {
		t.Fatalf("envelope = %+v, want stable order envelope", envelope)
	}

	var payload map[string]any
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload["order_id"] != "ord_1" || payload["user_id"] != "user_1" ||
		payload["status"] != "created" || payload["total_amount"].(float64) != 299900 {
		t.Fatalf("payload = %+v, want safe created fields", payload)
	}
	if _, exists := payload["address_snapshot"]; exists {
		t.Fatal("payload exposed address_snapshot")
	}
	if _, exists := payload["payment_token"]; exists {
		t.Fatal("payload exposed a payment token field")
	}
}

func TestBuildOrderPaidEventRejectsInvalidAmount(t *testing.T) {
	order := validEventOrder()
	order.Status = domain.OrderStatusPendingPayment
	order.PaymentID = "pay_1"
	order.TotalAmount = -1
	_, err := BuildOrderPaidEvent("evt_paid", "", time.Now(), order, "osh_paid")
	if !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("BuildOrderPaidEvent() error = %v, want ErrInvalidEvent", err)
	}
}

func TestEventTypeForStatusOnlyMapsTaskEightEvents(t *testing.T) {
	tests := map[domain.OrderStatus]struct {
		want string
		ok   bool
	}{
		domain.OrderStatusPaid:           {want: EventTypeOrderPaid, ok: true},
		domain.OrderStatusCancelled:      {want: EventTypeOrderCancelled, ok: true},
		domain.OrderStatusDelivered:      {want: EventTypeOrderDelivered, ok: true},
		domain.OrderStatusPacked:         {},
		domain.OrderStatusShipped:        {},
		domain.OrderStatusPaymentFailed:  {},
		domain.OrderStatusPendingPayment: {},
	}
	for status, test := range tests {
		got, ok := EventTypeForStatus(status)
		if got != test.want || ok != test.ok {
			t.Fatalf("EventTypeForStatus(%s) = %q/%t, want %q/%t", status, got, ok, test.want, test.ok)
		}
	}
}

func validEventOrder() domain.Order {
	now := time.Date(2026, time.May, 26, 10, 0, 0, 0, time.UTC)
	return domain.Order{
		OrderID: "ord_1", UserID: "user_1", CartID: "cart_1", Status: domain.OrderStatusCreated,
		Currency: "INR", SubtotalAmount: 299900, TotalAmount: 299900,
		Items: []domain.OrderItem{{
			OrderItemID: "oi_1", OrderID: "ord_1", SellerID: "seller_1", ProductID: "product_1",
			VariantID: "variant_1", Quantity: 1, Currency: "INR", UnitAmount: 299900,
			TotalAmount: 299900, FulfillmentStatus: domain.FulfillmentStatusPending,
		}},
		InitialHistory: domain.OrderStatusHistoryEntry{
			ID: "osh_created", OrderID: "ord_1", ToStatus: domain.OrderStatusCreated,
			Reason: "cart_to_order_created", ActorType: domain.OrderStatusActorBuyer,
			ActorID: "user_1", CreatedAt: now,
		},
	}
}
