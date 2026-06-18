package observability

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	gatewayauth "ecommerce/api-gateway/internal/auth"
)

func AccessLogMiddleware(cfg Config, logger *slog.Logger, labeler RouteLabeler) func(http.Handler) http.Handler {
	cfg = cfg.Normalize(cfg.ServiceName, cfg.Environment)
	if logger == nil {
		logger = slog.Default()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			recorder := NewStatusRecorder(w)
			route := RouteTemplateFromRequest(r, labeler)
			RecordRoute(recorder, route)

			next.ServeHTTP(recorder, r)

			route = fallbackRoute(route, recorder.Route(), RouteTemplateFromRequest(r, labeler))
			attrs := requestLogAttrs(r.Context(), cfg, r.Method, route, recorder, time.Since(started))
			switch {
			case recorder.Status() >= http.StatusInternalServerError:
				logger.ErrorContext(r.Context(), "http_request_failed", attrs...)
			case recorder.Status() >= http.StatusBadRequest:
				logger.WarnContext(r.Context(), "http_request_rejected", attrs...)
			default:
				logger.InfoContext(r.Context(), "http_request_completed", attrs...)
			}
		})
	}
}

func requestLogAttrs(ctx context.Context, cfg Config, method, route string, recorder *StatusRecorder, latency time.Duration) []any {
	errorInfo := recorder.ErrorInfo()
	spanInfo := recorder.SpanInfo()
	attrs := []any{
		"service", cfg.ServiceName,
		"environment", cfg.Environment,
		"request_id", RequestIDFromContext(ctx),
		"method", method,
		"route", emptyToUnknown(route),
		"status", recorder.Status(),
		"response_bytes", recorder.Bytes(),
		"latency_ms", latency.Milliseconds(),
		"error_code", emptyToNone(errorInfo.Code),
	}
	if spanInfo.TraceID != "" {
		attrs = append(attrs, "trace_id", spanInfo.TraceID)
	}
	if spanInfo.SpanID != "" {
		attrs = append(attrs, "span_id", spanInfo.SpanID)
	}
	if errorInfo.GRPCCode != "" {
		attrs = append(attrs, "grpc_code", errorInfo.GRPCCode)
	}
	if errorInfo.DownstreamService != "" {
		attrs = append(attrs, "downstream_service", errorInfo.DownstreamService)
	}
	if recorder.UserIDHash() != "" {
		attrs = append(attrs, "user_id_hash", recorder.UserIDHash())
	} else if userID, ok := gatewayauth.UserIDFromContext(ctx); ok {
		if hashed := HashUserID(userID, cfg.UserHashSalt); hashed != "" {
			attrs = append(attrs, "user_id_hash", hashed)
		}
	}
	return attrs
}

func fallbackRoute(values ...string) string {
	for _, value := range values {
		if value != "" && value != RouteUnknown && value != RouteUnmatched {
			return value
		}
	}
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return RouteUnknown
}
