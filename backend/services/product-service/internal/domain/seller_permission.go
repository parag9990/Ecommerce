package domain

import "strings"

const (
	RoleSeller              = "seller"
	RoleSellerManager       = "seller_manager"
	RoleSellerCatalogEditor = "seller_catalog_editor"
	RoleSuperadmin          = "superadmin"

	PermissionProductReadOwnSeller  = "product:read:own_seller"
	PermissionProductWriteOwnSeller = "product:write:own_seller"
)

type ActorContext struct {
	UserID      string   `json:"user_id"`
	SellerID    string   `json:"seller_id,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

func (a ActorContext) ActorID() string {
	if strings.TrimSpace(a.UserID) != "" {
		return strings.TrimSpace(a.UserID)
	}
	return strings.TrimSpace(a.SellerID)
}

func (a ActorContext) IsSuperadmin() bool {
	return a.HasRole(RoleSuperadmin)
}

func (a ActorContext) CanWriteSellerProducts() bool {
	if a.IsSuperadmin() || a.HasPermission(PermissionProductWriteOwnSeller) {
		return true
	}
	return a.HasRole(RoleSeller) ||
		a.HasRole(RoleSellerManager) ||
		a.HasRole(RoleSellerCatalogEditor)
}

func (a ActorContext) CanReadSellerProducts() bool {
	if a.IsSuperadmin() ||
		a.HasPermission(PermissionProductReadOwnSeller) ||
		a.HasPermission(PermissionProductWriteOwnSeller) {
		return true
	}
	return a.HasRole(RoleSeller) ||
		a.HasRole(RoleSellerManager) ||
		a.HasRole(RoleSellerCatalogEditor)
}

func (a ActorContext) HasRole(role string) bool {
	return containsNormalized(a.Roles, role)
}

func (a ActorContext) HasPermission(permission string) bool {
	return containsNormalized(a.Permissions, permission)
}

func containsNormalized(values []string, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	for _, value := range values {
		if strings.ToLower(strings.TrimSpace(value)) == target {
			return true
		}
	}
	return false
}
