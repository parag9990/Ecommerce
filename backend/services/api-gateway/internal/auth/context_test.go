package auth

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestClaimsContextRoundTripUsesRoleCopy(t *testing.T) {
	ctx := WithClaims(context.Background(), AccessClaims{
		SessionID: "sess_123",
		Roles:     []string{"buyer"},
		TokenType: AccessTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "user_123",
		},
	})

	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		t.Fatal("expected claims in context")
	}
	claims.Roles[0] = "admin"

	fresh, ok := ClaimsFromContext(ctx)
	if !ok {
		t.Fatal("expected fresh claims in context")
	}
	if fresh.Roles[0] != "buyer" {
		t.Fatalf("expected context roles to be immutable copy, got %q", fresh.Roles[0])
	}
}
