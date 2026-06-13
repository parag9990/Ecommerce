package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
)

func TestInteractionIngestionServiceStoresAndInvalidates(t *testing.T) {
	repo := &fakeInteractionRepository{}
	invalidator := &fakeInteractionInvalidator{}
	service, err := NewInteractionIngestionService(repo, nil, invalidator, nil, nil)
	if err != nil {
		t.Fatalf("NewInteractionIngestionService() error = %v", err)
	}

	result, err := service.IngestInteractions(context.Background(), []domain.UserInteraction{testInteraction("evt_1")})
	if err != nil {
		t.Fatalf("IngestInteractions() error = %v", err)
	}
	if result.Stored != 1 || result.Duplicates != 0 {
		t.Fatalf("result = %+v, want stored=1 duplicates=0", result)
	}
	if repo.calls != 1 {
		t.Fatalf("repo calls = %d, want 1", repo.calls)
	}
	if invalidator.calls != 1 {
		t.Fatalf("invalidator calls = %d, want 1", invalidator.calls)
	}
}

func TestInteractionIngestionServiceTreatsDuplicatesAsSuccess(t *testing.T) {
	repo := &fakeInteractionRepository{duplicate: true}
	invalidator := &fakeInteractionInvalidator{}
	service, err := NewInteractionIngestionService(repo, nil, invalidator, nil, nil)
	if err != nil {
		t.Fatalf("NewInteractionIngestionService() error = %v", err)
	}

	result, err := service.IngestInteractions(context.Background(), []domain.UserInteraction{testInteraction("evt_1")})
	if err != nil {
		t.Fatalf("IngestInteractions() error = %v", err)
	}
	if result.Stored != 0 || result.Duplicates != 1 {
		t.Fatalf("result = %+v, want stored=0 duplicates=1", result)
	}
	if invalidator.calls != 0 {
		t.Fatalf("invalidator calls = %d, want 0 for duplicate-only batch", invalidator.calls)
	}
}

func TestInteractionIngestionServiceCacheFailureIsNonFatal(t *testing.T) {
	service, err := NewInteractionIngestionService(
		&fakeInteractionRepository{},
		nil,
		&fakeInteractionInvalidator{err: errors.New("redis down")},
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewInteractionIngestionService() error = %v", err)
	}

	result, err := service.IngestInteractions(context.Background(), []domain.UserInteraction{testInteraction("evt_1")})
	if err != nil {
		t.Fatalf("IngestInteractions() error = %v", err)
	}
	if result.Stored != 1 {
		t.Fatalf("stored = %d, want 1", result.Stored)
	}
}

func TestInteractionIngestionServiceReturnsStorageError(t *testing.T) {
	service, err := NewInteractionIngestionService(
		&fakeInteractionRepository{err: errors.New("mongo down")},
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewInteractionIngestionService() error = %v", err)
	}

	_, err = service.IngestInteractions(context.Background(), []domain.UserInteraction{testInteraction("evt_1")})
	if !errors.Is(err, domain.ErrRecommendationStorage) {
		t.Fatalf("IngestInteractions() error = %v, want %v", err, domain.ErrRecommendationStorage)
	}
}

func TestInteractionIngestionServiceBuildsFeaturesForDuplicateRawRetry(t *testing.T) {
	repo := &fakeInteractionRepository{duplicate: true}
	builder := &fakeInteractionFeatureBuilder{result: domain.FeatureBuildResult{Applied: 1}}
	invalidator := &fakeInteractionInvalidator{}
	service, err := NewInteractionIngestionService(repo, builder, invalidator, nil, nil)
	if err != nil {
		t.Fatalf("NewInteractionIngestionService() error = %v", err)
	}

	result, err := service.IngestInteractions(context.Background(), []domain.UserInteraction{testInteraction("evt_retry")})
	if err != nil {
		t.Fatalf("IngestInteractions() error = %v", err)
	}
	if result.Stored != 0 || result.Duplicates != 1 || result.FeaturesApplied != 1 {
		t.Fatalf("result = %+v, want duplicate raw event with repaired feature apply", result)
	}
	if builder.calls != 1 || invalidator.calls != 1 {
		t.Fatalf("builder calls = %d invalidator calls = %d, want 1 and 1", builder.calls, invalidator.calls)
	}
}

type fakeInteractionRepository struct {
	calls     int
	duplicate bool
	err       error
}

func (f *fakeInteractionRepository) InsertInteraction(ctx context.Context, interaction domain.UserInteraction) (domain.InteractionInsertResult, error) {
	f.calls++
	if f.err != nil {
		return domain.InteractionInsertResult{}, f.err
	}
	return domain.InteractionInsertResult{Duplicate: f.duplicate}, nil
}

type fakeInteractionInvalidator struct {
	calls int
	err   error
}

type fakeInteractionFeatureBuilder struct {
	calls  int
	result domain.FeatureBuildResult
	err    error
}

func (f *fakeInteractionFeatureBuilder) ApplyInteractions(ctx context.Context, interactions []domain.UserInteraction) (domain.FeatureBuildResult, error) {
	f.calls++
	return f.result, f.err
}

func (f *fakeInteractionInvalidator) InvalidateInteractions(ctx context.Context, interactions []domain.UserInteraction) error {
	f.calls++
	return f.err
}

func testInteraction(eventID string) domain.UserInteraction {
	dedupeKey := domain.NewInteractionDedupeKey(eventID, domain.InteractionProductView, "prod_123", "var_1")
	return domain.UserInteraction{
		ID:                  domain.NewInteractionID(dedupeKey),
		DedupeKey:           dedupeKey,
		EventID:             eventID,
		SourceEventType:     domain.EventProductViewed,
		NormalizedEventType: domain.InteractionProductView,
		Version:             1,
		Producer:            "session-service",
		UserID:              "user_123",
		ProductID:           "prod_123",
		VariantID:           "var_1",
		Weight:              1,
		Quantity:            1,
		OccurredAt:          time.Date(2026, 5, 27, 10, 30, 0, 0, time.UTC),
		ReceivedAt:          time.Date(2026, 5, 27, 10, 30, 1, 0, time.UTC),
	}
}
