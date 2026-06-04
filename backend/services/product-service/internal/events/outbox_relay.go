package events

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"product-service/internal/repository"
)

const (
	defaultRelayPollInterval = time.Second
	defaultRelayBatchSize    = 100
	defaultRelayMaxAttempts  = 5
	defaultPublishTimeout    = 10 * time.Second
)

type OutboxRelayOptions struct {
	PollInterval   time.Duration
	BatchSize      int
	MaxAttempts    int
	PublishTimeout time.Duration
	RetryDelays    []time.Duration
}

type OutboxRelay struct {
	outbox    repository.ProductEventOutboxRepository
	publisher ProductEventPublisher
	clock     Clock
	logger    *slog.Logger
	options   OutboxRelayOptions
}

func NewOutboxRelay(
	outbox repository.ProductEventOutboxRepository,
	publisher ProductEventPublisher,
	clock Clock,
	logger *slog.Logger,
	options OutboxRelayOptions,
) (*OutboxRelay, error) {
	if outbox == nil {
		return nil, fmt.Errorf("product event outbox repository is required")
	}
	if publisher == nil {
		return nil, fmt.Errorf("product event publisher is required")
	}
	if clock == nil {
		clock = SystemClock{}
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &OutboxRelay{
		outbox:    outbox,
		publisher: publisher,
		clock:     clock,
		logger:    logger,
		options:   normalizeOutboxRelayOptions(options),
	}, nil
}

func (r *OutboxRelay) Run(ctx context.Context) {
	ticker := time.NewTicker(r.options.PollInterval)
	defer ticker.Stop()

	for {
		if err := r.RunOnce(ctx); err != nil && ctx.Err() == nil {
			r.logger.Error("product outbox relay cycle failed", "error", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (r *OutboxRelay) RunOnce(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	now := r.clock.Now().UTC()
	items, err := r.outbox.ListPendingOutboxEvents(ctx, r.options.BatchSize, now)
	if err != nil {
		return fmt.Errorf("list product outbox events: %w", err)
	}

	var firstErr error
	for _, item := range items {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := r.outbox.MarkOutboxEventPublishing(ctx, item.ID); err != nil {
			r.logger.Debug("product outbox event claim skipped", "event_id", item.ID, "error", err)
			continue
		}
		if report := item.Validate(); report.HasErrors() {
			err := r.outbox.MarkOutboxEventDeadLettered(ctx, item.ID, "invalid product outbox event payload")
			if err != nil && firstErr == nil {
				firstErr = err
			}
			continue
		}

		attempt := item.Attempts + 1
		publishCtx := ctx
		cancel := func() {}
		if r.options.PublishTimeout > 0 {
			publishCtx, cancel = context.WithTimeout(ctx, r.options.PublishTimeout)
		}
		err := r.publisher.Publish(publishCtx, item.Topic, item.Envelope())
		cancel()
		if err == nil {
			if markErr := r.outbox.MarkOutboxEventPublished(ctx, item.ID, r.clock.Now().UTC()); markErr != nil {
				r.logger.Error("mark product outbox event published failed", "event_id", item.ID, "error", markErr)
				if firstErr == nil {
					firstErr = markErr
				}
				continue
			}
			r.logger.Info(
				"product event published",
				"event_id", item.ID,
				"event_type", item.EventType,
				"topic", item.Topic,
				"attempt", attempt,
			)
			continue
		}

		if attempt >= r.options.MaxAttempts {
			if markErr := r.outbox.MarkOutboxEventDeadLettered(ctx, item.ID, err.Error()); markErr != nil {
				r.logger.Error("dead-letter product outbox event failed", "event_id", item.ID, "error", markErr)
				if firstErr == nil {
					firstErr = markErr
				}
				continue
			}
			r.logger.Error(
				"product event dead-lettered",
				"event_id", item.ID,
				"event_type", item.EventType,
				"topic", item.Topic,
				"attempt", attempt,
				"error", err,
			)
			continue
		}

		nextAttemptAt := r.clock.Now().UTC().Add(r.retryDelay(attempt))
		if markErr := r.outbox.MarkOutboxEventFailed(ctx, item.ID, nextAttemptAt, err.Error()); markErr != nil {
			r.logger.Error("schedule product outbox retry failed", "event_id", item.ID, "error", markErr)
			if firstErr == nil {
				firstErr = markErr
			}
			continue
		}
		r.logger.Warn(
			"product event publish failed",
			"event_id", item.ID,
			"event_type", item.EventType,
			"topic", item.Topic,
			"attempt", attempt,
			"next_attempt_at", nextAttemptAt,
			"error", err,
		)
	}
	return firstErr
}

func (r *OutboxRelay) retryDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return r.options.RetryDelays[0]
	}
	index := attempt - 1
	if index >= len(r.options.RetryDelays) {
		index = len(r.options.RetryDelays) - 1
	}
	return r.options.RetryDelays[index]
}

func normalizeOutboxRelayOptions(options OutboxRelayOptions) OutboxRelayOptions {
	if options.PollInterval <= 0 {
		options.PollInterval = defaultRelayPollInterval
	}
	if options.BatchSize <= 0 {
		options.BatchSize = defaultRelayBatchSize
	}
	if options.MaxAttempts <= 0 {
		options.MaxAttempts = defaultRelayMaxAttempts
	}
	if options.PublishTimeout <= 0 {
		options.PublishTimeout = defaultPublishTimeout
	}
	if len(options.RetryDelays) == 0 {
		options.RetryDelays = []time.Duration{
			5 * time.Second,
			30 * time.Second,
			2 * time.Minute,
			10 * time.Minute,
		}
	}
	return options
}
