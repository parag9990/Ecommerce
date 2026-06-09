package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/parag/ecommerce/backend/services/api-gateway/internal/clients"
	"github.com/parag/ecommerce/backend/services/api-gateway/internal/config"
	"github.com/parag/ecommerce/backend/services/api-gateway/internal/handlers"
	"github.com/parag/ecommerce/backend/services/api-gateway/internal/middleware"
	"github.com/parag/ecommerce/backend/services/api-gateway/internal/routes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if err := run(context.Background()); err != nil {
		slog.Error("api_gateway_exit", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := newLogger(cfg.Log.Level)
	logger.InfoContext(ctx, "api_gateway_starting", slog.String("config", cfg.String()))

	userConn, err := dialGRPC(ctx, cfg.UserService.Address)
	if err != nil {
		return err
	}
	defer userConn.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	userHandler := handlers.NewUserHandler(
		clients.NewGRPCUserClient(userConn),
		logger,
		handlers.WithTimeout(cfg.UserService.RequestTimeout),
		handlers.WithValidationPhoneRegion(cfg.Validation.PhoneRegion),
	)
	auth := middleware.NewAuthMiddleware(middleware.NewJWTVerifier(
		cfg.Auth.JWTHS256Secret,
		cfg.Auth.Issuer,
		cfg.Auth.Audience,
		cfg.Auth.ClockSkew,
		nil,
	))
	rbac := middleware.NewRBACMiddleware(logger)
	routes.RegisterUserRoutes(mux, userHandler, auth, rbac)

	handler := chain(
		mux,
		middleware.RequestID,
		middleware.Recovery(logger),
		middleware.Logging(logger),
	)

	server := &http.Server{
		Addr:              cfg.HTTP.Address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		logger.InfoContext(ctx, "api_gateway_http_listening", slog.String("address", cfg.HTTP.Address))
		serveErr <- server.ListenAndServe()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signals)

	select {
	case sig := <-signals:
		logger.InfoContext(ctx, "api_gateway_shutdown_requested", slog.String("signal", sig.String()))
		shutdownCtx, cancel := context.WithTimeout(ctx, cfg.HTTP.ShutdownTimeout)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve http: %w", err)
	}
}

func dialGRPC(ctx context.Context, address string) (*grpc.ClientConn, error) {
	dialCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		dialCtx,
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("dial user service at %s: %w", address, err)
	}
	return conn, nil
}

func chain(handler http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	return handler
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
