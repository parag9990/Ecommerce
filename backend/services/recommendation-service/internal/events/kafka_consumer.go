package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumerConfig struct {
	Brokers  []string
	Topic    string
	GroupID  string
	MinBytes int
	MaxBytes int
}

type KafkaInteractionConsumer struct {
	reader  *kafka.Reader
	handler *InteractionHandler
	metrics QueueLagMetrics
	logger  *slog.Logger
	topic   string
	groupID string
}

type QueueLagMetrics interface {
	SetConsumerLag(topic string, consumerGroup string, lag float64)
}

func NewKafkaInteractionConsumer(cfg KafkaConsumerConfig, handler *InteractionHandler, metrics QueueLagMetrics, logger *slog.Logger) (*KafkaInteractionConsumer, error) {
	if len(cfg.Brokers) == 0 {
		return nil, errors.New("kafka brokers are required")
	}
	if strings.TrimSpace(cfg.Topic) == "" {
		return nil, errors.New("kafka topic is required")
	}
	if strings.TrimSpace(cfg.GroupID) == "" {
		return nil, errors.New("kafka consumer group is required")
	}
	if cfg.MinBytes <= 0 {
		cfg.MinBytes = 1
	}
	if cfg.MaxBytes <= 0 {
		cfg.MaxBytes = 10 << 20
	}
	if handler == nil {
		return nil, errors.New("interaction handler is required")
	}
	if logger == nil {
		logger = slog.Default()
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    strings.TrimSpace(cfg.Topic),
		GroupID:  strings.TrimSpace(cfg.GroupID),
		MinBytes: cfg.MinBytes,
		MaxBytes: cfg.MaxBytes,
	})
	return &KafkaInteractionConsumer{
		reader:  reader,
		handler: handler,
		metrics: metrics,
		logger:  logger,
		topic:   strings.TrimSpace(cfg.Topic),
		groupID: strings.TrimSpace(cfg.GroupID),
	}, nil
}

func (c *KafkaInteractionConsumer) Run(ctx context.Context) error {
	if c == nil || c.reader == nil {
		return errors.New("kafka interaction consumer is not initialized")
	}
	c.logger.InfoContext(ctx, "recommendation.kafka.consumer_started",
		slog.String("topic", c.topic),
		slog.String("group_id", c.groupID),
	)
	defer c.logger.InfoContext(context.Background(), "recommendation.kafka.consumer_stopped",
		slog.String("topic", c.topic),
		slog.String("group_id", c.groupID),
	)

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return fmt.Errorf("fetch kafka message: %w", err)
		}

		handleErr := c.handler.Handle(ctx, Message{
			Topic:      msg.Topic,
			Key:        append([]byte(nil), msg.Key...),
			Value:      append([]byte(nil), msg.Value...),
			ReceivedAt: time.Now().UTC(),
		})
		if handleErr != nil {
			if errors.Is(handleErr, context.Canceled) || errors.Is(handleErr, context.DeadlineExceeded) {
				return nil
			}
			c.logger.ErrorContext(ctx, "recommendation.kafka.message_failed",
				slog.String("topic", msg.Topic),
				slog.Int("partition", msg.Partition),
				slog.Int64("offset", msg.Offset),
				slog.String("error", handleErr.Error()),
			)
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return fmt.Errorf("commit kafka message: %w", err)
		}
		c.recordLag()
	}
}

func (c *KafkaInteractionConsumer) Close() error {
	if c == nil || c.reader == nil {
		return nil
	}
	return c.reader.Close()
}

func (c *KafkaInteractionConsumer) recordLag() {
	if c.metrics == nil || c.reader == nil {
		return
	}
	stats := c.reader.Stats()
	if stats.Lag >= 0 {
		c.metrics.SetConsumerLag(c.topic, c.groupID, float64(stats.Lag))
	}
}
