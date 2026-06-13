package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
)

func TestIngestEventStoresSanitizedEventAndTouchesSession(t *testing.T) {
	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	events := &fakeEventRepository{}
	sessions := &fakeSessionRepository{}
	active := &fakeActiveSessionStore{}
	ingest, err := NewIngestUsecase(events, sessions, active, nil, IngestConfig{IPHashSalt: "test-salt"}, discardLogger())
	if err != nil {
		t.Fatalf("new ingest usecase: %v", err)
	}
	ingest.WithClock(fixedClock{now: now})
	ingest.WithIDGenerator(sequenceIDGenerator{})

	path := "/products/prod_123"
	bodyUserID := "user_body"
	out, err := ingest.IngestEvent(context.Background(), IngestEventInput{
		EventType:   string(domain.EventClick),
		AnonymousID: "anon_123",
		SessionID:   "sess_123",
		UserID:      &bodyUserID,
		OccurredAt:  now.Add(-time.Second),
		Path:        &path,
		Properties: map[string]any{
			"element_id": " add-to-cart ",
			"password":   "secret",
			"nested": map[string]any{
				"authorization": "Bearer secret",
				"safe":          " yes ",
			},
		},
		RequestID: "req_existing",
		UserAgent: "Mozilla/5.0",
		IPAddress: "203.0.113.10",
		Channel:   domain.ChannelUserAppWeb,
	})
	if err != nil {
		t.Fatalf("ingest event: %v", err)
	}
	if !out.Accepted || out.RequestID != "req_existing" || out.EventID != "evt_000001" {
		t.Fatalf("unexpected output: %+v", out)
	}
	if len(events.inserted) != 1 {
		t.Fatalf("expected one inserted event, got %d", len(events.inserted))
	}
	event := events.inserted[0]
	if _, ok := event.Properties["password"]; ok {
		t.Fatal("expected sensitive password property to be removed")
	}
	nested := event.Properties["nested"].(map[string]any)
	if _, ok := nested["authorization"]; ok {
		t.Fatal("expected nested authorization property to be removed")
	}
	if nested["safe"] != "yes" {
		t.Fatalf("expected nested safe value to be trimmed, got %#v", nested["safe"])
	}
	if len(sessions.touched) != 1 {
		t.Fatalf("expected session touch, got %d", len(sessions.touched))
	}
	touch := sessions.touched[0]
	if touch.IPHash == "203.0.113.10" || touch.IPHash == unavailableIPHash {
		t.Fatalf("expected hashed IP, got %q", touch.IPHash)
	}
	if active.touchedSessionID != "sess_123" {
		t.Fatalf("expected active session touch, got %q", active.touchedSessionID)
	}
}

func TestIngestEventRejectsUnsupportedEventType(t *testing.T) {
	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	ingest, err := NewIngestUsecase(&fakeEventRepository{}, &fakeSessionRepository{}, &fakeActiveSessionStore{}, nil, IngestConfig{}, discardLogger())
	if err != nil {
		t.Fatalf("new ingest usecase: %v", err)
	}
	ingest.WithClock(fixedClock{now: now})

	_, err = ingest.IngestEvent(context.Background(), IngestEventInput{
		EventType:   "remove_from_cart",
		AnonymousID: "anon_123",
		SessionID:   "sess_123",
		OccurredAt:  now,
		UserAgent:   "Mozilla/5.0",
	})
	if !errors.Is(err, ErrInvalidSessionInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestIngestEventToleratesActiveSessionFailure(t *testing.T) {
	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	events := &fakeEventRepository{}
	active := &fakeActiveSessionStore{touchErr: errors.New("redis unavailable")}
	ingest, err := NewIngestUsecase(events, &fakeSessionRepository{}, active, nil, IngestConfig{}, discardLogger())
	if err != nil {
		t.Fatalf("new ingest usecase: %v", err)
	}
	ingest.WithClock(fixedClock{now: now})
	ingest.WithIDGenerator(sequenceIDGenerator{})

	out, err := ingest.IngestEvent(context.Background(), validIngestInput(now))
	if err != nil {
		t.Fatalf("redis failure should not reject event, got %v", err)
	}
	if !out.Accepted || len(events.inserted) != 1 {
		t.Fatalf("expected accepted durable write, output=%+v inserted=%d", out, len(events.inserted))
	}
}

func TestIngestEventFailsWhenDurableInsertFails(t *testing.T) {
	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	events := &fakeEventRepository{insertErr: errors.New("mongo unavailable")}
	sessions := &fakeSessionRepository{}
	ingest, err := NewIngestUsecase(events, sessions, &fakeActiveSessionStore{}, nil, IngestConfig{}, discardLogger())
	if err != nil {
		t.Fatalf("new ingest usecase: %v", err)
	}
	ingest.WithClock(fixedClock{now: now})
	ingest.WithIDGenerator(sequenceIDGenerator{})

	_, err = ingest.IngestEvent(context.Background(), validIngestInput(now))
	if !errors.Is(err, ErrIngestStorageUnavailable) {
		t.Fatalf("expected storage unavailable error, got %v", err)
	}
	if len(sessions.touched) != 0 {
		t.Fatalf("session should not be touched after failed event insert")
	}
}

func validIngestInput(now time.Time) IngestEventInput {
	path := "/"
	return IngestEventInput{
		EventType:   string(domain.EventPageView),
		AnonymousID: "anon_123",
		SessionID:   "sess_123",
		OccurredAt:  now,
		Path:        &path,
		UserAgent:   "Mozilla/5.0",
	}
}

type sequenceIDGenerator struct {
	next int
}

func (g sequenceIDGenerator) NewID(prefix string) (string, error) {
	g.next++
	return prefix + "_000001", nil
}
