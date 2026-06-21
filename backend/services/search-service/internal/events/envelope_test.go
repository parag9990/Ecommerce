package events

import "testing"

func TestDecodeProductEnvelopeAcceptsProductServiceContract(t *testing.T) {
	raw := []byte(`{
		"event_id":"event_1",
		"event_type":"ProductPublished",
		"version":1,
		"source":"product-service",
		"producer":"product-service",
		"request_id":"request_1",
		"trace_id":"trace_1",
		"correlation_id":"request_1",
		"occurred_at":"2026-06-21T12:00:00Z",
		"payload":{
			"product_id":"product_1",
			"seller_id":"seller_1",
			"title":"Waterproof Phone",
			"category_id":"phones",
			"category_path":["electronics","phones"],
			"status":"published",
			"price":{"amount":249900,"currency":"INR"},
			"rating":4.5,
			"popularity_score":7,
			"in_stock":true,
			"updated_at":"2026-06-21T12:00:00Z"
		}
	}`)

	envelope, err := DecodeProductEnvelope(raw)
	if err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if err := envelope.ValidateMetadata(); err != nil {
		t.Fatalf("validate metadata: %v", err)
	}
	if err := envelope.Payload.ValidateForUpsert(); err != nil {
		t.Fatalf("validate payload: %v", err)
	}
	if envelope.Payload.Price == nil || *envelope.Payload.Price != 249900 {
		t.Fatalf("price = %#v", envelope.Payload.Price)
	}
	if len(envelope.Payload.CategoryIDs) != 2 || envelope.Payload.CategoryIDs[1] != "phones" {
		t.Fatalf("category ids = %#v", envelope.Payload.CategoryIDs)
	}
}

func TestDecodeProductEnvelopeUsesRequestIDAsLegacyCorrelationID(t *testing.T) {
	raw := []byte(`{
		"event_id":"event_1","event_type":"ProductUpdated","version":1,
		"source":"product-service","request_id":"request_1","trace_id":"trace_1",
		"occurred_at":"2026-06-21T12:00:00Z",
		"payload":{"product_id":"product_1","updated_at":"2026-06-21T12:00:00Z"}
	}`)

	envelope, err := DecodeProductEnvelope(raw)
	if err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.ProducerName() != "product-service" || envelope.CorrelationID != "request_1" {
		t.Fatalf("compatibility metadata = producer %q correlation %q", envelope.ProducerName(), envelope.CorrelationID)
	}
}
