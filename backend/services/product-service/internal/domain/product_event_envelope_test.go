package domain

import (
	"testing"
	"time"
)

func TestProductOutboxEnvelopePopulatesConsumerMetadata(t *testing.T) {
	event := ProductOutboxEvent{
		ID:         "event_1",
		EventType:  string(ProductEventPublished),
		Version:    ProductEventSchemaVersion,
		Source:     ProductEventDefaultSource,
		RequestID:  "request_1",
		TraceID:    "trace_1",
		OccurredAt: time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC),
		Payload:    map[string]any{"product_id": "product_1"},
	}

	envelope := event.Envelope()
	if envelope.Producer != ProductEventDefaultSource || envelope.CorrelationID != event.RequestID {
		t.Fatalf("envelope metadata = producer %q correlation %q", envelope.Producer, envelope.CorrelationID)
	}
}
