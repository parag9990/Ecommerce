package httptransport

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type HealthChecker interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	serviceName string
	checker     HealthChecker
	logger      *slog.Logger
	timeout     time.Duration
}

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Error   string `json:"error,omitempty"`
}

func NewHealthHandler(serviceName string, checker HealthChecker, logger *slog.Logger) *HealthHandler {
	if logger == nil {
		logger = slog.Default()
	}
	if serviceName == "" {
		serviceName = "wishlist-service"
	}
	return &HealthHandler{
		serviceName: serviceName,
		checker:     checker,
		logger:      logger,
		timeout:     2 * time.Second,
	}
}

func (h *HealthHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.Live)
	mux.HandleFunc("/readyz", h.Ready)
}

func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, healthResponse{
			Status:  "error",
			Service: h.serviceName,
			Error:   "method not allowed",
		})
		return
	}
	writeJSON(w, http.StatusOK, healthResponse{
		Status:  "ok",
		Service: h.serviceName,
	})
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, healthResponse{
			Status:  "error",
			Service: h.serviceName,
			Error:   "method not allowed",
		})
		return
	}
	if h.checker == nil {
		writeJSON(w, http.StatusServiceUnavailable, healthResponse{
			Status:  "error",
			Service: h.serviceName,
			Error:   "readiness checker is unavailable",
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	if err := h.checker.Ping(ctx); err != nil {
		h.logger.Warn("wishlist readiness check failed", "error", err)
		writeJSON(w, http.StatusServiceUnavailable, healthResponse{
			Status:  "error",
			Service: h.serviceName,
			Error:   "mongo unavailable",
		})
		return
	}

	writeJSON(w, http.StatusOK, healthResponse{
		Status:  "ok",
		Service: h.serviceName,
	})
}

func writeJSON(w http.ResponseWriter, status int, response healthResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}
