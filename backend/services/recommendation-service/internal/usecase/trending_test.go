package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
)

func TestTrendingServiceGeneratesSafeGlobalCategoryAndSellerSets(t *testing.T) {
	repo := &fakeRankingRepository{features: []domain.ProductFeature{
		rankingProduct("paid", "cat_a", "seller_a", domain.ProductCounters{Purchases7d: 20}),
		rankingProduct("views", "cat_a", "seller_b", domain.ProductCounters{Views24h: 100}),
		rankingProduct("other_category", "cat_b", "seller_a", domain.ProductCounters{CartAdds7d: 5}),
		{ProductID: "sold_out", CategoryID: "cat_a", SellerID: "seller_a", Status: domain.ProductStatusActive, StockStatus: "out_of_stock", QualityFlags: domain.ProductQualityFlags{IsRecommendable: true}, Counters: domain.ProductCounters{Purchases7d: 1000}},
	}}
	cache := &fakeRankingCache{}
	service := newTestTrendingService(t, repo, cache)

	result, err := service.GenerateAll(context.Background())
	if err != nil {
		t.Fatalf("GenerateAll() error = %v", err)
	}
	if result.EligibleCandidates != 3 || result.SetsGenerated != 5 {
		t.Fatalf("generation result = %+v, want 3 eligible candidates and 5 generated sets", result)
	}
	global := repo.byContext["home:global"]
	if len(global.Items) != 3 || global.Items[0].ProductID != "paid" {
		t.Fatalf("global items = %+v, want paid first and only eligible products", global.Items)
	}
	category := repo.byContext["category:cat_a"]
	if len(category.Items) != 2 || category.Items[0].Reason != "category_recent_activity" {
		t.Fatalf("category items = %+v, want two scoped items", category.Items)
	}
	if got := repo.byContext["seller:seller_b"].Items; len(got) != 1 || got[0].ProductID != "views" {
		t.Fatalf("seller_b items = %+v, want only views", got)
	}
	if cache.setCalls != result.SetsGenerated {
		t.Fatalf("cache set calls = %d, want %d", cache.setCalls, result.SetsGenerated)
	}
}

func TestTrendingServiceDoesNotCacheWhenDurableSaveFails(t *testing.T) {
	repo := &fakeRankingRepository{
		features:  []domain.ProductFeature{rankingProduct("paid", "cat_a", "seller_a", domain.ProductCounters{Purchases7d: 1})},
		upsertErr: errors.New("mongo down"),
	}
	cache := &fakeRankingCache{}
	service := newTestTrendingService(t, repo, cache)

	_, err := service.GenerateAll(context.Background())
	if !errors.Is(err, domain.ErrRecommendationStorage) {
		t.Fatalf("GenerateAll() error = %v, want storage failure", err)
	}
	if cache.setCalls != 0 {
		t.Fatalf("cache set calls = %d, want zero when durable save fails", cache.setCalls)
	}
}

func TestTrendingServiceKeepsDurableSetsWhenCacheWriteFails(t *testing.T) {
	repo := &fakeRankingRepository{features: []domain.ProductFeature{
		rankingProduct("paid", "cat_a", "seller_a", domain.ProductCounters{Purchases7d: 1}),
	}}
	cache := &fakeRankingCache{setErr: errors.New("redis down")}
	service := newTestTrendingService(t, repo, cache)

	result, err := service.GenerateAll(context.Background())
	if err != nil {
		t.Fatalf("GenerateAll() error = %v, want best-effort cache failure only", err)
	}
	if len(repo.upserted) != result.SetsGenerated || len(repo.upserted) == 0 {
		t.Fatalf("durable sets = %d, generation result = %+v", len(repo.upserted), result)
	}
}

func TestTrendingServiceResolveGeneratedSetFallsBackToGlobal(t *testing.T) {
	now := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
	repo := &fakeRankingRepository{byContext: map[string]domain.RecommendationSet{
		"home:global": {
			ID:          "reco_trending_global",
			ContextKey:  "home:global",
			Type:        domain.RecommendationTypeTrending,
			StrategyID:  domain.StrategyTrendingRecentActivity,
			Items:       []domain.RecommendationSetItem{{ProductID: "paid", Rank: 1, Score: 8}},
			GeneratedAt: now,
			ExpiresAt:   now.Add(time.Minute),
		},
	}}
	service := newTestTrendingService(t, repo, nil)

	result, err := service.ResolveGeneratedSet(context.Background(), domain.RankRequest{
		Scope:   domain.RankingScopeCategory,
		ScopeID: "cat_missing",
		Now:     now,
	})
	if err != nil {
		t.Fatalf("ResolveGeneratedSet() error = %v", err)
	}
	if !result.Fallback || result.Empty || result.Set.ContextKey != "home:global" {
		t.Fatalf("resolution = %+v, want global fallback", result)
	}
}

type fakeRankingRepository struct {
	features  []domain.ProductFeature
	listErr   error
	upsertErr error
	upserted  []domain.RecommendationSet
	byContext map[string]domain.RecommendationSet
}

func (f *fakeRankingRepository) ListEligibleProductFeatures(ctx context.Context) ([]domain.ProductFeature, error) {
	return f.features, f.listErr
}

func (f *fakeRankingRepository) UpsertRecommendationSet(ctx context.Context, set domain.RecommendationSet) error {
	if f.upsertErr != nil {
		return f.upsertErr
	}
	f.upserted = append(f.upserted, set)
	if f.byContext == nil {
		f.byContext = make(map[string]domain.RecommendationSet)
	}
	f.byContext[set.ContextKey] = set
	return nil
}

func (f *fakeRankingRepository) GetRecommendationSet(ctx context.Context, contextKey string, strategyID domain.StrategyID, now time.Time) (domain.RecommendationSet, error) {
	set, ok := f.byContext[contextKey]
	if !ok || set.StrategyID != strategyID || set.IsExpired(now) {
		return domain.RecommendationSet{}, domain.ErrRecommendationSetNotFound
	}
	return set, nil
}

type fakeRankingCache struct {
	setCalls int
	setErr   error
	values   map[string]domain.CachedRecommendation
}

func (f *fakeRankingCache) Set(ctx context.Context, key string, value domain.CachedRecommendation, ttl time.Duration) error {
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

func (f *fakeRankingCache) Get(ctx context.Context, key string) (domain.CachedRecommendation, time.Duration, error) {
	value, ok := f.values[key]
	if !ok {
		return domain.CachedRecommendation{}, 0, domain.ErrRecommendationCacheMiss
	}
	return value, time.Minute, nil
}

func newTestTrendingService(t *testing.T, repo *fakeRankingRepository, cache RankingResultCache) *TrendingService {
	t.Helper()
	service, err := NewTrendingService(repo, repo, cache, TrendingServiceConfig{
		Weights:             domain.DefaultPopularityWeights(),
		Limit:               12,
		MaxIdentifierLength: 128,
		TTL:                 15 * time.Minute,
		RebuildInterval:     15 * time.Minute,
	}, nil, nil)
	if err != nil {
		t.Fatalf("NewTrendingService() error = %v", err)
	}
	service.now = func() time.Time { return time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC) }
	return service
}

func rankingProduct(productID string, categoryID string, sellerID string, counters domain.ProductCounters) domain.ProductFeature {
	return domain.ProductFeature{
		ProductID:   productID,
		CategoryID:  categoryID,
		SellerID:    sellerID,
		Status:      domain.ProductStatusActive,
		StockStatus: domain.ProductStockStatusInStock,
		QualityFlags: domain.ProductQualityFlags{
			IsRecommendable: true,
		},
		Counters: counters,
	}
}
