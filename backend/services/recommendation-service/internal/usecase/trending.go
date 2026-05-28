package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
)

type TrendingServiceConfig struct {
	Weights             domain.PopularityWeights
	FormulaVersion      string
	Limit               int
	MaxIdentifierLength int
	CacheKeyPrefix      string
	TTL                 time.Duration
	RebuildInterval     time.Duration
}

type TrendingGenerationResult struct {
	EligibleCandidates int
	SetsGenerated      int
	ItemsGenerated     int
}

type RankedSetResolution struct {
	Set      domain.RecommendationSet
	Fallback bool
	Empty    bool
}

type TrendingService struct {
	features   RankingFeatureReader
	sets       RecommendationSetRepository
	cache      RankingResultCache
	metrics    RankingMetrics
	logger     *slog.Logger
	cfg        TrendingServiceConfig
	keyBuilder domain.CacheKeyBuilder
	now        func() time.Time
}

type rankingTarget struct {
	scope      domain.RankingScope
	scopeID    string
	contextKey string
	strategyID domain.StrategyID
	cacheKey   string
	reason     string
}

func NewTrendingService(
	features RankingFeatureReader,
	sets RecommendationSetRepository,
	cache RankingResultCache,
	cfg TrendingServiceConfig,
	metrics RankingMetrics,
	logger *slog.Logger,
) (*TrendingService, error) {
	if features == nil {
		return nil, errors.New("ranking feature reader is required")
	}
	if sets == nil {
		return nil, errors.New("recommendation set repository is required")
	}
	if cfg.Weights == (domain.PopularityWeights{}) {
		cfg.Weights = domain.DefaultPopularityWeights()
	}
	if err := cfg.Weights.Validate(); err != nil {
		return nil, err
	}
	cfg.FormulaVersion = strings.TrimSpace(cfg.FormulaVersion)
	if cfg.FormulaVersion == "" {
		cfg.FormulaVersion = domain.RuleBasedRankingFormulaVersion
	}
	if cfg.Weights != domain.DefaultPopularityWeights() && cfg.FormulaVersion == domain.RuleBasedRankingFormulaVersion {
		return nil, errors.New("ranking formula version must change when ranking weights differ from popularity_v1")
	}
	if cfg.Limit <= 0 {
		return nil, errors.New("ranking limit must be greater than zero")
	}
	if cfg.MaxIdentifierLength <= 0 {
		cfg.MaxIdentifierLength = 128
	}
	if cfg.TTL <= 0 {
		cfg.TTL = domain.DefaultRecommendationCacheTTL
	}
	if cfg.RebuildInterval < 0 {
		return nil, errors.New("ranking rebuild interval cannot be negative")
	}
	builder, err := domain.NewCacheKeyBuilder(cfg.CacheKeyPrefix, cfg.MaxIdentifierLength)
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &TrendingService{
		features:   features,
		sets:       sets,
		cache:      cache,
		metrics:    metrics,
		logger:     logger,
		cfg:        cfg,
		keyBuilder: builder,
		now:        func() time.Time { return time.Now().UTC() },
	}, nil
}

func (s *TrendingService) GenerateAll(ctx context.Context) (TrendingGenerationResult, error) {
	if err := ctx.Err(); err != nil {
		return TrendingGenerationResult{}, err
	}
	startedAt := s.now()
	features, err := s.features.ListEligibleProductFeatures(ctx)
	if err != nil {
		s.recordRun("failed", startedAt)
		return TrendingGenerationResult{}, fmt.Errorf("%w: list ranking feature candidates: %v", domain.ErrRecommendationStorage, err)
	}

	global := domain.ScoreAndSortRuleBasedProducts(features, s.cfg.Weights)
	categoryProducts := make(map[string][]domain.ScoredProduct)
	sellerProducts := make(map[string][]domain.ScoredProduct)
	for _, product := range global {
		if product.CategoryID != "" {
			categoryProducts[product.CategoryID] = append(categoryProducts[product.CategoryID], product)
		}
		if product.SellerID != "" {
			sellerProducts[product.SellerID] = append(sellerProducts[product.SellerID], product)
		}
	}

	result := TrendingGenerationResult{EligibleCandidates: len(global)}
	if s.metrics != nil {
		s.metrics.SetRankingCandidates(string(domain.RankingScopeGlobal), len(global))
		s.metrics.SetRankingCandidates(string(domain.RankingScopeCategory), groupedCandidateCount(categoryProducts))
		s.metrics.SetRankingCandidates(string(domain.RankingScopeSeller), groupedCandidateCount(sellerProducts))
	}
	globalTarget, _ := s.targetFor(domain.RankingScopeGlobal, "")
	if err := s.publishSet(ctx, globalTarget, global, startedAt, &result); err != nil {
		s.recordRun("failed", startedAt)
		return result, err
	}
	for _, categoryID := range sortedKeys(categoryProducts) {
		target, err := s.targetFor(domain.RankingScopeCategory, categoryID)
		if err != nil {
			s.logger.WarnContext(ctx, "recommendation.ranking.scope_skipped",
				slog.String("scope", string(domain.RankingScopeCategory)),
				slog.String("scope_id", categoryID),
				slog.String("error", err.Error()),
			)
			continue
		}
		if err := s.publishSet(ctx, target, categoryProducts[categoryID], startedAt, &result); err != nil {
			s.recordRun("failed", startedAt)
			return result, err
		}
	}
	for _, sellerID := range sortedKeys(sellerProducts) {
		target, err := s.targetFor(domain.RankingScopeSeller, sellerID)
		if err != nil {
			s.logger.WarnContext(ctx, "recommendation.ranking.scope_skipped",
				slog.String("scope", string(domain.RankingScopeSeller)),
				slog.String("scope_id", sellerID),
				slog.String("error", err.Error()),
			)
			continue
		}
		if err := s.publishSet(ctx, target, sellerProducts[sellerID], startedAt, &result); err != nil {
			s.recordRun("failed", startedAt)
			return result, err
		}
	}

	s.recordRun("success", startedAt)
	s.logger.InfoContext(ctx, "recommendation.ranking.completed",
		slog.String("formula_version", s.cfg.FormulaVersion),
		slog.Int("eligible_candidates", result.EligibleCandidates),
		slog.Int("sets_generated", result.SetsGenerated),
		slog.Int("items_generated", result.ItemsGenerated),
		slog.Int64("duration_ms", s.now().Sub(startedAt).Milliseconds()),
	)
	return result, nil
}

func (s *TrendingService) RunScheduled(ctx context.Context) {
	if _, err := s.GenerateAll(ctx); err != nil && !errors.Is(err, context.Canceled) {
		s.logger.ErrorContext(ctx, "recommendation.ranking.initial_generation_failed", slog.String("error", err.Error()))
	}
	if s.cfg.RebuildInterval <= 0 {
		return
	}
	ticker := time.NewTicker(s.cfg.RebuildInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := s.GenerateAll(ctx); err != nil && !errors.Is(err, context.Canceled) {
				s.logger.ErrorContext(ctx, "recommendation.ranking.generation_failed", slog.String("error", err.Error()))
			}
		}
	}
}

// ResolveGeneratedSet is the internal read path for generated lists and their safe fallback.
func (s *TrendingService) ResolveGeneratedSet(ctx context.Context, request domain.RankRequest) (RankedSetResolution, error) {
	if err := request.Validate(s.cfg.MaxIdentifierLength); err != nil {
		return RankedSetResolution{}, err
	}
	if err := ctx.Err(); err != nil {
		return RankedSetResolution{}, err
	}
	now := request.Now.UTC()
	if request.Now.IsZero() {
		now = s.now()
	}
	target, err := s.targetFor(request.Scope, strings.TrimSpace(request.ScopeID))
	if err != nil {
		return RankedSetResolution{}, err
	}
	primary, found := s.loadSet(ctx, target, now)
	if found && len(primary.Items) > 0 {
		return RankedSetResolution{Set: trimItems(primary, request.Limit)}, nil
	}
	if request.Scope == domain.RankingScopeGlobal {
		return RankedSetResolution{Set: trimItems(primary, request.Limit), Empty: true}, nil
	}
	if s.metrics != nil {
		s.metrics.RecordRankingFallback(string(request.Scope))
	}
	globalTarget, _ := s.targetFor(domain.RankingScopeGlobal, "")
	global, globalFound := s.loadSet(ctx, globalTarget, now)
	if globalFound && len(global.Items) > 0 {
		return RankedSetResolution{Set: trimItems(global, request.Limit), Fallback: true}, nil
	}
	return RankedSetResolution{Set: trimItems(primary, request.Limit), Empty: true}, nil
}

func (s *TrendingService) publishSet(
	ctx context.Context,
	target rankingTarget,
	ranked []domain.ScoredProduct,
	generatedAt time.Time,
	result *TrendingGenerationResult,
) error {
	items := recommendationItems(ranked, s.cfg.Limit, target.reason)
	set := domain.RecommendationSet{
		ID:         recommendationSetID(target),
		ContextKey: target.contextKey,
		Type:       domain.RecommendationTypeTrending,
		StrategyID: target.strategyID,
		Items:      items,
		Metadata: map[string]string{
			"scope":           string(target.scope),
			"formula_version": s.cfg.FormulaVersion,
			"limit":           strconv.Itoa(s.cfg.Limit),
		},
		GeneratedAt: generatedAt,
		ExpiresAt:   generatedAt.Add(s.cfg.TTL),
	}
	if target.scopeID != "" {
		set.Metadata["scope_id"] = target.scopeID
	}
	if err := s.sets.UpsertRecommendationSet(ctx, set); err != nil {
		if s.metrics != nil {
			s.metrics.RecordRankingSetUpsertError(string(target.scope))
		}
		return fmt.Errorf("%w: save %s trending set: %v", domain.ErrRecommendationStorage, target.scope, err)
	}
	result.SetsGenerated++
	result.ItemsGenerated += len(items)
	if s.metrics != nil {
		s.metrics.RecordRankingItemsGenerated(string(target.scope), len(items))
		if len(items) == 0 {
			s.metrics.RecordRankingEmptySet(string(target.scope))
		}
	}
	s.cacheSet(ctx, target, set)
	s.logger.InfoContext(ctx, "recommendation.rule_based_set_generated",
		slog.String("scope", string(target.scope)),
		slog.String("scope_id", target.scopeID),
		slog.String("strategy_id", string(target.strategyID)),
		slog.String("formula_version", s.cfg.FormulaVersion),
		slog.Int("candidate_count", len(ranked)),
		slog.Int("result_count", len(items)),
		slog.Int64("ttl_seconds", int64(s.cfg.TTL.Seconds())),
	)
	return nil
}

func (s *TrendingService) loadSet(ctx context.Context, target rankingTarget, now time.Time) (domain.RecommendationSet, bool) {
	if s.cache != nil {
		cached, _, err := s.cache.Get(ctx, target.cacheKey)
		if err == nil && cached.ExpiresAt.After(now) && cached.StrategyID == target.strategyID {
			return domain.RecommendationSet{
				ID:          cached.RecommendationID,
				ContextKey:  cached.ContextKey,
				Type:        cached.Type,
				StrategyID:  cached.StrategyID,
				Items:       cached.Items,
				GeneratedAt: cached.CachedAt,
				ExpiresAt:   cached.ExpiresAt,
			}, true
		}
		if err != nil && !errors.Is(err, domain.ErrRecommendationCacheMiss) {
			s.logger.WarnContext(ctx, "recommendation.ranking.cache_read_failed",
				slog.String("cache_key", target.cacheKey),
				slog.String("error", err.Error()),
			)
		}
	}
	set, err := s.sets.GetRecommendationSet(ctx, target.contextKey, target.strategyID, now)
	if err == nil {
		s.cacheSet(ctx, target, set)
		return set, true
	}
	if !errors.Is(err, domain.ErrRecommendationSetNotFound) {
		s.logger.WarnContext(ctx, "recommendation.ranking.set_read_failed",
			slog.String("context_key", target.contextKey),
			slog.String("error", err.Error()),
		)
	}
	return domain.RecommendationSet{}, false
}

func (s *TrendingService) cacheSet(ctx context.Context, target rankingTarget, set domain.RecommendationSet) {
	if s.cache == nil {
		return
	}
	cached := domain.CachedRecommendation{
		RecommendationID: set.ID,
		ContextKey:       set.ContextKey,
		Type:             set.Type,
		StrategyID:       set.StrategyID,
		Items:            set.Items,
		CachedAt:         set.GeneratedAt,
		ExpiresAt:        set.ExpiresAt,
	}
	if err := s.cache.Set(ctx, target.cacheKey, cached, s.cfg.TTL); err != nil {
		if s.metrics != nil {
			s.metrics.RecordRankingCacheSetError(string(target.scope))
		}
		s.logger.WarnContext(ctx, "recommendation.ranking.cache_set_failed",
			slog.String("scope", string(target.scope)),
			slog.String("cache_key", target.cacheKey),
			slog.String("error", err.Error()),
		)
	}
}

func (s *TrendingService) targetFor(scope domain.RankingScope, scopeID string) (rankingTarget, error) {
	switch scope {
	case domain.RankingScopeGlobal:
		key, err := s.keyBuilder.StrategyScopedKey(s.keyBuilder.TrendingGlobalKey(), domain.StrategyTrendingRecentActivity)
		if err != nil {
			return rankingTarget{}, err
		}
		return rankingTarget{
			scope:      scope,
			contextKey: domain.TrendingGlobalContextKey,
			strategyID: domain.StrategyTrendingRecentActivity,
			cacheKey:   key,
			reason:     "global_recent_activity",
		}, nil
	case domain.RankingScopeCategory:
		key, err := s.keyBuilder.TrendingCategoryKey(scopeID)
		if err != nil {
			return rankingTarget{}, err
		}
		key, err = s.keyBuilder.StrategyScopedKey(key, domain.StrategyTrendingFallbackCategory)
		if err != nil {
			return rankingTarget{}, err
		}
		return rankingTarget{
			scope:      scope,
			scopeID:    scopeID,
			contextKey: "category:" + scopeID,
			strategyID: domain.StrategyTrendingFallbackCategory,
			cacheKey:   key,
			reason:     "category_recent_activity",
		}, nil
	case domain.RankingScopeSeller:
		key, err := s.keyBuilder.TrendingSellerKey(scopeID)
		if err != nil {
			return rankingTarget{}, err
		}
		key, err = s.keyBuilder.StrategyScopedKey(key, domain.StrategyTrendingRecentActivity)
		if err != nil {
			return rankingTarget{}, err
		}
		return rankingTarget{
			scope:      scope,
			scopeID:    scopeID,
			contextKey: "seller:" + scopeID,
			strategyID: domain.StrategyTrendingRecentActivity,
			cacheKey:   key,
			reason:     "seller_recent_activity",
		}, nil
	default:
		return rankingTarget{}, fmt.Errorf("%w: invalid ranking scope", domain.ErrInvalidRecommendationRequest)
	}
}

func recommendationItems(products []domain.ScoredProduct, limit int, reason string) []domain.RecommendationSetItem {
	if len(products) < limit {
		limit = len(products)
	}
	items := make([]domain.RecommendationSetItem, 0, limit)
	for i := 0; i < limit; i++ {
		items = append(items, domain.RecommendationSetItem{
			ProductID: products[i].ProductID,
			Score:     products[i].Score,
			Rank:      i + 1,
			Reason:    reason,
		})
	}
	return items
}

func recommendationSetID(target rankingTarget) string {
	if target.scopeID == "" {
		return "reco_trending_global"
	}
	return "reco_trending_" + string(target.scope) + "_" + target.scopeID
}

func trimItems(set domain.RecommendationSet, limit int) domain.RecommendationSet {
	if limit > 0 && len(set.Items) > limit {
		set.Items = append([]domain.RecommendationSetItem(nil), set.Items[:limit]...)
	}
	return set
}

func sortedKeys[T any](items map[string]T) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func groupedCandidateCount(items map[string][]domain.ScoredProduct) int {
	count := 0
	for _, products := range items {
		count += len(products)
	}
	return count
}

func (s *TrendingService) recordRun(result string, startedAt time.Time) {
	if s.metrics == nil {
		return
	}
	s.metrics.RecordRankingRun(result)
	s.metrics.ObserveRankingJobDuration(s.now().Sub(startedAt))
}
