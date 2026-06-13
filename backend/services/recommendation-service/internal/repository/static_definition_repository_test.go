package repository

import (
	"context"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
)

func TestStaticDefinitionRepositoryCompleteness(t *testing.T) {
	repo, err := NewStaticDefinitionRepository()
	if err != nil {
		t.Fatalf("NewStaticDefinitionRepository() error = %v", err)
	}

	types, err := repo.ListTypes(context.Background())
	if err != nil {
		t.Fatalf("ListTypes() error = %v", err)
	}
	if len(types) != len(domain.SupportedRecommendationTypes()) {
		t.Fatalf("ListTypes() len = %d, want %d", len(types), len(domain.SupportedRecommendationTypes()))
	}

	contexts, err := repo.ListContexts(context.Background())
	if err != nil {
		t.Fatalf("ListContexts() error = %v", err)
	}
	if len(contexts) != len(domain.SupportedRecommendationContexts()) {
		t.Fatalf("ListContexts() len = %d, want %d", len(contexts), len(domain.SupportedRecommendationContexts()))
	}
}

func TestStaticDefinitionRepositoryReturnsDefensiveCopies(t *testing.T) {
	repo, err := NewStaticDefinitionRepository()
	if err != nil {
		t.Fatalf("NewStaticDefinitionRepository() error = %v", err)
	}

	first, err := repo.GetType(context.Background(), domain.RecommendationTypeSimilarProducts)
	if err != nil {
		t.Fatalf("GetType() error = %v", err)
	}
	first.BestContexts[0] = domain.ContextCheckout

	second, err := repo.GetType(context.Background(), domain.RecommendationTypeSimilarProducts)
	if err != nil {
		t.Fatalf("GetType() error = %v", err)
	}
	if second.BestContexts[0] != domain.ContextProductDetail {
		t.Fatalf("repository returned mutable definition, best context = %s", second.BestContexts[0])
	}
}
