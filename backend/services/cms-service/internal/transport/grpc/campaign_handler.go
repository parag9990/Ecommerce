package grpctransport

import (
	"context"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
	cmsv1 "github.com/example/ecommerce-platform/backend/services/cms-service/internal/gen/ecommerce/cms/v1"
)

var campaignReadInternalCallers = allowedCallers("api-gateway", "cart-service", "order-service")

func (s *Server) GetCampaign(ctx context.Context, req *cmsv1.GetCampaignRequest) (*cmsv1.GetCampaignResponse, error) {
	md, internal, err := s.internalCallerIfAuthenticated(ctx, campaignReadInternalCallers)
	if err != nil {
		return nil, err
	}
	input := getCampaignInputFromProto(req, md.Actor, md.RequestID)
	if internal {
		input.InternalCaller = md.ServiceName
	}
	result, err := s.campaigns.GetCampaign(ctx, input)
	if err != nil {
		return nil, grpcError(err)
	}
	campaign, err := campaignReadToProto(result)
	if err != nil {
		return nil, grpcError(err)
	}
	return &cmsv1.GetCampaignResponse{Campaign: campaign}, nil
}

func (s *Server) ListCampaigns(ctx context.Context, req *cmsv1.ListCampaignsRequest) (*cmsv1.ListCampaignsResponse, error) {
	md := metadataFromContext(ctx)
	if !md.Actor.Authenticated() {
		return nil, grpcError(domain.ErrUnauthenticated)
	}
	input := listCampaignsInputFromProto(req, md.Actor, md.RequestID)
	results, err := s.campaigns.ListCampaignReads(ctx, input)
	if err != nil {
		return nil, grpcError(err)
	}

	campaigns := make([]*cmsv1.Campaign, 0, len(results))
	for _, result := range results {
		campaign, err := campaignReadToProto(result)
		if err != nil {
			return nil, grpcError(err)
		}
		campaigns = append(campaigns, campaign)
	}
	limit, offset := pageLimitOffset(nil)
	if req != nil {
		limit, offset = pageLimitOffset(req.GetPage())
	}
	return &cmsv1.ListCampaignsResponse{
		Campaigns: campaigns,
		Page: &cmsv1.PageResponse{
			Limit:  int32(limit),
			Offset: int32(offset),
			Count:  int32(len(campaigns)),
		},
	}, nil
}
