package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type WebhookProcessingStatus string

const (
	WebhookProcessingStatusProcessed WebhookProcessingStatus = "processed"
	WebhookProcessingStatusDuplicate WebhookProcessingStatus = "duplicate"
	WebhookProcessingStatusIgnored   WebhookProcessingStatus = "ignored"
)

type PaymentWebhookEvent struct {
	WebhookEventID  string
	Provider        string
	ProviderEventID string
	EventType       string
	Processed       bool
	Payload         json.RawMessage
	ReceivedAt      time.Time
	ProcessedAt     *time.Time
}

func (e PaymentWebhookEvent) Normalized() PaymentWebhookEvent {
	e.WebhookEventID = strings.TrimSpace(e.WebhookEventID)
	e.Provider = strings.ToLower(strings.TrimSpace(e.Provider))
	e.ProviderEventID = strings.TrimSpace(e.ProviderEventID)
	e.EventType = strings.TrimSpace(e.EventType)
	if e.ProcessedAt != nil {
		processedAt := e.ProcessedAt.UTC()
		e.ProcessedAt = &processedAt
	}
	e.ReceivedAt = e.ReceivedAt.UTC()
	return e
}

func (e PaymentWebhookEvent) Validate() error {
	e = e.Normalized()
	if e.WebhookEventID == "" {
		return fmt.Errorf("%w: webhook_event_id is required", ErrInvalidPayment)
	}
	if len(e.WebhookEventID) > 64 {
		return fmt.Errorf("%w: webhook_event_id exceeds 64 characters", ErrInvalidPayment)
	}
	if e.Provider == "" {
		return fmt.Errorf("%w: provider is required", ErrInvalidPayment)
	}
	if len(e.Provider) > 64 {
		return fmt.Errorf("%w: provider exceeds 64 characters", ErrInvalidPayment)
	}
	if e.ProviderEventID == "" {
		return fmt.Errorf("%w: provider_event_id is required", ErrInvalidPayment)
	}
	if len(e.ProviderEventID) > 128 {
		return fmt.Errorf("%w: provider_event_id exceeds 128 characters", ErrInvalidPayment)
	}
	if e.EventType == "" {
		return fmt.Errorf("%w: event_type is required", ErrInvalidPayment)
	}
	if len(e.EventType) > 128 {
		return fmt.Errorf("%w: event_type exceeds 128 characters", ErrInvalidPayment)
	}
	if e.Processed && e.ProcessedAt == nil {
		return fmt.Errorf("%w: processed_at is required when webhook event is processed", ErrInvalidPayment)
	}
	if err := validateSanitizedJSONPayload(e.Payload, true, "payload"); err != nil {
		return err
	}
	return nil
}

// VerifiedPaymentWebhook is the provider-neutral notification used by the
// transaction layer after the adapter has authenticated and sanitized input.
type VerifiedPaymentWebhook struct {
	WebhookEventID    string
	Provider          string
	ProviderEventID   string
	EventType         string
	Event             PaymentEvent
	PaymentID         string
	OrderID           string
	ProviderIntentID  string
	ProviderPaymentID string
	Amount            Money
	FailureCode       string
	Payload           json.RawMessage
	OccurredAt        time.Time
	ReceivedAt        time.Time
}

func (e VerifiedPaymentWebhook) Normalized() VerifiedPaymentWebhook {
	e.WebhookEventID = strings.TrimSpace(e.WebhookEventID)
	e.Provider = strings.ToLower(strings.TrimSpace(e.Provider))
	e.ProviderEventID = strings.TrimSpace(e.ProviderEventID)
	e.EventType = strings.ToLower(strings.TrimSpace(e.EventType))
	e.PaymentID = strings.TrimSpace(e.PaymentID)
	e.OrderID = strings.TrimSpace(e.OrderID)
	e.ProviderIntentID = strings.TrimSpace(e.ProviderIntentID)
	e.ProviderPaymentID = strings.TrimSpace(e.ProviderPaymentID)
	e.Amount = e.Amount.Normalized()
	e.FailureCode = strings.TrimSpace(e.FailureCode)
	e.OccurredAt = e.OccurredAt.UTC()
	e.ReceivedAt = e.ReceivedAt.UTC()
	return e
}

func (e VerifiedPaymentWebhook) Validate() error {
	e = e.Normalized()
	stored := PaymentWebhookEvent{
		WebhookEventID:  e.WebhookEventID,
		Provider:        e.Provider,
		ProviderEventID: e.ProviderEventID,
		EventType:       e.EventType,
		Payload:         e.Payload,
		ReceivedAt:      e.ReceivedAt,
	}
	if err := stored.Validate(); err != nil {
		return err
	}
	switch e.Event {
	case PaymentEventProviderRequiresAction, PaymentEventProviderAuthorized, PaymentEventProviderCaptured, PaymentEventProviderFailed:
	default:
		return fmt.Errorf("%w: event %s is outside payment webhook scope", ErrInvalidWebhookEvent, e.Event)
	}
	if e.PaymentID == "" && e.ProviderIntentID == "" && e.ProviderPaymentID == "" && e.OrderID == "" {
		return fmt.Errorf("%w: payment lookup reference is required", ErrInvalidWebhookEvent)
	}
	if len(e.PaymentID) > 64 || len(e.OrderID) > 64 {
		return fmt.Errorf("%w: local payment reference exceeds schema limits", ErrInvalidWebhookEvent)
	}
	if len(e.ProviderIntentID) > 128 || len(e.ProviderPaymentID) > 128 {
		return fmt.Errorf("%w: provider payment reference exceeds schema limits", ErrInvalidWebhookEvent)
	}
	if len(e.FailureCode) > 128 {
		return fmt.Errorf("%w: failure_code exceeds schema limits", ErrInvalidWebhookEvent)
	}
	if e.Event == PaymentEventProviderCaptured && e.Amount.Amount <= 0 {
		return fmt.Errorf("%w: captured webhook amount is required", ErrInvalidWebhookEvent)
	}
	if e.Amount.Amount > 0 {
		if err := e.Amount.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (e VerifiedPaymentWebhook) StoredEvent() PaymentWebhookEvent {
	e = e.Normalized()
	return PaymentWebhookEvent{
		WebhookEventID:  e.WebhookEventID,
		Provider:        e.Provider,
		ProviderEventID: e.ProviderEventID,
		EventType:       e.EventType,
		Payload:         append(json.RawMessage(nil), e.Payload...),
		ReceivedAt:      e.ReceivedAt,
	}
}

type WebhookProcessingResult struct {
	WebhookEventID   string
	ProviderEventID  string
	ProcessingStatus WebhookProcessingStatus
	Payment          Payment
	PreviousStatus   PaymentStatus
	StatusChanged    bool
	IgnoreReason     string
	LateCapture      bool
	DuplicateCapture bool
}

type PaymentDomainEvent struct {
	EventID           string    `json:"event_id"`
	EventType         string    `json:"event_type"`
	RefundID          string    `json:"refund_id,omitempty"`
	RefundStatus      string    `json:"refund_status,omitempty"`
	PaymentID         string    `json:"payment_id"`
	OrderID           string    `json:"order_id"`
	Provider          string    `json:"provider"`
	ProviderPaymentID string    `json:"provider_payment_id,omitempty"`
	Amount            int64     `json:"amount"`
	Currency          string    `json:"currency"`
	Reason            string    `json:"reason,omitempty"`
	OccurredAt        time.Time `json:"occurred_at"`
}

func (r WebhookProcessingResult) DomainEvent(occurredAt time.Time) (PaymentDomainEvent, bool) {
	if !r.StatusChanged {
		return PaymentDomainEvent{}, false
	}
	eventType, ok := domainEventTypeForStatus(r.Payment.Status)
	if !ok {
		return PaymentDomainEvent{}, false
	}
	if r.DuplicateCapture {
		eventType = "DuplicateCaptureDetected"
	}
	if occurredAt.IsZero() {
		occurredAt = r.Payment.UpdatedAt
	}
	amount := r.Payment.Amount.Amount
	if r.Payment.Status == PaymentStatusCaptured && r.Payment.CapturedAmount > 0 {
		amount = r.Payment.CapturedAmount
	}
	return PaymentDomainEvent{
		EventID:           r.WebhookEventID,
		EventType:         eventType,
		PaymentID:         r.Payment.PaymentID,
		OrderID:           r.Payment.OrderID,
		Provider:          r.Payment.Provider,
		ProviderPaymentID: r.Payment.ProviderPaymentID,
		Amount:            amount,
		Currency:          r.Payment.Amount.Currency,
		OccurredAt:        occurredAt.UTC(),
	}, true
}

func domainEventTypeForStatus(status PaymentStatus) (string, bool) {
	switch status {
	case PaymentStatusRequiresAction:
		return "PaymentRequiresAction", true
	case PaymentStatusAuthorized:
		return "PaymentAuthorized", true
	case PaymentStatusCaptured:
		return "PaymentCaptured", true
	case PaymentStatusFailed:
		return "PaymentFailed", true
	default:
		return "", false
	}
}

func ApplyVerifiedPaymentWebhook(payment Payment, event VerifiedPaymentWebhook, at time.Time) (Payment, TransitionDecision, string) {
	payment = payment.Normalized()
	event = event.Normalized()
	switch {
	case payment.Provider != event.Provider:
		return payment, TransitionDecision{}, "provider mismatch"
	case event.PaymentID != "" && event.PaymentID != payment.PaymentID:
		return payment, TransitionDecision{}, "payment id mismatch"
	case event.OrderID != "" && event.OrderID != payment.OrderID:
		return payment, TransitionDecision{}, "order id mismatch"
	case event.ProviderIntentID != "" && payment.ProviderIntentID != "" && event.ProviderIntentID != payment.ProviderIntentID:
		return payment, TransitionDecision{}, "provider intent id mismatch"
	case event.Amount.Currency != "" && event.Amount.Currency != payment.Amount.Currency:
		return payment, TransitionDecision{}, "currency mismatch"
	case event.Amount.Amount > payment.Amount.Amount:
		return payment, TransitionDecision{}, "amount exceeds authorized payment amount"
	}

	previous := payment.Status
	var decision TransitionDecision
	if previous == PaymentStatusFailed && event.Event == PaymentEventProviderCaptured {
		// A verified provider capture is financial truth even when a prior
		// callback reported failure; the caller records this as a late capture.
		payment.Status = PaymentStatusCaptured
		payment.UpdatedAt = normalizeTransitionTime(at)
		decision = TransitionDecision{
			From:    previous,
			To:      PaymentStatusCaptured,
			Event:   event.Event,
			Allowed: true,
			Reason:  "verified late provider capture supersedes failed outcome",
		}
	} else {
		var err error
		decision, err = payment.ApplyEvent(event.Event, at)
		if err != nil {
			return payment, decision, "payment status transition is not allowed"
		}
	}
	if decision.Noop {
		return payment, decision, "payment already has webhook outcome status"
	}
	payment.ProviderPaymentID = coalesceWebhookValue(event.ProviderPaymentID, payment.ProviderPaymentID)
	if payment.ProviderIntentID == "" {
		payment.ProviderIntentID = event.ProviderIntentID
	}
	switch payment.Status {
	case PaymentStatusCaptured:
		payment.CapturedAmount = event.Amount.Amount
		if payment.CapturedAmount == 0 {
			payment.CapturedAmount = payment.Amount.Amount
		}
		payment.FailureCode = ""
		payment.FailureMessage = ""
	case PaymentStatusFailed:
		payment.FailureCode = event.FailureCode
		if payment.FailureCode == "" {
			payment.FailureCode = "provider_failed"
		}
		payment.FailureMessage = "Payment provider reported payment failure"
	}
	decision.From = previous
	return payment, decision, ""
}

func coalesceWebhookValue(value string, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}
