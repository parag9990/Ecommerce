package events

import (
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
)

func TestInteractionMapperMapsProductViewed(t *testing.T) {
	mapper := newTestMapper(t)
	envelope := decodeTestEnvelope(t, `{
		"event_id":"evt_view_001",
		"event_type":"ProductViewed",
		"version":1,
		"occurred_at":"2026-05-27T10:30:00Z",
		"producer":"session-service",
		"trace_id":"trace_abc",
		"payload":{
			"user_id":"user_123",
			"anonymous_id":"anon_456",
			"session_id":"sess_789",
			"product_id":"prod_123",
			"variant_id":"var_1",
			"category_id":"cat_shoes",
			"seller_id":"seller_456",
			"page":"product_detail",
			"metadata":{"email":"user@example.com","placement":"hero"}
		}
	}`)

	interactions, err := mapper.Map(envelope, time.Date(2026, 5, 27, 10, 30, 1, 0, time.UTC))
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}
	if len(interactions) != 1 {
		t.Fatalf("interactions = %d, want 1", len(interactions))
	}
	got := interactions[0]
	if got.NormalizedEventType != domain.InteractionProductView {
		t.Fatalf("normalized type = %s, want %s", got.NormalizedEventType, domain.InteractionProductView)
	}
	if got.Weight != 1 || got.Quantity != 1 {
		t.Fatalf("weight/quantity = %d/%d, want 1/1", got.Weight, got.Quantity)
	}
	if got.DedupeKey != "evt_view_001#product_view#prod_123#var_1" {
		t.Fatalf("dedupe key = %s", got.DedupeKey)
	}
	if _, ok := got.Metadata["email"]; ok {
		t.Fatal("sensitive metadata email should be removed")
	}
	if got.Metadata["page"] != "product_detail" || got.Metadata["placement"] != "hero" {
		t.Fatalf("metadata = %#v", got.Metadata)
	}
}

func TestInteractionMapperMapsPurchaseItems(t *testing.T) {
	mapper := newTestMapper(t)
	envelope := decodeTestEnvelope(t, `{
		"event_id":"evt_order_001",
		"event_type":"OrderPaid",
		"version":1,
		"occurred_at":"2026-05-27T11:00:00Z",
		"producer":"order-service",
		"payload":{
			"user_id":"user_123",
			"order_id":"order_123",
			"items":[
				{"product_id":"prod_1","variant_id":"var_1","category_id":"cat_a","seller_id":"seller_a","quantity":2,"unit_price_amount":299900,"currency":"INR"},
				{"product_id":"prod_2","variant_id":"var_2","category_id":"cat_b","seller_id":"seller_b","quantity":1,"unit_price_amount":159900,"currency":"INR"}
			]
		}
	}`)

	interactions, err := mapper.Map(envelope, time.Date(2026, 5, 27, 11, 0, 1, 0, time.UTC))
	if err != nil {
		t.Fatalf("Map() error = %v", err)
	}
	if len(interactions) != 2 {
		t.Fatalf("interactions = %d, want 2", len(interactions))
	}
	for _, interaction := range interactions {
		if interaction.NormalizedEventType != domain.InteractionPurchase {
			t.Fatalf("normalized type = %s, want purchase", interaction.NormalizedEventType)
		}
		if interaction.Weight != 8 {
			t.Fatalf("weight = %d, want 8", interaction.Weight)
		}
		if interaction.Metadata["order_id"] != "order_123" {
			t.Fatalf("metadata = %#v", interaction.Metadata)
		}
	}
	if interactions[0].DedupeKey == interactions[1].DedupeKey {
		t.Fatal("purchase item dedupe keys should differ per product/variant")
	}
}

func TestInteractionMapperRejectsMissingProductID(t *testing.T) {
	mapper := newTestMapper(t)
	envelope := decodeTestEnvelope(t, `{
		"event_id":"evt_bad_001",
		"event_type":"CartItemAdded",
		"version":1,
		"occurred_at":"2026-05-27T10:30:00Z",
		"producer":"cart-service",
		"payload":{"user_id":"user_123","quantity":1}
	}`)

	_, err := mapper.Map(envelope, time.Now())
	if !errors.Is(err, domain.ErrInvalidInteraction) {
		t.Fatalf("Map() error = %v, want %v", err, domain.ErrInvalidInteraction)
	}
}

func newTestMapper(t *testing.T) *InteractionMapper {
	t.Helper()
	mapper, err := NewInteractionMapper(InteractionMapperConfig{})
	if err != nil {
		t.Fatalf("NewInteractionMapper() error = %v", err)
	}
	return mapper
}

func decodeTestEnvelope(t *testing.T, raw string) Envelope {
	t.Helper()
	envelope, err := DecodeEnvelope([]byte(raw))
	if err != nil {
		t.Fatalf("DecodeEnvelope() error = %v", err)
	}
	return envelope
}
