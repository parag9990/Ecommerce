package httptransport

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gatewayauth "ecommerce/api-gateway/internal/auth"
	"ecommerce/api-gateway/internal/domain"
	"ecommerce/api-gateway/internal/ratelimit"
	"github.com/golang-jwt/jwt/v5"
)

func TestRateLimitPreAuthBlocksRequest(t *testing.T) {
	limiter := &stubRateLimiter{results: []ratelimit.Result{{
		Allowed:    false,
		Limit:      1,
		Remaining:  0,
		RetryAfter: time.Minute,
		ResetAfter: time.Minute,
	}}}
	middleware := testRateLimitMiddleware(t, limiter, []ratelimit.Policy{{
		Name:      "auth.login.ip",
		Dimension: ratelimit.DimensionIP,
		RouteIDs:  []string{"auth.login"},
		Limit:     1,
		Window:    time.Minute,
		Cost:      1,
	}})
	route := domain.RouteDefinition{ID: "auth.login", Method: domain.MethodPost, Path: "/api/v1/auth/login"}
	called := false
	handler := middleware.PreAuth(route)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.RemoteAddr = "203.0.113.10:12345"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if called {
		t.Fatal("expected blocked request not to reach handler")
	}
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
	if rec.Header().Get("RateLimit-Limit") != "1" || rec.Header().Get("Retry-After") != "60" {
		t.Fatalf("expected rate limit headers, got %+v", rec.Header())
	}
	assertErrorCode(t, rec, "RATE_LIMITED")
}

func TestRateLimitPostAuthUsesHashedUserIdentity(t *testing.T) {
	limiter := &stubRateLimiter{results: []ratelimit.Result{{
		Allowed:    true,
		Limit:      60,
		Remaining:  59,
		ResetAfter: time.Second,
	}}}
	middleware := testRateLimitMiddleware(t, limiter, []ratelimit.Policy{{
		Name:      "cart.user",
		Dimension: ratelimit.DimensionUser,
		RouteIDs:  []string{"cart.add_item"},
		Limit:     60,
		Window:    time.Minute,
		Cost:      1,
	}})
	route := domain.RouteDefinition{ID: "cart.add_item", Method: domain.MethodPost, Path: "/api/v1/cart/items"}
	called := false
	handler := middleware.PostAuth(route)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusAccepted)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", nil)
	req = req.WithContext(gatewayauth.WithClaims(req.Context(), gatewayauth.AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: "user_123"},
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called || rec.Code != http.StatusAccepted {
		t.Fatalf("expected request to continue, called=%t code=%d", called, rec.Code)
	}
	if len(limiter.requests) != 1 {
		t.Fatalf("expected one limiter request, got %d", len(limiter.requests))
	}
	if !strings.Contains(limiter.requests[0].Key, ":user:cart.user:") {
		t.Fatalf("expected user policy key, got %s", limiter.requests[0].Key)
	}
	if strings.Contains(limiter.requests[0].Key, "user_123") {
		t.Fatalf("key must not expose raw user id: %s", limiter.requests[0].Key)
	}
}

func TestRateLimitTargetPreservesRequestBody(t *testing.T) {
	limiter := &stubRateLimiter{results: []ratelimit.Result{{
		Allowed:    true,
		Limit:      3,
		Remaining:  2,
		ResetAfter: time.Second,
	}}}
	middleware := testRateLimitMiddleware(t, limiter, []ratelimit.Policy{{
		Name:         "auth.otp_send.target",
		Dimension:    ratelimit.DimensionTarget,
		RouteIDs:     []string{"auth.otp_send"},
		Limit:        3,
		Window:       time.Minute,
		Cost:         1,
		TargetFields: []string{"target"},
	}})
	route := domain.RouteDefinition{ID: "auth.otp_send", Method: domain.MethodPost, Path: "/api/v1/auth/otp/send"}
	const body = `{"target":"demo@example.com","channel":"email"}`
	handler := middleware.PreAuth(route)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read preserved body: %v", err)
		}
		if string(received) != body {
			t.Fatalf("expected preserved body %q, got %q", body, string(received))
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/otp/send", strings.NewReader(body))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected request to continue, got %d", rec.Code)
	}
	if len(limiter.requests) != 1 || strings.Contains(limiter.requests[0].Key, "demo@example.com") {
		t.Fatalf("expected hashed target key, got %+v", limiter.requests)
	}
}

func TestRateLimitFailModes(t *testing.T) {
	route := domain.RouteDefinition{ID: "search.products", Method: domain.MethodGet, Path: "/api/v1/search"}
	policies := []ratelimit.Policy{{
		Name:      "search.public.ip",
		Dimension: ratelimit.DimensionIP,
		RouteIDs:  []string{"search.products"},
		Limit:     120,
		Window:    time.Minute,
		Cost:      1,
	}}

	closed := testRateLimitMiddleware(t, &stubRateLimiter{err: errors.New("redis down")}, policies)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search", nil)
	req.RemoteAddr = "203.0.113.10:12345"
	rec := httptest.NewRecorder()
	closed.PreAuth(route)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("fail-closed request should not continue")
	})).ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected fail-closed 503, got %d", rec.Code)
	}

	open := testRateLimitMiddleware(t, &stubRateLimiter{err: errors.New("redis down")}, policies)
	open.failOpen = true
	req = httptest.NewRequest(http.MethodGet, "/api/v1/search", nil)
	req.RemoteAddr = "203.0.113.10:12345"
	rec = httptest.NewRecorder()
	open.PreAuth(route)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})).ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected fail-open request to continue, got %d", rec.Code)
	}
}

func testRateLimitMiddleware(t *testing.T, limiter ratelimit.Limiter, policies []ratelimit.Policy) *rateLimitMiddleware {
	t.Helper()
	resolver, err := newClientIPResolver(nil)
	if err != nil {
		t.Fatalf("new resolver: %v", err)
	}
	return &rateLimitMiddleware{
		limiter:              limiter,
		policies:             policies,
		keyPrefix:            "rl:test",
		targetBodyLimitBytes: 64 * 1024,
		clientIP:             resolver,
	}
}

type stubRateLimiter struct {
	results  []ratelimit.Result
	err      error
	requests []ratelimit.Request
}

func (s *stubRateLimiter) Allow(_ context.Context, req ratelimit.Request) (ratelimit.Result, error) {
	s.requests = append(s.requests, req)
	if s.err != nil {
		return ratelimit.Result{}, s.err
	}
	if len(s.results) == 0 {
		return ratelimit.Result{Allowed: true, Limit: req.Limit, Remaining: req.Limit - 1}, nil
	}
	result := s.results[0]
	if len(s.results) > 1 {
		s.results = s.results[1:]
	}
	return result, nil
}
