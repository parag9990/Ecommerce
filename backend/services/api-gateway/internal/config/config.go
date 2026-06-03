package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

type Config struct {
	HTTP        HTTPConfig
	UserService UserServiceConfig
	Auth        AuthConfig
	Validation  ValidationConfig
	Log         LogConfig
}

type HTTPConfig struct {
	Address         string
	ShutdownTimeout time.Duration
}

type UserServiceConfig struct {
	Address        string
	RequestTimeout time.Duration
}

type AuthConfig struct {
	JWTHS256Secret string
	Issuer         string
	Audience       string
	ClockSkew      time.Duration
}

type ValidationConfig struct {
	PhoneRegion string
}

type LogConfig struct {
	Level string
}

func Load() (Config, error) {
	cfg := Config{
		HTTP: HTTPConfig{
			Address:         firstStringEnv([]string{"API_GATEWAY_HTTP_ADDRESS", "HTTP_ADDR"}, ":8080"),
			ShutdownTimeout: firstDurationEnv([]string{"API_GATEWAY_SHUTDOWN_TIMEOUT", "HTTP_SHUTDOWN_TIMEOUT"}, 10*time.Second),
		},
		UserService: UserServiceConfig{
			Address:        firstStringEnv([]string{"USER_SERVICE_GRPC_ADDRESS", "USER_GRPC_ADDR"}, "localhost:50052"),
			RequestTimeout: firstDurationEnv([]string{"API_GATEWAY_USER_SERVICE_TIMEOUT", "GRPC_DIAL_TIMEOUT"}, 2*time.Second),
		},
		Auth: AuthConfig{
			JWTHS256Secret: firstStringEnv([]string{"API_GATEWAY_JWT_HS256_SECRET", "JWT_HS256_SECRET"}, ""),
			Issuer:         firstStringEnv([]string{"API_GATEWAY_JWT_ISSUER", "JWT_ISSUER"}, "ecommerce-auth"),
			Audience:       firstStringEnv([]string{"API_GATEWAY_JWT_AUDIENCE", "JWT_AUDIENCE"}, "ecommerce-api"),
			ClockSkew:      firstDurationEnv([]string{"API_GATEWAY_JWT_CLOCK_SKEW", "JWT_CLOCK_SKEW"}, 30*time.Second),
		},
		Validation: ValidationConfig{
			PhoneRegion: firstStringEnv([]string{"API_GATEWAY_VALIDATION_PHONE_REGION", "VALIDATION_PHONE_REGION"}, "IN"),
		},
		Log: LogConfig{
			Level: firstStringEnv([]string{"API_GATEWAY_LOG_LEVEL", "LOG_LEVEL"}, "info"),
		},
	}

	if cfg.Auth.JWTHS256Secret == "" {
		return Config{}, errors.New("API_GATEWAY_JWT_HS256_SECRET is required")
	}
	if cfg.HTTP.ShutdownTimeout <= 0 {
		return Config{}, errors.New("API_GATEWAY_SHUTDOWN_TIMEOUT must be positive")
	}
	if cfg.UserService.RequestTimeout <= 0 {
		return Config{}, errors.New("API_GATEWAY_USER_SERVICE_TIMEOUT must be positive")
	}
	if cfg.Auth.ClockSkew < 0 {
		return Config{}, errors.New("API_GATEWAY_JWT_CLOCK_SKEW cannot be negative")
	}
	if cfg.Validation.PhoneRegion == "" {
		return Config{}, errors.New("API_GATEWAY_VALIDATION_PHONE_REGION is required")
	}

	return cfg, nil
}

func firstStringEnv(keys []string, fallback string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return fallback
}

func firstDurationEnv(keys []string, fallback time.Duration) time.Duration {
	for _, key := range keys {
		value := os.Getenv(key)
		if value == "" {
			continue
		}

		parsed, err := time.ParseDuration(value)
		if err != nil {
			return fallback
		}
		return parsed
	}
	return fallback
}

func (cfg Config) String() string {
	return fmt.Sprintf("http=%s user_grpc=%s timeout=%s validation_phone_region=%s", cfg.HTTP.Address, cfg.UserService.Address, cfg.UserService.RequestTimeout, cfg.Validation.PhoneRegion)
}
