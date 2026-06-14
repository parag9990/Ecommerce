package domain

import (
	"fmt"
	"strings"
	"time"
)

// EventNotificationTrigger is the provider-neutral notification selected for a source event.
type EventNotificationTrigger struct {
	IdempotencyKey string
	SourceEventID  string
	SourceType     string
	TraceID        string
	UserID         string
	Channel        Channel
	Recipient      string
	TemplateKey    TemplateKey
	Variables      map[string]string
}

func (t EventNotificationTrigger) Validate() error {
	for value, field := range map[string]string{
		t.IdempotencyKey:      "idempotency_key",
		t.SourceEventID:       "source_event_id",
		t.SourceType:          "source_type",
		t.UserID:              "user_id",
		string(t.Channel):     "channel",
		string(t.TemplateKey): "template_key",
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: %s is required", ErrInvalidEventTrigger, field)
		}
	}
	if t.Channel != ChannelEmail {
		return fmt.Errorf("%w: event rule channel %q is not enabled", ErrInvalidEventTrigger, t.Channel)
	}
	if !validEmail(t.Recipient) {
		return fmt.Errorf("%w: %w: email address is required and must be valid",
			ErrInvalidEventTrigger, ErrInvalidRecipient)
	}
	expectedKey := t.SourceEventID + ":" + string(t.TemplateKey) + ":" + string(t.Channel)
	if t.IdempotencyKey != expectedKey {
		return fmt.Errorf("%w: idempotency key does not match the event notification", ErrInvalidEventTrigger)
	}

	required, exists := eventTemplateVariables[t.TemplateKey]
	if !exists {
		return fmt.Errorf("%w: %w: %q", ErrInvalidEventTrigger, ErrUnsupportedTemplateKey, t.TemplateKey)
	}
	if len(t.Variables) != len(required) {
		return fmt.Errorf("%w: unexpected template variables supplied", ErrInvalidEventTrigger)
	}
	for _, variable := range required {
		if strings.TrimSpace(t.Variables[variable]) == "" {
			return fmt.Errorf("%w: required variable %q is missing", ErrInvalidEventTrigger, variable)
		}
	}
	return nil
}

var eventTemplateVariables = map[TemplateKey][]string{
	TemplateOrderStatusUpdate:   {"name", "order_id", "status"},
	TemplatePaymentStatusUpdate: {"name", "order_id", "payment_status", "amount"},
	TemplateWelcomeUser:         {"name"},
	TemplateSellerApproved:      {"name"},
	TemplateAddressUpdated:      {"name"},
}

// EventDeliveryOutcome updates an event delivery claim after one provider attempt.
type EventDeliveryOutcome struct {
	Status            DeliveryStatus
	Provider          string
	ProviderMessageID string
	UpdatedAt         time.Time
}

func (o EventDeliveryOutcome) Validate() error {
	switch o.Status {
	case DeliveryStatusAccepted, DeliveryStatusRejected, DeliveryStatusFailed:
	default:
		return fmt.Errorf("%w: event outcome status %q is invalid", ErrInvalidDelivery, o.Status)
	}
	if strings.TrimSpace(o.Provider) == "" {
		return fmt.Errorf("%w: provider is required for an event outcome", ErrInvalidDelivery)
	}
	if o.UpdatedAt.IsZero() {
		return fmt.Errorf("%w: event outcome updated_at is required", ErrInvalidDelivery)
	}
	return nil
}
