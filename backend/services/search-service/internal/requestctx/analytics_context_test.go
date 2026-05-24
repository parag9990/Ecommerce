package requestctx

import (
	"context"
	"strings"
	"testing"
)

func TestAnalyticsContextNormalize(t *testing.T) {
	ctx := WithAnalyticsContext(context.Background(), AnalyticsContext{
		RequestID:   " req_1 ",
		AnonymousID: " anon_1 ",
		SessionID:   " sess_1 ",
		UserID:      " user_1 ",
		ClientPath:  "search?token=hidden",
	})

	analytics := Analytics(ctx)
	if analytics.RequestID != "req_1" || analytics.AnonymousID != "anon_1" || analytics.SessionID != "sess_1" || analytics.UserID != "user_1" {
		t.Fatalf("analytics = %#v", analytics)
	}
	if analytics.ClientPath != "/search" {
		t.Fatalf("client path = %q", analytics.ClientPath)
	}
	if !analytics.CanIngestSessionEvent() {
		t.Fatal("expected analytics context to be ingestable")
	}
}

func TestNormalizeClientPathDefaultsAndTruncates(t *testing.T) {
	if NormalizeClientPath("") != DefaultClientPath {
		t.Fatalf("empty path default = %q", NormalizeClientPath(""))
	}
	longPath := "/" + strings.Repeat("a", maxClientPathLen+10)
	if len([]rune(NormalizeClientPath(longPath))) != maxClientPathLen {
		t.Fatalf("path was not truncated")
	}
}
