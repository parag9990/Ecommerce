package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/config"
	"github.com/example/ecommerce-platform/backend/services/session-service/internal/repository"
	"github.com/example/ecommerce-platform/backend/services/session-service/internal/usecase"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type RetentionDependencies struct {
	MongoClient      *mongo.Client
	RedisClient      *redis.Client
	RetentionRepo    *repository.MongoRetentionRepository
	ActiveStore      *repository.RedisActiveSessionStore
	RetentionUsecase *usecase.RetentionUsecase
}

func NewRetentionDependencies(ctx context.Context, cfg config.Config, logger *slog.Logger) (*RetentionDependencies, error) {
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
	retentionRepo, err := repository.NewMongoRetentionRepository(database, logger)
	if err != nil {
		return nil, err
	}
	if err := retentionRepo.EnsureIndexes(ctx); err != nil {
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
	retentionUsecase, err := usecase.NewRetentionUsecase(retentionRepo, activeStore, cfg.RetentionUsecaseConfig(), logger)
	if err != nil {
		return nil, err
	}

	cleanupMongo = false
	cleanupRedis = false
	return &RetentionDependencies{
		MongoClient:      mongoClient,
		RedisClient:      redisClient,
		RetentionRepo:    retentionRepo,
		ActiveStore:      activeStore,
		RetentionUsecase: retentionUsecase,
	}, nil
}

func (d *RetentionDependencies) Close(ctx context.Context) error {
	if d == nil {
		return nil
	}
	var combined error
	if d.RedisClient != nil {
		if err := d.RedisClient.Close(); err != nil {
			combined = errors.Join(combined, fmt.Errorf("close redis: %w", err))
		}
	}
	if d.MongoClient != nil {
		if err := d.MongoClient.Disconnect(ctx); err != nil {
			combined = errors.Join(combined, fmt.Errorf("disconnect mongo: %w", err))
		}
	}
	return combined
}
