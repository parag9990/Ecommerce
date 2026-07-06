package usecase

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/session-service/internal/domain"
)

var ErrReportNotFound = errors.New("analytics report schedule not found")
var ErrReportStorageUnavailable = errors.New("analytics report storage unavailable")

type ReportRepository interface {
	BuildReport(ctx context.Context, query domain.ReportQuery, maxRows int) (domain.TabularReport, error)
	ListSchedules(ctx context.Context, limit int) ([]domain.ReportSchedule, error)
	CreateSchedule(ctx context.Context, schedule domain.ReportSchedule) error
	GetSchedule(ctx context.Context, id string) (domain.ReportSchedule, error)
	UpdateScheduleStatus(ctx context.Context, id string, status domain.ReportScheduleStatus, nextRunAt *time.Time, updatedAt time.Time) (domain.ReportSchedule, error)
	DeleteSchedule(ctx context.Context, id string) error
	CreateAuditEvent(ctx context.Context, event domain.AuditEvent) error
}

type ReportsConfig struct {
	MaxRangeDays int
	MaxRows      int
	MaxBytes     int
	ListLimit    int
}

type ReportsUsecase struct {
	repo  ReportRepository
	cfg   ReportsConfig
	clock Clock
}

type ExportReportInput struct {
	Actor      domain.Actor
	ReportType string
	Format     string
	From       time.Time
	To         time.Time
	Timezone   string
	Filters    domain.ReportFilters
	RequestID  string
}

type UpdateReportScheduleStatusInput struct {
	Actor     domain.Actor
	ID        string
	Status    string
	RequestID string
}

func NewReportsUsecase(repo ReportRepository, cfg ReportsConfig) (*ReportsUsecase, error) {
	if repo == nil {
		return nil, errors.New("report repository is required")
	}
	if cfg.MaxRangeDays == 0 {
		cfg.MaxRangeDays = 366
	}
	if cfg.MaxRows == 0 {
		cfg.MaxRows = 10000
	}
	if cfg.MaxBytes == 0 {
		cfg.MaxBytes = 5 << 20
	}
	if cfg.ListLimit == 0 {
		cfg.ListLimit = 100
	}
	if cfg.MaxRangeDays < 1 || cfg.MaxRows < 1 || cfg.MaxBytes < 1024 || cfg.ListLimit < 1 || cfg.ListLimit > 200 {
		return nil, errors.New("invalid reports configuration")
	}
	return &ReportsUsecase{repo: repo, cfg: cfg, clock: realClock{}}, nil
}

func (u *ReportsUsecase) Export(ctx context.Context, input ExportReportInput) (domain.ReportExport, error) {
	if err := requireAnyRole(input.Actor, analyticsAdminRoles...); err != nil {
		return domain.ReportExport{}, err
	}
	reportType := domain.AnalyticsReportType(strings.TrimSpace(input.ReportType))
	format := domain.ReportFormat(strings.TrimSpace(input.Format))
	if !reportType.Valid() {
		return domain.ReportExport{}, domain.NewFieldError("reportType", "unsupported report type")
	}
	if format != domain.ReportFormatCSV {
		return domain.ReportExport{}, domain.NewFieldError("format", "only csv is supported")
	}
	if err := validateRequiredDateRange(input.From.UTC(), input.To.UTC(), daysDuration(u.cfg.MaxRangeDays)); err != nil {
		return domain.ReportExport{}, err
	}
	timezone, err := normalizeTimezone(input.Timezone)
	if err != nil {
		return domain.ReportExport{}, err
	}
	query := domain.ReportQuery{ReportType: reportType, Format: format, From: input.From.UTC(), To: input.To.UTC(), Timezone: timezone, Filters: input.Filters}
	table, err := u.repo.BuildReport(ctx, query, u.cfg.MaxRows+1)
	if err != nil {
		return domain.ReportExport{}, fmt.Errorf("%w: build report: %w", ErrReportStorageUnavailable, err)
	}
	if len(table.Rows) > u.cfg.MaxRows {
		return domain.ReportExport{}, domain.NewFieldError("report", "row limit exceeded; narrow the date range")
	}
	data, err := encodeSafeCSV(table, u.cfg.MaxBytes)
	if err != nil {
		return domain.ReportExport{}, err
	}
	now := u.clock.Now().UTC()
	if err := u.repo.CreateAuditEvent(ctx, domain.AuditEvent{EventID: newID("audit"), ActorID: input.Actor.ID, Action: "session_analytics.report_exported", ResourceType: "analytics_report", ResourceID: string(reportType), RequestID: input.RequestID, CreatedAt: now, Metadata: map[string]any{"format": format, "rows": len(table.Rows), "bytes": len(data)}}); err != nil {
		return domain.ReportExport{}, fmt.Errorf("%w: write export audit: %w", ErrReportStorageUnavailable, err)
	}
	return domain.ReportExport{Filename: fmt.Sprintf("%s_%s_%s.csv", reportType, reportDateLabel(input.Filters.From, input.From), reportDateLabel(input.Filters.To, input.To)), ContentType: "text/csv; charset=utf-8", Data: data}, nil
}

func reportDateLabel(raw string, fallback time.Time) string {
	if parsed, err := time.Parse("2006-01-02", strings.TrimSpace(raw)); err == nil {
		return parsed.Format("2006-01-02")
	}
	return fallback.UTC().Format("2006-01-02")
}

func (u *ReportsUsecase) ListSchedules(ctx context.Context, actor domain.Actor) ([]domain.ReportSchedule, error) {
	if err := requireAnyRole(actor, analyticsAdminRoles...); err != nil {
		return nil, err
	}
	items, err := u.repo.ListSchedules(ctx, u.cfg.ListLimit)
	if err != nil {
		return nil, fmt.Errorf("%w: list schedules: %w", ErrReportStorageUnavailable, err)
	}
	if items == nil {
		items = []domain.ReportSchedule{}
	}
	return items, nil
}

func (u *ReportsUsecase) CreateSchedule(ctx context.Context, actor domain.Actor, input domain.CreateReportSchedule, requestID string) (domain.ReportSchedule, error) {
	if err := requireAnyRole(actor, analyticsAdminRoles...); err != nil {
		return domain.ReportSchedule{}, err
	}
	normalized, err := normalizeScheduleInput(input)
	if err != nil {
		return domain.ReportSchedule{}, err
	}
	now := u.clock.Now().UTC()
	next, err := nextReportRun(normalized, now)
	if err != nil {
		return domain.ReportSchedule{}, err
	}
	schedule := domain.ReportSchedule{ID: newID("rpt_sch"), Name: normalized.Name, ReportType: normalized.ReportType, Format: normalized.Format, Frequency: normalized.Frequency, Timezone: normalized.Timezone, TimeOfDay: normalized.TimeOfDay, DayOfWeek: normalized.DayOfWeek, DayOfMonth: normalized.DayOfMonth, Status: domain.ReportScheduleActive, Recipients: normalized.Recipients, Filters: normalized.Filters, NextRunAt: &next, CreatedBy: actor.ID, CreatedAt: now, UpdatedAt: now}
	if err := u.repo.CreateSchedule(ctx, schedule); err != nil {
		return domain.ReportSchedule{}, fmt.Errorf("%w: create schedule: %w", ErrReportStorageUnavailable, err)
	}
	if err := u.repo.CreateAuditEvent(ctx, domain.AuditEvent{EventID: newID("audit"), ActorID: actor.ID, Action: "session_analytics.report_schedule_created", ResourceType: "report_schedule", ResourceID: schedule.ID, RequestID: requestID, CreatedAt: now}); err != nil {
		return domain.ReportSchedule{}, fmt.Errorf("%w: write schedule audit: %w", ErrReportStorageUnavailable, err)
	}
	return schedule, nil
}

func (u *ReportsUsecase) UpdateScheduleStatus(ctx context.Context, input UpdateReportScheduleStatusInput) (domain.ReportSchedule, error) {
	if err := requireAnyRole(input.Actor, analyticsAdminRoles...); err != nil {
		return domain.ReportSchedule{}, err
	}
	id := strings.TrimSpace(input.ID)
	if id == "" {
		return domain.ReportSchedule{}, domain.NewFieldError("id", "is required")
	}
	status := domain.ReportScheduleStatus(strings.TrimSpace(input.Status))
	if status != domain.ReportScheduleActive && status != domain.ReportSchedulePaused {
		return domain.ReportSchedule{}, domain.NewFieldError("status", "must be active or paused")
	}
	now := u.clock.Now().UTC()
	var next *time.Time
	if status == domain.ReportScheduleActive {
		existing, err := u.repo.GetSchedule(ctx, id)
		if err != nil {
			return domain.ReportSchedule{}, err
		}
		value, err := nextReportRun(domain.CreateReportSchedule{Name: existing.Name, ReportType: existing.ReportType, Format: existing.Format, Frequency: existing.Frequency, Timezone: existing.Timezone, TimeOfDay: existing.TimeOfDay, DayOfWeek: existing.DayOfWeek, DayOfMonth: existing.DayOfMonth, Recipients: existing.Recipients, Filters: existing.Filters}, now)
		if err != nil {
			return domain.ReportSchedule{}, err
		}
		next = &value
	}
	out, err := u.repo.UpdateScheduleStatus(ctx, id, status, next, now)
	if err != nil {
		return domain.ReportSchedule{}, err
	}
	_ = u.repo.CreateAuditEvent(ctx, domain.AuditEvent{EventID: newID("audit"), ActorID: input.Actor.ID, Action: "session_analytics.report_schedule_status_updated", ResourceType: "report_schedule", ResourceID: id, RequestID: input.RequestID, CreatedAt: now, Metadata: map[string]any{"status": status}})
	return out, nil
}

func (u *ReportsUsecase) DeleteSchedule(ctx context.Context, actor domain.Actor, id string, requestID string) error {
	if err := requireAnyRole(actor, analyticsAdminRoles...); err != nil {
		return err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return domain.NewFieldError("id", "is required")
	}
	if err := u.repo.DeleteSchedule(ctx, id); err != nil {
		return err
	}
	return u.repo.CreateAuditEvent(ctx, domain.AuditEvent{EventID: newID("audit"), ActorID: actor.ID, Action: "session_analytics.report_schedule_deleted", ResourceType: "report_schedule", ResourceID: id, RequestID: requestID, CreatedAt: u.clock.Now().UTC()})
}

var reportTimePattern = regexp.MustCompile(`^(?:[01]\d|2[0-3]):[0-5]\d$`)
var reportDays = []domain.ReportDayOfWeek{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}

func normalizeScheduleInput(input domain.CreateReportSchedule) (domain.CreateReportSchedule, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Timezone = strings.TrimSpace(input.Timezone)
	input.TimeOfDay = strings.TrimSpace(input.TimeOfDay)
	if len(input.Name) < 3 || len(input.Name) > 80 {
		return input, domain.NewFieldError("name", "must be between 3 and 80 characters")
	}
	if !input.ReportType.Valid() {
		return input, domain.NewFieldError("reportType", "unsupported report type")
	}
	if input.Format != domain.ReportFormatCSV {
		return input, domain.NewFieldError("format", "only csv is supported")
	}
	if !input.Frequency.Valid() {
		return input, domain.NewFieldError("frequency", "must be daily, weekly, or monthly")
	}
	zone, err := normalizeTimezone(input.Timezone)
	if err != nil {
		return input, err
	}
	input.Timezone = zone
	if !reportTimePattern.MatchString(input.TimeOfDay) {
		return input, domain.NewFieldError("timeOfDay", "must use HH:mm")
	}
	if input.Frequency == domain.ReportFrequencyWeekly && !slices.Contains(reportDays, input.DayOfWeek) {
		return input, domain.NewFieldError("dayOfWeek", "is required for weekly schedules")
	}
	if input.Frequency == domain.ReportFrequencyMonthly && (input.DayOfMonth < 1 || input.DayOfMonth > 28) {
		return input, domain.NewFieldError("dayOfMonth", "must be between 1 and 28")
	}
	if len(input.Recipients) < 1 || len(input.Recipients) > 20 {
		return input, domain.NewFieldError("recipients", "must contain between 1 and 20 addresses")
	}
	seen := map[string]struct{}{}
	normalized := make([]string, 0, len(input.Recipients))
	for _, raw := range input.Recipients {
		address := strings.ToLower(strings.TrimSpace(raw))
		parsed, parseErr := mail.ParseAddress(address)
		if parseErr != nil || parsed.Address != address {
			return input, domain.NewFieldError("recipients", "contains an invalid email address")
		}
		if _, ok := seen[address]; !ok {
			seen[address] = struct{}{}
			normalized = append(normalized, address)
		}
	}
	input.Recipients = normalized
	return input, nil
}

func normalizeTimezone(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "UTC"
	}
	if _, err := time.LoadLocation(value); err != nil {
		return "", domain.NewFieldError("timezone", "is invalid")
	}
	return value, nil
}

func nextReportRun(input domain.CreateReportSchedule, now time.Time) (time.Time, error) {
	loc, err := time.LoadLocation(input.Timezone)
	if err != nil {
		return time.Time{}, err
	}
	parts := strings.Split(input.TimeOfDay, ":")
	hour := 0
	minute := 0
	_, _ = fmt.Sscanf(parts[0], "%d", &hour)
	_, _ = fmt.Sscanf(parts[1], "%d", &minute)
	localNow := now.In(loc)
	candidate := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), hour, minute, 0, 0, loc)
	switch input.Frequency {
	case domain.ReportFrequencyDaily:
		if !candidate.After(localNow) {
			candidate = candidate.AddDate(0, 0, 1)
		}
	case domain.ReportFrequencyWeekly:
		target := slices.Index(reportDays, input.DayOfWeek) + 1
		current := int(localNow.Weekday())
		if current == 0 {
			current = 7
		}
		days := (target - current + 7) % 7
		candidate = candidate.AddDate(0, 0, days)
		if !candidate.After(localNow) {
			candidate = candidate.AddDate(0, 0, 7)
		}
	case domain.ReportFrequencyMonthly:
		candidate = time.Date(localNow.Year(), localNow.Month(), input.DayOfMonth, hour, minute, 0, 0, loc)
		if !candidate.After(localNow) {
			candidate = candidate.AddDate(0, 1, 0)
		}
	}
	return candidate.UTC(), nil
}

func encodeSafeCSV(table domain.TabularReport, maxBytes int) ([]byte, error) {
	if len(table.Headers) == 0 {
		return nil, domain.NewFieldError("report", "has no columns")
	}
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	if err := writer.Write(safeCSVRow(table.Headers)); err != nil {
		return nil, err
	}
	for _, row := range table.Rows {
		if len(row) != len(table.Headers) {
			return nil, domain.NewFieldError("report", "contains an invalid row")
		}
		if err := writer.Write(safeCSVRow(row)); err != nil {
			return nil, err
		}
		writer.Flush()
		if buffer.Len() > maxBytes {
			return nil, domain.NewFieldError("report", "byte limit exceeded; narrow the date range")
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func safeCSVRow(row []string) []string {
	out := make([]string, len(row))
	for i, value := range row {
		value = strings.ReplaceAll(strings.ReplaceAll(value, "\r", " "), "\n", " ")
		trimmed := strings.TrimLeft(value, " \t")
		if trimmed != "" && strings.ContainsRune("=+-@", rune(trimmed[0])) {
			value = "'" + value
		}
		out[i] = value
	}
	return out
}
