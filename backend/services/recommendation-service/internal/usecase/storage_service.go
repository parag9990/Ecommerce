package usecase

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
)

type StorageServiceConfig struct {
	DatabaseName         string
	CacheKeyPrefix       string
	TTLPolicy            domain.CacheTTLPolicy
	InteractionRetention time.Duration
	MongoConfigured      bool
	RedisConfigured      bool
}

type StorageComponentState string

const (
	StorageComponentNotConfigured StorageComponentState = "not_configured"
	StorageComponentUnavailable   StorageComponentState = "unavailable"
	StorageComponentReady         StorageComponentState = "ready"
)

type StorageComponentStatus struct {
	Layer      domain.StorageLayer
	Configured bool
	State      StorageComponentState
	Latency    time.Duration
	Error      string
}

type StorageStatus struct {
	CheckedAt time.Time
	MongoDB   StorageComponentStatus
	Redis     StorageComponentStatus
}

type StorageService struct {
	mongo  RecommendationStorageRepository
	redis  RecommendationCacheRepository
	logger *slog.Logger
	cfg    StorageServiceConfig
}

func NewStorageService(
	mongoRepo RecommendationStorageRepository,
	redisRepo RecommendationCacheRepository,
	cfg StorageServiceConfig,
	logger *slog.Logger,
) (*StorageService, error) {
	if cfg.DatabaseName == "" {
		cfg.DatabaseName = domain.RecommendationDatabaseName
	}
	if cfg.CacheKeyPrefix == "" {
		cfg.CacheKeyPrefix = domain.DefaultCacheKeyPrefix
	}
	if err := domain.ValidateCacheKeyPrefix(cfg.CacheKeyPrefix); err != nil {
		return nil, err
	}
	if cfg.TTLPolicy.Default <= 0 {
		cfg.TTLPolicy.Default = domain.DefaultRecommendationCacheTTL
	}
	if cfg.TTLPolicy.Personalized <= 0 {
		cfg.TTLPolicy.Personalized = domain.DefaultPersonalizedCacheTTL
	}
	if cfg.TTLPolicy.Guest <= 0 {
		cfg.TTLPolicy.Guest = domain.DefaultGuestCacheTTL
	}
	if cfg.TTLPolicy.RebuildLock <= 0 {
		cfg.TTLPolicy.RebuildLock = time.Minute
	}
	if cfg.InteractionRetention <= 0 {
		cfg.InteractionRetention = domain.DefaultInteractionRetention
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &StorageService{
		mongo:  mongoRepo,
		redis:  redisRepo,
		logger: logger,
		cfg:    cfg,
	}, nil
}

func (s *StorageService) StoragePlan(ctx context.Context) (domain.StoragePlan, error) {
	if err := ctx.Err(); err != nil {
		return domain.StoragePlan{}, err
	}
	return domain.BuildStoragePlan(s.cfg.DatabaseName, s.cfg.CacheKeyPrefix, s.cfg.TTLPolicy, s.cfg.InteractionRetention), nil
}

func (s *StorageService) EnsureIndexes(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.mongo == nil {
		return nil
	}
	if err := s.mongo.EnsureIndexes(ctx); err != nil {
		s.logger.ErrorContext(ctx, "recommendation.storage.ensure_indexes_failed", slog.String("error", err.Error()))
		return err
	}
	return nil
}

func (s *StorageService) Status(ctx context.Context) (StorageStatus, error) {
	if err := ctx.Err(); err != nil {
		return StorageStatus{}, err
	}
	return StorageStatus{
		CheckedAt: time.Now().UTC(),
		MongoDB:   s.checkMongo(ctx),
		Redis:     s.checkRedis(ctx),
	}, nil
}

func (s *StorageService) checkMongo(ctx context.Context) StorageComponentStatus {
	status := StorageComponentStatus{
		Layer:      domain.StorageLayerMongoDB,
		Configured: s.cfg.MongoConfigured,
	}
	if s.mongo == nil {
		if status.Configured {
			status.State = StorageComponentUnavailable
			status.Error = "mongo repository is not initialized"
			return status
		}
		status.State = StorageComponentNotConfigured
		return status
	}
	start := time.Now()
	err := s.mongo.Ping(ctx)
	status.Latency = time.Since(start)
	if err != nil {
		status.State = StorageComponentUnavailable
		status.Error = err.Error()
		return status
	}
	status.Configured = true
	status.State = StorageComponentReady
	return status
}

func (s *StorageService) checkRedis(ctx context.Context) StorageComponentStatus {
	status := StorageComponentStatus{
		Layer:      domain.StorageLayerRedis,
		Configured: s.cfg.RedisConfigured,
	}
	if s.redis == nil {
		if status.Configured {
			status.State = StorageComponentUnavailable
			status.Error = "redis cache is not initialized"
			return status
		}
		status.State = StorageComponentNotConfigured
		return status
	}
	start := time.Now()
	err := s.redis.Ping(ctx)
	status.Latency = time.Since(start)
	if err != nil {
		status.State = StorageComponentUnavailable
		status.Error = err.Error()
		return status
	}
	status.Configured = true
	status.State = StorageComponentReady
	return status
}

func (s StorageStatus) Ready() bool {
	mongoReady := !s.MongoDB.Configured || s.MongoDB.State == StorageComponentReady
	redisReady := !s.Redis.Configured || s.Redis.State == StorageComponentReady
	return mongoReady && redisReady
}

func IsStorageUnavailable(err error) bool {
	return errors.Is(err, domain.ErrRecommendationStorage)
}
