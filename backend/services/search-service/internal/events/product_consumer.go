package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

const retryHeader = "x-search-indexer-retry-count"

type ProductIndexer interface {
	Handle(ctx context.Context, envelope Envelope[ProductIndexPayload]) (ProductIndexResult, error)
}

type ProcessedEventStore interface {
	WasProcessed(ctx context.Context, eventID string) (bool, error)
	MarkProcessed(ctx context.Context, eventID string) error
}

type ProductConsumerMetricsRecorder interface {
	RecordProductEvent(ctx context.Context, eventType string, outcome string)
	ObserveProductEventLag(ctx context.Context, eventType string, lag time.Duration)
	ObserveProductIndexDuration(ctx context.Context, eventType string, outcome string, duration time.Duration)
}

type NopProductConsumerMetricsRecorder struct{}

func (NopProductConsumerMetricsRecorder) RecordProductEvent(context.Context, string, string) {}
func (NopProductConsumerMetricsRecorder) ObserveProductEventLag(context.Context, string, time.Duration) {
}
func (NopProductConsumerMetricsRecorder) ObserveProductIndexDuration(context.Context, string, string, time.Duration) {
}

type RabbitMQConsumerConfig struct {
	URL                string
	Exchange           string
	Queue              string
	DeadLetterExchange string
	DeadLetterQueue    string
	ConsumerTag        string
	RoutingKeys        []string
	Prefetch           int
	MaxRetries         int
	ReconnectDelay     time.Duration
	RetryBaseDelay     time.Duration
	RetryMaxDelay      time.Duration
	MessageTimeout     time.Duration
	Metrics            ProductConsumerMetricsRecorder
}

type RabbitMQProductConsumer struct {
	cfg        RabbitMQConsumerConfig
	indexer    ProductIndexer
	eventStore ProcessedEventStore
	logger     *slog.Logger
	connected  atomic.Bool
}

func NewRabbitMQProductConsumer(cfg RabbitMQConsumerConfig, indexer ProductIndexer, eventStore ProcessedEventStore, logger *slog.Logger) (*RabbitMQProductConsumer, error) {
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if indexer == nil {
		return nil, errors.New("product indexer is required")
	}
	if eventStore == nil {
		return nil, errors.New("processed event store is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	if cfg.Metrics == nil {
		cfg.Metrics = NopProductConsumerMetricsRecorder{}
	}
	return &RabbitMQProductConsumer{
		cfg:        cfg,
		indexer:    indexer,
		eventStore: eventStore,
		logger:     logger,
	}, nil
}

func (c *RabbitMQProductConsumer) Run(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return nil
		}
		err := c.runOnce(ctx)
		if err == nil || errors.Is(err, context.Canceled) {
			return nil
		}
		c.logger.ErrorContext(ctx, "search.product_consumer.disconnected",
			slog.String("queue", c.cfg.Queue),
			slog.String("error", err.Error()),
			slog.Duration("reconnect_delay", c.cfg.ReconnectDelay),
		)
		if !sleepContext(ctx, c.cfg.ReconnectDelay) {
			return nil
		}
	}
}

func (c *RabbitMQProductConsumer) runOnce(ctx context.Context) error {
	conn, err := amqp.Dial(c.cfg.URL)
	if err != nil {
		return fmt.Errorf("connect rabbitmq: %w", err)
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("open rabbitmq channel: %w", err)
	}
	defer channel.Close()

	if err := c.setupTopology(channel); err != nil {
		return err
	}
	if err := channel.Qos(c.cfg.Prefetch, 0, false); err != nil {
		return fmt.Errorf("set rabbitmq qos: %w", err)
	}

	deliveries, err := channel.Consume(c.cfg.Queue, c.cfg.ConsumerTag, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("start rabbitmq consumer: %w", err)
	}
	c.connected.Store(true)
	defer c.connected.Store(false)

	c.logger.InfoContext(ctx, "search.product_consumer.started",
		slog.String("exchange", c.cfg.Exchange),
		slog.String("queue", c.cfg.Queue),
		slog.Int("prefetch", c.cfg.Prefetch),
	)

	notifyClose := channel.NotifyClose(make(chan *amqp.Error, 1))
	for {
		select {
		case <-ctx.Done():
			return nil
		case amqpErr := <-notifyClose:
			if amqpErr == nil {
				return nil
			}
			return fmt.Errorf("rabbitmq channel closed: %w", amqpErr)
		case delivery, ok := <-deliveries:
			if !ok {
				return errors.New("rabbitmq deliveries channel closed")
			}
			c.processDelivery(ctx, channel, delivery)
		}
	}
}

func (c *RabbitMQProductConsumer) Ready() bool {
	return c != nil && c.connected.Load()
}

func (c *RabbitMQProductConsumer) setupTopology(channel *amqp.Channel) error {
	if err := channel.ExchangeDeclare(c.cfg.Exchange, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare product exchange: %w", err)
	}
	if err := channel.ExchangeDeclare(c.cfg.DeadLetterExchange, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare product indexer dlx: %w", err)
	}
	if _, err := channel.QueueDeclare(c.cfg.DeadLetterQueue, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare product indexer dlq: %w", err)
	}
	if err := channel.QueueBind(c.cfg.DeadLetterQueue, c.cfg.DeadLetterQueue, c.cfg.DeadLetterExchange, false, nil); err != nil {
		return fmt.Errorf("bind product indexer dlq: %w", err)
	}

	args := amqp.Table{
		"x-dead-letter-exchange":    c.cfg.DeadLetterExchange,
		"x-dead-letter-routing-key": c.cfg.DeadLetterQueue,
	}
	if _, err := channel.QueueDeclare(c.cfg.Queue, true, false, false, false, args); err != nil {
		return fmt.Errorf("declare product indexer queue: %w", err)
	}
	for _, routingKey := range c.cfg.RoutingKeys {
		if err := channel.QueueBind(c.cfg.Queue, routingKey, c.cfg.Exchange, false, nil); err != nil {
			return fmt.Errorf("bind product indexer queue %q: %w", routingKey, err)
		}
	}
	return nil
}

func (c *RabbitMQProductConsumer) processDelivery(parent context.Context, channel *amqp.Channel, delivery amqp.Delivery) {
	messageCtx, cancel := context.WithTimeout(parent, c.cfg.MessageTimeout)
	defer cancel()

	envelope, err := DecodeProductEnvelope(delivery.Body)
	if err != nil {
		c.handleFailure(parent, channel, delivery, Envelope[ProductIndexPayload]{}, err, true)
		return
	}
	if err := envelope.ValidateMetadata(); err != nil {
		c.handleFailure(parent, channel, delivery, envelope, err, true)
		return
	}
	if lag := time.Since(envelope.OccurredAt); lag >= 0 {
		c.cfg.Metrics.ObserveProductEventLag(messageCtx, envelope.EventType, lag)
	}

	processed, err := c.eventStore.WasProcessed(messageCtx, envelope.EventID)
	if err != nil {
		c.handleFailure(parent, channel, delivery, envelope, err, false)
		return
	}
	if processed {
		c.cfg.Metrics.RecordProductEvent(messageCtx, envelope.EventType, "duplicate")
		c.logger.InfoContext(messageCtx, "search.product_consumer.duplicate_acked",
			slog.String("event_id", envelope.EventID),
			slog.String("event_type", envelope.EventType),
			slog.String("product_id", envelope.Payload.ProductID),
			slog.String("trace_id", envelope.TraceID),
		)
		ackDelivery(delivery, c.logger)
		return
	}

	indexStarted := time.Now()
	result, err := c.indexer.Handle(messageCtx, envelope)
	if err != nil {
		c.cfg.Metrics.ObserveProductIndexDuration(messageCtx, envelope.EventType, "failed", time.Since(indexStarted))
		c.handleFailure(parent, channel, delivery, envelope, err, IsPermanentProductIndexError(err))
		return
	}
	c.cfg.Metrics.ObserveProductIndexDuration(messageCtx, envelope.EventType, "success", time.Since(indexStarted))
	if err := c.eventStore.MarkProcessed(messageCtx, envelope.EventID); err != nil {
		c.handleFailure(parent, channel, delivery, envelope, err, false)
		return
	}

	c.logger.InfoContext(messageCtx, "search.product_consumer.acked",
		slog.String("event_id", envelope.EventID),
		slog.String("event_type", envelope.EventType),
		slog.String("product_id", result.ProductID),
		slog.String("action", result.Action),
		slog.String("trace_id", envelope.TraceID),
	)
	c.cfg.Metrics.RecordProductEvent(messageCtx, envelope.EventType, "indexed")
	ackDelivery(delivery, c.logger)
}

func (c *RabbitMQProductConsumer) handleFailure(ctx context.Context, channel *amqp.Channel, delivery amqp.Delivery, envelope Envelope[ProductIndexPayload], err error, permanent bool) {
	retryCount := retryCountFromHeaders(delivery.Headers)
	if permanent || retryCount >= c.cfg.MaxRetries {
		reason := "permanent_error"
		if !permanent {
			reason = "max_retries_exceeded"
		}
		if publishErr := c.publishToDLQ(ctx, channel, delivery, envelope, err, reason, retryCount); publishErr != nil {
			c.logger.ErrorContext(ctx, "search.product_consumer.dlq_publish_failed",
				slog.String("event_id", envelope.EventID),
				slog.String("error", publishErr.Error()),
			)
			nackDelivery(delivery, true, c.logger)
			return
		}
		c.logger.WarnContext(ctx, "search.product_consumer.dead_lettered",
			slog.String("event_id", envelope.EventID),
			slog.String("event_type", envelope.EventType),
			slog.String("product_id", envelope.Payload.ProductID),
			slog.String("reason", reason),
			slog.Int("retry_count", retryCount),
			slog.String("error", err.Error()),
		)
		c.cfg.Metrics.RecordProductEvent(ctx, envelope.EventType, "dead_lettered")
		ackDelivery(delivery, c.logger)
		return
	}

	nextRetryCount := retryCount + 1
	delay := c.retryDelay(nextRetryCount)
	c.logger.WarnContext(ctx, "search.product_consumer.retry_scheduled",
		slog.String("event_id", envelope.EventID),
		slog.String("event_type", envelope.EventType),
		slog.String("product_id", envelope.Payload.ProductID),
		slog.Int("retry_count", nextRetryCount),
		slog.Duration("delay", delay),
		slog.String("error", err.Error()),
	)
	c.cfg.Metrics.RecordProductEvent(ctx, envelope.EventType, "retry_scheduled")
	if !sleepContext(ctx, delay) {
		nackDelivery(delivery, true, c.logger)
		return
	}
	if publishErr := c.publishRetry(ctx, channel, delivery, envelope, err, nextRetryCount); publishErr != nil {
		c.logger.ErrorContext(ctx, "search.product_consumer.retry_publish_failed",
			slog.String("event_id", envelope.EventID),
			slog.String("error", publishErr.Error()),
		)
		nackDelivery(delivery, true, c.logger)
		return
	}
	ackDelivery(delivery, c.logger)
}

func (c *RabbitMQProductConsumer) publishRetry(ctx context.Context, channel *amqp.Channel, delivery amqp.Delivery, envelope Envelope[ProductIndexPayload], err error, retryCount int) error {
	headers := cloneHeaders(delivery.Headers)
	headers[retryHeader] = int32(retryCount)
	headers["x-search-indexer-last-error"] = truncate(err.Error(), 512)
	headers["x-search-indexer-retried-at"] = time.Now().UTC().Format(time.RFC3339Nano)

	return channel.PublishWithContext(ctx, c.cfg.Exchange, c.routingKey(delivery, envelope), false, false, amqp.Publishing{
		Headers:         headers,
		ContentType:     contentType(delivery.ContentType),
		ContentEncoding: delivery.ContentEncoding,
		DeliveryMode:    amqp.Persistent,
		CorrelationId:   delivery.CorrelationId,
		MessageId:       delivery.MessageId,
		Timestamp:       time.Now().UTC(),
		Body:            delivery.Body,
	})
}

func (c *RabbitMQProductConsumer) publishToDLQ(ctx context.Context, channel *amqp.Channel, delivery amqp.Delivery, envelope Envelope[ProductIndexPayload], err error, reason string, retryCount int) error {
	headers := cloneHeaders(delivery.Headers)
	headers["x-search-indexer-failure-reason"] = reason
	headers["x-search-indexer-error"] = truncate(err.Error(), 1024)
	headers["x-search-indexer-retry-count"] = int32(retryCount)
	headers["x-search-indexer-failed-at"] = time.Now().UTC().Format(time.RFC3339Nano)
	headers["x-original-exchange"] = delivery.Exchange
	headers["x-original-routing-key"] = delivery.RoutingKey

	return channel.PublishWithContext(ctx, c.cfg.DeadLetterExchange, c.cfg.DeadLetterQueue, false, false, amqp.Publishing{
		Headers:         headers,
		ContentType:     contentType(delivery.ContentType),
		ContentEncoding: delivery.ContentEncoding,
		DeliveryMode:    amqp.Persistent,
		CorrelationId:   delivery.CorrelationId,
		MessageId:       delivery.MessageId,
		Timestamp:       time.Now().UTC(),
		Body:            delivery.Body,
	})
}

func (c *RabbitMQProductConsumer) routingKey(delivery amqp.Delivery, envelope Envelope[ProductIndexPayload]) string {
	if strings.TrimSpace(delivery.RoutingKey) != "" {
		return delivery.RoutingKey
	}
	switch envelope.EventType {
	case domain.ProductEventPublished:
		return "product.published"
	case domain.ProductEventUpdated:
		return "product.updated"
	case domain.ProductEventPriceChanged:
		return "product.price_changed"
	case domain.ProductEventInventoryChanged:
		return "product.inventory_changed"
	case domain.ProductEventUnpublished:
		return "product.unpublished"
	case domain.ProductEventDeleted:
		return "product.deleted"
	case domain.ProductEventBlocked:
		return "product.blocked"
	default:
		return c.cfg.RoutingKeys[0]
	}
}

func (c *RabbitMQProductConsumer) retryDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return c.cfg.RetryBaseDelay
	}
	pow := math.Pow(2, float64(attempt-1))
	delay := time.Duration(float64(c.cfg.RetryBaseDelay) * pow)
	if delay > c.cfg.RetryMaxDelay {
		return c.cfg.RetryMaxDelay
	}
	return delay
}

func (cfg RabbitMQConsumerConfig) Validate() error {
	if strings.TrimSpace(cfg.URL) == "" {
		return errors.New("RABBITMQ_URL cannot be empty")
	}
	if strings.TrimSpace(cfg.Exchange) == "" {
		return errors.New("PRODUCT_EVENTS_EXCHANGE cannot be empty")
	}
	if strings.TrimSpace(cfg.Queue) == "" {
		return errors.New("PRODUCT_INDEXER_QUEUE cannot be empty")
	}
	if strings.TrimSpace(cfg.DeadLetterExchange) == "" {
		return errors.New("PRODUCT_INDEXER_DLX cannot be empty")
	}
	if strings.TrimSpace(cfg.DeadLetterQueue) == "" {
		return errors.New("PRODUCT_INDEXER_DLQ cannot be empty")
	}
	if len(cfg.RoutingKeys) == 0 {
		return errors.New("PRODUCT_INDEXER_ROUTING_KEYS cannot be empty")
	}
	if cfg.Prefetch <= 0 {
		return errors.New("PRODUCT_INDEXER_PREFETCH must be greater than zero")
	}
	if cfg.MaxRetries < 0 {
		return errors.New("PRODUCT_INDEXER_MAX_RETRIES cannot be negative")
	}
	if cfg.ReconnectDelay <= 0 {
		return errors.New("PRODUCT_INDEXER_RECONNECT_DELAY must be greater than zero")
	}
	if cfg.RetryBaseDelay <= 0 {
		return errors.New("PRODUCT_INDEXER_RETRY_BASE_DELAY must be greater than zero")
	}
	if cfg.RetryMaxDelay < cfg.RetryBaseDelay {
		return errors.New("PRODUCT_INDEXER_RETRY_MAX_DELAY must be greater than or equal to base delay")
	}
	if cfg.MessageTimeout <= 0 {
		return errors.New("PRODUCT_INDEXER_MESSAGE_TIMEOUT must be greater than zero")
	}
	return nil
}

func (cfg RabbitMQConsumerConfig) withDefaults() RabbitMQConsumerConfig {
	if cfg.ConsumerTag == "" {
		cfg.ConsumerTag = "search-service-product-indexer"
	}
	if cfg.Prefetch <= 0 {
		cfg.Prefetch = 20
	}
	if cfg.ReconnectDelay <= 0 {
		cfg.ReconnectDelay = 5 * time.Second
	}
	if cfg.RetryBaseDelay <= 0 {
		cfg.RetryBaseDelay = time.Second
	}
	if cfg.RetryMaxDelay <= 0 {
		cfg.RetryMaxDelay = 5 * time.Minute
	}
	if cfg.MessageTimeout <= 0 {
		cfg.MessageTimeout = 30 * time.Second
	}
	return cfg
}

func retryCountFromHeaders(headers amqp.Table) int {
	if headers == nil {
		return 0
	}
	value, ok := headers[retryHeader]
	if !ok {
		return 0
	}
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case string:
		parsed, _ := strconv.Atoi(typed)
		return parsed
	default:
		return 0
	}
}

func cloneHeaders(headers amqp.Table) amqp.Table {
	cloned := amqp.Table{}
	for key, value := range headers {
		cloned[key] = value
	}
	return cloned
}

func ackDelivery(delivery amqp.Delivery, logger *slog.Logger) {
	if err := delivery.Ack(false); err != nil {
		logger.Error("search.product_consumer.ack_failed", slog.String("error", err.Error()))
	}
}

func nackDelivery(delivery amqp.Delivery, requeue bool, logger *slog.Logger) {
	if err := delivery.Nack(false, requeue); err != nil {
		logger.Error("search.product_consumer.nack_failed",
			slog.Bool("requeue", requeue),
			slog.String("error", err.Error()),
		)
	}
}

func sleepContext(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func contentType(value string) string {
	if strings.TrimSpace(value) == "" {
		return "application/json"
	}
	return value
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}
