package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"time"
)

type MaskingMode string

const (
	MaskingModeHidden MaskingMode = "hidden"
	MaskingModeMasked MaskingMode = "masked"
	MaskingModeFull   MaskingMode = "full"
)

type LocationGranularity string

const (
	LocationGranularityNone    LocationGranularity = "none"
	LocationGranularityCountry LocationGranularity = "country"
	LocationGranularityCity    LocationGranularity = "city"
)

type PrivacyMaskingSettings struct {
	UserIDMode          MaskingMode         `json:"userIdMode"`
	AnonymousIDMode     MaskingMode         `json:"anonymousIdMode"`
	SessionIDMode       MaskingMode         `json:"sessionIdMode"`
	LocationGranularity LocationGranularity `json:"locationGranularity"`
	ShowSearchQueries   bool                `json:"showSearchQueries"`
	ShowIPHash          bool                `json:"showIpHash"`
}

type PrivacyPermissions struct {
	CanUpdateMasking   bool `json:"canUpdateMasking"`
	CanRequestDeletion bool `json:"canRequestDeletion"`
	CanUpdateRetention bool `json:"canUpdateRetention"`
}

type RetentionSettings struct {
	RawEventsDays             int `json:"rawEventsDays"`
	JourneySummariesDays      int `json:"journeySummariesDays"`
	HeatmapAggregatesDays     int `json:"heatmapAggregatesDays"`
	AnalyticsAggregatesMonths int `json:"analyticsAggregatesMonths"`
	ActiveSessionTTLMinutes   int `json:"activeSessionTtlMinutes"`
	DeletionRequestLogDays    int `json:"deletionRequestLogDays"`
}

type PrivacySettings struct {
	Masking     PrivacyMaskingSettings `json:"masking"`
	Retention   RetentionSettings      `json:"retention,omitempty"`
	Permissions PrivacyPermissions     `json:"permissions"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	UpdatedBy   string                 `json:"updatedBy"`
}

type DeletionTargetType string

const (
	DeletionTargetUserID      DeletionTargetType = "user_id"
	DeletionTargetAnonymousID DeletionTargetType = "anonymous_id"
	DeletionTargetSessionID   DeletionTargetType = "session_id"
)

type DeletionTarget struct {
	Type  DeletionTargetType `json:"targetType"`
	Value string             `json:"targetValue"`
}

type AggregateImpact string

const (
	AggregateImpactUnchanged             AggregateImpact = "unchanged"
	AggregateImpactAnonymized            AggregateImpact = "anonymized"
	AggregateImpactAnonymizedOrUnchanged AggregateImpact = "aggregates_anonymized_or_unchanged"
)

type DeletionPreview struct {
	TargetType                 DeletionTargetType `json:"targetType"`
	TargetValueMasked          string             `json:"targetValueMasked"`
	MatchedSessions            int64              `json:"matchedSessions"`
	MatchedEvents              int64              `json:"matchedEvents"`
	MatchedJourneySummaries    int64              `json:"matchedJourneySummaries"`
	MatchedActiveSessions      int64              `json:"matchedActiveSessions"`
	AggregateImpact            AggregateImpact    `json:"aggregateImpact"`
	EstimatedCompletionSeconds int                `json:"estimatedCompletionSeconds"`
}

type DeletionRequestStatus string

const (
	DeletionStatusQueued     DeletionRequestStatus = "queued"
	DeletionStatusProcessing DeletionRequestStatus = "processing"
	DeletionStatusCompleted  DeletionRequestStatus = "completed"
	DeletionStatusFailed     DeletionRequestStatus = "failed"
)

type DeletionRequest struct {
	RequestID               string                `json:"requestId"`
	TargetType              DeletionTargetType    `json:"targetType"`
	TargetHash              string                `json:"-"`
	TargetValueMasked       string                `json:"targetValueMasked"`
	Status                  DeletionRequestStatus `json:"status"`
	RequestedBy             string                `json:"requestedBy"`
	Reason                  string                `json:"reason"`
	MatchedSessions         int64                 `json:"matchedSessions"`
	MatchedEvents           int64                 `json:"matchedEvents"`
	MatchedJourneySummaries int64                 `json:"matchedJourneySummaries"`
	MatchedActiveSessions   int64                 `json:"matchedActiveSessions"`
	CreatedAt               time.Time             `json:"createdAt"`
	CompletedAt             *time.Time            `json:"completedAt,omitempty"`
	Error                   string                `json:"error,omitempty"`
}

type Actor struct {
	ID    string
	Roles []string
}

type AuditEvent struct {
	EventID      string
	ActorID      string
	Action       string
	ResourceType string
	ResourceID   string
	RequestID    string
	Reason       string
	Metadata     map[string]any
	CreatedAt    time.Time
}

var targetValuePattern = regexp.MustCompile(`^[A-Za-z0-9._:@-]+$`)

func DefaultMaskingSettings() PrivacyMaskingSettings {
	return PrivacyMaskingSettings{
		UserIDMode:          MaskingModeMasked,
		AnonymousIDMode:     MaskingModeMasked,
		SessionIDMode:       MaskingModeMasked,
		LocationGranularity: LocationGranularityCountry,
		ShowSearchQueries:   false,
		ShowIPHash:          false,
	}
}

func DefaultRetentionSettings() RetentionSettings {
	return RetentionSettings{
		RawEventsDays:             90,
		JourneySummariesDays:      365,
		HeatmapAggregatesDays:     365,
		AnalyticsAggregatesMonths: 36,
		ActiveSessionTTLMinutes:   45,
		DeletionRequestLogDays:    730,
	}
}

func DefaultPrivacySettings(now time.Time, updatedBy string) PrivacySettings {
	if updatedBy == "" {
		updatedBy = "system"
	}
	return PrivacySettings{
		Masking:   DefaultMaskingSettings(),
		Retention: DefaultRetentionSettings(),
		UpdatedAt: now.UTC(),
		UpdatedBy: updatedBy,
	}
}

func ValidateMaskingSettings(settings PrivacyMaskingSettings) error {
	if !validMaskingMode(settings.UserIDMode) {
		return NewFieldError("userIdMode", "unsupported masking mode")
	}
	if !validMaskingMode(settings.AnonymousIDMode) {
		return NewFieldError("anonymousIdMode", "unsupported masking mode")
	}
	if !validMaskingMode(settings.SessionIDMode) {
		return NewFieldError("sessionIdMode", "unsupported masking mode")
	}
	if !validLocationGranularity(settings.LocationGranularity) {
		return NewFieldError("locationGranularity", "unsupported location granularity")
	}
	return nil
}

func ValidateRetentionSettings(settings RetentionSettings) error {
	if settings.RawEventsDays < 7 || settings.RawEventsDays > 180 {
		return NewFieldError("rawEventsDays", "must be between 7 and 180 days")
	}
	if settings.JourneySummariesDays < 30 || settings.JourneySummariesDays > 730 {
		return NewFieldError("journeySummariesDays", "must be between 30 and 730 days")
	}
	if settings.HeatmapAggregatesDays < 30 || settings.HeatmapAggregatesDays > 730 {
		return NewFieldError("heatmapAggregatesDays", "must be between 30 and 730 days")
	}
	if settings.AnalyticsAggregatesMonths < 12 || settings.AnalyticsAggregatesMonths > 84 {
		return NewFieldError("analyticsAggregatesMonths", "must be between 12 and 84 months")
	}
	if settings.ActiveSessionTTLMinutes < 15 || settings.ActiveSessionTTLMinutes > 180 {
		return NewFieldError("activeSessionTtlMinutes", "must be between 15 and 180 minutes")
	}
	if settings.DeletionRequestLogDays < 365 || settings.DeletionRequestLogDays > 2555 {
		return NewFieldError("deletionRequestLogDays", "must be between 365 and 2555 days")
	}
	return nil
}

func ValidateDeletionTarget(target DeletionTarget) (DeletionTarget, error) {
	target.Type = DeletionTargetType(strings.TrimSpace(string(target.Type)))
	target.Value = strings.TrimSpace(target.Value)

	switch target.Type {
	case DeletionTargetUserID, DeletionTargetAnonymousID, DeletionTargetSessionID:
	default:
		return target, NewFieldError("targetType", "unsupported deletion target type")
	}

	if len(target.Value) < 3 || len(target.Value) > 160 {
		return target, NewFieldError("targetValue", "must be between 3 and 160 characters")
	}
	if !targetValuePattern.MatchString(target.Value) {
		return target, NewFieldError("targetValue", "contains unsupported characters")
	}
	return target, nil
}

func ValidateReason(reason string) (string, error) {
	trimmed := strings.TrimSpace(reason)
	if len(trimmed) < 10 {
		return "", ErrReasonRequired
	}
	if len(trimmed) > 512 {
		return "", NewFieldError("reason", "must be 512 characters or fewer")
	}
	return trimmed, nil
}

func MaskIdentifier(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "-"
	}
	if len(trimmed) <= 8 {
		return "****"
	}
	return trimmed[:6] + "..." + trimmed[len(trimmed)-4:]
}

func HashDeletionTarget(pepper string, target DeletionTarget) string {
	message := strings.ToLower(string(target.Type)) + ":" + target.Value
	mac := hmac.New(sha256.New, []byte(pepper))
	_, _ = mac.Write([]byte(message))
	return "sha256:" + hex.EncodeToString(mac.Sum(nil))
}

func validMaskingMode(value MaskingMode) bool {
	return value == MaskingModeHidden || value == MaskingModeMasked || value == MaskingModeFull
}

func validLocationGranularity(value LocationGranularity) bool {
	return value == LocationGranularityNone ||
		value == LocationGranularityCountry ||
		value == LocationGranularityCity
}
