package authctx

import (
	"context"
	"strings"
)

const (
	MetadataUserID    = "x-user-id"
	MetadataSessionID = "x-session-id"
	MetadataRoles     = "x-roles"
	MetadataSellerID  = "x-seller-id"
	MetadataTenantID  = "x-tenant-id"
)

func ToMetadata(claims Claims) map[string]string {
	claims = claims.Normalized()
	values := map[string]string{
		MetadataUserID:    claims.UserID,
		MetadataSessionID: claims.SessionID,
		MetadataRoles:     strings.Join(claims.Roles, ","),
	}
	if claims.SellerID != "" {
		values[MetadataSellerID] = claims.SellerID
	}
	if claims.TenantID != "" {
		values[MetadataTenantID] = claims.TenantID
	}
	return values
}

func FromMetadata(get func(string) []string) (Claims, bool) {
	if get == nil {
		return Claims{}, false
	}
	claims := Claims{
		UserID:    firstMetadataValue(get(MetadataUserID)),
		SessionID: firstMetadataValue(get(MetadataSessionID)),
		SellerID:  firstMetadataValue(get(MetadataSellerID)),
		TenantID:  firstMetadataValue(get(MetadataTenantID)),
		Roles:     splitRoles(firstMetadataValue(get(MetadataRoles))),
	}.Normalized()
	return claims, claims.Authenticated()
}

func ContextWithMetadata(ctx context.Context, get func(string) []string) context.Context {
	if claims, ok := FromMetadata(get); ok {
		return WithClaims(ctx, claims)
	}
	return ctx
}

func firstMetadataValue(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[0])
}

func splitRoles(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return NormalizeRoles(strings.Split(value, ","))
}
