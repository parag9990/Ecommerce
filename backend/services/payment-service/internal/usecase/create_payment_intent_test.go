package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider/providertest"
)

func TestCreatePaymentIntentPersistsProviderResultAndReturnsSafeClientPayload(t *testing.T) {
	repo := newFakePaymentIntentRepository()
	gateway := &providertest.FakeProvider{
		NameValue: provider.ProviderNameStripeLike,
		CreateIntentResponse: provider.CreateIntentResponse{
			Provider:         provider.ProviderNameStripeLike,
			ProviderIntentID: "pi_123",
			Status:           provider.IntentStatusRequiresAction,
			ClientSecret:     "client_secret_browser_only",
			FrontendPayload:  map[string]string{"checkout_session_id": "session_123", "theme": "blue"},
			RawProviderResponse: json.RawMessage(
				`{"id":"pi_123","status":"requires_action"}`,
			),
		},
	}
	uc := newPaymentIntentUsecaseForTest(t, repo, gateway, &staticPaymentIntentIDs{})

	out, err := uc.Execute(context.Background(), validPaymentIntentInput())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if out.PaymentID != "pay_test" || out.ProviderIntentID != "pi_123" || out.Status != domain.PaymentStatusRequiresAction {
		t.Fatalf("output = %+v, want created requires_action intent", out)
	}
	if out.ClientPayload.ClientSecret != "client_secret_browser_only" || out.ClientPayload.PublicKey != "pk_test" {
		t.Fatalf("client payload = %+v, want frontend client secret and public key", out.ClientPayload)
	}
	if gateway.CreateIntentCalls != 1 {
		t.Fatalf("CreateIntentCalls = %d, want 1", gateway.CreateIntentCalls)
	}
	if gateway.LastCreateIntent.IdempotencyKey != "payment_intent:ord_123:attempt_1" {
		t.Fatalf("provider idempotency key = %q", gateway.LastCreateIntent.IdempotencyKey)
	}
	if gateway.LastCreateIntent.Metadata["payment_id"] != "pay_test" || gateway.LastCreateIntent.Metadata["order_id"] != "ord_123" {
		t.Fatalf("provider metadata = %#v, want local references", gateway.LastCreateIntent.Metadata)
	}
	persisted := repo.payments["stripe_like/payment_intent:ord_123:attempt_1"]
	if persisted.ProviderIntentID != "pi_123" || persisted.Status != domain.PaymentStatusRequiresAction {
		t.Fatalf("persisted payment = %+v, want provider result", persisted)
	}
	if strings.Contains(string(repo.updatedAttempt.RawProviderResponse), "client_secret") {
		t.Fatalf("stored raw response contains browser secret: %s", repo.updatedAttempt.RawProviderResponse)
	}
}

func TestCreatePaymentIntentReturnsExistingIntentWithoutCallingProvider(t *testing.T) {
	repo := newFakePaymentIntentRepository()
	repo.payments["stripe_like/payment_intent:ord_123:attempt_1"] = existingPayment(t, "pi_existing", domain.PaymentStatusRequiresAction)
	gateway := &retrievingFakeProvider{
		FakeProvider: providertest.FakeProvider{NameValue: provider.ProviderNameStripeLike},
		response: provider.CreateIntentResponse{
			Provider:         provider.ProviderNameStripeLike,
			ProviderIntentID: "pi_existing",
			Status:           provider.IntentStatusRequiresAction,
			ClientSecret:     "retrieved_client_secret",
			RawProviderResponse: json.RawMessage(
				`{"id":"pi_existing","status":"requires_action"}`,
			),
		},
	}
	uc := newPaymentIntentUsecaseForTest(t, repo, gateway, &staticPaymentIntentIDs{})

	out, err := uc.Execute(context.Background(), validPaymentIntentInput())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !out.Replayed || out.ProviderIntentID != "pi_existing" {
		t.Fatalf("output = %+v, want idempotent replay", out)
	}
	if gateway.CreateIntentCalls != 0 || gateway.retrieveCalls != 1 {
		t.Fatalf("CreateIntentCalls/retrieveCalls = %d/%d, want 0/1", gateway.CreateIntentCalls, gateway.retrieveCalls)
	}
	if out.ClientPayload.ClientSecret != "retrieved_client_secret" {
		t.Fatalf("client secret = %q, want securely retrieved client payload", out.ClientPayload.ClientSecret)
	}
}

func TestCreatePaymentIntentRejectsIdempotencyConflict(t *testing.T) {
	repo := newFakePaymentIntentRepository()
	payment := existingPayment(t, "pi_existing", domain.PaymentStatusRequiresAction)
	payment.Amount.Amount = 2000
	repo.payments["stripe_like/payment_intent:ord_123:attempt_1"] = payment
	gateway := &providertest.FakeProvider{NameValue: provider.ProviderNameStripeLike}
	uc := newPaymentIntentUsecaseForTest(t, repo, gateway, &staticPaymentIntentIDs{})

	_, err := uc.Execute(context.Background(), validPaymentIntentInput())
	if !errors.Is(err, ErrPaymentIntentConflict) {
		t.Fatalf("Execute() error = %v, want idempotency conflict", err)
	}
	if gateway.CreateIntentCalls != 0 {
		t.Fatalf("CreateIntentCalls = %d, want 0", gateway.CreateIntentCalls)
	}
}

func TestCreatePaymentIntentRejectsInvalidInputBeforePersistence(t *testing.T) {
	repo := newFakePaymentIntentRepository()
	gateway := &providertest.FakeProvider{NameValue: provider.ProviderNameStripeLike}
	uc := newPaymentIntentUsecaseForTest(t, repo, gateway, &staticPaymentIntentIDs{})
	input := validPaymentIntentInput()
	input.Customer.Phone = "99999"

	_, err := uc.Execute(context.Background(), input)
	if !errors.Is(err, provider.ErrInvalidProviderRequest) {
		t.Fatalf("Execute() error = %v, want invalid provider request", err)
	}
	if repo.createCalls != 0 || gateway.CreateIntentCalls != 0 {
		t.Fatalf("create calls = %d/provider calls = %d, want no side effects", repo.createCalls, gateway.CreateIntentCalls)
	}
}

func TestCreatePaymentIntentRejectsProviderSpecificLimitsBeforePersistence(t *testing.T) {
	repo := newFakePaymentIntentRepository()
	gateway, err := provider.NewRazorpayLikeProvider(provider.ProviderConfig{
		PublicKey:     "rzp_key",
		SecretKey:     "rzp_secret",
		WebhookSecret: "rzp_whsec",
	}, nil)
	if err != nil {
		t.Fatalf("NewRazorpayLikeProvider() error = %v", err)
	}
	registry, err := provider.NewRegistry(gateway)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	cfg := provider.Config{
		DefaultProvider:   provider.ProviderNameRazorpayLike,
		AllowedProviders:  []string{provider.ProviderNameRazorpayLike},
		AllowedCurrencies: []string{"INR"},
		CaptureMode:       provider.CaptureModeAutomatic,
		Providers: map[string]provider.ProviderConfig{
			provider.ProviderNameRazorpayLike: {
				PublicKey:     "rzp_key",
				SecretKey:     "rzp_secret",
				WebhookSecret: "rzp_whsec",
			},
		},
	}
	uc, err := NewCreatePaymentIntentUsecase(repo, registry, cfg, nil, WithPaymentIntentIDGenerator(&staticPaymentIntentIDs{}))
	if err != nil {
		t.Fatalf("NewCreatePaymentIntentUsecase() error = %v", err)
	}
	input := validPaymentIntentInput()
	input.Provider = provider.ProviderNameRazorpayLike
	input.Metadata = map[string]string{
		"k1": "v", "k2": "v", "k3": "v", "k4": "v", "k5": "v",
		"k6": "v", "k7": "v", "k8": "v", "k9": "v", "k10": "v",
	}

	_, err = uc.Execute(context.Background(), input)
	if !errors.Is(err, provider.ErrInvalidProviderRequest) {
		t.Fatalf("Execute() error = %v, want provider limit validation error", err)
	}
	if repo.createCalls != 0 {
		t.Fatalf("create calls = %d, want no rows for invalid provider request", repo.createCalls)
	}
}

func TestCreatePaymentIntentMarksProviderFailureWithoutPersistingErrorSecret(t *testing.T) {
	logs := &bytes.Buffer{}
	repo := newFakePaymentIntentRepository()
	gateway := &providertest.FakeProvider{
		NameValue:       provider.ProviderNameStripeLike,
		CreateIntentErr: errors.New("client_secret=must_not_be_written"),
	}
	uc := newPaymentIntentUsecaseForTestWithLogger(t, repo, gateway, &staticPaymentIntentIDs{}, slog.New(slog.NewTextHandler(logs, nil)))

	_, err := uc.Execute(context.Background(), validPaymentIntentInput())
	if err == nil {
		t.Fatal("Execute() error = nil, want provider failure")
	}
	if repo.failureMessage != "payment provider could not create an intent" || repo.failureCode != "provider_error" {
		t.Fatalf("stored failure = %q/%q, want sanitized failure", repo.failureCode, repo.failureMessage)
	}
	if strings.Contains(logs.String(), "must_not_be_written") {
		t.Fatalf("logs include provider secret-bearing error: %s", logs.String())
	}
}

type fakePaymentIntentRepository struct {
	payments       map[string]domain.Payment
	createCalls    int
	updatedAttempt domain.PaymentAttempt
	failureCode    string
	failureMessage string
}

func newFakePaymentIntentRepository() *fakePaymentIntentRepository {
	return &fakePaymentIntentRepository{payments: make(map[string]domain.Payment)}
}

func (r *fakePaymentIntentRepository) FindPaymentByProviderIdempotencyKey(_ context.Context, providerName string, key string) (domain.Payment, error) {
	payment, ok := r.payments[providerName+"/"+key]
	if !ok {
		return domain.Payment{}, domain.ErrPaymentRecordNotFound
	}
	return payment, nil
}

func (r *fakePaymentIntentRepository) CreatePaymentWithAttempt(_ context.Context, payment domain.Payment, _ domain.PaymentAttempt) error {
	r.createCalls++
	key := payment.Provider + "/" + payment.IdempotencyKey
	if _, ok := r.payments[key]; ok {
		return domain.ErrDuplicatePaymentRecord
	}
	r.payments[key] = payment
	return nil
}

func (r *fakePaymentIntentRepository) UpdatePaymentIntent(_ context.Context, payment domain.Payment, attempt domain.PaymentAttempt) error {
	r.payments[payment.Provider+"/"+payment.IdempotencyKey] = payment
	r.updatedAttempt = attempt
	return nil
}

func (r *fakePaymentIntentRepository) FailPaymentIntent(_ context.Context, paymentID string, _ string, code string, message string, _ time.Time) error {
	r.failureCode = code
	r.failureMessage = message
	for key, payment := range r.payments {
		if payment.PaymentID == paymentID {
			payment.Status = domain.PaymentStatusFailed
			r.payments[key] = payment
		}
	}
	return nil
}

type staticPaymentIntentIDs struct{}

func (*staticPaymentIntentIDs) NewPaymentID() (string, error) { return "pay_test", nil }
func (*staticPaymentIntentIDs) NewAttemptID() (string, error) { return "pat_test", nil }

type retrievingFakeProvider struct {
	providertest.FakeProvider
	response      provider.CreateIntentResponse
	retrieveCalls int
}

func (p *retrievingFakeProvider) RetrieveIntent(_ context.Context, _ string) (provider.CreateIntentResponse, error) {
	p.retrieveCalls++
	return p.response, nil
}

func newPaymentIntentUsecaseForTest(t *testing.T, repo PaymentIntentRepository, gateway provider.Provider, ids PaymentIntentIDGenerator) *CreatePaymentIntentUsecase {
	t.Helper()
	return newPaymentIntentUsecaseForTestWithLogger(t, repo, gateway, ids, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
}

func newPaymentIntentUsecaseForTestWithLogger(t *testing.T, repo PaymentIntentRepository, gateway provider.Provider, ids PaymentIntentIDGenerator, logger *slog.Logger) *CreatePaymentIntentUsecase {
	t.Helper()
	registry, err := provider.NewRegistry(gateway)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	uc, err := NewCreatePaymentIntentUsecase(repo, registry, paymentIntentTestConfig(), logger, WithPaymentIntentIDGenerator(ids))
	if err != nil {
		t.Fatalf("NewCreatePaymentIntentUsecase() error = %v", err)
	}
	return uc
}

func paymentIntentTestConfig() provider.Config {
	return provider.Config{
		DefaultProvider:   provider.ProviderNameStripeLike,
		AllowedProviders:  []string{provider.ProviderNameStripeLike},
		AllowedCurrencies: []string{"INR", "USD"},
		CaptureMode:       provider.CaptureModeAutomatic,
		Providers: map[string]provider.ProviderConfig{
			provider.ProviderNameStripeLike: {
				PublicKey:     "pk_test",
				SecretKey:     "sk_test",
				WebhookSecret: "whsec_test",
			},
		},
	}
}

func validPaymentIntentInput() CreatePaymentIntentInput {
	return CreatePaymentIntentInput{
		OrderID:        "ord_123",
		UserID:         "usr_123",
		Amount:         1000,
		Currency:       "inr",
		IdempotencyKey: "payment_intent:ord_123:attempt_1",
		RequestID:      "request_123",
		Customer: PaymentIntentCustomer{
			Name:  "Buyer",
			Email: "buyer@example.com",
			Phone: "+919876543210",
		},
		Metadata: map[string]string{"checkout_reference": "checkout_123"},
	}
}

func existingPayment(t *testing.T, providerIntentID string, status domain.PaymentStatus) domain.Payment {
	t.Helper()
	payment, err := domain.NewPayment(domain.NewPaymentInput{
		PaymentID:        "pay_existing",
		OrderID:          "ord_123",
		UserID:           "usr_123",
		Provider:         provider.ProviderNameStripeLike,
		ProviderIntentID: providerIntentID,
		Amount:           domain.Money{Amount: 1000, Currency: "INR"},
		IdempotencyKey:   "payment_intent:ord_123:attempt_1",
		Now:              time.Now(),
	})
	if err != nil {
		t.Fatalf("NewPayment() error = %v", err)
	}
	payment.Status = status
	return payment
}
