package repository

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/typesense/typesense-go/v2/typesense"
	"github.com/typesense/typesense-go/v2/typesense/api"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

type TypesenseProductRepository struct {
	client     *typesense.Client
	collection string
}

func NewTypesenseProductRepository(client *typesense.Client, collection string) (*TypesenseProductRepository, error) {
	if client == nil {
		return nil, errors.New("typesense client is required")
	}
	collection = strings.TrimSpace(collection)
	if collection == "" {
		return nil, errors.New("typesense products collection is required")
	}
	return &TypesenseProductRepository{client: client, collection: collection}, nil
}

func (r *TypesenseProductRepository) UpsertProduct(ctx context.Context, doc domain.ProductDocument) error {
	if err := doc.Validate(); err != nil {
		return err
	}
	if _, err := r.client.Collection(r.collection).Documents().Upsert(ctx, doc); err != nil {
		return fmt.Errorf("%w: upsert product %q: %v", domain.ErrProductIndexUnavailable, doc.ID, err)
	}
	return nil
}

func (r *TypesenseProductRepository) DeleteProduct(ctx context.Context, productID string) error {
	productID = strings.TrimSpace(productID)
	if productID == "" {
		return fmt.Errorf("%w: product_id is required", domain.ErrInvalidProductEvent)
	}
	if _, err := r.client.Collection(r.collection).Document(productID).Delete(ctx); err != nil {
		if isTypesenseStatus(err, http.StatusNotFound) {
			return nil
		}
		return fmt.Errorf("%w: delete product %q: %v", domain.ErrProductIndexUnavailable, productID, err)
	}
	return nil
}

func (r *TypesenseProductRepository) SearchProducts(ctx context.Context, query domain.ProductSearchQuery) (domain.ProductSearchResult, error) {
	if strings.TrimSpace(query.Query) == "" {
		return domain.ProductSearchResult{}, fmt.Errorf("%w: query is required", domain.ErrInvalidSearchRequest)
	}
	if strings.TrimSpace(query.QueryBy) == "" {
		return domain.ProductSearchResult{}, fmt.Errorf("%w: query_by is required", domain.ErrInvalidSearchRequest)
	}
	if query.Page <= 0 {
		return domain.ProductSearchResult{}, fmt.Errorf("%w: page must be greater than zero", domain.ErrInvalidSearchRequest)
	}
	if query.PageSize <= 0 {
		return domain.ProductSearchResult{}, fmt.Errorf("%w: page_size must be greater than zero", domain.ErrInvalidSearchRequest)
	}

	includeFields := domain.ProductFieldID
	maxFacetValues := 50
	params := &api.SearchCollectionParams{
		Q:              stringPtr(query.Query),
		QueryBy:        stringPtr(query.QueryBy),
		QueryByWeights: stringPtr(query.QueryByWeights),
		NumTypos:       stringPtr(query.NumTypos),
		Prefix:         stringPtr(query.Prefix),
		FacetBy:        stringPtr(query.FacetBy),
		SortBy:         stringPtr(query.SortBy),
		Page:           intPtr(query.Page),
		PerPage:        intPtr(query.PageSize),
		IncludeFields:  stringPtr(includeFields),
		MaxFacetValues: intPtr(maxFacetValues),
	}
	if strings.TrimSpace(query.FilterBy) != "" {
		params.FilterBy = stringPtr(query.FilterBy)
	}

	result, err := r.client.Collection(r.collection).Documents().Search(ctx, params)
	if err != nil {
		return domain.ProductSearchResult{}, fmt.Errorf("%w: search products: %v", domain.ErrSearchBackendUnavailable, err)
	}
	return mapProductSearchResult(result), nil
}

func isTypesenseStatus(err error, status int) bool {
	var httpErr *typesense.HTTPError
	return errors.As(err, &httpErr) && httpErr.Status == status
}

func mapProductSearchResult(result *api.SearchResult) domain.ProductSearchResult {
	if result == nil {
		return domain.ProductSearchResult{
			IDs:    []string{},
			Facets: map[string][]domain.FacetValue{},
		}
	}

	ids := []string{}
	if result.Hits != nil {
		ids = make([]string, 0, len(*result.Hits))
		for _, hit := range *result.Hits {
			if hit.Document == nil {
				continue
			}
			id, ok := (*hit.Document)[domain.ProductFieldID].(string)
			if !ok || strings.TrimSpace(id) == "" {
				continue
			}
			ids = append(ids, id)
		}
	}

	total := 0
	if result.Found != nil {
		total = *result.Found
	}

	searchTimeMS := 0
	if result.SearchTimeMs != nil {
		searchTimeMS = *result.SearchTimeMs
	}

	return domain.ProductSearchResult{
		IDs:          ids,
		Facets:       mapFacetCounts(result.FacetCounts),
		Total:        total,
		SearchTimeMS: searchTimeMS,
	}
}

func mapFacetCounts(counts *[]api.FacetCounts) map[string][]domain.FacetValue {
	if counts == nil {
		return map[string][]domain.FacetValue{}
	}

	facets := make(map[string][]domain.FacetValue, len(*counts))
	for _, facet := range *counts {
		if facet.FieldName == nil || strings.TrimSpace(*facet.FieldName) == "" {
			continue
		}
		values := []domain.FacetValue{}
		if facet.Counts != nil {
			values = make([]domain.FacetValue, 0, len(*facet.Counts))
			for _, count := range *facet.Counts {
				if count.Value == nil || count.Count == nil {
					continue
				}
				values = append(values, domain.FacetValue{
					Value: *count.Value,
					Count: *count.Count,
				})
			}
		}
		facets[*facet.FieldName] = values
	}
	return facets
}

func stringPtr(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func intPtr(value int) *int {
	return &value
}
