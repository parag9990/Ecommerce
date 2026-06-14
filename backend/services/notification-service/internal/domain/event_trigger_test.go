package domain

import (
	"errors"
	"testing"
	"time"
)

func TestEventNotificationTriggerValidate(t *testing.T) {
	t.Parallel()

	trigger := validEventTrigger()
	if err := trigger.Validate(); err != nil {
		t.Fatalf("Validate() returned error for valid event trigger: %v", err)
	}
}

func TestEventNotificationTriggerRejectsUnsafeRequests(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		change func(*EventNotificationTrigger)
		want   error
	}{
		{name: "invalid recipient", change: func(trigger *EventNotificationTrigger) { trigger.Recipient = "bad" }, want: ErrInvalidRecipient},
		{name: "changed idempotency key", change: func(trigger *EventNotificationTrigger) { trigger.IdempotencyKey = "other" }, want: ErrInvalidEventTrigger},
		{name: "additional variable", change: func(trigger *EventNotificationTrigger) { trigger.Variables["token"] = "hidden" }, want: ErrInvalidEventTrigger},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			trigger := validEventTrigger()
			test.change(&trigger)
			if err := trigger.Validate(); !errors.Is(err, test.want) {
				t.Fatalf("Validate() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestEventDeliveryOutcomeValidate(t *testing.T) {
	t.Parallel()

	valid := EventDeliveryOutcome{
		Status:    DeliveryStatusAccepted,
		Provider:  "smtp",
		UpdatedAt: time.Now(),
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() returned error: %v", err)
	}
	valid.Status = DeliveryStatusProcessing
	if err := valid.Validate(); !errors.Is(err, ErrInvalidDelivery) {
		t.Fatalf("Validate() error = %v, want ErrInvalidDelivery", err)
	}
}

func validEventTrigger() EventNotificationTrigger {
	return EventNotificationTrigger{
		IdempotencyKey: "evt_1:order_status_update:email",
		SourceEventID:  "evt_1",
		SourceType:     "OrderPaid",
		TraceID:        "trace_1",
		UserID:         "user_1",
		Channel:        ChannelEmail,
		Recipient:      "buyer@example.com",
		TemplateKey:    TemplateOrderStatusUpdate,
		Variables: map[string]string{
			"name": "Riya", "order_id": "order_1", "status": "Paid",
		},
	}
}
