package events

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestEnvelopeJSONAndValidation(t *testing.T) {
	event := Envelope{ID: "evt-1", Type: "OrderCreated", Source: "order-service", Version: 1, OccurredAt: time.Unix(1, 0).UTC(), RequestID: "req-1", TraceID: "trace-1", Data: json.RawMessage(`{"order_id":"order-1"}`)}
	if err := event.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	encoded, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"event_id", "event_type", "producer", "request_id", "trace_id", "payload"} {
		if !strings.Contains(string(encoded), `"`+field+`"`) {
			t.Fatalf("event JSON %s missing %s", encoded, field)
		}
	}
}
