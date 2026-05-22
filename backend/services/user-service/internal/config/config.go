package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	GRPC     GRPCConfig
	Database DatabaseConfig
	Log      LogConfig
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
		Database: DatabaseConfig{
			DSN:             dsn,
			MaxOpenConns:    intEnv("USER_SERVICE_DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    intEnv("USER_SERVICE_DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: durationEnv("USER_SERVICE_DB_CONN_MAX_LIFETIME", 5*time.Minute),
			PingTimeout:     durationEnv("USER_SERVICE_DB_PING_TIMEOUT", 5*time.Second),
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
	if cfg.Database.PingTimeout <= 0 {
		return Config{}, errors.New("USER_SERVICE_DB_PING_TIMEOUT must be positive")
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

func (cfg Config) String() string {
	return fmt.Sprintf("grpc=%s reflection=%t db_max_open=%d", cfg.GRPC.Address, cfg.GRPC.Reflection, cfg.Database.MaxOpenConns)
}
