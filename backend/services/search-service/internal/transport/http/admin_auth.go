package httptransport

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/search-service/internal/domain"
)

const (
	adminPermissionSynonymsRead  = "search:synonyms:read"
	adminPermissionSynonymsWrite = "search:synonyms:write"
	adminPermissionReindexWrite  = "search:reindex:write"
)

var searchAdminRoles = map[string]struct{}{
	"catalog_admin": {},
	"superadmin":    {},
}

var supportedAdminPermissions = map[string]struct{}{
	adminPermissionSynonymsRead:  {},
	adminPermissionSynonymsWrite: {},
	adminPermissionReindexWrite:  {},
}

type AdminActor struct {
	ID    string
	Roles []string
}

type AdminAuthorizer interface {
	Authorize(ctx context.Context, r *http.Request, permission string) (AdminActor, error)
}

type HeaderAdminAuthorizer struct {
	enabled bool
}

func NewHeaderAdminAuthorizer(enabled bool) HeaderAdminAuthorizer {
	return HeaderAdminAuthorizer{enabled: enabled}
}

func (a HeaderAdminAuthorizer) Authorize(_ context.Context, r *http.Request, permission string) (AdminActor, error) {
	if !a.enabled {
		return AdminActor{ID: "local-dev", Roles: []string{"superadmin"}}, nil
	}
	if _, ok := supportedAdminPermissions[permission]; !ok {
		return AdminActor{}, fmt.Errorf("%w: unsupported admin permission", domain.ErrPermissionDenied)
	}

	actor := AdminActor{
		ID:    firstHeader(r.Header, "X-User-ID", "X-Actor-ID", "X-Admin-ID"),
		Roles: normalizeRoleHeader(firstHeader(r.Header, "X-Roles", "X-User-Roles", "X-Actor-Roles")),
	}
	if actor.ID == "" {
		return AdminActor{}, domain.ErrUnauthenticated
	}
	for _, role := range actor.Roles {
		if _, ok := searchAdminRoles[role]; ok {
			return actor, nil
		}
	}
	return AdminActor{}, domain.ErrPermissionDenied
}

func firstHeader(header http.Header, names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(header.Get(name)); value != "" {
			return value
		}
	}
	return ""
}

func normalizeRoleHeader(value string) []string {
	parts := strings.Split(value, ",")
	seen := make(map[string]struct{}, len(parts))
	roles := make([]string, 0, len(parts))
	for _, part := range parts {
		role := strings.ToLower(strings.TrimSpace(part))
		if role == "" {
			continue
		}
		if _, exists := seen[role]; exists {
			continue
		}
		seen[role] = struct{}{}
		roles = append(roles, role)
	}
	return roles
}
