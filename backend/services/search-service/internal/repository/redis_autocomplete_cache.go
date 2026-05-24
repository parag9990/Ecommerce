package repository

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type RedisAutocompleteCache struct {
	client *RedisClient
}

func NewRedisAutocompleteCache(client *RedisClient) (*RedisAutocompleteCache, error) {
	if client == nil {
		return nil, errors.New("redis client is required")
	}
	return &RedisAutocompleteCache{client: client}, nil
}

func (c *RedisAutocompleteCache) Get(ctx context.Context, key string) ([]string, bool, error) {
	if strings.TrimSpace(key) == "" {
		return nil, false, errors.New("autocomplete cache key is required")
	}
	raw, ok, err := c.client.Get(ctx, key)
	if err != nil || !ok {
		return nil, ok, err
	}

	var suggestions []string
	if err := json.Unmarshal([]byte(raw), &suggestions); err != nil {
		return nil, false, err
	}
	if suggestions == nil {
		suggestions = []string{}
	}
	return suggestions, true, nil
}

func (c *RedisAutocompleteCache) Set(ctx context.Context, key string, suggestions []string, ttl time.Duration) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("autocomplete cache key is required")
	}
	if ttl <= 0 {
		return errors.New("autocomplete cache ttl must be greater than zero")
	}
	if suggestions == nil {
		suggestions = []string{}
	}
	payload, err := json.Marshal(suggestions)
	if err != nil {
		return err
	}
	return c.client.SetEX(ctx, key, string(payload), ttl)
}
