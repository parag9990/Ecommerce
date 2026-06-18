package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
	transporthttp "ecommerce/superadmin-service/internal/transport/http"
)

func TestListUsersHandler(t *testing.T) {
	controls := &controlUsecaseStub{}
	mux := controlMux(controls)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users?q=alice&status=active&page=2&page_size=10", nil)
	addAdminHeaders(req, "superadmin_1", domain.RoleSuperadmin)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if controls.userListReq.Query != "alice" || controls.userListReq.Status != domain.UserStatusActive {
		t.Fatalf("request = %+v", controls.userListReq)
	}
	if controls.userListReq.Page != 2 || controls.userListReq.PageSize != 10 {
		t.Fatalf("pagination = %+v", controls.userListReq.Pagination)
	}
}

func TestUpdateUserStatusHandler(t *testing.T) {
	controls := &controlUsecaseStub{}
	mux := controlMux(controls)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/user_1/status", strings.NewReader(`{"status":"blocked","reason":"confirmed abuse by support"}`))
	addAdminHeaders(req, "superadmin_1", domain.RoleSuperadmin)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if controls.updatedUserID != "user_1" || controls.statusReq.Status != "blocked" {
		t.Fatalf("updated user id/status = %s/%+v", controls.updatedUserID, controls.statusReq)
	}
}

func TestUpdateSellerStatusHandlerReturnsValidationError(t *testing.T) {
	controls := &controlUsecaseStub{
		updateSellerErr: domain.NewInvalidStatusTransition("seller", "draft", "active"),
	}
	mux := controlMux(controls)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/sellers/seller_1/status", strings.NewReader(`{"status":"active","reason":"documents are not submitted"}`))
	addAdminHeaders(req, "superadmin_1", domain.RoleSuperadmin)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["code"] != string(domain.CodeInvalidStatusTransition) {
		t.Fatalf("error code = %s", body["code"])
	}
}

func controlMux(controls *controlUsecaseStub) *http.ServeMux {
	rbacHandler := transporthttp.NewRBACHandler(&authorizationStub{}, nil, logging.NewNop())
	controlHandler := transporthttp.NewControlHandler(controls, logging.NewNop())
	return transporthttp.NewServeMux(rbacHandler, controlHandler)
}

type controlUsecaseStub struct {
	userListReq     domain.AdminUserListRequest
	updatedUserID   string
	updatedSellerID string
	statusReq       domain.StatusUpdateRequest
	updateSellerErr error
}

func (s *controlUsecaseStub) ListUsersForAdmin(ctx context.Context, req domain.AdminUserListRequest) (domain.AdminUserListResponse, error) {
	s.userListReq = req
	return domain.AdminUserListResponse{
		Users: []domain.UserProfile{{UserID: "user_1", Status: domain.UserStatusActive}},
	}, nil
}

func (s *controlUsecaseStub) UpdateUserStatus(ctx context.Context, userID string, req domain.StatusUpdateRequest) (domain.SuccessResponse, error) {
	s.updatedUserID = userID
	s.statusReq = req
	return domain.SuccessResponse{Success: true}, nil
}

func (s *controlUsecaseStub) ListSellersForAdmin(ctx context.Context, req domain.AdminSellerListRequest) (domain.AdminSellerListResponse, error) {
	return domain.AdminSellerListResponse{
		Sellers: []domain.SellerProfile{{SellerID: "seller_1", Status: domain.SellerStatusPendingReview}},
	}, nil
}

func (s *controlUsecaseStub) UpdateSellerStatus(ctx context.Context, sellerID string, req domain.StatusUpdateRequest) (domain.SuccessResponse, error) {
	s.updatedSellerID = sellerID
	s.statusReq = req
	if s.updateSellerErr != nil {
		return domain.SuccessResponse{}, s.updateSellerErr
	}
	return domain.SuccessResponse{Success: true}, nil
}

type authorizationStub struct{}

func (s *authorizationStub) PermissionCatalog(ctx context.Context, actor domain.AdminActor) ([]domain.PermissionDefinition, error) {
	return domain.PermissionDefinitions(), nil
}

func (s *authorizationStub) ListRolePermissions(ctx context.Context, actor domain.AdminActor, role domain.AdminRole) ([]domain.Permission, error) {
	return domain.KnownPermissions(), nil
}

func (s *authorizationStub) RequirePermission(ctx context.Context, actor domain.AdminActor, permission domain.Permission) error {
	return nil
}
