package domain

import (
	"fmt"
	"strings"
	"time"
)

type RefundStatus string

const (
	RefundStatusUnspecified RefundStatus = ""
	RefundStatusRequested   RefundStatus = "requested"
	RefundStatusApproved    RefundStatus = "approved"
	RefundStatusRejected    RefundStatus = "rejected"
	RefundStatusProcessing  RefundStatus = "processing"
	RefundStatusSucceeded   RefundStatus = "succeeded"
	RefundStatusFailed      RefundStatus = "failed"
)

type Refund struct {
	RefundID         string
	PaymentID        string
	ProviderRefundID string
	Status           RefundStatus
	Amount           Money
	Reason           string
	RequestedBy      string
	ReviewedBy       string
	ReviewReason     string
	ReviewedAt       *time.Time
	IdempotencyKey   string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (r Refund) Normalized() Refund {
	r.RefundID = strings.TrimSpace(r.RefundID)
	r.PaymentID = strings.TrimSpace(r.PaymentID)
	r.ProviderRefundID = strings.TrimSpace(r.ProviderRefundID)
	r.Amount = r.Amount.Normalized()
	r.Reason = strings.TrimSpace(r.Reason)
	r.RequestedBy = strings.TrimSpace(r.RequestedBy)
	r.ReviewedBy = strings.TrimSpace(r.ReviewedBy)
	r.ReviewReason = strings.TrimSpace(r.ReviewReason)
	if r.ReviewedAt != nil {
		reviewedAt := r.ReviewedAt.UTC()
		r.ReviewedAt = &reviewedAt
	}
	r.IdempotencyKey = strings.TrimSpace(r.IdempotencyKey)
	r.CreatedAt = r.CreatedAt.UTC()
	r.UpdatedAt = r.UpdatedAt.UTC()
	return r
}

func (r Refund) Validate() error {
	r = r.Normalized()
	if r.RefundID == "" {
		return fmt.Errorf("%w: refund_id is required", ErrInvalidRefund)
	}
	if r.PaymentID == "" {
		return fmt.Errorf("%w: payment_id is required", ErrInvalidRefund)
	}
	if !IsKnownRefundStatus(r.Status) {
		return fmt.Errorf("%w: invalid refund status %s", ErrInvalidRefund, r.Status)
	}
	if err := r.Amount.Validate(); err != nil {
		return err
	}
	if r.Reason == "" {
		return fmt.Errorf("%w: reason is required", ErrInvalidRefund)
	}
	if len(r.Reason) > 512 {
		return fmt.Errorf("%w: reason exceeds 512 characters", ErrInvalidRefund)
	}
	if r.RequestedBy == "" {
		return fmt.Errorf("%w: requested_by is required", ErrInvalidRefund)
	}
	if len(r.RequestedBy) > 64 {
		return fmt.Errorf("%w: requested_by exceeds 64 characters", ErrInvalidRefund)
	}
	if r.ReviewedAt != nil && (r.ReviewedBy == "" || r.ReviewReason == "") {
		return fmt.Errorf("%w: reviewed_by and review_reason are required when reviewed_at is set", ErrInvalidRefund)
	}
	if len(r.ReviewedBy) > 64 {
		return fmt.Errorf("%w: reviewed_by exceeds 64 characters", ErrInvalidRefund)
	}
	if len(r.ReviewReason) > 512 {
		return fmt.Errorf("%w: review_reason exceeds 512 characters", ErrInvalidRefund)
	}
	if r.Status != RefundStatusRequested && (r.ReviewedAt == nil || r.ReviewedBy == "" || r.ReviewReason == "") {
		return fmt.Errorf("%w: reviewed audit fields are required after request status", ErrInvalidRefund)
	}
	if (r.Status == RefundStatusProcessing || r.Status == RefundStatusSucceeded) && r.ProviderRefundID == "" {
		return fmt.Errorf("%w: provider_refund_id is required for provider-confirmed status", ErrInvalidRefund)
	}
	if r.IdempotencyKey == "" {
		return fmt.Errorf("%w: idempotency_key is required", ErrInvalidRefund)
	}
	if len(r.ProviderRefundID) > 128 {
		return fmt.Errorf("%w: provider_refund_id exceeds 128 characters", ErrInvalidRefund)
	}
	if len(r.IdempotencyKey) > 128 {
		return fmt.Errorf("%w: idempotency_key exceeds 128 characters", ErrInvalidRefund)
	}
	return nil
}

func IsKnownRefundStatus(status RefundStatus) bool {
	switch status {
	case RefundStatusRequested, RefundStatusApproved, RefundStatusRejected, RefundStatusProcessing, RefundStatusSucceeded, RefundStatusFailed:
		return true
	default:
		return false
	}
}

func (r Refund) MatchesRequest(other Refund) bool {
	r = r.Normalized()
	other = other.Normalized()
	return r.PaymentID == other.PaymentID &&
		r.Amount == other.Amount &&
		r.Reason == other.Reason &&
		r.RequestedBy == other.RequestedBy &&
		r.IdempotencyKey == other.IdempotencyKey
}

func (r Refund) CanTransitionTo(status RefundStatus) bool {
	if r.Status == status {
		return true
	}
	switch r.Status {
	case RefundStatusRequested:
		return status == RefundStatusApproved || status == RefundStatusRejected
	case RefundStatusApproved:
		// A verified provider callback can arrive before the accepted response is persisted.
		return status == RefundStatusProcessing || status == RefundStatusSucceeded || status == RefundStatusFailed
	case RefundStatusProcessing:
		return status == RefundStatusSucceeded || status == RefundStatusFailed
	default:
		return false
	}
}

func ValidateRefundReservation(payment Payment, refund Refund, reservedOpenAmount int64) error {
	payment = payment.Normalized()
	refund = refund.Normalized()
	if refund.Status != RefundStatusRequested {
		return fmt.Errorf("%w: a reservation must start as requested", ErrInvalidRefund)
	}
	if payment.Status != PaymentStatusCaptured && payment.Status != PaymentStatusPartiallyRefunded {
		return fmt.Errorf("%w: status=%s", ErrPaymentNotRefundable, payment.Status)
	}
	if payment.ProviderPaymentID == "" {
		return fmt.Errorf("%w: provider_payment_id is required for a captured refund", ErrPaymentNotRefundable)
	}
	if refund.Amount.Currency != payment.Amount.Currency {
		return fmt.Errorf("%w: payment=%s refund=%s", ErrRefundCurrencyMismatch, payment.Amount.Currency, refund.Amount.Currency)
	}
	if reservedOpenAmount < 0 {
		return fmt.Errorf("%w: reserved open amount cannot be negative", ErrInvalidRefund)
	}
	available := payment.CapturedAmount - payment.RefundedAmount - reservedOpenAmount
	if available < 0 || refund.Amount.Amount > available {
		return fmt.Errorf("%w: available=%d requested=%d", ErrRefundExceedsAvailable, available, refund.Amount.Amount)
	}
	return nil
}
