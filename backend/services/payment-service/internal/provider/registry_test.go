package provider_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider/providertest"
)

func TestRegistryRejectsDuplicateProvider(t *testing.T) {
	_, err := provider.NewRegistry(
		&providertest.FakeProvider{NameValue: "stripe_like"},
		&providertest.FakeProvider{NameValue: " stripe_like "},
	)
	if !errors.Is(err, provider.ErrDuplicateProvider) {
		t.Fatalf("NewRegistry() error = %v, want duplicate provider", err)
	}
}

func TestRegistryReturnsProviderByNormalizedName(t *testing.T) {
	registry, err := provider.NewRegistry(&providertest.FakeProvider{NameValue: "stripe_like"})
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}

	got, err := registry.Get(" STRIPE_LIKE ")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Name() != "stripe_like" {
		t.Fatalf("provider name = %q, want stripe_like", got.Name())
	}
	if !reflect.DeepEqual(registry.Names(), []string{"stripe_like"}) {
		t.Fatalf("Names() = %#v, want stripe_like", registry.Names())
	}
}

func TestRegistryUnknownProviderReturnsTypedError(t *testing.T) {
	registry, err := provider.NewRegistry()
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}

	_, err = registry.Get("missing")
	if !errors.Is(err, provider.ErrProviderNotFound) {
		t.Fatalf("Get() error = %v, want provider not found", err)
	}
	code, ok := provider.CodeOf(err)
	if !ok || code != provider.ErrorCodeProviderNotFound {
		t.Fatalf("CodeOf() = %q/%v, want provider_not_found/true", code, ok)
	}
}

func TestNewRegistryFromConfigBuildsKnownProviders(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	registry, err := provider.NewRegistryFromConfig(provider.Config{
		DefaultProvider:   provider.ProviderNameStripeLike,
		AllowedProviders:  []string{provider.ProviderNameRazorpayLike, provider.ProviderNameStripeLike},
		AllowedCurrencies: []string{"INR", "usd"},
		CaptureMode:       provider.CaptureModeAutomatic,
		Providers: map[string]provider.ProviderConfig{
			provider.ProviderNameStripeLike: {
				PublicKey:     "pk_test",
				SecretKey:     "sk_test",
				WebhookSecret: "whsec_test",
			},
			provider.ProviderNameRazorpayLike: {
				PublicKey:     "rzp_test",
				SecretKey:     "rzp_secret",
				WebhookSecret: "rzp_whsec",
			},
		},
	}, logger)
	if err != nil {
		t.Fatalf("NewRegistryFromConfig() error = %v", err)
	}

	want := []string{provider.ProviderNameRazorpayLike, provider.ProviderNameStripeLike}
	if !reflect.DeepEqual(registry.Names(), want) {
		t.Fatalf("Names() = %#v, want %#v", registry.Names(), want)
	}
}

func TestNewRegistryFromConfigRejectsMissingSecrets(t *testing.T) {
	_, err := provider.NewRegistryFromConfig(provider.Config{
		DefaultProvider:   provider.ProviderNameStripeLike,
		AllowedProviders:  []string{provider.ProviderNameStripeLike},
		AllowedCurrencies: []string{"INR"},
		CaptureMode:       provider.CaptureModeAutomatic,
		Providers: map[string]provider.ProviderConfig{
			provider.ProviderNameStripeLike: {PublicKey: "pk_test"},
		},
	}, nil)
	if !errors.Is(err, provider.ErrInvalidProviderRequest) {
		t.Fatalf("NewRegistryFromConfig() error = %v, want invalid provider request", err)
	}
}

func TestConfigValidatesAllowedCurrency(t *testing.T) {
	cfg := provider.Config{
		AllowedCurrencies: []string{"inr", "usd"},
		CaptureMode:       provider.CaptureModeAutomatic,
	}
	if err := cfg.ValidateCurrency("INR"); err != nil {
		t.Fatalf("ValidateCurrency(INR) error = %v", err)
	}
	if err := cfg.ValidateCurrency("EUR"); !errors.Is(err, provider.ErrInvalidProviderRequest) {
		t.Fatalf("ValidateCurrency(EUR) error = %v, want invalid provider request", err)
	}
}

func TestProviderConfigRejectsInsecureRemoteEndpoint(t *testing.T) {
	err := (provider.ProviderConfig{
		PublicKey:     "pk_test",
		SecretKey:     "sk_test",
		WebhookSecret: "whsec_test",
		BaseURL:       "http://gateway.example.test",
	}).Validate(provider.ProviderNameStripeLike)
	if !errors.Is(err, provider.ErrInvalidProviderRequest) {
		t.Fatalf("Validate() error = %v, want insecure endpoint rejected", err)
	}
}

func TestStripeLikeCreateIntentMapsGatewayRequestAndReturnsSanitizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "sk_test" || password != "" {
			t.Errorf("BasicAuth() = %q/%q/%v, want configured secret", username, password, ok)
		}
		if r.Header.Get("Idempotency-Key") != "payment_intent:ord_123:1" {
			t.Errorf("Idempotency-Key = %q", r.Header.Get("Idempotency-Key"))
		}
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm() error = %v", err)
		}
		if r.Form.Get("amount") != "1000" || r.Form.Get("currency") != "inr" || r.Form.Get("metadata[payment_id]") != "pay_123" {
			t.Errorf("form = %#v, want amount/currency/metadata", r.Form)
		}
		_, _ = w.Write([]byte(`{"id":"pi_123","status":"requires_action","client_secret":"browser_only_secret"}`))
	}))
	defer server.Close()

	stripeProvider, err := provider.NewStripeLikeProvider(provider.ProviderConfig{
		PublicKey:     "pk_test",
		SecretKey:     "sk_test",
		WebhookSecret: "whsec_test",
		BaseURL:       server.URL,
	}, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
	if err != nil {
		t.Fatalf("NewStripeLikeProvider() error = %v", err)
	}

	res, err := stripeProvider.CreateIntent(context.Background(), validCreateIntentRequest())
	if err != nil {
		t.Fatalf("CreateIntent() error = %v", err)
	}
	if res.ProviderIntentID != "pi_123" || res.ClientSecret != "browser_only_secret" || res.Status != provider.IntentStatusRequiresAction {
		t.Fatalf("response = %+v, want normalized intent", res)
	}
	if strings.Contains(string(res.RawProviderResponse), "client_secret") || strings.Contains(string(res.RawProviderResponse), "browser_only_secret") {
		t.Fatalf("raw audit response contains client secret: %s", res.RawProviderResponse)
	}
}

func TestStripeLikeRetrieveIntentReturnsBrowserSecretWithoutStoringItInAuditPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/payment_intents/pi_existing" {
			t.Errorf("request = %s %s, want intent retrieval", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"id":"pi_existing","status":"requires_action","client_secret":"retrieved_browser_secret"}`))
	}))
	defer server.Close()

	stripeProvider, err := provider.NewStripeLikeProvider(provider.ProviderConfig{
		PublicKey:     "pk_test",
		SecretKey:     "sk_test",
		WebhookSecret: "whsec_test",
		BaseURL:       server.URL,
	}, nil)
	if err != nil {
		t.Fatalf("NewStripeLikeProvider() error = %v", err)
	}
	res, err := stripeProvider.RetrieveIntent(context.Background(), "pi_existing")
	if err != nil {
		t.Fatalf("RetrieveIntent() error = %v", err)
	}
	if res.ClientSecret != "retrieved_browser_secret" || strings.Contains(string(res.RawProviderResponse), "retrieved_browser_secret") {
		t.Fatalf("response = %+v/raw=%s, want client-only secret", res, res.RawProviderResponse)
	}
}

func TestRazorpayLikeCreateIntentMapsOrderRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "rzp_key" || password != "rzp_secret" {
			t.Errorf("BasicAuth() = %q/%q/%v, want configured credentials", username, password, ok)
		}
		if r.Header.Get("X-Razorpay-Idempotency-Key") == "" {
			t.Error("X-Razorpay-Idempotency-Key is empty")
		}
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("Decode() error = %v", err)
		}
		if request["currency"] != "INR" || request["receipt"] != "pay_123" {
			t.Errorf("request = %#v, want currency and receipt", request)
		}
		_, _ = w.Write([]byte(`{"id":"order_123","status":"created","receipt":"pay_123"}`))
	}))
	defer server.Close()

	razorpayProvider, err := provider.NewRazorpayLikeProvider(provider.ProviderConfig{
		PublicKey:     "rzp_key",
		SecretKey:     "rzp_secret",
		WebhookSecret: "rzp_whsec",
		BaseURL:       server.URL,
	}, nil)
	if err != nil {
		t.Fatalf("NewRazorpayLikeProvider() error = %v", err)
	}
	res, err := razorpayProvider.CreateIntent(context.Background(), validCreateIntentRequest())
	if err != nil {
		t.Fatalf("CreateIntent() error = %v", err)
	}
	if res.ProviderIntentID != "order_123" || res.Status != provider.IntentStatusInitiated || res.FrontendPayload["provider_order_id"] != "order_123" {
		t.Fatalf("response = %+v, want order payload", res)
	}
}

func TestConfiguredProviderNonIntentOperationRemainsDisabled(t *testing.T) {
	stripeProvider, err := provider.NewStripeLikeProvider(provider.ProviderConfig{
		PublicKey:     "pk_test",
		SecretKey:     "sk_test",
		WebhookSecret: "whsec_test",
	}, nil)
	if err != nil {
		t.Fatalf("NewStripeLikeProvider() error = %v", err)
	}
	_, err = stripeProvider.Capture(context.Background(), provider.CaptureRequest{
		PaymentID:         "pay_123",
		ProviderIntentID:  "pi_123",
		ProviderPaymentID: "ch_123",
		Amount:            provider.Money{AmountMinor: 1000, Currency: "INR"},
		IdempotencyKey:    "capture:pay_123:1",
	})
	if !errors.Is(err, provider.ErrUnsupportedOperation) {
		t.Fatalf("Capture() error = %v, want unsupported operation", err)
	}
}

func TestStripeLikeRefundSubmitsIdempotentRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/refunds" || r.Header.Get("Idempotency-Key") != "refund:pay_123:item_1" {
			t.Errorf("request = %s key=%q, want refund submission", r.URL.Path, r.Header.Get("Idempotency-Key"))
		}
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm() error = %v", err)
		}
		if r.Form.Get("charge") != "ch_123" || r.Form.Get("amount") != "250" || r.Form.Get("metadata[refund_id]") != "rfnd_123" {
			t.Errorf("refund form = %#v", r.Form)
		}
		_, _ = w.Write([]byte(`{"id":"re_123","status":"pending","amount":250,"currency":"inr"}`))
	}))
	defer server.Close()
	p, err := provider.NewStripeLikeProvider(provider.ProviderConfig{PublicKey: "pk_test", SecretKey: "sk_test", WebhookSecret: "whsec_test", BaseURL: server.URL}, nil)
	if err != nil {
		t.Fatalf("NewStripeLikeProvider() error = %v", err)
	}
	res, err := p.Refund(context.Background(), validRefundRequest())
	if err != nil {
		t.Fatalf("Refund() error = %v", err)
	}
	if res.ProviderRefundID != "re_123" || res.Status != provider.RefundStatusProcessing {
		t.Fatalf("refund response = %+v", res)
	}
}

func TestRazorpayLikeRefundSubmitsPaymentRefund(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/payments/ch_123/refund" || r.Header.Get("X-Razorpay-Idempotency-Key") == "" {
			t.Errorf("request = %s key=%q, want payment refund", r.URL.Path, r.Header.Get("X-Razorpay-Idempotency-Key"))
		}
		var got map[string]any
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("Decode() error = %v", err)
		}
		_, _ = w.Write([]byte(`{"id":"rfnd_gateway_1","status":"processed","amount":250,"currency":"INR"}`))
	}))
	defer server.Close()
	p, err := provider.NewRazorpayLikeProvider(provider.ProviderConfig{PublicKey: "rzp_key", SecretKey: "rzp_secret", WebhookSecret: "rzp_whsec", BaseURL: server.URL}, nil)
	if err != nil {
		t.Fatalf("NewRazorpayLikeProvider() error = %v", err)
	}
	res, err := p.Refund(context.Background(), validRefundRequest())
	if err != nil {
		t.Fatalf("Refund() error = %v", err)
	}
	if res.Status != provider.RefundStatusSucceeded {
		t.Fatalf("refund response = %+v, want succeeded", res)
	}
}

func validCreateIntentRequest() provider.CreateIntentRequest {
	return provider.CreateIntentRequest{
		PaymentID:      "pay_123",
		OrderID:        "ord_123",
		UserID:         "usr_123",
		Amount:         provider.Money{AmountMinor: 1000, Currency: "INR"},
		Customer:       provider.Customer{UserID: "usr_123", Email: "buyer@example.com"},
		IdempotencyKey: "payment_intent:ord_123:1",
		CaptureMode:    provider.CaptureModeAutomatic,
		Metadata:       provider.BuildSafeMetadata("pay_123", "ord_123", "usr_123", "req_123", "payment_intent:ord_123:1"),
	}
}

func validRefundRequest() provider.RefundRequest {
	return provider.RefundRequest{
		PaymentID: "pay_123", RefundID: "rfnd_123", ProviderPaymentID: "ch_123",
		Amount: provider.Money{AmountMinor: 250, Currency: "INR"}, Reason: "Cancelled item",
		IdempotencyKey: "refund:pay_123:item_1",
	}
}
