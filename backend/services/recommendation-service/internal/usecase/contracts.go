package usecase

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
)

type DefinitionRepository interface {
	ListTypes(ctx context.Context) ([]domain.TypeDefinition, error)
	GetType(ctx context.Context, typ domain.RecommendationType) (domain.TypeDefinition, error)
	ListContexts(ctx context.Context) ([]domain.ContextDefinition, error)
	GetContext(ctx context.Context, context domain.RecommendationContext) (domain.ContextDefinition, error)
	FallbackChain(ctx context.Context) ([]domain.FallbackStep, error)
}

type RecommendationStorageRepository interface {
	Ping(ctx context.Context) error
	EnsureIndexes(ctx context.Context) error
}

type RecommendationCacheRepository interface {
	Ping(ctx context.Context) error
}

type InteractionRepository interface {
	InsertInteraction(ctx context.Context, interaction domain.UserInteraction) (domain.InteractionInsertResult, error)
}

type FeatureRepository interface {
	ApplyFeatureBatch(ctx context.Context, batch domain.FeatureBatch) (domain.FeatureBuildResult, error)
	ListPendingFeatureBatches(ctx context.Context, limit int) ([][]domain.UserInteraction, error)
	RebuildProductWindows(ctx context.Context, job domain.FeatureJobRun) (domain.FeatureJobStats, error)
}

type InteractionFeatureBuilder interface {
	ApplyInteractions(ctx context.Context, interactions []domain.UserInteraction) (domain.FeatureBuildResult, error)
}

type InteractionCacheInvalidator interface {
	InvalidateInteractions(ctx context.Context, interactions []domain.UserInteraction) error
}

type InteractionIngestionMetrics interface {
	RecordDuplicate(eventType string)
	RecordMongoInsertError(errorType string)
	RecordCacheInvalidationError(action string)
}

type FeatureBuilderMetrics interface {
	RecordFeatureEventsProcessed(count int)
	RecordFeatureUpdateError(errorType string)
	RecordFeatureDocumentsUpdated(collection string, count int)
	ObserveFeatureJobDuration(jobType string, duration time.Duration)
	SetFeatureJobLag(lag time.Duration)
}

type RankingFeatureReader interface {
	ListEligibleProductFeatures(ctx context.Context) ([]domain.ProductFeature, error)
}

type PersonalizedFeatureReader interface {
	GetUserFeatureProfile(ctx context.Context, profileKey string) (domain.UserFeatureProfile, error)
	ListUserProductCounters(ctx context.Context, profileKey string, limit int) ([]domain.UserProductCounter, error)
	ListPersonalizedCandidates(ctx context.Context, query domain.PersonalizedCandidateQuery) ([]domain.ProductFeature, error)
}

type RecommendationSetRepository interface {
	GetRecommendationSet(ctx context.Context, contextKey string, strategyID domain.StrategyID, now time.Time) (domain.RecommendationSet, error)
	UpsertRecommendationSet(ctx context.Context, set domain.RecommendationSet) error
}

type RankingResultCache interface {
	Get(ctx context.Context, key string) (domain.CachedRecommendation, time.Duration, error)
	Set(ctx context.Context, key string, cached domain.CachedRecommendation, ttl time.Duration) error
}

type RankingMetrics interface {
	RecordRankingRun(result string)
	ObserveRankingJobDuration(duration time.Duration)
	SetRankingCandidates(scope string, count int)
	RecordRankingItemsGenerated(scope string, count int)
	RecordRankingEmptySet(scope string)
	RecordRankingSetUpsertError(scope string)
	RecordRankingCacheSetError(scope string)
	RecordRankingFallback(scope string)
}

type PersonalizedMetrics interface {
	RecordPersonalizedBuild(outcome string)
	ObservePersonalizedBuildDuration(duration time.Duration)
	RecordPersonalizedProfileMiss(reason string)
	ObservePersonalizedCandidates(count int)
	ObservePersonalizedResultItems(count int)
	RecordPersonalizedFallback(source string)
	RecordPersonalizedBackfill(source string, count int)
	RecordPersonalizedCacheHit(identityType string)
	RecordPersonalizedCacheError(action string)
	RecordPersonalizedEmptyResult()
	RecordPersonalizedSetUpsertError()
}

type DefinitionServiceConfig struct {
	DefaultLimit        int
	MaxLimit            int
	MaxIdentifierLength int
}

type DefinitionService struct {
	repo                DefinitionRepository
	logger              *slog.Logger
	defaultLimit        int
	maxLimit            int
	maxIdentifierLength int
}

func NewDefinitionService(repo DefinitionRepository, cfg DefinitionServiceConfig, logger *slog.Logger) (*DefinitionService, error) {
	if repo == nil {
		return nil, errors.New("definition repository is required")
	}
	if cfg.DefaultLimit <= 0 {
		return nil, errors.New("default limit must be greater than zero")
	}
	if cfg.MaxLimit <= 0 {
		return nil, errors.New("max limit must be greater than zero")
	}
	if cfg.DefaultLimit > cfg.MaxLimit {
		return nil, errors.New("default limit must be less than or equal to max limit")
	}
	if cfg.MaxIdentifierLength <= 0 {
		return nil, errors.New("max identifier length must be greater than zero")
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &DefinitionService{
		repo:                repo,
		logger:              logger,
		defaultLimit:        cfg.DefaultLimit,
		maxLimit:            cfg.MaxLimit,
		maxIdentifierLength: cfg.MaxIdentifierLength,
	}, nil
}
