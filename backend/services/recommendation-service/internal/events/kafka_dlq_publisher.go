package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/segmentio/kafka-go"
)

type KafkaDeadLetterPublisher struct {
	writer *kafka.Writer
	topic  string
}

func NewKafkaDeadLetterPublisher(brokers []string, topic string) (*KafkaDeadLetterPublisher, error) {
	if len(brokers) == 0 {
		return nil, errors.New("kafka brokers are required")
	}
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return nil, errors.New("kafka dlq topic is required")
	}
	return &KafkaDeadLetterPublisher{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireAll,
		},
		topic: topic,
	}, nil
}

func (p *KafkaDeadLetterPublisher) Publish(ctx context.Context, message DeadLetterMessage) error {
	if p == nil || p.writer == nil {
		return errors.New("kafka dead-letter publisher is not initialized")
	}
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal dead-letter message: %w", err)
	}
	key := message.EventID
	if key == "" {
		key = message.EventType
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic: p.topic,
		Key:   []byte(key),
		Value: payload,
		Time:  message.FailedAt,
	})
}

func (p *KafkaDeadLetterPublisher) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}
