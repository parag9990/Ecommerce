package repository

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

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

// EnsureAliasedCollection keeps the stable API collection name as an alias so
// future schema rebuilds can switch collections without interrupting reads.
// Existing physical collections are left in place and migrated by reindex.
func (r *TypesenseCollectionRepository) EnsureAliasedCollection(ctx context.Context, alias string, prefix string, collection domain.CollectionSchema, now time.Time) (string, error) {
	alias = strings.TrimSpace(alias)
	prefix = strings.TrimSpace(prefix)
	if alias == "" {
		return "", errors.New("typesense collection alias is required")
	}
	if prefix == "" {
		prefix = alias
	}

	resolved, err := r.client.Alias(alias).Retrieve(ctx)
	if err == nil {
		if resolved == nil || strings.TrimSpace(resolved.CollectionName) == "" {
			return "", fmt.Errorf("%w: alias %q has no target collection", domain.ErrSchemaUnavailable, alias)
		}
		return strings.TrimSpace(resolved.CollectionName), nil
	}
	if !isTypesenseStatus(err, http.StatusNotFound) {
		return "", fmt.Errorf("%w: resolve collection alias %q: %v", domain.ErrSchemaUnavailable, alias, err)
	}

	if _, err := r.client.Collection(alias).Retrieve(ctx); err == nil {
		return alias, nil
	} else if !isTypesenseStatus(err, http.StatusNotFound) {
		return "", fmt.Errorf("%w: retrieve legacy collection %q: %v", domain.ErrSchemaUnavailable, alias, err)
	}

	if now.IsZero() {
		now = time.Now()
	}
	target := prefix + "_" + now.UTC().Format("20060102_150405")
	collection.Name = target
	if err := r.EnsureCollection(ctx, collection); err != nil {
		return "", err
	}
	if _, err := r.client.Aliases().Upsert(ctx, alias, &api.CollectionAliasSchema{CollectionName: target}); err != nil {
		// Another replica may have completed bootstrap while this one created the
		// same physical collection.
		resolved, resolveErr := r.client.Alias(alias).Retrieve(ctx)
		if resolveErr == nil && resolved != nil && strings.TrimSpace(resolved.CollectionName) != "" {
			return strings.TrimSpace(resolved.CollectionName), nil
		}
		return "", fmt.Errorf("%w: create collection alias %q: %v", domain.ErrSchemaUnavailable, alias, err)
	}
	return target, nil
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
