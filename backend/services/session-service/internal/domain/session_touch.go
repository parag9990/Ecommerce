package domain

import (
	"fmt"
	"strings"
	"time"
)

type SessionTouch struct {
	SessionID             string     `json:"session_id" bson:"session_id"`
	AnonymousID           string     `json:"anonymous_id" bson:"anonymous_id"`
	UserID                *string    `json:"user_id,omitempty" bson:"user_id,omitempty"`
	EventType             EventType  `json:"event_type" bson:"event_type"`
	Path                  *string    `json:"path,omitempty" bson:"path,omitempty"`
	UserAgent             string     `json:"user_agent" bson:"user_agent"`
	UserAgentHash         *string    `json:"user_agent_hash,omitempty" bson:"user_agent_hash,omitempty"`
	IPHash                string     `json:"ip_hash" bson:"ip_hash"`
	IPVersion             IPVersion  `json:"ip_version" bson:"ip_version"`
	Channel               Channel    `json:"channel" bson:"channel"`
	DeviceType            DeviceType `json:"device_type" bson:"device_type"`
	Device                Device     `json:"device" bson:"device"`
	Client                Client     `json:"client" bson:"client"`
	Geo                   Geo        `json:"geo" bson:"geo"`
	DeviceFingerprintHash *string    `json:"device_fingerprint_hash,omitempty" bson:"device_fingerprint_hash,omitempty"`
	OccurredAt            time.Time  `json:"occurred_at" bson:"occurred_at"`
	ReceivedAt            time.Time  `json:"received_at" bson:"received_at"`
	SchemaVersion         int        `json:"schema_version" bson:"schema_version"`
}

type SessionTouchValidationError struct {
	Violations []FieldViolation
}

func (e SessionTouchValidationError) Error() string {
	if len(e.Violations) == 0 {
		return ErrInvalidSession.Error()
	}
	parts := make([]string, 0, len(e.Violations))
	for _, violation := range e.Violations {
		parts = append(parts, fmt.Sprintf("%s: %s", violation.Field, violation.Message))
	}
	return ErrInvalidSession.Error() + ": " + strings.Join(parts, "; ")
}

func (e SessionTouchValidationError) Is(target error) bool {
	return target == ErrInvalidSession
}

func (t SessionTouch) Normalize() SessionTouch {
	out := t
	out.SessionID = strings.TrimSpace(out.SessionID)
	out.AnonymousID = strings.TrimSpace(out.AnonymousID)
	out.UserID = trimStringPtr(out.UserID)
	out.EventType = EventType(strings.TrimSpace(string(out.EventType)))
	out.Path = trimStringPtr(out.Path)
	out.UserAgent = strings.TrimSpace(out.UserAgent)
	out.UserAgentHash = trimStringPtr(out.UserAgentHash)
	out.IPHash = strings.TrimSpace(out.IPHash)
	out.IPVersion = IPVersion(strings.TrimSpace(string(out.IPVersion)))
	out.Channel = Channel(strings.TrimSpace(string(out.Channel)))
	out.DeviceType = DeviceType(strings.TrimSpace(string(out.DeviceType)))
	out.Device = out.Device.Normalize()
	if out.Device.Type == "" {
		out.Device.Type = out.DeviceType
	}
	if out.Device.Type == "" {
		out.Device.Type = DeviceTypeUnknown
	}
	out.DeviceType = out.Device.Type
	out.Client = out.Client.Normalize()
	if out.Client.Channel == ChannelUnknown && out.Channel != "" {
		out.Client.Channel = out.Channel
	}
	out.Geo = out.Geo.Normalize()
	out.DeviceFingerprintHash = trimStringPtr(out.DeviceFingerprintHash)
	out.OccurredAt = normalizeTime(out.OccurredAt)
	out.ReceivedAt = normalizeTime(out.ReceivedAt)
	return out
}

func (t SessionTouch) Validate() error {
	cfg := DefaultValidationConfig()
	touch := t.Normalize()
	violations := make([]FieldViolation, 0)

	validateID(&violations, "session_id", touch.SessionID, true, cfg.MaxIDLength)
	validateID(&violations, "anonymous_id", touch.AnonymousID, true, cfg.MaxIDLength)
	validateOptionalID(&violations, "user_id", touch.UserID, cfg.MaxIDLength)
	validateEventType(&violations, touch.EventType, DefaultMaxEventTypeLength)
	validateOptionalPagePath(&violations, "path", touch.Path, cfg.MaxPageLength)
	validateRequiredText(&violations, "user_agent", touch.UserAgent, cfg.MaxUserAgentLength)
	validateHashPtr(&violations, "user_agent_hash", touch.UserAgentHash, cfg.MaxHashLength)
	validateHash(&violations, "ip_hash", touch.IPHash, true, cfg.MaxHashLength)
	if touch.IPVersion != "" && !touch.IPVersion.Valid() {
		addViolation(&violations, "ip_version", "must be ipv4, ipv6, or unknown")
	}
	if !touch.Channel.Valid() {
		addViolation(&violations, "channel", "must be a supported channel")
	}
	if !touch.DeviceType.Valid() {
		addViolation(&violations, "device_type", "must be desktop, mobile, tablet, bot, or unknown")
	}
	validateDevice(&violations, touch.Device, cfg)
	validateClient(&violations, touch.Client, cfg)
	validateGeo(&violations, touch.Geo, cfg)
	validateHashPtr(&violations, "device_fingerprint_hash", touch.DeviceFingerprintHash, cfg.MaxHashLength)
	if touch.OccurredAt.IsZero() {
		addViolation(&violations, "occurred_at", "is required")
	}
	if touch.ReceivedAt.IsZero() {
		addViolation(&violations, "received_at", "is required")
	}
	if touch.SchemaVersion != CurrentSessionSchemaVersion {
		addViolation(&violations, "schema_version", "must match current session schema version")
	}

	if len(violations) > 0 {
		return SessionTouchValidationError{Violations: violations}
	}
	return nil
}

func ActiveSessionSnapshotFromTouch(touch SessionTouch) ActiveSessionSnapshot {
	normalized := touch.Normalize()
	entryPage := "/"
	if normalized.Path != nil {
		entryPage = *normalized.Path
	}
	lastEventType := normalized.EventType
	return ActiveSessionSnapshot{
		SessionID:     normalized.SessionID,
		AnonymousID:   normalized.AnonymousID,
		UserID:        normalized.UserID,
		Status:        SessionStatusActive,
		EntryPage:     entryPage,
		CurrentPage:   normalized.Path,
		DeviceType:    normalized.DeviceType,
		Browser:       normalized.Device.Browser,
		OS:            normalized.Device.OS,
		Country:       normalized.Geo.Country,
		City:          normalized.Geo.City,
		Channel:       normalized.Channel,
		IPHash:        normalized.IPHash,
		LastEventType: &lastEventType,
		StartedAt:     normalized.OccurredAt,
		LastSeenAt:    normalized.OccurredAt,
	}
}
