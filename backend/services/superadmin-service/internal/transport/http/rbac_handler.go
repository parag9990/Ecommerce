package http

import (
	"context"
	"net/http"
	"strings"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
)

type AuthorizationUsecase interface {
	PermissionCatalog(ctx context.Context, actor domain.AdminActor) ([]domain.PermissionDefinition, error)
	ListRolePermissions(ctx context.Context, actor domain.AdminActor, role domain.AdminRole) ([]domain.Permission, error)
	RequirePermission(ctx context.Context, actor domain.AdminActor, permission domain.Permission) error
}

type ReadinessChecker interface {
	PingContext(ctx context.Context) error
}

type RBACHandler struct {
	authz  AuthorizationUsecase
	ready  ReadinessChecker
	logger logging.Logger
}

func NewRBACHandler(authz AuthorizationUsecase, ready ReadinessChecker, logger logging.Logger) *RBACHandler {
	if logger == nil {
		logger = logging.NewNop()
	}
	return &RBACHandler{authz: authz, ready: ready, logger: logger}
}

func (h *RBACHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.health)
	mux.HandleFunc("/readyz", h.readyz)

	protected := ActorMiddleware
	mux.Handle("/api/v1/admin/rbac/permissions", protected(http.HandlerFunc(h.listPermissions)))
	mux.Handle("/api/v1/admin/rbac/roles/", protected(http.HandlerFunc(h.listRolePermissions)))
}

func (h *RBACHandler) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *RBACHandler) readyz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Code: "METHOD_NOT_ALLOWED", Message: "method not allowed"})
		return
	}
	if h.ready != nil {
		if err := h.ready.PingContext(r.Context()); err != nil {
			h.logger.Error(r.Context(), "readiness check failed", "error", err)
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (h *RBACHandler) listPermissions(w http.ResponseWriter, r *http.Request) {
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

	definitions, err := h.authz.PermissionCatalog(r.Context(), actor)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, permissionCatalogResponse{Permissions: permissionDTOs(definitions)})
}

func (h *RBACHandler) listRolePermissions(w http.ResponseWriter, r *http.Request) {
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

	role, ok := roleFromPath(r.URL.Path)
	if !ok {
		writeError(w, r, domain.NewValidationError("expected /api/v1/admin/rbac/roles/{role}/permissions"))
		return
	}

	permissions, err := h.authz.ListRolePermissions(r.Context(), actor, role)
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, rolePermissionsResponse{
		Role:        string(role),
		Permissions: permissionsToStrings(permissions),
	})
}

func roleFromPath(path string) (domain.AdminRole, bool) {
	const prefix = "/api/v1/admin/rbac/roles/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}

	remaining := strings.TrimPrefix(path, prefix)
	roleValue, suffix, ok := strings.Cut(remaining, "/")
	if !ok || suffix != "permissions" || roleValue == "" {
		return "", false
	}
	return domain.AdminRole(roleValue), true
}
