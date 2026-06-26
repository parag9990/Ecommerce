package usecase

import (
	"context"
	"errors"
	"fmt"

	"ecommerce/superadmin-service/internal/domain"
	"ecommerce/superadmin-service/internal/logging"
)

type AuditLogAuthorizer interface {
	RequirePermission(ctx context.Context, actor domain.AdminActor, permission domain.Permission) error
	RequireHighRiskPermission(ctx context.Context, actor domain.AdminActor, permission domain.Permission, reason string) error
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

func (s *AuditLogService) ExportAuditLogs(ctx context.Context, req domain.AuditLogListRequest, reason string) (domain.AuditLogListResponse, error) {
	actor, err := actorFromContext(ctx)
	if err != nil {
		return domain.AuditLogListResponse{}, err
	}
	normalizedReason, err := domain.NormalizeMutationReason(reason)
	if err != nil {
		return domain.AuditLogListResponse{}, err
	}
	if err := s.authorizer.RequireHighRiskPermission(ctx, actor, domain.PermissionAuditLogsExport, normalizedReason); err != nil {
		return domain.AuditLogListResponse{}, err
	}

	normalized, err := req.Normalize()
	if err != nil {
		return domain.AuditLogListResponse{}, err
	}
	if normalized.PageSize < 1 || normalized.PageSize > domain.MaxAuditLogPageSize {
		normalized.PageSize = domain.MaxAuditLogPageSize
	}
	response, err := s.repo.ListAuditLogs(ctx, normalized)
	if err != nil {
		s.logger.Error(ctx, "audit log export list failed",
			"admin_id", actor.AdminID,
			"request_id", actor.RequestID,
			"error", err,
		)
		return domain.AuditLogListResponse{}, err
	}
	if err := s.repo.InsertAuditLog(ctx, domain.AuditRecord{
		ActorAdminID: actor.AdminID,
		Action:       "audit_logs.export",
		ResourceType: "audit_log",
		ResourceID:   "export",
		RequestID:    actor.RequestID,
		SessionID:    actor.SessionID,
		IPHash:       actor.IPHash,
		Reason:       normalizedReason,
		Metadata: map[string]string{
			"row_count": fmt.Sprintf("%d", len(response.Logs)),
		},
	}); err != nil {
		s.logger.Error(ctx, "audit log export audit record failed",
			"admin_id", actor.AdminID,
			"request_id", actor.RequestID,
			"error", err,
		)
		return domain.AuditLogListResponse{}, err
	}
	return response, nil
}
