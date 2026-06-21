package httptransport

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadinessHandlerReportsDependenciesReady(t *testing.T) {
	handler := readinessHandler([]ReadinessDependency{
		{Name: "mysql", Check: func(context.Context) error { return nil }},
		{Name: "redis", Check: func(context.Context) error { return nil }},
	})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if body := recorder.Body.String(); !strings.Contains(body, `"mysql":"ready"`) || !strings.Contains(body, `"redis":"ready"`) {
		t.Fatalf("body = %s", body)
	}
}

func TestReadinessHandlerFailsClosedWithoutLeakingDependencyError(t *testing.T) {
	const sensitiveError = "dial root:password@mysql"
	handler := readinessHandler([]ReadinessDependency{
		{Name: "mysql", Check: func(context.Context) error { return errors.New(sensitiveError) }},
	})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), sensitiveError) {
		t.Fatalf("readiness response leaked dependency error: %s", recorder.Body.String())
	}
}
