package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestHTTPMiddlewareRecordsStatus(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewHTTP(registry, "ecommerce", "test-service")
	handler := metrics.Middleware(nil, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) }))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/tea", nil))
	if got := testutil.ToFloat64(metrics.Requests.WithLabelValues(http.MethodGet, "/tea", "418")); got != 1 {
		t.Fatalf("requests = %v, want 1", got)
	}
}
