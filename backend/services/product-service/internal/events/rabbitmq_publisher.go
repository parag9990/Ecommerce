package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"product-service/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	url      string
	logger   *slog.Logger
	mu       sync.Mutex
	conn     *amqp.Connection
	channel  *amqp.Channel
	declared map[string]struct{}
}

func NewRabbitMQPublisher(amqpURL string, logger *slog.Logger) (*RabbitMQPublisher, error) {
	amqpURL = strings.TrimSpace(amqpURL)
	if amqpURL == "" {
		return nil, fmt.Errorf("rabbitmq url is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &RabbitMQPublisher{
		url:      amqpURL,
		logger:   logger,
		declared: make(map[string]struct{}),
	}, nil
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, topic string, envelope domain.EventEnvelope) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return fmt.Errorf("event topic is required")
	}
	body, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal product event envelope %q: %w", envelope.EventID, err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.ensureChannelLocked(); err != nil {
		return err
	}
	if err := p.ensureExchangeLocked(topic); err != nil {
		p.resetLocked()
		return err
	}
	err = p.channel.PublishWithContext(
		ctx,
		topic,
		productRoutingKey(envelope.EventType),
		true,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    envelope.EventID,
			Type:         envelope.EventType,
			Timestamp:    envelope.OccurredAt.UTC(),
			Headers: amqp.Table{
				"event_id":   envelope.EventID,
				"request_id": envelope.RequestID,
				"trace_id":   envelope.TraceID,
				"source":     envelope.Source,
				"version":    envelope.Version,
			},
			Body: body,
		},
	)
	if err != nil {
		p.resetLocked()
		return fmt.Errorf("publish product event %q to rabbitmq topic %q: %w", envelope.EventID, topic, err)
	}
	return nil
}

func (p *RabbitMQPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.resetLocked()
}

func (p *RabbitMQPublisher) ensureChannelLocked() error {
	if p.channel != nil {
		return nil
	}
	conn, err := amqp.Dial(p.url)
	if err != nil {
		return fmt.Errorf("connect rabbitmq: %w", err)
	}
	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("open rabbitmq channel: %w", err)
	}
	p.conn = conn
	p.channel = channel
	p.declared = make(map[string]struct{})
	p.logger.Debug("rabbitmq publisher connected")
	return nil
}

func (p *RabbitMQPublisher) ensureExchangeLocked(topic string) error {
	if _, ok := p.declared[topic]; ok {
		return nil
	}
	err := p.channel.ExchangeDeclare(
		topic,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare rabbitmq exchange %q: %w", topic, err)
	}
	p.declared[topic] = struct{}{}
	return nil
}

func productRoutingKey(eventType string) string {
	switch domain.ProductEventType(strings.TrimSpace(eventType)) {
	case domain.ProductEventCreated:
		return "product.created"
	case domain.ProductEventUpdated:
		return "product.updated"
	case domain.ProductEventPublished:
		return "product.published"
	case domain.ProductEventUnpublished:
		return "product.unpublished"
	case domain.ProductEventInventoryChanged:
		return "product.inventory_changed"
	default:
		return "product.unknown"
	}
}

func (p *RabbitMQPublisher) resetLocked() error {
	var err error
	if p.channel != nil {
		if closeErr := p.channel.Close(); closeErr != nil {
			err = closeErr
		}
	}
	if p.conn != nil {
		if closeErr := p.conn.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	p.conn = nil
	p.channel = nil
	p.declared = make(map[string]struct{})
	return err
}
