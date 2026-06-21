package main

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/parag/ecommerce/backend/services/user-service/internal/config"
	"github.com/parag/ecommerce/backend/services/user-service/internal/observability"
	platformhealth "github.com/parag/ecommerce/backend/shared/platform/health"
	platformmetrics "github.com/parag/ecommerce/backend/shared/platform/metrics"
)

func newAdminServer(cfg config.HTTPConfig, db *sql.DB, metrics *observability.Metrics) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("GET /health/live", platformhealth.LivenessHandler("user-service", ""))
	mux.Handle("GET /health/ready", platformhealth.ReadinessHandler("user-service", "", map[string]platformhealth.Check{
		"mysql": db.PingContext,
	}, 2*time.Second))
	mux.Handle("GET /metrics", platformmetrics.Handler(metrics.Registry))

	return &http.Server{
		Addr:              cfg.Address,
		Handler:           mux,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		IdleTimeout:       30 * time.Second,
	}
}

func shutdownAdminServer(ctx context.Context, server *http.Server, timeout time.Duration) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}
