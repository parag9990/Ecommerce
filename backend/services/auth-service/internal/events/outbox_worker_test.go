package events

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

type outboxRepositoryStub struct {
	events           []domain.OutboxEvent
	lockLimit        int
	workerID         string
	publishedEventID string
	publishedAt      time.Time
	failedEventID    string
	nextAttemptAt    time.Time
	deadEventID      string
	lastError        string
}

func (r *outboxRepositoryStub) LockPendingOutboxEvents(_ context.Context, limit int, _ time.Time, _ time.Time, workerID string) ([]domain.OutboxEvent, error) {
	r.lockLimit = limit
	r.workerID = workerID
	return append([]domain.OutboxEvent(nil), r.events...), nil
}

func (r *outboxRepositoryStub) MarkOutboxEventPublished(_ context.Context, eventID string, publishedAt time.Time) error {
	r.publishedEventID = eventID
	r.publishedAt = publishedAt
	return nil
}

func (r *outboxRepositoryStub) MarkOutboxEventFailed(_ context.Context, eventID string, nextAttemptAt time.Time, lastError string) error {
	r.failedEventID = eventID
	r.nextAttemptAt = nextAttemptAt
	r.lastError = lastError
	return nil
}

func (r *outboxRepositoryStub) MarkOutboxEventDeadLetter(_ context.Context, eventID string, lastError string) error {
	r.deadEventID = eventID
	r.lastError = lastError
	return nil
}

type publisherStub struct {
	topic string
	event domain.OutboxEvent
	err   error
}

func (p *publisherStub) Publish(_ context.Context, topic string, event domain.OutboxEvent) error {
	p.topic = topic
	p.event = event
	return p.err
}

type outboxFixedClock struct{ now time.Time }

func (c outboxFixedClock) Now() time.Time { return c.now }

func TestOutboxWorkerPublishesAndMarksEvent(t *testing.T) {
	now := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)
	repo := &outboxRepositoryStub{events: []domain.OutboxEvent{{EventID: "evt_1", EventType: "AuthLoginSucceeded"}}}
	publisher := &publisherStub{}
	worker := newTestOutboxWorker(t, repo, publisher, 5, now)

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce() error = %v", err)
	}
	if publisher.topic != "auth.events" || publisher.event.EventID != "evt_1" {
		t.Fatalf("publish call = topic %q event %+v", publisher.topic, publisher.event)
	}
	if repo.publishedEventID != "evt_1" || !repo.publishedAt.Equal(now) {
		t.Fatalf("published mark = %q at %v", repo.publishedEventID, repo.publishedAt)
	}
}

func TestOutboxWorkerRetriesThenDeadLetters(t *testing.T) {
	now := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)
	publishErr := errors.New("event ingress unavailable")

	retryRepo := &outboxRepositoryStub{events: []domain.OutboxEvent{{EventID: "evt_retry", Attempts: 1}}}
	retryWorker := newTestOutboxWorker(t, retryRepo, &publisherStub{err: publishErr}, 5, now)
	if err := retryWorker.RunOnce(context.Background()); err != nil {
		t.Fatalf("retry RunOnce() error = %v", err)
	}
	if retryRepo.failedEventID != "evt_retry" || !retryRepo.nextAttemptAt.Equal(now.Add(10*time.Second)) {
		t.Fatalf("retry mark = event %q next %v", retryRepo.failedEventID, retryRepo.nextAttemptAt)
	}
	if retryRepo.deadEventID != "" {
		t.Fatalf("retry event was dead-lettered: %q", retryRepo.deadEventID)
	}

	deadRepo := &outboxRepositoryStub{events: []domain.OutboxEvent{{EventID: "evt_dead", Attempts: 4}}}
	deadWorker := newTestOutboxWorker(t, deadRepo, &publisherStub{err: publishErr}, 5, now)
	if err := deadWorker.RunOnce(context.Background()); err != nil {
		t.Fatalf("dead-letter RunOnce() error = %v", err)
	}
	if deadRepo.deadEventID != "evt_dead" || deadRepo.failedEventID != "" {
		t.Fatalf("dead-letter mark = dead %q failed %q", deadRepo.deadEventID, deadRepo.failedEventID)
	}
}

func newTestOutboxWorker(t *testing.T, repo OutboxRepository, publisher Publisher, maxAttempts int, now time.Time) *OutboxWorker {
	t.Helper()
	worker, err := NewOutboxWorker(repo, publisher, OutboxWorkerConfig{
		Topic:            "auth.events",
		WorkerID:         "worker_test",
		BatchSize:        100,
		Interval:         time.Second,
		MaxAttempts:      maxAttempts,
		InitialBackoff:   5 * time.Second,
		MaxBackoff:       time.Minute,
		StaleLockTimeout: 5 * time.Minute,
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("NewOutboxWorker() error = %v", err)
	}
	worker.WithClock(outboxFixedClock{now: now})
	return worker
}
