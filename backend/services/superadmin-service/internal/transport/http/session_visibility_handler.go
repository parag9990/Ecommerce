package http

import (
	"context"
	"net/http"
	"strings"
	"time"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
	"ecommerce/superadmin-service/internal/usecase"
)

type SessionVisibilityUsecase interface {
	AuthorizeSessionAnalytics(ctx context.Context, req usecase.AuthorizeSessionAnalyticsRequest) (*usecase.AuthorizeSessionAnalyticsResult, error)
	DashboardAccess(ctx context.Context, actor domain.AdminActor) (*usecase.SessionDashboardAccess, error)
}

type SessionVisibilityHandler struct {
	visibility SessionVisibilityUsecase
	logger     logging.Logger
}

func NewSessionVisibilityHandler(visibility SessionVisibilityUsecase, logger logging.Logger) *SessionVisibilityHandler {
	if logger == nil {
		logger = logging.NewNop()
	}
	return &SessionVisibilityHandler{visibility: visibility, logger: logger}
}

func (h *SessionVisibilityHandler) Register(mux *http.ServeMux) {
	protected := ActorMiddleware
	mux.Handle("/api/v1/admin/session-analytics/access", protected(http.HandlerFunc(h.dashboardAccess)))
	mux.Handle("/api/v1/admin/session-analytics/authorize", protected(http.HandlerFunc(h.authorizeSessionAnalytics)))
}

func (h *SessionVisibilityHandler) dashboardAccess(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/admin/session-analytics/access" {
		writeError(w, r, domain.NewValidationError("expected /api/v1/admin/session-analytics/access"))
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	actor, ok := domain.ActorFromContext(r.Context())
	if !ok {
		writeError(w, r, domain.NewAdminContextMissing("admin actor is missing from request context"))
		return
	}

	access, err := h.visibility.DashboardAccess(r.Context(), actor)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, sessionDashboardAccessResponseFromUsecase(access))
}

func (h *SessionVisibilityHandler) authorizeSessionAnalytics(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/admin/session-analytics/authorize" {
		writeError(w, r, domain.NewValidationError("expected /api/v1/admin/session-analytics/authorize"))
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	actor, ok := domain.ActorFromContext(r.Context())
	if !ok {
		writeError(w, r, domain.NewAdminContextMissing("admin actor is missing from request context"))
		return
	}

	var payload authorizeSessionAnalyticsRequest
	if err := decodeJSONBody(r, &payload); err != nil {
		writeError(w, r, err)
		return
	}

	req, err := payload.toUsecaseRequest(actor)
	if err != nil {
		writeError(w, r, err)
		return
	}

	decision, err := h.visibility.AuthorizeSessionAnalytics(r.Context(), req)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, authorizeSessionAnalyticsResponseFromUsecase(decision))
}

type authorizeSessionAnalyticsRequest struct {
	Route       string `json:"route"`
	IncludeRisk bool   `json:"include_risk"`
	Export      bool   `json:"export"`
	From        string `json:"from"`
	To          string `json:"to"`
	Page        int    `json:"page"`
	PageSize    int    `json:"page_size"`
	SessionID   string `json:"session_id"`
	Path        string `json:"path"`
	DeviceType  string `json:"device_type"`
}

func (p authorizeSessionAnalyticsRequest) toUsecaseRequest(actor domain.AdminActor) (usecase.AuthorizeSessionAnalyticsRequest, error) {
	from, err := parseAnalyticsTime(p.From, "from")
	if err != nil {
		return usecase.AuthorizeSessionAnalyticsRequest{}, err
	}
	to, err := parseAnalyticsTime(p.To, "to")
	if err != nil {
		return usecase.AuthorizeSessionAnalyticsRequest{}, err
	}

	return usecase.AuthorizeSessionAnalyticsRequest{
		Actor:       actor,
		Route:       domain.SessionAnalyticsRoute(strings.TrimSpace(p.Route)),
		IncludeRisk: p.IncludeRisk,
		Export:      p.Export,
		Filters: domain.SessionAnalyticsFilters{
			From:       from,
			To:         to,
			Page:       p.Page,
			PageSize:   p.PageSize,
			SessionID:  strings.TrimSpace(p.SessionID),
			Path:       strings.TrimSpace(p.Path),
			DeviceType: domain.SessionAnalyticsDeviceType(strings.TrimSpace(strings.ToLower(p.DeviceType))),
		},
	}, nil
}

func parseAnalyticsTime(raw string, field string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, nil
	}
	if value, err := time.Parse(time.RFC3339, raw); err == nil {
		return value, nil
	}
	if value, err := time.Parse("2006-01-02", raw); err == nil {
		return value, nil
	}
	return time.Time{}, domain.NewValidationError(field + " must be RFC3339 or YYYY-MM-DD")
}

type authorizeSessionAnalyticsResponse struct {
	Allowed           bool              `json:"allowed"`
	Route             string            `json:"route"`
	Permission        string            `json:"permission"`
	RiskPermission    string            `json:"risk_permission"`
	RiskAllowed       bool              `json:"risk_allowed"`
	MaskPII           bool              `json:"mask_pii"`
	DownstreamHeaders map[string]string `json:"downstream_headers"`
}

func authorizeSessionAnalyticsResponseFromUsecase(result *usecase.AuthorizeSessionAnalyticsResult) authorizeSessionAnalyticsResponse {
	if result == nil {
		return authorizeSessionAnalyticsResponse{}
	}
	headers := make(map[string]string, len(result.DownstreamHeaders))
	for key, value := range result.DownstreamHeaders {
		headers[key] = value
	}
	return authorizeSessionAnalyticsResponse{
		Allowed:           result.Allowed,
		Route:             string(result.Route),
		Permission:        string(result.Permission),
		RiskPermission:    string(result.RiskPermission),
		RiskAllowed:       result.RiskAllowed,
		MaskPII:           result.MaskPII,
		DownstreamHeaders: headers,
	}
}

type sessionDashboardAccessResponse struct {
	CanViewDashboard bool                             `json:"can_view_dashboard"`
	RiskAllowed      bool                             `json:"risk_allowed"`
	MaskPII          bool                             `json:"mask_pii"`
	Modules          []sessionDashboardModuleResponse `json:"modules"`
}

type sessionDashboardModuleResponse struct {
	Key                string `json:"key"`
	Route              string `json:"route"`
	RequiredPermission string `json:"required_permission"`
	RequiresRisk       bool   `json:"requires_risk"`
	Visible            bool   `json:"visible"`
}

func sessionDashboardAccessResponseFromUsecase(access *usecase.SessionDashboardAccess) sessionDashboardAccessResponse {
	if access == nil {
		return sessionDashboardAccessResponse{}
	}
	modules := make([]sessionDashboardModuleResponse, 0, len(access.Modules))
	for _, module := range access.Modules {
		modules = append(modules, sessionDashboardModuleResponse{
			Key:                module.Key,
			Route:              string(module.Route),
			RequiredPermission: string(module.RequiredPermission),
			RequiresRisk:       module.RequiresRisk,
			Visible:            module.Visible,
		})
	}
	return sessionDashboardAccessResponse{
		CanViewDashboard: access.CanViewDashboard,
		RiskAllowed:      access.RiskAllowed,
		MaskPII:          access.MaskPII,
		Modules:          modules,
	}
}
