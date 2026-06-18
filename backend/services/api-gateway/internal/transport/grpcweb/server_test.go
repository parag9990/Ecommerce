package grpcweb

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"ecommerce/api-gateway/internal/config"
	"ecommerce/api-gateway/internal/domain"

	"google.golang.org/grpc/test/bufconn"
)

func TestNewServerRejectsPolicyWithUnavailableDownstream(t *testing.T) {
	policy := domain.GRPCWebMethodPolicy{
		FullMethod: "/ecommerce.session.v1.SessionService/IngestEvent",
		Downstream: "missing",
		AuthMode:   domain.GRPCWebAuthPublic,
		Timeout:    time.Second,
	}
	listener := bufconn.Listen(1024 * 1024)
	defer listener.Close()

	_, err := NewServer(context.Background(), ServerOptions{
		Config: config.GRPCWebConfig{
			MaxReceiveMsgBytes: 1024,
			MaxSendMsgBytes:    1024,
		},
		ServiceName: "api-gateway-test",
		Policies:    staticPolicyCatalog{policy: policy},
		Connections: testConnections{},
		Logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
		Listener:    listener,
	})
	if err == nil || !strings.Contains(err.Error(), "unavailable downstream") {
		t.Fatalf("expected unavailable downstream error, got %v", err)
	}
}
