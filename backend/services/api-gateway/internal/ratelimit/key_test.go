package ratelimit

import (
	"strings"
	"testing"
)

func TestBuildKeyHashesIdentity(t *testing.T) {
	key := BuildKey("rl:v1", DimensionTarget, "auth.otp_send.target", " User@Example.COM ")

	if !strings.HasPrefix(key, "rl:v1:target:auth.otp_send.target:") {
		t.Fatalf("unexpected key prefix: %s", key)
	}
	if strings.Contains(strings.ToLower(key), "user@example.com") {
		t.Fatalf("key must not contain raw target identity: %s", key)
	}
	if key != BuildKey("rl:v1", DimensionTarget, "auth.otp_send.target", "user@example.com") {
		t.Fatal("expected normalized identities to produce stable keys")
	}
}

func TestHashIdentityUsesUnknownBucketForEmptyValue(t *testing.T) {
	if HashIdentity("") != HashIdentity(" unknown ") {
		t.Fatal("expected empty identity to use stable unknown bucket")
	}
}
