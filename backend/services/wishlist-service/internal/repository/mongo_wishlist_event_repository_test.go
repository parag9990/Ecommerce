package repository

import (
	"testing"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/domain"
)

func TestWishlistEventCollectionValidatorMatchesTask8Contract(t *testing.T) {
	validator := WishlistEventCollectionValidator()
	schema := mustDocument(t, documentValue(t, validator, "$jsonSchema"))
	required := mustArray(t, documentValue(t, schema, "required"))
	for _, field := range []any{"_id", "event_type", "version", "topic", "payload", "status", "attempts", "next_retry_at", "occurred_at", "created_at", "updated_at"} {
		assertArrayContains(t, required, field)
	}

	properties := mustDocument(t, documentValue(t, schema, "properties"))
	eventType := mustDocument(t, documentValue(t, properties, "event_type"))
	eventTypeEnum := mustArray(t, documentValue(t, eventType, "enum"))
	assertArrayContains(t, eventTypeEnum, string(domain.WishlistEventItemAdded))
	assertArrayContains(t, eventTypeEnum, string(domain.WishlistEventItemRemoved))

	payload := mustDocument(t, documentValue(t, properties, "payload"))
	payloadRequired := mustArray(t, documentValue(t, payload, "required"))
	assertArrayContains(t, payloadRequired, "user_id")
	assertArrayContains(t, payloadRequired, "product_id")
	assertArrayContains(t, payloadRequired, "action")
	assertArrayContains(t, payloadRequired, "source")

	payloadProperties := mustDocument(t, documentValue(t, payload, "properties"))
	action := mustDocument(t, documentValue(t, payloadProperties, "action"))
	actionEnum := mustArray(t, documentValue(t, action, "enum"))
	assertArrayContains(t, actionEnum, string(domain.WishlistEventActionAdd))
	assertArrayContains(t, actionEnum, string(domain.WishlistEventActionRemove))
}

func TestWishlistEventIndexModelsMatchTask8Contract(t *testing.T) {
	models := WishlistEventIndexModels()
	if len(models) != 3 {
		t.Fatalf("len(WishlistEventIndexModels()) = %d, want 3", len(models))
	}

	pendingKeys := mustDocument(t, models[0].Keys)
	if got := documentValue(t, pendingKeys, "status"); got != 1 {
		t.Fatalf("pending index status key = %v, want 1", got)
	}
	if got := documentValue(t, pendingKeys, "next_retry_at"); got != 1 {
		t.Fatalf("pending index next_retry_at key = %v, want 1", got)
	}
	pendingOptions := materializeIndexOptions(t, models[0].Options)
	if pendingOptions.Name == nil || *pendingOptions.Name != WishlistEventPendingRetryIndexName {
		t.Fatalf("pending index name = %v, want %s", pendingOptions.Name, WishlistEventPendingRetryIndexName)
	}

	typeTimeKeys := mustDocument(t, models[1].Keys)
	if got := documentValue(t, typeTimeKeys, "event_type"); got != 1 {
		t.Fatalf("type index event_type key = %v, want 1", got)
	}
	if got := documentValue(t, typeTimeKeys, "occurred_at"); got != -1 {
		t.Fatalf("type index occurred_at key = %v, want -1", got)
	}

	ttlOptions := materializeIndexOptions(t, models[2].Options)
	if ttlOptions.Name == nil || *ttlOptions.Name != WishlistEventPublishedTTLIndexName {
		t.Fatalf("ttl index name = %v, want %s", ttlOptions.Name, WishlistEventPublishedTTLIndexName)
	}
	if ttlOptions.ExpireAfterSeconds == nil || *ttlOptions.ExpireAfterSeconds != publishedWishlistEventTTLSeconds {
		t.Fatalf("ttl seconds = %v, want %d", ttlOptions.ExpireAfterSeconds, publishedWishlistEventTTLSeconds)
	}
}

func TestWishlistEventDocumentRoundTrip(t *testing.T) {
	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	event := domain.WishlistAnalyticsEvent{
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
		Status:      domain.WishlistEventPending,
		NextRetryAt: now,
		OccurredAt:  now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	document, err := NewWishlistEventDocument(event)
	if err != nil {
		t.Fatalf("NewWishlistEventDocument returned error: %v", err)
	}
	if document.ID != "evt_wish_123" || document.Payload.LastKnownPrice.Amount != 299900 {
		t.Fatalf("document = %#v, want event id and price snapshot", document)
	}

	roundTrip, err := document.ToDomain()
	if err != nil {
		t.Fatalf("ToDomain returned error: %v", err)
	}
	if roundTrip.EventID != event.EventID || roundTrip.Payload.UserID != "user_123" || roundTrip.Payload.LastKnownPrice.Amount != 299900 {
		t.Fatalf("round-trip event = %#v, want original payload", roundTrip)
	}
}
