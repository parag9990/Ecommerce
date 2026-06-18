package ratelimit

import (
	"testing"
	"time"
)

func TestMatchPoliciesReturnsGlobalAndRouteSpecificIPPolicies(t *testing.T) {
	policies := DefaultPolicies(600, time.Minute)

	matches := MatchPolicies(policies, "auth.login", "POST", "/api/v1/auth/login", DimensionIP)

	names := policyNames(matches)
	if !containsName(names, "global.ip") || !containsName(names, "auth.login.ip") {
		t.Fatalf("expected global and login policies, got %v", names)
	}
}

func TestMatchPoliciesReturnsTargetPolicyForOTP(t *testing.T) {
	policies := DefaultPolicies(600, time.Minute)

	matches := MatchPolicies(policies, "auth.otp_send", "POST", "/api/v1/auth/otp/send", DimensionTarget)

	if len(matches) != 1 || matches[0].Name != "auth.otp_send.target" {
		t.Fatalf("expected otp target policy, got %+v", matches)
	}
}

func TestValidatePoliciesRejectsInvalidPolicy(t *testing.T) {
	err := ValidatePolicies([]Policy{{
		Name:      "bad",
		Dimension: DimensionUser,
		Limit:     0,
		Window:    time.Minute,
	}})
	if err == nil {
		t.Fatal("expected invalid policy to fail validation")
	}
}

func policyNames(policies []Policy) []string {
	names := make([]string, 0, len(policies))
	for _, policy := range policies {
		names = append(names, policy.Name)
	}
	return names
}

func containsName(names []string, want string) bool {
	for _, name := range names {
		if name == want {
			return true
		}
	}
	return false
}
