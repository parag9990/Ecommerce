package usecase

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
)

func TestReportsUsecaseExportEscapesSpreadsheetFormulas(t *testing.T) {
	repo := &reportsRepoFake{table: domain.TabularReport{Headers: []string{"name"}, Rows: [][]string{{"=HYPERLINK(\"https://example.invalid\")"}}}}
	service, err := NewReportsUsecase(repo, ReportsConfig{})
	if err != nil {
		t.Fatalf("new reports usecase: %v", err)
	}
	out, err := service.Export(context.Background(), ExportReportInput{Actor: domain.Actor{ID: "admin_1", Roles: []string{"admin"}}, ReportType: "overview", Format: "csv", From: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), To: time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC), Timezone: "UTC"})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if !strings.Contains(string(out.Data), "'=HYPERLINK") {
		t.Fatalf("formula was not escaped: %q", out.Data)
	}
	if len(repo.audits) != 1 {
		t.Fatalf("expected export audit, got %d", len(repo.audits))
	}
}

func TestReportsUsecaseCreatesValidatedWeeklySchedule(t *testing.T) {
	repo := &reportsRepoFake{}
	service, _ := NewReportsUsecase(repo, ReportsConfig{})
	out, err := service.CreateSchedule(context.Background(), domain.Actor{ID: "ops_1", Roles: []string{"operations_admin"}}, domain.CreateReportSchedule{Name: "Weekly funnel", ReportType: domain.ReportFunnel, Format: domain.ReportFormatCSV, Frequency: domain.ReportFrequencyWeekly, Timezone: "UTC", TimeOfDay: "09:00", DayOfWeek: "monday", Recipients: []string{"OPS@example.com", "ops@example.com"}}, "req_1")
	if err != nil {
		t.Fatalf("create schedule: %v", err)
	}
	if out.Status != domain.ReportScheduleActive || out.NextRunAt == nil {
		t.Fatalf("unexpected schedule: %+v", out)
	}
	if len(out.Recipients) != 1 || out.Recipients[0] != "ops@example.com" {
		t.Fatalf("recipients not normalized: %#v", out.Recipients)
	}
	if len(repo.schedules) != 1 {
		t.Fatal("schedule was not persisted")
	}
}

type reportsRepoFake struct {
	table     domain.TabularReport
	schedules []domain.ReportSchedule
	audits    []domain.AuditEvent
}

func (r *reportsRepoFake) BuildReport(context.Context, domain.ReportQuery, int) (domain.TabularReport, error) {
	return r.table, nil
}
func (r *reportsRepoFake) ListSchedules(context.Context, int) ([]domain.ReportSchedule, error) {
	return append([]domain.ReportSchedule(nil), r.schedules...), nil
}
func (r *reportsRepoFake) CreateSchedule(_ context.Context, s domain.ReportSchedule) error {
	r.schedules = append(r.schedules, s)
	return nil
}
func (r *reportsRepoFake) GetSchedule(_ context.Context, id string) (domain.ReportSchedule, error) {
	for _, item := range r.schedules {
		if item.ID == id {
			return item, nil
		}
	}
	return domain.ReportSchedule{}, domain.ErrNotFound
}
func (r *reportsRepoFake) UpdateScheduleStatus(_ context.Context, id string, status domain.ReportScheduleStatus, next *time.Time, updated time.Time) (domain.ReportSchedule, error) {
	return domain.ReportSchedule{ID: id, Status: status, NextRunAt: next, UpdatedAt: updated}, nil
}
func (r *reportsRepoFake) DeleteSchedule(context.Context, string) error { return nil }
func (r *reportsRepoFake) CreateAuditEvent(_ context.Context, event domain.AuditEvent) error {
	r.audits = append(r.audits, event)
	return nil
}
