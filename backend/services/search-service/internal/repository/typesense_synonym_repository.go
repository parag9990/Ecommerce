package repository

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/typesense/typesense-go/v2/typesense"
	"github.com/typesense/typesense-go/v2/typesense/api"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

type TypesenseSynonymRepository struct {
	client     *typesense.Client
	collection string
}

func NewTypesenseSynonymRepository(client *typesense.Client, collection string) (*TypesenseSynonymRepository, error) {
	if client == nil {
		return nil, errors.New("typesense client is required")
	}
	collection = strings.TrimSpace(collection)
	if collection == "" {
		return nil, errors.New("typesense products collection is required")
	}
	return &TypesenseSynonymRepository{client: client, collection: collection}, nil
}

func (r *TypesenseSynonymRepository) UpsertSynonym(ctx context.Context, synonym domain.SearchSynonym) (domain.SearchSynonym, error) {
	if err := synonym.Validate(); err != nil {
		return domain.SearchSynonym{}, err
	}

	root := synonym.Root
	payload := &api.SearchSynonymSchema{
		Root:     &root,
		Synonyms: append([]string(nil), synonym.Synonyms...),
	}
	saved, err := r.client.Collection(r.collection).Synonyms().Upsert(ctx, synonym.ID, payload)
	if err != nil {
		return domain.SearchSynonym{}, mapTypesenseSynonymError("upsert synonym", err)
	}
	return mapTypesenseSynonym(saved, synonym), nil
}

func (r *TypesenseSynonymRepository) GetSynonym(ctx context.Context, id string) (domain.SearchSynonym, bool, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return domain.SearchSynonym{}, false, domain.ErrInvalidSynonym
	}
	item, err := r.client.Collection(r.collection).Synonym(id).Retrieve(ctx)
	if err != nil {
		var httpErr *typesense.HTTPError
		if errors.As(err, &httpErr) && httpErr.Status == http.StatusNotFound {
			return domain.SearchSynonym{}, false, nil
		}
		return domain.SearchSynonym{}, false, mapTypesenseSynonymError("get synonym", err)
	}
	return mapTypesenseSynonym(item, domain.SearchSynonym{ID: id}), true, nil
}

func (r *TypesenseSynonymRepository) DeleteSynonym(ctx context.Context, id string) (domain.SearchSynonym, bool, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return domain.SearchSynonym{}, false, domain.ErrInvalidSynonym
	}
	item, err := r.client.Collection(r.collection).Synonym(id).Delete(ctx)
	if err != nil {
		var httpErr *typesense.HTTPError
		if errors.As(err, &httpErr) && httpErr.Status == http.StatusNotFound {
			return domain.SearchSynonym{}, false, nil
		}
		return domain.SearchSynonym{}, false, mapTypesenseSynonymError("delete synonym", err)
	}
	return mapTypesenseSynonym(item, domain.SearchSynonym{ID: id}), true, nil
}

func (r *TypesenseSynonymRepository) ListSynonyms(ctx context.Context, page domain.SearchSynonymPageRequest) ([]domain.SearchSynonym, error) {
	page = domain.NormalizeSearchSynonymPageRequest(page)
	items, err := r.client.Collection(r.collection).Synonyms().Retrieve(ctx)
	if err != nil {
		return nil, mapTypesenseSynonymError("list synonyms", err)
	}

	synonyms := make([]domain.SearchSynonym, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		synonym := mapTypesenseSynonym(item, domain.SearchSynonym{})
		if strings.TrimSpace(synonym.ID) == "" {
			continue
		}
		synonyms = append(synonyms, synonym)
	}
	sort.SliceStable(synonyms, func(i, j int) bool {
		if synonyms[i].Root == synonyms[j].Root {
			return synonyms[i].ID < synonyms[j].ID
		}
		return synonyms[i].Root < synonyms[j].Root
	})

	start := (page.Page - 1) * page.PageSize
	if start >= len(synonyms) {
		return []domain.SearchSynonym{}, nil
	}
	end := start + page.PageSize
	if end > len(synonyms) {
		end = len(synonyms)
	}
	return append([]domain.SearchSynonym(nil), synonyms[start:end]...), nil
}

func mapTypesenseSynonym(item *api.SearchSynonym, fallback domain.SearchSynonym) domain.SearchSynonym {
	if item == nil {
		return fallback
	}
	result := domain.SearchSynonym{
		ID:       fallback.ID,
		Root:     fallback.Root,
		Synonyms: append([]string(nil), fallback.Synonyms...),
	}
	if item.Id != nil {
		result.ID = strings.TrimSpace(*item.Id)
	}
	if item.Root != nil {
		result.Root = domain.NormalizeSynonymTerm(*item.Root)
	}
	if item.Synonyms != nil {
		result.Synonyms = make([]string, 0, len(item.Synonyms))
		for _, raw := range item.Synonyms {
			term := domain.NormalizeSynonymTerm(raw)
			if term != "" {
				result.Synonyms = append(result.Synonyms, term)
			}
		}
	}
	return result
}

func mapTypesenseSynonymError(operation string, err error) error {
	var httpErr *typesense.HTTPError
	if errors.As(err, &httpErr) {
		switch httpErr.Status {
		case http.StatusNotFound:
			return fmt.Errorf("%w: %s failed with status %d", domain.ErrSearchCollectionUnavailable, operation, httpErr.Status)
		case http.StatusUnauthorized, http.StatusForbidden:
			return fmt.Errorf("%w: %s credentials rejected with status %d", domain.ErrSearchBackendUnavailable, operation, httpErr.Status)
		default:
			return fmt.Errorf("%w: %s failed with status %d", domain.ErrSearchBackendUnavailable, operation, httpErr.Status)
		}
	}
	return fmt.Errorf("%w: %s: %v", domain.ErrSearchBackendUnavailable, operation, err)
}
