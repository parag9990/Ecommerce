package events

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/usecase"
)

func TestKafkaNotificationPublisherPublishesCommandEnvelope(t *testing.T) {
	writer := &fakeKafkaWriter{}
	publisher := newKafkaNotificationPublisherWithWriter(KafkaNotificationPublisherConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "notification.commands",
	}, writer, nil)

	occurredAt := time.Date(2026, 5, 26, 10, 30, 1, 0, time.UTC)
	err := publisher.Publish(context.Background(), usecase.NotificationCommand{
		EventID:        "notif_evt_123_user_123",
		UserID:         " user_123 ",
		TemplateKey:    "wishlist_price_drop",
		Channel:        "AUTO",
		IdempotencyKey: "wishlist_price_drop:user_123:prod_123:var_1:INR:299900",
		TraceID:        " trace_123 ",
		OccurredAt:     occurredAt,
		Variables: map[string]any{
			"product_id":       "prod_123",
			"old_price_amount": int64(349900),
			"new_price_amount": int64(299900),
		},
	})
	if err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}
	if len(writer.messages) != 1 {
		t.Fatalf("messages = %d, want 1", len(writer.messages))
	}
	message := writer.messages[0]
	if string(message.Key) != "user_123" {
		t.Fatalf("message key = %q, want user_123", message.Key)
	}

	var envelope NotificationCommandEnvelope
	if err := json.Unmarshal(message.Value, &envelope); err != nil {
		t.Fatalf("notification command JSON invalid: %v", err)
	}
	if envelope.EventType != NotificationCommandEventType || envelope.Producer != NotificationCommandProducer {
		t.Fatalf("envelope type/producer = %q/%q", envelope.EventType, envelope.Producer)
	}
	if envelope.Payload.UserID != "user_123" || envelope.Payload.Channel != "auto" {
		t.Fatalf("payload user/channel = %q/%q, want user_123/auto", envelope.Payload.UserID, envelope.Payload.Channel)
	}
	if envelope.Payload.IdempotencyKey != "wishlist_price_drop:user_123:prod_123:var_1:INR:299900" {
		t.Fatalf("idempotency key = %q", envelope.Payload.IdempotencyKey)
	}
}

func TestKafkaNotificationPublisherValidatesCommand(t *testing.T) {
	publisher := newKafkaNotificationPublisherWithWriter(KafkaNotificationPublisherConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "notification.commands",
	}, &fakeKafkaWriter{}, nil)

	err := publisher.Publish(context.Background(), usecase.NotificationCommand{
		EventID: "notif_123",
		UserID:  "user_123",
		Channel: "auto",
	})
	if err == nil {
		t.Fatal("Publish error = nil, want validation error")
	}
}

func TestKafkaNotificationPublisherPropagatesWriterError(t *testing.T) {
	writerErr := errors.New("broker unavailable")
	publisher := newKafkaNotificationPublisherWithWriter(KafkaNotificationPublisherConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "notification.commands",
	}, &fakeKafkaWriter{err: writerErr}, nil)

	err := publisher.Publish(context.Background(), usecase.NotificationCommand{
		EventID:        "notif_123",
		UserID:         "user_123",
		TemplateKey:    "wishlist_price_drop",
		Channel:        "auto",
		IdempotencyKey: "idem_123",
	})
	if !errors.Is(err, writerErr) {
		t.Fatalf("Publish error = %v, want writer error", err)
	}
}
