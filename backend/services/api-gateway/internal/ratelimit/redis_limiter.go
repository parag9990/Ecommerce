package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLimiter struct {
	client *redis.Client
	script *redis.Script
	now    func() time.Time
}

func NewRedisLimiter(client *redis.Client) *RedisLimiter {
	return &RedisLimiter{
		client: client,
		script: redis.NewScript(tokenBucketScript),
		now:    time.Now,
	}
}

func (l *RedisLimiter) Allow(ctx context.Context, req Request) (Result, error) {
	if err := validateRequest(req); err != nil {
		return Result{}, err
	}
	if l == nil || l.client == nil || l.script == nil {
		return Result{}, fmt.Errorf("redis limiter is not configured")
	}
	if req.Cost == 0 {
		req.Cost = 1
	}

	windowMs := req.Window.Milliseconds()
	refillPerMs := float64(req.Limit) / float64(windowMs)
	ttl := req.Window * 2
	if ttl < time.Second {
		ttl = time.Second
	}

	values, err := l.script.Run(ctx, l.client, []string{req.Key},
		req.Limit,
		refillPerMs,
		l.now().UnixMilli(),
		req.Cost,
		ttl.Milliseconds(),
	).Slice()
	if err != nil {
		return Result{}, err
	}
	if len(values) != 4 {
		return Result{}, fmt.Errorf("unexpected redis rate limit result length %d", len(values))
	}

	allowed, err := toInt64(values[0])
	if err != nil {
		return Result{}, fmt.Errorf("parse allowed result: %w", err)
	}
	remaining, err := toInt64(values[1])
	if err != nil {
		return Result{}, fmt.Errorf("parse remaining result: %w", err)
	}
	retryAfterMs, err := toInt64(values[2])
	if err != nil {
		return Result{}, fmt.Errorf("parse retry-after result: %w", err)
	}
	resetAfterMs, err := toInt64(values[3])
	if err != nil {
		return Result{}, fmt.Errorf("parse reset result: %w", err)
	}

	return Result{
		Allowed:    allowed == 1,
		Limit:      req.Limit,
		Remaining:  maxInt64(0, remaining),
		RetryAfter: time.Duration(maxInt64(0, retryAfterMs)) * time.Millisecond,
		ResetAfter: time.Duration(maxInt64(0, resetAfterMs)) * time.Millisecond,
	}, nil
}

func validateRequest(req Request) error {
	if req.Key == "" {
		return fmt.Errorf("%w: key is required", ErrInvalidRequest)
	}
	if req.Limit <= 0 {
		return fmt.Errorf("%w: limit must be positive", ErrInvalidRequest)
	}
	if req.Window <= 0 {
		return fmt.Errorf("%w: window must be positive", ErrInvalidRequest)
	}
	if req.Window.Milliseconds() <= 0 {
		return fmt.Errorf("%w: window must be at least 1ms", ErrInvalidRequest)
	}
	if req.Cost < 0 {
		return fmt.Errorf("%w: cost must not be negative", ErrInvalidRequest)
	}
	if req.Cost > req.Limit {
		return fmt.Errorf("%w: cost must not exceed limit", ErrInvalidRequest)
	}
	return nil
}

func toInt64(value any) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case uint64:
		if v > uint64(^uint64(0)>>1) {
			return 0, fmt.Errorf("uint64 value overflows int64")
		}
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	case []byte:
		return strconv.ParseInt(string(v), 10, 64)
	default:
		return 0, fmt.Errorf("unsupported type %T", value)
	}
}

func maxInt64(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
