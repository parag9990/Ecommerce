package provider

import (
	"fmt"
	"strings"
)

func BuildPaymentIntentIdempotencyKey(orderID string, attemptNo int) string {
	return fmt.Sprintf("payment_intent:%s:%d", idempotencyComponent(orderID), attemptNo)
}

func BuildWebhookIdempotencyKey(providerName string, eventID string) string {
	return fmt.Sprintf("webhook:%s:%s", idempotencyComponent(NormalizeProviderName(providerName)), idempotencyComponent(eventID))
}

func BuildRefundIdempotencyKey(paymentID string, refundKey string) string {
	return fmt.Sprintf("refund:%s:%s", idempotencyComponent(paymentID), idempotencyComponent(refundKey))
}

func idempotencyComponent(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, ":", "_")
	return value
}
