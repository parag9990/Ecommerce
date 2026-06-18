package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
	"ecommerce/superadmin-service/internal/rbac"
	transporthttp "ecommerce/superadmin-service/internal/transport/http"
	"ecommerce/superadmin-service/internal/usecase"
)

func TestPermissionCatalogHandler(t *testing.T) {
	mux := testMux(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/rbac/permissions", nil)
	addAdminHeaders(req, "superadmin_1", domain.RoleSuperadmin)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestPermissionCatalogHandlerRequiresActor(t *testing.T) {
	mux := testMux(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/rbac/permissions", nil)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestRolePermissionsHandler(t *testing.T) {
	mux := testMux(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/rbac/roles/finance_admin/permissions", nil)
	addAdminHeaders(req, "superadmin_1", domain.RoleSuperadmin)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func testMux(t *testing.T) *http.ServeMux {
	t.Helper()
	repo := rbac.NewStaticPermissionRepository([]rbac.StaticAdmin{
		{AdminID: "superadmin_1", Role: domain.RoleSuperadmin, Active: true},
	})
	authz, err := usecase.NewAuthorizationService(repo, logging.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	handler := transporthttp.NewRBACHandler(authz, nil, logging.NewNop())
	return transporthttp.NewServeMux(handler)
}

func addAdminHeaders(req *http.Request, adminID string, role domain.AdminRole) {
	req.Header.Set(transporthttp.HeaderAdminID, adminID)
	req.Header.Set(transporthttp.HeaderUserID, "user_"+adminID)
	req.Header.Set(transporthttp.HeaderAdminRoles, string(role))
	req.Header.Set(transporthttp.HeaderSessionID, "sess_1")
	req.Header.Set(transporthttp.HeaderRequestID, "req_1")
	req.Header.Set(transporthttp.HeaderMFAVerified, "true")
}
