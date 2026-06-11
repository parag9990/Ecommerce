package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider/providertest"
)

func TestRetryPaymentIntentCreatesProviderIntentFromReservedTrustedPayment(t *testing.T) {
	repo := newFakeRetryRepository(t, false)
	gateway := &providertest.FakeProvider{
		NameValue: provider.ProviderNameStripeLike,
		CreateIntentResponse: provider.CreateIntentResponse{
			Provider:         provider.ProviderNameStripeLike,
			ProviderIntentID: "pi_retry",
			Status:           provider.IntentStatusRequiresAction,
			ClientSecret:     "browser_retry_secret",
			RawProviderResponse: json.RawMessage(
				`{"id":"pi_retry","status":"requires_action"}`,
			),
		},
	}
	uc := newRetryUsecaseForTest(t, repo, gateway)
	out, err := uc.Execute(context.Background(), validRetryInput())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if out.PaymentID != "pay_retry" || out.AttemptNo != 2 || out.Replayed || out.ClientPayload.ClientSecret != "browser_retry_secret" {
		t.Fatalf("output = %+v, want created linked retry intent", out)
	}
	if gateway.CreateIntentCalls != 1 || gateway.LastCreateIntent.Amount.AmountMinor != 129900 ||
		gateway.LastCreateIntent.IdempotencyKey != "payment_intent:ord_123:2" {
		t.Fatalf("provider request = %+v calls=%d, want trusted amount and canonical key", gateway.LastCreateIntent, gateway.CreateIntentCalls)
	}
	if gateway.LastCreateIntent.Metadata["retry_of_payment_id"] != "pay_parent" ||
		gateway.LastCreateIntent.Metadata["attempt_no"] != "2" {
		t.Fatalf("metadata = %#v, want retry correlation", gateway.LastCreateIntent.Metadata)
	}
}

func TestRetryPaymentIntentReplayReturnsExistingProviderIntent(t *testing.T) {
	repo := newFakeRetryRepository(t, true)
	repo.reserved.Payment.ProviderIntentID = "pi_existing"
	repo.reserved.Payment.Status = domain.PaymentStatusRequiresAction
	gateway := &retrievingFakeProvider{
		FakeProvider: providertest.FakeProvider{NameValue: provider.ProviderNameStripeLike},
		response: provider.CreateIntentResponse{
			Provider:         provider.ProviderNameStripeLike,
			ProviderIntentID: "pi_existing",
			Status:           provider.IntentStatusRequiresAction,
			ClientSecret:     "retrieved_retry_secret",
			RawProviderResponse: json.RawMessage(
				`{"id":"pi_existing","status":"requires_action"}`,
			),
		},
	}
	uc := newRetryUsecaseForTest(t, repo, gateway)
	out, err := uc.Execute(context.Background(), validRetryInput())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !out.Replayed || out.PaymentID != "pay_retry" || out.ClientPayload.ClientSecret != "retrieved_retry_secret" {
		t.Fatalf("output = %+v, want recovered retry response", out)
	}
	if gateway.CreateIntentCalls != 0 || gateway.retrieveCalls != 1 {
		t.Fatalf("create/retrieve calls = %d/%d, want no duplicate create", gateway.CreateIntentCalls, gateway.retrieveCalls)
	}
}

func TestRetryPaymentIntentLeavesAmbiguousProviderOutcomePending(t *testing.T) {
	repo := newFakeRetryRepository(t, false)
	gateway := &providertest.FakeProvider{
		NameValue: provider.ProviderNameStripeLike,
		CreateIntentErr: provider.NewError(
			provider.ProviderNameStripeLike,
			"create_intent",
			provider.ErrorCodeUnavailable,
			"provider response unavailable",
			provider.WithRetryable(true),
		),
	}
	uc := newRetryUsecaseForTest(t, repo, gateway)
	if _, err := uc.Execute(context.Background(), validRetryInput()); err == nil {
		t.Fatal("Execute() error = nil, want provider uncertainty returned")
	}
	if repo.failCalls != 0 {
		t.Fatalf("FailPaymentIntent calls = %d, want pending ambiguous retry", repo.failCalls)
	}
}

func TestRetryPaymentIntentRejectsMissingRequestKeyBeforeReservation(t *testing.T) {
	repo := newFakeRetryRepository(t, false)
	uc := newRetryUsecaseForTest(t, repo, &providertest.FakeProvider{NameValue: provider.ProviderNameStripeLike})
	input := validRetryInput()
	input.RetryRequestKey = ""
	_, err := uc.Execute(context.Background(), input)
	if !errors.Is(err, provider.ErrInvalidProviderRequest) || repo.reserveCalls != 0 {
		t.Fatalf("Execute() error=%v reserve calls=%d, want validation before reservation", err, repo.reserveCalls)
	}
}

type fakeRetryRepository struct {
	reserved     domain.ReservedRetryPayment
	reserveInput domain.ReserveRetryPaymentInput
	reserveCalls int
	failCalls    int
	updated      domain.Payment
}

func newFakeRetryRepository(t *testing.T, replayed bool) *fakeRetryRepository {
	t.Helper()
	parent, err := domain.NewPayment(domain.NewPaymentInput{
		PaymentID: "pay_parent", OrderID: "ord_123", UserID: "usr_123", Provider: provider.ProviderNameStripeLike,
		Amount: domain.Money{Amount: 129900, Currency: "INR"}, IdempotencyKey: "payment_intent:ord_123:1",
	})
	if err != nil {
		t.Fatalf("NewPayment() error = %v", err)
	}
	parent.Status = domain.PaymentStatusFailed
	child, err := domain.NewRetryPayment(domain.NewRetryPaymentInput{
		Parent: parent, PaymentID: "pay_retry", RootPaymentID: parent.PaymentID, AttemptNo: 2,
		RetryRequestKey: "retry_client_123", ProviderIdempotencyKey: "payment_intent:ord_123:2",
	})
	if err != nil {
		t.Fatalf("NewRetryPayment() error = %v", err)
	}
	return &fakeRetryRepository{reserved: domain.ReservedRetryPayment{
		Payment:  child,
		Attempt:  domain.PaymentAttempt{AttemptID: "pat_retry", PaymentID: child.PaymentID, Status: domain.PaymentAttemptStatusInitiated},
		Replayed: replayed,
	}}
}

func (r *fakeRetryRepository) ReserveRetryPayment(_ context.Context, input domain.ReserveRetryPaymentInput) (domain.ReservedRetryPayment, error) {
	r.reserveCalls++
	r.reserveInput = input
	return r.reserved, nil
}

func (r *fakeRetryRepository) GetPaymentByID(_ context.Context, _ string) (domain.Payment, error) {
	if r.updated.PaymentID != "" {
		return r.updated, nil
	}
	return r.reserved.Payment, nil
}

func (r *fakeRetryRepository) UpdatePaymentIntent(_ context.Context, payment domain.Payment, _ domain.PaymentAttempt) error {
	r.updated = payment
	return nil
}

func (r *fakeRetryRepository) FailPaymentIntent(_ context.Context, _ string, _ string, _ string, _ string, _ time.Time) error {
	r.failCalls++
	return nil
}

func validRetryInput() RetryPaymentIntentInput {
	return RetryPaymentIntentInput{
		FailedPaymentID: "pay_parent",
		BuyerUserID:     "usr_123",
		RetryRequestKey: "retry_client_123",
		RequestID:       "request_123",
	}
}

func newRetryUsecaseForTest(t *testing.T, repo RetryPaymentRepository, gateway provider.Provider) *RetryPaymentIntentUsecase {
	t.Helper()
	registry, err := provider.NewRegistry(gateway)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	uc, err := NewRetryPaymentIntentUsecase(
		repo,
		registry,
		paymentIntentTestConfig(),
		domain.RetryPolicy{MaxAttempts: 3, Cooldown: 5 * time.Second},
		slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
		WithRetryPaymentIntentIDGenerator(&staticPaymentIntentIDs{}),
		WithRetryPaymentIntentClock(func() time.Time { return time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC) }),
	)
	if err != nil {
		t.Fatalf("NewRetryPaymentIntentUsecase() error = %v", err)
	}
	return uc
}
