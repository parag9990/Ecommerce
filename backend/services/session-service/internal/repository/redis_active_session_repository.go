package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func NewRedisClient(config RedisConfig) (*redis.Client, error) {
	if strings.TrimSpace(config.Addr) == "" {
		return nil, errors.New("redis address is required")
	}
	return redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}), nil
}

type RedisActiveSessionRepository struct {
	client *redis.Client
	prefix string
}

func NewRedisActiveSessionRepository(client *redis.Client, prefix string) (*RedisActiveSessionRepository, error) {
	if client == nil {
		return nil, errors.New("redis client is required")
	}
	prefix = strings.Trim(strings.TrimSpace(prefix), ":")
	if prefix == "" {
		prefix = "session"
	}
	return &RedisActiveSessionRepository{client: client, prefix: prefix}, nil
}

func (r *RedisActiveSessionRepository) PreviewDeletion(ctx context.Context, target domain.DeletionTarget) (int64, error) {
	keys, err := r.matchingKeys(ctx, target)
	if err != nil {
		return 0, err
	}
	return int64(len(keys)), nil
}

func (r *RedisActiveSessionRepository) DeleteMatching(ctx context.Context, target domain.DeletionTarget) (int64, error) {
	keys, err := r.matchingKeys(ctx, target)
	if err != nil {
		return 0, err
	}
	if len(keys) == 0 {
		return 0, nil
	}
	return r.client.Del(ctx, keys...).Result()
}

func (r *RedisActiveSessionRepository) matchingKeys(ctx context.Context, target domain.DeletionTarget) ([]string, error) {
	if target.Type == domain.DeletionTargetSessionID {
		return r.existingSessionKeys(ctx, target.Value)
	}

	fields := []string{redisTargetField(target.Type)}
	candidates, err := r.scanActiveSessionKeys(ctx)
	if err != nil {
		return nil, err
	}

	matches := make([]string, 0)
	for _, key := range candidates {
		for _, field := range fields {
			value, err := r.client.HGet(ctx, key, field).Result()
			if errors.Is(err, redis.Nil) {
				continue
			}
			if err != nil {
				return nil, err
			}
			if value == target.Value {
				matches = append(matches, key)
				break
			}
		}
	}
	return matches, nil
}

func (r *RedisActiveSessionRepository) existingSessionKeys(ctx context.Context, sessionID string) ([]string, error) {
	candidates := []string{
		r.prefix + ":active_session:" + sessionID,
		r.prefix + ":active:" + sessionID,
		r.prefix + ":session:" + sessionID,
	}
	keys := make([]string, 0, len(candidates))
	for _, key := range candidates {
		exists, err := r.client.Exists(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		if exists > 0 {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

func (r *RedisActiveSessionRepository) scanActiveSessionKeys(ctx context.Context) ([]string, error) {
	patterns := []string{
		r.prefix + ":active_session:*",
		r.prefix + ":active:*",
	}
	seen := make(map[string]struct{})
	keys := make([]string, 0)
	for _, pattern := range patterns {
		var cursor uint64
		for {
			batch, next, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
			if err != nil {
				return nil, err
			}
			for _, key := range batch {
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				keys = append(keys, key)
			}
			cursor = next
			if cursor == 0 {
				break
			}
		}
	}
	return keys, nil
}

func redisTargetField(targetType domain.DeletionTargetType) string {
	switch targetType {
	case domain.DeletionTargetUserID:
		return "user_id"
	case domain.DeletionTargetAnonymousID:
		return "anonymous_id"
	default:
		return "session_id"
	}
}
