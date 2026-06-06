package usecase

import (
	"context"
	"testing"
	"time"

	"ecommerce/api-gateway/internal/domain"
)

func TestGRPCWebPolicyCatalogEnforcesServiceAndMethodAllowlist(t *testing.T) {
	catalog := NewGRPCWebPolicyCatalogService(stubGRPCWebPolicyRepository{policies: []domain.GRPCWebMethodPolicy{{
		FullMethod: "/ecommerce.session.v1.SessionService/GetLiveSessions",
		Downstream: "session",
		AuthMode:   domain.GRPCWebAuthRequired,
		Roles:      []string{"admin"},
		Timeout:    time.Second,
	}}}, []string{"ecommerce.session.v1.SessionService"})

	if err := catalog.Load(context.Background()); err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	if _, err := catalog.FindByMethod(context.Background(), "/ecommerce.session.v1.SessionService/GetLiveSessions"); err != nil {
		t.Fatalf("find allowed method: %v", err)
	}
	if _, err := catalog.FindByMethod(context.Background(), "/ecommerce.session.v1.SessionService/Other"); err == nil {
		t.Fatal("expected unlisted method to be rejected")
	}
}

func TestGRPCWebPolicyCatalogRejectsUnexposedService(t *testing.T) {
	catalog := NewGRPCWebPolicyCatalogService(stubGRPCWebPolicyRepository{policies: []domain.GRPCWebMethodPolicy{{
		FullMethod: "/ecommerce.internal.v1.InternalService/Call",
		Downstream: "session",
		AuthMode:   domain.GRPCWebAuthPublic,
		Timeout:    time.Second,
	}}}, []string{"ecommerce.session.v1.SessionService"})

	if err := catalog.Load(context.Background()); err == nil {
		t.Fatal("expected unexposed service to be rejected")
	}
}

type stubGRPCWebPolicyRepository struct {
	policies []domain.GRPCWebMethodPolicy
}

func (r stubGRPCWebPolicyRepository) LoadPolicies(context.Context) ([]domain.GRPCWebMethodPolicy, error) {
	return r.policies, nil
}
