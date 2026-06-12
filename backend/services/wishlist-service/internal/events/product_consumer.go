package events

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/usecase"
)

type AvailabilitySyncUsecase interface {
	SyncProductAvailability(ctx context.Context, input usecase.SyncProductAvailabilityInput) (usecase.AvailabilitySyncResult, error)
}

type PriceChangeUsecase interface {
	HandleProductPriceChanged(ctx context.Context, input usecase.PriceChangeInput) (usecase.PriceDropResult, error)
}

type ProductEventMetrics interface {
	RecordConsumed(ctx context.Context, eventType, result string)
	RecordFailure(ctx context.Context, reason string)
	RecordAvailabilityUpdate(ctx context.Context, availability string, modifiedCount int64)
	RecordPriceDropCandidates(ctx context.Context, count int64)
	RecordPriceDropNotifications(ctx context.Context, result string, count int64)
	RecordPriceSnapshotUpdate(ctx context.Context, result string, modifiedCount int64)
	RecordProcessingDuration(ctx context.Context, eventType string, duration time.Duration)
	RecordEventLag(ctx context.Context, topic string, lag time.Duration)
}

type NoopProductEventMetrics struct{}

func (NoopProductEventMetrics) RecordConsumed(context.Context, string, string)                  {}
func (NoopProductEventMetrics) RecordFailure(context.Context, string)                           {}
func (NoopProductEventMetrics) RecordAvailabilityUpdate(context.Context, string, int64)         {}
func (NoopProductEventMetrics) RecordPriceDropCandidates(context.Context, int64)                {}
func (NoopProductEventMetrics) RecordPriceDropNotifications(context.Context, string, int64)     {}
func (NoopProductEventMetrics) RecordPriceSnapshotUpdate(context.Context, string, int64)        {}
func (NoopProductEventMetrics) RecordProcessingDuration(context.Context, string, time.Duration) {}
func (NoopProductEventMetrics) RecordEventLag(context.Context, string, time.Duration)           {}

type ProductMessage struct {
	Body       []byte
	Key        []byte
	Topic      string
	Partition  int
	Offset     int64
	Headers    map[string]string
	ReceivedAt time.Time
}

type ProductConsumer struct {
	sync      AvailabilitySyncUsecase
	priceDrop PriceChangeUsecase
	logger    *slog.Logger
	metrics   ProductEventMetrics
	clock     func() time.Time
}

type ProductConsumerOption func(*ProductConsumer)

func WithProductEventMetrics(metrics ProductEventMetrics) ProductConsumerOption {
	return func(c *ProductConsumer) {
		if metrics != nil {
			c.metrics = metrics
		}
	}
}

func WithProductConsumerClock(clock func() time.Time) ProductConsumerOption {
	return func(c *ProductConsumer) {
		if clock != nil {
			c.clock = clock
		}
	}
}

func WithPriceChangeUsecase(priceDrop PriceChangeUsecase) ProductConsumerOption {
	return func(c *ProductConsumer) {
		if priceDrop != nil {
			c.priceDrop = priceDrop
		}
	}
}

func NewProductConsumer(sync AvailabilitySyncUsecase, logger *slog.Logger, options ...ProductConsumerOption) (*ProductConsumer, error) {
	if sync == nil {
		return nil, errors.New("wishlist availability sync usecase is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	consumer := &ProductConsumer{
		sync:    sync,
		logger:  logger,
		metrics: NoopProductEventMetrics{},
		clock: func() time.Time {
			return time.Now().UTC()
		},
	}
	for _, option := range options {
		option(consumer)
	}
	return consumer, nil
}

func (c *ProductConsumer) HandleMessage(ctx context.Context, body []byte) error {
	return c.HandleMessageWithMetadata(ctx, ProductMessage{Body: body})
}

func (c *ProductConsumer) HandleMessageWithMetadata(ctx context.Context, message ProductMessage) error {
	if c == nil || c.sync == nil {
		return errors.New("product event consumer is not initialized")
	}
	ctx = contextOrBackground(ctx)
	startedAt := c.clock()

	input, ignored, err := DecodeProductAvailabilitySyncInput(message.Body)
	if !ignored && err == nil {
		return c.handleAvailability(ctx, message, input, startedAt)
	}
	if err != nil {
		reason := ProductEventErrorReason(err)
		c.logger.Warn("product event rejected",
			"topic", message.Topic,
			"partition", message.Partition,
			"offset", message.Offset,
			"reason", reason,
			"error", err,
		)
		c.metrics.RecordFailure(ctx, reason)
		c.metrics.RecordConsumed(ctx, "unknown", "failed")
		return err
	}

	if c.priceDrop == nil {
		c.logger.Debug("product event ignored by wishlist",
			"topic", message.Topic,
			"partition", message.Partition,
			"offset", message.Offset,
		)
		c.metrics.RecordConsumed(ctx, "unsupported", "ignored")
		return nil
	}

	priceInput, priceIgnored, err := DecodeProductPriceChangeInput(message.Body)
	if err != nil {
		reason := ProductEventErrorReason(err)
		c.logger.Warn("product event rejected",
			"topic", message.Topic,
			"partition", message.Partition,
			"offset", message.Offset,
			"reason", reason,
			"error", err,
		)
		c.metrics.RecordFailure(ctx, reason)
		c.metrics.RecordConsumed(ctx, "unknown", "failed")
		return err
	}
	if priceIgnored {
		c.logger.Debug("product event ignored by wishlist",
			"topic", message.Topic,
			"partition", message.Partition,
			"offset", message.Offset,
		)
		c.metrics.RecordConsumed(ctx, "unsupported", "ignored")
		return nil
	}

	return c.handlePriceChange(ctx, message, priceInput, startedAt)
}

func (c *ProductConsumer) handleAvailability(ctx context.Context, message ProductMessage, input usecase.SyncProductAvailabilityInput, startedAt time.Time) error {
	result, err := c.sync.SyncProductAvailability(ctx, input)
	if err != nil {
		c.logger.Warn("product event processing failed",
			"event_id", input.EventID,
			"event_type", input.EventType,
			"product_id", input.ProductID,
			"variant_id", input.VariantID,
			"trace_id", input.TraceID,
			"error", err,
		)
		c.metrics.RecordFailure(ctx, ProductEventErrorReason(err))
		c.metrics.RecordConsumed(ctx, input.EventType, "failed")
		return err
	}

	c.metrics.RecordConsumed(ctx, input.EventType, "processed")
	c.metrics.RecordAvailabilityUpdate(ctx, string(result.Availability), result.ModifiedCount)
	c.metrics.RecordProcessingDuration(ctx, input.EventType, c.clock().Sub(startedAt))
	if !input.OccurredAt.IsZero() && message.Topic != "" {
		c.metrics.RecordEventLag(ctx, message.Topic, c.clock().Sub(input.OccurredAt.UTC()))
	}
	return nil
}

func (c *ProductConsumer) handlePriceChange(ctx context.Context, message ProductMessage, input usecase.PriceChangeInput, startedAt time.Time) error {
	result, err := c.priceDrop.HandleProductPriceChanged(ctx, input)
	if err != nil {
		c.logger.Warn("product price event processing failed",
			"event_id", input.EventID,
			"event_type", input.EventType,
			"product_id", input.ProductID,
			"variant_id", input.VariantID,
			"trace_id", input.TraceID,
			"error", err,
		)
		c.metrics.RecordFailure(ctx, ProductEventErrorReason(err))
		c.metrics.RecordPriceDropNotifications(ctx, "failed", 0)
		c.metrics.RecordPriceSnapshotUpdate(ctx, "failed", 0)
		c.metrics.RecordConsumed(ctx, input.EventType, "failed")
		return err
	}

	c.metrics.RecordConsumed(ctx, input.EventType, "processed")
	c.metrics.RecordPriceDropCandidates(ctx, result.CandidateCount)
	c.metrics.RecordPriceDropNotifications(ctx, "published", result.NotificationCount)
	c.metrics.RecordPriceSnapshotUpdate(ctx, "updated", result.SnapshotModifiedCount)
	c.metrics.RecordProcessingDuration(ctx, input.EventType, c.clock().Sub(startedAt))
	if !input.OccurredAt.IsZero() && message.Topic != "" {
		c.metrics.RecordEventLag(ctx, message.Topic, c.clock().Sub(input.OccurredAt.UTC()))
	}
	return nil
}

func contextOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
