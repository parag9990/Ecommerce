package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	platformlog "github.com/parag/ecommerce/backend/shared/platform/logger"
)

const RequestIDHeader = "X-Request-ID"

const (
	defaultCORSAllowedMethods = "GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS"
	defaultCORSAllowedHeaders = "Accept, Authorization, Content-Type, Idempotency-Key, X-Client-App, X-Request-ID, X-Request-Source, X-CSRF-Token"
	defaultCORSExposedHeaders = "Content-Disposition, X-Request-ID"
)

var defaultCORSAllowedOrigins = []string{
	"http://localhost:3000",
	"http://localhost:3001",
	"http://localhost:3002",
	"http://localhost:3003",
	"http://localhost:5173",
	"http://127.0.0.1:3000",
	"http://127.0.0.1:3001",
	"http://127.0.0.1:3002",
	"http://127.0.0.1:3003",
	"http://127.0.0.1:5173",
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   string
	AllowedHeaders   string
	ExposedHeaders   string
	AllowCredentials bool
	MaxAgeSeconds    int
}

func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   corsOriginsFromEnv(),
		AllowedMethods:   defaultCORSAllowedMethods,
		AllowedHeaders:   defaultCORSAllowedHeaders,
		ExposedHeaders:   defaultCORSExposedHeaders,
		AllowCredentials: true,
		MaxAgeSeconds:    600,
	}
}

func CORS(cfg CORSConfig) func(http.Handler) http.Handler {
	allowedOrigins := make(map[string]struct{}, len(cfg.AllowedOrigins))
	for _, origin := range cfg.AllowedOrigins {
		if origin = normalizeOrigin(origin); origin != "" && origin != "*" {
			allowedOrigins[origin] = struct{}{}
		}
	}
	allowedMethods := valueOrDefault(cfg.AllowedMethods, defaultCORSAllowedMethods)
	allowedHeaders := valueOrDefault(cfg.AllowedHeaders, defaultCORSAllowedHeaders)
	exposedHeaders := valueOrDefault(cfg.ExposedHeaders, defaultCORSExposedHeaders)
	maxAge := cfg.MaxAgeSeconds
	if maxAge <= 0 {
		maxAge = 600
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := normalizeOrigin(r.Header.Get("Origin"))
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Add("Vary", "Origin")
			if r.Method == http.MethodOptions {
				w.Header().Add("Vary", "Access-Control-Request-Method")
				w.Header().Add("Vary", "Access-Control-Request-Headers")
			}

			if _, allowed := allowedOrigins[origin]; !allowed {
				if r.Method == http.MethodOptions {
					http.Error(w, "origin is not allowed", http.StatusForbidden)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", origin)
			if cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			w.Header().Set("Access-Control-Expose-Headers", exposedHeaders)
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
				w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(maxAge))
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

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

func corsOriginsFromEnv() []string {
	value := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if value == "" {
		return append([]string(nil), defaultCORSAllowedOrigins...)
	}
	var origins []string
	for _, item := range strings.Split(value, ",") {
		if item = normalizeOrigin(item); item != "" {
			origins = append(origins, item)
		}
	}
	return origins
}

func normalizeOrigin(origin string) string {
	return strings.TrimSuffix(strings.TrimSpace(origin), "/")
}

func valueOrDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
