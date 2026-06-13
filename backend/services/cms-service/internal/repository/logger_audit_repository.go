package repository

import (
	"context"
	"log/slog"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
)

type LoggerAuditRecorder struct {
	logger *slog.Logger
}

func NewLoggerAuditRecorder(logger *slog.Logger) *LoggerAuditRecorder {
	if logger == nil {
		logger = slog.Default()
	}
	return &LoggerAuditRecorder{logger: logger}
}

func (r *LoggerAuditRecorder) Record(ctx context.Context, event domain.AuditEvent) error {
	r.logger.InfoContext(ctx, "cms.audit",
		slog.String("audit_id", event.AuditID),
		slog.String("actor_user_id", event.ActorUserID),
		slog.String("actor_seller_id", event.ActorSellerID),
		slog.Any("actor_roles", domain.RoleStrings(event.ActorRoles)),
		slog.String("action", event.Action.String()),
		slog.String("resource_type", event.ResourceType),
		slog.String("resource_id", event.ResourceID),
		slog.String("resource_seller_id", event.ResourceSellerID),
		slog.String("request_id", event.RequestID),
		slog.String("decision", string(event.Decision)),
		slog.Any("before", event.Before),
		slog.Any("after", event.After),
		slog.Time("created_at", event.CreatedAt),
	)
	return nil
}
