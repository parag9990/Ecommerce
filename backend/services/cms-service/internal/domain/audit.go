package domain

import (
	"strings"
	"time"
)

type AuditDecision string

const (
	AuditDecisionAllowed AuditDecision = "allowed"
	AuditDecisionDenied  AuditDecision = "denied"
)

const (
	AuditActionProductSubmittedReview Permission = "cms:product:submitted_review"
	AuditActionProductApproved        Permission = "cms:product:approved"
	AuditActionProductRejected        Permission = "cms:product:rejected"
	AuditActionProductPublished       Permission = "cms:product:published"
	AuditActionProductUnpublished     Permission = "cms:product:unpublished"
	AuditActionCouponCreated          Permission = "cms:coupon:created"
	AuditActionCouponUpdated          Permission = "cms:coupon:updated"
	AuditActionCouponDisabled         Permission = "cms:coupon:disabled"
	AuditActionCouponRedeemed         Permission = "cms:coupon:redeemed"
	AuditActionCampaignCreated        Permission = "cms:campaigns:create"
	AuditActionCampaignUpdated        Permission = "cms:campaigns:update"
	AuditActionCampaignDisabled       Permission = "cms:campaigns:disable"
	AuditActionCampaignResumed        Permission = "cms:campaigns:resume"
	AuditActionCampaignCompleted      Permission = "cms:campaigns:complete"
)

type AuditEvent struct {
	AuditID          string
	ActorUserID      string
	ActorSellerID    string
	ActorRoles       []Role
	Action           Permission
	ResourceType     string
	ResourceID       string
	ResourceSellerID string
	RequestID        string
	TraceID          string
	Decision         AuditDecision
	Reason           string
	IPHash           string
	UserAgentHash    string
	Before           map[string]any
	After            map[string]any
	CreatedAt        time.Time
}

type AuditEntry struct {
	SequenceID    uint64
	AuditID       string
	SellerID      string
	ActorUserID   string
	ActorSellerID string
	ActorRoles    []Role
	Action        Permission
	ResourceType  string
	ResourceID    string
	RequestID     string
	TraceID       string
	Decision      AuditDecision
	Reason        string
	IPHash        string
	UserAgentHash string
	Before        map[string]any
	After         map[string]any
	CreatedAt     time.Time
}

type AuditLogFilter struct {
	SellerID     string
	ActorUserID  string
	Action       Permission
	ResourceType string
	ResourceID   string
	From         time.Time
	To           time.Time
	PageSize     int
	Cursor       string
}

type AuditLogPage struct {
	Entries    []AuditEntry
	NextCursor string
}

func AuditEntryFromEvent(event AuditEvent) AuditEntry {
	sellerID := firstAuditNonEmpty(event.ResourceSellerID, event.ActorSellerID)
	return AuditEntry{
		AuditID:       strings.TrimSpace(event.AuditID),
		SellerID:      sellerID,
		ActorUserID:   strings.TrimSpace(event.ActorUserID),
		ActorSellerID: strings.TrimSpace(event.ActorSellerID),
		ActorRoles:    NormalizeRoles(RoleStrings(event.ActorRoles)),
		Action:        Permission(strings.TrimSpace(event.Action.String())),
		ResourceType:  strings.TrimSpace(event.ResourceType),
		ResourceID:    strings.TrimSpace(event.ResourceID),
		RequestID:     strings.TrimSpace(event.RequestID),
		TraceID:       strings.TrimSpace(event.TraceID),
		Decision:      event.Decision.Normalized(),
		Reason:        strings.TrimSpace(event.Reason),
		IPHash:        strings.TrimSpace(event.IPHash),
		UserAgentHash: strings.TrimSpace(event.UserAgentHash),
		Before:        RedactAuditSnapshot(event.Before),
		After:         RedactAuditSnapshot(event.After),
		CreatedAt:     event.CreatedAt.UTC(),
	}.Normalized()
}

func (e AuditEntry) Normalized() AuditEntry {
	e.AuditID = strings.TrimSpace(e.AuditID)
	e.SellerID = strings.TrimSpace(e.SellerID)
	e.ActorUserID = strings.TrimSpace(e.ActorUserID)
	e.ActorSellerID = strings.TrimSpace(e.ActorSellerID)
	e.ActorRoles = NormalizeRoles(RoleStrings(e.ActorRoles))
	e.Action = Permission(strings.TrimSpace(e.Action.String()))
	e.ResourceType = strings.TrimSpace(e.ResourceType)
	e.ResourceID = strings.TrimSpace(e.ResourceID)
	e.RequestID = strings.TrimSpace(e.RequestID)
	e.TraceID = strings.TrimSpace(e.TraceID)
	e.Decision = e.Decision.Normalized()
	e.Reason = strings.TrimSpace(e.Reason)
	e.IPHash = strings.TrimSpace(e.IPHash)
	e.UserAgentHash = strings.TrimSpace(e.UserAgentHash)
	e.Before = RedactAuditSnapshot(e.Before)
	e.After = RedactAuditSnapshot(e.After)
	if !e.CreatedAt.IsZero() {
		e.CreatedAt = e.CreatedAt.UTC()
	}
	return e
}

func (d AuditDecision) Normalized() AuditDecision {
	switch AuditDecision(strings.ToLower(strings.TrimSpace(string(d)))) {
	case AuditDecisionDenied:
		return AuditDecisionDenied
	default:
		return AuditDecisionAllowed
	}
}

func RedactAuditSnapshot(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	out := make(map[string]any, len(input))
	for key, value := range input {
		if isSensitiveAuditKey(key) {
			out[key] = "[REDACTED]"
			continue
		}
		out[key] = redactAuditValue(value)
	}
	return out
}

func redactAuditValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return RedactAuditSnapshot(typed)
	case []any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, redactAuditValue(item))
		}
		return out
	default:
		return value
	}
}

func isSensitiveAuditKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")
	switch normalized {
	case "password", "passcode", "otp", "one_time_password",
		"token", "access_token", "refresh_token", "jwt", "authorization",
		"api_key", "secret", "client_secret", "card_number", "cvv", "cvc":
		return true
	default:
		return strings.Contains(normalized, "password") ||
			strings.Contains(normalized, "token") ||
			strings.Contains(normalized, "secret")
	}
}

func firstAuditNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
