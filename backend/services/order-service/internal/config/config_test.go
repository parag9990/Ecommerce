package config

import (
	"testing"
	"time"
)

func TestLoadCheckoutConfig(t *testing.T) {
	t.Setenv("ORDER_PRODUCT_SERVICE_TOKEN", "product-service-token")
	t.Setenv("ORDER_MYSQL_DSN", "user:password@tcp(localhost:3306)/order_db")
	t.Setenv("ORDER_INVENTORY_RESERVATION_TTL", "12m")
	t.Setenv("ORDER_INVENTORY_RELEASE_TIMEOUT", "750ms")
	t.Setenv("ORDER_IDEMPOTENCY_TTL", "36h")
	t.Setenv("ORDER_PAYMENT_RETURN_URL", "https://shop.example.test/checkout/result")
	t.Setenv("ORDER_PAYMENT_ALLOWED_CURRENCIES", "inr, usd, INR")
	t.Setenv("ORDER_PAYMENT_INVENTORY_ACTION_TIMEOUT", "900ms")
	t.Setenv("ORDER_PAYMENT_INTERNAL_TOKEN", "payment-internal-token-at-least-32-chars")
	t.Setenv("ORDER_PAYMENT_EVENTS_TOKEN", "payment-events-token-at-least-32-chars")
	t.Setenv("ORDER_GRPC_TRUSTED_CALLER_TOKEN", "trusted-caller-token-at-least-32-chars")
	t.Setenv("ORDER_PAGE_TOKEN_SIGNING_KEY", "page-token-signing-key-at-least-32-chars")
	t.Setenv("ORDER_KAFKA_BROKERS", "localhost:9092, kafka-b:9092")
	t.Setenv("ORDER_OUTBOX_BATCH_SIZE", "25")
	t.Setenv("ORDER_OUTBOX_INTERVAL", "1500ms")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if config.Checkout.InventoryReservationTTL != 12*time.Minute ||
		config.Checkout.InventoryReleaseTimeout != 750*time.Millisecond ||
		config.Checkout.IdempotencyTTL != 36*time.Hour {
		t.Fatalf("checkout config = %+v, want configured durations", config.Checkout)
	}
	usecaseConfig := config.CreateOrderFromCartConfig()
	if usecaseConfig.ReservationTTL != 12*time.Minute || usecaseConfig.ReleaseTimeout != 750*time.Millisecond {
		t.Fatalf("usecase config = %+v, want configured durations", usecaseConfig)
	}
	if config.CreateOrderConfig().IdempotencyTTL != 36*time.Hour {
		t.Fatalf("create order config = %+v, want configured idempotency TTL", config.CreateOrderConfig())
	}
	initiateConfig := config.InitiateOrderPaymentConfig()
	if initiateConfig.ReturnURL != "https://shop.example.test/checkout/result" ||
		len(initiateConfig.AllowedCurrencies) != 2 ||
		initiateConfig.InventoryReleaseTimeout != 900*time.Millisecond {
		t.Fatalf("payment initiate config = %+v, want configured payment settings", initiateConfig)
	}
	if config.ApplyPaymentResultConfig().InventoryActionTimeout != 900*time.Millisecond {
		t.Fatalf("payment result config = %+v, want configured timeout", config.ApplyPaymentResultConfig())
	}
	if config.GRPC.Address != ":9094" || len(config.PageTokenSigningKey()) < 32 {
		t.Fatalf("grpc config = %+v, want configured secure defaults", config.GRPC)
	}
	if !config.Events.Enabled || config.Events.Topic != "order.events" ||
		len(config.Events.KafkaBrokers) != 2 ||
		config.OrderOutboxWorkerConfig().BatchSize != 25 ||
		config.OrderOutboxWorkerConfig().Interval != 1500*time.Millisecond {
		t.Fatalf("events config = %+v, want enabled kafka outbox defaults with overrides", config.Events)
	}
}

func TestLoadRejectsMissingDatabaseDSNAndInvalidDuration(t *testing.T) {
	t.Setenv("ORDER_PRODUCT_SERVICE_TOKEN", "product-service-token")
	t.Setenv("ORDER_MYSQL_DSN", "")
	t.Setenv("ORDER_PAYMENT_RETURN_URL", "https://shop.example.test/checkout/result")
	t.Setenv("ORDER_GRPC_TRUSTED_CALLER_TOKEN", "trusted-caller-token-at-least-32-chars")
	t.Setenv("ORDER_PAGE_TOKEN_SIGNING_KEY", "page-token-signing-key-at-least-32-chars")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want missing DSN error")
	}

	t.Setenv("ORDER_MYSQL_DSN", "dsn")
	t.Setenv("ORDER_INVENTORY_RESERVATION_TTL", "not-a-duration")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want invalid duration error")
	}

	t.Setenv("ORDER_INVENTORY_RESERVATION_TTL", "10m")
	t.Setenv("ORDER_IDEMPOTENCY_TTL", "0s")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want non-positive idempotency TTL error")
	}

	t.Setenv("ORDER_IDEMPOTENCY_TTL", "24h")
	t.Setenv("ORDER_PAYMENT_RETURN_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want missing payment return URL error")
	}

	t.Setenv("ORDER_PAYMENT_RETURN_URL", "http://shop.example.test/checkout/result")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want non-HTTPS payment return URL error")
	}

	t.Setenv("ORDER_PAYMENT_RETURN_URL", "https://shop.example.test/checkout/result")
	t.Setenv("ORDER_GRPC_TRUSTED_CALLER_TOKEN", "short")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want short trusted caller token error")
	}
}
