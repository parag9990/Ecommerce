package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestJSONRouteRepositoryLoadContract(t *testing.T) {
	path := filepath.Join(t.TempDir(), "master-api.json")
	body := `{
		"project": "scalable-ecommerce-platform",
		"version": "1.0.0",
		"rest_endpoints": [
			{
				"id": "auth.login",
				"method": "POST",
				"path": "/api/v1/auth/login",
				"service": "auth-service",
				"grpc": "AuthService.Login",
				"auth": "public",
				"request_schema": "LoginRequest",
				"response_schema": "AuthSessionResponse"
			}
		]
	}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write contract: %v", err)
	}

	repo := NewJSONRouteRepository(path)
	contract, err := repo.LoadContract(context.Background())
	if err != nil {
		t.Fatalf("load contract: %v", err)
	}
	if got := len(contract.RestEndpoints); got != 1 {
		t.Fatalf("expected one route, got %d", got)
	}
	if got := contract.RestEndpoints[0].GRPCMethod; got != "AuthService.Login" {
		t.Fatalf("expected grpc mapping, got %q", got)
	}
}
