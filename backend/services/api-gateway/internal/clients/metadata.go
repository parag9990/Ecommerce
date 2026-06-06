package clients

import (
	"context"
	"strings"

	gatewayauth "ecommerce/api-gateway/internal/auth"
	"google.golang.org/grpc/metadata"
)

var propagatedMetadataKeys = map[string]struct{}{
	"x-request-id":      {},
	"x-correlation-id":  {},
	"traceparent":       {},
	"tracestate":        {},
	"x-b3-traceid":      {},
	"x-b3-spanid":       {},
	"x-b3-parentspanid": {},
	"x-b3-sampled":      {},
	"x-b3-flags":        {},
	"x-user-id":         {},
	"x-session-id":      {},
	"x-roles":           {},
	"x-seller-id":       {},
}

func WithRequestMetadata(ctx context.Context, requestID string) context.Context {
	return WithMetadata(ctx, map[string]string{"x-request-id": requestID})
}

func WithAuthMetadata(ctx context.Context) context.Context {
	claims, ok := gatewayauth.ClaimsFromContext(ctx)
	if !ok {
		return ctx
	}
	values := map[string]string{
		"x-user-id":    claims.UserID(),
		"x-session-id": claims.SessionID,
		"x-roles":      strings.Join(claims.Roles, ","),
	}
	if claims.SellerID != "" {
		values["x-seller-id"] = claims.SellerID
	}
	return WithMetadata(ctx, values)
}

func WithMetadata(ctx context.Context, values map[string]string) context.Context {
	if len(values) == 0 {
		return ctx
	}
	outgoing, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		outgoing = outgoing.Copy()
	} else {
		outgoing = metadata.MD{}
	}
	changed := false
	for key, value := range values {
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		if _, allowed := propagatedMetadataKeys[key]; !allowed {
			continue
		}
		outgoing.Set(key, value)
		changed = true
	}
	if !changed {
		return ctx
	}
	return metadata.NewOutgoingContext(ctx, outgoing)
}
