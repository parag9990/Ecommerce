package middleware

import (
	"log/slog"
	"net/http"

	"github.com/parag/ecommerce/backend/services/api-gateway/internal/authctx"
)

type RBACMiddleware struct {
	logger *slog.Logger
}

func NewRBACMiddleware(logger *slog.Logger) RBACMiddleware {
	if logger == nil {
		logger = slog.Default()
	}
	return RBACMiddleware{logger: logger}
}

func (m RBACMiddleware) RequireAnyRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := authctx.ClaimsFromContext(r.Context())
			if !ok || claims.UserID == "" {
				writeAPIError(w, r, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "Authentication required")
				return
			}

			for _, role := range claims.Roles {
				if _, ok := allowed[role]; ok {
					next.ServeHTTP(w, r)
					return
				}
			}

			m.logger.WarnContext(r.Context(), "rbac_denied",
				slog.String("user_id", claims.UserID),
				slog.String("request_id", RequestIDFromRequest(r)),
			)
			writeAPIError(w, r, http.StatusForbidden, "PERMISSION_DENIED", "Permission denied")
		})
	}
}
