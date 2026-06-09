package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
)

func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.ErrorContext(r.Context(), "http_panic_recovered",
						slog.String("panic_type", fmt.Sprintf("%T", recovered)),
						slog.String("request_id", RequestIDFromRequest(r)),
					)
					writeAPIError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
