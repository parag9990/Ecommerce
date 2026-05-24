package schema

import (
	"fmt"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

func PopularQueriesCollectionSchema() domain.CollectionSchema {
	return domain.CollectionSchema{
		Name: domain.PopularQueriesCollectionName,
		Fields: []domain.SchemaField{
			{Name: domain.PopularQueryFieldID, Type: domain.TypesenseTypeString},
			{Name: domain.PopularQueryFieldQuery, Type: domain.TypesenseTypeString},
			{Name: domain.PopularQueryFieldNormalizedQuery, Type: domain.TypesenseTypeString},
			{Name: domain.PopularQueryFieldScore, Type: domain.TypesenseTypeInt32, Sort: true},
			{Name: domain.PopularQueryFieldCount, Type: domain.TypesenseTypeInt32, Sort: true},
			{Name: domain.PopularQueryFieldLastSeenAt, Type: domain.TypesenseTypeInt64, Sort: true},
			{Name: domain.PopularQueryFieldIsActive, Type: domain.TypesenseTypeBool, Facet: true},
			{Name: domain.PopularQueryFieldLocale, Type: domain.TypesenseTypeString, Facet: true, Optional: true},
		},
		DefaultSortingField: domain.DefaultPopularQuerySortingField,
	}
}

func MustPopularQueriesCollectionSchema() domain.CollectionSchema {
	collection := PopularQueriesCollectionSchema()
	if err := collection.Validate(); err != nil {
		panic(fmt.Sprintf("invalid popular queries collection schema: %v", err))
	}
	return collection
}
