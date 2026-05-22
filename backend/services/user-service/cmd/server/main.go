package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/parag/ecommerce/backend/services/user-service/internal/config"
	"github.com/parag/ecommerce/backend/services/user-service/internal/repository"
	transportgrpc "github.com/parag/ecommerce/backend/services/user-service/internal/transport/grpc"
	"github.com/parag/ecommerce/backend/services/user-service/internal/usecase"
	userv1 "github.com/parag/ecommerce/backend/shared/gen/go/ecommerce/user/v1"
	grpcgo "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := run(context.Background()); err != nil {
		slog.Error("user_service_exit", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := newLogger(cfg.Log.Level)
	logger.InfoContext(ctx, "user_service_starting", slog.String("config", cfg.String()))

	db, err := openDatabase(ctx, cfg.Database)
	if err != nil {
		return err
	}
	defer db.Close()

	userRepo, err := repository.NewMySQLUserRepository(db, repository.WithLogger(logger))
	if err != nil {
		return err
	}
	sellerRepo, err := repository.NewMySQLSellerRepository(db, repository.WithLogger(logger))
	if err != nil {
		return err
	}
	userService, err := usecase.NewService(userRepo, sellerRepo, usecase.WithLogger(logger))
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", cfg.GRPC.Address)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", cfg.GRPC.Address, err)
	}

	grpcServer := grpcgo.NewServer(transportgrpc.ServerOptions(logger)...)
	userv1.RegisterUserServiceServer(grpcServer, transportgrpc.NewServer(userService, transportgrpc.WithLogger(logger)))
	if cfg.GRPC.Reflection {
		reflection.Register(grpcServer)
	}

	serveErr := make(chan error, 1)
	go func() {
		logger.InfoContext(ctx, "user_service_grpc_listening", slog.String("address", cfg.GRPC.Address))
		serveErr <- grpcServer.Serve(listener)
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	select {
	case sig := <-signals:
		logger.InfoContext(ctx, "user_service_shutdown_requested", slog.String("signal", sig.String()))
		return gracefulStop(grpcServer, cfg.GRPC.ShutdownTimeout)
	case err := <-serveErr:
		if errors.Is(err, grpcgo.ErrServerStopped) {
			return nil
		}
		return fmt.Errorf("serve grpc: %w", err)
	}
}

func openDatabase(ctx context.Context, cfg config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	pingCtx, cancel := context.WithTimeout(ctx, cfg.PingTimeout)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	return db, nil
}

func gracefulStop(server *grpcgo.Server, timeout time.Duration) error {
	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-done:
		return nil
	case <-timer.C:
		server.Stop()
		return errors.New("grpc graceful shutdown timed out")
	}
}

func newLogger(level string) *slog.Logger {
	var slogLevel slog.Level
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn", "warning":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel}))
}
