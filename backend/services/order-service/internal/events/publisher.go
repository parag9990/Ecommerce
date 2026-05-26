package events

import (
	"context"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/domain"
)

type Publisher interface {
	Publish(ctx context.Context, topic string, event domain.OutboxEvent) error
}
