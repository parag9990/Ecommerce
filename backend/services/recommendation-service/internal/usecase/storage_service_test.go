package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
)

func TestStorageServicePlanUsesTaskTwoDefaults(t *testing.T) {
	service, err := NewStorageService(nil, nil, StorageServiceConfig{}, nil)
	if err != nil {
		t.Fatalf("NewStorageService() error = %v", err)
	}

	plan, err := service.StoragePlan(context.Background())
	if err != nil {
		t.Fatalf("StoragePlan() error = %v", err)
	}
	if plan.DatabaseName != domain.RecommendationDatabaseName {
		t.Fatalf("database = %s, want %s", plan.DatabaseName, domain.RecommendationDatabaseName)
	}
	if plan.CacheKeyPrefix != domain.DefaultCacheKeyPrefix {
		t.Fatalf("cache prefix = %s, want %s", plan.CacheKeyPrefix, domain.DefaultCacheKeyPrefix)
	}
}

func TestStorageServiceStatusReportsConfiguredDependencyFailure(t *testing.T) {
	service, err := NewStorageService(
		fakeStorageRepository{err: errors.New("mongo down")},
		nil,
		StorageServiceConfig{MongoConfigured: true, RedisConfigured: false},
		nil,
	)
	if err != nil {
		t.Fatalf("NewStorageService() error = %v", err)
	}

	status, err := service.Status(context.Background())
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status.Ready() {
		t.Fatal("status should not be ready when configured mongo is unavailable")
	}
	if status.MongoDB.State != StorageComponentUnavailable {
		t.Fatalf("mongo state = %s, want %s", status.MongoDB.State, StorageComponentUnavailable)
	}
	if status.Redis.State != StorageComponentNotConfigured {
		t.Fatalf("redis state = %s, want %s", status.Redis.State, StorageComponentNotConfigured)
	}
}

type fakeStorageRepository struct {
	err error
}

func (f fakeStorageRepository) Ping(ctx context.Context) error {
	return f.err
}

func (f fakeStorageRepository) EnsureIndexes(ctx context.Context) error {
	return f.err
}
