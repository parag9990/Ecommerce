package usecase_test

import (
	"context"
	"testing"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
	"ecommerce/superadmin-service/internal/usecase"
)

func TestAuditLogServiceListRequiresReadPermission(t *testing.T) {
	service, err := usecase.NewAuditLogService(
		&controlAuthorizer{deny: domain.PermissionAuditLogsRead},
		&auditLogRepoStub{},
		logging.NewNop(),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.ListAuditLogs(actorContext(), domain.AuditLogListRequest{})
	if appErr, ok := domain.AsAppError(err); !ok || appErr.Code != domain.CodeForbidden {
		t.Fatalf("error = %v, want FORBIDDEN", err)
	}
}

func TestAuditLogServiceListNormalizesAndDelegates(t *testing.T) {
	repo := &auditLogRepoStub{
		response: domain.AuditLogListResponse{
			Logs: []domain.AuditLog{{AuditID: "aud_1", Action: "seller.status.suspended"}},
		},
	}
	service, err := usecase.NewAuditLogService(&controlAuthorizer{}, repo, logging.NewNop())
	if err != nil {
		t.Fatal(err)
	}

	got, err := service.ListAuditLogs(actorContext(), domain.AuditLogListRequest{
		ActorAdminID: " admin_2 ",
		ResourceType: " seller ",
	})
	if err != nil {
		t.Fatalf("ListAuditLogs returned error: %v", err)
	}
	if len(got.Logs) != 1 || got.Logs[0].AuditID != "aud_1" {
		t.Fatalf("response = %+v", got)
	}
	if repo.request.ActorAdminID != "admin_2" || repo.request.ResourceType != "seller" {
		t.Fatalf("request was not normalized: %+v", repo.request)
	}
	if repo.request.PageSize != domain.DefaultAuditLogPageSize {
		t.Fatalf("page size = %d, want %d", repo.request.PageSize, domain.DefaultAuditLogPageSize)
	}
}

type auditLogRepoStub struct {
	request  domain.AuditLogListRequest
	response domain.AuditLogListResponse
	err      error
}

func (r *auditLogRepoStub) InsertAuditLog(ctx context.Context, record domain.AuditRecord) error {
	return r.err
}

func (r *auditLogRepoStub) ListAuditLogs(ctx context.Context, req domain.AuditLogListRequest) (domain.AuditLogListResponse, error) {
	r.request = req
	if r.err != nil {
		return domain.AuditLogListResponse{}, r.err
	}
	return r.response, nil
}
