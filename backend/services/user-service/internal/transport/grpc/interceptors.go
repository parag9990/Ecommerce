package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/parag/ecommerce/backend/services/user-service/internal/audit"
	grpcgo "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ServerOptions(logger *slog.Logger) []grpcgo.ServerOption {
	return []grpcgo.ServerOption{
		grpcgo.ChainUnaryInterceptor(
			UnaryRecoveryInterceptor(logger),
			UnaryAuditActorInterceptor(),
			UnaryLoggingInterceptor(logger),
		),
	}
}

func UnaryAuditActorInterceptor() grpcgo.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpcgo.UnaryServerInfo, handler grpcgo.UnaryHandler) (any, error) {
		if actor, ok := actorFromMetadata(ctx); ok {
			ctx = audit.WithActor(ctx, actor)
		}
		return handler(ctx, req)
	}
}

func UnaryLoggingInterceptor(logger *slog.Logger) grpcgo.UnaryServerInterceptor {
	if logger == nil {
		logger = slog.Default()
	}

	return func(ctx context.Context, req any, info *grpcgo.UnaryServerInfo, handler grpcgo.UnaryHandler) (any, error) {
		startedAt := time.Now()
		resp, err := handler(ctx, req)
		code := status.Code(err)

		logger.InfoContext(ctx, "grpc_request",
			slog.String("method", info.FullMethod),
			slog.String("code", code.String()),
			slog.Duration("duration", time.Since(startedAt)),
			slog.String("request_id", requestIDFromContext(ctx)),
		)
		return resp, err
	}
}

func UnaryRecoveryInterceptor(logger *slog.Logger) grpcgo.UnaryServerInterceptor {
	if logger == nil {
		logger = slog.Default()
	}

	return func(ctx context.Context, req any, info *grpcgo.UnaryServerInfo, handler grpcgo.UnaryHandler) (resp any, err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.ErrorContext(ctx, "grpc_panic_recovered",
					slog.String("method", info.FullMethod),
					slog.String("panic_type", fmt.Sprintf("%T", recovered)),
					slog.String("request_id", requestIDFromContext(ctx)),
				)
				err = status.Error(codes.Internal, "internal error")
			}
		}()

		return handler(ctx, req)
	}
}
