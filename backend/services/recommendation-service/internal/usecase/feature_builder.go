package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
)

type FeatureBuilderConfig struct {
	GuestProfileRetention   time.Duration
	UserProductRetention    time.Duration
	ProcessedEventRetention time.Duration
	RecentProductsLimit     int
	ReconcileBatchSize      int
	ReconcileInterval       time.Duration
	WindowRebuildInterval   time.Duration
}

type FeatureBuilderService struct {
	repo    FeatureRepository
	metrics FeatureBuilderMetrics
	logger  *slog.Logger
	cfg     FeatureBuilderConfig
	now     func() time.Time
}

func NewFeatureBuilderService(repo FeatureRepository, cfg FeatureBuilderConfig, metrics FeatureBuilderMetrics, logger *slog.Logger) (*FeatureBuilderService, error) {
	if repo == nil {
		return nil, errors.New("feature repository is required")
	}
	if cfg.GuestProfileRetention <= 0 {
		cfg.GuestProfileRetention = domain.DefaultGuestProfileRetention
	}
	if cfg.UserProductRetention <= 0 {
		cfg.UserProductRetention = domain.DefaultUserProductRetention
	}
	if cfg.ProcessedEventRetention <= 0 {
		cfg.ProcessedEventRetention = domain.DefaultFeatureEventRetention
	}
	if cfg.RecentProductsLimit <= 0 {
		cfg.RecentProductsLimit = domain.DefaultRecentProductsLimit
	}
	if cfg.ReconcileBatchSize <= 0 {
		cfg.ReconcileBatchSize = 200
	}
	if cfg.ReconcileInterval < 0 || cfg.WindowRebuildInterval < 0 {
		return nil, errors.New("feature job intervals cannot be negative")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &FeatureBuilderService{
		repo:    repo,
		metrics: metrics,
		logger:  logger,
		cfg:     cfg,
		now:     func() time.Time { return time.Now().UTC() },
	}, nil
}

func (s *FeatureBuilderService) ApplyInteractions(ctx context.Context, interactions []domain.UserInteraction) (domain.FeatureBuildResult, error) {
	if err := ctx.Err(); err != nil {
		return domain.FeatureBuildResult{}, err
	}
	if len(interactions) == 0 {
		return domain.FeatureBuildResult{}, fmt.Errorf("%w: at least one interaction is required", domain.ErrInvalidFeature)
	}
	startedAt := s.now()
	batch := domain.FeatureBatch{
		ID:                      domain.NewFeatureJobID(domain.FeatureJobTypeIncremental, interactions, startedAt),
		Interactions:            interactions,
		StartedAt:               startedAt,
		ProcessedAt:             startedAt,
		GuestProfileRetention:   s.cfg.GuestProfileRetention,
		UserProductRetention:    s.cfg.UserProductRetention,
		ProcessedEventRetention: s.cfg.ProcessedEventRetention,
		RecentProductsLimit:     s.cfg.RecentProductsLimit,
	}
	if err := batch.Validate(); err != nil {
		return domain.FeatureBuildResult{}, err
	}

	result, err := s.repo.ApplyFeatureBatch(ctx, batch)
	duration := s.now().Sub(startedAt)
	if err != nil {
		if s.metrics != nil {
			s.metrics.RecordFeatureUpdateError(errorType(err))
			s.metrics.ObserveFeatureJobDuration(domain.FeatureJobTypeIncremental, duration)
		}
		return domain.FeatureBuildResult{}, fmt.Errorf("%w: apply interaction features: %v", domain.ErrRecommendationStorage, err)
	}
	s.recordSuccessMetrics(domain.FeatureJobTypeIncremental, result.Stats, duration)
	if s.metrics != nil && result.Applied > 0 {
		lag := startedAt.Sub(latestOccurredAt(interactions))
		if lag < 0 {
			lag = 0
		}
		s.metrics.SetFeatureJobLag(lag)
	}
	s.logger.InfoContext(ctx, "recommendation.feature_builder.completed",
		slog.String("job_type", domain.FeatureJobTypeIncremental),
		slog.Int("events_applied", result.Applied),
		slog.Int("duplicates", result.Duplicates),
		slog.Int("product_features_updated", result.Stats.ProductFeaturesUpdated),
		slog.Int("user_profiles_updated", result.Stats.UserProfilesUpdated),
		slog.Int("user_product_counters_updated", result.Stats.UserProductCountersUpdated),
		slog.Int("cooccurrence_pairs_updated", result.Stats.CooccurrencePairsUpdated),
		slog.Int64("duration_ms", duration.Milliseconds()),
	)
	return result, nil
}

func (s *FeatureBuilderService) RunReconciliation(ctx context.Context) {
	if s.cfg.ReconcileInterval <= 0 {
		return
	}
	s.reconcilePending(ctx)
	if s.cfg.WindowRebuildInterval > 0 {
		if _, err := s.RebuildProductWindows(ctx); err != nil {
			s.logger.ErrorContext(ctx, "recommendation.feature_builder.initial_window_rebuild_failed", slog.String("error", err.Error()))
		}
	}
	ticker := time.NewTicker(s.cfg.ReconcileInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.reconcilePending(ctx)
		}
	}
}

func (s *FeatureBuilderService) reconcilePending(ctx context.Context) {
	for {
		batches, err := s.repo.ListPendingFeatureBatches(ctx, s.cfg.ReconcileBatchSize)
		if err != nil {
			s.logger.ErrorContext(ctx, "recommendation.feature_builder.reconcile_read_failed", slog.String("error", err.Error()))
			return
		}
		if len(batches) == 0 {
			return
		}
		for _, interactions := range batches {
			if _, err := s.ApplyInteractions(ctx, interactions); err != nil {
				s.logger.ErrorContext(ctx, "recommendation.feature_builder.reconcile_apply_failed", slog.String("error", err.Error()))
				return
			}
		}
		if len(batches) < s.cfg.ReconcileBatchSize {
			return
		}
	}
}

func (s *FeatureBuilderService) RunWindowRebuilds(ctx context.Context) {
	if s.cfg.WindowRebuildInterval <= 0 {
		return
	}
	ticker := time.NewTicker(s.cfg.WindowRebuildInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := s.RebuildProductWindows(ctx); err != nil {
				s.logger.ErrorContext(ctx, "recommendation.feature_builder.window_rebuild_failed", slog.String("error", err.Error()))
			}
		}
	}
}

func (s *FeatureBuilderService) RebuildProductWindows(ctx context.Context) (domain.FeatureJobStats, error) {
	if err := ctx.Err(); err != nil {
		return domain.FeatureJobStats{}, err
	}
	startedAt := s.now()
	job := domain.FeatureJobRun{
		ID:        domain.NewFeatureJobID(domain.FeatureJobTypeWindowRebuild, nil, startedAt),
		JobType:   domain.FeatureJobTypeWindowRebuild,
		Status:    "running",
		StartedAt: startedAt,
		CreatedAt: startedAt,
	}
	stats, err := s.repo.RebuildProductWindows(ctx, job)
	duration := s.now().Sub(startedAt)
	if err != nil {
		if s.metrics != nil {
			s.metrics.RecordFeatureUpdateError(errorType(err))
			s.metrics.ObserveFeatureJobDuration(domain.FeatureJobTypeWindowRebuild, duration)
		}
		return domain.FeatureJobStats{}, fmt.Errorf("%w: rebuild product feature windows: %v", domain.ErrRecommendationStorage, err)
	}
	s.recordSuccessMetrics(domain.FeatureJobTypeWindowRebuild, stats, duration)
	s.logger.InfoContext(ctx, "recommendation.feature_builder.window_rebuilt",
		slog.Int("product_features_updated", stats.ProductFeaturesUpdated),
		slog.Int64("duration_ms", duration.Milliseconds()),
	)
	return stats, nil
}

func (s *FeatureBuilderService) recordSuccessMetrics(jobType string, stats domain.FeatureJobStats, duration time.Duration) {
	if s.metrics == nil {
		return
	}
	s.metrics.RecordFeatureEventsProcessed(stats.EventsRead)
	s.metrics.RecordFeatureDocumentsUpdated(domain.CollectionProductFeatures, stats.ProductFeaturesUpdated)
	s.metrics.RecordFeatureDocumentsUpdated(domain.CollectionUserFeatureProfiles, stats.UserProfilesUpdated)
	s.metrics.RecordFeatureDocumentsUpdated(domain.CollectionUserProductCounters, stats.UserProductCountersUpdated)
	s.metrics.RecordFeatureDocumentsUpdated(domain.CollectionProductCooccurrence, stats.CooccurrencePairsUpdated)
	s.metrics.ObserveFeatureJobDuration(jobType, duration)
}

func latestOccurredAt(interactions []domain.UserInteraction) time.Time {
	latest := interactions[0].OccurredAt
	for _, interaction := range interactions[1:] {
		if interaction.OccurredAt.After(latest) {
			latest = interaction.OccurredAt
		}
	}
	return latest
}
