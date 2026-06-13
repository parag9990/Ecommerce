package httptransport

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/usecase"
)

type Authorizer interface {
	Authorize(ctx context.Context, input usecase.AuthorizeInput) (usecase.AuthorizationDecision, error)
}

type ProductModerationService interface {
	SubmitProductForReview(ctx context.Context, input usecase.SubmitProductReviewInput) (domain.ProductModerationResult, error)
	DecideProductReview(ctx context.Context, input usecase.DecideProductReviewInput) (domain.ProductModerationResult, error)
	PublishApprovedProduct(ctx context.Context, input usecase.PublishProductInput) (domain.ProductModerationResult, error)
	UnpublishProduct(ctx context.Context, input usecase.UnpublishProductInput) (domain.ProductModerationResult, error)
	ListProductReviews(ctx context.Context, input usecase.ListProductReviewsInput) ([]domain.ProductModerationReview, error)
}

type CouponService interface {
	CreateCoupon(ctx context.Context, input usecase.CreateCouponInput) (domain.Coupon, error)
	UpdateCoupon(ctx context.Context, input usecase.UpdateCouponInput) (domain.Coupon, error)
	DisableCoupon(ctx context.Context, input usecase.DisableCouponInput) (domain.Coupon, error)
	ListCoupons(ctx context.Context, input usecase.ListCouponsInput) ([]domain.Coupon, error)
	ValidateCoupon(ctx context.Context, input usecase.ValidateCouponInput) (domain.CouponValidationResult, error)
	RecordRedemption(ctx context.Context, input usecase.RecordCouponRedemptionInput) (domain.CouponRedemption, error)
}

type CampaignService interface {
	CreateCampaign(ctx context.Context, input usecase.CreateCampaignInput) (domain.Campaign, error)
	UpdateCampaign(ctx context.Context, input usecase.UpdateCampaignInput) (domain.Campaign, error)
	DisableCampaign(ctx context.Context, input usecase.DisableCampaignInput) (domain.Campaign, error)
	ListCampaigns(ctx context.Context, input usecase.ListCampaignsInput) ([]domain.Campaign, error)
	ValidateCampaign(ctx context.Context, input usecase.ValidateCampaignInput) (domain.CampaignValidationResult, error)
}

type SellerAnalyticsService interface {
	GetSellerAnalytics(ctx context.Context, input usecase.GetSellerAnalyticsInput) (domain.SellerAnalytics, error)
}

type AuditLogService interface {
	ListAuditLogs(ctx context.Context, input usecase.ListAuditLogsInput) (domain.AuditLogPage, error)
}

type HandlerConfig struct {
	InternalAuthHeader string
	InternalAuthToken  string
	MaxBodyBytes       int64
}

type Handler struct {
	authorizer         Authorizer
	moderation         ProductModerationService
	coupons            CouponService
	campaigns          CampaignService
	analytics          SellerAnalyticsService
	auditLogs          AuditLogService
	internalAuthHeader string
	internalAuthToken  string
	maxBodyBytes       int64
	logger             *slog.Logger
}

func NewHandler(authorizer Authorizer, moderation ProductModerationService, coupons CouponService, campaigns CampaignService, auditLogs AuditLogService, cfg HandlerConfig, logger *slog.Logger) (*Handler, error) {
	if authorizer == nil {
		return nil, errors.New("authorizer is required")
	}
	if moderation == nil {
		return nil, errors.New("product moderation service is required")
	}
	if coupons == nil {
		return nil, errors.New("coupon service is required")
	}
	if campaigns == nil {
		return nil, errors.New("campaign service is required")
	}
	if auditLogs == nil {
		return nil, errors.New("audit log service is required")
	}
	if strings.TrimSpace(cfg.InternalAuthHeader) == "" {
		return nil, errors.New("internal auth header is required")
	}
	if cfg.MaxBodyBytes <= 0 {
		cfg.MaxBodyBytes = 1 << 20
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		authorizer:         authorizer,
		moderation:         moderation,
		coupons:            coupons,
		campaigns:          campaigns,
		auditLogs:          auditLogs,
		internalAuthHeader: strings.TrimSpace(cfg.InternalAuthHeader),
		internalAuthToken:  cfg.InternalAuthToken,
		maxBodyBytes:       cfg.MaxBodyBytes,
		logger:             logger,
	}, nil
}

func (h *Handler) SetSellerAnalyticsService(analytics SellerAnalyticsService) error {
	if analytics == nil {
		return errors.New("seller analytics service is required")
	}
	h.analytics = analytics
	return nil
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.Handle("/internal/v1/cms/authorize", h.requireInternalAuth(http.HandlerFunc(h.handleAuthorize)))
	mux.Handle("/internal/v1/cms/permissions/catalog", h.requireInternalAuth(http.HandlerFunc(h.handlePermissionCatalog)))
	mux.Handle("GET /api/v1/seller/dashboard/summary", h.requireInternalAuth(http.HandlerFunc(h.handleGetSellerAnalytics)))
	mux.Handle("GET /internal/v1/cms/seller/dashboard/summary", h.requireInternalAuth(http.HandlerFunc(h.handleGetSellerAnalytics)))
	mux.Handle("GET /api/v1/seller/audit-logs", h.requireInternalAuth(http.HandlerFunc(h.handleListAuditLogs)))
	mux.Handle("GET /internal/v1/cms/seller/audit-logs", h.requireInternalAuth(http.HandlerFunc(h.handleListAuditLogs)))
	mux.Handle("POST /internal/v1/cms/seller/products/{product_id}/submit-review", h.requireInternalAuth(http.HandlerFunc(h.handleSubmitProductReview)))
	mux.Handle("POST /internal/v1/cms/seller/products/{product_id}/publish", h.requireInternalAuth(http.HandlerFunc(h.handlePublishProduct)))
	mux.Handle("POST /internal/v1/cms/seller/products/{product_id}/unpublish", h.requireInternalAuth(http.HandlerFunc(h.handleUnpublishProduct)))
	mux.Handle("GET /internal/v1/cms/admin/catalog/reviews", h.requireInternalAuth(http.HandlerFunc(h.handleListProductReviews)))
	mux.Handle("POST /internal/v1/cms/admin/catalog/reviews/{review_id}/decision", h.requireInternalAuth(http.HandlerFunc(h.handleDecideProductReview)))
	mux.Handle("POST /internal/v1/cms/admin/catalog/products/{product_id}/unpublish", h.requireInternalAuth(http.HandlerFunc(h.handleForceUnpublishProduct)))
	mux.Handle("GET /internal/v1/cms/seller/coupons", h.requireInternalAuth(http.HandlerFunc(h.handleListCoupons)))
	mux.Handle("POST /internal/v1/cms/seller/coupons", h.requireInternalAuth(http.HandlerFunc(h.handleCreateCoupon)))
	mux.Handle("PATCH /internal/v1/cms/seller/coupons/{coupon_id}", h.requireInternalAuth(http.HandlerFunc(h.handleUpdateCoupon)))
	mux.Handle("POST /internal/v1/cms/seller/coupons/{coupon_id}/disable", h.requireInternalAuth(http.HandlerFunc(h.handleDisableCoupon)))
	mux.Handle("POST /internal/v1/cms/coupons/validate", h.requireInternalAuth(http.HandlerFunc(h.handleValidateCoupon)))
	mux.Handle("POST /internal/v1/cms/coupons/redemptions", h.requireInternalAuth(http.HandlerFunc(h.handleRecordCouponRedemption)))
	mux.Handle("GET /internal/v1/cms/seller/campaigns", h.requireInternalAuth(http.HandlerFunc(h.handleListCampaigns)))
	mux.Handle("POST /internal/v1/cms/seller/campaigns", h.requireInternalAuth(http.HandlerFunc(h.handleCreateCampaign)))
	mux.Handle("PATCH /internal/v1/cms/seller/campaigns/{campaign_id}", h.requireInternalAuth(http.HandlerFunc(h.handleUpdateCampaign)))
	mux.Handle("POST /internal/v1/cms/seller/campaigns/{campaign_id}/disable", h.requireInternalAuth(http.HandlerFunc(h.handleDisableCampaign)))
	mux.Handle("POST /internal/v1/cms/campaigns/validate", h.requireInternalAuth(http.HandlerFunc(h.handleValidateCampaign)))
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	writeJSON(w, http.StatusOK, successResponse{Success: true})
}

func (h *Handler) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req authorizeRequest
	if err := decodeJSON(w, r, h.maxBodyBytes, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	decision, err := h.authorizer.Authorize(r.Context(), usecase.AuthorizeInput{
		Actor:              actorFromRequest(r),
		ResourceSellerID:   req.ResourceSellerID,
		RequiredPermission: domain.Permission(req.RequiredPermission),
		ResourceType:       req.ResourceType,
		ResourceID:         req.ResourceID,
		TargetRole:         domain.Role(req.TargetRole),
		RequestID:          requestID(r),
	})

	response := authorizeResponseFromDecision(decision)
	if err != nil {
		status, code, message := mapAuthorizationError(err)
		h.logger.InfoContext(r.Context(), "cms.authorization.denied",
			slog.String("reason", decision.Reason),
			slog.String("permission", req.RequiredPermission),
			slog.String("request_id", decision.RequestID),
		)
		writeJSON(w, status, responseWithError(response, code, message))
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) handlePermissionCatalog(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	writeJSON(w, http.StatusOK, permissionCatalogDTO())
}

func (h *Handler) handleGetSellerAnalytics(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	if h.analytics == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "analytics_temporarily_unavailable", "Seller analytics are temporarily unavailable")
		return
	}

	from, err := optionalAnalyticsTimeQuery(r, "from")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: apiError{Code: "invalid_date_range", Message: err.Error()}})
		return
	}
	to, err := optionalAnalyticsTimeQuery(r, "to")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: apiError{Code: "invalid_date_range", Message: err.Error()}})
		return
	}
	topProductsLimit, err := optionalAnalyticsLimitQuery(r, "top_products_limit")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: apiError{Code: "invalid_top_products_limit", Message: err.Error()}})
		return
	}

	result, err := h.analytics.GetSellerAnalytics(r.Context(), usecase.GetSellerAnalyticsInput{
		Actor:            actorFromRequest(r),
		From:             from,
		To:               to,
		Currency:         r.URL.Query().Get("currency"),
		TopProductsLimit: topProductsLimit,
		RequestID:        requestID(r),
	})
	if err != nil {
		h.writeAnalyticsError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, sellerAnalyticsResponseFromDomain(result))
}

func (h *Handler) handleListAuditLogs(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	from, err := optionalAuditTimeQuery(r, "from")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: apiError{Code: "invalid_date_range", Message: err.Error()}})
		return
	}
	to, err := optionalAuditTimeQuery(r, "to")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: apiError{Code: "invalid_date_range", Message: err.Error()}})
		return
	}
	pageSize, err := optionalAuditPageSizeQuery(r, "page_size")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: apiError{Code: "invalid_page_size", Message: err.Error()}})
		return
	}

	page, err := h.auditLogs.ListAuditLogs(r.Context(), usecase.ListAuditLogsInput{
		Actor:        actorFromRequest(r),
		ActorUserID:  r.URL.Query().Get("actor_user_id"),
		Action:       domain.Permission(r.URL.Query().Get("action")),
		ResourceType: r.URL.Query().Get("resource_type"),
		ResourceID:   r.URL.Query().Get("resource_id"),
		From:         from,
		To:           to,
		PageSize:     pageSize,
		Cursor:       r.URL.Query().Get("cursor"),
		RequestID:    requestID(r),
	})
	if err != nil {
		h.writeAuditLogError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, auditLogListResponseFromDomain(page))
}

func (h *Handler) handleSubmitProductReview(w http.ResponseWriter, r *http.Request) {
	result, err := h.moderation.SubmitProductForReview(r.Context(), usecase.SubmitProductReviewInput{
		Actor:     actorFromRequest(r),
		ProductID: r.PathValue("product_id"),
		RequestID: requestID(r),
	})
	if err != nil {
		h.writeModerationError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, moderationResponseFromResult(result))
}

func (h *Handler) handleDecideProductReview(w http.ResponseWriter, r *http.Request) {
	var req reviewDecisionRequest
	if err := decodeJSON(w, r, h.maxBodyBytes, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	result, err := h.moderation.DecideProductReview(r.Context(), usecase.DecideProductReviewInput{
		Actor:     actorFromRequest(r),
		ReviewID:  r.PathValue("review_id"),
		Decision:  domain.ModerationDecision(req.Decision),
		Reason:    req.Reason,
		RequestID: requestID(r),
	})
	if err != nil {
		h.writeModerationError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, moderationResponseFromResult(result))
}

func (h *Handler) handlePublishProduct(w http.ResponseWriter, r *http.Request) {
	result, err := h.moderation.PublishApprovedProduct(r.Context(), usecase.PublishProductInput{
		Actor:     actorFromRequest(r),
		ProductID: r.PathValue("product_id"),
		RequestID: requestID(r),
	})
	if err != nil {
		h.writeModerationError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, moderationResponseFromResult(result))
}

func (h *Handler) handleUnpublishProduct(w http.ResponseWriter, r *http.Request) {
	req := unpublishRequest{}
	if r.Body != nil && r.ContentLength != 0 {
		if err := decodeJSON(w, r, h.maxBodyBytes, &req); err != nil {
			writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
	}
	result, err := h.moderation.UnpublishProduct(r.Context(), usecase.UnpublishProductInput{
		Actor:     actorFromRequest(r),
		ProductID: r.PathValue("product_id"),
		Reason:    req.Reason,
		RequestID: requestID(r),
	})
	if err != nil {
		h.writeModerationError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, moderationResponseFromResult(result))
}

func (h *Handler) handleForceUnpublishProduct(w http.ResponseWriter, r *http.Request) {
	var req unpublishRequest
	if err := decodeJSON(w, r, h.maxBodyBytes, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	result, err := h.moderation.UnpublishProduct(r.Context(), usecase.UnpublishProductInput{
		Actor:     actorFromRequest(r),
		ProductID: r.PathValue("product_id"),
		Reason:    req.Reason,
		Force:     true,
		RequestID: requestID(r),
	})
	if err != nil {
		h.writeModerationError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, moderationResponseFromResult(result))
}

func (h *Handler) handleListProductReviews(w http.ResponseWriter, r *http.Request) {
	limit := intQuery(r, "limit", 50)
	offset := intQuery(r, "offset", 0)
	reviews, err := h.moderation.ListProductReviews(r.Context(), usecase.ListProductReviewsInput{
		Actor:     actorFromRequest(r),
		Status:    domain.ProductReviewStatus(r.URL.Query().Get("status")),
		SellerID:  r.URL.Query().Get("seller_id"),
		Limit:     limit,
		Offset:    offset,
		RequestID: requestID(r),
	})
	if err != nil {
		h.writeModerationError(w, r, err)
		return
	}
	response := reviewListResponse{
		Reviews: make([]reviewResponse, 0, len(reviews)),
		Limit:   limit,
		Offset:  offset,
	}
	for _, review := range reviews {
		response.Reviews = append(response.Reviews, reviewResponseFromDomain(review))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) handleListCoupons(w http.ResponseWriter, r *http.Request) {
	limit := intQuery(r, "limit", 50)
	offset := intQuery(r, "offset", 0)
	coupons, err := h.coupons.ListCoupons(r.Context(), usecase.ListCouponsInput{
		Actor:     actorFromRequest(r),
		Status:    domain.CouponStatus(r.URL.Query().Get("status")),
		Code:      r.URL.Query().Get("code"),
		Limit:     limit,
		Offset:    offset,
		RequestID: requestID(r),
	})
	if err != nil {
		h.writeCouponError(w, r, err)
		return
	}
	response := couponListResponse{
		Coupons: make([]couponResponse, 0, len(coupons)),
		Limit:   limit,
		Offset:  offset,
	}
	for _, coupon := range coupons {
		response.Coupons = append(response.Coupons, couponResponseFromDomain(coupon))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) handleCreateCoupon(w http.ResponseWriter, r *http.Request) {
	var req couponCreateRequest
	if err := decodeJSON(w, r, h.maxBodyBytes, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	coupon, err := h.coupons.CreateCoupon(r.Context(), usecase.CreateCouponInput{
		Actor:             actorFromRequest(r),
		Code:              req.Code,
		DiscountType:      domain.DiscountType(req.DiscountType),
		DiscountValue:     req.DiscountValue,
		MaxDiscountAmount: amountPtr(req.MaxDiscountAmount),
		MinCartAmount:     amountOrZero(req.MinCartAmount),
		Currency:          currencyFromRequest(req.Currency, req.MinCartAmount, req.MaxDiscountAmount),
		UsageLimit:        req.UsageLimit,
		PerUserLimit:      req.PerUserLimit,
		Status:            domain.CouponStatus(req.Status),
		StartsAt:          req.StartsAt,
		EndsAt:            req.EndsAt,
		Rules:             couponRulesFromRequest(req.Rules),
		RequestID:         requestID(r),
	})
	if err != nil {
		h.writeCouponError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, couponResponseFromDomain(coupon))
}

func (h *Handler) handleUpdateCoupon(w http.ResponseWriter, r *http.Request) {
	var req couponPatchRequest
	if err := decodeJSON(w, r, h.maxBodyBytes, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	var discountType *domain.DiscountType
	if req.DiscountType != nil {
		value := domain.DiscountType(*req.DiscountType)
		discountType = &value
	}
	var status *domain.CouponStatus
	if req.Status != nil {
		value := domain.CouponStatus(*req.Status)
		status = &value
	}
	var minCartAmount *int64
	if req.MinCartAmount != nil {
		value := req.MinCartAmount.Amount
		minCartAmount = &value
	}
	var rules *[]domain.CouponRule
	if req.Rules != nil {
		value := couponRulesFromRequest(*req.Rules)
		rules = &value
	}

	coupon, err := h.coupons.UpdateCoupon(r.Context(), usecase.UpdateCouponInput{
		Actor:             actorFromRequest(r),
		CouponID:          r.PathValue("coupon_id"),
		Code:              req.Code,
		DiscountType:      discountType,
		DiscountValue:     req.DiscountValue,
		MaxDiscountAmount: amountPtr(req.MaxDiscountAmount),
		MinCartAmount:     minCartAmount,
		Currency:          req.Currency,
		UsageLimit:        req.UsageLimit,
		PerUserLimit:      req.PerUserLimit,
		Status:            status,
		StartsAt:          req.StartsAt,
		EndsAt:            req.EndsAt,
		Rules:             rules,
		RequestID:         requestID(r),
	})
	if err != nil {
		h.writeCouponError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, couponResponseFromDomain(coupon))
}

func (h *Handler) handleDisableCoupon(w http.ResponseWriter, r *http.Request) {
	coupon, err := h.coupons.DisableCoupon(r.Context(), usecase.DisableCouponInput{
		Actor:     actorFromRequest(r),
		CouponID:  r.PathValue("coupon_id"),
		RequestID: requestID(r),
	})
	if err != nil {
		h.writeCouponError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, couponResponseFromDomain(coupon))
}

func (h *Handler) handleValidateCoupon(w http.ResponseWriter, r *http.Request) {
	var req couponValidationRequest
	if err := decodeJSON(w, r, h.maxBodyBytes, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	result, err := h.coupons.ValidateCoupon(r.Context(), usecase.ValidateCouponInput{
		CouponCode:     req.CouponCode,
		CampaignID:     req.CampaignID,
		UserID:         req.UserID,
		CartID:         req.CartID,
		OrderID:        req.OrderID,
		Currency:       req.Currency,
		SubtotalAmount: req.SubtotalAmount,
		Items:          couponItemsFromRequest(req.Items),
		RequestID:      requestID(r),
	})
	if err != nil {
		h.writeCouponError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, couponValidationResponseFromDomain(result))
}

func (h *Handler) handleRecordCouponRedemption(w http.ResponseWriter, r *http.Request) {
	var req couponRedemptionRequest
	if err := decodeJSON(w, r, h.maxBodyBytes, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	discountAmount := req.DiscountAmount
	currency := req.Currency
	if req.Discount != nil {
		discountAmount = req.Discount.Amount
		if strings.TrimSpace(currency) == "" {
			currency = req.Discount.Currency
		}
	}
	requestIDValue := firstNonEmpty(req.RequestID, requestID(r))
	redemption, err := h.coupons.RecordRedemption(r.Context(), usecase.RecordCouponRedemptionInput{
		CouponID:       req.CouponID,
		CampaignID:     req.CampaignID,
		OrderID:        req.OrderID,
		UserID:         req.UserID,
		DiscountAmount: discountAmount,
		Currency:       currency,
		RequestID:      requestIDValue,
	})
	if err != nil {
		h.writeCouponError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, couponRedemptionResponseFromDomain(redemption))
}

func (h *Handler) handleListCampaigns(w http.ResponseWriter, r *http.Request) {
	limit := intQuery(r, "limit", 50)
	offset := intQuery(r, "offset", 0)
	var startsAfter *time.Time
	if value := strings.TrimSpace(r.URL.Query().Get("starts_after")); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "starts_after must be RFC3339")
			return
		}
		startsAfter = &parsed
	}
	var endsBefore *time.Time
	if value := strings.TrimSpace(r.URL.Query().Get("ends_before")); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", "ends_before must be RFC3339")
			return
		}
		endsBefore = &parsed
	}
	campaigns, err := h.campaigns.ListCampaigns(r.Context(), usecase.ListCampaignsInput{
		Actor:       actorFromRequest(r),
		Status:      domain.CampaignStatus(r.URL.Query().Get("status")),
		StartsAfter: startsAfter,
		EndsBefore:  endsBefore,
		Query:       r.URL.Query().Get("q"),
		Limit:       limit,
		Offset:      offset,
		RequestID:   requestID(r),
	})
	if err != nil {
		h.writeCampaignError(w, r, err)
		return
	}
	response := campaignListResponse{
		Campaigns: make([]campaignResponse, 0, len(campaigns)),
		Limit:     limit,
		Offset:    offset,
	}
	for _, campaign := range campaigns {
		response.Campaigns = append(response.Campaigns, campaignResponseFromDomain(campaign))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) handleCreateCampaign(w http.ResponseWriter, r *http.Request) {
	var req campaignCreateRequest
	if err := decodeJSON(w, r, h.maxBodyBytes, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	campaign, err := h.campaigns.CreateCampaign(r.Context(), usecase.CreateCampaignInput{
		Actor:        actorFromRequest(r),
		Name:         req.Name,
		BudgetAmount: amountPtr(req.Budget),
		Currency:     currencyFromRequest(req.Currency, req.Budget),
		StartsAt:     req.StartsAt,
		EndsAt:       req.EndsAt,
		Metadata:     campaignMetadataFromRequest(req.Metadata),
		RequestID:    requestID(r),
	})
	if err != nil {
		h.writeCampaignError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, campaignResponseFromDomain(campaign))
}

func (h *Handler) handleUpdateCampaign(w http.ResponseWriter, r *http.Request) {
	var req campaignPatchRequest
	if err := decodeJSON(w, r, h.maxBodyBytes, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	var status *domain.CampaignStatus
	if req.Status != nil {
		value := domain.CampaignStatus(*req.Status)
		status = &value
	}
	var metadata *domain.CampaignMetadata
	if req.Metadata != nil {
		value := campaignMetadataFromRequest(*req.Metadata)
		metadata = &value
	}
	campaign, err := h.campaigns.UpdateCampaign(r.Context(), usecase.UpdateCampaignInput{
		Actor:        actorFromRequest(r),
		CampaignID:   r.PathValue("campaign_id"),
		Name:         req.Name,
		BudgetAmount: amountPtr(req.Budget),
		Currency:     req.Currency,
		StartsAt:     req.StartsAt,
		EndsAt:       req.EndsAt,
		Metadata:     metadata,
		Status:       status,
		RequestID:    requestID(r),
	})
	if err != nil {
		h.writeCampaignError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, campaignResponseFromDomain(campaign))
}

func (h *Handler) handleDisableCampaign(w http.ResponseWriter, r *http.Request) {
	campaign, err := h.campaigns.DisableCampaign(r.Context(), usecase.DisableCampaignInput{
		Actor:      actorFromRequest(r),
		CampaignID: r.PathValue("campaign_id"),
		RequestID:  requestID(r),
	})
	if err != nil {
		h.writeCampaignError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, campaignResponseFromDomain(campaign))
}

func (h *Handler) handleValidateCampaign(w http.ResponseWriter, r *http.Request) {
	var req campaignValidationRequest
	if err := decodeJSON(w, r, h.maxBodyBytes, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}
	result, err := h.campaigns.ValidateCampaign(r.Context(), usecase.ValidateCampaignInput{
		CampaignID:     req.CampaignID,
		SellerID:       req.SellerID,
		CouponID:       req.CouponID,
		UserID:         req.UserID,
		Currency:       req.Currency,
		DiscountAmount: req.DiscountAmount,
		RequestID:      requestID(r),
	})
	if err != nil {
		h.writeCampaignError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, campaignValidationResponseFromDomain(result))
}

func authorizeResponseFromDecision(decision usecase.AuthorizationDecision) authorizeResponse {
	return authorizeResponse{
		Allowed:            decision.Allowed,
		Reason:             decision.Reason,
		RequiredPermission: decision.RequiredPermission.String(),
		AuditRequired:      decision.AuditRequired,
		EffectiveRoles:     domain.RoleStrings(decision.EffectiveRoles),
		RequestID:          decision.RequestID,
	}
}

func moderationResponseFromResult(result domain.ProductModerationResult) moderationResponse {
	return moderationResponse{
		ProductID: result.ProductID,
		SellerID:  result.SellerID,
		ReviewID:  result.ReviewID,
		Status:    result.Status.Normalized().String(),
		Message:   result.Message,
	}
}

func reviewResponseFromDomain(review domain.ProductModerationReview) reviewResponse {
	out := reviewResponse{
		ReviewID:        review.ReviewID,
		ProductID:       review.ProductID,
		SellerID:        review.SellerID,
		Status:          string(review.Status.Normalized()),
		SubmittedBy:     review.SubmittedBy,
		ReviewedBy:      review.ReviewedBy,
		RejectionReason: review.RejectionReason,
		SubmittedAt:     formatTime(review.SubmittedAt),
		CreatedAt:       formatTime(review.CreatedAt),
		UpdatedAt:       formatTime(review.UpdatedAt),
	}
	if !review.ReviewedAt.IsZero() {
		out.ReviewedAt = formatTime(review.ReviewedAt)
	}
	return out
}

func couponResponseFromDomain(coupon domain.Coupon) couponResponse {
	coupon = coupon.Normalized()
	response := couponResponse{
		CouponID:      coupon.CouponID,
		Code:          coupon.Code,
		DiscountType:  string(coupon.DiscountType),
		DiscountValue: coupon.DiscountValue,
		MinCartAmount: moneyResponse{
			Amount:   coupon.MinCartAmount,
			Currency: coupon.Currency,
		},
		Currency:     coupon.Currency,
		UsageLimit:   coupon.UsageLimit,
		PerUserLimit: coupon.PerUserLimit,
		Status:       string(coupon.Status),
		CreatedAt:    formatTime(coupon.CreatedAt),
		UpdatedAt:    formatTime(coupon.UpdatedAt),
		Rules:        make([]couponRuleResponse, 0, len(coupon.Rules)),
	}
	if coupon.SellerID != nil {
		response.SellerID = *coupon.SellerID
	}
	if coupon.MaxDiscountAmount != nil {
		response.MaxDiscountAmount = &moneyResponse{Amount: *coupon.MaxDiscountAmount, Currency: coupon.Currency}
	}
	if coupon.StartsAt != nil {
		response.StartsAt = formatTime(*coupon.StartsAt)
	}
	if coupon.EndsAt != nil {
		response.EndsAt = formatTime(*coupon.EndsAt)
	}
	for _, rule := range coupon.Rules {
		response.Rules = append(response.Rules, couponRuleResponseFromDomain(rule))
	}
	return response
}

func couponRuleResponseFromDomain(rule domain.CouponRule) couponRuleResponse {
	rule = rule.Normalized()
	return couponRuleResponse{
		RuleID:   rule.RuleID,
		RuleType: string(rule.Type),
		RuleValue: couponRuleValueResponse{
			ProductIDs:  rule.ProductIDs,
			CategoryIDs: rule.CategoryIDs,
			SellerIDs:   rule.SellerIDs,
		},
	}
}

func couponValidationResponseFromDomain(result domain.CouponValidationResult) couponValidationResponse {
	return couponValidationResponse{
		Valid:      result.Valid,
		CouponID:   result.CouponID,
		CampaignID: result.CampaignID,
		Discount: moneyResponse{
			Amount:   result.Discount.Amount,
			Currency: result.Discount.Currency,
		},
		Reason:           string(result.Reason),
		EligibleSubtotal: result.EligibleSubtotal,
	}
}

func couponRedemptionResponseFromDomain(redemption domain.CouponRedemption) couponRedemptionResponse {
	return couponRedemptionResponse{
		RedemptionID: redemption.RedemptionID,
		CouponID:     redemption.CouponID,
		CampaignID:   redemption.CampaignID,
		OrderID:      redemption.OrderID,
		UserID:       redemption.UserID,
		Discount: moneyResponse{
			Amount:   redemption.Discount.Amount,
			Currency: redemption.Discount.Currency,
		},
		RedeemedAt: formatTime(redemption.RedeemedAt),
	}
}

func campaignResponseFromDomain(campaign domain.Campaign) campaignResponse {
	campaign = campaign.Normalized()
	response := campaignResponse{
		CampaignID: campaign.CampaignID,
		Name:       campaign.Name,
		Status:     string(campaign.Status),
		Currency:   campaign.Currency,
		StartsAt:   formatTime(campaign.StartsAt),
		EndsAt:     formatTime(campaign.EndsAt),
		Metadata:   campaignMetadataResponseFromDomain(campaign.Metadata),
		CreatedBy:  campaign.CreatedBy,
		CreatedAt:  formatTime(campaign.CreatedAt),
		UpdatedAt:  formatTime(campaign.UpdatedAt),
	}
	if campaign.SellerID != nil {
		response.SellerID = *campaign.SellerID
	}
	if campaign.BudgetAmount != nil {
		response.Budget = &moneyResponse{Amount: *campaign.BudgetAmount, Currency: campaign.Currency}
	}
	return response
}

func campaignMetadataResponseFromDomain(metadata domain.CampaignMetadata) campaignMetadataResponse {
	metadata = metadata.Normalized()
	return campaignMetadataResponse{
		CouponIDs:                   metadata.CouponIDs,
		Channels:                    metadata.Channels,
		Description:                 metadata.Description,
		UsageLimit:                  metadata.UsageLimit,
		PerUserLimit:                metadata.PerUserLimit,
		BudgetAlertThresholdPercent: metadata.BudgetAlertThresholdPercent,
	}
}

func campaignValidationResponseFromDomain(result domain.CampaignValidationResult) campaignValidationResponse {
	response := campaignValidationResponse{
		Valid:           result.Valid,
		CampaignID:      result.CampaignID,
		CouponID:        result.CouponID,
		BudgetUnlimited: result.BudgetUnlimited,
		Reason:          string(result.Reason),
	}
	if !result.BudgetUnlimited {
		response.RemainingBudget = &moneyResponse{Amount: result.RemainingBudget.Amount, Currency: result.RemainingBudget.Currency}
	}
	return response
}

func sellerAnalyticsResponseFromDomain(result domain.SellerAnalytics) sellerAnalyticsResponse {
	response := sellerAnalyticsResponse{
		Revenue:        moneyResponse{Amount: result.Revenue.Amount, Currency: result.Revenue.Currency},
		Orders:         result.Orders,
		ConversionRate: result.ConversionRate,
		TopProducts:    make([]topProductMetricResponse, 0, len(result.TopProducts)),
	}
	for _, product := range result.TopProducts {
		response.TopProducts = append(response.TopProducts, topProductMetricResponse{
			ProductID: product.ProductID,
			Title:     product.Title,
			UnitsSold: product.UnitsSold,
			Orders:    product.Orders,
			Revenue:   moneyResponse{Amount: product.Revenue.Amount, Currency: product.Revenue.Currency},
		})
	}
	return response
}

func auditLogListResponseFromDomain(page domain.AuditLogPage) auditLogListResponse {
	response := auditLogListResponse{
		AuditLogs: make([]auditLogResponse, 0, len(page.Entries)),
	}
	if page.NextCursor != "" {
		response.NextCursor = &page.NextCursor
	}
	for _, entry := range page.Entries {
		response.AuditLogs = append(response.AuditLogs, auditLogResponseFromDomain(entry))
	}
	return response
}

func auditLogResponseFromDomain(entry domain.AuditEntry) auditLogResponse {
	entry = entry.Normalized()
	return auditLogResponse{
		AuditID:       entry.AuditID,
		SellerID:      entry.SellerID,
		ActorUserID:   entry.ActorUserID,
		ActorSellerID: entry.ActorSellerID,
		ActorRoles:    domain.RoleStrings(entry.ActorRoles),
		Action:        entry.Action.String(),
		ResourceType:  entry.ResourceType,
		ResourceID:    entry.ResourceID,
		RequestID:     entry.RequestID,
		TraceID:       entry.TraceID,
		Decision:      string(entry.Decision),
		Reason:        entry.Reason,
		Before:        entry.Before,
		After:         entry.After,
		CreatedAt:     formatTime(entry.CreatedAt),
	}
}

func couponRulesFromRequest(requests []couponRuleRequest) []domain.CouponRule {
	rules := make([]domain.CouponRule, 0, len(requests))
	for _, req := range requests {
		rules = append(rules, domain.CouponRule{
			Type:        domain.CouponRuleType(req.RuleType),
			ProductIDs:  req.RuleValue.ProductIDs,
			CategoryIDs: req.RuleValue.CategoryIDs,
			SellerIDs:   req.RuleValue.SellerIDs,
		})
	}
	return rules
}

func couponItemsFromRequest(requests []couponItemRequest) []domain.CouponCartItem {
	items := make([]domain.CouponCartItem, 0, len(requests))
	for _, req := range requests {
		items = append(items, domain.CouponCartItem{
			ProductID:          req.ProductID,
			VariantID:          req.VariantID,
			SellerID:           req.SellerID,
			CategoryIDs:        req.CategoryIDs,
			Quantity:           req.Quantity,
			UnitAmount:         req.UnitAmount,
			LineSubtotalAmount: req.LineSubtotalAmount,
		})
	}
	return items
}

func campaignMetadataFromRequest(req campaignMetadataRequest) domain.CampaignMetadata {
	return domain.CampaignMetadata{
		CouponIDs:                   req.CouponIDs,
		Channels:                    req.Channels,
		Description:                 req.Description,
		UsageLimit:                  req.UsageLimit,
		PerUserLimit:                req.PerUserLimit,
		BudgetAlertThresholdPercent: req.BudgetAlertThresholdPercent,
	}
}

func amountPtr(value *amountValue) *int64 {
	if value == nil {
		return nil
	}
	amount := value.Amount
	return &amount
}

func amountOrZero(value *amountValue) int64 {
	if value == nil {
		return 0
	}
	return value.Amount
}

func currencyFromRequest(explicit string, amounts ...*amountValue) string {
	if strings.TrimSpace(explicit) != "" {
		return explicit
	}
	for _, amount := range amounts {
		if amount != nil && strings.TrimSpace(amount.Currency) != "" {
			return amount.Currency
		}
	}
	return ""
}

func responseWithError(response authorizeResponse, code string, message string) map[string]any {
	return map[string]any{
		"allowed":             response.Allowed,
		"reason":              response.Reason,
		"required_permission": response.RequiredPermission,
		"audit_required":      response.AuditRequired,
		"effective_roles":     response.EffectiveRoles,
		"request_id":          response.RequestID,
		"error": apiError{
			Code:    code,
			Message: message,
		},
	}
}

func permissionCatalogDTO() permissionCatalogResponse {
	roles := []roleResponse{
		{Name: domain.RoleSeller.String(), Description: "Primary seller owner with full own-seller CMS access"},
		{Name: domain.RoleSellerManager.String(), Description: "Daily operations manager with broad catalog, order, analytics, and limited staff access"},
		{Name: domain.RoleSellerCatalogEditor.String(), Description: "Catalog, coupon, and campaign focused seller staff role"},
		{Name: domain.RoleSellerOrderManager.String(), Description: "Order handling, fulfillment, and return response seller staff role"},
		{Name: domain.RoleCatalogAdmin.String(), Description: "Catalog moderator who can approve or reject product reviews"},
		{Name: domain.RoleSuperadmin.String(), Description: "Platform administrator with catalog moderation override permissions"},
	}

	permissions := make([]permissionResponse, 0, len(domain.PermissionCatalog()))
	for _, permission := range domain.PermissionCatalog() {
		permissions = append(permissions, permissionResponse{
			Name:        permission.Permission.String(),
			Resource:    permission.Resource,
			Action:      permission.Action,
			Description: permission.Description,
			Risk:        string(permission.Risk),
			Mutation:    permission.Mutation,
		})
	}

	matrix := make(map[string][]string, len(domain.AllRoles()))
	for role, permissions := range domain.PermissionsByRole() {
		values := make([]string, 0, len(permissions))
		for _, permission := range permissions {
			values = append(values, permission.String())
		}
		matrix[role.String()] = values
	}

	routes := make([]routePermissionResponse, 0, len(domain.RoutePermissions()))
	for _, route := range domain.RoutePermissions() {
		routes = append(routes, routePermissionResponse{
			Method:     route.Method,
			Path:       route.Path,
			Permission: route.Permission.String(),
		})
	}

	return permissionCatalogResponse{
		Roles:       roles,
		Permissions: permissions,
		Matrix:      matrix,
		Routes:      routes,
	}
}

func mapAuthorizationError(err error) (int, string, string) {
	switch {
	case errors.Is(err, domain.ErrUnauthenticated):
		return http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "Authentication required"
	case errors.Is(err, domain.ErrUnknownPermission),
		errors.Is(err, domain.ErrUnknownRole),
		errors.Is(err, domain.ErrInvalidSellerScope),
		errors.Is(err, domain.ErrInvalidAuthorizationIn):
		return http.StatusBadRequest, "VALIDATION_ERROR", err.Error()
	case errors.Is(err, domain.ErrAuditWriteFailed):
		return http.StatusInternalServerError, "AUDIT_WRITE_FAILED", "Audit write failed"
	case errors.Is(err, domain.ErrStaffLookupFailed):
		return http.StatusServiceUnavailable, "STAFF_LOOKUP_UNAVAILABLE", "Staff status lookup is temporarily unavailable"
	default:
		return http.StatusForbidden, "PERMISSION_DENIED", "Permission denied"
	}
}

func (h *Handler) writeModerationError(w http.ResponseWriter, r *http.Request, err error) {
	status, apiErr := mapModerationError(err)
	h.logger.InfoContext(r.Context(), "cms.product_moderation.request_failed",
		slog.Int("status", status),
		slog.String("code", apiErr.Code),
		slog.String("request_id", requestID(r)),
		slog.String("error", err.Error()),
	)
	writeJSON(w, status, errorResponse{Error: apiErr})
}

func (h *Handler) writeCouponError(w http.ResponseWriter, r *http.Request, err error) {
	status, apiErr := mapCouponError(err)
	h.logger.InfoContext(r.Context(), "cms.coupon.request_failed",
		slog.Int("status", status),
		slog.String("code", apiErr.Code),
		slog.String("request_id", requestID(r)),
		slog.String("error", err.Error()),
	)
	writeJSON(w, status, errorResponse{Error: apiErr})
}

func (h *Handler) writeCampaignError(w http.ResponseWriter, r *http.Request, err error) {
	status, apiErr := mapCampaignError(err)
	h.logger.InfoContext(r.Context(), "cms.campaign.request_failed",
		slog.Int("status", status),
		slog.String("code", apiErr.Code),
		slog.String("request_id", requestID(r)),
		slog.String("error", err.Error()),
	)
	writeJSON(w, status, errorResponse{Error: apiErr})
}

func (h *Handler) writeAnalyticsError(w http.ResponseWriter, r *http.Request, err error) {
	status, apiErr := mapAnalyticsError(err)
	h.logger.InfoContext(r.Context(), "cms.seller_analytics.request_failed",
		slog.Int("status", status),
		slog.String("code", apiErr.Code),
		slog.String("request_id", requestID(r)),
		slog.String("error", err.Error()),
	)
	writeJSON(w, status, errorResponse{Error: apiErr})
}

func (h *Handler) writeAuditLogError(w http.ResponseWriter, r *http.Request, err error) {
	status, apiErr := mapAuditLogError(err)
	h.logger.InfoContext(r.Context(), "cms.audit_logs.request_failed",
		slog.Int("status", status),
		slog.String("code", apiErr.Code),
		slog.String("request_id", requestID(r)),
		slog.String("error", err.Error()),
	)
	writeJSON(w, status, errorResponse{Error: apiErr})
}

func mapModerationError(err error) (int, apiError) {
	var validationErr domain.ValidationError
	if errors.As(err, &validationErr) {
		fields := make([]fieldErrorResponse, 0, len(validationErr.Fields))
		for _, field := range validationErr.Fields {
			fields = append(fields, fieldErrorResponse{Field: field.Field, Reason: field.Reason})
		}
		return http.StatusUnprocessableEntity, apiError{Code: "VALIDATION_FAILED", Message: validationErr.Error(), Fields: fields}
	}
	switch {
	case errors.Is(err, domain.ErrUnauthenticated):
		return http.StatusUnauthorized, apiError{Code: "AUTHENTICATION_REQUIRED", Message: "Authentication required"}
	case errors.Is(err, domain.ErrForbidden), errors.Is(err, domain.ErrProductOwnershipMismatch):
		return http.StatusForbidden, apiError{Code: "PERMISSION_DENIED", Message: "Permission denied"}
	case errors.Is(err, domain.ErrProductNotFound):
		return http.StatusNotFound, apiError{Code: "NOT_FOUND", Message: "Product not found"}
	case errors.Is(err, domain.ErrReviewNotFound):
		return http.StatusNotFound, apiError{Code: "NOT_FOUND", Message: "Product moderation review not found"}
	case errors.Is(err, domain.ErrInvalidProductTransition), errors.Is(err, domain.ErrReviewAlreadyDecided):
		return http.StatusPreconditionFailed, apiError{Code: "FAILED_PRECONDITION", Message: err.Error()}
	case errors.Is(err, domain.ErrInvalidModerationDecision),
		errors.Is(err, domain.ErrRejectionReasonRequired),
		errors.Is(err, domain.ErrForceUnpublishReasonNeeded),
		errors.Is(err, domain.ErrInvalidProductStatus),
		errors.Is(err, domain.ErrValidationFailed):
		return http.StatusUnprocessableEntity, apiError{Code: "VALIDATION_FAILED", Message: err.Error()}
	case errors.Is(err, domain.ErrProductServiceUnavailable):
		return http.StatusServiceUnavailable, apiError{Code: "UNAVAILABLE", Message: "Product service is temporarily unavailable"}
	case errors.Is(err, domain.ErrAuditWriteFailed):
		return http.StatusInternalServerError, apiError{Code: "AUDIT_WRITE_FAILED", Message: "Audit write failed"}
	default:
		return http.StatusInternalServerError, apiError{Code: "INTERNAL_ERROR", Message: "Internal server error"}
	}
}

func mapAnalyticsError(err error) (int, apiError) {
	var validationErr domain.ValidationError
	if errors.As(err, &validationErr) {
		fields := make([]fieldErrorResponse, 0, len(validationErr.Fields))
		for _, field := range validationErr.Fields {
			fields = append(fields, fieldErrorResponse{Field: field.Field, Reason: field.Reason})
		}
		return http.StatusBadRequest, apiError{Code: "invalid_date_range", Message: validationErr.Error(), Fields: fields}
	}
	switch {
	case errors.Is(err, domain.ErrUnauthenticated):
		return http.StatusUnauthorized, apiError{Code: "authentication_required", Message: "Authentication required"}
	case errors.Is(err, domain.ErrSellerContextRequired):
		return http.StatusForbidden, apiError{Code: "seller_context_required", Message: "Seller context required"}
	case errors.Is(err, domain.ErrForbidden), errors.Is(err, domain.ErrInactiveStaff), errors.Is(err, domain.ErrUnknownRole):
		return http.StatusForbidden, apiError{Code: "permission_denied", Message: "Analytics read permission required"}
	case errors.Is(err, domain.ErrAnalyticsRangeTooLarge):
		return http.StatusBadRequest, apiError{Code: "date_range_too_large", Message: "Analytics date range is too large"}
	case errors.Is(err, domain.ErrInvalidAnalyticsDateRange):
		return http.StatusBadRequest, apiError{Code: "invalid_date_range", Message: err.Error()}
	case errors.Is(err, domain.ErrStaffLookupFailed), errors.Is(err, domain.ErrAnalyticsUnavailable):
		return http.StatusServiceUnavailable, apiError{Code: "analytics_temporarily_unavailable", Message: "Seller analytics are temporarily unavailable"}
	default:
		return http.StatusInternalServerError, apiError{Code: "internal_error", Message: "Internal server error"}
	}
}

func mapAuditLogError(err error) (int, apiError) {
	switch {
	case errors.Is(err, domain.ErrUnauthenticated):
		return http.StatusUnauthorized, apiError{Code: "authentication_required", Message: "Authentication required"}
	case errors.Is(err, domain.ErrSellerContextRequired):
		return http.StatusForbidden, apiError{Code: "seller_context_required", Message: "Seller context required"}
	case errors.Is(err, domain.ErrForbidden), errors.Is(err, domain.ErrInactiveStaff), errors.Is(err, domain.ErrUnknownRole):
		return http.StatusForbidden, apiError{Code: "permission_denied", Message: "Audit log read permission required"}
	case errors.Is(err, domain.ErrInvalidAuditLogFilter):
		return http.StatusBadRequest, apiError{Code: "invalid_audit_log_filter", Message: err.Error()}
	case errors.Is(err, domain.ErrStaffLookupFailed), errors.Is(err, domain.ErrAuditLogsUnavailable):
		return http.StatusServiceUnavailable, apiError{Code: "audit_logs_temporarily_unavailable", Message: "Audit logs are temporarily unavailable"}
	default:
		return http.StatusInternalServerError, apiError{Code: "internal_error", Message: "Internal server error"}
	}
}

func mapCouponError(err error) (int, apiError) {
	var validationErr domain.ValidationError
	if errors.As(err, &validationErr) {
		fields := make([]fieldErrorResponse, 0, len(validationErr.Fields))
		for _, field := range validationErr.Fields {
			fields = append(fields, fieldErrorResponse{Field: field.Field, Reason: field.Reason})
		}
		return http.StatusUnprocessableEntity, apiError{Code: "VALIDATION_FAILED", Message: validationErr.Error(), Fields: fields}
	}
	switch {
	case errors.Is(err, domain.ErrUnauthenticated):
		return http.StatusUnauthorized, apiError{Code: "AUTHENTICATION_REQUIRED", Message: "Authentication required"}
	case errors.Is(err, domain.ErrForbidden), errors.Is(err, domain.ErrCouponOwnershipMismatch):
		return http.StatusForbidden, apiError{Code: "PERMISSION_DENIED", Message: "Permission denied"}
	case errors.Is(err, domain.ErrCouponNotFound):
		return http.StatusNotFound, apiError{Code: "NOT_FOUND", Message: "Coupon not found"}
	case errors.Is(err, domain.ErrCouponCodeAlreadyExists):
		return http.StatusConflict, apiError{Code: "COUPON_CODE_EXISTS", Message: "Coupon code already exists"}
	case errors.Is(err, domain.ErrCouponRedemptionConflict):
		return http.StatusConflict, apiError{Code: "REDEMPTION_CONFLICT", Message: "Coupon redemption conflicts with an existing order redemption"}
	case errors.Is(err, domain.ErrCouponUsageLimitReached):
		return http.StatusPreconditionFailed, apiError{Code: "USAGE_LIMIT_REACHED", Message: "Coupon usage limit reached"}
	case errors.Is(err, domain.ErrCouponPerUserLimitReached):
		return http.StatusPreconditionFailed, apiError{Code: "PER_USER_LIMIT_REACHED", Message: "Coupon per-user limit reached"}
	case errors.Is(err, domain.ErrCouponCurrencyMismatch):
		return http.StatusUnprocessableEntity, apiError{Code: "CURRENCY_MISMATCH", Message: "Coupon currency mismatch"}
	case errors.Is(err, domain.ErrCampaignNotFound):
		return http.StatusNotFound, apiError{Code: "CAMPAIGN_NOT_FOUND", Message: "Campaign not found"}
	case errors.Is(err, domain.ErrCampaignBudgetExhausted):
		return http.StatusPreconditionFailed, apiError{Code: "CAMPAIGN_BUDGET_EXHAUSTED", Message: "Campaign budget exhausted"}
	case errors.Is(err, domain.ErrCampaignUsageLimitReached), errors.Is(err, domain.ErrCampaignPerUserLimitReached):
		return http.StatusPreconditionFailed, apiError{Code: "CAMPAIGN_USAGE_LIMIT_REACHED", Message: "Campaign usage limit reached"}
	case errors.Is(err, domain.ErrCampaignCouponNotLinked):
		return http.StatusUnprocessableEntity, apiError{Code: "COUPON_NOT_IN_CAMPAIGN", Message: "Coupon is not linked to campaign"}
	case errors.Is(err, domain.ErrCampaignCurrencyMismatch):
		return http.StatusUnprocessableEntity, apiError{Code: "CURRENCY_MISMATCH", Message: "Campaign currency mismatch"}
	case errors.Is(err, domain.ErrInvalidCouponStatus), errors.Is(err, domain.ErrInvalidCouponDiscountType), errors.Is(err, domain.ErrInvalidCouponRule), errors.Is(err, domain.ErrValidationFailed):
		return http.StatusUnprocessableEntity, apiError{Code: "VALIDATION_FAILED", Message: err.Error()}
	case errors.Is(err, domain.ErrAuditWriteFailed):
		return http.StatusInternalServerError, apiError{Code: "AUDIT_WRITE_FAILED", Message: "Audit write failed"}
	default:
		return http.StatusInternalServerError, apiError{Code: "INTERNAL_ERROR", Message: "Internal server error"}
	}
}

func mapCampaignError(err error) (int, apiError) {
	var validationErr domain.ValidationError
	if errors.As(err, &validationErr) {
		fields := make([]fieldErrorResponse, 0, len(validationErr.Fields))
		for _, field := range validationErr.Fields {
			fields = append(fields, fieldErrorResponse{Field: field.Field, Reason: field.Reason})
		}
		return http.StatusUnprocessableEntity, apiError{Code: "VALIDATION_FAILED", Message: validationErr.Error(), Fields: fields}
	}
	switch {
	case errors.Is(err, domain.ErrUnauthenticated):
		return http.StatusUnauthorized, apiError{Code: "AUTHENTICATION_REQUIRED", Message: "Authentication required"}
	case errors.Is(err, domain.ErrForbidden), errors.Is(err, domain.ErrCampaignOwnershipMismatch):
		return http.StatusForbidden, apiError{Code: "PERMISSION_DENIED", Message: "Permission denied"}
	case errors.Is(err, domain.ErrCampaignNotFound):
		return http.StatusNotFound, apiError{Code: "NOT_FOUND", Message: "Campaign not found"}
	case errors.Is(err, domain.ErrCampaignCompletedReadOnly):
		return http.StatusPreconditionFailed, apiError{Code: "CAMPAIGN_COMPLETED", Message: "Completed campaign is read-only"}
	case errors.Is(err, domain.ErrCampaignBudgetExhausted):
		return http.StatusPreconditionFailed, apiError{Code: "CAMPAIGN_BUDGET_EXHAUSTED", Message: "Campaign budget exhausted"}
	case errors.Is(err, domain.ErrCampaignUsageLimitReached), errors.Is(err, domain.ErrCampaignPerUserLimitReached):
		return http.StatusPreconditionFailed, apiError{Code: "CAMPAIGN_USAGE_LIMIT_REACHED", Message: "Campaign usage limit reached"}
	case errors.Is(err, domain.ErrCampaignCouponNotLinked):
		return http.StatusUnprocessableEntity, apiError{Code: "COUPON_NOT_IN_CAMPAIGN", Message: "Coupon is not linked to campaign"}
	case errors.Is(err, domain.ErrCampaignCurrencyMismatch):
		return http.StatusUnprocessableEntity, apiError{Code: "CURRENCY_MISMATCH", Message: "Campaign currency mismatch"}
	case errors.Is(err, domain.ErrInvalidCampaignStatus), errors.Is(err, domain.ErrInvalidCampaignStatusTransition), errors.Is(err, domain.ErrValidationFailed):
		return http.StatusUnprocessableEntity, apiError{Code: "VALIDATION_FAILED", Message: err.Error()}
	case errors.Is(err, domain.ErrAuditWriteFailed):
		return http.StatusInternalServerError, apiError{Code: "AUDIT_WRITE_FAILED", Message: "Audit write failed"}
	default:
		return http.StatusInternalServerError, apiError{Code: "INTERNAL_ERROR", Message: "Internal server error"}
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, maxBytes int64, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}
	w.Header().Set("Allow", method)
	writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	return false
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeAPIError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, errorResponse{
		Error: apiError{
			Code:    code,
			Message: message,
		},
	})
}

func intQuery(r *http.Request, key string, fallback int) int {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func optionalAnalyticsTimeQuery(r *http.Request, key string) (*time.Time, error) {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return nil, nil
	}
	parsed, err := parseAnalyticsTime(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func optionalAnalyticsLimitQuery(r *http.Request, key string) (int, error) {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New("top_products_limit must be an integer")
	}
	if parsed < 1 {
		return 0, errors.New("top_products_limit must be at least 1")
	}
	return parsed, nil
}

func optionalAuditTimeQuery(r *http.Request, key string) (*time.Time, error) {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return nil, nil
	}
	parsed, err := parseAuditTime(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func optionalAuditPageSizeQuery(r *http.Request, key string) (int, error) {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New("page_size must be an integer")
	}
	if parsed < 1 {
		return 0, errors.New("page_size must be at least 1")
	}
	return parsed, nil
}

func parseAnalyticsTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if parsed, err := time.Parse("2006-01-02", value); err == nil {
		return parsed.UTC(), nil
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.UTC(), nil
	}
	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return parsed.UTC(), nil
	}
	return time.Time{}, errors.New("analytics date must be YYYY-MM-DD or RFC3339")
}

func parseAuditTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if parsed, err := time.Parse("2006-01-02", value); err == nil {
		return parsed.UTC(), nil
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.UTC(), nil
	}
	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return parsed.UTC(), nil
	}
	return time.Time{}, errors.New("audit date must be YYYY-MM-DD or RFC3339")
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
