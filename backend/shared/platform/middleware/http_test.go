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

type allowLimiter struct{}

func (allowLimiter) Allow(context.Context, string) (bool, error) { return true, nil }

var _ RateLimiter = allowLimiter{}
