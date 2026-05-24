package usecase

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

func TestAutocompleteUsecaseReturnsCacheHit(t *testing.T) {
	uc, repo, cache := newAutocompleteTestUsecase(t)
	cache.getOK = true
	cache.getSuggestions = []string{"shoes", "shoe rack"}

	resp, err := uc.Execute(context.Background(), domain.AutocompleteRequest{Query: "sho", Limit: 2})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(resp.Suggestions) != 2 || resp.Suggestions[0] != "shoes" {
		t.Fatalf("suggestions = %#v", resp.Suggestions)
	}
	if repo.popularCalls != 0 || repo.productCalls != 0 {
		t.Fatalf("repository should not be called on cache hit: popular=%d product=%d", repo.popularCalls, repo.productCalls)
	}
}

func TestAutocompleteUsecaseSkipsProductLookupForShortQuery(t *testing.T) {
	uc, repo, _ := newAutocompleteTestUsecase(t)
	repo.popularCandidates = []domain.SuggestionCandidate{
		{Text: "shoes", Source: domain.AutocompleteSourcePopularQuery, Score: 100},
	}

	resp, err := uc.Execute(context.Background(), domain.AutocompleteRequest{Query: "s", Limit: 5})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if repo.popularCalls != 1 {
		t.Fatalf("popular calls = %d", repo.popularCalls)
	}
	if repo.productCalls != 0 {
		t.Fatalf("product calls = %d", repo.productCalls)
	}
	if len(resp.Suggestions) != 1 || resp.Suggestions[0] != "shoes" {
		t.Fatalf("suggestions = %#v", resp.Suggestions)
	}
}

func TestAutocompleteUsecaseFetchesProductAndPopularCandidates(t *testing.T) {
	uc, repo, _ := newAutocompleteTestUsecase(t)
	repo.popularCandidates = []domain.SuggestionCandidate{
		{Text: "shoes", Source: domain.AutocompleteSourcePopularQuery, Score: 100},
	}
	repo.productCandidates = []domain.SuggestionCandidate{
		{Text: "Nike Running Shoes", Source: domain.AutocompleteSourceProductTitle, Score: 80},
	}

	resp, err := uc.Execute(context.Background(), domain.AutocompleteRequest{Query: "sho", Limit: 5})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if repo.popularQuery != "sho" || repo.productQuery != "sho" {
		t.Fatalf("queries = popular %q product %q", repo.popularQuery, repo.productQuery)
	}
	if len(resp.Suggestions) != 2 {
		t.Fatalf("suggestions = %#v", resp.Suggestions)
	}
}

func TestAutocompleteUsecaseIgnoresCacheErrors(t *testing.T) {
	uc, repo, cache := newAutocompleteTestUsecase(t)
	cache.getErr = errors.New("redis down")
	cache.setErr = errors.New("redis still down")
	repo.popularCandidates = []domain.SuggestionCandidate{
		{Text: "shoes", Source: domain.AutocompleteSourcePopularQuery, Score: 100},
	}

	resp, err := uc.Execute(context.Background(), domain.AutocompleteRequest{Query: "sho", Limit: 5})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(resp.Suggestions) == 0 {
		t.Fatalf("expected suggestions despite cache errors")
	}
}

func TestAutocompleteUsecaseMapsTypesenseErrors(t *testing.T) {
	uc, repo, _ := newAutocompleteTestUsecase(t)
	repo.popularErr = errors.New("typesense down")

	_, err := uc.Execute(context.Background(), domain.AutocompleteRequest{Query: "sho", Limit: 5})
	if !errors.Is(err, domain.ErrSearchBackendUnavailable) {
		t.Fatalf("expected search backend unavailable, got %v", err)
	}
}

func TestAutocompleteUsecaseRejectsInvalidRequest(t *testing.T) {
	uc, _, _ := newAutocompleteTestUsecase(t)

	_, err := uc.Execute(context.Background(), domain.AutocompleteRequest{Limit: 99})
	if !errors.Is(err, domain.ErrInvalidAutocompleteRequest) {
		t.Fatalf("expected invalid autocomplete request, got %v", err)
	}
}

func newAutocompleteTestUsecase(t *testing.T) (*AutocompleteUsecase, *fakeAutocompleteRepo, *fakeAutocompleteCache) {
	t.Helper()
	repo := &fakeAutocompleteRepo{}
	cache := &fakeAutocompleteCache{}
	uc, err := NewAutocompleteUsecase(repo, cache, AutocompleteOptions{
		DefaultLimit:       8,
		MaxLimit:           10,
		TypesenseTimeout:   time.Second,
		CacheTimeout:       time.Second,
		PrefixCacheTTL:     time.Minute,
		EmptyQueryCacheTTL: 5 * time.Minute,
	}, slog.Default())
	if err != nil {
		t.Fatalf("usecase: %v", err)
	}
	return uc, repo, cache
}

type fakeAutocompleteRepo struct {
	popularCalls      int
	productCalls      int
	popularQuery      string
	productQuery      string
	popularCandidates []domain.SuggestionCandidate
	productCandidates []domain.SuggestionCandidate
	popularErr        error
	productErr        error
}

func (r *fakeAutocompleteRepo) ProductPrefixCandidates(_ context.Context, query string, _ int) ([]domain.SuggestionCandidate, error) {
	r.productCalls++
	r.productQuery = query
	if r.productErr != nil {
		return nil, r.productErr
	}
	return r.productCandidates, nil
}

func (r *fakeAutocompleteRepo) PopularQueryCandidates(_ context.Context, query string, _ int) ([]domain.SuggestionCandidate, error) {
	r.popularCalls++
	r.popularQuery = query
	if r.popularErr != nil {
		return nil, r.popularErr
	}
	return r.popularCandidates, nil
}

type fakeAutocompleteCache struct {
	getOK          bool
	getSuggestions []string
	getErr         error
	setErr         error
	setKey         string
	setTTL         time.Duration
}

func (c *fakeAutocompleteCache) Get(context.Context, string) ([]string, bool, error) {
	if c.getErr != nil {
		return nil, false, c.getErr
	}
	return c.getSuggestions, c.getOK, nil
}

func (c *fakeAutocompleteCache) Set(_ context.Context, key string, suggestions []string, ttl time.Duration) error {
	c.setKey = key
	c.setTTL = ttl
	if c.setErr != nil {
		return c.setErr
	}
	c.getSuggestions = append([]string(nil), suggestions...)
	return nil
}
