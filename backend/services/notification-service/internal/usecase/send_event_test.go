package usecase_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/provider"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/usecase"
)

type eventDeliveries struct {
	claim        domain.Delivery
	outcomeKey   string
	outcome      domain.EventDeliveryOutcome
	claimed      bool
	claimErr     error
	outcomeErr   error
	claimCalls   int
	outcomeCalls int
}

func (r *eventDeliveries) ClaimEventDelivery(_ context.Context, delivery domain.Delivery) (bool, error) {
	r.claimCalls++
	r.claim = delivery
	return r.claimed, r.claimErr
}

func (r *eventDeliveries) RecordEventDeliveryOutcome(
	_ context.Context,
	idempotencyKey string,
	outcome domain.EventDeliveryOutcome,
) error {
	r.outcomeCalls++
	r.outcomeKey = idempotencyKey
	r.outcome = outcome
	return r.outcomeErr
}

func TestSendEventDispatchesClaimedNotificationAndStoresMetadata(t *testing.T) {
	t.Parallel()

	deliveries := &eventDeliveries{claimed: true}
	sender := &otpSender{channel: domain.ChannelEmail, result: provider.Result{
		Status: provider.StatusAccepted, ProviderName: "smtp", ProviderMessageID: "message_1",
	}}
	service := newEventService(t, &otpRenderer{result: domain.RenderedMessage{
		Subject: "Order order_1 is Paid", Body: "Hi Riya, your order is paid.",
	}}, sender, deliveries)

	if err := service.TriggerFromEvent(context.Background(), eventTrigger()); err != nil {
		t.Fatalf("TriggerFromEvent() error = %v", err)
	}
	if deliveries.claimCalls != 1 || deliveries.claim.SourceEventID != "evt_1" ||
		deliveries.claim.Status != domain.DeliveryStatusProcessing || deliveries.claim.Attempts != 0 {
		t.Fatalf("delivery claim = %+v", deliveries.claim)
	}
	if sender.calls != 1 || sender.message.Recipient.Email != "buyer@example.com" {
		t.Fatalf("provider message = %+v", sender.message)
	}
	if deliveries.outcomeCalls != 1 || deliveries.outcomeKey != eventTrigger().IdempotencyKey ||
		deliveries.outcome.Status != domain.DeliveryStatusAccepted ||
		deliveries.outcome.ProviderMessageID != "message_1" {
		t.Fatalf("delivery outcome = %+v", deliveries.outcome)
	}
}

func TestSendEventDuplicateDoesNotInvokeProvider(t *testing.T) {
	t.Parallel()

	deliveries := &eventDeliveries{claimed: false}
	sender := &otpSender{channel: domain.ChannelEmail}
	service := newEventService(t, &otpRenderer{result: domain.RenderedMessage{Subject: "Paid", Body: "Paid"}}, sender, deliveries)

	err := service.TriggerFromEvent(context.Background(), eventTrigger())
	if !errors.Is(err, domain.ErrDuplicateEventDelivery) || sender.calls != 0 || deliveries.outcomeCalls != 0 {
		t.Fatalf("TriggerFromEvent() error = %v, sends = %d, outcomes = %d", err, sender.calls, deliveries.outcomeCalls)
	}
}

func TestSendEventRecordsTransientProviderFailureForReclaim(t *testing.T) {
	t.Parallel()

	deliveries := &eventDeliveries{claimed: true}
	sender := &otpSender{channel: domain.ChannelEmail, err: provider.ErrProviderUnavailable}
	service := newEventService(t, &otpRenderer{result: domain.RenderedMessage{Subject: "Paid", Body: "Paid"}}, sender, deliveries)

	err := service.TriggerFromEvent(context.Background(), eventTrigger())
	if !errors.Is(err, provider.ErrProviderUnavailable) ||
		deliveries.outcome.Status != domain.DeliveryStatusFailed || deliveries.outcomeCalls != 1 {
		t.Fatalf("TriggerFromEvent() error = %v, outcome = %+v", err, deliveries.outcome)
	}
}

func TestSendEventStoresSuppressedAuditWithoutRenderingOrProviderCall(t *testing.T) {
	t.Parallel()

	deliveries := &eventDeliveries{claimed: true}
	renderer := &otpRenderer{}
	sender := &otpSender{channel: domain.ChannelEmail}
	gate := newConsentGate(t, &preferenceReader{preference: domain.Preference{}})
	service := newEventServiceWithGate(t, renderer, sender, deliveries, gate)

	if err := service.TriggerFromEvent(context.Background(), eventTrigger()); err != nil {
		t.Fatalf("TriggerFromEvent() error = %v", err)
	}
	if deliveries.claim.Status != domain.DeliveryStatusSuppressed ||
		deliveries.claim.SuppressionReason != domain.ConsentReasonChannelOptedOut ||
		renderer.calls != 0 || sender.calls != 0 || deliveries.outcomeCalls != 0 {
		t.Fatalf("suppressed claim = %+v; renders = %d sends = %d outcomes = %d",
			deliveries.claim, renderer.calls, sender.calls, deliveries.outcomeCalls)
	}
}

func TestSendEventDoesNotClaimUnrenderableMessage(t *testing.T) {
	t.Parallel()

	deliveries := &eventDeliveries{claimed: true}
	service := newEventService(t, &otpRenderer{err: domain.ErrTemplateNotFound},
		&otpSender{channel: domain.ChannelEmail}, deliveries)

	err := service.TriggerFromEvent(context.Background(), eventTrigger())
	if !errors.Is(err, domain.ErrTemplateNotFound) || deliveries.claimCalls != 0 {
		t.Fatalf("TriggerFromEvent() error = %v, claim calls = %d", err, deliveries.claimCalls)
	}
}

func newEventService(
	t *testing.T,
	renderer usecase.Renderer,
	sender provider.Provider,
	deliveries *eventDeliveries,
) *usecase.SendEventService {
	t.Helper()
	return newEventServiceWithGate(t, renderer, sender, deliveries, allowedConsentGate(t))
}

func newEventServiceWithGate(
	t *testing.T,
	renderer usecase.Renderer,
	sender provider.Provider,
	deliveries *eventDeliveries,
	gate *usecase.ConsentGate,
) *usecase.SendEventService {
	t.Helper()
	service, err := usecase.NewSendEventService(
		renderer,
		&otpRegistry{sender: sender},
		deliveries,
		gate,
		&otpAnalytics{},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	if err != nil {
		t.Fatalf("NewSendEventService() error = %v", err)
	}
	service.WithClock(func() time.Time {
		return time.Date(2026, time.May, 27, 0, 0, 0, 0, time.UTC)
	})
	return service
}

func eventTrigger() domain.EventNotificationTrigger {
	return domain.EventNotificationTrigger{
		IdempotencyKey: "evt_1:order_status_update:email",
		SourceEventID:  "evt_1",
		SourceType:     "OrderPaid",
		TraceID:        "trace_1",
		UserID:         "user_1",
		Channel:        domain.ChannelEmail,
		Recipient:      "buyer@example.com",
		TemplateKey:    domain.TemplateOrderStatusUpdate,
		Variables: map[string]string{
			"name": "Riya", "order_id": "order_1", "status": "Paid",
		},
	}
}
