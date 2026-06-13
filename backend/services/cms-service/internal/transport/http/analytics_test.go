package httptransport

import (
	"net/http/httptest"
	"testing"
)

func TestOptionalAnalyticsTimeQueryParsesDateAndRFC3339(t *testing.T) {
	req := httptest.NewRequest("GET", "/?from=2026-05-01&to=2026-05-24T10:30:00Z", nil)

	from, err := optionalAnalyticsTimeQuery(req, "from")
	if err != nil {
		t.Fatalf("from: %v", err)
	}
	if got := from.Format("2006-01-02"); got != "2026-05-01" {
		t.Fatalf("expected date 2026-05-01, got %s", got)
	}
	to, err := optionalAnalyticsTimeQuery(req, "to")
	if err != nil {
		t.Fatalf("to: %v", err)
	}
	if got := to.Format("2006-01-02T15:04:05Z07:00"); got != "2026-05-24T10:30:00Z" {
		t.Fatalf("expected RFC3339 timestamp, got %s", got)
	}
}

func TestOptionalAnalyticsLimitQueryRejectsInvalidLimit(t *testing.T) {
	req := httptest.NewRequest("GET", "/?top_products_limit=0", nil)

	if _, err := optionalAnalyticsLimitQuery(req, "top_products_limit"); err == nil {
		t.Fatal("expected invalid limit error")
	}
}

func TestOptionalAuditQueryParsesDateAndPageSize(t *testing.T) {
	req := httptest.NewRequest("GET", "/?from=2026-05-01&page_size=25", nil)

	from, err := optionalAuditTimeQuery(req, "from")
	if err != nil {
		t.Fatalf("from: %v", err)
	}
	if got := from.Format("2006-01-02"); got != "2026-05-01" {
		t.Fatalf("expected date 2026-05-01, got %s", got)
	}
	pageSize, err := optionalAuditPageSizeQuery(req, "page_size")
	if err != nil {
		t.Fatalf("page_size: %v", err)
	}
	if pageSize != 25 {
		t.Fatalf("expected page size 25, got %d", pageSize)
	}
}
