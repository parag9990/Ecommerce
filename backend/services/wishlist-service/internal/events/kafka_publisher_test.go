package events

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/domain"
)

func TestKafkaWishlistAnalyticsPublisherPublishesRecommendationEnvelope(t *testing.T) {
	writer := &fakeKafkaWriter{}
	publisher := newKafkaWishlistAnalyticsPublisherWithWriter(KafkaWishlistAnalyticsPublisherConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "recommendation.events",
	}, writer, nil)

	event := testWishlistAnalyticsEvent(time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC))
	if err := publisher.Publish(context.Background(), event); err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}
	if len(writer.messages) != 1 {
		t.Fatalf("messages = %d, want 1", len(writer.messages))
	}
	message := writer.messages[0]
	if string(message.Key) != "user_123" {
		t.Fatalf("message key = %q, want user_123", message.Key)
	}

	var envelope WishlistAnalyticsEventEnvelope
	if err := json.Unmarshal(message.Value, &envelope); err != nil {
		t.Fatalf("wishlist analytics JSON invalid: %v", err)
	}
	if envelope.EventID != "evt_wish_123" || envelope.EventType != domain.WishlistEventItemAdded {
		t.Fatalf("envelope id/type = %q/%q", envelope.EventID, envelope.EventType)
	}
	if envelope.Producer != WishlistAnalyticsProducer || envelope.Version != domain.WishlistEventVersion {
		t.Fatalf("producer/version = %q/%d", envelope.Producer, envelope.Version)
	}
	if envelope.Payload.UserID != "user_123" || envelope.Payload.ProductID != "prod_123" {
		t.Fatalf("payload = %#v, want user_123/prod_123", envelope.Payload)
	}
	if envelope.Payload.LastKnownPrice == nil || envelope.Payload.LastKnownPrice.Amount != 299900 {
		t.Fatalf("payload price = %#v, want amount 299900", envelope.Payload.LastKnownPrice)
	}
	if string(message.Headers[0].Value) != "evt_wish_123" {
		t.Fatalf("event_id header = %q, want evt_wish_123", message.Headers[0].Value)
	}
}

func testWishlistAnalyticsEvent(now time.Time) domain.WishlistAnalyticsEvent {
	return domain.WishlistAnalyticsEvent{
		EventID:   "evt_wish_123",
		EventType: domain.WishlistEventItemAdded,
		Version:   domain.WishlistEventVersion,
		Topic:     "recommendation.events",
		Payload: domain.WishlistAnalyticsPayload{
			UserID:       "user_123",
			ProductID:    "prod_123",
			VariantID:    "var_1",
			Action:       domain.WishlistEventActionAdd,
			Source:       domain.WishlistEventSource,
			Availability: domain.AvailabilityInStock,
			LastKnownPrice: &domain.Money{
				Amount:   299900,
				Currency: "INR",
			},
		},
		TraceID:     "req_123",
		Status:      domain.WishlistEventPublishing,
		Attempts:    1,
		NextRetryAt: now,
		OccurredAt:  now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
