package usecase_test

import (
	"context"
	"testing"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
	"ecommerce/superadmin-service/internal/rbac"
	"ecommerce/superadmin-service/internal/usecase"
)

func TestRequirePermission(t *testing.T) {
	service := newAuthorizationService(t)
	actor := adminActor("finance_admin_1", domain.RoleFinance)

	if err := service.RequirePermission(context.Background(), actor, domain.PermissionRefundReview); err != nil {
		t.Fatalf("finance refund permission failed: %v", err)
	}

	err := service.RequirePermission(context.Background(), actor, domain.PermissionSellersStatusUpdate)
	if appErr, ok := domain.AsAppError(err); !ok || appErr.Code != domain.CodeForbidden {
		t.Fatalf("error = %v, want FORBIDDEN", err)
	}
}

func TestDisabledAdminDenied(t *testing.T) {
	repo := rbac.NewStaticPermissionRepository([]rbac.StaticAdmin{
		{AdminID: "disabled_admin", Role: domain.RoleSuperadmin, Active: false},
	})
	service, err := usecase.NewAuthorizationService(repo, logging.NewNop())
	if err != nil {
		t.Fatal(err)
	}

	err = service.RequirePermission(context.Background(), adminActor("disabled_admin", domain.RoleSuperadmin), domain.PermissionSettingsWrite)
	if appErr, ok := domain.AsAppError(err); !ok || appErr.Code != domain.CodeForbidden {
		t.Fatalf("error = %v, want FORBIDDEN", err)
	}
}

func TestRequireHighRiskPermission(t *testing.T) {
	service := newAuthorizationService(t)
	actor := adminActor("finance_admin_1", domain.RoleFinance)
	actor.MFAVerified = true

	if err := service.RequireHighRiskPermission(context.Background(), actor, domain.PermissionRefundReview, "refund policy met"); err != nil {
		t.Fatalf("RequireHighRiskPermission returned error: %v", err)
	}

	actor.MFAVerified = false
	err := service.RequireHighRiskPermission(context.Background(), actor, domain.PermissionRefundReview, "refund policy met")
	if appErr, ok := domain.AsAppError(err); !ok || appErr.Code != domain.CodeMFARequired {
		t.Fatalf("error = %v, want MFA_REQUIRED", err)
	}
}

func TestPermissionCatalogRequiresAdminUserRead(t *testing.T) {
	service := newAuthorizationService(t)
	actor := adminActor("finance_admin_1", domain.RoleFinance)

	_, err := service.PermissionCatalog(context.Background(), actor)
	if appErr, ok := domain.AsAppError(err); !ok || appErr.Code != domain.CodeForbidden {
		t.Fatalf("error = %v, want FORBIDDEN", err)
	}
}

func newAuthorizationService(t *testing.T) *usecase.AuthorizationService {
	t.Helper()
	repo := rbac.NewStaticPermissionRepository([]rbac.StaticAdmin{
		{AdminID: "superadmin_1", Role: domain.RoleSuperadmin, Active: true},
		{AdminID: "finance_admin_1", Role: domain.RoleFinance, Active: true},
	})
	service, err := usecase.NewAuthorizationService(repo, logging.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func adminActor(adminID string, role domain.AdminRole) domain.AdminActor {
	return domain.AdminActor{
		AdminID:   adminID,
		UserID:    "user_" + adminID,
		Roles:     []domain.AdminRole{role},
		SessionID: "sess_1",
		RequestID: "req_1",
	}
}
