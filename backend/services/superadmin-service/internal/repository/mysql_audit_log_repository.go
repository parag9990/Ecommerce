package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"ecommerce/superadmin-service/internal/domain"
)

type MySQLAuditLogRepository struct {
	db *sql.DB
}

func NewMySQLAuditLogRepository(db *sql.DB) (*MySQLAuditLogRepository, error) {
	if db == nil {
		return nil, errors.New("mysql audit log repository requires db")
	}
	return &MySQLAuditLogRepository{db: db}, nil
}

func (r *MySQLAuditLogRepository) InsertAuditLog(ctx context.Context, record domain.AuditRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}

	auditID, err := newAuditID()
	if err != nil {
		return err
	}
	beforeJSON, err := nullableAuditJSON(record.Before)
	if err != nil {
		return err
	}
	afterJSON, err := nullableAuditJSON(record.After)
	if err != nil {
		return err
	}
	metadataJSON, err := nullableAuditJSON(record.Metadata)
	if err != nil {
		return err
	}

	const query = `
		INSERT INTO admin_audit_logs (
		  audit_id,
		  actor_admin_id,
		  action,
		  resource_type,
		  resource_id,
		  request_id,
		  session_id,
		  ip_hash,
		  reason,
		  before_json,
		  after_json,
		  metadata_json
		) VALUES (?, ?, ?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?, ?)`

	if _, err := r.db.ExecContext(
		ctx,
		query,
		auditID,
		strings.TrimSpace(record.ActorAdminID),
		strings.TrimSpace(record.Action),
		strings.TrimSpace(record.ResourceType),
		strings.TrimSpace(record.ResourceID),
		strings.TrimSpace(record.RequestID),
		strings.TrimSpace(record.SessionID),
		strings.TrimSpace(record.IPHash),
		strings.TrimSpace(record.Reason),
		beforeJSON,
		afterJSON,
		metadataJSON,
	); err != nil {
		return fmt.Errorf("insert admin audit log: %w", err)
	}
	return nil
}

func (r *MySQLAuditLogRepository) ListAuditLogs(ctx context.Context, req domain.AuditLogListRequest) (domain.AuditLogListResponse, error) {
	req, err := req.Normalize()
	if err != nil {
		return domain.AuditLogListResponse{}, err
	}
	cursorID, err := domain.ParseAuditLogCursor(req.Pagination.Cursor)
	if err != nil {
		return domain.AuditLogListResponse{}, err
	}

	query, args := auditLogListQuery(req, cursorID)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return domain.AuditLogListResponse{}, fmt.Errorf("list admin audit logs: %w", err)
	}
	defer rows.Close()

	limit := req.Pagination.PageSize
	logs := make([]domain.AuditLog, 0, limit)
	var nextCursor string
	for rows.Next() {
		log, err := scanAuditLog(rows)
		if err != nil {
			return domain.AuditLogListResponse{}, err
		}
		if len(logs) == limit {
			if len(logs) > 0 {
				nextCursor = domain.NewAuditLogCursor(logs[len(logs)-1].SequenceID)
			}
			continue
		}
		logs = append(logs, log)
	}
	if err := rows.Err(); err != nil {
		return domain.AuditLogListResponse{}, fmt.Errorf("iterate admin audit logs: %w", err)
	}

	return domain.AuditLogListResponse{Logs: logs, NextCursor: nextCursor}, nil
}

type UnavailableAuditLogRepository struct {
	message string
}

func NewUnavailableAuditLogRepository(message string) *UnavailableAuditLogRepository {
	if strings.TrimSpace(message) == "" {
		message = "admin audit log storage is not configured"
	}
	return &UnavailableAuditLogRepository{message: message}
}

func (r *UnavailableAuditLogRepository) InsertAuditLog(ctx context.Context, record domain.AuditRecord) error {
	return domain.NewDownstreamUnavailable(r.message, nil)
}

func (r *UnavailableAuditLogRepository) ListAuditLogs(ctx context.Context, req domain.AuditLogListRequest) (domain.AuditLogListResponse, error) {
	return domain.AuditLogListResponse{}, domain.NewDownstreamUnavailable(r.message, nil)
}

func auditLogListQuery(req domain.AuditLogListRequest, cursorID uint64) (string, []any) {
	var builder strings.Builder
	args := make([]any, 0, 12)

	builder.WriteString(`
		SELECT id,
		       audit_id,
		       actor_admin_id,
		       action,
		       resource_type,
		       resource_id,
		       request_id,
		       session_id,
		       ip_hash,
		       reason,
		       before_json,
		       after_json,
		       metadata_json,
		       created_at
		FROM admin_audit_logs
		WHERE 1 = 1`)

	addAuditCondition := func(condition string, value any) {
		builder.WriteString("\n		  AND ")
		builder.WriteString(condition)
		args = append(args, value)
	}
	if req.ActorAdminID != "" {
		addAuditCondition("actor_admin_id = ?", req.ActorAdminID)
	}
	if req.Action != "" {
		addAuditCondition("action = ?", req.Action)
	}
	if req.ResourceType != "" {
		addAuditCondition("resource_type = ?", req.ResourceType)
	}
	if req.ResourceID != "" {
		addAuditCondition("resource_id = ?", req.ResourceID)
	}
	if req.RequestID != "" {
		addAuditCondition("request_id = ?", req.RequestID)
	}
	if !req.From.IsZero() {
		addAuditCondition("created_at >= ?", req.From.UTC())
	}
	if !req.To.IsZero() {
		addAuditCondition("created_at <= ?", req.To.UTC())
	}
	if cursorID > 0 {
		addAuditCondition("id < ?", cursorID)
	}

	builder.WriteString("\n		ORDER BY id DESC\n		LIMIT ? OFFSET ?")
	args = append(args, req.Pagination.PageSize+1, auditLogOffset(req, cursorID))
	return builder.String(), args
}

func auditLogOffset(req domain.AuditLogListRequest, cursorID uint64) int {
	if cursorID > 0 || req.Pagination.Page <= 1 {
		return 0
	}
	return (req.Pagination.Page - 1) * req.Pagination.PageSize
}

type auditLogScanner interface {
	Scan(dest ...any) error
}

func scanAuditLog(scanner auditLogScanner) (domain.AuditLog, error) {
	var log domain.AuditLog
	var requestID sql.NullString
	var sessionID sql.NullString
	var ipHash sql.NullString
	var reason sql.NullString
	var beforeJSON sql.NullString
	var afterJSON sql.NullString
	var metadataJSON sql.NullString
	var createdAt sql.NullTime

	if err := scanner.Scan(
		&log.SequenceID,
		&log.AuditID,
		&log.ActorAdminID,
		&log.Action,
		&log.ResourceType,
		&log.ResourceID,
		&requestID,
		&sessionID,
		&ipHash,
		&reason,
		&beforeJSON,
		&afterJSON,
		&metadataJSON,
		&createdAt,
	); err != nil {
		return domain.AuditLog{}, fmt.Errorf("scan admin audit log: %w", err)
	}

	log.RequestID = requestID.String
	log.SessionID = sessionID.String
	log.IPHash = ipHash.String
	log.Reason = reason.String
	if createdAt.Valid {
		log.CreatedAt = createdAt.Time
	}

	var err error
	if log.Before, err = decodeAuditJSON(beforeJSON, "before_json"); err != nil {
		return domain.AuditLog{}, err
	}
	if log.After, err = decodeAuditJSON(afterJSON, "after_json"); err != nil {
		return domain.AuditLog{}, err
	}
	if log.Metadata, err = decodeAuditJSON(metadataJSON, "metadata_json"); err != nil {
		return domain.AuditLog{}, err
	}
	return log, nil
}

func nullableAuditJSON(values map[string]string) (any, error) {
	safe := domain.SafeAuditSnapshot(values)
	if len(safe) == 0 {
		return nil, nil
	}
	encoded, err := json.Marshal(safe)
	if err != nil {
		return nil, domain.NewValidationError("audit snapshot must be JSON serializable")
	}
	return string(encoded), nil
}

func decodeAuditJSON(value sql.NullString, field string) (map[string]string, error) {
	if !value.Valid || strings.TrimSpace(value.String) == "" {
		return nil, nil
	}
	var decoded map[string]string
	if err := json.Unmarshal([]byte(value.String), &decoded); err != nil {
		return nil, domain.NewInternal("admin audit log "+field+" is invalid JSON", err)
	}
	return decoded, nil
}

func newAuditID() (string, error) {
	var randomBytes [16]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return "", domain.NewInternal("generate audit id failed", err)
	}
	return "aud_" + hex.EncodeToString(randomBytes[:]), nil
}
