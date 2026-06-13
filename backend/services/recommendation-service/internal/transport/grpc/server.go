package grpctransport

import (
	"context"
	"log/slog"
	"runtime/debug"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ServerConfig struct {
	MaxRecvBytes   int
	MaxSendBytes   int
	DefaultTimeout time.Duration
}

type ServerMetrics interface {
	RecordGRPCRequest(method string, code string)
	ObserveGRPCDuration(method string, duration time.Duration)
	IncGRPCInflight(method string)
	DecGRPCInflight(method string)
}

func NewServer(cfg ServerConfig, logger *slog.Logger, metrics ServerMetrics) *grpc.Server {
	if logger == nil {
		logger = slog.Default()
	}
	options := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			recoveryUnaryInterceptor(logger),
			defaultDeadlineUnaryInterceptor(cfg.DefaultTimeout),
			loggingUnaryInterceptor(logger, metrics),
		),
	}
	if cfg.MaxRecvBytes > 0 {
		options = append(options, grpc.MaxRecvMsgSize(cfg.MaxRecvBytes))
	}
	if cfg.MaxSendBytes > 0 {
		options = append(options, grpc.MaxSendMsgSize(cfg.MaxSendBytes))
	}
	return grpc.NewServer(options...)
}

func recoveryUnaryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.ErrorContext(ctx, "recommendation.grpc.panic_recovered",
					slog.String("method", methodName(info.FullMethod)),
					slog.Any("panic", recovered),
					slog.String("stack", string(debug.Stack())),
				)
				err = status.Error(codes.Internal, "internal recommendation error")
			}
		}()
		return handler(ctx, req)
	}
}

func defaultDeadlineUnaryInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if timeout <= 0 {
			return handler(ctx, req)
		}
		if _, ok := ctx.Deadline(); ok {
			return handler(ctx, req)
		}
		deadlineCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return handler(deadlineCtx, req)
	}
}

func loggingUnaryInterceptor(logger *slog.Logger, metrics ServerMetrics) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		startedAt := time.Now()
		method := methodName(info.FullMethod)
		if metrics != nil {
			metrics.IncGRPCInflight(method)
			defer metrics.DecGRPCInflight(method)
		}

		resp, err := handler(ctx, req)
		code := status.Code(err)
		duration := time.Since(startedAt)
		if metrics != nil {
			metrics.RecordGRPCRequest(method, code.String())
			metrics.ObserveGRPCDuration(method, duration)
		}
		level := slog.LevelInfo
		if code != codes.OK {
			level = slog.LevelWarn
		}
		logger.LogAttrs(ctx, level, "recommendation.grpc.request",
			slog.String("request_id", requestIDFromContext(ctx)),
			slog.String("method", method),
			slog.String("code", code.String()),
			slog.Int64("duration_ms", duration.Milliseconds()),
		)
		return resp, err
	}
}

func requestIDFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	for _, key := range []string{"x-request-id", "request-id", "traceparent"} {
		for _, value := range md.Get(key) {
			if value = strings.TrimSpace(value); value != "" {
				return value
			}
		}
	}
	return ""
}

func sessionIDFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	for _, key := range []string{"x-session-id", "session-id"} {
		for _, value := range md.Get(key) {
			if value = strings.TrimSpace(value); value != "" {
				return value
			}
		}
	}
	return ""
}

func methodName(fullMethod string) string {
	if fullMethod == "" {
		return ""
	}
	if index := strings.LastIndex(fullMethod, "/"); index >= 0 && index+1 < len(fullMethod) {
		return fullMethod[index+1:]
	}
	return fullMethod
}
