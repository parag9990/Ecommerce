package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/typesense/typesense-go/v2/typesense"
	"github.com/typesense/typesense-go/v2/typesense/api"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

const (
	autocompleteProductQueryBy        = "title,brand"
	autocompleteProductQueryByWeights = "5,3"
	autocompleteProductPrefix         = "true,true"
	autocompleteProductFilterBy       = "in_stock:=true"
	autocompleteProductSortBy         = "_text_match:desc,popularity_score:desc"
	autocompleteProductIncludeFields  = "title,brand,popularity_score"
	autocompletePopularQueryBy        = "query"
	autocompletePopularFilterBy       = "is_active:=true"
	autocompletePopularSortBy         = "score:desc,last_seen_at:desc"
	autocompletePopularIncludeFields  = "query,score"
)

type TypesenseAutocompleteRepository struct {
	client                   *typesense.Client
	productsCollection       string
	popularQueriesCollection string
}

func NewTypesenseAutocompleteRepository(client *typesense.Client, productsCollection string, popularQueriesCollection string) (*TypesenseAutocompleteRepository, error) {
	if client == nil {
		return nil, errors.New("typesense client is required")
	}
	productsCollection = strings.TrimSpace(productsCollection)
	if productsCollection == "" {
		return nil, errors.New("typesense products collection is required")
	}
	popularQueriesCollection = strings.TrimSpace(popularQueriesCollection)
	if popularQueriesCollection == "" {
		return nil, errors.New("typesense popular queries collection is required")
	}
	return &TypesenseAutocompleteRepository{
		client:                   client,
		productsCollection:       productsCollection,
		popularQueriesCollection: popularQueriesCollection,
	}, nil
}

func (r *TypesenseAutocompleteRepository) ProductPrefixCandidates(ctx context.Context, query string, limit int) ([]domain.SuggestionCandidate, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []domain.SuggestionCandidate{}, nil
	}
	if limit <= 0 {
		return []domain.SuggestionCandidate{}, nil
	}

	perPage := limit * 2
	if perPage < limit {
		perPage = limit
	}
	params := &api.SearchCollectionParams{
		Q:              stringPtr(query),
		QueryBy:        stringPtr(autocompleteProductQueryBy),
		QueryByWeights: stringPtr(autocompleteProductQueryByWeights),
		Prefix:         stringPtr(autocompleteProductPrefix),
		FilterBy:       stringPtr(autocompleteProductFilterBy),
		SortBy:         stringPtr(autocompleteProductSortBy),
		Page:           intPtr(1),
		PerPage:        intPtr(perPage),
		IncludeFields:  stringPtr(autocompleteProductIncludeFields),
	}

	result, err := r.client.Collection(r.productsCollection).Documents().Search(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("%w: autocomplete product prefix: %v", domain.ErrSearchBackendUnavailable, err)
	}
	return mapProductAutocompleteCandidates(result), nil
}

func (r *TypesenseAutocompleteRepository) PopularQueryCandidates(ctx context.Context, query string, limit int) ([]domain.SuggestionCandidate, error) {
	if limit <= 0 {
		return []domain.SuggestionCandidate{}, nil
	}

	searchQuery := strings.TrimSpace(query)
	prefix := "true"
	if searchQuery == "" {
		searchQuery = "*"
		prefix = "false"
	}

	params := &api.SearchCollectionParams{
		Q:             stringPtr(searchQuery),
		QueryBy:       stringPtr(autocompletePopularQueryBy),
		Prefix:        stringPtr(prefix),
		FilterBy:      stringPtr(autocompletePopularFilterBy),
		SortBy:        stringPtr(autocompletePopularSortBy),
		Page:          intPtr(1),
		PerPage:       intPtr(limit),
		IncludeFields: stringPtr(autocompletePopularIncludeFields),
	}

	result, err := r.client.Collection(r.popularQueriesCollection).Documents().Search(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("%w: autocomplete popular queries: %v", domain.ErrSearchBackendUnavailable, err)
	}
	return mapPopularQueryCandidates(result), nil
}

func mapProductAutocompleteCandidates(result *api.SearchResult) []domain.SuggestionCandidate {
	if result == nil || result.Hits == nil {
		return []domain.SuggestionCandidate{}
	}

	candidates := make([]domain.SuggestionCandidate, 0, len(*result.Hits)*2)
	for _, hit := range *result.Hits {
		if hit.Document == nil {
			continue
		}
		doc := *hit.Document
		score := documentInt(doc, domain.ProductFieldPopularityScore)
		if title := documentString(doc, domain.ProductFieldTitle); title != "" {
			candidates = append(candidates, domain.SuggestionCandidate{
				Text:   title,
				Source: domain.AutocompleteSourceProductTitle,
				Score:  score,
			})
		}
		if brand := documentString(doc, domain.ProductFieldBrand); brand != "" {
			candidates = append(candidates, domain.SuggestionCandidate{
				Text:   brand,
				Source: domain.AutocompleteSourceProductBrand,
				Score:  score,
			})
		}
	}
	return candidates
}

func mapPopularQueryCandidates(result *api.SearchResult) []domain.SuggestionCandidate {
	if result == nil || result.Hits == nil {
		return []domain.SuggestionCandidate{}
	}

	candidates := make([]domain.SuggestionCandidate, 0, len(*result.Hits))
	for _, hit := range *result.Hits {
		if hit.Document == nil {
			continue
		}
		doc := *hit.Document
		query := documentString(doc, domain.PopularQueryFieldQuery)
		if query == "" {
			continue
		}
		candidates = append(candidates, domain.SuggestionCandidate{
			Text:   query,
			Source: domain.AutocompleteSourcePopularQuery,
			Score:  documentInt(doc, domain.PopularQueryFieldScore),
		})
	}
	return candidates
}

func documentString(doc map[string]any, field string) string {
	value, ok := doc[field]
	if !ok || value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case json.Number:
		return strings.TrimSpace(v.String())
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func documentInt(doc map[string]any, field string) int {
	value, ok := doc[field]
	if !ok || value == nil {
		return 0
	}
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	case json.Number:
		i, err := v.Int64()
		if err == nil {
			return int(i)
		}
		f, err := v.Float64()
		if err == nil {
			return int(f)
		}
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(v))
		if err == nil {
			return i
		}
	}
	return 0
}
