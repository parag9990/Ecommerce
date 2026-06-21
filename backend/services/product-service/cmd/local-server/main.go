// Command local-server provides dependency probes until the Product Service's
// existing business handlers are wired to a network transport.
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

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	uri := env("PRODUCT_MONGO_URI", "mongodb://localhost:27017")
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		logger.Error("product.local.mongo_connect_failed", "error", err)
		os.Exit(1)
	}
	defer client.Disconnect(context.Background())

	probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	if err := client.Ping(probeCtx, nil); err != nil {
		cancel()
		logger.Error("product.local.mongo_ping_failed", "error", err)
		os.Exit(1)
	}
	cancel()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "mode": "local-runtime-shell"})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		probeCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := client.Ping(probeCtx, nil); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	server := &http.Server{Addr: env("PRODUCT_HTTP_ADDR", ":8082"), Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	errCh := make(chan error, 1)
	go func() {
		logger.Warn("product.local.runtime_shell_started", "addr", server.Addr, "business_transport", "not_wired")
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("product.local.shutdown_failed", "error", err)
			os.Exit(1)
		}
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("product.local.http_failed", "error", err)
			os.Exit(1)
		}
	}
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
