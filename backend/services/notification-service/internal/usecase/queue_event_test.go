package usecase_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/usecase"
)

type queuedDeliveries struct {
	delivery domain.Delivery
	publish  bool
	err      error
	calls    int
}

func (r *queuedDeliveries) PrepareEventDelivery(_ context.Context, delivery domain.Delivery) (bool, error) {
	r.calls++
	r.delivery = delivery
	return r.publish, r.err
}

type attemptPublisher struct {
	job   domain.RetryJob
	err   error
	calls int
}

func (p *attemptPublisher) PublishAttempt(_ context.Context, job domain.RetryJob) error {
	p.calls++
	p.job = job
	return p.err
}

type recipientProtector struct {
	protected string
	revealed  string
}

func (p *recipientProtector) Protect(recipient, _ string) (string, error) {
	p.protected = recipient
	return "encrypted-recipient", nil
}

func (p *recipientProtector) Reveal(string, string) (string, error) {
	return p.revealed, nil
}

func TestQueueEventStoresEncryptedIntentAndPublishesMetadataJob(t *testing.T) {
	t.Parallel()

	deliveries := &queuedDeliveries{publish: true}
	publisher := &attemptPublisher{}
	protector := &recipientProtector{}
	service := newQueueEventService(t, deliveries, publisher, protector)
	if err := service.TriggerFromEvent(context.Background(), eventTrigger()); err != nil {
		t.Fatalf("TriggerFromEvent() error = %v", err)
	}
	if deliveries.delivery.Status != domain.DeliveryStatusPending ||
		deliveries.delivery.RecipientCiphertext != "encrypted-recipient" ||
		deliveries.delivery.MaxAttempts != 4 ||
		deliveries.delivery.Payload["order_id"] != "order_1" {
		t.Fatalf("delivery = %+v", deliveries.delivery)
	}
	if protector.protected != "buyer@example.com" ||
		publisher.calls != 1 || publisher.job.Attempt != 1 ||
		publisher.job.IdempotencyKey != eventTrigger().IdempotencyKey {
		t.Fatalf("protected = %q, job = %+v", protector.protected, publisher.job)
	}
}

func TestQueueEventDoesNotPublishCompletedDuplicate(t *testing.T) {
	t.Parallel()

	publisher := &attemptPublisher{}
	service := newQueueEventService(t, &queuedDeliveries{publish: false}, publisher, &recipientProtector{})
	err := service.TriggerFromEvent(context.Background(), eventTrigger())
	if !errors.Is(err, domain.ErrDuplicateEventDelivery) || publisher.calls != 0 {
		t.Fatalf("TriggerFromEvent() error = %v, publishes = %d", err, publisher.calls)
	}
}

func newQueueEventService(
	t *testing.T,
	deliveries usecase.RetryEventRepository,
	publisher usecase.AttemptPublisher,
	protector usecase.RecipientProtector,
) *usecase.QueueEventService {
	t.Helper()
	service, err := usecase.NewQueueEventService(deliveries, publisher, protector, 4,
		slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("NewQueueEventService() error = %v", err)
	}
	service.WithClock(func() time.Time {
		return time.Date(2026, time.May, 27, 14, 0, 0, 0, time.UTC)
	})
	return service
}
