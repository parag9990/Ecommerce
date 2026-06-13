package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/typesense/typesense-go/v2/typesense"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/config"
	"github.com/example/ecommerce-platform/backend/services/search-service/internal/repository"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("search.reindex_alias.config_load_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	alias := flag.String("alias", cfg.Typesense.ProductsCollection, "Typesense alias to move")
	target := flag.String("target", "", "target collection for the alias")
	reason := flag.String("reason", "rollback", "audit reason for the alias move")
	actorID := flag.String("actor-id", "cli", "actor id recorded in logs")
	flag.Parse()

	targetCollection := strings.TrimSpace(*target)
	if targetCollection == "" {
		logger.Error("search.reindex_alias.invalid_request", slog.String("error", "--target is required"))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, cfg.Reindex.ImportTimeout)
	defer cancel()

	typesenseClient := typesense.NewClient(
		typesense.WithServer(cfg.Typesense.Endpoint()),
		typesense.WithAPIKey(cfg.Typesense.APIKey),
		typesense.WithConnectionTimeout(cfg.Typesense.RequestTimeout),
	)
	repo, err := repository.NewTypesenseReindexRepository(typesenseClient)
	if err != nil {
		logger.Error("search.reindex_alias.repository_init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	previous, err := repo.ResolveAlias(ctx, *alias)
	if err != nil {
		logger.Error("search.reindex_alias.resolve_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := repo.SwapAlias(ctx, *alias, targetCollection); err != nil {
		logger.Error("search.reindex_alias.swap_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("search.reindex_alias.completed",
		slog.String("alias", *alias),
		slog.String("target_collection", targetCollection),
		slog.String("previous_collection", previous),
		slog.String("reason", *reason),
		slog.String("actor_id", *actorID),
	)
}
