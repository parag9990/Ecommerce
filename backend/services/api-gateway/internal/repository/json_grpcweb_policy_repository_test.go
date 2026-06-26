package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ecommerce/api-gateway/internal/domain"
)

func TestJSONGRPCWebPolicyRepositoryLoadsPolicies(t *testing.T) {
	path := filepath.Join(t.TempDir(), "grpcweb-policies.json")
	data := []byte(`{"policies":[{"full_method":"/ecommerce.session.v1.SessionService/IngestEvent","downstream":"session","auth_mode":"optional","timeout":"800ms"}]}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write policy file: %v", err)
	}

	policies, err := NewJSONGRPCWebPolicyRepository(path).LoadPolicies(context.Background())
	if err != nil {
		t.Fatalf("load policies: %v", err)
	}
	if len(policies) != 1 {
		t.Fatalf("expected one policy, got %d", len(policies))
	}
	if policies[0].AuthMode != domain.GRPCWebAuthOptional || policies[0].Timeout != 800*time.Millisecond {
		t.Fatalf("unexpected policy: %+v", policies[0])
	}
}

func TestProjectGRPCWebPolicyFileLoads(t *testing.T) {
	path := filepath.Join("..", "..", "config", "grpcweb-policies.json")
	policies, err := NewJSONGRPCWebPolicyRepository(path).LoadPolicies(context.Background())
	if err != nil {
		t.Fatalf("load project grpc-web policies: %v", err)
	}
	if len(policies) != 7 {
		t.Fatalf("expected seven project grpc-web policies, got %d", len(policies))
	}
	for _, policy := range policies {
		if policy.FullMethod == "/ecommerce.recommendation.v1.RecommendationService/GetRecommendations" &&
			policy.Downstream == "recommendation" &&
			policy.AuthMode == domain.GRPCWebAuthOptional &&
			policy.Timeout == 800*time.Millisecond {
			return
		}
	}
	t.Fatal("expected recommendation grpc-web policy")
}

func TestJSONGRPCWebPolicyRepositoryRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "grpcweb-policies.json")
	data := []byte(`{"policies":[{"full_method":"/ecommerce.session.v1.SessionService/IngestEvent","downstream":"session","auth_mode":"optional","timeout":"800ms","typo":true}]}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write policy file: %v", err)
	}

	if _, err := NewJSONGRPCWebPolicyRepository(path).LoadPolicies(context.Background()); err == nil {
		t.Fatal("expected unknown policy field to be rejected")
	}
}
