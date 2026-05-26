package events

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
)

func TestOutboxWorkerPublishesAndMarksPublished(t *testing.T) {
	now := time.Date(2026, time.May, 26, 10, 0, 0, 0, time.UTC)
	repo := &fakeOutboxRepo{events: []domain.OutboxEvent{validOutboxEvent(0)}}
	publisher := &fakePublisher{}
	worker := newTestWorker(t, repo, publisher, now, 3)

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce() error = %v", err)
	}
	if publisher.calls != 1 || publisher.topic != "order.events" ||
		publisher.event.AggregateID != "ord_1" {
		t.Fatalf("publisher = %+v, want one publish using order aggregate", publisher)
	}
	if repo.publishedID != "evt_1" || !repo.publishedAt.Equal(now) {
		t.Fatalf("published = %q/%v, want evt_1 at fixed clock", repo.publishedID, repo.publishedAt)
	}
}

func TestOutboxWorkerRetriesTemporaryFailure(t *testing.T) {
	now := time.Date(2026, time.May, 26, 10, 0, 0, 0, time.UTC)
	repo := &fakeOutboxRepo{events: []domain.OutboxEvent{validOutboxEvent(1)}}
	publisher := &fakePublisher{err: errors.New("broker down")}
	worker := newTestWorker(t, repo, publisher, now, 5)

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce() error = %v", err)
	}
	if repo.failedID != "evt_1" || !repo.nextAttemptAt.Equal(now.Add(2*time.Second)) {
		t.Fatalf("failed = %q next=%v, want exponential retry", repo.failedID, repo.nextAttemptAt)
	}
	if repo.deadLetterID != "" {
		t.Fatalf("dead letter id = %q, want retry instead", repo.deadLetterID)
	}
}

func TestOutboxWorkerDeadLettersAfterMaxAttempts(t *testing.T) {
	now := time.Date(2026, time.May, 26, 10, 0, 0, 0, time.UTC)
	repo := &fakeOutboxRepo{events: []domain.OutboxEvent{validOutboxEvent(2)}}
	publisher := &fakePublisher{err: errors.New("invalid payload")}
	worker := newTestWorker(t, repo, publisher, now, 3)

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce() error = %v", err)
	}
	if repo.deadLetterID != "evt_1" || repo.failedID != "" {
		t.Fatalf("repo failed/dead = %q/%q, want dead letter only", repo.failedID, repo.deadLetterID)
	}
}

func newTestWorker(t *testing.T, repo OutboxRepository, publisher Publisher, now time.Time, maxAttempts int) *OutboxWorker {
	t.Helper()
	worker, err := NewOutboxWorker(repo, publisher, OutboxWorkerConfig{
		Topic:            "order.events",
		WorkerID:         "worker_1",
		BatchSize:        10,
		Interval:         time.Second,
		MaxAttempts:      maxAttempts,
		InitialBackoff:   time.Second,
		MaxBackoff:       10 * time.Second,
		StaleLockTimeout: time.Minute,
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("NewOutboxWorker() error = %v", err)
	}
	worker.WithClock(fixedWorkerClock{now: now})
	return worker
}

func validOutboxEvent(attempts int) domain.OutboxEvent {
	return domain.OutboxEvent{
		EventID: "evt_1", EventType: "OrderCreated", Version: 1,
		AggregateType: "order", AggregateID: "ord_1", RoutingKey: "order.created",
		DeduplicationKey: "ord_1:created", SourceHistoryID: "osh_1",
		Payload: []byte(`{"event_id":"evt_1"}`), Status: domain.OutboxStatusProcessing,
		Attempts: attempts, CreatedAt: time.Date(2026, time.May, 26, 9, 0, 0, 0, time.UTC),
	}
}

type fixedWorkerClock struct {
	now time.Time
}

func (c fixedWorkerClock) Now() time.Time {
	return c.now
}

type fakePublisher struct {
	err   error
	calls int
	topic string
	event domain.OutboxEvent
}

func (f *fakePublisher) Publish(_ context.Context, topic string, event domain.OutboxEvent) error {
	f.calls++
	f.topic = topic
	f.event = event
	return f.err
}

type fakeOutboxRepo struct {
	events        []domain.OutboxEvent
	publishedID   string
	publishedAt   time.Time
	failedID      string
	nextAttemptAt time.Time
	deadLetterID  string
}

func (f *fakeOutboxRepo) LockPendingOutboxEvents(context.Context, int, time.Time, time.Time, string) ([]domain.OutboxEvent, error) {
	return f.events, nil
}

func (f *fakeOutboxRepo) MarkOutboxEventPublished(_ context.Context, eventID string, publishedAt time.Time) error {
	f.publishedID = eventID
	f.publishedAt = publishedAt
	return nil
}

func (f *fakeOutboxRepo) MarkOutboxEventFailed(_ context.Context, eventID string, nextAttemptAt time.Time, _ string) error {
	f.failedID = eventID
	f.nextAttemptAt = nextAttemptAt
	return nil
}

func (f *fakeOutboxRepo) MarkOutboxEventDeadLetter(_ context.Context, eventID string, _ string) error {
	f.deadLetterID = eventID
	return nil
}
