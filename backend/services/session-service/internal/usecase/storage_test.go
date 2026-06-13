package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
)

func TestStorageUsecaseInsertEventAppliesDefaults(t *testing.T) {
	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	events := &fakeEventRepository{}
	storage, err := NewStorageUsecase(&fakeSessionRepository{}, events, &fakeActiveSessionStore{}, StorageConfig{}, discardLogger())
	if err != nil {
		t.Fatalf("new storage usecase: %v", err)
	}
	storage.WithClock(fixedClock{now: now})

	event, err := storage.InsertEvent(context.Background(), domain.SessionEvent{
		EventID:     "evt_123",
		SessionID:   "sess_123",
		AnonymousID: "anon_123",
		EventType:   domain.EventType("page_view"),
		OccurredAt:  now.Add(-time.Second),
	})
	if err != nil {
		t.Fatalf("insert event: %v", err)
	}
	if event.SchemaVersion != domain.CurrentSessionEventSchemaVersion {
		t.Fatalf("expected schema default, got %d", event.SchemaVersion)
	}
	if !event.ReceivedAt.Equal(now) {
		t.Fatalf("expected received_at default %s, got %s", now, event.ReceivedAt)
	}
	if len(events.inserted) != 1 || events.inserted[0].EventID != "evt_123" {
		t.Fatalf("expected repository insert, got %+v", events.inserted)
	}
}

func TestStorageUsecaseTouchActiveSessionUsesConfiguredTTL(t *testing.T) {
	active := &fakeActiveSessionStore{}
	storage, err := NewStorageUsecase(&fakeSessionRepository{}, &fakeEventRepository{}, active, StorageConfig{ActiveSessionTTL: 42 * time.Minute}, discardLogger())
	if err != nil {
		t.Fatalf("new storage usecase: %v", err)
	}

	snapshot := validActiveSnapshot()
	if _, err := storage.TouchActiveSession(context.Background(), snapshot); err != nil {
		t.Fatalf("touch active session: %v", err)
	}
	if active.touchedTTL != 42*time.Minute {
		t.Fatalf("expected configured ttl, got %s", active.touchedTTL)
	}
}

func TestStorageUsecaseMarkSessionEndedDeletesRedisBestEffort(t *testing.T) {
	sessions := &fakeSessionRepository{}
	active := &fakeActiveSessionStore{deleteErr: errors.New("redis unavailable")}
	storage, err := NewStorageUsecase(sessions, &fakeEventRepository{}, active, StorageConfig{}, discardLogger())
	if err != nil {
		t.Fatalf("new storage usecase: %v", err)
	}

	err = storage.MarkSessionEnded(context.Background(), "sess_123", time.Date(2026, 5, 22, 10, 30, 0, 0, time.UTC), domain.EndReasonLogout)
	if err != nil {
		t.Fatalf("mark ended should tolerate active store delete failure, got %v", err)
	}
	if sessions.endedSessionID != "sess_123" {
		t.Fatalf("expected durable mark ended, got %q", sessions.endedSessionID)
	}
	if active.deletedSessionID != "sess_123" {
		t.Fatalf("expected active session delete attempt, got %q", active.deletedSessionID)
	}
}

type fakeSessionRepository struct {
	upserted       []domain.Session
	touched        []domain.SessionTouch
	endedSessionID string
}

func (r *fakeSessionRepository) UpsertSession(ctx context.Context, session domain.Session) error {
	r.upserted = append(r.upserted, session)
	return nil
}

func (r *fakeSessionRepository) TouchSession(ctx context.Context, touch domain.SessionTouch) error {
	r.touched = append(r.touched, touch)
	return nil
}

func (r *fakeSessionRepository) FindSessionByID(ctx context.Context, sessionID string) (domain.Session, error) {
	return domain.Session{SessionID: sessionID}, nil
}

func (r *fakeSessionRepository) ListUserSessions(ctx context.Context, userID string, limit int) ([]domain.Session, error) {
	return []domain.Session{{UserID: &userID}}, nil
}

func (r *fakeSessionRepository) MarkEnded(ctx context.Context, sessionID string, endedAt time.Time, reason domain.SessionEndReason) error {
	r.endedSessionID = sessionID
	return nil
}

type fakeEventRepository struct {
	inserted  []domain.SessionEvent
	insertErr error
}

func (r *fakeEventRepository) InsertEvent(ctx context.Context, event domain.SessionEvent) error {
	if r.insertErr != nil {
		return r.insertErr
	}
	r.inserted = append(r.inserted, event)
	return nil
}

func (r *fakeEventRepository) ListEventsBySession(ctx context.Context, sessionID string, limit int) ([]domain.SessionEvent, error) {
	return []domain.SessionEvent{{SessionID: sessionID}}, nil
}

type fakeActiveSessionStore struct {
	touchedSessionID string
	touchedTTL       time.Duration
	touchErr         error
	deletedSessionID string
	deleteErr        error
}

func (s *fakeActiveSessionStore) Touch(ctx context.Context, snapshot domain.ActiveSessionSnapshot, ttl time.Duration) error {
	s.touchedSessionID = snapshot.SessionID
	s.touchedTTL = ttl
	return s.touchErr
}

func (s *fakeActiveSessionStore) Get(ctx context.Context, sessionID string) (domain.ActiveSessionSnapshot, error) {
	snapshot := validActiveSnapshot()
	snapshot.SessionID = sessionID
	return snapshot, nil
}

func (s *fakeActiveSessionStore) Delete(ctx context.Context, sessionID string) error {
	s.deletedSessionID = sessionID
	return s.deleteErr
}

func (s *fakeActiveSessionStore) ListByUser(ctx context.Context, userID string) ([]string, error) {
	return []string{"sess_123"}, nil
}

func (s *fakeActiveSessionStore) ListByAnonymousID(ctx context.Context, anonymousID string) ([]string, error) {
	return []string{"sess_123"}, nil
}

func validActiveSnapshot() domain.ActiveSessionSnapshot {
	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	return domain.ActiveSessionSnapshot{
		SessionID:   "sess_123",
		AnonymousID: "anon_123",
		Status:      domain.SessionStatusActive,
		EntryPage:   "/",
		DeviceType:  domain.DeviceTypeUnknown,
		Channel:     domain.ChannelUnknown,
		IPHash:      "b94d27b9934d3e08a52e52d7da7dabfade64b6d5",
		StartedAt:   now,
		LastSeenAt:  now.Add(time.Minute),
	}
}
