package provider

import (
	"context"
	"encoding/json"
	"strings"
)

const (
	ProviderNameStripeLike   = "stripe_like"
	ProviderNameRazorpayLike = "razorpay_like"
)

type Provider interface {
	Name() string
	CreateIntent(ctx context.Context, req CreateIntentRequest) (CreateIntentResponse, error)
	Capture(ctx context.Context, req CaptureRequest) (CaptureResponse, error)
	Refund(ctx context.Context, req RefundRequest) (RefundResponse, error)
	VerifyWebhook(ctx context.Context, req VerifyWebhookRequest) (WebhookEvent, error)
}

type Money struct {
	AmountMinor int64
	Currency    string
}

func (m Money) Normalized() Money {
	return Money{
		AmountMinor: m.AmountMinor,
		Currency:    strings.ToUpper(strings.TrimSpace(m.Currency)),
	}
}

type Customer struct {
	UserID string
	Email  string
	Phone  string
	Name   string
}

func (c Customer) Normalized() Customer {
	return Customer{
		UserID: strings.TrimSpace(c.UserID),
		Email:  strings.TrimSpace(c.Email),
		Phone:  strings.TrimSpace(c.Phone),
		Name:   strings.TrimSpace(c.Name),
	}
}

type CaptureMode string

const (
	CaptureModeAutomatic CaptureMode = "automatic"
	CaptureModeManual    CaptureMode = "manual"
)

func (m CaptureMode) Normalized() CaptureMode {
	return CaptureMode(strings.ToLower(strings.TrimSpace(string(m))))
}

type CreateIntentRequest struct {
	PaymentID      string
	OrderID        string
	UserID         string
	Amount         Money
	Customer       Customer
	IdempotencyKey string
	CaptureMode    CaptureMode
	Metadata       map[string]string
}

func (r CreateIntentRequest) Normalized() CreateIntentRequest {
	r.PaymentID = strings.TrimSpace(r.PaymentID)
	r.OrderID = strings.TrimSpace(r.OrderID)
	r.UserID = strings.TrimSpace(r.UserID)
	r.Amount = r.Amount.Normalized()
	r.Customer = r.Customer.Normalized()
	r.IdempotencyKey = strings.TrimSpace(r.IdempotencyKey)
	r.CaptureMode = r.CaptureMode.Normalized()
	r.Metadata = NormalizeMetadata(r.Metadata)
	return r
}

type IntentStatus string

const (
	IntentStatusInitiated      IntentStatus = "initiated"
	IntentStatusRequiresAction IntentStatus = "requires_action"
	IntentStatusAuthorized     IntentStatus = "authorized"
	IntentStatusCaptured       IntentStatus = "captured"
	IntentStatusFailed         IntentStatus = "failed"
)

func (s IntentStatus) Normalized() IntentStatus {
	return IntentStatus(strings.ToLower(strings.TrimSpace(string(s))))
}

type CreateIntentResponse struct {
	Provider            string
	ProviderIntentID    string
	ProviderPaymentID   string
	Status              IntentStatus
	ClientSecret        string
	RedirectURL         string
	FrontendPayload     map[string]string
	RawProviderResponse json.RawMessage
}

func (r CreateIntentResponse) Normalized() CreateIntentResponse {
	r.Provider = NormalizeProviderName(r.Provider)
	r.ProviderIntentID = strings.TrimSpace(r.ProviderIntentID)
	r.ProviderPaymentID = strings.TrimSpace(r.ProviderPaymentID)
	r.Status = r.Status.Normalized()
	r.ClientSecret = strings.TrimSpace(r.ClientSecret)
	r.RedirectURL = strings.TrimSpace(r.RedirectURL)
	r.FrontendPayload = NormalizeMetadata(r.FrontendPayload)
	return r
}

type CaptureRequest struct {
	PaymentID         string
	ProviderIntentID  string
	ProviderPaymentID string
	Amount            Money
	IdempotencyKey    string
	Metadata          map[string]string
}

func (r CaptureRequest) Normalized() CaptureRequest {
	r.PaymentID = strings.TrimSpace(r.PaymentID)
	r.ProviderIntentID = strings.TrimSpace(r.ProviderIntentID)
	r.ProviderPaymentID = strings.TrimSpace(r.ProviderPaymentID)
	r.Amount = r.Amount.Normalized()
	r.IdempotencyKey = strings.TrimSpace(r.IdempotencyKey)
	r.Metadata = NormalizeMetadata(r.Metadata)
	return r
}

type CaptureResponse struct {
	ProviderPaymentID   string
	Status              IntentStatus
	CapturedAmount      Money
	RawProviderResponse json.RawMessage
}

func (r CaptureResponse) Normalized() CaptureResponse {
	r.ProviderPaymentID = strings.TrimSpace(r.ProviderPaymentID)
	r.Status = r.Status.Normalized()
	r.CapturedAmount = r.CapturedAmount.Normalized()
	return r
}

type RefundRequest struct {
	PaymentID         string
	RefundID          string
	ProviderPaymentID string
	Amount            Money
	Reason            string
	IdempotencyKey    string
	Metadata          map[string]string
}

func (r RefundRequest) Normalized() RefundRequest {
	r.PaymentID = strings.TrimSpace(r.PaymentID)
	r.RefundID = strings.TrimSpace(r.RefundID)
	r.ProviderPaymentID = strings.TrimSpace(r.ProviderPaymentID)
	r.Amount = r.Amount.Normalized()
	r.Reason = strings.TrimSpace(r.Reason)
	r.IdempotencyKey = strings.TrimSpace(r.IdempotencyKey)
	r.Metadata = NormalizeMetadata(r.Metadata)
	return r
}

type RefundStatus string

const (
	RefundStatusProcessing RefundStatus = "processing"
	RefundStatusSucceeded  RefundStatus = "succeeded"
	RefundStatusFailed     RefundStatus = "failed"
)

func (s RefundStatus) Normalized() RefundStatus {
	return RefundStatus(strings.ToLower(strings.TrimSpace(string(s))))
}

type RefundResponse struct {
	ProviderRefundID    string
	Status              RefundStatus
	RefundedAmount      Money
	RawProviderResponse json.RawMessage
}

func (r RefundResponse) Normalized() RefundResponse {
	r.ProviderRefundID = strings.TrimSpace(r.ProviderRefundID)
	r.Status = r.Status.Normalized()
	r.RefundedAmount = r.RefundedAmount.Normalized()
	return r
}

type VerifyWebhookRequest struct {
	Headers map[string]string
	Body    []byte
}

func (r VerifyWebhookRequest) Normalized() VerifyWebhookRequest {
	headers := make(map[string]string, len(r.Headers))
	for key, value := range r.Headers {
		headers[strings.ToLower(strings.TrimSpace(key))] = strings.TrimSpace(value)
	}
	r.Headers = headers
	return r
}

type WebhookEventType string

const (
	WebhookEventPaymentRequiresAction WebhookEventType = "payment.requires_action"
	WebhookEventPaymentAuthorized     WebhookEventType = "payment.authorized"
	WebhookEventPaymentCaptured       WebhookEventType = "payment.captured"
	WebhookEventPaymentFailed         WebhookEventType = "payment.failed"
	WebhookEventRefundSucceeded       WebhookEventType = "refund.succeeded"
	WebhookEventRefundFailed          WebhookEventType = "refund.failed"
)

func (t WebhookEventType) Normalized() WebhookEventType {
	return WebhookEventType(strings.ToLower(strings.TrimSpace(string(t))))
}

type WebhookEvent struct {
	Provider          string
	ProviderEventID   string
	Type              WebhookEventType
	RefundID          string
	PaymentID         string
	OrderID           string
	ProviderIntentID  string
	ProviderPaymentID string
	ProviderRefundID  string
	Amount            Money
	FailureCode       string
	OccurredAtUnix    int64
	RawPayload        json.RawMessage
}

func (e WebhookEvent) Normalized() WebhookEvent {
	e.Provider = NormalizeProviderName(e.Provider)
	e.ProviderEventID = strings.TrimSpace(e.ProviderEventID)
	e.Type = e.Type.Normalized()
	e.RefundID = strings.TrimSpace(e.RefundID)
	e.PaymentID = strings.TrimSpace(e.PaymentID)
	e.OrderID = strings.TrimSpace(e.OrderID)
	e.ProviderIntentID = strings.TrimSpace(e.ProviderIntentID)
	e.ProviderPaymentID = strings.TrimSpace(e.ProviderPaymentID)
	e.ProviderRefundID = strings.TrimSpace(e.ProviderRefundID)
	e.Amount = e.Amount.Normalized()
	e.FailureCode = strings.TrimSpace(e.FailureCode)
	return e
}

func NormalizeProviderName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
