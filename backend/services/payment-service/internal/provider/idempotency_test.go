package provider_test

import (
	"testing"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider"
)

func TestBuildIdempotencyKeys(t *testing.T) {
	tests := map[string]string{
		provider.BuildPaymentIntentIdempotencyKey("ord:123", 2):     "payment_intent:ord_123:2",
		provider.BuildWebhookIdempotencyKey("Stripe_Like", "evt:1"): "webhook:stripe_like:evt_1",
		provider.BuildRefundIdempotencyKey("pay:1", "refund:2"):     "refund:pay_1:refund_2",
	}
	for got, want := range tests {
		if got != want {
			t.Fatalf("idempotency key = %q, want %q", got, want)
		}
		if err := provider.ValidateIdempotencyKey(got); err != nil {
			t.Fatalf("ValidateIdempotencyKey(%q) error = %v", got, err)
		}
	}
}
