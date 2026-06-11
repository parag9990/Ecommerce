package domain

import (
	"fmt"
	"strings"
	"time"
)

type RetryPolicy struct {
	MaxAttempts uint32
	Cooldown    time.Duration
}

func (p RetryPolicy) Validate() error {
	if p.MaxAttempts < 2 {
		return fmt.Errorf("%w: max retry attempts must be at least 2", ErrInvalidPayment)
	}
	if p.Cooldown < 0 {
		return fmt.Errorf("%w: retry cooldown cannot be negative", ErrInvalidPayment)
	}
	return nil
}

type ReserveRetryPaymentInput struct {
	FailedPaymentID string
	BuyerUserID     string
	RetryRequestKey string
	PaymentID       string
	AttemptID       string
	Policy          RetryPolicy
	Now             time.Time
}

func (i ReserveRetryPaymentInput) Normalized() ReserveRetryPaymentInput {
	i.FailedPaymentID = strings.TrimSpace(i.FailedPaymentID)
	i.BuyerUserID = strings.TrimSpace(i.BuyerUserID)
	i.RetryRequestKey = strings.TrimSpace(i.RetryRequestKey)
	i.PaymentID = strings.TrimSpace(i.PaymentID)
	i.AttemptID = strings.TrimSpace(i.AttemptID)
	i.Now = i.Now.UTC()
	return i
}

func (i ReserveRetryPaymentInput) Validate() error {
	i = i.Normalized()
	switch {
	case i.FailedPaymentID == "":
		return fmt.Errorf("%w: failed payment_id is required", ErrInvalidPayment)
	case len(i.FailedPaymentID) > 64:
		return fmt.Errorf("%w: failed payment_id exceeds 64 characters", ErrInvalidPayment)
	case i.BuyerUserID == "":
		return fmt.Errorf("%w: buyer user_id is required", ErrInvalidPayment)
	case len(i.BuyerUserID) > 64:
		return fmt.Errorf("%w: buyer user_id exceeds 64 characters", ErrInvalidPayment)
	case i.RetryRequestKey == "":
		return fmt.Errorf("%w: retry_request_key is required", ErrInvalidPayment)
	case len(i.RetryRequestKey) > 128:
		return fmt.Errorf("%w: retry_request_key exceeds 128 characters", ErrInvalidPayment)
	case i.PaymentID == "" || i.AttemptID == "":
		return fmt.Errorf("%w: retry payment_id and attempt_id are required", ErrInvalidPayment)
	case len(i.PaymentID) > 64 || len(i.AttemptID) > 64:
		return fmt.Errorf("%w: retry generated identifiers exceed 64 characters", ErrInvalidPayment)
	}
	return i.Policy.Validate()
}

type ReservedRetryPayment struct {
	Payment  Payment
	Attempt  PaymentAttempt
	Replayed bool
}

func (p Payment) RetryRootPaymentID() string {
	rootID := strings.TrimSpace(p.RootPaymentID)
	if rootID == "" {
		return strings.TrimSpace(p.PaymentID)
	}
	return rootID
}

type NewRetryPaymentInput struct {
	Parent                 Payment
	PaymentID              string
	RootPaymentID          string
	AttemptNo              uint32
	RetryRequestKey        string
	ProviderIdempotencyKey string
	Now                    time.Time
}

func NewRetryPayment(input NewRetryPaymentInput) (Payment, error) {
	parent := input.Parent.Normalized()
	if parent.Status != PaymentStatusFailed || parent.CapturedAmount > 0 {
		return Payment{}, ErrPaymentNotRetryable
	}
	payment, err := NewPayment(NewPaymentInput{
		PaymentID:      input.PaymentID,
		OrderID:        parent.OrderID,
		UserID:         parent.UserID,
		Provider:       parent.Provider,
		Amount:         parent.Amount,
		IdempotencyKey: input.ProviderIdempotencyKey,
		Now:            input.Now,
	})
	if err != nil {
		return Payment{}, err
	}
	payment.RetryOfPaymentID = parent.PaymentID
	payment.RootPaymentID = strings.TrimSpace(input.RootPaymentID)
	payment.AttemptNo = input.AttemptNo
	payment.RetryRequestKey = strings.TrimSpace(input.RetryRequestKey)
	if err := payment.ValidateForPersistence(); err != nil {
		return Payment{}, err
	}
	return payment.Normalized(), nil
}
