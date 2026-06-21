package grpcclient

import (
	"context"
	"testing"

	"github.com/parag/ecommerce/backend/shared/platform/authctx"
	platformlog "github.com/parag/ecommerce/backend/shared/platform/logger"
	"google.golang.org/grpc/metadata"
)

func TestMetadataRoundTrip(t *testing.T) {
	ctx := platformlog.WithRequestID(context.Background(), "req-1")
	ctx = authctx.WithClaims(ctx, authctx.Claims{UserID: "user-1", Roles: []string{"buyer"}})
	outgoing, ok := metadata.FromOutgoingContext(OutgoingContext(ctx))
	if !ok || outgoing.Get(headerRequestID)[0] != "req-1" {
		t.Fatalf("outgoing metadata = %#v", outgoing)
	}
	incoming := metadata.NewIncomingContext(context.Background(), outgoing)
	incoming = IncomingContext(incoming)
	claims, ok := authctx.ClaimsFrom(incoming)
	if !ok || claims.UserID != "user-1" || !claims.HasRole("buyer") {
		t.Fatalf("claims = %#v, %t", claims, ok)
	}
	if got := platformlog.RequestIDFromContext(incoming); got != "req-1" {
		t.Fatalf("request id = %q", got)
	}
}
