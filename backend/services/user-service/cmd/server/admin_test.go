package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/parag/ecommerce/backend/services/user-service/internal/config"
	"github.com/parag/ecommerce/backend/services/user-service/internal/observability"
)

func TestAdminServerExposesHealthAndMetrics(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New returned error: %v", err)
	}
	defer db.Close()

	server := newAdminServer(config.HTTPConfig{
		Address:           ":0",
		ReadHeaderTimeout: time.Second,
	}, db, observability.NewMetrics(), nil, nil, nil)

	live := httptest.NewRecorder()
	server.Handler.ServeHTTP(live, httptest.NewRequest(http.MethodGet, "/health/live", nil))
	if live.Code != http.StatusOK || !strings.Contains(live.Body.String(), `"status":"live"`) {
		t.Fatalf("liveness = %d %s", live.Code, live.Body.String())
	}

	metrics := httptest.NewRecorder()
	server.Handler.ServeHTTP(metrics, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if metrics.Code != http.StatusOK || !strings.Contains(metrics.Body.String(), "go_goroutines") {
		t.Fatalf("metrics = %d %s", metrics.Code, metrics.Body.String())
	}
}
