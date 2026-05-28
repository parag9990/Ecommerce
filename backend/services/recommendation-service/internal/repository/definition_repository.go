package repository

import (
	"context"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
)

type DefinitionRepository interface {
	ListTypes(ctx context.Context) ([]domain.TypeDefinition, error)
	GetType(ctx context.Context, typ domain.RecommendationType) (domain.TypeDefinition, error)
	ListContexts(ctx context.Context) ([]domain.ContextDefinition, error)
	GetContext(ctx context.Context, context domain.RecommendationContext) (domain.ContextDefinition, error)
	FallbackChain(ctx context.Context) ([]domain.FallbackStep, error)
}
