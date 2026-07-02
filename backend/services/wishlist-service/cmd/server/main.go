package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/clients"
	"ecommerce/backend/services/wishlist-service/internal/config"
	"ecommerce/backend/services/wishlist-service/internal/events"
	"ecommerce/backend/services/wishlist-service/internal/observability"
	"ecommerce/backend/services/wishlist-service/internal/repository"
	httptransport "ecommerce/backend/services/wishlist-service/internal/transport/http"
	"ecommerce/backend/services/wishlist-service/internal/usecase"

	platformmiddleware "github.com/parag/ecommerce/backend/shared/platform/middleware"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	if err := run(); err != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).Error("wishlist service stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := newLogger(cfg.LogLevel)
	logger.Info("starting wishlist service",
		"service", cfg.ServiceName,
		"env", cfg.AppEnv,
		"mongo_database", cfg.Mongo.Database,
		"mongo_collection", cfg.Mongo.Collection,
		"product_service_base_url", cfg.ProductService.BaseURL,
		"cart_service_base_url", cfg.CartService.BaseURL,
		"events_backend", cfg.Events.Backend,
		"analytics_events_enabled", cfg.Events.Analytics.Enabled,
		"analytics_event_publisher", cfg.Events.Analytics.Publisher,
	)

	startupCtx, cancelStartup := context.WithTimeout(context.Background(), cfg.StartupTimeout)
	defer cancelStartup()

	client, err := mongo.Connect(options.Client().
		ApplyURI(cfg.Mongo.URI).
		SetConnectTimeout(cfg.Mongo.ConnectTimeout).
		SetServerSelectionTimeout(cfg.Mongo.ServerSelectionTimeout))
	if err != nil {
		return err
	}
	defer disconnectMongo(client, cfg.ShutdownTimeout, logger)

	wishlistRepository, err := repository.NewMongoWishlistRepository(
		client.Database(cfg.Mongo.Database),
		cfg.Mongo.Collection,
		logger,
	)
	if err != nil {
		return err
	}
	wishlistRepository.SetPriceDropBatchSize(cfg.Events.PriceDrop.BatchSize)

	setupService, err := usecase.NewCollectionSetupService(wishlistRepository, logger)
	if err != nil {
		return err
	}
	if err := setupService.EnsureReady(startupCtx); err != nil {
		return err
	}

	var wishlistEventRepository *repository.MongoWishlistEventRepository
	if cfg.Events.Analytics.Enabled {
		wishlistEventRepository, err = repository.NewMongoWishlistEventRepository(
			client.Database(cfg.Mongo.Database),
			cfg.Events.Analytics.OutboxCollection,
			logger,
		)
		if err != nil {
			return err
		}
		wishlistEventRepository.SetClaimLease(cfg.Events.Analytics.ClaimLease)
		if err := wishlistEventRepository.EnsureCollection(startupCtx); err != nil {
			return err
		}
	}

	productValidator, err := clients.NewHTTPProductValidator(
		cfg.ProductService.BaseURL,
		cfg.ProductService.Timeout,
		logger,
	)
	if err != nil {
		return err
	}
	cartClient, err := clients.NewHTTPCartClient(
		cfg.CartService.BaseURL,
		cfg.CartService.Timeout,
		logger,
	)
	if err != nil {
		return err
	}
	wishlistOptions := []usecase.WishlistServiceOption{
		usecase.WithWishlistEventTopic(cfg.Events.Analytics.Topic),
	}
	if wishlistEventRepository != nil {
		wishlistOptions = append(wishlistOptions, usecase.WithWishlistEventRepository(wishlistEventRepository))
	}
	wishlistService, err := usecase.NewWishlistService(wishlistRepository, productValidator, cartClient, logger, wishlistOptions...)
	if err != nil {
		return err
	}
	metrics := observability.NewMetrics()

	productEventRunner, err := newProductEventRunner(cfg, wishlistService, wishlistRepository, metrics, logger)
	if err != nil {
		return err
	}
	analyticsEventRunner, err := newWishlistAnalyticsRunner(cfg, wishlistEventRepository, metrics, logger)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", metrics.Handler())
	healthHandler := httptransport.NewHealthHandler(cfg.ServiceName, wishlistRepository, logger)
	healthHandler.Register(mux)
	wishlistHandler, err := httptransport.NewWishlistHandler(wishlistService, httptransport.WishlistHandlerConfig{
		UserIDHeader:     cfg.Auth.UserIDHeader,
		RolesHeader:      cfg.Auth.RolesHeader,
		RequestIDHeader:  cfg.Auth.RequestIDHeader,
		RequireBuyerRole: cfg.Auth.RequireBuyerRole,
	}, logger)
	if err != nil {
		return err
	}
	wishlistHandler.Register(mux)

	server := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      platformmiddleware.CORS(platformmiddleware.DefaultCORSConfig())(loggingMiddleware(logger, mux)),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("wishlist service listening", "addr", cfg.HTTP.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	runtimeCtx, cancelRuntime := context.WithCancel(context.Background())
	defer cancelRuntime()
	backgroundRunners := make([]backgroundRunner, 0, 2)
	if productEventRunner != nil {
		backgroundRunners = append(backgroundRunners, productEventRunner)
	}
	if analyticsEventRunner != nil {
		backgroundRunners = append(backgroundRunners, analyticsEventRunner)
	}
	eventsErr := make(chan error, len(backgroundRunners))
	for _, runner := range backgroundRunners {
		go func(runner backgroundRunner) {
			eventsErr <- runner.Run(runtimeCtx)
		}(runner)
	}

	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		return err
	case err := <-eventsErr:
		if err != nil {
			return err
		}
		return errors.New("wishlist background runner stopped unexpectedly")
	case <-stopCtx.Done():
		logger.Info("wishlist service shutdown requested")
	}

	cancelRuntime()
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}
	for _, runner := range backgroundRunners {
		if err := runner.Close(shutdownCtx); err != nil {
			logger.Warn("wishlist background runner close failed", "error", err)
		}
	}
	logger.Info("wishlist service stopped")
	return nil
}

type backgroundRunner interface {
	Run(ctx context.Context) error
	Close(ctx context.Context) error
}

func newProductEventRunner(cfg config.Config, wishlistService *usecase.WishlistService, priceDropRepository usecase.PriceDropRepository, metrics events.ProductEventMetrics, logger *slog.Logger) (backgroundRunner, error) {
	switch cfg.Events.Backend {
	case "", config.EventsBackendDisabled:
		logger.Info("wishlist product event consumer disabled")
		return nil, nil
	case config.EventsBackendKafka:
		notificationPublisher, err := events.NewKafkaNotificationPublisher(events.KafkaNotificationPublisherConfig{
			Brokers: cfg.Events.Kafka.Brokers,
			Topic:   cfg.Events.NotificationCommands.Topic,
		}, logger)
		if err != nil {
			return nil, err
		}
		priceDropService, err := usecase.NewPriceDropService(priceDropRepository, notificationPublisher, logger, usecase.WithPriceDropConfig(usecase.PriceDropConfig{
			TemplateKey:    cfg.Events.PriceDrop.TemplateKey,
			MinDeltaAmount: cfg.Events.PriceDrop.MinDeltaAmount,
		}))
		if err != nil {
			_ = notificationPublisher.Close(context.Background())
			return nil, err
		}
		productConsumer, err := events.NewProductConsumer(wishlistService, logger, events.WithPriceChangeUsecase(priceDropService), events.WithProductEventMetrics(metrics))
		if err != nil {
			_ = notificationPublisher.Close(context.Background())
			return nil, err
		}
		productEventConsumer, err := events.NewKafkaProductEventConsumer(events.KafkaProductEventConsumerConfig{
			Brokers:      cfg.Events.Kafka.Brokers,
			Topic:        cfg.Events.ProductEvents.Topic,
			GroupID:      cfg.Events.ProductEvents.GroupID,
			DLQTopic:     cfg.Events.ProductEvents.DLQTopic,
			MaxAttempts:  cfg.Events.ProductEvents.MaxAttempts,
			RetryBackoff: cfg.Events.ProductEvents.RetryBackoff,
		}, productConsumer, logger)
		if err != nil {
			_ = notificationPublisher.Close(context.Background())
			return nil, err
		}
		return &wishlistEventRuntime{
			productEvents:        productEventConsumer,
			notificationCommands: notificationPublisher,
		}, nil
	default:
		return nil, errors.New("unsupported wishlist events backend")
	}
}

func newWishlistAnalyticsRunner(cfg config.Config, eventRepository events.WishlistOutboxRepository, metrics events.WishlistOutboxMetrics, logger *slog.Logger) (backgroundRunner, error) {
	if !cfg.Events.Analytics.Enabled {
		logger.Info("wishlist analytics events disabled")
		return nil, nil
	}
	if eventRepository == nil {
		return nil, errors.New("wishlist analytics event repository is required")
	}
	switch cfg.Events.Analytics.Publisher {
	case "", config.EventsBackendDisabled:
		logger.Info("wishlist analytics event publisher disabled")
		return nil, nil
	case config.EventsBackendKafka:
		publisher, err := events.NewKafkaWishlistAnalyticsPublisher(events.KafkaWishlistAnalyticsPublisherConfig{
			Brokers: cfg.Events.Kafka.Brokers,
			Topic:   cfg.Events.Analytics.Topic,
		}, logger)
		if err != nil {
			return nil, err
		}
		worker, err := events.NewWishlistOutboxWorker(eventRepository, publisher, logger, events.WishlistOutboxWorkerConfig{
			BatchSize:   cfg.Events.Analytics.BatchSize,
			MaxAttempts: cfg.Events.Analytics.MaxAttempts,
			PollEvery:   cfg.Events.Analytics.PollInterval,
		}, events.WithWishlistOutboxMetrics(metrics))
		if err != nil {
			_ = publisher.Close(context.Background())
			return nil, err
		}
		return &wishlistAnalyticsRuntime{
			worker:    worker,
			publisher: publisher,
		}, nil
	default:
		return nil, errors.New("unsupported wishlist analytics event publisher")
	}
}

type wishlistEventRuntime struct {
	productEvents        backgroundRunner
	notificationCommands interface {
		Close(ctx context.Context) error
	}
}

func (r *wishlistEventRuntime) Run(ctx context.Context) error {
	if r == nil || r.productEvents == nil {
		return nil
	}
	return r.productEvents.Run(ctx)
}

func (r *wishlistEventRuntime) Close(ctx context.Context) error {
	if r == nil {
		return nil
	}
	var closeErr error
	if r.productEvents != nil {
		closeErr = errors.Join(closeErr, r.productEvents.Close(ctx))
	}
	if r.notificationCommands != nil {
		closeErr = errors.Join(closeErr, r.notificationCommands.Close(ctx))
	}
	return closeErr
}

type wishlistAnalyticsRuntime struct {
	worker    *events.WishlistOutboxWorker
	publisher events.WishlistAnalyticsPublisher
}

func (r *wishlistAnalyticsRuntime) Run(ctx context.Context) error {
	if r == nil || r.worker == nil {
		return nil
	}
	return r.worker.Run(ctx)
}

func (r *wishlistAnalyticsRuntime) Close(ctx context.Context) error {
	if r == nil || r.publisher == nil {
		return nil
	}
	return r.publisher.Close(ctx)
}

func newLogger(level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}

func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("http request handled",
			"method", r.Method,
			"path", r.URL.Path,
			"duration_ms", time.Since(startedAt).Milliseconds(),
		)
	})
}

func disconnectMongo(client *mongo.Client, timeout time.Duration, logger *slog.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := client.Disconnect(ctx); err != nil {
		logger.Warn("mongo disconnect failed", "error", err)
	}
}
