package observability

import "testing"

func TestSensitiveFieldsAreRedacted(t *testing.T) {
	if got := RedactField("Authorization", "Bearer secret"); got != redactedValue {
		t.Fatalf("authorization field was not redacted: %v", got)
	}
	if got := RedactField("password", "secret"); got != redactedValue {
		t.Fatalf("password field was not redacted: %v", got)
	}
	if got := RedactField("route", "/api/v1/products"); got != "/api/v1/products" {
		t.Fatalf("safe field was unexpectedly redacted: %v", got)
	}
}
