package authctx

import (
	"context"
	"testing"
)

func TestClaimsRoundTripAndRoleMatching(t *testing.T) {
	ctx := WithClaims(context.Background(), Claims{UserID: " user-1 ", Roles: []string{" buyer ", ""}})
	claims, ok := ClaimsFrom(ctx)
	if !ok || claims.UserID != "user-1" {
		t.Fatalf("ClaimsFrom() = %#v, %t", claims, ok)
	}
	if !claims.HasRole("BUYER") {
		t.Fatal("expected case-insensitive buyer role")
	}
}
