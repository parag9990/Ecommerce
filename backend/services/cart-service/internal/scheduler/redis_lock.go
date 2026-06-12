package scheduler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type LockHandle interface {
	Release(ctx context.Context) error
}

type Locker interface {
	Acquire(ctx context.Context, key string, owner string, ttl time.Duration) (LockHandle, bool, error)
}

type RedisLocker struct {
	client *redis.Client
}

type redisLockHandle struct {
	client *redis.Client
	key    string
	owner  string
}

func NewRedisLocker(client *redis.Client) (*RedisLocker, error) {
	if client == nil {
		return nil, errors.New("redis client is required")
	}
	return &RedisLocker{client: client}, nil
}

func (l *RedisLocker) Acquire(ctx context.Context, key string, owner string, ttl time.Duration) (LockHandle, bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	key = strings.TrimSpace(key)
	owner = strings.TrimSpace(owner)
	if key == "" {
		return nil, false, errors.New("lock key is required")
	}
	if owner == "" {
		return nil, false, errors.New("lock owner is required")
	}
	if ttl <= 0 {
		return nil, false, errors.New("lock ttl must be greater than zero")
	}
	acquired, err := l.client.SetNX(ctx, key, owner, ttl).Result()
	if err != nil {
		return nil, false, fmt.Errorf("acquire redis lock: %w", err)
	}
	if !acquired {
		return nil, false, nil
	}
	return &redisLockHandle{client: l.client, key: key, owner: owner}, true, nil
}

func (h *redisLockHandle) Release(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	const releaseScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
  return redis.call("DEL", KEYS[1])
end
return 0
`
	if err := h.client.Eval(ctx, releaseScript, []string{h.key}, h.owner).Err(); err != nil {
		return fmt.Errorf("release redis lock: %w", err)
	}
	return nil
}
