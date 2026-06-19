package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
)

func TestPrivacyUsecaseGetSettingsAddsPermissions(t *testing.T) {
	uc := newTestPrivacyUsecase(t, &fakePrivacyRepo{}, nil)

	settings, err := uc.GetPrivacySettings(context.Background(), domain.Actor{
		ID:    "admin_1",
		Roles: []string{"admin"},
	})
	if err != nil {
		t.Fatalf("GetPrivacySettings returned error: %v", err)
	}

	if settings.Masking.UserIDMode != domain.MaskingModeMasked {
		t.Fatalf("expected masked default, got %q", settings.Masking.UserIDMode)
	}
	if settings.Permissions.CanUpdateMasking {
		t.Fatal("plain admin should not update masking")
	}
}

func TestPrivacyUsecaseUpdateRetentionRequiresSuperadminAndReason(t *testing.T) {
	uc := newTestPrivacyUsecase(t, &fakePrivacyRepo{}, nil)

	_, err := uc.UpdateRetentionSettings(context.Background(), UpdateRetentionSettingsInput{
		Actor: domain.Actor{ID: "admin_1", Roles: []string{"admin"}},
		Retention: domain.RetentionSettings{
			RawEventsDays:             90,
			JourneySummariesDays:      365,
			HeatmapAggregatesDays:     365,
			AnalyticsAggregatesMonths: 36,
			ActiveSessionTTLMinutes:   45,
			DeletionRequestLogDays:    730,
		},
		Reason: "Support ticket SUP-123",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}

	_, err = uc.UpdateRetentionSettings(context.Background(), UpdateRetentionSettingsInput{
		Actor: domain.Actor{ID: "root_1", Roles: []string{"superadmin"}},
		Retention: domain.RetentionSettings{
			RawEventsDays:             90,
			JourneySummariesDays:      365,
			HeatmapAggregatesDays:     365,
			AnalyticsAggregatesMonths: 36,
			ActiveSessionTTLMinutes:   45,
			DeletionRequestLogDays:    730,
		},
		Reason: "short",
	})
	if !errors.Is(err, domain.ErrReasonRequired) {
		t.Fatalf("expected reason required, got %v", err)
	}
}

func TestPrivacyUsecaseCreateDeletionRequestMasksHashesAndAudits(t *testing.T) {
	repo := &fakePrivacyRepo{
		preview: domain.DeletionPreview{
			MatchedSessions:         2,
			MatchedEvents:           40,
			MatchedJourneySummaries: 2,
		},
		applied: domain.DeletionPreview{
			MatchedSessions:         2,
			MatchedEvents:           40,
			MatchedJourneySummaries: 2,
		},
	}
	active := &fakeActiveSessionRepo{matched: 1, deleted: 1}
	uc := newTestPrivacyUsecase(t, repo, active)

	request, err := uc.CreateDeletionRequest(context.Background(), CreateDeletionRequestInput{
		Actor: domain.Actor{
			ID:    "ops_1",
			Roles: []string{"operations_admin"},
		},
		Target: domain.DeletionTarget{
			Type:  domain.DeletionTargetUserID,
			Value: "user_123456789",
		},
		Reason:    "Support ticket SUP-123 user requested analytics deletion",
		Confirmed: true,
		RequestID: "req_123",
	})
	if err != nil {
		t.Fatalf("CreateDeletionRequest returned error: %v", err)
	}

	if request.Status != domain.DeletionStatusCompleted {
		t.Fatalf("expected completed status, got %q", request.Status)
	}
	if request.TargetValueMasked != "user_1...6789" {
		t.Fatalf("unexpected masked value %q", request.TargetValueMasked)
	}
	if request.TargetHash == "" || strings.Contains(request.TargetHash, "user_123456789") {
		t.Fatalf("target hash should be opaque, got %q", request.TargetHash)
	}
	if len(repo.requests) != 1 {
		t.Fatalf("expected deletion request to be stored")
	}
	if len(repo.auditEvents) != 1 || repo.auditEvents[0].Action != "session_analytics.deletion_requested" {
		t.Fatalf("expected deletion audit event, got %#v", repo.auditEvents)
	}
}

func TestPrivacyUsecasePreviewDeletionRejectsInvalidTarget(t *testing.T) {
	uc := newTestPrivacyUsecase(t, &fakePrivacyRepo{}, nil)

	_, err := uc.PreviewDeletion(context.Background(), PreviewDeletionInput{
		Actor:  domain.Actor{ID: "ops_1", Roles: []string{"operations_admin"}},
		Target: domain.DeletionTarget{Type: "email", Value: "buyer@example.com"},
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func newTestPrivacyUsecase(t *testing.T, repo *fakePrivacyRepo, active ActiveSessionRepository) *PrivacyUsecase {
	t.Helper()
	uc, err := NewPrivacyUsecase(repo, active, PrivacyConfig{
		HashPepper:        "unit-test-pepper",
		DeletionListLimit: 50,
	}, nil)
	if err != nil {
		t.Fatalf("NewPrivacyUsecase returned error: %v", err)
	}
	uc.clock = fixedClock{now: time.Date(2026, 5, 28, 4, 0, 0, 0, time.UTC)}
	return uc
}

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

type fakePrivacyRepo struct {
	settings    domain.PrivacySettings
	hasSettings bool
	preview     domain.DeletionPreview
	applied     domain.DeletionPreview
	requests    []domain.DeletionRequest
	auditEvents []domain.AuditEvent
}

func (r *fakePrivacyRepo) GetPrivacySettings(context.Context) (domain.PrivacySettings, error) {
	if !r.hasSettings {
		return domain.PrivacySettings{}, domain.ErrNotFound
	}
	return r.settings, nil
}

func (r *fakePrivacyRepo) UpsertPrivacySettings(_ context.Context, settings domain.PrivacySettings) error {
	r.settings = settings
	r.hasSettings = true
	return nil
}

func (r *fakePrivacyRepo) PreviewDeletion(_ context.Context, target domain.DeletionTarget) (domain.DeletionPreview, error) {
	preview := r.preview
	preview.TargetType = target.Type
	preview.TargetValueMasked = domain.MaskIdentifier(target.Value)
	return preview, nil
}

func (r *fakePrivacyRepo) ApplyDeletion(_ context.Context, target domain.DeletionTarget) (domain.DeletionPreview, error) {
	applied := r.applied
	applied.TargetType = target.Type
	applied.TargetValueMasked = domain.MaskIdentifier(target.Value)
	return applied, nil
}

func (r *fakePrivacyRepo) CreateDeletionRequest(_ context.Context, request domain.DeletionRequest) error {
	r.requests = append(r.requests, request)
	return nil
}

func (r *fakePrivacyRepo) UpdateDeletionRequestStatus(_ context.Context, requestID string, status domain.DeletionRequestStatus, completedAt *time.Time, message string) error {
	for index := range r.requests {
		if r.requests[index].RequestID == requestID {
			r.requests[index].Status = status
			r.requests[index].CompletedAt = completedAt
			r.requests[index].Error = message
			return nil
		}
	}
	return domain.ErrNotFound
}

func (r *fakePrivacyRepo) ListDeletionRequests(context.Context, int) ([]domain.DeletionRequest, error) {
	return r.requests, nil
}

func (r *fakePrivacyRepo) CreateAuditEvent(_ context.Context, event domain.AuditEvent) error {
	r.auditEvents = append(r.auditEvents, event)
	return nil
}

type fakeActiveSessionRepo struct {
	matched int64
	deleted int64
}

func (r *fakeActiveSessionRepo) PreviewDeletion(context.Context, domain.DeletionTarget) (int64, error) {
	return r.matched, nil
}

func (r *fakeActiveSessionRepo) DeleteMatching(context.Context, domain.DeletionTarget) (int64, error) {
	return r.deleted, nil
}
