package httptransport

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/usecase"
)

func TestValidateCouponLegacyRouteAcceptsCartPayload(t *testing.T) {
	coupons := &capturingCouponService{
		result: domain.CouponValidationResult{
			Valid:    true,
			CouponID: "coupon_1",
			Discount: domain.Money{Amount: 100, Currency: "INR"},
		},
	}
	handler, err := NewHandler(
		noopAuthorizer{},
		noopModerationService{},
		coupons,
		noopCampaignService{},
		noopAuditLogService{},
		HandlerConfig{InternalAuthHeader: "X-Internal-Token", MaxBodyBytes: 1 << 20},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}

	body := []byte(`{
		"code":"save10",
		"guest_session_id":"guest_1",
		"cart_id":"cart_1",
		"subtotal":{"amount":1500,"currency":"INR"},
		"items":[{
			"product_id":"prod_1",
			"seller_id":"seller_1",
			"category_id":"cat_1",
			"quantity":2,
			"unit_price":{"amount":500,"currency":"INR"},
			"line_subtotal":{"amount":1000,"currency":"INR"}
		}]
	}`)
	req := httptest.NewRequest(http.MethodPost, "/internal/v1/coupons/validate", bytes.NewReader(body))
	res := httptest.NewRecorder()

	NewRouter(handler).ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
	if coupons.input.CouponCode != "save10" {
		t.Fatalf("coupon code = %q", coupons.input.CouponCode)
	}
	if coupons.input.UserID != "guest_1" {
		t.Fatalf("user id = %q", coupons.input.UserID)
	}
	if coupons.input.Currency != "INR" || coupons.input.SubtotalAmount != 1500 {
		t.Fatalf("money input = currency %q subtotal %d", coupons.input.Currency, coupons.input.SubtotalAmount)
	}
	if len(coupons.input.Items) != 1 {
		t.Fatalf("items length = %d", len(coupons.input.Items))
	}
	item := coupons.input.Items[0]
	if item.UnitAmount != 500 || item.LineSubtotalAmount != 1000 {
		t.Fatalf("item money = unit %d line %d", item.UnitAmount, item.LineSubtotalAmount)
	}
	if len(item.CategoryIDs) != 1 || item.CategoryIDs[0] != "cat_1" {
		t.Fatalf("category ids = %#v", item.CategoryIDs)
	}
}

func TestSellerCouponContractRouteUsesInternalAuth(t *testing.T) {
	handler, err := NewHandler(
		noopAuthorizer{},
		noopModerationService{},
		&capturingCouponService{},
		noopCampaignService{},
		noopAuditLogService{},
		HandlerConfig{
			InternalAuthHeader: "X-Internal-Token",
			InternalAuthToken:  "secret",
			MaxBodyBytes:       1 << 20,
		},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}

	router := NewRouter(handler)
	unauthorized := httptest.NewRecorder()
	router.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/v1/seller/coupons", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want 401", unauthorized.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/seller/coupons", nil)
	req.Header.Set("X-Internal-Token", "secret")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("authorized status = %d, body = %s", res.Code, res.Body.String())
	}
}

type noopAuthorizer struct{}

func (noopAuthorizer) Authorize(context.Context, usecase.AuthorizeInput) (usecase.AuthorizationDecision, error) {
	return usecase.AuthorizationDecision{Allowed: true}, nil
}

type noopModerationService struct{}

func (noopModerationService) SubmitProductForReview(context.Context, usecase.SubmitProductReviewInput) (domain.ProductModerationResult, error) {
	return domain.ProductModerationResult{}, nil
}

func (noopModerationService) DecideProductReview(context.Context, usecase.DecideProductReviewInput) (domain.ProductModerationResult, error) {
	return domain.ProductModerationResult{}, nil
}

func (noopModerationService) PublishApprovedProduct(context.Context, usecase.PublishProductInput) (domain.ProductModerationResult, error) {
	return domain.ProductModerationResult{}, nil
}

func (noopModerationService) UnpublishProduct(context.Context, usecase.UnpublishProductInput) (domain.ProductModerationResult, error) {
	return domain.ProductModerationResult{}, nil
}

func (noopModerationService) ListProductReviews(context.Context, usecase.ListProductReviewsInput) ([]domain.ProductModerationReview, error) {
	return nil, nil
}

type capturingCouponService struct {
	input  usecase.ValidateCouponInput
	result domain.CouponValidationResult
}

func (s *capturingCouponService) CreateCoupon(context.Context, usecase.CreateCouponInput) (domain.Coupon, error) {
	return domain.Coupon{}, nil
}

func (s *capturingCouponService) UpdateCoupon(context.Context, usecase.UpdateCouponInput) (domain.Coupon, error) {
	return domain.Coupon{}, nil
}

func (s *capturingCouponService) DisableCoupon(context.Context, usecase.DisableCouponInput) (domain.Coupon, error) {
	return domain.Coupon{}, nil
}

func (s *capturingCouponService) ListCoupons(context.Context, usecase.ListCouponsInput) ([]domain.Coupon, error) {
	return nil, nil
}

func (s *capturingCouponService) ValidateCoupon(_ context.Context, input usecase.ValidateCouponInput) (domain.CouponValidationResult, error) {
	s.input = input
	return s.result, nil
}

func (s *capturingCouponService) RecordRedemption(context.Context, usecase.RecordCouponRedemptionInput) (domain.CouponRedemption, error) {
	return domain.CouponRedemption{}, nil
}

type noopCampaignService struct{}

func (noopCampaignService) CreateCampaign(context.Context, usecase.CreateCampaignInput) (domain.Campaign, error) {
	return domain.Campaign{}, nil
}

func (noopCampaignService) UpdateCampaign(context.Context, usecase.UpdateCampaignInput) (domain.Campaign, error) {
	return domain.Campaign{}, nil
}

func (noopCampaignService) DisableCampaign(context.Context, usecase.DisableCampaignInput) (domain.Campaign, error) {
	return domain.Campaign{}, nil
}

func (noopCampaignService) ListCampaigns(context.Context, usecase.ListCampaignsInput) ([]domain.Campaign, error) {
	return nil, nil
}

func (noopCampaignService) ValidateCampaign(context.Context, usecase.ValidateCampaignInput) (domain.CampaignValidationResult, error) {
	return domain.CampaignValidationResult{}, nil
}

type noopAuditLogService struct{}

func (noopAuditLogService) ListAuditLogs(context.Context, usecase.ListAuditLogsInput) (domain.AuditLogPage, error) {
	return domain.AuditLogPage{}, nil
}
