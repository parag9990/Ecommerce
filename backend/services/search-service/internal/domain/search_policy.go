package domain

import (
	"fmt"
	"slices"
	"strings"
)

const (
	SortRelevance  = "relevance"
	SortPriceAsc   = "price_asc"
	SortPriceDesc  = "price_desc"
	SortRatingDesc = "rating_desc"
	SortNewest     = "newest"
	SortPopular    = "popular"
)

type SearchPolicy struct {
	CollectionName   string            `json:"collection_name"`
	Searchable       []SearchableField `json:"searchable_fields"`
	FacetFields      []string          `json:"facet_fields"`
	SortOptions      []SortOption      `json:"sort_options"`
	DefaultSort      string            `json:"default_sort"`
	DefaultFacetBy   string            `json:"default_facet_by"`
	DefaultQueryBy   string            `json:"default_query_by"`
	QueryByWeights   string            `json:"query_by_weights"`
	NumTypos         string            `json:"num_typos"`
	Prefix           string            `json:"prefix"`
	DefaultPage      int               `json:"default_page"`
	DefaultPageSize  int               `json:"default_page_size"`
	AllowedPageSizes []int             `json:"allowed_page_sizes"`
	APIContract      SearchAPIContract `json:"api_contract"`
}

type SearchableField struct {
	Name     string `json:"name"`
	Weight   int    `json:"weight"`
	NumTypos int    `json:"num_typos"`
	Prefix   bool   `json:"prefix"`
}

type SortOption struct {
	Key         string `json:"key"`
	TypesenseBy string `json:"typesense_by"`
	Description string `json:"description"`
}

type SearchAPIContract struct {
	RESTMethod     string `json:"rest_method"`
	RESTPath       string `json:"rest_path"`
	GRPCMethod     string `json:"grpc_method"`
	RequestSchema  string `json:"request_schema"`
	ResponseSchema string `json:"response_schema"`
}

func (p SearchPolicy) Validate(schema CollectionSchema) error {
	if strings.TrimSpace(p.CollectionName) == "" {
		return fmt.Errorf("%w: collection name is required", ErrInvalidSearchPolicy)
	}
	if p.CollectionName != schema.Name {
		return fmt.Errorf("%w: collection name %q does not match schema %q", ErrInvalidSearchPolicy, p.CollectionName, schema.Name)
	}
	if len(p.Searchable) == 0 {
		return fmt.Errorf("%w: searchable fields are required", ErrInvalidSearchPolicy)
	}
	if len(p.FacetFields) == 0 {
		return fmt.Errorf("%w: facet fields are required", ErrInvalidSearchPolicy)
	}
	if len(p.SortOptions) == 0 {
		return fmt.Errorf("%w: sort options are required", ErrInvalidSearchPolicy)
	}
	if strings.TrimSpace(p.DefaultSort) == "" {
		return fmt.Errorf("%w: default sort is required", ErrInvalidSearchPolicy)
	}
	if p.DefaultPage <= 0 {
		return fmt.Errorf("%w: default page must be greater than zero", ErrInvalidSearchPolicy)
	}
	if p.DefaultPageSize <= 0 {
		return fmt.Errorf("%w: default page size must be greater than zero", ErrInvalidSearchPolicy)
	}

	fields := make(map[string]SchemaField, len(schema.Fields))
	for _, field := range schema.Fields {
		fields[field.Name] = field
	}
	for _, searchable := range p.Searchable {
		field, ok := fields[searchable.Name]
		if !ok {
			return fmt.Errorf("%w: searchable field %q is not in schema", ErrInvalidSearchPolicy, searchable.Name)
		}
		if field.Type != TypesenseTypeString {
			return fmt.Errorf("%w: searchable field %q must be string", ErrInvalidSearchPolicy, searchable.Name)
		}
		if searchable.Weight <= 0 {
			return fmt.Errorf("%w: searchable field %q weight must be greater than zero", ErrInvalidSearchPolicy, searchable.Name)
		}
		if searchable.NumTypos < 0 || searchable.NumTypos > 2 {
			return fmt.Errorf("%w: searchable field %q num_typos must be between 0 and 2", ErrInvalidSearchPolicy, searchable.Name)
		}
	}
	for _, facet := range p.FacetFields {
		field, ok := fields[facet]
		if !ok {
			return fmt.Errorf("%w: facet field %q is not in schema", ErrInvalidSearchPolicy, facet)
		}
		if !field.Facet {
			return fmt.Errorf("%w: facet field %q is not marked as facet", ErrInvalidSearchPolicy, facet)
		}
	}

	sortKeys := make(map[string]struct{}, len(p.SortOptions))
	for _, option := range p.SortOptions {
		if strings.TrimSpace(option.Key) == "" || strings.TrimSpace(option.TypesenseBy) == "" {
			return fmt.Errorf("%w: sort options require key and typesense mapping", ErrInvalidSearchPolicy)
		}
		if _, ok := sortKeys[option.Key]; ok {
			return fmt.Errorf("%w: duplicate sort key %q", ErrInvalidSearchPolicy, option.Key)
		}
		sortKeys[option.Key] = struct{}{}
	}
	if _, ok := sortKeys[p.DefaultSort]; !ok {
		return fmt.Errorf("%w: default sort %q is not a supported sort option", ErrInvalidSearchPolicy, p.DefaultSort)
	}
	if !slices.Contains(p.AllowedPageSizes, p.DefaultPageSize) {
		return fmt.Errorf("%w: default page size must be allowed", ErrInvalidSearchPolicy)
	}
	if err := p.APIContract.Validate(); err != nil {
		return err
	}
	return nil
}

func (c SearchAPIContract) Validate() error {
	if strings.TrimSpace(c.RESTMethod) == "" {
		return fmt.Errorf("%w: rest method is required", ErrInvalidSearchPolicy)
	}
	if strings.TrimSpace(c.RESTPath) == "" {
		return fmt.Errorf("%w: rest path is required", ErrInvalidSearchPolicy)
	}
	if strings.TrimSpace(c.GRPCMethod) == "" {
		return fmt.Errorf("%w: grpc method is required", ErrInvalidSearchPolicy)
	}
	if strings.TrimSpace(c.RequestSchema) == "" || strings.TrimSpace(c.ResponseSchema) == "" {
		return fmt.Errorf("%w: request and response schemas are required", ErrInvalidSearchPolicy)
	}
	return nil
}
