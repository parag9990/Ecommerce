package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
)

const (
	defaultHeatmapBatchSize                     = 1000
	defaultHeatmapWorkerName                    = "heatmap_aggregator"
	defaultHeatmapAggregationInitialLookback    = time.Hour
	defaultHeatmapAggregationCheckpointLookback = 5 * time.Minute
)

type HeatmapAggregationConfig struct {
	BatchSize          int
	ClickBucketSize    int
	WorkerName         string
	InitialLookback    time.Duration
	CheckpointLookback time.Duration
}

type AggregateHeatmapInput struct {
	From           time.Time
	To             time.Time
	UseCheckpoint  bool
	SaveCheckpoint bool
}

type AggregateHeatmapOutput struct {
	From            time.Time
	To              time.Time
	EventsScanned   int
	PointsUpserted  int
	EventsSkipped   int
	CheckpointSaved bool
}

type HeatmapAggregationUsecase struct {
	events HeatmapEventRepository
	points HeatmapRepository
	cfg    HeatmapAggregationConfig
	logger *slog.Logger
	clock  Clock
}

func NewHeatmapAggregationUsecase(events HeatmapEventRepository, points HeatmapRepository, cfg HeatmapAggregationConfig, logger *slog.Logger) (*HeatmapAggregationUsecase, error) {
	if events == nil {
		return nil, errors.New("heatmap event repository is required")
	}
	if points == nil {
		return nil, errors.New("heatmap repository is required")
	}
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &HeatmapAggregationUsecase{
		events: events,
		points: points,
		cfg:    cfg,
		logger: logger,
		clock:  realClock{},
	}, nil
}

func (u *HeatmapAggregationUsecase) WithClock(clock Clock) {
	if clock != nil {
		u.clock = clock
	}
}

func (u *HeatmapAggregationUsecase) AggregateHeatmap(ctx context.Context, input AggregateHeatmapInput) (AggregateHeatmapOutput, error) {
	if err := ctx.Err(); err != nil {
		return AggregateHeatmapOutput{}, err
	}
	started := time.Now()
	from, to, err := u.normalizeWindow(ctx, input)
	if err != nil {
		return AggregateHeatmapOutput{}, err
	}

	output := AggregateHeatmapOutput{From: from, To: to}
	var cursor *domain.HeatmapEventCursor
	for {
		events, err := u.events.ListHeatmapEvents(ctx, from, to, u.cfg.BatchSize, cursor)
		if err != nil {
			return output, fmt.Errorf("%w: list heatmap events: %w", ErrHeatmapStorageUnavailable, err)
		}
		if len(events) == 0 {
			break
		}
		for _, event := range events {
			output.EventsScanned++
			point, err := domain.HeatmapPointFromEvent(event, u.clock.Now().UTC(), u.cfg.ClickBucketSize)
			if err != nil {
				output.EventsSkipped++
				u.logger.DebugContext(ctx, "session.heatmap.event_skipped",
					slog.String("event_id", event.EventID),
					slog.String("session_id", event.SessionID),
					slog.String("event_type", string(event.EventType)),
					slog.String("error", err.Error()),
				)
				continue
			}
			processed, err := u.points.UpsertHeatmapPoint(ctx, point, event.EventID, event.SessionID)
			if err != nil {
				return output, fmt.Errorf("%w: upsert heatmap point: %w", ErrHeatmapStorageUnavailable, err)
			}
			if processed {
				output.PointsUpserted++
			}
		}
		last := events[len(events)-1].Normalize()
		cursor = &domain.HeatmapEventCursor{
			OccurredAt: last.OccurredAt,
			EventID:    last.EventID,
		}
		if len(events) < u.cfg.BatchSize {
			break
		}
	}

	if input.SaveCheckpoint {
		if err := u.points.SaveHeatmapCheckpoint(ctx, domain.HeatmapAggregationCheckpoint{
			WorkerName:      u.cfg.WorkerName,
			LastProcessedAt: to,
			UpdatedAt:       u.clock.Now().UTC(),
		}); err != nil {
			return output, fmt.Errorf("%w: save heatmap checkpoint: %w", ErrHeatmapStorageUnavailable, err)
		}
		output.CheckpointSaved = true
	}
	u.logger.InfoContext(ctx, "session.heatmap.aggregated",
		slog.String("from", output.From.Format(time.RFC3339Nano)),
		slog.String("to", output.To.Format(time.RFC3339Nano)),
		slog.Int("events_scanned", output.EventsScanned),
		slog.Int("points_upserted", output.PointsUpserted),
		slog.Int("events_skipped", output.EventsSkipped),
		slog.Bool("checkpoint_saved", output.CheckpointSaved),
		slog.Int64("duration_ms", time.Since(started).Milliseconds()),
	)
	return output, nil
}

func (u *HeatmapAggregationUsecase) normalizeWindow(ctx context.Context, input AggregateHeatmapInput) (time.Time, time.Time, error) {
	now := u.clock.Now().UTC()
	to := input.To.UTC()
	if to.IsZero() {
		to = now
	}
	from := input.From.UTC()
	if input.UseCheckpoint {
		checkpoint, err := u.points.FindHeatmapCheckpoint(ctx, u.cfg.WorkerName)
		if err == nil && !checkpoint.LastProcessedAt.IsZero() {
			from = checkpoint.LastProcessedAt.Add(-u.cfg.CheckpointLookback).UTC()
		} else if err != nil && !errors.Is(err, domain.ErrHeatmapNotFound) {
			return time.Time{}, time.Time{}, fmt.Errorf("%w: find heatmap checkpoint: %w", ErrHeatmapStorageUnavailable, err)
		}
	}
	if from.IsZero() {
		from = to.Add(-u.cfg.InitialLookback).UTC()
	}
	if !from.Before(to) {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: heatmap aggregation from must be before to", ErrInvalidSessionInput)
	}
	return from, to, nil
}

func (c HeatmapAggregationConfig) Validate() error {
	if c.BatchSize <= 0 {
		return errors.New("SESSION_HEATMAP_AGGREGATION_BATCH_SIZE must be greater than zero")
	}
	if c.ClickBucketSize <= 0 || c.ClickBucketSize > 100 {
		return errors.New("SESSION_HEATMAP_CLICK_BUCKET_SIZE must be between 1 and 100")
	}
	if strings.TrimSpace(c.WorkerName) == "" {
		return errors.New("SESSION_HEATMAP_AGGREGATION_WORKER_NAME cannot be empty")
	}
	if c.InitialLookback <= 0 {
		return errors.New("SESSION_HEATMAP_AGGREGATION_INITIAL_LOOKBACK must be greater than zero")
	}
	if c.CheckpointLookback < 0 {
		return errors.New("SESSION_HEATMAP_AGGREGATION_CHECKPOINT_LOOKBACK cannot be negative")
	}
	return nil
}

func (c HeatmapAggregationConfig) withDefaults() HeatmapAggregationConfig {
	if c.BatchSize == 0 {
		c.BatchSize = defaultHeatmapBatchSize
	}
	if c.ClickBucketSize == 0 {
		c.ClickBucketSize = domain.DefaultHeatmapClickBucketSize
	}
	if strings.TrimSpace(c.WorkerName) == "" {
		c.WorkerName = defaultHeatmapWorkerName
	}
	if c.InitialLookback == 0 {
		c.InitialLookback = defaultHeatmapAggregationInitialLookback
	}
	if c.CheckpointLookback == 0 {
		c.CheckpointLookback = defaultHeatmapAggregationCheckpointLookback
	}
	return c
}
