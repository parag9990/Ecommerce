package events

import (
	"context"
	"errors"
	"testing"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/domain"
)

func TestWishlistOutboxWorkerPublishSuccessMarksPublished(t *testing.T) {
	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	repository := &fakeWishlistOutboxRepository{
		events: []domain.WishlistAnalyticsEvent{testWishlistAnalyticsEvent(now)},
	}
	publisher := &fakeWishlistAnalyticsPublisher{}
	worker := mustWishlistOutboxWorker(t, repository, publisher, now)

	if err := worker.publishBatch(context.Background()); err != nil {
		t.Fatalf("publishBatch returned error: %v", err)
	}
	if len(publisher.published) != 1 {
		t.Fatalf("published events = %d, want 1", len(publisher.published))
	}
	if repository.publishedEventID != "evt_wish_123" {
		t.Fatalf("published event id = %q, want evt_wish_123", repository.publishedEventID)
	}
}

func TestWishlistOutboxWorkerPublishFailureSchedulesRetry(t *testing.T) {
	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	repository := &fakeWishlistOutboxRepository{
		events: []domain.WishlistAnalyticsEvent{testWishlistAnalyticsEvent(now)},
	}
	publisher := &fakeWishlistAnalyticsPublisher{err: errors.New("broker unavailable")}
	worker := mustWishlistOutboxWorker(t, repository, publisher, now)

	if err := worker.publishBatch(context.Background()); err != nil {
		t.Fatalf("publishBatch returned error: %v", err)
	}
	if repository.retryEventID != "evt_wish_123" {
		t.Fatalf("retry event id = %q, want evt_wish_123", repository.retryEventID)
	}
	if repository.retryAttempts != 2 {
		t.Fatalf("retry attempts = %d, want 2", repository.retryAttempts)
	}
	if !repository.nextRetryAt.Equal(now.Add(30 * time.Second)) {
		t.Fatalf("next retry = %v, want %v", repository.nextRetryAt, now.Add(30*time.Second))
	}
}

func TestWishlistOutboxWorkerPublishFailureMarksFailedAtMaxAttempts(t *testing.T) {
	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	event := testWishlistAnalyticsEvent(now)
	event.Attempts = 4
	repository := &fakeWishlistOutboxRepository{events: []domain.WishlistAnalyticsEvent{event}}
	publisher := &fakeWishlistAnalyticsPublisher{err: errors.New("broker unavailable")}
	worker := mustWishlistOutboxWorker(t, repository, publisher, now)

	if err := worker.publishBatch(context.Background()); err != nil {
		t.Fatalf("publishBatch returned error: %v", err)
	}
	if repository.failedEventID != "evt_wish_123" {
		t.Fatalf("failed event id = %q, want evt_wish_123", repository.failedEventID)
	}
	if repository.failedAttempts != 5 {
		t.Fatalf("failed attempts = %d, want 5", repository.failedAttempts)
	}
}

func mustWishlistOutboxWorker(t *testing.T, repository WishlistOutboxRepository, publisher WishlistAnalyticsPublisher, now time.Time) *WishlistOutboxWorker {
	t.Helper()
	worker, err := NewWishlistOutboxWorker(repository, publisher, nil, WishlistOutboxWorkerConfig{
		BatchSize:   10,
		MaxAttempts: 5,
		PollEvery:   time.Second,
	}, WithWishlistOutboxClock(func() time.Time { return now }))
	if err != nil {
		t.Fatalf("NewWishlistOutboxWorker returned error: %v", err)
	}
	return worker
}

type fakeWishlistOutboxRepository struct {
	events           []domain.WishlistAnalyticsEvent
	publishedEventID string
	retryEventID     string
	retryAttempts    int
	nextRetryAt      time.Time
	failedEventID    string
	failedAttempts   int
}

func (f *fakeWishlistOutboxRepository) ClaimPending(ctx context.Context, limit int, now time.Time) ([]domain.WishlistAnalyticsEvent, error) {
	return f.events, nil
}

func (f *fakeWishlistOutboxRepository) MarkPublished(ctx context.Context, eventID string, publishedAt time.Time) error {
	f.publishedEventID = eventID
	return nil
}

func (f *fakeWishlistOutboxRepository) MarkRetry(ctx context.Context, eventID string, attempts int, nextRetryAt time.Time, lastError string) error {
	f.retryEventID = eventID
	f.retryAttempts = attempts
	f.nextRetryAt = nextRetryAt
	return nil
}

func (f *fakeWishlistOutboxRepository) MarkFailed(ctx context.Context, eventID string, attempts int, lastError string, failedAt time.Time) error {
	f.failedEventID = eventID
	f.failedAttempts = attempts
	return nil
}

type fakeWishlistAnalyticsPublisher struct {
	published []domain.WishlistAnalyticsEvent
	err       error
	closed    bool
}

func (f *fakeWishlistAnalyticsPublisher) Publish(ctx context.Context, event domain.WishlistAnalyticsEvent) error {
	if f.err != nil {
		return f.err
	}
	f.published = append(f.published, event)
	return nil
}

func (f *fakeWishlistAnalyticsPublisher) Close(ctx context.Context) error {
	f.closed = true
	return nil
}
