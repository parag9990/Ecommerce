package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
)

type CollectionManager interface {
	EnsureCartCollections(ctx context.Context) (domain.CollectionReport, error)
	Ping(ctx context.Context) error
}

type SchemaUsecase struct {
	manager CollectionManager
	logger  *slog.Logger
}

func NewSchemaUsecase(manager CollectionManager, logger *slog.Logger) (*SchemaUsecase, error) {
	if manager == nil {
		return nil, errors.New("collection manager is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &SchemaUsecase{manager: manager, logger: logger}, nil
}

func (u *SchemaUsecase) Spec(ctx context.Context) (domain.CollectionSpec, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-ctx.Done():
		return domain.CollectionSpec{}, ctx.Err()
	default:
	}
	return domain.CartCollectionSpec(), nil
}

func (u *SchemaUsecase) EnsureCollections(ctx context.Context) (domain.CollectionReport, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	report, err := u.manager.EnsureCartCollections(ctx)
	if err != nil {
		u.logger.Error("cart.schema.ensure_failed", slog.String("error", err.Error()))
		return report, err
	}
	u.logger.Info(
		"cart.schema.ensure_succeeded",
		slog.String("database", report.DatabaseName),
		slog.String("collection", report.CollectionName),
		slog.Bool("collection_created", report.CollectionCreated),
		slog.Bool("validator_updated", report.ValidatorUpdated),
		slog.Int("index_count", len(report.IndexesEnsured)),
	)
	return report, nil
}

func (u *SchemaUsecase) Ready(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return u.manager.Ping(ctx)
}
