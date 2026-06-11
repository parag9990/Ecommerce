package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

const (
	maxIDLength          = 128
	maxMetadataKeyLength = 64
	maxMetadataValLength = 512
	maxReasonLength      = 512
)

var sensitiveKeys = []string{
	"api_key",
	"api_secret",
	"authorization",
	"card_number",
	"client_secret",
	"cvv",
	"cvc",
	"otp",
	"password",
	"secret_key",
	"token",
	"webhook_secret",
}

func ValidateCreateIntentRequest(req CreateIntentRequest) error {
	req = req.Normalized()
	if err := validateID("payment_id", req.PaymentID); err != nil {
		return err
	}
	if err := validateID("order_id", req.OrderID); err != nil {
		return err
	}
	if err := validateID("user_id", req.UserID); err != nil {
		return err
	}
	if err := ValidateMoney(req.Amount); err != nil {
		return err
	}
	if err := ValidateCaptureMode(req.CaptureMode); err != nil {
		return err
	}
	if err := ValidateIdempotencyKey(req.IdempotencyKey); err != nil {
		return err
	}
	if err := ValidateMetadata(req.Metadata); err != nil {
		return err
	}
	return validateCustomer(req.Customer)
}

func ValidateCaptureRequest(req CaptureRequest) error {
	req = req.Normalized()
	if err := validateID("payment_id", req.PaymentID); err != nil {
		return err
	}
	if req.ProviderIntentID == "" && req.ProviderPaymentID == "" {
		return fmt.Errorf("%w: provider_intent_id or provider_payment_id is required", ErrInvalidProviderRequest)
	}
	if err := validateOptionalID("provider_intent_id", req.ProviderIntentID); err != nil {
		return err
	}
	if err := validateOptionalID("provider_payment_id", req.ProviderPaymentID); err != nil {
		return err
	}
	if err := ValidateMoney(req.Amount); err != nil {
		return err
	}
	if err := ValidateIdempotencyKey(req.IdempotencyKey); err != nil {
		return err
	}
	return ValidateMetadata(req.Metadata)
}

func ValidateRefundRequest(req RefundRequest) error {
	req = req.Normalized()
	if err := validateID("payment_id", req.PaymentID); err != nil {
		return err
	}
	if err := validateID("refund_id", req.RefundID); err != nil {
		return err
	}
	if err := validateID("provider_payment_id", req.ProviderPaymentID); err != nil {
		return err
	}
	if err := ValidateMoney(req.Amount); err != nil {
		return err
	}
	if req.Reason == "" {
		return fmt.Errorf("%w: reason is required", ErrInvalidProviderRequest)
	}
	if len(req.Reason) > maxReasonLength {
		return fmt.Errorf("%w: reason exceeds %d characters", ErrInvalidProviderRequest, maxReasonLength)
	}
	if err := ValidateIdempotencyKey(req.IdempotencyKey); err != nil {
		return err
	}
	return ValidateMetadata(req.Metadata)
}

func ValidateVerifyWebhookRequest(req VerifyWebhookRequest) error {
	req = req.Normalized()
	if len(req.Body) == 0 {
		return fmt.Errorf("%w: webhook body is required", ErrInvalidProviderRequest)
	}
	if len(req.Headers) == 0 {
		return fmt.Errorf("%w: webhook headers are required", ErrInvalidProviderRequest)
	}
	return nil
}

func ValidateCreateIntentResponse(res CreateIntentResponse) error {
	res = res.Normalized()
	if err := validateID("provider", res.Provider); err != nil {
		return err
	}
	if err := validateID("provider_intent_id", res.ProviderIntentID); err != nil {
		return err
	}
	if err := validateOptionalID("provider_payment_id", res.ProviderPaymentID); err != nil {
		return err
	}
	if !IsKnownIntentStatus(res.Status) {
		return fmt.Errorf("%w: invalid intent status %s", ErrInvalidProviderRequest, res.Status)
	}
	if res.RedirectURL != "" {
		if _, err := url.ParseRequestURI(res.RedirectURL); err != nil {
			return fmt.Errorf("%w: redirect_url must be valid", ErrInvalidProviderRequest)
		}
	}
	if err := ValidateMetadata(res.FrontendPayload); err != nil {
		return err
	}
	return ValidateSanitizedJSONPayload(res.RawProviderResponse, false, "raw_provider_response")
}

func ValidateCaptureResponse(res CaptureResponse) error {
	res = res.Normalized()
	if err := validateID("provider_payment_id", res.ProviderPaymentID); err != nil {
		return err
	}
	if !IsKnownIntentStatus(res.Status) {
		return fmt.Errorf("%w: invalid intent status %s", ErrInvalidProviderRequest, res.Status)
	}
	if err := ValidateMoney(res.CapturedAmount); err != nil {
		return err
	}
	return ValidateSanitizedJSONPayload(res.RawProviderResponse, false, "raw_provider_response")
}

func ValidateRefundResponse(res RefundResponse) error {
	res = res.Normalized()
	if err := validateID("provider_refund_id", res.ProviderRefundID); err != nil {
		return err
	}
	if !IsKnownRefundStatus(res.Status) {
		return fmt.Errorf("%w: invalid refund status %s", ErrInvalidProviderRequest, res.Status)
	}
	if err := ValidateMoney(res.RefundedAmount); err != nil {
		return err
	}
	return ValidateSanitizedJSONPayload(res.RawProviderResponse, false, "raw_provider_response")
}

func ValidateWebhookEvent(event WebhookEvent) error {
	event = event.Normalized()
	if err := validateID("provider", event.Provider); err != nil {
		return err
	}
	if err := validateID("provider_event_id", event.ProviderEventID); err != nil {
		return err
	}
	if !IsKnownWebhookEventType(event.Type) {
		return fmt.Errorf("%w: invalid webhook event type %s", ErrInvalidProviderRequest, event.Type)
	}
	if event.RefundID == "" && event.PaymentID == "" && event.OrderID == "" && event.ProviderIntentID == "" && event.ProviderPaymentID == "" && event.ProviderRefundID == "" {
		return fmt.Errorf("%w: payment lookup reference is required", ErrInvalidProviderRequest)
	}
	if err := validateOptionalID("refund_id", event.RefundID); err != nil {
		return err
	}
	if err := validateOptionalID("payment_id", event.PaymentID); err != nil {
		return err
	}
	if err := validateOptionalID("order_id", event.OrderID); err != nil {
		return err
	}
	if err := validateOptionalID("provider_intent_id", event.ProviderIntentID); err != nil {
		return err
	}
	if err := validateOptionalID("provider_payment_id", event.ProviderPaymentID); err != nil {
		return err
	}
	if err := validateOptionalID("provider_refund_id", event.ProviderRefundID); err != nil {
		return err
	}
	if err := validateOptionalID("failure_code", event.FailureCode); err != nil {
		return err
	}
	if (event.Type == WebhookEventPaymentCaptured || event.Type == WebhookEventRefundSucceeded || event.Type == WebhookEventRefundFailed) && event.Amount.AmountMinor <= 0 {
		return fmt.Errorf("%w: outcome webhook amount is required", ErrInvalidProviderRequest)
	}
	if event.Amount.AmountMinor > 0 {
		if err := ValidateMoney(event.Amount); err != nil {
			return err
		}
	}
	return ValidateSanitizedJSONPayload(event.RawPayload, true, "raw_payload")
}

func ValidateMoney(m Money) error {
	m = m.Normalized()
	if m.AmountMinor <= 0 {
		return fmt.Errorf("%w: amount must be greater than zero", ErrInvalidProviderRequest)
	}
	if !isCurrencyCode(m.Currency) {
		return fmt.Errorf("%w: currency must be a 3-letter ISO code", ErrInvalidProviderRequest)
	}
	return nil
}

func ValidateCaptureMode(mode CaptureMode) error {
	switch mode.Normalized() {
	case CaptureModeAutomatic, CaptureModeManual:
		return nil
	default:
		return fmt.Errorf("%w: capture_mode must be automatic or manual", ErrInvalidProviderRequest)
	}
}

func ValidateIdempotencyKey(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("%w: idempotency_key is required", ErrInvalidProviderRequest)
	}
	if len(key) > maxIDLength {
		return fmt.Errorf("%w: idempotency_key exceeds %d characters", ErrInvalidProviderRequest, maxIDLength)
	}
	if containsSensitiveKey(key) {
		return fmt.Errorf("%w: idempotency_key contains sensitive token marker", ErrInvalidProviderRequest)
	}
	return nil
}

func NormalizeMetadata(metadata map[string]string) map[string]string {
	if len(metadata) == 0 {
		return nil
	}
	normalized := make(map[string]string, len(metadata))
	for key, value := range metadata {
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		if key != "" {
			normalized[key] = value
		}
	}
	return normalized
}

func ValidateMetadata(metadata map[string]string) error {
	for key, value := range NormalizeMetadata(metadata) {
		if key == "" {
			return fmt.Errorf("%w: metadata key is required", ErrInvalidProviderRequest)
		}
		if len(key) > maxMetadataKeyLength {
			return fmt.Errorf("%w: metadata key %q exceeds %d characters", ErrInvalidProviderRequest, key, maxMetadataKeyLength)
		}
		if len(value) > maxMetadataValLength {
			return fmt.Errorf("%w: metadata value for %q exceeds %d characters", ErrInvalidProviderRequest, key, maxMetadataValLength)
		}
		if containsSensitiveKey(key) {
			return fmt.Errorf("%w: metadata contains sensitive key %q", ErrInvalidProviderRequest, key)
		}
	}
	return nil
}

func BuildSafeMetadata(paymentID string, orderID string, userID string, requestID string, idempotencyKey string) map[string]string {
	return NormalizeMetadata(map[string]string{
		"payment_id":      paymentID,
		"order_id":        orderID,
		"user_id":         userID,
		"request_id":      requestID,
		"idempotency_key": idempotencyKey,
		"service_name":    "payment-service",
	})
}

func ValidateSanitizedJSONPayload(raw []byte, required bool, field string) error {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		if required {
			return fmt.Errorf("%w: %s is required", ErrInvalidProviderRequest, field)
		}
		return nil
	}
	if !json.Valid(raw) {
		return fmt.Errorf("%w: %s must be valid JSON", ErrInvalidProviderRequest, field)
	}
	lower := strings.ToLower(string(raw))
	for _, key := range sensitiveKeys {
		if strings.Contains(lower, `"`+key+`"`) {
			return fmt.Errorf("%w: %s contains sensitive key %q", ErrInvalidProviderRequest, field, key)
		}
	}
	return nil
}

func IsKnownIntentStatus(status IntentStatus) bool {
	switch status.Normalized() {
	case IntentStatusInitiated, IntentStatusRequiresAction, IntentStatusAuthorized, IntentStatusCaptured, IntentStatusFailed:
		return true
	default:
		return false
	}
}

func IsKnownRefundStatus(status RefundStatus) bool {
	switch status.Normalized() {
	case RefundStatusProcessing, RefundStatusSucceeded, RefundStatusFailed:
		return true
	default:
		return false
	}
}

func IsKnownWebhookEventType(eventType WebhookEventType) bool {
	switch eventType.Normalized() {
	case WebhookEventPaymentRequiresAction, WebhookEventPaymentAuthorized, WebhookEventPaymentCaptured, WebhookEventPaymentFailed, WebhookEventRefundSucceeded, WebhookEventRefundFailed:
		return true
	default:
		return false
	}
}

func validateCustomer(customer Customer) error {
	customer = customer.Normalized()
	if customer.UserID != "" {
		if err := validateID("customer.user_id", customer.UserID); err != nil {
			return err
		}
	}
	if len(customer.Email) > 320 {
		return fmt.Errorf("%w: customer email exceeds 320 characters", ErrInvalidProviderRequest)
	}
	if len(customer.Phone) > 32 {
		return fmt.Errorf("%w: customer phone exceeds 32 characters", ErrInvalidProviderRequest)
	}
	if len(customer.Name) > 128 {
		return fmt.Errorf("%w: customer name exceeds 128 characters", ErrInvalidProviderRequest)
	}
	return nil
}

func validateID(field string, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%w: %s is required", ErrInvalidProviderRequest, field)
	}
	if len(value) > maxIDLength {
		return fmt.Errorf("%w: %s exceeds %d characters", ErrInvalidProviderRequest, field, maxIDLength)
	}
	if containsSensitiveKey(value) {
		return fmt.Errorf("%w: %s contains sensitive token marker", ErrInvalidProviderRequest, field)
	}
	return nil
}

func validateOptionalID(field string, value string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return validateID(field, value)
}

func isCurrencyCode(currency string) bool {
	if len(currency) != 3 {
		return false
	}
	for _, r := range currency {
		if !unicode.IsUpper(r) || r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

func containsSensitiveKey(value string) bool {
	value = strings.ToLower(value)
	for _, key := range sensitiveKeys {
		if strings.Contains(value, key) {
			return true
		}
	}
	return false
}
