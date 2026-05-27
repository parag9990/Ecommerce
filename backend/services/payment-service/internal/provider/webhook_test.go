package provider_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider"
)

func TestStripeLikeVerifyWebhookAuthenticatesAndNormalizesPayment(t *testing.T) {
	p, err := provider.NewStripeLikeProvider(provider.ProviderConfig{
		PublicKey:                 "pk_test",
		SecretKey:                 "sk_test",
		WebhookSecret:             "whsec_test",
		WebhookTimestampTolerance: time.Minute,
	}, nil)
	if err != nil {
		t.Fatalf("NewStripeLikeProvider() error = %v", err)
	}
	body := []byte(`{"id":"evt_123","type":"payment_intent.succeeded","created":1710000000,"data":{"object":{"id":"pi_123","latest_charge":"ch_123","amount":1000,"amount_received":1000,"currency":"inr","metadata":{"payment_id":"pay_123","order_id":"ord_123"}}}}`)
	timestamp := time.Now().Unix()
	signature := hmacHex("whsec_test", []byte(fmt.Sprintf("%d.%s", timestamp, body)))

	event, err := p.VerifyWebhook(context.Background(), provider.VerifyWebhookRequest{
		Headers: map[string]string{"Stripe-Signature": fmt.Sprintf("t=%d,v1=%s", timestamp, signature)},
		Body:    body,
	})
	if err != nil {
		t.Fatalf("VerifyWebhook() error = %v", err)
	}
	if event.Type != provider.WebhookEventPaymentCaptured || event.PaymentID != "pay_123" || event.ProviderIntentID != "pi_123" {
		t.Fatalf("event = %+v, want normalized captured payment", event)
	}
}

func TestStripeLikeVerifyWebhookRejectsInvalidOrStaleSignature(t *testing.T) {
	p, err := provider.NewStripeLikeProvider(provider.ProviderConfig{
		PublicKey:                 "pk_test",
		SecretKey:                 "sk_test",
		WebhookSecret:             "whsec_test",
		WebhookTimestampTolerance: time.Second,
	}, nil)
	if err != nil {
		t.Fatalf("NewStripeLikeProvider() error = %v", err)
	}
	_, err = p.VerifyWebhook(context.Background(), provider.VerifyWebhookRequest{
		Headers: map[string]string{"Stripe-Signature": "t=1,v1=bad"},
		Body:    []byte(`{"id":"evt_123"}`),
	})
	if !errors.Is(err, provider.ErrInvalidWebhookSignature) {
		t.Fatalf("VerifyWebhook() error = %v, want invalid signature", err)
	}
}

func TestStripeLikeVerifyWebhookReportsSignedMalformedPayloadAsInvalidRequest(t *testing.T) {
	p, err := provider.NewStripeLikeProvider(provider.ProviderConfig{
		PublicKey:     "pk_test",
		SecretKey:     "sk_test",
		WebhookSecret: "whsec_test",
	}, nil)
	if err != nil {
		t.Fatalf("NewStripeLikeProvider() error = %v", err)
	}
	body := []byte(`{"malformed"`)
	timestamp := time.Now().Unix()
	signature := hmacHex("whsec_test", []byte(fmt.Sprintf("%d.%s", timestamp, body)))
	_, err = p.VerifyWebhook(context.Background(), provider.VerifyWebhookRequest{
		Headers: map[string]string{"Stripe-Signature": fmt.Sprintf("t=%d,v1=%s", timestamp, signature)},
		Body:    body,
	})
	if !errors.Is(err, provider.ErrInvalidProviderRequest) {
		t.Fatalf("VerifyWebhook() error = %v, want invalid provider request", err)
	}
}

func TestRazorpayLikeVerifyWebhookAuthenticatesAndNormalizesPayment(t *testing.T) {
	p, err := provider.NewRazorpayLikeProvider(provider.ProviderConfig{
		PublicKey:     "rzp_key",
		SecretKey:     "rzp_secret",
		WebhookSecret: "rzp_whsec",
	}, nil)
	if err != nil {
		t.Fatalf("NewRazorpayLikeProvider() error = %v", err)
	}
	body := []byte(`{"event":"payment.captured","created_at":1710000000,"payload":{"payment":{"entity":{"id":"pay_gateway_123","order_id":"order_123","amount":1000,"currency":"INR","notes":{"payment_id":"pay_123","order_id":"ord_123"}}}}}`)
	event, err := p.VerifyWebhook(context.Background(), provider.VerifyWebhookRequest{
		Headers: map[string]string{"X-Razorpay-Signature": hmacHex("rzp_whsec", body)},
		Body:    body,
	})
	if err != nil {
		t.Fatalf("VerifyWebhook() error = %v", err)
	}
	if event.Type != provider.WebhookEventPaymentCaptured || event.PaymentID != "pay_123" || event.ProviderPaymentID != "pay_gateway_123" {
		t.Fatalf("event = %+v, want normalized captured payment", event)
	}
}

func TestStripeLikeVerifyWebhookNormalizesRefundOutcome(t *testing.T) {
	p, err := provider.NewStripeLikeProvider(provider.ProviderConfig{PublicKey: "pk_test", SecretKey: "sk_test", WebhookSecret: "whsec_test"}, nil)
	if err != nil {
		t.Fatalf("NewStripeLikeProvider() error = %v", err)
	}
	body := []byte(`{"id":"evt_refund","type":"refund.succeeded","created":1710000000,"data":{"object":{"id":"re_123","status":"succeeded","charge":"ch_123","amount":250,"currency":"inr","metadata":{"refund_id":"rfnd_123","payment_id":"pay_123"}}}}`)
	timestamp := time.Now().Unix()
	signature := hmacHex("whsec_test", []byte(fmt.Sprintf("%d.%s", timestamp, body)))
	event, err := p.VerifyWebhook(context.Background(), provider.VerifyWebhookRequest{Headers: map[string]string{"Stripe-Signature": fmt.Sprintf("t=%d,v1=%s", timestamp, signature)}, Body: body})
	if err != nil {
		t.Fatalf("VerifyWebhook() error = %v", err)
	}
	if event.Type != provider.WebhookEventRefundSucceeded || event.RefundID != "rfnd_123" || event.ProviderRefundID != "re_123" {
		t.Fatalf("refund event = %+v", event)
	}
}

func TestRazorpayLikeVerifyWebhookNormalizesRefundOutcome(t *testing.T) {
	p, err := provider.NewRazorpayLikeProvider(provider.ProviderConfig{PublicKey: "rzp_key", SecretKey: "rzp_secret", WebhookSecret: "rzp_whsec"}, nil)
	if err != nil {
		t.Fatalf("NewRazorpayLikeProvider() error = %v", err)
	}
	body := []byte(`{"event":"refund.processed","created_at":1710000000,"payload":{"refund":{"entity":{"id":"rfnd_gateway_1","payment_id":"ch_123","amount":250,"currency":"INR","notes":{"refund_id":"rfnd_123","payment_id":"pay_123"}}}}}`)
	event, err := p.VerifyWebhook(context.Background(), provider.VerifyWebhookRequest{Headers: map[string]string{"X-Razorpay-Signature": hmacHex("rzp_whsec", body)}, Body: body})
	if err != nil {
		t.Fatalf("VerifyWebhook() error = %v", err)
	}
	if event.Type != provider.WebhookEventRefundSucceeded || event.ProviderRefundID != "rfnd_gateway_1" || event.PaymentID != "pay_123" {
		t.Fatalf("refund event = %+v", event)
	}
}

func hmacHex(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
