package repository

import (
	"context"
	"errors"
	"strings"
	"time"
)

type RedisZeroResultDedupeStore struct {
	client *RedisClient
}

func NewRedisZeroResultDedupeStore(client *RedisClient) (*RedisZeroResultDedupeStore, error) {
	if client == nil {
		return nil, errors.New("redis client is required")
	}
	return &RedisZeroResultDedupeStore{client: client}, nil
}

func (s *RedisZeroResultDedupeStore) MarkFirstSeen(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if strings.TrimSpace(key) == "" {
		return false, errors.New("zero-result dedupe key is required")
	}
	if ttl <= 0 {
		return false, errors.New("zero-result dedupe ttl must be greater than zero")
	}
	return s.client.SetNX(ctx, key, "1", ttl)
}

func (s *RedisZeroResultDedupeStore) Release(ctx context.Context, key string) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("zero-result dedupe key is required")
	}
	return s.client.Delete(ctx, key)
}
