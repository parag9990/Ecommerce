package domain

import "time"

type Role string

const (
	RoleBuyer               Role = "buyer"
	RoleSeller              Role = "seller"
	RoleSellerManager       Role = "seller_manager"
	RoleSellerCatalogEditor Role = "seller_catalog_editor"
	RoleSellerOrderManager  Role = "seller_order_manager"
	RoleAdmin               Role = "admin"
	RoleOperationsAdmin     Role = "operations_admin"
	RoleFinanceAdmin        Role = "finance_admin"
	RoleCatalogAdmin        Role = "catalog_admin"
	RoleReadonlyAdmin       Role = "readonly_admin"
	RoleSuperadmin          Role = "superadmin"

	// RoleSuperAdmin is kept as an alias for older call sites.
	RoleSuperAdmin = RoleSuperadmin
)

const (
	ScopeTypeSeller = "seller"
)

type RoleAssignment struct {
	AccountID  string
	Role       Role
	ScopeType  string
	ScopeID    string
	AssignedBy string
	Reason     string
	AssignedAt time.Time
	RevokedAt  *time.Time
}

func (r Role) String() string {
	return string(r)
}

func IsKnownRole(role Role) bool {
	switch role {
	case RoleBuyer,
		RoleSeller,
		RoleSellerManager,
		RoleSellerCatalogEditor,
		RoleSellerOrderManager,
		RoleAdmin,
		RoleOperationsAdmin,
		RoleFinanceAdmin,
		RoleCatalogAdmin,
		RoleReadonlyAdmin,
		RoleSuperadmin:
		return true
	default:
		return false
	}
}

func IsSellerScopedRole(role Role) bool {
	switch role {
	case RoleSeller, RoleSellerManager, RoleSellerCatalogEditor, RoleSellerOrderManager:
		return true
	default:
		return false
	}
}

func IsAdminRole(role Role) bool {
	switch role {
	case RoleAdmin, RoleOperationsAdmin, RoleFinanceAdmin, RoleCatalogAdmin, RoleReadonlyAdmin, RoleSuperadmin:
		return true
	default:
		return false
	}
}

func IsPrivilegedRole(role Role) bool {
	return IsAdminRole(role) || IsSellerScopedRole(role)
}

func RoleStrings(roles []Role) []string {
	out := make([]string, 0, len(roles))
	for _, role := range roles {
		out = append(out, role.String())
	}
	return out
}
