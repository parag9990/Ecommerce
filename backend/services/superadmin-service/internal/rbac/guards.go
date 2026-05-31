package rbac

import (
	"strings"

	"ecommerce/superadmin-service/internal/domain"
)

type HighRiskPolicy struct {
	Permission       domain.Permission
	RequireMFA       bool
	RequireReason    bool
	RequireRequestID bool
}

func PolicyForPermission(permission domain.Permission) HighRiskPolicy {
	switch permission {
	case domain.PermissionRefundReview, domain.PermissionSettingsWrite:
		return HighRiskPolicy{Permission: permission, RequireMFA: true, RequireReason: true, RequireRequestID: true}
	case domain.PermissionUsersStatusUpdate,
		domain.PermissionSellersStatusUpdate,
		domain.PermissionOrderManualReview,
		domain.PermissionCatalogWrite,
		domain.PermissionSearchSynonymsWrite,
		domain.PermissionAuditLogsExport:
		return HighRiskPolicy{Permission: permission, RequireReason: true, RequireRequestID: true}
	default:
		return HighRiskPolicy{Permission: permission}
	}
}

func RequireHighRiskContext(actor domain.AdminActor, reason string) error {
	return ValidateHighRiskContext(HighRiskPolicy{
		RequireMFA:       true,
		RequireReason:    true,
		RequireRequestID: true,
	}, actor, reason)
}

func ValidateHighRiskContext(policy HighRiskPolicy, actor domain.AdminActor, reason string) error {
	if policy.RequireMFA && !actor.MFAVerified {
		return domain.NewMFARequired()
	}
	if policy.RequireReason && strings.TrimSpace(reason) == "" {
		return domain.NewReasonRequired()
	}
	if policy.RequireRequestID && strings.TrimSpace(actor.RequestID) == "" {
		return domain.NewRequestContextMissing()
	}
	return nil
}
