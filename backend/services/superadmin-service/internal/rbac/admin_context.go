package rbac

import (
	"context"
	"strconv"
	"strings"

	"ecommerce/superadmin-service/internal/domain"
)

const (
	MetadataAdminID      = "x-admin-id"
	MetadataAdminRoles   = "x-admin-roles"
	MetadataRequestID    = "x-request-id"
	MetadataSessionID    = "x-session-id"
	MetadataActionReason = "x-action-reason"
	MetadataPermission   = "x-admin-permission"
	MetadataRiskAllowed  = "x-admin-risk-allowed"
	MetadataMaskPII      = "x-admin-mask-pii"
)

type adminMetadataContextKey struct{}

func AdminMetadata(actor domain.AdminActor, reason string) map[string]string {
	return map[string]string{
		MetadataAdminID:      actor.AdminID,
		MetadataAdminRoles:   strings.Join(domain.RolesToStrings(actor.Roles), ","),
		MetadataRequestID:    actor.RequestID,
		MetadataSessionID:    actor.SessionID,
		MetadataActionReason: strings.TrimSpace(reason),
	}
}

func SessionAnalyticsMetadata(actor domain.AdminActor, permission domain.Permission, riskAllowed bool, maskPII bool) map[string]string {
	values := AdminMetadata(actor, "")
	delete(values, MetadataActionReason)
	values[MetadataPermission] = string(permission)
	values[MetadataRiskAllowed] = strconv.FormatBool(riskAllowed)
	values[MetadataMaskPII] = strconv.FormatBool(maskPII)
	return values
}

func AttachAdminContext(ctx context.Context, actor domain.AdminActor, reason string) context.Context {
	return context.WithValue(ctx, adminMetadataContextKey{}, AdminMetadata(actor, reason))
}

func AdminMetadataFromContext(ctx context.Context) (map[string]string, bool) {
	values, ok := ctx.Value(adminMetadataContextKey{}).(map[string]string)
	if !ok {
		return nil, false
	}

	out := make(map[string]string, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out, true
}
