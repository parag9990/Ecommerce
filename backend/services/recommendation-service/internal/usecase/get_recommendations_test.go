package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
)

func TestGetRecommendationsServesSimilarProductsFromCache(t *testing.T) {
	now := recommendationServingTestNow()
	cache := &fakeServingCache{values: map[string]domain.CachedRecommendation{
		"reco:v1:similar:product:prod_1:strategy:similar_v1_attributes": {
			RecommendationID: "reco_similar_prod_1",
			ContextKey:       "similar:product:prod_1",
			Type:             domain.RecommendationTypeSimilarProducts,
			StrategyID:       domain.StrategySimilarAttributes,
			Items:            []domain.RecommendationSetItem{{ProductID: "prod_2", Rank: 1, Score: 9.5, Reason: "same_category"}},
			CachedAt:         now.Add(-time.Minute),
			ExpiresAt:        now.Add(10 * time.Minute),
		},
	}}
	service := newTestRecommendationServingService(t, nil, cache, nil, nil)
	service.now = func() time.Time { return now }

	got, err := service.GetRecommendations(context.Background(), GetRecommendationsInput{
		Context:   domain.ContextProductDetail,
		ProductID: "prod_1",
		Limit:     6,
	})
	if err != nil {
		t.Fatalf("GetRecommendations() error = %v", err)
	}
	if got.Source != recommendationSourceRedis || got.Result.Type != domain.RecommendationTypeSimilarProducts {
		t.Fatalf("output source/type = %s/%s, want redis/similar", got.Source, got.Result.Type)
	}
	if len(got.Result.Items) != 1 || got.Result.Items[0].ProductID != "prod_2" {
		t.Fatalf("items = %+v, want cached similar product", got.Result.Items)
	}
}

func TestGetRecommendationsImplicitPersonalizedUnsupportedContextFallsBackToCategoryTrending(t *testing.T) {
	now := recommendationServingTestNow()
	sets := &fakeServingSetRepository{sets: map[string]domain.RecommendationSet{
		servingSetKey("category:cat_1", domain.StrategyTrendingFallbackCategory): servingSet(
			"reco_category_cat_1",
			"category:cat_1",
			domain.RecommendationTypeTrending,
			domain.StrategyTrendingFallbackCategory,
			now,
			"prod_1",
		),
	}}
	service := newTestRecommendationServingService(t, sets, nil, nil, fakePersonalizedBuilder{err: domain.ErrUnsupportedContext})
	service.now = func() time.Time { return now }

	got, err := service.GetRecommendations(context.Background(), GetRecommendationsInput{
		UserID:     "user_1",
		Context:    domain.ContextCategoryListing,
		CategoryID: "cat_1",
		Limit:      4,
	})
	if err != nil {
		t.Fatalf("GetRecommendations() error = %v", err)
	}
	if !got.Fallback || got.Result.Type != domain.RecommendationTypeTrending {
		t.Fatalf("fallback/type = %t/%s, want trending fallback", got.Fallback, got.Result.Type)
	}
	if got.Result.StrategyID != domain.StrategyTrendingFallbackCategory {
		t.Fatalf("strategy = %s, want category fallback strategy", got.Result.StrategyID)
	}
}

func TestGetRecommendationsMissingSimilarSetFallsBackToGlobalTrending(t *testing.T) {
	now := recommendationServingTestNow()
	sets := &fakeServingSetRepository{sets: map[string]domain.RecommendationSet{
		servingSetKey(domain.TrendingGlobalContextKey, domain.StrategyTrendingRecentActivity): servingSet(
			"reco_trending_global",
			domain.TrendingGlobalContextKey,
			domain.RecommendationTypeTrending,
			domain.StrategyTrendingRecentActivity,
			now,
			"prod_global",
		),
	}}
	service := newTestRecommendationServingService(t, sets, nil, nil, nil)
	service.now = func() time.Time { return now }

	got, err := service.GetRecommendations(context.Background(), GetRecommendationsInput{
		Context:   domain.ContextProductDetail,
		ProductID: "prod_missing",
		Limit:     4,
	})
	if err != nil {
		t.Fatalf("GetRecommendations() error = %v", err)
	}
	if !got.Fallback || got.Result.Type != domain.RecommendationTypeTrending {
		t.Fatalf("fallback/type = %t/%s, want global trending fallback", got.Fallback, got.Result.Type)
	}
	if len(got.Result.Items) != 1 || got.Result.Items[0].ProductID != "prod_global" {
		t.Fatalf("items = %+v, want global trending fallback item", got.Result.Items)
	}
}

func TestGetRecommendationsUsesAssignedTrendingStrategy(t *testing.T) {
	now := recommendationServingTestNow()
	sets := &fakeServingSetRepository{sets: map[string]domain.RecommendationSet{
		servingSetKey(domain.TrendingGlobalContextKey, domain.StrategyTrendingRecentActivity): servingSet(
			"reco_trending_global",
			domain.TrendingGlobalContextKey,
			domain.RecommendationTypeTrending,
			domain.StrategyTrendingRecentActivity,
			now,
			"prod_global",
		),
	}}
	service := newTestRecommendationServingService(t, sets, nil, nil, nil)
	service.now = func() time.Time { return now }
	service.assigner = fakeStrategyAssigner{assignment: domain.StrategyAssignment{
		ExperimentID: "reco_home_strategy_2026_05",
		VariantID:    "control",
		StrategyID:   domain.StrategyTrendingRecentActivity,
		Assigned:     true,
	}}
	service.abFailOpen = true

	got, err := service.GetRecommendations(context.Background(), GetRecommendationsInput{
		UserID:  "user_1",
		Context: domain.ContextHomeFeed,
		Limit:   4,
	})
	if err != nil {
		t.Fatalf("GetRecommendations() error = %v", err)
	}
	if got.Result.Type != domain.RecommendationTypeTrending || got.Result.StrategyID != domain.StrategyTrendingRecentActivity {
		t.Fatalf("type/strategy = %s/%s, want trending/%s", got.Result.Type, got.Result.StrategyID, domain.StrategyTrendingRecentActivity)
	}
	if len(got.Result.Items) != 1 || got.Result.Items[0].ProductID != "prod_global" {
		t.Fatalf("items = %+v, want assigned trending item", got.Result.Items)
	}
}

func TestGetRecommendationsAssignmentFailureFallsBackToDefaultStrategy(t *testing.T) {
	now := recommendationServingTestNow()
	sets := &fakeServingSetRepository{sets: map[string]domain.RecommendationSet{
		servingSetKey("home:user:user_1", domain.StrategyPersonalizedBehavior): servingSet(
			"reco_personalized_user_1",
			"home:user:user_1",
			domain.RecommendationTypePersonalized,
			domain.StrategyPersonalizedBehavior,
			now,
			"prod_personalized",
		),
	}}
	service := newTestRecommendationServingService(t, sets, nil, nil, nil)
	service.now = func() time.Time { return now }
	service.assigner = fakeStrategyAssigner{err: domain.ErrRecommendationStorage}
	service.abFailOpen = true

	got, err := service.GetRecommendations(context.Background(), GetRecommendationsInput{
		UserID:  "user_1",
		Context: domain.ContextHomeFeed,
		Type:    domain.RecommendationTypePersonalized,
		Limit:   4,
	})
	if err != nil {
		t.Fatalf("GetRecommendations() error = %v", err)
	}
	if got.Result.StrategyID != domain.StrategyPersonalizedBehavior {
		t.Fatalf("strategy = %s, want default personalized", got.Result.StrategyID)
	}
}

type fakeServingSetRepository struct {
	sets map[string]domain.RecommendationSet
	err  error
}

func (f *fakeServingSetRepository) GetRecommendationSet(ctx context.Context, contextKey string, strategyID domain.StrategyID, now time.Time) (domain.RecommendationSet, error) {
	if f.err != nil {
		return domain.RecommendationSet{}, f.err
	}
	set, ok := f.sets[servingSetKey(contextKey, strategyID)]
	if !ok || set.IsExpired(now) {
		return domain.RecommendationSet{}, domain.ErrRecommendationSetNotFound
	}
	return set, nil
}

func (f *fakeServingSetRepository) UpsertRecommendationSet(ctx context.Context, set domain.RecommendationSet) error {
	return errors.New("not implemented")
}

type fakeServingCache struct {
	values map[string]domain.CachedRecommendation
	err    error
}

func (f *fakeServingCache) Get(ctx context.Context, key string) (domain.CachedRecommendation, time.Duration, error) {
	if f.err != nil {
		return domain.CachedRecommendation{}, 0, f.err
	}
	value, ok := f.values[key]
	if !ok {
		return domain.CachedRecommendation{}, 0, domain.ErrRecommendationCacheMiss
	}
	return value, time.Until(value.ExpiresAt), nil
}

func (f *fakeServingCache) Set(ctx context.Context, key string, cached domain.CachedRecommendation, ttl time.Duration) error {
	if f.values == nil {
		f.values = make(map[string]domain.CachedRecommendation)
	}
	f.values[key] = cached
	return nil
}

type fakePersonalizedBuilder struct {
	set domain.RecommendationSet
	err error
}

func (f fakePersonalizedBuilder) Build(ctx context.Context, input PersonalizedBuildInput) (domain.RecommendationSet, error) {
	return f.set, f.err
}

type fakeStrategyAssigner struct {
	assignment domain.StrategyAssignment
	err        error
}

func (f fakeStrategyAssigner) Assign(ctx context.Context, input StrategyAssignmentInput) (domain.StrategyAssignment, error) {
	if f.err != nil {
		return domain.StrategyAssignment{}, f.err
	}
	return f.assignment, nil
}

func newTestRecommendationServingService(
	t *testing.T,
	sets RecommendationSetRepository,
	cache RankingResultCache,
	trending GeneratedSetResolver,
	personalized PersonalizedRecommendationBuilder,
) *GetRecommendationsService {
	t.Helper()
	service, err := NewGetRecommendationsService(
		newTestDefinitionService(t),
		sets,
		cache,
		trending,
		personalized,
		nil,
		GetRecommendationsConfig{
			CacheKeyPrefix:      domain.DefaultCacheKeyPrefix,
			MaxIdentifierLength: 128,
			TTLPolicy: domain.CacheTTLPolicy{
				Default:      domain.DefaultRecommendationCacheTTL,
				Personalized: domain.DefaultPersonalizedCacheTTL,
				Guest:        domain.DefaultGuestCacheTTL,
				RebuildLock:  time.Minute,
			},
		},
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewGetRecommendationsService() error = %v", err)
	}
	return service
}

func servingSet(id string, contextKey string, typ domain.RecommendationType, strategyID domain.StrategyID, now time.Time, productIDs ...string) domain.RecommendationSet {
	items := make([]domain.RecommendationSetItem, 0, len(productIDs))
	for index, productID := range productIDs {
		items = append(items, domain.RecommendationSetItem{
			ProductID: productID,
			Rank:      index + 1,
			Score:     float64(100 - index),
			Reason:    "test",
		})
	}
	return domain.RecommendationSet{
		ID:          id,
		ContextKey:  contextKey,
		Type:        typ,
		StrategyID:  strategyID,
		Items:       items,
		GeneratedAt: now,
		ExpiresAt:   now.Add(time.Hour),
	}
}

func servingSetKey(contextKey string, strategyID domain.StrategyID) string {
	return contextKey + "|" + string(strategyID)
}

func recommendationServingTestNow() time.Time {
	return time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
}
