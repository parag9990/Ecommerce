package rbac

import (
	"sort"

	"ecommerce/superadmin-service/internal/domain"
)

var rolePermissionMatrix = map[domain.AdminRole][]domain.Permission{
	domain.RoleSuperadmin: domain.KnownPermissions(),
	domain.RoleOperations: {
		domain.PermissionUsersRead,
		domain.PermissionUsersStatusUpdate,
		domain.PermissionSellersRead,
		domain.PermissionSellersStatusUpdate,
		domain.PermissionOrdersRead,
		domain.PermissionOrderManualReview,
		domain.PermissionSessionsRead,
		domain.PermissionSessionRiskRead,
		domain.PermissionCatalogRead,
	},
	domain.RoleFinance: {
		domain.PermissionPaymentsRead,
		domain.PermissionRefundReview,
	},
	domain.RoleCatalog: {
		domain.PermissionSellersRead,
		domain.PermissionSellersStatusUpdate,
		domain.PermissionCatalogRead,
		domain.PermissionCatalogWrite,
		domain.PermissionSearchSynonymsRead,
		domain.PermissionSearchSynonymsWrite,
	},
	domain.RoleReadonly: {
		domain.PermissionUsersRead,
		domain.PermissionSellersRead,
		domain.PermissionOrdersRead,
		domain.PermissionSessionsRead,
		domain.PermissionCatalogRead,
		domain.PermissionSearchSynonymsRead,
		domain.PermissionAuditLogsRead,
	},
}

func RolePermissionMatrix() map[domain.AdminRole][]domain.Permission {
	out := make(map[domain.AdminRole][]domain.Permission, len(rolePermissionMatrix))
	for role, permissions := range rolePermissionMatrix {
		out[role] = cloneAndSortPermissions(permissions)
	}
	return out
}

func PermissionsForRole(role domain.AdminRole) []domain.Permission {
	return cloneAndSortPermissions(rolePermissionMatrix[role])
}

func StaticRoleAllows(role domain.AdminRole, permission domain.Permission) bool {
	for _, candidate := range rolePermissionMatrix[role] {
		if candidate == permission {
			return true
		}
	}
	return false
}

func ActorHasStaticPermission(actor domain.AdminActor, permission domain.Permission) bool {
	for _, role := range actor.Roles {
		if StaticRoleAllows(role, permission) {
			return true
		}
	}
	return false
}

func cloneAndSortPermissions(permissions []domain.Permission) []domain.Permission {
	out := make([]domain.Permission, len(permissions))
	copy(out, permissions)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
