package grpctransport

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	recommendationv1 "github.com/example/ecommerce-platform/backend/proto-gen/go/ecommerce/recommendation/v1"
	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/usecase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestGetRecommendationsMapsDomainResult(t *testing.T) {
	now := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
	reader := &fakeRecommendationReader{
		out: usecase.GetRecommendationsOutput{
			Result: domain.RecommendationResult{
				RecommendationID: "reco_123",
				Type:             domain.RecommendationTypeTrending,
				StrategyID:       domain.StrategyTrendingRecentActivity,
				GeneratedAt:      now,
				Items: []domain.RecommendationItem{
					{ProductID: "prod_1", Score: 42, Reason: "trending_recent_activity"},
				},
			},
			CacheTTL: 5 * time.Minute,
		},
	}
	handler, err := NewHandler(reader, nil)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-request-id", "req_123", "x-session-id", "sess_123"))

	got, err := handler.GetRecommendations(ctx, &recommendationv1.GetRecommendationsRequest{
		Context: recommendationv1.RecommendationContext_RECOMMENDATION_CONTEXT_HOME_FEED,
		Limit:   12,
	})
	if err != nil {
		t.Fatalf("GetRecommendations() error = %v", err)
	}
	if reader.input.RequestID != "req_123" {
		t.Fatalf("request id = %q, want req_123", reader.input.RequestID)
	}
	if reader.input.SessionID != "sess_123" {
		t.Fatalf("session id = %q, want sess_123", reader.input.SessionID)
	}
	if got.GetRecommendationId() != "reco_123" ||
		got.GetType() != recommendationv1.RecommendationType_RECOMMENDATION_TYPE_TRENDING ||
		got.GetStrategyId() != string(domain.StrategyTrendingRecentActivity) ||
		got.GetCacheTtlSeconds() != 300 {
		t.Fatalf("response = %+v, want mapped domain result", got)
	}
	if len(got.GetItems()) != 1 || got.GetItems()[0].GetProductId() != "prod_1" || got.GetItems()[0].GetRank() != 1 {
		t.Fatalf("items = %+v, want ranked product", got.GetItems())
	}
}

func TestGetRecommendationsRejectsNilRequest(t *testing.T) {
	handler, err := NewHandler(&fakeRecommendationReader{}, nil)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	_, err = handler.GetRecommendations(context.Background(), nil)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("code = %s, want InvalidArgument", status.Code(err))
	}
}

func TestGetRecommendationsSanitizesStorageErrors(t *testing.T) {
	handler, err := NewHandler(&fakeRecommendationReader{err: domain.ErrRecommendationStorage}, nil)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	_, err = handler.GetRecommendations(context.Background(), &recommendationv1.GetRecommendationsRequest{
		Context: recommendationv1.RecommendationContext_RECOMMENDATION_CONTEXT_HOME_FEED,
	})
	if status.Code(err) != codes.Unavailable {
		t.Fatalf("code = %s, want Unavailable", status.Code(err))
	}
	if got := status.Convert(err).Message(); got != "recommendations temporarily unavailable" {
		t.Fatalf("message = %q, want sanitized unavailable message", got)
	}
}

func TestRecommendationServiceBufconnContract(t *testing.T) {
	now := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
	reader := &fakeRecommendationReader{
		out: usecase.GetRecommendationsOutput{
			Result: domain.RecommendationResult{
				RecommendationID: "reco_bufconn",
				Type:             domain.RecommendationTypePersonalized,
				StrategyID:       domain.StrategyPersonalizedBehavior,
				GeneratedAt:      now,
				Items:            []domain.RecommendationItem{{ProductID: "prod_1", Score: 1}},
			},
		},
	}
	handler, err := NewHandler(reader, nil)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	listener := bufconn.Listen(1024 * 1024)
	server := NewServer(ServerConfig{DefaultTimeout: time.Second}, nil, nil)
	recommendationv1.RegisterRecommendationServiceServer(server, handler)
	go func() {
		_ = server.Serve(listener)
	}()
	defer server.Stop()

	conn, err := grpc.NewClient(
		"passthrough:///bufconn",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("grpc.NewClient() error = %v", err)
	}
	defer conn.Close()

	client := recommendationv1.NewRecommendationServiceClient(conn)
	got, err := client.GetRecommendations(context.Background(), &recommendationv1.GetRecommendationsRequest{
		UserId:  "user_1",
		Context: recommendationv1.RecommendationContext_RECOMMENDATION_CONTEXT_HOME_FEED,
		Limit:   12,
	})
	if err != nil {
		t.Fatalf("client.GetRecommendations() error = %v", err)
	}
	if got.GetRecommendationId() != "reco_bufconn" || len(got.GetItems()) != 1 {
		t.Fatalf("response = %+v, want bufconn response", got)
	}
}

type fakeRecommendationReader struct {
	input usecase.GetRecommendationsInput
	out   usecase.GetRecommendationsOutput
	err   error
}

func (f *fakeRecommendationReader) GetRecommendations(ctx context.Context, input usecase.GetRecommendationsInput) (usecase.GetRecommendationsOutput, error) {
	f.input = input
	if f.err != nil {
		return usecase.GetRecommendationsOutput{}, f.err
	}
	if f.out.Result.RecommendationID == "" {
		return usecase.GetRecommendationsOutput{}, errors.New("missing fake output")
	}
	return f.out, nil
}
