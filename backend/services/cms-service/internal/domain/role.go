package domain

import "strings"

type Role string

const (
	RoleSeller              Role = "seller"
	RoleSellerManager       Role = "seller_manager"
	RoleSellerCatalogEditor Role = "seller_catalog_editor"
	RoleSellerOrderManager  Role = "seller_order_manager"
	RoleCatalogAdmin        Role = "catalog_admin"
	RoleSuperadmin          Role = "superadmin"
)

var orderedRoles = []Role{
	RoleSeller,
	RoleSellerManager,
	RoleSellerCatalogEditor,
	RoleSellerOrderManager,
	RoleCatalogAdmin,
	RoleSuperadmin,
}

func (r Role) String() string {
	return string(r)
}

func AllRoles() []Role {
	return append([]Role(nil), orderedRoles...)
}

func IsKnownRole(role Role) bool {
	switch role {
	case RoleSeller, RoleSellerManager, RoleSellerCatalogEditor, RoleSellerOrderManager, RoleCatalogAdmin, RoleSuperadmin:
		return true
	default:
		return false
	}
}

func NormalizeRoles(values []string) []Role {
	seen := make(map[Role]struct{}, len(values))
	roles := make([]Role, 0, len(values))
	for _, value := range values {
		role := Role(strings.TrimSpace(value))
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

func RoleStrings(roles []Role) []string {
	out := make([]string, 0, len(roles))
	for _, role := range roles {
		out = append(out, role.String())
	}
	return out
}
