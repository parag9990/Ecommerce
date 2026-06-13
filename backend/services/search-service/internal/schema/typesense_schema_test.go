package schema

import (
	"reflect"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

func TestProductsCollectionSchemaMatchesTaskContract(t *testing.T) {
	collection := ProductsCollectionSchema()
	if err := collection.Validate(); err != nil {
		t.Fatalf("schema should validate: %v", err)
	}
	if collection.Name != domain.ProductsCollectionName {
		t.Fatalf("collection name = %q, want %q", collection.Name, domain.ProductsCollectionName)
	}
	if collection.DefaultSortingField != domain.ProductFieldPopularityScore {
		t.Fatalf("default sorting field = %q", collection.DefaultSortingField)
	}

	wantFields := []domain.SchemaField{
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
	}
	if !reflect.DeepEqual(collection.Fields, wantFields) {
		t.Fatalf("fields mismatch\n got: %#v\nwant: %#v", collection.Fields, wantFields)
	}
}

func TestProductSearchPolicyMatchesTaskContract(t *testing.T) {
	collection := ProductsCollectionSchema()
	policy := ProductSearchPolicy()
	if err := policy.Validate(collection); err != nil {
		t.Fatalf("policy should validate: %v", err)
	}

	if policy.DefaultQueryBy != "title,brand,description" {
		t.Fatalf("DefaultQueryBy = %q", policy.DefaultQueryBy)
	}
	if policy.QueryByWeights != "5,3,1" {
		t.Fatalf("QueryByWeights = %q", policy.QueryByWeights)
	}
	if policy.NumTypos != "2,1,1" {
		t.Fatalf("NumTypos = %q", policy.NumTypos)
	}
	if policy.Prefix != "true,true,false" {
		t.Fatalf("Prefix = %q", policy.Prefix)
	}
	if policy.DefaultFacetBy != "brand,category_ids,seller_id,price,rating,in_stock" {
		t.Fatalf("DefaultFacetBy = %q", policy.DefaultFacetBy)
	}

	sortMap := map[string]string{}
	for _, option := range policy.SortOptions {
		sortMap[option.Key] = option.TypesenseBy
	}
	wantSorts := map[string]string{
		domain.SortRelevance:  "_text_match:desc,popularity_score:desc",
		domain.SortPopular:    "popularity_score:desc",
		domain.SortPriceAsc:   "price:asc",
		domain.SortPriceDesc:  "price:desc",
		domain.SortRatingDesc: "rating:desc,popularity_score:desc",
		domain.SortNewest:     "created_at:desc",
	}
	if !reflect.DeepEqual(sortMap, wantSorts) {
		t.Fatalf("sort map mismatch\n got: %#v\nwant: %#v", sortMap, wantSorts)
	}
}

func TestProductSynonymModelMatchesTaskContract(t *testing.T) {
	model := ProductSynonymModel()
	if err := model.Validate(); err != nil {
		t.Fatalf("synonym model should validate: %v", err)
	}

	if model.Shape.Root != "mobile" {
		t.Fatalf("shape root = %q", model.Shape.Root)
	}
	if !reflect.DeepEqual(model.Shape.Synonyms, []string{"phone", "smartphone", "cellphone"}) {
		t.Fatalf("shape synonyms = %#v", model.Shape.Synonyms)
	}
}

func TestPopularQueriesCollectionSchemaMatchesTaskContract(t *testing.T) {
	collection := PopularQueriesCollectionSchema()
	if err := collection.Validate(); err != nil {
		t.Fatalf("schema should validate: %v", err)
	}
	if collection.Name != domain.PopularQueriesCollectionName {
		t.Fatalf("collection name = %q", collection.Name)
	}
	if collection.DefaultSortingField != domain.PopularQueryFieldScore {
		t.Fatalf("default sorting field = %q", collection.DefaultSortingField)
	}

	wantFields := []domain.SchemaField{
		{Name: domain.PopularQueryFieldID, Type: domain.TypesenseTypeString},
		{Name: domain.PopularQueryFieldQuery, Type: domain.TypesenseTypeString},
		{Name: domain.PopularQueryFieldNormalizedQuery, Type: domain.TypesenseTypeString},
		{Name: domain.PopularQueryFieldScore, Type: domain.TypesenseTypeInt32, Sort: true},
		{Name: domain.PopularQueryFieldCount, Type: domain.TypesenseTypeInt32, Sort: true},
		{Name: domain.PopularQueryFieldLastSeenAt, Type: domain.TypesenseTypeInt64, Sort: true},
		{Name: domain.PopularQueryFieldIsActive, Type: domain.TypesenseTypeBool, Facet: true},
		{Name: domain.PopularQueryFieldLocale, Type: domain.TypesenseTypeString, Facet: true, Optional: true},
	}
	if !reflect.DeepEqual(collection.Fields, wantFields) {
		t.Fatalf("fields mismatch\n got: %#v\nwant: %#v", collection.Fields, wantFields)
	}
}
