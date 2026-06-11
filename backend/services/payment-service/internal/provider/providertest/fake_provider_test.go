package providertest_test

import (
	"context"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider/providertest"
)

func TestFakeProviderCreateIntent(t *testing.T) {
	fake := &providertest.FakeProvider{}
	res, err := fake.CreateIntent(context.Background(), provider.CreateIntentRequest{
		PaymentID:      "pay_123",
		OrderID:        "ord_123",
		UserID:         "usr_123",
		Amount:         provider.Money{AmountMinor: 1000, Currency: "inr"},
		Customer:       provider.Customer{UserID: "usr_123"},
		IdempotencyKey: "payment_intent:ord_123:1",
		CaptureMode:    provider.CaptureModeAutomatic,
	})
	if err != nil {
		t.Fatalf("CreateIntent() error = %v", err)
	}
	if res.Provider != providertest.ProviderNameFake {
		t.Fatalf("Provider = %q, want fake", res.Provider)
	}
	if res.ProviderIntentID != "fake_intent_pay_123" {
		t.Fatalf("ProviderIntentID = %q, want fake_intent_pay_123", res.ProviderIntentID)
	}
	if fake.CreateIntentCalls != 1 {
		t.Fatalf("CreateIntentCalls = %d, want 1", fake.CreateIntentCalls)
	}
}

func TestFakeProviderVerifyWebhook(t *testing.T) {
	fake := &providertest.FakeProvider{}
	event, err := fake.VerifyWebhook(context.Background(), provider.VerifyWebhookRequest{
		Headers: map[string]string{"X-Test-Signature": "valid"},
		Body:    []byte(`{"type":"payment.captured"}`),
	})
	if err != nil {
		t.Fatalf("VerifyWebhook() error = %v", err)
	}
	if event.Provider != providertest.ProviderNameFake {
		t.Fatalf("Provider = %q, want fake", event.Provider)
	}
	if event.Type != provider.WebhookEventPaymentCaptured {
		t.Fatalf("Type = %q, want payment.captured", event.Type)
	}
}
