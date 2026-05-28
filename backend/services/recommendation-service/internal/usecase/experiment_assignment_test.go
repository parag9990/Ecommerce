package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
)

func TestExperimentAssignmentServiceAssignsAndPersistsVariant(t *testing.T) {
	now := recommendationServingTestNow()
	repo := newFakeExperimentAssignmentRepository()
	service := newTestExperimentAssignmentService(t, repo)

	got, err := service.Assign(context.Background(), StrategyAssignmentInput{
		RequestID:          "req_1",
		DefaultStrategyID:  domain.StrategyPersonalizedBehavior,
		AssignmentSnapshot: now,
		Request: domain.RecommendationRequest{
			UserID:  "user_123",
			Context: domain.ContextHomeFeed,
			Type:    domain.RecommendationTypePersonalized,
			Limit:   12,
		},
	})
	if err != nil {
		t.Fatalf("Assign() error = %v", err)
	}
	if !got.Assigned || got.ExperimentID != "reco_home_strategy_2026_05" || got.StrategyID == "" {
		t.Fatalf("assignment = %+v, want persisted assignment", got)
	}
	if len(repo.assignments) != 1 {
		t.Fatalf("stored assignments = %d, want 1", len(repo.assignments))
	}
}

func TestExperimentAssignmentServiceReusesStoredAssignment(t *testing.T) {
	now := recommendationServingTestNow()
	repo := newFakeExperimentAssignmentRepository()
	stored := domain.NewExperimentAssignment(
		testExperimentDefinition(),
		domain.AssignmentIdentity{AssignmentKey: "user:user_123", UserID: "user_123"},
		domain.ExperimentVariant{VariantID: "treatment", StrategyID: domain.StrategyPersonalizedBehavior, Weight: 50},
		now.Add(-time.Minute),
		time.Hour,
	)
	repo.assignments[stored.ExperimentID+"|"+stored.AssignmentKey] = stored
	service := newTestExperimentAssignmentService(t, repo)

	got, err := service.Assign(context.Background(), StrategyAssignmentInput{
		DefaultStrategyID:  domain.StrategyPersonalizedBehavior,
		AssignmentSnapshot: now,
		Request: domain.RecommendationRequest{
			UserID:  "user_123",
			Context: domain.ContextHomeFeed,
			Type:    domain.RecommendationTypePersonalized,
		},
	})
	if err != nil {
		t.Fatalf("Assign() error = %v", err)
	}
	if got.VariantID != "treatment" || got.StrategyID != domain.StrategyPersonalizedBehavior {
		t.Fatalf("assignment = %+v, want stored treatment", got)
	}
	if repo.upserts != 0 {
		t.Fatalf("upserts = %d, want stored assignment reuse", repo.upserts)
	}
}

func TestExperimentAssignmentServiceReturnsDefaultWhenNoIdentity(t *testing.T) {
	service := newTestExperimentAssignmentService(t, newFakeExperimentAssignmentRepository())
	got, err := service.Assign(context.Background(), StrategyAssignmentInput{
		DefaultStrategyID:  domain.StrategyPersonalizedBehavior,
		AssignmentSnapshot: recommendationServingTestNow(),
		Request: domain.RecommendationRequest{
			Context: domain.ContextHomeFeed,
			Type:    domain.RecommendationTypePersonalized,
		},
	})
	if err != nil {
		t.Fatalf("Assign() error = %v", err)
	}
	if got.Assigned || got.StrategyID != domain.StrategyPersonalizedBehavior {
		t.Fatalf("assignment = %+v, want unassigned default strategy", got)
	}
}

func TestExperimentAssignmentServiceReturnsErrorWhenRepositoryUnavailable(t *testing.T) {
	service := newTestExperimentAssignmentService(t, nil)
	_, err := service.Assign(context.Background(), StrategyAssignmentInput{
		DefaultStrategyID:  domain.StrategyPersonalizedBehavior,
		AssignmentSnapshot: recommendationServingTestNow(),
		Request: domain.RecommendationRequest{
			UserID:  "user_123",
			Context: domain.ContextHomeFeed,
			Type:    domain.RecommendationTypePersonalized,
		},
	})
	if !errors.Is(err, domain.ErrRecommendationStorage) {
		t.Fatalf("Assign() error = %v, want %v", err, domain.ErrRecommendationStorage)
	}
}

type fakeExperimentAssignmentRepository struct {
	assignments map[string]domain.ExperimentAssignment
	getErr      error
	upsertErr   error
	upserts     int
}

func newFakeExperimentAssignmentRepository() *fakeExperimentAssignmentRepository {
	return &fakeExperimentAssignmentRepository{assignments: make(map[string]domain.ExperimentAssignment)}
}

func (f *fakeExperimentAssignmentRepository) GetExperimentAssignment(ctx context.Context, experimentID string, assignmentKey string, now time.Time) (domain.ExperimentAssignment, error) {
	if f.getErr != nil {
		return domain.ExperimentAssignment{}, f.getErr
	}
	assignment, ok := f.assignments[experimentID+"|"+assignmentKey]
	if !ok || !assignment.ExpiresAt.After(now) {
		return domain.ExperimentAssignment{}, domain.ErrExperimentAssignmentNotFound
	}
	return assignment, nil
}

func (f *fakeExperimentAssignmentRepository) UpsertExperimentAssignment(ctx context.Context, assignment domain.ExperimentAssignment) error {
	if f.upsertErr != nil {
		return f.upsertErr
	}
	f.upserts++
	f.assignments[assignment.ExperimentID+"|"+assignment.AssignmentKey] = assignment
	return nil
}

func newTestExperimentAssignmentService(t *testing.T, repo ExperimentAssignmentRepository) *ExperimentAssignmentService {
	t.Helper()
	service, err := NewExperimentAssignmentService(repo, ExperimentAssignmentConfig{
		Enabled:             true,
		Experiments:         []domain.ExperimentDefinition{testExperimentDefinition()},
		AssignmentTTL:       time.Hour,
		DefaultSalt:         "salt",
		MaxIdentifierLength: 128,
	}, nil, nil)
	if err != nil {
		t.Fatalf("NewExperimentAssignmentService() error = %v", err)
	}
	return service
}

func testExperimentDefinition() domain.ExperimentDefinition {
	return domain.ExperimentDefinition{
		ExperimentID: "reco_home_strategy_2026_05",
		Context:      domain.ContextHomeFeed,
		Type:         domain.RecommendationTypePersonalized,
		Status:       domain.ExperimentActive,
		Salt:         "salt",
		Variants: []domain.ExperimentVariant{
			{VariantID: "control", StrategyID: domain.StrategyTrendingRecentActivity, Weight: 50},
			{VariantID: "treatment", StrategyID: domain.StrategyPersonalizedBehavior, Weight: 50},
		},
	}
}
