package observability

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestUnaryClientInterceptorPropagatesRequestIDAndTraceparent(t *testing.T) {
	provider := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	defer provider.Shutdown(context.Background())
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	cfg := DefaultConfig("api-gateway", "test")
	var captured metadata.MD
	interceptor := UnaryClientInterceptor(cfg, nil, "product")
	ctx := WithRequestID(context.Background(), "req_trace_123")

	err := interceptor(ctx, "/ecommerce.product.v1.ProductService/GetProduct", nil, nil, nil, func(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			t.Fatal("expected outgoing metadata")
		}
		captured = md
		return nil
	})
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}
	if got := captured.Get("x-request-id"); len(got) != 1 || got[0] != "req_trace_123" {
		t.Fatalf("x-request-id metadata = %v", got)
	}
	if got := captured.Get("traceparent"); len(got) != 1 || got[0] == "" {
		t.Fatalf("traceparent metadata missing: %v", got)
	}
}
