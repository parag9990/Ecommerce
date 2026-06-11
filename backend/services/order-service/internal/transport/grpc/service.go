package ordergrpc

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/order-service/internal/authctx"
	"github.com/example/ecommerce-platform/backend/services/order-service/internal/usecase"
	orderv1 "github.com/example/ecommerce-platform/backend/shared/gen/go/ecommerce/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RuntimeConfig struct {
	TrustedCallerToken string
}

func NewGRPCServer(handler *Server, ids usecase.IDGenerator, config RuntimeConfig, logger *slog.Logger) (*grpc.Server, error) {
	if handler == nil {
		return nil, errors.New("order gRPC handler is required")
	}
	if ids == nil {
		return nil, errors.New("request id generator is required")
	}
	if len(strings.TrimSpace(config.TrustedCallerToken)) < 32 {
		return nil, errors.New("trusted caller token must be at least 32 characters")
	}
	if logger == nil {
		logger = slog.Default()
	}
	server := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recoveryInterceptor(logger),
		authInterceptor(config.TrustedCallerToken, ids),
		loggingInterceptor(logger),
	))
	orderv1.RegisterOrderServiceServer(server, handler)
	return server, nil
}

func Serve(ctx context.Context, listener net.Listener, server *grpc.Server) error {
	if listener == nil || server == nil {
		return errors.New("listener and gRPC server are required")
	}
	stopped := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			server.GracefulStop()
		case <-stopped:
		}
	}()
	err := server.Serve(listener)
	close(stopped)
	if errors.Is(err, grpc.ErrServerStopped) {
		return nil
	}
	return err
}

func ListenAndServe(ctx context.Context, address string, server *grpc.Server) error {
	address = strings.TrimSpace(address)
	if address == "" {
		return errors.New("gRPC listen address is required")
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer listener.Close()
	return Serve(ctx, listener, server)
}

func authInterceptor(trustedToken string, ids usecase.IDGenerator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, request any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		requestID, err := ids.NewID("req")
		if err != nil {
			return nil, status.Error(codes.Internal, "internal order service error")
		}
		authenticatedCtx, err := authctx.AuthenticateIncoming(ctx, trustedToken, requestID)
		if err != nil {
			return nil, toStatusError(err)
		}
		return handler(authenticatedCtx, request)
	}
}

func recoveryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, request any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (response any, err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("order.grpc.panic_recovered", slog.String("method", info.FullMethod))
				err = status.Error(codes.Internal, "internal order service error")
			}
		}()
		return handler(ctx, request)
	}
}

func loggingInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, request any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		startedAt := time.Now()
		response, err := handler(ctx, request)
		actor, _ := authctx.ActorFromContext(ctx)
		logger.Info("order.grpc.request",
			slog.String("method", info.FullMethod),
			slog.String("request_id", actor.RequestID),
			slog.String("actor_id", actor.UserID),
			slog.String("code", status.Code(err).String()),
			slog.Duration("duration", time.Since(startedAt)),
		)
		return response, err
	}
}
