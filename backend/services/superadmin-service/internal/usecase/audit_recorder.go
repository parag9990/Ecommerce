package usecase

import (
	"context"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
)

type LoggingAuditRecorder struct {
	logger logging.Logger
}

type MySQLAuditRecorder struct {
	repo AuditLogRepository
}

func NewLoggingAuditRecorder(logger logging.Logger) *LoggingAuditRecorder {
	if logger == nil {
		logger = logging.NewNop()
	}
	return &LoggingAuditRecorder{logger: logger}
}

func NewMySQLAuditRecorder(repo AuditLogRepository) (*MySQLAuditRecorder, error) {
	if repo == nil {
		return nil, domain.NewValidationError("audit log repository is required")
	}
	return &MySQLAuditRecorder{repo: repo}, nil
}

func (r *LoggingAuditRecorder) RecordAdminMutation(ctx context.Context, record domain.AuditRecord) error {
	r.logger.Info(ctx, "admin mutation audit event",
		"actor_admin_id", record.ActorAdminID,
		"action", record.Action,
		"resource_type", record.ResourceType,
		"resource_id", record.ResourceID,
		"request_id", record.RequestID,
		"session_id", record.SessionID,
		"ip_hash", record.IPHash,
		"reason", record.Reason,
		"before", record.Before,
		"after", record.After,
		"metadata", record.Metadata,
	)
	return nil
}

func (r *MySQLAuditRecorder) RecordAdminMutation(ctx context.Context, record domain.AuditRecord) error {
	return r.repo.InsertAuditLog(ctx, record)
}

func auditRecordFailure(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := domain.AsAppError(err); ok {
		return err
	}
	return domain.NewInternal("admin audit record failed", err)
}
