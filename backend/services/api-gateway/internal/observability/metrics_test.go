package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce/api-gateway/internal/domain"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestMetricsMiddlewareUsesRouteTemplateLabel(t *testing.T) {
	cfg := DefaultConfig("api-gateway", "test")
	registry := prometheus.NewRegistry()
	metrics, err := NewMetrics(cfg, registry)
	if err != nil {
		t.Fatalf("new metrics: %v", err)
	}
	labeler := NewRouteLabeler([]domain.RouteDefinition{{
		Method: domain.MethodGet,
		Path:   "/api/v1/products/{product_id}",
	}})
	handler := MetricsMiddleware(metrics, labeler)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/prod_12345678", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}
	labels := findCounterLabels(t, metricFamilies, "http_requests_total")
	if labels["route"] != "/api/v1/products/{product_id}" {
		t.Fatalf("route label = %q", labels["route"])
	}
	if labels["route"] == "/api/v1/products/prod_12345678" {
		t.Fatal("raw resource id leaked into route label")
	}
	if labels["status_code"] != "404" {
		t.Fatalf("status_code label = %q", labels["status_code"])
	}
}

func findCounterLabels(t *testing.T, families []*dto.MetricFamily, name string) map[string]string {
	t.Helper()
	for _, family := range families {
		if family.GetName() != name {
			continue
		}
		for _, metric := range family.GetMetric() {
			labels := make(map[string]string, len(metric.GetLabel()))
			for _, label := range metric.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}
			return labels
		}
	}
	t.Fatalf("metric %s not found", name)
	return nil
}
