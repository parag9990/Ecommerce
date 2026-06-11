package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider"
)

var (
	ErrPaymentIntentProcessing         = errors.New("payment intent is already processing")
	ErrPaymentIntentConflict           = errors.New("idempotency key is already used for another payment request")
	ErrPaymentIntentPreviouslyFailed   = errors.New("payment intent previously failed")
	ErrPaymentClientPayloadUnavailable = errors.New("payment intent client payload is unavailable")
	ErrPaymentProviderUnavailable      = errors.New("payment provider is not configured")
)

var e164PhonePattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

type PaymentIntentRepository interface {
	FindPaymentByProviderIdempotencyKey(ctx context.Context, providerName string, idempotencyKey string) (domain.Payment, error)
	CreatePaymentWithAttempt(ctx context.Context, payment domain.Payment, attempt domain.PaymentAttempt) error
	UpdatePaymentIntent(ctx context.Context, payment domain.Payment, attempt domain.PaymentAttempt) error
	FailPaymentIntent(ctx context.Context, paymentID string, attemptID string, failureCode string, failureMessage string, at time.Time) error
}

type PaymentProviderRegistry interface {
	Get(name string) (provider.Provider, error)
}

type PaymentIntentClientPayloadReader interface {
	RetrieveIntent(ctx context.Context, providerIntentID string) (provider.CreateIntentResponse, error)
}

type PaymentIntentRequestValidator interface {
	ValidateCreateIntent(req provider.CreateIntentRequest) error
}

type PaymentIntentIDGenerator interface {
	NewPaymentID() (string, error)
	NewAttemptID() (string, error)
}

type CreatePaymentIntentUsecase struct {
	repository PaymentIntentRepository
	providers  PaymentProviderRegistry
	config     provider.Config
	ids        PaymentIntentIDGenerator
	clock      func() time.Time
	logger     *slog.Logger
}

type CreatePaymentIntentOption func(*CreatePaymentIntentUsecase)

func WithPaymentIntentIDGenerator(ids PaymentIntentIDGenerator) CreatePaymentIntentOption {
	return func(u *CreatePaymentIntentUsecase) {
		if ids != nil {
			u.ids = ids
		}
	}
}

func WithPaymentIntentClock(clock func() time.Time) CreatePaymentIntentOption {
	return func(u *CreatePaymentIntentUsecase) {
		if clock != nil {
			u.clock = clock
		}
	}
}

type PaymentIntentCustomer struct {
	Name  string
	Email string
	Phone string
}

type CreatePaymentIntentInput struct {
	OrderID        string
	UserID         string
	Amount         int64
	Currency       string
	Customer       PaymentIntentCustomer
	IdempotencyKey string
	Provider       string
	Metadata       map[string]string
	RequestID      string
}

type PaymentClientPayload struct {
	PublicKey         string
	ClientSecret      string
	ProviderOrderID   string
	CheckoutSessionID string
	RedirectURL       string
	Extra             map[string]string
}

type CreatePaymentIntentOutput struct {
	PaymentID        string
	OrderID          string
	Provider         string
	ProviderIntentID string
	Status           domain.PaymentStatus
	Amount           int64
	Currency         string
	ClientPayload    PaymentClientPayload
	Replayed         bool
}

func NewCreatePaymentIntentUsecase(
	repository PaymentIntentRepository,
	providers PaymentProviderRegistry,
	cfg provider.Config,
	logger *slog.Logger,
	opts ...CreatePaymentIntentOption,
) (*CreatePaymentIntentUsecase, error) {
	if repository == nil {
		return nil, errors.New("payment intent repository is required")
	}
	if providers == nil {
		return nil, errors.New("payment provider registry is required")
	}
	cfg = cfg.Normalized()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	usecase := &CreatePaymentIntentUsecase{
		repository: repository,
		providers:  providers,
		config:     cfg,
		ids:        cryptoPaymentIntentIDGenerator{},
		clock:      func() time.Time { return time.Now().UTC() },
		logger:     logger,
	}
	for _, opt := range opts {
		opt(usecase)
	}
	return usecase, nil
}

func (u *CreatePaymentIntentUsecase) Execute(ctx context.Context, input CreatePaymentIntentInput) (CreatePaymentIntentOutput, error) {
	if err := ctx.Err(); err != nil {
		return CreatePaymentIntentOutput{}, err
	}

	input = normalizeCreatePaymentIntentInput(input)
	providerName, gateway, err := u.resolveProvider(input.Provider)
	if err != nil {
		return CreatePaymentIntentOutput{}, err
	}
	if err := u.validateInput(input); err != nil {
		return CreatePaymentIntentOutput{}, err
	}

	existing, err := u.repository.FindPaymentByProviderIdempotencyKey(ctx, providerName, input.IdempotencyKey)
	if err == nil {
		return u.existingIntentOutput(ctx, existing, input, gateway)
	}
	if !errors.Is(err, domain.ErrPaymentRecordNotFound) {
		return CreatePaymentIntentOutput{}, err
	}

	payment, attempt, err := u.newRecords(input, providerName)
	if err != nil {
		return CreatePaymentIntentOutput{}, err
	}
	providerReq := u.providerRequest(payment, input)
	if err := provider.ValidateCreateIntentRequest(providerReq); err != nil {
		return CreatePaymentIntentOutput{}, err
	}
	if validator, ok := gateway.(PaymentIntentRequestValidator); ok {
		if err := validator.ValidateCreateIntent(providerReq); err != nil {
			return CreatePaymentIntentOutput{}, err
		}
	}
	if err := u.repository.CreatePaymentWithAttempt(ctx, payment, attempt); err != nil {
		if errors.Is(err, domain.ErrDuplicatePaymentRecord) {
			existing, loadErr := u.repository.FindPaymentByProviderIdempotencyKey(ctx, providerName, input.IdempotencyKey)
			if loadErr != nil {
				return CreatePaymentIntentOutput{}, loadErr
			}
			return u.existingIntentOutput(ctx, existing, input, gateway)
		}
		return CreatePaymentIntentOutput{}, err
	}

	res, err := gateway.CreateIntent(ctx, providerReq)
	if err != nil {
		u.recordProviderFailure(ctx, payment, attempt, err)
		return CreatePaymentIntentOutput{}, err
	}
	res = res.Normalized()
	if res.Provider != providerName {
		err := fmt.Errorf("%w: provider response identifies %q, want %q", provider.ErrInvalidProviderRequest, res.Provider, providerName)
		u.recordProviderFailure(ctx, payment, attempt, err)
		return CreatePaymentIntentOutput{}, err
	}
	if err := provider.ValidateCreateIntentResponse(res); err != nil {
		u.recordProviderFailure(ctx, payment, attempt, err)
		return CreatePaymentIntentOutput{}, err
	}

	payment.ProviderIntentID = res.ProviderIntentID
	payment.ProviderPaymentID = res.ProviderPaymentID
	status, err := intentPaymentStatus(res.Status)
	if err != nil {
		u.recordProviderFailure(ctx, payment, attempt, err)
		return CreatePaymentIntentOutput{}, err
	}
	if _, err := payment.ApplyTransition(status, u.clock()); err != nil {
		u.recordProviderFailure(ctx, payment, attempt, err)
		return CreatePaymentIntentOutput{}, err
	}
	attempt.ProviderAttemptID = res.ProviderIntentID
	attempt.RawProviderResponse = append(attempt.RawProviderResponse[:0], res.RawProviderResponse...)
	attempt.Status = domain.PaymentAttemptStatusSucceeded
	if status == domain.PaymentStatusFailed {
		attempt.Status = domain.PaymentAttemptStatusFailed
	}
	if err := u.repository.UpdatePaymentIntent(ctx, payment, attempt); err != nil {
		u.logger.ErrorContext(ctx, "payment.intent.persist_result_failed",
			slog.String("payment_id", payment.PaymentID),
			slog.String("provider", providerName),
			slog.String("provider_intent_id", res.ProviderIntentID),
			slog.String("error", err.Error()),
		)
		return CreatePaymentIntentOutput{}, err
	}

	out := u.outputFromProviderResponse(payment, res)
	u.logger.InfoContext(ctx, "payment.intent.created",
		slog.String("payment_id", payment.PaymentID),
		slog.String("order_id", payment.OrderID),
		slog.String("provider", providerName),
		slog.String("provider_intent_id", payment.ProviderIntentID),
		slog.String("status", string(payment.Status)),
		slog.String("currency", payment.Amount.Currency),
		slog.Int64("amount_minor", payment.Amount.Amount),
	)
	return out, nil
}

func (u *CreatePaymentIntentUsecase) resolveProvider(requested string) (string, provider.Provider, error) {
	name := provider.NormalizeProviderName(requested)
	if name == "" {
		name = u.config.DefaultProvider
	}
	if name == "" {
		return "", nil, ErrPaymentProviderUnavailable
	}
	gateway, err := u.providers.Get(name)
	if err != nil {
		return "", nil, err
	}
	return name, gateway, nil
}

func (u *CreatePaymentIntentUsecase) validateInput(input CreatePaymentIntentInput) error {
	switch {
	case input.OrderID == "":
		return fmt.Errorf("%w: order_id is required", provider.ErrInvalidProviderRequest)
	case len(input.OrderID) > 64:
		return fmt.Errorf("%w: order_id exceeds 64 characters", provider.ErrInvalidProviderRequest)
	case input.UserID == "":
		return fmt.Errorf("%w: user_id is required", provider.ErrInvalidProviderRequest)
	case len(input.UserID) > 64:
		return fmt.Errorf("%w: user_id exceeds 64 characters", provider.ErrInvalidProviderRequest)
	case input.Amount <= 0:
		return fmt.Errorf("%w: amount must be greater than zero", provider.ErrInvalidProviderRequest)
	}
	if err := u.config.ValidateCurrency(input.Currency); err != nil {
		return err
	}
	if err := provider.ValidateIdempotencyKey(input.IdempotencyKey); err != nil {
		return err
	}
	if err := provider.ValidateMetadata(input.Metadata); err != nil {
		return err
	}
	if len(input.Customer.Name) > 256 {
		return fmt.Errorf("%w: customer.name exceeds 256 characters", provider.ErrInvalidProviderRequest)
	}
	if input.Customer.Email != "" {
		address, err := mail.ParseAddress(input.Customer.Email)
		if err != nil || !strings.EqualFold(address.Address, input.Customer.Email) {
			return fmt.Errorf("%w: customer.email is invalid", provider.ErrInvalidProviderRequest)
		}
	}
	if input.Customer.Phone != "" && !e164PhonePattern.MatchString(input.Customer.Phone) {
		return fmt.Errorf("%w: customer.phone must be in E.164 format", provider.ErrInvalidProviderRequest)
	}
	return nil
}

func (u *CreatePaymentIntentUsecase) newRecords(input CreatePaymentIntentInput, providerName string) (domain.Payment, domain.PaymentAttempt, error) {
	paymentID, err := u.ids.NewPaymentID()
	if err != nil {
		return domain.Payment{}, domain.PaymentAttempt{}, err
	}
	attemptID, err := u.ids.NewAttemptID()
	if err != nil {
		return domain.Payment{}, domain.PaymentAttempt{}, err
	}
	now := u.clock().UTC()
	payment, err := domain.NewPayment(domain.NewPaymentInput{
		PaymentID:      paymentID,
		OrderID:        input.OrderID,
		UserID:         input.UserID,
		Provider:       providerName,
		Amount:         domain.Money{Amount: input.Amount, Currency: input.Currency},
		IdempotencyKey: input.IdempotencyKey,
		Now:            now,
	})
	if err != nil {
		return domain.Payment{}, domain.PaymentAttempt{}, err
	}
	attempt := domain.PaymentAttempt{
		AttemptID: attemptID,
		PaymentID: payment.PaymentID,
		Status:    domain.PaymentAttemptStatusInitiated,
		CreatedAt: now,
	}
	if err := attempt.Validate(); err != nil {
		return domain.Payment{}, domain.PaymentAttempt{}, err
	}
	return payment, attempt, nil
}

func (u *CreatePaymentIntentUsecase) providerRequest(payment domain.Payment, input CreatePaymentIntentInput) provider.CreateIntentRequest {
	metadata := make(map[string]string, len(input.Metadata)+6)
	for key, value := range input.Metadata {
		metadata[key] = value
	}
	for key, value := range provider.BuildSafeMetadata(payment.PaymentID, payment.OrderID, payment.UserID, input.RequestID, payment.IdempotencyKey) {
		metadata[key] = value
	}
	return provider.CreateIntentRequest{
		PaymentID: payment.PaymentID,
		OrderID:   payment.OrderID,
		UserID:    payment.UserID,
		Amount: provider.Money{
			AmountMinor: payment.Amount.Amount,
			Currency:    payment.Amount.Currency,
		},
		Customer: provider.Customer{
			UserID: payment.UserID,
			Name:   input.Customer.Name,
			Email:  input.Customer.Email,
			Phone:  input.Customer.Phone,
		},
		IdempotencyKey: payment.IdempotencyKey,
		CaptureMode:    u.config.CaptureMode,
		Metadata:       metadata,
	}
}

func (u *CreatePaymentIntentUsecase) existingIntentOutput(ctx context.Context, payment domain.Payment, input CreatePaymentIntentInput, gateway provider.Provider) (CreatePaymentIntentOutput, error) {
	if payment.OrderID != input.OrderID ||
		payment.UserID != input.UserID ||
		payment.Amount.Amount != input.Amount ||
		payment.Amount.Currency != input.Currency {
		return CreatePaymentIntentOutput{}, ErrPaymentIntentConflict
	}
	if payment.ProviderIntentID == "" {
		if payment.Status == domain.PaymentStatusFailed {
			return CreatePaymentIntentOutput{}, ErrPaymentIntentPreviouslyFailed
		}
		return CreatePaymentIntentOutput{}, ErrPaymentIntentProcessing
	}
	u.logger.InfoContext(ctx, "payment.intent.idempotency_hit",
		slog.String("payment_id", payment.PaymentID),
		slog.String("order_id", payment.OrderID),
		slog.String("provider", payment.Provider),
		slog.String("status", string(payment.Status)),
	)
	out := u.outputFromPayment(payment)
	if payment.Provider == provider.ProviderNameStripeLike {
		reader, ok := gateway.(PaymentIntentClientPayloadReader)
		if !ok {
			return CreatePaymentIntentOutput{}, ErrPaymentClientPayloadUnavailable
		}
		res, err := reader.RetrieveIntent(ctx, payment.ProviderIntentID)
		if err != nil {
			return CreatePaymentIntentOutput{}, err
		}
		res = res.Normalized()
		if err := provider.ValidateCreateIntentResponse(res); err != nil {
			return CreatePaymentIntentOutput{}, ErrPaymentClientPayloadUnavailable
		}
		if res.Provider != payment.Provider || res.ProviderIntentID != payment.ProviderIntentID || res.ClientSecret == "" {
			return CreatePaymentIntentOutput{}, ErrPaymentClientPayloadUnavailable
		}
		out.ClientPayload.ClientSecret = res.ClientSecret
		out.ClientPayload.RedirectURL = res.RedirectURL
		out.ClientPayload.Extra = copySafeFrontendPayload(res.FrontendPayload)
	}
	out.Replayed = true
	return out, nil
}

func (u *CreatePaymentIntentUsecase) outputFromProviderResponse(payment domain.Payment, res provider.CreateIntentResponse) CreatePaymentIntentOutput {
	out := u.outputFromPayment(payment)
	out.ClientPayload.ClientSecret = res.ClientSecret
	out.ClientPayload.RedirectURL = res.RedirectURL
	out.ClientPayload.Extra = copySafeFrontendPayload(res.FrontendPayload)
	if orderID := strings.TrimSpace(res.FrontendPayload["provider_order_id"]); orderID != "" {
		out.ClientPayload.ProviderOrderID = orderID
	}
	if sessionID := strings.TrimSpace(res.FrontendPayload["checkout_session_id"]); sessionID != "" {
		out.ClientPayload.CheckoutSessionID = sessionID
	}
	return out
}

func (u *CreatePaymentIntentUsecase) outputFromPayment(payment domain.Payment) CreatePaymentIntentOutput {
	clientPayload := PaymentClientPayload{PublicKey: u.publicKey(payment.Provider)}
	if payment.Provider == provider.ProviderNameRazorpayLike {
		clientPayload.ProviderOrderID = payment.ProviderIntentID
	}
	return CreatePaymentIntentOutput{
		PaymentID:        payment.PaymentID,
		OrderID:          payment.OrderID,
		Provider:         payment.Provider,
		ProviderIntentID: payment.ProviderIntentID,
		Status:           payment.Status,
		Amount:           payment.Amount.Amount,
		Currency:         payment.Amount.Currency,
		ClientPayload:    clientPayload,
	}
}

func (u *CreatePaymentIntentUsecase) publicKey(providerName string) string {
	cfg, ok := u.config.ProviderConfig(providerName)
	if !ok {
		return ""
	}
	return cfg.PublicKey
}

func (u *CreatePaymentIntentUsecase) recordProviderFailure(ctx context.Context, payment domain.Payment, attempt domain.PaymentAttempt, err error) {
	code := "provider_error"
	if providerCode, ok := provider.CodeOf(err); ok {
		code = string(providerCode)
	}
	if persistenceErr := u.repository.FailPaymentIntent(
		ctx,
		payment.PaymentID,
		attempt.AttemptID,
		code,
		"payment provider could not create an intent",
		u.clock(),
	); persistenceErr != nil {
		u.logger.ErrorContext(ctx, "payment.intent.persist_failure_failed",
			slog.String("payment_id", payment.PaymentID),
			slog.String("provider", payment.Provider),
			slog.String("error", persistenceErr.Error()),
		)
	}
	u.logger.WarnContext(ctx, "payment.intent.provider_failed",
		slog.String("payment_id", payment.PaymentID),
		slog.String("order_id", payment.OrderID),
		slog.String("provider", payment.Provider),
		slog.String("failure_code", code),
	)
}

func normalizeCreatePaymentIntentInput(input CreatePaymentIntentInput) CreatePaymentIntentInput {
	input.OrderID = strings.TrimSpace(input.OrderID)
	input.UserID = strings.TrimSpace(input.UserID)
	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	input.Provider = provider.NormalizeProviderName(input.Provider)
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	input.RequestID = strings.TrimSpace(input.RequestID)
	input.Customer.Name = strings.TrimSpace(input.Customer.Name)
	input.Customer.Email = strings.TrimSpace(input.Customer.Email)
	input.Customer.Phone = strings.TrimSpace(input.Customer.Phone)
	input.Metadata = provider.NormalizeMetadata(input.Metadata)
	return input
}

func intentPaymentStatus(status provider.IntentStatus) (domain.PaymentStatus, error) {
	switch status.Normalized() {
	case provider.IntentStatusInitiated:
		return domain.PaymentStatusInitiated, nil
	case provider.IntentStatusRequiresAction:
		return domain.PaymentStatusRequiresAction, nil
	case provider.IntentStatusAuthorized:
		return domain.PaymentStatusAuthorized, nil
	case provider.IntentStatusCaptured:
		return domain.PaymentStatusCaptured, nil
	case provider.IntentStatusFailed:
		return domain.PaymentStatusFailed, nil
	default:
		return domain.PaymentStatusUnspecified, fmt.Errorf("%w: unknown intent status %q", provider.ErrInvalidProviderRequest, status)
	}
}

func copySafeFrontendPayload(payload map[string]string) map[string]string {
	if len(payload) == 0 {
		return nil
	}
	excluded := map[string]bool{
		"client_secret":       true,
		"public_key":          true,
		"provider_order_id":   true,
		"checkout_session_id": true,
	}
	extra := make(map[string]string, len(payload))
	for key, value := range payload {
		if !excluded[key] {
			extra[key] = value
		}
	}
	if len(extra) == 0 {
		return nil
	}
	return extra
}

type cryptoPaymentIntentIDGenerator struct{}

func (cryptoPaymentIntentIDGenerator) NewPaymentID() (string, error) {
	return randomPrefixedID("pay_")
}

func (cryptoPaymentIntentIDGenerator) NewAttemptID() (string, error) {
	return randomPrefixedID("pat_")
}

func randomPrefixedID(prefix string) (string, error) {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate payment identifier: %w", err)
	}
	return prefix + hex.EncodeToString(random), nil
}
