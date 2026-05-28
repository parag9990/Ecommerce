package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
)

func TestFeatureBuilderServiceBuildsValidatedBatch(t *testing.T) {
	repo := &fakeFeatureRepository{result: domain.FeatureBuildResult{Applied: 1, Stats: domain.FeatureJobStats{EventsRead: 1}}}
	service, err := NewFeatureBuilderService(repo, FeatureBuilderConfig{}, nil, nil)
	if err != nil {
		t.Fatalf("NewFeatureBuilderService() error = %v", err)
	}
	service.now = func() time.Time { return time.Date(2026, 5, 27, 11, 0, 0, 0, time.UTC) }

	result, err := service.ApplyInteractions(context.Background(), []domain.UserInteraction{testInteraction("evt_feature")})
	if err != nil {
		t.Fatalf("ApplyInteractions() error = %v", err)
	}
	if result.Applied != 1 || repo.calls != 1 {
		t.Fatalf("result = %+v calls = %d, want one applied batch", result, repo.calls)
	}
	if repo.batch.GuestProfileRetention != domain.DefaultGuestProfileRetention ||
		repo.batch.UserProductRetention != domain.DefaultUserProductRetention ||
		repo.batch.ProcessedEventRetention != domain.DefaultFeatureEventRetention {
		t.Fatalf("batch retention values = %+v, want domain defaults", repo.batch)
	}
}

func TestFeatureBuilderServiceWrapsStorageFailure(t *testing.T) {
	repo := &fakeFeatureRepository{err: errors.New("write failed")}
	service, err := NewFeatureBuilderService(repo, FeatureBuilderConfig{}, nil, nil)
	if err != nil {
		t.Fatalf("NewFeatureBuilderService() error = %v", err)
	}
	_, err = service.ApplyInteractions(context.Background(), []domain.UserInteraction{testInteraction("evt_feature")})
	if !errors.Is(err, domain.ErrRecommendationStorage) {
		t.Fatalf("ApplyInteractions() error = %v, want storage error", err)
	}
}

type fakeFeatureRepository struct {
	calls  int
	batch  domain.FeatureBatch
	result domain.FeatureBuildResult
	err    error
}

func (f *fakeFeatureRepository) ApplyFeatureBatch(ctx context.Context, batch domain.FeatureBatch) (domain.FeatureBuildResult, error) {
	f.calls++
	f.batch = batch
	return f.result, f.err
}

func (f *fakeFeatureRepository) ListPendingFeatureBatches(ctx context.Context, limit int) ([][]domain.UserInteraction, error) {
	return nil, f.err
}

func (f *fakeFeatureRepository) RebuildProductWindows(ctx context.Context, job domain.FeatureJobRun) (domain.FeatureJobStats, error) {
	return domain.FeatureJobStats{}, f.err
}
