package grpcclient

import (
	"context"
	"strings"
	"time"

	"github.com/parag/ecommerce/backend/shared/platform/authctx"
	platformlog "github.com/parag/ecommerce/backend/shared/platform/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	headerRequestID = "x-request-id"
	headerUserID    = "x-user-id"
	headerSellerID  = "x-seller-id"
	headerSessionID = "x-session-id"
	headerRoles     = "x-roles"
)

func PropagationUnaryClientInterceptor(defaultTimeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, request, reply any, connection *grpc.ClientConn, invoker grpc.UnaryInvoker, options ...grpc.CallOption) error {
		if _, hasDeadline := ctx.Deadline(); !hasDeadline && defaultTimeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
			defer cancel()
		}
		ctx = OutgoingContext(ctx)
		return invoker(ctx, method, request, reply, connection, options...)
	}
}

func OutgoingContext(ctx context.Context) context.Context {
	md, _ := metadata.FromOutgoingContext(ctx)
	md = md.Copy()
	if requestID := platformlog.RequestIDFromContext(ctx); requestID != "" {
		md.Set(headerRequestID, requestID)
	}
	if claims, ok := authctx.ClaimsFrom(ctx); ok {
		md.Set(headerUserID, claims.UserID)
		setIfPresent(md, headerSellerID, claims.SellerID)
		setIfPresent(md, headerSessionID, claims.SessionID)
		if len(claims.Roles) > 0 {
			md.Set(headerRoles, strings.Join(claims.Roles, ","))
		}
	}
	otel.GetTextMapPropagator().Inject(ctx, metadataCarrier(md))
	return metadata.NewOutgoingContext(ctx, md)
}

func IncomingContext(ctx context.Context) context.Context {
	md, _ := metadata.FromIncomingContext(ctx)
	ctx = otel.GetTextMapPropagator().Extract(ctx, metadataCarrier(md))
	if values := md.Get(headerRequestID); len(values) > 0 {
		ctx = platformlog.WithRequestID(ctx, values[0])
	}
	if values := md.Get(headerUserID); len(values) > 0 && strings.TrimSpace(values[0]) != "" {
		claims := authctx.Claims{UserID: values[0]}
		if values := md.Get(headerSellerID); len(values) > 0 {
			claims.SellerID = values[0]
		}
		if values := md.Get(headerSessionID); len(values) > 0 {
			claims.SessionID = values[0]
		}
		if values := md.Get(headerRoles); len(values) > 0 {
			claims.Roles = strings.Split(values[0], ",")
		}
		ctx = authctx.WithClaims(ctx, claims)
	}
	return ctx
}

func PropagationUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, request any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return handler(IncomingContext(ctx), request)
	}
}

type metadataCarrier metadata.MD

func (carrier metadataCarrier) Get(key string) string {
	values := metadata.MD(carrier).Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (carrier metadataCarrier) Set(key, value string) { metadata.MD(carrier).Set(key, value) }

func (carrier metadataCarrier) Keys() []string {
	keys := make([]string, 0, len(carrier))
	for key := range carrier {
		keys = append(keys, key)
	}
	return keys
}

var _ propagation.TextMapCarrier = metadataCarrier{}

func setIfPresent(md metadata.MD, key, value string) {
	if value = strings.TrimSpace(value); value != "" {
		md.Set(key, value)
	}
}
