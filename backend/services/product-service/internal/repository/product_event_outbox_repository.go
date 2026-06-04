package repository

import (
	"context"
	"time"

	"product-service/internal/domain"
)

type ProductFinder interface {
	FindProductByID(ctx context.Context, productID string) (*domain.Product, error)
}

type ProductEventOutboxWriter interface {
	InsertOutboxEvent(ctx context.Context, event *domain.ProductOutboxEvent) error
}

type ProductEventOutboxRepository interface {
	ProductEventOutboxWriter
	ListPendingOutboxEvents(ctx context.Context, limit int, now time.Time) ([]domain.ProductOutboxEvent, error)
	MarkOutboxEventPublishing(ctx context.Context, eventID string) error
	MarkOutboxEventPublished(ctx context.Context, eventID string, publishedAt time.Time) error
	MarkOutboxEventFailed(ctx context.Context, eventID string, nextAttemptAt time.Time, lastError string) error
	MarkOutboxEventDeadLettered(ctx context.Context, eventID string, lastError string) error
}

type ProductWriteRepositoryWithOutbox interface {
	InsertProductWithOutbox(ctx context.Context, product *domain.Product, event *domain.ProductOutboxEvent) error
	UpdateProductWithOutbox(ctx context.Context, product *domain.Product, event *domain.ProductOutboxEvent, expectedStatuses ...domain.ProductStatus) error
}
