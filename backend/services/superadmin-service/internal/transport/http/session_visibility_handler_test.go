package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
	"ecommerce/superadmin-service/internal/rbac"
	transporthttp "ecommerce/superadmin-service/internal/transport/http"
	"ecommerce/superadmin-service/internal/usecase"
)

func TestAuthorizeSessionAnalyticsHandlerAllowsReadonlyMasked(t *testing.T) {
	mux := sessionVisibilityMux(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/session-analytics/authorize", strings.NewReader(`{"route":"analytics.sessions","page_size":20}`))
	addAdminHeaders(req, "readonly_1", domain.RoleReadonly)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Allowed           bool              `json:"allowed"`
		RiskAllowed       bool              `json:"risk_allowed"`
		MaskPII           bool              `json:"mask_pii"`
		DownstreamHeaders map[string]string `json:"downstream_headers"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Allowed || body.RiskAllowed || !body.MaskPII {
		t.Fatalf("body = %+v", body)
	}
	if body.DownstreamHeaders["x-admin-mask-pii"] != "true" {
		t.Fatalf("downstream headers = %+v", body.DownstreamHeaders)
	}
}

func TestAuthorizeSessionAnalyticsHandlerBlocksReadonlyRisk(t *testing.T) {
	mux := sessionVisibilityMux(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/session-analytics/authorize", strings.NewReader(`{"route":"analytics.sessions","include_risk":true}`))
	addAdminHeaders(req, "readonly_1", domain.RoleReadonly)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), string(domain.PermissionSessionRiskRead)) {
		t.Fatalf("body = %s, want required risk permission", rec.Body.String())
	}
}

func TestAuthorizeSessionAnalyticsHandlerValidatesFilters(t *testing.T) {
	mux := sessionVisibilityMux(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/session-analytics/authorize", strings.NewReader(`{"route":"analytics.heatmaps","path":"products/1"}`))
	addAdminHeaders(req, "superadmin_1", domain.RoleSuperadmin)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestSessionDashboardAccessHandler(t *testing.T) {
	mux := sessionVisibilityMux(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/session-analytics/access", nil)
	addAdminHeaders(req, "operations_1", domain.RoleOperations)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var body struct {
		CanViewDashboard bool `json:"can_view_dashboard"`
		RiskAllowed      bool `json:"risk_allowed"`
		Modules          []struct {
			Key     string `json:"key"`
			Visible bool   `json:"visible"`
		} `json:"modules"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.CanViewDashboard || !body.RiskAllowed {
		t.Fatalf("body = %+v", body)
	}
	if !dashboardModuleVisible(body.Modules, "suspicious_activity") {
		t.Fatalf("modules = %+v", body.Modules)
	}
}

func sessionVisibilityMux(t *testing.T) *http.ServeMux {
	t.Helper()
	repo := rbac.NewStaticPermissionRepository([]rbac.StaticAdmin{
		{AdminID: "superadmin_1", Role: domain.RoleSuperadmin, Active: true},
		{AdminID: "operations_1", Role: domain.RoleOperations, Active: true},
		{AdminID: "readonly_1", Role: domain.RoleReadonly, Active: true},
	})
	authz, err := usecase.NewAuthorizationService(repo, logging.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	visibility, err := usecase.NewSessionVisibilityService(authz, usecase.SessionVisibilityConfig{
		MaxDateRange: 30 * 24 * time.Hour,
		MaxPageSize:  100,
	}, logging.NewNop())
	if err != nil {
		t.Fatal(err)
	}

	rbacHandler := transporthttp.NewRBACHandler(authz, nil, logging.NewNop())
	sessionHandler := transporthttp.NewSessionVisibilityHandler(visibility, logging.NewNop())
	return transporthttp.NewServeMux(rbacHandler, sessionHandler)
}

func dashboardModuleVisible(modules []struct {
	Key     string `json:"key"`
	Visible bool   `json:"visible"`
}, key string) bool {
	for _, module := range modules {
		if module.Key == key {
			return module.Visible
		}
	}
	return false
}
