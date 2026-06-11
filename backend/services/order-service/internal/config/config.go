package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/events"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/orderevents"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/usecase"
)

const (
	defaultReservationTTL                = 15 * time.Minute
	defaultIdempotencyTTL                = 24 * time.Hour
	defaultReleaseTimeout                = 3 * time.Second
	defaultPaymentInventoryActionTimeout = 3 * time.Second
	defaultPaymentAllowedCurrencies      = "INR,USD"
	defaultGRPCAddress                   = ":9094"
	defaultOrderEventsEnabled            = true
	defaultOrderOutboxBatchSize          = 100
	defaultOrderOutboxInterval           = time.Second
	defaultOrderOutboxMaxAttempts        = 8
	defaultOrderOutboxInitialBackoff     = 5 * time.Second
	defaultOrderOutboxMaxBackoff         = 10 * time.Minute
	defaultOrderOutboxStaleLockTimeout   = 5 * time.Minute
	defaultOrderKafkaWriteTimeout        = 10 * time.Second
)

type Config struct {
	Database DatabaseConfig
	Checkout CheckoutConfig
	Payment  PaymentConfig
	GRPC     GRPCConfig
	Events   EventConfig
}

type DatabaseConfig struct {
	DSN string
}

type CheckoutConfig struct {
	InventoryReservationTTL time.Duration
	InventoryReleaseTimeout time.Duration
	IdempotencyTTL          time.Duration
}

type PaymentConfig struct {
	ReturnURL              string
	AllowedCurrencies      []string
	InventoryActionTimeout time.Duration
}

type GRPCConfig struct {
	Address             string
	TrustedCallerToken  string
	PageTokenSigningKey string
}

type EventConfig struct {
	Enabled           bool
	Topic             string
	KafkaBrokers      []string
	KafkaWriteTimeout time.Duration
	Outbox            OutboxConfig
}

type OutboxConfig struct {
	BatchSize        int
	Interval         time.Duration
	MaxAttempts      int
	InitialBackoff   time.Duration
	MaxBackoff       time.Duration
	StaleLockTimeout time.Duration
}

func Load() (Config, error) {
	reservationTTL, err := envDuration("ORDER_INVENTORY_RESERVATION_TTL", defaultReservationTTL)
	if err != nil {
		return Config{}, err
	}
	releaseTimeout, err := envDuration("ORDER_INVENTORY_RELEASE_TIMEOUT", defaultReleaseTimeout)
	if err != nil {
		return Config{}, err
	}
	idempotencyTTL, err := envDuration("ORDER_IDEMPOTENCY_TTL", defaultIdempotencyTTL)
	if err != nil {
		return Config{}, err
	}
	actionTimeout, err := envDuration("ORDER_PAYMENT_INVENTORY_ACTION_TIMEOUT", defaultPaymentInventoryActionTimeout)
	if err != nil {
		return Config{}, err
	}
	outboxInterval, err := envDuration("ORDER_OUTBOX_INTERVAL", defaultOrderOutboxInterval)
	if err != nil {
		return Config{}, err
	}
	outboxInitialBackoff, err := envDuration("ORDER_OUTBOX_INITIAL_BACKOFF", defaultOrderOutboxInitialBackoff)
	if err != nil {
		return Config{}, err
	}
	outboxMaxBackoff, err := envDuration("ORDER_OUTBOX_MAX_BACKOFF", defaultOrderOutboxMaxBackoff)
	if err != nil {
		return Config{}, err
	}
	outboxStaleLockTimeout, err := envDuration("ORDER_OUTBOX_STALE_LOCK_TIMEOUT", defaultOrderOutboxStaleLockTimeout)
	if err != nil {
		return Config{}, err
	}
	kafkaWriteTimeout, err := envDuration("ORDER_KAFKA_WRITE_TIMEOUT", defaultOrderKafkaWriteTimeout)
	if err != nil {
		return Config{}, err
	}
	eventsEnabled, err := envBool("ORDER_EVENTS_ENABLED", defaultOrderEventsEnabled)
	if err != nil {
		return Config{}, err
	}
	config := Config{
		Database: DatabaseConfig{DSN: strings.TrimSpace(os.Getenv("ORDER_MYSQL_DSN"))},
		Checkout: CheckoutConfig{
			InventoryReservationTTL: reservationTTL,
			InventoryReleaseTimeout: releaseTimeout,
			IdempotencyTTL:          idempotencyTTL,
		},
		Payment: PaymentConfig{
			ReturnURL:              strings.TrimSpace(os.Getenv("ORDER_PAYMENT_RETURN_URL")),
			AllowedCurrencies:      csvValues(os.Getenv("ORDER_PAYMENT_ALLOWED_CURRENCIES"), defaultPaymentAllowedCurrencies),
			InventoryActionTimeout: actionTimeout,
		},
		GRPC: GRPCConfig{
			Address:             envValue("ORDER_GRPC_ADDR", defaultGRPCAddress),
			TrustedCallerToken:  strings.TrimSpace(os.Getenv("ORDER_GRPC_TRUSTED_CALLER_TOKEN")),
			PageTokenSigningKey: strings.TrimSpace(os.Getenv("ORDER_PAGE_TOKEN_SIGNING_KEY")),
		},
		Events: EventConfig{
			Enabled:           eventsEnabled,
			Topic:             envValue("ORDER_EVENTS_TOPIC", orderevents.OrderEventsTopic),
			KafkaBrokers:      csvRawValues(os.Getenv("ORDER_KAFKA_BROKERS")),
			KafkaWriteTimeout: kafkaWriteTimeout,
			Outbox: OutboxConfig{
				BatchSize:        envInt("ORDER_OUTBOX_BATCH_SIZE", defaultOrderOutboxBatchSize),
				Interval:         outboxInterval,
				MaxAttempts:      envInt("ORDER_OUTBOX_MAX_ATTEMPTS", defaultOrderOutboxMaxAttempts),
				InitialBackoff:   outboxInitialBackoff,
				MaxBackoff:       outboxMaxBackoff,
				StaleLockTimeout: outboxStaleLockTimeout,
			},
		},
	}
	if err := config.Validate(); err != nil {
		return Config{}, err
	}
	return config, nil
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Database.DSN) == "" {
		return errors.New("ORDER_MYSQL_DSN is required")
	}
	if c.Checkout.InventoryReservationTTL <= 0 {
		return errors.New("ORDER_INVENTORY_RESERVATION_TTL must be greater than zero")
	}
	if c.Checkout.InventoryReleaseTimeout <= 0 {
		return errors.New("ORDER_INVENTORY_RELEASE_TIMEOUT must be greater than zero")
	}
	if c.Checkout.IdempotencyTTL <= 0 {
		return errors.New("ORDER_IDEMPOTENCY_TTL must be greater than zero")
	}
	returnURL, err := url.ParseRequestURI(strings.TrimSpace(c.Payment.ReturnURL))
	if err != nil || returnURL.Scheme != "https" || returnURL.Host == "" {
		return errors.New("ORDER_PAYMENT_RETURN_URL must be an absolute HTTPS URL")
	}
	if len(c.Payment.AllowedCurrencies) == 0 {
		return errors.New("ORDER_PAYMENT_ALLOWED_CURRENCIES must include at least one currency")
	}
	for _, currency := range c.Payment.AllowedCurrencies {
		if !isISOCurrency(currency) {
			return fmt.Errorf("ORDER_PAYMENT_ALLOWED_CURRENCIES includes invalid currency %q", currency)
		}
	}
	if c.Payment.InventoryActionTimeout <= 0 {
		return errors.New("ORDER_PAYMENT_INVENTORY_ACTION_TIMEOUT must be greater than zero")
	}
	if strings.TrimSpace(c.GRPC.Address) == "" {
		return errors.New("ORDER_GRPC_ADDR is required")
	}
	if len(c.GRPC.TrustedCallerToken) < 32 {
		return errors.New("ORDER_GRPC_TRUSTED_CALLER_TOKEN must be at least 32 characters")
	}
	if len(c.GRPC.PageTokenSigningKey) < 32 {
		return errors.New("ORDER_PAGE_TOKEN_SIGNING_KEY must be at least 32 characters")
	}
	if err := c.Events.Validate(); err != nil {
		return err
	}
	return nil
}

func (c EventConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if strings.TrimSpace(c.Topic) == "" {
		return errors.New("ORDER_EVENTS_TOPIC is required when order events are enabled")
	}
	if len(c.KafkaBrokers) == 0 {
		return errors.New("ORDER_KAFKA_BROKERS is required when order events are enabled")
	}
	for _, broker := range c.KafkaBrokers {
		if strings.TrimSpace(broker) == "" {
			return errors.New("ORDER_KAFKA_BROKERS contains an empty broker")
		}
	}
	if c.KafkaWriteTimeout <= 0 {
		return errors.New("ORDER_KAFKA_WRITE_TIMEOUT must be greater than zero")
	}
	if c.Outbox.BatchSize <= 0 {
		return errors.New("ORDER_OUTBOX_BATCH_SIZE must be greater than zero")
	}
	if c.Outbox.Interval <= 0 {
		return errors.New("ORDER_OUTBOX_INTERVAL must be greater than zero")
	}
	if c.Outbox.MaxAttempts <= 0 {
		return errors.New("ORDER_OUTBOX_MAX_ATTEMPTS must be greater than zero")
	}
	if c.Outbox.InitialBackoff <= 0 {
		return errors.New("ORDER_OUTBOX_INITIAL_BACKOFF must be greater than zero")
	}
	if c.Outbox.MaxBackoff < c.Outbox.InitialBackoff {
		return errors.New("ORDER_OUTBOX_MAX_BACKOFF must be greater than or equal to ORDER_OUTBOX_INITIAL_BACKOFF")
	}
	if c.Outbox.StaleLockTimeout <= 0 {
		return errors.New("ORDER_OUTBOX_STALE_LOCK_TIMEOUT must be greater than zero")
	}
	return nil
}

func (c Config) CreateOrderFromCartConfig() usecase.CreateOrderFromCartConfig {
	return usecase.CreateOrderFromCartConfig{
		ReservationTTL: c.Checkout.InventoryReservationTTL,
		ReleaseTimeout: c.Checkout.InventoryReleaseTimeout,
	}
}

func (c Config) CreateOrderConfig() usecase.CreateOrderConfig {
	return usecase.CreateOrderConfig{IdempotencyTTL: c.Checkout.IdempotencyTTL}
}

func (c Config) InitiateOrderPaymentConfig() usecase.InitiateOrderPaymentConfig {
	return usecase.InitiateOrderPaymentConfig{
		ReturnURL:               c.Payment.ReturnURL,
		AllowedCurrencies:       append([]string(nil), c.Payment.AllowedCurrencies...),
		InventoryReleaseTimeout: c.Payment.InventoryActionTimeout,
	}
}

func (c Config) ApplyPaymentResultConfig() usecase.ApplyPaymentResultConfig {
	return usecase.ApplyPaymentResultConfig{InventoryActionTimeout: c.Payment.InventoryActionTimeout}
}

func (c Config) OrderOutboxWorkerConfig() events.OutboxWorkerConfig {
	return events.OutboxWorkerConfig{
		Topic:            c.Events.Topic,
		BatchSize:        c.Events.Outbox.BatchSize,
		Interval:         c.Events.Outbox.Interval,
		MaxAttempts:      c.Events.Outbox.MaxAttempts,
		InitialBackoff:   c.Events.Outbox.InitialBackoff,
		MaxBackoff:       c.Events.Outbox.MaxBackoff,
		StaleLockTimeout: c.Events.Outbox.StaleLockTimeout,
	}
}

func (c Config) KafkaPublisherConfig() events.KafkaPublisherConfig {
	return events.KafkaPublisherConfig{
		Brokers:      append([]string(nil), c.Events.KafkaBrokers...),
		Topic:        c.Events.Topic,
		WriteTimeout: c.Events.KafkaWriteTimeout,
	}
}

func (c Config) PageTokenSigningKey() []byte {
	return []byte(c.GRPC.PageTokenSigningKey)
}

func envDuration(name string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", name, err)
	}
	return parsed, nil
}

func envValue(name string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func envBool(name string, fallback bool) (bool, error) {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(name)))
	if value == "" {
		return fallback, nil
	}
	switch value {
	case "1", "true", "t", "yes", "y", "on":
		return true, nil
	case "0", "false", "f", "no", "n", "off":
		return false, nil
	default:
		return false, fmt.Errorf("%s must be a boolean", name)
	}
}

func envInt(name string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

func csvValues(value string, fallback string) []string {
	if strings.TrimSpace(value) == "" {
		value = fallback
	}
	seen := make(map[string]struct{})
	var values []string
	for _, entry := range strings.Split(value, ",") {
		entry = strings.ToUpper(strings.TrimSpace(entry))
		if entry == "" {
			continue
		}
		if _, exists := seen[entry]; exists {
			continue
		}
		seen[entry] = struct{}{}
		values = append(values, entry)
	}
	return values
}

func csvRawValues(value string) []string {
	seen := make(map[string]struct{})
	var values []string
	for _, entry := range strings.Split(value, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if _, exists := seen[entry]; exists {
			continue
		}
		seen[entry] = struct{}{}
		values = append(values, entry)
	}
	return values
}

func isISOCurrency(value string) bool {
	if len(value) != 3 {
		return false
	}
	for _, character := range value {
		if character < 'A' || character > 'Z' {
			return false
		}
	}
	return true
}
