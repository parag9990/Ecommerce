package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	platformlog "github.com/parag/ecommerce/backend/shared/platform/logger"
)

const RequestIDHeader = "X-Request-ID"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := strings.TrimSpace(r.Header.Get(RequestIDHeader))
		if requestID == "" || len(requestID) > 128 {
			requestID = randomID()
		}
		w.Header().Set(RequestIDHeader, requestID)
		next.ServeHTTP(w, r.WithContext(platformlog.WithRequestID(r.Context(), requestID)))
	})
}

// RequestIDFromContext returns the correlation identifier installed by RequestID.
func RequestIDFromContext(ctx context.Context) string {
	return platformlog.RequestIDFromContext(ctx)
}

func Recover(logger *slog.Logger, next http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.ErrorContext(r.Context(), "http.panic_recovered", "method", r.Method, "path", r.URL.Path, "stack", string(debug.Stack()))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":{"code":"INTERNAL_ERROR","message":"internal server error"}}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func AccessLog(logger *slog.Logger, next http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		platformlog.FromContext(r.Context(), logger).InfoContext(r.Context(), "http.request", "method", r.Method, "route", r.URL.Path, "status", recorder.status, "duration_ms", time.Since(started).Milliseconds())
	})
}

type RateLimiter interface {
	Allow(context.Context, string) (bool, error)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func Chain(handler http.Handler, wrappers ...func(http.Handler) http.Handler) http.Handler {
	for index := len(wrappers) - 1; index >= 0; index-- {
		handler = wrappers[index](handler)
	}
	return handler
}

func randomID() string {
	var bytes [16]byte
	_, _ = rand.Read(bytes[:])
	return hex.EncodeToString(bytes[:])
}
