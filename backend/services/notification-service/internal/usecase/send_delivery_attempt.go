package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/provider"
)

type RetryDeliveryLookup interface {
	FindDeliveryByID(ctx context.Context, id string) (domain.Delivery, error)
}

type DeliveryAttemptService struct {
	renderer   Renderer
	providers  EventProviderRegistry
	deliveries RetryDeliveryLookup
	protector  RecipientProtector
	consent    *ConsentGate
}

func NewDeliveryAttemptService(
	renderer Renderer,
	providers EventProviderRegistry,
	deliveries RetryDeliveryLookup,
	protector RecipientProtector,
	consent *ConsentGate,
) (*DeliveryAttemptService, error) {
	if nilDependency(renderer) || nilDependency(providers) ||
		nilDependency(deliveries) || nilDependency(protector) || consent == nil {
		return nil, errors.New("notification delivery attempt dependencies are required")
	}
	return &DeliveryAttemptService{
		renderer: renderer, providers: providers, deliveries: deliveries, protector: protector, consent: consent,
	}, nil
}

func (s *DeliveryAttemptService) SendDeliveryAttempt(ctx context.Context, deliveryID string) (provider.Result, error) {
	if ctx == nil {
		return provider.Result{}, fmt.Errorf("%w: context is required", domain.ErrInvalidRetryInstruction)
	}
	delivery, err := s.deliveries.FindDeliveryByID(ctx, strings.TrimSpace(deliveryID))
	if err != nil {
		if errors.Is(err, domain.ErrDeliveryNotFound) {
			return provider.Result{}, fmt.Errorf("%w: delivery is missing", domain.ErrInvalidRetryInstruction)
		}
		return provider.Result{}, err
	}
	if delivery.TemplateKey == string(domain.TemplateOTPVerification) {
		return provider.Result{}, domain.ErrDurableOTPRetryForbidden
	}
	if strings.TrimSpace(delivery.RecipientCiphertext) == "" || delivery.MaxAttempts < 1 {
		return provider.Result{}, fmt.Errorf("%w: delivery is not retry-safe", domain.ErrInvalidRetryInstruction)
	}
	decision, err := s.consent.Evaluate(ctx, delivery.UserID, delivery.Channel,
		domain.TemplateKey(delivery.TemplateKey), false)
	if err != nil {
		return provider.Result{}, err
	}
	if !decision.Allowed {
		return provider.Result{}, &domain.DeliverySuppressedError{
			Reason: decision.Reason, Purpose: decision.Purpose,
		}
	}
	recipient, err := s.protector.Reveal(delivery.RecipientCiphertext, delivery.ID)
	if err != nil {
		return provider.Result{}, fmt.Errorf("%w: recipient unavailable", domain.ErrInvalidRetryInstruction)
	}
	variables, err := retryVariables(delivery.Payload)
	if err != nil {
		return provider.Result{}, err
	}
	rendered, err := s.renderer.Render(ctx, domain.RenderRequest{
		TemplateKey: domain.TemplateKey(delivery.TemplateKey),
		Channel:     delivery.Channel,
		Variables:   variables,
	})
	if err != nil {
		return provider.Result{}, err
	}
	message, err := eventMessage(domain.EventNotificationTrigger{
		Channel: delivery.Channel, Recipient: recipient, TraceID: delivery.TraceID,
	}, rendered)
	if err != nil {
		return provider.Result{}, fmt.Errorf("%w: %w", domain.ErrInvalidRetryInstruction, err)
	}
	sender, err := s.providers.Resolve(delivery.Channel)
	if err != nil {
		return provider.Result{}, err
	}
	result, sendErr := sender.Send(ctx, message)
	if strings.TrimSpace(result.ProviderName) == "" {
		result.ProviderName = sender.Name()
		result.Channel = delivery.Channel
	}
	return result, sendErr
}

func retryVariables(payload map[string]any) (map[string]string, error) {
	if len(payload) == 0 {
		return nil, fmt.Errorf("%w: template variables are missing", domain.ErrInvalidRetryInstruction)
	}
	variables := make(map[string]string, len(payload))
	for key, value := range payload {
		text, ok := value.(string)
		if !ok || strings.TrimSpace(key) == "" || strings.TrimSpace(text) == "" {
			return nil, fmt.Errorf("%w: stored template variables are invalid", domain.ErrInvalidRetryInstruction)
		}
		variables[key] = text
	}
	return variables, nil
}
