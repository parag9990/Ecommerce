package httptransport

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/authctx"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/authorization"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

func (h *Handler) AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			writeAPIError(w, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "Authentication required")
			return
		}

		claims, err := h.tokenUsecase.VerifyAccessToken(r.Context(), token)
		if err != nil {
			h.logger.InfoContext(r.Context(), "auth.http.token_denied",
				slog.String("request_id", requestID(r)),
			)
			writeAPIError(w, http.StatusUnauthorized, "INVALID_ACCESS_TOKEN", "Invalid access token")
			return
		}

		ctx := authctx.WithClaims(r.Context(), authctx.FromTokenClaims(claims))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) RequireAuthLevel(level authorization.AuthLevel) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := authorization.RequireAuthLevel(r.Context(), level); err != nil {
				h.writeAuthorizationError(w, r, err)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (h *Handler) RequireRoles(allowedRoles ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := authorization.RequireAnyRole(r.Context(), allowedRoles...); err != nil {
				h.writeAuthorizationError(w, r, err)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (h *Handler) writeAuthorizationError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, domain.ErrUnauthenticated) {
		writeAPIError(w, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "Authentication required")
		return
	}
	h.logger.InfoContext(r.Context(), "auth.http.rbac_denied",
		slog.String("request_id", requestID(r)),
	)
	writeAPIError(w, http.StatusForbidden, "PERMISSION_DENIED", "Permission denied")
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(strings.TrimSpace(header))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func requestID(r *http.Request) string {
	return strings.TrimSpace(r.Header.Get("X-Request-ID"))
}
