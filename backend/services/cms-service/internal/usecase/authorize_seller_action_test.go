package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
)

type auditRecorderFunc func(context.Context, domain.AuditEvent) error

func (f auditRecorderFunc) Record(ctx context.Context, event domain.AuditEvent) error {
	return f(ctx, event)
}

type staffRepositoryFunc func(context.Context, string, string) (domain.SellerStaff, error)

func (f staffRepositoryFunc) GetBySellerAndUser(ctx context.Context, sellerID string, userID string) (domain.SellerStaff, error) {
	return f(ctx, sellerID, userID)
}

func TestAuthorizerAllowsSellerUpdatingOwnSettings(t *testing.T) {
	authorizer := newTestAuthorizer(t, nil, nil)

	decision, err := authorizer.Authorize(context.Background(), AuthorizeInput{
		Actor:              activeActor(domain.RoleSeller),
		ResourceSellerID:   "seller_1",
		RequiredPermission: domain.PermissionSettingsUpdate,
		ResourceType:       "settings",
		ResourceID:         "seller_1",
	})
	if err != nil {
		t.Fatalf("expected allow: %v", err)
	}
	if !decision.Allowed || !decision.AuditRequired {
		t.Fatalf("expected allowed audited decision, got %+v", decision)
	}
}

func TestAuthorizerDeniesCatalogEditorUpdatingSettings(t *testing.T) {
	authorizer := newTestAuthorizer(t, nil, nil)

	decision, err := authorizer.Authorize(context.Background(), AuthorizeInput{
		Actor:              activeActor(domain.RoleSellerCatalogEditor),
		ResourceSellerID:   "seller_1",
		RequiredPermission: domain.PermissionSettingsUpdate,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden, got decision=%+v err=%v", decision, err)
	}
	if decision.Allowed {
		t.Fatal("expected denied decision")
	}
}

func TestAuthorizerAllowsOrderManagerReadingOwnOrder(t *testing.T) {
	authorizer := newTestAuthorizer(t, nil, nil)

	decision, err := authorizer.Authorize(context.Background(), AuthorizeInput{
		Actor:              activeActor(domain.RoleSellerOrderManager),
		ResourceSellerID:   "seller_1",
		RequiredPermission: domain.PermissionOrdersRead,
		ResourceType:       "order",
		ResourceID:         "order_1",
	})
	if err != nil {
		t.Fatalf("expected allow: %v", err)
	}
	if !decision.Allowed || decision.AuditRequired {
		t.Fatalf("expected read allow without audit, got %+v", decision)
	}
}

func TestAuthorizerDeniesOrderManagerUpdatingProduct(t *testing.T) {
	authorizer := newTestAuthorizer(t, nil, nil)

	_, err := authorizer.Authorize(context.Background(), AuthorizeInput{
		Actor:              activeActor(domain.RoleSellerOrderManager),
		ResourceSellerID:   "seller_1",
		RequiredPermission: domain.PermissionProductsUpdate,
		ResourceType:       "product",
		ResourceID:         "product_1",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestAuthorizerDeniesCatalogEditorReadingOrders(t *testing.T) {
	authorizer := newTestAuthorizer(t, nil, nil)

	_, err := authorizer.Authorize(context.Background(), AuthorizeInput{
		Actor:              activeActor(domain.RoleSellerCatalogEditor),
		ResourceSellerID:   "seller_1",
		RequiredPermission: domain.PermissionOrdersRead,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestAuthorizerDeniesCrossSellerAccess(t *testing.T) {
	authorizer := newTestAuthorizer(t, nil, nil)

	decision, err := authorizer.Authorize(context.Background(), AuthorizeInput{
		Actor:              activeActor(domain.RoleSellerOrderManager),
		ResourceSellerID:   "seller_2",
		RequiredPermission: domain.PermissionOrdersRead,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden, got decision=%+v err=%v", decision, err)
	}
	if decision.Reason != "cross_seller_access" {
		t.Fatalf("expected cross seller reason, got %q", decision.Reason)
	}
}

func TestAuthorizerDeniesInactiveStaff(t *testing.T) {
	authorizer := newTestAuthorizer(t, nil, nil)
	actor := activeActor(domain.RoleSeller)
	actor.StaffStatus = domain.StaffStatusDisabled

	_, err := authorizer.Authorize(context.Background(), AuthorizeInput{
		Actor:              actor,
		ResourceSellerID:   "seller_1",
		RequiredPermission: domain.PermissionSettingsRead,
	})
	if !errors.Is(err, domain.ErrInactiveStaff) {
		t.Fatalf("expected inactive staff, got %v", err)
	}
}

func TestAuthorizerDeniesUnknownPermission(t *testing.T) {
	authorizer := newTestAuthorizer(t, nil, nil)

	_, err := authorizer.Authorize(context.Background(), AuthorizeInput{
		Actor:              activeActor(domain.RoleSeller),
		ResourceSellerID:   "seller_1",
		RequiredPermission: "cms:unknown:read",
	})
	if !errors.Is(err, domain.ErrUnknownPermission) {
		t.Fatalf("expected unknown permission, got %v", err)
	}
}

func TestAuthorizerDeniesManagerInvitingSellerRole(t *testing.T) {
	authorizer := newTestAuthorizer(t, nil, nil)

	_, err := authorizer.Authorize(context.Background(), AuthorizeInput{
		Actor:              activeActor(domain.RoleSellerManager),
		ResourceSellerID:   "seller_1",
		RequiredPermission: domain.PermissionStaffInvite,
		TargetRole:         domain.RoleSeller,
	})
	if !errors.Is(err, domain.ErrInvalidRoleConstraint) {
		t.Fatalf("expected role constraint, got %v", err)
	}
}

func TestAuthorizerUsesStaffRepositoryRoleWhenConfigured(t *testing.T) {
	repo := staffRepositoryFunc(func(context.Context, string, string) (domain.SellerStaff, error) {
		return domain.SellerStaff{
			SellerID: "seller_1",
			UserID:   "user_1",
			Role:     domain.RoleSellerOrderManager,
			Status:   domain.StaffStatusActive,
		}, nil
	})
	authorizer := newTestAuthorizer(t, repo, nil)
	actor := activeActor(domain.RoleSellerCatalogEditor)

	_, err := authorizer.Authorize(context.Background(), AuthorizeInput{
		Actor:              actor,
		ResourceSellerID:   "seller_1",
		RequiredPermission: domain.PermissionProductsUpdate,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected repository role to deny product update, got %v", err)
	}
}

func TestAuthorizerFailsClosedWhenAuditRecorderFails(t *testing.T) {
	authorizer := newTestAuthorizer(t, nil, auditRecorderFunc(func(context.Context, domain.AuditEvent) error {
		return errors.New("sink down")
	}))

	_, err := authorizer.Authorize(context.Background(), AuthorizeInput{
		Actor:              activeActor(domain.RoleSeller),
		ResourceSellerID:   "seller_1",
		RequiredPermission: domain.PermissionProductsCreate,
		ResourceType:       "product",
		ResourceID:         "product_1",
	})
	if !errors.Is(err, domain.ErrAuditWriteFailed) {
		t.Fatalf("expected audit write failure, got %v", err)
	}
}

func newTestAuthorizer(t *testing.T, staffRepo SellerStaffRepository, auditRecorder AuditRecorder) *Authorizer {
	t.Helper()
	if auditRecorder == nil {
		auditRecorder = auditRecorderFunc(func(context.Context, domain.AuditEvent) error { return nil })
	}
	authorizer, err := NewAuthorizer(staffRepo, auditRecorder, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("NewAuthorizer: %v", err)
	}
	return authorizer
}

func activeActor(role domain.Role) domain.ActorContext {
	return domain.ActorContext{
		UserID:      "user_1",
		SellerID:    "seller_1",
		Roles:       []domain.Role{role},
		StaffStatus: domain.StaffStatusActive,
		RequestID:   "req_1",
	}
}
