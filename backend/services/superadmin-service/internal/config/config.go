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
	ServiceName                 string
	AppEnv                      string
	HTTPAddr                    string
	LogLevel                    string
	DatabaseDSN                 string
	DBMaxOpenConns              int
	DBMaxIdleConns              int
	DBConnMaxLifetime           time.Duration
	DBPingTimeout               time.Duration
	RequireDatabase             bool
	UserServiceBaseURL          string
	UserServiceTimeout          time.Duration
	RequireUserService          bool
	OrderServiceBaseURL         string
	OrderServiceTimeout         time.Duration
	RequireOrderService         bool
	PaymentServiceBaseURL       string
	PaymentServiceTimeout       time.Duration
	RequirePaymentService       bool
	PlatformSettingsCacheTTL    time.Duration
	PlatformSettingsEventTopic  string
	SessionAnalyticsMaxRange    time.Duration
	SessionAnalyticsMaxPageSize int
	HTTPReadHeaderTimeout       time.Duration
	HTTPShutdownTimeout         time.Duration
}

func Load() Config {
	return Config{
		ServiceName:                 envString("SERVICE_NAME", "superadmin-service"),
		AppEnv:                      envString("APP_ENV", "local"),
		HTTPAddr:                    envString("HTTP_ADDR", ":8088"),
		LogLevel:                    envString("LOG_LEVEL", envString("SUPERADMIN_LOG_LEVEL", "info")),
		DatabaseDSN:                 envString("SUPERADMIN_DATABASE_DSN", envString("MYSQL_DSN", "")),
		DBMaxOpenConns:              envInt("SUPERADMIN_DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:              envInt("SUPERADMIN_DB_MAX_IDLE_CONNS", 25),
		DBConnMaxLifetime:           envDuration("SUPERADMIN_DB_CONN_MAX_LIFETIME", 5*time.Minute),
		DBPingTimeout:               envDuration("SUPERADMIN_DB_PING_TIMEOUT", 5*time.Second),
		RequireDatabase:             envBool("SUPERADMIN_REQUIRE_DATABASE", false),
		UserServiceBaseURL:          envString("USER_SERVICE_ADMIN_BASE_URL", ""),
		UserServiceTimeout:          envDuration("USER_SERVICE_TIMEOUT", 5*time.Second),
		RequireUserService:          envBool("SUPERADMIN_REQUIRE_USER_SERVICE", false),
		OrderServiceBaseURL:         envString("ORDER_SERVICE_ADMIN_BASE_URL", ""),
		OrderServiceTimeout:         envDuration("ORDER_SERVICE_TIMEOUT", 5*time.Second),
		RequireOrderService:         envBool("SUPERADMIN_REQUIRE_ORDER_SERVICE", false),
		PaymentServiceBaseURL:       envString("PAYMENT_SERVICE_ADMIN_BASE_URL", ""),
		PaymentServiceTimeout:       envDuration("PAYMENT_SERVICE_TIMEOUT", 5*time.Second),
		RequirePaymentService:       envBool("SUPERADMIN_REQUIRE_PAYMENT_SERVICE", false),
		PlatformSettingsCacheTTL:    envDuration("SUPERADMIN_PLATFORM_SETTINGS_CACHE_TTL", 5*time.Minute),
		PlatformSettingsEventTopic:  envString("SUPERADMIN_PLATFORM_SETTINGS_EVENT_TOPIC", "platform.settings.updated"),
		SessionAnalyticsMaxRange:    envDuration("SUPERADMIN_SESSION_ANALYTICS_MAX_RANGE", 30*24*time.Hour),
		SessionAnalyticsMaxPageSize: envInt("SUPERADMIN_SESSION_ANALYTICS_MAX_PAGE_SIZE", 100),
		HTTPReadHeaderTimeout:       envDuration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second),
		HTTPShutdownTimeout:         envDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
	}
}

func (c Config) Validate() error {
	var problems []string

	if strings.TrimSpace(c.ServiceName) == "" {
		problems = append(problems, "SERVICE_NAME is required")
	}
	if strings.TrimSpace(c.HTTPAddr) == "" {
		problems = append(problems, "HTTP_ADDR is required")
	}
	if c.DBMaxOpenConns < 1 {
		problems = append(problems, "SUPERADMIN_DB_MAX_OPEN_CONNS must be greater than zero")
	}
	if c.DBMaxIdleConns < 0 {
		problems = append(problems, "SUPERADMIN_DB_MAX_IDLE_CONNS cannot be negative")
	}
	if c.DBMaxIdleConns > c.DBMaxOpenConns {
		problems = append(problems, "SUPERADMIN_DB_MAX_IDLE_CONNS cannot exceed SUPERADMIN_DB_MAX_OPEN_CONNS")
	}
	if c.DBPingTimeout <= 0 {
		problems = append(problems, "SUPERADMIN_DB_PING_TIMEOUT must be positive")
	}
	if c.UserServiceTimeout <= 0 {
		problems = append(problems, "USER_SERVICE_TIMEOUT must be positive")
	}
	if c.OrderServiceTimeout <= 0 {
		problems = append(problems, "ORDER_SERVICE_TIMEOUT must be positive")
	}
	if c.PaymentServiceTimeout <= 0 {
		problems = append(problems, "PAYMENT_SERVICE_TIMEOUT must be positive")
	}
	if c.PlatformSettingsCacheTTL < 0 {
		problems = append(problems, "SUPERADMIN_PLATFORM_SETTINGS_CACHE_TTL cannot be negative")
	}
	if strings.TrimSpace(c.PlatformSettingsEventTopic) == "" {
		problems = append(problems, "SUPERADMIN_PLATFORM_SETTINGS_EVENT_TOPIC is required")
	}
	if c.SessionAnalyticsMaxRange <= 0 {
		problems = append(problems, "SUPERADMIN_SESSION_ANALYTICS_MAX_RANGE must be positive")
	}
	if c.SessionAnalyticsMaxPageSize < 1 {
		problems = append(problems, "SUPERADMIN_SESSION_ANALYTICS_MAX_PAGE_SIZE must be greater than zero")
	}
	if c.HTTPReadHeaderTimeout <= 0 {
		problems = append(problems, "HTTP_READ_HEADER_TIMEOUT must be positive")
	}
	if c.HTTPShutdownTimeout <= 0 {
		problems = append(problems, "HTTP_SHUTDOWN_TIMEOUT must be positive")
	}
	if requiresDatabase(c) && strings.TrimSpace(c.DatabaseDSN) == "" {
		problems = append(problems, "SUPERADMIN_DATABASE_DSN is required outside local/test unless SUPERADMIN_REQUIRE_DATABASE=false")
	}
	if requiresUserService(c) && strings.TrimSpace(c.UserServiceBaseURL) == "" {
		problems = append(problems, "USER_SERVICE_ADMIN_BASE_URL is required outside local/test unless SUPERADMIN_REQUIRE_USER_SERVICE=false")
	}
	if requiresOrderService(c) && strings.TrimSpace(c.OrderServiceBaseURL) == "" {
		problems = append(problems, "ORDER_SERVICE_ADMIN_BASE_URL is required outside local/test unless SUPERADMIN_REQUIRE_ORDER_SERVICE=false")
	}
	if requiresPaymentService(c) && strings.TrimSpace(c.PaymentServiceBaseURL) == "" {
		problems = append(problems, "PAYMENT_SERVICE_ADMIN_BASE_URL is required outside local/test unless SUPERADMIN_REQUIRE_PAYMENT_SERVICE=false")
	}

	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}

func requiresDatabase(c Config) bool {
	if c.RequireDatabase {
		return true
	}
	env := strings.ToLower(strings.TrimSpace(c.AppEnv))
	return env != "" && env != "local" && env != "test"
}

func requiresUserService(c Config) bool {
	if c.RequireUserService {
		return true
	}
	env := strings.ToLower(strings.TrimSpace(c.AppEnv))
	return env != "" && env != "local" && env != "test"
}

func requiresOrderService(c Config) bool {
	if c.RequireOrderService {
		return true
	}
	env := strings.ToLower(strings.TrimSpace(c.AppEnv))
	return env != "" && env != "local" && env != "test"
}

func requiresPaymentService(c Config) bool {
	if c.RequirePaymentService {
		return true
	}
	env := strings.ToLower(strings.TrimSpace(c.AppEnv))
	return env != "" && env != "local" && env != "test"
}

func envString(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
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

func envDuration(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid duration for %s=%q, using %s\n", key, raw, fallback)
		return fallback
	}
	return value
}
