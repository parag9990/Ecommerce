package events

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
	kafka "github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	writer *kafka.Writer
}

type KafkaPublisherConfig struct {
	Brokers      []string
	Topic        string
	WriteTimeout time.Duration
}

func NewKafkaPublisher(config KafkaPublisherConfig) (*KafkaPublisher, error) {
	brokers := normalizeBrokers(config.Brokers)
	if len(brokers) == 0 {
		return nil, errors.New("at least one kafka broker is required")
	}
	topic := strings.TrimSpace(config.Topic)
	if topic == "" {
		return nil, errors.New("kafka topic is required")
	}
	writeTimeout := config.WriteTimeout
	if writeTimeout <= 0 {
		writeTimeout = 10 * time.Second
	}
	return &KafkaPublisher{writer: &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
		WriteTimeout: writeTimeout,
	}}, nil
}

func (p *KafkaPublisher) Publish(ctx context.Context, _ string, event domain.OutboxEvent) error {
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.AggregateID),
		Value: event.Payload,
		Headers: []kafka.Header{
			{Key: "event_id", Value: []byte(event.EventID)},
			{Key: "event_type", Value: []byte(event.EventType)},
			{Key: "routing_key", Value: []byte(event.RoutingKey)},
		},
	})
}

func (p *KafkaPublisher) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

func normalizeBrokers(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	var brokers []string
	for _, value := range values {
		for _, entry := range strings.Split(value, ",") {
			entry = strings.TrimSpace(entry)
			if entry == "" {
				continue
			}
			if _, exists := seen[entry]; exists {
				continue
			}
			seen[entry] = struct{}{}
			brokers = append(brokers, entry)
		}
	}
	return brokers
}
