package grpctransport

import (
	"context"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
	cmsv1 "github.com/example/ecommerce-platform/backend/services/cms-service/internal/gen/ecommerce/cms/v1"
)

func (s *Server) ListAuditLogs(ctx context.Context, req *cmsv1.ListAuditLogsRequest) (*cmsv1.ListAuditLogsResponse, error) {
	md := metadataFromContext(ctx)
	if !md.Actor.Authenticated() {
		return nil, grpcError(domain.ErrUnauthenticated)
	}
	if req != nil && strings.TrimSpace(req.GetSellerId()) != "" && strings.TrimSpace(req.GetSellerId()) != md.Actor.SellerID {
		return nil, grpcError(domain.ErrForbidden)
	}

	page, err := s.auditLogs.ListAuditLogs(ctx, listAuditLogsInputFromProto(req, md.Actor, md.RequestID))
	if err != nil {
		return nil, grpcError(err)
	}
	response, err := auditLogPageToProto(page)
	if err != nil {
		return nil, grpcError(err)
	}
	return response, nil
}
