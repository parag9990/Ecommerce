package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	GRPC         GRPCConfig
	HTTP         HTTPConfig
	Database     DatabaseConfig
	Validation   ValidationConfig
	Events       EventsConfig
	OutboxWorker OutboxWorkerConfig
	RabbitMQ     RabbitMQConfig
	Kafka        KafkaConfig
	Log          LogConfig
}

type HTTPConfig struct {
	Address           string
	ReadHeaderTimeout time.Duration
	AdminToken        string
}

type GRPCConfig struct {
	Address         string
	Reflection      bool
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	PingTimeout     time.Duration
}

type ValidationConfig struct {
	PhoneRegion string
}

type EventsConfig struct {
	Enabled         bool
	Provider        string
	Topic           string
	DeadLetterTopic string
}

type OutboxWorkerConfig struct {
	Enabled        bool
	BatchSize      int
	PollInterval   time.Duration
	LockTTL        time.Duration
	MaxAttempts    int
	PublishTimeout time.Duration
}

type RabbitMQConfig struct {
	URL string
}

type KafkaConfig struct {
	Brokers []string
}

type LogConfig struct {
	Level string
}

func Load() (Config, error) {
	dsn := firstEnv("USER_SERVICE_DATABASE_DSN", "MYSQL_DSN")
	if dsn == "" {
		return Config{}, errors.New("USER_SERVICE_DATABASE_DSN is required")
	}

	cfg := Config{
		GRPC: GRPCConfig{
			Address:         stringEnv("USER_SERVICE_GRPC_ADDRESS", ":50052"),
			Reflection:      boolEnv("USER_SERVICE_GRPC_REFLECTION", true),
			ShutdownTimeout: durationEnv("USER_SERVICE_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		HTTP: HTTPConfig{
			Address:           stringEnv("USER_SERVICE_HTTP_ADDRESS", ":9091"),
			ReadHeaderTimeout: durationEnv("USER_SERVICE_HTTP_READ_HEADER_TIMEOUT", 5*time.Second),
			AdminToken:        strings.TrimSpace(os.Getenv("USER_SERVICE_ADMIN_TOKEN")),
		},
		Database: DatabaseConfig{
			DSN:             dsn,
			MaxOpenConns:    intEnv("USER_SERVICE_DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    intEnv("USER_SERVICE_DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: durationEnv("USER_SERVICE_DB_CONN_MAX_LIFETIME", 5*time.Minute),
			PingTimeout:     durationEnv("USER_SERVICE_DB_PING_TIMEOUT", 5*time.Second),
		},
		Validation: ValidationConfig{
			PhoneRegion: stringEnv("USER_SERVICE_VALIDATION_PHONE_REGION", "IN"),
		},
		Events: EventsConfig{
			Enabled:         boolEnv("USER_EVENTS_ENABLED", true),
			Provider:        strings.ToLower(strings.TrimSpace(stringEnv("USER_EVENTS_PROVIDER", "rabbitmq"))),
			Topic:           stringEnv("USER_EVENTS_TOPIC", "user.events"),
			DeadLetterTopic: stringEnv("USER_EVENTS_DLQ", "user.events.dlq"),
		},
		OutboxWorker: OutboxWorkerConfig{
			Enabled:        boolEnv("OUTBOX_WORKER_ENABLED", false),
			BatchSize:      intEnv("OUTBOX_WORKER_BATCH_SIZE", 100),
			PollInterval:   durationEnv("OUTBOX_WORKER_POLL_INTERVAL", 2*time.Second),
			LockTTL:        durationEnv("OUTBOX_WORKER_LOCK_TTL", 30*time.Second),
			MaxAttempts:    intEnv("OUTBOX_MAX_ATTEMPTS", 10),
			PublishTimeout: durationEnv("OUTBOX_PUBLISH_TIMEOUT", 10*time.Second),
		},
		RabbitMQ: RabbitMQConfig{
			URL: stringEnv("RABBITMQ_URL", ""),
		},
		Kafka: KafkaConfig{
			Brokers: splitCSVEnv("KAFKA_BROKERS"),
		},
		Log: LogConfig{
			Level: stringEnv("USER_SERVICE_LOG_LEVEL", "info"),
		},
	}

	if cfg.Database.MaxOpenConns < 1 {
		return Config{}, errors.New("USER_SERVICE_DB_MAX_OPEN_CONNS must be greater than zero")
	}
	if cfg.Database.MaxIdleConns < 0 {
		return Config{}, errors.New("USER_SERVICE_DB_MAX_IDLE_CONNS cannot be negative")
	}
	if cfg.GRPC.ShutdownTimeout <= 0 {
		return Config{}, errors.New("USER_SERVICE_SHUTDOWN_TIMEOUT must be positive")
	}
	if strings.TrimSpace(cfg.HTTP.Address) == "" {
		return Config{}, errors.New("USER_SERVICE_HTTP_ADDRESS is required")
	}
	if cfg.HTTP.ReadHeaderTimeout <= 0 {
		return Config{}, errors.New("USER_SERVICE_HTTP_READ_HEADER_TIMEOUT must be positive")
	}
	if len(cfg.HTTP.AdminToken) < 32 {
		return Config{}, errors.New("USER_SERVICE_ADMIN_TOKEN must be at least 32 characters")
	}
	if cfg.Database.PingTimeout <= 0 {
		return Config{}, errors.New("USER_SERVICE_DB_PING_TIMEOUT must be positive")
	}
	if cfg.Validation.PhoneRegion == "" {
		return Config{}, errors.New("USER_SERVICE_VALIDATION_PHONE_REGION is required")
	}
	if cfg.Events.Enabled && cfg.Events.Topic == "" {
		return Config{}, errors.New("USER_EVENTS_TOPIC is required when events are enabled")
	}
	if cfg.OutboxWorker.Enabled {
		if !cfg.Events.Enabled {
			return Config{}, errors.New("OUTBOX_WORKER_ENABLED requires USER_EVENTS_ENABLED")
		}
		if cfg.OutboxWorker.BatchSize <= 0 {
			return Config{}, errors.New("OUTBOX_WORKER_BATCH_SIZE must be greater than zero")
		}
		if cfg.OutboxWorker.PollInterval <= 0 {
			return Config{}, errors.New("OUTBOX_WORKER_POLL_INTERVAL must be positive")
		}
		if cfg.OutboxWorker.LockTTL <= 0 {
			return Config{}, errors.New("OUTBOX_WORKER_LOCK_TTL must be positive")
		}
		if cfg.OutboxWorker.MaxAttempts <= 0 {
			return Config{}, errors.New("OUTBOX_MAX_ATTEMPTS must be greater than zero")
		}
		switch cfg.Events.Provider {
		case "rabbitmq":
			if cfg.RabbitMQ.URL == "" {
				return Config{}, errors.New("RABBITMQ_URL is required when RabbitMQ outbox worker is enabled")
			}
		case "kafka":
			if len(cfg.Kafka.Brokers) == 0 {
				return Config{}, errors.New("KAFKA_BROKERS is required when Kafka outbox worker is enabled")
			}
		default:
			return Config{}, fmt.Errorf("unsupported USER_EVENTS_PROVIDER %q", cfg.Events.Provider)
		}
	}

	return cfg, nil
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}

func stringEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func intEnv(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func boolEnv(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func splitCSVEnv(key string) []string {
	value := os.Getenv(key)
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}

func (cfg Config) String() string {
	return fmt.Sprintf("grpc=%s http=%s reflection=%t db_max_open=%d validation_phone_region=%s events_enabled=%t events_provider=%s outbox_worker=%t", cfg.GRPC.Address, cfg.HTTP.Address, cfg.GRPC.Reflection, cfg.Database.MaxOpenConns, cfg.Validation.PhoneRegion, cfg.Events.Enabled, cfg.Events.Provider, cfg.OutboxWorker.Enabled)
}
