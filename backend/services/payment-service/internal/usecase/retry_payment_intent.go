package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider"
)

type RetryPaymentRepository interface {
	ReserveRetryPayment(ctx context.Context, input domain.ReserveRetryPaymentInput) (domain.ReservedRetryPayment, error)
	GetPaymentByID(ctx context.Context, paymentID string) (domain.Payment, error)
	UpdatePaymentIntent(ctx context.Context, payment domain.Payment, attempt domain.PaymentAttempt) error
	FailPaymentIntent(ctx context.Context, paymentID string, attemptID string, failureCode string, failureMessage string, at time.Time) error
}

type RetryPaymentIntentUsecase struct {
	repository RetryPaymentRepository
	providers  PaymentProviderRegistry
	config     provider.Config
	policy     domain.RetryPolicy
	ids        PaymentIntentIDGenerator
	clock      func() time.Time
	logger     *slog.Logger
}

type RetryPaymentIntentOption func(*RetryPaymentIntentUsecase)

func WithRetryPaymentIntentIDGenerator(ids PaymentIntentIDGenerator) RetryPaymentIntentOption {
	return func(u *RetryPaymentIntentUsecase) {
		if ids != nil {
			u.ids = ids
		}
	}
}

func WithRetryPaymentIntentClock(clock func() time.Time) RetryPaymentIntentOption {
	return func(u *RetryPaymentIntentUsecase) {
		if clock != nil {
			u.clock = clock
		}
	}
}

type RetryPaymentIntentInput struct {
	FailedPaymentID string
	BuyerUserID     string
	RetryRequestKey string
	RequestID       string
}

type RetryPaymentIntentOutput struct {
	PaymentID        string
	OrderID          string
	RetryOfPaymentID string
	RootPaymentID    string
	AttemptNo        uint32
	Provider         string
	ProviderIntentID string
	Status           domain.PaymentStatus
	Amount           int64
	Currency         string
	ClientPayload    PaymentClientPayload
	Replayed         bool
}

func NewRetryPaymentIntentUsecase(
	repository RetryPaymentRepository,
	providers PaymentProviderRegistry,
	cfg provider.Config,
	policy domain.RetryPolicy,
	logger *slog.Logger,
	opts ...RetryPaymentIntentOption,
) (*RetryPaymentIntentUsecase, error) {
	if repository == nil {
		return nil, errors.New("retry payment repository is required")
	}
	if providers == nil {
		return nil, errors.New("payment provider registry is required")
	}
	cfg = cfg.Normalized()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	u := &RetryPaymentIntentUsecase{
		repository: repository,
		providers:  providers,
		config:     cfg,
		policy:     policy,
		ids:        cryptoPaymentIntentIDGenerator{},
		clock:      func() time.Time { return time.Now().UTC() },
		logger:     logger,
	}
	for _, opt := range opts {
		opt(u)
	}
	return u, nil
}

func (u *RetryPaymentIntentUsecase) Execute(ctx context.Context, input RetryPaymentIntentInput) (RetryPaymentIntentOutput, error) {
	if err := ctx.Err(); err != nil {
		return RetryPaymentIntentOutput{}, err
	}
	input = normalizeRetryPaymentIntentInput(input)
	if err := validateRetryPaymentIntentInput(input); err != nil {
		return RetryPaymentIntentOutput{}, err
	}
	paymentID, err := u.ids.NewPaymentID()
	if err != nil {
		return RetryPaymentIntentOutput{}, err
	}
	attemptID, err := u.ids.NewAttemptID()
	if err != nil {
		return RetryPaymentIntentOutput{}, err
	}
	reserved, err := u.repository.ReserveRetryPayment(ctx, domain.ReserveRetryPaymentInput{
		FailedPaymentID: input.FailedPaymentID,
		BuyerUserID:     input.BuyerUserID,
		RetryRequestKey: input.RetryRequestKey,
		PaymentID:       paymentID,
		AttemptID:       attemptID,
		Policy:          u.policy,
		Now:             u.clock(),
	})
	if err != nil {
		return RetryPaymentIntentOutput{}, err
	}

	gateway, err := u.providers.Get(reserved.Payment.Provider)
	if err != nil {
		u.failBeforeProviderCall(ctx, reserved, "provider_unavailable")
		return RetryPaymentIntentOutput{}, err
	}
	if reserved.Replayed && reserved.Payment.ProviderIntentID != "" {
		return u.existingRetryOutput(ctx, reserved.Payment, gateway)
	}
	if reserved.Replayed && reserved.Payment.Status == domain.PaymentStatusFailed {
		return RetryPaymentIntentOutput{}, ErrPaymentIntentPreviouslyFailed
	}

	request := u.providerRequest(reserved.Payment, input.RequestID)
	if err := provider.ValidateCreateIntentRequest(request); err != nil {
		u.failBeforeProviderCall(ctx, reserved, "provider_request_invalid")
		return RetryPaymentIntentOutput{}, err
	}
	if validator, ok := gateway.(PaymentIntentRequestValidator); ok {
		if err := validator.ValidateCreateIntent(request); err != nil {
			u.failBeforeProviderCall(ctx, reserved, "provider_request_invalid")
			return RetryPaymentIntentOutput{}, err
		}
	}

	response, err := gateway.CreateIntent(ctx, request)
	if err != nil {
		if isDefinitiveRetryProviderFailure(err) {
			u.recordProviderFailure(ctx, reserved, err)
		} else {
			u.logger.WarnContext(ctx, "payment.retry.provider_outcome_pending",
				slog.String("payment_id", reserved.Payment.PaymentID),
				slog.String("order_id", reserved.Payment.OrderID),
				slog.String("provider", reserved.Payment.Provider),
			)
		}
		return RetryPaymentIntentOutput{}, err
	}
	response = response.Normalized()
	if response.Provider != reserved.Payment.Provider {
		err := fmt.Errorf("%w: provider response identifies %q, want %q", provider.ErrInvalidProviderRequest, response.Provider, reserved.Payment.Provider)
		u.logAmbiguousResponse(ctx, reserved.Payment, err)
		return RetryPaymentIntentOutput{}, err
	}
	if err := provider.ValidateCreateIntentResponse(response); err != nil {
		u.logAmbiguousResponse(ctx, reserved.Payment, err)
		return RetryPaymentIntentOutput{}, err
	}

	payment := reserved.Payment
	payment.ProviderIntentID = response.ProviderIntentID
	payment.ProviderPaymentID = response.ProviderPaymentID
	status, err := intentPaymentStatus(response.Status)
	if err != nil {
		u.logAmbiguousResponse(ctx, payment, err)
		return RetryPaymentIntentOutput{}, err
	}
	if _, err := payment.ApplyTransition(status, u.clock()); err != nil {
		u.logAmbiguousResponse(ctx, payment, err)
		return RetryPaymentIntentOutput{}, err
	}
	attempt := reserved.Attempt
	attempt.ProviderAttemptID = response.ProviderIntentID
	attempt.RawProviderResponse = append(attempt.RawProviderResponse[:0], response.RawProviderResponse...)
	attempt.Status = domain.PaymentAttemptStatusSucceeded
	if status == domain.PaymentStatusFailed {
		attempt.Status = domain.PaymentAttemptStatusFailed
	}
	if err := u.repository.UpdatePaymentIntent(ctx, payment, attempt); err != nil {
		if errors.Is(err, domain.ErrPaymentRecordNotFound) {
			current, loadErr := u.repository.GetPaymentByID(ctx, payment.PaymentID)
			if loadErr == nil && current.ProviderIntentID != "" {
				return u.existingRetryOutput(ctx, current, gateway)
			}
		}
		return RetryPaymentIntentOutput{}, err
	}
	out := u.outputFromProviderResponse(payment, response, reserved.Replayed)
	u.logger.InfoContext(ctx, "payment.retry.created",
		slog.String("payment_id", payment.PaymentID),
		slog.String("retry_of_payment_id", payment.RetryOfPaymentID),
		slog.String("root_payment_id", payment.RootPaymentID),
		slog.String("order_id", payment.OrderID),
		slog.Uint64("attempt_no", uint64(payment.AttemptNo)),
		slog.String("provider", payment.Provider),
		slog.Bool("idempotency_replayed", reserved.Replayed),
	)
	return out, nil
}

func normalizeRetryPaymentIntentInput(input RetryPaymentIntentInput) RetryPaymentIntentInput {
	input.FailedPaymentID = strings.TrimSpace(input.FailedPaymentID)
	input.BuyerUserID = strings.TrimSpace(input.BuyerUserID)
	input.RetryRequestKey = strings.TrimSpace(input.RetryRequestKey)
	input.RequestID = strings.TrimSpace(input.RequestID)
	return input
}

func validateRetryPaymentIntentInput(input RetryPaymentIntentInput) error {
	switch {
	case input.FailedPaymentID == "":
		return fmt.Errorf("%w: payment_id is required", provider.ErrInvalidProviderRequest)
	case len(input.FailedPaymentID) > 64:
		return fmt.Errorf("%w: payment_id exceeds 64 characters", provider.ErrInvalidProviderRequest)
	case input.BuyerUserID == "":
		return fmt.Errorf("%w: buyer user_id is required", provider.ErrInvalidProviderRequest)
	case len(input.BuyerUserID) > 64:
		return fmt.Errorf("%w: buyer user_id exceeds 64 characters", provider.ErrInvalidProviderRequest)
	}
	return provider.ValidateIdempotencyKey(input.RetryRequestKey)
}

func (u *RetryPaymentIntentUsecase) providerRequest(payment domain.Payment, requestID string) provider.CreateIntentRequest {
	metadata := provider.BuildSafeMetadata(payment.PaymentID, payment.OrderID, payment.UserID, requestID, payment.IdempotencyKey)
	metadata["retry_of_payment_id"] = payment.RetryOfPaymentID
	metadata["root_payment_id"] = payment.RootPaymentID
	metadata["attempt_no"] = strconv.FormatUint(uint64(payment.AttemptNo), 10)
	return provider.CreateIntentRequest{
		PaymentID: payment.PaymentID,
		OrderID:   payment.OrderID,
		UserID:    payment.UserID,
		Amount: provider.Money{
			AmountMinor: payment.Amount.Amount,
			Currency:    payment.Amount.Currency,
		},
		IdempotencyKey: payment.IdempotencyKey,
		CaptureMode:    u.config.CaptureMode,
		Metadata:       metadata,
	}
}

func (u *RetryPaymentIntentUsecase) existingRetryOutput(ctx context.Context, payment domain.Payment, gateway provider.Provider) (RetryPaymentIntentOutput, error) {
	if payment.ProviderIntentID == "" {
		return RetryPaymentIntentOutput{}, ErrPaymentIntentProcessing
	}
	out := u.outputFromPayment(payment, true)
	if payment.Provider == provider.ProviderNameStripeLike {
		reader, ok := gateway.(PaymentIntentClientPayloadReader)
		if !ok {
			return RetryPaymentIntentOutput{}, ErrPaymentClientPayloadUnavailable
		}
		response, err := reader.RetrieveIntent(ctx, payment.ProviderIntentID)
		if err != nil {
			return RetryPaymentIntentOutput{}, err
		}
		response = response.Normalized()
		if err := provider.ValidateCreateIntentResponse(response); err != nil ||
			response.Provider != payment.Provider ||
			response.ProviderIntentID != payment.ProviderIntentID ||
			response.ClientSecret == "" {
			return RetryPaymentIntentOutput{}, ErrPaymentClientPayloadUnavailable
		}
		out.ClientPayload.ClientSecret = response.ClientSecret
		out.ClientPayload.RedirectURL = response.RedirectURL
		out.ClientPayload.Extra = copySafeFrontendPayload(response.FrontendPayload)
	}
	u.logger.InfoContext(ctx, "payment.retry.idempotency_hit",
		slog.String("payment_id", payment.PaymentID),
		slog.String("retry_of_payment_id", payment.RetryOfPaymentID),
		slog.String("order_id", payment.OrderID),
	)
	return out, nil
}

func (u *RetryPaymentIntentUsecase) outputFromProviderResponse(payment domain.Payment, response provider.CreateIntentResponse, replayed bool) RetryPaymentIntentOutput {
	out := u.outputFromPayment(payment, replayed)
	out.ClientPayload.ClientSecret = response.ClientSecret
	out.ClientPayload.RedirectURL = response.RedirectURL
	out.ClientPayload.Extra = copySafeFrontendPayload(response.FrontendPayload)
	if orderID := strings.TrimSpace(response.FrontendPayload["provider_order_id"]); orderID != "" {
		out.ClientPayload.ProviderOrderID = orderID
	}
	if sessionID := strings.TrimSpace(response.FrontendPayload["checkout_session_id"]); sessionID != "" {
		out.ClientPayload.CheckoutSessionID = sessionID
	}
	return out
}

func (u *RetryPaymentIntentUsecase) outputFromPayment(payment domain.Payment, replayed bool) RetryPaymentIntentOutput {
	clientPayload := PaymentClientPayload{}
	if cfg, ok := u.config.ProviderConfig(payment.Provider); ok {
		clientPayload.PublicKey = cfg.PublicKey
	}
	if payment.Provider == provider.ProviderNameRazorpayLike {
		clientPayload.ProviderOrderID = payment.ProviderIntentID
	}
	return RetryPaymentIntentOutput{
		PaymentID:        payment.PaymentID,
		OrderID:          payment.OrderID,
		RetryOfPaymentID: payment.RetryOfPaymentID,
		RootPaymentID:    payment.RootPaymentID,
		AttemptNo:        payment.AttemptNo,
		Provider:         payment.Provider,
		ProviderIntentID: payment.ProviderIntentID,
		Status:           payment.Status,
		Amount:           payment.Amount.Amount,
		Currency:         payment.Amount.Currency,
		ClientPayload:    clientPayload,
		Replayed:         replayed,
	}
}

func (u *RetryPaymentIntentUsecase) failBeforeProviderCall(ctx context.Context, reserved domain.ReservedRetryPayment, code string) {
	if err := u.repository.FailPaymentIntent(
		ctx,
		reserved.Payment.PaymentID,
		reserved.Attempt.AttemptID,
		code,
		"payment provider could not create a retry intent",
		u.clock(),
	); err != nil {
		u.logger.ErrorContext(ctx, "payment.retry.persist_failure_failed",
			slog.String("payment_id", reserved.Payment.PaymentID),
			slog.String("error", err.Error()),
		)
	}
}

func (u *RetryPaymentIntentUsecase) recordProviderFailure(ctx context.Context, reserved domain.ReservedRetryPayment, err error) {
	code := "provider_error"
	if providerCode, ok := provider.CodeOf(err); ok {
		code = string(providerCode)
	}
	u.failBeforeProviderCall(ctx, reserved, code)
	u.logger.WarnContext(ctx, "payment.retry.provider_failed",
		slog.String("payment_id", reserved.Payment.PaymentID),
		slog.String("order_id", reserved.Payment.OrderID),
		slog.String("provider", reserved.Payment.Provider),
		slog.String("failure_code", code),
	)
}

func (u *RetryPaymentIntentUsecase) logAmbiguousResponse(ctx context.Context, payment domain.Payment, err error) {
	u.logger.ErrorContext(ctx, "payment.retry.provider_response_unresolved",
		slog.String("payment_id", payment.PaymentID),
		slog.String("order_id", payment.OrderID),
		slog.String("provider", payment.Provider),
		slog.String("error", err.Error()),
	)
}

func isDefinitiveRetryProviderFailure(err error) bool {
	code, ok := provider.CodeOf(err)
	if !ok {
		return false
	}
	switch code {
	case provider.ErrorCodeValidation, provider.ErrorCodeAuthentication, provider.ErrorCodeDeclined:
		return true
	default:
		return false
	}
}
