package config

import (
	"log/slog"
	"testing"
	"time"
)

func TestLoadUsesWishlistDefaults(t *testing.T) {
	t.Setenv("SERVICE_NAME", "")
	t.Setenv("APP_ENV", "")
	t.Setenv("WISHLIST_HTTP_ADDR", "")
	t.Setenv("WISHLIST_MONGO_URI", "")
	t.Setenv("WISHLIST_MONGO_DATABASE", "")
	t.Setenv("WISHLIST_MONGO_COLLECTION", "")
	t.Setenv("WISHLIST_LOG_LEVEL", "")
	t.Setenv("WISHLIST_PRODUCT_SERVICE_BASE_URL", "")
	t.Setenv("PRODUCT_SERVICE_BASE_URL", "")
	t.Setenv("WISHLIST_PRODUCT_SERVICE_TIMEOUT", "")
	t.Setenv("WISHLIST_PRODUCT_CALL_TIMEOUT_MS", "")
	t.Setenv("WISHLIST_CART_SERVICE_BASE_URL", "")
	t.Setenv("CART_SERVICE_BASE_URL", "")
	t.Setenv("WISHLIST_CART_SERVICE_TIMEOUT", "")
	t.Setenv("WISHLIST_CART_CALL_TIMEOUT_MS", "")
	t.Setenv("WISHLIST_AUTH_USER_ID_HEADER", "")
	t.Setenv("WISHLIST_AUTH_ROLES_HEADER", "")
	t.Setenv("WISHLIST_REQUEST_ID_HEADER", "")
	t.Setenv("WISHLIST_EVENTS_BACKEND", "")
	t.Setenv("WISHLIST_KAFKA_BROKERS", "")
	t.Setenv("WISHLIST_PRODUCT_EVENTS_TOPIC", "")
	t.Setenv("WISHLIST_PRODUCT_EVENTS_GROUP_ID", "")
	t.Setenv("WISHLIST_PRICE_EVENTS_GROUP_ID", "")
	t.Setenv("WISHLIST_PRODUCT_EVENTS_DLQ_TOPIC", "")
	t.Setenv("WISHLIST_PRICE_EVENTS_DLQ_TOPIC", "")
	t.Setenv("WISHLIST_PRODUCT_EVENTS_MAX_ATTEMPTS", "")
	t.Setenv("WISHLIST_PRODUCT_EVENTS_RETRY_BACKOFF", "")
	t.Setenv("WISHLIST_NOTIFICATION_COMMANDS_TOPIC", "")
	t.Setenv("WISHLIST_PRICE_DROP_TEMPLATE_KEY", "")
	t.Setenv("WISHLIST_PRICE_DROP_BATCH_SIZE", "")
	t.Setenv("WISHLIST_PRICE_DROP_MIN_DELTA_AMOUNT", "")
	t.Setenv("WISHLIST_ANALYTICS_EVENTS_ENABLED", "")
	t.Setenv("WISHLIST_EVENT_OUTBOX_COLLECTION", "")
	t.Setenv("WISHLIST_EVENT_TOPIC", "")
	t.Setenv("WISHLIST_EVENT_PUBLISHER", "")
	t.Setenv("WISHLIST_EVENT_POLL_INTERVAL", "")
	t.Setenv("WISHLIST_EVENT_BATCH_SIZE", "")
	t.Setenv("WISHLIST_EVENT_MAX_ATTEMPTS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.ServiceName != DefaultServiceName {
		t.Fatalf("ServiceName = %q, want %q", cfg.ServiceName, DefaultServiceName)
	}
	if cfg.HTTP.Addr != DefaultHTTPAddr {
		t.Fatalf("HTTP.Addr = %q, want %q", cfg.HTTP.Addr, DefaultHTTPAddr)
	}
	if cfg.Mongo.Database != DefaultMongoDatabase {
		t.Fatalf("Mongo.Database = %q, want %q", cfg.Mongo.Database, DefaultMongoDatabase)
	}
	if cfg.Mongo.Collection != DefaultMongoCollection {
		t.Fatalf("Mongo.Collection = %q, want %q", cfg.Mongo.Collection, DefaultMongoCollection)
	}
	if cfg.ProductService.BaseURL != DefaultProductBaseURL {
		t.Fatalf("ProductService.BaseURL = %q, want %q", cfg.ProductService.BaseURL, DefaultProductBaseURL)
	}
	if cfg.CartService.BaseURL != DefaultCartBaseURL {
		t.Fatalf("CartService.BaseURL = %q, want %q", cfg.CartService.BaseURL, DefaultCartBaseURL)
	}
	if cfg.Auth.UserIDHeader != DefaultUserIDHeader {
		t.Fatalf("Auth.UserIDHeader = %q, want %q", cfg.Auth.UserIDHeader, DefaultUserIDHeader)
	}
	if !cfg.Auth.RequireBuyerRole {
		t.Fatal("Auth.RequireBuyerRole = false, want true")
	}
	if cfg.Events.Backend != EventsBackendDisabled {
		t.Fatalf("Events.Backend = %q, want %q", cfg.Events.Backend, EventsBackendDisabled)
	}
	if len(cfg.Events.Kafka.Brokers) != 1 || cfg.Events.Kafka.Brokers[0] != DefaultKafkaBrokers {
		t.Fatalf("Kafka.Brokers = %#v, want %q", cfg.Events.Kafka.Brokers, DefaultKafkaBrokers)
	}
	if cfg.Events.ProductEvents.Topic != DefaultProductEventsTopic {
		t.Fatalf("ProductEvents.Topic = %q, want %q", cfg.Events.ProductEvents.Topic, DefaultProductEventsTopic)
	}
	if cfg.Events.ProductEvents.DLQTopic != DefaultProductEventsDLQTopic {
		t.Fatalf("ProductEvents.DLQTopic = %q, want %q", cfg.Events.ProductEvents.DLQTopic, DefaultProductEventsDLQTopic)
	}
	if cfg.Events.NotificationCommands.Topic != DefaultNotificationCommandsTopic {
		t.Fatalf("NotificationCommands.Topic = %q, want %q", cfg.Events.NotificationCommands.Topic, DefaultNotificationCommandsTopic)
	}
	if cfg.Events.PriceDrop.TemplateKey != DefaultPriceDropTemplateKey {
		t.Fatalf("PriceDrop.TemplateKey = %q, want %q", cfg.Events.PriceDrop.TemplateKey, DefaultPriceDropTemplateKey)
	}
	if cfg.Events.PriceDrop.BatchSize != DefaultPriceDropBatchSize {
		t.Fatalf("PriceDrop.BatchSize = %d, want %d", cfg.Events.PriceDrop.BatchSize, DefaultPriceDropBatchSize)
	}
	if cfg.Events.PriceDrop.MinDeltaAmount != DefaultPriceDropMinDeltaAmount {
		t.Fatalf("PriceDrop.MinDeltaAmount = %d, want %d", cfg.Events.PriceDrop.MinDeltaAmount, DefaultPriceDropMinDeltaAmount)
	}
	if !cfg.Events.Analytics.Enabled {
		t.Fatal("Analytics.Enabled = false, want true")
	}
	if cfg.Events.Analytics.OutboxCollection != DefaultWishlistEventCollection {
		t.Fatalf("Analytics.OutboxCollection = %q, want %q", cfg.Events.Analytics.OutboxCollection, DefaultWishlistEventCollection)
	}
	if cfg.Events.Analytics.Topic != DefaultWishlistEventTopic {
		t.Fatalf("Analytics.Topic = %q, want %q", cfg.Events.Analytics.Topic, DefaultWishlistEventTopic)
	}
	if cfg.Events.Analytics.Publisher != DefaultWishlistEventPublisher {
		t.Fatalf("Analytics.Publisher = %q, want %q", cfg.Events.Analytics.Publisher, DefaultWishlistEventPublisher)
	}
	if cfg.Events.Analytics.PollInterval != DefaultWishlistEventPollInterval {
		t.Fatalf("Analytics.PollInterval = %v, want %v", cfg.Events.Analytics.PollInterval, DefaultWishlistEventPollInterval)
	}
	if cfg.Events.Analytics.BatchSize != DefaultWishlistEventBatchSize {
		t.Fatalf("Analytics.BatchSize = %d, want %d", cfg.Events.Analytics.BatchSize, DefaultWishlistEventBatchSize)
	}
	if cfg.Events.Analytics.MaxAttempts != DefaultWishlistEventMaxAttempts {
		t.Fatalf("Analytics.MaxAttempts = %d, want %d", cfg.Events.Analytics.MaxAttempts, DefaultWishlistEventMaxAttempts)
	}
}

func TestLoadParsesDurationsAndLogLevel(t *testing.T) {
	t.Setenv("WISHLIST_LOG_LEVEL", "debug")
	t.Setenv("WISHLIST_MONGO_CONNECT_TIMEOUT", "250")
	t.Setenv("WISHLIST_MONGO_SERVER_SELECTION_TIMEOUT", "3s")
	t.Setenv("WISHLIST_PRODUCT_SERVICE_TIMEOUT", "")
	t.Setenv("WISHLIST_PRODUCT_CALL_TIMEOUT_MS", "750")
	t.Setenv("WISHLIST_CART_SERVICE_TIMEOUT", "1250")
	t.Setenv("WISHLIST_CART_CALL_TIMEOUT_MS", "")
	t.Setenv("WISHLIST_AUTH_REQUIRE_BUYER_ROLE", "false")
	t.Setenv("WISHLIST_EVENTS_BACKEND", "kafka")
	t.Setenv("WISHLIST_KAFKA_BROKERS", "localhost:9092, localhost:9093")
	t.Setenv("WISHLIST_PRODUCT_EVENTS_TOPIC", "product.events")
	t.Setenv("WISHLIST_PRICE_EVENTS_GROUP_ID", "wishlist-price-drop-test")
	t.Setenv("WISHLIST_PRICE_EVENTS_DLQ_TOPIC", "product.events.wishlist.price.dlq")
	t.Setenv("WISHLIST_PRODUCT_EVENTS_MAX_ATTEMPTS", "5")
	t.Setenv("WISHLIST_PRODUCT_EVENTS_RETRY_BACKOFF", "2s")
	t.Setenv("WISHLIST_NOTIFICATION_COMMANDS_TOPIC", "notification.commands.test")
	t.Setenv("WISHLIST_PRICE_DROP_TEMPLATE_KEY", "wishlist_price_drop_test")
	t.Setenv("WISHLIST_PRICE_DROP_BATCH_SIZE", "250")
	t.Setenv("WISHLIST_PRICE_DROP_MIN_DELTA_AMOUNT", "100")
	t.Setenv("WISHLIST_ANALYTICS_EVENTS_ENABLED", "true")
	t.Setenv("WISHLIST_EVENT_OUTBOX_COLLECTION", "wishlist_events_test")
	t.Setenv("WISHLIST_EVENT_TOPIC", "recommendation.events.test")
	t.Setenv("WISHLIST_EVENT_PUBLISHER", "kafka")
	t.Setenv("WISHLIST_EVENT_POLL_INTERVAL", "750ms")
	t.Setenv("WISHLIST_EVENT_BATCH_SIZE", "25")
	t.Setenv("WISHLIST_EVENT_MAX_ATTEMPTS", "7")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Fatalf("LogLevel = %v, want debug", cfg.LogLevel)
	}
	if cfg.Mongo.ConnectTimeout != 250*time.Millisecond {
		t.Fatalf("ConnectTimeout = %v, want 250ms", cfg.Mongo.ConnectTimeout)
	}
	if cfg.Mongo.ServerSelectionTimeout != 3*time.Second {
		t.Fatalf("ServerSelectionTimeout = %v, want 3s", cfg.Mongo.ServerSelectionTimeout)
	}
	if cfg.ProductService.Timeout != 750*time.Millisecond {
		t.Fatalf("ProductService.Timeout = %v, want 750ms", cfg.ProductService.Timeout)
	}
	if cfg.CartService.Timeout != 1250*time.Millisecond {
		t.Fatalf("CartService.Timeout = %v, want 1250ms", cfg.CartService.Timeout)
	}
	if cfg.Auth.RequireBuyerRole {
		t.Fatal("Auth.RequireBuyerRole = true, want false")
	}
	if cfg.Events.Backend != EventsBackendKafka {
		t.Fatalf("Events.Backend = %q, want kafka", cfg.Events.Backend)
	}
	if len(cfg.Events.Kafka.Brokers) != 2 || cfg.Events.Kafka.Brokers[0] != "localhost:9092" || cfg.Events.Kafka.Brokers[1] != "localhost:9093" {
		t.Fatalf("Kafka.Brokers = %#v, want two trimmed brokers", cfg.Events.Kafka.Brokers)
	}
	if cfg.Events.ProductEvents.GroupID != "wishlist-price-drop-test" {
		t.Fatalf("ProductEvents.GroupID = %q, want wishlist-price-drop-test", cfg.Events.ProductEvents.GroupID)
	}
	if cfg.Events.ProductEvents.DLQTopic != "product.events.wishlist.price.dlq" {
		t.Fatalf("ProductEvents.DLQTopic = %q, want product.events.wishlist.price.dlq", cfg.Events.ProductEvents.DLQTopic)
	}
	if cfg.Events.ProductEvents.MaxAttempts != 5 {
		t.Fatalf("ProductEvents.MaxAttempts = %d, want 5", cfg.Events.ProductEvents.MaxAttempts)
	}
	if cfg.Events.ProductEvents.RetryBackoff != 2*time.Second {
		t.Fatalf("ProductEvents.RetryBackoff = %v, want 2s", cfg.Events.ProductEvents.RetryBackoff)
	}
	if cfg.Events.NotificationCommands.Topic != "notification.commands.test" {
		t.Fatalf("NotificationCommands.Topic = %q, want notification.commands.test", cfg.Events.NotificationCommands.Topic)
	}
	if cfg.Events.PriceDrop.TemplateKey != "wishlist_price_drop_test" {
		t.Fatalf("PriceDrop.TemplateKey = %q, want wishlist_price_drop_test", cfg.Events.PriceDrop.TemplateKey)
	}
	if cfg.Events.PriceDrop.BatchSize != 250 {
		t.Fatalf("PriceDrop.BatchSize = %d, want 250", cfg.Events.PriceDrop.BatchSize)
	}
	if cfg.Events.PriceDrop.MinDeltaAmount != 100 {
		t.Fatalf("PriceDrop.MinDeltaAmount = %d, want 100", cfg.Events.PriceDrop.MinDeltaAmount)
	}
	if cfg.Events.Analytics.OutboxCollection != "wishlist_events_test" {
		t.Fatalf("Analytics.OutboxCollection = %q, want wishlist_events_test", cfg.Events.Analytics.OutboxCollection)
	}
	if cfg.Events.Analytics.Topic != "recommendation.events.test" {
		t.Fatalf("Analytics.Topic = %q, want recommendation.events.test", cfg.Events.Analytics.Topic)
	}
	if cfg.Events.Analytics.PollInterval != 750*time.Millisecond {
		t.Fatalf("Analytics.PollInterval = %v, want 750ms", cfg.Events.Analytics.PollInterval)
	}
	if cfg.Events.Analytics.BatchSize != 25 || cfg.Events.Analytics.MaxAttempts != 7 {
		t.Fatalf("Analytics batch/max = %d/%d, want 25/7", cfg.Events.Analytics.BatchSize, cfg.Events.Analytics.MaxAttempts)
	}
}
