package config

import (
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider"
)

func TestLoadAllowsNoProviderByDefault(t *testing.T) {
	clearProviderEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.PaymentGateway.Normalized().DefaultProvider != "" {
		t.Fatalf("DefaultProvider = %q, want empty", cfg.PaymentGateway.DefaultProvider)
	}
	if len(cfg.PaymentGateway.AllowedProviders) != 0 {
		t.Fatalf("AllowedProviders = %#v, want empty", cfg.PaymentGateway.AllowedProviders)
	}
	if cfg.Reconciliation.Enabled {
		t.Fatal("Reconciliation.Enabled = true, want disabled by default")
	}
}

func TestLoadPaymentGatewayProviderConfig(t *testing.T) {
	clearProviderEnv(t)
	t.Setenv("PAYMENT_DEFAULT_PROVIDER", provider.ProviderNameStripeLike)
	t.Setenv("PAYMENT_ALLOWED_PROVIDERS", "stripe_like")
	t.Setenv("PAYMENT_ALLOWED_CURRENCIES", "INR,usd")
	t.Setenv("PAYMENT_CAPTURE_MODE", "manual")
	t.Setenv("PAYMENT_PROVIDER_TIMEOUT", "2s")
	t.Setenv("STRIPE_LIKE_PUBLIC_KEY", "pk_test")
	t.Setenv("STRIPE_LIKE_SECRET_KEY", "sk_test")
	t.Setenv("STRIPE_LIKE_WEBHOOK_SECRET", "whsec_test")
	t.Setenv("PAYMENT_INTERNAL_API_TOKEN", "internal-payment-token-at-least-32-characters")
	t.Setenv("PAYMENT_EVENTS_ENDPOINT", "http://127.0.0.1:9080/events")
	t.Setenv("PAYMENT_EVENTS_AUTH_TOKEN", "event-publisher-token-at-least-32-characters")
	t.Setenv("PAYMENT_REFUND_MANUAL_REVIEW_THRESHOLD_MINOR", "50000")
	t.Setenv("PAYMENT_RETRY_MAX_ATTEMPTS", "4")
	t.Setenv("PAYMENT_RETRY_COOLDOWN_SECONDS", "12")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	paymentGateway := cfg.PaymentGateway.Normalized()
	if paymentGateway.DefaultProvider != provider.ProviderNameStripeLike {
		t.Fatalf("DefaultProvider = %q, want stripe_like", paymentGateway.DefaultProvider)
	}
	if paymentGateway.CaptureMode != provider.CaptureModeManual {
		t.Fatalf("CaptureMode = %q, want manual", paymentGateway.CaptureMode)
	}
	if paymentGateway.AllowedCurrencies[1] != "USD" {
		t.Fatalf("AllowedCurrencies = %#v, want uppercase USD", paymentGateway.AllowedCurrencies)
	}
	stripeConfig, ok := paymentGateway.ProviderConfig(provider.ProviderNameStripeLike)
	if !ok {
		t.Fatal("ProviderConfig(stripe_like) missing")
	}
	if stripeConfig.Timeout != 2*time.Second {
		t.Fatalf("Timeout = %s, want 2s", stripeConfig.Timeout)
	}
	if cfg.Refund.ManualReviewThresholdMinor != 50000 {
		t.Fatalf("ManualReviewThresholdMinor = %d, want 50000", cfg.Refund.ManualReviewThresholdMinor)
	}
	if cfg.Retry.MaxAttempts != 4 || cfg.Retry.Cooldown != 12*time.Second {
		t.Fatalf("Retry = %+v, want four attempts and 12-second cooldown", cfg.Retry)
	}
}

func TestLoadRejectsEnabledGatewayWithoutInternalAPIToken(t *testing.T) {
	clearProviderEnv(t)
	t.Setenv("PAYMENT_DEFAULT_PROVIDER", provider.ProviderNameStripeLike)
	t.Setenv("PAYMENT_ALLOWED_PROVIDERS", "stripe_like")
	t.Setenv("STRIPE_LIKE_PUBLIC_KEY", "pk_test")
	t.Setenv("STRIPE_LIKE_SECRET_KEY", "sk_test")
	t.Setenv("STRIPE_LIKE_WEBHOOK_SECRET", "whsec_test")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want internal API token validation error")
	}
}

func TestLoadRejectsEnabledGatewayWithoutEventPublisher(t *testing.T) {
	clearProviderEnv(t)
	t.Setenv("PAYMENT_DEFAULT_PROVIDER", provider.ProviderNameStripeLike)
	t.Setenv("PAYMENT_ALLOWED_PROVIDERS", "stripe_like")
	t.Setenv("STRIPE_LIKE_PUBLIC_KEY", "pk_test")
	t.Setenv("STRIPE_LIKE_SECRET_KEY", "sk_test")
	t.Setenv("STRIPE_LIKE_WEBHOOK_SECRET", "whsec_test")
	t.Setenv("PAYMENT_INTERNAL_API_TOKEN", "internal-payment-token-at-least-32-characters")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want event publisher validation error")
	}
}

func TestLoadRejectsNegativeRefundReviewThreshold(t *testing.T) {
	clearProviderEnv(t)
	t.Setenv("PAYMENT_REFUND_MANUAL_REVIEW_THRESHOLD_MINOR", "-1")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want refund threshold validation error")
	}
}

func TestLoadRejectsInvalidRetryPolicy(t *testing.T) {
	clearProviderEnv(t)
	t.Setenv("PAYMENT_RETRY_MAX_ATTEMPTS", "1")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want retry attempt validation error")
	}
	clearProviderEnv(t)
	t.Setenv("PAYMENT_RETRY_COOLDOWN_SECONDS", "-1")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want retry cooldown validation error")
	}
}

func TestLoadReconciliationConfig(t *testing.T) {
	clearProviderEnv(t)
	t.Setenv("PAYMENT_RECONCILIATION_ENABLED", "true")
	t.Setenv("PAYMENT_RECONCILIATION_PROVIDER", "Stripe_Like")
	t.Setenv("PAYMENT_RECONCILIATION_REPORT_LAG_HOURS", "36")
	t.Setenv("PAYMENT_RECONCILIATION_BATCH_SIZE", "250")
	t.Setenv("PAYMENT_RECONCILIATION_TIMEOUT", "15m")
	t.Setenv("PAYMENT_RECONCILIATION_ALERT_TOPIC", "finance.payment.alerts")
	t.Setenv("PAYMENT_EVENTS_ENDPOINT", "http://127.0.0.1:9080/events")
	t.Setenv("PAYMENT_EVENTS_AUTH_TOKEN", "event-publisher-token-at-least-32-characters")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Reconciliation.Provider != "stripe_like" || cfg.Reconciliation.ReportLag != 36*time.Hour {
		t.Fatalf("Reconciliation = %+v, want normalized provider and 36-hour lag", cfg.Reconciliation)
	}
	if cfg.Reconciliation.BatchSize != 250 || cfg.Reconciliation.Timeout != 15*time.Minute {
		t.Fatalf("Reconciliation = %+v, want configured paging and timeout", cfg.Reconciliation)
	}
}

func TestLoadRejectsEnabledReconciliationWithoutAlertPublisher(t *testing.T) {
	clearProviderEnv(t)
	t.Setenv("PAYMENT_RECONCILIATION_ENABLED", "true")
	t.Setenv("PAYMENT_RECONCILIATION_PROVIDER", "stripe_like")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want reconciliation publisher validation error")
	}
}

func TestLoadAllowsReconciliationProviderToBePassedByCommand(t *testing.T) {
	clearProviderEnv(t)
	t.Setenv("PAYMENT_RECONCILIATION_ENABLED", "true")
	t.Setenv("PAYMENT_EVENTS_ENDPOINT", "http://127.0.0.1:9080/events")
	t.Setenv("PAYMENT_EVENTS_AUTH_TOKEN", "event-publisher-token-at-least-32-characters")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want command to be able to supply provider", err)
	}
	if cfg.Reconciliation.Provider != "" {
		t.Fatalf("Reconciliation.Provider = %q, want empty command default", cfg.Reconciliation.Provider)
	}
}

func clearProviderEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"PAYMENT_DEFAULT_PROVIDER",
		"PAYMENT_ALLOWED_PROVIDERS",
		"PAYMENT_ALLOWED_CURRENCIES",
		"PAYMENT_CAPTURE_MODE",
		"PAYMENT_PROVIDER_TIMEOUT",
		"PAYMENT_WEBHOOK_TIMESTAMP_TOLERANCE",
		"PAYMENT_WEBHOOK_MAX_BODY_BYTES",
		"PAYMENT_EVENTS_ENDPOINT",
		"PAYMENT_EVENTS_AUTH_TOKEN",
		"PAYMENT_EVENTS_TIMEOUT",
		"STRIPE_LIKE_PUBLIC_KEY",
		"STRIPE_LIKE_SECRET_KEY",
		"STRIPE_LIKE_WEBHOOK_SECRET",
		"STRIPE_LIKE_BASE_URL",
		"STRIPE_LIKE_TIMEOUT",
		"RAZORPAY_LIKE_PUBLIC_KEY",
		"RAZORPAY_LIKE_SECRET_KEY",
		"RAZORPAY_LIKE_WEBHOOK_SECRET",
		"RAZORPAY_LIKE_BASE_URL",
		"RAZORPAY_LIKE_TIMEOUT",
		"PAYMENT_INTERNAL_API_TOKEN",
		"PAYMENT_REFUND_MANUAL_REVIEW_THRESHOLD_MINOR",
		"PAYMENT_RETRY_MAX_ATTEMPTS",
		"PAYMENT_RETRY_COOLDOWN_SECONDS",
		"PAYMENT_RECONCILIATION_ENABLED",
		"PAYMENT_RECONCILIATION_PROVIDER",
		"PAYMENT_RECONCILIATION_REPORT_FILE",
		"PAYMENT_RECONCILIATION_REPORT_LAG_HOURS",
		"PAYMENT_RECONCILIATION_BATCH_SIZE",
		"PAYMENT_RECONCILIATION_TIMEOUT",
		"PAYMENT_RECONCILIATION_MAX_REPORT_FILE_BYTES",
		"PAYMENT_RECONCILIATION_ALERT_TOPIC",
	}
	for _, key := range keys {
		t.Setenv(key, "")
	}
}
