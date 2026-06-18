package grpcweb

import (
	"context"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	gatewayauth "ecommerce/api-gateway/internal/auth"
	"ecommerce/api-gateway/internal/clients"
	"ecommerce/api-gateway/internal/config"
	"ecommerce/api-gateway/internal/domain"
	"ecommerce/api-gateway/internal/observability"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestTransparentProxyForwardsAllowlistedMethodAndSafeMetadata(t *testing.T) {
	const method = "/ecommerce.session.v1.SessionService/IngestEvent"
	downstreamMetadata := make(chan metadata.MD, 1)
	downstreamConn, downstreamCleanup := startRawServer(t, func(_ any, stream grpc.ServerStream) error {
		md, _ := metadata.FromIncomingContext(stream.Context())
		downstreamMetadata <- md.Copy()
		message := &rawMessage{}
		if err := stream.RecvMsg(message); err != nil {
			return err
		}
		return stream.SendMsg(message)
	})
	defer downstreamCleanup()

	policy := domain.GRPCWebMethodPolicy{
		FullMethod: method,
		Downstream: string(clients.DownstreamSession),
		AuthMode:   domain.GRPCWebAuthPublic,
		Timeout:    time.Second,
	}
	gatewayConn, gatewayCleanup := startGatewayServer(t, staticPolicyCatalog{policy: policy}, testConnections{
		clients.DownstreamSession: downstreamConn,
	}, nil)
	defer gatewayCleanup()

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(
		"x-request-id", "req_proxy_test_123",
		"x-user-id", "spoofed-user",
		"authorization", "Bearer unverified-token",
	))
	stream, err := gatewayConn.NewStream(ctx, &grpc.StreamDesc{ClientStreams: true, ServerStreams: true}, method, grpc.ForceCodec(rawCodec{}))
	if err != nil {
		t.Fatalf("open gateway stream: %v", err)
	}
	request := &rawMessage{data: []byte{0x08, 0x01}}
	if err := stream.SendMsg(request); err != nil {
		t.Fatalf("send request: %v", err)
	}
	if err := stream.CloseSend(); err != nil {
		t.Fatalf("close request stream: %v", err)
	}
	response := &rawMessage{}
	if err := stream.RecvMsg(response); err != nil {
		t.Fatalf("receive response: %v", err)
	}
	if string(response.data) != string(request.data) {
		t.Fatalf("expected raw response %v, got %v", request.data, response.data)
	}
	if err := stream.RecvMsg(&rawMessage{}); err != io.EOF {
		t.Fatalf("expected response EOF, got %v", err)
	}
	headers, err := stream.Header()
	if err != nil {
		t.Fatalf("read response headers: %v", err)
	}
	if got := headers.Get("x-request-id"); len(got) != 1 || got[0] != "req_proxy_test_123" {
		t.Fatalf("expected request id response header, got %v", got)
	}

	select {
	case md := <-downstreamMetadata:
		if got := md.Get("x-request-id"); len(got) != 1 || got[0] != "req_proxy_test_123" {
			t.Fatalf("expected request id downstream, got %v", got)
		}
		if got := md.Get("x-user-id"); len(got) != 0 {
			t.Fatalf("spoofed identity metadata must be stripped, got %v", got)
		}
		if got := md.Get("authorization"); len(got) != 0 {
			t.Fatalf("unverified authorization metadata must be stripped, got %v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("downstream did not receive request")
	}
}

func TestTransparentProxyForwardsVerifiedAuthAndTrustedIdentity(t *testing.T) {
	const method = "/ecommerce.session.v1.SessionService/GetLiveSessions"
	downstreamMetadata := make(chan metadata.MD, 1)
	downstreamConn, downstreamCleanup := startRawServer(t, func(_ any, stream grpc.ServerStream) error {
		md, _ := metadata.FromIncomingContext(stream.Context())
		downstreamMetadata <- md.Copy()
		message := &rawMessage{}
		if err := stream.RecvMsg(message); err != nil {
			return err
		}
		return stream.SendMsg(message)
	})
	defer downstreamCleanup()

	policy := domain.GRPCWebMethodPolicy{
		FullMethod: method,
		Downstream: string(clients.DownstreamSession),
		AuthMode:   domain.GRPCWebAuthRequired,
		Roles:      []string{"admin"},
		Timeout:    time.Second,
	}
	verifier := stubTokenVerifier{claims: gatewayauth.AccessClaims{
		SessionID: "session_123",
		Roles:     []string{"admin"},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "user_123",
		},
	}}
	gatewayConn, gatewayCleanup := startGatewayServer(t, staticPolicyCatalog{policy: policy}, testConnections{
		clients.DownstreamSession: downstreamConn,
	}, verifier)
	defer gatewayCleanup()

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer valid-token"))
	stream, err := gatewayConn.NewStream(ctx, &grpc.StreamDesc{ClientStreams: true, ServerStreams: true}, method, grpc.ForceCodec(rawCodec{}))
	if err != nil {
		t.Fatalf("open gateway stream: %v", err)
	}
	if err := stream.SendMsg(&rawMessage{data: []byte{0x08, 0x01}}); err != nil {
		t.Fatalf("send request: %v", err)
	}
	if err := stream.CloseSend(); err != nil {
		t.Fatalf("close request stream: %v", err)
	}
	if err := stream.RecvMsg(&rawMessage{}); err != nil {
		t.Fatalf("receive response: %v", err)
	}

	select {
	case md := <-downstreamMetadata:
		assertSingleMetadata(t, md, "authorization", "Bearer valid-token")
		assertSingleMetadata(t, md, "x-user-id", "user_123")
		assertSingleMetadata(t, md, "x-session-id", "session_123")
		assertSingleMetadata(t, md, "x-roles", "admin")
	case <-time.After(time.Second):
		t.Fatal("downstream did not receive authenticated request")
	}
}

func TestTransparentProxyRejectsMethodOutsideAllowlist(t *testing.T) {
	downstreamConn, downstreamCleanup := startRawServer(t, func(any, grpc.ServerStream) error {
		return nil
	})
	defer downstreamCleanup()

	policy := domain.GRPCWebMethodPolicy{
		FullMethod: "/ecommerce.session.v1.SessionService/IngestEvent",
		Downstream: string(clients.DownstreamSession),
		AuthMode:   domain.GRPCWebAuthPublic,
		Timeout:    time.Second,
	}
	gatewayConn, cleanup := startGatewayServer(t, staticPolicyCatalog{policy: policy}, testConnections{
		clients.DownstreamSession: downstreamConn,
	}, nil)
	defer cleanup()

	stream, err := gatewayConn.NewStream(context.Background(), &grpc.StreamDesc{ClientStreams: true, ServerStreams: true}, "/ecommerce.internal.v1.InternalService/Call", grpc.ForceCodec(rawCodec{}))
	if err != nil {
		t.Fatalf("open gateway stream: %v", err)
	}
	_ = stream.CloseSend()
	err = stream.RecvMsg(&rawMessage{})
	if status.Code(err).String() != "Unimplemented" {
		t.Fatalf("expected Unimplemented, got %v", err)
	}
	headers, headerErr := stream.Header()
	if headerErr != nil {
		t.Fatalf("read rejected response headers: %v", headerErr)
	}
	if got := headers.Get("x-request-id"); len(got) != 1 || !observability.ValidRequestID(got[0]) {
		t.Fatalf("expected generated request id response header, got %v", got)
	}
}

func startGatewayServer(t *testing.T, policies staticPolicyCatalog, connections testConnections, verifier TokenVerifier) (*grpc.ClientConn, func()) {
	t.Helper()
	listener := bufconn.Listen(1024 * 1024)
	server, err := NewServer(context.Background(), ServerOptions{
		Config: config.GRPCWebConfig{
			Address:            ":0",
			MaxReceiveMsgBytes: 1024 * 1024,
			MaxSendMsgBytes:    1024 * 1024,
		},
		ServiceName:   "api-gateway-test",
		Policies:      policies,
		Connections:   connections,
		TokenVerifier: verifier,
		Logger:        slog.New(slog.NewTextHandler(io.Discard, nil)),
		Listener:      listener,
	})
	if err != nil {
		t.Fatalf("new gateway server: %v", err)
	}
	go func() {
		_ = server.Serve()
	}()

	conn := newBufConn(t, listener)
	return conn, func() {
		_ = conn.Close()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
		_ = listener.Close()
	}
}

func startRawServer(t *testing.T, handler grpc.StreamHandler) (*grpc.ClientConn, func()) {
	t.Helper()
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer(grpc.ForceServerCodec(rawCodec{}), grpc.UnknownServiceHandler(handler))
	go func() {
		_ = server.Serve(listener)
	}()
	conn := newBufConn(t, listener)
	return conn, func() {
		_ = conn.Close()
		server.Stop()
		_ = listener.Close()
	}
}

func newBufConn(t *testing.T, listener *bufconn.Listener) *grpc.ClientConn {
	t.Helper()
	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("new bufconn client: %v", err)
	}
	return conn
}

type staticPolicyCatalog struct {
	policy domain.GRPCWebMethodPolicy
}

func (c staticPolicyCatalog) Load(context.Context) error {
	return nil
}

func (c staticPolicyCatalog) FindByMethod(_ context.Context, fullMethod string) (domain.GRPCWebMethodPolicy, error) {
	if fullMethod != c.policy.FullMethod {
		return domain.GRPCWebMethodPolicy{}, status.Error(12, "not allowed")
	}
	return c.policy, nil
}

func (c staticPolicyCatalog) List(context.Context) ([]domain.GRPCWebMethodPolicy, error) {
	return []domain.GRPCWebMethodPolicy{c.policy}, nil
}

func (c staticPolicyCatalog) RequiresToken(context.Context) (bool, error) {
	return c.policy.RequiresToken(), nil
}

type testConnections map[clients.Downstream]*grpc.ClientConn

func (c testConnections) Connection(service clients.Downstream) (*grpc.ClientConn, bool) {
	connection, ok := c[service]
	return connection, ok
}

func assertSingleMetadata(t *testing.T, md metadata.MD, key, want string) {
	t.Helper()
	got := md.Get(key)
	if len(got) != 1 || got[0] != want {
		t.Fatalf("expected %s=%q, got %v", key, want, got)
	}
}
