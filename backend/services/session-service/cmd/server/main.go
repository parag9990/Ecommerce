package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/config"
	"github.com/example/ecommerce-platform/backend/services/session-service/internal/repository"
	httptransport "github.com/example/ecommerce-platform/backend/services/session-service/internal/transport/http"
	"github.com/example/ecommerce-platform/backend/services/session-service/internal/usecase"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("session.config.load_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	mongoClient, err := connectMongo(ctx, cfg)
	if err != nil {
		logger.Error("session.mongo.connect_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		disconnectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mongoClient.Disconnect(disconnectCtx); err != nil {
			logger.Error("session.mongo.disconnect_failed", slog.String("error", err.Error()))
		}
	}()

	redisClient, err := repository.NewRedisClient(repository.RedisConfig{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	})
	if err != nil {
		logger.Error("session.redis.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	redisCtx, cancelRedis := context.WithTimeout(ctx, 5*time.Second)
	defer cancelRedis()
	if err := redisClient.Ping(redisCtx).Err(); err != nil {
		logger.Error("session.redis.ping_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer redisClient.Close()

	privacyRepo, err := repository.NewMongoPrivacyRepository(mongoClient.Database(cfg.Mongo.Database))
	if err != nil {
		logger.Error("session.privacy_repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	indexCtx, cancelIndex := context.WithTimeout(ctx, 10*time.Second)
	defer cancelIndex()
	if err := privacyRepo.EnsureIndexes(indexCtx); err != nil {
		logger.Error("session.privacy_repository.index_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	activeRepo, err := repository.NewRedisActiveSessionRepository(redisClient, cfg.Redis.KeyPrefix)
	if err != nil {
		logger.Error("session.active_session_repository.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	privacyUsecase, err := usecase.NewPrivacyUsecase(
		privacyRepo,
		activeRepo,
		usecase.PrivacyConfig{
			HashPepper:        cfg.Privacy.HashPepper,
			DeletionListLimit: cfg.Privacy.DeletionListLimit,
		},
		logger,
	)
	if err != nil {
		logger.Error("session.privacy_usecase.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	handler, err := httptransport.NewHandler(privacyUsecase, logger)
	if err != nil {
		logger.Error("session.http.handler_init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	server := &http.Server{
		Addr:         cfg.HTTP.Address,
		Handler:      httptransport.NewRouter(handler),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	go func() {
		logger.Info("session.http.starting", slog.String("addr", cfg.HTTP.Address))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("session.http.failed", slog.String("error", err.Error()))
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("session.http.shutdown_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("session.http.stopped")
}

func connectMongo(ctx context.Context, cfg config.Config) (*mongo.Client, error) {
	connectCtx, cancel := context.WithTimeout(ctx, cfg.Mongo.ConnectTimeout)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(cfg.Mongo.URI))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(connectCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}
	return client, nil
}
