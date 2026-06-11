package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestApplyVerifiedPaymentWebhookCapturesMatchingPayment(t *testing.T) {
	payment := webhookTestPayment(PaymentStatusAuthorized)
	event := VerifiedPaymentWebhook{
		Provider:          "stripe_like",
		ProviderEventID:   "evt_123",
		PaymentID:         payment.PaymentID,
		ProviderIntentID:  payment.ProviderIntentID,
		ProviderPaymentID: "ch_123",
		Event:             PaymentEventProviderCaptured,
		Amount:            Money{Amount: 1000, Currency: "INR"},
	}

	updated, decision, reason := ApplyVerifiedPaymentWebhook(payment, event, time.Now())
	if reason != "" || !decision.Allowed || decision.Noop {
		t.Fatalf("decision = %+v reason = %q, want changed capture", decision, reason)
	}
	if updated.Status != PaymentStatusCaptured || updated.CapturedAmount != 1000 || updated.ProviderPaymentID != "ch_123" {
		t.Fatalf("updated payment = %+v, want captured amount/provider id", updated)
	}
}

func TestApplyVerifiedPaymentWebhookIgnoresDowngradeAndMismatch(t *testing.T) {
	payment := webhookTestPayment(PaymentStatusCaptured)
	failed := VerifiedPaymentWebhook{
		Provider:        "stripe_like",
		ProviderEventID: "evt_failed",
		PaymentID:       payment.PaymentID,
		Event:           PaymentEventProviderFailed,
		Amount:          Money{Amount: 1000, Currency: "INR"},
	}
	updated, _, reason := ApplyVerifiedPaymentWebhook(payment, failed, time.Now())
	if updated.Status != PaymentStatusCaptured || reason == "" {
		t.Fatalf("failed after capture result = %+v/%q, want ignored capture", updated, reason)
	}

	failed.Amount.Currency = "USD"
	_, _, reason = ApplyVerifiedPaymentWebhook(webhookTestPayment(PaymentStatusAuthorized), failed, time.Now())
	if reason != "currency mismatch" {
		t.Fatalf("mismatch reason = %q, want currency mismatch", reason)
	}
}

func TestApplyVerifiedPaymentWebhookRecordsVerifiedLateCapture(t *testing.T) {
	payment := webhookTestPayment(PaymentStatusFailed)
	event := VerifiedPaymentWebhook{
		Provider:          "stripe_like",
		ProviderEventID:   "evt_late_capture",
		PaymentID:         payment.PaymentID,
		ProviderIntentID:  payment.ProviderIntentID,
		ProviderPaymentID: "ch_late",
		Event:             PaymentEventProviderCaptured,
		Amount:            Money{Amount: 1000, Currency: "INR"},
	}
	updated, decision, reason := ApplyVerifiedPaymentWebhook(payment, event, time.Now())
	if reason != "" || !decision.Allowed || updated.Status != PaymentStatusCaptured || updated.CapturedAmount != 1000 {
		t.Fatalf("late capture = %+v/%+v/%q, want recorded provider truth", updated, decision, reason)
	}
}

func TestWebhookDuplicateCapturePublishesAlertEventType(t *testing.T) {
	result := WebhookProcessingResult{
		WebhookEventID:   "whe_123",
		StatusChanged:    true,
		DuplicateCapture: true,
		Payment: Payment{
			PaymentID: "pay_retry", OrderID: "ord_123", Provider: "stripe_like",
			Status: PaymentStatusCaptured, Amount: Money{Amount: 1000, Currency: "INR"},
		},
	}
	event, ok := result.DomainEvent(time.Now())
	if !ok || event.EventType != "DuplicateCaptureDetected" {
		t.Fatalf("DomainEvent() = %+v/%v, want duplicate capture alert", event, ok)
	}
}

func TestVerifiedPaymentWebhookRejectsRefundEventFromWebhookTaskScope(t *testing.T) {
	event := VerifiedPaymentWebhook{
		WebhookEventID:   "whe_123",
		Provider:         "stripe_like",
		ProviderEventID:  "evt_123",
		EventType:        "refund.succeeded",
		Event:            PaymentEventFullRefundSucceeded,
		ProviderIntentID: "pi_123",
		Payload:          json.RawMessage(`{"type":"refund.succeeded"}`),
	}
	if err := event.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want refund outside task-5 scope rejected")
	}
}

func webhookTestPayment(status PaymentStatus) Payment {
	return Payment{
		PaymentID:        "pay_123",
		OrderID:          "ord_123",
		UserID:           "usr_123",
		Provider:         "stripe_like",
		ProviderIntentID: "pi_123",
		Status:           status,
		Amount:           Money{Amount: 1000, Currency: "INR"},
		IdempotencyKey:   "payment_intent:ord_123:1",
	}
}
