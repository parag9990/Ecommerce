package repository

import (
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	defaultRedisAddr         = "localhost:6379"
	defaultRedisDialTimeout  = 2 * time.Second
	defaultRedisReadTimeout  = 2 * time.Second
	defaultRedisWriteTimeout = 2 * time.Second
)

type RedisClientConfig struct {
	Addr         string
	Password     string
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func NewRedisClient(cfg RedisClientConfig) (*redis.Client, error) {
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}), nil
}

func (c RedisClientConfig) Validate() error {
	if strings.TrimSpace(c.Addr) == "" {
		return errors.New("SESSION_REDIS_ADDR cannot be empty")
	}
	if c.DB < 0 {
		return errors.New("SESSION_REDIS_DB cannot be negative")
	}
	if c.DialTimeout <= 0 {
		return errors.New("SESSION_REDIS_DIAL_TIMEOUT must be greater than zero")
	}
	if c.ReadTimeout <= 0 {
		return errors.New("SESSION_REDIS_READ_TIMEOUT must be greater than zero")
	}
	if c.WriteTimeout <= 0 {
		return errors.New("SESSION_REDIS_WRITE_TIMEOUT must be greater than zero")
	}
	return nil
}

func (c RedisClientConfig) withDefaults() RedisClientConfig {
	if strings.TrimSpace(c.Addr) == "" {
		c.Addr = defaultRedisAddr
	}
	if c.DialTimeout == 0 {
		c.DialTimeout = defaultRedisDialTimeout
	}
	if c.ReadTimeout == 0 {
		c.ReadTimeout = defaultRedisReadTimeout
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = defaultRedisWriteTimeout
	}
	return c
}
