package providertest

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider"
)

const ProviderNameFake = "fake"

type FakeProvider struct {
	NameValue             string
	CreateIntentResponse  provider.CreateIntentResponse
	CaptureResponse       provider.CaptureResponse
	RefundResponse        provider.RefundResponse
	WebhookEvent          provider.WebhookEvent
	CreateIntentErr       error
	CaptureErr            error
	RefundErr             error
	VerifyWebhookErr      error
	CreateIntentCalls     int
	CaptureCalls          int
	RefundCalls           int
	VerifyWebhookCalls    int
	LastCreateIntent      provider.CreateIntentRequest
	LastCapture           provider.CaptureRequest
	LastRefund            provider.RefundRequest
	LastVerifyWebhook     provider.VerifyWebhookRequest
	SkipRequestValidation bool
}

func (p *FakeProvider) Name() string {
	if strings.TrimSpace(p.NameValue) != "" {
		return strings.TrimSpace(p.NameValue)
	}
	return ProviderNameFake
}

func (p *FakeProvider) CreateIntent(ctx context.Context, req provider.CreateIntentRequest) (provider.CreateIntentResponse, error) {
	if err := ctx.Err(); err != nil {
		return provider.CreateIntentResponse{}, err
	}
	p.CreateIntentCalls++
	p.LastCreateIntent = req.Normalized()
	if !p.SkipRequestValidation {
		if err := provider.ValidateCreateIntentRequest(req); err != nil {
			return provider.CreateIntentResponse{}, err
		}
	}
	if p.CreateIntentErr != nil {
		return provider.CreateIntentResponse{}, p.CreateIntentErr
	}
	res := p.CreateIntentResponse.Normalized()
	if res.Provider == "" {
		res.Provider = p.Name()
	}
	if res.ProviderIntentID == "" {
		res.ProviderIntentID = "fake_intent_" + p.LastCreateIntent.PaymentID
	}
	if res.Status == "" {
		res.Status = provider.IntentStatusInitiated
	}
	if len(res.FrontendPayload) == 0 {
		res.FrontendPayload = map[string]string{"mode": "fake"}
	}
	if len(res.RawProviderResponse) == 0 {
		res.RawProviderResponse = json.RawMessage(`{"provider":"fake"}`)
	}
	if err := provider.ValidateCreateIntentResponse(res); err != nil {
		return provider.CreateIntentResponse{}, err
	}
	return res, nil
}

func (p *FakeProvider) Capture(ctx context.Context, req provider.CaptureRequest) (provider.CaptureResponse, error) {
	if err := ctx.Err(); err != nil {
		return provider.CaptureResponse{}, err
	}
	p.CaptureCalls++
	p.LastCapture = req.Normalized()
	if !p.SkipRequestValidation {
		if err := provider.ValidateCaptureRequest(req); err != nil {
			return provider.CaptureResponse{}, err
		}
	}
	if p.CaptureErr != nil {
		return provider.CaptureResponse{}, p.CaptureErr
	}
	res := p.CaptureResponse.Normalized()
	if res.ProviderPaymentID == "" {
		res.ProviderPaymentID = p.LastCapture.ProviderPaymentID
	}
	if res.ProviderPaymentID == "" {
		res.ProviderPaymentID = "fake_payment_" + p.LastCapture.PaymentID
	}
	if res.Status == "" {
		res.Status = provider.IntentStatusCaptured
	}
	if res.CapturedAmount.AmountMinor == 0 {
		res.CapturedAmount = p.LastCapture.Amount
	}
	if len(res.RawProviderResponse) == 0 {
		res.RawProviderResponse = json.RawMessage(`{"provider":"fake","status":"captured"}`)
	}
	if err := provider.ValidateCaptureResponse(res); err != nil {
		return provider.CaptureResponse{}, err
	}
	return res, nil
}

func (p *FakeProvider) Refund(ctx context.Context, req provider.RefundRequest) (provider.RefundResponse, error) {
	if err := ctx.Err(); err != nil {
		return provider.RefundResponse{}, err
	}
	p.RefundCalls++
	p.LastRefund = req.Normalized()
	if !p.SkipRequestValidation {
		if err := provider.ValidateRefundRequest(req); err != nil {
			return provider.RefundResponse{}, err
		}
	}
	if p.RefundErr != nil {
		return provider.RefundResponse{}, p.RefundErr
	}
	res := p.RefundResponse.Normalized()
	if res.ProviderRefundID == "" {
		res.ProviderRefundID = "fake_refund_" + p.LastRefund.RefundID
	}
	if res.Status == "" {
		res.Status = provider.RefundStatusSucceeded
	}
	if res.RefundedAmount.AmountMinor == 0 {
		res.RefundedAmount = p.LastRefund.Amount
	}
	if len(res.RawProviderResponse) == 0 {
		res.RawProviderResponse = json.RawMessage(`{"provider":"fake","status":"succeeded"}`)
	}
	if err := provider.ValidateRefundResponse(res); err != nil {
		return provider.RefundResponse{}, err
	}
	return res, nil
}

func (p *FakeProvider) VerifyWebhook(ctx context.Context, req provider.VerifyWebhookRequest) (provider.WebhookEvent, error) {
	if err := ctx.Err(); err != nil {
		return provider.WebhookEvent{}, err
	}
	p.VerifyWebhookCalls++
	p.LastVerifyWebhook = req.Normalized()
	if !p.SkipRequestValidation {
		if err := provider.ValidateVerifyWebhookRequest(req); err != nil {
			return provider.WebhookEvent{}, err
		}
	}
	if p.VerifyWebhookErr != nil {
		return provider.WebhookEvent{}, p.VerifyWebhookErr
	}
	event := p.WebhookEvent.Normalized()
	if event.Provider == "" {
		event.Provider = p.Name()
	}
	if event.ProviderEventID == "" {
		event.ProviderEventID = "fake_event"
	}
	if event.Type == "" {
		event.Type = provider.WebhookEventPaymentCaptured
	}
	if event.ProviderPaymentID == "" {
		event.ProviderPaymentID = "fake_payment"
	}
	if event.Amount.AmountMinor == 0 {
		event.Amount = provider.Money{AmountMinor: 100, Currency: "INR"}
	}
	if len(event.RawPayload) == 0 {
		event.RawPayload = json.RawMessage(`{"provider":"fake","type":"payment.captured"}`)
	}
	if err := provider.ValidateWebhookEvent(event); err != nil {
		return provider.WebhookEvent{}, err
	}
	return event, nil
}
