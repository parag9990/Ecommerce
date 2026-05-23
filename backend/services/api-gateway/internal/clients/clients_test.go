package clients

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	"ecommerce/api-gateway/internal/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

func TestNewWithDialerInitializesAllServiceClients(t *testing.T) {
	dialer := &stubDialer{}
	registry, err := NewWithDialer(context.Background(), ServiceDescriptorsFromConfig(testClientConfig(t)), dialer, discardLogger())
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	defer registry.Close()

	for _, service := range serviceOrder {
		client, ok := registry.Client(service)
		if !ok {
			t.Fatalf("expected %s client to be initialized", service)
		}
		if client.Descriptor().Name != service {
			t.Fatalf("expected descriptor service %s, got %s", service, client.Descriptor().Name)
		}
		if _, ok := registry.Connection(service); !ok {
			t.Fatalf("expected %s connection to be stored", service)
		}
	}
	if registry.Auth == nil || registry.Product == nil || registry.Superadmin == nil {
		t.Fatal("expected typed service client fields to be populated")
	}
	if err := registry.Close(); err != nil {
		t.Fatalf("second close should be idempotent: %v", err)
	}
}

func TestNewWithDialerClosesPartialRegistryOnFailure(t *testing.T) {
	dialer := &stubDialer{fail: DownstreamProduct}

	_, err := NewWithDialer(context.Background(), ServiceDescriptorsFromConfig(testClientConfig(t)), dialer, nil)
	if err == nil {
		t.Fatal("expected registry initialization to fail")
	}
	if !strings.Contains(err.Error(), "dial product service") {
		t.Fatalf("expected service-specific dial error, got %v", err)
	}
}

func TestWithMetadataPropagatesAllowedHeadersOnly(t *testing.T) {
	ctx := WithMetadata(context.Background(), map[string]string{
		"X-Request-Id":  " req-123 ",
		"Traceparent":   "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-00",
		"Authorization": "Bearer secret",
	})

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("expected outgoing metadata")
	}
	if got := md.Get("x-request-id"); len(got) != 1 || got[0] != "req-123" {
		t.Fatalf("expected request id metadata, got %v", got)
	}
	if got := md.Get("traceparent"); len(got) != 1 {
		t.Fatalf("expected traceparent metadata, got %v", got)
	}
	if got := md.Get("authorization"); len(got) != 0 {
		t.Fatalf("authorization must not be propagated, got %v", got)
	}
}

func TestTimeoutForUsesServiceStrategy(t *testing.T) {
	tests := map[Downstream]time.Duration{
		DownstreamSearch:  300 * time.Millisecond,
		DownstreamProduct: 500 * time.Millisecond,
		DownstreamOrder:   1500 * time.Millisecond,
		DownstreamCart:    700 * time.Millisecond,
	}
	for service, want := range tests {
		if got := TimeoutFor(service); got != want {
			t.Fatalf("expected %s timeout %s, got %s", service, want, got)
		}
	}

	ctx, cancel := WithDeadline(context.Background(), DownstreamSearch)
	defer cancel()
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected deadline to be set")
	}
	if remaining := time.Until(deadline); remaining <= 0 || remaining > TimeoutFor(DownstreamSearch) {
		t.Fatalf("deadline outside expected bounds: %s", remaining)
	}
}

func TestHealthCheckReportsServingAndNotServing(t *testing.T) {
	conn, cleanup := newHealthConn(t, map[string]healthv1.HealthCheckResponse_ServingStatus{
		"ecommerce.auth.v1.AuthService":       healthv1.HealthCheckResponse_SERVING,
		"ecommerce.product.v1.ProductService": healthv1.HealthCheckResponse_NOT_SERVING,
	})
	defer cleanup()

	registry := &Clients{
		conns: map[Downstream]*grpc.ClientConn{
			DownstreamAuth:    conn,
			DownstreamProduct: conn,
		},
		descriptors: map[Downstream]ServiceDescriptor{
			DownstreamAuth:    newServiceDescriptor(DownstreamAuth, "bufnet", "ecommerce.auth.v1.AuthService"),
			DownstreamProduct: newServiceDescriptor(DownstreamProduct, "bufnet", "ecommerce.product.v1.ProductService"),
		},
	}

	report := registry.Check(context.Background())
	if report[DownstreamAuth].Status != HealthStatusServing {
		t.Fatalf("expected auth serving, got %+v", report[DownstreamAuth])
	}
	if report[DownstreamProduct].Status != HealthStatusNotServing {
		t.Fatalf("expected product not serving, got %+v", report[DownstreamProduct])
	}
	if report.Ready() {
		t.Fatal("partial unhealthy registry should not be ready")
	}
}

type stubDialer struct {
	fail Downstream
}

func (d *stubDialer) Dial(_ context.Context, descriptor ServiceDescriptor) (*grpc.ClientConn, error) {
	if descriptor.Name == d.fail {
		return nil, errors.New("dial failed")
	}
	return grpc.NewClient(
		"passthrough:///"+string(descriptor.Name),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

func testClientConfig(t *testing.T) config.Config {
	t.Helper()
	return config.Config{
		ServiceName:          "api-gateway",
		Environment:          "test",
		HTTPAddress:          ":0",
		APIBasePath:          "/api/v1",
		APIContractPath:      "testdata/master-api.json",
		ReadHeaderTimeout:    time.Second,
		ShutdownTimeout:      time.Second,
		GRPCDialTimeout:      time.Second,
		AuthGRPCAddr:         "auth-service:9090",
		UserGRPCAddr:         "user-service:9090",
		ProductGRPCAddr:      "product-service:9090",
		CartGRPCAddr:         "cart-service:9090",
		WishlistGRPCAddr:     "wishlist-service:9090",
		OrderGRPCAddr:        "order-service:9090",
		PaymentGRPCAddr:      "payment-service:9090",
		SearchGRPCAddr:       "search-service:9090",
		CMSGRPCAddr:          "cms-service:9090",
		SessionGRPCAddr:      "session-service:9090",
		NotificationGRPCAddr: "notification-service:9090",
		SuperadminGRPCAddr:   "superadmin-service:9090",
	}
}

func newHealthConn(t *testing.T, statuses map[string]healthv1.HealthCheckResponse_ServingStatus) (*grpc.ClientConn, func()) {
	t.Helper()
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	healthServer := health.NewServer()
	for service, status := range statuses {
		healthServer.SetServingStatus(service, status)
	}
	healthv1.RegisterHealthServer(server, healthServer)
	go func() {
		_ = server.Serve(listener)
	}()

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

	return conn, func() {
		_ = conn.Close()
		server.Stop()
		_ = listener.Close()
	}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}
