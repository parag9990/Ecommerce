package config

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
)

func TestLoadFromEnv(t *testing.T) {
	t.Parallel()

	values := mongoEnvironment()
	values["NOTIFICATION_EMAIL_ENABLED"] = "true"
	values["NOTIFICATION_EMAIL_PROVIDER"] = " mailpit "
	values["NOTIFICATION_EMAIL_SMTP_HOST"] = "localhost"
	values["NOTIFICATION_EMAIL_SMTP_PORT"] = "1025"
	values["NOTIFICATION_EMAIL_FROM"] = "no-reply@example.com"
	values["NOTIFICATION_EMAIL_SMTP_TLS_MODE"] = "none"
	values["NOTIFICATION_SMS_ENABLED"] = "false"
	values["NOTIFICATION_PUSH_ENABLED"] = "true"
	values["NOTIFICATION_PUSH_PROVIDER"] = "firebase"
	values["NOTIFICATION_WHATSAPP_LIKE_ENABLED"] = "false"
	cfg, err := LoadFromEnv(mapLookup(values))
	if err != nil {
		t.Fatalf("LoadFromEnv() returned error: %v", err)
	}
	if cfg.Mongo.Database != defaultMongoDatabase ||
		cfg.Mongo.TemplatesCollection != defaultMongoTemplatesCollection ||
		cfg.Mongo.DeliveriesCollection != defaultMongoDeliveriesCollection ||
		cfg.Mongo.PreferencesCollection != defaultMongoPreferencesCollection ||
		cfg.Mongo.ProviderEventsCollection != defaultMongoProviderEventsCollection {
		t.Fatalf("mongo defaults = %+v", cfg.Mongo)
	}
	if cfg.GRPC.Address != defaultGRPCAddress || cfg.GRPC.StartupTimeout != defaultStartupTimeout {
		t.Fatalf("grpc defaults = %+v", cfg.GRPC)
	}
	if !cfg.ChannelEnabled(domain.ChannelEmail) || cfg.ProviderName(domain.ChannelEmail) != "mailpit" {
		t.Fatalf("email config = %+v, want enabled mailpit", cfg.Email)
	}
	if cfg.ChannelEnabled(domain.ChannelSMS) {
		t.Fatal("SMS channel should be disabled")
	}
	if !cfg.ChannelEnabled(domain.ChannelPush) || cfg.ProviderName(domain.ChannelPush) != "firebase" {
		t.Fatalf("push config = %+v, want enabled firebase", cfg.Push)
	}
}

func TestLoadFromEnvRejectsMissingMongoURI(t *testing.T) {
	t.Parallel()

	_, err := LoadFromEnv(func(string) (string, bool) { return "", false })
	if err == nil {
		t.Fatal("LoadFromEnv() returned nil error without Mongo URI")
	}
}

func TestLoadFromEnvRejectsInvalidMongoCollection(t *testing.T) {
	t.Parallel()

	values := mongoEnvironment()
	values["NOTIFICATION_MONGO_TEMPLATES_COLLECTION"] = "system.templates"
	_, err := LoadFromEnv(mapLookup(values))
	if err == nil {
		t.Fatal("LoadFromEnv() returned nil error with reserved Mongo collection")
	}
}

func TestLoadFromEnvSupportsMongoCollectionOverrides(t *testing.T) {
	t.Parallel()

	values := map[string]string{
		"NOTIFICATION_MONGO_URI":                        "mongodb://localhost:27017/alerts",
		"NOTIFICATION_MONGO_DATABASE":                   "alerts",
		"NOTIFICATION_MONGO_TEMPLATES_COLLECTION":       "templates",
		"NOTIFICATION_MONGO_DELIVERIES_COLLECTION":      "deliveries",
		"NOTIFICATION_MONGO_PREFERENCES_COLLECTION":     "preferences",
		"NOTIFICATION_MONGO_PROVIDER_EVENTS_COLLECTION": "provider_events",
	}
	cfg, err := LoadFromEnv(mapLookup(values))
	if err != nil {
		t.Fatalf("LoadFromEnv() returned error: %v", err)
	}
	if cfg.Mongo.Database != "alerts" || cfg.Mongo.TemplatesCollection != "templates" ||
		cfg.Mongo.DeliveriesCollection != "deliveries" || cfg.Mongo.PreferencesCollection != "preferences" ||
		cfg.Mongo.ProviderEventsCollection != "provider_events" {
		t.Fatalf("mongo config = %+v", cfg.Mongo)
	}
}

func TestLoadFromEnvRejectsDuplicateMongoCollections(t *testing.T) {
	t.Parallel()

	values := mongoEnvironment()
	values["NOTIFICATION_MONGO_DELIVERIES_COLLECTION"] = "notification_preferences"
	if _, err := LoadFromEnv(mapLookup(values)); err == nil {
		t.Fatal("LoadFromEnv() accepted duplicate Mongo collections")
	}
}

func TestLoadFromEnvRejectsEnabledChannelWithoutProvider(t *testing.T) {
	t.Parallel()

	values := mongoEnvironment()
	values["NOTIFICATION_SMS_ENABLED"] = "true"
	_, err := LoadFromEnv(mapLookup(values))
	if err == nil {
		t.Fatal("LoadFromEnv() returned nil error for enabled channel without provider")
	}
}

func TestLoadFromEnvLoadsEnabledOTPProviderSettings(t *testing.T) {
	t.Parallel()

	values := mongoEnvironment()
	values["NOTIFICATION_EMAIL_ENABLED"] = "true"
	values["NOTIFICATION_EMAIL_PROVIDER"] = "smtp"
	values["NOTIFICATION_EMAIL_SMTP_HOST"] = "smtp.example.com"
	values["NOTIFICATION_EMAIL_SMTP_PORT"] = "465"
	values["NOTIFICATION_EMAIL_FROM"] = "no-reply@example.com"
	values["NOTIFICATION_EMAIL_SMTP_TLS_MODE"] = "tls"
	values["NOTIFICATION_EMAIL_TIMEOUT"] = "4s"
	values["NOTIFICATION_SMS_ENABLED"] = "true"
	values["NOTIFICATION_SMS_PROVIDER"] = "gateway"
	values["NOTIFICATION_SMS_HTTP_ENDPOINT"] = "https://sms.example.com/send"
	values["NOTIFICATION_SMS_HTTP_BEARER_TOKEN"] = "from-secret-manager"
	values["NOTIFICATION_SMS_TIMEOUT"] = "3s"

	cfg, err := LoadFromEnv(mapLookup(values))
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}
	if cfg.EmailSMTP.Port != 465 || cfg.EmailSMTP.Timeout != 4*time.Second ||
		cfg.SMSHTTP.Endpoint != "https://sms.example.com/send" || cfg.SMSHTTP.Timeout != 3*time.Second {
		t.Fatalf("provider settings = SMTP %+v SMS %+v", cfg.EmailSMTP, cfg.SMSHTTP)
	}
}

func TestLoadFromEnvRejectsInsecureRemoteSMSOrPlaintextSMTPCredentials(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		values map[string]string
	}{
		{
			name: "remote SMS without HTTPS",
			values: map[string]string{
				"NOTIFICATION_SMS_ENABLED":       "true",
				"NOTIFICATION_SMS_PROVIDER":      "gateway",
				"NOTIFICATION_SMS_HTTP_ENDPOINT": "http://sms.example.com/send",
			},
		},
		{
			name: "SMTP credentials without TLS",
			values: map[string]string{
				"NOTIFICATION_EMAIL_ENABLED":       "true",
				"NOTIFICATION_EMAIL_PROVIDER":      "smtp",
				"NOTIFICATION_EMAIL_SMTP_HOST":     "localhost",
				"NOTIFICATION_EMAIL_SMTP_PORT":     "1025",
				"NOTIFICATION_EMAIL_FROM":          "no-reply@example.com",
				"NOTIFICATION_EMAIL_SMTP_TLS_MODE": "none",
				"NOTIFICATION_EMAIL_SMTP_USERNAME": "user",
				"NOTIFICATION_EMAIL_SMTP_PASSWORD": "password",
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			values := mongoEnvironment()
			for key, value := range test.values {
				values[key] = value
			}
			if _, err := LoadFromEnv(mapLookup(values)); err == nil {
				t.Fatal("LoadFromEnv() returned nil error")
			}
		})
	}
}

func TestLoadFromEnvRejectsInvalidBoolean(t *testing.T) {
	t.Parallel()

	values := mongoEnvironment()
	values["NOTIFICATION_EMAIL_ENABLED"] = "sometimes"
	_, err := LoadFromEnv(mapLookup(values))
	if err == nil {
		t.Fatal("LoadFromEnv() returned nil error for invalid enabled value")
	}
}

func TestLoadFromEnvLoadsEnabledEventConsumer(t *testing.T) {
	t.Parallel()

	values := mongoEnvironment()
	values["NOTIFICATION_EVENT_CONSUMER_ENABLED"] = "true"
	values["NOTIFICATION_RABBITMQ_URL"] = "amqp://guest:guest@localhost:5672/ecommerce"
	values["NOTIFICATION_CONSUMER_PREFETCH"] = "25"
	values["NOTIFICATION_EMAIL_ENABLED"] = "true"
	values["NOTIFICATION_EMAIL_PROVIDER"] = "smtp"
	values["NOTIFICATION_EMAIL_SMTP_HOST"] = "localhost"
	values["NOTIFICATION_EMAIL_SMTP_PORT"] = "1025"
	values["NOTIFICATION_EMAIL_FROM"] = "no-reply@example.com"
	values["NOTIFICATION_EMAIL_SMTP_TLS_MODE"] = "none"
	values["NOTIFICATION_DELIVERY_ENCRYPTION_KEY"] = deliveryEncryptionKey()
	cfg, err := LoadFromEnv(mapLookup(values))
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}
	if !cfg.RabbitMQ.Enabled || cfg.RabbitMQ.Prefetch != 25 ||
		cfg.RabbitMQ.OrderEventsQueue != defaultOrderEventsQueue {
		t.Fatalf("RabbitMQ config = %+v", cfg.RabbitMQ)
	}
	if cfg.Retry.MaxAttempts != 4 || len(cfg.Retry.Delays) != 3 ||
		cfg.Retry.Delays[0] != 30*time.Second {
		t.Fatalf("retry config = %+v", cfg.Retry)
	}
}

func TestLoadFromEnvRejectsIncompleteEnabledEventConsumer(t *testing.T) {
	t.Parallel()

	tests := []map[string]string{
		{"NOTIFICATION_EVENT_CONSUMER_ENABLED": "true"},
		{"NOTIFICATION_EVENT_CONSUMER_ENABLED": "true", "NOTIFICATION_RABBITMQ_URL": "http://localhost:5672"},
	}
	for _, additions := range tests {
		values := mongoEnvironment()
		for key, value := range additions {
			values[key] = value
		}
		if _, err := LoadFromEnv(mapLookup(values)); err == nil {
			t.Fatal("LoadFromEnv() returned nil error for incomplete event consumer configuration")
		}
	}
}

func TestLoadFromEnvValidatesRetryConfiguration(t *testing.T) {
	t.Parallel()

	values := mongoEnvironment()
	values["NOTIFICATION_EVENT_CONSUMER_ENABLED"] = "true"
	values["NOTIFICATION_RABBITMQ_URL"] = "amqp://guest:guest@localhost:5672/ecommerce"
	values["NOTIFICATION_EMAIL_ENABLED"] = "true"
	values["NOTIFICATION_EMAIL_PROVIDER"] = "smtp"
	values["NOTIFICATION_EMAIL_SMTP_HOST"] = "localhost"
	values["NOTIFICATION_EMAIL_SMTP_PORT"] = "1025"
	values["NOTIFICATION_EMAIL_FROM"] = "no-reply@example.com"
	values["NOTIFICATION_EMAIL_SMTP_TLS_MODE"] = "none"
	values["NOTIFICATION_DELIVERY_ENCRYPTION_KEY"] = deliveryEncryptionKey()
	values["NOTIFICATION_RETRY_MAX_ATTEMPTS"] = "3"
	values["NOTIFICATION_RETRY_DELAY_1"] = "10s"
	values["NOTIFICATION_RETRY_DELAY_2"] = "1m"
	values["NOTIFICATION_RETRY_PREFETCH"] = "7"

	cfg, err := LoadFromEnv(mapLookup(values))
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}
	if cfg.Retry.MaxAttempts != 3 || cfg.Retry.Prefetch != 7 ||
		cfg.Retry.Delays[1] != time.Minute {
		t.Fatalf("retry config = %+v", cfg.Retry)
	}
	values["NOTIFICATION_RETRY_DELAY_2"] = "5s"
	if _, err := LoadFromEnv(mapLookup(values)); err == nil {
		t.Fatal("LoadFromEnv() accepted non-increasing retry delays")
	}
}

func TestLoadFromEnvLoadsAnalyticsWebhookConfiguration(t *testing.T) {
	t.Parallel()

	values := mongoEnvironment()
	values["NOTIFICATION_EMAIL_ENABLED"] = "true"
	values["NOTIFICATION_EMAIL_PROVIDER"] = "smtp"
	values["NOTIFICATION_EMAIL_SMTP_HOST"] = "localhost"
	values["NOTIFICATION_EMAIL_SMTP_PORT"] = "1025"
	values["NOTIFICATION_EMAIL_FROM"] = "no-reply@example.com"
	values["NOTIFICATION_EMAIL_SMTP_TLS_MODE"] = "none"
	values["NOTIFICATION_METRICS_ENABLED"] = "true"
	values["NOTIFICATION_PROVIDER_WEBHOOKS_ENABLED"] = "true"
	values["NOTIFICATION_EMAIL_WEBHOOK_SIGNING_SECRET"] = "from-secret-manager"
	values["NOTIFICATION_EMAIL_OPEN_TRACKING_ENABLED"] = "true"
	values["NOTIFICATION_WEBHOOK_REPLAY_WINDOW_SECONDS"] = "120"

	cfg, err := LoadFromEnv(mapLookup(values))
	if err != nil {
		t.Fatalf("LoadFromEnv() analytics error = %v", err)
	}
	if !cfg.Analytics.MetricsEnabled || !cfg.Analytics.WebhooksEnabled ||
		cfg.Analytics.SigningSecret(domain.ChannelEmail) != "from-secret-manager" ||
		!cfg.Analytics.AllowsOpenTracking(domain.ChannelEmail) ||
		cfg.Analytics.WebhookReplayWindow != 2*time.Minute {
		t.Fatalf("analytics config = %+v", cfg.Analytics)
	}
}

func TestLoadFromEnvRejectsUnsafeAnalyticsConfiguration(t *testing.T) {
	t.Parallel()

	tests := []map[string]string{
		{"NOTIFICATION_PROVIDER_WEBHOOKS_ENABLED": "true"},
		{"NOTIFICATION_EMAIL_OPEN_TRACKING_ENABLED": "true"},
		{"NOTIFICATION_SMS_OPEN_TRACKING_ENABLED": "true", "NOTIFICATION_PROVIDER_WEBHOOKS_ENABLED": "true"},
	}
	for _, addition := range tests {
		values := mongoEnvironment()
		for key, value := range addition {
			values[key] = value
		}
		if _, err := LoadFromEnv(mapLookup(values)); err == nil {
			t.Fatalf("LoadFromEnv() accepted unsafe analytics settings: %+v", addition)
		}
	}
}

func deliveryEncryptionKey() string {
	return base64.StdEncoding.EncodeToString([]byte("01234567890123456789012345678901"))
}

func mongoEnvironment() map[string]string {
	return map[string]string{
		"NOTIFICATION_MONGO_URI": "mongodb://localhost:27017/notification_db",
	}
}

func mapLookup(values map[string]string) LookupEnv {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
