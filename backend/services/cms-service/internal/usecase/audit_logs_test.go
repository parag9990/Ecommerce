package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
)

type auditLogRepositoryFunc func(context.Context, domain.AuditLogFilter) (domain.AuditLogPage, error)

func (f auditLogRepositoryFunc) ListAuditLogs(ctx context.Context, filter domain.AuditLogFilter) (domain.AuditLogPage, error) {
	return f(ctx, filter)
}

func TestAuditLogUsecaseScopesToActorSellerAndAppliesDefaults(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	var captured domain.AuditLogFilter
	uc := newTestAuditLogUsecase(t, auditLogRepositoryFunc(func(_ context.Context, filter domain.AuditLogFilter) (domain.AuditLogPage, error) {
		captured = filter
		return domain.AuditLogPage{
			Entries: []domain.AuditEntry{{AuditID: "audit_1", SellerID: filter.SellerID, CreatedAt: now}},
		}, nil
	}))
	uc.now = func() time.Time { return now }

	page, err := uc.ListAuditLogs(context.Background(), ListAuditLogsInput{
		Actor:  activeActor(domain.RoleSellerManager),
		Action: domain.PermissionCampaignsUpdate,
	})
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}
	if len(page.Entries) != 1 {
		t.Fatalf("expected one audit entry, got %+v", page)
	}
	if captured.SellerID != "seller_1" || captured.Action != domain.PermissionCampaignsUpdate {
		t.Fatalf("unexpected filter: %+v", captured)
	}
	if captured.PageSize != 20 {
		t.Fatalf("expected default page size 20, got %d", captured.PageSize)
	}
	if got := captured.From.Format("2006-01-02"); got != "2026-04-24" {
		t.Fatalf("expected default from 2026-04-24, got %s", got)
	}
}

func TestAuditLogUsecaseDeniesCatalogEditor(t *testing.T) {
	uc := newTestAuditLogUsecase(t, auditLogRepositoryFunc(func(context.Context, domain.AuditLogFilter) (domain.AuditLogPage, error) {
		t.Fatal("repository should not be called")
		return domain.AuditLogPage{}, nil
	}))

	_, err := uc.ListAuditLogs(context.Background(), ListAuditLogsInput{
		Actor: activeActor(domain.RoleSellerCatalogEditor),
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestAuditLogUsecaseRejectsInvalidDateRange(t *testing.T) {
	uc := newTestAuditLogUsecase(t, auditLogRepositoryFunc(func(context.Context, domain.AuditLogFilter) (domain.AuditLogPage, error) {
		t.Fatal("repository should not be called")
		return domain.AuditLogPage{}, nil
	}))
	from := time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC)

	_, err := uc.ListAuditLogs(context.Background(), ListAuditLogsInput{
		Actor: activeActor(domain.RoleSeller),
		From:  &from,
		To:    &to,
	})
	if !errors.Is(err, domain.ErrInvalidAuditLogFilter) {
		t.Fatalf("expected invalid filter, got %v", err)
	}
}

func newTestAuditLogUsecase(t *testing.T, repo AuditLogRepository) *AuditLogUsecase {
	t.Helper()
	uc, err := NewAuditLogUsecase(
		newTestAuthorizer(t, nil, nil),
		repo,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		AuditLogOptions{DefaultRangeDays: 30, DefaultPageSize: 20, MaxPageSize: 100},
	)
	if err != nil {
		t.Fatalf("NewAuditLogUsecase: %v", err)
	}
	return uc
}
