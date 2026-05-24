package usecase

import (
	"context"
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

type SchemaRepository interface {
	GetProductsCollectionSchema(ctx context.Context) (domain.CollectionSchema, error)
	GetProductSearchPolicy(ctx context.Context) (domain.SearchPolicy, error)
	GetProductSynonymModel(ctx context.Context) (domain.SynonymModel, error)
}

type ProductSearchRepository interface {
	SearchProducts(ctx context.Context, query domain.ProductSearchQuery) (domain.ProductSearchResult, error)
}

type ProductHydrator interface {
	BatchGetProducts(ctx context.Context, ids []string) ([]domain.ProductSummary, error)
}

type ProductCatalogReader interface {
	ListSearchableProducts(ctx context.Context, req domain.ProductExportRequest) (domain.ProductExportPage, error)
}

type ReindexRepository interface {
	EnsureCollection(ctx context.Context, collection domain.CollectionSchema) error
	ImportProducts(ctx context.Context, collection string, docs []domain.ProductDocument) error
	CountDocuments(ctx context.Context, collection string) (int, error)
	SmokeSearch(ctx context.Context, collection string) error
	SwapAlias(ctx context.Context, alias string, collection string) error
	ResolveAlias(ctx context.Context, alias string) (string, error)
	CleanupOldCollections(ctx context.Context, prefix string, activeCollection string, preserveCollection string, retention time.Duration, now time.Time) ([]string, error)
}

type ReindexLock interface {
	Acquire(ctx context.Context, jobID string, ttl time.Duration) (bool, error)
	Refresh(ctx context.Context, jobID string, ttl time.Duration) (bool, error)
	Release(ctx context.Context, jobID string) error
}

type AutocompleteSearchRepository interface {
	ProductPrefixCandidates(ctx context.Context, query string, limit int) ([]domain.SuggestionCandidate, error)
	PopularQueryCandidates(ctx context.Context, query string, limit int) ([]domain.SuggestionCandidate, error)
}

type AutocompleteCache interface {
	Get(ctx context.Context, key string) ([]string, bool, error)
	Set(ctx context.Context, key string, suggestions []string, ttl time.Duration) error
}

type SynonymRepository interface {
	UpsertSynonym(ctx context.Context, synonym domain.SearchSynonym) (domain.SearchSynonym, error)
	ListSynonyms(ctx context.Context, page domain.SearchSynonymPageRequest) ([]domain.SearchSynonym, error)
}

type SearchMetricsRecorder interface {
	RecordSearch(ctx context.Context, outcome SearchMetrics)
}

type SearchMetrics struct {
	Query       string
	Sort        string
	Page        int
	PageSize    int
	Total       int
	ResultCount int
	TypesenseMS int
	HydrationMS int
	DurationMS  int
	ZeroResults bool
	ErrorCode   string
}

type NopSearchMetricsRecorder struct{}

func (NopSearchMetricsRecorder) RecordSearch(context.Context, SearchMetrics) {}

type AutocompleteMetricsRecorder interface {
	RecordAutocomplete(ctx context.Context, outcome AutocompleteMetrics)
}

type AutocompleteMetrics struct {
	QueryLength     int
	Limit           int
	CacheHit        bool
	SuggestionCount int
	DurationMS      int
	ErrorCode       string
}

type NopAutocompleteMetricsRecorder struct{}

func (NopAutocompleteMetricsRecorder) RecordAutocomplete(context.Context, AutocompleteMetrics) {}

type ZeroResultSearchTracker interface {
	Track(ctx context.Context, event domain.ZeroResultSearchEvent)
}

type ZeroResultDedupeStore interface {
	MarkFirstSeen(ctx context.Context, key string, ttl time.Duration) (bool, error)
}

type SessionEventSink interface {
	IngestSearchEvent(ctx context.Context, event domain.ZeroResultSearchEvent) error
}

type ZeroResultMetricsRecorder interface {
	RecordZeroResult(ctx context.Context, outcome ZeroResultMetrics)
}

type ZeroResultMetrics struct {
	Outcome    string
	Reason     string
	DurationMS int
}

type NopZeroResultMetricsRecorder struct{}

func (NopZeroResultMetricsRecorder) RecordZeroResult(context.Context, ZeroResultMetrics) {}

type ReindexMetricsRecorder interface {
	RecordReindexRun(ctx context.Context, outcome ReindexMetrics)
}

type ReindexMetrics struct {
	Mode            string
	Status          string
	Stage           string
	ProductsRead    int
	ProductsIndexed int
	ProductsSkipped int
	DurationMS      int
	ErrorCode       string
}

type NopReindexMetricsRecorder struct{}

func (NopReindexMetricsRecorder) RecordReindexRun(context.Context, ReindexMetrics) {}
