package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/app"
	"github.com/example/ecommerce-platform/backend/services/session-service/internal/config"
	httptransport "github.com/example/ecommerce-platform/backend/services/session-service/internal/transport/http"
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

	deps, err := app.NewDependencies(ctx, cfg, logger)
	if err != nil {
		logger.Error("session.dependencies.init_failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
		defer cancel()
		if err := deps.Close(shutdownCtx); err != nil {
			logger.Error("session.dependencies.close_failed", slog.String("error", err.Error()))
		}
	}()

	server := &http.Server{
		Addr:         cfg.HTTP.Address,
		Handler:      httptransport.NewRouter(deps.HTTPHandler),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	go func() {
		logger.Info("session.http.started", slog.String("addr", cfg.HTTP.Address))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("session.http.listen_failed", slog.String("error", err.Error()))
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("session.http.shutdown_failed", slog.String("error", err.Error()))
	}
}
