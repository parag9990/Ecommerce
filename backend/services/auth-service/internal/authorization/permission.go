package authorization

import (
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/authctx"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

type Permission string

const (
	PermissionProfileReadSelf     Permission = "profile:read:self"
	PermissionCartWriteSelf       Permission = "cart:write:self"
	PermissionProductWriteOwn     Permission = "product:write:own_seller"
	PermissionOrderManageOwn      Permission = "order:manage:own_seller"
	PermissionPaymentRefundReview Permission = "payment:refund:review"
	PermissionPlatformWrite       Permission = "platform:settings:write"
)

var permissionsByRole = map[domain.Role][]Permission{
	domain.RoleBuyer: {
		PermissionProfileReadSelf,
		PermissionCartWriteSelf,
	},
	domain.RoleSeller: {
		PermissionProfileReadSelf,
		PermissionCartWriteSelf,
		PermissionProductWriteOwn,
		PermissionOrderManageOwn,
	},
	domain.RoleSellerManager: {
		PermissionProfileReadSelf,
		PermissionProductWriteOwn,
		PermissionOrderManageOwn,
	},
	domain.RoleSellerCatalogEditor: {
		PermissionProfileReadSelf,
		PermissionProductWriteOwn,
	},
	domain.RoleSellerOrderManager: {
		PermissionProfileReadSelf,
		PermissionOrderManageOwn,
	},
	domain.RoleAdmin: {
		PermissionProfileReadSelf,
	},
	domain.RoleOperationsAdmin: {
		PermissionProfileReadSelf,
	},
	domain.RoleFinanceAdmin: {
		PermissionProfileReadSelf,
		PermissionPaymentRefundReview,
	},
	domain.RoleCatalogAdmin: {
		PermissionProfileReadSelf,
		PermissionProductWriteOwn,
	},
	domain.RoleReadonlyAdmin: {
		PermissionProfileReadSelf,
	},
	domain.RoleSuperadmin: {
		PermissionProfileReadSelf,
		PermissionProductWriteOwn,
		PermissionOrderManageOwn,
		PermissionPaymentRefundReview,
		PermissionPlatformWrite,
	},
}

func RoleHasPermission(role domain.Role, permission Permission) bool {
	for _, current := range permissionsByRole[role] {
		if current == permission {
			return true
		}
	}
	return false
}

func ClaimsHavePermission(claims authctx.Claims, permission Permission) bool {
	for _, role := range claims.Roles {
		if RoleHasPermission(domain.Role(role), permission) {
			return true
		}
	}
	return false
}
