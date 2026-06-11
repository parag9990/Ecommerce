package authorization

import (
	"context"
	"errors"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/authctx"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

func TestRequireAuthLevelRoleBypassMatrix(t *testing.T) {
	tests := []struct {
		name      string
		claims    authctx.Claims
		level     AuthLevel
		wantError error
	}{
		{
			name:      "missing token denied on buyer route",
			claims:    authctx.Claims{},
			level:     AuthLevelBuyer,
			wantError: domain.ErrUnauthenticated,
		},
		{
			name:      "buyer allowed on buyer route",
			claims:    authctx.Claims{UserID: "user_buyer", Roles: []string{"buyer"}},
			level:     AuthLevelBuyer,
			wantError: nil,
		},
		{
			name:      "buyer denied on seller route",
			claims:    authctx.Claims{UserID: "user_buyer", Roles: []string{"buyer"}},
			level:     AuthLevelSeller,
			wantError: domain.ErrForbidden,
		},
		{
			name:      "seller denied on admin route",
			claims:    authctx.Claims{UserID: "user_seller", Roles: []string{"seller"}},
			level:     AuthLevelAdmin,
			wantError: domain.ErrForbidden,
		},
		{
			name:      "admin denied on superadmin route",
			claims:    authctx.Claims{UserID: "user_admin", Roles: []string{"admin"}},
			level:     AuthLevelSuperadmin,
			wantError: domain.ErrForbidden,
		},
		{
			name:      "superadmin allowed on superadmin route",
			claims:    authctx.Claims{UserID: "user_super", Roles: []string{"superadmin"}},
			level:     AuthLevelSuperadmin,
			wantError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.claims.UserID != "" || len(tt.claims.Roles) > 0 {
				ctx = authctx.WithClaims(ctx, tt.claims)
			}

			_, err := RequireAuthLevel(ctx, tt.level)
			if tt.wantError == nil {
				if err != nil {
					t.Fatalf("RequireAuthLevel() error = %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantError) {
				t.Fatalf("RequireAuthLevel() error = %v, want %v", err, tt.wantError)
			}
		})
	}
}

func TestRequireSellerScopeDeniesDifferentSeller(t *testing.T) {
	ctx := authctx.WithClaims(context.Background(), authctx.Claims{
		UserID:   "user_seller_1",
		SellerID: "seller_1",
		Roles:    []string{"seller"},
	})

	_, err := RequireSellerScope(ctx, "seller_2")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("RequireSellerScope() error = %v, want forbidden", err)
	}
}

func TestRequireSellerScopeAllowsSuperadminAcrossSellers(t *testing.T) {
	ctx := authctx.WithClaims(context.Background(), authctx.Claims{
		UserID: "user_super",
		Roles:  []string{"superadmin"},
	})

	if _, err := RequireSellerScope(ctx, "seller_2"); err != nil {
		t.Fatalf("RequireSellerScope() error = %v", err)
	}
}

func TestRequirePermissionDeniesRoleBypass(t *testing.T) {
	tests := []struct {
		name       string
		roles      []string
		permission Permission
		wantError  error
	}{
		{
			name:       "buyer cannot write seller product",
			roles:      []string{"buyer"},
			permission: PermissionProductWriteOwn,
			wantError:  domain.ErrForbidden,
		},
		{
			name:       "seller order manager cannot write catalog",
			roles:      []string{"seller_order_manager"},
			permission: PermissionProductWriteOwn,
			wantError:  domain.ErrForbidden,
		},
		{
			name:       "finance admin can review refunds",
			roles:      []string{"finance_admin"},
			permission: PermissionPaymentRefundReview,
			wantError:  nil,
		},
		{
			name:       "admin cannot write platform settings",
			roles:      []string{"admin"},
			permission: PermissionPlatformWrite,
			wantError:  domain.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := authctx.WithClaims(context.Background(), authctx.Claims{
				UserID: "user_123",
				Roles:  tt.roles,
			})

			_, err := RequirePermission(ctx, tt.permission)
			if tt.wantError == nil {
				if err != nil {
					t.Fatalf("RequirePermission() error = %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantError) {
				t.Fatalf("RequirePermission() error = %v, want %v", err, tt.wantError)
			}
		})
	}
}
