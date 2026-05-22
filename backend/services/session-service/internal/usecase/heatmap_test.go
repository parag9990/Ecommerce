package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
)

func TestHeatmapUsecaseBuildsResponse(t *testing.T) {
	repo := &heatmapRepoFake{
		points: []domain.HeatmapPoint{
			{
				HeatmapType:    domain.HeatmapTypeClick,
				Path:           "/products/prod_123",
				NormalizedPath: "/products/:product_id",
				DeviceType:     domain.DeviceTypeMobile,
				ViewportBucket: "mobile_360_480",
				Day:            "2026-05-22",
				X:              55,
				Y:              70,
				Weight:         42,
				SampleEvents:   42,
				SchemaVersion:  domain.CurrentHeatmapSchemaVersion,
				FirstSeenAt:    time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC),
				LastSeenAt:     time.Date(2026, 5, 22, 11, 0, 0, 0, time.UTC),
				UpdatedAt:      time.Date(2026, 5, 22, 11, 1, 0, 0, time.UTC),
			},
		},
	}
	uc, err := NewHeatmapUsecase(repo, HeatmapQueryConfig{MaxDateRangeDays: 31, MaxPoints: 5000}, discardLogger())
	if err != nil {
		t.Fatalf("new heatmap usecase: %v", err)
	}

	out, err := uc.GetHeatmap(context.Background(), GetHeatmapInput{
		Path:       "/products/prod_123",
		DeviceType: "mobile",
		From:       time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC),
		To:         time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("get heatmap: %v", err)
	}
	if repo.filter.HeatmapType != domain.HeatmapTypeClick {
		t.Fatalf("expected default click heatmap type, got %q", repo.filter.HeatmapType)
	}
	if len(out.Points) != 1 || out.Points[0].Weight != 42 || out.MaxWeight != 42 || out.TotalEvents != 42 {
		t.Fatalf("unexpected heatmap output: %+v", out)
	}
}

func TestHeatmapUsecaseRejectsInvalidDateRange(t *testing.T) {
	uc, err := NewHeatmapUsecase(&heatmapRepoFake{}, HeatmapQueryConfig{MaxDateRangeDays: 1, MaxPoints: 5000}, discardLogger())
	if err != nil {
		t.Fatalf("new heatmap usecase: %v", err)
	}

	_, err = uc.GetHeatmap(context.Background(), GetHeatmapInput{
		Path:       "/products/prod_123",
		DeviceType: "mobile",
		From:       time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC),
		To:         time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC),
	})
	if !errors.Is(err, ErrInvalidSessionInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestHeatmapAggregationUsecaseAggregatesValidEventsAndSkipsInvalid(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	clickPath := "/products/prod_123"
	scrollPath := "/products/prod_456"
	events := &heatmapEventRepoFake{
		events: []domain.SessionEvent{
			{
				EventID:     "evt_click_1",
				SessionID:   "sess_1",
				AnonymousID: "anon_1",
				EventType:   domain.EventClick,
				Path:        &clickPath,
				Properties: map[string]any{
					"x":               210,
					"y":               600,
					"viewport_width":  390,
					"viewport_height": 844,
				},
				Device:     &domain.Device{Type: domain.DeviceTypeMobile},
				OccurredAt: now.Add(-10 * time.Minute),
				ReceivedAt: now.Add(-10*time.Minute + time.Second),
			},
			{
				EventID:     "evt_scroll_1",
				SessionID:   "sess_2",
				AnonymousID: "anon_2",
				EventType:   domain.EventScroll,
				Path:        &scrollPath,
				Properties: map[string]any{
					"depth_percent": 92,
				},
				Device:     &domain.Device{Type: domain.DeviceTypeDesktop},
				OccurredAt: now.Add(-9 * time.Minute),
				ReceivedAt: now.Add(-9*time.Minute + time.Second),
			},
			{
				EventID:     "evt_click_bad",
				SessionID:   "sess_3",
				AnonymousID: "anon_3",
				EventType:   domain.EventClick,
				Path:        &clickPath,
				Properties: map[string]any{
					"x": 1,
				},
				OccurredAt: now.Add(-8 * time.Minute),
			},
		},
	}
	points := &heatmapRepoFake{}
	uc, err := NewHeatmapAggregationUsecase(events, points, HeatmapAggregationConfig{BatchSize: 10, ClickBucketSize: 5}, discardLogger())
	if err != nil {
		t.Fatalf("new heatmap aggregation usecase: %v", err)
	}
	uc.WithClock(fixedClock{now: now})

	out, err := uc.AggregateHeatmap(context.Background(), AggregateHeatmapInput{
		From:           now.Add(-time.Hour),
		To:             now,
		SaveCheckpoint: true,
	})
	if err != nil {
		t.Fatalf("aggregate heatmap: %v", err)
	}
	if out.EventsScanned != 3 || out.PointsUpserted != 2 || out.EventsSkipped != 1 || !out.CheckpointSaved {
		t.Fatalf("unexpected aggregation output: %+v", out)
	}
	if len(points.upserted) != 2 {
		t.Fatalf("expected two upserted points, got %+v", points.upserted)
	}
	if points.upserted[0].point.HeatmapType != domain.HeatmapTypeClick || points.upserted[1].point.HeatmapType != domain.HeatmapTypeScroll {
		t.Fatalf("unexpected upserted point types: %+v", points.upserted)
	}
	if points.savedCheckpoint.LastProcessedAt != now {
		t.Fatalf("unexpected checkpoint: %+v", points.savedCheckpoint)
	}
}

type heatmapRepoFake struct {
	points          []domain.HeatmapPoint
	filter          domain.HeatmapFilter
	limit           int
	listErr         error
	upserted        []heatmapPointUpsert
	upsertErr       error
	checkpoint      domain.HeatmapAggregationCheckpoint
	checkpointErr   error
	savedCheckpoint domain.HeatmapAggregationCheckpoint
	saveErr         error
}

type heatmapPointUpsert struct {
	point     domain.HeatmapPoint
	sessionID string
}

func (r *heatmapRepoFake) UpsertHeatmapPoint(ctx context.Context, point domain.HeatmapPoint, sessionID string) error {
	if r.upsertErr != nil {
		return r.upsertErr
	}
	r.upserted = append(r.upserted, heatmapPointUpsert{point: point, sessionID: sessionID})
	return nil
}

func (r *heatmapRepoFake) ListHeatmapPoints(ctx context.Context, filter domain.HeatmapFilter, limit int) ([]domain.HeatmapPoint, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	r.filter = filter
	r.limit = limit
	return r.points, nil
}

func (r *heatmapRepoFake) FindHeatmapCheckpoint(ctx context.Context, workerName string) (domain.HeatmapAggregationCheckpoint, error) {
	if r.checkpointErr != nil {
		return domain.HeatmapAggregationCheckpoint{}, r.checkpointErr
	}
	return r.checkpoint, nil
}

func (r *heatmapRepoFake) SaveHeatmapCheckpoint(ctx context.Context, checkpoint domain.HeatmapAggregationCheckpoint) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.savedCheckpoint = checkpoint
	return nil
}

type heatmapEventRepoFake struct {
	events []domain.SessionEvent
	err    error
}

func (r *heatmapEventRepoFake) ListHeatmapEvents(ctx context.Context, from time.Time, to time.Time, limit int, cursorAfter *domain.HeatmapEventCursor) ([]domain.SessionEvent, error) {
	if r.err != nil {
		return nil, r.err
	}
	if cursorAfter == nil {
		if len(r.events) <= limit {
			return r.events, nil
		}
		return r.events[:limit], nil
	}
	return nil, nil
}
