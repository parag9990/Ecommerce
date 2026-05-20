package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/authctx"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/authorization"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

const defaultRoleReasonMaxLength = 512

type RoleManagementRepository interface {
	ResolveAccountID(ctx context.Context, userOrAccountID string) (string, error)
	AssignRole(ctx context.Context, assignment domain.RoleAssignment) error
	RevokeRole(ctx context.Context, accountID string, role domain.Role, scopeType string, scopeID string, revokedAt time.Time) error
	GetRoleAssignments(ctx context.Context, accountID string, includeRevoked bool) ([]domain.RoleAssignment, error)
}

type RoleUsecaseConfig struct {
	ReasonMaxLength int
}

type RoleUsecase struct {
	repo            RoleManagementRepository
	logger          *slog.Logger
	clock           Clock
	reasonMaxLength int
}

func NewRoleUsecase(repo RoleManagementRepository, cfg RoleUsecaseConfig, logger *slog.Logger) (*RoleUsecase, error) {
	if repo == nil {
		return nil, errors.New("role repository is required")
	}
	if cfg.ReasonMaxLength <= 0 {
		cfg.ReasonMaxLength = defaultRoleReasonMaxLength
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &RoleUsecase{
		repo:            repo,
		logger:          logger,
		clock:           realClock{},
		reasonMaxLength: cfg.ReasonMaxLength,
	}, nil
}

func (u *RoleUsecase) WithClock(clock Clock) {
	if clock != nil {
		u.clock = clock
	}
}

type AssignRoleInput struct {
	TargetUserID string
	Role         string
	ScopeType    string
	ScopeID      string
	Reason       string
	RequestID    string
}

type RevokeRoleInput struct {
	TargetUserID string
	Role         string
	ScopeType    string
	ScopeID      string
	Reason       string
	RequestID    string
}

type GetUserRolesInput struct {
	UserID         string
	IncludeRevoked bool
}

type UserRoles struct {
	UserID    string
	AccountID string
	Roles     []domain.RoleAssignment
}

func (u *RoleUsecase) AssignRole(ctx context.Context, input AssignRoleInput) (domain.RoleAssignment, error) {
	actor, err := authorization.RequireAuthenticated(ctx)
	if err != nil {
		return domain.RoleAssignment{}, err
	}

	targetUserID := strings.TrimSpace(input.TargetUserID)
	if targetUserID == "" {
		return domain.RoleAssignment{}, domain.ErrInvalidRoleAssignment
	}

	role, scopeType, scopeID, reason, err := u.normalizeMutation(input.Role, input.ScopeType, input.ScopeID, input.Reason)
	if err != nil {
		return domain.RoleAssignment{}, err
	}
	if err := validateRoleScope(role, scopeType, scopeID); err != nil {
		return domain.RoleAssignment{}, err
	}
	if err := canAssignRole(actor, targetUserID, role, scopeType, scopeID); err != nil {
		return domain.RoleAssignment{}, err
	}

	accountID, err := u.repo.ResolveAccountID(ctx, targetUserID)
	if err != nil {
		return domain.RoleAssignment{}, fmt.Errorf("resolve target account: %w", err)
	}

	assignment := domain.RoleAssignment{
		AccountID:  accountID,
		Role:       role,
		ScopeType:  scopeType,
		ScopeID:    scopeID,
		AssignedBy: actor.UserID,
		Reason:     reason,
		AssignedAt: u.clock.Now(),
	}
	if err := u.repo.AssignRole(ctx, assignment); err != nil {
		return domain.RoleAssignment{}, err
	}

	u.logger.InfoContext(ctx, "auth.role.assign",
		slog.String("actor_user_id", actor.UserID),
		slog.Any("actor_roles", actor.Roles),
		slog.String("target_user_id", targetUserID),
		slog.String("target_account_id", accountID),
		slog.String("role", role.String()),
		slog.String("scope_type", scopeType),
		slog.String("scope_id", scopeID),
		slog.String("request_id", strings.TrimSpace(input.RequestID)),
		slog.String("result", "success"),
	)

	return assignment, nil
}

func (u *RoleUsecase) RevokeRole(ctx context.Context, input RevokeRoleInput) error {
	actor, err := authorization.RequireAuthenticated(ctx)
	if err != nil {
		return err
	}

	targetUserID := strings.TrimSpace(input.TargetUserID)
	if targetUserID == "" {
		return domain.ErrInvalidRoleAssignment
	}

	role, scopeType, scopeID, _, err := u.normalizeMutation(input.Role, input.ScopeType, input.ScopeID, input.Reason)
	if err != nil {
		return err
	}
	if err := validateRoleScope(role, scopeType, scopeID); err != nil {
		return err
	}
	if err := canRevokeRole(actor, targetUserID, role, scopeType, scopeID); err != nil {
		return err
	}

	accountID, err := u.repo.ResolveAccountID(ctx, targetUserID)
	if err != nil {
		return fmt.Errorf("resolve target account: %w", err)
	}

	if err := u.repo.RevokeRole(ctx, accountID, role, scopeType, scopeID, u.clock.Now()); err != nil {
		return err
	}

	u.logger.InfoContext(ctx, "auth.role.revoke",
		slog.String("actor_user_id", actor.UserID),
		slog.Any("actor_roles", actor.Roles),
		slog.String("target_user_id", targetUserID),
		slog.String("target_account_id", accountID),
		slog.String("role", role.String()),
		slog.String("scope_type", scopeType),
		slog.String("scope_id", scopeID),
		slog.String("request_id", strings.TrimSpace(input.RequestID)),
		slog.String("result", "success"),
	)

	return nil
}

func (u *RoleUsecase) GetUserRoles(ctx context.Context, input GetUserRolesInput) (UserRoles, error) {
	actor, err := authorization.RequireAuthenticated(ctx)
	if err != nil {
		return UserRoles{}, err
	}

	targetUserID := strings.TrimSpace(input.UserID)
	if targetUserID == "" {
		targetUserID = actor.UserID
	}
	if targetUserID == "" {
		return UserRoles{}, domain.ErrInvalidRoleAssignment
	}

	includeRevoked := input.IncludeRevoked && actor.HasRole(domain.RoleSuperadmin.String())
	if targetUserID != actor.UserID && !actorCanReadUserRoles(actor) {
		return UserRoles{}, domain.ErrForbidden
	}

	accountID, err := u.repo.ResolveAccountID(ctx, targetUserID)
	if err != nil {
		return UserRoles{}, fmt.Errorf("resolve target account: %w", err)
	}

	roles, err := u.repo.GetRoleAssignments(ctx, accountID, includeRevoked)
	if err != nil {
		return UserRoles{}, err
	}

	return UserRoles{
		UserID:    targetUserID,
		AccountID: accountID,
		Roles:     roles,
	}, nil
}

func (u *RoleUsecase) RequireFreshRole(ctx context.Context, userID string, allowed ...domain.Role) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return domain.ErrUnauthenticated
	}

	accountID, err := u.repo.ResolveAccountID(ctx, userID)
	if err != nil {
		return fmt.Errorf("resolve target account: %w", err)
	}
	roles, err := u.repo.GetRoleAssignments(ctx, accountID, false)
	if err != nil {
		return err
	}

	allowedSet := make(map[domain.Role]struct{}, len(allowed))
	for _, role := range allowed {
		allowedSet[role] = struct{}{}
	}
	for _, assignment := range roles {
		if _, ok := allowedSet[assignment.Role]; ok {
			return nil
		}
	}
	return domain.ErrForbidden
}

func (u *RoleUsecase) normalizeMutation(roleValue string, scopeTypeValue string, scopeIDValue string, reasonValue string) (domain.Role, string, string, string, error) {
	role := domain.Role(strings.TrimSpace(roleValue))
	if !domain.IsKnownRole(role) {
		return "", "", "", "", domain.ErrUnknownRole
	}
	scopeType := strings.TrimSpace(scopeTypeValue)
	scopeID := strings.TrimSpace(scopeIDValue)
	reason := strings.TrimSpace(reasonValue)
	if reason == "" {
		return "", "", "", "", domain.ErrAuditReasonRequired
	}
	if len(reason) > u.reasonMaxLength {
		return "", "", "", "", domain.ErrInvalidRoleAssignment
	}
	return role, scopeType, scopeID, reason, nil
}

func validateRoleScope(role domain.Role, scopeType string, scopeID string) error {
	if domain.IsSellerScopedRole(role) {
		if scopeType != domain.ScopeTypeSeller || scopeID == "" {
			return domain.ErrInvalidRoleScope
		}
		return nil
	}
	if scopeType != "" || scopeID != "" {
		return domain.ErrInvalidRoleScope
	}
	return nil
}

func canAssignRole(actor authctx.Claims, targetUserID string, role domain.Role, scopeType string, scopeID string) error {
	if strings.TrimSpace(actor.UserID) == strings.TrimSpace(targetUserID) && role != domain.RoleBuyer {
		return domain.ErrForbidden
	}

	switch role {
	case domain.RoleBuyer:
		return requireActorRole(actor, domain.RoleOperationsAdmin, domain.RoleSuperadmin)
	case domain.RoleSeller:
		return requireActorRole(actor, domain.RoleOperationsAdmin, domain.RoleSuperadmin)
	case domain.RoleSellerManager, domain.RoleSellerCatalogEditor, domain.RoleSellerOrderManager:
		if actor.HasRole(domain.RoleSuperadmin.String()) {
			return nil
		}
		if actor.HasAnyRoleList(domain.RoleSeller, domain.RoleSellerManager) && strings.TrimSpace(actor.SellerID) == scopeID {
			return nil
		}
		return domain.ErrForbidden
	case domain.RoleAdmin,
		domain.RoleOperationsAdmin,
		domain.RoleFinanceAdmin,
		domain.RoleCatalogAdmin,
		domain.RoleReadonlyAdmin,
		domain.RoleSuperadmin:
		return requireActorRole(actor, domain.RoleSuperadmin)
	default:
		return domain.ErrUnknownRole
	}
}

func canRevokeRole(actor authctx.Claims, targetUserID string, role domain.Role, scopeType string, scopeID string) error {
	if strings.TrimSpace(actor.UserID) == strings.TrimSpace(targetUserID) {
		return domain.ErrForbidden
	}
	return canAssignRole(actor, targetUserID, role, scopeType, scopeID)
}

func requireActorRole(actor authctx.Claims, roles ...domain.Role) error {
	if actor.HasAnyRoleList(roles...) {
		return nil
	}
	return domain.ErrForbidden
}

func actorCanReadUserRoles(actor authctx.Claims) bool {
	return actor.HasAnyRoleList(
		domain.RoleAdmin,
		domain.RoleOperationsAdmin,
		domain.RoleFinanceAdmin,
		domain.RoleCatalogAdmin,
		domain.RoleReadonlyAdmin,
		domain.RoleSuperadmin,
	)
}
