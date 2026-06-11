package usecase

import (
	"context"
	"log/slog"
	"testing"

	"product-service/internal/repository"
)

func TestCollectionSchemaServiceDescribesCollections(t *testing.T) {
	service := NewCollectionSchemaService(staticCollectionManager{}, slog.Default())

	descriptions, err := service.DescribeProductCollections(context.Background())
	if err != nil {
		t.Fatalf("DescribeProductCollections returned error: %v", err)
	}
	if len(descriptions) != 1 {
		t.Fatalf("description count = %d, want 1", len(descriptions))
	}
	if descriptions[0].Name != repository.CollectionProducts {
		t.Fatalf("collection name = %s, want %s", descriptions[0].Name, repository.CollectionProducts)
	}
}

func TestCollectionSchemaServiceEnsuresCollections(t *testing.T) {
	service := NewCollectionSchemaService(staticCollectionManager{}, slog.Default())

	result, err := service.EnsureProductCollections(context.Background())
	if err != nil {
		t.Fatalf("EnsureProductCollections returned error: %v", err)
	}
	if result.Database != repository.ProductDatabaseName {
		t.Fatalf("database = %s, want %s", result.Database, repository.ProductDatabaseName)
	}
	if len(result.Collections) != 1 || !result.Collections[0].ValidatorApplied {
		t.Fatalf("unexpected setup result: %+v", result.Collections)
	}
}

func TestCollectionSchemaServicePropagatesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service := NewCollectionSchemaService(staticCollectionManager{}, slog.Default())

	if _, err := service.EnsureProductCollections(ctx); err == nil {
		t.Fatal("expected context cancellation error")
	}
}

type staticCollectionManager struct{}

func (staticCollectionManager) EnsureCollections(context.Context) (repository.CollectionSetupResult, error) {
	return repository.CollectionSetupResult{
		Database: repository.ProductDatabaseName,
		Collections: []repository.CollectionSetupItem{
			{
				Name:             repository.CollectionProducts,
				Created:          true,
				ValidatorApplied: true,
				IndexNames:       []string{"idx_products_seller_status_updated"},
			},
		},
	}, nil
}

func (staticCollectionManager) DescribeCollections() []repository.CollectionDescription {
	return []repository.CollectionDescription{
		{
			Name:         repository.CollectionProducts,
			IndexNames:   []string{"idx_products_seller_status_updated"},
			HasValidator: true,
		},
	}
}
