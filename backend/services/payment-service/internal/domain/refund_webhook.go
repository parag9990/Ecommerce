package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// VerifiedRefundWebhook is a provider-authenticated terminal refund outcome.
type VerifiedRefundWebhook struct {
	WebhookEventID    string
	Provider          string
	ProviderEventID   string
	EventType         string
	Outcome           RefundStatus
	RefundID          string
	PaymentID         string
	ProviderPaymentID string
	ProviderRefundID  string
	Amount            Money
	FailureCode       string
	Payload           json.RawMessage
	OccurredAt        time.Time
	ReceivedAt        time.Time
}

func (e VerifiedRefundWebhook) Normalized() VerifiedRefundWebhook {
	e.WebhookEventID = strings.TrimSpace(e.WebhookEventID)
	e.Provider = strings.ToLower(strings.TrimSpace(e.Provider))
	e.ProviderEventID = strings.TrimSpace(e.ProviderEventID)
	e.EventType = strings.ToLower(strings.TrimSpace(e.EventType))
	e.RefundID = strings.TrimSpace(e.RefundID)
	e.PaymentID = strings.TrimSpace(e.PaymentID)
	e.ProviderPaymentID = strings.TrimSpace(e.ProviderPaymentID)
	e.ProviderRefundID = strings.TrimSpace(e.ProviderRefundID)
	e.Amount = e.Amount.Normalized()
	e.FailureCode = strings.TrimSpace(e.FailureCode)
	e.OccurredAt = e.OccurredAt.UTC()
	e.ReceivedAt = e.ReceivedAt.UTC()
	return e
}

func (e VerifiedRefundWebhook) Validate() error {
	e = e.Normalized()
	if err := (PaymentWebhookEvent{
		WebhookEventID:  e.WebhookEventID,
		Provider:        e.Provider,
		ProviderEventID: e.ProviderEventID,
		EventType:       e.EventType,
		Payload:         e.Payload,
		ReceivedAt:      e.ReceivedAt,
	}).Validate(); err != nil {
		return err
	}
	if e.Outcome != RefundStatusSucceeded && e.Outcome != RefundStatusFailed {
		return fmt.Errorf("%w: terminal refund outcome is required", ErrInvalidWebhookEvent)
	}
	if e.RefundID == "" && e.ProviderRefundID == "" {
		return fmt.Errorf("%w: refund lookup reference is required", ErrInvalidWebhookEvent)
	}
	if len(e.RefundID) > 64 || len(e.PaymentID) > 64 {
		return fmt.Errorf("%w: local refund reference exceeds schema limits", ErrInvalidWebhookEvent)
	}
	if len(e.ProviderRefundID) > 128 || len(e.ProviderPaymentID) > 128 {
		return fmt.Errorf("%w: provider refund reference exceeds schema limits", ErrInvalidWebhookEvent)
	}
	if err := e.Amount.Validate(); err != nil {
		return err
	}
	return nil
}

func (e VerifiedRefundWebhook) StoredEvent() PaymentWebhookEvent {
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

type RefundWebhookProcessingResult struct {
	WebhookEventID   string
	ProviderEventID  string
	ProcessingStatus WebhookProcessingStatus
	Refund           Refund
	Payment          Payment
	PreviousStatus   RefundStatus
	StatusChanged    bool
	IgnoreReason     string
}

func (r RefundWebhookProcessingResult) DomainEvent(occurredAt time.Time) (PaymentDomainEvent, bool) {
	if !r.StatusChanged {
		return PaymentDomainEvent{}, false
	}
	eventType := "RefundFailed"
	if r.Refund.Status == RefundStatusSucceeded {
		eventType = "RefundSucceeded"
	}
	if occurredAt.IsZero() {
		occurredAt = r.Refund.UpdatedAt
	}
	return RefundDomainEvent(r.WebhookEventID, eventType, r.Refund, r.Payment, occurredAt), true
}

func RefundDomainEvent(eventID string, eventType string, refund Refund, payment Payment, occurredAt time.Time) PaymentDomainEvent {
	return PaymentDomainEvent{
		EventID:           strings.TrimSpace(eventID),
		EventType:         strings.TrimSpace(eventType),
		RefundID:          refund.RefundID,
		RefundStatus:      string(refund.Status),
		PaymentID:         payment.PaymentID,
		OrderID:           payment.OrderID,
		Provider:          payment.Provider,
		ProviderPaymentID: payment.ProviderPaymentID,
		Amount:            refund.Amount.Amount,
		Currency:          refund.Amount.Currency,
		Reason:            refund.ReviewReason,
		OccurredAt:        occurredAt.UTC(),
	}
}

func ApplyVerifiedRefundOutcome(payment Payment, refund Refund, event VerifiedRefundWebhook, at time.Time) (Payment, Refund, string) {
	payment = payment.Normalized()
	refund = refund.Normalized()
	event = event.Normalized()
	switch {
	case payment.Provider != event.Provider:
		return payment, refund, "provider mismatch"
	case refund.PaymentID != payment.PaymentID:
		return payment, refund, "refund payment mismatch"
	case event.PaymentID != "" && event.PaymentID != payment.PaymentID:
		return payment, refund, "payment id mismatch"
	case event.RefundID != "" && event.RefundID != refund.RefundID:
		return payment, refund, "refund id mismatch"
	case event.ProviderRefundID != "" && refund.ProviderRefundID != "" && event.ProviderRefundID != refund.ProviderRefundID:
		return payment, refund, "provider refund id mismatch"
	case event.ProviderPaymentID != "" && payment.ProviderPaymentID != "" && event.ProviderPaymentID != payment.ProviderPaymentID:
		return payment, refund, "provider payment id mismatch"
	case event.Amount != refund.Amount:
		return payment, refund, "refund amount mismatch"
	case !refund.CanTransitionTo(event.Outcome):
		return payment, refund, "refund status transition is not allowed"
	}
	if refund.Status == event.Outcome {
		return payment, refund, "refund already has terminal outcome status"
	}

	refund.Status = event.Outcome
	if refund.ProviderRefundID == "" {
		refund.ProviderRefundID = event.ProviderRefundID
	}
	refund.UpdatedAt = normalizeTransitionTime(at)
	if event.Outcome == RefundStatusFailed {
		return payment, refund, ""
	}

	newRefunded := payment.RefundedAmount + refund.Amount.Amount
	if newRefunded > payment.CapturedAmount {
		return payment, refund, "refunded amount exceeds captured amount"
	}
	payment.RefundedAmount = newRefunded
	target := PaymentStatusPartiallyRefunded
	if newRefunded == payment.CapturedAmount {
		target = PaymentStatusRefunded
	}
	if _, err := payment.ApplyTransition(target, at); err != nil {
		return payment, refund, "payment refund transition is not allowed"
	}
	payment.UpdatedAt = normalizeTransitionTime(at)
	return payment, refund, ""
}
