package repository

import (
	"context"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/schema"
)

func TestStaticSchemaRepositoryReturnsDefensiveCopies(t *testing.T) {
	collection, policy, synonyms := schema.MustProductSchemaContract()
	repo, err := NewStaticSchemaRepository(collection, policy, synonyms)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	firstCollection, err := repo.GetProductsCollectionSchema(context.Background())
	if err != nil {
		t.Fatalf("first collection: %v", err)
	}
	firstPolicy, err := repo.GetProductSearchPolicy(context.Background())
	if err != nil {
		t.Fatalf("first policy: %v", err)
	}
	firstSynonyms, err := repo.GetProductSynonymModel(context.Background())
	if err != nil {
		t.Fatalf("first synonyms: %v", err)
	}

	firstCollection.Fields[0].Name = "mutated"
	firstPolicy.Searchable[0].Name = "mutated"
	firstSynonyms.Examples[0].Synonyms[0] = "mutated"

	secondCollection, err := repo.GetProductsCollectionSchema(context.Background())
	if err != nil {
		t.Fatalf("second collection: %v", err)
	}
	secondPolicy, err := repo.GetProductSearchPolicy(context.Background())
	if err != nil {
		t.Fatalf("second policy: %v", err)
	}
	secondSynonyms, err := repo.GetProductSynonymModel(context.Background())
	if err != nil {
		t.Fatalf("second synonyms: %v", err)
	}

	if secondCollection.Fields[0].Name != domain.ProductFieldID {
		t.Fatalf("schema field mutation leaked: %q", secondCollection.Fields[0].Name)
	}
	if secondPolicy.Searchable[0].Name != domain.ProductFieldTitle {
		t.Fatalf("policy mutation leaked: %q", secondPolicy.Searchable[0].Name)
	}
	if secondSynonyms.Examples[0].Synonyms[0] != "phone" {
		t.Fatalf("synonym mutation leaked: %q", secondSynonyms.Examples[0].Synonyms[0])
	}
}
