package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

type RedisRecommendationCacheConfig struct {
	KeyPrefix       string
	DefaultTTL      time.Duration
	PersonalizedTTL time.Duration
	GuestTTL        time.Duration
	RebuildLockTTL  time.Duration
	DirtyTTL        time.Duration
}

type RedisRecommendationCache struct {
	client     *redis.Client
	ttl        domain.CacheTTLPolicy
	lockTTL    time.Duration
	dirtyTTL   time.Duration
	keyBuilder domain.CacheKeyBuilder
	logger     *slog.Logger
}

func NewRedisRecommendationCache(client *redis.Client, cfg RedisRecommendationCacheConfig, logger *slog.Logger) (*RedisRecommendationCache, error) {
	if client == nil {
		return nil, errors.New("redis client is required")
	}
	if cfg.DefaultTTL <= 0 {
		cfg.DefaultTTL = domain.DefaultRecommendationCacheTTL
	}
	if cfg.PersonalizedTTL <= 0 {
		cfg.PersonalizedTTL = domain.DefaultPersonalizedCacheTTL
	}
	if cfg.GuestTTL <= 0 {
		cfg.GuestTTL = domain.DefaultGuestCacheTTL
	}
	if cfg.RebuildLockTTL <= 0 {
		cfg.RebuildLockTTL = time.Minute
	}
	if cfg.DirtyTTL <= 0 {
		cfg.DirtyTTL = domain.DefaultRecommendationCacheTTL
	}
	keyBuilder, err := domain.NewCacheKeyBuilder(cfg.KeyPrefix, 128)
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &RedisRecommendationCache{
		client: client,
		ttl: domain.CacheTTLPolicy{
			Default:      cfg.DefaultTTL,
			Personalized: cfg.PersonalizedTTL,
			Guest:        cfg.GuestTTL,
			RebuildLock:  cfg.RebuildLockTTL,
		},
		lockTTL:    cfg.RebuildLockTTL,
		dirtyTTL:   cfg.DirtyTTL,
		keyBuilder: keyBuilder,
		logger:     logger,
	}, nil
}

func (c *RedisRecommendationCache) Ping(ctx context.Context) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("%w: redis cache is not configured", domain.ErrRecommendationStorage)
	}
	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("%w: ping redis: %v", domain.ErrRecommendationStorage, err)
	}
	return nil
}

func (c *RedisRecommendationCache) Get(ctx context.Context, key string) (domain.CachedRecommendation, time.Duration, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return domain.CachedRecommendation{}, 0, fmt.Errorf("%w: cache key is required", domain.ErrInvalidCacheKey)
	}

	raw, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return domain.CachedRecommendation{}, 0, domain.ErrRecommendationCacheMiss
		}
		return domain.CachedRecommendation{}, 0, fmt.Errorf("%w: get cache key: %v", domain.ErrRecommendationStorage, err)
	}

	var cached domain.CachedRecommendation
	if err := json.Unmarshal([]byte(raw), &cached); err != nil {
		if delErr := c.Delete(ctx, key); delErr != nil {
			c.logger.WarnContext(ctx, "recommendation.cache_corrupt_delete_failed",
				slog.String("cache_key", key),
				slog.String("error", delErr.Error()),
			)
		}
		return domain.CachedRecommendation{}, 0, fmt.Errorf("%w: key=%s", domain.ErrRecommendationCacheCorrupt, key)
	}
	if err := cached.Validate(); err != nil {
		_ = c.Delete(ctx, key)
		return domain.CachedRecommendation{}, 0, fmt.Errorf("%w: key=%s", domain.ErrRecommendationCacheCorrupt, key)
	}

	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		c.logger.WarnContext(ctx, "recommendation.cache_ttl_failed",
			slog.String("cache_key", key),
			slog.String("error", err.Error()),
		)
	}
	return cached, ttl, nil
}

func (c *RedisRecommendationCache) Set(ctx context.Context, key string, cached domain.CachedRecommendation, ttl time.Duration) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("%w: cache key is required", domain.ErrInvalidCacheKey)
	}
	if err := cached.Validate(); err != nil {
		return err
	}
	if ttl <= 0 {
		ttl = c.ttl.TTLFor(cached.Type, strings.Contains(key, ":anon:"))
	}

	payload, err := json.Marshal(cached)
	if err != nil {
		return fmt.Errorf("%w: marshal cache payload: %v", domain.ErrRecommendationStorage, err)
	}
	if err := c.client.Set(ctx, key, payload, ttl).Err(); err != nil {
		return fmt.Errorf("%w: set cache key: %v", domain.ErrRecommendationStorage, err)
	}
	return nil
}

func (c *RedisRecommendationCache) Delete(ctx context.Context, keys ...string) error {
	trimmed := make([]string, 0, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key != "" {
			trimmed = append(trimmed, key)
		}
	}
	if len(trimmed) == 0 {
		return nil
	}
	if err := c.client.Del(ctx, trimmed...).Err(); err != nil {
		return fmt.Errorf("%w: delete cache key: %v", domain.ErrRecommendationStorage, err)
	}
	return nil
}

func (c *RedisRecommendationCache) InvalidateInteractions(ctx context.Context, interactions []domain.UserInteraction) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("%w: redis cache is not configured", domain.ErrRecommendationStorage)
	}
	if len(interactions) == 0 {
		return nil
	}

	deleteKeys := make([]string, 0)
	dirtyKeys := make([]string, 0)
	seenDelete := make(map[string]struct{})
	seenDirty := make(map[string]struct{})

	addDelete := func(key string, err error) {
		if err != nil {
			c.logger.DebugContext(ctx, "recommendation.cache.invalidate_key_skipped", slog.String("error", err.Error()))
			return
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return
		}
		if _, ok := seenDelete[key]; ok {
			return
		}
		seenDelete[key] = struct{}{}
		deleteKeys = append(deleteKeys, key)
	}
	addStrategyDelete := func(key string, err error, strategyID domain.StrategyID) {
		addDelete(key, err)
		if err != nil {
			return
		}
		scoped, scopedErr := c.keyBuilder.StrategyScopedKey(key, strategyID)
		addDelete(scoped, scopedErr)
	}
	addDirty := func(key string, err error) {
		if err != nil {
			c.logger.DebugContext(ctx, "recommendation.cache.dirty_key_skipped", slog.String("error", err.Error()))
			return
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return
		}
		if _, ok := seenDirty[key]; ok {
			return
		}
		seenDirty[key] = struct{}{}
		dirtyKeys = append(dirtyKeys, key)
	}

	for _, interaction := range interactions {
		if strings.TrimSpace(interaction.UserID) != "" {
			key, err := c.keyBuilder.PersonalizedUserKey(interaction.UserID)
			addStrategyDelete(key, err, domain.StrategyPersonalizedBehavior)
		}
		if strings.TrimSpace(interaction.AnonymousID) != "" {
			key, err := c.keyBuilder.PersonalizedAnonymousKey(interaction.AnonymousID)
			addStrategyDelete(key, err, domain.StrategyPersonalizedBehavior)
		}
		if strings.TrimSpace(interaction.ProductID) != "" {
			similarKey, err := c.keyBuilder.SimilarProductKey(interaction.ProductID)
			addStrategyDelete(similarKey, err, domain.StrategySimilarAttributes)
			fbtKey, err := c.keyBuilder.FrequentlyBoughtTogetherProductKey(interaction.ProductID)
			addStrategyDelete(fbtKey, err, domain.StrategyFBTOrderCooccurrence)
			addDirty(c.keyBuilder.DirtyProductKey(interaction.ProductID))
		}
		if strings.TrimSpace(interaction.CategoryID) != "" {
			addDirty(c.keyBuilder.DirtyCategoryKey(interaction.CategoryID))
		}
		if strings.TrimSpace(interaction.SellerID) != "" {
			addDirty(c.keyBuilder.DirtySellerKey(interaction.SellerID))
		}
	}

	if len(deleteKeys) > 0 {
		if err := c.Delete(ctx, deleteKeys...); err != nil {
			return err
		}
	}
	if len(dirtyKeys) > 0 {
		pipe := c.client.Pipeline()
		for _, key := range dirtyKeys {
			pipe.Set(ctx, key, "1", c.dirtyTTL)
		}
		if _, err := pipe.Exec(ctx); err != nil {
			return fmt.Errorf("%w: mark dirty cache keys: %v", domain.ErrRecommendationStorage, err)
		}
	}
	return nil
}

func (c *RedisRecommendationCache) AcquireRebuildLock(ctx context.Context, lockKey string) (bool, error) {
	lockKey = strings.TrimSpace(lockKey)
	if lockKey == "" {
		return false, fmt.Errorf("%w: lock key is required", domain.ErrInvalidCacheKey)
	}
	ok, err := c.client.SetNX(ctx, lockKey, "1", c.lockTTL).Result()
	if err != nil {
		return false, fmt.Errorf("%w: acquire rebuild lock: %v", domain.ErrRecommendationStorage, err)
	}
	return ok, nil
}
