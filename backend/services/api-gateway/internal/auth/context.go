package auth

import "context"

type claimsContextKey struct{}

func WithClaims(ctx context.Context, claims AccessClaims) context.Context {
	claims.Roles = cloneRoles(claims.Roles)
	return context.WithValue(ctx, claimsContextKey{}, claims)
}

func ClaimsFromContext(ctx context.Context) (AccessClaims, bool) {
	claims, ok := ctx.Value(claimsContextKey{}).(AccessClaims)
	if !ok {
		return AccessClaims{}, false
	}
	claims.Roles = cloneRoles(claims.Roles)
	return claims, true
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	claims, ok := ClaimsFromContext(ctx)
	if !ok || claims.UserID() == "" {
		return "", false
	}
	return claims.UserID(), true
}
