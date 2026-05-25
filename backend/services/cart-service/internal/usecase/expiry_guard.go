package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
)

type activeCartExpirer interface {
	ExpireActiveCart(ctx context.Context, cartID string, expectedVersion int64, now time.Time) (bool, error)
}

type activeCartInvalidator interface {
	DeleteActiveCart(ctx context.Context, owner domain.CartOwner) error
	DeleteCartSummary(ctx context.Context, owner domain.CartOwner) error
}

func expireActiveCartForOwner(ctx context.Context, repository activeCartExpirer, cache activeCartInvalidator, logger *slog.Logger, owner domain.CartOwner, cart *domain.Cart, now time.Time, operation string) error {
	if logger == nil {
		logger = slog.Default()
	}
	expired, err := repository.ExpireActiveCart(ctx, cart.ID, cart.Version, now)
	if err != nil {
		return err
	}
	if !expired {
		return domain.ErrCartVersionConflict
	}
	invalidateExpiredCartCache(ctx, cache, logger, owner, cart.ID, operation)
	logger.Info(
		"cart.expiry.active_cart_marked",
		slog.String("cart_id", cart.ID),
		slog.String("owner_type", string(owner.Type)),
		slog.String("operation", operation),
	)
	return nil
}

func invalidateExpiredCartCache(ctx context.Context, cache activeCartInvalidator, logger *slog.Logger, owner domain.CartOwner, cartID string, operation string) int {
	if logger == nil {
		logger = slog.Default()
	}
	failures := 0
	if err := cache.DeleteActiveCart(ctx, owner); err != nil {
		failures++
		logger.Warn("cart.expiry.cache_delete_failed", slog.String("operation", operation), slog.String("cache_key_type", "active"), slog.String("cart_id", cartID), slog.String("error", err.Error()))
	}
	if err := cache.DeleteCartSummary(ctx, owner); err != nil {
		failures++
		logger.Warn("cart.expiry.cache_delete_failed", slog.String("operation", operation), slog.String("cache_key_type", "summary"), slog.String("cart_id", cartID), slog.String("error", err.Error()))
	}
	return failures
}
