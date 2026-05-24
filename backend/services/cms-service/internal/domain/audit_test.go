package domain

import "testing"

func TestRedactAuditSnapshotRemovesSensitiveFields(t *testing.T) {
	snapshot := map[string]any{
		"status":        "active",
		"password":      "secret",
		"refresh_token": "token",
		"nested": map[string]any{
			"api_key": "key",
			"name":    "visible",
		},
	}

	redacted := RedactAuditSnapshot(snapshot)
	if redacted["status"] != "active" {
		t.Fatalf("expected non-sensitive value to remain, got %+v", redacted)
	}
	if redacted["password"] != "[REDACTED]" || redacted["refresh_token"] != "[REDACTED]" {
		t.Fatalf("expected sensitive top-level fields to be redacted, got %+v", redacted)
	}
	nested, ok := redacted["nested"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested map, got %T", redacted["nested"])
	}
	if nested["api_key"] != "[REDACTED]" || nested["name"] != "visible" {
		t.Fatalf("expected nested sensitive fields to be redacted, got %+v", nested)
	}
}

func TestAuditEntryFromEventNormalizesSellerAndDecision(t *testing.T) {
	entry := AuditEntryFromEvent(AuditEvent{
		AuditID:          " audit_1 ",
		ActorUserID:      " user_1 ",
		ActorSellerID:    "seller_actor",
		Action:           PermissionCouponsCreate,
		ResourceType:     " coupon ",
		ResourceID:       " coupon_1 ",
		ResourceSellerID: " seller_1 ",
		Decision:         "",
	})

	if entry.AuditID != "audit_1" || entry.SellerID != "seller_1" || entry.Decision != AuditDecisionAllowed {
		t.Fatalf("unexpected normalized entry: %+v", entry)
	}
}
