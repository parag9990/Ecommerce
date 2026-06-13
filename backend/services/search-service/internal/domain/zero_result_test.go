package domain

import (
	"strings"
	"testing"
)

func TestZeroResultSearchEventShouldTrack(t *testing.T) {
	event := ZeroResultSearchEvent{
		Query:       " Waterproof   Laptop Bag ",
		AnonymousID: "anon_1",
		SessionID:   "sess_1",
	}
	if !event.ShouldTrack() {
		t.Fatalf("expected event to be trackable, reason=%s", event.SkipReason())
	}
	normalized := event.Normalize()
	if normalized.NormalizedQuery != "waterproof laptop bag" {
		t.Fatalf("normalized query = %q", normalized.NormalizedQuery)
	}
}

func TestZeroResultSearchEventSkipsBrowseAndMissingSession(t *testing.T) {
	tests := map[string]ZeroResultSearchEvent{
		"browse": {
			Query:       DefaultSearchQuery,
			AnonymousID: "anon_1",
			SessionID:   "sess_1",
		},
		"missing anonymous id": {
			Query:     "laptop bag",
			SessionID: "sess_1",
		},
		"missing session id": {
			Query:       "laptop bag",
			AnonymousID: "anon_1",
		},
	}
	for name, event := range tests {
		t.Run(name, func(t *testing.T) {
			if event.ShouldTrack() {
				t.Fatal("expected event to be skipped")
			}
		})
	}
}

func TestZeroResultDedupeKeyIncludesSessionQueryAndFilters(t *testing.T) {
	base := ZeroResultSearchEvent{
		Query:       "Laptop Bag",
		Filters:     map[string]string{SearchFilterBrand: "Acme"},
		AnonymousID: "anon_1",
		SessionID:   "sess_1",
	}
	same := ZeroResultSearchEvent{
		Query:       " laptop   bag ",
		Filters:     map[string]string{SearchFilterBrand: "Acme"},
		AnonymousID: "anon_1",
		SessionID:   "sess_1",
	}
	differentFilter := base
	differentFilter.Filters = map[string]string{SearchFilterBrand: "Other"}

	if base.DedupeKey() != same.DedupeKey() {
		t.Fatalf("expected equivalent query to produce same key")
	}
	if base.DedupeKey() == differentFilter.DedupeKey() {
		t.Fatalf("expected filters to affect dedupe key")
	}
	if !strings.HasPrefix(base.DedupeKey(), zeroResultDedupePrefix) {
		t.Fatalf("dedupe key prefix = %q", base.DedupeKey())
	}
}

func TestAnalyticsFiltersAllowlist(t *testing.T) {
	minPrice := 100.0
	maxPrice := 250.5
	rating := 4.2
	inStock := true

	filters := AnalyticsFilters(SearchInput{Filters: SearchFilters{
		Brands:     []string{"Acme", "Contoso"},
		CategoryID: "cat_bags",
		SellerID:   "seller_1",
		MinPrice:   &minPrice,
		MaxPrice:   &maxPrice,
		MinRating:  &rating,
		InStock:    &inStock,
	}})

	if filters[SearchFilterBrand] != "Acme,Contoso" {
		t.Fatalf("brand filter = %q", filters[SearchFilterBrand])
	}
	if filters[SearchFilterMinPrice] != "100.00" || filters[SearchFilterMaxPrice] != "250.50" {
		t.Fatalf("price filters = %#v", filters)
	}
	if filters[SearchFilterMinRating] != "4.2" || filters[SearchFilterInStock] != "true" {
		t.Fatalf("rating/in_stock filters = %#v", filters)
	}
}
