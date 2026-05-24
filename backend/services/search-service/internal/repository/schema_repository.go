package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

type SchemaRepository interface {
	GetProductsCollectionSchema(ctx context.Context) (domain.CollectionSchema, error)
	GetProductSearchPolicy(ctx context.Context) (domain.SearchPolicy, error)
	GetProductSynonymModel(ctx context.Context) (domain.SynonymModel, error)
}

type StaticSchemaRepository struct {
	collection domain.CollectionSchema
	policy     domain.SearchPolicy
	synonyms   domain.SynonymModel
}

func NewStaticSchemaRepository(collection domain.CollectionSchema, policy domain.SearchPolicy, synonyms domain.SynonymModel) (*StaticSchemaRepository, error) {
	if err := collection.Validate(); err != nil {
		return nil, fmt.Errorf("validate collection schema: %w", err)
	}
	if err := policy.Validate(collection); err != nil {
		return nil, fmt.Errorf("validate search policy: %w", err)
	}
	if err := synonyms.Validate(); err != nil {
		return nil, fmt.Errorf("validate synonym model: %w", err)
	}
	return &StaticSchemaRepository{
		collection: collection,
		policy:     policy,
		synonyms:   synonyms,
	}, nil
}

func (r *StaticSchemaRepository) GetProductsCollectionSchema(ctx context.Context) (domain.CollectionSchema, error) {
	if err := contextError(ctx); err != nil {
		return domain.CollectionSchema{}, err
	}
	return cloneCollectionSchema(r.collection), nil
}

func (r *StaticSchemaRepository) GetProductSearchPolicy(ctx context.Context) (domain.SearchPolicy, error) {
	if err := contextError(ctx); err != nil {
		return domain.SearchPolicy{}, err
	}
	return cloneSearchPolicy(r.policy), nil
}

func (r *StaticSchemaRepository) GetProductSynonymModel(ctx context.Context) (domain.SynonymModel, error) {
	if err := contextError(ctx); err != nil {
		return domain.SynonymModel{}, err
	}
	return cloneSynonymModel(r.synonyms), nil
}

func contextError(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return errors.Join(domain.ErrSchemaUnavailable, err)
	}
	return nil
}

func cloneCollectionSchema(schema domain.CollectionSchema) domain.CollectionSchema {
	schema.Fields = append([]domain.SchemaField(nil), schema.Fields...)
	return schema
}

func cloneSearchPolicy(policy domain.SearchPolicy) domain.SearchPolicy {
	policy.Searchable = append([]domain.SearchableField(nil), policy.Searchable...)
	policy.FacetFields = append([]string(nil), policy.FacetFields...)
	policy.SortOptions = append([]domain.SortOption(nil), policy.SortOptions...)
	policy.AllowedPageSizes = append([]int(nil), policy.AllowedPageSizes...)
	return policy
}

func cloneSynonymModel(model domain.SynonymModel) domain.SynonymModel {
	model.Shape.Synonyms = append([]string(nil), model.Shape.Synonyms...)
	model.Examples = append([]domain.Synonym(nil), model.Examples...)
	for i := range model.Examples {
		model.Examples[i].Synonyms = append([]string(nil), model.Examples[i].Synonyms...)
	}
	return model
}
