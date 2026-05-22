package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
)

func TestAnalyticsUsecaseGetLiveMetrics(t *testing.T) {
	live := &analyticsLiveRepoFake{activeUsers: 10, activeSessions: 12, eventsPerMinute: 7.5}
	uc, err := NewAnalyticsUsecase(live, &analyticsSessionRepoFake{}, &analyticsAggregateRepoFake{}, AnalyticsConfig{}, discardLogger())
	if err != nil {
		t.Fatalf("new analytics usecase: %v", err)
	}
	uc.WithClock(fixedClock{now: time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)})

	out, err := uc.GetLiveMetrics(context.Background(), GetLiveMetricsInput{})
	if err != nil {
		t.Fatalf("get live metrics: %v", err)
	}
	if out.ActiveUsers != 10 || out.ActiveSessions != 12 || out.EventsPerMinute != 7.5 || out.WindowSeconds != 300 {
		t.Fatalf("unexpected live metrics: %+v", out)
	}
}

func TestAnalyticsUsecaseListSessionsValidatesPageSize(t *testing.T) {
	uc, err := NewAnalyticsUsecase(&analyticsLiveRepoFake{}, &analyticsSessionRepoFake{}, &analyticsAggregateRepoFake{}, AnalyticsConfig{DefaultPageSize: 10, MaxPageSize: 10}, discardLogger())
	if err != nil {
		t.Fatalf("new analytics usecase: %v", err)
	}

	_, err = uc.ListSessions(context.Background(), ListSessionsInput{Page: 1, PageSize: 11})
	if !errors.Is(err, ErrInvalidSessionInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestAnalyticsUsecaseListSessionsUsesDefaults(t *testing.T) {
	now := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	repo := &analyticsSessionRepoFake{
		sessions: []domain.Session{analyticsTestSession(now)},
		total:    1,
	}
	uc, err := NewAnalyticsUsecase(&analyticsLiveRepoFake{}, repo, &analyticsAggregateRepoFake{}, AnalyticsConfig{DefaultPageSize: 25, MaxPageSize: 50}, discardLogger())
	if err != nil {
		t.Fatalf("new analytics usecase: %v", err)
	}
	uc.WithClock(fixedClock{now: now})

	out, err := uc.ListSessions(context.Background(), ListSessionsInput{DeviceType: string(domain.DeviceTypeDesktop)})
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if repo.filter.Page != 1 || repo.filter.PageSize != 25 || repo.filter.DeviceType != domain.DeviceTypeDesktop {
		t.Fatalf("unexpected filter: %+v", repo.filter)
	}
	if out.Total != 1 || len(out.Sessions) != 1 {
		t.Fatalf("unexpected output: %+v", out)
	}
}

func TestAnalyticsUsecaseFunnelUsesAggregateFirst(t *testing.T) {
	from := time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC)
	agg := &analyticsAggregateRepoFake{
		aggregateSteps: []domain.FunnelStep{
			{Name: "product_view", EventType: domain.EventProductView, UniqueSessions: 100, Count: 120},
			{Name: "add_to_cart", EventType: domain.EventAddToCart, UniqueSessions: 25, Count: 30},
		},
	}
	uc, err := NewAnalyticsUsecase(&analyticsLiveRepoFake{}, &analyticsSessionRepoFake{}, agg, AnalyticsConfig{}, discardLogger())
	if err != nil {
		t.Fatalf("new analytics usecase: %v", err)
	}

	out, err := uc.GetFunnelReport(context.Background(), GetFunnelReportInput{
		From:  from,
		To:    from.Add(24 * time.Hour),
		Steps: []string{"product_view", "add_to_cart"},
	})
	if err != nil {
		t.Fatalf("get funnel: %v", err)
	}
	if agg.rawCalled {
		t.Fatal("raw fallback should not be called when aggregate exists")
	}
	if out.Source != "mongo_aggregate" || out.Steps[1].ConversionFromPrevious != 25 {
		t.Fatalf("unexpected funnel output: %+v", out)
	}
}

func TestAnalyticsUsecaseFunnelUsesRawFallbackForShortRange(t *testing.T) {
	from := time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC)
	agg := &analyticsAggregateRepoFake{
		rawSteps: []domain.FunnelStep{
			{Name: "product_view", EventType: domain.EventProductView, UniqueSessions: 10},
			{Name: "add_to_cart", EventType: domain.EventAddToCart, UniqueSessions: 5},
		},
	}
	uc, err := NewAnalyticsUsecase(&analyticsLiveRepoFake{}, &analyticsSessionRepoFake{}, agg, AnalyticsConfig{RawFallbackRange: time.Hour}, discardLogger())
	if err != nil {
		t.Fatalf("new analytics usecase: %v", err)
	}

	out, err := uc.GetFunnelReport(context.Background(), GetFunnelReportInput{
		From:  from,
		To:    from.Add(time.Hour),
		Steps: []string{"product_view", "add_to_cart"},
	})
	if err != nil {
		t.Fatalf("get funnel: %v", err)
	}
	if !agg.rawCalled || out.Source != "raw_fallback" || out.OverallConversion != 50 {
		t.Fatalf("unexpected fallback output: %+v rawCalled=%v", out, agg.rawCalled)
	}
}

func TestAnalyticsUsecaseFunnelRejectsLargeRawFallback(t *testing.T) {
	from := time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC)
	uc, err := NewAnalyticsUsecase(&analyticsLiveRepoFake{}, &analyticsSessionRepoFake{}, &analyticsAggregateRepoFake{}, AnalyticsConfig{RawFallbackRange: time.Hour}, discardLogger())
	if err != nil {
		t.Fatalf("new analytics usecase: %v", err)
	}

	_, err = uc.GetFunnelReport(context.Background(), GetFunnelReportInput{
		From:  from,
		To:    from.Add(2 * time.Hour),
		Steps: []string{"product_view", "add_to_cart"},
	})
	if !errors.Is(err, ErrAnalyticsAggregateNotReady) {
		t.Fatalf("expected aggregate not ready, got %v", err)
	}
}

type analyticsLiveRepoFake struct {
	activeUsers     int64
	activeSessions  int64
	eventsPerMinute float64
	err             error
}

func (r *analyticsLiveRepoFake) CountActiveUsers(ctx context.Context, window time.Duration) (int64, error) {
	return r.activeUsers, r.err
}

func (r *analyticsLiveRepoFake) CountActiveSessions(ctx context.Context, window time.Duration) (int64, error) {
	return r.activeSessions, r.err
}

func (r *analyticsLiveRepoFake) EventsPerMinute(ctx context.Context, window time.Duration) (float64, error) {
	return r.eventsPerMinute, r.err
}

type analyticsSessionRepoFake struct {
	filter   domain.SessionListFilter
	sessions []domain.Session
	total    int64
	err      error
}

func (r *analyticsSessionRepoFake) ListSessions(ctx context.Context, filter domain.SessionListFilter) ([]domain.Session, int64, error) {
	r.filter = filter
	if r.err != nil {
		return nil, 0, r.err
	}
	return append([]domain.Session(nil), r.sessions...), r.total, nil
}

type analyticsAggregateRepoFake struct {
	filter         domain.FunnelReportFilter
	aggregateSteps []domain.FunnelStep
	rawSteps       []domain.FunnelStep
	aggregateErr   error
	rawErr         error
	rawCalled      bool
}

func (r *analyticsAggregateRepoFake) GetFunnelAggregate(ctx context.Context, filter domain.FunnelReportFilter) ([]domain.FunnelStep, error) {
	r.filter = filter
	if r.aggregateErr != nil {
		return nil, r.aggregateErr
	}
	return append([]domain.FunnelStep(nil), r.aggregateSteps...), nil
}

func (r *analyticsAggregateRepoFake) BuildFunnelFromRawEvents(ctx context.Context, filter domain.FunnelReportFilter) ([]domain.FunnelStep, error) {
	r.rawCalled = true
	if r.rawErr != nil {
		return nil, r.rawErr
	}
	return append([]domain.FunnelStep(nil), r.rawSteps...), nil
}

func analyticsTestSession(now time.Time) domain.Session {
	return domain.Session{
		SessionID:     "sess_analytics",
		AnonymousID:   "anon_analytics",
		SchemaVersion: domain.CurrentSessionSchemaVersion,
		Status:        domain.SessionStatusActive,
		Channel:       domain.ChannelUserAppWeb,
		EntryPage:     "/",
		UserAgent:     "Mozilla/5.0",
		Device:        domain.Device{Type: domain.DeviceTypeDesktop},
		IPHash:        "7c9e6679f7425d42a01e0cfbdac7c7b39f4f5f9c31f5b22d9efb9b9f7436e75a",
		IPVersion:     domain.IPVersionIPv4,
		Geo:           domain.Geo{Source: domain.GeoSourceUnknown},
		RiskLevel:     domain.RiskLevelLow,
		RiskReasons:   []string{},
		StartedAt:     now.Add(-time.Hour),
		LastSeenAt:    now,
		CreatedAt:     now.Add(-time.Hour),
		UpdatedAt:     now,
	}
}
