package repository

import (
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

func TestMongoFeatureRepositoryIndexModelsMatchFeatureStoreCollections(t *testing.T) {
	repo := &MongoFeatureRepository{interactionTTL: domain.DefaultInteractionRetention}
	models := repo.indexModels()

	tests := []struct {
		collection string
		want       int
	}{
		{collection: domain.CollectionUserInteractions, want: 6},
		{collection: domain.CollectionRecommendationSets, want: 3},
		{collection: domain.CollectionABTestAssignments, want: 5},
		{collection: domain.CollectionProductFeatures, want: 9},
		{collection: domain.CollectionUserFeatureProfiles, want: 4},
		{collection: domain.CollectionUserProductCounters, want: 5},
		{collection: domain.CollectionProductCooccurrence, want: 4},
		{collection: domain.CollectionFeatureJobRuns, want: 2},
		{collection: domain.CollectionFeatureProcessedEvents, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.collection, func(t *testing.T) {
			if got := len(models[tt.collection]); got != tt.want {
				t.Fatalf("index count = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestRedisRecommendationCacheDefaultsTTLPolicy(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer client.Close()

	cache, err := NewRedisRecommendationCache(client, RedisRecommendationCacheConfig{}, nil)
	if err != nil {
		t.Fatalf("NewRedisRecommendationCache() error = %v", err)
	}
	if cache.ttl.Default != 15*time.Minute {
		t.Fatalf("default ttl = %s, want 15m", cache.ttl.Default)
	}
	if cache.ttl.Personalized != 5*time.Minute {
		t.Fatalf("personalized ttl = %s, want 5m", cache.ttl.Personalized)
	}
	if cache.lockTTL != time.Minute {
		t.Fatalf("lock ttl = %s, want 1m", cache.lockTTL)
	}
}
