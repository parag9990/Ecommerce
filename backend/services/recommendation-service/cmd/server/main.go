package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	recommendationv1 "github.com/example/ecommerce-platform/backend/proto-gen/go/ecommerce/recommendation/v1"
	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/config"
	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/events"
	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/observability"
	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/repository"
	grpctransport "github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/transport/grpc"
	httptransport "github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/transport/http"
	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/usecase"
	platformmiddleware "github.com/parag/ecommerce/backend/shared/platform/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type eventConsumer interface {
	Run(ctx context.Context) error
	Close() error
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		logger.Error("recommendation.config.load_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.Service.LogLevel}))

	definitionRepo, err := repository.NewStaticDefinitionRepository()
	if err != nil {
		logger.Error("recommendation.definition_repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	definitionService, err := usecase.NewDefinitionService(
		definitionRepo,
		usecase.DefinitionServiceConfig{
			DefaultLimit:        cfg.Recommendation.DefaultLimit,
			MaxLimit:            cfg.Recommendation.MaxLimit,
			MaxIdentifierLength: cfg.Recommendation.MaxIdentifierLength,
		},
		logger,
	)
	if err != nil {
		logger.Error("recommendation.definition_service.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	mongoRepo, closeMongo := initMongoRepository(context.Background(), cfg, logger)
	defer closeMongo()
	redisCache, closeRedis := initRedisCache(context.Background(), cfg, logger)
	defer closeRedis()

	metrics, err := observability.NewConsumerMetrics(nil)
	if err != nil {
		logger.Error("recommendation.metrics.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	var storageRepo usecase.RecommendationStorageRepository
	if mongoRepo != nil {
		storageRepo = mongoRepo
	}
	var cacheRepo usecase.RecommendationCacheRepository
	if redisCache != nil {
		cacheRepo = redisCache
	}
	storageService, err := usecase.NewStorageService(
		storageRepo,
		cacheRepo,
		usecase.StorageServiceConfig{
			DatabaseName:         cfg.Storage.Mongo.Database,
			CacheKeyPrefix:       cfg.Storage.Cache.KeyPrefix,
			InteractionRetention: cfg.Storage.Mongo.InteractionRetention,
			MongoConfigured:      cfg.Storage.Mongo.Enabled(),
			RedisConfigured:      cfg.Storage.Redis.Enabled(),
			TTLPolicy:            domainCacheTTLPolicy(cfg),
		},
		logger,
	)
	if err != nil {
		logger.Error("recommendation.storage_service.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if cfg.Storage.Mongo.EnsureIndexes {
		if err := storageService.EnsureIndexes(context.Background()); err != nil {
			logStorageInitError(logger, cfg, "mongo_indexes", err)
		}
	}
	featureBuilder, err := initFeatureBuilder(cfg, mongoRepo, metrics, logger)
	if err != nil {
		logger.Error("recommendation.feature_builder.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	trendingService, err := initTrendingService(cfg, mongoRepo, redisCache, metrics, logger)
	if err != nil {
		logger.Error("recommendation.ranking.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	personalizedService, err := initPersonalizedService(cfg, mongoRepo, redisCache, metrics, logger)
	if err != nil {
		logger.Error("recommendation.personalization.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	assignmentService, err := initExperimentAssignmentService(cfg, mongoRepo, metrics, logger)
	if err != nil {
		logger.Error("recommendation.ab_assignment.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	recommendationReader, err := usecase.NewGetRecommendationsService(
		definitionService,
		storageRepoAsSets(mongoRepo),
		cacheRepoAsRankingCache(redisCache),
		trendingService,
		personalizedService,
		assignmentService,
		usecase.GetRecommendationsConfig{
			CacheKeyPrefix:      cfg.Storage.Cache.KeyPrefix,
			MaxIdentifierLength: cfg.Recommendation.MaxIdentifierLength,
			TTLPolicy:           domainCacheTTLPolicy(cfg),
			ABTestingFailOpen:   cfg.ABTesting.FailOpen,
		},
		metrics,
		logger,
	)
	if err != nil {
		logger.Error("recommendation.serving.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	handler, err := httptransport.NewHandler(definitionService, storageService, logger, cfg.HTTP.MaxBodyBytes)
	if err != nil {
		logger.Error("recommendation.http.handler_init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	interactionConsumer, closeInteractionConsumer, err := initInteractionConsumer(cfg, mongoRepo, featureBuilder, redisCache, metrics, logger)
	if err != nil {
		logger.Error("recommendation.interaction_consumer.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer closeInteractionConsumer()

	server := &http.Server{
		Addr:         cfg.HTTP.Address,
		Handler:      platformmiddleware.CORS(platformmiddleware.DefaultCORSConfig())(httptransport.NewRouter(handler)),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}
	grpcHandler, err := grpctransport.NewHandler(recommendationReader, logger)
	if err != nil {
		logger.Error("recommendation.grpc.handler_init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	grpcServer := grpctransport.NewServer(grpctransport.ServerConfig{
		MaxRecvBytes:   cfg.GRPC.MaxRecvBytes,
		MaxSendBytes:   cfg.GRPC.MaxSendBytes,
		DefaultTimeout: cfg.GRPC.DefaultDeadline,
	}, logger, metrics)
	recommendationv1.RegisterRecommendationServiceServer(grpcServer, grpcHandler)
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("ecommerce.recommendation.v1.RecommendationService", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	grpcListener, err := net.Listen("tcp", cfg.GRPC.Address)
	if err != nil {
		logger.Error("recommendation.grpc.listen_failed",
			slog.String("addr", cfg.GRPC.Address),
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if interactionConsumer != nil {
		go func() {
			if err := interactionConsumer.Run(ctx); err != nil {
				logger.Error("recommendation.interaction_consumer.run_failed", slog.String("error", err.Error()))
				stop()
			}
		}()
	}
	if featureBuilder != nil {
		go featureBuilder.RunReconciliation(ctx)
		go featureBuilder.RunWindowRebuilds(ctx)
	}
	if trendingService != nil {
		go trendingService.RunScheduled(ctx)
	}

	go func() {
		logger.Info("recommendation.http.started",
			slog.String("addr", cfg.HTTP.Address),
			slog.String("service", cfg.Service.Name),
			slog.String("env", cfg.Service.Env),
		)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("recommendation.http.listen_failed", slog.String("error", err.Error()))
			stop()
		}
	}()
	go func() {
		logger.Info("recommendation.grpc.started",
			slog.String("addr", cfg.GRPC.Address),
			slog.String("service", cfg.Service.Name),
			slog.String("env", cfg.Service.Env),
		)
		if err := grpcServer.Serve(grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			logger.Error("recommendation.grpc.serve_failed", slog.String("error", err.Error()))
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancelShutdown()
	shutdownGRPCServer(shutdownCtx, grpcServer, healthServer, logger)
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("recommendation.http.shutdown_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("recommendation.service.stopped")
}

func storageRepoAsSets(repo *repository.MongoFeatureRepository) usecase.RecommendationSetRepository {
	if repo == nil {
		return nil
	}
	return repo
}

func cacheRepoAsRankingCache(cache *repository.RedisRecommendationCache) usecase.RankingResultCache {
	if cache == nil {
		return nil
	}
	return cache
}

func shutdownGRPCServer(ctx context.Context, server *grpc.Server, healthServer *health.Server, logger *slog.Logger) {
	if server == nil {
		return
	}
	if healthServer != nil {
		healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
		healthServer.SetServingStatus("ecommerce.recommendation.v1.RecommendationService", healthpb.HealthCheckResponse_NOT_SERVING)
	}
	stopped := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(stopped)
	}()
	select {
	case <-stopped:
		logger.Info("recommendation.grpc.stopped")
	case <-ctx.Done():
		server.Stop()
		logger.Warn("recommendation.grpc.stop_forced", slog.String("error", ctx.Err().Error()))
	}
}

func initMongoRepository(ctx context.Context, cfg config.Config, logger *slog.Logger) (*repository.MongoFeatureRepository, func()) {
	if !cfg.Storage.Mongo.Enabled() {
		logger.Info("recommendation.mongo.not_configured")
		return nil, func() {}
	}
	client, err := repository.NewMongoClient(ctx, repository.MongoClientConfig{
		URI:            cfg.Storage.Mongo.URI,
		ConnectTimeout: cfg.Storage.Mongo.ConnectTimeout,
		PingTimeout:    cfg.Storage.Mongo.PingTimeout,
	}, logger)
	if err != nil {
		logStorageInitError(logger, cfg, "mongo", err)
		return nil, func() {}
	}

	repo, err := repository.NewMongoFeatureRepository(
		client.Database(cfg.Storage.Mongo.Database),
		repository.MongoFeatureRepositoryConfig{
			InteractionRetention: cfg.Storage.Mongo.InteractionRetention,
		},
		logger,
	)
	if err != nil {
		_ = client.Disconnect(context.Background())
		logStorageInitError(logger, cfg, "mongo_repository", err)
		return nil, func() {}
	}

	return repo, func() {
		if err := client.Disconnect(context.Background()); err != nil {
			logger.Error("recommendation.mongo.disconnect_failed", slog.String("error", err.Error()))
		}
	}
}

func initRedisCache(ctx context.Context, cfg config.Config, logger *slog.Logger) (*repository.RedisRecommendationCache, func()) {
	if !cfg.Storage.Redis.Enabled() {
		logger.Info("recommendation.redis.not_configured")
		return nil, func() {}
	}
	client, err := repository.NewRedisClient(ctx, repository.RedisClientConfig{
		Addr:         cfg.Storage.Redis.Addr,
		Password:     cfg.Storage.Redis.Password,
		DB:           cfg.Storage.Redis.DB,
		DialTimeout:  cfg.Storage.Redis.DialTimeout,
		ReadTimeout:  cfg.Storage.Redis.ReadTimeout,
		WriteTimeout: cfg.Storage.Redis.WriteTimeout,
		PingTimeout:  cfg.Storage.Redis.PingTimeout,
	}, logger)
	if err != nil {
		logStorageInitError(logger, cfg, "redis", err)
		return nil, func() {}
	}

	cache, err := repository.NewRedisRecommendationCache(client, repository.RedisRecommendationCacheConfig{
		KeyPrefix:       cfg.Storage.Cache.KeyPrefix,
		DefaultTTL:      cfg.Storage.Cache.DefaultTTL,
		PersonalizedTTL: cfg.Storage.Cache.PersonalizedTTL,
		GuestTTL:        cfg.Storage.Cache.GuestTTL,
		RebuildLockTTL:  cfg.Storage.Cache.RebuildLockTTL,
		DirtyTTL:        cfg.Storage.Cache.DirtyTTL,
	}, logger)
	if err != nil {
		_ = client.Close()
		logStorageInitError(logger, cfg, "redis_cache", err)
		return nil, func() {}
	}
	return cache, func() {
		if err := client.Close(); err != nil {
			logger.Error("recommendation.redis.close_failed", slog.String("error", err.Error()))
		}
	}
}

func initInteractionConsumer(
	cfg config.Config,
	mongoRepo *repository.MongoFeatureRepository,
	featureBuilder *usecase.FeatureBuilderService,
	redisCache *repository.RedisRecommendationCache,
	metrics *observability.ConsumerMetrics,
	logger *slog.Logger,
) (eventConsumer, func(), error) {
	if !cfg.Events.Active() {
		logger.Info("recommendation.interaction_consumer.not_configured")
		return nil, func() {}, nil
	}
	if mongoRepo == nil {
		return nil, func() {}, errors.New("event ingestion requires configured MongoDB storage")
	}

	var invalidator usecase.InteractionCacheInvalidator
	if redisCache != nil {
		invalidator = redisCache
	}
	ingestionService, err := usecase.NewInteractionIngestionService(mongoRepo, featureBuilder, invalidator, metrics, logger)
	if err != nil {
		return nil, func() {}, err
	}
	mapper, err := events.NewInteractionMapper(events.InteractionMapperConfig{
		MaxIdentifierLength: cfg.Recommendation.MaxIdentifierLength,
		SupportedVersion:    cfg.Events.SupportedVersion,
	})
	if err != nil {
		return nil, func() {}, err
	}
	dlqPublisher, err := events.NewKafkaDeadLetterPublisher(cfg.Events.Kafka.Brokers, cfg.Events.Kafka.DLQTopic)
	if err != nil {
		return nil, func() {}, err
	}
	handler, err := events.NewInteractionHandler(
		mapper,
		ingestionService,
		dlqPublisher,
		metrics,
		events.InteractionHandlerConfig{
			ServiceName:        cfg.Service.Name,
			Topic:              cfg.Events.Kafka.Topic,
			SupportedVersion:   cfg.Events.SupportedVersion,
			MaxMessageBytes:    cfg.Events.MaxMessageBytes,
			MaxRetryAttempts:   cfg.Events.MaxRetryAttempts,
			RetryBackoffs:      cfg.Events.RetryBackoffs,
			UnknownEventPolicy: cfg.Events.UnknownEventPolicy,
		},
		logger,
	)
	if err != nil {
		_ = dlqPublisher.Close()
		return nil, func() {}, err
	}
	consumer, err := events.NewKafkaInteractionConsumer(events.KafkaConsumerConfig{
		Brokers:  cfg.Events.Kafka.Brokers,
		Topic:    cfg.Events.Kafka.Topic,
		GroupID:  cfg.Events.Kafka.GroupID,
		MinBytes: cfg.Events.Kafka.MinBytes,
		MaxBytes: cfg.Events.Kafka.MaxBytes,
	}, handler, metrics, logger)
	if err != nil {
		_ = dlqPublisher.Close()
		return nil, func() {}, err
	}

	closeFn := func() {
		if err := consumer.Close(); err != nil {
			logger.Error("recommendation.kafka.consumer_close_failed", slog.String("error", err.Error()))
		}
		if err := dlqPublisher.Close(); err != nil {
			logger.Error("recommendation.kafka.dlq_close_failed", slog.String("error", err.Error()))
		}
	}
	return consumer, closeFn, nil
}

func initFeatureBuilder(
	cfg config.Config,
	mongoRepo *repository.MongoFeatureRepository,
	metrics *observability.ConsumerMetrics,
	logger *slog.Logger,
) (*usecase.FeatureBuilderService, error) {
	if !cfg.Features.Enabled {
		logger.Info("recommendation.feature_builder.disabled")
		return nil, nil
	}
	if mongoRepo == nil {
		if cfg.Events.Active() {
			return nil, errors.New("feature building requires configured MongoDB storage")
		}
		logger.Warn("recommendation.feature_builder.not_configured", slog.String("reason", "mongo storage is not configured"))
		return nil, nil
	}
	return usecase.NewFeatureBuilderService(mongoRepo, usecase.FeatureBuilderConfig{
		GuestProfileRetention:   cfg.Features.GuestProfileRetention,
		UserProductRetention:    cfg.Features.UserProductRetention,
		ProcessedEventRetention: cfg.Features.ProcessedEventRetention,
		RecentProductsLimit:     cfg.Features.RecentProductsLimit,
		ReconcileBatchSize:      cfg.Features.ReconcileBatchSize,
		ReconcileInterval:       cfg.Features.ReconcileInterval,
		WindowRebuildInterval:   cfg.Features.WindowRebuildInterval,
	}, metrics, logger)
}

func initTrendingService(
	cfg config.Config,
	mongoRepo *repository.MongoFeatureRepository,
	redisCache *repository.RedisRecommendationCache,
	metrics *observability.ConsumerMetrics,
	logger *slog.Logger,
) (*usecase.TrendingService, error) {
	if !cfg.Ranking.Enabled {
		logger.Info("recommendation.ranking.disabled")
		return nil, nil
	}
	if mongoRepo == nil {
		logger.Warn("recommendation.ranking.not_configured", slog.String("reason", "mongo storage is not configured"))
		return nil, nil
	}
	var cache usecase.RankingResultCache
	if redisCache != nil {
		cache = redisCache
	}
	return usecase.NewTrendingService(
		mongoRepo,
		mongoRepo,
		cache,
		usecase.TrendingServiceConfig{
			Weights:             cfg.Ranking.Weights,
			FormulaVersion:      cfg.Ranking.FormulaVersion,
			Limit:               cfg.Recommendation.DefaultLimit,
			MaxIdentifierLength: cfg.Recommendation.MaxIdentifierLength,
			CacheKeyPrefix:      cfg.Storage.Cache.KeyPrefix,
			TTL:                 cfg.Storage.Cache.DefaultTTL,
			RebuildInterval:     cfg.Ranking.RebuildInterval,
		},
		metrics,
		logger,
	)
}

func initPersonalizedService(
	cfg config.Config,
	mongoRepo *repository.MongoFeatureRepository,
	redisCache *repository.RedisRecommendationCache,
	metrics *observability.ConsumerMetrics,
	logger *slog.Logger,
) (*usecase.PersonalizedService, error) {
	if !cfg.Personalization.Enabled {
		logger.Info("recommendation.personalization.disabled")
		return nil, nil
	}
	if mongoRepo == nil {
		logger.Warn("recommendation.personalization.not_configured", slog.String("reason", "mongo storage is not configured"))
		return nil, nil
	}
	var cache usecase.RankingResultCache
	if redisCache != nil {
		cache = redisCache
	}
	return usecase.NewPersonalizedService(
		mongoRepo,
		mongoRepo,
		cache,
		usecase.PersonalizedServiceConfig{
			FormulaVersion:          cfg.Personalization.FormulaVersion,
			Limit:                   cfg.Recommendation.DefaultLimit,
			MaxLimit:                cfg.Recommendation.MaxLimit,
			MaxIdentifierLength:     cfg.Recommendation.MaxIdentifierLength,
			CacheKeyPrefix:          cfg.Storage.Cache.KeyPrefix,
			TTL:                     cfg.Storage.Cache.PersonalizedTTL,
			GuestTTL:                cfg.Storage.Cache.GuestTTL,
			MinPositiveInteractions: cfg.Personalization.MinPositiveInteractions,
			ProfileMaxAge:           cfg.Personalization.ProfileMaxAge,
			MaxCandidates:           cfg.Personalization.MaxCandidates,
			MaxItemsPerSeller:       cfg.Personalization.MaxItemsPerSeller,
			TopCategories:           cfg.Personalization.TopCategories,
			TopSellers:              cfg.Personalization.TopSellers,
			TopBrands:               cfg.Personalization.TopBrands,
			DirectProductLimit:      cfg.Personalization.DirectProductLimit,
			Weights:                 cfg.Personalization.Weights,
		},
		metrics,
		logger,
	)
}

func initExperimentAssignmentService(
	cfg config.Config,
	mongoRepo *repository.MongoFeatureRepository,
	metrics *observability.ConsumerMetrics,
	logger *slog.Logger,
) (*usecase.ExperimentAssignmentService, error) {
	if !cfg.ABTesting.Enabled {
		logger.Info("recommendation.ab_testing.disabled")
		return nil, nil
	}
	if len(cfg.ABTesting.Experiments) == 0 {
		logger.Info("recommendation.ab_testing.no_experiments")
	}
	if mongoRepo == nil {
		logger.Warn("recommendation.ab_testing.repository_not_configured", slog.String("behavior", "fail_open_to_default_strategy"))
	}
	return usecase.NewExperimentAssignmentService(
		mongoRepo,
		usecase.ExperimentAssignmentConfig{
			Enabled:             cfg.ABTesting.Enabled,
			Experiments:         cfg.ABTesting.Experiments,
			AssignmentTTL:       cfg.ABTesting.AssignmentTTL,
			DefaultSalt:         cfg.ABTesting.DefaultSalt,
			MaxIdentifierLength: cfg.ABTesting.MaxIdentifierLength,
		},
		metrics,
		logger,
	)
}

func logStorageInitError(logger *slog.Logger, cfg config.Config, component string, err error) {
	level := slog.LevelWarn
	message := "recommendation.storage.init_degraded"
	if cfg.Storage.FailFast {
		level = slog.LevelError
		message = "recommendation.storage.init_failed"
	}
	logger.Log(context.Background(), level, message,
		slog.String("component", component),
		slog.String("error", err.Error()),
	)
	if cfg.Storage.FailFast {
		os.Exit(1)
	}
}

func domainCacheTTLPolicy(cfg config.Config) domain.CacheTTLPolicy {
	return domain.CacheTTLPolicy{
		Default:      cfg.Storage.Cache.DefaultTTL,
		Personalized: cfg.Storage.Cache.PersonalizedTTL,
		Guest:        cfg.Storage.Cache.GuestTTL,
		RebuildLock:  cfg.Storage.Cache.RebuildLockTTL,
	}
}
