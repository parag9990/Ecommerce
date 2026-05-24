package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

type ProductSchemaContract struct {
	Version             string                  `json:"version"`
	Collection          domain.CollectionSchema `json:"collection"`
	TypesenseDefinition map[string]any          `json:"typesense_definition"`
	SearchPolicy        domain.SearchPolicy     `json:"search_policy"`
	SynonymModel        domain.SynonymModel     `json:"synonym_model"`
}

type SchemaUsecase struct {
	repo   SchemaRepository
	logger *slog.Logger
}

func NewSchemaUsecase(repo SchemaRepository, logger *slog.Logger) (*SchemaUsecase, error) {
	if repo == nil {
		return nil, errors.New("schema repository is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &SchemaUsecase{repo: repo, logger: logger}, nil
}

func (u *SchemaUsecase) GetProductSchemaContract(ctx context.Context) (ProductSchemaContract, error) {
	collection, err := u.repo.GetProductsCollectionSchema(ctx)
	if err != nil {
		u.logger.Error("search.schema.collection_load_failed", slog.String("error", err.Error()))
		return ProductSchemaContract{}, errors.Join(domain.ErrSchemaUnavailable, err)
	}
	if err := collection.Validate(); err != nil {
		u.logger.Error("search.schema.collection_invalid", slog.String("error", err.Error()))
		return ProductSchemaContract{}, err
	}

	policy, err := u.repo.GetProductSearchPolicy(ctx)
	if err != nil {
		u.logger.Error("search.schema.policy_load_failed", slog.String("error", err.Error()))
		return ProductSchemaContract{}, errors.Join(domain.ErrSchemaUnavailable, err)
	}
	if err := policy.Validate(collection); err != nil {
		u.logger.Error("search.schema.policy_invalid", slog.String("error", err.Error()))
		return ProductSchemaContract{}, err
	}

	synonyms, err := u.repo.GetProductSynonymModel(ctx)
	if err != nil {
		u.logger.Error("search.schema.synonyms_load_failed", slog.String("error", err.Error()))
		return ProductSchemaContract{}, errors.Join(domain.ErrSchemaUnavailable, err)
	}
	if err := synonyms.Validate(); err != nil {
		u.logger.Error("search.schema.synonyms_invalid", slog.String("error", err.Error()))
		return ProductSchemaContract{}, err
	}

	return ProductSchemaContract{
		Version:             domain.ProductSchemaContractVersion,
		Collection:          collection,
		TypesenseDefinition: collection.TypesenseDefinition(),
		SearchPolicy:        policy,
		SynonymModel:        synonyms,
	}, nil
}
