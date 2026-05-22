package domain

import (
	"fmt"
	"strings"
	"time"
)

type ActiveSessionSnapshot struct {
	SessionID     string        `json:"session_id" bson:"session_id"`
	AnonymousID   string        `json:"anonymous_id" bson:"anonymous_id"`
	UserID        *string       `json:"user_id,omitempty" bson:"user_id,omitempty"`
	Status        SessionStatus `json:"status" bson:"status"`
	EntryPage     string        `json:"entry_page" bson:"entry_page"`
	CurrentPage   *string       `json:"current_page,omitempty" bson:"current_page,omitempty"`
	DeviceType    DeviceType    `json:"device_type" bson:"device_type"`
	Browser       *string       `json:"browser,omitempty" bson:"browser,omitempty"`
	OS            *string       `json:"os,omitempty" bson:"os,omitempty"`
	Country       *string       `json:"country,omitempty" bson:"country,omitempty"`
	City          *string       `json:"city,omitempty" bson:"city,omitempty"`
	Channel       Channel       `json:"channel" bson:"channel"`
	IPHash        string        `json:"ip_hash" bson:"ip_hash"`
	LastEventType *EventType    `json:"last_event_type,omitempty" bson:"last_event_type,omitempty"`
	StartedAt     time.Time     `json:"started_at" bson:"started_at"`
	LastSeenAt    time.Time     `json:"last_seen_at" bson:"last_seen_at"`
}

type ActiveSessionValidationError struct {
	Violations []FieldViolation
}

func (e ActiveSessionValidationError) Error() string {
	if len(e.Violations) == 0 {
		return ErrInvalidActiveSession.Error()
	}
	parts := make([]string, 0, len(e.Violations))
	for _, violation := range e.Violations {
		parts = append(parts, fmt.Sprintf("%s: %s", violation.Field, violation.Message))
	}
	return ErrInvalidActiveSession.Error() + ": " + strings.Join(parts, "; ")
}

func (e ActiveSessionValidationError) Is(target error) bool {
	return target == ErrInvalidActiveSession
}

func (s ActiveSessionSnapshot) Normalize() ActiveSessionSnapshot {
	out := s
	out.SessionID = strings.TrimSpace(out.SessionID)
	out.AnonymousID = strings.TrimSpace(out.AnonymousID)
	out.UserID = trimStringPtr(out.UserID)
	out.Status = SessionStatus(strings.TrimSpace(string(out.Status)))
	out.EntryPage = strings.TrimSpace(out.EntryPage)
	out.CurrentPage = trimStringPtr(out.CurrentPage)
	out.DeviceType = DeviceType(strings.TrimSpace(string(out.DeviceType)))
	out.Browser = trimStringPtr(out.Browser)
	out.OS = trimStringPtr(out.OS)
	out.Country = trimStringPtr(out.Country)
	out.City = trimStringPtr(out.City)
	out.Channel = Channel(strings.TrimSpace(string(out.Channel)))
	out.IPHash = strings.TrimSpace(out.IPHash)
	out.LastEventType = trimEventTypePtr(out.LastEventType)
	out.StartedAt = normalizeTime(out.StartedAt)
	out.LastSeenAt = normalizeTime(out.LastSeenAt)
	return out
}

func (s ActiveSessionSnapshot) Validate() error {
	snapshot := s.Normalize()
	cfg := DefaultValidationConfig()
	violations := make([]FieldViolation, 0)

	validateID(&violations, "session_id", snapshot.SessionID, true, cfg.MaxIDLength)
	validateID(&violations, "anonymous_id", snapshot.AnonymousID, true, cfg.MaxIDLength)
	validateOptionalID(&violations, "user_id", snapshot.UserID, cfg.MaxIDLength)
	if !snapshot.Status.Valid() {
		addViolation(&violations, "status", "must be active, ended, expired, or revoked")
	}
	if snapshot.Status != SessionStatusActive {
		addViolation(&violations, "status", "active session snapshot must have active status")
	}
	validatePagePath(&violations, "entry_page", snapshot.EntryPage, true, cfg.MaxPageLength)
	validateOptionalPagePath(&violations, "current_page", snapshot.CurrentPage, cfg.MaxPageLength)
	if !snapshot.DeviceType.Valid() {
		addViolation(&violations, "device_type", "must be desktop, mobile, tablet, bot, or unknown")
	}
	validateOptionalText(&violations, "browser", snapshot.Browser, cfg.MaxMetadataLength)
	validateOptionalText(&violations, "os", snapshot.OS, cfg.MaxMetadataLength)
	validateOptionalText(&violations, "country", snapshot.Country, cfg.MaxGeoValueLength)
	validateOptionalText(&violations, "city", snapshot.City, cfg.MaxGeoValueLength)
	if !snapshot.Channel.Valid() {
		addViolation(&violations, "channel", "must be a supported channel")
	}
	validateHash(&violations, "ip_hash", snapshot.IPHash, true, cfg.MaxHashLength)
	if snapshot.LastEventType != nil {
		validateEventType(&violations, *snapshot.LastEventType, DefaultMaxEventTypeLength)
	}
	if snapshot.StartedAt.IsZero() {
		addViolation(&violations, "started_at", "is required")
	}
	if snapshot.LastSeenAt.IsZero() {
		addViolation(&violations, "last_seen_at", "is required")
	}
	if !snapshot.StartedAt.IsZero() && !snapshot.LastSeenAt.IsZero() && snapshot.LastSeenAt.Before(snapshot.StartedAt) {
		addViolation(&violations, "last_seen_at", "cannot be before started_at")
	}

	if len(violations) > 0 {
		return ActiveSessionValidationError{Violations: violations}
	}
	return nil
}

func ActiveSessionSnapshotFromSession(session Session, currentPage *string) ActiveSessionSnapshot {
	normalized := session.Normalize()
	return ActiveSessionSnapshot{
		SessionID:   normalized.SessionID,
		AnonymousID: normalized.AnonymousID,
		UserID:      normalized.UserID,
		Status:      normalized.Status,
		EntryPage:   normalized.EntryPage,
		CurrentPage: trimStringPtr(currentPage),
		DeviceType:  normalized.Device.Type,
		Browser:     normalized.Device.Browser,
		OS:          normalized.Device.OS,
		Country:     normalized.Geo.Country,
		City:        normalized.Geo.City,
		Channel:     normalized.Channel,
		IPHash:      normalized.IPHash,
		StartedAt:   normalized.StartedAt,
		LastSeenAt:  normalized.LastSeenAt,
	}
}

func trimEventTypePtr(value *EventType) *EventType {
	if value == nil {
		return nil
	}
	eventType := EventType(strings.TrimSpace(string(*value)))
	if eventType == "" {
		return nil
	}
	return &eventType
}
