package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/domain"
)

const (
	defaultWishlistOutboxBatchSize   = 50
	defaultWishlistOutboxMaxAttempts = 5
	defaultWishlistOutboxPollEvery   = 5 * time.Second
)

type WishlistOutboxRepository interface {
	ClaimPending(ctx context.Context, limit int, now time.Time) ([]domain.WishlistAnalyticsEvent, error)
	MarkPublished(ctx context.Context, eventID string, publishedAt time.Time) error
	MarkRetry(ctx context.Context, eventID string, attempts int, nextRetryAt time.Time, lastError string) error
	MarkFailed(ctx context.Context, eventID string, attempts int, lastError string, failedAt time.Time) error
}

type WishlistOutboxMetrics interface {
	RecordBatchSize(ctx context.Context, size int)
	RecordPublished(ctx context.Context, eventType string)
	RecordPublishFailure(ctx context.Context, eventType string)
	RecordPublishDuration(ctx context.Context, eventType string, duration time.Duration)
}

type NoopWishlistOutboxMetrics struct{}

func (NoopWishlistOutboxMetrics) RecordBatchSize(context.Context, int)                         {}
func (NoopWishlistOutboxMetrics) RecordPublished(context.Context, string)                      {}
func (NoopWishlistOutboxMetrics) RecordPublishFailure(context.Context, string)                 {}
func (NoopWishlistOutboxMetrics) RecordPublishDuration(context.Context, string, time.Duration) {}

type WishlistOutboxWorkerConfig struct {
	BatchSize   int
	MaxAttempts int
	PollEvery   time.Duration
}

type WishlistOutboxWorker struct {
	repository  WishlistOutboxRepository
	publisher   WishlistAnalyticsPublisher
	logger      *slog.Logger
	metrics     WishlistOutboxMetrics
	clock       func() time.Time
	batchSize   int
	maxAttempts int
	pollEvery   time.Duration
}

type WishlistOutboxWorkerOption func(*WishlistOutboxWorker)

func WithWishlistOutboxMetrics(metrics WishlistOutboxMetrics) WishlistOutboxWorkerOption {
	return func(w *WishlistOutboxWorker) {
		if metrics != nil {
			w.metrics = metrics
		}
	}
}

func WithWishlistOutboxClock(clock func() time.Time) WishlistOutboxWorkerOption {
	return func(w *WishlistOutboxWorker) {
		if clock != nil {
			w.clock = clock
		}
	}
}

func NewWishlistOutboxWorker(repository WishlistOutboxRepository, publisher WishlistAnalyticsPublisher, logger *slog.Logger, config WishlistOutboxWorkerConfig, options ...WishlistOutboxWorkerOption) (*WishlistOutboxWorker, error) {
	if repository == nil {
		return nil, errors.New("wishlist outbox repository is required")
	}
	if publisher == nil {
		return nil, errors.New("wishlist analytics publisher is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	if config.BatchSize <= 0 {
		config.BatchSize = defaultWishlistOutboxBatchSize
	}
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = defaultWishlistOutboxMaxAttempts
	}
	if config.PollEvery <= 0 {
		config.PollEvery = defaultWishlistOutboxPollEvery
	}

	worker := &WishlistOutboxWorker{
		repository:  repository,
		publisher:   publisher,
		logger:      logger,
		metrics:     NoopWishlistOutboxMetrics{},
		clock:       func() time.Time { return time.Now().UTC() },
		batchSize:   config.BatchSize,
		maxAttempts: config.MaxAttempts,
		pollEvery:   config.PollEvery,
	}
	for _, option := range options {
		option(worker)
	}
	return worker, nil
}

func (w *WishlistOutboxWorker) Run(ctx context.Context) error {
	if w == nil || w.repository == nil || w.publisher == nil {
		return errors.New("wishlist outbox worker is not initialized")
	}
	ctx = contextOrBackground(ctx)
	w.logger.Info("wishlist analytics outbox worker started",
		"batch_size", w.batchSize,
		"max_attempts", w.maxAttempts,
		"poll_every", w.pollEvery.String(),
	)

	ticker := time.NewTicker(w.pollEvery)
	defer ticker.Stop()

	for {
		if err := w.publishBatch(ctx); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			w.logger.Warn("wishlist analytics publish batch failed", "error", err)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (w *WishlistOutboxWorker) publishBatch(ctx context.Context) error {
	now := w.clock().UTC()
	events, err := w.repository.ClaimPending(ctx, w.batchSize, now)
	if err != nil {
		return err
	}
	w.metrics.RecordBatchSize(ctx, len(events))

	for _, event := range events {
		startedAt := w.clock().UTC()
		if err := w.publisher.Publish(ctx, event); err != nil {
			if markErr := w.markPublishFailure(ctx, event, err, startedAt); markErr != nil {
				return markErr
			}
			continue
		}

		publishedAt := w.clock().UTC()
		if err := w.repository.MarkPublished(ctx, event.EventID, publishedAt); err != nil {
			return fmt.Errorf("mark wishlist analytics event %q published: %w", event.EventID, err)
		}
		w.metrics.RecordPublished(ctx, string(event.EventType))
		w.metrics.RecordPublishDuration(ctx, string(event.EventType), publishedAt.Sub(startedAt))
		w.logger.Info("wishlist analytics event published",
			"event_id", event.EventID,
			"event_type", event.EventType,
			"topic", event.Topic,
			"attempts", event.Attempts,
		)
	}
	return nil
}

func (w *WishlistOutboxWorker) markPublishFailure(ctx context.Context, event domain.WishlistAnalyticsEvent, cause error, failedAt time.Time) error {
	attempts := event.Attempts + 1
	w.metrics.RecordPublishFailure(ctx, string(event.EventType))
	if attempts >= w.maxAttempts {
		if err := w.repository.MarkFailed(ctx, event.EventID, attempts, errorString(cause), failedAt); err != nil {
			return fmt.Errorf("mark wishlist analytics event %q failed: %w", event.EventID, err)
		}
		w.logger.Error("wishlist analytics event failed permanently",
			"event_id", event.EventID,
			"event_type", event.EventType,
			"attempts", attempts,
			"last_error", errorString(cause),
		)
		return nil
	}

	nextRetryAt := failedAt.Add(wishlistOutboxBackoff(attempts))
	if err := w.repository.MarkRetry(ctx, event.EventID, attempts, nextRetryAt, errorString(cause)); err != nil {
		return fmt.Errorf("schedule wishlist analytics event %q retry: %w", event.EventID, err)
	}
	w.logger.Warn("wishlist analytics event publish will retry",
		"event_id", event.EventID,
		"event_type", event.EventType,
		"attempts", attempts,
		"next_retry_at", nextRetryAt,
		"error", cause,
	)
	return nil
}

func wishlistOutboxBackoff(attempts int) time.Duration {
	switch {
	case attempts <= 1:
		return 5 * time.Second
	case attempts == 2:
		return 30 * time.Second
	case attempts == 3:
		return 2 * time.Minute
	default:
		return 10 * time.Minute
	}
}
