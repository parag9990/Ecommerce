package analytics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
)

type recorderStore struct {
	delivery domain.Delivery
	event    domain.ProviderEvent
	result   domain.DeliveryEventApplyResult
	findErr  error
}

func (s *recorderStore) FindDeliveryByID(context.Context, string) (domain.Delivery, error) {
	return s.delivery, s.findErr
}

func (s *recorderStore) FindDeliveryByProviderMessageID(context.Context, string, string, domain.Channel) (domain.Delivery, error) {
	return s.delivery, s.findErr
}

func (s *recorderStore) RecordEventAndApplyMilestone(_ context.Context, event domain.ProviderEvent) (domain.DeliveryEventApplyResult, error) {
	s.event = event
	return s.result, nil
}

type recorderObserver struct {
	events []domain.ProviderEvent
}

func (o *recorderObserver) ObserveDeliveryEvent(event domain.ProviderEvent) {
	o.events = append(o.events, event)
}
func (*recorderObserver) ObserveWebhook(string, string)                {}
func (*recorderObserver) ObserveWebhookDuration(string, time.Duration) {}

func TestRecorderAppliesCallbackUsingStoredDeliveryDimensions(t *testing.T) {
	t.Parallel()

	store := &recorderStore{delivery: analyticsDelivery(), result: domain.DeliveryEventApplyResult{MilestoneChanged: true}}
	observer := &recorderObserver{}
	recorder, err := NewRecorder(store, observer)
	if err != nil {
		t.Fatalf("NewRecorder() error = %v", err)
	}
	recorder.WithClock(func() time.Time {
		return time.Date(2026, time.May, 27, 12, 1, 0, 0, time.UTC)
	})
	_, err = recorder.RecordCallback(context.Background(), domain.ProviderCallback{
		Provider: "smtp", ProviderEventID: "evt_delivered", ProviderMessageID: "msg_1",
		Type: domain.DeliveryEventDelivered, Channel: domain.ChannelEmail,
		OccurredAt: time.Date(2026, time.May, 27, 12, 0, 10, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("RecordCallback() error = %v", err)
	}
	if store.event.DeliveryID != "delivery_1" || store.event.TemplateKey != "order_status_update" ||
		store.event.CampaignID != "campaign_1" || len(observer.events) != 1 {
		t.Fatalf("recorded event = %+v, observed = %+v", store.event, observer.events)
	}
}

func TestRecorderDoesNotCountDuplicateMilestone(t *testing.T) {
	t.Parallel()

	store := &recorderStore{delivery: analyticsDelivery(), result: domain.DeliveryEventApplyResult{Duplicate: true}}
	observer := &recorderObserver{}
	recorder, _ := NewRecorder(store, observer)
	_, err := recorder.RecordSent(context.Background(), "delivery_1", "smtp", "msg_1", time.Now().UTC())
	if err != nil || len(observer.events) != 0 {
		t.Fatalf("RecordSent() error = %v, observations = %+v", err, observer.events)
	}
}

func TestRecorderSanitizesTerminalFailureAndPropagatesMissingDelivery(t *testing.T) {
	t.Parallel()

	store := &recorderStore{delivery: analyticsDelivery(), result: domain.DeliveryEventApplyResult{MilestoneChanged: true}}
	recorder, _ := NewRecorder(store, &recorderObserver{})
	_, err := recorder.RecordFailed(context.Background(), "delivery_1", "smtp", "Hard Bounce: user@example.com", time.Now().UTC())
	if err != nil || store.event.FailureCode != "hard_bounce:_user_example_com" {
		t.Fatalf("RecordFailed() error = %v, event = %+v", err, store.event)
	}

	store.findErr = domain.ErrDeliveryNotFound
	if _, err := recorder.RecordSent(context.Background(), "missing", "smtp", "msg", time.Now().UTC()); !errors.Is(err, domain.ErrDeliveryNotFound) {
		t.Fatalf("RecordSent() missing error = %v", err)
	}
}

func TestRecorderRejectsOTPObservedOpen(t *testing.T) {
	t.Parallel()

	delivery := analyticsDelivery()
	delivery.TemplateKey = string(domain.TemplateOTPVerification)
	store := &recorderStore{delivery: delivery}
	recorder, _ := NewRecorder(store, &recorderObserver{})
	_, err := recorder.RecordCallback(context.Background(), domain.ProviderCallback{
		Provider: "smtp", ProviderEventID: "open_1", ProviderMessageID: "msg_1",
		Type: domain.DeliveryEventOpened, Channel: domain.ChannelEmail, OccurredAt: time.Now().UTC(),
	})
	if !errors.Is(err, domain.ErrInvalidProviderEvent) {
		t.Fatalf("RecordCallback() OTP open error = %v, want ErrInvalidProviderEvent", err)
	}
}

func analyticsDelivery() domain.Delivery {
	now := time.Date(2026, time.May, 27, 12, 0, 0, 0, time.UTC)
	return domain.Delivery{
		ID: "delivery_1", UserID: "user_1", Channel: domain.ChannelEmail,
		TemplateKey: "order_status_update", CampaignID: "campaign_1",
		Status: domain.DeliveryStatusAccepted, Provider: "smtp", ProviderMessageID: "msg_1",
		Attempts: 1, CreatedAt: now, UpdatedAt: now,
	}
}
