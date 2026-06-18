package usecase

import (
	"context"
	"errors"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
)

type AuditLogAuthorizer interface {
	RequirePermission(ctx context.Context, actor domain.AdminActor, permission domain.Permission) error
}

type AuditLogRepository interface {
	InsertAuditLog(ctx context.Context, record domain.AuditRecord) error
	ListAuditLogs(ctx context.Context, req domain.AuditLogListRequest) (domain.AuditLogListResponse, error)
}

type AuditLogService struct {
	authorizer AuditLogAuthorizer
	repo       AuditLogRepository
	logger     logging.Logger
}

func NewAuditLogService(authorizer AuditLogAuthorizer, repo AuditLogRepository, logger logging.Logger) (*AuditLogService, error) {
	if authorizer == nil {
		return nil, errors.New("audit log service requires authorizer")
	}
	if repo == nil {
		return nil, errors.New("audit log service requires repository")
	}
	if logger == nil {
		logger = logging.NewNop()
	}
	return &AuditLogService{authorizer: authorizer, repo: repo, logger: logger}, nil
}

func (s *AuditLogService) ListAuditLogs(ctx context.Context, req domain.AuditLogListRequest) (domain.AuditLogListResponse, error) {
	actor, err := actorFromContext(ctx)
	if err != nil {
		return domain.AuditLogListResponse{}, err
	}
	if err := s.authorizer.RequirePermission(ctx, actor, domain.PermissionAuditLogsRead); err != nil {
		return domain.AuditLogListResponse{}, err
	}

	normalized, err := req.Normalize()
	if err != nil {
		return domain.AuditLogListResponse{}, err
	}
	response, err := s.repo.ListAuditLogs(ctx, normalized)
	if err != nil {
		s.logger.Error(ctx, "audit log list failed",
			"admin_id", actor.AdminID,
			"request_id", actor.RequestID,
			"error", err,
		)
		return domain.AuditLogListResponse{}, err
	}
	return response, nil
}
