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

type PersonalizedServiceConfig struct {
	FormulaVersion          string
	Limit                   int
	MaxLimit                int
	MaxIdentifierLength     int
	CacheKeyPrefix          string
	TTL                     time.Duration
	GuestTTL                time.Duration
	MinPositiveInteractions int
	ProfileMaxAge           time.Duration
	MaxCandidates           int
	MaxItemsPerSeller       int
	TopCategories           int
	TopSellers              int
	TopBrands               int
	DirectProductLimit      int
	Weights                 domain.PersonalizationWeights
}

type PersonalizedBuildInput struct {
	UserID      string
	AnonymousID string
	CategoryID  string
	Context     domain.RecommendationContext
	Limit       int
	Now         time.Time
}

type ProfileIdentity struct {
	ProfileKey  string
	UserID      string
	AnonymousID string
	IsGuest     bool
}

type PersonalizedService struct {
	features   PersonalizedFeatureReader
	sets       RecommendationSetRepository
	cache      RankingResultCache
	metrics    PersonalizedMetrics
	logger     *slog.Logger
	cfg        PersonalizedServiceConfig
	keyBuilder domain.CacheKeyBuilder
	now        func() time.Time
}

type personalizedFallbackTarget struct {
	source     string
	reason     string
	contextKey string
	strategyID domain.StrategyID
}

type personalizedFallbackResult struct {
	items  []domain.RecommendationSetItem
	source string
	count  int
}

func NewPersonalizedService(
	features PersonalizedFeatureReader,
	sets RecommendationSetRepository,
	cache RankingResultCache,
	cfg PersonalizedServiceConfig,
	metrics PersonalizedMetrics,
	logger *slog.Logger,
) (*PersonalizedService, error) {
	if features == nil {
		return nil, errors.New("personalized feature reader is required")
	}
	if sets == nil {
		return nil, errors.New("recommendation set repository is required")
	}
	cfg.FormulaVersion = strings.TrimSpace(cfg.FormulaVersion)
	if cfg.FormulaVersion == "" {
		cfg.FormulaVersion = domain.PersonalizationFormulaVersion
	}
	if cfg.Weights == (domain.PersonalizationWeights{}) {
		cfg.Weights = domain.DefaultPersonalizationWeights()
	}
	if err := cfg.Weights.Validate(); err != nil {
		return nil, err
	}
	if cfg.Weights != domain.DefaultPersonalizationWeights() && cfg.FormulaVersion == domain.PersonalizationFormulaVersion {
		return nil, errors.New("personalization formula version must change when behavior_v1 weights differ")
	}
	if cfg.Limit <= 0 {
		cfg.Limit = 12
	}
	if cfg.MaxLimit <= 0 {
		cfg.MaxLimit = 100
	}
	if cfg.Limit > cfg.MaxLimit {
		return nil, errors.New("personalized default limit must be less than or equal to max limit")
	}
	if cfg.MaxIdentifierLength <= 0 {
		cfg.MaxIdentifierLength = 128
	}
	if cfg.TTL <= 0 {
		cfg.TTL = domain.DefaultPersonalizedCacheTTL
	}
	if cfg.GuestTTL <= 0 {
		cfg.GuestTTL = domain.DefaultGuestCacheTTL
	}
	if cfg.MinPositiveInteractions < 0 {
		return nil, errors.New("personalized minimum positive interactions cannot be negative")
	}
	if cfg.MinPositiveInteractions == 0 {
		cfg.MinPositiveInteractions = domain.DefaultPersonalizedMinInteractions
	}
	if cfg.ProfileMaxAge <= 0 {
		cfg.ProfileMaxAge = domain.DefaultPersonalizedProfileMaxAge
	}
	if cfg.MaxCandidates <= 0 {
		cfg.MaxCandidates = domain.DefaultPersonalizedMaxCandidates
	}
	if cfg.MaxCandidates < cfg.Limit {
		return nil, errors.New("personalized max candidates must be greater than or equal to the default limit")
	}
	if cfg.MaxItemsPerSeller <= 0 {
		cfg.MaxItemsPerSeller = domain.DefaultPersonalizedMaxItemsPerSeller
	}
	if cfg.TopCategories < 0 || cfg.TopSellers < 0 || cfg.TopBrands < 0 || cfg.DirectProductLimit < 0 {
		return nil, errors.New("personalized candidate source limits cannot be negative")
	}
	if cfg.TopCategories == 0 {
		cfg.TopCategories = domain.DefaultPersonalizedTopCategories
	}
	if cfg.TopSellers == 0 {
		cfg.TopSellers = domain.DefaultPersonalizedTopSellers
	}
	if cfg.TopBrands == 0 {
		cfg.TopBrands = domain.DefaultPersonalizedTopBrands
	}
	if cfg.DirectProductLimit == 0 {
		cfg.DirectProductLimit = domain.DefaultPersonalizedDirectProductLimit
	}

	builder, err := domain.NewCacheKeyBuilder(cfg.CacheKeyPrefix, cfg.MaxIdentifierLength)
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &PersonalizedService{
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

func ResolveProfileIdentity(userID string, anonymousID string) (ProfileIdentity, bool) {
	userID = strings.TrimSpace(userID)
	anonymousID = strings.TrimSpace(anonymousID)
	if userID != "" {
		return ProfileIdentity{ProfileKey: "user:" + userID, UserID: userID}, true
	}
	if anonymousID != "" {
		return ProfileIdentity{ProfileKey: "anon:" + anonymousID, AnonymousID: anonymousID, IsGuest: true}, true
	}
	return ProfileIdentity{}, false
}

func (s *PersonalizedService) Build(ctx context.Context, input PersonalizedBuildInput) (domain.RecommendationSet, error) {
	if err := ctx.Err(); err != nil {
		return domain.RecommendationSet{}, err
	}
	startedAt := s.now()
	input, err := s.normalizeInput(input)
	if err != nil {
		s.recordBuild("failed", startedAt)
		return domain.RecommendationSet{}, err
	}
	now := input.Now
	if now.IsZero() {
		now = startedAt
	}

	identity, hasIdentity := ResolveProfileIdentity(input.UserID, input.AnonymousID)
	if !hasIdentity {
		s.recordProfileMiss("identity_missing")
		result := s.buildUnidentifiedColdStart(ctx, input, now)
		s.recordBuild("cold_start", startedAt)
		return result, nil
	}

	if cached, ok := s.loadCached(ctx, identity, input.Limit, now); ok {
		if s.metrics != nil {
			s.metrics.RecordPersonalizedCacheHit(identityType(identity))
		}
		s.recordBuild("cache_hit", startedAt)
		return cached, nil
	}

	profile, err := s.features.GetUserFeatureProfile(ctx, identity.ProfileKey)
	if err != nil {
		if errors.Is(err, domain.ErrFeatureProfileNotFound) {
			s.recordProfileMiss("profile_missing")
		} else {
			s.logger.WarnContext(ctx, "recommendation.personalized.profile_read_failed", slog.String("error", err.Error()))
			s.recordProfileMiss("profile_read_error")
		}
		return s.persistColdStart(ctx, input, identity, now, "profile_unavailable", startedAt)
	}

	counters, err := s.features.ListUserProductCounters(ctx, identity.ProfileKey, s.cfg.DirectProductLimit)
	if err != nil {
		s.logger.WarnContext(ctx, "recommendation.personalized.affinity_read_failed", slog.String("error", err.Error()))
		s.recordProfileMiss("affinity_read_error")
		return s.persistColdStart(ctx, input, identity, now, "affinity_unavailable", startedAt)
	}
	if !s.usableProfile(profile, counters, now) {
		s.recordProfileMiss("profile_weak_or_stale")
		return s.persistColdStart(ctx, input, identity, now, "profile_weak_or_stale", startedAt)
	}

	query := s.candidateQuery(profile, counters)
	if query.Empty() {
		s.recordProfileMiss("candidate_query_empty")
		return s.persistColdStart(ctx, input, identity, now, "candidate_query_empty", startedAt)
	}
	candidates, err := s.features.ListPersonalizedCandidates(ctx, query)
	if err != nil {
		s.logger.WarnContext(ctx, "recommendation.personalized.candidate_read_failed", slog.String("error", err.Error()))
		return s.persistColdStart(ctx, input, identity, now, "candidate_read_error", startedAt)
	}
	if s.metrics != nil {
		s.metrics.ObservePersonalizedCandidates(len(candidates))
	}

	ranked := domain.ScoreAndSortPersonalizedProducts(profile, counters, candidates, s.cfg.Weights)
	selected := s.applyDiversity(ranked, input.Limit)
	items := itemsFromPersonalizedProducts(selected)
	personalizedCount := len(items)
	fallback := s.backfill(ctx, input, now, items, input.Limit)
	items = fallback.items

	set := s.newPersonalizedSet(identity, items, now, map[string]string{
		"fallback_used":           strconv.FormatBool(fallback.count > 0),
		"fallback_source":         fallback.source,
		"personalized_item_count": strconv.Itoa(personalizedCount),
		"backfill_item_count":     strconv.Itoa(fallback.count),
		"candidate_count":         strconv.Itoa(len(candidates)),
	})
	if err := s.persistAndCache(ctx, identity, set); err != nil {
		s.recordBuild("failed", startedAt)
		return domain.RecommendationSet{}, err
	}

	if fallback.count > 0 && s.metrics != nil {
		s.metrics.RecordPersonalizedBackfill(fallback.source, fallback.count)
	}
	if len(items) == 0 && s.metrics != nil {
		s.metrics.RecordPersonalizedEmptyResult()
	}
	s.logger.InfoContext(ctx, "recommendation.personalized_set_generated",
		slog.String("identity_type", identityType(identity)),
		slog.String("strategy_id", string(domain.StrategyPersonalizedBehavior)),
		slog.String("formula_version", s.cfg.FormulaVersion),
		slog.Int("candidate_count", len(candidates)),
		slog.Int("personalized_item_count", personalizedCount),
		slog.Int("backfill_item_count", fallback.count),
		slog.String("fallback_source", fallback.source),
		slog.Int64("ttl_seconds", int64(s.ttlFor(identity).Seconds())),
		slog.Int64("duration_ms", s.now().Sub(startedAt).Milliseconds()),
	)
	s.recordBuild("success", startedAt)
	return set, nil
}

func (s *PersonalizedService) normalizeInput(input PersonalizedBuildInput) (PersonalizedBuildInput, error) {
	input.UserID = strings.TrimSpace(input.UserID)
	input.AnonymousID = strings.TrimSpace(input.AnonymousID)
	input.CategoryID = strings.TrimSpace(input.CategoryID)
	input.Context = domain.RecommendationContext(strings.TrimSpace(string(input.Context)))
	if input.Context == "" {
		input.Context = domain.ContextHomeFeed
	}
	if input.Context != domain.ContextHomeFeed {
		return PersonalizedBuildInput{}, fmt.Errorf("%w: personalized_v1_behavior currently supports home_feed context", domain.ErrUnsupportedContext)
	}
	if input.Limit < 0 {
		return PersonalizedBuildInput{}, fmt.Errorf("%w: limit cannot be negative", domain.ErrInvalidLimit)
	}
	if input.Limit == 0 {
		input.Limit = s.cfg.Limit
	}
	if input.Limit > s.cfg.MaxLimit {
		return PersonalizedBuildInput{}, fmt.Errorf("%w: limit cannot exceed %d", domain.ErrInvalidLimit, s.cfg.MaxLimit)
	}
	for field, value := range map[string]string{
		"user_id":      input.UserID,
		"anonymous_id": input.AnonymousID,
		"category_id":  input.CategoryID,
	} {
		if len(value) > s.cfg.MaxIdentifierLength {
			return PersonalizedBuildInput{}, fmt.Errorf("%w: %s is too long", domain.ErrInvalidRecommendationRequest, field)
		}
	}
	if !input.Now.IsZero() {
		input.Now = input.Now.UTC()
	}
	return input, nil
}

func (s *PersonalizedService) usableProfile(profile domain.UserFeatureProfile, counters []domain.UserProductCounter, now time.Time) bool {
	if strings.TrimSpace(profile.ProfileKey) == "" || profile.LastEventAt.IsZero() {
		return false
	}
	if now.Sub(profile.LastEventAt.UTC()) > s.cfg.ProfileMaxAge {
		return false
	}
	if domain.HasPositiveProfileAffinity(profile) {
		return true
	}
	return domain.PositiveUserProductInteractions(counters) >= s.cfg.MinPositiveInteractions
}

func (s *PersonalizedService) candidateQuery(profile domain.UserFeatureProfile, counters []domain.UserProductCounter) domain.PersonalizedCandidateQuery {
	productIDs := make([]string, 0, s.cfg.DirectProductLimit)
	for _, product := range profile.RecentProducts {
		productID := strings.TrimSpace(product.ProductID)
		if productID != "" {
			productIDs = append(productIDs, productID)
		}
		if len(productIDs) >= s.cfg.DirectProductLimit {
			break
		}
	}
	for _, counter := range counters {
		productID := strings.TrimSpace(counter.ProductID)
		if productID != "" {
			productIDs = append(productIDs, productID)
		}
	}
	return domain.PersonalizedCandidateQuery{
		CategoryIDs: topScoreKeys(profile.CategoryScores, s.cfg.TopCategories),
		SellerIDs:   topScoreKeys(profile.SellerScores, s.cfg.TopSellers),
		BrandIDs:    topScoreKeys(profile.BrandScores, s.cfg.TopBrands),
		ProductIDs:  productIDs,
		Limit:       s.cfg.MaxCandidates,
	}.Normalize()
}

func (s *PersonalizedService) applyDiversity(ranked []domain.ScoredPersonalizedProduct, limit int) []domain.ScoredPersonalizedProduct {
	selected := make([]domain.ScoredPersonalizedProduct, 0, limit)
	seen := make(map[string]struct{}, limit)
	sellerCounts := make(map[string]int)
	for _, item := range ranked {
		if _, ok := seen[item.ProductID]; ok {
			continue
		}
		if item.SellerID != "" && sellerCounts[item.SellerID] >= s.cfg.MaxItemsPerSeller {
			continue
		}
		seen[item.ProductID] = struct{}{}
		if item.SellerID != "" {
			sellerCounts[item.SellerID]++
		}
		selected = append(selected, item)
		if len(selected) == limit {
			break
		}
	}
	return selected
}

func (s *PersonalizedService) persistColdStart(
	ctx context.Context,
	input PersonalizedBuildInput,
	identity ProfileIdentity,
	now time.Time,
	reason string,
	startedAt time.Time,
) (domain.RecommendationSet, error) {
	fallback := s.backfill(ctx, input, now, nil, input.Limit)
	set := s.newPersonalizedSet(identity, fallback.items, now, map[string]string{
		"fallback_used":           strconv.FormatBool(fallback.count > 0),
		"fallback_source":         fallback.source,
		"personalized_item_count": "0",
		"backfill_item_count":     strconv.Itoa(fallback.count),
		"cold_start_reason":       reason,
	})
	if err := s.persistAndCache(ctx, identity, set); err != nil {
		s.recordBuild("failed", startedAt)
		return domain.RecommendationSet{}, err
	}
	if fallback.count > 0 && s.metrics != nil {
		s.metrics.RecordPersonalizedFallback(fallback.source)
	}
	if len(set.Items) == 0 && s.metrics != nil {
		s.metrics.RecordPersonalizedEmptyResult()
	}
	s.recordBuild("cold_start", startedAt)
	return set, nil
}

func (s *PersonalizedService) buildUnidentifiedColdStart(ctx context.Context, input PersonalizedBuildInput, now time.Time) domain.RecommendationSet {
	fallback := s.backfill(ctx, input, now, nil, input.Limit)
	items := rerankItems(fallback.items)
	if fallback.count > 0 && s.metrics != nil {
		s.metrics.RecordPersonalizedFallback(fallback.source)
	}
	if len(items) == 0 && s.metrics != nil {
		s.metrics.RecordPersonalizedEmptyResult()
	}
	return domain.RecommendationSet{
		ID:         "reco_personalized_cold_start",
		ContextKey: "home:anonymous",
		Type:       domain.RecommendationTypePersonalized,
		StrategyID: domain.StrategyPersonalizedBehavior,
		Items:      items,
		Metadata: map[string]string{
			"formula_version":         s.cfg.FormulaVersion,
			"fallback_used":           strconv.FormatBool(fallback.count > 0),
			"fallback_source":         fallback.source,
			"personalized_item_count": "0",
			"backfill_item_count":     strconv.Itoa(fallback.count),
			"cold_start_reason":       "identity_missing",
		},
		GeneratedAt: now,
		ExpiresAt:   now.Add(s.cfg.TTL),
	}
}

func (s *PersonalizedService) backfill(
	ctx context.Context,
	input PersonalizedBuildInput,
	now time.Time,
	existing []domain.RecommendationSetItem,
	limit int,
) personalizedFallbackResult {
	items := append([]domain.RecommendationSetItem(nil), existing...)
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		seen[item.ProductID] = struct{}{}
	}
	result := personalizedFallbackResult{items: items}
	if len(items) >= limit {
		result.items = rerankItems(items)
		return result
	}

	for _, target := range fallbackTargets(input.CategoryID) {
		set, err := s.sets.GetRecommendationSet(ctx, target.contextKey, target.strategyID, now)
		if err != nil {
			if !errors.Is(err, domain.ErrRecommendationSetNotFound) {
				s.logger.WarnContext(ctx, "recommendation.personalized.fallback_read_failed",
					slog.String("source", target.source),
					slog.String("error", err.Error()),
				)
			}
			continue
		}
		before := len(items)
		for _, item := range set.Items {
			productID := strings.TrimSpace(item.ProductID)
			if productID == "" {
				continue
			}
			if _, ok := seen[productID]; ok {
				continue
			}
			seen[productID] = struct{}{}
			items = append(items, domain.RecommendationSetItem{
				ProductID: productID,
				Score:     item.Score,
				Reason:    target.reason,
			})
			if len(items) == limit {
				break
			}
		}
		if added := len(items) - before; added > 0 {
			if result.source == "" {
				result.source = target.source
			}
			result.count += added
		}
		if len(items) == limit {
			break
		}
	}
	result.items = rerankItems(items)
	return result
}

func (s *PersonalizedService) newPersonalizedSet(identity ProfileIdentity, items []domain.RecommendationSetItem, now time.Time, metadata map[string]string) domain.RecommendationSet {
	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["formula_version"] = s.cfg.FormulaVersion
	metadata["profile_key"] = identity.ProfileKey
	metadata["identity_type"] = identityType(identity)
	metadata["limit"] = strconv.Itoa(len(items))

	return domain.RecommendationSet{
		ID:          "reco_personalized_" + strings.NewReplacer(":", "_").Replace(identity.ProfileKey),
		ContextKey:  "home:" + identity.ProfileKey,
		Type:        domain.RecommendationTypePersonalized,
		StrategyID:  domain.StrategyPersonalizedBehavior,
		Items:       rerankItems(items),
		Metadata:    metadata,
		GeneratedAt: now,
		ExpiresAt:   now.Add(s.ttlFor(identity)),
	}
}

func (s *PersonalizedService) persistAndCache(ctx context.Context, identity ProfileIdentity, set domain.RecommendationSet) error {
	if err := s.sets.UpsertRecommendationSet(ctx, set); err != nil {
		if s.metrics != nil {
			s.metrics.RecordPersonalizedSetUpsertError()
		}
		return fmt.Errorf("%w: save personalized set: %v", domain.ErrRecommendationStorage, err)
	}
	if s.metrics != nil {
		s.metrics.ObservePersonalizedResultItems(len(set.Items))
	}
	s.cacheSet(ctx, identity, set)
	return nil
}

func (s *PersonalizedService) loadCached(ctx context.Context, identity ProfileIdentity, limit int, now time.Time) (domain.RecommendationSet, bool) {
	if s.cache == nil {
		return domain.RecommendationSet{}, false
	}
	key, err := s.cacheKey(identity)
	if err != nil {
		if s.metrics != nil {
			s.metrics.RecordPersonalizedCacheError("key")
		}
		s.logger.WarnContext(ctx, "recommendation.personalized.cache_key_failed", slog.String("error", err.Error()))
		return domain.RecommendationSet{}, false
	}
	cached, _, err := s.cache.Get(ctx, key)
	if err != nil {
		if !errors.Is(err, domain.ErrRecommendationCacheMiss) {
			if s.metrics != nil {
				s.metrics.RecordPersonalizedCacheError("get")
			}
			s.logger.WarnContext(ctx, "recommendation.personalized.cache_read_failed",
				slog.String("cache_key", key),
				slog.String("error", err.Error()),
			)
		}
		return domain.RecommendationSet{}, false
	}
	if cached.ExpiresAt.IsZero() || !cached.ExpiresAt.After(now) || cached.StrategyID != domain.StrategyPersonalizedBehavior {
		return domain.RecommendationSet{}, false
	}
	set := domain.RecommendationSet{
		ID:          cached.RecommendationID,
		ContextKey:  cached.ContextKey,
		Type:        cached.Type,
		StrategyID:  cached.StrategyID,
		Items:       cached.Items,
		GeneratedAt: cached.CachedAt,
		ExpiresAt:   cached.ExpiresAt,
	}
	return trimSetItems(set, limit), true
}

func (s *PersonalizedService) cacheSet(ctx context.Context, identity ProfileIdentity, set domain.RecommendationSet) {
	if s.cache == nil {
		return
	}
	key, err := s.cacheKey(identity)
	if err != nil {
		if s.metrics != nil {
			s.metrics.RecordPersonalizedCacheError("key")
		}
		s.logger.WarnContext(ctx, "recommendation.personalized.cache_key_failed", slog.String("error", err.Error()))
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
	if err := s.cache.Set(ctx, key, cached, s.ttlFor(identity)); err != nil {
		if s.metrics != nil {
			s.metrics.RecordPersonalizedCacheError("set")
		}
		s.logger.WarnContext(ctx, "recommendation.personalized.cache_set_failed",
			slog.String("identity_type", identityType(identity)),
			slog.String("cache_key", key),
			slog.String("error", err.Error()),
		)
	}
}

func (s *PersonalizedService) cacheKey(identity ProfileIdentity) (string, error) {
	var (
		key string
		err error
	)
	if identity.IsGuest {
		key, err = s.keyBuilder.PersonalizedAnonymousKey(identity.AnonymousID)
	} else {
		key, err = s.keyBuilder.PersonalizedUserKey(identity.UserID)
	}
	if err != nil {
		return "", err
	}
	return s.keyBuilder.StrategyScopedKey(key, domain.StrategyPersonalizedBehavior)
}

func (s *PersonalizedService) ttlFor(identity ProfileIdentity) time.Duration {
	if identity.IsGuest && s.cfg.GuestTTL > 0 {
		return s.cfg.GuestTTL
	}
	return s.cfg.TTL
}

func (s *PersonalizedService) recordProfileMiss(reason string) {
	if s.metrics != nil {
		s.metrics.RecordPersonalizedProfileMiss(reason)
	}
}

func (s *PersonalizedService) recordBuild(outcome string, startedAt time.Time) {
	if s.metrics == nil {
		return
	}
	s.metrics.RecordPersonalizedBuild(outcome)
	s.metrics.ObservePersonalizedBuildDuration(s.now().Sub(startedAt))
}

func fallbackTargets(categoryID string) []personalizedFallbackTarget {
	targets := make([]personalizedFallbackTarget, 0, 2)
	categoryID = strings.TrimSpace(categoryID)
	if categoryID != "" {
		targets = append(targets, personalizedFallbackTarget{
			source:     string(domain.StrategyTrendingFallbackCategory),
			reason:     domain.FallbackReasonCategoryPopular,
			contextKey: "category:" + categoryID,
			strategyID: domain.StrategyTrendingFallbackCategory,
		})
	}
	targets = append(targets, personalizedFallbackTarget{
		source:     string(domain.StrategyTrendingFallbackGlobal),
		reason:     domain.FallbackReasonGlobalTrending,
		contextKey: domain.TrendingGlobalContextKey,
		strategyID: domain.StrategyTrendingRecentActivity,
	})
	return targets
}

func itemsFromPersonalizedProducts(products []domain.ScoredPersonalizedProduct) []domain.RecommendationSetItem {
	items := make([]domain.RecommendationSetItem, 0, len(products))
	for _, product := range products {
		items = append(items, domain.RecommendationSetItem{
			ProductID: product.ProductID,
			Score:     product.Score,
			Reason:    product.Reason,
		})
	}
	return rerankItems(items)
}

func rerankItems(items []domain.RecommendationSetItem) []domain.RecommendationSetItem {
	ranked := append([]domain.RecommendationSetItem(nil), items...)
	for i := range ranked {
		ranked[i].Rank = i + 1
	}
	return ranked
}

func trimSetItems(set domain.RecommendationSet, limit int) domain.RecommendationSet {
	if limit > 0 && len(set.Items) > limit {
		set.Items = append([]domain.RecommendationSetItem(nil), set.Items[:limit]...)
	}
	set.Items = rerankItems(set.Items)
	return set
}

func topScoreKeys(scores map[string]float64, limit int) []string {
	if limit <= 0 || len(scores) == 0 {
		return nil
	}
	type scoredKey struct {
		key   string
		score float64
	}
	items := make([]scoredKey, 0, len(scores))
	for key, score := range scores {
		key = strings.TrimSpace(key)
		if key == "" || score <= 0 {
			continue
		}
		items = append(items, scoredKey{key: key, score: score})
	}
	sort.SliceStable(items, func(i int, j int) bool {
		if items[i].score != items[j].score {
			return items[i].score > items[j].score
		}
		return items[i].key < items[j].key
	})
	if len(items) < limit {
		limit = len(items)
	}
	keys := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		keys = append(keys, items[i].key)
	}
	return keys
}

func identityType(identity ProfileIdentity) string {
	if identity.IsGuest {
		return "anonymous"
	}
	return "user"
}
