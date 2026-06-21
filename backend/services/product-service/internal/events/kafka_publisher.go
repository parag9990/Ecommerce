package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"product-service/internal/domain"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPublisher(brokers []string, writeTimeout time.Duration) (*KafkaPublisher, error) {
	cleaned := make([]string, 0, len(brokers))
	for _, broker := range brokers {
		if broker = strings.TrimSpace(broker); broker != "" {
			cleaned = append(cleaned, broker)
		}
	}
	if len(cleaned) == 0 {
		return nil, errors.New("kafka brokers are required")
	}
	if writeTimeout <= 0 {
		writeTimeout = 10 * time.Second
	}
	return &KafkaPublisher{writer: &kafka.Writer{
		Addr:         kafka.TCP(cleaned...),
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireAll,
		WriteTimeout: writeTimeout,
	}}, nil
}

func (p *KafkaPublisher) Publish(ctx context.Context, topic string, envelope domain.EventEnvelope) error {
	if p == nil || p.writer == nil {
		return errors.New("kafka publisher is not initialized")
	}
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return errors.New("event topic is required")
	}
	body, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal product event %q: %w", envelope.EventID, err)
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(envelope.EventID),
		Value: body,
		Time:  envelope.OccurredAt.UTC(),
		Headers: []kafka.Header{
			{Key: "event_id", Value: []byte(envelope.EventID)},
			{Key: "event_type", Value: []byte(envelope.EventType)},
			{Key: "request_id", Value: []byte(envelope.RequestID)},
			{Key: "trace_id", Value: []byte(envelope.TraceID)},
		},
	})
}

func (p *KafkaPublisher) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}
