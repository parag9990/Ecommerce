package repository

import (
	"errors"
	"net/http"
	"testing"

	"github.com/typesense/typesense-go/v2/typesense"
	"github.com/typesense/typesense-go/v2/typesense/api"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

func TestMapTypesenseSynonymUsesFallbackForUpsertResponse(t *testing.T) {
	id := "syn_mobile"
	root := "Mobile"
	item := &api.SearchSynonym{
		Id:       &id,
		Root:     &root,
		Synonyms: []string{" Phone ", "SMARTPHONE"},
	}

	got := mapTypesenseSynonym(item, domain.SearchSynonym{})
	if got.ID != "syn_mobile" || got.Root != "mobile" {
		t.Fatalf("mapped synonym = %#v", got)
	}
	if len(got.Synonyms) != 2 || got.Synonyms[0] != "phone" || got.Synonyms[1] != "smartphone" {
		t.Fatalf("mapped synonyms = %#v", got.Synonyms)
	}
}

func TestMapTypesenseSynonymFallsBackWhenResponseOmitsFields(t *testing.T) {
	fallback := domain.SearchSynonym{
		ID:       "syn_mobile",
		Root:     "mobile",
		Synonyms: []string{"phone"},
	}

	got := mapTypesenseSynonym(&api.SearchSynonym{}, fallback)
	if got.ID != fallback.ID || got.Root != fallback.Root || got.Synonyms[0] != "phone" {
		t.Fatalf("mapped synonym = %#v", got)
	}
}

func TestMapTypesenseSynonymErrorPreservesDomainSentinel(t *testing.T) {
	notFound := &typesense.HTTPError{Status: http.StatusNotFound}
	err := mapTypesenseSynonymError("list synonyms", notFound)
	if !errors.Is(err, domain.ErrSearchCollectionUnavailable) {
		t.Fatalf("expected collection unavailable, got %v", err)
	}

	unauthorized := &typesense.HTTPError{Status: http.StatusUnauthorized}
	err = mapTypesenseSynonymError("upsert synonym", unauthorized)
	if !errors.Is(err, domain.ErrSearchBackendUnavailable) {
		t.Fatalf("expected backend unavailable, got %v", err)
	}
}
