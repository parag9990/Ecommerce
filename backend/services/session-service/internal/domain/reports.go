package domain

import "time"

type AnalyticsReportType string

const (
	ReportOverview       AnalyticsReportType = "overview"
	ReportActiveSessions AnalyticsReportType = "active_sessions"
	ReportJourneySummary AnalyticsReportType = "journey_summary"
	ReportFunnel         AnalyticsReportType = "funnel"
	ReportHeatmap        AnalyticsReportType = "heatmap"
	ReportRetention      AnalyticsReportType = "retention"
)

func (t AnalyticsReportType) Valid() bool {
	switch t {
	case ReportOverview, ReportActiveSessions, ReportJourneySummary, ReportFunnel, ReportHeatmap, ReportRetention:
		return true
	default:
		return false
	}
}

type ReportFormat string

const ReportFormatCSV ReportFormat = "csv"

type ReportFrequency string

const (
	ReportFrequencyDaily   ReportFrequency = "daily"
	ReportFrequencyWeekly  ReportFrequency = "weekly"
	ReportFrequencyMonthly ReportFrequency = "monthly"
)

func (f ReportFrequency) Valid() bool {
	return f == ReportFrequencyDaily || f == ReportFrequencyWeekly || f == ReportFrequencyMonthly
}

type ReportScheduleStatus string

const (
	ReportScheduleActive ReportScheduleStatus = "active"
	ReportSchedulePaused ReportScheduleStatus = "paused"
	ReportScheduleFailed ReportScheduleStatus = "failed"
)

type ReportDayOfWeek string

type ReportFilters struct {
	From       string `json:"from,omitempty" bson:"from,omitempty"`
	To         string `json:"to,omitempty" bson:"to,omitempty"`
	Timezone   string `json:"timezone,omitempty" bson:"timezone,omitempty"`
	Channel    string `json:"channel,omitempty" bson:"channel,omitempty"`
	Country    string `json:"country,omitempty" bson:"country,omitempty"`
	DeviceType string `json:"deviceType,omitempty" bson:"device_type,omitempty"`
	Source     string `json:"source,omitempty" bson:"source,omitempty"`
	UserType   string `json:"userType,omitempty" bson:"user_type,omitempty"`
}

type ReportQuery struct {
	ReportType AnalyticsReportType
	Format     ReportFormat
	From       time.Time
	To         time.Time
	Timezone   string
	Filters    ReportFilters
}

type TabularReport struct {
	Headers []string
	Rows    [][]string
}

type ReportExport struct {
	Filename    string
	ContentType string
	Data        []byte
}

type ReportSchedule struct {
	ID         string               `json:"id" bson:"_id"`
	Name       string               `json:"name" bson:"name"`
	ReportType AnalyticsReportType  `json:"reportType" bson:"report_type"`
	Format     ReportFormat         `json:"format" bson:"format"`
	Frequency  ReportFrequency      `json:"frequency" bson:"frequency"`
	Timezone   string               `json:"timezone" bson:"timezone"`
	TimeOfDay  string               `json:"timeOfDay" bson:"time_of_day"`
	DayOfWeek  ReportDayOfWeek      `json:"dayOfWeek,omitempty" bson:"day_of_week,omitempty"`
	DayOfMonth int                  `json:"dayOfMonth,omitempty" bson:"day_of_month,omitempty"`
	Status     ReportScheduleStatus `json:"status" bson:"status"`
	Recipients []string             `json:"recipients" bson:"recipients"`
	Filters    ReportFilters        `json:"filters" bson:"filters"`
	LastRunAt  *time.Time           `json:"lastRunAt,omitempty" bson:"last_run_at,omitempty"`
	NextRunAt  *time.Time           `json:"nextRunAt,omitempty" bson:"next_run_at,omitempty"`
	LastError  string               `json:"lastError,omitempty" bson:"last_error,omitempty"`
	CreatedBy  string               `json:"-" bson:"created_by"`
	CreatedAt  time.Time            `json:"createdAt" bson:"created_at"`
	UpdatedAt  time.Time            `json:"updatedAt,omitempty" bson:"updated_at"`
}

type CreateReportSchedule struct {
	Name       string
	ReportType AnalyticsReportType
	Format     ReportFormat
	Frequency  ReportFrequency
	Timezone   string
	TimeOfDay  string
	DayOfWeek  ReportDayOfWeek
	DayOfMonth int
	Recipients []string
	Filters    ReportFilters
}
