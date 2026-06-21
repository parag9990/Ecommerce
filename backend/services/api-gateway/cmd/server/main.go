package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"ecommerce/api-gateway/internal/clients"
	"ecommerce/api-gateway/internal/config"
	"ecommerce/api-gateway/internal/logger"
	"ecommerce/api-gateway/internal/observability"
	"ecommerce/api-gateway/internal/ratelimit"
	"ecommerce/api-gateway/internal/repository"
	gatewayserver "ecommerce/api-gateway/internal/server"
	grpcwebtransport "ecommerce/api-gateway/internal/transport/grpcweb"
	httptransport "ecommerce/api-gateway/internal/transport/http"
	"ecommerce/api-gateway/internal/usecase"
	"github.com/redis/go-redis/v9"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "api-gateway failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load(ctx)
	if err != nil {
		return err
	}

	log := logger.New(cfg.LogLevel)
	observabilityConfig := cfg.Observability.Normalize(cfg.ServiceName, cfg.Environment)
	tracingProvider, err := observability.InitTracing(ctx, observabilityConfig)
	if err != nil {
		return fmt.Errorf("initialize tracing: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), observabilityConfig.TraceShutdownTimeout)
		defer cancel()
		if err := tracingProvider.Shutdown(shutdownCtx); err != nil {
			log.ErrorContext(context.Background(), "tracing_shutdown_failed", "error", err)
		}
	}()

	metrics, err := observability.NewMetrics(observabilityConfig, nil)
	if err != nil {
		return fmt.Errorf("initialize metrics: %w", err)
	}
	metricsServer := gatewayserver.NewMetricsServer(observabilityConfig, metrics)
	var metricsErrCh <-chan error
	if metricsServer != nil {
		metricsErrCh = gatewayserver.StartMetricsServer(ctx, metricsServer, log)
	}

	routeRepo := repository.NewJSONRouteRepository(cfg.APIContractPath)
	routeCatalog := usecase.NewRouteCatalogService(routeRepo, cfg.APIBasePath)
	if err := routeCatalog.Load(ctx); err != nil {
		return fmt.Errorf("load route catalog: %w", err)
	}

	var rateLimiter ratelimit.Limiter
	var redisClient *redis.Client
	if cfg.RateLimit.Enabled {
		redisOptions := &redis.Options{
			Addr:        cfg.Redis.Addr,
			Password:    cfg.Redis.Password,
			DB:          cfg.Redis.DB,
			DialTimeout: cfg.Redis.DialTimeout,
		}
		if cfg.Redis.TLSEnabled {
			redisOptions.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		}
		redisClient = redis.NewClient(redisOptions)
		redisCtx, cancel := context.WithTimeout(ctx, cfg.Redis.DialTimeout)
		if err := redisClient.Ping(redisCtx).Err(); err != nil {
			cancel()
			_ = redisClient.Close()
			return fmt.Errorf("redis ping failed: %w", err)
		}
		cancel()
		rateLimiter = ratelimit.NewRedisLimiter(redisClient)
		defer func() {
			if err := redisClient.Close(); err != nil {
				log.ErrorContext(context.Background(), "redis_client_close_failed", "error", err)
			}
		}()
	}

	grpcClients, err := clients.New(ctx, cfg, log, clients.WithMetrics(metrics))
	if err != nil {
		return fmt.Errorf("initialize grpc clients: %w", err)
	}
	defer func() {
		if err := grpcClients.Close(); err != nil {
			log.ErrorContext(context.Background(), "grpc_clients_close_failed", "error", err)
		}
	}()

	var grpcWebServer *grpcwebtransport.Server
	var grpcWebErrCh <-chan error
	if cfg.GRPCWeb.Enabled {
		policyRepo := repository.NewJSONGRPCWebPolicyRepository(cfg.GRPCWeb.PolicyPath)
		policyCatalog := usecase.NewGRPCWebPolicyCatalogService(policyRepo, cfg.GRPCWeb.ExposedServices)
		if err := policyCatalog.Load(ctx); err != nil {
			return fmt.Errorf("load grpc-web policy catalog: %w", err)
		}
		requiresToken, err := policyCatalog.RequiresToken(ctx)
		if err != nil {
			return fmt.Errorf("inspect grpc-web policy catalog: %w", err)
		}
		var tokenVerifier grpcwebtransport.TokenVerifier
		if requiresToken {
			tokenVerifier, err = grpcwebtransport.NewTokenVerifier(cfg, log)
			if err != nil {
				return err
			}
		}
		grpcWebServer, err = grpcwebtransport.NewServer(ctx, grpcwebtransport.ServerOptions{
			Config:        cfg.GRPCWeb,
			ServiceName:   cfg.ServiceName,
			Observability: observabilityConfig,
			Policies:      policyCatalog,
			Connections:   grpcClients,
			TokenVerifier: tokenVerifier,
			Logger:        log,
			Metrics:       metrics,
		})
		if err != nil {
			return fmt.Errorf("initialize grpc-web facade: %w", err)
		}
		grpcWebServeCh := make(chan error, 1)
		grpcWebErrCh = grpcWebServeCh
		go func() {
			log.InfoContext(ctx, "grpc_web_facade_starting",
				"addr", grpcWebServer.Address(),
				"policy_path", cfg.GRPCWeb.PolicyPath,
			)
			grpcWebServeCh <- grpcWebServer.Serve()
		}()
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
			defer cancel()
			if err := grpcWebServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, context.Canceled) {
				log.ErrorContext(context.Background(), "grpc_web_facade_shutdown_failed", "error", err)
			}
		}()
	}

	router, err := httptransport.NewRouterWithOptions(ctx, cfg, routeCatalog, log, grpcClients, httptransport.RouterOptions{
		RateLimiter:  rateLimiter,
		Metrics:      metrics,
		UserClient:   grpcClients.User,
		SearchClient: grpcClients.Search,
	})
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           router,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		log.InfoContext(ctx, "api_gateway_starting",
			"addr", cfg.HTTPAddress,
			"environment", cfg.Environment,
			"contract_path", cfg.APIContractPath,
		)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		log.InfoContext(shutdownCtx, "api_gateway_stopping")
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown http server: %w", err)
		}
		if grpcWebServer != nil {
			if err := grpcWebServer.Shutdown(shutdownCtx); err != nil {
				return fmt.Errorf("shutdown grpc-web facade: %w", err)
			}
		}
		if metricsServer != nil {
			if err := metricsServer.Shutdown(shutdownCtx); err != nil {
				return fmt.Errorf("shutdown metrics server: %w", err)
			}
		}
		if err := <-errCh; err != nil {
			return err
		}
		if metricsErrCh != nil {
			if err := <-metricsErrCh; err != nil {
				return fmt.Errorf("serve metrics: %w", err)
			}
		}
		if grpcWebErrCh != nil {
			if err := <-grpcWebErrCh; err != nil {
				return fmt.Errorf("serve grpc-web facade: %w", err)
			}
		}
		log.InfoContext(context.Background(), "api_gateway_stopped")
		return nil
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("serve http: %w", err)
		}
		return nil
	case err := <-metricsErrCh:
		if err != nil {
			return fmt.Errorf("serve metrics: %w", err)
		}
		return nil
	case err := <-grpcWebErrCh:
		if err != nil {
			return fmt.Errorf("serve grpc-web facade: %w", err)
		}
		return nil
	}
}
