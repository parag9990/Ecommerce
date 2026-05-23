package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"ecommerce/api-gateway/internal/observability"
)

func NewMetricsServer(cfg observability.Config, metrics *observability.Metrics) *http.Server {
	if metrics == nil || !cfg.MetricsEnabled {
		return nil
	}
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", metrics.Handler())
	return &http.Server{
		Addr:    cfg.MetricsAddress,
		Handler: mux,
	}
}

func StartMetricsServer(ctx context.Context, srv *http.Server, logger *slog.Logger) <-chan error {
	errCh := make(chan error, 1)
	if srv == nil {
		errCh <- nil
		close(errCh)
		return errCh
	}
	go func() {
		if logger != nil {
			logger.InfoContext(ctx, "metrics_server_starting", "addr", srv.Addr)
		}
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()
	return errCh
}
