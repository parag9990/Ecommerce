package clients

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestUnaryDeadlineInterceptorBoundsCall(t *testing.T) {
	const timeout = 20 * time.Millisecond
	interceptor := UnaryDeadlineInterceptor(timeout)

	started := time.Now()
	err := interceptor(
		context.Background(),
		"/ecommerce.product.v1.ProductService/GetProduct",
		nil,
		nil,
		nil,
		func(ctx context.Context, _ string, _ any, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			<-ctx.Done()
			return ctx.Err()
		},
	)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
	if elapsed := time.Since(started); elapsed < timeout || elapsed > 10*timeout {
		t.Fatalf("call completed outside expected deadline bounds: %s", elapsed)
	}
}

func TestStreamDeadlineInterceptorBoundsStreamContext(t *testing.T) {
	const timeout = 20 * time.Millisecond
	interceptor := StreamDeadlineInterceptor(timeout)
	stream := &stubClientStream{}

	got, err := interceptor(
		context.Background(),
		&grpc.StreamDesc{ServerStreams: true},
		nil,
		"/ecommerce.search.v1.SearchService/StreamResults",
		func(ctx context.Context, _ *grpc.StreamDesc, _ *grpc.ClientConn, _ string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
			stream.ctx = ctx
			return stream, nil
		},
	)
	if err != nil {
		t.Fatalf("create stream: %v", err)
	}

	select {
	case <-got.Context().Done():
		if !errors.Is(got.Context().Err(), context.DeadlineExceeded) {
			t.Fatalf("expected deadline exceeded, got %v", got.Context().Err())
		}
	case <-time.After(10 * timeout):
		t.Fatal("stream context was not bounded by the service deadline")
	}
}

func TestStreamDeadlineInterceptorReleasesContextOnCompletion(t *testing.T) {
	interceptor := StreamDeadlineInterceptor(time.Second)
	stream := &stubClientStream{}

	got, err := interceptor(
		context.Background(),
		&grpc.StreamDesc{ServerStreams: true},
		nil,
		"/ecommerce.search.v1.SearchService/StreamResults",
		func(ctx context.Context, _ *grpc.StreamDesc, _ *grpc.ClientConn, _ string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
			stream.ctx = ctx
			return stream, nil
		},
	)
	if err != nil {
		t.Fatalf("create stream: %v", err)
	}
	if err := got.RecvMsg(nil); !errors.Is(err, io.EOF) {
		t.Fatalf("expected completed stream, got %v", err)
	}

	select {
	case <-got.Context().Done():
		if !errors.Is(got.Context().Err(), context.Canceled) {
			t.Fatalf("expected completed stream context to be canceled, got %v", got.Context().Err())
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("completed stream did not release its deadline context")
	}
}

func TestDeadlineInterceptorPreservesShorterParentDeadline(t *testing.T) {
	parent, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := UnaryDeadlineInterceptor(time.Second)(
		parent,
		"/ecommerce.auth.v1.AuthService/Verify",
		nil,
		nil,
		nil,
		func(ctx context.Context, _ string, _ any, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			<-ctx.Done()
			return ctx.Err()
		},
	)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected parent deadline exceeded, got %v", err)
	}
}

func TestDeadlineInterceptorBoundsExplicitLongerDeadline(t *testing.T) {
	parent, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	const serviceTimeout = 20 * time.Millisecond
	started := time.Now()

	err := UnaryDeadlineInterceptor(serviceTimeout)(
		parent,
		"/ecommerce.order.v1.OrderService/CreateOrder",
		nil,
		nil,
		nil,
		func(ctx context.Context, _ string, _ any, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			got, ok := ctx.Deadline()
			if !ok {
				t.Fatal("expected service deadline to be set")
			}
			remaining := time.Until(got)
			if remaining <= 0 || remaining > serviceTimeout {
				t.Fatalf("expected service deadline within %s, got %s", serviceTimeout, remaining)
			}
			<-ctx.Done()
			return ctx.Err()
		},
	)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected service deadline exceeded, got %v", err)
	}
	if elapsed := time.Since(started); elapsed < serviceTimeout || elapsed > 10*serviceTimeout {
		t.Fatalf("call completed outside expected deadline bounds: %s", elapsed)
	}
}

type stubClientStream struct {
	ctx context.Context
}

func (s *stubClientStream) Header() (metadata.MD, error) { return nil, nil }
func (s *stubClientStream) Trailer() metadata.MD         { return nil }
func (s *stubClientStream) CloseSend() error             { return nil }
func (s *stubClientStream) Context() context.Context     { return s.ctx }
func (s *stubClientStream) SendMsg(any) error            { return nil }
func (s *stubClientStream) RecvMsg(any) error            { return io.EOF }
