package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
)

type RetryEventRepository interface {
	PrepareEventDelivery(ctx context.Context, delivery domain.Delivery) (bool, error)
}

type AttemptPublisher interface {
	PublishAttempt(ctx context.Context, job domain.RetryJob) error
}

type RecipientProtector interface {
	Protect(recipient, deliveryID string) (string, error)
	Reveal(ciphertext, deliveryID string) (string, error)
}

// QueueEventService stores a retry-safe intent and enqueues its first attempt.
type QueueEventService struct {
	deliveries  RetryEventRepository
	publisher   AttemptPublisher
	protector   RecipientProtector
	maxAttempts int
	logger      *slog.Logger
	now         func() time.Time
}

func NewQueueEventService(
	deliveries RetryEventRepository,
	publisher AttemptPublisher,
	protector RecipientProtector,
	maxAttempts int,
	logger *slog.Logger,
) (*QueueEventService, error) {
	if nilServiceDependency(deliveries) || nilServiceDependency(publisher) || nilServiceDependency(protector) {
		return nil, errors.New("queued event notification dependencies are required")
	}
	if maxAttempts < 1 {
		return nil, errors.New("queued event notification max attempts must be positive")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &QueueEventService{
		deliveries: deliveries, publisher: publisher, protector: protector,
		maxAttempts: maxAttempts, logger: logger, now: time.Now,
	}, nil
}

func (s *QueueEventService) WithClock(now func() time.Time) {
	if now != nil {
		s.now = now
	}
}

func (s *QueueEventService) TriggerFromEvent(ctx context.Context, req domain.EventNotificationTrigger) error {
	if ctx == nil {
		return fmt.Errorf("%w: context is required", domain.ErrInvalidEventTrigger)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := req.Validate(); err != nil {
		return err
	}
	deliveryID := eventDeliveryID(req.IdempotencyKey)
	recipient, err := s.protector.Protect(req.Recipient, deliveryID)
	if err != nil {
		return fmt.Errorf("protect event notification recipient: %w", err)
	}
	now := s.now().UTC()
	delivery := domain.Delivery{
		ID:                  deliveryID,
		UserID:              strings.TrimSpace(req.UserID),
		Channel:             req.Channel,
		TemplateKey:         string(req.TemplateKey),
		Status:              domain.DeliveryStatusPending,
		Attempts:            0,
		Payload:             eventPayload(req.Variables),
		IdempotencyKey:      req.IdempotencyKey,
		SourceEventID:       req.SourceEventID,
		SourceEventType:     req.SourceType,
		TraceID:             strings.TrimSpace(req.TraceID),
		RecipientCiphertext: recipient,
		MaxAttempts:         s.maxAttempts,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	shouldPublish, err := s.deliveries.PrepareEventDelivery(ctx, delivery)
	if err != nil {
		return fmt.Errorf("prepare event notification delivery: %w", err)
	}
	if !shouldPublish {
		return domain.ErrDuplicateEventDelivery
	}
	job := domain.RetryJob{
		DeliveryID: delivery.ID, IdempotencyKey: delivery.IdempotencyKey,
		Attempt: 1, MaxAttempts: s.maxAttempts, TraceID: delivery.TraceID, QueuedAt: now,
	}
	if err := s.publisher.PublishAttempt(ctx, job); err != nil {
		return fmt.Errorf("enqueue event notification attempt: %w", err)
	}
	s.logger.InfoContext(ctx, "notification.event.attempt_enqueued",
		slog.String("delivery_id", delivery.ID),
		slog.String("event_id", delivery.SourceEventID),
		slog.String("trace_id", delivery.TraceID),
		slog.String("channel", string(delivery.Channel)),
		slog.Int("attempt", 1),
	)
	return nil
}

func eventPayload(variables map[string]string) map[string]any {
	payload := make(map[string]any, len(variables))
	for key, value := range variables {
		payload[key] = value
	}
	return payload
}

func eventDeliveryID(idempotencyKey string) string {
	sum := sha256.Sum256([]byte(idempotencyKey))
	return "delivery_event_" + hex.EncodeToString(sum[:16])
}

func nilServiceDependency(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
