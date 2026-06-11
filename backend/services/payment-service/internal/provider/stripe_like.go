package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const defaultStripeLikeBaseURL = "https://api.stripe.com"

type StripeLikeProvider struct {
	cfg    ProviderConfig
	client *http.Client
	logger *slog.Logger
}

func NewStripeLikeProvider(cfg ProviderConfig, logger *slog.Logger) (*StripeLikeProvider, error) {
	cfg = cfg.Normalized()
	if err := cfg.Validate(ProviderNameStripeLike); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &StripeLikeProvider{
		cfg:    cfg,
		client: &http.Client{Timeout: cfg.Timeout},
		logger: logger,
	}, nil
}

func (p *StripeLikeProvider) Name() string {
	return ProviderNameStripeLike
}

func (p *StripeLikeProvider) CreateIntent(ctx context.Context, req CreateIntentRequest) (CreateIntentResponse, error) {
	if err := ctx.Err(); err != nil {
		return CreateIntentResponse{}, err
	}
	if err := p.ValidateCreateIntent(req); err != nil {
		return CreateIntentResponse{}, err
	}
	req = req.Normalized()
	endpoint, err := providerEndpoint(p.cfg.BaseURL, defaultStripeLikeBaseURL, "/v1/payment_intents")
	if err != nil {
		return CreateIntentResponse{}, err
	}

	values := url.Values{
		"amount":         {strconv.FormatInt(req.Amount.AmountMinor, 10)},
		"currency":       {strings.ToLower(req.Amount.Currency)},
		"capture_method": {string(req.CaptureMode)},
	}
	if req.Customer.Email != "" {
		values.Set("receipt_email", req.Customer.Email)
	}
	for key, value := range req.Metadata {
		values.Set("metadata["+key+"]", value)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return CreateIntentResponse{}, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Idempotency-Key", req.IdempotencyKey)
	request.SetBasicAuth(p.cfg.SecretKey, "")

	body, err := executeProviderRequest(ctx, p.client, request, p.Name(), "create_intent")
	if err != nil {
		return CreateIntentResponse{}, err
	}
	return p.intentResponse(body, "create_intent")
}

func (p *StripeLikeProvider) ValidateCreateIntent(req CreateIntentRequest) error {
	if err := ValidateCreateIntentRequest(req); err != nil {
		return err
	}
	req = req.Normalized()
	if len(req.Metadata) > 50 {
		return fmt.Errorf("%w: Stripe metadata exceeds 50 entries", ErrInvalidProviderRequest)
	}
	for key, value := range req.Metadata {
		if len(key) > 40 || len(value) > 500 {
			return fmt.Errorf("%w: Stripe metadata entry %q exceeds provider limits", ErrInvalidProviderRequest, key)
		}
	}
	return nil
}

func (p *StripeLikeProvider) RetrieveIntent(ctx context.Context, providerIntentID string) (CreateIntentResponse, error) {
	if err := ctx.Err(); err != nil {
		return CreateIntentResponse{}, err
	}
	if err := validateID("provider_intent_id", providerIntentID); err != nil {
		return CreateIntentResponse{}, err
	}
	endpoint, err := providerEndpoint(p.cfg.BaseURL, defaultStripeLikeBaseURL, "/v1/payment_intents/"+url.PathEscape(strings.TrimSpace(providerIntentID)))
	if err != nil {
		return CreateIntentResponse{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return CreateIntentResponse{}, err
	}
	request.SetBasicAuth(p.cfg.SecretKey, "")
	body, err := executeProviderRequest(ctx, p.client, request, p.Name(), "retrieve_intent")
	if err != nil {
		return CreateIntentResponse{}, err
	}
	return p.intentResponse(body, "retrieve_intent")
}

func (p *StripeLikeProvider) intentResponse(body []byte, operation string) (CreateIntentResponse, error) {
	var response struct {
		ID           string `json:"id"`
		Status       string `json:"status"`
		ClientSecret string `json:"client_secret"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return CreateIntentResponse{}, NewError(p.Name(), operation, ErrorCodeUnavailable, "payment provider returned invalid JSON", WithCause(err))
	}
	status, ok := NormalizeStripeLikeStatus(response.Status)
	if !ok {
		return CreateIntentResponse{}, NewError(p.Name(), operation, ErrorCodeUnknownStatus, "payment provider returned an unknown intent status")
	}
	audit, err := sanitizedAuditPayload(map[string]any{
		"id":     response.ID,
		"status": response.Status,
	})
	if err != nil {
		return CreateIntentResponse{}, err
	}
	result := CreateIntentResponse{
		Provider:            p.Name(),
		ProviderIntentID:    response.ID,
		Status:              status,
		ClientSecret:        response.ClientSecret,
		RawProviderResponse: audit,
	}
	if err := ValidateCreateIntentResponse(result); err != nil {
		return CreateIntentResponse{}, err
	}
	return result, nil
}

func (p *StripeLikeProvider) Capture(ctx context.Context, req CaptureRequest) (CaptureResponse, error) {
	if err := ctx.Err(); err != nil {
		return CaptureResponse{}, err
	}
	if err := ValidateCaptureRequest(req); err != nil {
		return CaptureResponse{}, err
	}
	p.logUnsupported(ctx, "capture")
	return CaptureResponse{}, NewUnsupportedOperationError(p.Name(), "capture")
}

func (p *StripeLikeProvider) Refund(ctx context.Context, req RefundRequest) (RefundResponse, error) {
	if err := ctx.Err(); err != nil {
		return RefundResponse{}, err
	}
	if err := ValidateRefundRequest(req); err != nil {
		return RefundResponse{}, err
	}
	req = req.Normalized()
	endpoint, err := providerEndpoint(p.cfg.BaseURL, defaultStripeLikeBaseURL, "/v1/refunds")
	if err != nil {
		return RefundResponse{}, err
	}
	values := url.Values{
		"charge":               {req.ProviderPaymentID},
		"amount":               {strconv.FormatInt(req.Amount.AmountMinor, 10)},
		"metadata[refund_id]":  {req.RefundID},
		"metadata[payment_id]": {req.PaymentID},
		"metadata[reason]":     {req.Reason},
	}
	for key, value := range req.Metadata {
		values.Set("metadata["+key+"]", value)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return RefundResponse{}, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Idempotency-Key", req.IdempotencyKey)
	request.SetBasicAuth(p.cfg.SecretKey, "")
	body, err := executeProviderRequest(ctx, p.client, request, p.Name(), "refund")
	if err != nil {
		return RefundResponse{}, err
	}
	var response struct {
		ID       string `json:"id"`
		Status   string `json:"status"`
		Amount   int64  `json:"amount"`
		Currency string `json:"currency"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return RefundResponse{}, NewError(p.Name(), "refund", ErrorCodeUnavailable, "payment provider returned invalid JSON", WithCause(err))
	}
	status, ok := NormalizeStripeLikeRefundStatus(response.Status)
	if !ok {
		return RefundResponse{}, NewError(p.Name(), "refund", ErrorCodeUnknownStatus, "payment provider returned an unknown refund status")
	}
	audit, err := sanitizedAuditPayload(map[string]any{"id": response.ID, "status": response.Status, "amount": response.Amount, "currency": response.Currency})
	if err != nil {
		return RefundResponse{}, err
	}
	result := RefundResponse{
		ProviderRefundID:    response.ID,
		Status:              status,
		RefundedAmount:      Money{AmountMinor: response.Amount, Currency: response.Currency},
		RawProviderResponse: audit,
	}
	if err := ValidateRefundResponse(result); err != nil {
		return RefundResponse{}, err
	}
	return result.Normalized(), nil
}

func (p *StripeLikeProvider) VerifyWebhook(ctx context.Context, req VerifyWebhookRequest) (WebhookEvent, error) {
	if err := ctx.Err(); err != nil {
		return WebhookEvent{}, err
	}
	if err := ValidateVerifyWebhookRequest(req); err != nil {
		return WebhookEvent{}, err
	}
	req = req.Normalized()
	timestamp, signatures, err := stripeSignatureParts(req.Headers["stripe-signature"])
	if err != nil || !validWebhookTimestamp(timestamp, p.cfg.WebhookTimestampTolerance) {
		return WebhookEvent{}, invalidWebhookSignature(p.Name())
	}
	signedPayload := []byte(strconv.FormatInt(timestamp, 10) + "." + string(req.Body))
	verified := false
	for _, signature := range signatures {
		if validHMACSHA256(p.cfg.WebhookSecret, signedPayload, signature) {
			verified = true
			break
		}
	}
	if !verified {
		return WebhookEvent{}, invalidWebhookSignature(p.Name())
	}

	var envelope struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Created int64  `json:"created"`
		Data    struct {
			Object struct {
				ID               string            `json:"id"`
				Status           string            `json:"status"`
				Amount           int64             `json:"amount"`
				AmountReceived   int64             `json:"amount_received"`
				Currency         string            `json:"currency"`
				LatestCharge     string            `json:"latest_charge"`
				Charge           string            `json:"charge"`
				PaymentIntent    string            `json:"payment_intent"`
				Metadata         map[string]string `json:"metadata"`
				LastPaymentError struct {
					Code string `json:"code"`
				} `json:"last_payment_error"`
			} `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(req.Body, &envelope); err != nil {
		return WebhookEvent{}, invalidWebhookPayload(p.Name(), "payment webhook payload is malformed", err)
	}
	eventType, ok := stripeWebhookType(envelope.Type, envelope.Data.Object.Status)
	if !ok {
		return WebhookEvent{}, invalidWebhookPayload(p.Name(), "payment webhook event type is not supported", ErrInvalidProviderRequest)
	}
	object := envelope.Data.Object
	amount := object.Amount
	if eventType == WebhookEventPaymentCaptured && object.AmountReceived > 0 {
		amount = object.AmountReceived
	}
	audit, err := sanitizedAuditPayload(map[string]any{
		"provider_event_id":   envelope.ID,
		"event_type":          envelope.Type,
		"refund_id":           object.Metadata["refund_id"],
		"payment_id":          object.Metadata["payment_id"],
		"order_id":            object.Metadata["order_id"],
		"provider_intent_id":  coalesce(object.PaymentIntent, object.ID),
		"provider_payment_id": coalesce(object.Charge, object.LatestCharge),
		"provider_refund_id":  refundObjectID(eventType, object.ID),
		"amount":              amount,
		"currency":            object.Currency,
	})
	if err != nil {
		return WebhookEvent{}, err
	}
	event := WebhookEvent{
		Provider:          p.Name(),
		ProviderEventID:   envelope.ID,
		Type:              eventType,
		RefundID:          object.Metadata["refund_id"],
		PaymentID:         object.Metadata["payment_id"],
		OrderID:           object.Metadata["order_id"],
		ProviderIntentID:  coalesce(object.PaymentIntent, object.ID),
		ProviderPaymentID: coalesce(object.Charge, object.LatestCharge),
		ProviderRefundID:  refundObjectID(eventType, object.ID),
		Amount:            Money{AmountMinor: amount, Currency: object.Currency},
		FailureCode:       object.LastPaymentError.Code,
		OccurredAtUnix:    envelope.Created,
		RawPayload:        audit,
	}
	if err := ValidateWebhookEvent(event); err != nil {
		return WebhookEvent{}, invalidWebhookPayload(p.Name(), "payment webhook event is invalid", err)
	}
	return event.Normalized(), nil
}

func (p *StripeLikeProvider) logUnsupported(ctx context.Context, operation string) {
	p.logger.InfoContext(ctx, "payment.provider.operation_not_enabled",
		slog.String("provider", p.Name()),
		slog.String("operation", operation),
	)
}

func stripeWebhookType(raw string, status string) (WebhookEventType, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "payment_intent.requires_action":
		return WebhookEventPaymentRequiresAction, true
	case "payment_intent.authorized", "payment_intent.amount_capturable_updated":
		return WebhookEventPaymentAuthorized, true
	case "payment_intent.succeeded":
		return WebhookEventPaymentCaptured, true
	case "payment_intent.payment_failed", "payment_intent.canceled":
		return WebhookEventPaymentFailed, true
	case "refund.succeeded":
		return WebhookEventRefundSucceeded, true
	case "refund.failed":
		return WebhookEventRefundFailed, true
	case "refund.updated":
		normalized, ok := NormalizeStripeLikeRefundStatus(status)
		if !ok || normalized == RefundStatusProcessing {
			return "", false
		}
		if normalized == RefundStatusSucceeded {
			return WebhookEventRefundSucceeded, true
		}
		return WebhookEventRefundFailed, true
	default:
		return "", false
	}
}

func refundObjectID(eventType WebhookEventType, objectID string) string {
	if eventType == WebhookEventRefundSucceeded || eventType == WebhookEventRefundFailed {
		return objectID
	}
	return ""
}

func coalesce(value string, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}
