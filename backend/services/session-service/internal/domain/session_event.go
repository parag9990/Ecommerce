package domain

import (
	"fmt"
	"strings"
	"time"
)

const (
	CurrentSessionEventSchemaVersion = 1

	DefaultMaxEventTypeLength        = 64
	DefaultMaxEventPropertyKeyLength = 128
	DefaultMaxEventStringLength      = 2048
	DefaultMaxEventProperties        = 50
	DefaultMaxEventPropertyDepth     = 6
)

type EventType string

const (
	EventPageView      EventType = "page_view"
	EventProductView   EventType = "product_view"
	EventSearch        EventType = "search"
	EventClick         EventType = "click"
	EventScroll        EventType = "scroll"
	EventAddToCart     EventType = "add_to_cart"
	EventCheckoutStep  EventType = "checkout_step"
	EventPaymentResult EventType = "payment_result"
)

func (t EventType) SupportedForIngestion() bool {
	switch t {
	case EventPageView,
		EventProductView,
		EventSearch,
		EventClick,
		EventScroll,
		EventAddToCart,
		EventCheckoutStep,
		EventPaymentResult:
		return true
	default:
		return false
	}
}

type SessionEvent struct {
	EventID               string         `json:"event_id" bson:"event_id"`
	SessionID             string         `json:"session_id" bson:"session_id"`
	AnonymousID           string         `json:"anonymous_id" bson:"anonymous_id"`
	UserID                *string        `json:"user_id,omitempty" bson:"user_id,omitempty"`
	EventType             EventType      `json:"event_type" bson:"event_type"`
	Path                  *string        `json:"path,omitempty" bson:"path,omitempty"`
	Properties            map[string]any `json:"properties,omitempty" bson:"properties,omitempty"`
	UserAgentHash         *string        `json:"user_agent_hash,omitempty" bson:"user_agent_hash,omitempty"`
	IPHash                *string        `json:"ip_hash,omitempty" bson:"ip_hash,omitempty"`
	IPVersion             IPVersion      `json:"ip_version,omitempty" bson:"ip_version,omitempty"`
	DeviceFingerprintHash *string        `json:"device_fingerprint_hash,omitempty" bson:"device_fingerprint_hash,omitempty"`
	Device                *Device        `json:"device,omitempty" bson:"device,omitempty"`
	Client                *Client        `json:"client,omitempty" bson:"client,omitempty"`
	Geo                   *Geo           `json:"geo,omitempty" bson:"geo,omitempty"`
	OccurredAt            time.Time      `json:"occurred_at" bson:"occurred_at"`
	ReceivedAt            time.Time      `json:"received_at" bson:"received_at"`
	SchemaVersion         int            `json:"schema_version" bson:"schema_version"`
	RequestID             *string        `json:"request_id,omitempty" bson:"request_id,omitempty"`
	RetainUntil           *time.Time     `json:"retain_until,omitempty" bson:"retain_until,omitempty"`
	LegalHold             bool           `json:"legal_hold,omitempty" bson:"legal_hold,omitempty"`
	RetentionClass        RetentionClass `json:"retention_class,omitempty" bson:"retention_class,omitempty"`
}

type SessionEventValidationConfig struct {
	MaxIDLength             int
	MaxEventTypeLength      int
	MaxPageLength           int
	MaxPropertyKeyLength    int
	MaxPropertyStringLength int
	MaxProperties           int
	MaxPropertyDepth        int
}

func DefaultSessionEventValidationConfig() SessionEventValidationConfig {
	return SessionEventValidationConfig{
		MaxIDLength:             DefaultMaxIDLength,
		MaxEventTypeLength:      DefaultMaxEventTypeLength,
		MaxPageLength:           DefaultMaxPageLength,
		MaxPropertyKeyLength:    DefaultMaxEventPropertyKeyLength,
		MaxPropertyStringLength: DefaultMaxEventStringLength,
		MaxProperties:           DefaultMaxEventProperties,
		MaxPropertyDepth:        DefaultMaxEventPropertyDepth,
	}
}

func (c SessionEventValidationConfig) WithDefaults() SessionEventValidationConfig {
	defaults := DefaultSessionEventValidationConfig()
	if c.MaxIDLength <= 0 {
		c.MaxIDLength = defaults.MaxIDLength
	}
	if c.MaxEventTypeLength <= 0 {
		c.MaxEventTypeLength = defaults.MaxEventTypeLength
	}
	if c.MaxPageLength <= 0 {
		c.MaxPageLength = defaults.MaxPageLength
	}
	if c.MaxPropertyKeyLength <= 0 {
		c.MaxPropertyKeyLength = defaults.MaxPropertyKeyLength
	}
	if c.MaxPropertyStringLength <= 0 {
		c.MaxPropertyStringLength = defaults.MaxPropertyStringLength
	}
	if c.MaxProperties <= 0 {
		c.MaxProperties = defaults.MaxProperties
	}
	if c.MaxPropertyDepth <= 0 {
		c.MaxPropertyDepth = defaults.MaxPropertyDepth
	}
	return c
}

type SessionEventValidationError struct {
	Violations []FieldViolation
}

func (e SessionEventValidationError) Error() string {
	if len(e.Violations) == 0 {
		return ErrInvalidSessionEvent.Error()
	}
	parts := make([]string, 0, len(e.Violations))
	for _, violation := range e.Violations {
		parts = append(parts, fmt.Sprintf("%s: %s", violation.Field, violation.Message))
	}
	return ErrInvalidSessionEvent.Error() + ": " + strings.Join(parts, "; ")
}

func (e SessionEventValidationError) Is(target error) bool {
	return target == ErrInvalidSessionEvent
}

func (e SessionEvent) Normalize() SessionEvent {
	out := e
	out.EventID = strings.TrimSpace(out.EventID)
	out.SessionID = strings.TrimSpace(out.SessionID)
	out.AnonymousID = strings.TrimSpace(out.AnonymousID)
	out.UserID = trimStringPtr(out.UserID)
	out.EventType = EventType(strings.TrimSpace(string(out.EventType)))
	out.Path = trimStringPtr(out.Path)
	out.Properties = normalizeProperties(out.Properties)
	out.UserAgentHash = trimStringPtr(out.UserAgentHash)
	out.IPHash = trimStringPtr(out.IPHash)
	out.IPVersion = IPVersion(strings.TrimSpace(string(out.IPVersion)))
	out.DeviceFingerprintHash = trimStringPtr(out.DeviceFingerprintHash)
	out.Device = normalizeDevicePtr(out.Device)
	out.Client = normalizeClientPtr(out.Client)
	out.Geo = normalizeGeoPtr(out.Geo)
	out.OccurredAt = normalizeTime(out.OccurredAt)
	out.ReceivedAt = normalizeTime(out.ReceivedAt)
	out.RequestID = trimStringPtr(out.RequestID)
	out.RetainUntil = normalizeTimePtr(out.RetainUntil)
	out.RetentionClass = RetentionClass(strings.TrimSpace(string(out.RetentionClass)))
	return out
}

func (e SessionEvent) Validate() error {
	return e.ValidateWithConfig(DefaultSessionEventValidationConfig())
}

func (e SessionEvent) ValidateWithConfig(cfg SessionEventValidationConfig) error {
	cfg = cfg.WithDefaults()
	event := e.Normalize()
	violations := make([]FieldViolation, 0)

	validateID(&violations, "event_id", event.EventID, true, cfg.MaxIDLength)
	validateID(&violations, "session_id", event.SessionID, true, cfg.MaxIDLength)
	validateID(&violations, "anonymous_id", event.AnonymousID, true, cfg.MaxIDLength)
	validateOptionalID(&violations, "user_id", event.UserID, cfg.MaxIDLength)
	validateOptionalID(&violations, "request_id", event.RequestID, cfg.MaxIDLength)
	validateEventType(&violations, event.EventType, cfg.MaxEventTypeLength)
	validateOptionalPagePath(&violations, "path", event.Path, cfg.MaxPageLength)
	validateHashPtr(&violations, "user_agent_hash", event.UserAgentHash, DefaultMaxHashLength)
	validateHashPtr(&violations, "ip_hash", event.IPHash, DefaultMaxHashLength)
	if event.IPVersion != "" && !event.IPVersion.Valid() {
		addViolation(&violations, "ip_version", "must be ipv4, ipv6, or unknown")
	}
	validateHashPtr(&violations, "device_fingerprint_hash", event.DeviceFingerprintHash, DefaultMaxHashLength)
	if event.Device != nil {
		validateDevice(&violations, *event.Device, DefaultValidationConfig())
	}
	if event.Client != nil {
		validateClient(&violations, *event.Client, DefaultValidationConfig())
	}
	if event.Geo != nil {
		validateGeo(&violations, *event.Geo, DefaultValidationConfig())
	}
	validateEventTimestamps(&violations, event)
	validateEventProperties(&violations, "properties", event.Properties, cfg)
	validateOptionalRetentionMetadata(&violations, event.RetentionClass, event.LegalHold, nil, nil)

	if event.SchemaVersion != CurrentSessionEventSchemaVersion {
		addViolation(&violations, "schema_version", "must match current session event schema version")
	}
	if len(violations) > 0 {
		return SessionEventValidationError{Violations: violations}
	}
	return nil
}

func validateEventType(violations *[]FieldViolation, eventType EventType, maxLength int) {
	value := strings.TrimSpace(string(eventType))
	if value == "" {
		addViolation(violations, "event_type", "is required")
		return
	}
	if len(value) > maxLength {
		addViolation(violations, "event_type", "is too long")
		return
	}
	for _, r := range value {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		switch r {
		case '_', '-', '.':
			continue
		default:
			addViolation(violations, "event_type", "must use lowercase letters, numbers, underscore, dash, or dot")
			return
		}
	}
}

func validateEventTimestamps(violations *[]FieldViolation, event SessionEvent) {
	if event.OccurredAt.IsZero() {
		addViolation(violations, "occurred_at", "is required")
	}
	if event.ReceivedAt.IsZero() {
		addViolation(violations, "received_at", "is required")
	}
	if !event.OccurredAt.IsZero() && !event.ReceivedAt.IsZero() && event.ReceivedAt.Before(event.OccurredAt.Add(-24*time.Hour)) {
		addViolation(violations, "received_at", "cannot be more than 24 hours before occurred_at")
	}
}

func validateEventProperties(violations *[]FieldViolation, field string, value any, cfg SessionEventValidationConfig) {
	counter := 0
	validatePropertyValue(violations, field, value, cfg, 0, &counter)
}

func validatePropertyValue(violations *[]FieldViolation, field string, value any, cfg SessionEventValidationConfig, depth int, counter *int) {
	if value == nil {
		return
	}
	if depth > cfg.MaxPropertyDepth {
		addViolation(violations, field, "is nested too deeply")
		return
	}
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			(*counter)++
			if *counter > cfg.MaxProperties {
				addViolation(violations, field, "has too many properties")
				return
			}
			propertyField := field + "." + key
			validatePropertyKey(violations, propertyField, key, cfg)
			validatePropertyValue(violations, propertyField, nested, cfg, depth+1, counter)
		}
	case []any:
		for i, nested := range typed {
			validatePropertyValue(violations, fmt.Sprintf("%s[%d]", field, i), nested, cfg, depth+1, counter)
		}
	case string:
		validateText(violations, field, typed, cfg.MaxPropertyStringLength)
	}
}

func validatePropertyKey(violations *[]FieldViolation, field string, key string, cfg SessionEventValidationConfig) {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		addViolation(violations, field, "property key is required")
		return
	}
	if len(trimmed) > cfg.MaxPropertyKeyLength {
		addViolation(violations, field, "property key is too long")
	}
	if containsControl(trimmed) {
		addViolation(violations, field, "property key contains control characters")
	}
	if sensitivePropertyName(trimmed) {
		addViolation(violations, field, "sensitive property keys are not allowed")
	}
}

func sensitivePropertyName(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")
	switch normalized {
	case "password", "otp", "card_number", "cvv", "pin", "token", "authorization", "cookie", "private_message", "raw_ip":
		return true
	default:
		return false
	}
}

func normalizeProperties(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for key, value := range in {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		out[key] = normalizePropertyValue(value)
	}
	return out
}

func normalizeDevicePtr(value *Device) *Device {
	if value == nil {
		return nil
	}
	normalized := value.Normalize()
	if normalized.Type == "" &&
		normalized.Browser == nil &&
		normalized.BrowserVersion == nil &&
		normalized.OS == nil &&
		normalized.OSVersion == nil &&
		normalized.Model == nil &&
		normalized.Vendor == nil &&
		!normalized.IsBot {
		return nil
	}
	if normalized.Type == "" {
		normalized.Type = DeviceTypeUnknown
	}
	return &normalized
}

func normalizeClientPtr(value *Client) *Client {
	if value == nil {
		return nil
	}
	normalized := value.Normalize()
	if normalized.Channel == ChannelUnknown &&
		normalized.Locale == nil &&
		normalized.Timezone == nil &&
		normalized.ScreenWidth == 0 &&
		normalized.ScreenHeight == 0 &&
		normalized.ViewportWidth == 0 &&
		normalized.ViewportHeight == 0 {
		return nil
	}
	return &normalized
}

func normalizeGeoPtr(value *Geo) *Geo {
	if value == nil {
		return nil
	}
	normalized := value.Normalize()
	if normalized.Source == GeoSourceUnknown &&
		normalized.Country == nil &&
		normalized.Region == nil &&
		normalized.City == nil &&
		normalized.Timezone == nil {
		return nil
	}
	return &normalized
}

func normalizePropertyValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return normalizeProperties(typed)
	case []any:
		out := make([]any, len(typed))
		for i, nested := range typed {
			out[i] = normalizePropertyValue(nested)
		}
		return out
	case string:
		return strings.TrimSpace(typed)
	default:
		return typed
	}
}
