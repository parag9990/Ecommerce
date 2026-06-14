package retry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher interface {
	PublishAttempt(ctx context.Context, job domain.RetryJob) error
	PublishRetry(ctx context.Context, route string, job domain.RetryJob) error
	PublishDeadLetter(ctx context.Context, route string, job domain.RetryJob) error
}

type RabbitPublisher struct {
	channel        *amqp.Channel
	returns        <-chan amqp.Return
	confirmTimeout time.Duration
	mu             sync.Mutex
}

func NewRabbitPublisher(channel *amqp.Channel, confirmTimeout time.Duration) (*RabbitPublisher, error) {
	if channel == nil {
		return nil, errors.New("notification publisher channel is required")
	}
	if confirmTimeout <= 0 {
		return nil, errors.New("notification publisher confirm timeout must be positive")
	}
	if err := channel.Confirm(false); err != nil {
		return nil, fmt.Errorf("enable notification publisher confirms: %w", err)
	}
	return &RabbitPublisher{
		channel:        channel,
		returns:        channel.NotifyReturn(make(chan amqp.Return, 1)),
		confirmTimeout: confirmTimeout,
	}, nil
}

func (p *RabbitPublisher) PublishAttempt(ctx context.Context, job domain.RetryJob) error {
	if err := job.Validate(); err != nil {
		return err
	}
	return p.publish(ctx, AttemptExchange, "send", job)
}

func (p *RabbitPublisher) PublishRetry(ctx context.Context, route string, job domain.RetryJob) error {
	if err := job.Validate(); err != nil {
		return err
	}
	return p.publish(ctx, RetryExchange, route, job)
}

func (p *RabbitPublisher) PublishDeadLetter(ctx context.Context, route string, job domain.RetryJob) error {
	switch route {
	case "failed", "invalid", "security", "rejected":
	default:
		return fmt.Errorf("unsupported notification dead-letter route %q", route)
	}
	return p.publish(ctx, DeadExchange, route, job)
}

func (p *RabbitPublisher) publish(
	ctx context.Context,
	exchange string,
	route string,
	job domain.RetryJob,
) error {
	if p == nil || p.channel == nil {
		return errors.New("notification publisher is not configured")
	}
	if ctx == nil {
		return errors.New("notification publish context is required")
	}
	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("encode notification retry job: %w", err)
	}
	waitCtx, cancel := context.WithTimeout(ctx, p.confirmTimeout)
	defer cancel()

	p.mu.Lock()
	defer p.mu.Unlock()
	p.discardReturnedMessages()
	confirmation, err := p.channel.PublishWithDeferredConfirmWithContext(
		waitCtx,
		exchange,
		route,
		true,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    fmt.Sprintf("%s:%d:%d", job.DeliveryID, job.Attempt, time.Now().UnixNano()),
			Timestamp:    time.Now().UTC(),
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("publish notification job: %w", err)
	}
	confirmed, err := confirmation.WaitContext(waitCtx)
	if err != nil {
		return fmt.Errorf("wait for notification publish confirmation: %w", err)
	}
	if !confirmed {
		return errors.New("notification job publish was negatively acknowledged")
	}
	select {
	case returned := <-p.returns:
		return fmt.Errorf("notification job was unroutable: exchange %q route %q code %d",
			returned.Exchange, returned.RoutingKey, returned.ReplyCode)
	default:
		return nil
	}
}

func (p *RabbitPublisher) discardReturnedMessages() {
	for {
		select {
		case <-p.returns:
		default:
			return
		}
	}
}
