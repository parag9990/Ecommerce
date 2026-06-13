package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
)

type SellerStaffRepository interface {
	GetBySellerAndUser(ctx context.Context, sellerID string, userID string) (domain.SellerStaff, error)
}

type AuditRecorder interface {
	Record(ctx context.Context, event domain.AuditEvent) error
}

type Authorizer struct {
	staffRepo     SellerStaffRepository
	auditRecorder AuditRecorder
	logger        *slog.Logger
	now           func() time.Time
}

type AuthorizeInput struct {
	Actor              domain.ActorContext
	ResourceSellerID   string
	RequiredPermission domain.Permission
	ResourceType       string
	ResourceID         string
	TargetRole         domain.Role
	RequestID          string
}

type AuthorizationDecision struct {
	Allowed            bool
	Reason             string
	RequiredPermission domain.Permission
	AuditRequired      bool
	EffectiveRoles     []domain.Role
	RequestID          string
}

func NewAuthorizer(staffRepo SellerStaffRepository, auditRecorder AuditRecorder, logger *slog.Logger) (*Authorizer, error) {
	if auditRecorder == nil {
		return nil, domain.ErrAuditRecorderRequired
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Authorizer{
		staffRepo:     staffRepo,
		auditRecorder: auditRecorder,
		logger:        logger,
		now:           func() time.Time { return time.Now().UTC() },
	}, nil
}

func (a *Authorizer) Authorize(ctx context.Context, input AuthorizeInput) (AuthorizationDecision, error) {
	input = input.normalized()
	decision := AuthorizationDecision{
		RequiredPermission: input.RequiredPermission,
		AuditRequired:      domain.IsMutationPermission(input.RequiredPermission),
		RequestID:          input.requestID(),
	}

	if !domain.IsKnownPermission(input.RequiredPermission) {
		return a.deny(decision, "unknown_permission", domain.ErrUnknownPermission)
	}

	actor := input.Actor.Normalized()
	if !actor.Authenticated() || actor.SellerID == "" {
		return a.deny(decision, "unauthenticated", domain.ErrUnauthenticated)
	}
	if input.ResourceSellerID == "" {
		return a.deny(decision, "invalid_seller_scope", domain.ErrInvalidSellerScope)
	}
	if actor.SellerID != input.ResourceSellerID {
		return a.deny(decision, "cross_seller_access", domain.ErrForbidden)
	}

	effectiveRoles, err := a.effectiveRoles(ctx, actor, input.ResourceSellerID)
	if err != nil {
		reason := "inactive_staff"
		if errors.Is(err, domain.ErrUnknownRole) {
			reason = "unknown_role"
		} else if errors.Is(err, domain.ErrStaffLookupFailed) {
			reason = "staff_lookup_failed"
		}
		return a.deny(decision, reason, err)
	}
	decision.EffectiveRoles = effectiveRoles

	if violates, err := violatesRoleConstraint(input, effectiveRoles); violates {
		return a.deny(decision, "role_constraint_denied", err)
	}

	if !rolesHavePermission(effectiveRoles, input.RequiredPermission) {
		return a.deny(decision, "permission_denied", domain.ErrForbidden)
	}

	if decision.AuditRequired {
		if err := a.recordAudit(ctx, input, actor, effectiveRoles, domain.AuditDecisionAllowed); err != nil {
			a.logger.ErrorContext(ctx, "cms.audit.write_failed",
				slog.String("request_id", decision.RequestID),
				slog.String("permission", input.RequiredPermission.String()),
				slog.String("error", err.Error()),
			)
			return a.deny(decision, "audit_write_failed", fmt.Errorf("%w: %v", domain.ErrAuditWriteFailed, err))
		}
	}

	decision.Allowed = true
	decision.Reason = "allowed"
	return decision, nil
}

func (a *Authorizer) effectiveRoles(ctx context.Context, actor domain.ActorContext, resourceSellerID string) ([]domain.Role, error) {
	if a.staffRepo == nil {
		if !actor.StaffStatus.IsActive() {
			return nil, domain.ErrInactiveStaff
		}
		return knownRoles(actor.Roles)
	}

	staff, err := a.staffRepo.GetBySellerAndUser(ctx, resourceSellerID, actor.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrSellerStaffNotFound) {
			return nil, domain.ErrInactiveStaff
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrStaffLookupFailed, err)
	}
	if !staff.Active() {
		return nil, domain.ErrInactiveStaff
	}
	if !domain.IsKnownRole(staff.Role) {
		return nil, domain.ErrUnknownRole
	}
	return []domain.Role{staff.Role}, nil
}

func (a *Authorizer) recordAudit(ctx context.Context, input AuthorizeInput, actor domain.ActorContext, roles []domain.Role, decision domain.AuditDecision) error {
	info, _ := domain.PermissionMetadata(input.RequiredPermission)
	resourceType := input.ResourceType
	if resourceType == "" {
		resourceType = info.Resource
	}
	return a.auditRecorder.Record(ctx, domain.AuditEvent{
		AuditID:          newAuditID(),
		ActorUserID:      actor.UserID,
		ActorSellerID:    actor.SellerID,
		ActorRoles:       roles,
		Action:           input.RequiredPermission,
		ResourceType:     resourceType,
		ResourceID:       input.ResourceID,
		ResourceSellerID: input.ResourceSellerID,
		RequestID:        input.requestID(),
		Decision:         decision,
		CreatedAt:        a.now(),
	})
}

func (a *Authorizer) deny(decision AuthorizationDecision, reason string, err error) (AuthorizationDecision, error) {
	decision.Allowed = false
	decision.Reason = reason
	return decision, err
}

func (i AuthorizeInput) normalized() AuthorizeInput {
	i.Actor = i.Actor.Normalized()
	i.ResourceSellerID = strings.TrimSpace(i.ResourceSellerID)
	i.RequiredPermission = domain.Permission(strings.TrimSpace(string(i.RequiredPermission)))
	i.ResourceType = strings.TrimSpace(i.ResourceType)
	i.ResourceID = strings.TrimSpace(i.ResourceID)
	i.TargetRole = domain.Role(strings.TrimSpace(string(i.TargetRole)))
	i.RequestID = strings.TrimSpace(i.RequestID)
	return i
}

func (i AuthorizeInput) requestID() string {
	if i.RequestID != "" {
		return i.RequestID
	}
	return i.Actor.RequestID
}

func rolesHavePermission(roles []domain.Role, permission domain.Permission) bool {
	return domain.RolesHavePermission(roles, permission)
}

func knownRoles(roles []domain.Role) ([]domain.Role, error) {
	out := make([]domain.Role, 0, len(roles))
	for _, role := range roles {
		if !domain.IsKnownRole(role) {
			continue
		}
		out = append(out, role)
	}
	if len(out) == 0 {
		return nil, domain.ErrUnknownRole
	}
	return out, nil
}

func violatesRoleConstraint(input AuthorizeInput, roles []domain.Role) (bool, error) {
	if input.RequiredPermission != domain.PermissionStaffInvite || input.TargetRole == "" {
		return false, nil
	}
	if !domain.IsKnownRole(input.TargetRole) {
		return true, domain.ErrUnknownRole
	}
	for _, role := range roles {
		if role == domain.RoleSeller {
			return false, nil
		}
	}
	if input.TargetRole == domain.RoleSeller {
		return true, domain.ErrInvalidRoleConstraint
	}
	return false, nil
}

func newAuditID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("audit_%d", time.Now().UTC().UnixNano())
	}
	return "audit_" + hex.EncodeToString(bytes[:])
}
