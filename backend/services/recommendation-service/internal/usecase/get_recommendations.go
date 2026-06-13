package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
)

const (
	recommendationSourceRedis  = "redis"
	recommendationSourceMongo  = "mongo"
	recommendationSourceRanker = "ranker"
	recommendationSourceEmpty  = "empty"
)

type GetRecommendationsInput struct {
	RequestID      string
	UserID         string
	AnonymousID    string
	SessionID      string
	ProductID      string
	CategoryID     string
	SellerID       string
	CartProductIDs []string
	Context        domain.RecommendationContext
	Type           domain.RecommendationType
	Limit          int
}

type GetRecommendationsOutput struct {
	Result   domain.RecommendationResult
	CacheTTL time.Duration
	Source   string
	Fallback bool
}

type RecommendationReader interface {
	GetRecommendations(ctx context.Context, input GetRecommendationsInput) (GetRecommendationsOutput, error)
}

type RecommendationDefinitionResolver interface {
	Resolve(ctx context.Context, input ResolveInput) (Resolution, error)
}

type GeneratedSetResolver interface {
	ResolveGeneratedSet(ctx context.Context, request domain.RankRequest) (RankedSetResolution, error)
}

type PersonalizedRecommendationBuilder interface {
	Build(ctx context.Context, input PersonalizedBuildInput) (domain.RecommendationSet, error)
}

type StrategyAssigner interface {
	Assign(ctx context.Context, input StrategyAssignmentInput) (domain.StrategyAssignment, error)
}

type RecommendationServeMetrics interface {
	RecordRecommendationServeSource(source string)
	RecordRecommendationServeFallback(kind string)
	RecordRecommendationServeEmpty(context string)
	RecordRecommendationStrategyServed(strategyID string, context string, source string)
	RecordABAssignmentError(reason string)
}

type GetRecommendationsConfig struct {
	CacheKeyPrefix      string
	MaxIdentifierLength int
	TTLPolicy           domain.CacheTTLPolicy
	ABTestingFailOpen   bool
}

type GetRecommendationsService struct {
	definitions  RecommendationDefinitionResolver
	sets         RecommendationSetRepository
	cache        RankingResultCache
	trending     GeneratedSetResolver
	personalized PersonalizedRecommendationBuilder
	assigner     StrategyAssigner
	metrics      RecommendationServeMetrics
	logger       *slog.Logger
	keyBuilder   domain.CacheKeyBuilder
	ttlPolicy    domain.CacheTTLPolicy
	abFailOpen   bool
	now          func() time.Time
}

type recommendationTarget struct {
	contextKey string
	strategyID domain.StrategyID
	cacheKey   string
	typ        domain.RecommendationType
	anonymous  bool
	scoped     bool
}

func NewGetRecommendationsService(
	definitions RecommendationDefinitionResolver,
	sets RecommendationSetRepository,
	cache RankingResultCache,
	trending GeneratedSetResolver,
	personalized PersonalizedRecommendationBuilder,
	assigner StrategyAssigner,
	cfg GetRecommendationsConfig,
	metrics RecommendationServeMetrics,
	logger *slog.Logger,
) (*GetRecommendationsService, error) {
	if definitions == nil {
		return nil, errors.New("recommendation definition resolver is required")
	}
	if cfg.MaxIdentifierLength <= 0 {
		cfg.MaxIdentifierLength = 128
	}
	if cfg.TTLPolicy.Default <= 0 {
		cfg.TTLPolicy.Default = domain.DefaultRecommendationCacheTTL
	}
	if cfg.TTLPolicy.Personalized <= 0 {
		cfg.TTLPolicy.Personalized = domain.DefaultPersonalizedCacheTTL
	}
	if cfg.TTLPolicy.Guest <= 0 {
		cfg.TTLPolicy.Guest = domain.DefaultGuestCacheTTL
	}
	if cfg.TTLPolicy.RebuildLock <= 0 {
		cfg.TTLPolicy.RebuildLock = time.Minute
	}
	builder, err := domain.NewCacheKeyBuilder(cfg.CacheKeyPrefix, cfg.MaxIdentifierLength)
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &GetRecommendationsService{
		definitions:  definitions,
		sets:         sets,
		cache:        cache,
		trending:     trending,
		personalized: personalized,
		assigner:     assigner,
		metrics:      metrics,
		logger:       logger,
		keyBuilder:   builder,
		ttlPolicy:    cfg.TTLPolicy,
		abFailOpen:   cfg.ABTestingFailOpen,
		now:          func() time.Time { return time.Now().UTC() },
	}, nil
}

func (s *GetRecommendationsService) GetRecommendations(ctx context.Context, input GetRecommendationsInput) (GetRecommendationsOutput, error) {
	if err := ctx.Err(); err != nil {
		return GetRecommendationsOutput{}, err
	}

	resolution, err := s.definitions.Resolve(ctx, ResolveInput{
		RequestID:      input.RequestID,
		UserID:         input.UserID,
		AnonymousID:    input.AnonymousID,
		SessionID:      input.SessionID,
		ProductID:      input.ProductID,
		CategoryID:     input.CategoryID,
		SellerID:       input.SellerID,
		CartProductIDs: input.CartProductIDs,
		Context:        input.Context,
		Limit:          input.Limit,
		Type:           input.Type,
	})
	if err != nil {
		return GetRecommendationsOutput{}, err
	}

	now := s.now()
	req := resolution.Request
	explicitType := input.Type.IsValid()
	strategyID, assignment, err := s.assignedStrategy(ctx, input.RequestID, req, resolution.StrategyID, now)
	if err != nil {
		return GetRecommendationsOutput{}, err
	}
	strategyType, err := domain.RecommendationTypeForStrategy(strategyID)
	if err != nil {
		return GetRecommendationsOutput{}, err
	}
	req.Type = strategyType

	var out GetRecommendationsOutput
	switch req.Type {
	case domain.RecommendationTypePersonalized:
		out, err = s.servePersonalizedStrategy(ctx, input.RequestID, req, strategyID, explicitType, now)
	case domain.RecommendationTypeTrending:
		out, err = s.serveTrendingStrategy(ctx, req, strategyID, now)
	case domain.RecommendationTypeSimilarProducts, domain.RecommendationTypeFrequentlyBoughtTogether:
		out, err = s.serveStoredOrTrendingFallback(ctx, req, strategyID, now)
	default:
		err = fmt.Errorf("%w: %s", domain.ErrUnsupportedRecommendationType, req.Type)
	}
	if err != nil {
		return GetRecommendationsOutput{}, err
	}

	s.recordOutput(ctx, input.RequestID, req, assignment, out)
	return out, nil
}

func (s *GetRecommendationsService) assignedStrategy(
	ctx context.Context,
	requestID string,
	req domain.RecommendationRequest,
	defaultStrategyID domain.StrategyID,
	now time.Time,
) (domain.StrategyID, domain.StrategyAssignment, error) {
	if defaultStrategyID == "" {
		var err error
		defaultStrategyID, err = domain.DefaultStrategyForType(req.Type)
		if err != nil {
			return "", domain.StrategyAssignment{}, err
		}
	}
	assignment := domain.StrategyAssignment{StrategyID: defaultStrategyID}
	if s.assigner == nil {
		return defaultStrategyID, assignment, nil
	}

	selected, err := s.assigner.Assign(ctx, StrategyAssignmentInput{
		RequestID:          requestID,
		Request:            req,
		DefaultStrategyID:  defaultStrategyID,
		AssignmentSnapshot: now,
	})
	if err != nil {
		if s.metrics != nil {
			s.metrics.RecordABAssignmentError("assign")
		}
		s.logger.WarnContext(ctx, "recommendation.ab.assignment_failed",
			slog.String("request_id", requestID),
			slog.String("context", string(req.Context)),
			slog.String("default_strategy_id", string(defaultStrategyID)),
			slog.String("error", err.Error()),
		)
		if !s.abFailOpen {
			return "", domain.StrategyAssignment{}, err
		}
		return defaultStrategyID, assignment, nil
	}
	if selected.StrategyID == "" {
		selected.StrategyID = defaultStrategyID
	}
	return selected.StrategyID, selected, nil
}

func (s *GetRecommendationsService) servePersonalizedStrategy(
	ctx context.Context,
	requestID string,
	req domain.RecommendationRequest,
	strategyID domain.StrategyID,
	explicitType bool,
	now time.Time,
) (GetRecommendationsOutput, error) {
	if strategyID != domain.StrategyPersonalizedBehavior {
		return s.serveStoredOrTrendingFallback(ctx, req, strategyID, now)
	}
	return s.servePersonalized(ctx, requestID, req, explicitType, now)
}

func (s *GetRecommendationsService) serveTrendingStrategy(
	ctx context.Context,
	req domain.RecommendationRequest,
	strategyID domain.StrategyID,
	now time.Time,
) (GetRecommendationsOutput, error) {
	scope, _ := trendingScope(req)
	defaultStrategy := strategyForTrendingScope(scope)
	if strategyID == "" || strategyID == defaultStrategy || strategyID == domain.StrategyTrendingRecentActivity || strategyID == domain.StrategyTrendingFallbackCategory {
		return s.serveTrending(ctx, req, now)
	}
	return s.serveStored(ctx, req, strategyID, now)
}

func (s *GetRecommendationsService) servePersonalized(
	ctx context.Context,
	requestID string,
	req domain.RecommendationRequest,
	explicitType bool,
	now time.Time,
) (GetRecommendationsOutput, error) {
	if s.personalized != nil {
		set, err := s.personalized.Build(ctx, PersonalizedBuildInput{
			UserID:      req.UserID,
			AnonymousID: req.AnonymousID,
			CategoryID:  req.CategoryID,
			Context:     req.Context,
			Limit:       req.Limit,
			Now:         now,
		})
		if err == nil {
			return outputFromSet(set, ttlUntil(set.ExpiresAt, now), recommendationSourceRanker, false, now), nil
		}
		if !explicitType && errors.Is(err, domain.ErrUnsupportedContext) {
			s.logger.InfoContext(ctx, "recommendation.personalized_context_fallback",
				slog.String("request_id", requestID),
				slog.String("context", string(req.Context)),
				slog.String("fallback_type", string(domain.RecommendationTypeTrending)),
			)
			out, fallbackErr := s.serveTrending(ctx, asTrendingRequest(req), now)
			out.Fallback = true
			return out, fallbackErr
		}
		return GetRecommendationsOutput{}, err
	}

	out, err := s.serveStored(ctx, req, domain.StrategyPersonalizedBehavior, now)
	if err != nil {
		return GetRecommendationsOutput{}, err
	}
	if len(out.Result.Items) > 0 || explicitType {
		return out, nil
	}

	fallback, err := s.serveTrending(ctx, asTrendingRequest(req), now)
	fallback.Fallback = true
	return fallback, err
}

func (s *GetRecommendationsService) serveTrending(ctx context.Context, req domain.RecommendationRequest, now time.Time) (GetRecommendationsOutput, error) {
	scope, scopeID := trendingScope(req)
	if s.trending != nil {
		resolution, err := s.trending.ResolveGeneratedSet(ctx, domain.RankRequest{
			Scope:   scope,
			ScopeID: scopeID,
			Limit:   req.Limit,
			Now:     now,
		})
		if err != nil {
			return GetRecommendationsOutput{}, err
		}
		if resolution.Empty && strings.TrimSpace(resolution.Set.ID) == "" {
			out := emptyOutput(req, strategyForTrendingScope(scope), now)
			out.Fallback = resolution.Fallback
			return out, nil
		}
		source := recommendationSourceMongo
		if resolution.Empty {
			source = recommendationSourceEmpty
		}
		return outputFromSet(resolution.Set, ttlUntil(resolution.Set.ExpiresAt, now), source, resolution.Fallback, now), nil
	}

	out, err := s.serveStored(ctx, req, strategyForTrendingScope(scope), now)
	if err != nil {
		return GetRecommendationsOutput{}, err
	}
	if len(out.Result.Items) > 0 || !targetIsScoped(req) {
		return out, nil
	}

	global := req
	global.CategoryID = ""
	global.SellerID = ""
	fallback, err := s.serveStored(ctx, global, domain.StrategyTrendingRecentActivity, now)
	fallback.Fallback = true
	return fallback, err
}

func (s *GetRecommendationsService) serveStoredOrTrendingFallback(
	ctx context.Context,
	req domain.RecommendationRequest,
	strategyID domain.StrategyID,
	now time.Time,
) (GetRecommendationsOutput, error) {
	out, err := s.serveStored(ctx, req, strategyID, now)
	if err != nil {
		return GetRecommendationsOutput{}, err
	}
	if len(out.Result.Items) > 0 {
		return out, nil
	}
	fallback, err := s.serveTrending(ctx, asTrendingRequest(req), now)
	fallback.Fallback = true
	return fallback, err
}

func (s *GetRecommendationsService) serveStored(
	ctx context.Context,
	req domain.RecommendationRequest,
	strategyID domain.StrategyID,
	now time.Time,
) (GetRecommendationsOutput, error) {
	target, err := s.targetForRequest(req, strategyID)
	if err != nil {
		return GetRecommendationsOutput{}, err
	}

	var cacheErr error
	if s.cache != nil && target.cacheKey != "" {
		cached, ttl, err := s.cache.Get(ctx, target.cacheKey)
		if err == nil && cached.ExpiresAt.After(now) && cached.StrategyID == target.strategyID {
			return outputFromSet(setFromCached(cached), ttl, recommendationSourceRedis, false, now), nil
		}
		if err != nil && !errors.Is(err, domain.ErrRecommendationCacheMiss) && !errors.Is(err, domain.ErrRecommendationCacheCorrupt) {
			cacheErr = err
			s.logger.WarnContext(ctx, "recommendation.serve.cache_read_failed",
				slog.String("cache_key", target.cacheKey),
				slog.String("error", err.Error()),
			)
		}
	}

	if s.sets != nil {
		set, err := s.sets.GetRecommendationSet(ctx, target.contextKey, target.strategyID, now)
		if err == nil {
			s.cacheSetBestEffort(ctx, target, set, now)
			return outputFromSet(set, ttlUntil(set.ExpiresAt, now), recommendationSourceMongo, false, now), nil
		}
		if !errors.Is(err, domain.ErrRecommendationSetNotFound) {
			return GetRecommendationsOutput{}, err
		}
	}

	if cacheErr != nil && s.sets == nil {
		return GetRecommendationsOutput{}, cacheErr
	}
	return emptyOutput(req, target.strategyID, now), nil
}

func (s *GetRecommendationsService) targetForRequest(req domain.RecommendationRequest, strategyID domain.StrategyID) (recommendationTarget, error) {
	if strategyID == "" {
		var err error
		strategyID, err = domain.DefaultStrategyForType(req.Type)
		if err != nil {
			return recommendationTarget{}, err
		}
	}

	switch req.Type {
	case domain.RecommendationTypeTrending:
		scope, scopeID := trendingScope(req)
		return s.trendingTarget(scope, scopeID, strategyID)
	case domain.RecommendationTypePersonalized:
		if strings.TrimSpace(req.UserID) != "" {
			key, err := s.keyBuilder.PersonalizedUserKey(req.UserID)
			if err != nil {
				return recommendationTarget{}, err
			}
			key, err = s.keyBuilder.StrategyScopedKey(key, strategyID)
			if err != nil {
				return recommendationTarget{}, err
			}
			return recommendationTarget{
				contextKey: "home:user:" + strings.TrimSpace(req.UserID),
				strategyID: strategyID,
				cacheKey:   key,
				typ:        req.Type,
			}, nil
		}
		key, err := s.keyBuilder.PersonalizedAnonymousKey(req.AnonymousID)
		if err != nil {
			return recommendationTarget{}, err
		}
		key, err = s.keyBuilder.StrategyScopedKey(key, strategyID)
		if err != nil {
			return recommendationTarget{}, err
		}
		return recommendationTarget{
			contextKey: "home:anon:" + strings.TrimSpace(req.AnonymousID),
			strategyID: strategyID,
			cacheKey:   key,
			typ:        req.Type,
			anonymous:  true,
		}, nil
	case domain.RecommendationTypeSimilarProducts:
		productID := strings.TrimSpace(req.ProductID)
		key, err := s.keyBuilder.SimilarProductKey(productID)
		if err != nil {
			return recommendationTarget{}, err
		}
		key, err = s.keyBuilder.StrategyScopedKey(key, strategyID)
		if err != nil {
			return recommendationTarget{}, err
		}
		return recommendationTarget{
			contextKey: "similar:product:" + productID,
			strategyID: strategyID,
			cacheKey:   key,
			typ:        req.Type,
		}, nil
	case domain.RecommendationTypeFrequentlyBoughtTogether:
		productID := anchorProductID(req)
		key, err := s.keyBuilder.FrequentlyBoughtTogetherProductKey(productID)
		if err != nil {
			return recommendationTarget{}, err
		}
		key, err = s.keyBuilder.StrategyScopedKey(key, strategyID)
		if err != nil {
			return recommendationTarget{}, err
		}
		return recommendationTarget{
			contextKey: "fbt:product:" + productID,
			strategyID: strategyID,
			cacheKey:   key,
			typ:        req.Type,
		}, nil
	default:
		return recommendationTarget{}, fmt.Errorf("%w: %s", domain.ErrUnsupportedRecommendationType, req.Type)
	}
}

func (s *GetRecommendationsService) trendingTarget(scope domain.RankingScope, scopeID string, strategyID domain.StrategyID) (recommendationTarget, error) {
	switch scope {
	case domain.RankingScopeGlobal:
		if strategyID == "" {
			strategyID = domain.StrategyTrendingRecentActivity
		}
		key, err := s.keyBuilder.StrategyScopedKey(s.keyBuilder.TrendingGlobalKey(), strategyID)
		if err != nil {
			return recommendationTarget{}, err
		}
		return recommendationTarget{
			contextKey: domain.TrendingGlobalContextKey,
			strategyID: strategyID,
			cacheKey:   key,
			typ:        domain.RecommendationTypeTrending,
		}, nil
	case domain.RankingScopeCategory:
		if strategyID == "" {
			strategyID = domain.StrategyTrendingFallbackCategory
		}
		key, err := s.keyBuilder.TrendingCategoryKey(scopeID)
		if err != nil {
			return recommendationTarget{}, err
		}
		key, err = s.keyBuilder.StrategyScopedKey(key, strategyID)
		if err != nil {
			return recommendationTarget{}, err
		}
		return recommendationTarget{
			contextKey: "category:" + strings.TrimSpace(scopeID),
			strategyID: strategyID,
			cacheKey:   key,
			typ:        domain.RecommendationTypeTrending,
			scoped:     true,
		}, nil
	case domain.RankingScopeSeller:
		if strategyID == "" {
			strategyID = domain.StrategyTrendingRecentActivity
		}
		key, err := s.keyBuilder.TrendingSellerKey(scopeID)
		if err != nil {
			return recommendationTarget{}, err
		}
		key, err = s.keyBuilder.StrategyScopedKey(key, strategyID)
		if err != nil {
			return recommendationTarget{}, err
		}
		return recommendationTarget{
			contextKey: "seller:" + strings.TrimSpace(scopeID),
			strategyID: strategyID,
			cacheKey:   key,
			typ:        domain.RecommendationTypeTrending,
			scoped:     true,
		}, nil
	default:
		return recommendationTarget{}, fmt.Errorf("%w: invalid ranking scope", domain.ErrInvalidRecommendationRequest)
	}
}

func (s *GetRecommendationsService) cacheSetBestEffort(ctx context.Context, target recommendationTarget, set domain.RecommendationSet, now time.Time) {
	if s.cache == nil || target.cacheKey == "" || len(set.Items) == 0 {
		return
	}
	ttl := ttlUntil(set.ExpiresAt, now)
	if ttl <= 0 {
		ttl = s.ttlPolicy.TTLFor(set.Type, target.anonymous)
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
	if err := s.cache.Set(ctx, target.cacheKey, cached, ttl); err != nil {
		s.logger.WarnContext(ctx, "recommendation.serve.cache_set_failed",
			slog.String("cache_key", target.cacheKey),
			slog.String("error", err.Error()),
		)
	}
}

func (s *GetRecommendationsService) recordOutput(ctx context.Context, requestID string, req domain.RecommendationRequest, assignment domain.StrategyAssignment, out GetRecommendationsOutput) {
	source := out.Source
	if source == "" {
		source = recommendationSourceEmpty
	}
	if s.metrics != nil {
		s.metrics.RecordRecommendationServeSource(source)
		if out.Fallback {
			s.metrics.RecordRecommendationServeFallback(string(out.Result.Type))
		}
		if len(out.Result.Items) == 0 {
			s.metrics.RecordRecommendationServeEmpty(string(req.Context))
		}
		s.metrics.RecordRecommendationStrategyServed(string(out.Result.StrategyID), string(req.Context), source)
	}
	s.logger.InfoContext(ctx, "recommendation.served",
		slog.String("request_id", requestID),
		slog.String("context", string(req.Context)),
		slog.String("recommendation_type", string(out.Result.Type)),
		slog.String("strategy_id", string(out.Result.StrategyID)),
		slog.String("experiment_id", assignment.ExperimentID),
		slog.String("variant_id", assignment.VariantID),
		slog.Bool("ab_assigned", assignment.Assigned),
		slog.String("source", source),
		slog.Bool("fallback", out.Fallback),
		slog.Int("item_count", len(out.Result.Items)),
		slog.Int64("cache_ttl_seconds", int64(out.CacheTTL.Seconds())),
	)
}

func outputFromSet(set domain.RecommendationSet, ttl time.Duration, source string, fallback bool, now time.Time) GetRecommendationsOutput {
	if set.GeneratedAt.IsZero() {
		set.GeneratedAt = now
	}
	items := make([]domain.RecommendationItem, 0, len(set.Items))
	for _, item := range set.Items {
		productID := strings.TrimSpace(item.ProductID)
		if productID == "" {
			continue
		}
		items = append(items, domain.RecommendationItem{
			ProductID: productID,
			Score:     item.Score,
			Reason:    item.Reason,
		})
	}
	return GetRecommendationsOutput{
		Result: domain.RecommendationResult{
			RecommendationID: set.ID,
			Type:             set.Type,
			StrategyID:       set.StrategyID,
			Items:            items,
			GeneratedAt:      set.GeneratedAt,
		},
		CacheTTL: nonNegativeTTL(ttl),
		Source:   source,
		Fallback: fallback,
	}
}

func emptyOutput(req domain.RecommendationRequest, strategyID domain.StrategyID, now time.Time) GetRecommendationsOutput {
	if strategyID == "" {
		strategyID, _ = domain.DefaultStrategyForType(req.Type)
	}
	return GetRecommendationsOutput{
		Result: domain.RecommendationResult{
			RecommendationID: emptyRecommendationID(req),
			Type:             req.Type,
			StrategyID:       strategyID,
			GeneratedAt:      now,
		},
		Source: recommendationSourceEmpty,
	}
}

func setFromCached(cached domain.CachedRecommendation) domain.RecommendationSet {
	return domain.RecommendationSet{
		ID:          cached.RecommendationID,
		ContextKey:  cached.ContextKey,
		Type:        cached.Type,
		StrategyID:  cached.StrategyID,
		Items:       cached.Items,
		GeneratedAt: cached.CachedAt,
		ExpiresAt:   cached.ExpiresAt,
	}
}

func asTrendingRequest(req domain.RecommendationRequest) domain.RecommendationRequest {
	req.Type = domain.RecommendationTypeTrending
	req.UserID = ""
	req.AnonymousID = ""
	req.ProductID = ""
	req.CartProductIDs = nil
	return req
}

func trendingScope(req domain.RecommendationRequest) (domain.RankingScope, string) {
	if sellerID := strings.TrimSpace(req.SellerID); sellerID != "" {
		return domain.RankingScopeSeller, sellerID
	}
	if categoryID := strings.TrimSpace(req.CategoryID); categoryID != "" {
		return domain.RankingScopeCategory, categoryID
	}
	return domain.RankingScopeGlobal, ""
}

func strategyForTrendingScope(scope domain.RankingScope) domain.StrategyID {
	if scope == domain.RankingScopeCategory {
		return domain.StrategyTrendingFallbackCategory
	}
	return domain.StrategyTrendingRecentActivity
}

func targetIsScoped(req domain.RecommendationRequest) bool {
	scope, _ := trendingScope(req)
	return scope != domain.RankingScopeGlobal
}

func anchorProductID(req domain.RecommendationRequest) string {
	if productID := strings.TrimSpace(req.ProductID); productID != "" {
		return productID
	}
	for _, productID := range req.CartProductIDs {
		if productID = strings.TrimSpace(productID); productID != "" {
			return productID
		}
	}
	return ""
}

func ttlUntil(expiresAt time.Time, now time.Time) time.Duration {
	if expiresAt.IsZero() {
		return 0
	}
	ttl := expiresAt.Sub(now)
	if ttl < 0 {
		return 0
	}
	return ttl
}

func nonNegativeTTL(ttl time.Duration) time.Duration {
	if ttl < 0 {
		return 0
	}
	return ttl
}

func emptyRecommendationID(req domain.RecommendationRequest) string {
	parts := []string{"reco", "empty", string(req.Type), string(req.Context)}
	if value := strings.TrimSpace(req.CategoryID); value != "" {
		parts = append(parts, "category", value)
	}
	if value := strings.TrimSpace(req.SellerID); value != "" {
		parts = append(parts, "seller", value)
	}
	if value := anchorProductID(req); value != "" {
		parts = append(parts, "product", value)
	}
	id := strings.Join(parts, "_")
	id = strings.NewReplacer(":", "_", "/", "_", " ", "_", "\t", "_", "\n", "_", "\r", "_").Replace(id)
	return id
}
