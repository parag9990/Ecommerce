package domain

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
	"unicode"
)

const (
	CurrentSessionSchemaVersion = 1

	DefaultMaxIDLength             = 128
	DefaultMaxHashLength           = 256
	DefaultMaxPageLength           = 2048
	DefaultMaxReferrerLength       = 2048
	DefaultMaxUserAgentLength      = 1024
	DefaultMaxMetadataLength       = 128
	DefaultMaxUTMValueLength       = 256
	DefaultMaxGeoValueLength       = 128
	DefaultMaxRiskReasons          = 20
	DefaultMaxRiskReasonTextLength = 256
)

type SessionStatus string

const (
	SessionStatusActive  SessionStatus = "active"
	SessionStatusEnded   SessionStatus = "ended"
	SessionStatusExpired SessionStatus = "expired"
	SessionStatusRevoked SessionStatus = "revoked"
)

func (s SessionStatus) Valid() bool {
	switch s {
	case SessionStatusActive, SessionStatusEnded, SessionStatusExpired, SessionStatusRevoked:
		return true
	default:
		return false
	}
}

type DeviceType string

const (
	DeviceTypeDesktop DeviceType = "desktop"
	DeviceTypeMobile  DeviceType = "mobile"
	DeviceTypeTablet  DeviceType = "tablet"
	DeviceTypeBot     DeviceType = "bot"
	DeviceTypeUnknown DeviceType = "unknown"
)

func (t DeviceType) Valid() bool {
	switch t {
	case DeviceTypeDesktop, DeviceTypeMobile, DeviceTypeTablet, DeviceTypeBot, DeviceTypeUnknown:
		return true
	default:
		return false
	}
}

type IPVersion string

const (
	IPVersionIPv4    IPVersion = "ipv4"
	IPVersionIPv6    IPVersion = "ipv6"
	IPVersionUnknown IPVersion = "unknown"
)

func (v IPVersion) Valid() bool {
	switch v {
	case IPVersionIPv4, IPVersionIPv6, IPVersionUnknown:
		return true
	default:
		return false
	}
}

type Channel string

const (
	ChannelUserAppWeb                Channel = "user_app_web"
	ChannelSellerDashboardWeb        Channel = "seller_dashboard_web"
	ChannelSuperadminPanelWeb        Channel = "superadmin_panel_web"
	ChannelSessionAnalyticsDashboard Channel = "session_analytics_dashboard_web"
	ChannelMobileWeb                 Channel = "mobile_web"
	ChannelAndroidApp                Channel = "android_app"
	ChannelIOSApp                    Channel = "ios_app"
	ChannelUnknown                   Channel = "unknown"
)

func (c Channel) Valid() bool {
	switch c {
	case ChannelUserAppWeb,
		ChannelSellerDashboardWeb,
		ChannelSuperadminPanelWeb,
		ChannelSessionAnalyticsDashboard,
		ChannelMobileWeb,
		ChannelAndroidApp,
		ChannelIOSApp,
		ChannelUnknown:
		return true
	default:
		return false
	}
}

type SessionEndReason string

const (
	EndReasonLogout            SessionEndReason = "logout"
	EndReasonInactivityTimeout SessionEndReason = "inactivity_timeout"
	EndReasonSessionReplaced   SessionEndReason = "session_replaced"
	EndReasonSecurityRevoke    SessionEndReason = "security_revoke"
	EndReasonAdminRevoke       SessionEndReason = "admin_revoke"
	EndReasonUnknown           SessionEndReason = "unknown"
)

func (r SessionEndReason) Valid() bool {
	switch r {
	case EndReasonLogout,
		EndReasonInactivityTimeout,
		EndReasonSessionReplaced,
		EndReasonSecurityRevoke,
		EndReasonAdminRevoke,
		EndReasonUnknown:
		return true
	default:
		return false
	}
}

type RiskLevel string

const (
	RiskLevelLow     RiskLevel = "low"
	RiskLevelMedium  RiskLevel = "medium"
	RiskLevelHigh    RiskLevel = "high"
	RiskLevelUnknown RiskLevel = "unknown"
)

func (r RiskLevel) Valid() bool {
	switch r {
	case RiskLevelLow, RiskLevelMedium, RiskLevelHigh, RiskLevelUnknown:
		return true
	default:
		return false
	}
}

type Session struct {
	SessionID             string            `json:"session_id" bson:"session_id"`
	AnonymousID           string            `json:"anonymous_id" bson:"anonymous_id"`
	UserID                *string           `json:"user_id,omitempty" bson:"user_id,omitempty"`
	SchemaVersion         int               `json:"schema_version" bson:"schema_version"`
	Status                SessionStatus     `json:"status" bson:"status"`
	Channel               Channel           `json:"channel" bson:"channel"`
	EntryPage             string            `json:"entry_page" bson:"entry_page"`
	ExitPage              *string           `json:"exit_page,omitempty" bson:"exit_page,omitempty"`
	Referrer              *string           `json:"referrer,omitempty" bson:"referrer,omitempty"`
	UTM                   UTM               `json:"utm" bson:"utm"`
	UserAgent             string            `json:"user_agent" bson:"user_agent"`
	UserAgentHash         *string           `json:"user_agent_hash,omitempty" bson:"user_agent_hash,omitempty"`
	Device                Device            `json:"device" bson:"device"`
	Client                Client            `json:"client" bson:"client"`
	IPHash                string            `json:"ip_hash" bson:"ip_hash"`
	IPVersion             IPVersion         `json:"ip_version,omitempty" bson:"ip_version,omitempty"`
	Geo                   Geo               `json:"geo" bson:"geo"`
	DeviceFingerprintHash *string           `json:"device_fingerprint_hash,omitempty" bson:"device_fingerprint_hash,omitempty"`
	AuthSessionID         *string           `json:"auth_session_id,omitempty" bson:"auth_session_id,omitempty"`
	RiskLevel             RiskLevel         `json:"risk_level" bson:"risk_level"`
	RiskReasons           []string          `json:"risk_reasons" bson:"risk_reasons"`
	StartedAt             time.Time         `json:"started_at" bson:"started_at"`
	LastSeenAt            time.Time         `json:"last_seen_at" bson:"last_seen_at"`
	EndedAt               *time.Time        `json:"ended_at,omitempty" bson:"ended_at,omitempty"`
	RevokedAt             *time.Time        `json:"revoked_at,omitempty" bson:"revoked_at,omitempty"`
	EndReason             *SessionEndReason `json:"end_reason,omitempty" bson:"end_reason,omitempty"`
	CreatedAt             time.Time         `json:"created_at" bson:"created_at"`
	UpdatedAt             time.Time         `json:"updated_at" bson:"updated_at"`
	RetainUntil           *time.Time        `json:"retain_until,omitempty" bson:"retain_until,omitempty"`
	AnonymizedAt          *time.Time        `json:"anonymized_at,omitempty" bson:"anonymized_at,omitempty"`
	DeletionRequestedAt   *time.Time        `json:"deletion_requested_at,omitempty" bson:"deletion_requested_at,omitempty"`
	LegalHold             bool              `json:"legal_hold,omitempty" bson:"legal_hold,omitempty"`
	LegalHoldReason       *string           `json:"legal_hold_reason,omitempty" bson:"legal_hold_reason,omitempty"`
	LegalHoldSetBy        *string           `json:"legal_hold_set_by,omitempty" bson:"legal_hold_set_by,omitempty"`
	LegalHoldSetAt        *time.Time        `json:"legal_hold_set_at,omitempty" bson:"legal_hold_set_at,omitempty"`
	RetentionClass        RetentionClass    `json:"retention_class,omitempty" bson:"retention_class,omitempty"`
}

type Device struct {
	Type           DeviceType `json:"type" bson:"type"`
	Browser        *string    `json:"browser,omitempty" bson:"browser,omitempty"`
	BrowserVersion *string    `json:"browser_version,omitempty" bson:"browser_version,omitempty"`
	OS             *string    `json:"os,omitempty" bson:"os,omitempty"`
	OSVersion      *string    `json:"os_version,omitempty" bson:"os_version,omitempty"`
	Model          *string    `json:"model,omitempty" bson:"model,omitempty"`
	Vendor         *string    `json:"vendor,omitempty" bson:"vendor,omitempty"`
	IsBot          bool       `json:"is_bot" bson:"is_bot"`
}

type Geo struct {
	Country  *string   `json:"country,omitempty" bson:"country,omitempty"`
	Region   *string   `json:"region,omitempty" bson:"region,omitempty"`
	City     *string   `json:"city,omitempty" bson:"city,omitempty"`
	Timezone *string   `json:"timezone,omitempty" bson:"timezone,omitempty"`
	Source   GeoSource `json:"source" bson:"source"`
}

type GeoSource string

const (
	GeoSourceGeoIP    GeoSource = "geoip"
	GeoSourceHeader   GeoSource = "header"
	GeoSourceUnknown  GeoSource = "unknown"
	GeoSourceDisabled GeoSource = "disabled"
)

func (s GeoSource) Valid() bool {
	switch s {
	case GeoSourceGeoIP, GeoSourceHeader, GeoSourceUnknown, GeoSourceDisabled:
		return true
	default:
		return false
	}
}

type Client struct {
	Channel        Channel `json:"channel" bson:"channel"`
	Locale         *string `json:"locale,omitempty" bson:"locale,omitempty"`
	Timezone       *string `json:"timezone,omitempty" bson:"timezone,omitempty"`
	ScreenWidth    int     `json:"screen_width,omitempty" bson:"screen_width,omitempty"`
	ScreenHeight   int     `json:"screen_height,omitempty" bson:"screen_height,omitempty"`
	ViewportWidth  int     `json:"viewport_width,omitempty" bson:"viewport_width,omitempty"`
	ViewportHeight int     `json:"viewport_height,omitempty" bson:"viewport_height,omitempty"`
}

type UTM struct {
	Source   *string `json:"source,omitempty" bson:"source,omitempty"`
	Medium   *string `json:"medium,omitempty" bson:"medium,omitempty"`
	Campaign *string `json:"campaign,omitempty" bson:"campaign,omitempty"`
	Term     *string `json:"term,omitempty" bson:"term,omitempty"`
	Content  *string `json:"content,omitempty" bson:"content,omitempty"`
}

type ValidationConfig struct {
	MaxIDLength             int
	MaxHashLength           int
	MaxPageLength           int
	MaxReferrerLength       int
	MaxUserAgentLength      int
	MaxMetadataLength       int
	MaxUTMValueLength       int
	MaxGeoValueLength       int
	MaxRiskReasons          int
	MaxRiskReasonTextLength int
}

func DefaultValidationConfig() ValidationConfig {
	return ValidationConfig{
		MaxIDLength:             DefaultMaxIDLength,
		MaxHashLength:           DefaultMaxHashLength,
		MaxPageLength:           DefaultMaxPageLength,
		MaxReferrerLength:       DefaultMaxReferrerLength,
		MaxUserAgentLength:      DefaultMaxUserAgentLength,
		MaxMetadataLength:       DefaultMaxMetadataLength,
		MaxUTMValueLength:       DefaultMaxUTMValueLength,
		MaxGeoValueLength:       DefaultMaxGeoValueLength,
		MaxRiskReasons:          DefaultMaxRiskReasons,
		MaxRiskReasonTextLength: DefaultMaxRiskReasonTextLength,
	}
}

func (c ValidationConfig) WithDefaults() ValidationConfig {
	defaults := DefaultValidationConfig()
	if c.MaxIDLength <= 0 {
		c.MaxIDLength = defaults.MaxIDLength
	}
	if c.MaxHashLength <= 0 {
		c.MaxHashLength = defaults.MaxHashLength
	}
	if c.MaxPageLength <= 0 {
		c.MaxPageLength = defaults.MaxPageLength
	}
	if c.MaxReferrerLength <= 0 {
		c.MaxReferrerLength = defaults.MaxReferrerLength
	}
	if c.MaxUserAgentLength <= 0 {
		c.MaxUserAgentLength = defaults.MaxUserAgentLength
	}
	if c.MaxMetadataLength <= 0 {
		c.MaxMetadataLength = defaults.MaxMetadataLength
	}
	if c.MaxUTMValueLength <= 0 {
		c.MaxUTMValueLength = defaults.MaxUTMValueLength
	}
	if c.MaxGeoValueLength <= 0 {
		c.MaxGeoValueLength = defaults.MaxGeoValueLength
	}
	if c.MaxRiskReasons <= 0 {
		c.MaxRiskReasons = defaults.MaxRiskReasons
	}
	if c.MaxRiskReasonTextLength <= 0 {
		c.MaxRiskReasonTextLength = defaults.MaxRiskReasonTextLength
	}
	return c
}

type FieldViolation struct {
	Field   string
	Message string
}

type ValidationError struct {
	Violations []FieldViolation
}

func (e ValidationError) Error() string {
	if len(e.Violations) == 0 {
		return ErrInvalidSession.Error()
	}
	parts := make([]string, 0, len(e.Violations))
	for _, violation := range e.Violations {
		parts = append(parts, fmt.Sprintf("%s: %s", violation.Field, violation.Message))
	}
	return ErrInvalidSession.Error() + ": " + strings.Join(parts, "; ")
}

func (e ValidationError) Is(target error) bool {
	return target == ErrInvalidSession
}

func (s Session) Normalize() Session {
	out := s
	out.SessionID = strings.TrimSpace(out.SessionID)
	out.AnonymousID = strings.TrimSpace(out.AnonymousID)
	out.UserID = trimStringPtr(out.UserID)
	out.Channel = Channel(strings.TrimSpace(string(out.Channel)))
	out.Status = SessionStatus(strings.TrimSpace(string(out.Status)))
	out.EntryPage = strings.TrimSpace(out.EntryPage)
	out.ExitPage = trimStringPtr(out.ExitPage)
	out.Referrer = trimStringPtr(out.Referrer)
	out.UTM = out.UTM.Normalize()
	out.UserAgent = strings.TrimSpace(out.UserAgent)
	out.UserAgentHash = trimStringPtr(out.UserAgentHash)
	out.Device = out.Device.Normalize()
	out.Client = out.Client.Normalize()
	out.IPHash = strings.TrimSpace(out.IPHash)
	out.IPVersion = IPVersion(strings.TrimSpace(string(out.IPVersion)))
	out.Geo = out.Geo.Normalize()
	out.DeviceFingerprintHash = trimStringPtr(out.DeviceFingerprintHash)
	out.AuthSessionID = trimStringPtr(out.AuthSessionID)
	out.RiskLevel = RiskLevel(strings.TrimSpace(string(out.RiskLevel)))
	out.RiskReasons = normalizeStringSlice(out.RiskReasons)
	out.StartedAt = normalizeTime(out.StartedAt)
	out.LastSeenAt = normalizeTime(out.LastSeenAt)
	out.EndedAt = normalizeTimePtr(out.EndedAt)
	out.RevokedAt = normalizeTimePtr(out.RevokedAt)
	out.EndReason = trimEndReasonPtr(out.EndReason)
	out.CreatedAt = normalizeTime(out.CreatedAt)
	out.UpdatedAt = normalizeTime(out.UpdatedAt)
	out.RetainUntil = normalizeTimePtr(out.RetainUntil)
	out.AnonymizedAt = normalizeTimePtr(out.AnonymizedAt)
	out.DeletionRequestedAt = normalizeTimePtr(out.DeletionRequestedAt)
	out.LegalHoldReason = trimStringPtr(out.LegalHoldReason)
	out.LegalHoldSetBy = trimStringPtr(out.LegalHoldSetBy)
	out.LegalHoldSetAt = normalizeTimePtr(out.LegalHoldSetAt)
	out.RetentionClass = RetentionClass(strings.TrimSpace(string(out.RetentionClass)))
	return out
}

func (s Session) Validate() error {
	return s.ValidateWithConfig(DefaultValidationConfig())
}

func (s Session) ValidateWithConfig(cfg ValidationConfig) error {
	cfg = cfg.WithDefaults()
	session := s.Normalize()
	violations := make([]FieldViolation, 0)

	validateID(&violations, "session_id", session.SessionID, true, cfg.MaxIDLength)
	validateID(&violations, "anonymous_id", session.AnonymousID, true, cfg.MaxIDLength)
	validateOptionalID(&violations, "user_id", session.UserID, cfg.MaxIDLength)
	validateOptionalID(&violations, "auth_session_id", session.AuthSessionID, cfg.MaxIDLength)

	if session.SchemaVersion != CurrentSessionSchemaVersion {
		addViolation(&violations, "schema_version", "must match current session schema version")
	}
	if !session.Status.Valid() {
		addViolation(&violations, "status", "must be active, ended, expired, or revoked")
	}
	if !session.Channel.Valid() {
		addViolation(&violations, "channel", "must be a supported channel")
	}
	validatePagePath(&violations, "entry_page", session.EntryPage, true, cfg.MaxPageLength)
	validateOptionalPagePath(&violations, "exit_page", session.ExitPage, cfg.MaxPageLength)
	validateOptionalText(&violations, "referrer", session.Referrer, cfg.MaxReferrerLength)
	validateUTM(&violations, session.UTM, cfg)

	validateRequiredText(&violations, "user_agent", session.UserAgent, cfg.MaxUserAgentLength)
	validateHashPtr(&violations, "user_agent_hash", session.UserAgentHash, cfg.MaxHashLength)
	validateDevice(&violations, session.Device, cfg)
	validateClient(&violations, session.Client, cfg)
	validateHash(&violations, "ip_hash", session.IPHash, true, cfg.MaxHashLength)
	if session.IPVersion != "" && !session.IPVersion.Valid() {
		addViolation(&violations, "ip_version", "must be ipv4, ipv6, or unknown")
	}
	validateGeo(&violations, session.Geo, cfg)
	validateHashPtr(&violations, "device_fingerprint_hash", session.DeviceFingerprintHash, cfg.MaxHashLength)

	if !session.RiskLevel.Valid() {
		addViolation(&violations, "risk_level", "must be low, medium, high, or unknown")
	}
	validateRiskReasons(&violations, session.RiskReasons, cfg)
	validateTimestamps(&violations, session)
	validateLifecycle(&violations, session)
	validateOptionalRetentionMetadata(&violations, session.RetentionClass, session.LegalHold, session.LegalHoldReason, session.LegalHoldSetAt)

	if len(violations) > 0 {
		return ValidationError{Violations: violations}
	}
	return nil
}

func (s Session) Active() bool {
	return s.Status == SessionStatusActive && s.EndedAt == nil && s.RevokedAt == nil
}

func (s *Session) LinkUser(userID string, at time.Time) error {
	if s == nil {
		return ErrNilSession
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return ValidationError{Violations: []FieldViolation{{Field: "user_id", Message: "is required"}}}
	}
	timestamp := normalizeTime(at)
	if timestamp.IsZero() {
		timestamp = time.Now().UTC()
	}
	s.UserID = &userID
	if s.LastSeenAt.IsZero() || s.LastSeenAt.Before(timestamp) {
		s.LastSeenAt = timestamp
	}
	s.UpdatedAt = timestamp
	return s.Validate()
}

func (s *Session) End(reason SessionEndReason, endedAt time.Time, exitPage *string) error {
	if s == nil {
		return ErrNilSession
	}
	reason = normalizeEndReason(reason, EndReasonUnknown)
	timestamp := normalizeTime(endedAt)
	s.Status = SessionStatusEnded
	s.EndedAt = &timestamp
	s.RevokedAt = nil
	s.EndReason = &reason
	s.ExitPage = trimStringPtr(exitPage)
	if s.LastSeenAt.IsZero() || s.LastSeenAt.Before(timestamp) {
		s.LastSeenAt = timestamp
	}
	s.UpdatedAt = timestamp
	return s.Validate()
}

func (s *Session) Expire(expiredAt time.Time, exitPage *string) error {
	if s == nil {
		return ErrNilSession
	}
	reason := EndReasonInactivityTimeout
	timestamp := normalizeTime(expiredAt)
	s.Status = SessionStatusExpired
	s.EndedAt = &timestamp
	s.RevokedAt = nil
	s.EndReason = &reason
	s.ExitPage = trimStringPtr(exitPage)
	if s.LastSeenAt.IsZero() || s.LastSeenAt.Before(timestamp) {
		s.LastSeenAt = timestamp
	}
	s.UpdatedAt = timestamp
	return s.Validate()
}

func (s *Session) Revoke(reason SessionEndReason, revokedAt time.Time) error {
	if s == nil {
		return ErrNilSession
	}
	reason = normalizeEndReason(reason, EndReasonSecurityRevoke)
	timestamp := normalizeTime(revokedAt)
	s.Status = SessionStatusRevoked
	s.EndedAt = nil
	s.RevokedAt = &timestamp
	s.EndReason = &reason
	if s.LastSeenAt.IsZero() || s.LastSeenAt.Before(timestamp) {
		s.LastSeenAt = timestamp
	}
	s.UpdatedAt = timestamp
	return s.Validate()
}

func (d Device) Normalize() Device {
	return Device{
		Type:           DeviceType(strings.TrimSpace(string(d.Type))),
		Browser:        trimStringPtr(d.Browser),
		BrowserVersion: trimStringPtr(d.BrowserVersion),
		OS:             trimStringPtr(d.OS),
		OSVersion:      trimStringPtr(d.OSVersion),
		Model:          trimStringPtr(d.Model),
		Vendor:         trimStringPtr(d.Vendor),
		IsBot:          d.IsBot,
	}
}

func (g Geo) Normalize() Geo {
	source := GeoSource(strings.TrimSpace(string(g.Source)))
	if source == "" {
		source = GeoSourceUnknown
	}
	return Geo{
		Country:  trimStringPtr(g.Country),
		Region:   trimStringPtr(g.Region),
		City:     trimStringPtr(g.City),
		Timezone: trimStringPtr(g.Timezone),
		Source:   source,
	}
}

func (c Client) Normalize() Client {
	channel := Channel(strings.TrimSpace(string(c.Channel)))
	if channel == "" {
		channel = ChannelUnknown
	}
	return Client{
		Channel:        channel,
		Locale:         normalizeLocalePtr(c.Locale),
		Timezone:       normalizeTimezonePtr(c.Timezone),
		ScreenWidth:    normalizeDimension(c.ScreenWidth),
		ScreenHeight:   normalizeDimension(c.ScreenHeight),
		ViewportWidth:  normalizeDimension(c.ViewportWidth),
		ViewportHeight: normalizeDimension(c.ViewportHeight),
	}
}

func (u UTM) Normalize() UTM {
	return UTM{
		Source:   trimStringPtr(u.Source),
		Medium:   trimStringPtr(u.Medium),
		Campaign: trimStringPtr(u.Campaign),
		Term:     trimStringPtr(u.Term),
		Content:  trimStringPtr(u.Content),
	}
}

func validateTimestamps(violations *[]FieldViolation, session Session) {
	if session.StartedAt.IsZero() {
		addViolation(violations, "started_at", "is required")
	}
	if session.LastSeenAt.IsZero() {
		addViolation(violations, "last_seen_at", "is required")
	}
	if session.CreatedAt.IsZero() {
		addViolation(violations, "created_at", "is required")
	}
	if session.UpdatedAt.IsZero() {
		addViolation(violations, "updated_at", "is required")
	}
	if !session.StartedAt.IsZero() && !session.LastSeenAt.IsZero() && session.LastSeenAt.Before(session.StartedAt) {
		addViolation(violations, "last_seen_at", "cannot be before started_at")
	}
	if session.EndedAt != nil && !session.StartedAt.IsZero() && session.EndedAt.Before(session.StartedAt) {
		addViolation(violations, "ended_at", "cannot be before started_at")
	}
	if session.RevokedAt != nil && !session.StartedAt.IsZero() && session.RevokedAt.Before(session.StartedAt) {
		addViolation(violations, "revoked_at", "cannot be before started_at")
	}
	if !session.CreatedAt.IsZero() && !session.UpdatedAt.IsZero() && session.UpdatedAt.Before(session.CreatedAt) {
		addViolation(violations, "updated_at", "cannot be before created_at")
	}
}

func validateLifecycle(violations *[]FieldViolation, session Session) {
	if session.EndReason != nil && !session.EndReason.Valid() {
		addViolation(violations, "end_reason", "must be a supported end reason")
	}

	switch session.Status {
	case SessionStatusActive:
		if session.EndedAt != nil {
			addViolation(violations, "ended_at", "must be empty while session is active")
		}
		if session.RevokedAt != nil {
			addViolation(violations, "revoked_at", "must be empty while session is active")
		}
		if session.EndReason != nil {
			addViolation(violations, "end_reason", "must be empty while session is active")
		}
	case SessionStatusEnded:
		if session.EndedAt == nil {
			addViolation(violations, "ended_at", "is required when status is ended")
		}
		if session.RevokedAt != nil {
			addViolation(violations, "revoked_at", "must be empty when status is ended")
		}
	case SessionStatusExpired:
		if session.EndedAt == nil {
			addViolation(violations, "ended_at", "is required when status is expired")
		}
		if session.RevokedAt != nil {
			addViolation(violations, "revoked_at", "must be empty when status is expired")
		}
		if session.EndReason == nil || *session.EndReason != EndReasonInactivityTimeout {
			addViolation(violations, "end_reason", "must be inactivity_timeout when status is expired")
		}
	case SessionStatusRevoked:
		if session.RevokedAt == nil {
			addViolation(violations, "revoked_at", "is required when status is revoked")
		}
	}
}

func validateDevice(violations *[]FieldViolation, device Device, cfg ValidationConfig) {
	if !device.Type.Valid() {
		addViolation(violations, "device.type", "must be desktop, mobile, tablet, bot, or unknown")
	}
	validateOptionalText(violations, "device.browser", device.Browser, cfg.MaxMetadataLength)
	validateOptionalText(violations, "device.browser_version", device.BrowserVersion, cfg.MaxMetadataLength)
	validateOptionalText(violations, "device.os", device.OS, cfg.MaxMetadataLength)
	validateOptionalText(violations, "device.os_version", device.OSVersion, cfg.MaxMetadataLength)
	validateOptionalText(violations, "device.model", device.Model, cfg.MaxMetadataLength)
	validateOptionalText(violations, "device.vendor", device.Vendor, cfg.MaxMetadataLength)
}

func validateClient(violations *[]FieldViolation, client Client, cfg ValidationConfig) {
	if !client.Channel.Valid() {
		addViolation(violations, "client.channel", "must be a supported channel")
	}
	validateOptionalText(violations, "client.locale", client.Locale, cfg.MaxMetadataLength)
	validateOptionalText(violations, "client.timezone", client.Timezone, cfg.MaxMetadataLength)
	validateDimension(violations, "client.screen_width", client.ScreenWidth)
	validateDimension(violations, "client.screen_height", client.ScreenHeight)
	validateDimension(violations, "client.viewport_width", client.ViewportWidth)
	validateDimension(violations, "client.viewport_height", client.ViewportHeight)
}

func validateUTM(violations *[]FieldViolation, utm UTM, cfg ValidationConfig) {
	validateOptionalText(violations, "utm.source", utm.Source, cfg.MaxUTMValueLength)
	validateOptionalText(violations, "utm.medium", utm.Medium, cfg.MaxUTMValueLength)
	validateOptionalText(violations, "utm.campaign", utm.Campaign, cfg.MaxUTMValueLength)
	validateOptionalText(violations, "utm.term", utm.Term, cfg.MaxUTMValueLength)
	validateOptionalText(violations, "utm.content", utm.Content, cfg.MaxUTMValueLength)
}

func validateGeo(violations *[]FieldViolation, geo Geo, cfg ValidationConfig) {
	validateOptionalText(violations, "geo.country", geo.Country, cfg.MaxGeoValueLength)
	validateOptionalText(violations, "geo.region", geo.Region, cfg.MaxGeoValueLength)
	validateOptionalText(violations, "geo.city", geo.City, cfg.MaxGeoValueLength)
	validateOptionalText(violations, "geo.timezone", geo.Timezone, cfg.MaxGeoValueLength)
	if !geo.Source.Valid() {
		addViolation(violations, "geo.source", "must be geoip, header, unknown, or disabled")
	}
}

func validateRiskReasons(violations *[]FieldViolation, reasons []string, cfg ValidationConfig) {
	if len(reasons) > cfg.MaxRiskReasons {
		addViolation(violations, "risk_reasons", "has too many values")
		return
	}
	for i, reason := range reasons {
		field := fmt.Sprintf("risk_reasons[%d]", i)
		validateRequiredText(violations, field, reason, cfg.MaxRiskReasonTextLength)
	}
}

func validateID(violations *[]FieldViolation, field string, value string, required bool, maxLength int) {
	value = strings.TrimSpace(value)
	if value == "" {
		if required {
			addViolation(violations, field, "is required")
		}
		return
	}
	if len(value) > maxLength {
		addViolation(violations, field, "is too long")
	}
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		switch r {
		case '_', '-', '.', ':':
			continue
		default:
			addViolation(violations, field, "contains unsupported characters")
			return
		}
	}
}

func validateOptionalID(violations *[]FieldViolation, field string, value *string, maxLength int) {
	if value == nil {
		return
	}
	validateID(violations, field, *value, true, maxLength)
}

func validateRequiredText(violations *[]FieldViolation, field string, value string, maxLength int) {
	value = strings.TrimSpace(value)
	if value == "" {
		addViolation(violations, field, "is required")
		return
	}
	validateText(violations, field, value, maxLength)
}

func validateOptionalText(violations *[]FieldViolation, field string, value *string, maxLength int) {
	if value == nil {
		return
	}
	validateText(violations, field, strings.TrimSpace(*value), maxLength)
}

func validateText(violations *[]FieldViolation, field string, value string, maxLength int) {
	if len(value) > maxLength {
		addViolation(violations, field, "is too long")
	}
	if containsControl(value) {
		addViolation(violations, field, "contains control characters")
	}
}

func validatePagePath(violations *[]FieldViolation, field string, value string, required bool, maxLength int) {
	value = strings.TrimSpace(value)
	if value == "" {
		if required {
			addViolation(violations, field, "is required")
		}
		return
	}
	if len(value) > maxLength {
		addViolation(violations, field, "is too long")
		return
	}
	if containsControl(value) || strings.ContainsAny(value, " \t\r\n") {
		addViolation(violations, field, "must be a valid path")
		return
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.IsAbs() || !strings.HasPrefix(parsed.Path, "/") || strings.HasPrefix(value, "//") {
		addViolation(violations, field, "must be a valid path")
	}
}

func validateOptionalPagePath(violations *[]FieldViolation, field string, value *string, maxLength int) {
	if value == nil {
		return
	}
	validatePagePath(violations, field, *value, true, maxLength)
}

func validateHash(violations *[]FieldViolation, field string, value string, required bool, maxLength int) {
	value = strings.TrimSpace(value)
	if value == "" {
		if required {
			addViolation(violations, field, "is required")
		}
		return
	}
	validateText(violations, field, value, maxLength)
	if net.ParseIP(value) != nil {
		addViolation(violations, field, "must not contain a raw IP address")
	}
}

func validateHashPtr(violations *[]FieldViolation, field string, value *string, maxLength int) {
	if value == nil {
		return
	}
	validateHash(violations, field, *value, true, maxLength)
}

func validateOptionalRetentionMetadata(violations *[]FieldViolation, class RetentionClass, legalHold bool, reason *string, setAt *time.Time) {
	if class != "" && !class.Valid() {
		addViolation(violations, "retention_class", "must be a supported retention class")
	}
	if legalHold {
		if reason == nil || strings.TrimSpace(*reason) == "" {
			addViolation(violations, "legal_hold_reason", "is required when legal_hold is true")
		}
		if setAt == nil || setAt.IsZero() {
			addViolation(violations, "legal_hold_set_at", "is required when legal_hold is true")
		}
	}
}

func addViolation(violations *[]FieldViolation, field string, message string) {
	*violations = append(*violations, FieldViolation{Field: field, Message: message})
}

func trimStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeLocalePtr(value *string) *string {
	value = trimStringPtr(value)
	if value == nil {
		return nil
	}
	locale := strings.ReplaceAll(*value, "_", "-")
	locale = strings.Trim(locale, ",; ")
	if locale == "" {
		return nil
	}
	return &locale
}

func normalizeTimezonePtr(value *string) *string {
	value = trimStringPtr(value)
	if value == nil {
		return nil
	}
	timezone := strings.Trim(*value, `"'`)
	timezone = strings.TrimSpace(timezone)
	if timezone == "" {
		return nil
	}
	return &timezone
}

func normalizeDimension(value int) int {
	if value <= 0 || value > 10000 {
		return 0
	}
	return value
}

func validateDimension(violations *[]FieldViolation, field string, value int) {
	if value == 0 {
		return
	}
	if value < 0 || value > 10000 {
		addViolation(violations, field, "must be between 1 and 10000")
	}
}

func trimEndReasonPtr(value *SessionEndReason) *SessionEndReason {
	if value == nil {
		return nil
	}
	reason := normalizeEndReason(*value, "")
	if reason == "" {
		return nil
	}
	return &reason
}

func normalizeEndReason(value SessionEndReason, fallback SessionEndReason) SessionEndReason {
	reason := SessionEndReason(strings.TrimSpace(string(value)))
	if reason == "" {
		return fallback
	}
	return reason
}

func normalizeStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		out = append(out, value)
	}
	return out
}

func normalizeTime(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}
	return value.UTC()
}

func normalizeTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	normalized := normalizeTime(*value)
	if normalized.IsZero() {
		return nil
	}
	return &normalized
}

func containsControl(value string) bool {
	for _, r := range value {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}
