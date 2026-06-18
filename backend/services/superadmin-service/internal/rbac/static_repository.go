package rbac

import (
	"context"
	"sort"

	"ecommerce/superadmin-service/internal/domain"
)

type StaticAdmin struct {
	AdminID string
	Role    domain.AdminRole
	Active  bool
}

type StaticPermissionRepository struct {
	admins map[string]StaticAdmin
}

func NewStaticPermissionRepository(admins []StaticAdmin) *StaticPermissionRepository {
	byID := make(map[string]StaticAdmin, len(admins))
	for _, admin := range admins {
		if admin.AdminID == "" {
			continue
		}
		byID[admin.AdminID] = admin
	}
	return &StaticPermissionRepository{admins: byID}
}

func (r *StaticPermissionRepository) AdminHasPermission(ctx context.Context, adminID string, permission domain.Permission) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	admin, ok := r.admins[adminID]
	if !ok || !admin.Active {
		return false, nil
	}
	return StaticRoleAllows(admin.Role, permission), nil
}

func (r *StaticPermissionRepository) ListPermissionsByRole(ctx context.Context, role domain.AdminRole) ([]domain.Permission, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return PermissionsForRole(role), nil
}

func (r *StaticPermissionRepository) ListPermissionDefinitions(ctx context.Context) ([]domain.PermissionDefinition, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	definitions := domain.PermissionDefinitions()
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].Key < definitions[j].Key })
	return definitions, nil
}
