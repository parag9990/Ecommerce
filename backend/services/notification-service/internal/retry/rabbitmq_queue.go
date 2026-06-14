package retry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

const maxRetryJobBytes = 16 * 1024

type RabbitQueue struct {
	connection *amqp.Connection
	consumer   *amqp.Channel
	publisher  *RabbitPublisher
	logger     *slog.Logger
}

func OpenRabbitQueue(
	ctx context.Context,
	url string,
	prefetch int,
	policy Policy,
	confirmTimeout time.Duration,
	logger *slog.Logger,
) (*RabbitQueue, error) {
	if ctx == nil {
		return nil, errors.New("notification retry RabbitMQ startup context is required")
	}
	if prefetch < 1 {
		return nil, errors.New("notification retry RabbitMQ prefetch must be positive")
	}
	if _, err := NewPolicy(policy.MaxAttempts, policy.Delays); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	dialer := &net.Dialer{}
	connection, err := amqp.DialConfig(strings.TrimSpace(url), amqp.Config{
		Dial: func(network, address string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, address)
		},
	})
	if err != nil {
		return nil, fmt.Errorf("connect notification retry RabbitMQ: %w", err)
	}
	closeConnection := func() { _ = connection.Close() }
	consumer, err := connection.Channel()
	if err != nil {
		closeConnection()
		return nil, fmt.Errorf("open notification retry consumer channel: %w", err)
	}
	closeConsumer := func() { _ = consumer.Close() }
	if err := consumer.Qos(prefetch, 0, false); err != nil {
		closeConsumer()
		closeConnection()
		return nil, fmt.Errorf("set notification retry prefetch: %w", err)
	}
	if err := DeclareTopology(consumer, policy); err != nil {
		closeConsumer()
		closeConnection()
		return nil, err
	}
	publishChannel, err := connection.Channel()
	if err != nil {
		closeConsumer()
		closeConnection()
		return nil, fmt.Errorf("open notification retry publisher channel: %w", err)
	}
	publisher, err := NewRabbitPublisher(publishChannel, confirmTimeout)
	if err != nil {
		_ = publishChannel.Close()
		closeConsumer()
		closeConnection()
		return nil, err
	}
	return &RabbitQueue{
		connection: connection, consumer: consumer, publisher: publisher, logger: logger,
	}, nil
}

func (q *RabbitQueue) Publisher() Publisher {
	if q == nil {
		return nil
	}
	return q.publisher
}

func (q *RabbitQueue) Run(ctx context.Context, worker *Worker) error {
	if q == nil || q.consumer == nil || worker == nil {
		return errors.New("notification retry queue dependencies are required")
	}
	if ctx == nil {
		return errors.New("notification retry queue context is required")
	}
	deliveries, err := q.consumer.ConsumeWithContext(ctx, AttemptQueue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume notification delivery attempts: %w", err)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case delivery, ok := <-deliveries:
			if !ok {
				if ctx.Err() != nil {
					return nil
				}
				return errors.New("notification attempt delivery stream closed")
			}
			if err := q.handleDelivery(ctx, worker, delivery); err != nil {
				return err
			}
		}
	}
}

func (q *RabbitQueue) handleDelivery(ctx context.Context, worker *Worker, delivery amqp.Delivery) error {
	job, invalidRoute, err := decodeRetryJob(delivery.Body)
	if err != nil {
		if err := worker.PublishInvalid(ctx, job, invalidRoute); err != nil {
			return requeueDelivery(delivery, err)
		}
		return ackDelivery(delivery)
	}
	err = worker.Process(ctx, job)
	switch {
	case err == nil:
		return ackDelivery(delivery)
	case errors.Is(err, ErrAttemptBusy):
		q.logger.WarnContext(ctx, "notification.retry.attempt_busy",
			slog.String("delivery_id", job.DeliveryID),
			slog.Int("attempt", job.Attempt),
		)
		timer := time.NewTimer(time.Second)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			return requeueDelivery(delivery, err)
		}
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		if ctx.Err() != nil {
			return nil
		}
		return requeueDelivery(delivery, err)
	default:
		q.logger.WarnContext(ctx, "notification.retry.requeue",
			slog.String("delivery_id", job.DeliveryID),
			slog.Int("attempt", job.Attempt),
			slog.String("failure_category", "temporary_processing"),
		)
		return requeueDelivery(delivery, err)
	}
}

func decodeRetryJob(body []byte) (domain.RetryJob, string, error) {
	var job domain.RetryJob
	route := "invalid"
	if len(body) == 0 || len(body) > maxRetryJobBytes {
		return job, route, domain.ErrInvalidRetryJob
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err == nil {
		for key := range fields {
			normalized := strings.ToLower(strings.NewReplacer("-", "_", ".", "_", " ", "_").Replace(key))
			for _, prohibited := range []string{"otp", "password", "secret", "token", "authorization", "recipient", "email", "phone", "body", "content"} {
				if strings.Contains(normalized, prohibited) {
					route = "security"
					break
				}
			}
		}
		_ = json.Unmarshal(body, &job)
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&job); err != nil {
		return job, route, fmt.Errorf("%w: %v", domain.ErrInvalidRetryJob, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return job, route, domain.ErrInvalidRetryJob
	}
	if err := job.Validate(); err != nil {
		return job, route, err
	}
	return job, "", nil
}

func ackDelivery(delivery amqp.Delivery) error {
	if err := delivery.Ack(false); err != nil {
		return fmt.Errorf("ack notification attempt delivery: %w", err)
	}
	return nil
}

func requeueDelivery(delivery amqp.Delivery, cause error) error {
	if err := delivery.Nack(false, true); err != nil {
		return fmt.Errorf("nack notification attempt delivery after %v: %w", cause, err)
	}
	return nil
}

func (q *RabbitQueue) Close() error {
	if q == nil {
		return nil
	}
	var closeErr error
	if q.publisher != nil && q.publisher.channel != nil {
		closeErr = q.publisher.channel.Close()
	}
	if q.consumer != nil {
		if err := q.consumer.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
	}
	if q.connection != nil {
		if err := q.connection.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
	}
	return closeErr
}
