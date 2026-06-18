package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
	transporthttp "ecommerce/superadmin-service/internal/transport/http"
)

func TestAuditLogHandlerListsLogs(t *testing.T) {
	usecase := &auditLogUsecaseStub{
		response: domain.AuditLogListResponse{
			Logs: []domain.AuditLog{{
				AuditID:      "aud_1",
				ActorAdminID: "admin_1",
				Action:       "seller.status.suspended",
				ResourceType: "seller",
				ResourceID:   "seller_1",
				RequestID:    "req_1",
				CreatedAt:    time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC),
			}},
		},
	}
	handler := transporthttp.NewAuditLogHandler(usecase, logging.NewNop())
	mux := http.NewServeMux()
	handler.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs?actor_id=admin_1&action=seller.status.suspended&resource_type=seller&resource_id=seller_1&page_size=25&from=2026-05-01T00:00:00Z&to=2026-05-31T23:59:59Z", nil)
	addAdminHeaders(req, "superadmin_1", domain.RoleSuperadmin)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if usecase.request.ActorAdminID != "admin_1" || usecase.request.Action != "seller.status.suspended" {
		t.Fatalf("request = %+v", usecase.request)
	}
	if usecase.request.PageSize != 25 {
		t.Fatalf("page_size = %d, want 25", usecase.request.PageSize)
	}

	var body domain.AuditLogListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response JSON invalid: %v", err)
	}
	if len(body.Logs) != 1 || body.Logs[0].AuditID != "aud_1" {
		t.Fatalf("body = %+v", body)
	}
}

func TestAuditLogHandlerRejectsBadTimeFilter(t *testing.T) {
	handler := transporthttp.NewAuditLogHandler(&auditLogUsecaseStub{}, logging.NewNop())
	mux := http.NewServeMux()
	handler.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs?from=yesterday", nil)
	addAdminHeaders(req, "superadmin_1", domain.RoleSuperadmin)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

type auditLogUsecaseStub struct {
	request  domain.AuditLogListRequest
	response domain.AuditLogListResponse
	err      error
}

func (s *auditLogUsecaseStub) ListAuditLogs(ctx context.Context, req domain.AuditLogListRequest) (domain.AuditLogListResponse, error) {
	if _, ok := domain.ActorFromContext(ctx); !ok {
		return domain.AuditLogListResponse{}, domain.NewAdminContextMissing("admin actor is missing from request context")
	}
	s.request = req
	if s.err != nil {
		return domain.AuditLogListResponse{}, s.err
	}
	return s.response, nil
}
