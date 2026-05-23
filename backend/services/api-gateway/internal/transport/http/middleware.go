package httptransport

import (
	"context"
	"log/slog"
	"net/http"

	gatewayerrors "ecommerce/api-gateway/internal/errors"
	"ecommerce/api-gateway/internal/observability"
)

const requestIDHeader = observability.HeaderRequestID

func RequestIDMiddleware(next http.Handler) http.Handler {
	return observability.RequestIDMiddleware(requestIDHeader)(next)
}

func AccessLogMiddleware(log *slog.Logger) func(http.Handler) http.Handler {
	return AccessLogMiddlewareWithConfig(observability.DefaultConfig("api-gateway", "local"), log, nil)
}

func AccessLogMiddlewareWithConfig(cfg observability.Config, log *slog.Logger, labeler observability.RouteLabeler) func(http.Handler) http.Handler {
	return observability.AccessLogMiddleware(cfg, log, labeler)
}

func MetricsMiddleware(metrics *observability.Metrics, labeler observability.RouteLabeler) func(http.Handler) http.Handler {
	return observability.MetricsMiddleware(metrics, labeler)
}

func TracingMiddleware(cfg observability.Config, labeler observability.RouteLabeler) func(http.Handler) http.Handler {
	return observability.TracingMiddleware(cfg, labeler)
}

func RecoveryMiddleware(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					if log != nil {
						log.ErrorContext(r.Context(), "http_panic_recovered",
							"panic", recovered,
							"request_id", RequestIDFromContext(r.Context()),
						)
					}
					writeMappedError(w, r, gatewayerrors.New(http.StatusInternalServerError, gatewayerrors.CodeInternal, "Internal server error", nil, nil))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func RequestIDFromContext(ctx context.Context) string {
	return observability.RequestIDFromContext(ctx)
}
