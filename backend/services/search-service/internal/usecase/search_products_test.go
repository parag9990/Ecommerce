package usecase

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/repository"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/requestctx"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/schema"
)

func TestSearchProductsUsecasePreservesTypesenseRanking(t *testing.T) {
	uc, searchRepo, hydrator := newSearchProductsTestUsecase(t)
	searchRepo.result = domain.ProductSearchResult{
		IDs: []string{"prod_2", "prod_1", "prod_missing"},
		Facets: map[string][]domain.FacetValue{
			domain.ProductFieldBrand:       {{Value: "Nike", Count: 2}},
			domain.ProductFieldCategoryIDs: {{Value: "cat_shoes", Count: 2}},
			"private_field":                {{Value: "secret", Count: 99}},
		},
		Total:        3,
		SearchTimeMS: 12,
	}
	hydrator.products = []domain.ProductSummary{
		{ProductID: "prod_1", Title: "First"},
		{ProductID: "prod_2", Title: "Second"},
	}

	resp, err := uc.Execute(context.Background(), domain.SearchRequest{
		Query: " running   shoes ",
		Filters: map[string]string{
			domain.SearchFilterBrand:    "Nike,Puma",
			domain.SearchFilterInStock:  "true",
			domain.SearchFilterMinPrice: "1000",
		},
		Sort:     domain.SortPriceAsc,
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	if searchRepo.query.Query != "running shoes" {
		t.Fatalf("query = %q", searchRepo.query.Query)
	}
	if searchRepo.query.SortBy != "price:asc" {
		t.Fatalf("sort_by = %q", searchRepo.query.SortBy)
	}
	if searchRepo.query.FilterBy != "brand:=[`Nike`,`Puma`] && price:>=1000 && in_stock:=true" {
		t.Fatalf("filter_by = %q", searchRepo.query.FilterBy)
	}
	if len(hydrator.ids) != 3 || hydrator.ids[0] != "prod_2" {
		t.Fatalf("hydrator ids = %#v", hydrator.ids)
	}
	if len(resp.Products) != 2 || resp.Products[0].ProductID != "prod_2" || resp.Products[1].ProductID != "prod_1" {
		t.Fatalf("products were not ranked order: %#v", resp.Products)
	}
	if _, ok := resp.Facets["private_field"]; ok {
		t.Fatalf("unexpected private facet: %#v", resp.Facets)
	}
	if resp.Total != 3 {
		t.Fatalf("total = %d", resp.Total)
	}
}

func TestSearchProductsUsecaseReturnsEmptyResultsWithoutHydration(t *testing.T) {
	uc, searchRepo, hydrator := newSearchProductsTestUsecase(t)
	searchRepo.result = domain.ProductSearchResult{IDs: []string{}, Facets: map[string][]domain.FacetValue{}, Total: 0}

	resp, err := uc.Execute(context.Background(), domain.SearchRequest{})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(resp.Products) != 0 || resp.Total != 0 {
		t.Fatalf("unexpected response: %#v", resp)
	}
	if hydrator.called {
		t.Fatal("hydrator should not be called for empty hits")
	}
}

func TestSearchProductsUsecaseTracksZeroResults(t *testing.T) {
	tracker := &capturingZeroResultTracker{}
	uc, searchRepo, _ := newSearchProductsTestUsecaseWithTracker(t, tracker)
	searchRepo.result = domain.ProductSearchResult{IDs: []string{}, Facets: map[string][]domain.FacetValue{}, Total: 0}
	ctx := requestctx.WithRequestID(context.Background(), "req_ctx")
	ctx = requestctx.WithAnalyticsContext(ctx, requestctx.AnalyticsContext{
		RequestID:   "req_1",
		AnonymousID: "anon_1",
		SessionID:   "sess_1",
		UserID:      "user_1",
		ClientPath:  "/search",
	})

	resp, err := uc.Execute(ctx, domain.SearchRequest{
		Query: " Waterproof   Laptop Bag ",
		Filters: map[string]string{
			domain.SearchFilterBrand:    "Acme",
			domain.SearchFilterMinPrice: "100",
		},
		Sort:     domain.SortRelevance,
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if resp.Total != 0 {
		t.Fatalf("total = %d", resp.Total)
	}
	if len(tracker.events) != 1 {
		t.Fatalf("tracker events = %d", len(tracker.events))
	}
	event := tracker.events[0]
	if event.NormalizedQuery != "waterproof laptop bag" || event.RequestID != "req_1" || event.SessionID != "sess_1" {
		t.Fatalf("event = %#v", event)
	}
	if event.Filters[domain.SearchFilterBrand] != "Acme" || event.Filters[domain.SearchFilterMinPrice] != "100.00" {
		t.Fatalf("filters = %#v", event.Filters)
	}
}

func TestSearchProductsUsecaseDoesNotTrackPageBeyondResults(t *testing.T) {
	tracker := &capturingZeroResultTracker{}
	uc, searchRepo, hydrator := newSearchProductsTestUsecaseWithTracker(t, tracker)
	searchRepo.result = domain.ProductSearchResult{IDs: []string{}, Facets: map[string][]domain.FacetValue{}, Total: 12}

	resp, err := uc.Execute(context.Background(), domain.SearchRequest{Query: "shoes", Page: 99})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if resp.Total != 12 {
		t.Fatalf("total = %d", resp.Total)
	}
	if hydrator.called {
		t.Fatal("hydrator should not be called for empty page")
	}
	if len(tracker.events) != 0 {
		t.Fatalf("tracker events = %#v", tracker.events)
	}
}

func TestSearchProductsUsecaseMapsDependencyErrors(t *testing.T) {
	t.Run("typesense", func(t *testing.T) {
		uc, searchRepo, _ := newSearchProductsTestUsecase(t)
		searchRepo.err = errors.New("down")

		_, err := uc.Execute(context.Background(), domain.SearchRequest{})
		if !errors.Is(err, domain.ErrSearchBackendUnavailable) {
			t.Fatalf("expected search backend unavailable, got %v", err)
		}
	})

	t.Run("hydration", func(t *testing.T) {
		uc, searchRepo, hydrator := newSearchProductsTestUsecase(t)
		searchRepo.result = domain.ProductSearchResult{IDs: []string{"prod_1"}, Total: 1}
		hydrator.err = errors.New("down")

		_, err := uc.Execute(context.Background(), domain.SearchRequest{})
		if !errors.Is(err, domain.ErrProductHydrationUnavailable) {
			t.Fatalf("expected product hydration unavailable, got %v", err)
		}
	})
}

func newSearchProductsTestUsecase(t *testing.T) (*SearchProductsUsecase, *fakeProductSearchRepo, *fakeProductHydrator) {
	return newSearchProductsTestUsecaseWithTracker(t, NopZeroResultTracker{})
}

func newSearchProductsTestUsecaseWithTracker(t *testing.T, tracker ZeroResultSearchTracker) (*SearchProductsUsecase, *fakeProductSearchRepo, *fakeProductHydrator) {
	t.Helper()
	collection, policy, synonyms := schema.MustProductSchemaContract()
	schemaRepo, err := repository.NewStaticSchemaRepository(collection, policy, synonyms)
	if err != nil {
		t.Fatalf("schema repo: %v", err)
	}
	searchRepo := &fakeProductSearchRepo{}
	hydrator := &fakeProductHydrator{}
	uc, err := NewSearchProductsUsecase(schemaRepo, searchRepo, hydrator, SearchProductsOptions{
		TypesenseTimeout:  time.Second,
		HydrationTimeout:  time.Second,
		ZeroResultTracker: tracker,
	}, slog.Default())
	if err != nil {
		t.Fatalf("usecase: %v", err)
	}
	return uc, searchRepo, hydrator
}

type capturingZeroResultTracker struct {
	events []domain.ZeroResultSearchEvent
}

func (t *capturingZeroResultTracker) Track(_ context.Context, event domain.ZeroResultSearchEvent) {
	t.events = append(t.events, event.Normalize())
}

type fakeProductSearchRepo struct {
	query  domain.ProductSearchQuery
	result domain.ProductSearchResult
	err    error
}

func (r *fakeProductSearchRepo) SearchProducts(_ context.Context, query domain.ProductSearchQuery) (domain.ProductSearchResult, error) {
	r.query = query
	if r.err != nil {
		return domain.ProductSearchResult{}, r.err
	}
	return r.result, nil
}

type fakeProductHydrator struct {
	ids      []string
	products []domain.ProductSummary
	err      error
	called   bool
}

func (h *fakeProductHydrator) BatchGetProducts(_ context.Context, ids []string) ([]domain.ProductSummary, error) {
	h.called = true
	h.ids = append([]string(nil), ids...)
	if h.err != nil {
		return nil, h.err
	}
	return h.products, nil
}
