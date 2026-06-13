package repository

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/typesense/typesense-go/v2/typesense"
	"github.com/typesense/typesense-go/v2/typesense/api"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

func TestTypesenseSynonymRepositoryIntegration(t *testing.T) {
	if os.Getenv("TYPESENSE_INTEGRATION") != "1" {
		t.Skip("set TYPESENSE_INTEGRATION=1 to run against a live Typesense instance")
	}

	apiKey := strings.TrimSpace(os.Getenv("TYPESENSE_API_KEY"))
	if apiKey == "" {
		t.Fatal("TYPESENSE_API_KEY is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := "synonym_it_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	client := typesense.NewClient(
		typesense.WithServer(typesenseIntegrationEndpoint()),
		typesense.WithAPIKey(apiKey),
		typesense.WithConnectionTimeout(2*time.Second),
	)

	_, err := client.Collections().Create(ctx, &api.CollectionSchema{
		Name: collection,
		Fields: []api.Field{
			{Name: "title", Type: "string"},
		},
	})
	if err != nil {
		t.Fatalf("create collection: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cancelCleanup := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelCleanup()
		_, _ = client.Collection(collection).Delete(cleanupCtx)
	})

	repo, err := NewTypesenseSynonymRepository(client, collection)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}

	synonym, err := domain.NewSearchSynonym(domain.SearchSynonymInput{
		Root:     "mobile",
		Synonyms: []string{"phone", "smartphone"},
	})
	if err != nil {
		t.Fatalf("new synonym: %v", err)
	}

	if _, err := repo.UpsertSynonym(ctx, synonym); err != nil {
		t.Fatalf("upsert synonym: %v", err)
	}
	got, err := repo.ListSynonyms(ctx, domain.SearchSynonymPageRequest{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list synonyms: %v", err)
	}
	if len(got) != 1 || got[0].ID != "syn_mobile" || got[0].Root != "mobile" {
		t.Fatalf("synonyms = %#v", got)
	}
}

func typesenseIntegrationEndpoint() string {
	if endpoint := strings.TrimSpace(os.Getenv("TYPESENSE_URL")); endpoint != "" {
		return strings.TrimRight(endpoint, "/")
	}
	protocol := strings.TrimSpace(os.Getenv("TYPESENSE_PROTOCOL"))
	if protocol == "" {
		protocol = "http"
	}
	host := strings.TrimSpace(os.Getenv("TYPESENSE_HOST"))
	if host == "" {
		host = "localhost"
	}
	port := strings.TrimSpace(os.Getenv("TYPESENSE_PORT"))
	if port == "" {
		port = "8108"
	}
	return protocol + "://" + host + ":" + port
}
