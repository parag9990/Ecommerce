// Command local-server provides dependency probes while the two incomplete
// Session Service runtime implementations are reconciled.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mongoClient, err := mongo.Connect(options.Client().ApplyURI(env("SESSION_MONGO_URI", "mongodb://localhost:27017")))
	if err != nil {
		logger.Error("session.local.mongo_connect_failed", "error", err)
		os.Exit(1)
	}
	defer mongoClient.Disconnect(context.Background())
	redisClient := redis.NewClient(&redis.Options{
		Addr:     env("SESSION_REDIS_ADDR", "localhost:6379"),
		Password: env("SESSION_REDIS_PASSWORD", ""),
	})
	defer redisClient.Close()

	if err := probe(ctx, mongoClient, redisClient); err != nil {
		logger.Error("session.local.dependency_probe_failed", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "mode": "local-runtime-shell"})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := probe(r.Context(), mongoClient, redisClient); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	server := &http.Server{Addr: env("SESSION_HTTP_ADDR", ":8086"), Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	errCh := make(chan error, 1)
	go func() {
		logger.Warn("session.local.runtime_shell_started", "addr", server.Addr, "business_transport", "not_wired")
		errCh <- server.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("session.local.shutdown_failed", "error", err)
			os.Exit(1)
		}
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("session.local.http_failed", "error", err)
			os.Exit(1)
		}
	}
}

func probe(ctx context.Context, mongoClient *mongo.Client, redisClient *redis.Client) error {
	probeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := mongoClient.Ping(probeCtx, nil); err != nil {
		return err
	}
	return redisClient.Ping(probeCtx).Err()
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
