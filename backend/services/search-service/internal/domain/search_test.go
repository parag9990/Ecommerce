package domain

import (
	"errors"
	"testing"
)

func TestNormalizeSearchRequestDefaultsBrowseMode(t *testing.T) {
	policy := testSearchPolicy()

	input, err := NormalizeSearchRequest(SearchRequest{}, policy, SearchLimits{})
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if input.Query != DefaultSearchQuery {
		t.Fatalf("query = %q", input.Query)
	}
	if input.Page != 1 {
		t.Fatalf("page = %d", input.Page)
	}
	if input.PageSize != 20 {
		t.Fatalf("page size = %d", input.PageSize)
	}
	if input.Sort != SortRelevance {
		t.Fatalf("sort = %q", input.Sort)
	}
}

func TestNormalizeSearchRequestRejectsPageSizeTooLarge(t *testing.T) {
	policy := testSearchPolicy()

	_, err := NormalizeSearchRequest(SearchRequest{PageSize: 101}, policy, SearchLimits{MaxPageSize: 100})
	if !errors.Is(err, ErrInvalidSearchRequest) {
		t.Fatalf("expected invalid search request, got %v", err)
	}
}

func TestNormalizeSearchRequestRejectsUnknownSort(t *testing.T) {
	policy := testSearchPolicy()

	_, err := NormalizeSearchRequest(SearchRequest{Sort: "price_down"}, policy, SearchLimits{})
	if !errors.Is(err, ErrUnsupportedSearchSort) {
		t.Fatalf("expected unsupported sort, got %v", err)
	}
}

func testSearchPolicy() SearchPolicy {
	return SearchPolicy{
		DefaultSort:     SortRelevance,
		DefaultPage:     1,
		DefaultPageSize: 20,
		SortOptions: []SortOption{
			{Key: SortRelevance, TypesenseBy: "_text_match:desc,popularity_score:desc"},
			{Key: SortPriceAsc, TypesenseBy: "price:asc"},
		},
	}
}

func TestBuildProductFilterBy(t *testing.T) {
	minPrice := 1000.0
	maxPrice := 5000.0
	minRating := 4.0
	inStock := true

	filterBy, err := BuildProductFilterBy(SearchFilters{
		Brands:     []string{"Nike", "Puma"},
		CategoryID: "cat_shoes",
		SellerID:   "seller_1",
		MinPrice:   &minPrice,
		MaxPrice:   &maxPrice,
		MinRating:  &minRating,
		InStock:    &inStock,
	})
	if err != nil {
		t.Fatalf("filter by: %v", err)
	}

	want := "brand:=[`Nike`,`Puma`] && category_ids:=`cat_shoes` && seller_id:=`seller_1` && price:>=1000 && price:<=5000 && rating:>=4 && in_stock:=true"
	if filterBy != want {
		t.Fatalf("filter_by = %q, want %q", filterBy, want)
	}
}

func TestParseSearchFiltersRejectsInvalidValues(t *testing.T) {
	_, err := ParseSearchFilters(map[string]string{
		SearchFilterMinPrice:  "500",
		SearchFilterMaxPrice:  "100",
		SearchFilterMinRating: "6",
	})
	if !errors.Is(err, ErrInvalidSearchFilter) {
		t.Fatalf("expected invalid filter, got %v", err)
	}
}
