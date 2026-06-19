package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPAddress     = ":8086"
	defaultShutdownTimeout = 10 * time.Second
)

type Config struct {
	HTTP    HTTPConfig
	Mongo   MongoConfig
	Redis   RedisConfig
	Privacy PrivacyConfig
}

type HTTPConfig struct {
	Address         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type MongoConfig struct {
	URI            string
	Database       string
	ConnectTimeout time.Duration
}

type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	KeyPrefix    string
}

type PrivacyConfig struct {
	HashPepper        string
	DeletionListLimit int
}

func Load() (Config, error) {
	cfg := Config{
		HTTP: HTTPConfig{
			Address:         envString("SESSION_HTTP_ADDR", defaultHTTPAddress),
			ReadTimeout:     envDuration("SESSION_HTTP_READ_TIMEOUT", 5*time.Second),
			WriteTimeout:    envDuration("SESSION_HTTP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:     envDuration("SESSION_HTTP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: envDuration("SESSION_SHUTDOWN_TIMEOUT", defaultShutdownTimeout),
		},
		Mongo: MongoConfig{
			URI:            envString("SESSION_MONGO_URI", "mongodb://localhost:27017"),
			Database:       envString("SESSION_MONGO_DATABASE", "session_db"),
			ConnectTimeout: envDuration("SESSION_MONGO_CONNECT_TIMEOUT", 5*time.Second),
		},
		Redis: RedisConfig{
			Addr:         envString("SESSION_REDIS_ADDR", "localhost:6379"),
			Password:     os.Getenv("SESSION_REDIS_PASSWORD"),
			DB:           envInt("SESSION_REDIS_DB", 2),
			DialTimeout:  envDuration("SESSION_REDIS_DIAL_TIMEOUT", 2*time.Second),
			ReadTimeout:  envDuration("SESSION_REDIS_READ_TIMEOUT", 2*time.Second),
			WriteTimeout: envDuration("SESSION_REDIS_WRITE_TIMEOUT", 2*time.Second),
			KeyPrefix:    envString("SESSION_REDIS_KEY_PREFIX", "session"),
		},
		Privacy: PrivacyConfig{
			HashPepper:        os.Getenv("SESSION_PRIVACY_HASH_PEPPER"),
			DeletionListLimit: envInt("SESSION_PRIVACY_DELETION_LIST_LIMIT", 50),
		},
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	var errs []error
	if strings.TrimSpace(c.HTTP.Address) == "" {
		errs = append(errs, errors.New("SESSION_HTTP_ADDR is required"))
	}
	if strings.TrimSpace(c.Mongo.URI) == "" {
		errs = append(errs, errors.New("SESSION_MONGO_URI is required"))
	}
	if strings.TrimSpace(c.Mongo.Database) == "" {
		errs = append(errs, errors.New("SESSION_MONGO_DATABASE is required"))
	}
	if strings.TrimSpace(c.Redis.Addr) == "" {
		errs = append(errs, errors.New("SESSION_REDIS_ADDR is required"))
	}
	if strings.TrimSpace(c.Privacy.HashPepper) == "" {
		errs = append(errs, errors.New("SESSION_PRIVACY_HASH_PEPPER is required"))
	}
	if c.Privacy.DeletionListLimit < 1 || c.Privacy.DeletionListLimit > 200 {
		errs = append(errs, errors.New("SESSION_PRIVACY_DELETION_LIST_LIMIT must be between 1 and 200"))
	}
	return errors.Join(errs...)
}

func envString(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err == nil {
		return parsed
	}
	seconds, err := strconv.Atoi(value)
	if err == nil {
		return time.Duration(seconds) * time.Second
	}
	return fallback
}

func (c Config) String() string {
	return fmt.Sprintf("http=%s mongo_db=%s redis=%s", c.HTTP.Address, c.Mongo.Database, c.Redis.Addr)
}
