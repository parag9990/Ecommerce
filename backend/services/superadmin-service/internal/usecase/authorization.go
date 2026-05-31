package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
	"ecommerce/superadmin-service/internal/rbac"
)

type PermissionRepository interface {
	AdminHasPermission(ctx context.Context, adminID string, permission domain.Permission) (bool, error)
	ListPermissionsByRole(ctx context.Context, role domain.AdminRole) ([]domain.Permission, error)
	ListPermissionDefinitions(ctx context.Context) ([]domain.PermissionDefinition, error)
}

type AuthorizationService struct {
	repo   PermissionRepository
	logger logging.Logger
}

func NewAuthorizationService(repo PermissionRepository, logger logging.Logger) (*AuthorizationService, error) {
	if repo == nil {
		return nil, errors.New("authorization service requires permission repository")
	}
	if logger == nil {
		logger = logging.NewNop()
	}
	return &AuthorizationService{repo: repo, logger: logger}, nil
}

func (s *AuthorizationService) RequirePermission(ctx context.Context, actor domain.AdminActor, permission domain.Permission) error {
	if err := actor.ValidateForAdminRoute(); err != nil {
		return err
	}
	if !permission.Valid() {
		return domain.NewValidationError(fmt.Sprintf("unknown permission %q", permission))
	}

	allowed, err := s.repo.AdminHasPermission(ctx, actor.AdminID, permission)
	if err != nil {
		s.logger.Error(ctx, "admin permission lookup failed",
			"admin_id", actor.AdminID,
			"permission", permission,
			"request_id", actor.RequestID,
			"error", err,
		)
		return domain.NewInternal("permission lookup failed", err)
	}
	if !allowed {
		s.logger.Warn(ctx, "admin permission denied",
			"admin_id", actor.AdminID,
			"permission", permission,
			"request_id", actor.RequestID,
		)
		return domain.NewForbidden(permission)
	}

	return nil
}

func (s *AuthorizationService) HasPermission(ctx context.Context, actor domain.AdminActor, permission domain.Permission) (bool, error) {
	err := s.RequirePermission(ctx, actor, permission)
	if err == nil {
		return true, nil
	}
	if appErr, ok := domain.AsAppError(err); ok && appErr.Code == domain.CodeForbidden {
		return false, nil
	}
	return false, err
}

func (s *AuthorizationService) RequireHighRiskPermission(ctx context.Context, actor domain.AdminActor, permission domain.Permission, reason string) error {
	if err := s.RequirePermission(ctx, actor, permission); err != nil {
		return err
	}
	if err := rbac.ValidateHighRiskContext(rbac.PolicyForPermission(permission), actor, reason); err != nil {
		return err
	}
	return nil
}

func (s *AuthorizationService) PermissionCatalog(ctx context.Context, actor domain.AdminActor) ([]domain.PermissionDefinition, error) {
	if err := s.RequirePermission(ctx, actor, domain.PermissionAdminUsersRead); err != nil {
		return nil, err
	}

	definitions, err := s.repo.ListPermissionDefinitions(ctx)
	if err != nil {
		s.logger.Error(ctx, "permission catalog lookup failed",
			"admin_id", actor.AdminID,
			"request_id", actor.RequestID,
			"error", err,
		)
		return nil, domain.NewInternal("permission catalog lookup failed", err)
	}
	return definitions, nil
}

func (s *AuthorizationService) ListRolePermissions(ctx context.Context, actor domain.AdminActor, role domain.AdminRole) ([]domain.Permission, error) {
	if err := s.RequirePermission(ctx, actor, domain.PermissionAdminUsersRead); err != nil {
		return nil, err
	}
	if strings.TrimSpace(string(role)) == "" {
		return nil, domain.NewValidationError("role is required")
	}
	if !role.Valid() {
		return nil, domain.NewValidationError(fmt.Sprintf("unknown role %q", role))
	}

	permissions, err := s.repo.ListPermissionsByRole(ctx, role)
	if err != nil {
		s.logger.Error(ctx, "role permission lookup failed",
			"admin_id", actor.AdminID,
			"role", role,
			"request_id", actor.RequestID,
			"error", err,
		)
		return nil, domain.NewInternal("role permission lookup failed", err)
	}
	return permissions, nil
}

type AuthorizeRequest struct {
	Actor      domain.AdminActor
	Permission domain.Permission
	Reason     string
	HighRisk   bool
}

type AuthorizeResult struct {
	Allowed bool
}

func (s *AuthorizationService) Authorize(ctx context.Context, req AuthorizeRequest) (*AuthorizeResult, error) {
	if req.HighRisk {
		if err := s.RequireHighRiskPermission(ctx, req.Actor, req.Permission, req.Reason); err != nil {
			return &AuthorizeResult{Allowed: false}, err
		}
		return &AuthorizeResult{Allowed: true}, nil
	}

	allowed, err := s.HasPermission(ctx, req.Actor, req.Permission)
	if err != nil {
		return &AuthorizeResult{Allowed: false}, err
	}
	return &AuthorizeResult{Allowed: allowed}, nil
}
