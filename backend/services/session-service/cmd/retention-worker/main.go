package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/app"
	"github.com/example/ecommerce-platform/backend/services/session-service/internal/config"
	"github.com/example/ecommerce-platform/backend/services/session-service/internal/worker"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("session.retention.config.load_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	deps, err := app.NewRetentionDependencies(ctx, cfg, logger)
	if err != nil {
		logger.Error("session.retention.dependencies.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
		defer cancel()
		if err := deps.Close(shutdownCtx); err != nil {
			logger.Error("session.retention.dependencies.close_failed", slog.String("error", err.Error()))
		}
	}()

	retentionWorker := worker.NewRetentionWorker(deps.RetentionUsecase, cfg.Retention.WorkerInterval, logger)
	if cfg.Retention.WorkerRunOnce {
		if err := retentionWorker.RunOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("session.retention.worker_run_failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
		return
	}
	if err := retentionWorker.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("session.retention.worker_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
