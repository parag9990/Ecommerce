package indexer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
	searchevents "github.com/example/ecommerce-platform/backend/services/search-service/internal/events"
)

const (
	ActionUpsert = "upsert"
	ActionDelete = "delete"
)

type ProductIndexRepository interface {
	UpsertProduct(ctx context.Context, doc domain.ProductDocument) error
	DeleteProduct(ctx context.Context, productID string) error
}

type ProductIndexer struct {
	repo   ProductIndexRepository
	logger *slog.Logger
}

func NewProductIndexer(repo ProductIndexRepository, logger *slog.Logger) (*ProductIndexer, error) {
	if repo == nil {
		return nil, errors.New("product index repository is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &ProductIndexer{repo: repo, logger: logger}, nil
}

func (i *ProductIndexer) Handle(ctx context.Context, envelope searchevents.Envelope[searchevents.ProductIndexPayload]) (searchevents.ProductIndexResult, error) {
	startedAt := time.Now()
	result := searchevents.ProductIndexResult{
		EventID:   envelope.EventID,
		EventType: envelope.EventType,
		ProductID: envelope.Payload.ProductID,
	}

	if err := envelope.ValidateMetadata(); err != nil {
		i.logFailure(ctx, result, startedAt, err)
		return result, err
	}
	if err := ctx.Err(); err != nil {
		i.logFailure(ctx, result, startedAt, err)
		return result, err
	}

	payload := envelope.Payload.Normalized()
	result.ProductID = payload.ProductID
	if err := payload.ValidateForRouting(); err != nil {
		i.logFailure(ctx, result, startedAt, err)
		return result, err
	}

	if domain.IsProductDeleteEvent(envelope.EventType) {
		result.Action = ActionDelete
		if err := i.repo.DeleteProduct(ctx, payload.ProductID); err != nil {
			i.logFailure(ctx, result, startedAt, err)
			return result, err
		}
		i.logSuccess(ctx, envelope, result, startedAt)
		return result, nil
	}

	if !payload.IsSearchable() {
		result.Action = ActionDelete
		if err := i.repo.DeleteProduct(ctx, payload.ProductID); err != nil {
			i.logFailure(ctx, result, startedAt, err)
			return result, err
		}
		i.logSuccess(ctx, envelope, result, startedAt)
		return result, nil
	}

	if err := payload.ValidateForUpsert(); err != nil {
		i.logFailure(ctx, result, startedAt, err)
		return result, err
	}

	document := MapProductToDocument(payload, envelope.OccurredAt)
	if err := document.Validate(); err != nil {
		err = fmt.Errorf("%w: %v", domain.ErrInvalidProductEvent, err)
		i.logFailure(ctx, result, startedAt, err)
		return result, err
	}

	result.Action = ActionUpsert
	if err := i.repo.UpsertProduct(ctx, document); err != nil {
		i.logFailure(ctx, result, startedAt, err)
		return result, err
	}
	i.logSuccess(ctx, envelope, result, startedAt)
	return result, nil
}

func (i *ProductIndexer) logSuccess(ctx context.Context, envelope searchevents.Envelope[searchevents.ProductIndexPayload], result searchevents.ProductIndexResult, startedAt time.Time) {
	i.logger.InfoContext(ctx, "search.product_indexer.event_processed",
		slog.String("event_id", envelope.EventID),
		slog.String("event_type", envelope.EventType),
		slog.String("product_id", result.ProductID),
		slog.String("action", result.Action),
		slog.String("request_id", envelope.RequestID),
		slog.String("trace_id", envelope.TraceID),
		slog.Int64("duration_ms", time.Since(startedAt).Milliseconds()),
	)
}

func (i *ProductIndexer) logFailure(ctx context.Context, result searchevents.ProductIndexResult, startedAt time.Time, err error) {
	i.logger.ErrorContext(ctx, "search.product_indexer.event_failed",
		slog.String("event_id", result.EventID),
		slog.String("event_type", result.EventType),
		slog.String("product_id", result.ProductID),
		slog.String("action", result.Action),
		slog.String("error", err.Error()),
		slog.Int64("duration_ms", time.Since(startedAt).Milliseconds()),
	)
}
