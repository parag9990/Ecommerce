// Command local-server provides dependency probes until the Order Service's
// existing gRPC application is wired to concrete cart, product, and payment adapters.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := sql.Open("mysql", env("ORDER_MYSQL_DSN", "root:local_password@tcp(localhost:3306)/order_db?parseTime=true"))
	if err != nil {
		logger.Error("order.local.mysql_open_failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	if err := db.PingContext(probeCtx); err != nil {
		cancel()
		logger.Error("order.local.mysql_ping_failed", "error", err)
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
		if err := db.PingContext(probeCtx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	server := &http.Server{Addr: env("ORDER_LOCAL_HTTP_ADDR", ":8090"), Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	errCh := make(chan error, 1)
	go func() {
		logger.Warn("order.local.runtime_shell_started", "addr", server.Addr, "business_transport", "not_wired")
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("order.local.shutdown_failed", "error", err)
			os.Exit(1)
		}
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("order.local.http_failed", "error", err)
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
