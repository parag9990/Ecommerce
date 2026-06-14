package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/provider"
)

type EventProviderRegistry interface {
	Resolve(channel domain.Channel) (provider.Provider, error)
}

type EventDeliveryRepository interface {
	ClaimEventDelivery(ctx context.Context, delivery domain.Delivery) (bool, error)
	RecordEventDeliveryOutcome(ctx context.Context, idempotencyKey string, outcome domain.EventDeliveryOutcome) error
}

// SendEventService renders and dispatches one allowlisted, idempotently claimed event notification.
type SendEventService struct {
	renderer   Renderer
	providers  EventProviderRegistry
	deliveries EventDeliveryRepository
	analytics  DeliveryAnalyticsRecorder
	consent    *ConsentGate
	logger     *slog.Logger
	now        func() time.Time
}

func NewSendEventService(
	renderer Renderer,
	providers EventProviderRegistry,
	deliveries EventDeliveryRepository,
	consent *ConsentGate,
	analytics DeliveryAnalyticsRecorder,
	logger *slog.Logger,
) (*SendEventService, error) {
	if nilDependency(renderer) {
		return nil, errors.New("event template renderer is required")
	}
	if nilDependency(providers) {
		return nil, errors.New("event provider registry is required")
	}
	if nilDependency(deliveries) {
		return nil, errors.New("event delivery repository is required")
	}
	if consent == nil {
		return nil, errors.New("event notification consent gate is required")
	}
	if nilDependency(analytics) {
		return nil, errors.New("event notification analytics recorder is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &SendEventService{
		renderer:   renderer,
		providers:  providers,
		deliveries: deliveries,
		consent:    consent,
		analytics:  analytics,
		logger:     logger,
		now:        time.Now,
	}, nil
}

func (s *SendEventService) WithClock(now func() time.Time) {
	if now != nil {
		s.now = now
	}
}

func (s *SendEventService) TriggerFromEvent(ctx context.Context, req domain.EventNotificationTrigger) error {
	if ctx == nil {
		return fmt.Errorf("%w: context is required", domain.ErrInvalidEventTrigger)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := req.Validate(); err != nil {
		return err
	}
	decision, err := s.consent.Evaluate(ctx, req.UserID, req.Channel, req.TemplateKey, false)
	if err != nil {
		return fmt.Errorf("evaluate event notification preference: %w", err)
	}
	if !decision.Allowed {
		claimed, err := s.deliveries.ClaimEventDelivery(ctx, newSuppressedEventDelivery(req, decision.Reason, s.now().UTC()))
		if err != nil {
			return fmt.Errorf("record suppressed event delivery: %w", err)
		}
		if !claimed {
			return domain.ErrDuplicateEventDelivery
		}
		s.logger.InfoContext(ctx, "notification.event.suppressed",
			slog.String("event_id", req.SourceEventID),
			slog.String("trace_id", strings.TrimSpace(req.TraceID)),
			slog.String("channel", string(req.Channel)),
			slog.String("template_key", string(req.TemplateKey)),
			slog.String("reason", string(decision.Reason)),
		)
		return nil
	}

	rendered, err := s.renderer.Render(ctx, domain.RenderRequest{
		TemplateKey: req.TemplateKey,
		Channel:     req.Channel,
		Variables:   req.Variables,
	})
	if err != nil {
		return fmt.Errorf("render event template: %w", err)
	}
	message, err := eventMessage(req, rendered)
	if err != nil {
		return fmt.Errorf("build event message: %w", err)
	}
	sender, err := s.providers.Resolve(req.Channel)
	if err != nil {
		return fmt.Errorf("resolve event provider: %w", err)
	}

	now := s.now().UTC()
	claimed, err := s.deliveries.ClaimEventDelivery(ctx, newEventDeliveryClaim(req, now))
	if err != nil {
		return fmt.Errorf("claim event delivery: %w", err)
	}
	if !claimed {
		return domain.ErrDuplicateEventDelivery
	}

	result, sendErr := sender.Send(ctx, message)
	if errors.Is(sendErr, context.Canceled) || errors.Is(sendErr, context.DeadlineExceeded) {
		return sendErr
	}
	status, safeSendErr := eventDeliveryOutcome(result, sendErr)
	providerName := strings.TrimSpace(result.ProviderName)
	if providerName == "" {
		providerName = sender.Name()
	}
	if err := s.deliveries.RecordEventDeliveryOutcome(ctx, req.IdempotencyKey, domain.EventDeliveryOutcome{
		Status:            status,
		Provider:          providerName,
		ProviderMessageID: strings.TrimSpace(result.ProviderMessageID),
		UpdatedAt:         s.now().UTC(),
	}); err != nil {
		return fmt.Errorf("record event delivery outcome: %w", err)
	}
	switch status {
	case domain.DeliveryStatusAccepted:
		if _, err := s.analytics.RecordSent(ctx, eventDeliveryID(req.IdempotencyKey), providerName,
			strings.TrimSpace(result.ProviderMessageID), s.now().UTC()); err != nil {
			return fmt.Errorf("record sent event notification milestone: %w", err)
		}
	case domain.DeliveryStatusRejected:
		if _, err := s.analytics.RecordFailed(ctx, eventDeliveryID(req.IdempotencyKey), providerName,
			"provider_rejected_recipient", s.now().UTC()); err != nil {
			return fmt.Errorf("record failed event notification milestone: %w", err)
		}
	}

	s.logger.InfoContext(ctx, "notification.event.delivery_recorded",
		slog.String("event_id", req.SourceEventID),
		slog.String("event_type", req.SourceType),
		slog.String("trace_id", strings.TrimSpace(req.TraceID)),
		slog.String("channel", string(req.Channel)),
		slog.String("template_key", string(req.TemplateKey)),
		slog.String("status", string(status)),
	)
	if safeSendErr != nil {
		return fmt.Errorf("send event notification: %w", safeSendErr)
	}
	return nil
}

func eventMessage(req domain.EventNotificationTrigger, rendered domain.RenderedMessage) (domain.Message, error) {
	message := domain.Message{
		Channel:       req.Channel,
		CorrelationID: strings.TrimSpace(req.TraceID),
		Content: domain.Content{
			Subject:  rendered.Subject,
			TextBody: rendered.Body,
		},
	}
	switch req.Channel {
	case domain.ChannelEmail:
		message.Recipient.Email = strings.TrimSpace(req.Recipient)
	default:
		return domain.Message{}, domain.ErrUnsupportedChannel
	}
	if err := message.Validate(); err != nil {
		return domain.Message{}, err
	}
	return message, nil
}

func newEventDeliveryClaim(req domain.EventNotificationTrigger, now time.Time) domain.Delivery {
	return domain.Delivery{
		ID:              eventDeliveryID(req.IdempotencyKey),
		UserID:          strings.TrimSpace(req.UserID),
		Channel:         req.Channel,
		TemplateKey:     string(req.TemplateKey),
		Status:          domain.DeliveryStatusProcessing,
		Attempts:        0,
		IdempotencyKey:  req.IdempotencyKey,
		SourceEventID:   req.SourceEventID,
		SourceEventType: req.SourceType,
		TraceID:         strings.TrimSpace(req.TraceID),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func newSuppressedEventDelivery(
	req domain.EventNotificationTrigger,
	reason domain.ConsentReason,
	now time.Time,
) domain.Delivery {
	delivery := newEventDeliveryClaim(req, now)
	delivery.Status = domain.DeliveryStatusSuppressed
	delivery.SuppressionReason = reason
	return delivery
}

func eventDeliveryOutcome(result provider.Result, sendErr error) (domain.DeliveryStatus, error) {
	if sendErr == nil && result.Status == provider.StatusAccepted {
		return domain.DeliveryStatusAccepted, nil
	}
	if errors.Is(sendErr, provider.ErrProviderRejected) {
		return domain.DeliveryStatusRejected, provider.ErrProviderRejected
	}
	return domain.DeliveryStatusFailed, provider.ErrProviderUnavailable
}
