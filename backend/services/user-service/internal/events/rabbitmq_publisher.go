package events

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

const rabbitMQExchangeType = "fanout"

type RabbitMQPublisher struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
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

	publisher := &RabbitMQPublisher{
		conn:     conn,
		channel:  channel,
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
	if err := p.declareExchange(topic); err != nil {
		return err
	}

	return p.channel.PublishWithContext(ctx, topic, "", false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         payload,
	})
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
	if _, ok := p.declared[topic]; ok {
		return nil
	}
	if err := p.channel.ExchangeDeclare(topic, rabbitMQExchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare rabbitmq exchange %q: %w", topic, err)
	}
	p.declared[topic] = struct{}{}
	return nil
}
