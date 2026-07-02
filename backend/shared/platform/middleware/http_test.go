package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDGeneratesAndReturnsHeader(t *testing.T) {
	recorder := httptest.NewRecorder()
	RequestID(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Header().Get(RequestIDHeader) == "" {
		t.Fatal("request id header is empty")
	}
}

func TestRequestIDPreservesValidHeaderAndContext(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set(RequestIDHeader, "req-test")
	RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := RequestIDFromContext(r.Context()); got != "req-test" {
			t.Fatalf("request id = %q", got)
		}
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(recorder, request)
	if got := recorder.Header().Get(RequestIDHeader); got != "req-test" {
		t.Fatalf("response request id = %q", got)
	}
}

func TestRecoverReturnsSafeJSON(t *testing.T) {
	recorder := httptest.NewRecorder()
	Recover(slog.New(slog.DiscardHandler), http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("secret panic") })).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusInternalServerError || recorder.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("response = %d, %q", recorder.Code, recorder.Header().Get("Content-Type"))
	}
}

func TestCORSAllowsCredentialedRequest(t *testing.T) {
	handler := CORS(CORSConfig{AllowedOrigins: []string{"http://localhost:3000"}, AllowCredentials: true})(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }),
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	handler.ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("allow origin = %q", got)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("allow credentials = %q", got)
	}
}

func TestCORSHandlesCustomHeaderPreflight(t *testing.T) {
	handler := CORS(CORSConfig{AllowedOrigins: []string{"http://localhost:3001"}, AllowCredentials: true})(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Fatal("preflight must not reach next handler")
		}),
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/seller/products", nil)
	request.Header.Set("Origin", "http://localhost:3001")
	request.Header.Set("Access-Control-Request-Method", http.MethodGet)
	request.Header.Set("Access-Control-Request-Headers", "x-client-app,x-request-id")
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d", recorder.Code)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Headers"); got != defaultCORSAllowedHeaders {
		t.Fatalf("allow headers = %q", got)
	}
}

func TestCORSRejectsUnknownPreflightOrigin(t *testing.T) {
	handler := CORS(CORSConfig{AllowedOrigins: []string{"http://localhost:3000"}})(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Fatal("rejected preflight must not reach next handler")
		}),
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/products", nil)
	request.Header.Set("Origin", "https://example.test")
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d", recorder.Code)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("allow origin = %q", got)
	}
}

type allowLimiter struct{}

func (allowLimiter) Allow(context.Context, string) (bool, error) { return true, nil }

var _ RateLimiter = allowLimiter{}
