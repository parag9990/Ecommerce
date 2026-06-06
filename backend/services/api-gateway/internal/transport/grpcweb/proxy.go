package grpcweb

import (
	"context"
	"errors"
	"io"
	"strings"

	gatewayauth "ecommerce/api-gateway/internal/auth"
	"ecommerce/api-gateway/internal/clients"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var forwardedRequestMetadata = map[string]struct{}{
	"authorization":     {},
	"x-request-id":      {},
	"x-correlation-id":  {},
	"traceparent":       {},
	"tracestate":        {},
	"x-b3-traceid":      {},
	"x-b3-spanid":       {},
	"x-b3-parentspanid": {},
	"x-b3-sampled":      {},
	"x-b3-flags":        {},
}

var forwardedResponseMetadata = map[string]struct{}{
	"x-correlation-id":        {},
	"traceparent":             {},
	"tracestate":              {},
	"grpc-status-details-bin": {},
}

type ConnectionProvider interface {
	Connection(service clients.Downstream) (*grpc.ClientConn, bool)
}

type transparentProxy struct {
	connections ConnectionProvider
	codec       rawCodec
}

func newTransparentProxy(connections ConnectionProvider) *transparentProxy {
	return &transparentProxy{connections: connections}
}

func (p *transparentProxy) Handler(_ any, serverStream grpc.ServerStream) error {
	method, ok := grpc.MethodFromServerStream(serverStream)
	if !ok {
		return status.Error(codes.Internal, "unable to identify grpc method")
	}
	policy, ok := policyFromContext(serverStream.Context())
	if !ok {
		return status.Error(codes.PermissionDenied, "method policy is missing")
	}
	connection, ok := p.connections.Connection(clients.Downstream(policy.Downstream))
	if !ok || connection == nil {
		return status.Error(codes.Unavailable, "service unavailable")
	}

	ctx, cancel := context.WithTimeout(serverStream.Context(), policy.Timeout)
	defer cancel()
	ctx = outgoingProxyContext(ctx)

	clientStream, err := connection.NewStream(
		ctx,
		&grpc.StreamDesc{ClientStreams: true, ServerStreams: true},
		method,
		grpc.ForceCodec(p.codec),
	)
	if err != nil {
		return sanitizeDownstreamError(err)
	}

	requestErr := make(chan error, 1)
	responseErr := make(chan error, 1)
	go func() {
		requestErr <- forwardRequests(serverStream, clientStream)
	}()
	go func() {
		responseErr <- forwardResponses(clientStream, serverStream)
	}()

	for completed := 0; completed < 2; completed++ {
		select {
		case err := <-requestErr:
			if err != nil {
				return sanitizeDownstreamError(err)
			}
		case err := <-responseErr:
			if err != nil {
				return sanitizeDownstreamError(err)
			}
		}
	}
	return nil
}

func outgoingProxyContext(ctx context.Context) context.Context {
	incoming, _ := metadata.FromIncomingContext(ctx)
	outgoing := filterMetadata(incoming, forwardedRequestMetadata)
	if _, authenticated := gatewayauth.ClaimsFromContext(ctx); !authenticated {
		outgoing.Delete("authorization")
	}
	ctx = metadata.NewOutgoingContext(ctx, outgoing)
	return clients.WithAuthMetadata(ctx)
}

func forwardRequests(serverStream grpc.ServerStream, clientStream grpc.ClientStream) error {
	for {
		message := &rawMessage{}
		err := serverStream.RecvMsg(message)
		if errors.Is(err, io.EOF) {
			return clientStream.CloseSend()
		}
		if err != nil {
			return err
		}
		if err := clientStream.SendMsg(message); err != nil {
			return err
		}
	}
}

func forwardResponses(clientStream grpc.ClientStream, serverStream grpc.ServerStream) error {
	headers, err := clientStream.Header()
	if err != nil {
		return err
	}
	if filtered := filterMetadata(headers, forwardedResponseMetadata); len(filtered) > 0 {
		if err := serverStream.SendHeader(filtered); err != nil {
			return err
		}
	}
	for {
		message := &rawMessage{}
		err := clientStream.RecvMsg(message)
		if errors.Is(err, io.EOF) {
			serverStream.SetTrailer(filterMetadata(clientStream.Trailer(), forwardedResponseMetadata))
			return nil
		}
		if err != nil {
			serverStream.SetTrailer(filterMetadata(clientStream.Trailer(), forwardedResponseMetadata))
			return err
		}
		if err := serverStream.SendMsg(message); err != nil {
			return err
		}
	}
}

func filterMetadata(source metadata.MD, allowed map[string]struct{}) metadata.MD {
	filtered := metadata.MD{}
	for key, values := range source {
		key = strings.ToLower(strings.TrimSpace(key))
		if _, ok := allowed[key]; !ok {
			continue
		}
		filtered[key] = append([]string(nil), values...)
	}
	return filtered
}

func sanitizeDownstreamError(err error) error {
	if err == nil {
		return nil
	}
	code := status.Code(err)
	switch code {
	case codes.Canceled:
		return status.Error(code, "request canceled")
	case codes.Unknown, codes.Internal, codes.DataLoss:
		return status.Error(codes.Internal, "internal service error")
	case codes.DeadlineExceeded:
		return status.Error(code, "request deadline exceeded")
	case codes.Unavailable:
		return status.Error(code, "service unavailable")
	case codes.Unauthenticated:
		return status.Error(code, "unauthenticated")
	case codes.PermissionDenied:
		return status.Error(code, "permission denied")
	case codes.ResourceExhausted:
		return status.Error(code, "resource exhausted")
	default:
		return err
	}
}
