package httptransport

import (
	"net/http"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
)

const (
	headerUserID      = "X-User-ID"
	headerSessionID   = "X-Session-ID"
	headerRoles       = "X-Roles"
	headerSellerID    = "X-Seller-ID"
	headerTenantID    = "X-Tenant-ID"
	headerStaffStatus = "X-Staff-Status"
	headerRequestID   = "X-Request-ID"
	headerTraceID     = "X-Trace-ID"
)

func (h *Handler) requireInternalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.internalAuthToken == "" {
			next.ServeHTTP(w, r)
			return
		}
		if r.Header.Get(h.internalAuthHeader) != h.internalAuthToken {
			writeAPIError(w, http.StatusUnauthorized, "INTERNAL_AUTH_REQUIRED", "Internal authorization required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func actorFromRequest(r *http.Request) domain.ActorContext {
	return domain.ActorContext{
		UserID:      strings.TrimSpace(r.Header.Get(headerUserID)),
		SessionID:   strings.TrimSpace(r.Header.Get(headerSessionID)),
		SellerID:    strings.TrimSpace(r.Header.Get(headerSellerID)),
		TenantID:    strings.TrimSpace(r.Header.Get(headerTenantID)),
		RequestID:   requestID(r),
		Roles:       domain.NormalizeRoles(splitCommaHeader(r.Header.Get(headerRoles))),
		StaffStatus: domain.StaffStatus(r.Header.Get(headerStaffStatus)),
	}.Normalized()
}

func requestID(r *http.Request) string {
	if requestID := strings.TrimSpace(r.Header.Get(headerRequestID)); requestID != "" {
		return requestID
	}
	if traceID := strings.TrimSpace(r.Header.Get(headerTraceID)); traceID != "" {
		return traceID
	}
	return ""
}

func splitCommaHeader(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
