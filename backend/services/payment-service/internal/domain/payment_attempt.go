package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type PaymentAttemptStatus string

const (
	PaymentAttemptStatusUnspecified PaymentAttemptStatus = ""
	PaymentAttemptStatusInitiated   PaymentAttemptStatus = "initiated"
	PaymentAttemptStatusSucceeded   PaymentAttemptStatus = "succeeded"
	PaymentAttemptStatusFailed      PaymentAttemptStatus = "failed"
)

type PaymentAttempt struct {
	AttemptID           string
	PaymentID           string
	ProviderAttemptID   string
	Status              PaymentAttemptStatus
	FailureCode         string
	FailureMessage      string
	RawProviderResponse json.RawMessage
	CreatedAt           time.Time
}

func (a PaymentAttempt) Normalized() PaymentAttempt {
	a.AttemptID = strings.TrimSpace(a.AttemptID)
	a.PaymentID = strings.TrimSpace(a.PaymentID)
	a.ProviderAttemptID = strings.TrimSpace(a.ProviderAttemptID)
	a.FailureCode = strings.TrimSpace(a.FailureCode)
	a.FailureMessage = strings.TrimSpace(a.FailureMessage)
	a.CreatedAt = a.CreatedAt.UTC()
	return a
}

func (a PaymentAttempt) Validate() error {
	a = a.Normalized()
	if a.AttemptID == "" {
		return fmt.Errorf("%w: attempt_id is required", ErrInvalidPayment)
	}
	if a.PaymentID == "" {
		return fmt.Errorf("%w: payment_id is required", ErrInvalidPayment)
	}
	if !IsKnownPaymentAttemptStatus(a.Status) {
		return fmt.Errorf("%w: invalid payment attempt status %s", ErrInvalidPayment, a.Status)
	}
	if len(a.ProviderAttemptID) > 128 {
		return fmt.Errorf("%w: provider_attempt_id exceeds 128 characters", ErrInvalidPayment)
	}
	if len(a.FailureCode) > 128 {
		return fmt.Errorf("%w: failure_code exceeds 128 characters", ErrInvalidPayment)
	}
	if len(a.FailureMessage) > 512 {
		return fmt.Errorf("%w: failure_message exceeds 512 characters", ErrInvalidPayment)
	}
	if err := validateSanitizedJSONPayload(a.RawProviderResponse, false, "raw_provider_response"); err != nil {
		return err
	}
	return nil
}

func IsKnownPaymentAttemptStatus(status PaymentAttemptStatus) bool {
	switch status {
	case PaymentAttemptStatusInitiated, PaymentAttemptStatusSucceeded, PaymentAttemptStatusFailed:
		return true
	default:
		return false
	}
}
