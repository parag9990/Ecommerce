package http

import (
	"context"
	"net/http"
	"strings"
	"time"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
)

type AuditLogUsecase interface {
	ListAuditLogs(ctx context.Context, req domain.AuditLogListRequest) (domain.AuditLogListResponse, error)
}

type AuditLogHandler struct {
	auditLogs AuditLogUsecase
	logger    logging.Logger
}

func NewAuditLogHandler(auditLogs AuditLogUsecase, logger logging.Logger) *AuditLogHandler {
	if logger == nil {
		logger = logging.NewNop()
	}
	return &AuditLogHandler{auditLogs: auditLogs, logger: logger}
}

func (h *AuditLogHandler) Register(mux *http.ServeMux) {
	mux.Handle("/api/v1/admin/audit-logs", ActorMiddleware(http.HandlerFunc(h.listAuditLogs)))
}

func (h *AuditLogHandler) listAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/admin/audit-logs" {
		writeError(w, r, domain.NewValidationError("expected /api/v1/admin/audit-logs"))
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	req, err := auditLogListRequestFromQuery(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	response, err := h.auditLogs.ListAuditLogs(r.Context(), req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func auditLogListRequestFromQuery(r *http.Request) (domain.AuditLogListRequest, error) {
	query := r.URL.Query()
	page, err := positiveIntQuery(query.Get("page"), "page")
	if err != nil {
		return domain.AuditLogListRequest{}, err
	}
	pageSize, err := positiveIntQuery(query.Get("page_size"), "page_size")
	if err != nil {
		return domain.AuditLogListRequest{}, err
	}
	from, err := optionalRFC3339Query(query.Get("from"), "from")
	if err != nil {
		return domain.AuditLogListRequest{}, err
	}
	to, err := optionalRFC3339Query(query.Get("to"), "to")
	if err != nil {
		return domain.AuditLogListRequest{}, err
	}

	req := domain.AuditLogListRequest{
		ActorAdminID: query.Get("actor_id"),
		Action:       query.Get("action"),
		ResourceType: query.Get("resource_type"),
		ResourceID:   query.Get("resource_id"),
		RequestID:    query.Get("request_id"),
		From:         from,
		To:           to,
		Pagination: domain.Pagination{
			Page:     page,
			PageSize: pageSize,
			Cursor:   query.Get("cursor"),
		},
	}
	return req.Normalize()
}

func optionalRFC3339Query(raw string, field string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, domain.NewValidationError(field + " must be an RFC3339 timestamp")
	}
	return parsed.UTC(), nil
}
