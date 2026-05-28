package usecase

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/repository"
)

func TestDefinitionServiceResolveSimilarProducts(t *testing.T) {
	service := newTestDefinitionService(t)

	resolution, err := service.Resolve(context.Background(), ResolveInput{
		RequestID: "req_123",
		Context:   domain.ContextProductDetail,
		ProductID: "prod_123",
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if resolution.Type != domain.RecommendationTypeSimilarProducts {
		t.Fatalf("Type = %s, want %s", resolution.Type, domain.RecommendationTypeSimilarProducts)
	}
	if resolution.StrategyID != domain.StrategySimilarAttributes {
		t.Fatalf("StrategyID = %s, want %s", resolution.StrategyID, domain.StrategySimilarAttributes)
	}
	if resolution.Request.Limit != 12 {
		t.Fatalf("Limit = %d, want 12", resolution.Request.Limit)
	}
	if len(resolution.Fallbacks) == 0 {
		t.Fatal("fallbacks are empty")
	}
}

func TestDefinitionServiceResolveCartRequiresProductSignal(t *testing.T) {
	service := newTestDefinitionService(t)

	_, err := service.Resolve(context.Background(), ResolveInput{
		Context: domain.ContextCart,
	})
	if !errors.Is(err, domain.ErrInvalidRecommendationRequest) {
		t.Fatalf("Resolve() error = %v, want %v", err, domain.ErrInvalidRecommendationRequest)
	}
}

func TestDefinitionServiceRejectsOverLimit(t *testing.T) {
	service := newTestDefinitionService(t)

	_, err := service.Resolve(context.Background(), ResolveInput{
		Context: domain.ContextHomeFeed,
		Type:    domain.RecommendationTypeTrending,
		Limit:   101,
	})
	if !errors.Is(err, domain.ErrInvalidLimit) {
		t.Fatalf("Resolve() error = %v, want %v", err, domain.ErrInvalidLimit)
	}
}

func newTestDefinitionService(t *testing.T) *DefinitionService {
	t.Helper()

	repo, err := repository.NewStaticDefinitionRepository()
	if err != nil {
		t.Fatalf("NewStaticDefinitionRepository() error = %v", err)
	}
	service, err := NewDefinitionService(repo, DefinitionServiceConfig{
		DefaultLimit:        12,
		MaxLimit:            100,
		MaxIdentifierLength: 128,
	}, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
	if err != nil {
		t.Fatalf("NewDefinitionService() error = %v", err)
	}
	return service
}
