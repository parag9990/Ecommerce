package schema

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

func ProductsCollectionSchema() domain.CollectionSchema {
	return domain.CollectionSchema{
		Name: domain.ProductsCollectionName,
		Fields: []domain.SchemaField{
			{Name: domain.ProductFieldID, Type: domain.TypesenseTypeString},
			{Name: domain.ProductFieldTitle, Type: domain.TypesenseTypeString},
			{Name: domain.ProductFieldDescription, Type: domain.TypesenseTypeString, Optional: true},
			{Name: domain.ProductFieldBrand, Type: domain.TypesenseTypeString, Facet: true, Optional: true},
			{Name: domain.ProductFieldCategoryIDs, Type: domain.TypesenseTypeStringArray, Facet: true},
			{Name: domain.ProductFieldSellerID, Type: domain.TypesenseTypeString, Facet: true},
			{Name: domain.ProductFieldPrice, Type: domain.TypesenseTypeFloat, Facet: true, Sort: true},
			{Name: domain.ProductFieldRating, Type: domain.TypesenseTypeFloat, Facet: true, Sort: true, Optional: true},
			{Name: domain.ProductFieldPopularityScore, Type: domain.TypesenseTypeInt32, Facet: true, Sort: true},
			{Name: domain.ProductFieldInStock, Type: domain.TypesenseTypeBool, Facet: true},
			{Name: domain.ProductFieldCreatedAt, Type: domain.TypesenseTypeInt64, Sort: true},
		},
		DefaultSortingField: domain.DefaultProductSortingField,
	}
}

func ProductSearchPolicy() domain.SearchPolicy {
	searchable := []domain.SearchableField{
		{Name: domain.ProductFieldTitle, Weight: 5, NumTypos: 2, Prefix: true},
		{Name: domain.ProductFieldBrand, Weight: 3, NumTypos: 1, Prefix: true},
		{Name: domain.ProductFieldDescription, Weight: 1, NumTypos: 1, Prefix: false},
	}
	facets := []string{
		domain.ProductFieldBrand,
		domain.ProductFieldCategoryIDs,
		domain.ProductFieldSellerID,
		domain.ProductFieldPrice,
		domain.ProductFieldRating,
		domain.ProductFieldInStock,
	}

	return domain.SearchPolicy{
		CollectionName: domain.ProductsCollectionName,
		Searchable:     searchable,
		FacetFields:    facets,
		SortOptions: []domain.SortOption{
			{Key: domain.SortRelevance, TypesenseBy: "_text_match:desc,popularity_score:desc", Description: "Best match with popularity tie-breaker"},
			{Key: domain.SortPopular, TypesenseBy: "popularity_score:desc", Description: "Most popular first"},
			{Key: domain.SortPriceAsc, TypesenseBy: "price:asc", Description: "Cheapest first"},
			{Key: domain.SortPriceDesc, TypesenseBy: "price:desc", Description: "Most expensive first"},
			{Key: domain.SortRatingDesc, TypesenseBy: "rating:desc,popularity_score:desc", Description: "Best rated first"},
			{Key: domain.SortNewest, TypesenseBy: "created_at:desc", Description: "Newest arrivals first"},
		},
		DefaultSort:      domain.SortRelevance,
		DefaultFacetBy:   strings.Join(facets, ","),
		DefaultQueryBy:   queryBy(searchable),
		QueryByWeights:   queryByWeights(searchable),
		NumTypos:         numTypos(searchable),
		Prefix:           prefixes(searchable),
		DefaultPage:      1,
		DefaultPageSize:  20,
		AllowedPageSizes: []int{10, 20, 50, 100},
		APIContract: domain.SearchAPIContract{
			RESTMethod:     "GET",
			RESTPath:       "/api/v1/search",
			GRPCMethod:     "SearchService.SearchProducts",
			RequestSchema:  "SearchRequest",
			ResponseSchema: "SearchResponse",
		},
	}
}

func ProductSynonymModel() domain.SynonymModel {
	return domain.SynonymModel{
		Shape: domain.Synonym{
			Root:     "mobile",
			Synonyms: []string{"phone", "smartphone", "cellphone"},
		},
		Examples: []domain.Synonym{
			{Root: "mobile", Synonyms: []string{"phone", "smartphone", "cellphone"}},
			{Root: "shoes", Synonyms: []string{"sneakers", "footwear", "trainers"}},
			{Root: "tv", Synonyms: []string{"television", "smart tv", "led tv"}},
			{Root: "laptop", Synonyms: []string{"notebook", "ultrabook"}},
		},
	}
}

func MustProductSchemaContract() (domain.CollectionSchema, domain.SearchPolicy, domain.SynonymModel) {
	collection := ProductsCollectionSchema()
	policy := ProductSearchPolicy()
	synonyms := ProductSynonymModel()
	if err := collection.Validate(); err != nil {
		panic(fmt.Sprintf("invalid product collection schema: %v", err))
	}
	if err := policy.Validate(collection); err != nil {
		panic(fmt.Sprintf("invalid product search policy: %v", err))
	}
	if err := synonyms.Validate(); err != nil {
		panic(fmt.Sprintf("invalid product synonym model: %v", err))
	}
	return collection, policy, synonyms
}

func queryBy(fields []domain.SearchableField) string {
	names := make([]string, 0, len(fields))
	for _, field := range fields {
		names = append(names, field.Name)
	}
	return strings.Join(names, ",")
}

func queryByWeights(fields []domain.SearchableField) string {
	values := make([]string, 0, len(fields))
	for _, field := range fields {
		values = append(values, strconv.Itoa(field.Weight))
	}
	return strings.Join(values, ",")
}

func numTypos(fields []domain.SearchableField) string {
	values := make([]string, 0, len(fields))
	for _, field := range fields {
		values = append(values, strconv.Itoa(field.NumTypos))
	}
	return strings.Join(values, ",")
}

func prefixes(fields []domain.SearchableField) string {
	values := make([]string, 0, len(fields))
	for _, field := range fields {
		values = append(values, strconv.FormatBool(field.Prefix))
	}
	return strings.Join(values, ",")
}
