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
	"time"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/clients"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/config"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/events"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/indexer"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/observability"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/repository"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/schema"
	grpctransport "github.com/example/ecommerce-platform/backend/services/search-service/internal/transport/grpc"
	httptransport "github.com/example/ecommerce-platform/backend/services/search-service/internal/transport/http"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/usecase"
	searchv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/search/v1"
	platformmiddleware "github.com/parag/ecommerce/backend/shared/platform/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/typesense/typesense-go/v2/typesense"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("search.config.load_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	metrics := observability.NewMetrics(prometheus.DefaultRegisterer)

	collection, policy, synonyms := schema.MustProductSchemaContract()
	collection.Name = cfg.Typesense.ProductsCollection
	policy.CollectionName = cfg.Typesense.ProductsCollection
	popularQueriesCollection := schema.MustPopularQueriesCollectionSchema()
	popularQueriesCollection.Name = cfg.Typesense.PopularQueriesCollection
	schemaRepo, err := repository.NewStaticSchemaRepository(collection, policy, synonyms)
	if err != nil {
		logger.Error("search.schema_repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	schemaUsecase, err := usecase.NewSchemaUsecase(schemaRepo, logger)
	if err != nil {
		logger.Error("search.schema_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	typesenseClient := typesense.NewClient(
		typesense.WithServer(cfg.Typesense.Endpoint()),
		typesense.WithAPIKey(cfg.Typesense.APIKey),
		typesense.WithConnectionTimeout(cfg.Typesense.RequestTimeout),
	)
	collectionRepo, err := repository.NewTypesenseCollectionRepository(typesenseClient)
	if err != nil {
		logger.Error("search.typesense_collection_repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	ensureCtx, cancelEnsure := context.WithTimeout(ctx, 10*time.Second)
	activeProductsCollection, err := collectionRepo.EnsureAliasedCollection(ensureCtx, cfg.Typesense.ProductsCollection, cfg.Reindex.CollectionPrefix, collection, time.Now())
	if err != nil {
		cancelEnsure()
		logger.Error("search.typesense.collection_ensure_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("search.typesense.collection_ready",
		slog.String("alias", cfg.Typesense.ProductsCollection),
		slog.String("active_collection", activeProductsCollection),
	)
	if err := collectionRepo.EnsureCollection(ensureCtx, popularQueriesCollection); err != nil {
		cancelEnsure()
		logger.Error("search.typesense.popular_queries_collection_ensure_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	cancelEnsure()

	productSearchRepo, err := repository.NewTypesenseProductRepository(typesenseClient, cfg.Typesense.ProductsCollection)
	if err != nil {
		logger.Error("search.product_search_repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	redisClient, err := repository.NewRedisClient(repository.RedisClientConfig{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	})
	if err != nil {
		logger.Error("search.redis_client.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	zeroResultTracker, err := buildZeroResultTracker(cfg, redisClient, metrics, logger)
	if err != nil {
		logger.Error("search.zero_result_tracker.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	productClient, err := clients.NewHTTPProductClient(clients.HTTPProductClientConfig{
		BaseURL:      cfg.Product.URL,
		BatchGetPath: cfg.Product.BatchGetPath,
		ReadyPath:    cfg.Product.ReadyPath,
		Timeout:      cfg.Product.Timeout,
		ServiceToken: cfg.Product.ServiceToken,
	}, nil)
	if err != nil {
		logger.Error("search.product_client.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	searchUsecase, err := usecase.NewSearchProductsUsecase(schemaRepo, productSearchRepo, productClient, usecase.SearchProductsOptions{
		Limits: domain.SearchLimits{
			DefaultPage:     domain.DefaultSearchPage,
			DefaultPageSize: cfg.Search.DefaultPageSize,
			MaxPageSize:     cfg.Search.MaxPageSize,
		},
		TypesenseTimeout:  cfg.Typesense.RequestTimeout,
		HydrationTimeout:  cfg.Product.Timeout,
		ZeroResultTracker: zeroResultTracker,
		Metrics:           metrics,
	}, logger)
	if err != nil {
		logger.Error("search.search_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	autocompleteSearchRepo, err := repository.NewTypesenseAutocompleteRepository(typesenseClient, cfg.Typesense.ProductsCollection, cfg.Typesense.PopularQueriesCollection)
	if err != nil {
		logger.Error("search.autocomplete_repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	autocompleteCache, err := repository.NewRedisAutocompleteCache(redisClient)
	if err != nil {
		logger.Error("search.autocomplete_cache.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	autocompleteUsecase, err := usecase.NewAutocompleteUsecase(autocompleteSearchRepo, autocompleteCache, usecase.AutocompleteOptions{
		DefaultLimit:       cfg.Search.AutocompleteDefaultLimit,
		MaxLimit:           cfg.Search.AutocompleteMaxLimit,
		TypesenseTimeout:   cfg.Search.AutocompleteSearchTimeout,
		CacheTimeout:       cfg.Search.AutocompleteCacheTimeout,
		PrefixCacheTTL:     cfg.Search.AutocompletePrefixCacheTTL,
		EmptyQueryCacheTTL: cfg.Search.AutocompleteEmptyCacheTTL,
		Metrics:            metrics,
	}, logger)
	if err != nil {
		logger.Error("search.autocomplete_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	synonymRepo, err := repository.NewTypesenseSynonymRepository(typesenseClient, cfg.Typesense.ProductsCollection)
	if err != nil {
		logger.Error("search.synonym_repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	createSynonymUsecase, err := usecase.NewCreateSynonymUsecase(synonymRepo, usecase.CreateSynonymOptions{
		Timeout: cfg.Admin.Timeout,
	}, logger)
	if err != nil {
		logger.Error("search.create_synonym_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	updateSynonymUsecase, err := usecase.NewUpdateSynonymUsecase(synonymRepo, usecase.UpdateSynonymOptions{
		Timeout: cfg.Admin.Timeout,
	}, logger)
	if err != nil {
		logger.Error("search.update_synonym_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	deleteSynonymUsecase, err := usecase.NewDeleteSynonymUsecase(synonymRepo, usecase.DeleteSynonymOptions{
		Timeout: cfg.Admin.Timeout,
	}, logger)
	if err != nil {
		logger.Error("search.delete_synonym_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	listSynonymsUsecase, err := usecase.NewListSynonymsUsecase(synonymRepo, usecase.ListSynonymsOptions{
		Timeout: cfg.Admin.Timeout,
	}, logger)
	if err != nil {
		logger.Error("search.list_synonyms_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	productExportClient, err := clients.NewHTTPProductExportClient(clients.HTTPProductExportClientConfig{
		BaseURL:          cfg.Product.URL,
		SearchExportPath: cfg.Product.SearchExportPath,
		Timeout:          cfg.Product.SearchExportTimeout,
		ServiceToken:     cfg.Product.ServiceToken,
	}, nil)
	if err != nil {
		logger.Error("search.product_export_client.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	reindexRepo, err := repository.NewTypesenseReindexRepository(typesenseClient)
	if err != nil {
		logger.Error("search.reindex_repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	reindexLock, err := repository.NewRedisReindexLock(redisClient, "")
	if err != nil {
		logger.Error("search.reindex_lock.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	reindexUsecase, err := usecase.NewReindexCatalogUsecase(schemaRepo, productExportClient, reindexRepo, reindexLock, usecase.ReindexCatalogOptions{
		ProductsAlias:          cfg.Typesense.ProductsCollection,
		DefaultMode:            cfg.Reindex.Mode,
		DefaultBatchSize:       cfg.Reindex.BatchSize,
		MaxBatchSize:           cfg.Reindex.MaxBatchSize,
		CollectionPrefix:       cfg.Reindex.CollectionPrefix,
		LockTTL:                cfg.Reindex.LockTTL,
		JobTimeout:             cfg.Reindex.JobTimeout,
		ProductPageTimeout:     cfg.Product.SearchExportTimeout,
		ImportTimeout:          cfg.Reindex.ImportTimeout,
		OldCollectionRetention: cfg.Reindex.OldCollectionRetention,
		Metrics:                metrics,
	}, logger)
	if err != nil {
		logger.Error("search.reindex_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	reindexStarter, err := usecase.NewAsyncReindexStarter(ctx, reindexUsecase, usecase.AsyncReindexStarterOptions{
		DefaultMode:      cfg.Reindex.Mode,
		DefaultBatchSize: cfg.Reindex.BatchSize,
		MaxBatchSize:     cfg.Reindex.MaxBatchSize,
		JobTimeout:       cfg.Reindex.JobTimeout,
	}, logger)
	if err != nil {
		logger.Error("search.reindex_starter.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	var productConsumer *events.RabbitMQProductConsumer
	handler, err := httptransport.NewHandler(
		schemaUsecase,
		searchUsecase,
		autocompleteUsecase,
		logger,
		httptransport.WithSynonymUsecases(createSynonymUsecase, updateSynonymUsecase, deleteSynonymUsecase, listSynonymsUsecase),
		httptransport.WithReindexStarter(reindexStarter),
		httptransport.WithAdminAuthorizer(httptransport.NewHeaderAdminAuthorizer(cfg.Admin.AuthEnabled)),
		httptransport.WithAdminRateLimiter(httptransport.NewFixedWindowAdminRateLimiter(cfg.Admin.MutationRateLimit, cfg.Admin.MutationRateWindow)),
		httptransport.WithReadinessChecker(httptransport.ReadinessFunc(func(checkCtx context.Context) error {
			if err := redisClient.Ping(checkCtx); err != nil {
				return err
			}
			if _, err := collectionRepo.EnsureAliasedCollection(checkCtx, cfg.Typesense.ProductsCollection, cfg.Reindex.CollectionPrefix, collection, time.Now()); err != nil {
				return err
			}
			if err := collectionRepo.EnsureCollection(checkCtx, popularQueriesCollection); err != nil {
				return err
			}
			if err := productClient.CheckReady(checkCtx); err != nil {
				return err
			}
			if cfg.Indexer.Enabled && (productConsumer == nil || !productConsumer.Ready()) {
				return errors.New("product indexer consumer is not connected")
			}
			return nil
		})),
		httptransport.WithMetricsHandler(promhttp.Handler()),
	)
	if err != nil {
		logger.Error("search.http.handler_init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	grpcHandler, err := grpctransport.NewHandler(searchUsecase, autocompleteUsecase, createSynonymUsecase, updateSynonymUsecase, deleteSynonymUsecase, listSynonymsUsecase)
	if err != nil {
		logger.Error("search.grpc.handler_init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	grpcListener, err := net.Listen("tcp", cfg.GRPC.Address)
	if err != nil {
		logger.Error("search.grpc.listen_failed", slog.String("addr", cfg.GRPC.Address), slog.String("error", err.Error()))
		os.Exit(1)
	}
	grpcServer := grpc.NewServer()
	searchv1.RegisterSearchServiceServer(grpcServer, grpcHandler)
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthv1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(searchv1.SearchService_ServiceDesc.ServiceName, healthv1.HealthCheckResponse_SERVING)
	healthv1.RegisterHealthServer(grpcServer, healthServer)

	if cfg.Indexer.Enabled {
		productConsumer, err = buildProductIndexerConsumer(ctx, cfg, typesenseClient, metrics, logger)
		if err != nil {
			logger.Error("search.product_consumer.init_failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	} else {
		logger.Warn("search.product_consumer.disabled")
	}

	server := &http.Server{
		Addr:         cfg.HTTP.Address,
		Handler:      platformmiddleware.CORS(platformmiddleware.DefaultCORSConfig())(httptransport.NewRouter(handler)),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	go func() {
		logger.Info("search.http.started",
			slog.String("addr", cfg.HTTP.Address),
			slog.String("typesense_endpoint", cfg.Typesense.Endpoint()),
		)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("search.http.listen_failed", slog.String("error", err.Error()))
			stop()
		}
	}()
	go func() {
		logger.Info("search.grpc.started", slog.String("addr", cfg.GRPC.Address))
		if err := grpcServer.Serve(grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			logger.Error("search.grpc.serve_failed", slog.String("error", err.Error()))
			stop()
		}
	}()

	if productConsumer != nil {
		go func() {
			if err := productConsumer.Run(ctx); err != nil {
				logger.Error("search.product_consumer.run_failed", slog.String("error", err.Error()))
				stop()
			}
		}()
	}

	<-ctx.Done()

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancelShutdown()
	healthServer.SetServingStatus("", healthv1.HealthCheckResponse_NOT_SERVING)
	healthServer.SetServingStatus(searchv1.SearchService_ServiceDesc.ServiceName, healthv1.HealthCheckResponse_NOT_SERVING)
	grpcStopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(grpcStopped)
	}()
	select {
	case <-grpcStopped:
	case <-shutdownCtx.Done():
		grpcServer.Stop()
	}
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("search.http.shutdown_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := zeroResultTracker.Close(shutdownCtx); err != nil {
		logger.Warn("search.zero_result_tracker.shutdown_failed", slog.String("error", err.Error()))
	}

	logger.Info("search.http.stopped")
}

func buildZeroResultTracker(cfg config.Config, redisClient *repository.RedisClient, metrics usecase.ZeroResultMetricsRecorder, logger *slog.Logger) (*usecase.ZeroResultTracker, error) {
	options := usecase.ZeroResultTrackerOptions{
		Enabled:     cfg.Search.ZeroResultTrackingEnabled,
		DedupeTTL:   cfg.Search.ZeroResultDedupeTTL,
		SendTimeout: cfg.Search.ZeroResultSendTimeout,
		QueueSize:   cfg.Search.ZeroResultQueueSize,
		WorkerCount: cfg.Search.ZeroResultWorkerCount,
		Metrics:     metrics,
	}
	if !cfg.Search.ZeroResultTrackingEnabled {
		return usecase.NewZeroResultTracker(nil, nil, options, logger)
	}

	sessionClient, err := clients.NewHTTPSessionEventClient(clients.HTTPSessionEventClientConfig{
		BaseURL:    cfg.Session.URL,
		IngestPath: cfg.Session.IngestPath,
		Timeout:    cfg.Session.Timeout,
	}, nil)
	if err != nil {
		return nil, err
	}
	dedupe, err := repository.NewRedisZeroResultDedupeStore(redisClient)
	if err != nil {
		return nil, err
	}
	return usecase.NewZeroResultTracker(dedupe, sessionClient, options, logger)
}

func buildProductIndexerConsumer(ctx context.Context, cfg config.Config, typesenseClient *typesense.Client, metrics *observability.Metrics, logger *slog.Logger) (*events.RabbitMQProductConsumer, error) {
	productRepo, err := repository.NewTypesenseProductRepository(typesenseClient, cfg.Typesense.ProductsCollection)
	if err != nil {
		return nil, err
	}
	productIndexer, err := indexer.NewProductIndexer(productRepo, logger)
	if err != nil {
		return nil, err
	}

	redisClient, err := repository.NewRedisClient(repository.RedisClientConfig{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	})
	if err != nil {
		return nil, err
	}
	pingCtx, cancelPing := context.WithTimeout(ctx, 5*time.Second)
	defer cancelPing()
	if err := redisClient.Ping(pingCtx); err != nil {
		return nil, err
	}

	processedEvents, err := repository.NewRedisProcessedEventStore(redisClient, cfg.Indexer.ProcessedEventTTL)
	if err != nil {
		return nil, err
	}
	return events.NewRabbitMQProductConsumer(events.RabbitMQConsumerConfig{
		URL:                cfg.Queue.RabbitMQURL,
		Exchange:           cfg.Queue.ProductEventsExchange,
		Queue:              cfg.Queue.ProductIndexerQueue,
		DeadLetterExchange: cfg.Queue.ProductIndexerDLX,
		DeadLetterQueue:    cfg.Queue.ProductIndexerDLQ,
		ConsumerTag:        cfg.Queue.ProductIndexerConsumerTag,
		RoutingKeys:        cfg.Queue.ProductIndexerRoutingKeys,
		Prefetch:           cfg.Queue.ProductIndexerPrefetch,
		MaxRetries:         cfg.Queue.ProductIndexerMaxRetries,
		ReconnectDelay:     cfg.Queue.ProductIndexerReconnectDelay,
		RetryBaseDelay:     cfg.Queue.ProductIndexerRetryBaseDelay,
		RetryMaxDelay:      cfg.Queue.ProductIndexerRetryMaxDelay,
		MessageTimeout:     cfg.Indexer.MessageTimeout,
		Metrics:            metrics,
	}, productIndexer, processedEvents, logger)
}
