package grpctransport

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
	cmsv1 "github.com/example/ecommerce-platform/backend/services/cms-service/internal/gen/ecommerce/cms/v1"
	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func validateCouponInputFromProto(req *cmsv1.ValidateCouponRequest, requestID string) (usecase.ValidateCouponInput, error) {
	if req == nil {
		return usecase.ValidateCouponInput{}, status.Error(codes.InvalidArgument, "request is required")
	}
	currency := req.GetCurrency()
	subtotal := req.GetCartSubtotal()
	if currency == "" && subtotal != nil {
		currency = subtotal.GetCurrency()
	}
	items := make([]domain.CouponCartItem, 0, len(req.GetItems()))
	for _, item := range req.GetItems() {
		if item == nil {
			continue
		}
		categoryIDs := []string{}
		if item.GetCategoryId() != "" {
			categoryIDs = append(categoryIDs, item.GetCategoryId())
		}
		lineTotal := item.GetLineTotal()
		unitPrice := item.GetUnitPrice()
		items = append(items, domain.CouponCartItem{
			ProductID:          item.GetProductId(),
			VariantID:          item.GetVariantId(),
			SellerID:           item.GetSellerId(),
			CategoryIDs:        categoryIDs,
			Quantity:           int64(item.GetQuantity()),
			UnitAmount:         moneyAmount(unitPrice),
			LineSubtotalAmount: moneyAmount(lineTotal),
		})
	}
	return usecase.ValidateCouponInput{
		CouponCode:     req.GetCouponCode(),
		CampaignID:     req.GetCampaignId(),
		UserID:         req.GetUserId(),
		CartID:         req.GetCartId(),
		OrderID:        req.GetOrderId(),
		Currency:       currency,
		SubtotalAmount: moneyAmount(subtotal),
		Items:          items,
		RequestID:      requestID,
	}, nil
}

func validateCouponResponseToProto(result domain.CouponValidationResult) *cmsv1.ValidateCouponResponse {
	return &cmsv1.ValidateCouponResponse{
		Valid:            result.Valid,
		CouponId:         result.CouponID,
		Discount:         moneyToProto(result.Discount),
		Reason:           couponReasonForProto(result.Reason),
		CampaignId:       result.CampaignID,
		AppliedRuleCodes: append([]string(nil), result.AppliedRuleCodes...),
	}
}

func getSellerSettingsInputFromProto(req *cmsv1.GetSellerSettingsRequest, actor domain.ActorContext, requestID string) usecase.GetSellerSettingsInput {
	sellerID := ""
	if req != nil {
		sellerID = req.GetSellerId()
	}
	return usecase.GetSellerSettingsInput{
		Actor:     actor,
		SellerID:  sellerID,
		RequestID: requestID,
	}
}

func sellerSettingsToProto(settings domain.SellerSettings) (*cmsv1.SellerSettings, error) {
	settings = settings.Normalized()
	metadata, err := structpb.NewStruct(settings.Settings)
	if err != nil {
		return nil, err
	}
	return &cmsv1.SellerSettings{
		SellerId:       settings.SellerID,
		ReturnPolicy:   settings.ReturnPolicy,
		ShippingPolicy: settings.ShippingPolicy,
		UpdatedAt:      timeToProto(settings.UpdatedAt),
		SupportEmail:   settings.SupportEmail,
		Settings:       metadata,
	}, nil
}

func getCampaignInputFromProto(req *cmsv1.GetCampaignRequest, actor domain.ActorContext, requestID string) usecase.GetCampaignInput {
	if req == nil {
		return usecase.GetCampaignInput{Actor: actor, RequestID: requestID}
	}
	return usecase.GetCampaignInput{
		Actor:      actor,
		SellerID:   req.GetSellerId(),
		CampaignID: req.GetCampaignId(),
		RequestID:  requestID,
	}
}

func listCampaignsInputFromProto(req *cmsv1.ListCampaignsRequest, actor domain.ActorContext, requestID string) usecase.ListCampaignsInput {
	limit, offset := pageLimitOffset(nil)
	status := cmsv1.CampaignStatus_CAMPAIGN_STATUS_UNSPECIFIED
	if req != nil {
		limit, offset = pageLimitOffset(req.GetPage())
		status = req.GetStatus()
	}
	return usecase.ListCampaignsInput{
		Actor:     actor,
		Status:    campaignStatusFromProto(status),
		Limit:     limit,
		Offset:    offset,
		RequestID: requestID,
	}
}

func listAuditLogsInputFromProto(req *cmsv1.ListAuditLogsRequest, actor domain.ActorContext, requestID string) usecase.ListAuditLogsInput {
	if req == nil {
		return usecase.ListAuditLogsInput{Actor: actor, RequestID: requestID}
	}
	return usecase.ListAuditLogsInput{
		Actor:        actor,
		ActorUserID:  req.GetActorUserId(),
		Action:       domain.Permission(req.GetAction()),
		ResourceType: req.GetResourceType(),
		ResourceID:   req.GetResourceId(),
		From:         timestampPtr(req.GetFrom()),
		To:           timestampPtr(req.GetTo()),
		PageSize:     int(req.GetPageSize()),
		Cursor:       req.GetCursor(),
		RequestID:    requestID,
	}
}

func auditLogPageToProto(page domain.AuditLogPage) (*cmsv1.ListAuditLogsResponse, error) {
	logs := make([]*cmsv1.AuditLog, 0, len(page.Entries))
	for _, entry := range page.Entries {
		log, err := auditLogToProto(entry)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return &cmsv1.ListAuditLogsResponse{
		AuditLogs:  logs,
		NextCursor: page.NextCursor,
	}, nil
}

func auditLogToProto(entry domain.AuditEntry) (*cmsv1.AuditLog, error) {
	entry = entry.Normalized()
	before, err := mapToStruct(entry.Before)
	if err != nil {
		return nil, err
	}
	after, err := mapToStruct(entry.After)
	if err != nil {
		return nil, err
	}
	return &cmsv1.AuditLog{
		AuditId:      entry.AuditID,
		SellerId:     entry.SellerID,
		ActorUserId:  entry.ActorUserID,
		ActorRoles:   domain.RoleStrings(entry.ActorRoles),
		Action:       entry.Action.String(),
		ResourceType: entry.ResourceType,
		ResourceId:   entry.ResourceID,
		RequestId:    entry.RequestID,
		TraceId:      entry.TraceID,
		Decision:     string(entry.Decision),
		Reason:       entry.Reason,
		Before:       before,
		After:        after,
		CreatedAt:    timeToProto(entry.CreatedAt),
	}, nil
}

func campaignReadToProto(read usecase.CampaignRead) (*cmsv1.Campaign, error) {
	campaign := read.Campaign.Normalized()
	metadata, err := campaignMetadataToStruct(campaign.Metadata)
	if err != nil {
		return nil, err
	}
	return &cmsv1.Campaign{
		CampaignId: campaign.CampaignID,
		SellerId:   campaignSellerID(campaign),
		Name:       campaign.Name,
		Status:     campaignStatusToProto(campaign.Status),
		StartsAt:   timeToProto(campaign.StartsAt),
		EndsAt:     timeToProto(campaign.EndsAt),
		Budget:     optionalMoneyToProto(campaign.BudgetAmount, campaign.Currency),
		Spent:      moneyToProto(read.Spent),
		CouponIds:  append([]string(nil), campaign.Metadata.CouponIDs...),
		Metadata:   metadata,
		CreatedAt:  timeToProto(campaign.CreatedAt),
		UpdatedAt:  timeToProto(campaign.UpdatedAt),
	}, nil
}

func campaignSellerID(campaign domain.Campaign) string {
	if campaign.SellerID == nil {
		return ""
	}
	return strings.TrimSpace(*campaign.SellerID)
}

func campaignMetadataToStruct(metadata domain.CampaignMetadata) (*structpb.Struct, error) {
	metadata = metadata.Normalized()
	raw, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}
	values := map[string]any{}
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, err
	}
	return structpb.NewStruct(values)
}

func campaignStatusFromProto(status cmsv1.CampaignStatus) domain.CampaignStatus {
	switch status {
	case cmsv1.CampaignStatus_CAMPAIGN_STATUS_DRAFT:
		return domain.CampaignStatusDraft
	case cmsv1.CampaignStatus_CAMPAIGN_STATUS_ACTIVE:
		return domain.CampaignStatusActive
	case cmsv1.CampaignStatus_CAMPAIGN_STATUS_PAUSED:
		return domain.CampaignStatusPaused
	case cmsv1.CampaignStatus_CAMPAIGN_STATUS_COMPLETED:
		return domain.CampaignStatusCompleted
	default:
		return ""
	}
}

func campaignStatusToProto(status domain.CampaignStatus) cmsv1.CampaignStatus {
	switch status.Normalized() {
	case domain.CampaignStatusDraft:
		return cmsv1.CampaignStatus_CAMPAIGN_STATUS_DRAFT
	case domain.CampaignStatusActive:
		return cmsv1.CampaignStatus_CAMPAIGN_STATUS_ACTIVE
	case domain.CampaignStatusPaused:
		return cmsv1.CampaignStatus_CAMPAIGN_STATUS_PAUSED
	case domain.CampaignStatusCompleted:
		return cmsv1.CampaignStatus_CAMPAIGN_STATUS_COMPLETED
	default:
		return cmsv1.CampaignStatus_CAMPAIGN_STATUS_UNSPECIFIED
	}
}

func pageLimitOffset(page *cmsv1.PageRequest) (int, int) {
	if page == nil {
		return 50, 0
	}
	limit := int(page.GetLimit())
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	offset := int(page.GetOffset())
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func moneyAmount(value *cmsv1.Money) int64 {
	if value == nil {
		return 0
	}
	return value.GetAmount()
}

func moneyToProto(value domain.Money) *cmsv1.Money {
	currency := domain.NormalizeCurrency(value.Currency)
	return &cmsv1.Money{Amount: value.Amount, Currency: currency}
}

func optionalMoneyToProto(amount *int64, currency string) *cmsv1.Money {
	if amount == nil {
		return nil
	}
	return &cmsv1.Money{Amount: *amount, Currency: domain.NormalizeCurrency(currency)}
}

func timeToProto(value time.Time) *timestamppb.Timestamp {
	if value.IsZero() {
		return nil
	}
	return timestamppb.New(value.UTC())
}

func timestampPtr(value *timestamppb.Timestamp) *time.Time {
	if value == nil {
		return nil
	}
	parsed := value.AsTime().UTC()
	return &parsed
}

func mapToStruct(value map[string]any) (*structpb.Struct, error) {
	if len(value) == 0 {
		return nil, nil
	}
	return structpb.NewStruct(value)
}

func couponReasonForProto(reason domain.CouponInvalidReason) string {
	switch reason {
	case "":
		return ""
	case domain.CouponInvalidReasonNotActive:
		return "coupon_inactive"
	case domain.CouponInvalidReasonUsageLimitReached:
		return "global_usage_limit_reached"
	case domain.CouponInvalidReasonPerUserLimitReached:
		return "user_usage_limit_reached"
	case domain.CouponInvalidReasonScopeNotMatched:
		return "scope_not_matched"
	default:
		return string(reason)
	}
}
