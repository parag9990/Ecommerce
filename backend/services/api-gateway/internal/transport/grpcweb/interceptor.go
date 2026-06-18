package grpcweb

import (
	"context"
	"log/slog"
	"strings"
	"time"

	gatewayauth "ecommerce/api-gateway/internal/auth"
	"ecommerce/api-gateway/internal/domain"
	"ecommerce/api-gateway/internal/observability"
	"ecommerce/api-gateway/internal/usecase"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type policyContextKey struct{}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *wrappedServerStream) Context() context.Context {
	return s.ctx
}

func securityInterceptor(
	serviceName string,
	observabilityConfig observability.Config,
	policies usecase.GRPCWebPolicyCatalog,
	verifier TokenVerifier,
	logger *slog.Logger,
	metrics *observability.Metrics,
) grpc.StreamServerInterceptor {
	observabilityConfig = observabilityConfig.Normalize(serviceName, observabilityConfig.Environment)
	if logger == nil {
		logger = slog.Default()
	}
	tracer := otel.Tracer(serviceName)

	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		started := time.Now()
		ctx := stream.Context()
		incoming, _ := metadata.FromIncomingContext(ctx)
		ctx = otel.GetTextMapPropagator().Extract(ctx, observability.MetadataCarrier{MD: incoming})
		ctx, span := tracer.Start(ctx, "grpc "+info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		requestID := firstMetadata(incoming, "x-request-id")
		if !observability.ValidRequestID(requestID) {
			requestID = observability.NewRequestID()
		}
		ctx = observability.WithRequestID(ctx, requestID)
		_ = stream.SetHeader(metadata.Pairs("x-request-id", requestID))

		policy, err := policies.FindByMethod(ctx, info.FullMethod)
		if err != nil {
			err = status.Error(codes.Unimplemented, "method is not exposed")
		} else {
			ctx, err = authorize(ctx, incoming, policy, verifier)
		}
		if err == nil {
			ctx = context.WithValue(ctx, policyContextKey{}, policy)
			err = handler(srv, &wrappedServerStream{ServerStream: stream, ctx: ctx})
		}

		code := status.Code(err)
		switch code {
		case codes.Unauthenticated:
			metrics.ObserveAuthFailure(info.FullMethod, "jwt_verification_failed")
		case codes.PermissionDenied:
			metrics.ObserveAuthFailure(info.FullMethod, "rbac_denied")
		}
		metrics.ObserveGRPCServer(info.FullMethod, code.String(), started)
		span.SetAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.method", info.FullMethod),
			attribute.String("rpc.grpc.status_code", code.String()),
			attribute.String("request_id", requestID),
		)
		if policy.Downstream != "" {
			span.SetAttributes(attribute.String("downstream_service", policy.Downstream))
		}
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, code.String())
		}
		logGRPCWebRequest(logger, ctx, observabilityConfig, policy, info.FullMethod, code, time.Since(started))
		return err
	}
}

func authorize(ctx context.Context, incoming metadata.MD, policy domain.GRPCWebMethodPolicy, verifier TokenVerifier) (context.Context, error) {
	if policy.AuthMode == domain.GRPCWebAuthPublic {
		return ctx, nil
	}
	if len(incoming.Get("authorization")) > 1 {
		return ctx, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	authorization := firstMetadata(incoming, "authorization")
	token := gatewayauth.BearerToken(authorization)
	if token == "" {
		if policy.AuthMode == domain.GRPCWebAuthOptional && strings.TrimSpace(authorization) == "" {
			return ctx, nil
		}
		return ctx, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	if verifier == nil {
		return ctx, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	claims, err := verifier.Verify(ctx, token)
	if err != nil {
		return ctx, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	if len(policy.Roles) > 0 && !claims.HasAnyRole(policy.Roles) {
		return ctx, status.Error(codes.PermissionDenied, "permission denied")
	}
	return gatewayauth.WithClaims(ctx, claims), nil
}

func logGRPCWebRequest(logger *slog.Logger, ctx context.Context, cfg observability.Config, policy domain.GRPCWebMethodPolicy, method string, code codes.Code, latency time.Duration) {
	attrs := []any{
		"service", cfg.ServiceName,
		"environment", cfg.Environment,
		"request_id", observability.RequestIDFromContext(ctx),
		"grpc_method", method,
		"grpc_code", code.String(),
		"latency_ms", latency.Milliseconds(),
	}
	if policy.Downstream != "" {
		attrs = append(attrs, "downstream_service", policy.Downstream)
	}
	if claims, ok := gatewayauth.ClaimsFromContext(ctx); ok {
		attrs = append(attrs, "user_id_hash", observability.HashUserID(claims.UserID(), cfg.UserHashSalt))
	}
	switch code {
	case codes.OK:
		logger.InfoContext(ctx, "grpc_web_request_completed", attrs...)
	case codes.Internal, codes.Unknown, codes.DataLoss, codes.Unavailable:
		logger.ErrorContext(ctx, "grpc_web_request_failed", attrs...)
	default:
		logger.WarnContext(ctx, "grpc_web_request_rejected", attrs...)
	}
}

func policyFromContext(ctx context.Context) (domain.GRPCWebMethodPolicy, bool) {
	policy, ok := ctx.Value(policyContextKey{}).(domain.GRPCWebMethodPolicy)
	return policy, ok
}

func firstMetadata(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[0])
}
