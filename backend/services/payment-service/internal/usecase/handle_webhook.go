package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider"
)

const PaymentEventsTopic = "payment.events"

type WebhookRepository interface {
	ProcessWebhook(ctx context.Context, event domain.VerifiedPaymentWebhook) (domain.WebhookProcessingResult, error)
	ProcessRefundWebhook(ctx context.Context, event domain.VerifiedRefundWebhook) (domain.RefundWebhookProcessingResult, error)
}

type PaymentEventPublisher interface {
	Publish(ctx context.Context, topic string, event domain.PaymentDomainEvent) error
}

type WebhookIDGenerator interface {
	NewWebhookEventID() (string, error)
}

type HandleWebhookUsecase struct {
	repository WebhookRepository
	providers  PaymentProviderRegistry
	publisher  PaymentEventPublisher
	ids        WebhookIDGenerator
	clock      func() time.Time
	logger     *slog.Logger
}

type HandleWebhookOption func(*HandleWebhookUsecase)

func WithWebhookIDGenerator(ids WebhookIDGenerator) HandleWebhookOption {
	return func(u *HandleWebhookUsecase) {
		if ids != nil {
			u.ids = ids
		}
	}
}

func WithWebhookClock(clock func() time.Time) HandleWebhookOption {
	return func(u *HandleWebhookUsecase) {
		if clock != nil {
			u.clock = clock
		}
	}
}

type HandleWebhookInput struct {
	Provider string
	Headers  map[string]string
	RawBody  []byte
}

type HandleWebhookOutput struct {
	WebhookEventID   string
	ProviderEventID  string
	ProcessingStatus domain.WebhookProcessingStatus
	PaymentID        string
	Status           domain.PaymentStatus
	RefundID         string
	RefundStatus     domain.RefundStatus
}

func NewHandleWebhookUsecase(
	repository WebhookRepository,
	providers PaymentProviderRegistry,
	publisher PaymentEventPublisher,
	logger *slog.Logger,
	opts ...HandleWebhookOption,
) (*HandleWebhookUsecase, error) {
	if repository == nil {
		return nil, errors.New("webhook repository is required")
	}
	if providers == nil {
		return nil, errors.New("payment provider registry is required")
	}
	if publisher == nil {
		return nil, errors.New("payment event publisher is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	u := &HandleWebhookUsecase{
		repository: repository,
		providers:  providers,
		publisher:  publisher,
		ids:        cryptoWebhookIDGenerator{},
		clock:      func() time.Time { return time.Now().UTC() },
		logger:     logger,
	}
	for _, opt := range opts {
		opt(u)
	}
	return u, nil
}

func (u *HandleWebhookUsecase) Execute(ctx context.Context, input HandleWebhookInput) (HandleWebhookOutput, error) {
	if err := ctx.Err(); err != nil {
		return HandleWebhookOutput{}, err
	}
	input.Provider = provider.NormalizeProviderName(input.Provider)
	if input.Provider == "" {
		return HandleWebhookOutput{}, fmt.Errorf("%w: provider is required", provider.ErrInvalidProviderRequest)
	}
	gateway, err := u.providers.Get(input.Provider)
	if err != nil {
		return HandleWebhookOutput{}, err
	}
	verified, err := gateway.VerifyWebhook(ctx, provider.VerifyWebhookRequest{
		Headers: input.Headers,
		Body:    append([]byte(nil), input.RawBody...),
	})
	if err != nil {
		return HandleWebhookOutput{}, err
	}
	verified = verified.Normalized()
	if verified.Provider != input.Provider {
		return HandleWebhookOutput{}, fmt.Errorf("%w: verified webhook provider does not match route provider", provider.ErrInvalidProviderRequest)
	}
	if verified.Type == provider.WebhookEventRefundSucceeded || verified.Type == provider.WebhookEventRefundFailed {
		return u.executeRefundWebhook(ctx, verified)
	}
	event, err := u.verifiedDomainEvent(verified)
	if err != nil {
		return HandleWebhookOutput{}, err
	}
	result, err := u.repository.ProcessWebhook(ctx, event)
	if err != nil {
		return HandleWebhookOutput{}, err
	}

	if published, ok := result.DomainEvent(event.OccurredAt); ok {
		if err := u.publisher.Publish(ctx, PaymentEventsTopic, published); err != nil {
			u.logger.ErrorContext(ctx, "payment.webhook.event_publish_failed",
				slog.String("webhook_event_id", result.WebhookEventID),
				slog.String("payment_id", result.Payment.PaymentID),
				slog.String("provider", event.Provider),
				slog.String("event_type", published.EventType),
				slog.String("error", err.Error()),
			)
		}
	}
	u.logger.InfoContext(ctx, "payment.webhook.processed",
		slog.String("webhook_event_id", result.WebhookEventID),
		slog.String("provider_event_id", result.ProviderEventID),
		slog.String("provider", event.Provider),
		slog.String("payment_id", result.Payment.PaymentID),
		slog.String("processing_status", string(result.ProcessingStatus)),
		slog.String("payment_status", string(result.Payment.Status)),
	)
	return HandleWebhookOutput{
		WebhookEventID:   result.WebhookEventID,
		ProviderEventID:  result.ProviderEventID,
		ProcessingStatus: result.ProcessingStatus,
		PaymentID:        result.Payment.PaymentID,
		Status:           result.Payment.Status,
	}, nil
}

func (u *HandleWebhookUsecase) executeRefundWebhook(ctx context.Context, verified provider.WebhookEvent) (HandleWebhookOutput, error) {
	event, err := u.verifiedRefundDomainEvent(verified)
	if err != nil {
		return HandleWebhookOutput{}, err
	}
	result, err := u.repository.ProcessRefundWebhook(ctx, event)
	if err != nil {
		return HandleWebhookOutput{}, err
	}
	if published, ok := result.DomainEvent(event.OccurredAt); ok {
		if err := u.publisher.Publish(ctx, PaymentEventsTopic, published); err != nil {
			u.logger.ErrorContext(ctx, "payment.refund_webhook.event_publish_failed",
				slog.String("webhook_event_id", result.WebhookEventID),
				slog.String("refund_id", result.Refund.RefundID),
				slog.String("payment_id", result.Payment.PaymentID),
				slog.String("provider", event.Provider),
				slog.String("error", err.Error()),
			)
		}
	}
	u.logger.InfoContext(ctx, "payment.refund_webhook.processed",
		slog.String("webhook_event_id", result.WebhookEventID),
		slog.String("provider_event_id", result.ProviderEventID),
		slog.String("provider", event.Provider),
		slog.String("refund_id", result.Refund.RefundID),
		slog.String("payment_id", result.Payment.PaymentID),
		slog.String("processing_status", string(result.ProcessingStatus)),
		slog.String("refund_status", string(result.Refund.Status)),
	)
	return HandleWebhookOutput{
		WebhookEventID:   result.WebhookEventID,
		ProviderEventID:  result.ProviderEventID,
		ProcessingStatus: result.ProcessingStatus,
		PaymentID:        result.Payment.PaymentID,
		Status:           result.Payment.Status,
		RefundID:         result.Refund.RefundID,
		RefundStatus:     result.Refund.Status,
	}, nil
}

func (u *HandleWebhookUsecase) verifiedDomainEvent(event provider.WebhookEvent) (domain.VerifiedPaymentWebhook, error) {
	internalEvent, err := webhookPaymentEvent(event.Type)
	if err != nil {
		return domain.VerifiedPaymentWebhook{}, err
	}
	webhookEventID, err := u.ids.NewWebhookEventID()
	if err != nil {
		return domain.VerifiedPaymentWebhook{}, err
	}
	receivedAt := u.clock().UTC()
	occurredAt := receivedAt
	if event.OccurredAtUnix > 0 {
		occurredAt = time.Unix(event.OccurredAtUnix, 0).UTC()
	}
	verified := domain.VerifiedPaymentWebhook{
		WebhookEventID:    webhookEventID,
		Provider:          event.Provider,
		ProviderEventID:   event.ProviderEventID,
		EventType:         string(event.Type),
		Event:             internalEvent,
		PaymentID:         event.PaymentID,
		OrderID:           event.OrderID,
		ProviderIntentID:  event.ProviderIntentID,
		ProviderPaymentID: event.ProviderPaymentID,
		Amount: domain.Money{
			Amount:   event.Amount.AmountMinor,
			Currency: event.Amount.Currency,
		},
		FailureCode: event.FailureCode,
		Payload:     append([]byte(nil), event.RawPayload...),
		OccurredAt:  occurredAt,
		ReceivedAt:  receivedAt,
	}
	if err := verified.Validate(); err != nil {
		return domain.VerifiedPaymentWebhook{}, err
	}
	return verified.Normalized(), nil
}

func (u *HandleWebhookUsecase) verifiedRefundDomainEvent(event provider.WebhookEvent) (domain.VerifiedRefundWebhook, error) {
	outcome := domain.RefundStatusFailed
	if event.Type == provider.WebhookEventRefundSucceeded {
		outcome = domain.RefundStatusSucceeded
	}
	webhookEventID, err := u.ids.NewWebhookEventID()
	if err != nil {
		return domain.VerifiedRefundWebhook{}, err
	}
	receivedAt := u.clock().UTC()
	occurredAt := receivedAt
	if event.OccurredAtUnix > 0 {
		occurredAt = time.Unix(event.OccurredAtUnix, 0).UTC()
	}
	verified := domain.VerifiedRefundWebhook{
		WebhookEventID:    webhookEventID,
		Provider:          event.Provider,
		ProviderEventID:   event.ProviderEventID,
		EventType:         string(event.Type),
		Outcome:           outcome,
		RefundID:          event.RefundID,
		PaymentID:         event.PaymentID,
		ProviderPaymentID: event.ProviderPaymentID,
		ProviderRefundID:  event.ProviderRefundID,
		Amount:            domain.Money{Amount: event.Amount.AmountMinor, Currency: event.Amount.Currency},
		FailureCode:       event.FailureCode,
		Payload:           append([]byte(nil), event.RawPayload...),
		OccurredAt:        occurredAt,
		ReceivedAt:        receivedAt,
	}
	if err := verified.Validate(); err != nil {
		return domain.VerifiedRefundWebhook{}, err
	}
	return verified.Normalized(), nil
}

func webhookPaymentEvent(eventType provider.WebhookEventType) (domain.PaymentEvent, error) {
	switch eventType.Normalized() {
	case provider.WebhookEventPaymentRequiresAction:
		return domain.PaymentEventProviderRequiresAction, nil
	case provider.WebhookEventPaymentAuthorized:
		return domain.PaymentEventProviderAuthorized, nil
	case provider.WebhookEventPaymentCaptured:
		return domain.PaymentEventProviderCaptured, nil
	case provider.WebhookEventPaymentFailed:
		return domain.PaymentEventProviderFailed, nil
	default:
		return "", fmt.Errorf("%w: event type %s is outside webhook payment outcome scope", domain.ErrInvalidWebhookEvent, eventType)
	}
}

type cryptoWebhookIDGenerator struct{}

func (cryptoWebhookIDGenerator) NewWebhookEventID() (string, error) {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate webhook event identifier: %w", err)
	}
	return "whe_" + strings.ToLower(hex.EncodeToString(random)), nil
}
