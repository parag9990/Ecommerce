package httptransport

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce/api-gateway/internal/config"
)

func TestCORSMiddlewareAllowsConfiguredPanelOrigin(t *testing.T) {
	nextCalled := false
	handler := CORSMiddleware(config.CORSConfig{AllowedOrigins: []string{"http://localhost:3003"}})(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		}),
	)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	request.Header.Set("Origin", "http://localhost:3003")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK || !nextCalled {
		t.Fatalf("expected request to reach next handler, code=%d next_called=%v", response.Code, nextCalled)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3003" {
		t.Fatalf("unexpected allow origin %q", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("unexpected allow credentials %q", got)
	}
}

func TestCORSMiddlewareHandlesAllowedPreflight(t *testing.T) {
	handler := CORSMiddleware(config.CORSConfig{AllowedOrigins: []string{"http://localhost:3003"}})(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Fatal("preflight must not reach the protected route")
		}),
	)

	request := httptest.NewRequest(http.MethodOptions, "/api/v1/admin/users", nil)
	request.Header.Set("Origin", "http://localhost:3003")
	request.Header.Set("Access-Control-Request-Method", http.MethodGet)
	request.Header.Set("Access-Control-Request-Headers", "authorization,x-client-app,x-request-id")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected 204 preflight response, got %d", response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Headers"); got != corsAllowedHeaders {
		t.Fatalf("unexpected allow headers %q", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("unexpected allow credentials %q", got)
	}
}

func TestCORSMiddlewareRejectsUnknownPreflightOrigin(t *testing.T) {
	handler := CORSMiddleware(config.CORSConfig{AllowedOrigins: []string{"http://localhost:3003"}})(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Fatal("rejected preflight must not reach the protected route")
		}),
	)

	request := httptest.NewRequest(http.MethodOptions, "/api/v1/admin/users", nil)
	request.Header.Set("Origin", "https://attacker.example")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for unknown origin, got %d", response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unknown origin must not be reflected, got %q", got)
	}
}
