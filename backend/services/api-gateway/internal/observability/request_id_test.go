package observability

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestIDMiddlewarePreservesValidIncomingID(t *testing.T) {
	handler := RequestIDMiddleware(HeaderRequestID)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := RequestIDFromContext(r.Context()); got != "req_test_123" {
			t.Fatalf("request id from context = %q", got)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	req.Header.Set(HeaderRequestID, "req_test_123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(HeaderRequestID); got != "req_test_123" {
		t.Fatalf("response request id = %q", got)
	}
}

func TestRequestIDMiddlewareReplacesUnsafeIncomingID(t *testing.T) {
	handler := RequestIDMiddleware(HeaderRequestID)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := RequestIDFromContext(r.Context()); got == "bad id with spaces" || !strings.HasPrefix(got, "req_") {
			t.Fatalf("unsafe request id was not replaced: %q", got)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	req.Header.Set(HeaderRequestID, "bad id with spaces")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(HeaderRequestID); got == "bad id with spaces" || !strings.HasPrefix(got, "req_") {
		t.Fatalf("unsafe response request id was not replaced: %q", got)
	}
}
