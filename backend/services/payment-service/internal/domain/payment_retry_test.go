package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewRetryPaymentCopiesTrustedParentAndRecordsLineage(t *testing.T) {
	parent, err := NewPayment(NewPaymentInput{
		PaymentID:      "pay_parent",
		OrderID:        "ord_123",
		UserID:         "usr_123",
		Provider:       "stripe_like",
		Amount:         Money{Amount: 129900, Currency: "inr"},
		IdempotencyKey: "payment_intent:ord_123:1",
		Now:            time.Now(),
	})
	if err != nil {
		t.Fatalf("NewPayment() error = %v", err)
	}
	parent.Status = PaymentStatusFailed
	child, err := NewRetryPayment(NewRetryPaymentInput{
		Parent:                 parent,
		PaymentID:              "pay_retry",
		RootPaymentID:          parent.PaymentID,
		AttemptNo:              2,
		RetryRequestKey:        "retry_request_123",
		ProviderIdempotencyKey: "payment_intent:ord_123:2",
		Now:                    time.Now(),
	})
	if err != nil {
		t.Fatalf("NewRetryPayment() error = %v", err)
	}
	if child.RetryOfPaymentID != parent.PaymentID || child.RootPaymentID != parent.PaymentID || child.AttemptNo != 2 {
		t.Fatalf("child lineage = %+v, want linked attempt 2", child)
	}
	if child.Amount != parent.Amount || child.UserID != parent.UserID || child.Status != PaymentStatusInitiated {
		t.Fatalf("child = %+v, want trusted parent payment values", child)
	}
}

func TestNewRetryPaymentRejectsNonFailedParentAndIncompleteLineage(t *testing.T) {
	parent, err := NewPayment(NewPaymentInput{
		PaymentID: "pay_parent", OrderID: "ord_123", UserID: "usr_123", Provider: "stripe_like",
		Amount: Money{Amount: 1000, Currency: "INR"}, IdempotencyKey: "payment_intent:ord_123:1",
	})
	if err != nil {
		t.Fatalf("NewPayment() error = %v", err)
	}
	if _, err := NewRetryPayment(NewRetryPaymentInput{Parent: parent}); !errors.Is(err, ErrPaymentNotRetryable) {
		t.Fatalf("NewRetryPayment() error = %v, want non-retryable", err)
	}
	parent.Status = PaymentStatusFailed
	child, err := NewRetryPayment(NewRetryPaymentInput{
		Parent: parent, PaymentID: "pay_retry", RootPaymentID: parent.PaymentID, AttemptNo: 2,
		RetryRequestKey: "retry_request_123", ProviderIdempotencyKey: "payment_intent:ord_123:2",
	})
	if err != nil {
		t.Fatalf("NewRetryPayment() error = %v", err)
	}
	child.RootPaymentID = ""
	if err := child.ValidateForPersistence(); !errors.Is(err, ErrInvalidPayment) {
		t.Fatalf("ValidateForPersistence() error = %v, want invalid incomplete lineage", err)
	}
}
