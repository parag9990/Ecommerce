package domain

import (
	"errors"
	"testing"
	"time"
)

func TestProviderEventValidate(t *testing.T) {
	t.Parallel()

	event := validProviderEvent()
	if err := event.Validate(); err != nil {
		t.Fatalf("Validate() returned error = %v", err)
	}
	event.Type = DeliveryEventFailed
	event.FailureCode = ""
	if err := event.Validate(); !errors.Is(err, ErrInvalidProviderEvent) {
		t.Fatalf("Validate() failed event error = %v, want ErrInvalidProviderEvent", err)
	}
}

func TestProviderCallbackValidateRequiresCorrelation(t *testing.T) {
	t.Parallel()

	callback := ProviderCallback{
		Provider: "smtp", ProviderEventID: "evt_1", ProviderMessageID: "msg_1",
		Type: DeliveryEventDelivered, Channel: ChannelEmail, OccurredAt: time.Now().UTC(),
	}
	if err := callback.Validate(); err != nil {
		t.Fatalf("Validate() returned error = %v", err)
	}
	callback.ProviderMessageID = ""
	if err := callback.Validate(); !errors.Is(err, ErrInvalidProviderEvent) {
		t.Fatalf("Validate() missing message id error = %v, want ErrInvalidProviderEvent", err)
	}
	callback.ProviderMessageID = "msg_1"
	callback.FailureCode = "bounce"
	if err := callback.Validate(); !errors.Is(err, ErrInvalidProviderEvent) {
		t.Fatalf("Validate() invalid failure code error = %v, want ErrInvalidProviderEvent", err)
	}
}

func validProviderEvent() ProviderEvent {
	now := time.Date(2026, time.May, 27, 12, 0, 0, 0, time.UTC)
	return ProviderEvent{
		ID: "provider_event_1", Provider: "smtp", ProviderEventID: "evt_1",
		ProviderMessageID: "msg_1", DeliveryID: "delivery_1",
		Type: DeliveryEventDelivered, Channel: ChannelEmail, TemplateKey: "order_status_update",
		OccurredAt: now, ReceivedAt: now.Add(time.Second),
	}
}
