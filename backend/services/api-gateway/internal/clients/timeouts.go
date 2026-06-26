package clients

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"
)

func TimeoutFor(service Downstream) time.Duration {
	switch service {
	case DownstreamSession:
		return 200 * time.Millisecond
	case DownstreamAuth, DownstreamSearch:
		return 300 * time.Millisecond
	case DownstreamRecommendation:
		return 800 * time.Millisecond
	case DownstreamProduct:
		return 500 * time.Millisecond
	case DownstreamOrder, DownstreamPayment:
		return 1500 * time.Millisecond
	case DownstreamCMS, DownstreamSuperadmin:
		return time.Second
	default:
		return 700 * time.Millisecond
	}
}

func WithDeadline(ctx context.Context, service Downstream) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, TimeoutFor(service))
}

// UnaryDeadlineInterceptor enforces the service timeout unless the caller has a sooner deadline.
func UnaryDeadlineInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req any,
		reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		callCtx, cancel := withDefaultDeadline(ctx, timeout)
		defer cancel()
		return invoker(callCtx, method, req, reply, cc, opts...)
	}
}

// StreamDeadlineInterceptor keeps the bounded deadline alive until the stream completes.
func StreamDeadlineInterceptor(timeout time.Duration) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		callCtx, cancel := withDefaultDeadline(ctx, timeout)
		stream, err := streamer(callCtx, desc, cc, method, opts...)
		if err != nil {
			cancel()
			return nil, err
		}
		return &deadlineClientStream{
			ClientStream: stream,
			cancel:       cancel,
		}, nil
	}
}

func withDefaultDeadline(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = TimeoutFor("")
	}
	if deadline, exists := ctx.Deadline(); exists && time.Until(deadline) <= timeout {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, timeout)
}

type deadlineClientStream struct {
	grpc.ClientStream
	cancel context.CancelFunc
	once   sync.Once
}

func (s *deadlineClientStream) RecvMsg(message any) error {
	err := s.ClientStream.RecvMsg(message)
	if err != nil {
		s.finish()
	}
	return err
}

func (s *deadlineClientStream) finish() {
	s.once.Do(s.cancel)
}
