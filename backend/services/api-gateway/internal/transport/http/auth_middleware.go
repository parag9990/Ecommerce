package httptransport

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	gatewayauth "ecommerce/api-gateway/internal/auth"
	"ecommerce/api-gateway/internal/domain"
	gatewayerrors "ecommerce/api-gateway/internal/errors"
	"ecommerce/api-gateway/internal/observability"
)

type TokenVerifier interface {
	Verify(ctx context.Context, raw string) (gatewayauth.AccessClaims, error)
}

func AuthRequired(verifier TokenVerifier, logger *slog.Logger, route domain.RouteDefinition, metrics *observability.Metrics, userHashSalt string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := gatewayauth.BearerToken(r.Header.Get("Authorization"))
			if raw == "" {
				observeAuthFailure(metrics, route, "missing_bearer_token")
				writeError(w, r, http.StatusUnauthorized, gatewayerrors.CodeUnauthorized, "Authentication required")
				return
			}

			claims, err := verifier.Verify(r.Context(), raw)
			if err != nil {
				observeAuthFailure(metrics, route, "jwt_verification_failed")
				logAuthFailure(logger, r, route, "jwt_verification_failed", err)
				writeError(w, r, http.StatusUnauthorized, gatewayerrors.CodeUnauthorized, "Authentication required")
				return
			}

			observability.RecordUserIDHash(w, observability.HashUserID(claims.UserID(), userHashSalt))
			next.ServeHTTP(w, r.WithContext(gatewayauth.WithClaims(r.Context(), claims)))
		})
	}
}

func OptionalAuth(verifier TokenVerifier, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := gatewayauth.BearerToken(r.Header.Get("Authorization"))
			if raw == "" {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := verifier.Verify(r.Context(), raw)
			if err != nil {
				logAuthFailure(logger, r, domain.RouteDefinition{}, "optional_jwt_verification_failed", err)
				next.ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r.WithContext(gatewayauth.WithClaims(r.Context(), claims)))
		})
	}
}

func RequireRoles(logger *slog.Logger, route domain.RouteDefinition, metrics *observability.Metrics, userHashSalt string, allowed ...string) func(http.Handler) http.Handler {
	allowed = normalizeAllowedRoles(allowed)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := gatewayauth.ClaimsFromContext(r.Context())
			if !ok {
				observeAuthFailure(metrics, route, "auth_context_missing")
				logAuthFailure(logger, r, route, "auth_context_missing", gatewayauth.ErrMissingClaim)
				writeError(w, r, http.StatusUnauthorized, gatewayerrors.CodeUnauthorized, "Authentication required")
				return
			}
			if claims.HasAnyRole(allowed) {
				next.ServeHTTP(w, r)
				return
			}

			observeAuthFailure(metrics, route, "rbac_denied")
			logAuthFailure(logger, r, route, "rbac_denied", nil,
				"user_id_hash", observability.HashUserID(claims.UserID(), userHashSalt),
				"roles", strings.Join(claims.Roles, ","),
				"allowed_roles", strings.Join(allowed, ","),
			)
			writeError(w, r, http.StatusForbidden, gatewayerrors.CodeForbidden, "You do not have permission to perform this action")
		})
	}
}

func RequireWebhookSignature(signatureHeader string, logger *slog.Logger, route domain.RouteDefinition, metrics *observability.Metrics) func(http.Handler) http.Handler {
	signatureHeader = strings.TrimSpace(signatureHeader)
	if signatureHeader == "" {
		signatureHeader = "X-Provider-Signature"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			provider := strings.TrimSpace(r.PathValue("provider"))
			signature := strings.TrimSpace(r.Header.Get(signatureHeader))
			if provider == "" || signature == "" {
				observeAuthFailure(metrics, route, "webhook_signature_missing")
				logAuthFailure(logger, r, route, "webhook_signature_missing", nil, "provider", provider, "signature_header", signatureHeader)
				writeError(w, r, http.StatusUnauthorized, gatewayerrors.CodeUnauthorized, "Authentication required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func normalizeAllowedRoles(roles []string) []string {
	seen := make(map[string]struct{}, len(roles))
	out := make([]string, 0, len(roles))
	for _, role := range roles {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}
		if _, exists := seen[role]; exists {
			continue
		}
		seen[role] = struct{}{}
		out = append(out, role)
	}
	return out
}

func logAuthFailure(logger *slog.Logger, r *http.Request, route domain.RouteDefinition, message string, err error, attrs ...any) {
	if logger == nil {
		return
	}
	args := []any{
		"method", r.Method,
		"route", routeTemplate(route),
		"request_id", RequestIDFromContext(r.Context()),
	}
	if err != nil && !errors.Is(err, gatewayauth.ErrInvalidToken) {
		args = append(args, "error", err)
	}
	args = append(args, attrs...)
	logger.WarnContext(r.Context(), message, args...)
}

func observeAuthFailure(metrics *observability.Metrics, route domain.RouteDefinition, reason string) {
	if metrics == nil {
		return
	}
	metrics.ObserveAuthFailure(routeTemplate(route), reason)
}

func routeTemplate(route domain.RouteDefinition) string {
	if strings.TrimSpace(route.Path) == "" {
		return observability.RouteUnknown
	}
	return route.Path
}
