package http

import (
	"context"
	"net/http"
	"time"
)

type ReadinessCheck func(context.Context) error

type HealthHandler struct {
	check   ReadinessCheck
	timeout time.Duration
}

func NewHealthHandler(check ReadinessCheck, timeout time.Duration) *HealthHandler {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	return &HealthHandler{check: check, timeout: timeout}
}

func (h *HealthHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.Live)
	mux.HandleFunc("GET /readyz", h.Ready)
}

func (h *HealthHandler) Live(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (h *HealthHandler) Ready(w http.ResponseWriter, request *http.Request) {
	if h.check == nil {
		http.Error(w, "service is not ready", http.StatusServiceUnavailable)
		return
	}
	ctx, cancel := context.WithTimeout(request.Context(), h.timeout)
	defer cancel()
	if err := h.check(ctx); err != nil {
		http.Error(w, "service is not ready", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
