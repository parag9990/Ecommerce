package httptransport

import (
	"log/slog"
	"net/http"
	"strings"

	"ecommerce/api-gateway/internal/domain"
	"ecommerce/api-gateway/internal/observability"
	"ecommerce/api-gateway/internal/validation"
)

type RequestValidator interface {
	Validate(w http.ResponseWriter, r *http.Request, route domain.RouteDefinition) (*http.Request, *validation.Error)
}

func RequestValidationMiddleware(route domain.RouteDefinition, validator RequestValidator, logger *slog.Logger, metrics *observability.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if validator == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			validated, err := validator.Validate(w, r, route)
			if err != nil {
				observeValidationFailure(metrics, route, err)
				logValidationFailure(logger, r, route, err)
				writeValidationError(w, validated, err)
				return
			}
			next.ServeHTTP(w, validated)
		})
	}
}

func writeValidationError(w http.ResponseWriter, r *http.Request, err *validation.Error) {
	if err == nil {
		return
	}
	writeMappedError(w, r, err)
}

func logValidationFailure(logger *slog.Logger, r *http.Request, route domain.RouteDefinition, err *validation.Error) {
	if logger == nil || err == nil {
		return
	}
	logger.WarnContext(r.Context(), "request_validation_failed",
		"route_id", route.ID,
		"method", r.Method,
		"route", route.Path,
		"status", err.Status,
		"detail_count", len(err.Details),
		"request_id", RequestIDFromContext(r.Context()),
	)
}

func observeValidationFailure(metrics *observability.Metrics, route domain.RouteDefinition, err *validation.Error) {
	if metrics == nil || err == nil {
		return
	}
	fieldGroup := "request"
	if len(err.Details) > 0 && err.Details[0].Field != "" {
		field := strings.ToLower(err.Details[0].Field)
		switch {
		case field == "body" || strings.Contains(field, "."):
			fieldGroup = "body"
		case field == "query":
			fieldGroup = "query"
		case strings.Contains(field, "header") || strings.Contains(field, "content-type") || strings.Contains(field, "idempotency-key"):
			fieldGroup = "header"
		default:
			fieldGroup = "field"
		}
	}
	metrics.ObserveValidationFailure(route.Path, fieldGroup)
}
