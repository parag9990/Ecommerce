package httptransport

import (
	"context"
	"log/slog"
	"net/http"

	"ecommerce/api-gateway/internal/clients"
	"ecommerce/api-gateway/internal/domain"
	"ecommerce/api-gateway/internal/usecase"
)

type DownstreamHealthChecker interface {
	Check(ctx context.Context) clients.HealthReport
}

type Handler struct {
	catalog          usecase.RouteCatalog
	logger           *slog.Logger
	serviceName      string
	downstreamHealth DownstreamHealthChecker
}

func NewHandler(catalog usecase.RouteCatalog, logger *slog.Logger, serviceName string, downstreamHealth DownstreamHealthChecker) *Handler {
	return &Handler{
		catalog:          catalog,
		logger:           logger,
		serviceName:      serviceName,
		downstreamHealth: downstreamHealth,
	}
}

func (h *Handler) HealthLive(w http.ResponseWriter, r *http.Request) {
	writeSuccess(w, r, http.StatusOK, HealthDTO{
		Status:  "ok",
		Service: h.serviceName,
	})
}

func (h *Handler) HealthReady(w http.ResponseWriter, r *http.Request) {
	routes, err := h.catalog.ListRoutes(r.Context())
	if err != nil {
		h.logger.ErrorContext(r.Context(), "route_catalog_unavailable",
			"error", err,
			"request_id", RequestIDFromContext(r.Context()),
		)
		writeError(w, r, http.StatusServiceUnavailable, "ROUTE_CATALOG_UNAVAILABLE", "Route catalog is not ready")
		return
	}
	if h.downstreamHealth != nil {
		report := h.downstreamHealth.Check(r.Context())
		if !report.Ready() {
			h.logger.ErrorContext(r.Context(), "downstream_services_unavailable",
				"dependencies", report.Statuses(),
				"errors", report.Errors(),
				"request_id", RequestIDFromContext(r.Context()),
			)
			writeError(w, r, http.StatusServiceUnavailable, "DOWNSTREAM_SERVICES_UNAVAILABLE", "One or more downstream services are not ready", report.Errors()...)
			return
		}
		writeSuccess(w, r, http.StatusOK, HealthDTO{
			Status:       "ready",
			Service:      h.serviceName,
			Routes:       len(routes),
			Dependencies: report.Statuses(),
		})
		return
	}
	writeSuccess(w, r, http.StatusOK, HealthDTO{
		Status:  "ready",
		Service: h.serviceName,
		Routes:  len(routes),
	})
}

func (h *Handler) RouteDefined(route domain.RouteDefinition) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.logger.InfoContext(r.Context(), "route_contract_matched",
			"route_id", route.ID,
			"method", route.Method,
			"path", route.Path,
			"auth", route.AuthLevel,
			"target_service", route.Service,
			"grpc_method", route.GRPCMethod,
			"request_id", RequestIDFromContext(r.Context()),
		)
		writeError(
			w,
			r,
			http.StatusNotImplemented,
			"ROUTE_BRIDGE_NOT_CONFIGURED",
			"Route is defined in the API Gateway contract; downstream bridge implementation belongs to a later gateway task.",
		)
	}
}
