package httptransport

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	gatewayauth "ecommerce/api-gateway/internal/auth"
	"github.com/golang-jwt/jwt/v5"
)

func TestSessionProxyReplacesSpoofedIdentityHeaders(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-User-ID") != "admin_1" || r.Header.Get("X-User-Roles") != "admin" {
			t.Errorf("unexpected identity headers: %#v", r.Header)
		}
		if r.URL.Path != "/api/v1/analytics/live" || r.URL.Query().Get("from") != "2026-05-01" {
			t.Errorf("request target not preserved: %s", r.URL.String())
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer upstream.Close()
	proxy, err := NewSessionProxy(upstream.URL, upstream.Client(), nil)
	if err != nil {
		t.Fatalf("new proxy: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/live?from=2026-05-01", nil)
	req.Header.Set("X-User-ID", "attacker")
	req.Header.Set("X-Roles", "superadmin")
	claims := gatewayauth.AccessClaims{Roles: []string{"admin"}, RegisteredClaims: jwt.RegisteredClaims{Subject: "admin_1"}}
	req = req.WithContext(gatewayauth.WithClaims(req.Context(), claims))
	rr := httptest.NewRecorder()
	proxy.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestSuperadminProxyBuildsTrustedAdminContext(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Admin-ID") != "admin_1" || r.Header.Get("X-Session-ID") != "sess_1" || r.Header.Get("X-MFA-Verified") != "true" {
			t.Errorf("headers=%v", r.Header)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()
	proxy, err := NewServiceHTTPProxy("superadmin", upstream.URL, upstream.Client(), nil)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	request.Header.Set("X-Admin-ID", "attacker")
	request.Header.Set("X-MFA-Verified", "true")
	request = request.WithContext(gatewayauth.WithClaims(request.Context(), gatewayauth.AccessClaims{SessionID: "sess_1", Roles: []string{"superadmin"}, AMR: []string{"pwd", "mfa"}, RegisteredClaims: jwt.RegisteredClaims{Subject: "admin_1"}}))
	response := httptest.NewRecorder()
	proxy.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("status=%d", response.Code)
	}
}

func TestPaymentProxyInjectsInternalBearerAndActorRole(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer payment-internal-token-at-least-32-chars" {
			t.Errorf("authorization=%q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-Actor-ID") != "buyer_1" || r.Header.Get("X-Actor-Role") != "buyer" {
			t.Errorf("actor headers=%v", r.Header)
		}
		if r.Header.Get("X-User-ID") != "buyer_1" || r.Header.Get("X-User-Roles") != "buyer" {
			t.Errorf("user headers=%v", r.Header)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer upstream.Close()

	proxy, err := NewServiceHTTPProxy(
		"payment",
		upstream.URL,
		upstream.Client(),
		nil,
		WithBearerAuth("payment-internal-token-at-least-32-chars"),
		WithActorRoleHeader(),
	)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/payments/pay_1/retry", nil)
	request.Header.Set("Authorization", "Bearer user-jwt")
	request.Header.Set("X-Actor-Role", "superadmin")
	request = request.WithContext(gatewayauth.WithClaims(request.Context(), gatewayauth.AccessClaims{Roles: []string{"buyer"}, RegisteredClaims: jwt.RegisteredClaims{Subject: "buyer_1"}}))
	response := httptest.NewRecorder()
	proxy.ServeHTTP(response, request)
	if response.Code != http.StatusAccepted {
		t.Fatalf("status=%d", response.Code)
	}
}
