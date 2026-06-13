package domain

import "testing"

func TestRolePermissionMatrixMatchesSellerTaskScope(t *testing.T) {
	tests := []struct {
		role       Role
		permission Permission
		allowed    bool
	}{
		{RoleSeller, PermissionSettingsUpdate, true},
		{RoleSellerManager, PermissionStaffUpdateRole, false},
		{RoleSellerCatalogEditor, PermissionCouponsCreate, true},
		{RoleSellerCatalogEditor, PermissionOrdersRead, false},
		{RoleSellerOrderManager, PermissionOrdersUpdateFulfill, true},
		{RoleSellerOrderManager, PermissionProductsUpdate, false},
	}

	for _, tt := range tests {
		if got := RoleHasPermission(tt.role, tt.permission); got != tt.allowed {
			t.Fatalf("RoleHasPermission(%q, %q) = %v, want %v", tt.role, tt.permission, got, tt.allowed)
		}
	}
}

func TestEveryRolePermissionIsKnown(t *testing.T) {
	for role, permissions := range PermissionsByRole() {
		if !IsKnownRole(role) {
			t.Fatalf("unknown role in matrix: %q", role)
		}
		for _, permission := range permissions {
			if !IsKnownPermission(permission) {
				t.Fatalf("unknown permission %q for role %q", permission, role)
			}
		}
	}
}
