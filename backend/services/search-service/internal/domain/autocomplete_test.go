package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeAutocompleteRequest(t *testing.T) {
	input, err := NormalizeAutocompleteRequest(AutocompleteRequest{Query: "  Nike   Shoes  "}, 8, 10)
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if input.Query != "Nike Shoes" {
		t.Fatalf("query = %q", input.Query)
	}
	if input.NormalizedQuery != "nike shoes" {
		t.Fatalf("normalized query = %q", input.NormalizedQuery)
	}
	if input.Limit != 8 {
		t.Fatalf("limit = %d", input.Limit)
	}
}

func TestNormalizeAutocompleteRequestAllowsEmptyQuery(t *testing.T) {
	input, err := NormalizeAutocompleteRequest(AutocompleteRequest{}, 8, 10)
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if input.NormalizedQuery != "" || input.Limit != 8 {
		t.Fatalf("input = %#v", input)
	}
}

func TestNormalizeAutocompleteRequestRejectsInvalidValues(t *testing.T) {
	t.Run("query too long", func(t *testing.T) {
		_, err := NormalizeAutocompleteRequest(AutocompleteRequest{Query: strings.Repeat("a", MaxAutocompleteQueryLength+1)}, 8, 10)
		if !errors.Is(err, ErrInvalidAutocompleteRequest) {
			t.Fatalf("expected invalid autocomplete request, got %v", err)
		}
	})

	t.Run("limit too large", func(t *testing.T) {
		_, err := NormalizeAutocompleteRequest(AutocompleteRequest{Limit: 11}, 8, 10)
		if !errors.Is(err, ErrInvalidAutocompleteRequest) {
			t.Fatalf("expected invalid autocomplete request, got %v", err)
		}
	})
}

func TestMergeAutocompleteCandidatesDedupesCaseInsensitive(t *testing.T) {
	got := MergeAutocompleteCandidates("nike", 5,
		[]SuggestionCandidate{{Text: "Nike Shoes", Source: AutocompleteSourceProductTitle, Score: 20}},
		[]SuggestionCandidate{{Text: "nike shoes", Source: AutocompleteSourcePopularQuery, Score: 50}},
	)
	if len(got) != 1 || got[0] != "nike shoes" {
		t.Fatalf("suggestions = %#v", got)
	}
}

func TestMergeAutocompleteCandidatesRanksAndLimits(t *testing.T) {
	got := MergeAutocompleteCandidates("sho", 2,
		[]SuggestionCandidate{
			{Text: "Shoe Rack Organizer", Source: AutocompleteSourceProductTitle, Score: 20},
			{Text: "Nike", Source: AutocompleteSourceProductBrand, Score: 100},
		},
		[]SuggestionCandidate{
			{Text: "shoes", Source: AutocompleteSourcePopularQuery, Score: 80},
			{Text: "shoe rack", Source: AutocompleteSourcePopularQuery, Score: 70},
		},
	)
	want := []string{"shoes", "shoe rack"}
	if len(got) != len(want) {
		t.Fatalf("suggestions = %#v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("suggestions = %#v, want %#v", got, want)
		}
	}
}

func TestBuildAutocompleteCacheKeyNormalizesQuery(t *testing.T) {
	key := BuildAutocompleteCacheKey(" Nike   Shoes ", 8)
	if key != "autocomplete:v1:nike shoes:8" {
		t.Fatalf("key = %q", key)
	}
}
