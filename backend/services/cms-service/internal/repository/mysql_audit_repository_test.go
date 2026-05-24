package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
)

func TestAuditCursorRoundTrip(t *testing.T) {
	createdAt := time.Date(2026, 5, 24, 10, 30, 0, 123000000, time.UTC)
	encoded, err := encodeAuditCursor(auditCursor{CreatedAt: createdAt, ID: 42})
	if err != nil {
		t.Fatalf("encodeAuditCursor: %v", err)
	}
	decoded, ok, err := decodeAuditCursor(encoded)
	if err != nil {
		t.Fatalf("decodeAuditCursor: %v", err)
	}
	if !ok || decoded.ID != 42 || !decoded.CreatedAt.Equal(createdAt) {
		t.Fatalf("unexpected decoded cursor: ok=%v cursor=%+v", ok, decoded)
	}
}

func TestAuditCursorRejectsInvalidValue(t *testing.T) {
	_, _, err := decodeAuditCursor("not-base64")
	if !errors.Is(err, domain.ErrInvalidAuditLogFilter) {
		t.Fatalf("expected invalid audit log filter, got %v", err)
	}
}
