package metrics

import (
	"context"
	"log/slog"
	"time"

	platformlog "github.com/parag/ecommerce/backend/shared/platform/logger"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type GRPC struct {
	ServerRequests *prometheus.CounterVec
	ServerDuration *prometheus.HistogramVec
	ClientRequests *prometheus.CounterVec
	ClientDuration *prometheus.HistogramVec
}

func NewGRPC(registry prometheus.Registerer, namespace, service string) *GRPC {
	metrics := &GRPC{
		ServerRequests: prometheus.NewCounterVec(prometheus.CounterOpts{Namespace: namespace, Subsystem: "grpc_server", Name: "requests_total", ConstLabels: prometheus.Labels{"service": service}}, []string{"grpc_method", "grpc_code"}),
		ServerDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{Namespace: namespace, Subsystem: "grpc_server", Name: "duration_seconds", ConstLabels: prometheus.Labels{"service": service}}, []string{"grpc_method"}),
		ClientRequests: prometheus.NewCounterVec(prometheus.CounterOpts{Namespace: namespace, Subsystem: "grpc_client", Name: "requests_total", ConstLabels: prometheus.Labels{"service": service}}, []string{"target_service", "grpc_method", "grpc_code"}),
		ClientDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{Namespace: namespace, Subsystem: "grpc_client", Name: "duration_seconds", ConstLabels: prometheus.Labels{"service": service}}, []string{"target_service", "grpc_method"}),
	}
	registry.MustRegister(metrics.ServerRequests, metrics.ServerDuration, metrics.ClientRequests, metrics.ClientDuration)
	return metrics
}

func (metrics *GRPC) UnaryServerInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, request any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		started := time.Now()
		response, err := handler(ctx, request)
		code := status.Code(err).String()
		metrics.ServerRequests.WithLabelValues(info.FullMethod, code).Inc()
		metrics.ServerDuration.WithLabelValues(info.FullMethod).Observe(time.Since(started).Seconds())
		span := trace.SpanFromContext(ctx)
		if err != nil {
			span.RecordError(err)
		}
		platformlog.FromContext(ctx, logger).LogAttrs(ctx, levelForError(err), "grpc.server.request", slog.String("grpc_method", info.FullMethod), slog.String("grpc_code", code), slog.Int64("duration_ms", time.Since(started).Milliseconds()))
		return response, err
	}
}

func (metrics *GRPC) UnaryClientInterceptor(target string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, request, reply any, connection *grpc.ClientConn, invoker grpc.UnaryInvoker, options ...grpc.CallOption) error {
		started := time.Now()
		err := invoker(ctx, method, request, reply, connection, options...)
		code := status.Code(err).String()
		metrics.ClientRequests.WithLabelValues(target, method, code).Inc()
		metrics.ClientDuration.WithLabelValues(target, method).Observe(time.Since(started).Seconds())
		return err
	}
}

func levelForError(err error) slog.Level {
	if err != nil {
		return slog.LevelError
	}
	return slog.LevelInfo
}
