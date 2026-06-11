package usecase

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider/providertest"
)

func TestRefundPaymentRequiresReviewWithoutCallingProvider(t *testing.T) {
	repo := newRefundRepositoryFake()
	gateway := &providertest.FakeProvider{NameValue: provider.ProviderNameStripeLike}
	uc := newRefundUsecaseForTest(t, repo, gateway, RefundPolicy{ManualReviewThresholdMinor: 0})

	out, err := uc.Execute(context.Background(), validRefundInput())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if out.Refund.Status != domain.RefundStatusRequested || gateway.RefundCalls != 0 {
		t.Fatalf("output=%+v calls=%d, want requested manual review", out, gateway.RefundCalls)
	}
}

func TestRefundPaymentAutoApprovalCompletesSynchronousPartialRefund(t *testing.T) {
	repo := newRefundRepositoryFake()
	gateway := &providertest.FakeProvider{
		NameValue: provider.ProviderNameStripeLike,
		RefundResponse: provider.RefundResponse{
			ProviderRefundID: "re_123",
			Status:           provider.RefundStatusSucceeded,
			RefundedAmount:   provider.Money{AmountMinor: 250, Currency: "INR"},
		},
	}
	publisher := &fakePaymentEventPublisher{}
	uc := newRefundUsecaseForTestWithPublisher(t, repo, gateway, publisher, RefundPolicy{ManualReviewThresholdMinor: 500})
	out, err := uc.Execute(context.Background(), validRefundInput())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if out.Refund.Status != domain.RefundStatusSucceeded || repo.payment.RefundedAmount != 250 || repo.payment.Status != domain.PaymentStatusPartiallyRefunded {
		t.Fatalf("output=%+v payment=%+v, want succeeded partial refund", out, repo.payment)
	}
	if gateway.RefundCalls != 1 || publisher.calls < 4 {
		t.Fatalf("refund calls=%d events=%d, want provider submission and lifecycle events", gateway.RefundCalls, publisher.calls)
	}
}

func TestRefundPaymentRejectsChangedIdempotentPayload(t *testing.T) {
	repo := newRefundRepositoryFake()
	repo.refund = domain.Refund{
		RefundID: "rfnd_existing", PaymentID: "pay_123", Status: domain.RefundStatusRequested,
		Amount: domain.Money{Amount: 100, Currency: "INR"}, Reason: "Different amount",
		RequestedBy: "finance_1", IdempotencyKey: "refund:pay_123:item_1",
	}
	uc := newRefundUsecaseForTest(t, repo, &providertest.FakeProvider{NameValue: provider.ProviderNameStripeLike}, RefundPolicy{})
	if _, err := uc.Execute(context.Background(), validRefundInput()); !errors.Is(err, ErrRefundIdempotencyConflict) {
		t.Fatalf("Execute() error = %v, want idempotency conflict", err)
	}
}

func TestReviewRefundEnforcesMakerChecker(t *testing.T) {
	repo := newRefundRepositoryFake()
	repo.refund = domain.Refund{
		RefundID: "rfnd_123", PaymentID: "pay_123", Status: domain.RefundStatusRequested,
		Amount: domain.Money{Amount: 250, Currency: "INR"}, Reason: "Cancelled item",
		RequestedBy: "finance_1", IdempotencyKey: "refund:pay_123:item_1",
	}
	uc := newRefundUsecaseForTest(t, repo, &providertest.FakeProvider{NameValue: provider.ProviderNameStripeLike}, RefundPolicy{})
	_, err := uc.Review(context.Background(), ReviewRefundInput{RefundID: "rfnd_123", Decision: "approved", Reason: "Approved", ReviewedBy: "finance_1"})
	if !errors.Is(err, ErrRefundMakerChecker) {
		t.Fatalf("Review() error = %v, want maker-checker failure", err)
	}
}

type staticRefundIDs struct{}

func (staticRefundIDs) NewRefundID() (string, error)      { return "rfnd_123", nil }
func (staticRefundIDs) NewRefundEventID() (string, error) { return "rfe_123", nil }

type refundRepositoryFake struct {
	payment domain.Payment
	refund  domain.Refund
}

func newRefundRepositoryFake() *refundRepositoryFake {
	return &refundRepositoryFake{payment: domain.Payment{
		PaymentID: "pay_123", OrderID: "ord_123", UserID: "usr_123", Provider: provider.ProviderNameStripeLike,
		ProviderPaymentID: "ch_123", Status: domain.PaymentStatusCaptured,
		Amount: domain.Money{Amount: 1000, Currency: "INR"}, CapturedAmount: 1000,
	}}
}

func (r *refundRepositoryFake) GetPaymentByID(context.Context, string) (domain.Payment, error) {
	return r.payment, nil
}
func (r *refundRepositoryFake) GetRefundByID(context.Context, string) (domain.Refund, error) {
	if r.refund.RefundID == "" {
		return domain.Refund{}, domain.ErrRefundRecordNotFound
	}
	return r.refund, nil
}
func (r *refundRepositoryFake) FindRefundByIdempotencyKey(context.Context, string, string) (domain.Refund, error) {
	if r.refund.RefundID == "" {
		return domain.Refund{}, domain.ErrRefundRecordNotFound
	}
	return r.refund, nil
}
func (r *refundRepositoryFake) ReserveRefund(_ context.Context, refund domain.Refund) (domain.Payment, error) {
	if err := domain.ValidateRefundReservation(r.payment, refund, 0); err != nil {
		return domain.Payment{}, err
	}
	r.refund = refund
	return r.payment, nil
}
func (r *refundRepositoryFake) ReviewRefund(_ context.Context, _ string, approved bool, reviewer string, reason string, at time.Time) (domain.Refund, domain.Payment, error) {
	if approved {
		r.refund.Status = domain.RefundStatusApproved
	} else {
		r.refund.Status = domain.RefundStatusRejected
	}
	r.refund.ReviewedBy = reviewer
	r.refund.ReviewReason = reason
	r.refund.ReviewedAt = &at
	r.refund.UpdatedAt = at
	return r.refund, r.payment, nil
}
func (r *refundRepositoryFake) MarkRefundProcessing(_ context.Context, _ string, providerRefundID string, at time.Time) (domain.Refund, error) {
	r.refund.Status = domain.RefundStatusProcessing
	r.refund.ProviderRefundID = providerRefundID
	r.refund.UpdatedAt = at
	return r.refund, nil
}
func (r *refundRepositoryFake) CompleteRefund(_ context.Context, _ string, providerRefundID string, outcome domain.RefundStatus, at time.Time) (domain.Refund, domain.Payment, error) {
	event := domain.VerifiedRefundWebhook{Provider: r.payment.Provider, Outcome: outcome, RefundID: r.refund.RefundID, ProviderRefundID: providerRefundID, Amount: r.refund.Amount}
	payment, refund, reason := domain.ApplyVerifiedRefundOutcome(r.payment, r.refund, event, at)
	if reason != "" {
		return domain.Refund{}, domain.Payment{}, errors.New(reason)
	}
	r.payment, r.refund = payment, refund
	return refund, payment, nil
}

func validRefundInput() RefundPaymentInput {
	return RefundPaymentInput{PaymentID: "pay_123", AmountMinor: 250, Currency: "INR", Reason: "Cancelled item", RequestedBy: "finance_1", IdempotencyKey: "refund:pay_123:item_1"}
}

func newRefundUsecaseForTest(t *testing.T, repo RefundRepository, gateway provider.Provider, policy RefundPolicy) *RefundPaymentUsecase {
	return newRefundUsecaseForTestWithPublisher(t, repo, gateway, &fakePaymentEventPublisher{}, policy)
}

func newRefundUsecaseForTestWithPublisher(t *testing.T, repo RefundRepository, gateway provider.Provider, publisher PaymentEventPublisher, policy RefundPolicy) *RefundPaymentUsecase {
	t.Helper()
	registry, err := provider.NewRegistry(gateway)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	uc, err := NewRefundPaymentUsecase(repo, registry, publisher, policy, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)), WithRefundIDGenerator(staticRefundIDs{}), WithRefundClock(func() time.Time {
		return time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	}))
	if err != nil {
		t.Fatalf("NewRefundPaymentUsecase() error = %v", err)
	}
	return uc
}
