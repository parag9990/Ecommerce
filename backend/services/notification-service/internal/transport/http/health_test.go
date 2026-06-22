package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthHandlerReportsLivenessAndReadiness(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		check      ReadinessCheck
		path       string
		wantStatus int
	}{
		{name: "live", check: func(context.Context) error { return errors.New("dependency down") }, path: "/healthz", wantStatus: http.StatusNoContent},
		{name: "ready", check: func(context.Context) error { return nil }, path: "/readyz", wantStatus: http.StatusNoContent},
		{name: "not ready", check: func(context.Context) error { return errors.New("dependency down") }, path: "/readyz", wantStatus: http.StatusServiceUnavailable},
		{name: "missing check", path: "/readyz", wantStatus: http.StatusServiceUnavailable},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			mux := http.NewServeMux()
			NewHealthHandler(test.check, time.Second).Register(mux)
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
		})
	}
}

func TestHealthHandlerBoundsReadinessCheck(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	NewHealthHandler(func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}, 5*time.Millisecond).Register(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}
