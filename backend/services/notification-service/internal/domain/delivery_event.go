package domain

import (
	"fmt"
	"strings"
	"time"
)

type DeliveryEventType string

const (
	DeliveryEventSent      DeliveryEventType = "sent"
	DeliveryEventDelivered DeliveryEventType = "delivered"
	DeliveryEventFailed    DeliveryEventType = "failed"
	DeliveryEventOpened    DeliveryEventType = "opened"
)

func (t DeliveryEventType) IsSupported() bool {
	switch t {
	case DeliveryEventSent, DeliveryEventDelivered, DeliveryEventFailed, DeliveryEventOpened:
		return true
	default:
		return false
	}
}

// ProviderEvent is a sanitized, normalized lifecycle event. It intentionally
// contains no recipient address, rendered content, raw callback, or signature.
type ProviderEvent struct {
	ID                string            `bson:"_id" json:"id"`
	Provider          string            `bson:"provider" json:"provider"`
	ProviderEventID   string            `bson:"provider_event_id" json:"provider_event_id"`
	ProviderMessageID string            `bson:"provider_message_id,omitempty" json:"provider_message_id,omitempty"`
	DeliveryID        string            `bson:"delivery_id" json:"delivery_id"`
	Type              DeliveryEventType `bson:"type" json:"type"`
	Channel           Channel           `bson:"channel" json:"channel"`
	TemplateKey       string            `bson:"template_key" json:"template_key"`
	CampaignID        string            `bson:"campaign_id,omitempty" json:"campaign_id,omitempty"`
	OccurredAt        time.Time         `bson:"occurred_at" json:"occurred_at"`
	ReceivedAt        time.Time         `bson:"received_at" json:"received_at"`
	FailureCode       string            `bson:"failure_code,omitempty" json:"failure_code,omitempty"`
}

func (e ProviderEvent) Validate() error {
	for value, field := range map[string]string{
		e.ID:              "id",
		e.Provider:        "provider",
		e.ProviderEventID: "provider_event_id",
		e.DeliveryID:      "delivery_id",
		e.TemplateKey:     "template_key",
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: %s is required", ErrInvalidProviderEvent, field)
		}
	}
	if !e.Type.IsSupported() {
		return fmt.Errorf("%w: unsupported type %q", ErrInvalidProviderEvent, e.Type)
	}
	if !e.Channel.IsSupported() {
		return fmt.Errorf("%w: %w: %q", ErrInvalidProviderEvent, ErrUnsupportedChannel, e.Channel)
	}
	if e.OccurredAt.IsZero() || e.ReceivedAt.IsZero() {
		return fmt.Errorf("%w: occurred_at and received_at are required", ErrInvalidProviderEvent)
	}
	if e.Type == DeliveryEventFailed && strings.TrimSpace(e.FailureCode) == "" {
		return fmt.Errorf("%w: failed events require failure_code", ErrInvalidProviderEvent)
	}
	if e.Type != DeliveryEventFailed && strings.TrimSpace(e.FailureCode) != "" {
		return fmt.Errorf("%w: failure_code is only valid for failed events", ErrInvalidProviderEvent)
	}
	return nil
}

// ProviderCallback is the minimal normalized input accepted after webhook
// authentication. Delivery-owned dimensions are populated from storage.
type ProviderCallback struct {
	Provider          string
	ProviderEventID   string
	ProviderMessageID string
	Type              DeliveryEventType
	Channel           Channel
	OccurredAt        time.Time
	FailureCode       string
}

func (c ProviderCallback) Validate() error {
	for value, field := range map[string]string{
		c.Provider:          "provider",
		c.ProviderEventID:   "provider_event_id",
		c.ProviderMessageID: "provider_message_id",
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: %s is required", ErrInvalidProviderEvent, field)
		}
	}
	if !c.Type.IsSupported() || !c.Channel.IsSupported() || c.OccurredAt.IsZero() {
		return fmt.Errorf("%w: callback type, channel, and occurred_at are required", ErrInvalidProviderEvent)
	}
	if c.Type != DeliveryEventFailed && strings.TrimSpace(c.FailureCode) != "" {
		return fmt.Errorf("%w: failure_code is only valid for failed callbacks", ErrInvalidProviderEvent)
	}
	return nil
}

type DeliveryEventApplyResult struct {
	Duplicate        bool
	MilestoneChanged bool
}
