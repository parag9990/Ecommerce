package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/usecase"

	"github.com/segmentio/kafka-go"
)

const (
	NotificationCommandVersion   = 1
	NotificationCommandEventType = "SendNotification"
	NotificationCommandProducer  = "wishlist-service"
)

type KafkaNotificationPublisherConfig struct {
	Brokers []string
	Topic   string
}

type NotificationCommandEnvelope struct {
	EventID    string                     `json:"event_id"`
	EventType  string                     `json:"event_type"`
	Version    int                        `json:"version"`
	OccurredAt time.Time                  `json:"occurred_at"`
	Producer   string                     `json:"producer"`
	TraceID    string                     `json:"trace_id,omitempty"`
	Payload    NotificationCommandPayload `json:"payload"`
}

type NotificationCommandPayload struct {
	UserID         string         `json:"user_id"`
	Channel        string         `json:"channel"`
	TemplateKey    string         `json:"template_key"`
	IdempotencyKey string         `json:"idempotency_key"`
	Variables      map[string]any `json:"variables,omitempty"`
}

type kafkaNotificationWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type KafkaNotificationPublisher struct {
	config KafkaNotificationPublisherConfig
	writer kafkaNotificationWriter
	logger *slog.Logger
	clock  func() time.Time
}

func NewKafkaNotificationPublisher(config KafkaNotificationPublisherConfig, logger *slog.Logger) (*KafkaNotificationPublisher, error) {
	normalized, err := normalizeKafkaNotificationPublisherConfig(config)
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	writer := &kafka.Writer{
		Addr:         kafka.TCP(normalized.Brokers...),
		Topic:        normalized.Topic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireAll,
	}
	return newKafkaNotificationPublisherWithWriter(normalized, writer, logger), nil
}

func newKafkaNotificationPublisherWithWriter(config KafkaNotificationPublisherConfig, writer kafkaNotificationWriter, logger *slog.Logger) *KafkaNotificationPublisher {
	if logger == nil {
		logger = slog.Default()
	}
	return &KafkaNotificationPublisher{
		config: config,
		writer: writer,
		logger: logger,
		clock: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (p *KafkaNotificationPublisher) Publish(ctx context.Context, command usecase.NotificationCommand) error {
	if p == nil || p.writer == nil {
		return errors.New("notification publisher is not initialized")
	}
	normalized, err := normalizeNotificationCommand(command, p.clock)
	if err != nil {
		return err
	}

	envelope := NotificationCommandEnvelope{
		EventID:    normalized.EventID,
		EventType:  NotificationCommandEventType,
		Version:    NotificationCommandVersion,
		OccurredAt: normalized.OccurredAt,
		Producer:   NotificationCommandProducer,
		TraceID:    normalized.TraceID,
		Payload: NotificationCommandPayload{
			UserID:         normalized.UserID,
			Channel:        normalized.Channel,
			TemplateKey:    normalized.TemplateKey,
			IdempotencyKey: normalized.IdempotencyKey,
			Variables:      normalized.Variables,
		},
	}
	body, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("encode notification command: %w", err)
	}

	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(normalized.UserID),
		Value: body,
		Time:  normalized.OccurredAt,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(NotificationCommandEventType)},
			{Key: "idempotency_key", Value: []byte(normalized.IdempotencyKey)},
		},
	}); err != nil {
		return fmt.Errorf("publish notification command to %q: %w", p.config.Topic, err)
	}

	p.logger.Debug("notification command published",
		"event_id", normalized.EventID,
		"user_id", normalized.UserID,
		"template_key", normalized.TemplateKey,
		"channel", normalized.Channel,
		"topic", p.config.Topic,
	)
	return nil
}

func (p *KafkaNotificationPublisher) Close(context.Context) error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

func normalizeKafkaNotificationPublisherConfig(config KafkaNotificationPublisherConfig) (KafkaNotificationPublisherConfig, error) {
	normalized := config
	normalized.Brokers = normalizeStringSlice(normalized.Brokers)
	normalized.Topic = strings.TrimSpace(normalized.Topic)

	var validationErrors []string
	if len(normalized.Brokers) == 0 {
		validationErrors = append(validationErrors, "kafka brokers are required")
	}
	if normalized.Topic == "" {
		validationErrors = append(validationErrors, "notification commands topic is required")
	}
	if len(validationErrors) > 0 {
		return KafkaNotificationPublisherConfig{}, errors.New(strings.Join(validationErrors, "; "))
	}
	return normalized, nil
}

func normalizeNotificationCommand(command usecase.NotificationCommand, clock func() time.Time) (usecase.NotificationCommand, error) {
	command.EventID = strings.TrimSpace(command.EventID)
	command.UserID = strings.TrimSpace(command.UserID)
	command.TemplateKey = strings.TrimSpace(command.TemplateKey)
	command.Channel = strings.ToLower(strings.TrimSpace(command.Channel))
	command.IdempotencyKey = strings.TrimSpace(command.IdempotencyKey)
	command.TraceID = strings.TrimSpace(command.TraceID)
	if command.OccurredAt.IsZero() {
		if clock == nil {
			clock = func() time.Time { return time.Now().UTC() }
		}
		command.OccurredAt = clock().UTC()
	} else {
		command.OccurredAt = command.OccurredAt.UTC()
	}

	var validationErrors []string
	if command.EventID == "" {
		validationErrors = append(validationErrors, "notification command event_id is required")
	}
	if command.UserID == "" {
		validationErrors = append(validationErrors, "notification command user_id is required")
	}
	if command.TemplateKey == "" {
		validationErrors = append(validationErrors, "notification command template_key is required")
	}
	if command.Channel == "" {
		validationErrors = append(validationErrors, "notification command channel is required")
	}
	if command.IdempotencyKey == "" {
		validationErrors = append(validationErrors, "notification command idempotency_key is required")
	}
	if len(validationErrors) > 0 {
		return usecase.NotificationCommand{}, errors.New(strings.Join(validationErrors, "; "))
	}
	return command, nil
}
