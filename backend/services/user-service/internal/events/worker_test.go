package events

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestNewEnvelopeGeneratesUniqueEventIDs(t *testing.T) {
	first, err := NewEnvelope("UserCreated", "user_123", "req_1", "trace_1", fixedEventTime(), map[string]string{"user_id": "user_123"})
	if err != nil {
		t.Fatalf("NewEnvelope returned error: %v", err)
	}
	second, err := NewEnvelope("UserCreated", "user_123", "req_1", "trace_1", fixedEventTime(), map[string]string{"user_id": "user_123"})
	if err != nil {
		t.Fatalf("NewEnvelope returned error: %v", err)
	}
	if first.EventID == second.EventID {
		t.Fatalf("event ids should be unique, got %q", first.EventID)
	}
	if first.Source != SourceUserService || first.Version != EventSchemaVersion1 {
		t.Fatalf("unexpected envelope metadata: %#v", first)
	}
}

func TestOutboxWorkerMarksPublishedOnSuccess(t *testing.T) {
	repo := &fakeOutboxRepository{
		locked: []OutboxEvent{{
			EventID:   "evt_123",
			EventType: "UserCreated",
			Topic:     "user.events",
			Payload:   []byte(`{"event_id":"evt_123"}`),
		}},
	}
	worker := newTestWorker(t, repo, &fakePublisher{})

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
	if repo.publishedEventID != "evt_123" {
		t.Fatalf("published event id = %q, want evt_123", repo.publishedEventID)
	}
}

func TestOutboxWorkerMarksDeadAfterMaxAttempts(t *testing.T) {
	repo := &fakeOutboxRepository{
		locked: []OutboxEvent{{
			EventID:   "evt_123",
			EventType: "UserCreated",
			Topic:     "user.events",
			Payload:   []byte(`{"event_id":"evt_123"}`),
			Attempts:  1,
		}},
	}
	publisher := &fakePublisher{err: errors.New("broker down")}
	worker := newTestWorker(t, repo, publisher)
	worker.cfg.MaxAttempts = 2
	worker.cfg.DeadLetterTopic = "user.events.dlq"

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
	if !repo.failure.Dead {
		t.Fatalf("expected dead failure, got %#v", repo.failure)
	}
	if publisher.publishedTopics[1] != "user.events.dlq" {
		t.Fatalf("expected DLQ publish, topics = %#v", publisher.publishedTopics)
	}
}

func newTestWorker(t *testing.T, repo *fakeOutboxRepository, publisher *fakePublisher) *OutboxWorker {
	t.Helper()
	worker, err := NewOutboxWorker(repo, publisher, WorkerConfig{
		Clock:  fixedEventClock{at: fixedEventTime()},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	if err != nil {
		t.Fatalf("NewOutboxWorker returned error: %v", err)
	}
	return worker
}

func fixedEventTime() time.Time {
	return time.Date(2026, 5, 21, 10, 30, 0, 0, time.UTC)
}

type fixedEventClock struct {
	at time.Time
}

func (c fixedEventClock) Now() time.Time {
	return c.at
}

type fakeOutboxRepository struct {
	locked           []OutboxEvent
	publishedEventID string
	failure          OutboxFailure
}

func (r *fakeOutboxRepository) Insert(context.Context, OutboxEvent) error {
	return nil
}

func (r *fakeOutboxRepository) LockPending(context.Context, int, time.Duration) ([]OutboxEvent, error) {
	return r.locked, nil
}

func (r *fakeOutboxRepository) MarkPublished(ctx context.Context, eventID string, publishedAt time.Time) error {
	r.publishedEventID = eventID
	return nil
}

func (r *fakeOutboxRepository) MarkFailed(ctx context.Context, eventID string, failure OutboxFailure) error {
	r.failure = failure
	return nil
}

type fakePublisher struct {
	err             error
	publishedTopics []string
}

func (p *fakePublisher) Publish(ctx context.Context, topic string, payload []byte) error {
	p.publishedTopics = append(p.publishedTopics, topic)
	return p.err
}
