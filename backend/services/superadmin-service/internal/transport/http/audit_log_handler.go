package http

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"time"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
)

type AuditLogUsecase interface {
	ListAuditLogs(ctx context.Context, req domain.AuditLogListRequest) (domain.AuditLogListResponse, error)
	ExportAuditLogs(ctx context.Context, req domain.AuditLogListRequest, reason string) (domain.AuditLogListResponse, error)
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
	mux.Handle("/api/v1/admin/audit-logs/export", ActorMiddleware(http.HandlerFunc(h.exportAuditLogs)))
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

func (h *AuditLogHandler) exportAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/admin/audit-logs/export" {
		writeError(w, r, domain.NewValidationError("expected /api/v1/admin/audit-logs/export"))
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}

	var request auditExportRequest
	if err := decodeJSONBody(r, &request); err != nil {
		writeError(w, r, err)
		return
	}
	req, err := auditLogListRequestFromExport(request.Filters)
	if err != nil {
		writeError(w, r, err)
		return
	}
	response, err := h.auditLogs.ExportAuditLogs(r.Context(), req, request.Reason)
	if err != nil {
		writeError(w, r, err)
		return
	}
	payload, err := auditLogsCSV(response.Logs)
	if err != nil {
		writeError(w, r, domain.NewInternal("audit export CSV generation failed", err))
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="admin-audit-logs.csv"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}

type auditExportRequest struct {
	Filters auditExportFilters `json:"filters"`
	Reason  string             `json:"reason"`
}

type auditExportFilters struct {
	ActorID      string `json:"actor_id"`
	Action       string `json:"action"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	RequestID    string `json:"request_id"`
	From         string `json:"from"`
	To           string `json:"to"`
	Page         int    `json:"page"`
	PageSize     int    `json:"page_size"`
	Cursor       string `json:"cursor"`
}

func auditLogListRequestFromExport(filters auditExportFilters) (domain.AuditLogListRequest, error) {
	from, err := optionalRFC3339Query(filters.From, "from")
	if err != nil {
		return domain.AuditLogListRequest{}, err
	}
	to, err := optionalRFC3339Query(filters.To, "to")
	if err != nil {
		return domain.AuditLogListRequest{}, err
	}
	req := domain.AuditLogListRequest{
		ActorAdminID: filters.ActorID,
		Action:       filters.Action,
		ResourceType: filters.ResourceType,
		ResourceID:   filters.ResourceID,
		RequestID:    filters.RequestID,
		From:         from,
		To:           to,
		Pagination: domain.Pagination{
			Page:     filters.Page,
			PageSize: filters.PageSize,
			Cursor:   filters.Cursor,
		},
	}
	return req.Normalize()
}

func auditLogsCSV(logs []domain.AuditLog) ([]byte, error) {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	if err := writer.Write([]string{"audit_id", "actor_admin_id", "action", "resource_type", "resource_id", "request_id", "ip_hash", "reason", "created_at"}); err != nil {
		return nil, err
	}
	for _, log := range logs {
		if err := writer.Write([]string{
			csvCell(log.AuditID),
			csvCell(log.ActorAdminID),
			csvCell(log.Action),
			csvCell(log.ResourceType),
			csvCell(log.ResourceID),
			csvCell(log.RequestID),
			csvCell(log.IPHash),
			csvCell(log.Reason),
			csvCell(log.CreatedAt.UTC().Format(time.RFC3339)),
		}); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("write CSV: %w", err)
	}
	return buffer.Bytes(), nil
}

func csvCell(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	switch value[0] {
	case '=', '+', '-', '@':
		return "'" + value
	default:
		return value
	}
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
