package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/config"
	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/repository"
	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/scheduler"
	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/usecase"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("cart.config.load_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger = logger.With(
		slog.String("service", cfg.ServiceName),
		slog.String("environment", cfg.Environment),
		slog.String("worker", "cart-expiry-worker"),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	mongoClient, err := connectMongo(ctx, cfg, logger)
	if err != nil {
		logger.Error("cart.mongo.connect_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		disconnectCtx, cancel := context.WithTimeout(context.Background(), cfg.Mongo.ConnectTimeout)
		defer cancel()
		if err := mongoClient.Disconnect(disconnectCtx); err != nil {
			logger.Error("cart.mongo.disconnect_failed", slog.String("error", err.Error()))
		}
	}()

	redisClient := connectRedis(cfg)
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Error("cart.redis.close_failed", slog.String("error", err.Error()))
		}
	}()
	redisPingCtx, cancelRedisPing := context.WithTimeout(ctx, cfg.Redis.PingTimeout)
	defer cancelRedisPing()
	if err := redisClient.Ping(redisPingCtx).Err(); err != nil {
		logger.Error("cart.redis.ping_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	db := mongoClient.Database(cfg.Mongo.Database)
	if cfg.BootstrapOnStartup {
		collectionManager, err := repository.NewMongoCollectionManager(db, logger)
		if err != nil {
			logger.Error("cart.collection_manager.init_failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
		schemaUsecase, err := usecase.NewSchemaUsecase(collectionManager, logger)
		if err != nil {
			logger.Error("cart.schema_usecase.init_failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
		if _, err := schemaUsecase.EnsureCollections(ctx); err != nil {
			logger.Error("cart.schema.bootstrap_startup_failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}

	cartRepository, err := repository.NewMongoCartRepository(db, logger)
	if err != nil {
		logger.Error("cart.repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	cartCache, err := repository.NewRedisCartCache(redisClient, cfg.Cache.ActiveUserTTL, cfg.Cache.ActiveGuestTTL, cfg.Cache.SummaryTTL)
	if err != nil {
		logger.Error("cart.cache.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	cleanupUsecase, err := usecase.NewCleanupExpiredCartsUsecase(usecase.CleanupExpiredCartsDependencies{
		Repository: cartRepository,
		Cache:      cartCache,
		Clock:      usecase.SystemClock{},
		Logger:     logger,
		BatchSize:  cfg.CartExpiry.CleanupBatchSize,
	})
	if err != nil {
		logger.Error("cart.expiry_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	locker, err := scheduler.NewRedisLocker(redisClient)
	if err != nil {
		logger.Error("cart.expiry_locker.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	expiryScheduler, err := scheduler.NewExpiryScheduler(scheduler.ExpirySchedulerDependencies{
		Cleanup: cleanupUsecase,
		Locker:  locker,
		Logger:  logger,
		Config: scheduler.ExpirySchedulerConfig{
			Interval:   cfg.CartExpiry.CleanupInterval,
			RunTimeout: cfg.CartExpiry.CleanupRunTimeout,
			LockTTL:    cfg.CartExpiry.CleanupLockTTL,
			LockKey:    cfg.CartExpiry.CleanupLockKey,
		},
	})
	if err != nil {
		logger.Error("cart.expiry_scheduler.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	result, err := expiryScheduler.RunOnce(ctx)
	if err != nil {
		logger.Error("cart.expiry.worker_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info(
		"cart.expiry.worker_finished",
		slog.Bool("lock_skipped", result.LockSkipped),
		slog.Int("scanned_count", result.Report.ScannedCount),
		slog.Int64("expired_count", result.Report.ExpiredCount),
		slog.Int("cache_delete_failures", result.Report.CacheDeleteFailures),
	)
}

func connectMongo(ctx context.Context, cfg config.Config, logger *slog.Logger) (*mongo.Client, error) {
	connectCtx, cancelConnect := context.WithTimeout(ctx, cfg.Mongo.ConnectTimeout)
	defer cancelConnect()
	client, err := mongo.Connect(connectCtx, options.Client().ApplyURI(cfg.Mongo.URI).SetConnectTimeout(cfg.Mongo.ConnectTimeout))
	if err != nil {
		return nil, err
	}
	pingCtx, cancelPing := context.WithTimeout(ctx, cfg.Mongo.PingTimeout)
	defer cancelPing()
	if err := client.Ping(pingCtx, nil); err != nil {
		disconnectCtx, cancelDisconnect := context.WithTimeout(context.Background(), cfg.Mongo.ConnectTimeout)
		defer cancelDisconnect()
		if disconnectErr := client.Disconnect(disconnectCtx); disconnectErr != nil {
			logger.Error("cart.mongo.disconnect_failed", slog.String("error", disconnectErr.Error()))
		}
		return nil, err
	}
	return client, nil
}

func connectRedis(cfg config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Address,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	})
}
