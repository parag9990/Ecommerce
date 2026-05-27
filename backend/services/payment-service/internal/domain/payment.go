package domain

import (
	"fmt"
	"strings"
	"time"
)

type Money struct {
	Amount   int64
	Currency string
}

func (m Money) Validate() error {
	if m.Amount <= 0 {
		return fmt.Errorf("%w: amount must be greater than zero", ErrInvalidPayment)
	}
	if len(strings.TrimSpace(m.Currency)) != 3 {
		return fmt.Errorf("%w: currency must be a 3-letter ISO code", ErrInvalidPayment)
	}
	return nil
}

func (m Money) Normalized() Money {
	return Money{
		Amount:   m.Amount,
		Currency: strings.ToUpper(strings.TrimSpace(m.Currency)),
	}
}

type Payment struct {
	PaymentID         string
	OrderID           string
	UserID            string
	Provider          string
	ProviderPaymentID string
	ProviderIntentID  string
	Amount            Money
	Status            PaymentStatus
	CapturedAmount    int64
	RefundedAmount    int64
	IdempotencyKey    string
	RetryOfPaymentID  string
	RootPaymentID     string
	AttemptNo         uint32
	RetryRequestKey   string
	FailureCode       string
	FailureMessage    string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type NewPaymentInput struct {
	PaymentID        string
	OrderID          string
	UserID           string
	Provider         string
	ProviderIntentID string
	Amount           Money
	IdempotencyKey   string
	Now              time.Time
}

func NewPayment(input NewPaymentInput) (Payment, error) {
	payment := Payment{
		PaymentID:        strings.TrimSpace(input.PaymentID),
		OrderID:          strings.TrimSpace(input.OrderID),
		UserID:           strings.TrimSpace(input.UserID),
		Provider:         strings.ToLower(strings.TrimSpace(input.Provider)),
		Amount:           input.Amount.Normalized(),
		Status:           PaymentStatusInitiated,
		ProviderIntentID: strings.TrimSpace(input.ProviderIntentID),
		IdempotencyKey:   strings.TrimSpace(input.IdempotencyKey),
		AttemptNo:        1,
		CreatedAt:        input.Now.UTC(),
		UpdatedAt:        input.Now.UTC(),
	}
	if payment.CreatedAt.IsZero() {
		now := time.Now().UTC()
		payment.CreatedAt = now
		payment.UpdatedAt = now
	}
	if err := payment.Validate(); err != nil {
		return Payment{}, err
	}
	return payment, nil
}

func (p Payment) Normalized() Payment {
	p.PaymentID = strings.TrimSpace(p.PaymentID)
	p.OrderID = strings.TrimSpace(p.OrderID)
	p.UserID = strings.TrimSpace(p.UserID)
	p.Provider = strings.ToLower(strings.TrimSpace(p.Provider))
	p.ProviderPaymentID = strings.TrimSpace(p.ProviderPaymentID)
	p.ProviderIntentID = strings.TrimSpace(p.ProviderIntentID)
	p.Amount = p.Amount.Normalized()
	p.IdempotencyKey = strings.TrimSpace(p.IdempotencyKey)
	p.RetryOfPaymentID = strings.TrimSpace(p.RetryOfPaymentID)
	p.RootPaymentID = strings.TrimSpace(p.RootPaymentID)
	p.RetryRequestKey = strings.TrimSpace(p.RetryRequestKey)
	p.FailureCode = strings.TrimSpace(p.FailureCode)
	p.FailureMessage = strings.TrimSpace(p.FailureMessage)
	p.CreatedAt = p.CreatedAt.UTC()
	p.UpdatedAt = p.UpdatedAt.UTC()
	return p
}

func (p Payment) Validate() error {
	p = p.Normalized()
	if strings.TrimSpace(p.PaymentID) == "" {
		return fmt.Errorf("%w: payment_id is required", ErrInvalidPayment)
	}
	if strings.TrimSpace(p.OrderID) == "" {
		return fmt.Errorf("%w: order_id is required", ErrInvalidPayment)
	}
	if strings.TrimSpace(p.UserID) == "" {
		return fmt.Errorf("%w: user_id is required", ErrInvalidPayment)
	}
	if strings.TrimSpace(p.Provider) == "" {
		return fmt.Errorf("%w: provider is required", ErrInvalidPayment)
	}
	if err := p.Amount.Validate(); err != nil {
		return err
	}
	if !NewStateMachine().IsKnownStatus(p.Status) || p.Status == PaymentStatusUnspecified {
		return InvalidPaymentStatusError{Status: p.Status}
	}
	if p.CapturedAmount < 0 {
		return fmt.Errorf("%w: captured amount cannot be negative", ErrInvalidPayment)
	}
	if p.RefundedAmount < 0 {
		return fmt.Errorf("%w: refunded amount cannot be negative", ErrInvalidPayment)
	}
	if p.RefundedAmount > p.CapturedAmount && p.CapturedAmount > 0 {
		return fmt.Errorf("%w: refunded amount cannot exceed captured amount", ErrInvalidPayment)
	}
	return nil
}

func (p Payment) ValidateForPersistence() error {
	p = p.Normalized()
	if err := p.Validate(); err != nil {
		return err
	}
	if !IsPersistedPaymentStatus(p.Status) {
		return fmt.Errorf("%w: payment status %s cannot be persisted", ErrInvalidPayment, p.Status)
	}
	if p.IdempotencyKey == "" {
		return fmt.Errorf("%w: idempotency_key is required", ErrInvalidPayment)
	}
	if len(p.IdempotencyKey) > 128 {
		return fmt.Errorf("%w: idempotency_key exceeds 128 characters", ErrInvalidPayment)
	}
	if len(p.ProviderPaymentID) > 128 {
		return fmt.Errorf("%w: provider_payment_id exceeds 128 characters", ErrInvalidPayment)
	}
	if len(p.ProviderIntentID) > 128 {
		return fmt.Errorf("%w: provider_intent_id exceeds 128 characters", ErrInvalidPayment)
	}
	if len(p.FailureCode) > 128 {
		return fmt.Errorf("%w: failure_code exceeds 128 characters", ErrInvalidPayment)
	}
	if len(p.FailureMessage) > 512 {
		return fmt.Errorf("%w: failure_message exceeds 512 characters", ErrInvalidPayment)
	}
	if err := p.validateRetryLineage(); err != nil {
		return err
	}
	return nil
}

func (p Payment) validateRetryLineage() error {
	switch {
	case p.AttemptNo == 0:
		return fmt.Errorf("%w: attempt_no must be greater than zero", ErrInvalidPayment)
	case p.RetryOfPaymentID == "" && p.RootPaymentID == "" && p.RetryRequestKey == "":
		if p.AttemptNo != 1 {
			return fmt.Errorf("%w: an initial payment must have attempt_no 1", ErrInvalidPayment)
		}
		return nil
	case p.RetryOfPaymentID == "" || p.RootPaymentID == "" || p.RetryRequestKey == "":
		return fmt.Errorf("%w: retry lineage fields must be set together", ErrInvalidPayment)
	case p.AttemptNo <= 1:
		return fmt.Errorf("%w: a retry payment must have attempt_no greater than 1", ErrInvalidPayment)
	case p.RetryOfPaymentID == p.PaymentID:
		return fmt.Errorf("%w: payment cannot retry itself", ErrInvalidPayment)
	case len(p.RetryOfPaymentID) > 64 || len(p.RootPaymentID) > 64:
		return fmt.Errorf("%w: retry payment references exceed 64 characters", ErrInvalidPayment)
	case len(p.RetryRequestKey) > 128:
		return fmt.Errorf("%w: retry_request_key exceeds 128 characters", ErrInvalidPayment)
	default:
		return nil
	}
}

func (p *Payment) ApplyTransition(to PaymentStatus, at time.Time) (TransitionDecision, error) {
	if p == nil {
		return TransitionDecision{}, fmt.Errorf("%w: payment is nil", ErrInvalidPayment)
	}
	machine := NewStateMachine()
	decision := machine.EvaluateTransition(p.Status, to)
	if !decision.Allowed {
		return decision, InvalidPaymentTransitionError{From: p.Status, To: to}
	}
	if !decision.Noop {
		p.Status = to
		p.UpdatedAt = normalizeTransitionTime(at)
	}
	return decision, nil
}

func (p *Payment) ApplyEvent(event PaymentEvent, at time.Time) (TransitionDecision, error) {
	if p == nil {
		return TransitionDecision{}, fmt.Errorf("%w: payment is nil", ErrInvalidPayment)
	}
	machine := NewStateMachine()
	decision := machine.EvaluateEvent(p.Status, event)
	if !decision.Allowed {
		return decision, InvalidPaymentTransitionError{From: p.Status, To: decision.To, Event: event}
	}
	if !decision.Noop {
		p.Status = decision.To
		p.UpdatedAt = normalizeTransitionTime(at)
	}
	return decision, nil
}

func (p Payment) SuggestedOrderStatus() (OrderPaymentStatus, bool) {
	return SuggestedOrderStatus(p.Status)
}

func (p Payment) PaidForFulfillment() bool {
	return p.Status == PaymentStatusCaptured ||
		p.Status == PaymentStatusPartiallyRefunded
}

func normalizeTransitionTime(at time.Time) time.Time {
	if at.IsZero() {
		return time.Now().UTC()
	}
	return at.UTC()
}
