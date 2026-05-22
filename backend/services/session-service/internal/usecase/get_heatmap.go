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
	defaultHeatmapMaxDateRangeDays = 31
	defaultHeatmapMaxPoints        = 5000
)

type HeatmapQueryConfig struct {
	MaxDateRangeDays int
	MaxPoints        int
}

type GetHeatmapInput struct {
	HeatmapType    string
	Path           string
	DeviceType     string
	From           time.Time
	To             time.Time
	ViewportBucket string
	Limit          int
}

type HeatmapPointOutput struct {
	X      int
	Y      int
	Weight int
}

type HeatmapOutput struct {
	Path        string
	DeviceType  domain.DeviceType
	HeatmapType domain.HeatmapType
	From        string
	To          string
	Points      []HeatmapPointOutput
	MaxWeight   int
	TotalEvents int
}

type HeatmapUsecase struct {
	repo   HeatmapRepository
	cfg    HeatmapQueryConfig
	logger *slog.Logger
}

func NewHeatmapUsecase(repo HeatmapRepository, cfg HeatmapQueryConfig, logger *slog.Logger) (*HeatmapUsecase, error) {
	if repo == nil {
		return nil, errors.New("heatmap repository is required")
	}
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &HeatmapUsecase{
		repo:   repo,
		cfg:    cfg,
		logger: logger,
	}, nil
}

func (u *HeatmapUsecase) GetHeatmap(ctx context.Context, input GetHeatmapInput) (HeatmapOutput, error) {
	if err := ctx.Err(); err != nil {
		return HeatmapOutput{}, err
	}
	started := time.Now()
	filter, limit, err := u.normalizeInput(input)
	if err != nil {
		return HeatmapOutput{}, err
	}

	points, err := u.repo.ListHeatmapPoints(ctx, filter, limit)
	if err != nil {
		return HeatmapOutput{}, fmt.Errorf("%w: list heatmap points: %w", ErrHeatmapStorageUnavailable, err)
	}
	output := buildHeatmapOutput(filter, points)
	u.logger.InfoContext(ctx, "session.heatmap.fetched",
		slog.String("path", filter.Path),
		slog.String("device_type", string(filter.DeviceType)),
		slog.String("heatmap_type", string(filter.HeatmapType)),
		slog.String("from", filter.FromDay),
		slog.String("to", filter.ToDay),
		slog.Int("points_count", len(output.Points)),
		slog.Int64("duration_ms", time.Since(started).Milliseconds()),
	)
	return output, nil
}

func (u *HeatmapUsecase) normalizeInput(input GetHeatmapInput) (domain.HeatmapFilter, int, error) {
	heatmapType := domain.NormalizeHeatmapType(input.HeatmapType)
	path := strings.TrimSpace(input.Path)
	deviceType := domain.DeviceType(strings.TrimSpace(input.DeviceType))
	from := input.From.UTC()
	to := input.To.UTC()
	filter := domain.HeatmapFilter{
		HeatmapType:    heatmapType,
		Path:           path,
		DeviceType:     deviceType,
		ViewportBucket: strings.TrimSpace(input.ViewportBucket),
		FromDay:        domain.DayString(from),
		ToDay:          domain.DayString(to),
	}
	if input.From.IsZero() {
		filter.FromDay = ""
	}
	if input.To.IsZero() {
		filter.ToDay = ""
	}
	if err := filter.Validate(); err != nil {
		return domain.HeatmapFilter{}, 0, fmt.Errorf("%w: %w", ErrInvalidSessionInput, err)
	}
	if from.After(to) {
		return domain.HeatmapFilter{}, 0, fmt.Errorf("%w: from cannot be after to", ErrInvalidSessionInput)
	}
	if int(to.Sub(from).Hours()/24)+1 > u.cfg.MaxDateRangeDays {
		return domain.HeatmapFilter{}, 0, fmt.Errorf("%w: date range cannot exceed %d days", ErrInvalidSessionInput, u.cfg.MaxDateRangeDays)
	}

	limit := input.Limit
	if limit <= 0 || limit > u.cfg.MaxPoints {
		limit = u.cfg.MaxPoints
	}
	return filter, limit, nil
}

func buildHeatmapOutput(filter domain.HeatmapFilter, points []domain.HeatmapPoint) HeatmapOutput {
	out := HeatmapOutput{
		Path:        filter.Path,
		DeviceType:  filter.DeviceType,
		HeatmapType: filter.HeatmapType,
		From:        filter.FromDay,
		To:          filter.ToDay,
		Points:      make([]HeatmapPointOutput, 0, len(points)),
	}
	for _, point := range points {
		normalized := point.Normalize()
		out.Points = append(out.Points, HeatmapPointOutput{
			X:      normalized.X,
			Y:      normalized.Y,
			Weight: normalized.Weight,
		})
		if normalized.Weight > out.MaxWeight {
			out.MaxWeight = normalized.Weight
		}
		out.TotalEvents += normalized.SampleEvents
	}
	return out
}

func (c HeatmapQueryConfig) Validate() error {
	if c.MaxDateRangeDays <= 0 {
		return errors.New("SESSION_HEATMAP_MAX_DATE_RANGE_DAYS must be greater than zero")
	}
	if c.MaxPoints <= 0 {
		return errors.New("SESSION_HEATMAP_MAX_POINTS must be greater than zero")
	}
	return nil
}

func (c HeatmapQueryConfig) withDefaults() HeatmapQueryConfig {
	if c.MaxDateRangeDays == 0 {
		c.MaxDateRangeDays = defaultHeatmapMaxDateRangeDays
	}
	if c.MaxPoints == 0 {
		c.MaxPoints = defaultHeatmapMaxPoints
	}
	return c
}
