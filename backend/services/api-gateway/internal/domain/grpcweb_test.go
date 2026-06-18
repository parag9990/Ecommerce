package domain

import (
	"testing"
	"time"
)

func TestGRPCWebMethodPolicyValidate(t *testing.T) {
	policy := GRPCWebMethodPolicy{
		FullMethod: "/ecommerce.session.v1.SessionService/GetLiveSessions",
		Downstream: "session",
		AuthMode:   GRPCWebAuthRequired,
		Roles:      []string{"admin", "superadmin"},
		Timeout:    1500 * time.Millisecond,
	}

	if err := policy.Validate(); err != nil {
		t.Fatalf("validate policy: %v", err)
	}
	if policy.Service() != "ecommerce.session.v1.SessionService" {
		t.Fatalf("unexpected service %q", policy.Service())
	}
}

func TestGRPCWebMethodPolicyRejectsRolesWithoutRequiredAuth(t *testing.T) {
	policy := GRPCWebMethodPolicy{
		FullMethod: "/ecommerce.session.v1.SessionService/IngestEvent",
		Downstream: "session",
		AuthMode:   GRPCWebAuthOptional,
		Roles:      []string{"buyer"},
		Timeout:    time.Second,
	}

	if err := policy.Validate(); err == nil {
		t.Fatal("expected roles on optional auth policy to fail")
	}
}
