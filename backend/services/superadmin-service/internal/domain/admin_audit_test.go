package domain

import "testing"

func TestAuditRecordValidateRequiresAuditContext(t *testing.T) {
	record := AuditRecord{
		ActorAdminID: "admin_1",
		Action:       "seller.status.suspended",
		ResourceType: "seller",
		ResourceID:   "seller_1",
		RequestID:    "req_1",
		Reason:       "verified policy violation",
	}

	if err := record.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	record.RequestID = ""
	if err := record.Validate(); err == nil {
		t.Fatal("Validate returned nil, want missing request_id error")
	}
}

func TestSafeAuditSnapshotRemovesSensitiveKeys(t *testing.T) {
	got := SafeAuditSnapshot(map[string]string{
		"status":        "active",
		"access_token":  "secret",
		"Card-Number":   "4111111111111111",
		"update_reason": "manual review",
	})

	if got["status"] != "active" || got["update_reason"] != "manual review" {
		t.Fatalf("safe fields missing: %+v", got)
	}
	if _, ok := got["access_token"]; ok {
		t.Fatalf("access_token was not removed: %+v", got)
	}
	if _, ok := got["Card-Number"]; ok {
		t.Fatalf("Card-Number was not removed: %+v", got)
	}
}

func TestAuditLogCursorRoundTrip(t *testing.T) {
	cursor := NewAuditLogCursor(12345)
	if cursor == "" {
		t.Fatal("cursor is empty")
	}

	got, err := ParseAuditLogCursor(cursor)
	if err != nil {
		t.Fatalf("ParseAuditLogCursor returned error: %v", err)
	}
	if got != 12345 {
		t.Fatalf("cursor id = %d, want 12345", got)
	}
}

func TestAuditLogListRequestNormalizeDefaults(t *testing.T) {
	got, err := (AuditLogListRequest{}).Normalize()
	if err != nil {
		t.Fatalf("Normalize returned error: %v", err)
	}
	if got.Page != 1 || got.PageSize != DefaultAuditLogPageSize {
		t.Fatalf("pagination = %+v", got.Pagination)
	}
}
