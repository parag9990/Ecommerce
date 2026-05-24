package usecase

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/repository"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/schema"
)

func TestSchemaUsecaseReturnsProductSchemaContract(t *testing.T) {
	collection, policy, synonyms := schema.MustProductSchemaContract()
	repo, err := repository.NewStaticSchemaRepository(collection, policy, synonyms)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	uc, err := NewSchemaUsecase(repo, slog.Default())
	if err != nil {
		t.Fatalf("new usecase: %v", err)
	}

	contract, err := uc.GetProductSchemaContract(context.Background())
	if err != nil {
		t.Fatalf("contract: %v", err)
	}
	if contract.Version != domain.ProductSchemaContractVersion {
		t.Fatalf("version = %q", contract.Version)
	}
	if contract.Collection.Name != domain.ProductsCollectionName {
		t.Fatalf("collection = %q", contract.Collection.Name)
	}
	if contract.TypesenseDefinition["name"] != domain.ProductsCollectionName {
		t.Fatalf("typesense definition name = %#v", contract.TypesenseDefinition["name"])
	}
}

func TestSchemaUsecasePreservesCanceledContext(t *testing.T) {
	collection, policy, synonyms := schema.MustProductSchemaContract()
	repo, err := repository.NewStaticSchemaRepository(collection, policy, synonyms)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	uc, err := NewSchemaUsecase(repo, slog.Default())
	if err != nil {
		t.Fatalf("new usecase: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = uc.GetProductSchemaContract(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if !errors.Is(err, domain.ErrSchemaUnavailable) {
		t.Fatalf("expected ErrSchemaUnavailable, got %v", err)
	}
}
