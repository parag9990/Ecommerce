package events

import (
	"context"

	"ecommerce/backend/services/wishlist-service/internal/domain"
)

type WishlistAnalyticsPublisher interface {
	Publish(ctx context.Context, event domain.WishlistAnalyticsEvent) error
	Close(ctx context.Context) error
}
