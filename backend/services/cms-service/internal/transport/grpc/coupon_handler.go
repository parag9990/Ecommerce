package grpctransport

import (
	"context"

	cmsv1 "github.com/example/ecommerce-platform/backend/services/cms-service/internal/gen/ecommerce/cms/v1"
)

var validateCouponCallers = allowedCallers("cart-service", "order-service")

func (s *Server) ValidateCoupon(ctx context.Context, req *cmsv1.ValidateCouponRequest) (*cmsv1.ValidateCouponResponse, error) {
	md, err := s.requireInternalCaller(ctx, validateCouponCallers)
	if err != nil {
		return nil, err
	}
	input, err := validateCouponInputFromProto(req, md.RequestID)
	if err != nil {
		return nil, err
	}
	result, err := s.coupons.ValidateCoupon(ctx, input)
	if err != nil {
		return nil, grpcError(err)
	}
	return validateCouponResponseToProto(result), nil
}
