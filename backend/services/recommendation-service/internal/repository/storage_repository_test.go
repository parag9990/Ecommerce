package repository

import (
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
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

func TestProductFeatureSetOnInsertMakesInteractionDerivedProductsRankable(t *testing.T) {
	now := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
	values := bsonDMap(productFeatureSetOnInsert(domain.UserInteraction{ProductID: "prod_1"}, now))

	if values["_id"] != "prod_1" || values["product_id"] != "prod_1" {
		t.Fatalf("product identity fields = %+v, want prod_1", values)
	}
	if values["status"] != domain.ProductStatusActive {
		t.Fatalf("status = %v, want %s", values["status"], domain.ProductStatusActive)
	}
	if values["stock_status"] != domain.ProductStockStatusInStock {
		t.Fatalf("stock_status = %v, want %s", values["stock_status"], domain.ProductStockStatusInStock)
	}
	flags, ok := values["quality_flags"].(domain.ProductQualityFlags)
	if !ok || !flags.IsRecommendable || flags.IsDeleted {
		t.Fatalf("quality_flags = %#v, want recommendable and not deleted", values["quality_flags"])
	}
	if values["created_at"] != now {
		t.Fatalf("created_at = %v, want %v", values["created_at"], now)
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

func bsonDMap(values bson.D) map[string]any {
	result := make(map[string]any, len(values))
	for _, value := range values {
		result[value.Key] = value.Value
	}
	return result
}
