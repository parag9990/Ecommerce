package clients

import (
	"context"
	"testing"

	gatewayauth "ecommerce/api-gateway/internal/auth"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

func TestWithAuthMetadataPropagatesSafeClaims(t *testing.T) {
	ctx := gatewayauth.WithClaims(context.Background(), gatewayauth.AccessClaims{
		SessionID: "sess_123",
		Roles:     []string{"seller", "seller_manager"},
		SellerID:  "seller_123",
		TokenType: gatewayauth.AccessTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "user_123",
		},
	})

	ctx = WithAuthMetadata(ctx)
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("expected outgoing metadata")
	}
	assertMetadataValue(t, md, "x-user-id", "user_123")
	assertMetadataValue(t, md, "x-session-id", "sess_123")
	assertMetadataValue(t, md, "x-roles", "seller,seller_manager")
	assertMetadataValue(t, md, "x-seller-id", "seller_123")
}

func TestWithMetadataReplacesTrustedValueAndPreservesOtherMetadata(t *testing.T) {
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(
		"x-user-id", "untrusted-user",
		"authorization", "Bearer internal-token",
	))

	ctx = WithMetadata(ctx, map[string]string{"x-user-id": "trusted-user"})
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("expected outgoing metadata")
	}
	assertMetadataValue(t, md, "x-user-id", "trusted-user")
	assertMetadataValue(t, md, "authorization", "Bearer internal-token")
}

func assertMetadataValue(t *testing.T, md metadata.MD, key string, want string) {
	t.Helper()
	values := md.Get(key)
	if len(values) != 1 || values[0] != want {
		t.Fatalf("expected %s=%q, got %v", key, want, values)
	}
}
