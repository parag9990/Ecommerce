package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
)

func TestPersonalizedServiceBuildRanksBehaviorAndBackfills(t *testing.T) {
	now := personalizedTestNow()
	repo := &fakePersonalizedRepository{
		profile: domain.UserFeatureProfile{
			ProfileKey:     "user:u1",
			CategoryScores: map[string]float64{"cat_a": 10},
			SellerScores:   map[string]float64{"seller_a": 5},
			BrandScores:    map[string]float64{"brand_a": 4},
			PriceAffinity:  &domain.PriceAffinity{PreferredBuckets: []string{"2000_2999"}},
			RecentProducts: []domain.RecentProduct{{ProductID: "direct_1", OccurredAt: now.Add(-time.Hour)}},
			LastEventAt:    now.Add(-time.Hour),
		},
		counters: []domain.UserProductCounter{
			{ProductID: "direct_1", WeightedScoreInput: 10, Counters: domain.InteractionCounters{CartAdds: 1}},
		},
		candidates: []domain.ProductFeature{
			personalizedCandidate("direct_1", "cat_a", "seller_a", "brand_a", "2000_2999", domain.ProductCounters{Purchases7d: 1}),
			personalizedCandidate("category_1", "cat_a", "seller_b", "brand_b", "3000_4999", domain.ProductCounters{Purchases7d: 20}),
			{ProductID: "sold_out", CategoryID: "cat_a", Status: domain.ProductStatusActive, StockStatus: "out_of_stock", QualityFlags: domain.ProductQualityFlags{IsRecommendable: true}},
		},
		sets: map[string]domain.RecommendationSet{
			recommendationSetKey("category:cat_a", domain.StrategyTrendingFallbackCategory): fallbackSet("category:cat_a", domain.StrategyTrendingFallbackCategory, now, "fallback_1"),
		},
	}
	cache := &fakePersonalizedCache{}
	service := newTestPersonalizedService(t, repo, cache)

	got, err := service.Build(context.Background(), PersonalizedBuildInput{
		UserID:     "u1",
		CategoryID: "cat_a",
		Limit:      3,
		Now:        now,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if got.Type != domain.RecommendationTypePersonalized || got.StrategyID != domain.StrategyPersonalizedBehavior {
		t.Fatalf("set identity = %s/%s, want personalized behavior", got.Type, got.StrategyID)
	}
	if len(got.Items) != 3 {
		t.Fatalf("items = %+v, want 3 with fallback backfill", got.Items)
	}
	if got.Items[0].ProductID != "direct_1" || got.Items[0].Reason != "preferred_category_and_cart_affinity" {
		t.Fatalf("top item = %+v, want direct behavior item first", got.Items[0])
	}
	if got.Items[2].ProductID != "fallback_1" || got.Items[2].Reason != domain.FallbackReasonCategoryPopular {
		t.Fatalf("backfill item = %+v, want category fallback", got.Items[2])
	}
	if got.Metadata["fallback_used"] != "true" || got.Metadata["fallback_source"] != string(domain.StrategyTrendingFallbackCategory) {
		t.Fatalf("metadata = %+v, want category fallback metadata", got.Metadata)
	}
	if len(repo.upserted) != 1 || cache.setCalls != 1 {
		t.Fatalf("upserts = %d cache sets = %d, want durable save and cache publish", len(repo.upserted), cache.setCalls)
	}
}

func TestPersonalizedServiceUsesGlobalFallbackWhenProfileMissing(t *testing.T) {
	now := personalizedTestNow()
	repo := &fakePersonalizedRepository{
		profileErr: domain.ErrFeatureProfileNotFound,
		sets: map[string]domain.RecommendationSet{
			recommendationSetKey(domain.TrendingGlobalContextKey, domain.StrategyTrendingRecentActivity): fallbackSet(domain.TrendingGlobalContextKey, domain.StrategyTrendingRecentActivity, now, "popular_1"),
		},
	}
	service := newTestPersonalizedService(t, repo, &fakePersonalizedCache{})

	got, err := service.Build(context.Background(), PersonalizedBuildInput{UserID: "new_user", Limit: 2, Now: now})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if len(got.Items) != 1 || got.Items[0].ProductID != "popular_1" {
		t.Fatalf("items = %+v, want global fallback product", got.Items)
	}
	if got.Items[0].Reason != domain.FallbackReasonGlobalTrending {
		t.Fatalf("reason = %s, want global fallback reason", got.Items[0].Reason)
	}
	if got.Metadata["cold_start_reason"] != "profile_unavailable" {
		t.Fatalf("metadata = %+v, want cold-start reason", got.Metadata)
	}
}

func TestPersonalizedServiceDoesNotCacheWhenDurableSaveFails(t *testing.T) {
	now := personalizedTestNow()
	repo := fakePersonalizedRepoWithProfile(now)
	repo.upsertErr = errors.New("mongo down")
	cache := &fakePersonalizedCache{}
	service := newTestPersonalizedService(t, repo, cache)

	_, err := service.Build(context.Background(), PersonalizedBuildInput{UserID: "u1", Limit: 2, Now: now})
	if !errors.Is(err, domain.ErrRecommendationStorage) {
		t.Fatalf("Build() error = %v, want storage error", err)
	}
	if cache.setCalls != 0 {
		t.Fatalf("cache set calls = %d, want zero after durable save failure", cache.setCalls)
	}
}

func TestPersonalizedServiceKeepsDurableResultWhenCacheWriteFails(t *testing.T) {
	now := personalizedTestNow()
	repo := fakePersonalizedRepoWithProfile(now)
	cache := &fakePersonalizedCache{setErr: errors.New("redis down")}
	service := newTestPersonalizedService(t, repo, cache)

	got, err := service.Build(context.Background(), PersonalizedBuildInput{UserID: "u1", Limit: 2, Now: now})
	if err != nil {
		t.Fatalf("Build() error = %v, want best-effort cache failure only", err)
	}
	if len(got.Items) == 0 || len(repo.upserted) != 1 || cache.setCalls != 1 {
		t.Fatalf("items=%+v upserts=%d cache calls=%d, want durable result despite cache failure", got.Items, len(repo.upserted), cache.setCalls)
	}
}

func TestPersonalizedServiceReturnsCachedSetBeforeFeatureReads(t *testing.T) {
	now := personalizedTestNow()
	cache := &fakePersonalizedCache{values: map[string]domain.CachedRecommendation{
		"reco:v1:personalized:user:u1:strategy:personalized_v1_behavior": {
			RecommendationID: "cached",
			ContextKey:       "home:user:u1",
			Type:             domain.RecommendationTypePersonalized,
			StrategyID:       domain.StrategyPersonalizedBehavior,
			Items:            []domain.RecommendationSetItem{{ProductID: "cached_1", Score: 10, Rank: 1}},
			CachedAt:         now,
			ExpiresAt:        now.Add(time.Minute),
		},
	}}
	repo := &fakePersonalizedRepository{profileErr: errors.New("should not read profile")}
	service := newTestPersonalizedService(t, repo, cache)

	got, err := service.Build(context.Background(), PersonalizedBuildInput{UserID: "u1", Limit: 1, Now: now})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if got.ID != "cached" || repo.profileReads != 0 || len(repo.upserted) != 0 {
		t.Fatalf("got=%+v profile reads=%d upserts=%d, want cache hit only", got, repo.profileReads, len(repo.upserted))
	}
}

func TestPersonalizedServiceAppliesSellerDiversity(t *testing.T) {
	now := personalizedTestNow()
	repo := &fakePersonalizedRepository{
		profile: domain.UserFeatureProfile{
			ProfileKey:     "user:u1",
			CategoryScores: map[string]float64{"cat_a": 10},
			LastEventAt:    now.Add(-time.Hour),
		},
		candidates: []domain.ProductFeature{
			personalizedCandidate("seller_a_1", "cat_a", "seller_a", "", "", domain.ProductCounters{Purchases7d: 30}),
			personalizedCandidate("seller_a_2", "cat_a", "seller_a", "", "", domain.ProductCounters{Purchases7d: 20}),
			personalizedCandidate("seller_a_3", "cat_a", "seller_a", "", "", domain.ProductCounters{Purchases7d: 10}),
			personalizedCandidate("seller_b_1", "cat_a", "seller_b", "", "", domain.ProductCounters{Purchases7d: 1}),
		},
	}
	service := newTestPersonalizedService(t, repo, nil)
	service.cfg.MaxItemsPerSeller = 2

	got, err := service.Build(context.Background(), PersonalizedBuildInput{UserID: "u1", Limit: 3, Now: now})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if len(got.Items) != 3 {
		t.Fatalf("items = %+v, want 3", got.Items)
	}
	if got.Items[2].ProductID != "seller_b_1" {
		t.Fatalf("third item = %+v, want seller diversity to include seller_b_1", got.Items[2])
	}
}

type fakePersonalizedRepository struct {
	profile      domain.UserFeatureProfile
	profileErr   error
	profileReads int

	counters   []domain.UserProductCounter
	counterErr error

	candidates   []domain.ProductFeature
	candidateErr error
	lastQuery    domain.PersonalizedCandidateQuery

	sets      map[string]domain.RecommendationSet
	upsertErr error
	upserted  []domain.RecommendationSet
}

func (f *fakePersonalizedRepository) GetUserFeatureProfile(ctx context.Context, profileKey string) (domain.UserFeatureProfile, error) {
	f.profileReads++
	if f.profileErr != nil {
		return domain.UserFeatureProfile{}, f.profileErr
	}
	return f.profile, nil
}

func (f *fakePersonalizedRepository) ListUserProductCounters(ctx context.Context, profileKey string, limit int) ([]domain.UserProductCounter, error) {
	return append([]domain.UserProductCounter(nil), f.counters...), f.counterErr
}

func (f *fakePersonalizedRepository) ListPersonalizedCandidates(ctx context.Context, query domain.PersonalizedCandidateQuery) ([]domain.ProductFeature, error) {
	f.lastQuery = query
	return append([]domain.ProductFeature(nil), f.candidates...), f.candidateErr
}

func (f *fakePersonalizedRepository) GetRecommendationSet(ctx context.Context, contextKey string, strategyID domain.StrategyID, now time.Time) (domain.RecommendationSet, error) {
	set, ok := f.sets[recommendationSetKey(contextKey, strategyID)]
	if !ok || set.IsExpired(now) {
		return domain.RecommendationSet{}, domain.ErrRecommendationSetNotFound
	}
	return set, nil
}

func (f *fakePersonalizedRepository) UpsertRecommendationSet(ctx context.Context, set domain.RecommendationSet) error {
	if f.upsertErr != nil {
		return f.upsertErr
	}
	f.upserted = append(f.upserted, set)
	if f.sets == nil {
		f.sets = make(map[string]domain.RecommendationSet)
	}
	f.sets[recommendationSetKey(set.ContextKey, set.StrategyID)] = set
	return nil
}

type fakePersonalizedCache struct {
	values   map[string]domain.CachedRecommendation
	setCalls int
	setErr   error
}

func (f *fakePersonalizedCache) Get(ctx context.Context, key string) (domain.CachedRecommendation, time.Duration, error) {
	value, ok := f.values[key]
	if !ok {
		return domain.CachedRecommendation{}, 0, domain.ErrRecommendationCacheMiss
	}
	return value, time.Until(value.ExpiresAt), nil
}

func (f *fakePersonalizedCache) Set(ctx context.Context, key string, value domain.CachedRecommendation, ttl time.Duration) error {
	f.setCalls++
	if f.setErr != nil {
		return f.setErr
	}
	if f.values == nil {
		f.values = make(map[string]domain.CachedRecommendation)
	}
	f.values[key] = value
	return nil
}

func newTestPersonalizedService(t *testing.T, repo *fakePersonalizedRepository, cache RankingResultCache) *PersonalizedService {
	t.Helper()
	service, err := NewPersonalizedService(repo, repo, cache, PersonalizedServiceConfig{
		Limit:                   3,
		MaxLimit:                10,
		MaxIdentifierLength:     128,
		TTL:                     5 * time.Minute,
		GuestTTL:                5 * time.Minute,
		MinPositiveInteractions: 3,
		ProfileMaxAge:           30 * 24 * time.Hour,
		MaxCandidates:           20,
		MaxItemsPerSeller:       3,
		TopCategories:           3,
		TopSellers:              2,
		TopBrands:               2,
		DirectProductLimit:      20,
		Weights:                 domain.DefaultPersonalizationWeights(),
	}, nil, nil)
	if err != nil {
		t.Fatalf("NewPersonalizedService() error = %v", err)
	}
	service.now = personalizedTestNow
	return service
}

func fakePersonalizedRepoWithProfile(now time.Time) *fakePersonalizedRepository {
	return &fakePersonalizedRepository{
		profile: domain.UserFeatureProfile{
			ProfileKey:     "user:u1",
			CategoryScores: map[string]float64{"cat_a": 10},
			LastEventAt:    now.Add(-time.Hour),
		},
		candidates: []domain.ProductFeature{
			personalizedCandidate("product_1", "cat_a", "seller_a", "", "", domain.ProductCounters{Purchases7d: 10}),
		},
	}
}

func personalizedCandidate(productID string, categoryID string, sellerID string, brandID string, priceBucket string, counters domain.ProductCounters) domain.ProductFeature {
	var price *domain.ProductPrice
	if priceBucket != "" {
		price = &domain.ProductPrice{Bucket: priceBucket}
	}
	return domain.ProductFeature{
		ProductID:   productID,
		CategoryID:  categoryID,
		SellerID:    sellerID,
		BrandID:     brandID,
		Status:      domain.ProductStatusActive,
		StockStatus: domain.ProductStockStatusInStock,
		Price:       price,
		QualityFlags: domain.ProductQualityFlags{
			IsRecommendable: true,
		},
		Counters: counters,
	}
}

func fallbackSet(contextKey string, strategyID domain.StrategyID, now time.Time, productIDs ...string) domain.RecommendationSet {
	items := make([]domain.RecommendationSetItem, 0, len(productIDs))
	for i, productID := range productIDs {
		items = append(items, domain.RecommendationSetItem{ProductID: productID, Score: float64(100 - i), Rank: i + 1})
	}
	return domain.RecommendationSet{
		ID:          "fallback_" + contextKey,
		ContextKey:  contextKey,
		Type:        domain.RecommendationTypeTrending,
		StrategyID:  strategyID,
		Items:       items,
		GeneratedAt: now,
		ExpiresAt:   now.Add(time.Hour),
	}
}

func recommendationSetKey(contextKey string, strategyID domain.StrategyID) string {
	return contextKey + "|" + string(strategyID)
}

func personalizedTestNow() time.Time {
	return time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
}
