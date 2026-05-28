package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
)

type InteractionIngestionResult struct {
	Stored            int
	Duplicates        int
	FeaturesApplied   int
	FeatureDuplicates int
}

type InteractionIngestionService struct {
	repo        InteractionRepository
	builder     InteractionFeatureBuilder
	invalidator InteractionCacheInvalidator
	logger      *slog.Logger
	metrics     InteractionIngestionMetrics
}

func NewInteractionIngestionService(
	repo InteractionRepository,
	builder InteractionFeatureBuilder,
	invalidator InteractionCacheInvalidator,
	metrics InteractionIngestionMetrics,
	logger *slog.Logger,
) (*InteractionIngestionService, error) {
	if repo == nil {
		return nil, errors.New("interaction repository is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &InteractionIngestionService{
		repo:        repo,
		builder:     builder,
		invalidator: invalidator,
		logger:      logger,
		metrics:     metrics,
	}, nil
}

func (s *InteractionIngestionService) IngestInteractions(ctx context.Context, interactions []domain.UserInteraction) (InteractionIngestionResult, error) {
	if err := ctx.Err(); err != nil {
		return InteractionIngestionResult{}, err
	}
	if len(interactions) == 0 {
		return InteractionIngestionResult{}, fmt.Errorf("%w: at least one interaction is required", domain.ErrInvalidInteraction)
	}

	result := InteractionIngestionResult{}
	for _, interaction := range interactions {
		if err := interaction.Validate(); err != nil {
			return result, err
		}

		insertResult, err := s.repo.InsertInteraction(ctx, interaction)
		if err != nil {
			if s.metrics != nil {
				s.metrics.RecordMongoInsertError(errorType(err))
			}
			return result, fmt.Errorf("%w: insert interaction: %v", domain.ErrRecommendationStorage, err)
		}
		if insertResult.Duplicate {
			result.Duplicates++
			if s.metrics != nil {
				s.metrics.RecordDuplicate(string(interaction.SourceEventType))
			}
			continue
		}
		result.Stored++
	}

	if s.builder != nil {
		buildResult, err := s.builder.ApplyInteractions(ctx, interactions)
		if err != nil {
			return result, err
		}
		result.FeaturesApplied = buildResult.Applied
		result.FeatureDuplicates = buildResult.Duplicates
	}

	if s.invalidator != nil && (result.Stored > 0 || result.FeaturesApplied > 0) {
		if err := s.invalidator.InvalidateInteractions(ctx, interactions); err != nil {
			if s.metrics != nil {
				s.metrics.RecordCacheInvalidationError("interaction_invalidation")
			}
			s.logger.WarnContext(ctx, "recommendation.interaction.cache_invalidation_failed",
				slog.String("event_id", interactions[0].EventID),
				slog.String("event_type", string(interactions[0].SourceEventType)),
				slog.String("error", err.Error()),
			)
		}
	}

	s.logger.InfoContext(ctx, "recommendation.interaction.ingested",
		slog.String("event_id", interactions[0].EventID),
		slog.String("event_type", string(interactions[0].SourceEventType)),
		slog.Int("stored", result.Stored),
		slog.Int("duplicates", result.Duplicates),
		slog.Int("features_applied", result.FeaturesApplied),
		slog.Int("feature_duplicates", result.FeatureDuplicates),
	)
	return result, nil
}

func errorType(err error) string {
	switch {
	case errors.Is(err, context.Canceled):
		return "context_canceled"
	case errors.Is(err, context.DeadlineExceeded):
		return "deadline_exceeded"
	case errors.Is(err, domain.ErrInvalidInteraction):
		return "invalid_interaction"
	case errors.Is(err, domain.ErrRecommendationStorage):
		return "storage"
	default:
		return "unknown"
	}
}
