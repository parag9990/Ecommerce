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

func TestHandleWebhookProcessesVerifiedCaptureAndPublishesEvent(t *testing.T) {
	repo := &fakeWebhookRepository{result: processedWebhookResult()}
	publisher := &fakePaymentEventPublisher{}
	gateway := validWebhookFakeProvider()
	uc := newWebhookUsecaseForTest(t, repo, publisher, gateway)

	out, err := uc.Execute(context.Background(), HandleWebhookInput{
		Provider: provider.ProviderNameStripeLike,
		Headers:  map[string]string{"Stripe-Signature": "verified-by-adapter"},
		RawBody:  []byte(`{"id":"evt_123"}`),
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if out.ProcessingStatus != domain.WebhookProcessingStatusProcessed || out.PaymentID != "pay_123" {
		t.Fatalf("output = %+v, want processed payment", out)
	}
	if repo.event.Event != domain.PaymentEventProviderCaptured || repo.event.WebhookEventID != "whe_test" {
		t.Fatalf("stored event = %+v, want mapped captured webhook", repo.event)
	}
	if publisher.calls != 1 || publisher.event.EventType != "PaymentCaptured" || publisher.topic != PaymentEventsTopic {
		t.Fatalf("publisher = %+v calls=%d, want PaymentCaptured publish", publisher.event, publisher.calls)
	}
}

func TestHandleWebhookDuplicateDoesNotPublishAgain(t *testing.T) {
	repo := &fakeWebhookRepository{result: domain.WebhookProcessingResult{
		WebhookEventID:   "whe_existing",
		ProviderEventID:  "evt_123",
		ProcessingStatus: domain.WebhookProcessingStatusDuplicate,
	}}
	publisher := &fakePaymentEventPublisher{}
	uc := newWebhookUsecaseForTest(t, repo, publisher, validWebhookFakeProvider())

	out, err := uc.Execute(context.Background(), HandleWebhookInput{
		Provider: provider.ProviderNameStripeLike,
		Headers:  map[string]string{"Stripe-Signature": "verified-by-adapter"},
		RawBody:  []byte(`{"id":"evt_123"}`),
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if out.ProcessingStatus != domain.WebhookProcessingStatusDuplicate || publisher.calls != 0 {
		t.Fatalf("output=%+v publish calls=%d, want duplicate without publish", out, publisher.calls)
	}
}

func TestHandleWebhookRejectsInvalidSignatureBeforePersistence(t *testing.T) {
	repo := &fakeWebhookRepository{}
	gateway := validWebhookFakeProvider()
	gateway.VerifyWebhookErr = provider.ErrInvalidWebhookSignature
	uc := newWebhookUsecaseForTest(t, repo, &fakePaymentEventPublisher{}, gateway)

	_, err := uc.Execute(context.Background(), HandleWebhookInput{
		Provider: provider.ProviderNameStripeLike,
		Headers:  map[string]string{"Stripe-Signature": "invalid"},
		RawBody:  []byte(`{"id":"evt_123"}`),
	})
	if !errors.Is(err, provider.ErrInvalidWebhookSignature) || repo.calls != 0 {
		t.Fatalf("Execute() error=%v repo calls=%d, want signature rejection before persistence", err, repo.calls)
	}
}

func TestHandleWebhookProcessesVerifiedRefundAndPublishesEvent(t *testing.T) {
	repo := &fakeWebhookRepository{refundResult: domain.RefundWebhookProcessingResult{
		WebhookEventID:   "whe_test",
		ProviderEventID:  "evt_refund",
		ProcessingStatus: domain.WebhookProcessingStatusProcessed,
		StatusChanged:    true,
		Refund: domain.Refund{
			RefundID:  "rfnd_123",
			PaymentID: "pay_123",
			Status:    domain.RefundStatusSucceeded,
			Amount:    domain.Money{Amount: 250, Currency: "INR"},
		},
		Payment: domain.Payment{
			PaymentID: "pay_123",
			OrderID:   "ord_123",
			Provider:  provider.ProviderNameStripeLike,
			Status:    domain.PaymentStatusPartiallyRefunded,
		},
	}}
	publisher := &fakePaymentEventPublisher{}
	gateway := &providertest.FakeProvider{
		NameValue: provider.ProviderNameStripeLike,
		WebhookEvent: provider.WebhookEvent{
			Provider:         provider.ProviderNameStripeLike,
			ProviderEventID:  "evt_refund",
			Type:             provider.WebhookEventRefundSucceeded,
			RefundID:         "rfnd_123",
			PaymentID:        "pay_123",
			ProviderRefundID: "re_123",
			Amount:           provider.Money{AmountMinor: 250, Currency: "INR"},
			RawPayload:       json.RawMessage(`{"type":"refund.succeeded","refund_id":"rfnd_123"}`),
		},
	}
	uc := newWebhookUsecaseForTest(t, repo, publisher, gateway)
	out, err := uc.Execute(context.Background(), HandleWebhookInput{Provider: provider.ProviderNameStripeLike, Headers: map[string]string{"signature": "ok"}, RawBody: []byte(`{"event":"refund"}`)})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if out.RefundID != "rfnd_123" || out.RefundStatus != domain.RefundStatusSucceeded || repo.refundCalls != 1 {
		t.Fatalf("output=%+v refund calls=%d, want successful refund webhook", out, repo.refundCalls)
	}
	if publisher.event.EventType != "RefundSucceeded" || publisher.event.RefundID != "rfnd_123" {
		t.Fatalf("published event = %+v, want refund outcome", publisher.event)
	}
}

type fakeWebhookRepository struct {
	calls        int
	event        domain.VerifiedPaymentWebhook
	result       domain.WebhookProcessingResult
	err          error
	refundCalls  int
	refundEvent  domain.VerifiedRefundWebhook
	refundResult domain.RefundWebhookProcessingResult
	refundErr    error
}

func (r *fakeWebhookRepository) ProcessRefundWebhook(_ context.Context, event domain.VerifiedRefundWebhook) (domain.RefundWebhookProcessingResult, error) {
	r.refundCalls++
	r.refundEvent = event
	return r.refundResult, r.refundErr
}

func (r *fakeWebhookRepository) ProcessWebhook(_ context.Context, event domain.VerifiedPaymentWebhook) (domain.WebhookProcessingResult, error) {
	r.calls++
	r.event = event
	return r.result, r.err
}

type fakePaymentEventPublisher struct {
	calls int
	topic string
	event domain.PaymentDomainEvent
	err   error
}

func (p *fakePaymentEventPublisher) Publish(_ context.Context, topic string, event domain.PaymentDomainEvent) error {
	p.calls++
	p.topic = topic
	p.event = event
	return p.err
}

type staticWebhookIDs struct{}

func (staticWebhookIDs) NewWebhookEventID() (string, error) { return "whe_test", nil }

func validWebhookFakeProvider() *providertest.FakeProvider {
	return &providertest.FakeProvider{
		NameValue: provider.ProviderNameStripeLike,
		WebhookEvent: provider.WebhookEvent{
			Provider:          provider.ProviderNameStripeLike,
			ProviderEventID:   "evt_123",
			Type:              provider.WebhookEventPaymentCaptured,
			PaymentID:         "pay_123",
			ProviderIntentID:  "pi_123",
			ProviderPaymentID: "ch_123",
			Amount:            provider.Money{AmountMinor: 1000, Currency: "INR"},
			OccurredAtUnix:    time.Date(2026, 5, 26, 10, 30, 0, 0, time.UTC).Unix(),
			RawPayload:        json.RawMessage(`{"type":"payment.captured","payment_id":"pay_123"}`),
		},
	}
}

func processedWebhookResult() domain.WebhookProcessingResult {
	return domain.WebhookProcessingResult{
		WebhookEventID:   "whe_test",
		ProviderEventID:  "evt_123",
		ProcessingStatus: domain.WebhookProcessingStatusProcessed,
		StatusChanged:    true,
		Payment: domain.Payment{
			PaymentID:         "pay_123",
			OrderID:           "ord_123",
			Provider:          provider.ProviderNameStripeLike,
			ProviderPaymentID: "ch_123",
			Status:            domain.PaymentStatusCaptured,
			Amount:            domain.Money{Amount: 1000, Currency: "INR"},
		},
	}
}

func newWebhookUsecaseForTest(t *testing.T, repo WebhookRepository, publisher PaymentEventPublisher, gateway provider.Provider) *HandleWebhookUsecase {
	t.Helper()
	registry, err := provider.NewRegistry(gateway)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	uc, err := NewHandleWebhookUsecase(
		repo,
		registry,
		publisher,
		slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
		WithWebhookIDGenerator(staticWebhookIDs{}),
		WithWebhookClock(func() time.Time { return time.Date(2026, 5, 26, 10, 31, 0, 0, time.UTC) }),
	)
	if err != nil {
		t.Fatalf("NewHandleWebhookUsecase() error = %v", err)
	}
	return uc
}
