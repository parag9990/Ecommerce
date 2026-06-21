package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/schema"
	"github.com/typesense/typesense-go/v2/typesense"
)

func TestTypesenseAliasBootstrapAndLegacyMigrationIntegration(t *testing.T) {
	if os.Getenv("TYPESENSE_INTEGRATION") != "1" {
		t.Skip("set TYPESENSE_INTEGRATION=1 to run against a live Typesense instance")
	}

	endpoint := envOrDefault("TYPESENSE_URL", "http://127.0.0.1:8108")
	apiKey := envOrDefault("TYPESENSE_API_KEY", "local_typesense_key")
	client := typesense.NewClient(typesense.WithServer(endpoint), typesense.WithAPIKey(apiKey))
	collectionRepo, err := NewTypesenseCollectionRepository(client)
	if err != nil {
		t.Fatalf("collection repository: %v", err)
	}
	reindexRepo, err := NewTypesenseReindexRepository(client)
	if err != nil {
		t.Fatalf("reindex repository: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	stamp := time.Now().UTC().Format("20060102_150405_000000000")
	alias := "test_products_alias_" + stamp
	prefix := alias
	legacy := "test_products_legacy_" + stamp
	target := legacy + "_target"
	bootstrapTarget := prefix + "_20260621_120000"
	for _, name := range []string{bootstrapTarget, legacy, target} {
		name := name
		t.Cleanup(func() { _, _ = client.Collection(name).Delete(context.Background()) })
	}
	t.Cleanup(func() { _, _ = client.Alias(alias).Delete(context.Background()) })
	t.Cleanup(func() { _, _ = client.Alias(legacy).Delete(context.Background()) })

	productSchema := schema.ProductsCollectionSchema()
	bootstrapAt := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)
	active, err := collectionRepo.EnsureAliasedCollection(ctx, alias, prefix, productSchema, bootstrapAt)
	if err != nil {
		t.Fatalf("bootstrap alias: %v", err)
	}
	if active != bootstrapTarget {
		t.Fatalf("active collection = %q, want %q", active, bootstrapTarget)
	}
	resolved, err := reindexRepo.ResolveAlias(ctx, alias)
	if err != nil || resolved != active {
		t.Fatalf("resolved bootstrap alias = %q, err=%v", resolved, err)
	}

	legacySchema := productSchema
	legacySchema.Name = legacy
	if err := reindexRepo.EnsureCollection(ctx, legacySchema); err != nil {
		t.Fatalf("create legacy collection: %v", err)
	}
	targetSchema := productSchema
	targetSchema.Name = target
	if err := reindexRepo.EnsureCollection(ctx, targetSchema); err != nil {
		t.Fatalf("create target collection: %v", err)
	}

	synonymRepo, err := NewTypesenseSynonymRepository(client, legacy)
	if err != nil {
		t.Fatalf("synonym repository: %v", err)
	}
	synonym := domain.SearchSynonym{ID: "syn_mobile", Root: "mobile", Synonyms: []string{"phone", "smartphone"}}
	if _, err := synonymRepo.UpsertSynonym(ctx, synonym); err != nil {
		t.Fatalf("seed legacy synonym: %v", err)
	}
	previous, err := reindexRepo.ResolveAlias(ctx, legacy)
	if err != nil || previous != legacy {
		t.Fatalf("resolve legacy collection = %q, err=%v", previous, err)
	}
	if err := reindexRepo.CopySynonyms(ctx, legacy, target); err != nil {
		t.Fatalf("copy synonyms: %v", err)
	}
	if err := reindexRepo.SwapAlias(ctx, legacy, target); err != nil {
		t.Fatalf("migrate legacy collection to alias: %v", err)
	}
	resolved, err = reindexRepo.ResolveAlias(ctx, legacy)
	if err != nil || resolved != target {
		t.Fatalf("resolved migrated alias = %q, err=%v", resolved, err)
	}
	targetSynonyms, err := NewTypesenseSynonymRepository(client, target)
	if err != nil {
		t.Fatalf("target synonym repository: %v", err)
	}
	items, err := targetSynonyms.ListSynonyms(ctx, domain.SearchSynonymPageRequest{Page: 1, PageSize: 10})
	if err != nil || len(items) != 1 || items[0].ID != synonym.ID {
		t.Fatalf("target synonyms = %#v, err=%v", items, err)
	}
}

func envOrDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
