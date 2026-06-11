package domain

import (
	"errors"
	"testing"
	"time"
)

func TestValidateRefundReservationProtectsCapturedBalance(t *testing.T) {
	payment := refundablePayment()
	refund := testRefund(RefundStatusRequested, 400)
	if err := ValidateRefundReservation(payment, refund, 100); err != nil {
		t.Fatalf("ValidateRefundReservation() error = %v", err)
	}
	if err := ValidateRefundReservation(payment, testRefund(RefundStatusRequested, 901), 100); !errors.Is(err, ErrRefundExceedsAvailable) {
		t.Fatalf("over-refund error = %v, want exceeds available", err)
	}
	payment.Status = PaymentStatusAuthorized
	if err := ValidateRefundReservation(payment, refund, 0); !errors.Is(err, ErrPaymentNotRefundable) {
		t.Fatalf("uncaptured error = %v, want not refundable", err)
	}
}

func TestRefundValidationRequiresLifecycleAuditEvidence(t *testing.T) {
	refund := testRefund(RefundStatusApproved, 250)
	if err := refund.Validate(); !errors.Is(err, ErrInvalidRefund) {
		t.Fatalf("approved refund without review error = %v, want invalid refund", err)
	}
	at := time.Now()
	refund.ReviewedBy = "finance_2"
	refund.ReviewReason = "Approved"
	refund.ReviewedAt = &at
	refund.Status = RefundStatusProcessing
	if err := refund.Validate(); !errors.Is(err, ErrInvalidRefund) {
		t.Fatalf("processing refund without provider id error = %v, want invalid refund", err)
	}
}

func TestApplyVerifiedRefundOutcomeUpdatesOnlySuccessfulTotals(t *testing.T) {
	payment := refundablePayment()
	refund := testRefund(RefundStatusProcessing, 250)
	event := VerifiedRefundWebhook{
		Provider: "stripe_like", Outcome: RefundStatusSucceeded, RefundID: refund.RefundID,
		ProviderRefundID: "re_123", Amount: refund.Amount,
	}
	updatedPayment, updatedRefund, reason := ApplyVerifiedRefundOutcome(payment, refund, event, time.Now())
	if reason != "" || updatedRefund.Status != RefundStatusSucceeded {
		t.Fatalf("updated refund=%+v reason=%q, want succeeded", updatedRefund, reason)
	}
	if updatedPayment.RefundedAmount != 250 || updatedPayment.Status != PaymentStatusPartiallyRefunded {
		t.Fatalf("updated payment=%+v, want partial totals", updatedPayment)
	}

	failed := testRefund(RefundStatusProcessing, 250)
	event.Outcome = RefundStatusFailed
	failedPayment, failedRefund, reason := ApplyVerifiedRefundOutcome(payment, failed, event, time.Now())
	if reason != "" || failedRefund.Status != RefundStatusFailed || failedPayment.RefundedAmount != 0 {
		t.Fatalf("failure outcome=%+v/%+v/%q, want no aggregate update", failedPayment, failedRefund, reason)
	}
}

func TestApplyVerifiedRefundOutcomeAcceptsTerminalCallbackRacingProcessingPersistence(t *testing.T) {
	payment := refundablePayment()
	refund := testRefund(RefundStatusApproved, 1000)
	event := VerifiedRefundWebhook{Provider: "stripe_like", Outcome: RefundStatusSucceeded, RefundID: refund.RefundID, ProviderRefundID: "re_early", Amount: refund.Amount}
	updatedPayment, updatedRefund, reason := ApplyVerifiedRefundOutcome(payment, refund, event, time.Now())
	if reason != "" || updatedRefund.Status != RefundStatusSucceeded || updatedPayment.Status != PaymentStatusRefunded {
		t.Fatalf("outcome=%+v/%+v/%q, want verified full refund despite response race", updatedPayment, updatedRefund, reason)
	}
}

func refundablePayment() Payment {
	return Payment{
		PaymentID: "pay_123", OrderID: "ord_123", UserID: "usr_123", Provider: "stripe_like",
		ProviderPaymentID: "ch_123", Status: PaymentStatusCaptured,
		Amount: Money{Amount: 1000, Currency: "INR"}, CapturedAmount: 1000,
	}
}

func testRefund(status RefundStatus, amount int64) Refund {
	return Refund{
		RefundID: "rfnd_123", PaymentID: "pay_123", Status: status,
		Amount: Money{Amount: amount, Currency: "INR"}, Reason: "Cancelled item",
		RequestedBy: "finance_1", IdempotencyKey: "refund:pay_123:item_1",
	}
}
