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

type TypesenseCollectionRepository struct {
	client *typesense.Client
}

func NewTypesenseCollectionRepository(client *typesense.Client) (*TypesenseCollectionRepository, error) {
	if client == nil {
		return nil, errors.New("typesense client is required")
	}
	return &TypesenseCollectionRepository{client: client}, nil
}

func (r *TypesenseCollectionRepository) EnsureCollection(ctx context.Context, collection domain.CollectionSchema) error {
	if err := collection.Validate(); err != nil {
		return err
	}
	if _, err := r.client.Collection(collection.Name).Retrieve(ctx); err == nil {
		return nil
	} else if !isTypesenseStatus(err, http.StatusNotFound) {
		return fmt.Errorf("%w: retrieve collection %q: %v", domain.ErrSchemaUnavailable, collection.Name, err)
	}

	if _, err := r.client.Collections().Create(ctx, toTypesenseCollectionSchema(collection)); err != nil {
		if isTypesenseStatus(err, http.StatusConflict) {
			return nil
		}
		return fmt.Errorf("%w: create collection %q: %v", domain.ErrSchemaUnavailable, collection.Name, err)
	}
	return nil
}

func toTypesenseCollectionSchema(collection domain.CollectionSchema) *api.CollectionSchema {
	fields := make([]api.Field, 0, len(collection.Fields))
	for _, field := range collection.Fields {
		fields = append(fields, api.Field{
			Name:     field.Name,
			Type:     field.Type,
			Facet:    optionalBool(field.Facet),
			Sort:     optionalBool(field.Sort),
			Optional: optionalBool(field.Optional),
		})
	}

	defaultSortingField := strings.TrimSpace(collection.DefaultSortingField)
	return &api.CollectionSchema{
		Name:                collection.Name,
		Fields:              fields,
		DefaultSortingField: &defaultSortingField,
	}
}

func optionalBool(value bool) *bool {
	if !value {
		return nil
	}
	return &value
}
