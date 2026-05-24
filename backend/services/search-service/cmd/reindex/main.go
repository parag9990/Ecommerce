package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/typesense/typesense-go/v2/typesense"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/clients"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/config"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/repository"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/schema"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("search.reindex.config_load_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	req := parseFlags(cfg)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, cfg.Reindex.JobTimeout)
	defer cancel()

	reindexUsecase, err := buildReindexUsecase(cfg, logger)
	if err != nil {
		logger.Error("search.reindex.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	result, err := reindexUsecase.Execute(ctx, req)
	if err != nil {
		logger.Error("search.reindex.failed",
			slog.String("error", err.Error()),
			slog.String("mode", req.Mode),
			slog.String("reason", req.Reason),
		)
		os.Exit(1)
	}

	logger.Info("search.reindex.completed",
		slog.String("job_id", result.JobID),
		slog.String("mode", result.Mode),
		slog.String("target_collection", result.TargetCollection),
		slog.String("previous_collection", result.PreviousCollection),
		slog.Int("products_read", result.ProductsRead),
		slog.Int("products_indexed", result.ProductsIndexed),
		slog.Int("products_skipped", result.ProductsSkipped),
		slog.Bool("alias_swapped", result.AliasSwapped),
		slog.Bool("dry_run", result.DryRun),
		slog.Int64("duration_ms", result.Duration.Milliseconds()),
	)
}

func parseFlags(cfg config.Config) domain.ReindexRequest {
	mode := flag.String("mode", cfg.Reindex.Mode, "reindex mode: alias or in_place")
	batchSize := flag.Int("batch-size", cfg.Reindex.BatchSize, "product export page size")
	reason := flag.String("reason", "manual", "audit reason for the reindex")
	actorID := flag.String("actor-id", "cli", "actor id recorded in reindex logs")
	dryRun := flag.Bool("dry-run", false, "read and validate Product Service pages without writing or swapping aliases")
	target := flag.String("target", "", "optional target collection name")
	flag.Parse()

	return domain.ReindexRequest{
		Mode:             *mode,
		BatchSize:        *batchSize,
		Reason:           *reason,
		ActorID:          *actorID,
		DryRun:           *dryRun,
		TargetCollection: *target,
	}
}

func buildReindexUsecase(cfg config.Config, logger *slog.Logger) (*usecase.ReindexCatalogUsecase, error) {
	collection, policy, synonyms := schema.MustProductSchemaContract()
	collection.Name = cfg.Typesense.ProductsCollection
	policy.CollectionName = cfg.Typesense.ProductsCollection
	schemaRepo, err := repository.NewStaticSchemaRepository(collection, policy, synonyms)
	if err != nil {
		return nil, err
	}

	typesenseClient := typesense.NewClient(
		typesense.WithServer(cfg.Typesense.Endpoint()),
		typesense.WithAPIKey(cfg.Typesense.APIKey),
		typesense.WithConnectionTimeout(cfg.Typesense.RequestTimeout),
	)
	reindexRepo, err := repository.NewTypesenseReindexRepository(typesenseClient)
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
	pingCtx, cancelPing := context.WithTimeout(context.Background(), cfg.Redis.DialTimeout+cfg.Redis.ReadTimeout)
	defer cancelPing()
	if err := redisClient.Ping(pingCtx); err != nil {
		return nil, err
	}
	reindexLock, err := repository.NewRedisReindexLock(redisClient, "")
	if err != nil {
		return nil, err
	}

	productExportClient, err := clients.NewHTTPProductExportClient(clients.HTTPProductExportClientConfig{
		BaseURL:          cfg.Product.URL,
		SearchExportPath: cfg.Product.SearchExportPath,
		Timeout:          cfg.Product.SearchExportTimeout,
	}, nil)
	if err != nil {
		return nil, err
	}

	return usecase.NewReindexCatalogUsecase(schemaRepo, productExportClient, reindexRepo, reindexLock, usecase.ReindexCatalogOptions{
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
	}, logger)
}
