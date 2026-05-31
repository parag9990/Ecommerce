package http

import (
	"context"
	"net/http"
	"strings"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
	"ecommerce/superadmin-service/internal/validation"
)

const (
	HeaderAdminID     = "X-Admin-Id"
	HeaderUserID      = "X-User-Id"
	HeaderSubjectID   = "X-Subject-Id"
	HeaderAdminRoles  = "X-Admin-Roles"
	HeaderRoles       = "X-Roles"
	HeaderSessionID   = "X-Session-Id"
	HeaderRequestID   = "X-Request-Id"
	HeaderIPHash      = "X-IP-Hash"
	HeaderMFAVerified = "X-MFA-Verified"
)

type PermissionAuthorizer interface {
	RequirePermission(ctx context.Context, actor domain.AdminActor, permission domain.Permission) error
}

func ActorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor := actorFromHeaders(r)
		next.ServeHTTP(w, r.WithContext(domain.ContextWithActor(r.Context(), actor)))
	})
}

func RequirePermission(authorizer PermissionAuthorizer, permission domain.Permission, logger logging.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = logging.NewNop()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor, ok := domain.ActorFromContext(r.Context())
			if !ok {
				writeError(w, r, domain.NewAdminContextMissing("admin actor is missing from request context"))
				return
			}

			if err := authorizer.RequirePermission(r.Context(), actor, permission); err != nil {
				logger.Warn(r.Context(), "admin middleware denied request",
					"permission", permission,
					"request_id", actor.RequestID,
					"error", err,
				)
				writeError(w, r, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func actorFromHeaders(r *http.Request) domain.AdminActor {
	userID := strings.TrimSpace(r.Header.Get(HeaderUserID))
	if userID == "" {
		userID = strings.TrimSpace(r.Header.Get(HeaderSubjectID))
	}

	rolesHeader := r.Header.Get(HeaderAdminRoles)
	if strings.TrimSpace(rolesHeader) == "" {
		rolesHeader = r.Header.Get(HeaderRoles)
	}

	return domain.AdminActor{
		AdminID:     strings.TrimSpace(r.Header.Get(HeaderAdminID)),
		UserID:      userID,
		Roles:       domain.RolesFromStrings(validation.SplitCSV(rolesHeader)),
		SessionID:   strings.TrimSpace(r.Header.Get(HeaderSessionID)),
		RequestID:   requestIDFromRequest(r),
		IPHash:      strings.TrimSpace(r.Header.Get(HeaderIPHash)),
		MFAVerified: parseBoolHeader(r.Header.Get(HeaderMFAVerified)),
	}
}

func requestIDFromRequest(r *http.Request) string {
	return strings.TrimSpace(r.Header.Get(HeaderRequestID))
}

func parseBoolHeader(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "yes", "y":
		return true
	default:
		return false
	}
}
