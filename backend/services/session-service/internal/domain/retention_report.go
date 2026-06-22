package domain

import "time"

type RetentionInterval string

const (
	RetentionIntervalDay   RetentionInterval = "day"
	RetentionIntervalWeek  RetentionInterval = "week"
	RetentionIntervalMonth RetentionInterval = "month"
)

func (i RetentionInterval) Valid() bool {
	return i == RetentionIntervalDay || i == RetentionIntervalWeek || i == RetentionIntervalMonth
}

type RetentionReportFilter struct {
	From       time.Time
	To         time.Time
	Interval   RetentionInterval
	Window     int
	DeviceType DeviceType
	Channel    Channel
	Source     string
	UserType   string
}

type RetentionSummary struct {
	NewUsers         int64   `json:"newUsers"`
	ReturningUsers   int64   `json:"returningUsers"`
	ReturningRate    float64 `json:"returningRate"`
	AverageRetention float64 `json:"averageRetention"`
	BestCohort       string  `json:"bestCohort,omitempty"`
	WorstCohort      string  `json:"worstCohort,omitempty"`
}

type NewReturningBucket struct {
	Bucket         string `json:"bucket"`
	Label          string `json:"label,omitempty"`
	NewUsers       int64  `json:"newUsers"`
	ReturningUsers int64  `json:"returningUsers"`
	TotalUsers     int64  `json:"totalUsers"`
}

type RetentionBucket struct {
	Offset     int     `json:"offset"`
	Label      string  `json:"label"`
	Users      int64   `json:"users"`
	Rate       float64 `json:"rate"`
	Suppressed bool    `json:"suppressed"`
}

type RetentionCohort struct {
	CohortKey   string            `json:"cohortKey"`
	CohortLabel string            `json:"cohortLabel"`
	CohortSize  int64             `json:"cohortSize"`
	Buckets     []RetentionBucket `json:"buckets"`
	Suppressed  bool              `json:"suppressed"`
}

type RetentionReportMeta struct {
	Interval            RetentionInterval `json:"interval"`
	Window              int               `json:"window"`
	From                string            `json:"from"`
	To                  string            `json:"to"`
	GeneratedAt         time.Time         `json:"generatedAt"`
	Partial             bool              `json:"partial"`
	SmallCountThreshold int64             `json:"smallCountThreshold"`
	Suppressed          bool              `json:"suppressed"`
}

type RetentionReport struct {
	Summary        RetentionSummary     `json:"summary"`
	NewVsReturning []NewReturningBucket `json:"newVsReturning"`
	Cohorts        []RetentionCohort    `json:"cohorts"`
	Meta           RetentionReportMeta  `json:"meta"`
}
