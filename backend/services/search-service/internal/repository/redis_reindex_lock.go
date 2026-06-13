package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

const defaultReindexLockKey = "search:reindex:products:lock"

type RedisReindexLock struct {
	redis *RedisClient
	key   string
}

func NewRedisReindexLock(redis *RedisClient, key string) (*RedisReindexLock, error) {
	if redis == nil {
		return nil, errors.New("redis client is required")
	}
	key = strings.TrimSpace(key)
	if key == "" {
		key = defaultReindexLockKey
	}
	return &RedisReindexLock{redis: redis, key: key}, nil
}

func (l *RedisReindexLock) Acquire(ctx context.Context, jobID string, ttl time.Duration) (bool, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return false, fmt.Errorf("%w: job_id is required for lock", domain.ErrInvalidReindexRequest)
	}
	acquired, err := l.redis.SetNX(ctx, l.key, jobID, ttl)
	if err != nil {
		return false, fmt.Errorf("%w: acquire redis lock: %v", domain.ErrSearchBackendUnavailable, err)
	}
	return acquired, nil
}

func (l *RedisReindexLock) Refresh(ctx context.Context, jobID string, ttl time.Duration) (bool, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return false, fmt.Errorf("%w: job_id is required for lock refresh", domain.ErrInvalidReindexRequest)
	}
	if ttl <= 0 {
		return false, errors.New("redis lock ttl must be greater than zero")
	}
	value, err := l.redis.do(ctx,
		"EVAL",
		"if redis.call('GET', KEYS[1]) == ARGV[1] then return redis.call('EXPIRE', KEYS[1], ARGV[2]) else return 0 end",
		"1",
		l.key,
		jobID,
		fmt.Sprintf("%d", secondsCeil(ttl)),
	)
	if err != nil {
		return false, fmt.Errorf("%w: refresh redis lock: %v", domain.ErrSearchBackendUnavailable, err)
	}
	count, err := value.intValue()
	if err != nil {
		return false, fmt.Errorf("%w: parse redis lock refresh response: %v", domain.ErrSearchBackendUnavailable, err)
	}
	return count == 1, nil
}

func (l *RedisReindexLock) Release(ctx context.Context, jobID string) error {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return fmt.Errorf("%w: job_id is required for lock release", domain.ErrInvalidReindexRequest)
	}
	if _, err := l.redis.do(ctx,
		"EVAL",
		"if redis.call('GET', KEYS[1]) == ARGV[1] then return redis.call('DEL', KEYS[1]) else return 0 end",
		"1",
		l.key,
		jobID,
	); err != nil {
		return fmt.Errorf("%w: release redis lock: %v", domain.ErrSearchBackendUnavailable, err)
	}
	return nil
}
