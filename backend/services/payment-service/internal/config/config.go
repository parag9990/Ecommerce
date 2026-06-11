package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/events"
	"github.com/example/ecommerce-platform/backend/services/payment-service/internal/provider"
)

const (
	defaultHTTPAddress                = ":8080"
	defaultMySQLDSN                   = "root:root@tcp(127.0.0.1:3306)/payment_db?parseTime=true&charset=utf8mb4,utf8&loc=UTC"
	defaultShutdownTimeout            = 10 * time.Second
	defaultReconciliationReportLag    = 24 * time.Hour
	defaultReconciliationBatchSize    = 500
	defaultReconciliationTimeout      = 20 * time.Minute
	defaultReconciliationMaxFileBytes = 64 << 20
	defaultReconciliationAlertTopic   = "payment.reconciliation.alerts"
)

type Config struct {
	HTTP           HTTPConfig
	Database       DatabaseConfig
	InternalAPI    InternalAPIConfig
	PaymentGateway provider.Config
	Webhook        WebhookConfig
	Refund         RefundConfig
	Retry          RetryConfig
	Reconciliation ReconciliationConfig
	EventPublisher events.HTTPPublisherConfig
}

type HTTPConfig struct {
	Address         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type InternalAPIConfig struct {
	Token string
}

type WebhookConfig struct {
	MaxBodyBytes int64
}

type RefundConfig struct {
	ManualReviewThresholdMinor int64
}

type RetryConfig struct {
	MaxAttempts int
	Cooldown    time.Duration
}

type ReconciliationConfig struct {
	Enabled            bool
	Provider           string
	ReportFile         string
	ReportLag          time.Duration
	BatchSize          int
	Timeout            time.Duration
	MaxReportFileBytes int64
	AlertTopic         string
}

func Load() (Config, error) {
	webhookTolerance := envDuration("PAYMENT_WEBHOOK_TIMESTAMP_TOLERANCE", 5*time.Minute)
	cfg := Config{
		HTTP: HTTPConfig{
			Address:         envString("PAYMENT_HTTP_ADDR", envString("HTTP_ADDR", defaultHTTPAddress)),
			ReadTimeout:     envDuration("PAYMENT_HTTP_READ_TIMEOUT", 5*time.Second),
			WriteTimeout:    envDuration("PAYMENT_HTTP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:     envDuration("PAYMENT_HTTP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: envDuration("PAYMENT_SHUTDOWN_TIMEOUT", defaultShutdownTimeout),
		},
		Database: DatabaseConfig{
			DSN:             envString("PAYMENT_MYSQL_DSN", envString("MYSQL_DSN", defaultMySQLDSN)),
			MaxOpenConns:    envInt("PAYMENT_MYSQL_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    envInt("PAYMENT_MYSQL_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: envDuration("PAYMENT_MYSQL_CONN_MAX_LIFETIME", 30*time.Minute),
			ConnMaxIdleTime: envDuration("PAYMENT_MYSQL_CONN_MAX_IDLE_TIME", 5*time.Minute),
		},
		InternalAPI: InternalAPIConfig{
			Token: envString("PAYMENT_INTERNAL_API_TOKEN", ""),
		},
		Webhook: WebhookConfig{
			MaxBodyBytes: envInt64("PAYMENT_WEBHOOK_MAX_BODY_BYTES", 1<<20),
		},
		Refund: RefundConfig{
			ManualReviewThresholdMinor: envInt64("PAYMENT_REFUND_MANUAL_REVIEW_THRESHOLD_MINOR", 0),
		},
		Retry: RetryConfig{
			MaxAttempts: envInt("PAYMENT_RETRY_MAX_ATTEMPTS", 3),
			Cooldown:    time.Duration(envInt("PAYMENT_RETRY_COOLDOWN_SECONDS", 5)) * time.Second,
		},
		Reconciliation: ReconciliationConfig{
			Enabled:            envBool("PAYMENT_RECONCILIATION_ENABLED", false),
			Provider:           strings.ToLower(strings.TrimSpace(envString("PAYMENT_RECONCILIATION_PROVIDER", ""))),
			ReportFile:         strings.TrimSpace(envString("PAYMENT_RECONCILIATION_REPORT_FILE", "")),
			ReportLag:          envHoursDuration("PAYMENT_RECONCILIATION_REPORT_LAG_HOURS", defaultReconciliationReportLag),
			BatchSize:          envInt("PAYMENT_RECONCILIATION_BATCH_SIZE", defaultReconciliationBatchSize),
			Timeout:            envDuration("PAYMENT_RECONCILIATION_TIMEOUT", defaultReconciliationTimeout),
			MaxReportFileBytes: envInt64("PAYMENT_RECONCILIATION_MAX_REPORT_FILE_BYTES", defaultReconciliationMaxFileBytes),
			AlertTopic:         strings.TrimSpace(envString("PAYMENT_RECONCILIATION_ALERT_TOPIC", defaultReconciliationAlertTopic)),
		},
		EventPublisher: events.HTTPPublisherConfig{
			Endpoint:  envString("PAYMENT_EVENTS_ENDPOINT", ""),
			AuthToken: envString("PAYMENT_EVENTS_AUTH_TOKEN", ""),
			Timeout:   envDuration("PAYMENT_EVENTS_TIMEOUT", 5*time.Second),
		},
		PaymentGateway: provider.Config{
			DefaultProvider:   envString("PAYMENT_DEFAULT_PROVIDER", ""),
			AllowedProviders:  envList("PAYMENT_ALLOWED_PROVIDERS", nil),
			AllowedCurrencies: envList("PAYMENT_ALLOWED_CURRENCIES", []string{"INR", "USD"}),
			CaptureMode:       provider.CaptureMode(envString("PAYMENT_CAPTURE_MODE", string(provider.CaptureModeAutomatic))),
			Providers: map[string]provider.ProviderConfig{
				provider.ProviderNameStripeLike: {
					PublicKey:                 envString("STRIPE_LIKE_PUBLIC_KEY", ""),
					SecretKey:                 envString("STRIPE_LIKE_SECRET_KEY", ""),
					WebhookSecret:             envString("STRIPE_LIKE_WEBHOOK_SECRET", ""),
					BaseURL:                   envString("STRIPE_LIKE_BASE_URL", ""),
					Timeout:                   envDuration("STRIPE_LIKE_TIMEOUT", envDuration("PAYMENT_PROVIDER_TIMEOUT", 5*time.Second)),
					WebhookTimestampTolerance: webhookTolerance,
				},
				provider.ProviderNameRazorpayLike: {
					PublicKey:                 envString("RAZORPAY_LIKE_PUBLIC_KEY", ""),
					SecretKey:                 envString("RAZORPAY_LIKE_SECRET_KEY", ""),
					WebhookSecret:             envString("RAZORPAY_LIKE_WEBHOOK_SECRET", ""),
					BaseURL:                   envString("RAZORPAY_LIKE_BASE_URL", ""),
					Timeout:                   envDuration("RAZORPAY_LIKE_TIMEOUT", envDuration("PAYMENT_PROVIDER_TIMEOUT", 5*time.Second)),
					WebhookTimestampTolerance: webhookTolerance,
				},
			},
		},
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if c.HTTP.Address == "" {
		return errors.New("PAYMENT_HTTP_ADDR cannot be empty")
	}
	if c.HTTP.ReadTimeout <= 0 {
		return errors.New("PAYMENT_HTTP_READ_TIMEOUT must be greater than zero")
	}
	if c.HTTP.WriteTimeout <= 0 {
		return errors.New("PAYMENT_HTTP_WRITE_TIMEOUT must be greater than zero")
	}
	if c.HTTP.IdleTimeout <= 0 {
		return errors.New("PAYMENT_HTTP_IDLE_TIMEOUT must be greater than zero")
	}
	if c.HTTP.ShutdownTimeout <= 0 {
		return errors.New("PAYMENT_SHUTDOWN_TIMEOUT must be greater than zero")
	}
	if c.Database.DSN == "" {
		return errors.New("PAYMENT_MYSQL_DSN cannot be empty")
	}
	if c.Database.MaxOpenConns <= 0 {
		return errors.New("PAYMENT_MYSQL_MAX_OPEN_CONNS must be greater than zero")
	}
	if c.Database.MaxIdleConns < 0 {
		return errors.New("PAYMENT_MYSQL_MAX_IDLE_CONNS cannot be negative")
	}
	if c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		return errors.New("PAYMENT_MYSQL_MAX_IDLE_CONNS cannot exceed PAYMENT_MYSQL_MAX_OPEN_CONNS")
	}
	if c.Database.ConnMaxLifetime <= 0 {
		return errors.New("PAYMENT_MYSQL_CONN_MAX_LIFETIME must be greater than zero")
	}
	if c.Database.ConnMaxIdleTime <= 0 {
		return errors.New("PAYMENT_MYSQL_CONN_MAX_IDLE_TIME must be greater than zero")
	}
	if c.Webhook.MaxBodyBytes <= 0 {
		return errors.New("PAYMENT_WEBHOOK_MAX_BODY_BYTES must be greater than zero")
	}
	if c.Refund.ManualReviewThresholdMinor < 0 {
		return errors.New("PAYMENT_REFUND_MANUAL_REVIEW_THRESHOLD_MINOR cannot be negative")
	}
	if c.Retry.MaxAttempts < 2 {
		return errors.New("PAYMENT_RETRY_MAX_ATTEMPTS must be at least 2")
	}
	if c.Retry.Cooldown < 0 {
		return errors.New("PAYMENT_RETRY_COOLDOWN_SECONDS cannot be negative")
	}
	if c.Reconciliation.ReportLag < 0 {
		return errors.New("PAYMENT_RECONCILIATION_REPORT_LAG_HOURS cannot be negative")
	}
	if c.Reconciliation.BatchSize <= 0 || c.Reconciliation.BatchSize > 10000 {
		return errors.New("PAYMENT_RECONCILIATION_BATCH_SIZE must be between 1 and 10000")
	}
	if c.Reconciliation.Timeout <= 0 {
		return errors.New("PAYMENT_RECONCILIATION_TIMEOUT must be greater than zero")
	}
	if c.Reconciliation.MaxReportFileBytes <= 0 {
		return errors.New("PAYMENT_RECONCILIATION_MAX_REPORT_FILE_BYTES must be greater than zero")
	}
	if c.Reconciliation.Enabled {
		if len(c.Reconciliation.Provider) > 64 {
			return errors.New("PAYMENT_RECONCILIATION_PROVIDER cannot exceed 64 characters")
		}
		if c.Reconciliation.AlertTopic == "" {
			return errors.New("PAYMENT_RECONCILIATION_ALERT_TOPIC is required when reconciliation is enabled")
		}
		if err := c.EventPublisher.Validate(); err != nil {
			return err
		}
	}
	if err := c.PaymentGateway.Validate(); err != nil {
		return err
	}
	if len(c.PaymentGateway.Normalized().AllowedProviders) > 0 {
		internalToken := strings.TrimSpace(c.InternalAPI.Token)
		if internalToken == "" {
			return errors.New("PAYMENT_INTERNAL_API_TOKEN is required when payment providers are enabled")
		}
		if len(internalToken) < 32 {
			return errors.New("PAYMENT_INTERNAL_API_TOKEN must be at least 32 characters")
		}
		if err := c.EventPublisher.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func envString(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err == nil {
		return value
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}

func envHoursDuration(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err == nil {
		return value
	}
	hours, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return time.Duration(hours) * time.Hour
}

func envInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func envInt64(key string, fallback int64) int64 {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fallback
	}
	return value
}

func envBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return value
}

func envList(key string, fallback []string) []string {
	raw := os.Getenv(key)
	if raw == "" {
		return append([]string(nil), fallback...)
	}
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			values = append(values, part)
		}
	}
	return values
}
