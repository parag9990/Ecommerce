package httptransport

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/session-service/internal/usecase"
)

type createReportScheduleRequest struct {
	Name       string                     `json:"name"`
	ReportType domain.AnalyticsReportType `json:"report_type"`
	Format     domain.ReportFormat        `json:"format"`
	Frequency  domain.ReportFrequency     `json:"frequency"`
	Timezone   string                     `json:"timezone"`
	TimeOfDay  string                     `json:"time_of_day"`
	DayOfWeek  domain.ReportDayOfWeek     `json:"day_of_week"`
	DayOfMonth int                        `json:"day_of_month"`
	Recipients []string                   `json:"recipients"`
	Filters    reportFiltersRequest       `json:"filters"`
}

type reportFiltersRequest struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Timezone   string `json:"timezone"`
	DeviceType string `json:"device"`
	Channel    string `json:"channel"`
	Source     string `json:"source"`
	Country    string `json:"country"`
	UserType   string `json:"user_type"`
}

type updateReportScheduleRequest struct {
	Status string `json:"status"`
}

type reportSchedulesResponse struct {
	Items []domain.ReportSchedule `json:"items"`
}

func (h *Handler) handleExportReport(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) || !requireAdminAccess(w, r) {
		return
	}
	if h.reports == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "REPORTS_UNAVAILABLE", "Reports service is not configured")
		return
	}
	query := r.URL.Query()
	from, err := parseRequiredTimeQuery(query.Get("from"), "from")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	to, err := parseRequiredTimeQuery(query.Get("to"), "to")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	if len(strings.TrimSpace(query.Get("to"))) == len("2006-01-02") {
		to = to.AddDate(0, 0, 1)
	}
	out, err := h.reports.Export(r.Context(), usecase.ExportReportInput{
		Actor: actorFromRequest(r), ReportType: query.Get("report_type"), Format: query.Get("format"), From: from, To: to,
		Timezone: query.Get("timezone"), RequestID: requestID(r),
		Filters: domain.ReportFilters{From: query.Get("from"), To: query.Get("to"), Timezone: query.Get("timezone"), DeviceType: query.Get("device"), Channel: query.Get("channel"), Source: query.Get("source"), Country: query.Get("country"), UserType: query.Get("user_type")},
	})
	if err != nil {
		h.writeReportsError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", out.ContentType)
	w.Header().Set("Content-Disposition", `attachment; filename="`+out.Filename+`"`)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(out.Data)
}

func (h *Handler) handleReportSchedules(w http.ResponseWriter, r *http.Request) {
	if !requireAdminAccess(w, r) {
		return
	}
	if h.reports == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "REPORTS_UNAVAILABLE", "Reports service is not configured")
		return
	}
	actor := actorFromRequest(r)
	switch r.Method {
	case http.MethodGet:
		items, err := h.reports.ListSchedules(r.Context(), actor)
		if err != nil {
			h.writeReportsError(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, reportSchedulesResponse{Items: items})
	case http.MethodPost:
		if !requireJSONRequest(w, r) {
			return
		}
		var req createReportScheduleRequest
		if err := h.decodeJSON(w, r, &req); err != nil {
			writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
		out, err := h.reports.CreateSchedule(r.Context(), actor, domain.CreateReportSchedule{Name: req.Name, ReportType: req.ReportType, Format: req.Format, Frequency: req.Frequency, Timezone: req.Timezone, TimeOfDay: req.TimeOfDay, DayOfWeek: req.DayOfWeek, DayOfMonth: req.DayOfMonth, Recipients: req.Recipients, Filters: req.Filters.domain()}, requestID(r))
		if err != nil {
			h.writeReportsError(w, r, err)
			return
		}
		writeJSON(w, http.StatusCreated, out)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPost)
	}
}

func (h *Handler) handleReportSchedule(w http.ResponseWriter, r *http.Request) {
	if !requireAdminAccess(w, r) {
		return
	}
	if h.reports == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "REPORTS_UNAVAILABLE", "Reports service is not configured")
		return
	}
	id, ok := reportScheduleIDFromPath(r.URL.Path)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "Report schedule not found")
		return
	}
	actor := actorFromRequest(r)
	switch r.Method {
	case http.MethodPatch:
		if !requireJSONRequest(w, r) {
			return
		}
		var req updateReportScheduleRequest
		if err := h.decodeJSON(w, r, &req); err != nil {
			writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
		out, err := h.reports.UpdateScheduleStatus(r.Context(), usecase.UpdateReportScheduleStatusInput{Actor: actor, ID: id, Status: req.Status, RequestID: requestID(r)})
		if err != nil {
			h.writeReportsError(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, out)
	case http.MethodDelete:
		if err := h.reports.DeleteSchedule(r.Context(), actor, id, requestID(r)); err != nil {
			h.writeReportsError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeMethodNotAllowed(w, http.MethodPatch+", "+http.MethodDelete)
	}
}

func (r reportFiltersRequest) domain() domain.ReportFilters {
	return domain.ReportFilters{From: r.From, To: r.To, Timezone: r.Timezone, DeviceType: r.DeviceType, Channel: r.Channel, Source: r.Source, Country: r.Country, UserType: r.UserType}
}

func reportScheduleIDFromPath(path string) (string, bool) {
	const prefix = "/api/v1/analytics/reports/schedules/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	id := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	return id, id != "" && !strings.Contains(id, "/")
}

func (h *Handler) writeReportsError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrUnauthenticated):
		writeAPIError(w, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "Authentication required")
	case errors.Is(err, domain.ErrForbidden):
		writeAPIError(w, http.StatusForbidden, "PERMISSION_DENIED", "Permission denied")
	case errors.Is(err, domain.ErrInvalidInput), errors.Is(err, usecase.ErrInvalidSessionInput):
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, domain.ErrNotFound), errors.Is(err, usecase.ErrReportNotFound):
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "Report schedule not found")
	case errors.Is(err, usecase.ErrReportStorageUnavailable):
		writeAPIError(w, http.StatusServiceUnavailable, "REPORT_STORAGE_UNAVAILABLE", "Report storage is temporarily unavailable")
	default:
		h.logger.ErrorContext(r.Context(), "session.http.reports_failed", slog.String("request_id", requestID(r)), slog.String("error", err.Error()))
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}
