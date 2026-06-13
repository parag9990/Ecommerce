package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultLiveMetricsScanCount int64 = 500

type RedisLiveMetricsConfig struct {
	KeyPrefix string
}

type RedisLiveMetricsRepository struct {
	client    *redis.Client
	keyPrefix string
	logger    *slog.Logger
}

func NewRedisLiveMetricsRepository(client *redis.Client, cfg RedisLiveMetricsConfig, logger *slog.Logger) (*RedisLiveMetricsRepository, error) {
	if client == nil {
		return nil, errors.New("redis client is required")
	}
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &RedisLiveMetricsRepository{
		client:    client,
		keyPrefix: cfg.KeyPrefix,
		logger:    logger,
	}, nil
}

func (r *RedisLiveMetricsRepository) CountActiveUsers(ctx context.Context, window time.Duration) (int64, error) {
	if window <= 0 {
		return 0, errors.New("live metrics window must be greater than zero")
	}
	if err := r.trimActiveZSet(ctx, r.activeUsersKey(), window); err != nil {
		return 0, err
	}
	exists, err := r.client.Exists(ctx, r.activeUsersKey()).Result()
	if err != nil {
		return 0, fmt.Errorf("check active users key: %w", err)
	}
	if exists == 0 {
		return r.countActiveUserHashes(ctx)
	}
	count, err := r.client.ZCard(ctx, r.activeUsersKey()).Result()
	if err != nil {
		return 0, fmt.Errorf("count active users: %w", err)
	}
	return count, nil
}

func (r *RedisLiveMetricsRepository) countActiveUserHashes(ctx context.Context) (int64, error) {
	var cursor uint64
	pattern := r.keyPrefix + ":active:*"
	users := make(map[string]struct{})
	for {
		keys, next, err := r.client.Scan(ctx, cursor, pattern, defaultLiveMetricsScanCount).Result()
		if err != nil {
			return 0, fmt.Errorf("scan active user keys: %w", err)
		}
		for _, key := range keys {
			userID, err := r.client.HGet(ctx, key, "user_id").Result()
			if errors.Is(err, redis.Nil) {
				continue
			}
			if err != nil {
				return 0, fmt.Errorf("read active user hash: %w", err)
			}
			userID = strings.TrimSpace(userID)
			if userID != "" {
				users[userID] = struct{}{}
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return int64(len(users)), nil
}

func (r *RedisLiveMetricsRepository) CountActiveSessions(ctx context.Context, window time.Duration) (int64, error) {
	if window <= 0 {
		return 0, errors.New("live metrics window must be greater than zero")
	}
	if err := r.trimActiveZSet(ctx, r.activeSessionsKey(), window); err != nil {
		return 0, err
	}
	exists, err := r.client.Exists(ctx, r.activeSessionsKey()).Result()
	if err != nil {
		return 0, fmt.Errorf("check active sessions key: %w", err)
	}
	if exists > 0 {
		count, err := r.client.ZCard(ctx, r.activeSessionsKey()).Result()
		if err != nil {
			return 0, fmt.Errorf("count active sessions: %w", err)
		}
		return count, nil
	}
	return r.countActiveSessionHashes(ctx)
}

func (r *RedisLiveMetricsRepository) EventsPerMinute(ctx context.Context, window time.Duration) (float64, error) {
	if window <= 0 {
		return 0, errors.New("live metrics window must be greater than zero")
	}
	now := time.Now().UTC().Truncate(time.Minute)
	minutes := int(window/time.Minute) + 1
	keys := make([]string, 0, minutes)
	for i := 0; i < minutes; i++ {
		keys = append(keys, r.eventCounterKey(now.Add(-time.Duration(i)*time.Minute)))
	}
	values, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("read event counters: %w", err)
	}
	var total int64
	for _, value := range values {
		switch typed := value.(type) {
		case nil:
			continue
		case string:
			parsed, err := strconv.ParseInt(typed, 10, 64)
			if err != nil {
				r.logger.WarnContext(ctx, "session.redis.invalid_event_counter", slog.String("value", typed))
				continue
			}
			total += parsed
		case int64:
			total += typed
		}
	}
	minuteWindow := window.Minutes()
	if minuteWindow <= 0 {
		minuteWindow = 1
	}
	return float64(total) / minuteWindow, nil
}

func (r *RedisLiveMetricsRepository) trimActiveZSet(ctx context.Context, key string, window time.Duration) error {
	cutoff := time.Now().UTC().Add(-window).Unix()
	if err := r.client.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(cutoff, 10)).Err(); err != nil {
		return fmt.Errorf("trim active metrics key %s: %w", key, err)
	}
	return nil
}

func (r *RedisLiveMetricsRepository) countActiveSessionHashes(ctx context.Context) (int64, error) {
	var cursor uint64
	var count int64
	pattern := r.keyPrefix + ":active:*"
	for {
		keys, next, err := r.client.Scan(ctx, cursor, pattern, defaultLiveMetricsScanCount).Result()
		if err != nil {
			return 0, fmt.Errorf("scan active session keys: %w", err)
		}
		count += int64(len(keys))
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return count, nil
}

func (r *RedisLiveMetricsRepository) activeSessionsKey() string {
	return r.keyPrefix + ":active_sessions"
}

func (r *RedisLiveMetricsRepository) activeUsersKey() string {
	return r.keyPrefix + ":active_users"
}

func (r *RedisLiveMetricsRepository) eventCounterKey(at time.Time) string {
	return r.keyPrefix + ":counter:events:" + at.UTC().Format("200601021504")
}

func (c RedisLiveMetricsConfig) Validate() error {
	if strings.TrimSpace(c.KeyPrefix) == "" {
		return errors.New("SESSION_REDIS_KEY_PREFIX cannot be empty")
	}
	if strings.ContainsAny(c.KeyPrefix, " \t\r\n") {
		return errors.New("SESSION_REDIS_KEY_PREFIX cannot contain whitespace")
	}
	return nil
}

func (c RedisLiveMetricsConfig) withDefaults() RedisLiveMetricsConfig {
	c.KeyPrefix = strings.TrimSpace(c.KeyPrefix)
	if c.KeyPrefix == "" {
		c.KeyPrefix = defaultActiveSessionKeyPrefix
	}
	return c
}
