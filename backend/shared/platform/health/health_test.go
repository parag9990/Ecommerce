package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLiveAndReadyAreSeparate(t *testing.T) {
	live := httptest.NewRecorder()
	LivenessHandler("service", "test").ServeHTTP(live, httptest.NewRequest(http.MethodGet, "/health/live", nil))
	if live.Code != http.StatusOK || !strings.Contains(live.Body.String(), `"status":"live"`) {
		t.Fatalf("liveness = %d %s", live.Code, live.Body.String())
	}

	ready := httptest.NewRecorder()
	ReadinessHandler("service", "test", map[string]Check{"db": func(context.Context) error { return errors.New("down") }}, time.Second).ServeHTTP(ready, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if ready.Code != http.StatusServiceUnavailable || !strings.Contains(ready.Body.String(), `"db":"unavailable"`) {
		t.Fatalf("readiness = %d %s", ready.Code, ready.Body.String())
	}
}
