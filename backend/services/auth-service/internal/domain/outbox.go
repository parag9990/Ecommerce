package domain

import "time"

type OutboxStatus string

const (
	OutboxStatusPending    OutboxStatus = "pending"
	OutboxStatusProcessing OutboxStatus = "processing"
	OutboxStatusPublished  OutboxStatus = "published"
	OutboxStatusFailed     OutboxStatus = "failed"
	OutboxStatusDeadLetter OutboxStatus = "dead_letter"
)

type OutboxEvent struct {
	EventID       string
	EventType     string
	Version       int
	AggregateType string
	AggregateID   string
	RoutingKey    string
	Payload       []byte
	TraceID       string
	Status        OutboxStatus
	Attempts      int
	NextAttemptAt *time.Time
	PublishedAt   *time.Time
	LockedAt      *time.Time
	LockedBy      string
	LastError     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
