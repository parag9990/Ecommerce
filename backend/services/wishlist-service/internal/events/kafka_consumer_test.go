package events

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
)

func TestKafkaProductEventConsumerSendsPermanentErrorToDLQAndCommits(t *testing.T) {
	handler := mustProductConsumer(t, &fakeAvailabilitySync{})
	reader := &fakeKafkaReader{}
	writer := &fakeKafkaWriter{}
	consumer := newKafkaProductEventConsumerWithIO(KafkaProductEventConsumerConfig{
		Topic:        "product.events",
		GroupID:      "wishlist-service",
		DLQTopic:     "product.events.wishlist.dlq",
		MaxAttempts:  3,
		RetryBackoff: time.Millisecond,
	}, handler, reader, writer, nil)

	err := consumer.processKafkaMessage(context.Background(), kafka.Message{
		Topic:     "product.events",
		Partition: 2,
		Offset:    42,
		Value:     []byte(`{"event_id":"evt_1","event_type":"ProductDeleted","version":1,"producer":"product-service","payload":{}}`),
	})
	if err != nil {
		t.Fatalf("processKafkaMessage returned error: %v", err)
	}
	if reader.commits != 1 {
		t.Fatalf("commits = %d, want 1", reader.commits)
	}
	if len(writer.messages) != 1 {
		t.Fatalf("DLQ messages = %d, want 1", len(writer.messages))
	}
	if string(writer.messages[0].Headers[len(writer.messages[0].Headers)-2].Value) != "missing_product_id" {
		t.Fatalf("DLQ reason header = %q, want missing_product_id", writer.messages[0].Headers[len(writer.messages[0].Headers)-2].Value)
	}
}

func TestKafkaProductEventConsumerRetriesThenDLQOnProcessingFailure(t *testing.T) {
	syncErr := errors.New("mongo unavailable")
	handler := mustProductConsumer(t, &fakeAvailabilitySync{err: syncErr})
	reader := &fakeKafkaReader{}
	writer := &fakeKafkaWriter{}
	consumer := newKafkaProductEventConsumerWithIO(KafkaProductEventConsumerConfig{
		Topic:        "product.events",
		GroupID:      "wishlist-service",
		DLQTopic:     "product.events.wishlist.dlq",
		MaxAttempts:  2,
		RetryBackoff: time.Millisecond,
	}, handler, reader, writer, nil)

	err := consumer.processKafkaMessage(context.Background(), kafka.Message{
		Topic:     "product.events",
		Partition: 1,
		Offset:    12,
		Value: []byte(`{
			"event_id":"evt_123",
			"event_type":"ProductDeleted",
			"version":1,
			"producer":"product-service",
			"payload":{"product_id":"prod_123"}
		}`),
	})
	if err != nil {
		t.Fatalf("processKafkaMessage returned error: %v", err)
	}
	if reader.commits != 1 {
		t.Fatalf("commits = %d, want 1", reader.commits)
	}
	if len(writer.messages) != 1 {
		t.Fatalf("DLQ messages = %d, want 1", len(writer.messages))
	}
}

type fakeKafkaReader struct {
	message kafka.Message
	err     error
	commits int
	closed  bool
}

func (f *fakeKafkaReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	if f.err != nil {
		return kafka.Message{}, f.err
	}
	return f.message, nil
}

func (f *fakeKafkaReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	f.commits += len(msgs)
	return nil
}

func (f *fakeKafkaReader) Close() error {
	f.closed = true
	return nil
}

type fakeKafkaWriter struct {
	messages []kafka.Message
	err      error
	closed   bool
}

func (f *fakeKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	if f.err != nil {
		return f.err
	}
	f.messages = append(f.messages, msgs...)
	return nil
}

func (f *fakeKafkaWriter) Close() error {
	f.closed = true
	return nil
}
