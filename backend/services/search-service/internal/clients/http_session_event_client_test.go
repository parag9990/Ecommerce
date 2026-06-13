package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

func TestHTTPSessionEventClientIngestSearchEvent(t *testing.T) {
	var received sessionEventRequest
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != defaultSessionEventIngestPath {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.Header.Get("X-Request-ID") != "req_123" {
			t.Fatalf("request id = %q", r.Header.Get("X-Request-ID"))
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		return jsonResponse(t, http.StatusAccepted, map[string]bool{"accepted": true}), nil
	})}

	client, err := NewHTTPSessionEventClient(HTTPSessionEventClientConfig{BaseURL: "http://session-service", Timeout: time.Second}, httpClient)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	event := domain.ZeroResultSearchEvent{
		Query:           "Laptop Bag",
		NormalizedQuery: "laptop bag",
		Filters:         map[string]string{domain.SearchFilterBrand: "Acme"},
		Sort:            domain.SortRelevance,
		Page:            1,
		PageSize:        20,
		RequestID:       "req_123",
		AnonymousID:     "anon_1",
		SessionID:       "sess_1",
		Path:            "/search",
		OccurredAt:      time.Date(2026, 5, 24, 10, 30, 0, 0, time.UTC),
	}
	if err := client.IngestSearchEvent(context.Background(), event); err != nil {
		t.Fatalf("ingest: %v", err)
	}

	if received.EventType != domain.ZeroResultEventTypeSearch || received.AnonymousID != "anon_1" || received.SessionID != "sess_1" {
		t.Fatalf("received event = %#v", received)
	}
	if received.Properties["zero_result"] != true || received.Properties["result_count"].(float64) != 0 {
		t.Fatalf("properties = %#v", received.Properties)
	}
}
