package observability

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type MetadataCarrier struct {
	MD metadata.MD
}

func (c MetadataCarrier) Get(key string) string {
	values := c.MD.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (c MetadataCarrier) Set(key, value string) {
	if key == "" || value == "" {
		return
	}
	c.MD.Set(key, value)
}

func (c MetadataCarrier) Keys() []string {
	keys := make([]string, 0, len(c.MD))
	for key := range c.MD {
		keys = append(keys, key)
	}
	return keys
}

func UnaryClientInterceptor(cfg Config, metrics *Metrics, targetService string) grpc.UnaryClientInterceptor {
	cfg = cfg.Normalize(cfg.ServiceName, cfg.Environment)
	tracer := otel.Tracer(cfg.ServiceName)
	return func(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		started := time.Now()
		ctx, span := tracer.Start(ctx, "grpc "+method, trace.WithSpanKind(trace.SpanKindClient))
		defer span.End()

		ctx = injectOutgoingMetadata(ctx)
		err := invoker(ctx, method, req, reply, cc, opts...)
		grpcCode := status.Code(err).String()
		metrics.ObserveGRPCClient(targetService, method, grpcCode, started)

		span.SetAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.method", method),
			attribute.String("rpc.grpc.status_code", grpcCode),
			attribute.String("downstream_service", targetService),
			attribute.String("request_id", RequestIDFromContext(ctx)),
		)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, grpcCode)
			metrics.ObserveDownstreamError(targetService, grpcCode)
		}
		return err
	}
}

func StreamClientInterceptor(cfg Config, metrics *Metrics, targetService string) grpc.StreamClientInterceptor {
	cfg = cfg.Normalize(cfg.ServiceName, cfg.Environment)
	tracer := otel.Tracer(cfg.ServiceName)
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		started := time.Now()
		ctx, span := tracer.Start(ctx, "grpc "+method, trace.WithSpanKind(trace.SpanKindClient))
		defer span.End()

		ctx = injectOutgoingMetadata(ctx)
		stream, err := streamer(ctx, desc, cc, method, opts...)
		grpcCode := status.Code(err).String()
		metrics.ObserveGRPCClient(targetService, method, grpcCode, started)

		span.SetAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.method", method),
			attribute.String("rpc.grpc.status_code", grpcCode),
			attribute.String("downstream_service", targetService),
			attribute.String("request_id", RequestIDFromContext(ctx)),
		)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, grpcCode)
			metrics.ObserveDownstreamError(targetService, grpcCode)
		}
		return stream, err
	}
}

func injectOutgoingMetadata(ctx context.Context) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		md = md.Copy()
	} else {
		md = metadata.MD{}
	}
	if requestID := RequestIDFromContext(ctx); requestID != "" {
		md.Set("x-request-id", requestID)
	}
	otel.GetTextMapPropagator().Inject(ctx, MetadataCarrier{MD: md})
	return metadata.NewOutgoingContext(ctx, md)
}
