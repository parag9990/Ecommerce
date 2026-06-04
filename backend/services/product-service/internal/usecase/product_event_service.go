package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"product-service/internal/domain"
	"product-service/internal/mapper"
	"product-service/internal/repository"
)

type ProductEventRecorder interface {
	BuildProductEvent(
		ctx context.Context,
		eventType domain.ProductEventType,
		product domain.Product,
		occurredAt time.Time,
	) (*domain.ProductOutboxEvent, error)
	QueueProductEvent(ctx context.Context, event domain.ProductOutboxEvent) error
}

type ProductEventServiceOptions struct {
	Disabled bool
	Topic    string
	Source   string
}

type ProductEventService struct {
	outbox  repository.ProductEventOutboxWriter
	ids     ProductEventIDGenerator
	clock   Clock
	logger  *slog.Logger
	options ProductEventServiceOptions
}

func NewProductEventService(
	outbox repository.ProductEventOutboxWriter,
	ids ProductEventIDGenerator,
	clock Clock,
	logger *slog.Logger,
	options ProductEventServiceOptions,
) (*ProductEventService, error) {
	if !options.Disabled && outbox == nil {
		return nil, fmt.Errorf("product event outbox repository is required")
	}
	if ids == nil {
		ids = NewRandomIDGenerator()
	}
	if clock == nil {
		clock = SystemClock{}
	}
	if logger == nil {
		logger = slog.Default()
	}
	options = normalizeProductEventOptions(options)
	return &ProductEventService{
		outbox:  outbox,
		ids:     ids,
		clock:   clock,
		logger:  logger,
		options: options,
	}, nil
}

func (s *ProductEventService) BuildProductEvent(
	ctx context.Context,
	eventType domain.ProductEventType,
	product domain.Product,
	occurredAt time.Time,
) (*domain.ProductOutboxEvent, error) {
	if s == nil || s.options.Disabled {
		return nil, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !eventType.Valid() {
		return nil, serviceError(ErrorKindInvalidArgument, ErrorCodeValidation, "product event type is not supported", nil)
	}
	if occurredAt.IsZero() {
		occurredAt = s.clock.Now().UTC()
	} else {
		occurredAt = occurredAt.UTC()
	}

	payload := mapper.ToProductSearchEventPayload(product)
	if report := payload.Validate(); report.HasErrors() {
		return nil, validationFailed(report)
	}

	eventID := s.ids.NewProductEventID()
	requestID := RequestIDFromContext(ctx)
	if requestID == "" {
		requestID = eventID
	}
	traceID := TraceIDFromContext(ctx)
	if traceID == "" {
		traceID = requestID
	}

	event := domain.ProductOutboxEvent{
		ID:            eventID,
		Topic:         s.options.Topic,
		EventType:     string(eventType),
		Version:       domain.ProductEventSchemaVersion,
		Source:        s.options.Source,
		RequestID:     requestID,
		TraceID:       traceID,
		Payload:       payload.ToMap(),
		Status:        domain.OutboxStatusPending,
		Attempts:      0,
		NextAttemptAt: occurredAt,
		OccurredAt:    occurredAt,
	}
	if report := event.Validate(); report.HasErrors() {
		return nil, validationFailed(report)
	}
	return &event, nil
}

func (s *ProductEventService) QueueProductEvent(ctx context.Context, event domain.ProductOutboxEvent) error {
	if s == nil || s.options.Disabled {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := s.outbox.InsertOutboxEvent(ctx, &event); err != nil {
		return fmt.Errorf("queue product event %q: %w", event.ID, err)
	}
	s.logger.Info(
		"product event queued",
		"event_id", event.ID,
		"event_type", event.EventType,
		"product_id", event.Payload["product_id"],
		"topic", event.Topic,
		"request_id", event.RequestID,
		"trace_id", event.TraceID,
	)
	return nil
}

func normalizeProductEventOptions(options ProductEventServiceOptions) ProductEventServiceOptions {
	if strings.TrimSpace(options.Topic) == "" {
		options.Topic = domain.ProductEventDefaultTopic
	} else {
		options.Topic = strings.TrimSpace(options.Topic)
	}
	if strings.TrimSpace(options.Source) == "" {
		options.Source = domain.ProductEventDefaultSource
	} else {
		options.Source = strings.TrimSpace(options.Source)
	}
	return options
}
