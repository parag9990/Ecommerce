package usecase

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"unicode/utf8"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/requestctx"
)

const (
	defaultTypesenseSearchTimeout  = 300 * time.Millisecond
	defaultProductHydrationTimeout = 300 * time.Millisecond
)

type SearchProductsOptions struct {
	Limits            domain.SearchLimits
	TypesenseTimeout  time.Duration
	HydrationTimeout  time.Duration
	Metrics           SearchMetricsRecorder
	ZeroResultTracker ZeroResultSearchTracker
}

type SearchProductsUsecase struct {
	schemaRepo      SchemaRepository
	productSearch   ProductSearchRepository
	productHydrator ProductHydrator
	options         SearchProductsOptions
	logger          *slog.Logger
}

func NewSearchProductsUsecase(schemaRepo SchemaRepository, productSearch ProductSearchRepository, productHydrator ProductHydrator, options SearchProductsOptions, logger *slog.Logger) (*SearchProductsUsecase, error) {
	if schemaRepo == nil {
		return nil, errors.New("schema repository is required")
	}
	if productSearch == nil {
		return nil, errors.New("product search repository is required")
	}
	if productHydrator == nil {
		return nil, errors.New("product hydrator is required")
	}
	if options.TypesenseTimeout <= 0 {
		options.TypesenseTimeout = defaultTypesenseSearchTimeout
	}
	if options.HydrationTimeout <= 0 {
		options.HydrationTimeout = defaultProductHydrationTimeout
	}
	if options.Metrics == nil {
		options.Metrics = NopSearchMetricsRecorder{}
	}
	if options.ZeroResultTracker == nil {
		options.ZeroResultTracker = NopZeroResultTracker{}
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &SearchProductsUsecase{
		schemaRepo:      schemaRepo,
		productSearch:   productSearch,
		productHydrator: productHydrator,
		options:         options,
		logger:          logger,
	}, nil
}

func (u *SearchProductsUsecase) Execute(ctx context.Context, req domain.SearchRequest) (domain.SearchResponse, error) {
	started := time.Now()
	metrics := SearchMetrics{}
	defer func() {
		metrics.DurationMS = elapsedMS(started)
		metrics.ZeroResults = metrics.Total == 0 && metrics.ErrorCode == ""
		u.options.Metrics.RecordSearch(ctx, metrics)
	}()

	policy, err := u.schemaRepo.GetProductSearchPolicy(ctx)
	if err != nil {
		metrics.ErrorCode = "schema_unavailable"
		return domain.SearchResponse{}, err
	}

	input, err := domain.NormalizeSearchRequest(req, policy, u.options.Limits)
	if err != nil {
		metrics.ErrorCode = "invalid_request"
		return domain.SearchResponse{}, err
	}
	metrics.Query = input.Query
	metrics.Sort = input.Sort
	metrics.Page = input.Page
	metrics.PageSize = input.PageSize

	sortBy, err := domain.ResolveSearchSort(policy, input.Sort)
	if err != nil {
		metrics.ErrorCode = "unsupported_sort"
		return domain.SearchResponse{}, err
	}
	filterBy, err := domain.BuildProductFilterBy(input.Filters)
	if err != nil {
		metrics.ErrorCode = "invalid_filter"
		return domain.SearchResponse{}, err
	}

	searchStarted := time.Now()
	searchCtx, cancelSearch := context.WithTimeout(ctx, u.options.TypesenseTimeout)
	searchResult, err := u.productSearch.SearchProducts(searchCtx, domain.ProductSearchQuery{
		Query:          input.Query,
		QueryBy:        policy.DefaultQueryBy,
		QueryByWeights: policy.QueryByWeights,
		NumTypos:       policy.NumTypos,
		Prefix:         policy.Prefix,
		FacetBy:        policy.DefaultFacetBy,
		FilterBy:       filterBy,
		SortBy:         sortBy,
		Page:           input.Page,
		PageSize:       input.PageSize,
	})
	cancelSearch()
	metrics.TypesenseMS = elapsedMS(searchStarted)
	if err != nil {
		metrics.ErrorCode = "search_backend_unavailable"
		return domain.SearchResponse{}, errors.Join(domain.ErrSearchBackendUnavailable, err)
	}

	facets := filterFacets(searchResult.Facets, domain.AllowedFacetSet(policy))
	metrics.Total = searchResult.Total
	if searchResult.Total == 0 {
		u.trackZeroResult(ctx, input)
	}
	if len(searchResult.IDs) == 0 {
		u.logSearch(ctx, input, searchResult, 0, metrics.TypesenseMS, 0)
		return domain.SearchResponse{
			Products: []domain.ProductSummary{},
			Facets:   facets,
			Total:    searchResult.Total,
		}, nil
	}

	hydrationStarted := time.Now()
	hydrationCtx, cancelHydration := context.WithTimeout(ctx, u.options.HydrationTimeout)
	products, err := u.productHydrator.BatchGetProducts(hydrationCtx, searchResult.IDs)
	cancelHydration()
	metrics.HydrationMS = elapsedMS(hydrationStarted)
	if err != nil {
		metrics.ErrorCode = "product_hydration_unavailable"
		return domain.SearchResponse{}, errors.Join(domain.ErrProductHydrationUnavailable, err)
	}

	orderedProducts := orderProductsByIDs(searchResult.IDs, products)
	metrics.ResultCount = len(orderedProducts)
	u.logSearch(ctx, input, searchResult, len(orderedProducts), metrics.TypesenseMS, metrics.HydrationMS)

	return domain.SearchResponse{
		Products: orderedProducts,
		Facets:   facets,
		Total:    searchResult.Total,
	}, nil
}

func (u *SearchProductsUsecase) trackZeroResult(ctx context.Context, input domain.SearchInput) {
	analytics := requestctx.Analytics(ctx)
	reqID := analytics.RequestID
	if reqID == "" {
		reqID = requestID(ctx)
	}
	u.options.ZeroResultTracker.Track(ctx, domain.ZeroResultSearchEvent{
		Query:           input.Query,
		NormalizedQuery: domain.NormalizeAnalyticsQuery(input.Query),
		Filters:         domain.AnalyticsFilters(input),
		Sort:            input.Sort,
		Page:            input.Page,
		PageSize:        input.PageSize,
		RequestID:       reqID,
		AnonymousID:     analytics.AnonymousID,
		SessionID:       analytics.SessionID,
		UserID:          analytics.UserID,
		Path:            analytics.ClientPath,
		OccurredAt:      time.Now().UTC(),
	})
}

func filterFacets(facets map[string][]domain.FacetValue, allowed map[string]struct{}) map[string][]domain.FacetValue {
	filtered := make(map[string][]domain.FacetValue, len(allowed))
	for field := range allowed {
		values := facets[field]
		if values == nil {
			values = []domain.FacetValue{}
		}
		filtered[field] = append([]domain.FacetValue(nil), values...)
	}
	return filtered
}

func orderProductsByIDs(ids []string, products []domain.ProductSummary) []domain.ProductSummary {
	byID := make(map[string]domain.ProductSummary, len(products))
	for _, product := range products {
		if product.ProductID == "" {
			continue
		}
		byID[product.ProductID] = product
	}

	ordered := make([]domain.ProductSummary, 0, len(ids))
	seen := map[string]struct{}{}
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		product, ok := byID[id]
		if !ok {
			continue
		}
		ordered = append(ordered, product)
	}
	return ordered
}

func (u *SearchProductsUsecase) logSearch(ctx context.Context, input domain.SearchInput, result domain.ProductSearchResult, productCount int, typesenseMS int, hydrationMS int) {
	normalizedQuery := domain.NormalizeAnalyticsQuery(input.Query)
	u.logger.Info("search.products",
		slog.String("request_id", requestID(ctx)),
		slog.String("query_hash", domain.ShortHash(normalizedQuery)),
		slog.Int("query_length", utf8.RuneCountInString(normalizedQuery)),
		slog.String("sort", input.Sort),
		slog.Int("page", input.Page),
		slog.Int("page_size", input.PageSize),
		slog.Int("total", result.Total),
		slog.Int("product_count", productCount),
		slog.Int("typesense_ms", typesenseMS),
		slog.Int("hydration_ms", hydrationMS),
		slog.Int("typesense_reported_ms", result.SearchTimeMS),
	)
}

func elapsedMS(started time.Time) int {
	return int(time.Since(started) / time.Millisecond)
}

func requestID(ctx context.Context) string {
	return requestctx.RequestID(ctx)
}
