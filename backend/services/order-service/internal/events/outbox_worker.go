package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
)

type OutboxRepository interface {
	LockPendingOutboxEvents(ctx context.Context, limit int, now time.Time, staleBefore time.Time, workerID string) ([]domain.OutboxEvent, error)
	MarkOutboxEventPublished(ctx context.Context, eventID string, publishedAt time.Time) error
	MarkOutboxEventFailed(ctx context.Context, eventID string, nextAttemptAt time.Time, lastError string) error
	MarkOutboxEventDeadLetter(ctx context.Context, eventID string, lastError string) error
}

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now().UTC()
}

type OutboxWorkerConfig struct {
	Topic            string
	WorkerID         string
	BatchSize        int
	Interval         time.Duration
	MaxAttempts      int
	InitialBackoff   time.Duration
	MaxBackoff       time.Duration
	StaleLockTimeout time.Duration
}

type OutboxWorker struct {
	repo      OutboxRepository
	publisher Publisher
	cfg       OutboxWorkerConfig
	logger    *slog.Logger
	clock     Clock
}

func NewOutboxWorker(repo OutboxRepository, publisher Publisher, cfg OutboxWorkerConfig, logger *slog.Logger) (*OutboxWorker, error) {
	if repo == nil {
		return nil, errors.New("order outbox repository is required")
	}
	if publisher == nil {
		return nil, errors.New("order event publisher is required")
	}
	cfg.Topic = strings.TrimSpace(cfg.Topic)
	if cfg.Topic == "" {
		return nil, errors.New("order events topic is required")
	}
	if cfg.BatchSize <= 0 {
		return nil, errors.New("order outbox batch size must be greater than zero")
	}
	if cfg.Interval <= 0 {
		return nil, errors.New("order outbox interval must be greater than zero")
	}
	if cfg.MaxAttempts <= 0 {
		return nil, errors.New("order outbox max attempts must be greater than zero")
	}
	if cfg.InitialBackoff <= 0 {
		return nil, errors.New("order outbox initial backoff must be greater than zero")
	}
	if cfg.MaxBackoff < cfg.InitialBackoff {
		return nil, errors.New("order outbox max backoff must be greater than or equal to initial backoff")
	}
	if cfg.StaleLockTimeout <= 0 {
		return nil, errors.New("order outbox stale lock timeout must be greater than zero")
	}
	if strings.TrimSpace(cfg.WorkerID) == "" {
		host, _ := os.Hostname()
		if host == "" {
			host = "order-service"
		}
		cfg.WorkerID = fmt.Sprintf("%s-%d", host, time.Now().UnixNano())
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &OutboxWorker{
		repo:      repo,
		publisher: publisher,
		cfg:       cfg,
		logger:    logger,
		clock:     realClock{},
	}, nil
}

func (w *OutboxWorker) WithClock(clock Clock) {
	if clock != nil {
		w.clock = clock
	}
}

func (w *OutboxWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.cfg.Interval)
	defer ticker.Stop()

	for {
		if err := w.RunOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
			w.logger.ErrorContext(ctx, "order.outbox.run_failed",
				slog.String("error", err.Error()),
			)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (w *OutboxWorker) RunOnce(ctx context.Context) error {
	now := w.clock.Now().UTC()
	events, err := w.repo.LockPendingOutboxEvents(
		ctx,
		w.cfg.BatchSize,
		now,
		now.Add(-w.cfg.StaleLockTimeout),
		w.cfg.WorkerID,
	)
	if err != nil {
		return err
	}

	for _, event := range events {
		if err := w.publishOne(ctx, event); err != nil {
			w.logger.WarnContext(ctx, "order.outbox.publish_failed",
				slog.String("event_id", event.EventID),
				slog.String("event_type", event.EventType),
				slog.Int("attempts", event.Attempts+1),
				slog.String("error", err.Error()),
			)
			continue
		}
	}
	return nil
}

func (w *OutboxWorker) publishOne(ctx context.Context, event domain.OutboxEvent) error {
	err := w.publisher.Publish(ctx, w.cfg.Topic, event)
	if err == nil {
		publishedAt := w.clock.Now().UTC()
		if markErr := w.repo.MarkOutboxEventPublished(ctx, event.EventID, publishedAt); markErr != nil {
			return markErr
		}
		w.logger.InfoContext(ctx, "order.outbox.publish_succeeded",
			slog.String("event_id", event.EventID),
			slog.String("event_type", event.EventType),
			slog.String("topic", w.cfg.Topic),
			slog.Int("attempts", event.Attempts+1),
		)
		return nil
	}

	nextAttemptNumber := event.Attempts + 1
	if nextAttemptNumber >= w.cfg.MaxAttempts {
		if markErr := w.repo.MarkOutboxEventDeadLetter(ctx, event.EventID, err.Error()); markErr != nil {
			return markErr
		}
		w.logger.ErrorContext(ctx, "order.outbox.dead_lettered",
			slog.String("event_id", event.EventID),
			slog.String("event_type", event.EventType),
			slog.Int("attempts", nextAttemptNumber),
		)
		return err
	}

	nextAttemptAt := w.clock.Now().UTC().Add(w.backoff(nextAttemptNumber))
	if markErr := w.repo.MarkOutboxEventFailed(ctx, event.EventID, nextAttemptAt, err.Error()); markErr != nil {
		return markErr
	}
	return err
}

func (w *OutboxWorker) backoff(attempt int) time.Duration {
	if attempt <= 1 {
		return w.cfg.InitialBackoff
	}
	delay := w.cfg.InitialBackoff
	for i := 1; i < attempt; i++ {
		delay *= 2
		if delay >= w.cfg.MaxBackoff {
			return w.cfg.MaxBackoff
		}
	}
	return delay
}
