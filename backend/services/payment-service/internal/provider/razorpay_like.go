package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

const defaultRazorpayLikeBaseURL = "https://api.razorpay.com"

type RazorpayLikeProvider struct {
	cfg    ProviderConfig
	client *http.Client
	logger *slog.Logger
}

func NewRazorpayLikeProvider(cfg ProviderConfig, logger *slog.Logger) (*RazorpayLikeProvider, error) {
	cfg = cfg.Normalized()
	if err := cfg.Validate(ProviderNameRazorpayLike); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &RazorpayLikeProvider{
		cfg:    cfg,
		client: &http.Client{Timeout: cfg.Timeout},
		logger: logger,
	}, nil
}

func (p *RazorpayLikeProvider) Name() string {
	return ProviderNameRazorpayLike
}

func (p *RazorpayLikeProvider) CreateIntent(ctx context.Context, req CreateIntentRequest) (CreateIntentResponse, error) {
	if err := ctx.Err(); err != nil {
		return CreateIntentResponse{}, err
	}
	if err := p.ValidateCreateIntent(req); err != nil {
		return CreateIntentResponse{}, err
	}
	req = req.Normalized()
	endpoint, err := providerEndpoint(p.cfg.BaseURL, defaultRazorpayLikeBaseURL, "/v1/orders")
	if err != nil {
		return CreateIntentResponse{}, err
	}
	body, err := json.Marshal(map[string]any{
		"amount":   req.Amount.AmountMinor,
		"currency": req.Amount.Currency,
		"receipt":  req.PaymentID,
		"notes":    req.Metadata,
	})
	if err != nil {
		return CreateIntentResponse{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return CreateIntentResponse{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Razorpay-Idempotency-Key", req.IdempotencyKey)
	request.SetBasicAuth(p.cfg.PublicKey, p.cfg.SecretKey)

	responseBody, err := executeProviderRequest(ctx, p.client, request, p.Name(), "create_intent")
	if err != nil {
		return CreateIntentResponse{}, err
	}
	var response struct {
		ID      string `json:"id"`
		Status  string `json:"status"`
		Receipt string `json:"receipt"`
	}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return CreateIntentResponse{}, NewError(p.Name(), "create_intent", ErrorCodeUnavailable, "payment provider returned invalid JSON", WithCause(err))
	}
	status, ok := NormalizeRazorpayLikeStatus(response.Status)
	if !ok {
		return CreateIntentResponse{}, NewError(p.Name(), "create_intent", ErrorCodeUnknownStatus, "payment provider returned an unknown intent status")
	}
	audit, err := sanitizedAuditPayload(map[string]any{
		"id":      response.ID,
		"status":  response.Status,
		"receipt": response.Receipt,
	})
	if err != nil {
		return CreateIntentResponse{}, err
	}
	result := CreateIntentResponse{
		Provider:         p.Name(),
		ProviderIntentID: response.ID,
		Status:           status,
		FrontendPayload: map[string]string{
			"provider_order_id": response.ID,
		},
		RawProviderResponse: audit,
	}
	if err := ValidateCreateIntentResponse(result); err != nil {
		return CreateIntentResponse{}, err
	}
	return result, nil
}

func (p *RazorpayLikeProvider) ValidateCreateIntent(req CreateIntentRequest) error {
	if err := ValidateCreateIntentRequest(req); err != nil {
		return err
	}
	req = req.Normalized()
	if len(req.PaymentID) > 40 {
		return fmt.Errorf("%w: payment_id exceeds Razorpay receipt limit", ErrInvalidProviderRequest)
	}
	if len(req.Metadata) > 15 {
		return fmt.Errorf("%w: Razorpay notes exceed 15 entries", ErrInvalidProviderRequest)
	}
	for key, value := range req.Metadata {
		if len(key) > 256 || len(value) > 256 {
			return fmt.Errorf("%w: Razorpay note %q exceeds 256 characters", ErrInvalidProviderRequest, key)
		}
	}
	return nil
}

func (p *RazorpayLikeProvider) Capture(ctx context.Context, req CaptureRequest) (CaptureResponse, error) {
	if err := ctx.Err(); err != nil {
		return CaptureResponse{}, err
	}
	if err := ValidateCaptureRequest(req); err != nil {
		return CaptureResponse{}, err
	}
	p.logUnsupported(ctx, "capture")
	return CaptureResponse{}, NewUnsupportedOperationError(p.Name(), "capture")
}

func (p *RazorpayLikeProvider) Refund(ctx context.Context, req RefundRequest) (RefundResponse, error) {
	if err := ctx.Err(); err != nil {
		return RefundResponse{}, err
	}
	if err := ValidateRefundRequest(req); err != nil {
		return RefundResponse{}, err
	}
	req = req.Normalized()
	endpoint, err := providerEndpoint(p.cfg.BaseURL, defaultRazorpayLikeBaseURL, "/v1/payments/"+url.PathEscape(req.ProviderPaymentID)+"/refund")
	if err != nil {
		return RefundResponse{}, err
	}
	notes := make(map[string]string, len(req.Metadata)+3)
	for key, value := range req.Metadata {
		notes[key] = value
	}
	notes["refund_id"] = req.RefundID
	notes["payment_id"] = req.PaymentID
	notes["reason"] = req.Reason
	body, err := json.Marshal(map[string]any{"amount": req.Amount.AmountMinor, "notes": notes})
	if err != nil {
		return RefundResponse{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return RefundResponse{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Razorpay-Idempotency-Key", req.IdempotencyKey)
	request.SetBasicAuth(p.cfg.PublicKey, p.cfg.SecretKey)
	responseBody, err := executeProviderRequest(ctx, p.client, request, p.Name(), "refund")
	if err != nil {
		return RefundResponse{}, err
	}
	var response struct {
		ID       string `json:"id"`
		Status   string `json:"status"`
		Amount   int64  `json:"amount"`
		Currency string `json:"currency"`
	}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return RefundResponse{}, NewError(p.Name(), "refund", ErrorCodeUnavailable, "payment provider returned invalid JSON", WithCause(err))
	}
	status, ok := NormalizeRazorpayLikeRefundStatus(response.Status)
	if !ok {
		return RefundResponse{}, NewError(p.Name(), "refund", ErrorCodeUnknownStatus, "payment provider returned an unknown refund status")
	}
	audit, err := sanitizedAuditPayload(map[string]any{"id": response.ID, "status": response.Status, "amount": response.Amount, "currency": response.Currency})
	if err != nil {
		return RefundResponse{}, err
	}
	result := RefundResponse{ProviderRefundID: response.ID, Status: status, RefundedAmount: Money{AmountMinor: response.Amount, Currency: response.Currency}, RawProviderResponse: audit}
	if err := ValidateRefundResponse(result); err != nil {
		return RefundResponse{}, err
	}
	return result.Normalized(), nil
}

func (p *RazorpayLikeProvider) VerifyWebhook(ctx context.Context, req VerifyWebhookRequest) (WebhookEvent, error) {
	if err := ctx.Err(); err != nil {
		return WebhookEvent{}, err
	}
	if err := ValidateVerifyWebhookRequest(req); err != nil {
		return WebhookEvent{}, err
	}
	req = req.Normalized()
	if !validHMACSHA256(p.cfg.WebhookSecret, req.Body, req.Headers["x-razorpay-signature"]) {
		return WebhookEvent{}, invalidWebhookSignature(p.Name())
	}

	var envelope struct {
		Event     string `json:"event"`
		CreatedAt int64  `json:"created_at"`
		Payload   struct {
			Payment struct {
				Entity struct {
					ID        string            `json:"id"`
					OrderID   string            `json:"order_id"`
					Amount    int64             `json:"amount"`
					Currency  string            `json:"currency"`
					Notes     map[string]string `json:"notes"`
					ErrorCode string            `json:"error_code"`
				} `json:"entity"`
			} `json:"payment"`
			Refund struct {
				Entity struct {
					ID        string            `json:"id"`
					PaymentID string            `json:"payment_id"`
					Amount    int64             `json:"amount"`
					Currency  string            `json:"currency"`
					Notes     map[string]string `json:"notes"`
				} `json:"entity"`
			} `json:"refund"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(req.Body, &envelope); err != nil {
		return WebhookEvent{}, invalidWebhookPayload(p.Name(), "payment webhook payload is malformed", err)
	}
	eventType, ok := razorpayWebhookType(envelope.Event)
	if !ok {
		return WebhookEvent{}, invalidWebhookPayload(p.Name(), "payment webhook event type is not supported", ErrInvalidProviderRequest)
	}
	entity := envelope.Payload.Payment.Entity
	refundEntity := envelope.Payload.Refund.Entity
	refundEvent := eventType == WebhookEventRefundSucceeded || eventType == WebhookEventRefundFailed
	if refundEvent {
		entity.ID = refundEntity.PaymentID
		entity.Amount = refundEntity.Amount
		entity.Currency = refundEntity.Currency
		entity.Notes = refundEntity.Notes
	}
	audit, err := sanitizedAuditPayload(map[string]any{
		"event_type":          envelope.Event,
		"refund_id":           entity.Notes["refund_id"],
		"payment_id":          entity.Notes["payment_id"],
		"order_id":            entity.Notes["order_id"],
		"provider_intent_id":  entity.OrderID,
		"provider_payment_id": entity.ID,
		"provider_refund_id":  refundEntity.ID,
		"amount":              entity.Amount,
		"currency":            entity.Currency,
	})
	if err != nil {
		return WebhookEvent{}, err
	}
	event := WebhookEvent{
		Provider:          p.Name(),
		ProviderEventID:   providerWebhookEventID(envelope.Event, coalesce(refundEntity.ID, entity.ID)),
		Type:              eventType,
		RefundID:          entity.Notes["refund_id"],
		PaymentID:         entity.Notes["payment_id"],
		OrderID:           entity.Notes["order_id"],
		ProviderIntentID:  entity.OrderID,
		ProviderPaymentID: entity.ID,
		ProviderRefundID:  refundEntity.ID,
		Amount:            Money{AmountMinor: entity.Amount, Currency: entity.Currency},
		FailureCode:       entity.ErrorCode,
		OccurredAtUnix:    envelope.CreatedAt,
		RawPayload:        audit,
	}
	if err := ValidateWebhookEvent(event); err != nil {
		return WebhookEvent{}, invalidWebhookPayload(p.Name(), "payment webhook event is invalid", err)
	}
	return event.Normalized(), nil
}

func (p *RazorpayLikeProvider) logUnsupported(ctx context.Context, operation string) {
	p.logger.InfoContext(ctx, "payment.provider.operation_not_enabled",
		slog.String("provider", p.Name()),
		slog.String("operation", operation),
	)
}

func razorpayWebhookType(raw string) (WebhookEventType, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "payment.authorized":
		return WebhookEventPaymentAuthorized, true
	case "payment.captured", "order.paid":
		return WebhookEventPaymentCaptured, true
	case "payment.failed":
		return WebhookEventPaymentFailed, true
	case "refund.processed":
		return WebhookEventRefundSucceeded, true
	case "refund.failed":
		return WebhookEventRefundFailed, true
	default:
		return "", false
	}
}

func providerWebhookEventID(eventType string, entityID string) string {
	return strings.TrimSpace(eventType) + ":" + strings.TrimSpace(entityID)
}
