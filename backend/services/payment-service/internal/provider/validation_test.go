package provider_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider"
)

func TestValidateCreateIntentRequestRejectsInvalidAmount(t *testing.T) {
	req := validCreateIntentRequest()
	req.Amount.AmountMinor = 0

	err := provider.ValidateCreateIntentRequest(req)
	if !errors.Is(err, provider.ErrInvalidProviderRequest) {
		t.Fatalf("ValidateCreateIntentRequest() error = %v, want invalid provider request", err)
	}
}

func TestValidateCreateIntentRequestRejectsSensitiveMetadata(t *testing.T) {
	req := validCreateIntentRequest()
	req.Metadata["secret_key"] = "sk_test"

	err := provider.ValidateCreateIntentRequest(req)
	if !errors.Is(err, provider.ErrInvalidProviderRequest) {
		t.Fatalf("ValidateCreateIntentRequest() error = %v, want invalid provider request", err)
	}
}

func TestValidateRawProviderResponseRejectsSensitiveKeys(t *testing.T) {
	err := provider.ValidateSanitizedJSONPayload(json.RawMessage(`{"client_secret":"should_not_store"}`), false, "raw_provider_response")
	if !errors.Is(err, provider.ErrInvalidProviderRequest) {
		t.Fatalf("ValidateSanitizedJSONPayload() error = %v, want invalid provider request", err)
	}
}

func TestValidateWebhookEventRequiresProviderObjectID(t *testing.T) {
	err := provider.ValidateWebhookEvent(provider.WebhookEvent{
		Provider:        provider.ProviderNameStripeLike,
		ProviderEventID: "evt_123",
		Type:            provider.WebhookEventPaymentCaptured,
		Amount:          provider.Money{AmountMinor: 1000, Currency: "INR"},
		RawPayload:      json.RawMessage(`{"type":"payment.captured"}`),
	})
	if !errors.Is(err, provider.ErrInvalidProviderRequest) {
		t.Fatalf("ValidateWebhookEvent() error = %v, want invalid provider request", err)
	}
}

func TestValidateCapturedWebhookRequiresAmount(t *testing.T) {
	err := provider.ValidateWebhookEvent(provider.WebhookEvent{
		Provider:         provider.ProviderNameStripeLike,
		ProviderEventID:  "evt_123",
		Type:             provider.WebhookEventPaymentCaptured,
		ProviderIntentID: "pi_123",
		RawPayload:       json.RawMessage(`{"type":"payment.captured"}`),
	})
	if !errors.Is(err, provider.ErrInvalidProviderRequest) {
		t.Fatalf("ValidateWebhookEvent() error = %v, want missing captured amount rejected", err)
	}
}

func TestValidateCreateIntentResponseAllowsFrontendClientSecret(t *testing.T) {
	res := provider.CreateIntentResponse{
		Provider:         provider.ProviderNameStripeLike,
		ProviderIntentID: "pi_123",
		Status:           provider.IntentStatusRequiresAction,
		ClientSecret:     "client_secret_frontend_safe",
		FrontendPayload: map[string]string{
			"publishable_key": "pk_test",
		},
		RawProviderResponse: json.RawMessage(`{"id":"pi_123","status":"requires_action"}`),
	}
	if err := provider.ValidateCreateIntentResponse(res); err != nil {
		t.Fatalf("ValidateCreateIntentResponse() error = %v", err)
	}
}
