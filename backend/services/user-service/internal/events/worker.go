package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

const (
	defaultWorkerBatchSize   = 100
	defaultWorkerPoll        = 2 * time.Second
	defaultWorkerLockTTL     = 30 * time.Second
	defaultWorkerMaxAttempts = 10
	defaultPublishTimeout    = 10 * time.Second
	maxFailureMessageLength  = 1024
)

type Publisher interface {
	Publish(ctx context.Context, topic string, payload []byte) error
}

type ClosePublisher interface {
	Close() error
}

type WorkerConfig struct {
	BatchSize       int
	PollInterval    time.Duration
	LockTTL         time.Duration
	MaxAttempts     int
	PublishTimeout  time.Duration
	DeadLetterTopic string
	Clock           Clock
	Logger          *slog.Logger
}

type OutboxWorker struct {
	repo      OutboxRepository
	publisher Publisher
	cfg       WorkerConfig
	clock     Clock
	logger    *slog.Logger
}

func NewOutboxWorker(repo OutboxRepository, publisher Publisher, cfg WorkerConfig) (*OutboxWorker, error) {
	if repo == nil {
		return nil, errors.New("outbox repository is required")
	}
	if publisher == nil {
		return nil, errors.New("event publisher is required")
	}
	normalized := normalizeWorkerConfig(cfg)
	return &OutboxWorker{
		repo:      repo,
		publisher: publisher,
		cfg:       normalized,
		clock:     normalized.Clock,
		logger:    normalized.Logger,
	}, nil
}

func (w *OutboxWorker) Run(ctx context.Context) error {
	if err := w.RunOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
		w.logger.ErrorContext(ctx, "user_outbox_worker_run_once_failed", slog.String("error_type", fmt.Sprintf("%T", err)))
	}

	ticker := time.NewTicker(w.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := w.RunOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
				w.logger.ErrorContext(ctx, "user_outbox_worker_run_once_failed", slog.String("error_type", fmt.Sprintf("%T", err)))
			}
		}
	}
}

func (w *OutboxWorker) RunOnce(ctx context.Context) error {
	rows, err := w.repo.LockPending(ctx, w.cfg.BatchSize, w.cfg.LockTTL)
	if err != nil {
		return fmt.Errorf("lock pending outbox events: %w", err)
	}
	for _, row := range rows {
		w.publishOne(ctx, row)
	}
	return nil
}

func (w *OutboxWorker) publishOne(ctx context.Context, row OutboxEvent) {
	publishCtx, cancel := context.WithTimeout(ctx, w.cfg.PublishTimeout)
	err := w.publisher.Publish(publishCtx, row.Topic, row.Payload)
	cancel()

	if err == nil {
		if markErr := w.repo.MarkPublished(ctx, row.EventID, w.clock.Now()); markErr != nil {
			w.logger.ErrorContext(ctx, "user_event_mark_published_failed",
				slog.String("event_id", row.EventID),
				slog.String("event_type", row.EventType),
				slog.String("error_type", fmt.Sprintf("%T", markErr)),
			)
			return
		}
		w.logger.InfoContext(ctx, "user_event_published",
			slog.String("event_id", row.EventID),
			slog.String("event_type", row.EventType),
			slog.String("aggregate_id", row.AggregateID),
		)
		return
	}

	nextAttempts := row.Attempts + 1
	dead := nextAttempts >= w.cfg.MaxAttempts
	now := w.clock.Now()
	nextAttempt := now.Add(nextBackoff(row.Attempts))
	if dead {
		nextAttempt = time.Time{}
		w.publishDeadLetter(ctx, row)
	}

	failure := OutboxFailure{
		LastError:     truncateFailure(err.Error()),
		NextAttemptAt: nextAttempt,
		Dead:          dead,
		FailedAt:      now,
	}
	if markErr := w.repo.MarkFailed(ctx, row.EventID, failure); markErr != nil {
		w.logger.ErrorContext(ctx, "user_event_mark_failed_failed",
			slog.String("event_id", row.EventID),
			slog.String("event_type", row.EventType),
			slog.String("error_type", fmt.Sprintf("%T", markErr)),
		)
		return
	}

	w.logger.WarnContext(ctx, "user_event_publish_failed",
		slog.String("event_id", row.EventID),
		slog.String("event_type", row.EventType),
		slog.String("aggregate_id", row.AggregateID),
		slog.Int("attempts", nextAttempts),
		slog.Bool("dead", dead),
		slog.String("error_type", fmt.Sprintf("%T", err)),
	)
}

func (w *OutboxWorker) publishDeadLetter(ctx context.Context, row OutboxEvent) {
	topic := strings.TrimSpace(w.cfg.DeadLetterTopic)
	if topic == "" {
		return
	}

	publishCtx, cancel := context.WithTimeout(ctx, w.cfg.PublishTimeout)
	err := w.publisher.Publish(publishCtx, topic, row.Payload)
	cancel()
	if err != nil {
		w.logger.ErrorContext(ctx, "user_event_dead_letter_publish_failed",
			slog.String("event_id", row.EventID),
			slog.String("event_type", row.EventType),
			slog.String("error_type", fmt.Sprintf("%T", err)),
		)
	}
}

func normalizeWorkerConfig(cfg WorkerConfig) WorkerConfig {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = defaultWorkerBatchSize
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = defaultWorkerPoll
	}
	if cfg.LockTTL <= 0 {
		cfg.LockTTL = defaultWorkerLockTTL
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = defaultWorkerMaxAttempts
	}
	if cfg.PublishTimeout <= 0 {
		cfg.PublishTimeout = defaultPublishTimeout
	}
	if cfg.Clock == nil {
		cfg.Clock = systemClock{}
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return cfg
}

func nextBackoff(attempts int) time.Duration {
	switch attempts {
	case 0:
		return 5 * time.Second
	case 1:
		return 30 * time.Second
	case 2:
		return 2 * time.Minute
	case 3:
		return 10 * time.Minute
	case 4:
		return 30 * time.Minute
	default:
		delay := time.Duration(1<<(min(attempts-4, 5))) * time.Hour
		if delay > 24*time.Hour {
			return 24 * time.Hour
		}
		return delay
	}
}

func truncateFailure(message string) string {
	message = strings.TrimSpace(message)
	if len(message) <= maxFailureMessageLength {
		return message
	}
	return message[:maxFailureMessageLength]
}
