package grpctransport

import (
	"context"

	cmsv1 "github.com/example/ecommerce-platform/backend/services/cms-service/internal/gen/ecommerce/cms/v1"
)

var sellerSettingsInternalCallers = allowedCallers("api-gateway")

func (s *Server) GetSellerSettings(ctx context.Context, req *cmsv1.GetSellerSettingsRequest) (*cmsv1.GetSellerSettingsResponse, error) {
	md, internal, err := s.internalCallerIfAuthenticated(ctx, sellerSettingsInternalCallers)
	if err != nil {
		return nil, err
	}
	input := getSellerSettingsInputFromProto(req, md.Actor, md.RequestID)
	if internal {
		input.InternalCaller = md.ServiceName
	}
	settings, err := s.settings.GetSellerSettings(ctx, input)
	if err != nil {
		return nil, grpcError(err)
	}
	out, err := sellerSettingsToProto(settings)
	if err != nil {
		return nil, grpcError(err)
	}
	return &cmsv1.GetSellerSettingsResponse{Settings: out}, nil
}
