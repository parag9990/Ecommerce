package grpcweb

import (
	"context"
	"errors"
	"testing"

	gatewayauth "ecommerce/api-gateway/internal/auth"
	"ecommerce/api-gateway/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAuthorizeRequiredMethodEnforcesRoles(t *testing.T) {
	policy := domain.GRPCWebMethodPolicy{
		AuthMode: domain.GRPCWebAuthRequired,
		Roles:    []string{"admin", "superadmin"},
	}
	verifier := stubTokenVerifier{claims: gatewayauth.AccessClaims{
		Roles: []string{"buyer"},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "user_123",
		},
	}}

	_, err := authorize(context.Background(), metadata.Pairs("authorization", "Bearer valid-token"), policy, verifier)
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("expected permission denied, got %v", err)
	}
}

func TestAuthorizeRequiredMethodInjectsVerifiedClaims(t *testing.T) {
	policy := domain.GRPCWebMethodPolicy{
		AuthMode: domain.GRPCWebAuthRequired,
		Roles:    []string{"admin"},
	}
	verifier := stubTokenVerifier{claims: gatewayauth.AccessClaims{
		Roles: []string{"admin"},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "user_123",
		},
	}}

	ctx, err := authorize(context.Background(), metadata.Pairs("authorization", "Bearer valid-token"), policy, verifier)
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	claims, ok := gatewayauth.ClaimsFromContext(ctx)
	if !ok || claims.UserID() != "user_123" {
		t.Fatalf("expected verified claims in context, got %+v", claims)
	}
}

func TestAuthorizeOptionalMethodRejectsInvalidProvidedToken(t *testing.T) {
	policy := domain.GRPCWebMethodPolicy{AuthMode: domain.GRPCWebAuthOptional}
	verifier := stubTokenVerifier{err: errors.New("invalid token")}

	_, err := authorize(context.Background(), metadata.Pairs("authorization", "Bearer invalid-token"), policy, verifier)
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
}

func TestAuthorizeRejectsMultipleAuthorizationValues(t *testing.T) {
	policy := domain.GRPCWebMethodPolicy{AuthMode: domain.GRPCWebAuthRequired}
	incoming := metadata.Pairs(
		"authorization", "Bearer first-token",
		"authorization", "Bearer second-token",
	)

	_, err := authorize(context.Background(), incoming, policy, stubTokenVerifier{})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
}

type stubTokenVerifier struct {
	claims gatewayauth.AccessClaims
	err    error
}

func (v stubTokenVerifier) Verify(context.Context, string) (gatewayauth.AccessClaims, error) {
	return v.claims, v.err
}
