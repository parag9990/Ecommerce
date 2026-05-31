package domain

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultAuditLogPageSize = 50
	MaxAuditLogPageSize     = 100

	maxAuditActorIDLength      = 64
	maxAuditActionLength       = 128
	maxAuditResourceTypeLength = 64
	maxAuditResourceIDLength   = 128
	maxAuditRequestIDLength    = 64
	maxAuditSessionIDLength    = 64
	maxAuditIPHashLength       = 128
	maxAuditCursorLength       = 256
)

type AuditRecord struct {
	ActorAdminID string            `json:"actor_admin_id"`
	Action       string            `json:"action"`
	ResourceType string            `json:"resource_type"`
	ResourceID   string            `json:"resource_id"`
	RequestID    string            `json:"request_id,omitempty"`
	SessionID    string            `json:"session_id,omitempty"`
	IPHash       string            `json:"ip_hash,omitempty"`
	Reason       string            `json:"reason,omitempty"`
	Before       map[string]string `json:"before_json,omitempty"`
	After        map[string]string `json:"after_json,omitempty"`
	Metadata     map[string]string `json:"metadata_json,omitempty"`
}

func (r AuditRecord) Validate() error {
	if _, err := requiredAuditField(r.ActorAdminID, "actor_admin_id", maxAuditActorIDLength); err != nil {
		return err
	}
	if _, err := requiredAuditField(r.Action, "action", maxAuditActionLength); err != nil {
		return err
	}
	if _, err := requiredAuditField(r.ResourceType, "resource_type", maxAuditResourceTypeLength); err != nil {
		return err
	}
	if _, err := requiredAuditField(r.ResourceID, "resource_id", maxAuditResourceIDLength); err != nil {
		return err
	}
	if _, err := requiredAuditField(r.RequestID, "request_id", maxAuditRequestIDLength); err != nil {
		return err
	}
	if _, err := requiredAuditField(r.Reason, "reason", MaxReasonLength); err != nil {
		return err
	}
	if _, err := optionalAuditField(r.SessionID, "session_id", maxAuditSessionIDLength); err != nil {
		return err
	}
	if _, err := optionalAuditField(r.IPHash, "ip_hash", maxAuditIPHashLength); err != nil {
		return err
	}
	return nil
}

type AuditLog struct {
	SequenceID   uint64            `json:"-"`
	AuditID      string            `json:"audit_id"`
	ActorAdminID string            `json:"actor_admin_id"`
	Action       string            `json:"action"`
	ResourceType string            `json:"resource_type"`
	ResourceID   string            `json:"resource_id"`
	RequestID    string            `json:"request_id,omitempty"`
	SessionID    string            `json:"session_id,omitempty"`
	IPHash       string            `json:"ip_hash,omitempty"`
	Reason       string            `json:"reason,omitempty"`
	Before       map[string]string `json:"before_json,omitempty"`
	After        map[string]string `json:"after_json,omitempty"`
	Metadata     map[string]string `json:"metadata_json,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
}

type AuditLogListRequest struct {
	ActorAdminID string
	Action       string
	ResourceType string
	ResourceID   string
	RequestID    string
	From         time.Time
	To           time.Time
	Pagination
}

func (r AuditLogListRequest) Normalize() (AuditLogListRequest, error) {
	var err error
	if r.ActorAdminID, err = optionalAuditField(r.ActorAdminID, "actor_id", maxAuditActorIDLength); err != nil {
		return AuditLogListRequest{}, err
	}
	if r.Action, err = optionalAuditField(r.Action, "action", maxAuditActionLength); err != nil {
		return AuditLogListRequest{}, err
	}
	if r.ResourceType, err = optionalAuditField(r.ResourceType, "resource_type", maxAuditResourceTypeLength); err != nil {
		return AuditLogListRequest{}, err
	}
	if r.ResourceID, err = optionalAuditField(r.ResourceID, "resource_id", maxAuditResourceIDLength); err != nil {
		return AuditLogListRequest{}, err
	}
	if r.RequestID, err = optionalAuditField(r.RequestID, "request_id", maxAuditRequestIDLength); err != nil {
		return AuditLogListRequest{}, err
	}
	if !r.From.IsZero() && !r.To.IsZero() && r.From.After(r.To) {
		return AuditLogListRequest{}, NewValidationError("from must be before or equal to to")
	}
	if r.Pagination.Page < 1 {
		r.Pagination.Page = 1
	}
	if r.Pagination.PageSize < 1 {
		r.Pagination.PageSize = DefaultAuditLogPageSize
	}
	if r.Pagination.PageSize > MaxAuditLogPageSize {
		r.Pagination.PageSize = MaxAuditLogPageSize
	}
	r.Pagination.Cursor = strings.TrimSpace(r.Pagination.Cursor)
	if len(r.Pagination.Cursor) > maxAuditCursorLength {
		return AuditLogListRequest{}, NewValidationError(fmt.Sprintf("cursor must be at most %d characters", maxAuditCursorLength))
	}
	if _, err := ParseAuditLogCursor(r.Pagination.Cursor); err != nil {
		return AuditLogListRequest{}, err
	}
	return r, nil
}

type AuditLogListResponse struct {
	Logs       []AuditLog `json:"logs"`
	NextCursor string     `json:"next_cursor,omitempty"`
}

func SafeAuditSnapshot(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		cleanKey := strings.TrimSpace(key)
		if cleanKey == "" || isSensitiveAuditKey(cleanKey) {
			continue
		}
		out[cleanKey] = value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func NewAuditLogCursor(sequenceID uint64) string {
	if sequenceID == 0 {
		return ""
	}
	raw := strconv.FormatUint(sequenceID, 10)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func ParseAuditLogCursor(cursor string) (uint64, error) {
	cursor = strings.TrimSpace(cursor)
	if cursor == "" {
		return 0, nil
	}
	if id, err := strconv.ParseUint(cursor, 10, 64); err == nil {
		return id, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return 0, NewValidationError("cursor is invalid")
	}
	id, err := strconv.ParseUint(string(decoded), 10, 64)
	if err != nil || id == 0 {
		return 0, NewValidationError("cursor is invalid")
	}
	return id, nil
}

func requiredAuditField(value string, field string, maxLength int) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", NewValidationError(field + " is required")
	}
	return auditFieldWithinLimit(value, field, maxLength)
}

func optionalAuditField(value string, field string, maxLength int) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	return auditFieldWithinLimit(value, field, maxLength)
}

func auditFieldWithinLimit(value string, field string, maxLength int) (string, error) {
	if len(value) > maxLength {
		return "", NewValidationError(fmt.Sprintf("%s must be at most %d characters", field, maxLength))
	}
	return value, nil
}

func isSensitiveAuditKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")

	sensitiveFragments := []string{
		"password",
		"otp",
		"token",
		"authorization",
		"cookie",
		"card_number",
		"cvv",
		"secret",
		"api_key",
		"private_key",
	}
	for _, fragment := range sensitiveFragments {
		if strings.Contains(normalized, fragment) {
			return true
		}
	}
	return false
}

func UserStatusAuditAction(status UserStatus) string {
	return "user.status." + string(status)
}

func SellerStatusAuditAction(status SellerStatus) string {
	return "seller.status." + string(status)
}
