package clients

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestGRPCDialerConnectsToReadyService(t *testing.T) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	go func() {
		_ = server.Serve(listener)
	}()
	t.Cleanup(func() {
		server.Stop()
		_ = listener.Close()
	})

	dialer := NewGRPCDialer(DialOptions{DialTimeout: time.Second}, nil,
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		}),
	)
	conn, err := dialer.Dial(context.Background(), newServiceDescriptor(
		DownstreamProduct,
		"passthrough:///bufnet",
		"ecommerce.product.v1.ProductService",
	))
	if err != nil {
		t.Fatalf("dial ready service: %v", err)
	}
	t.Cleanup(func() {
		_ = conn.Close()
	})
}

func TestGRPCDialerReturnsWhenDialTimeoutExpires(t *testing.T) {
	const timeout = 40 * time.Millisecond
	dialer := NewGRPCDialer(DialOptions{DialTimeout: timeout}, nil,
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		}),
	)

	started := time.Now()
	_, err := dialer.Dial(context.Background(), newServiceDescriptor(
		DownstreamProduct,
		"passthrough:///unreachable",
		"ecommerce.product.v1.ProductService",
	))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected dial deadline exceeded, got %v", err)
	}
	if elapsed := time.Since(started); elapsed < timeout || elapsed > 2*time.Second {
		t.Fatalf("dial returned outside expected timeout bounds: %s", elapsed)
	}
}

func TestGRPCDialerCanStartWithUnavailableService(t *testing.T) {
	dialer := NewGRPCDialer(DialOptions{
		DialTimeout:      40 * time.Millisecond,
		AllowUnavailable: true,
	}, nil, grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}))

	conn, err := dialer.Dial(context.Background(), newServiceDescriptor(
		DownstreamProduct,
		"passthrough:///unreachable",
		"ecommerce.product.v1.ProductService",
	))
	if err != nil {
		t.Fatalf("dial unavailable service in degraded mode: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
}
