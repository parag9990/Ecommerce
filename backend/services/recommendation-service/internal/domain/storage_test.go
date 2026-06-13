package domain

import (
	"errors"
	"testing"
	"time"
)

func TestCacheKeyBuilderBuildsTaskTwoKeys(t *testing.T) {
	builder, err := NewCacheKeyBuilder(DefaultCacheKeyPrefix, 128)
	if err != nil {
		t.Fatalf("NewCacheKeyBuilder() error = %v", err)
	}

	tests := []struct {
		name string
		got  func() (string, error)
		want string
	}{
		{
			name: "global trending",
			got: func() (string, error) {
				return builder.TrendingGlobalKey(), nil
			},
			want: "reco:v1:trending:global",
		},
		{
			name: "category trending",
			got: func() (string, error) {
				return builder.TrendingCategoryKey("cat_shoes")
			},
			want: "reco:v1:trending:category:cat_shoes",
		},
		{
			name: "similar product",
			got: func() (string, error) {
				return builder.SimilarProductKey("prod_123")
			},
			want: "reco:v1:similar:product:prod_123",
		},
		{
			name: "personalized user",
			got: func() (string, error) {
				return builder.PersonalizedUserKey("user_123")
			},
			want: "reco:v1:personalized:user:user_123",
		},
		{
			name: "guest personalized",
			got: func() (string, error) {
				return builder.PersonalizedAnonymousKey("anon_123")
			},
			want: "reco:v1:personalized:anon:anon_123",
		},
		{
			name: "frequently bought together",
			got: func() (string, error) {
				return builder.FrequentlyBoughtTogetherProductKey("prod_123")
			},
			want: "reco:v1:fbt:product:prod_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.got()
			if err != nil {
				t.Fatalf("key builder error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("key = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestCacheKeyBuilderRejectsUnsafeIdentifiers(t *testing.T) {
	builder, err := NewCacheKeyBuilder(DefaultCacheKeyPrefix, 8)
	if err != nil {
		t.Fatalf("NewCacheKeyBuilder() error = %v", err)
	}
	if _, err := builder.SimilarProductKey("prod:123"); !errors.Is(err, ErrInvalidCacheKey) {
		t.Fatalf("SimilarProductKey() error = %v, want %v", err, ErrInvalidCacheKey)
	}
	if _, err := builder.SimilarProductKey("product_123"); !errors.Is(err, ErrInvalidCacheKey) {
		t.Fatalf("SimilarProductKey() error = %v, want %v", err, ErrInvalidCacheKey)
	}
}

func TestCacheKeyBuilderBuildsStrategyScopedKeys(t *testing.T) {
	builder, err := NewCacheKeyBuilder(DefaultCacheKeyPrefix, 128)
	if err != nil {
		t.Fatalf("NewCacheKeyBuilder() error = %v", err)
	}
	base, err := builder.PersonalizedUserKey("user_123")
	if err != nil {
		t.Fatalf("PersonalizedUserKey() error = %v", err)
	}
	got, err := builder.StrategyScopedKey(base, StrategyPersonalizedBehavior)
	if err != nil {
		t.Fatalf("StrategyScopedKey() error = %v", err)
	}
	want := "reco:v1:personalized:user:user_123:strategy:personalized_v1_behavior"
	if got != want {
		t.Fatalf("strategy key = %s, want %s", got, want)
	}
}

func TestStoragePlanContainsMongoAndRedisDecision(t *testing.T) {
	plan := BuildStoragePlan(
		RecommendationDatabaseName,
		DefaultCacheKeyPrefix,
		CacheTTLPolicy{
			Default:      15 * time.Minute,
			Personalized: 5 * time.Minute,
			Guest:        5 * time.Minute,
		},
		DefaultInteractionRetention,
	)
	if plan.DatabaseName != "recommendation_db" {
		t.Fatalf("database = %s, want recommendation_db", plan.DatabaseName)
	}
	if len(plan.MongoCollections) != 9 {
		t.Fatalf("mongo collections = %d, want 9", len(plan.MongoCollections))
	}
	if len(plan.RedisCaches) == 0 {
		t.Fatal("redis caches are empty")
	}
	if plan.TTLPolicy.Default != 15*time.Minute {
		t.Fatalf("default ttl = %s, want 15m", plan.TTLPolicy.Default)
	}
}
