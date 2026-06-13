package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/typesense/typesense-go/v2/typesense"
	"github.com/typesense/typesense-go/v2/typesense/api"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

type TypesenseReindexRepository struct {
	client *typesense.Client
}

func NewTypesenseReindexRepository(client *typesense.Client) (*TypesenseReindexRepository, error) {
	if client == nil {
		return nil, errors.New("typesense client is required")
	}
	return &TypesenseReindexRepository{client: client}, nil
}

func (r *TypesenseReindexRepository) EnsureCollection(ctx context.Context, collection domain.CollectionSchema) error {
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

func (r *TypesenseReindexRepository) ImportProducts(ctx context.Context, collection string, docs []domain.ProductDocument) error {
	collection = strings.TrimSpace(collection)
	if collection == "" {
		return fmt.Errorf("%w: collection is required", domain.ErrInvalidReindexRequest)
	}
	if len(docs) == 0 {
		return nil
	}

	importDocs := make([]interface{}, 0, len(docs))
	for _, doc := range docs {
		if err := doc.Validate(); err != nil {
			return err
		}
		importDocs = append(importDocs, doc)
	}

	responses, err := r.client.Collection(collection).Documents().Import(ctx, importDocs, &api.ImportDocumentsParams{
		Action:    stringPtr("upsert"),
		BatchSize: intPtr(len(importDocs)),
	})
	if err != nil {
		return fmt.Errorf("%w: import products into %q: %v", domain.ErrProductIndexUnavailable, collection, err)
	}
	failures := make([]string, 0)
	for i, response := range responses {
		if response == nil {
			failures = append(failures, fmt.Sprintf("document[%d]: empty import response", i))
			continue
		}
		if response.Success {
			continue
		}
		failures = append(failures, formatImportFailure(i, response))
	}
	if len(failures) > 0 {
		return fmt.Errorf("%w: import products into %q failed for %d documents: %s", domain.ErrProductIndexUnavailable, collection, len(failures), strings.Join(failures, "; "))
	}
	return nil
}

func (r *TypesenseReindexRepository) CountDocuments(ctx context.Context, collection string) (int, error) {
	collection = strings.TrimSpace(collection)
	if collection == "" {
		return 0, fmt.Errorf("%w: collection is required", domain.ErrInvalidReindexRequest)
	}
	resp, err := r.client.Collection(collection).Retrieve(ctx)
	if err != nil {
		return 0, fmt.Errorf("%w: count documents in %q: %v", domain.ErrSearchCollectionUnavailable, collection, err)
	}
	if resp.NumDocuments == nil {
		return 0, nil
	}
	return int(*resp.NumDocuments), nil
}

func (r *TypesenseReindexRepository) SmokeSearch(ctx context.Context, collection string) error {
	collection = strings.TrimSpace(collection)
	if collection == "" {
		return fmt.Errorf("%w: collection is required", domain.ErrInvalidReindexRequest)
	}
	result, err := r.client.Collection(collection).Documents().Search(ctx, &api.SearchCollectionParams{
		Q:             stringPtr("*"),
		QueryBy:       stringPtr(domain.ProductFieldTitle),
		IncludeFields: stringPtr(domain.ProductFieldID),
		PerPage:       intPtr(1),
	})
	if err != nil {
		return fmt.Errorf("%w: smoke search %q: %v", domain.ErrSearchBackendUnavailable, collection, err)
	}
	if result == nil || result.Found == nil || *result.Found <= 0 {
		return fmt.Errorf("%w: smoke search returned no products", domain.ErrReindexValidationFailed)
	}
	return nil
}

func (r *TypesenseReindexRepository) SwapAlias(ctx context.Context, alias string, collection string) error {
	alias = strings.TrimSpace(alias)
	collection = strings.TrimSpace(collection)
	if alias == "" || collection == "" {
		return fmt.Errorf("%w: alias and collection are required", domain.ErrInvalidReindexRequest)
	}
	if _, err := r.client.Aliases().Upsert(ctx, alias, &api.CollectionAliasSchema{CollectionName: collection}); err != nil {
		return fmt.Errorf("%w: swap alias %q to %q: %v", domain.ErrSearchCollectionUnavailable, alias, collection, err)
	}
	return nil
}

func (r *TypesenseReindexRepository) ResolveAlias(ctx context.Context, alias string) (string, error) {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return "", fmt.Errorf("%w: alias is required", domain.ErrInvalidReindexRequest)
	}
	resolved, err := r.client.Alias(alias).Retrieve(ctx)
	if err != nil {
		if isTypesenseStatus(err, http.StatusNotFound) {
			return "", nil
		}
		return "", fmt.Errorf("%w: resolve alias %q: %v", domain.ErrSearchCollectionUnavailable, alias, err)
	}
	if resolved == nil {
		return "", nil
	}
	return strings.TrimSpace(resolved.CollectionName), nil
}

func (r *TypesenseReindexRepository) CleanupOldCollections(ctx context.Context, prefix string, activeCollection string, preserveCollection string, retention time.Duration, now time.Time) ([]string, error) {
	prefix = strings.TrimSpace(prefix)
	activeCollection = strings.TrimSpace(activeCollection)
	preserveCollection = strings.TrimSpace(preserveCollection)
	if prefix == "" || retention <= 0 {
		return []string{}, nil
	}
	if now.IsZero() {
		now = time.Now()
	}

	collections, err := r.client.Collections().Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: list collections for cleanup: %v", domain.ErrSearchCollectionUnavailable, err)
	}

	cutoff := now.UTC().Add(-retention)
	deleted := make([]string, 0)
	for _, collection := range collections {
		if collection == nil {
			continue
		}
		name := strings.TrimSpace(collection.Name)
		if !strings.HasPrefix(name, prefix+"_") || name == activeCollection || name == preserveCollection {
			continue
		}
		createdAt, ok := collectionCreatedAt(collection, name, prefix)
		if !ok || createdAt.After(cutoff) {
			continue
		}
		if _, err := r.client.Collection(name).Delete(ctx); err != nil {
			return deleted, fmt.Errorf("%w: delete old collection %q: %v", domain.ErrSearchCollectionUnavailable, name, err)
		}
		deleted = append(deleted, name)
	}
	return deleted, nil
}

func formatImportFailure(index int, response *api.ImportDocumentResponse) string {
	productID := importFailureProductID(response.Document)
	if productID == "" {
		productID = fmt.Sprintf("document[%d]", index)
	}
	if strings.TrimSpace(response.Error) == "" {
		return productID
	}
	return productID + ": " + strings.TrimSpace(response.Error)
}

func importFailureProductID(raw string) string {
	var doc struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		return ""
	}
	return strings.TrimSpace(doc.ID)
}

func collectionCreatedAt(collection *api.CollectionResponse, name string, prefix string) (time.Time, bool) {
	if collection.CreatedAt != nil && *collection.CreatedAt > 0 {
		return time.Unix(*collection.CreatedAt, 0).UTC(), true
	}
	raw := strings.TrimPrefix(name, prefix+"_")
	createdAt, err := time.ParseInLocation("20060102_150405", raw, time.UTC)
	if err != nil {
		return time.Time{}, false
	}
	return createdAt, true
}
