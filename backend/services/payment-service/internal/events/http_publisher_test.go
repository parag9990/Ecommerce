package events

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/domain"
)

func TestHTTPPublisherPostsIdempotentPaymentEvent(t *testing.T) {
	var got publishedEnvelope
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer publisher-token-at-least-32-characters" {
			t.Errorf("Authorization header not set")
		}
		if r.Header.Get("Idempotency-Key") != "whe_123" {
			t.Errorf("Idempotency-Key = %q, want whe_123", r.Header.Get("Idempotency-Key"))
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("Decode() error = %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	publisher, err := NewHTTPPublisher(HTTPPublisherConfig{
		Endpoint:  server.URL,
		AuthToken: "publisher-token-at-least-32-characters",
		Timeout:   time.Second,
	})
	if err != nil {
		t.Fatalf("NewHTTPPublisher() error = %v", err)
	}
	event := domain.PaymentDomainEvent{EventID: "whe_123", EventType: "PaymentCaptured", PaymentID: "pay_123"}
	if err := publisher.Publish(context.Background(), "payment.events", event); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if got.Topic != "payment.events" || got.Event.PaymentID != "pay_123" {
		t.Fatalf("published envelope = %+v, want payment event", got)
	}
}

func TestHTTPPublisherRejectsInsecureRemoteEndpoint(t *testing.T) {
	_, err := NewHTTPPublisher(HTTPPublisherConfig{
		Endpoint:  "http://events.example.test/payment",
		AuthToken: "publisher-token-at-least-32-characters",
	})
	if err == nil {
		t.Fatal("NewHTTPPublisher() error = nil, want insecure endpoint rejected")
	}
}

func TestHTTPPublisherPostsReconciliationAlertWithDeterministicKey(t *testing.T) {
	var got reconciliationAlertEnvelope
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Idempotency-Key") != "ral_123" {
			t.Errorf("Idempotency-Key = %q, want ral_123", r.Header.Get("Idempotency-Key"))
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("Decode() error = %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	publisher, err := NewHTTPPublisher(HTTPPublisherConfig{
		Endpoint:  server.URL,
		AuthToken: "publisher-token-at-least-32-characters",
		Timeout:   time.Second,
	})
	if err != nil {
		t.Fatalf("NewHTTPPublisher() error = %v", err)
	}
	alert := domain.ReconciliationAlert{
		EventID:          "ral_123",
		EventType:        "PaymentReconciliationMismatchDetected",
		ReconciliationID: "rec_123",
		Provider:         "stripe_like",
		SettlementID:     "stl_123",
		Status:           domain.ReconciliationStatusMismatch,
		ReasonCodes:      []string{"amount_mismatch"},
		Severity:         "critical",
		OccurredAt:       time.Now().UTC(),
	}
	if err := publisher.PublishMismatch(context.Background(), "payment.reconciliation.alerts", alert); err != nil {
		t.Fatalf("PublishMismatch() error = %v", err)
	}
	if got.Topic != "payment.reconciliation.alerts" || got.Event.ReconciliationID != "rec_123" {
		t.Fatalf("published envelope = %+v, want reconciliation event", got)
	}
}
