package domain

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidPayment           = errors.New("invalid payment")
	ErrInvalidPaymentStatus     = errors.New("invalid payment status")
	ErrInvalidPaymentEvent      = errors.New("invalid payment event")
	ErrInvalidPaymentTransition = errors.New("invalid payment transition")
	ErrUnknownProviderEvent     = errors.New("unknown provider event")
	ErrDuplicatePaymentRecord   = errors.New("duplicate payment record")
	ErrPaymentRecordNotFound    = errors.New("payment record not found")
	ErrPaymentSchemaUnavailable = errors.New("payment schema unavailable")
	ErrPaymentNotRetryable      = errors.New("payment is not retryable")
	ErrPaymentRetryInProgress   = errors.New("another payment attempt is active")
	ErrPaymentAlreadyCaptured   = errors.New("payment has already been captured")
	ErrPaymentRetryLimitReached = errors.New("payment retry limit reached")
	ErrPaymentRetryCooldown     = errors.New("payment retry cooldown is active")
	ErrInvalidWebhookEvent      = errors.New("invalid payment webhook event")
	ErrWebhookPaymentAmbiguous  = errors.New("webhook payment lookup is ambiguous")
	ErrInvalidRefund            = errors.New("invalid refund")
	ErrRefundRecordNotFound     = errors.New("refund record not found")
	ErrDuplicateRefundRecord    = errors.New("duplicate refund record")
	ErrPaymentNotRefundable     = errors.New("payment is not refundable")
	ErrRefundExceedsAvailable   = errors.New("refund exceeds available captured amount")
	ErrRefundCurrencyMismatch   = errors.New("refund currency does not match payment")
	ErrInvalidRefundTransition  = errors.New("invalid refund transition")
	ErrInvalidReconciliation    = errors.New("invalid payment reconciliation")
	ErrReconciliationNotFound   = errors.New("payment reconciliation not found")
	ErrReconciliationConflict   = errors.New("payment reconciliation conflicts with previously stored evidence")
	ErrProviderPaymentAmbiguous = errors.New("provider payment reference maps to multiple local payments")
)

type InvalidPaymentStatusError struct {
	Status PaymentStatus
}

func (e InvalidPaymentStatusError) Error() string {
	return fmt.Sprintf("%s: %s", ErrInvalidPaymentStatus.Error(), e.Status)
}

func (e InvalidPaymentStatusError) Unwrap() error {
	return ErrInvalidPaymentStatus
}

type InvalidPaymentEventError struct {
	Event PaymentEvent
}

func (e InvalidPaymentEventError) Error() string {
	return fmt.Sprintf("%s: %s", ErrInvalidPaymentEvent.Error(), e.Event)
}

func (e InvalidPaymentEventError) Unwrap() error {
	return ErrInvalidPaymentEvent
}

type InvalidPaymentTransitionError struct {
	From  PaymentStatus
	To    PaymentStatus
	Event PaymentEvent
}

func (e InvalidPaymentTransitionError) Error() string {
	if e.Event != "" {
		return fmt.Sprintf("%s: %s --%s--> %s", ErrInvalidPaymentTransition.Error(), e.From, e.Event, e.To)
	}
	return fmt.Sprintf("%s: %s -> %s", ErrInvalidPaymentTransition.Error(), e.From, e.To)
}

func (e InvalidPaymentTransitionError) Unwrap() error {
	return ErrInvalidPaymentTransition
}
