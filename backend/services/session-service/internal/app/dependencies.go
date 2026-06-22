package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/config"
	deviceinfra "github.com/example/ecommerce-platform/backend/services/session-service/internal/device"
	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/session-service/internal/repository"
	httptransport "github.com/example/ecommerce-platform/backend/services/session-service/internal/transport/http"
	"github.com/example/ecommerce-platform/backend/services/session-service/internal/usecase"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Dependencies struct {
	MongoClient               *mongo.Client
	RedisClient               *redis.Client
	SessionRepo               *repository.MongoSessionRepository
	EventRepo                 *repository.MongoEventRepository
	JourneyRepo               *repository.MongoJourneySummaryRepository
	HeatmapRepo               *repository.MongoHeatmapRepository
	AnalyticsRepo             *repository.MongoAnalyticsRepository
	RetentionRepo             *repository.MongoRetentionRepository
	PrivacyRepo               *repository.MongoPrivacyRepository
	ReportsRepo               *repository.MongoReportsRepository
	ActiveStore               *repository.RedisActiveSessionStore
	LiveMetricsRepo           *repository.RedisLiveMetricsRepository
	StorageUsecase            *usecase.StorageUsecase
	IngestUsecase             *usecase.IngestUsecase
	JourneyUsecase            *usecase.JourneyUsecase
	HeatmapUsecase            *usecase.HeatmapUsecase
	HeatmapAggregationUsecase *usecase.HeatmapAggregationUsecase
	AnalyticsUsecase          *usecase.AnalyticsUsecase
	RetentionUsecase          *usecase.RetentionUsecase
	PrivacyUsecase            *usecase.PrivacyUsecase
	ReportsUsecase            *usecase.ReportsUsecase
	HTTPHandler               *httptransport.Handler
	heatmapWorkerCancel       context.CancelFunc
	geoCloser                 func() error
}

func NewDependencies(ctx context.Context, cfg config.Config, logger *slog.Logger) (*Dependencies, error) {
	if logger == nil {
		logger = slog.Default()
	}

	mongoClient, err := repository.NewMongoClient(ctx, cfg.Storage.MongoConfig())
	if err != nil {
		return nil, err
	}
	cleanupMongo := true
	defer func() {
		if cleanupMongo {
			_ = mongoClient.Disconnect(context.Background())
		}
	}()

	database, err := repository.MongoDatabase(mongoClient, cfg.Storage.MongoDatabase)
	if err != nil {
		return nil, err
	}
	sessionRepo, err := repository.NewMongoSessionRepository(
		database,
		cfg.SessionModel.UsecaseConfig().Validation,
		logger,
		repository.WithSessionMetadataRetentionDays(cfg.Retention.SessionMetadataRetentionDays),
	)
	if err != nil {
		return nil, err
	}
	eventRepo, err := repository.NewMongoEventRepository(database, cfg.EventValidationConfig(), cfg.Storage.RawEventTTLDays, logger)
	if err != nil {
		return nil, err
	}
	journeyRepo, err := repository.NewMongoJourneySummaryRepository(
		database,
		domainJourneySummaryValidation(cfg),
		logger,
		repository.WithJourneySummaryRetentionDays(cfg.Retention.JourneySummaryRetentionDays),
	)
	if err != nil {
		return nil, err
	}
	heatmapRepo, err := repository.NewMongoHeatmapRepository(database, logger, repository.WithHeatmapRetentionDays(cfg.Retention.HeatmapRetentionDays))
	if err != nil {
		return nil, err
	}
	analyticsRepo, err := repository.NewMongoAnalyticsRepository(database, logger)
	if err != nil {
		return nil, err
	}
	retentionRepo, err := repository.NewMongoRetentionRepository(database, logger)
	if err != nil {
		return nil, err
	}
	privacyRepo, err := repository.NewMongoPrivacyRepository(database)
	if err != nil {
		return nil, err
	}
	reportsRepo, err := repository.NewMongoReportsRepository(database)
	if err != nil {
		return nil, err
	}
	if err := sessionRepo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	if err := eventRepo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	if err := journeyRepo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	if err := heatmapRepo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	if err := analyticsRepo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	if err := retentionRepo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	if err := privacyRepo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}
	if err := reportsRepo.EnsureIndexes(ctx); err != nil {
		return nil, err
	}

	redisClient, err := repository.NewRedisClient(cfg.Storage.RedisClientConfig())
	if err != nil {
		return nil, err
	}
	cleanupRedis := true
	defer func() {
		if cleanupRedis {
			_ = redisClient.Close()
		}
	}()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	activeStore, err := repository.NewRedisActiveSessionStore(redisClient, cfg.Storage.ActiveSessionStoreConfig(), logger)
	if err != nil {
		return nil, err
	}
	if savedPrivacy, err := privacyRepo.GetPrivacySettings(ctx); err == nil {
		if err := privacyRepo.ApplyRetentionSettings(ctx, savedPrivacy.Retention); err != nil {
			return nil, fmt.Errorf("apply saved privacy retention policy: %w", err)
		}
		if err := activeStore.ApplyRetentionTTL(ctx, time.Duration(savedPrivacy.Retention.ActiveSessionTTLMinutes)*time.Minute); err != nil {
			return nil, fmt.Errorf("apply saved active session ttl: %w", err)
		}
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("load saved privacy settings: %w", err)
	}
	liveMetricsRepo, err := repository.NewRedisLiveMetricsRepository(redisClient, repository.RedisLiveMetricsConfig{
		KeyPrefix: cfg.Storage.RedisKeyPrefix,
	}, logger)
	if err != nil {
		return nil, err
	}
	geoResolver, geoCloser, err := newGeoResolver(cfg)
	if err != nil {
		return nil, err
	}
	cleanupGeo := true
	defer func() {
		if cleanupGeo && geoCloser != nil {
			_ = geoCloser()
		}
	}()
	deviceEnricher, err := usecase.NewDeviceEnricher(
		deviceinfra.UserAgentParser{},
		geoResolver,
		deviceinfra.NewPrivacyHasher(cfg.DeviceTracking.PrivacyHashPepper),
		cfg.DeviceEnricherConfig(),
		logger,
	)
	if err != nil {
		return nil, err
	}
	storageUsecase, err := usecase.NewStorageUsecase(
		sessionRepo,
		eventRepo,
		activeStore,
		cfg.StorageUsecaseConfig(),
		logger,
	)
	if err != nil {
		return nil, err
	}
	ingestUsecase, err := usecase.NewIngestUsecase(
		eventRepo,
		sessionRepo,
		activeStore,
		nil,
		cfg.IngestUsecaseConfig(),
		logger,
		usecase.WithDeviceEnricher(deviceEnricher),
	)
	if err != nil {
		return nil, err
	}
	journeyUsecase, err := usecase.NewJourneyUsecase(
		sessionRepo,
		eventRepo,
		journeyRepo,
		cfg.JourneyUsecaseConfig(),
		logger,
	)
	if err != nil {
		return nil, err
	}
	heatmapUsecase, err := usecase.NewHeatmapUsecase(
		heatmapRepo,
		cfg.HeatmapQueryConfig(),
		logger,
	)
	if err != nil {
		return nil, err
	}
	heatmapAggregationUsecase, err := usecase.NewHeatmapAggregationUsecase(
		eventRepo,
		heatmapRepo,
		cfg.HeatmapAggregationConfig(),
		logger,
	)
	if err != nil {
		return nil, err
	}
	analyticsUsecase, err := usecase.NewAnalyticsUsecase(
		liveMetricsRepo,
		sessionRepo,
		analyticsRepo,
		cfg.AnalyticsUsecaseConfig(),
		logger,
	)
	if err != nil {
		return nil, err
	}
	retentionUsecase, err := usecase.NewRetentionUsecase(
		retentionRepo,
		activeStore,
		cfg.RetentionUsecaseConfig(),
		logger,
	)
	if err != nil {
		return nil, err
	}
	privacyUsecase, err := usecase.NewPrivacyUsecase(
		privacyRepo,
		activeStore,
		cfg.Privacy.UsecaseConfig(),
		logger,
	)
	if err != nil {
		return nil, err
	}
	reportsUsecase, err := usecase.NewReportsUsecase(reportsRepo, cfg.Reports.UsecaseConfig())
	if err != nil {
		return nil, err
	}

	requestContextConfig, err := httptransport.NewRequestContextConfig(cfg.DeviceTracking.TrustedProxyCIDRs)
	if err != nil {
		return nil, err
	}
	handler, err := httptransport.NewHandler(map[string]httptransport.ProbeFunc{
		"mongo": func(probeCtx context.Context) error {
			return mongoClient.Ping(probeCtx, nil)
		},
		"redis": func(probeCtx context.Context) error {
			return redisClient.Ping(probeCtx).Err()
		},
	}, ingestUsecase, journeyUsecase, heatmapUsecase, analyticsUsecase, cfg.Ingest.MaxBodyBytes, logger, requestContextConfig)
	if err != nil {
		return nil, err
	}
	handler.SetRetentionUsecase(retentionUsecase)
	handler.SetPrivacyUsecase(privacyUsecase)
	handler.SetReportsUsecase(reportsUsecase)
	sessionReferenceCodec, err := domain.NewSessionReferenceCodec(cfg.Privacy.HashPepper)
	if err != nil {
		return nil, err
	}
	handler.SetSessionReferenceCodec(sessionReferenceCodec)

	var heatmapWorkerCancel context.CancelFunc
	if cfg.Heatmap.AggregationEnabled {
		var workerCtx context.Context
		workerCtx, heatmapWorkerCancel = context.WithCancel(ctx)
		startHeatmapWorker(workerCtx, heatmapAggregationUsecase, cfg, logger)
	}

	cleanupMongo = false
	cleanupRedis = false
	cleanupGeo = false
	return &Dependencies{
		MongoClient:               mongoClient,
		RedisClient:               redisClient,
		SessionRepo:               sessionRepo,
		EventRepo:                 eventRepo,
		JourneyRepo:               journeyRepo,
		HeatmapRepo:               heatmapRepo,
		AnalyticsRepo:             analyticsRepo,
		RetentionRepo:             retentionRepo,
		PrivacyRepo:               privacyRepo,
		ReportsRepo:               reportsRepo,
		ActiveStore:               activeStore,
		LiveMetricsRepo:           liveMetricsRepo,
		StorageUsecase:            storageUsecase,
		IngestUsecase:             ingestUsecase,
		JourneyUsecase:            journeyUsecase,
		HeatmapUsecase:            heatmapUsecase,
		HeatmapAggregationUsecase: heatmapAggregationUsecase,
		AnalyticsUsecase:          analyticsUsecase,
		RetentionUsecase:          retentionUsecase,
		PrivacyUsecase:            privacyUsecase,
		ReportsUsecase:            reportsUsecase,
		HTTPHandler:               handler,
		heatmapWorkerCancel:       heatmapWorkerCancel,
		geoCloser:                 geoCloser,
	}, nil
}

func startHeatmapWorker(ctx context.Context, aggregator *usecase.HeatmapAggregationUsecase, cfg config.Config, logger *slog.Logger) {
	if aggregator == nil {
		return
	}
	run := func() {
		_, err := aggregator.AggregateHeatmap(ctx, usecase.AggregateHeatmapInput{
			UseCheckpoint:  true,
			SaveCheckpoint: true,
		})
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.WarnContext(ctx, "session.heatmap.worker_run_failed", slog.String("error", err.Error()))
		}
	}
	go func() {
		run()
		ticker := time.NewTicker(cfg.Heatmap.AggregationInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run()
			}
		}
	}()
}

func newGeoResolver(cfg config.Config) (usecase.GeoResolver, func() error, error) {
	if !cfg.DeviceTracking.Enabled || !cfg.DeviceTracking.GeoIPEnabled {
		return deviceinfra.DisabledGeoResolver{}, nil, nil
	}
	resolver, err := deviceinfra.NewMaxMindGeoResolver(cfg.DeviceTracking.GeoIPDBPath)
	if err != nil {
		return nil, nil, err
	}
	return resolver, resolver.Close, nil
}

func domainJourneySummaryValidation(cfg config.Config) domain.JourneySummaryValidationConfig {
	return domain.JourneySummaryValidationConfig{
		Session:       cfg.SessionModel.UsecaseConfig().Validation,
		MaxMilestones: domain.DefaultMaxJourneyMilestones,
		MaxTopPaths:   cfg.Journey.SummaryTopPathsLimit,
	}.WithDefaults()
}

func (d *Dependencies) Close(ctx context.Context) error {
	if d == nil {
		return nil
	}
	var combined error
	if d.heatmapWorkerCancel != nil {
		d.heatmapWorkerCancel()
	}
	if d.RedisClient != nil {
		if err := d.RedisClient.Close(); err != nil {
			combined = errors.Join(combined, fmt.Errorf("close redis: %w", err))
		}
	}
	if d.geoCloser != nil {
		if err := d.geoCloser(); err != nil {
			combined = errors.Join(combined, fmt.Errorf("close geo resolver: %w", err))
		}
	}
	if d.MongoClient != nil {
		if err := d.MongoClient.Disconnect(ctx); err != nil {
			combined = errors.Join(combined, fmt.Errorf("disconnect mongo: %w", err))
		}
	}
	return combined
}
