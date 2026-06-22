package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	EventsBackendDisabled            = "disabled"
	EventsBackendKafka               = "kafka"
	DefaultServiceName               = "wishlist-service"
	DefaultAppEnv                    = "local"
	DefaultHTTPAddr                  = ":8084"
	DefaultMongoURI                  = "mongodb://localhost:27017"
	DefaultMongoDatabase             = "wishlist_db"
	DefaultMongoCollection           = "wishlists"
	DefaultProductBaseURL            = "http://localhost:8082"
	DefaultCartBaseURL               = "http://localhost:8083"
	DefaultUserIDHeader              = "X-User-ID"
	DefaultRolesHeader               = "X-User-Roles"
	DefaultRequestIDHeader           = "X-Request-ID"
	DefaultEventsBackend             = EventsBackendDisabled
	DefaultKafkaBrokers              = "localhost:9092"
	DefaultProductEventsTopic        = "product.events"
	DefaultProductEventsGroupID      = "wishlist-service"
	DefaultProductEventsDLQTopic     = "product.events.wishlist.dlq"
	DefaultProductEventsMaxAttempts  = 3
	DefaultProductEventsRetryBackoff = 500 * time.Millisecond
	DefaultNotificationCommandsTopic = "notification.commands"
	DefaultPriceDropTemplateKey      = "wishlist_price_drop"
	DefaultPriceDropBatchSize        = 500
	DefaultPriceDropMinDeltaAmount   = int64(1)
	DefaultAnalyticsEventsEnabled    = true
	DefaultWishlistEventCollection   = "wishlist_events"
	DefaultWishlistEventTopic        = "recommendation.events"
	DefaultWishlistEventPublisher    = EventsBackendKafka
	DefaultWishlistEventPollInterval = 5 * time.Second
	DefaultWishlistEventBatchSize    = 50
	DefaultWishlistEventMaxAttempts  = 5
	DefaultWishlistEventClaimLease   = 30 * time.Second
)

type Config struct {
	ServiceName     string
	AppEnv          string
	LogLevel        slog.Level
	HTTP            HTTPConfig
	Mongo           MongoConfig
	ProductService  ProductServiceConfig
	CartService     CartServiceConfig
	Auth            AuthConfig
	Events          EventsConfig
	StartupTimeout  time.Duration
	ShutdownTimeout time.Duration
}

type HTTPConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type MongoConfig struct {
	URI                    string
	Database               string
	Collection             string
	ConnectTimeout         time.Duration
	ServerSelectionTimeout time.Duration
}

type ProductServiceConfig struct {
	BaseURL string
	Timeout time.Duration
}

type CartServiceConfig struct {
	BaseURL string
	Timeout time.Duration
}

type AuthConfig struct {
	UserIDHeader     string
	RolesHeader      string
	RequestIDHeader  string
	RequireBuyerRole bool
}

type EventsConfig struct {
	Backend              string
	Kafka                KafkaConfig
	ProductEvents        ProductEventsConfig
	NotificationCommands NotificationCommandsConfig
	PriceDrop            PriceDropConfig
	Analytics            AnalyticsEventsConfig
}

type KafkaConfig struct {
	Brokers []string
}

type ProductEventsConfig struct {
	Topic        string
	GroupID      string
	DLQTopic     string
	MaxAttempts  int
	RetryBackoff time.Duration
}

type NotificationCommandsConfig struct {
	Topic string
}

type PriceDropConfig struct {
	TemplateKey    string
	BatchSize      int
	MinDeltaAmount int64
}

type AnalyticsEventsConfig struct {
	Enabled          bool
	OutboxCollection string
	Topic            string
	Publisher        string
	PollInterval     time.Duration
	BatchSize        int
	MaxAttempts      int
	ClaimLease       time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		ServiceName: getEnv("SERVICE_NAME", DefaultServiceName),
		AppEnv:      getEnv("APP_ENV", DefaultAppEnv),
		HTTP: HTTPConfig{
			Addr: getEnv("WISHLIST_HTTP_ADDR", DefaultHTTPAddr),
		},
		Mongo: MongoConfig{
			URI:        getEnv("WISHLIST_MONGO_URI", DefaultMongoURI),
			Database:   getEnv("WISHLIST_MONGO_DATABASE", DefaultMongoDatabase),
			Collection: getEnv("WISHLIST_MONGO_COLLECTION", DefaultMongoCollection),
		},
		ProductService: ProductServiceConfig{
			BaseURL: strings.TrimRight(getEnv("WISHLIST_PRODUCT_SERVICE_BASE_URL", getEnv("PRODUCT_SERVICE_BASE_URL", DefaultProductBaseURL)), "/"),
		},
		CartService: CartServiceConfig{
			BaseURL: strings.TrimRight(getEnv("WISHLIST_CART_SERVICE_BASE_URL", getEnv("CART_SERVICE_BASE_URL", DefaultCartBaseURL)), "/"),
		},
		Auth: AuthConfig{
			UserIDHeader:    getEnv("WISHLIST_AUTH_USER_ID_HEADER", DefaultUserIDHeader),
			RolesHeader:     getEnv("WISHLIST_AUTH_ROLES_HEADER", DefaultRolesHeader),
			RequestIDHeader: getEnv("WISHLIST_REQUEST_ID_HEADER", DefaultRequestIDHeader),
		},
		Events: EventsConfig{
			Backend: strings.ToLower(getEnv("WISHLIST_EVENTS_BACKEND", DefaultEventsBackend)),
			Kafka: KafkaConfig{
				Brokers: splitCSV(getEnv("WISHLIST_KAFKA_BROKERS", DefaultKafkaBrokers)),
			},
			ProductEvents: ProductEventsConfig{
				Topic:    getEnv("WISHLIST_PRODUCT_EVENTS_TOPIC", DefaultProductEventsTopic),
				GroupID:  getEnv("WISHLIST_PRICE_EVENTS_GROUP_ID", getEnv("WISHLIST_PRODUCT_EVENTS_GROUP_ID", DefaultProductEventsGroupID)),
				DLQTopic: getEnv("WISHLIST_PRICE_EVENTS_DLQ_TOPIC", getEnv("WISHLIST_PRODUCT_EVENTS_DLQ_TOPIC", DefaultProductEventsDLQTopic)),
			},
			NotificationCommands: NotificationCommandsConfig{
				Topic: getEnv("WISHLIST_NOTIFICATION_COMMANDS_TOPIC", DefaultNotificationCommandsTopic),
			},
			PriceDrop: PriceDropConfig{
				TemplateKey: getEnv("WISHLIST_PRICE_DROP_TEMPLATE_KEY", DefaultPriceDropTemplateKey),
			},
			Analytics: AnalyticsEventsConfig{
				OutboxCollection: getEnv("WISHLIST_EVENT_OUTBOX_COLLECTION", DefaultWishlistEventCollection),
				Topic:            getEnv("WISHLIST_EVENT_TOPIC", DefaultWishlistEventTopic),
				Publisher:        strings.ToLower(getEnv("WISHLIST_EVENT_PUBLISHER", DefaultWishlistEventPublisher)),
			},
		},
	}

	var err error
	if cfg.LogLevel, err = parseLogLevel(getEnv("WISHLIST_LOG_LEVEL", getEnv("LOG_LEVEL", "info"))); err != nil {
		return Config{}, err
	}
	if cfg.HTTP.ReadTimeout, err = durationFromEnv("WISHLIST_HTTP_READ_TIMEOUT", 5*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.HTTP.WriteTimeout, err = durationFromEnv("WISHLIST_HTTP_WRITE_TIMEOUT", 10*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.HTTP.IdleTimeout, err = durationFromEnv("WISHLIST_HTTP_IDLE_TIMEOUT", 60*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.Mongo.ConnectTimeout, err = durationFromEnv("WISHLIST_MONGO_CONNECT_TIMEOUT", 5*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.Mongo.ServerSelectionTimeout, err = durationFromEnv("WISHLIST_MONGO_SERVER_SELECTION_TIMEOUT", 5*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.ProductService.Timeout, err = durationFromEnvAny([]string{"WISHLIST_PRODUCT_SERVICE_TIMEOUT", "WISHLIST_PRODUCT_CALL_TIMEOUT_MS"}, time.Second); err != nil {
		return Config{}, err
	}
	if cfg.CartService.Timeout, err = durationFromEnvAny([]string{"WISHLIST_CART_SERVICE_TIMEOUT", "WISHLIST_CART_CALL_TIMEOUT_MS"}, time.Second); err != nil {
		return Config{}, err
	}
	if cfg.Auth.RequireBuyerRole, err = boolFromEnv("WISHLIST_AUTH_REQUIRE_BUYER_ROLE", true); err != nil {
		return Config{}, err
	}
	if cfg.Events.Analytics.Enabled, err = boolFromEnv("WISHLIST_ANALYTICS_EVENTS_ENABLED", DefaultAnalyticsEventsEnabled); err != nil {
		return Config{}, err
	}
	if cfg.Events.Analytics.PollInterval, err = durationFromEnv("WISHLIST_EVENT_POLL_INTERVAL", DefaultWishlistEventPollInterval); err != nil {
		return Config{}, err
	}
	if cfg.Events.Analytics.BatchSize, err = intFromEnv("WISHLIST_EVENT_BATCH_SIZE", DefaultWishlistEventBatchSize); err != nil {
		return Config{}, err
	}
	if cfg.Events.Analytics.MaxAttempts, err = intFromEnv("WISHLIST_EVENT_MAX_ATTEMPTS", DefaultWishlistEventMaxAttempts); err != nil {
		return Config{}, err
	}
	if cfg.Events.Analytics.ClaimLease, err = durationFromEnv("WISHLIST_EVENT_CLAIM_LEASE", DefaultWishlistEventClaimLease); err != nil {
		return Config{}, err
	}
	if cfg.Events.ProductEvents.MaxAttempts, err = intFromEnv("WISHLIST_PRODUCT_EVENTS_MAX_ATTEMPTS", DefaultProductEventsMaxAttempts); err != nil {
		return Config{}, err
	}
	if cfg.Events.ProductEvents.RetryBackoff, err = durationFromEnv("WISHLIST_PRODUCT_EVENTS_RETRY_BACKOFF", DefaultProductEventsRetryBackoff); err != nil {
		return Config{}, err
	}
	if cfg.Events.PriceDrop.BatchSize, err = intFromEnv("WISHLIST_PRICE_DROP_BATCH_SIZE", DefaultPriceDropBatchSize); err != nil {
		return Config{}, err
	}
	if cfg.Events.PriceDrop.MinDeltaAmount, err = int64FromEnv("WISHLIST_PRICE_DROP_MIN_DELTA_AMOUNT", DefaultPriceDropMinDeltaAmount); err != nil {
		return Config{}, err
	}
	if cfg.StartupTimeout, err = durationFromEnv("WISHLIST_STARTUP_TIMEOUT", 10*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.ShutdownTimeout, err = durationFromEnv("WISHLIST_SHUTDOWN_TIMEOUT", 10*time.Second); err != nil {
		return Config{}, err
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	var validationErrors []string
	if strings.TrimSpace(c.ServiceName) == "" {
		validationErrors = append(validationErrors, "SERVICE_NAME is required")
	}
	if strings.TrimSpace(c.HTTP.Addr) == "" {
		validationErrors = append(validationErrors, "WISHLIST_HTTP_ADDR is required")
	}
	if strings.TrimSpace(c.Mongo.URI) == "" {
		validationErrors = append(validationErrors, "WISHLIST_MONGO_URI is required")
	}
	if strings.TrimSpace(c.Mongo.Database) == "" {
		validationErrors = append(validationErrors, "WISHLIST_MONGO_DATABASE is required")
	}
	if strings.TrimSpace(c.Mongo.Collection) == "" {
		validationErrors = append(validationErrors, "WISHLIST_MONGO_COLLECTION is required")
	}
	if c.Mongo.ConnectTimeout <= 0 {
		validationErrors = append(validationErrors, "WISHLIST_MONGO_CONNECT_TIMEOUT must be positive")
	}
	if c.Mongo.ServerSelectionTimeout <= 0 {
		validationErrors = append(validationErrors, "WISHLIST_MONGO_SERVER_SELECTION_TIMEOUT must be positive")
	}
	if strings.TrimSpace(c.ProductService.BaseURL) == "" {
		validationErrors = append(validationErrors, "WISHLIST_PRODUCT_SERVICE_BASE_URL is required")
	} else if err := validateHTTPURL(c.ProductService.BaseURL); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("WISHLIST_PRODUCT_SERVICE_BASE_URL is invalid: %v", err))
	}
	if c.ProductService.Timeout <= 0 {
		validationErrors = append(validationErrors, "WISHLIST_PRODUCT_SERVICE_TIMEOUT must be positive")
	}
	if strings.TrimSpace(c.CartService.BaseURL) == "" {
		validationErrors = append(validationErrors, "WISHLIST_CART_SERVICE_BASE_URL is required")
	} else if err := validateHTTPURL(c.CartService.BaseURL); err != nil {
		validationErrors = append(validationErrors, fmt.Sprintf("WISHLIST_CART_SERVICE_BASE_URL is invalid: %v", err))
	}
	if c.CartService.Timeout <= 0 {
		validationErrors = append(validationErrors, "WISHLIST_CART_SERVICE_TIMEOUT must be positive")
	}
	if strings.TrimSpace(c.Auth.UserIDHeader) == "" {
		validationErrors = append(validationErrors, "WISHLIST_AUTH_USER_ID_HEADER is required")
	}
	if strings.TrimSpace(c.Auth.RolesHeader) == "" {
		validationErrors = append(validationErrors, "WISHLIST_AUTH_ROLES_HEADER is required")
	}
	if strings.TrimSpace(c.Auth.RequestIDHeader) == "" {
		validationErrors = append(validationErrors, "WISHLIST_REQUEST_ID_HEADER is required")
	}
	if c.Events.Analytics.Enabled {
		if strings.TrimSpace(c.Events.Analytics.OutboxCollection) == "" {
			validationErrors = append(validationErrors, "WISHLIST_EVENT_OUTBOX_COLLECTION is required when WISHLIST_ANALYTICS_EVENTS_ENABLED=true")
		}
		if strings.TrimSpace(c.Events.Analytics.Topic) == "" {
			validationErrors = append(validationErrors, "WISHLIST_EVENT_TOPIC is required when WISHLIST_ANALYTICS_EVENTS_ENABLED=true")
		}
		if c.Events.Analytics.PollInterval <= 0 {
			validationErrors = append(validationErrors, "WISHLIST_EVENT_POLL_INTERVAL must be positive")
		}
		if c.Events.Analytics.BatchSize <= 0 {
			validationErrors = append(validationErrors, "WISHLIST_EVENT_BATCH_SIZE must be positive")
		}
		if c.Events.Analytics.MaxAttempts <= 0 {
			validationErrors = append(validationErrors, "WISHLIST_EVENT_MAX_ATTEMPTS must be positive")
		}
		if c.Events.Analytics.ClaimLease <= 0 {
			validationErrors = append(validationErrors, "WISHLIST_EVENT_CLAIM_LEASE must be positive")
		}
	}
	switch strings.ToLower(strings.TrimSpace(c.Events.Analytics.Publisher)) {
	case EventsBackendDisabled, "":
	case EventsBackendKafka:
		if c.Events.Analytics.Enabled && len(c.Events.Kafka.Brokers) == 0 {
			validationErrors = append(validationErrors, "WISHLIST_KAFKA_BROKERS is required when WISHLIST_EVENT_PUBLISHER=kafka")
		}
	default:
		validationErrors = append(validationErrors, "WISHLIST_EVENT_PUBLISHER must be disabled or kafka")
	}
	switch strings.ToLower(strings.TrimSpace(c.Events.Backend)) {
	case EventsBackendDisabled, "":
	case EventsBackendKafka:
		if len(c.Events.Kafka.Brokers) == 0 {
			validationErrors = append(validationErrors, "WISHLIST_KAFKA_BROKERS is required when WISHLIST_EVENTS_BACKEND=kafka")
		}
		if strings.TrimSpace(c.Events.ProductEvents.Topic) == "" {
			validationErrors = append(validationErrors, "WISHLIST_PRODUCT_EVENTS_TOPIC is required when WISHLIST_EVENTS_BACKEND=kafka")
		}
		if strings.TrimSpace(c.Events.ProductEvents.GroupID) == "" {
			validationErrors = append(validationErrors, "WISHLIST_PRODUCT_EVENTS_GROUP_ID is required when WISHLIST_EVENTS_BACKEND=kafka")
		}
		if strings.TrimSpace(c.Events.ProductEvents.DLQTopic) == "" {
			validationErrors = append(validationErrors, "WISHLIST_PRODUCT_EVENTS_DLQ_TOPIC is required when WISHLIST_EVENTS_BACKEND=kafka")
		}
		if strings.TrimSpace(c.Events.NotificationCommands.Topic) == "" {
			validationErrors = append(validationErrors, "WISHLIST_NOTIFICATION_COMMANDS_TOPIC is required when WISHLIST_EVENTS_BACKEND=kafka")
		}
		if strings.TrimSpace(c.Events.PriceDrop.TemplateKey) == "" {
			validationErrors = append(validationErrors, "WISHLIST_PRICE_DROP_TEMPLATE_KEY is required when WISHLIST_EVENTS_BACKEND=kafka")
		}
		if c.Events.PriceDrop.BatchSize <= 0 {
			validationErrors = append(validationErrors, "WISHLIST_PRICE_DROP_BATCH_SIZE must be positive")
		}
		if c.Events.PriceDrop.MinDeltaAmount <= 0 {
			validationErrors = append(validationErrors, "WISHLIST_PRICE_DROP_MIN_DELTA_AMOUNT must be positive")
		}
		if c.Events.ProductEvents.MaxAttempts <= 0 {
			validationErrors = append(validationErrors, "WISHLIST_PRODUCT_EVENTS_MAX_ATTEMPTS must be positive")
		}
		if c.Events.ProductEvents.RetryBackoff <= 0 {
			validationErrors = append(validationErrors, "WISHLIST_PRODUCT_EVENTS_RETRY_BACKOFF must be positive")
		}
	default:
		validationErrors = append(validationErrors, "WISHLIST_EVENTS_BACKEND must be disabled or kafka")
	}
	if c.StartupTimeout <= 0 {
		validationErrors = append(validationErrors, "WISHLIST_STARTUP_TIMEOUT must be positive")
	}
	if c.ShutdownTimeout <= 0 {
		validationErrors = append(validationErrors, "WISHLIST_SHUTDOWN_TIMEOUT must be positive")
	}
	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return fallback
}

func durationFromEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := getEnv(key, "")
	if value == "" {
		return fallback, nil
	}
	return parseDurationValue(key, value)
}

func durationFromEnvAny(keys []string, fallback time.Duration) (time.Duration, error) {
	for _, key := range keys {
		value := getEnv(key, "")
		if value == "" {
			continue
		}
		return parseDurationValue(key, value)
	}
	return fallback, nil
}

func intFromEnv(key string, fallback int) (int, error) {
	value := getEnv(key, "")
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return parsed, nil
}

func int64FromEnv(key string, fallback int64) (int64, error) {
	value := getEnv(key, "")
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return parsed, nil
}

func parseDurationValue(key, value string) (time.Duration, error) {
	duration, err := time.ParseDuration(value)
	if err == nil {
		return duration, nil
	}
	milliseconds, convErr := strconv.Atoi(value)
	if convErr != nil {
		return 0, fmt.Errorf("%s must be a duration such as 5s or milliseconds as an integer: %w", key, err)
	}
	return time.Duration(milliseconds) * time.Millisecond, nil
}

func boolFromEnv(key string, fallback bool) (bool, error) {
	value := getEnv(key, "")
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", key, err)
	}
	return parsed, nil
}

func validateHTTPURL(value string) error {
	parsed, err := url.ParseRequestURI(value)
	if err != nil {
		return err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("scheme must be http or https")
	}
	if parsed.Host == "" {
		return fmt.Errorf("host is required")
	}
	return nil
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func parseLogLevel(value string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unsupported log level %q", value)
	}
}
