package domain

import "strings"

type ProviderEventMapping struct {
	ProviderEvent string
	InternalEvent PaymentEvent
	NextStatus    PaymentStatus
}

func ProviderEventMappings() []ProviderEventMapping {
	mappings := make([]ProviderEventMapping, 0, len(providerEventMappings))
	for _, mapping := range providerEventMappings {
		mappings = append(mappings, mapping)
	}
	return mappings
}

func NormalizeProviderEvent(raw string) (ProviderEventMapping, error) {
	event := strings.ToLower(strings.TrimSpace(raw))
	for _, mapping := range providerEventMappings {
		if mapping.ProviderEvent == event {
			return mapping, nil
		}
	}
	return ProviderEventMapping{}, ErrUnknownProviderEvent
}

func NextStatusFromProviderEvent(event PaymentEvent) (PaymentStatus, bool) {
	status, ok := paymentEventTargets[event]
	return status, ok
}

var providerEventMappings = []ProviderEventMapping{
	{
		ProviderEvent: "payment_intent.created",
		InternalEvent: PaymentEventAttemptCreated,
		NextStatus:    PaymentStatusInitiated,
	},
	{
		ProviderEvent: "payment_intent.requires_action",
		InternalEvent: PaymentEventProviderRequiresAction,
		NextStatus:    PaymentStatusRequiresAction,
	},
	{
		ProviderEvent: "payment_intent.authorized",
		InternalEvent: PaymentEventProviderAuthorized,
		NextStatus:    PaymentStatusAuthorized,
	},
	{
		ProviderEvent: "payment_intent.succeeded",
		InternalEvent: PaymentEventProviderCaptured,
		NextStatus:    PaymentStatusCaptured,
	},
	{
		ProviderEvent: "payment.captured",
		InternalEvent: PaymentEventProviderCaptured,
		NextStatus:    PaymentStatusCaptured,
	},
	{
		ProviderEvent: "payment_intent.payment_failed",
		InternalEvent: PaymentEventProviderFailed,
		NextStatus:    PaymentStatusFailed,
	},
	{
		ProviderEvent: "payment.cancelled",
		InternalEvent: PaymentEventProviderFailed,
		NextStatus:    PaymentStatusFailed,
	},
	{
		ProviderEvent: "refund.partially_processed",
		InternalEvent: PaymentEventPartialRefundSucceeded,
		NextStatus:    PaymentStatusPartiallyRefunded,
	},
	{
		ProviderEvent: "refund.succeeded",
		InternalEvent: PaymentEventFullRefundSucceeded,
		NextStatus:    PaymentStatusRefunded,
	},
}
