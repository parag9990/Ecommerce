package usecase

import (
	"context"
	"errors"
	"log/slog"
	"time"
	"unicode/utf8"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

const (
	defaultAutocompleteTypesenseTimeout   = 150 * time.Millisecond
	defaultAutocompleteCacheTimeout       = 50 * time.Millisecond
	defaultAutocompletePrefixCacheTTL     = time.Minute
	defaultAutocompleteEmptyQueryCacheTTL = 5 * time.Minute
)

type AutocompleteOptions struct {
	DefaultLimit       int
	MaxLimit           int
	TypesenseTimeout   time.Duration
	CacheTimeout       time.Duration
	PrefixCacheTTL     time.Duration
	EmptyQueryCacheTTL time.Duration
	Metrics            AutocompleteMetricsRecorder
}

type AutocompleteUsecase struct {
	repo    AutocompleteSearchRepository
	cache   AutocompleteCache
	options AutocompleteOptions
	logger  *slog.Logger
}

func NewAutocompleteUsecase(repo AutocompleteSearchRepository, cache AutocompleteCache, options AutocompleteOptions, logger *slog.Logger) (*AutocompleteUsecase, error) {
	if repo == nil {
		return nil, errors.New("autocomplete repository is required")
	}
	if cache == nil {
		return nil, errors.New("autocomplete cache is required")
	}
	if options.DefaultLimit <= 0 {
		options.DefaultLimit = domain.DefaultAutocompleteLimit
	}
	if options.MaxLimit <= 0 {
		options.MaxLimit = domain.MaxAutocompleteLimit
	}
	if options.DefaultLimit > options.MaxLimit {
		options.DefaultLimit = options.MaxLimit
	}
	if options.TypesenseTimeout <= 0 {
		options.TypesenseTimeout = defaultAutocompleteTypesenseTimeout
	}
	if options.CacheTimeout <= 0 {
		options.CacheTimeout = defaultAutocompleteCacheTimeout
	}
	if options.PrefixCacheTTL <= 0 {
		options.PrefixCacheTTL = defaultAutocompletePrefixCacheTTL
	}
	if options.EmptyQueryCacheTTL <= 0 {
		options.EmptyQueryCacheTTL = defaultAutocompleteEmptyQueryCacheTTL
	}
	if options.Metrics == nil {
		options.Metrics = NopAutocompleteMetricsRecorder{}
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &AutocompleteUsecase{
		repo:    repo,
		cache:   cache,
		options: options,
		logger:  logger,
	}, nil
}

func (u *AutocompleteUsecase) Execute(ctx context.Context, req domain.AutocompleteRequest) (domain.AutocompleteResponse, error) {
	started := time.Now()
	metrics := AutocompleteMetrics{}
	defer func() {
		metrics.DurationMS = elapsedMS(started)
		u.options.Metrics.RecordAutocomplete(ctx, metrics)
	}()

	input, err := domain.NormalizeAutocompleteRequest(req, u.options.DefaultLimit, u.options.MaxLimit)
	if err != nil {
		metrics.ErrorCode = "invalid_request"
		return domain.AutocompleteResponse{}, err
	}
	metrics.QueryLength = utf8.RuneCountInString(input.NormalizedQuery)
	metrics.Limit = input.Limit

	cacheKey := domain.BuildAutocompleteCacheKey(input.NormalizedQuery, input.Limit)
	suggestions, ok, err := u.getCachedSuggestions(ctx, cacheKey)
	if err != nil {
		u.logger.Warn("search.autocomplete.cache_get_failed",
			slog.String("request_id", requestID(ctx)),
			slog.String("error", err.Error()),
		)
	}
	if ok {
		metrics.CacheHit = true
		metrics.SuggestionCount = len(suggestions)
		u.logAutocomplete(ctx, input, true, len(suggestions), started)
		return domain.AutocompleteResponse{Suggestions: suggestions}, nil
	}

	popular, productCandidates, err := u.fetchCandidates(ctx, input)
	if err != nil {
		metrics.ErrorCode = "search_backend_unavailable"
		return domain.AutocompleteResponse{}, err
	}

	suggestions = domain.MergeAutocompleteCandidates(input.NormalizedQuery, input.Limit, popular, productCandidates)
	metrics.SuggestionCount = len(suggestions)
	if err := u.setCachedSuggestions(ctx, cacheKey, suggestions, autocompleteCacheTTL(input.NormalizedQuery, u.options)); err != nil {
		u.logger.Warn("search.autocomplete.cache_set_failed",
			slog.String("request_id", requestID(ctx)),
			slog.String("error", err.Error()),
		)
	}

	u.logAutocomplete(ctx, input, false, len(suggestions), started)
	return domain.AutocompleteResponse{Suggestions: suggestions}, nil
}

func (u *AutocompleteUsecase) fetchCandidates(ctx context.Context, input domain.AutocompleteInput) ([]domain.SuggestionCandidate, []domain.SuggestionCandidate, error) {
	searchCtx, cancel := context.WithTimeout(ctx, u.options.TypesenseTimeout)
	defer cancel()

	popular, err := u.repo.PopularQueryCandidates(searchCtx, input.NormalizedQuery, input.Limit)
	if err != nil {
		return nil, nil, errors.Join(domain.ErrSearchBackendUnavailable, err)
	}

	if utf8.RuneCountInString(input.NormalizedQuery) < domain.MinAutocompleteProductPrefix {
		return popular, []domain.SuggestionCandidate{}, nil
	}

	productCandidates, err := u.repo.ProductPrefixCandidates(searchCtx, input.NormalizedQuery, input.Limit)
	if err != nil {
		return nil, nil, errors.Join(domain.ErrSearchBackendUnavailable, err)
	}
	return popular, productCandidates, nil
}

func (u *AutocompleteUsecase) getCachedSuggestions(ctx context.Context, key string) ([]string, bool, error) {
	cacheCtx, cancel := context.WithTimeout(ctx, u.options.CacheTimeout)
	defer cancel()
	return u.cache.Get(cacheCtx, key)
}

func (u *AutocompleteUsecase) setCachedSuggestions(ctx context.Context, key string, suggestions []string, ttl time.Duration) error {
	cacheCtx, cancel := context.WithTimeout(ctx, u.options.CacheTimeout)
	defer cancel()
	return u.cache.Set(cacheCtx, key, suggestions, ttl)
}

func autocompleteCacheTTL(normalizedQuery string, options AutocompleteOptions) time.Duration {
	if normalizedQuery == "" {
		return options.EmptyQueryCacheTTL
	}
	return options.PrefixCacheTTL
}

func (u *AutocompleteUsecase) logAutocomplete(ctx context.Context, input domain.AutocompleteInput, cacheHit bool, suggestionCount int, started time.Time) {
	u.logger.Info("search.autocomplete",
		slog.String("request_id", requestID(ctx)),
		slog.Int("query_length", utf8.RuneCountInString(input.NormalizedQuery)),
		slog.Int("limit", input.Limit),
		slog.Bool("cache_hit", cacheHit),
		slog.Int("suggestion_count", suggestionCount),
		slog.Int("duration_ms", elapsedMS(started)),
	)
}
