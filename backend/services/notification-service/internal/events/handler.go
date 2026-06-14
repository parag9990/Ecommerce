package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/provider"
)

type Outcome string

const (
	OutcomeProcessed   Outcome = "processed"
	OutcomeDuplicate   Outcome = "duplicate"
	OutcomeUnsupported Outcome = "unsupported"
)

type TriggerSender interface {
	TriggerFromEvent(ctx context.Context, req domain.EventNotificationTrigger) error
}

type Handler struct {
	sender TriggerSender
	logger *slog.Logger
}

func NewHandler(sender TriggerSender, logger *slog.Logger) (*Handler, error) {
	if sender == nil {
		return nil, errors.New("event notification sender is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{sender: sender, logger: logger}, nil
}

func (h *Handler) Handle(ctx context.Context, body []byte) (Outcome, error) {
	if ctx == nil {
		return "", fmt.Errorf("%w: context is required", ErrInvalidEvent)
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	envelope, err := DecodeEnvelope(body)
	if err != nil {
		return "", err
	}
	trigger, supported, err := BuildTrigger(envelope)
	if err != nil {
		h.logOutcome(ctx, envelope, "", "invalid")
		return "", err
	}
	if !supported {
		h.logOutcome(ctx, envelope, OutcomeUnsupported, "")
		return OutcomeUnsupported, nil
	}

	err = h.sender.TriggerFromEvent(ctx, trigger)
	switch {
	case err == nil:
		h.logOutcome(ctx, envelope, OutcomeProcessed, "")
		return OutcomeProcessed, nil
	case errors.Is(err, domain.ErrDuplicateEventDelivery):
		h.logOutcome(ctx, envelope, OutcomeDuplicate, "")
		return OutcomeDuplicate, nil
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return "", err
	case nonRetryableTriggerFailure(err):
		h.logOutcome(ctx, envelope, "", "non_retryable")
		return "", fmt.Errorf("%w: trigger event notification: %w", ErrNonRetryableProcessing, err)
	default:
		h.logOutcome(ctx, envelope, "", "temporary")
		return "", fmt.Errorf("%w: trigger event notification: %w", ErrTemporaryProcessing, err)
	}
}

func nonRetryableTriggerFailure(err error) bool {
	return errors.Is(err, domain.ErrInvalidEventTrigger) ||
		errors.Is(err, domain.ErrInvalidDelivery) ||
		errors.Is(err, domain.ErrDurableOTPRetryForbidden) ||
		errors.Is(err, domain.ErrInvalidRecipient) ||
		errors.Is(err, domain.ErrInvalidContent) ||
		errors.Is(err, domain.ErrUnsupportedChannel) ||
		errors.Is(err, domain.ErrUnsupportedTemplateKey) ||
		errors.Is(err, domain.ErrInvalidRenderRequest) ||
		errors.Is(err, domain.ErrTemplateNotFound) ||
		errors.Is(err, domain.ErrTemplateRender) ||
		errors.Is(err, provider.ErrProviderRejected) ||
		errors.Is(err, provider.ErrProviderNotConfigured) ||
		errors.Is(err, provider.ErrChannelDisabled)
}

func (h *Handler) logOutcome(ctx context.Context, envelope Envelope, outcome Outcome, category string) {
	attributes := []any{
		slog.String("event_id", envelope.EventID),
		slog.String("event_type", envelope.EventType),
		slog.String("trace_id", envelope.TraceID),
	}
	if outcome != "" {
		attributes = append(attributes, slog.String("outcome", string(outcome)))
		h.logger.InfoContext(ctx, "notification.event.handled", attributes...)
		return
	}
	attributes = append(attributes, slog.String("failure_category", category))
	h.logger.WarnContext(ctx, "notification.event.failed", attributes...)
}
