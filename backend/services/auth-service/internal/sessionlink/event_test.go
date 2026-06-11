package sessionlink

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

func TestBuildLoginSucceededOutboxEvent(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	outboxEvent, err := BuildLoginSucceededOutboxEvent(LoginSucceededEvent{
		EventID:     "evt_test",
		TraceID:     "trace_123",
		OccurredAt:  now,
		AccountID:   "auth_123",
		UserID:      "user_123",
		SessionID:   "sess_123",
		AnonymousID: "anon_456",
		Roles:       []string{"buyer", "buyer"},
		Device: DeviceInfo{
			DeviceFingerprintHash: "device_hash",
			UserAgent:             "Mozilla/5.0",
			Channel:               "web",
			Locale:                "en-US",
		},
		Network: NetworkInfo{IPHash: "ip_hash"},
		Auth:    AuthInfo{Method: AuthMethodPassword},
	})
	if err != nil {
		t.Fatalf("BuildLoginSucceededOutboxEvent() error = %v", err)
	}

	if outboxEvent.EventType != EventTypeAuthLoginSucceeded {
		t.Fatalf("event type = %q", outboxEvent.EventType)
	}
	if outboxEvent.RoutingKey != RoutingKeyAuthLoginSucceeded {
		t.Fatalf("routing key = %q", outboxEvent.RoutingKey)
	}
	if outboxEvent.Status != domain.OutboxStatusPending {
		t.Fatalf("status = %q", outboxEvent.Status)
	}

	var envelope EventEnvelope
	if err := json.Unmarshal(outboxEvent.Payload, &envelope); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if envelope.EventID != "evt_test" || envelope.AggregateID != "sess_123" {
		t.Fatalf("envelope = %+v", envelope)
	}

	var payload LoginSucceededEvent
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.SessionID != "sess_123" || payload.UserID != "user_123" || payload.AnonymousID != "anon_456" {
		t.Fatalf("payload = %+v", payload)
	}
	if len(payload.Roles) != 1 || payload.Roles[0] != "buyer" {
		t.Fatalf("roles = %#v", payload.Roles)
	}
}

func TestLoginEventDoesNotContainSecretsOrRawNetworkValues(t *testing.T) {
	outboxEvent, err := BuildLoginSucceededOutboxEvent(LoginSucceededEvent{
		EventID:     "evt_test",
		OccurredAt:  time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC),
		AccountID:   "auth_123",
		UserID:      "user_123",
		SessionID:   "sess_123",
		AnonymousID: "anon_456",
		Roles:       []string{"buyer"},
		Device:      DeviceInfo{DeviceFingerprintHash: "hashed_device"},
		Network:     NetworkInfo{IPHash: "hashed_ip"},
		Auth:        AuthInfo{Method: AuthMethodPassword},
	})
	if err != nil {
		t.Fatalf("BuildLoginSucceededOutboxEvent() error = %v", err)
	}

	for _, forbidden := range [][]byte{
		[]byte("plain-password"),
		[]byte("refresh-token"),
		[]byte("access-token"),
		[]byte("123456"),
		[]byte("203.0.113.10"),
		[]byte("raw-fingerprint"),
	} {
		if bytes.Contains(outboxEvent.Payload, forbidden) {
			t.Fatalf("event payload contains forbidden value %q: %s", forbidden, outboxEvent.Payload)
		}
	}
}

func TestPrivacyHasherIsStableAndPeppered(t *testing.T) {
	first, err := NewPrivacyHasher("pepper-one")
	if err != nil {
		t.Fatalf("NewPrivacyHasher() error = %v", err)
	}
	second, err := NewPrivacyHasher("pepper-two")
	if err != nil {
		t.Fatalf("NewPrivacyHasher() error = %v", err)
	}

	a := first.HashIP("203.0.113.10")
	b := first.HashIP("203.0.113.10")
	c := second.HashIP("203.0.113.10")
	if a == "" || a != b {
		t.Fatalf("hash should be stable, got %q and %q", a, b)
	}
	if a == c {
		t.Fatal("different peppers produced the same hash")
	}
}

func TestOutboxLinkerInsertsLoginEvent(t *testing.T) {
	repo := &fakeOutboxRepository{}
	linker, err := NewOutboxLinker(repo, nil)
	if err != nil {
		t.Fatalf("NewOutboxLinker() error = %v", err)
	}

	err = linker.RecordLoginSucceeded(context.Background(), LoginSucceededEvent{
		EventID:    "evt_test",
		OccurredAt: time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC),
		AccountID:  "auth_123",
		UserID:     "user_123",
		SessionID:  "sess_123",
		Roles:      []string{"buyer"},
		Auth:       AuthInfo{Method: AuthMethodPassword},
	})
	if err != nil {
		t.Fatalf("RecordLoginSucceeded() error = %v", err)
	}
	if repo.event.EventID != "evt_test" || repo.event.EventType != EventTypeAuthLoginSucceeded {
		t.Fatalf("inserted event = %+v", repo.event)
	}
}

type fakeOutboxRepository struct {
	event domain.OutboxEvent
}

func (r *fakeOutboxRepository) InsertOutboxEvent(ctx context.Context, event domain.OutboxEvent) error {
	r.event = event
	return nil
}
