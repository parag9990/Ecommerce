package authctx

import (
	"context"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

type contextKey struct{}

type Claims struct {
	UserID    string
	SessionID string
	SellerID  string
	TenantID  string
	TokenID   string
	Roles     []string
}

func FromTokenClaims(claims domain.TokenClaims) Claims {
	return Claims{
		UserID:    strings.TrimSpace(claims.UserID),
		SessionID: strings.TrimSpace(claims.SessionID),
		SellerID:  strings.TrimSpace(claims.SellerID),
		TenantID:  strings.TrimSpace(claims.TenantID),
		TokenID:   strings.TrimSpace(claims.TokenID),
		Roles:     NormalizeRoles(claims.Roles),
	}
}

func WithClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, contextKey{}, claims.Normalized())
}

func FromContext(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(contextKey{}).(Claims)
	if !ok {
		return Claims{}, false
	}
	return claims.Normalized(), true
}

func (c Claims) Normalized() Claims {
	c.UserID = strings.TrimSpace(c.UserID)
	c.SessionID = strings.TrimSpace(c.SessionID)
	c.SellerID = strings.TrimSpace(c.SellerID)
	c.TenantID = strings.TrimSpace(c.TenantID)
	c.TokenID = strings.TrimSpace(c.TokenID)
	c.Roles = NormalizeRoles(c.Roles)
	return c
}

func (c Claims) Authenticated() bool {
	return strings.TrimSpace(c.UserID) != ""
}

func (c Claims) HasRole(role string) bool {
	role = strings.TrimSpace(role)
	if role == "" {
		return false
	}
	for _, current := range c.Roles {
		if current == role {
			return true
		}
	}
	return false
}

func (c Claims) HasAnyRole(allowed map[string]struct{}) bool {
	for _, role := range c.Roles {
		if _, ok := allowed[role]; ok {
			return true
		}
	}
	return false
}

func (c Claims) HasAnyRoleList(allowed ...domain.Role) bool {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, role := range allowed {
		allowedSet[role.String()] = struct{}{}
	}
	return c.HasAnyRole(allowedSet)
}

func NormalizeRoles(roles []string) []string {
	seen := make(map[string]struct{}, len(roles))
	out := make([]string, 0, len(roles))
	for _, role := range roles {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}
		if _, exists := seen[role]; exists {
			continue
		}
		seen[role] = struct{}{}
		out = append(out, role)
	}
	return out
}
