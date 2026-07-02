package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

const rabbitMQExchangeType = "topic"

type RabbitMQPublisher struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	confirms <-chan amqp.Confirmation
	mu       sync.Mutex
	declared map[string]struct{}
}

func NewRabbitMQPublisher(url string, topics ...string) (*RabbitMQPublisher, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return nil, errors.New("rabbitmq url is required")
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}
	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open rabbitmq channel: %w", err)
	}
	if err := channel.Confirm(false); err != nil {
		_ = channel.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("enable rabbitmq publisher confirms: %w", err)
	}

	publisher := &RabbitMQPublisher{
		conn:     conn,
		channel:  channel,
		confirms: channel.NotifyPublish(make(chan amqp.Confirmation, 1)),
		declared: make(map[string]struct{}),
	}
	for _, topic := range topics {
		if err := publisher.declareExchange(strings.TrimSpace(topic)); err != nil {
			_ = publisher.Close()
			return nil, err
		}
	}
	return publisher, nil
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, topic string, payload []byte) error {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return errors.New("rabbitmq topic is required")
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.declareExchangeLocked(topic); err != nil {
		return err
	}

	if err := p.channel.PublishWithContext(ctx, topic, routingKeyFromPayload(payload), false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         payload,
	}); err != nil {
		return err
	}

	select {
	case confirmation, ok := <-p.confirms:
		if !ok {
			return errors.New("rabbitmq publisher confirmation channel closed")
		}
		if !confirmation.Ack {
			return fmt.Errorf("rabbitmq broker rejected delivery %d", confirmation.DeliveryTag)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *RabbitMQPublisher) Close() error {
	var closeErr error
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			closeErr = err
		}
	}
	if p.conn != nil {
		if err := p.conn.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
	}
	return closeErr
}

func (p *RabbitMQPublisher) declareExchange(topic string) error {
	if topic == "" {
		return nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	return p.declareExchangeLocked(topic)
}

func (p *RabbitMQPublisher) declareExchangeLocked(topic string) error {
	if _, ok := p.declared[topic]; ok {
		return nil
	}
	if err := p.channel.ExchangeDeclare(topic, rabbitMQExchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare rabbitmq exchange %q: %w", topic, err)
	}
	p.declared[topic] = struct{}{}
	return nil
}

func routingKeyFromPayload(payload []byte) string {
	var envelope struct {
		EventType string `json:"event_type"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return ""
	}
	return strings.TrimSpace(envelope.EventType)
}
