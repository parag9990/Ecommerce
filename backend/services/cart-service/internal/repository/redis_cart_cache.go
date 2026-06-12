package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

type RedisCartCache struct {
	client         *redis.Client
	activeUserTTL  time.Duration
	activeGuestTTL time.Duration
	summaryTTL     time.Duration
}

func NewRedisCartCache(client *redis.Client, activeUserTTL time.Duration, activeGuestTTL time.Duration, summaryTTL time.Duration) (*RedisCartCache, error) {
	if client == nil {
		return nil, errors.New("redis client is required")
	}
	if activeUserTTL <= 0 || activeGuestTTL <= 0 || summaryTTL <= 0 {
		return nil, errors.New("redis cache ttl values must be greater than zero")
	}
	return &RedisCartCache{
		client:         client,
		activeUserTTL:  activeUserTTL,
		activeGuestTTL: activeGuestTTL,
		summaryTTL:     summaryTTL,
	}, nil
}

func (c *RedisCartCache) SetActiveCart(ctx context.Context, owner domain.CartOwner, cart *domain.Cart) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := owner.Validate(); err != nil {
		return err
	}
	if cart == nil {
		return fmt.Errorf("%w: cart is nil", domain.ErrInvalidCart)
	}
	payload, err := json.Marshal(cart)
	if err != nil {
		return fmt.Errorf("%w: marshal cart: %v", domain.ErrCacheUnavailable, err)
	}
	ttl := c.activeGuestTTL
	if owner.IsUser() {
		ttl = c.activeUserTTL
	}
	if err := c.client.Set(ctx, activeCartKey(owner), payload, ttl).Err(); err != nil {
		return fmt.Errorf("%w: set active cart: %v", domain.ErrCacheUnavailable, err)
	}
	return nil
}

func (c *RedisCartCache) SetCartSummary(ctx context.Context, owner domain.CartOwner, summary domain.CartSummary) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := owner.Validate(); err != nil {
		return err
	}
	payload, err := json.Marshal(summary)
	if err != nil {
		return fmt.Errorf("%w: marshal cart summary: %v", domain.ErrCacheUnavailable, err)
	}
	if err := c.client.Set(ctx, summaryKey(owner), payload, c.summaryTTL).Err(); err != nil {
		return fmt.Errorf("%w: set cart summary: %v", domain.ErrCacheUnavailable, err)
	}
	return nil
}

func (c *RedisCartCache) DeleteActiveCart(ctx context.Context, owner domain.CartOwner) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := owner.Validate(); err != nil {
		return err
	}
	if err := c.client.Del(ctx, activeCartKey(owner)).Err(); err != nil {
		return fmt.Errorf("%w: delete active cart: %v", domain.ErrCacheUnavailable, err)
	}
	return nil
}

func (c *RedisCartCache) DeleteCartSummary(ctx context.Context, owner domain.CartOwner) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := owner.Validate(); err != nil {
		return err
	}
	if err := c.client.Del(ctx, summaryKey(owner)).Err(); err != nil {
		return fmt.Errorf("%w: delete cart summary: %v", domain.ErrCacheUnavailable, err)
	}
	return nil
}

func activeCartKey(owner domain.CartOwner) string {
	if owner.IsUser() {
		return "cart:active:user:" + owner.UserID
	}
	return "cart:active:guest:" + owner.GuestSessionID
}

func summaryKey(owner domain.CartOwner) string {
	if owner.IsUser() {
		return "cart:summary:user:" + owner.UserID
	}
	return "cart:summary:guest:" + owner.GuestSessionID
}
