package events

import (
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

func TestDecodeProductEnvelopeAndValidateMetadata(t *testing.T) {
	raw := []byte(`{
		"event_id":"evt_1",
		"event_type":"ProductPublished",
		"version":1,
		"producer":"product-service",
		"request_id":"req_1",
		"trace_id":"trace_1",
		"correlation_id":"prod_1",
		"occurred_at":"2026-05-23T10:30:00Z",
		"payload":{
			"product_id":"prod_1",
			"title":"Running Shoes",
			"category_ids":["cat_shoes"],
			"seller_id":"seller_1",
			"price":999,
			"popularity_score":10,
			"in_stock":true,
			"status":"published",
			"updated_at":"2026-05-23T10:29:00Z"
		}
	}`)

	envelope, err := DecodeProductEnvelope(raw)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if err := envelope.ValidateMetadata(); err != nil {
		t.Fatalf("validate metadata: %v", err)
	}
	if envelope.Payload.Price == nil || *envelope.Payload.Price != 999 {
		t.Fatalf("price not decoded: %#v", envelope.Payload.Price)
	}
}

func TestEnvelopeValidateRejectsUnsupportedEvent(t *testing.T) {
	envelope := Envelope[ProductIndexPayload]{
		EventID:       "evt_1",
		EventType:     "ProductArchived",
		Version:       SupportedProductEventVersion,
		Producer:      "product-service",
		RequestID:     "req_1",
		TraceID:       "trace_1",
		CorrelationID: "prod_1",
		OccurredAt:    time.Date(2026, 5, 23, 10, 30, 0, 0, time.UTC),
	}

	err := envelope.ValidateMetadata()
	if !errors.Is(err, domain.ErrUnsupportedProductEvent) {
		t.Fatalf("expected ErrUnsupportedProductEvent, got %v", err)
	}
	if !IsPermanentProductIndexError(err) {
		t.Fatalf("unsupported event should be permanent")
	}
}

func TestProductIndexPayloadSearchableRule(t *testing.T) {
	price := 100.0
	popularity := int32(1)
	inStock := false
	payload := ProductIndexPayload{
		ProductID:       " prod_1 ",
		Title:           " Shoes ",
		CategoryIDs:     []string{" cat_1 ", "", "cat_1"},
		SellerID:        " seller_1 ",
		Price:           &price,
		PopularityScore: &popularity,
		InStock:         &inStock,
		Status:          " Published ",
		UpdatedAt:       time.Date(2026, 5, 23, 10, 30, 0, 0, time.UTC),
	}

	normalized := payload.Normalized()
	if !normalized.IsSearchable() {
		t.Fatal("expected payload to be searchable")
	}
	if len(normalized.CategoryIDs) != 1 || normalized.CategoryIDs[0] != "cat_1" {
		t.Fatalf("category ids not normalized: %#v", normalized.CategoryIDs)
	}
	if err := normalized.ValidateForUpsert(); err != nil {
		t.Fatalf("validate upsert: %v", err)
	}
}

func TestProductIndexPayloadRequiresUpdatedAt(t *testing.T) {
	payload := ProductIndexPayload{ProductID: "prod_1"}
	err := payload.ValidateForRouting()
	if !errors.Is(err, domain.ErrInvalidProductEvent) {
		t.Fatalf("expected ErrInvalidProductEvent, got %v", err)
	}
}
