package usecase

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
)

func TestRetentionUsecaseRunsCleanupInPolicyOrder(t *testing.T) {
	repo := &fakeRetentionRepository{
		legalHoldSkipped: 2,
		sessions:         3,
		journeys:         4,
		heatmapPoints:    5,
		heatmapMarkers:   6,
		aggregates:       7,
	}
	usecase := newTestRetentionUsecase(t, repo, nil)
	dryRun := true

	out, err := usecase.RunRetentionCleanup(context.Background(), RunRetentionCleanupInput{
		RequestID: "ret_test",
		DryRun:    &dryRun,
		Now:       time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("run cleanup: %v", err)
	}
	if !repo.dryRun || repo.sessionDays != 365 || repo.journeyDays != 365 || repo.heatmapDays != 730 || repo.aggregateYears != 5 {
		t.Fatalf("unexpected repository policy call: %+v", repo)
	}
	if out.SessionsAnonymized != 3 || out.JourneysAnonymized != 4 || out.HeatmapPointsPurged != 5 || out.AnalyticsAggregatesPurged != 7 || out.LegalHoldRecordsSkipped != 2 {
		t.Fatalf("unexpected cleanup result: %+v", out)
	}
}

func TestRetentionUsecaseDeleteUserSessionDataAnonymizesAndAudits(t *testing.T) {
	repo := &fakeRetentionRepository{
		rawDeleted:         10,
		sessionsAnonymized: 2,
		journeysAnonymized: 1,
	}
	active := &fakeRetentionActiveStore{
		byUser: map[string][]string{
			"user_123": {"sess_1", "sess_2"},
		},
		byAnon: map[string][]string{
			"anon_123": {"sess_2", "sess_3"},
		},
	}
	usecase := newTestRetentionUsecase(t, repo, active)

	out, err := usecase.DeleteUserSessionData(context.Background(), DeleteUserSessionDataInput{
		RequestID:    "del_test",
		UserID:       "user_123",
		AnonymousIDs: []string{"anon_123"},
		Reason:       "privacy_request",
		RequestedBy:  "admin_123",
	})
	if err != nil {
		t.Fatalf("delete user session data: %v", err)
	}
	if repo.audit.RequestID != "del_test" || repo.auditResult.RawEventsDeleted != 10 {
		t.Fatalf("expected audit write, got request=%+v result=%+v", repo.audit, repo.auditResult)
	}
	if out.SessionsAnonymized != 2 || out.JourneysAnonymized != 1 || out.RedisKeysDeleted != 3 {
		t.Fatalf("unexpected deletion result: %+v", out)
	}
	if len(active.deleted) != 3 {
		t.Fatalf("expected deduplicated redis deletes, got %+v", active.deleted)
	}
}

func newTestRetentionUsecase(t *testing.T, repo RetentionRepository, active ActiveSessionStore) *RetentionUsecase {
	t.Helper()
	u, err := NewRetentionUsecase(repo, active, RetentionConfig{
		RawEventTTLDays:              90,
		SessionMetadataRetentionDays: 365,
		JourneySummaryRetentionDays:  365,
		HeatmapRetentionDays:         730,
		AggregateRetentionYears:      5,
		WorkerBatchSize:              1000,
		WorkerInterval:               time.Hour,
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("new retention usecase: %v", err)
	}
	u.WithClock(retentionFixedClock{at: time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)})
	return u
}

type retentionFixedClock struct {
	at time.Time
}

func (c retentionFixedClock) Now() time.Time {
	return c.at
}

type fakeRetentionRepository struct {
	legalHoldSkipped  int64
	sessions          int64
	journeys          int64
	heatmapPoints     int64
	heatmapMarkers    int64
	aggregates        int64
	aggregatesWithPII int64

	rawDeleted         int64
	sessionsDeleted    int64
	sessionsAnonymized int64
	journeysDeleted    int64
	journeysAnonymized int64

	dryRun         bool
	sessionDays    int
	journeyDays    int
	heatmapDays    int
	aggregateYears int

	audit       domain.DeleteUserSessionDataRequest
	auditResult domain.DeleteUserSessionDataResult
}

func (f *fakeRetentionRepository) AnonymizeExpiredSessions(ctx context.Context, now time.Time, retentionDays int, limit int, dryRun bool) (int64, error) {
	f.sessionDays = retentionDays
	f.dryRun = dryRun
	return f.sessions, nil
}

func (f *fakeRetentionRepository) AnonymizeExpiredJourneys(ctx context.Context, now time.Time, retentionDays int, limit int, dryRun bool) (int64, error) {
	f.journeyDays = retentionDays
	return f.journeys, nil
}

func (f *fakeRetentionRepository) PurgeExpiredHeatmap(ctx context.Context, now time.Time, retentionDays int, limit int, dryRun bool) (int64, int64, error) {
	f.heatmapDays = retentionDays
	return f.heatmapPoints, f.heatmapMarkers, nil
}

func (f *fakeRetentionRepository) PurgeExpiredAggregates(ctx context.Context, now time.Time, retentionYears int, limit int, dryRun bool) (int64, error) {
	f.aggregateYears = retentionYears
	return f.aggregates, nil
}

func (f *fakeRetentionRepository) CountAggregatesWithPII(ctx context.Context) (int64, error) {
	return f.aggregatesWithPII, nil
}

func (f *fakeRetentionRepository) CountLegalHoldRetentionDue(ctx context.Context, now time.Time, sessionDays int, journeyDays int) (int64, error) {
	return f.legalHoldSkipped, nil
}

func (f *fakeRetentionRepository) DeleteRawEventsByIdentity(ctx context.Context, identity domain.SessionDataIdentity, dryRun bool) (int64, error) {
	return f.rawDeleted, nil
}

func (f *fakeRetentionRepository) DeleteSessionsByIdentity(ctx context.Context, identity domain.SessionDataIdentity, dryRun bool) (int64, error) {
	return f.sessionsDeleted, nil
}

func (f *fakeRetentionRepository) AnonymizeSessionsByIdentity(ctx context.Context, identity domain.SessionDataIdentity, at time.Time, dryRun bool) (int64, error) {
	return f.sessionsAnonymized, nil
}

func (f *fakeRetentionRepository) DeleteJourneysByIdentity(ctx context.Context, identity domain.SessionDataIdentity, dryRun bool) (int64, error) {
	return f.journeysDeleted, nil
}

func (f *fakeRetentionRepository) AnonymizeJourneysByIdentity(ctx context.Context, identity domain.SessionDataIdentity, at time.Time, dryRun bool) (int64, error) {
	return f.journeysAnonymized, nil
}

func (f *fakeRetentionRepository) WriteDeletionAudit(ctx context.Context, request domain.DeleteUserSessionDataRequest, result domain.DeleteUserSessionDataResult) error {
	f.audit = request
	f.auditResult = result
	return nil
}

func (f *fakeRetentionRepository) SetLegalHold(ctx context.Context, request domain.SetLegalHoldRequest) (domain.SetLegalHoldResult, error) {
	return domain.SetLegalHoldResult{SessionID: request.SessionID, LegalHold: request.Hold, SessionsMatched: 1, UpdatedAt: request.SetAt}, nil
}

type fakeRetentionActiveStore struct {
	byUser  map[string][]string
	byAnon  map[string][]string
	deleted []string
}

func (f *fakeRetentionActiveStore) Touch(ctx context.Context, snapshot domain.ActiveSessionSnapshot, ttl time.Duration) error {
	return nil
}

func (f *fakeRetentionActiveStore) Get(ctx context.Context, sessionID string) (domain.ActiveSessionSnapshot, error) {
	return domain.ActiveSessionSnapshot{}, domain.ErrSessionNotFound
}

func (f *fakeRetentionActiveStore) Delete(ctx context.Context, sessionID string) error {
	f.deleted = append(f.deleted, sessionID)
	return nil
}

func (f *fakeRetentionActiveStore) ListByUser(ctx context.Context, userID string) ([]string, error) {
	return append([]string(nil), f.byUser[userID]...), nil
}

func (f *fakeRetentionActiveStore) ListByAnonymousID(ctx context.Context, anonymousID string) ([]string, error) {
	return append([]string(nil), f.byAnon[anonymousID]...), nil
}
