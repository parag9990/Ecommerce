package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestSessionEventValidateAcceptsFlexibleProperties(t *testing.T) {
	event := validSessionEvent()

	if err := event.Validate(); err != nil {
		t.Fatalf("expected valid event, got %v", err)
	}
}

func TestSessionEventValidateRejectsSensitivePropertyKeys(t *testing.T) {
	event := validSessionEvent()
	event.Properties["password"] = "secret"

	err := event.Validate()
	if !errors.Is(err, ErrInvalidSessionEvent) {
		t.Fatalf("expected invalid event error, got %v", err)
	}
	if !strings.Contains(err.Error(), "password") {
		t.Fatalf("expected sensitive property violation, got %v", err)
	}
}

func TestActiveSessionSnapshotFromSessionIsValid(t *testing.T) {
	session := validSession()
	currentPage := "/products/prod_123"

	snapshot := ActiveSessionSnapshotFromSession(session, &currentPage)
	if err := snapshot.Validate(); err != nil {
		t.Fatalf("expected active snapshot to be valid, got %v", err)
	}
	if snapshot.CurrentPage == nil || *snapshot.CurrentPage != currentPage {
		t.Fatalf("expected current page to be copied, got %#v", snapshot.CurrentPage)
	}
}

func validSessionEvent() SessionEvent {
	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	path := "/products/prod_123"
	return SessionEvent{
		EventID:       "evt_01HX9ZPHVZV7M8B8QX5Y6Z1K2A",
		SessionID:     "sess_01HX9ZK6N5M2V4B7S8T9Q0R1P2",
		AnonymousID:   "anon_01HX9ZJY35F7K2M4N6P8R0T1V2",
		EventType:     EventType("click"),
		Path:          &path,
		Properties:    map[string]any{"element_id": "add-to-cart", "x": 42.5},
		OccurredAt:    now,
		ReceivedAt:    now.Add(time.Second),
		SchemaVersion: CurrentSessionEventSchemaVersion,
	}
}
