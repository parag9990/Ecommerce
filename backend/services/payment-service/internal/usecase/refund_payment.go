package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider"
)

var (
	ErrRefundIdempotencyConflict = errors.New("refund idempotency key is already used for different data")
	ErrRefundMakerChecker        = errors.New("refund requester cannot approve a manual-review refund")
	ErrRefundReviewDecision      = errors.New("refund review decision is invalid")
	ErrRefundProviderMismatch    = errors.New("payment provider returned inconsistent refund data")
)

type RefundRepository interface {
	GetPaymentByID(ctx context.Context, paymentID string) (domain.Payment, error)
	GetRefundByID(ctx context.Context, refundID string) (domain.Refund, error)
	FindRefundByIdempotencyKey(ctx context.Context, paymentID string, idempotencyKey string) (domain.Refund, error)
	ReserveRefund(ctx context.Context, refund domain.Refund) (domain.Payment, error)
	ReviewRefund(ctx context.Context, refundID string, approved bool, reviewedBy string, reviewReason string, at time.Time) (domain.Refund, domain.Payment, error)
	MarkRefundProcessing(ctx context.Context, refundID string, providerRefundID string, at time.Time) (domain.Refund, error)
	CompleteRefund(ctx context.Context, refundID string, providerRefundID string, outcome domain.RefundStatus, at time.Time) (domain.Refund, domain.Payment, error)
}

type RefundIDGenerator interface {
	NewRefundID() (string, error)
	NewRefundEventID() (string, error)
}

type RefundPolicy struct {
	// Zero is fail-safe: every refund must be reviewed before provider submission.
	ManualReviewThresholdMinor int64
}

func (p RefundPolicy) Validate() error {
	if p.ManualReviewThresholdMinor < 0 {
		return errors.New("PAYMENT_REFUND_MANUAL_REVIEW_THRESHOLD_MINOR cannot be negative")
	}
	return nil
}

func (p RefundPolicy) RequiresManualReview(amount int64) bool {
	return p.ManualReviewThresholdMinor == 0 || amount >= p.ManualReviewThresholdMinor
}

type RefundPaymentUsecase struct {
	repository RefundRepository
	providers  PaymentProviderRegistry
	publisher  PaymentEventPublisher
	policy     RefundPolicy
	ids        RefundIDGenerator
	clock      func() time.Time
	logger     *slog.Logger
}

type RefundPaymentOption func(*RefundPaymentUsecase)

func WithRefundIDGenerator(ids RefundIDGenerator) RefundPaymentOption {
	return func(u *RefundPaymentUsecase) {
		if ids != nil {
			u.ids = ids
		}
	}
}

func WithRefundClock(clock func() time.Time) RefundPaymentOption {
	return func(u *RefundPaymentUsecase) {
		if clock != nil {
			u.clock = clock
		}
	}
}

type RefundPaymentInput struct {
	PaymentID      string
	AmountMinor    int64
	Currency       string
	Reason         string
	RequestedBy    string
	IdempotencyKey string
	RequestID      string
}

type ReviewRefundInput struct {
	RefundID   string
	Decision   string
	Reason     string
	ReviewedBy string
}

type RefundPaymentOutput struct {
	Refund   domain.Refund
	Replayed bool
}

func NewRefundPaymentUsecase(
	repository RefundRepository,
	providers PaymentProviderRegistry,
	publisher PaymentEventPublisher,
	policy RefundPolicy,
	logger *slog.Logger,
	opts ...RefundPaymentOption,
) (*RefundPaymentUsecase, error) {
	if repository == nil {
		return nil, errors.New("refund repository is required")
	}
	if providers == nil {
		return nil, errors.New("payment provider registry is required")
	}
	if publisher == nil {
		return nil, errors.New("payment event publisher is required")
	}
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	u := &RefundPaymentUsecase{
		repository: repository,
		providers:  providers,
		publisher:  publisher,
		policy:     policy,
		ids:        cryptoRefundIDGenerator{},
		clock:      func() time.Time { return time.Now().UTC() },
		logger:     logger,
	}
	for _, opt := range opts {
		opt(u)
	}
	return u, nil
}

func (u *RefundPaymentUsecase) Execute(ctx context.Context, input RefundPaymentInput) (RefundPaymentOutput, error) {
	if err := ctx.Err(); err != nil {
		return RefundPaymentOutput{}, err
	}
	input = normalizeRefundPaymentInput(input)
	if err := validateRefundPaymentInput(input); err != nil {
		return RefundPaymentOutput{}, err
	}
	request := domain.Refund{
		PaymentID:      input.PaymentID,
		Status:         domain.RefundStatusRequested,
		Amount:         domain.Money{Amount: input.AmountMinor, Currency: input.Currency},
		Reason:         input.Reason,
		RequestedBy:    input.RequestedBy,
		IdempotencyKey: input.IdempotencyKey,
	}
	existing, err := u.repository.FindRefundByIdempotencyKey(ctx, input.PaymentID, input.IdempotencyKey)
	if err == nil {
		if !existing.MatchesRequest(request) {
			return RefundPaymentOutput{}, ErrRefundIdempotencyConflict
		}
		return u.replayExisting(ctx, existing, input.RequestID)
	}
	if !errors.Is(err, domain.ErrRefundRecordNotFound) {
		return RefundPaymentOutput{}, err
	}

	refundID, err := u.ids.NewRefundID()
	if err != nil {
		return RefundPaymentOutput{}, err
	}
	now := u.clock().UTC()
	request.RefundID = refundID
	request.CreatedAt = now
	request.UpdatedAt = now
	if err := request.Validate(); err != nil {
		return RefundPaymentOutput{}, err
	}
	payment, err := u.repository.ReserveRefund(ctx, request)
	if errors.Is(err, domain.ErrDuplicateRefundRecord) {
		existing, loadErr := u.repository.FindRefundByIdempotencyKey(ctx, input.PaymentID, input.IdempotencyKey)
		if loadErr != nil {
			return RefundPaymentOutput{}, loadErr
		}
		if !existing.MatchesRequest(request) {
			return RefundPaymentOutput{}, ErrRefundIdempotencyConflict
		}
		return u.replayExisting(ctx, existing, input.RequestID)
	}
	if err != nil {
		return RefundPaymentOutput{}, err
	}
	u.publishRefundEvent(ctx, "RefundRequested", request, payment, now)
	u.logger.InfoContext(ctx, "payment.refund.requested",
		slog.String("refund_id", request.RefundID),
		slog.String("payment_id", request.PaymentID),
		slog.String("requested_by", request.RequestedBy),
		slog.Int64("amount_minor", request.Amount.Amount),
		slog.String("currency", request.Amount.Currency),
	)
	if u.policy.RequiresManualReview(request.Amount.Amount) {
		return RefundPaymentOutput{Refund: request}, nil
	}
	approved, payment, err := u.repository.ReviewRefund(ctx, request.RefundID, true, "system:auto_approval", "Approved by configured automatic refund policy", now)
	if err != nil {
		return RefundPaymentOutput{}, err
	}
	u.publishRefundEvent(ctx, "RefundApproved", approved, payment, now)
	return u.submitApprovedRefund(ctx, approved, payment, input.RequestID, false)
}

func (u *RefundPaymentUsecase) Get(ctx context.Context, refundID string) (domain.Refund, error) {
	if err := ctx.Err(); err != nil {
		return domain.Refund{}, err
	}
	return u.repository.GetRefundByID(ctx, strings.TrimSpace(refundID))
}

func (u *RefundPaymentUsecase) Review(ctx context.Context, input ReviewRefundInput) (RefundPaymentOutput, error) {
	if err := ctx.Err(); err != nil {
		return RefundPaymentOutput{}, err
	}
	input.RefundID = strings.TrimSpace(input.RefundID)
	input.Decision = strings.ToLower(strings.TrimSpace(input.Decision))
	input.Reason = strings.TrimSpace(input.Reason)
	input.ReviewedBy = strings.TrimSpace(input.ReviewedBy)
	if input.RefundID == "" || input.ReviewedBy == "" || input.Reason == "" || len(input.Reason) > 512 {
		return RefundPaymentOutput{}, fmt.Errorf("%w: refund_id, reviewer, and reason are required", provider.ErrInvalidProviderRequest)
	}
	if input.Decision != string(domain.RefundStatusApproved) && input.Decision != string(domain.RefundStatusRejected) {
		return RefundPaymentOutput{}, ErrRefundReviewDecision
	}
	existing, err := u.repository.GetRefundByID(ctx, input.RefundID)
	if err != nil {
		return RefundPaymentOutput{}, err
	}
	if input.Decision == string(domain.RefundStatusApproved) &&
		u.policy.RequiresManualReview(existing.Amount.Amount) &&
		existing.RequestedBy == input.ReviewedBy {
		return RefundPaymentOutput{}, ErrRefundMakerChecker
	}
	approved := input.Decision == string(domain.RefundStatusApproved)
	refund, payment, err := u.repository.ReviewRefund(ctx, input.RefundID, approved, input.ReviewedBy, input.Reason, u.clock())
	if err != nil {
		return RefundPaymentOutput{}, err
	}
	eventType := "RefundRejected"
	if approved {
		eventType = "RefundApproved"
	}
	u.publishRefundEvent(ctx, eventType, refund, payment, refund.UpdatedAt)
	if !approved {
		return RefundPaymentOutput{Refund: refund}, nil
	}
	return u.submitApprovedRefund(ctx, refund, payment, "", false)
}

func (u *RefundPaymentUsecase) replayExisting(ctx context.Context, refund domain.Refund, requestID string) (RefundPaymentOutput, error) {
	if refund.Status != domain.RefundStatusApproved {
		return RefundPaymentOutput{Refund: refund, Replayed: true}, nil
	}
	payment, err := u.repository.GetPaymentByID(ctx, refund.PaymentID)
	if err != nil {
		return RefundPaymentOutput{}, err
	}
	return u.submitApprovedRefund(ctx, refund, payment, requestID, true)
}

func (u *RefundPaymentUsecase) submitApprovedRefund(ctx context.Context, refund domain.Refund, payment domain.Payment, requestID string, replayed bool) (RefundPaymentOutput, error) {
	gateway, err := u.providers.Get(payment.Provider)
	if err != nil {
		return RefundPaymentOutput{}, err
	}
	req := provider.RefundRequest{
		PaymentID:         payment.PaymentID,
		RefundID:          refund.RefundID,
		ProviderPaymentID: payment.ProviderPaymentID,
		Amount:            provider.Money{AmountMinor: refund.Amount.Amount, Currency: refund.Amount.Currency},
		Reason:            refund.Reason,
		IdempotencyKey:    refund.IdempotencyKey,
		Metadata: provider.NormalizeMetadata(map[string]string{
			"payment_id": payment.PaymentID,
			"order_id":   payment.OrderID,
			"refund_id":  refund.RefundID,
			"request_id": requestID,
		}),
	}
	if err := provider.ValidateRefundRequest(req); err != nil {
		return RefundPaymentOutput{}, err
	}
	res, err := gateway.Refund(ctx, req)
	if err != nil {
		if !provider.IsRetryable(err) {
			failed, failedPayment, persistErr := u.repository.CompleteRefund(ctx, refund.RefundID, "", domain.RefundStatusFailed, u.clock())
			if persistErr == nil {
				u.publishRefundEvent(ctx, "RefundFailed", failed, failedPayment, failed.UpdatedAt)
			}
		}
		return RefundPaymentOutput{}, err
	}
	res = res.Normalized()
	if err := provider.ValidateRefundResponse(res); err != nil {
		return RefundPaymentOutput{}, err
	}
	if res.RefundedAmount.AmountMinor != refund.Amount.Amount || res.RefundedAmount.Currency != refund.Amount.Currency {
		return RefundPaymentOutput{}, ErrRefundProviderMismatch
	}
	processing, err := u.repository.MarkRefundProcessing(ctx, refund.RefundID, res.ProviderRefundID, u.clock())
	if err != nil {
		return RefundPaymentOutput{}, err
	}
	u.publishRefundEvent(ctx, "RefundProcessing", processing, payment, processing.UpdatedAt)
	if res.Status == provider.RefundStatusProcessing {
		return RefundPaymentOutput{Refund: processing, Replayed: replayed}, nil
	}
	outcome := domain.RefundStatusSucceeded
	eventType := "RefundSucceeded"
	if res.Status == provider.RefundStatusFailed {
		outcome = domain.RefundStatusFailed
		eventType = "RefundFailed"
	}
	completed, updatedPayment, err := u.repository.CompleteRefund(ctx, refund.RefundID, res.ProviderRefundID, outcome, u.clock())
	if err != nil {
		return RefundPaymentOutput{}, err
	}
	u.publishRefundEvent(ctx, eventType, completed, updatedPayment, completed.UpdatedAt)
	return RefundPaymentOutput{Refund: completed, Replayed: replayed}, nil
}

func (u *RefundPaymentUsecase) publishRefundEvent(ctx context.Context, eventType string, refund domain.Refund, payment domain.Payment, at time.Time) {
	eventID, err := u.ids.NewRefundEventID()
	if err != nil {
		u.logger.ErrorContext(ctx, "payment.refund.event_id_failed", slog.String("refund_id", refund.RefundID), slog.String("error", err.Error()))
		return
	}
	event := domain.RefundDomainEvent(eventID, eventType, refund, payment, at)
	if err := u.publisher.Publish(ctx, PaymentEventsTopic, event); err != nil {
		u.logger.ErrorContext(ctx, "payment.refund.event_publish_failed",
			slog.String("refund_id", refund.RefundID),
			slog.String("payment_id", refund.PaymentID),
			slog.String("event_type", eventType),
			slog.String("error", err.Error()),
		)
	}
}

func normalizeRefundPaymentInput(input RefundPaymentInput) RefundPaymentInput {
	input.PaymentID = strings.TrimSpace(input.PaymentID)
	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	input.Reason = strings.TrimSpace(input.Reason)
	input.RequestedBy = strings.TrimSpace(input.RequestedBy)
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	input.RequestID = strings.TrimSpace(input.RequestID)
	return input
}

func validateRefundPaymentInput(input RefundPaymentInput) error {
	if input.PaymentID == "" || len(input.PaymentID) > 64 {
		return fmt.Errorf("%w: valid payment_id is required", provider.ErrInvalidProviderRequest)
	}
	if input.RequestedBy == "" || len(input.RequestedBy) > 64 {
		return fmt.Errorf("%w: valid requested_by is required", provider.ErrInvalidProviderRequest)
	}
	if input.Reason == "" || len(input.Reason) > 512 {
		return fmt.Errorf("%w: reason is required and cannot exceed 512 characters", provider.ErrInvalidProviderRequest)
	}
	if err := provider.ValidateMoney(provider.Money{AmountMinor: input.AmountMinor, Currency: input.Currency}); err != nil {
		return err
	}
	return provider.ValidateIdempotencyKey(input.IdempotencyKey)
}

type cryptoRefundIDGenerator struct{}

func (cryptoRefundIDGenerator) NewRefundID() (string, error) {
	return randomRefundIdentifier("rfnd_")
}

func (cryptoRefundIDGenerator) NewRefundEventID() (string, error) {
	return randomRefundIdentifier("rfe_")
}

func randomRefundIdentifier(prefix string) (string, error) {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate refund identifier: %w", err)
	}
	return prefix + strings.ToLower(hex.EncodeToString(random)), nil
}
