package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/authctx"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

func TestAssignRoleBuyerCannotEscalateToAdmin(t *testing.T) {
	repo := newSecurityRoleManagementRepository()
	uc := newSecurityRoleUsecase(t, repo)
	ctx := authctx.WithClaims(context.Background(), authctx.Claims{
		UserID: "user_buyer",
		Roles:  []string{"buyer"},
	})

	_, err := uc.AssignRole(ctx, AssignRoleInput{
		TargetUserID: "user_target",
		Role:         domain.RoleAdmin.String(),
		Reason:       "security test",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("AssignRole() error = %v, want forbidden", err)
	}
	if repo.assignCalled {
		t.Fatal("repository AssignRole must not be called for denied escalation")
	}
}

func TestAssignRoleSellerCannotAssignScopedRoleForDifferentSeller(t *testing.T) {
	repo := newSecurityRoleManagementRepository()
	uc := newSecurityRoleUsecase(t, repo)
	ctx := authctx.WithClaims(context.Background(), authctx.Claims{
		UserID:   "user_seller_manager",
		SellerID: "seller_1",
		Roles:    []string{"seller_manager"},
	})

	_, err := uc.AssignRole(ctx, AssignRoleInput{
		TargetUserID: "user_target",
		Role:         domain.RoleSellerCatalogEditor.String(),
		ScopeType:    domain.ScopeTypeSeller,
		ScopeID:      "seller_2",
		Reason:       "catalog coverage",
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("AssignRole() error = %v, want forbidden", err)
	}
	if repo.assignCalled {
		t.Fatal("repository AssignRole must not be called for cross-seller escalation")
	}
}

func TestAssignRoleSuperadminCanAssignAdmin(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	repo := newSecurityRoleManagementRepository()
	uc := newSecurityRoleUsecase(t, repo)
	uc.WithClock(fixedClock{now: now})
	ctx := authctx.WithClaims(context.Background(), authctx.Claims{
		UserID: "user_super",
		Roles:  []string{"superadmin"},
	})

	assignment, err := uc.AssignRole(ctx, AssignRoleInput{
		TargetUserID: "user_target",
		Role:         domain.RoleAdmin.String(),
		Reason:       "break glass approval",
	})
	if err != nil {
		t.Fatalf("AssignRole() error = %v", err)
	}
	if !repo.assignCalled {
		t.Fatal("repository AssignRole was not called")
	}
	if assignment.AccountID != "auth_target" || assignment.AssignedBy != "user_super" || assignment.AssignedAt != now {
		t.Fatalf("assignment = %+v", assignment)
	}
}

func TestGetUserRolesBuyerCannotReadAnotherUserRoles(t *testing.T) {
	repo := newSecurityRoleManagementRepository()
	uc := newSecurityRoleUsecase(t, repo)
	ctx := authctx.WithClaims(context.Background(), authctx.Claims{
		UserID: "user_buyer",
		Roles:  []string{"buyer"},
	})

	_, err := uc.GetUserRoles(ctx, GetUserRolesInput{UserID: "user_target"})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("GetUserRoles() error = %v, want forbidden", err)
	}
	if repo.resolveCalled {
		t.Fatal("repository ResolveAccountID must not be called for denied read")
	}
}

func TestRequireFreshRoleIgnoresRevokedAssignments(t *testing.T) {
	revokedAt := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	repo := newSecurityRoleManagementRepository()
	repo.assignments["auth_target"] = []domain.RoleAssignment{
		{
			AccountID: "auth_target",
			Role:      domain.RoleAdmin,
			RevokedAt: &revokedAt,
		},
	}
	uc := newSecurityRoleUsecase(t, repo)

	err := uc.RequireFreshRole(context.Background(), "user_target", domain.RoleAdmin)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("RequireFreshRole() error = %v, want forbidden", err)
	}
}

func newSecurityRoleUsecase(t *testing.T, repo *securityRoleManagementRepository) *RoleUsecase {
	t.Helper()

	uc, err := NewRoleUsecase(repo, RoleUsecaseConfig{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("NewRoleUsecase() error = %v", err)
	}
	return uc
}

type securityRoleManagementRepository struct {
	resolved      map[string]string
	assignments   map[string][]domain.RoleAssignment
	resolveCalled bool
	assignCalled  bool
}

func newSecurityRoleManagementRepository() *securityRoleManagementRepository {
	return &securityRoleManagementRepository{
		resolved: map[string]string{
			"user_target": "auth_target",
		},
		assignments: map[string][]domain.RoleAssignment{},
	}
}

func (r *securityRoleManagementRepository) ResolveAccountID(ctx context.Context, userOrAccountID string) (string, error) {
	r.resolveCalled = true
	accountID, ok := r.resolved[userOrAccountID]
	if !ok {
		return "", domain.ErrAccountNotFound
	}
	return accountID, nil
}

func (r *securityRoleManagementRepository) AssignRole(ctx context.Context, assignment domain.RoleAssignment) error {
	r.assignCalled = true
	r.assignments[assignment.AccountID] = append(r.assignments[assignment.AccountID], assignment)
	return nil
}

func (r *securityRoleManagementRepository) RevokeRole(ctx context.Context, accountID string, role domain.Role, scopeType string, scopeID string, revokedAt time.Time) error {
	for i, assignment := range r.assignments[accountID] {
		if assignment.Role == role && assignment.ScopeType == scopeType && assignment.ScopeID == scopeID && assignment.RevokedAt == nil {
			assignment.RevokedAt = &revokedAt
			r.assignments[accountID][i] = assignment
			return nil
		}
	}
	return domain.ErrRoleAssignmentNotFound
}

func (r *securityRoleManagementRepository) GetRoleAssignments(ctx context.Context, accountID string, includeRevoked bool) ([]domain.RoleAssignment, error) {
	source := r.assignments[accountID]
	roles := make([]domain.RoleAssignment, 0, len(source))
	for _, assignment := range source {
		if !includeRevoked && assignment.RevokedAt != nil {
			continue
		}
		roles = append(roles, assignment)
	}
	return roles, nil
}
