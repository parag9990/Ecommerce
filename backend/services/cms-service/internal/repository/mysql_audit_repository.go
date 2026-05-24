package repository

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
)

const defaultAuditRepositoryPageSize = 20
const maxAuditRepositoryPageSize = 100

type MySQLAuditRecorder struct {
	db *sql.DB
}

type auditCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uint64    `json:"id"`
}

func NewMySQLAuditRecorder(db *sql.DB) (*MySQLAuditRecorder, error) {
	if db == nil {
		return nil, errors.New("mysql db is required")
	}
	return &MySQLAuditRecorder{db: db}, nil
}

func (r *MySQLAuditRecorder) Record(ctx context.Context, event domain.AuditEvent) error {
	entry := domain.AuditEntryFromEvent(event)
	if entry.AuditID == "" {
		return errors.New("audit_id is required")
	}
	if entry.SellerID == "" {
		return errors.New("seller_id is required")
	}
	if entry.ActorUserID == "" {
		return errors.New("actor_user_id is required")
	}
	if entry.Action == "" {
		return errors.New("audit action is required")
	}
	if entry.ResourceType == "" {
		entry.ResourceType = "unknown"
	}
	if entry.ResourceID == "" {
		entry.ResourceID = firstNonEmpty(entry.SellerID, "unknown")
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}

	actorRolesJSON, err := marshalOptionalStrings(domain.RoleStrings(entry.ActorRoles))
	if err != nil {
		return err
	}
	beforeJSON, err := marshalOptionalMap(entry.Before)
	if err != nil {
		return err
	}
	afterJSON, err := marshalOptionalMap(entry.After)
	if err != nil {
		return err
	}

	const query = `
INSERT INTO cms_audit_logs (
  audit_id,
  seller_id,
  actor_user_id,
  actor_roles_json,
  action,
  resource_type,
  resource_id,
  request_id,
  trace_id,
  decision,
  reason,
  ip_hash,
  user_agent_hash,
  before_json,
  after_json,
  created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(
		ctx,
		query,
		entry.AuditID,
		entry.SellerID,
		entry.ActorUserID,
		actorRolesJSON,
		entry.Action.String(),
		entry.ResourceType,
		entry.ResourceID,
		nullString(entry.RequestID),
		nullString(entry.TraceID),
		entry.Decision.Normalized(),
		nullString(entry.Reason),
		nullString(entry.IPHash),
		nullString(entry.UserAgentHash),
		beforeJSON,
		afterJSON,
		entry.CreatedAt.UTC(),
	)
	return err
}

func (r *MySQLAuditRecorder) ListAuditLogs(ctx context.Context, filter domain.AuditLogFilter) (domain.AuditLogPage, error) {
	filter = normalizeAuditFilter(filter)
	limit := filter.PageSize
	if limit <= 0 {
		limit = defaultAuditRepositoryPageSize
	}
	if limit > maxAuditRepositoryPageSize {
		limit = maxAuditRepositoryPageSize
	}

	cursor, hasCursor, err := decodeAuditCursor(filter.Cursor)
	if err != nil {
		return domain.AuditLogPage{}, err
	}

	query := strings.Builder{}
	query.WriteString(`
SELECT
  id,
  audit_id,
  seller_id,
  actor_user_id,
  actor_roles_json,
  action,
  resource_type,
  resource_id,
  request_id,
  trace_id,
  decision,
  reason,
  ip_hash,
  user_agent_hash,
  before_json,
  after_json,
  created_at
FROM cms_audit_logs
WHERE seller_id = ?`)
	args := []any{filter.SellerID}
	if filter.ActorUserID != "" {
		query.WriteString(" AND actor_user_id = ?")
		args = append(args, filter.ActorUserID)
	}
	if filter.Action != "" {
		query.WriteString(" AND action = ?")
		args = append(args, filter.Action.String())
	}
	if filter.ResourceType != "" {
		query.WriteString(" AND resource_type = ?")
		args = append(args, filter.ResourceType)
	}
	if filter.ResourceID != "" {
		query.WriteString(" AND resource_id = ?")
		args = append(args, filter.ResourceID)
	}
	if !filter.From.IsZero() {
		query.WriteString(" AND created_at >= ?")
		args = append(args, filter.From.UTC())
	}
	if !filter.To.IsZero() {
		query.WriteString(" AND created_at < ?")
		args = append(args, filter.To.UTC())
	}
	if hasCursor {
		query.WriteString(" AND (created_at < ? OR (created_at = ? AND id < ?))")
		args = append(args, cursor.CreatedAt.UTC(), cursor.CreatedAt.UTC(), cursor.ID)
	}
	query.WriteString(" ORDER BY created_at DESC, id DESC LIMIT ?")
	args = append(args, limit+1)

	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return domain.AuditLogPage{}, err
	}
	defer rows.Close()

	entries := make([]domain.AuditEntry, 0, limit+1)
	for rows.Next() {
		entry, err := scanAuditEntry(rows)
		if err != nil {
			return domain.AuditLogPage{}, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return domain.AuditLogPage{}, err
	}

	page := domain.AuditLogPage{Entries: entries}
	if len(page.Entries) > limit {
		page.Entries = page.Entries[:limit]
		last := page.Entries[len(page.Entries)-1]
		nextCursor, err := encodeAuditCursor(auditCursor{CreatedAt: last.CreatedAt, ID: last.SequenceID})
		if err != nil {
			return domain.AuditLogPage{}, err
		}
		page.NextCursor = nextCursor
	}
	return page, nil
}

type auditScanner interface {
	Scan(dest ...any) error
}

func scanAuditEntry(scanner auditScanner) (domain.AuditEntry, error) {
	var entry domain.AuditEntry
	var action string
	var decision string
	var actorRolesRaw []byte
	var beforeRaw []byte
	var afterRaw []byte
	var requestID sql.NullString
	var traceID sql.NullString
	var reason sql.NullString
	var ipHash sql.NullString
	var userAgentHash sql.NullString

	err := scanner.Scan(
		&entry.SequenceID,
		&entry.AuditID,
		&entry.SellerID,
		&entry.ActorUserID,
		&actorRolesRaw,
		&action,
		&entry.ResourceType,
		&entry.ResourceID,
		&requestID,
		&traceID,
		&decision,
		&reason,
		&ipHash,
		&userAgentHash,
		&beforeRaw,
		&afterRaw,
		&entry.CreatedAt,
	)
	if err != nil {
		return domain.AuditEntry{}, err
	}
	entry.Action = domain.Permission(action)
	entry.Decision = domain.AuditDecision(decision).Normalized()
	entry.RequestID = requestID.String
	entry.TraceID = traceID.String
	entry.Reason = reason.String
	entry.IPHash = ipHash.String
	entry.UserAgentHash = userAgentHash.String
	roles, err := decodeRoleJSON(actorRolesRaw)
	if err != nil {
		return domain.AuditEntry{}, err
	}
	entry.ActorRoles = roles
	before, err := decodeAuditMap(beforeRaw)
	if err != nil {
		return domain.AuditEntry{}, err
	}
	entry.Before = before
	after, err := decodeAuditMap(afterRaw)
	if err != nil {
		return domain.AuditEntry{}, err
	}
	entry.After = after
	return entry.Normalized(), nil
}

func normalizeAuditFilter(filter domain.AuditLogFilter) domain.AuditLogFilter {
	filter.SellerID = strings.TrimSpace(filter.SellerID)
	filter.ActorUserID = strings.TrimSpace(filter.ActorUserID)
	filter.Action = domain.Permission(strings.TrimSpace(filter.Action.String()))
	filter.ResourceType = strings.TrimSpace(filter.ResourceType)
	filter.ResourceID = strings.TrimSpace(filter.ResourceID)
	filter.Cursor = strings.TrimSpace(filter.Cursor)
	if !filter.From.IsZero() {
		filter.From = filter.From.UTC()
	}
	if !filter.To.IsZero() {
		filter.To = filter.To.UTC()
	}
	return filter
}

func marshalOptionalStrings(values []string) ([]byte, error) {
	if len(values) == 0 {
		return nil, nil
	}
	return json.Marshal(values)
}

func marshalOptionalMap(value map[string]any) ([]byte, error) {
	if len(value) == 0 {
		return nil, nil
	}
	return json.Marshal(value)
}

func decodeRoleJSON(raw []byte) ([]domain.Role, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	values := []string{}
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, err
	}
	return domain.NormalizeRoles(values), nil
}

func decodeAuditMap(raw []byte) (map[string]any, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	value := map[string]any{}
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func encodeAuditCursor(cursor auditCursor) (string, error) {
	if cursor.ID == 0 || cursor.CreatedAt.IsZero() {
		return "", nil
	}
	payload := map[string]string{
		"created_at": cursor.CreatedAt.UTC().Format(time.RFC3339Nano),
		"id":         strconv.FormatUint(cursor.ID, 10),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func decodeAuditCursor(value string) (auditCursor, bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return auditCursor{}, false, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return auditCursor{}, false, fmt.Errorf("%w: invalid cursor", domain.ErrInvalidAuditLogFilter)
	}
	payload := map[string]string{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return auditCursor{}, false, fmt.Errorf("%w: invalid cursor", domain.ErrInvalidAuditLogFilter)
	}
	id, err := strconv.ParseUint(payload["id"], 10, 64)
	if err != nil || id == 0 {
		return auditCursor{}, false, fmt.Errorf("%w: invalid cursor", domain.ErrInvalidAuditLogFilter)
	}
	createdAt, err := time.Parse(time.RFC3339Nano, payload["created_at"])
	if err != nil || createdAt.IsZero() {
		return auditCursor{}, false, fmt.Errorf("%w: invalid cursor", domain.ErrInvalidAuditLogFilter)
	}
	return auditCursor{CreatedAt: createdAt.UTC(), ID: id}, true, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
