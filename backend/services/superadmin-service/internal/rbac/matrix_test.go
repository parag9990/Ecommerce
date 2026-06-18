package rbac_test

import (
	"testing"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/rbac"
)

func TestSuperadminHasEveryPermission(t *testing.T) {
	for _, permission := range domain.KnownPermissions() {
		if !rbac.StaticRoleAllows(domain.RoleSuperadmin, permission) {
			t.Fatalf("superadmin missing permission %s", permission)
		}
	}
}

func TestRolePermissionMatrix(t *testing.T) {
	cases := []struct {
		name       string
		role       domain.AdminRole
		permission domain.Permission
		allowed    bool
	}{
		{name: "operations can read users", role: domain.RoleOperations, permission: domain.PermissionUsersRead, allowed: true},
		{name: "operations cannot review refunds", role: domain.RoleOperations, permission: domain.PermissionRefundReview, allowed: false},
		{name: "finance can review refunds", role: domain.RoleFinance, permission: domain.PermissionRefundReview, allowed: true},
		{name: "finance cannot update seller status", role: domain.RoleFinance, permission: domain.PermissionSellersStatusUpdate, allowed: false},
		{name: "catalog can write catalog moderation", role: domain.RoleCatalog, permission: domain.PermissionCatalogWrite, allowed: true},
		{name: "catalog cannot write platform settings", role: domain.RoleCatalog, permission: domain.PermissionSettingsWrite, allowed: false},
		{name: "readonly can read audit logs", role: domain.RoleReadonly, permission: domain.PermissionAuditLogsRead, allowed: true},
		{name: "readonly cannot block users", role: domain.RoleReadonly, permission: domain.PermissionUsersStatusUpdate, allowed: false},
		{name: "broad admin role has no exact permissions", role: domain.RoleAdmin, permission: domain.PermissionUsersRead, allowed: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := rbac.StaticRoleAllows(tc.role, tc.permission)
			if got != tc.allowed {
				t.Fatalf("StaticRoleAllows(%s, %s) = %v, want %v", tc.role, tc.permission, got, tc.allowed)
			}
		})
	}
}

func TestPermissionDefinitionsAreComplete(t *testing.T) {
	seen := make(map[domain.Permission]struct{})
	for _, definition := range domain.PermissionDefinitions() {
		if definition.Key == "" {
			t.Fatal("permission key cannot be empty")
		}
		if definition.Description == "" {
			t.Fatalf("permission %s description cannot be empty", definition.Key)
		}
		if definition.Risk == "" {
			t.Fatalf("permission %s risk cannot be empty", definition.Key)
		}
		if _, ok := seen[definition.Key]; ok {
			t.Fatalf("duplicate permission definition %s", definition.Key)
		}
		seen[definition.Key] = struct{}{}
	}

	for role, permissions := range rbac.RolePermissionMatrix() {
		for _, permission := range permissions {
			if _, ok := seen[permission]; !ok {
				t.Fatalf("role %s references unknown permission %s", role, permission)
			}
		}
	}
}
