package authorization

import (
	"context"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/authctx"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

type AuthLevel string

const (
	AuthLevelPublic     AuthLevel = "public"
	AuthLevelBuyer      AuthLevel = "buyer"
	AuthLevelSeller     AuthLevel = "seller"
	AuthLevelAdmin      AuthLevel = "admin"
	AuthLevelSuperadmin AuthLevel = "superadmin"
	AuthLevelWebhook    AuthLevel = "webhook"
	AuthLevelInternal   AuthLevel = "internal"
)

var rolesByAuthLevel = map[AuthLevel][]domain.Role{
	AuthLevelBuyer: {
		domain.RoleBuyer,
		domain.RoleSeller,
		domain.RoleAdmin,
		domain.RoleSuperadmin,
	},
	AuthLevelSeller: {
		domain.RoleSeller,
		domain.RoleSellerManager,
		domain.RoleSellerCatalogEditor,
		domain.RoleSellerOrderManager,
		domain.RoleSuperadmin,
	},
	AuthLevelAdmin: {
		domain.RoleAdmin,
		domain.RoleOperationsAdmin,
		domain.RoleFinanceAdmin,
		domain.RoleCatalogAdmin,
		domain.RoleSuperadmin,
	},
	AuthLevelSuperadmin: {
		domain.RoleSuperadmin,
	},
}

func RolesForAuthLevel(level AuthLevel) ([]domain.Role, bool) {
	roles, ok := rolesByAuthLevel[level]
	if !ok {
		return nil, false
	}
	return append([]domain.Role(nil), roles...), true
}

func RequireAuthLevel(ctx context.Context, level AuthLevel) (authctx.Claims, error) {
	switch level {
	case AuthLevelPublic, AuthLevelWebhook:
		return authctx.Claims{}, nil
	case AuthLevelInternal:
		return RequireAuthenticated(ctx)
	default:
		roles, ok := RolesForAuthLevel(level)
		if !ok {
			return authctx.Claims{}, domain.ErrForbidden
		}
		return RequireAnyRole(ctx, roles...)
	}
}

func RequireAuthenticated(ctx context.Context) (authctx.Claims, error) {
	claims, ok := authctx.FromContext(ctx)
	if !ok || !claims.Authenticated() {
		return authctx.Claims{}, domain.ErrUnauthenticated
	}
	return claims, nil
}

func RequireAnyRole(ctx context.Context, roles ...domain.Role) (authctx.Claims, error) {
	claims, err := RequireAuthenticated(ctx)
	if err != nil {
		return authctx.Claims{}, err
	}
	if len(roles) == 0 {
		return claims, nil
	}
	if !claims.HasAnyRoleList(roles...) {
		return authctx.Claims{}, domain.ErrForbidden
	}
	return claims, nil
}

func RequirePermission(ctx context.Context, permission Permission) (authctx.Claims, error) {
	claims, err := RequireAuthenticated(ctx)
	if err != nil {
		return authctx.Claims{}, err
	}
	if !ClaimsHavePermission(claims, permission) {
		return authctx.Claims{}, domain.ErrForbidden
	}
	return claims, nil
}

func RequireSellerScope(ctx context.Context, resourceSellerID string, roles ...domain.Role) (authctx.Claims, error) {
	if len(roles) == 0 {
		roles = []domain.Role{
			domain.RoleSeller,
			domain.RoleSellerManager,
			domain.RoleSellerCatalogEditor,
			domain.RoleSellerOrderManager,
			domain.RoleSuperadmin,
		}
	}
	claims, err := RequireAnyRole(ctx, roles...)
	if err != nil {
		return authctx.Claims{}, err
	}
	if claims.HasRole(domain.RoleSuperadmin.String()) {
		return claims, nil
	}
	if strings.TrimSpace(resourceSellerID) == "" || strings.TrimSpace(claims.SellerID) != strings.TrimSpace(resourceSellerID) {
		return authctx.Claims{}, domain.ErrForbidden
	}
	return claims, nil
}
