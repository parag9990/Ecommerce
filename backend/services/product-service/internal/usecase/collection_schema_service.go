package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"product-service/internal/repository"
)

type CollectionSchemaUseCase interface {
	EnsureProductCollections(ctx context.Context) (CollectionSetupResult, error)
	DescribeProductCollections(ctx context.Context) ([]CollectionDescription, error)
}

type CollectionSetupResult struct {
	Database    string
	Collections []CollectionSetupItem
}

type CollectionSetupItem struct {
	Name             string
	Created          bool
	ValidatorApplied bool
	IndexNames       []string
}

type CollectionDescription struct {
	Name         string
	IndexNames   []string
	HasValidator bool
}

type CollectionSchemaService struct {
	manager repository.CollectionSchemaManager
	logger  *slog.Logger
}

func NewCollectionSchemaService(manager repository.CollectionSchemaManager, logger *slog.Logger) *CollectionSchemaService {
	if logger == nil {
		logger = slog.Default()
	}
	return &CollectionSchemaService{
		manager: manager,
		logger:  logger,
	}
}

func (s *CollectionSchemaService) EnsureProductCollections(ctx context.Context) (CollectionSetupResult, error) {
	if err := ctx.Err(); err != nil {
		return CollectionSetupResult{}, err
	}
	if s.manager == nil {
		return CollectionSetupResult{}, fmt.Errorf("collection schema manager is required")
	}
	result, err := s.manager.EnsureCollections(ctx)
	if err != nil {
		s.logger.Error("ensure product mongo collections failed", "error", err)
		return CollectionSetupResult{}, err
	}
	mapped := collectionSetupResultFromRepository(result)
	s.logger.Info(
		"product mongo collections ensured",
		"database", mapped.Database,
		"collection_count", len(mapped.Collections),
	)
	return mapped, nil
}

func (s *CollectionSchemaService) DescribeProductCollections(ctx context.Context) ([]CollectionDescription, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s.manager == nil {
		return nil, fmt.Errorf("collection schema manager is required")
	}
	descriptions := s.manager.DescribeCollections()
	mapped := make([]CollectionDescription, 0, len(descriptions))
	for _, description := range descriptions {
		mapped = append(mapped, CollectionDescription{
			Name:         description.Name,
			IndexNames:   append([]string(nil), description.IndexNames...),
			HasValidator: description.HasValidator,
		})
	}
	return mapped, nil
}

func collectionSetupResultFromRepository(result repository.CollectionSetupResult) CollectionSetupResult {
	mapped := CollectionSetupResult{
		Database:    result.Database,
		Collections: make([]CollectionSetupItem, 0, len(result.Collections)),
	}
	for _, item := range result.Collections {
		mapped.Collections = append(mapped.Collections, CollectionSetupItem{
			Name:             item.Name,
			Created:          item.Created,
			ValidatorApplied: item.ValidatorApplied,
			IndexNames:       append([]string(nil), item.IndexNames...),
		})
	}
	return mapped
}
