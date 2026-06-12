package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	defaultKafkaMaxAttempts  = 3
	defaultKafkaRetryBackoff = 500 * time.Millisecond
)

type KafkaProductEventConsumerConfig struct {
	Brokers      []string
	Topic        string
	GroupID      string
	DLQTopic     string
	MaxAttempts  int
	RetryBackoff time.Duration
}

type DeadLetterEnvelope struct {
	FailedAt        time.Time         `json:"failed_at"`
	SourceTopic     string            `json:"source_topic"`
	SourcePartition int               `json:"source_partition"`
	SourceOffset    int64             `json:"source_offset"`
	Key             string            `json:"key,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	Reason          string            `json:"reason"`
	Permanent       bool              `json:"permanent"`
	Attempts        int               `json:"attempts"`
	Error           string            `json:"error"`
	Body            string            `json:"body"`
}

type kafkaProductEventReader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type kafkaProductEventWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type KafkaProductEventConsumer struct {
	config  KafkaProductEventConsumerConfig
	reader  kafkaProductEventReader
	writer  kafkaProductEventWriter
	handler *ProductConsumer
	logger  *slog.Logger
	clock   func() time.Time
}

func NewKafkaProductEventConsumer(config KafkaProductEventConsumerConfig, handler *ProductConsumer, logger *slog.Logger) (*KafkaProductEventConsumer, error) {
	normalized, err := normalizeKafkaProductEventConsumerConfig(config)
	if err != nil {
		return nil, err
	}
	if handler == nil {
		return nil, errors.New("product event handler is required")
	}
	if logger == nil {
		logger = slog.Default()
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     normalized.Brokers,
		Topic:       normalized.Topic,
		GroupID:     normalized.GroupID,
		StartOffset: kafka.LastOffset,
		MinBytes:    1,
		MaxBytes:    10e6,
	})
	writer := &kafka.Writer{
		Addr:         kafka.TCP(normalized.Brokers...),
		Topic:        normalized.DLQTopic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireAll,
	}

	return newKafkaProductEventConsumerWithIO(normalized, handler, reader, writer, logger), nil
}

func newKafkaProductEventConsumerWithIO(config KafkaProductEventConsumerConfig, handler *ProductConsumer, reader kafkaProductEventReader, writer kafkaProductEventWriter, logger *slog.Logger) *KafkaProductEventConsumer {
	if logger == nil {
		logger = slog.Default()
	}
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = defaultKafkaMaxAttempts
	}
	if config.RetryBackoff <= 0 {
		config.RetryBackoff = defaultKafkaRetryBackoff
	}
	return &KafkaProductEventConsumer{
		config:  config,
		reader:  reader,
		writer:  writer,
		handler: handler,
		logger:  logger,
		clock: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (c *KafkaProductEventConsumer) Run(ctx context.Context) error {
	if c == nil || c.reader == nil || c.handler == nil {
		return errors.New("kafka product event consumer is not initialized")
	}
	ctx = contextOrBackground(ctx)
	c.logger.Info("wishlist product event consumer started",
		"topic", c.config.Topic,
		"group_id", c.config.GroupID,
		"dlq_topic", c.config.DLQTopic,
	)
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, context.Canceled) {
				return nil
			}
			return fmt.Errorf("fetch product event from kafka: %w", err)
		}
		if err := c.processKafkaMessage(ctx, message); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
	}
}

func (c *KafkaProductEventConsumer) Close(context.Context) error {
	if c == nil {
		return nil
	}
	var closeErr error
	if c.reader != nil {
		closeErr = errors.Join(closeErr, c.reader.Close())
	}
	if c.writer != nil {
		closeErr = errors.Join(closeErr, c.writer.Close())
	}
	return closeErr
}

func (c *KafkaProductEventConsumer) processKafkaMessage(ctx context.Context, message kafka.Message) error {
	headers := kafkaHeadersToMap(message.Headers)
	productMessage := ProductMessage{
		Body:       message.Value,
		Key:        message.Key,
		Topic:      firstNonEmpty(message.Topic, c.config.Topic),
		Partition:  message.Partition,
		Offset:     message.Offset,
		Headers:    headers,
		ReceivedAt: c.clock(),
	}

	var lastErr error
	for attempt := 1; attempt <= c.config.MaxAttempts; attempt++ {
		err := c.handler.HandleMessageWithMetadata(ctx, productMessage)
		if err == nil {
			return c.commit(ctx, message)
		}

		lastErr = err
		if IsPermanentProductEventError(err) {
			if err := c.publishDeadLetter(ctx, message, err, attempt, true); err != nil {
				return err
			}
			return c.commit(ctx, message)
		}
		if attempt < c.config.MaxAttempts {
			c.logger.Warn("product event processing will retry",
				"topic", productMessage.Topic,
				"partition", productMessage.Partition,
				"offset", productMessage.Offset,
				"attempt", attempt,
				"max_attempts", c.config.MaxAttempts,
				"error", err,
			)
			if err := sleepWithContext(ctx, time.Duration(attempt)*c.config.RetryBackoff); err != nil {
				return err
			}
		}
	}

	if err := c.publishDeadLetter(ctx, message, lastErr, c.config.MaxAttempts, false); err != nil {
		return err
	}
	return c.commit(ctx, message)
}

func (c *KafkaProductEventConsumer) publishDeadLetter(ctx context.Context, message kafka.Message, cause error, attempts int, permanent bool) error {
	if c.writer == nil {
		return errors.New("product event DLQ writer is not initialized")
	}
	reason := ProductEventErrorReason(cause)
	envelope := DeadLetterEnvelope{
		FailedAt:        c.clock(),
		SourceTopic:     firstNonEmpty(message.Topic, c.config.Topic),
		SourcePartition: message.Partition,
		SourceOffset:    message.Offset,
		Key:             string(message.Key),
		Headers:         kafkaHeadersToMap(message.Headers),
		Reason:          reason,
		Permanent:       permanent,
		Attempts:        attempts,
		Error:           errorString(cause),
		Body:            string(message.Value),
	}
	body, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("encode product event dead letter: %w", err)
	}

	headers := append([]kafka.Header{}, message.Headers...)
	headers = append(headers,
		kafka.Header{Key: "x-dead-letter-reason", Value: []byte(reason)},
		kafka.Header{Key: "x-dead-letter-source-topic", Value: []byte(envelope.SourceTopic)},
	)
	if err := c.writer.WriteMessages(ctx, kafka.Message{
		Key:     message.Key,
		Value:   body,
		Headers: headers,
		Time:    c.clock(),
	}); err != nil {
		return fmt.Errorf("publish product event to DLQ %q: %w", c.config.DLQTopic, err)
	}
	c.logger.Warn("product event sent to DLQ",
		"source_topic", envelope.SourceTopic,
		"source_partition", envelope.SourcePartition,
		"source_offset", envelope.SourceOffset,
		"dlq_topic", c.config.DLQTopic,
		"reason", reason,
		"attempts", attempts,
		"permanent", permanent,
	)
	return nil
}

func (c *KafkaProductEventConsumer) commit(ctx context.Context, message kafka.Message) error {
	if err := c.reader.CommitMessages(ctx, message); err != nil {
		return fmt.Errorf("commit product event offset topic %q partition %d offset %d: %w", firstNonEmpty(message.Topic, c.config.Topic), message.Partition, message.Offset, err)
	}
	return nil
}

func normalizeKafkaProductEventConsumerConfig(config KafkaProductEventConsumerConfig) (KafkaProductEventConsumerConfig, error) {
	normalized := config
	normalized.Topic = strings.TrimSpace(normalized.Topic)
	normalized.GroupID = strings.TrimSpace(normalized.GroupID)
	normalized.DLQTopic = strings.TrimSpace(normalized.DLQTopic)
	normalized.Brokers = normalizeStringSlice(normalized.Brokers)
	if normalized.MaxAttempts <= 0 {
		normalized.MaxAttempts = defaultKafkaMaxAttempts
	}
	if normalized.RetryBackoff <= 0 {
		normalized.RetryBackoff = defaultKafkaRetryBackoff
	}

	var validationErrors []string
	if len(normalized.Brokers) == 0 {
		validationErrors = append(validationErrors, "kafka brokers are required")
	}
	if normalized.Topic == "" {
		validationErrors = append(validationErrors, "product events topic is required")
	}
	if normalized.GroupID == "" {
		validationErrors = append(validationErrors, "product events group id is required")
	}
	if normalized.DLQTopic == "" {
		validationErrors = append(validationErrors, "product events DLQ topic is required")
	}
	if len(validationErrors) > 0 {
		return KafkaProductEventConsumerConfig{}, errors.New(strings.Join(validationErrors, "; "))
	}
	return normalized, nil
}

func kafkaHeadersToMap(headers []kafka.Header) map[string]string {
	if len(headers) == 0 {
		return nil
	}
	result := make(map[string]string, len(headers))
	for _, header := range headers {
		result[header.Key] = string(header.Value)
	}
	return result
}

func normalizeStringSlice(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

func sleepWithContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
