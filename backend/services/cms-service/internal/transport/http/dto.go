package httptransport

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type authorizeRequest struct {
	ResourceSellerID   string `json:"resource_seller_id"`
	RequiredPermission string `json:"required_permission"`
	ResourceType       string `json:"resource_type,omitempty"`
	ResourceID         string `json:"resource_id,omitempty"`
	TargetRole         string `json:"target_role,omitempty"`
}

type authorizeResponse struct {
	Allowed            bool     `json:"allowed"`
	Reason             string   `json:"reason"`
	RequiredPermission string   `json:"required_permission"`
	AuditRequired      bool     `json:"audit_required"`
	EffectiveRoles     []string `json:"effective_roles,omitempty"`
	RequestID          string   `json:"request_id,omitempty"`
}

type permissionCatalogResponse struct {
	Roles       []roleResponse            `json:"roles"`
	Permissions []permissionResponse      `json:"permissions"`
	Matrix      map[string][]string       `json:"matrix"`
	Routes      []routePermissionResponse `json:"routes"`
}

type roleResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type permissionResponse struct {
	Name        string `json:"name"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description"`
	Risk        string `json:"risk"`
	Mutation    bool   `json:"mutation"`
}

type routePermissionResponse struct {
	Method     string `json:"method"`
	Path       string `json:"path"`
	Permission string `json:"permission"`
}

type successResponse struct {
	Success bool `json:"success"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string               `json:"code"`
	Message string               `json:"message"`
	Fields  []fieldErrorResponse `json:"fields,omitempty"`
}

type moderationResponse struct {
	ProductID string `json:"product_id"`
	SellerID  string `json:"seller_id,omitempty"`
	ReviewID  string `json:"review_id,omitempty"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

type reviewDecisionRequest struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason,omitempty"`
}

type unpublishRequest struct {
	Reason string `json:"reason,omitempty"`
}

type reviewListResponse struct {
	Reviews []reviewResponse `json:"reviews"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
}

type reviewResponse struct {
	ReviewID        string `json:"review_id"`
	ProductID       string `json:"product_id"`
	SellerID        string `json:"seller_id"`
	Status          string `json:"status"`
	SubmittedBy     string `json:"submitted_by"`
	ReviewedBy      string `json:"reviewed_by,omitempty"`
	RejectionReason string `json:"rejection_reason,omitempty"`
	SubmittedAt     string `json:"submitted_at"`
	ReviewedAt      string `json:"reviewed_at,omitempty"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type fieldErrorResponse struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

type couponCreateRequest struct {
	Code              string              `json:"code"`
	DiscountType      string              `json:"discount_type"`
	DiscountValue     int64               `json:"discount_value"`
	MaxDiscountAmount *amountValue        `json:"max_discount_amount,omitempty"`
	MinCartAmount     *amountValue        `json:"min_cart_amount,omitempty"`
	Currency          string              `json:"currency,omitempty"`
	UsageLimit        *int64              `json:"usage_limit,omitempty"`
	PerUserLimit      *int64              `json:"per_user_limit,omitempty"`
	Status            string              `json:"status,omitempty"`
	StartsAt          *time.Time          `json:"starts_at,omitempty"`
	EndsAt            *time.Time          `json:"ends_at,omitempty"`
	Rules             []couponRuleRequest `json:"rules,omitempty"`
}

type couponPatchRequest struct {
	Code              *string              `json:"code,omitempty"`
	DiscountType      *string              `json:"discount_type,omitempty"`
	DiscountValue     *int64               `json:"discount_value,omitempty"`
	MaxDiscountAmount *amountValue         `json:"max_discount_amount,omitempty"`
	MinCartAmount     *amountValue         `json:"min_cart_amount,omitempty"`
	Currency          *string              `json:"currency,omitempty"`
	UsageLimit        *int64               `json:"usage_limit,omitempty"`
	PerUserLimit      *int64               `json:"per_user_limit,omitempty"`
	Status            *string              `json:"status,omitempty"`
	StartsAt          *time.Time           `json:"starts_at,omitempty"`
	EndsAt            *time.Time           `json:"ends_at,omitempty"`
	Rules             *[]couponRuleRequest `json:"rules,omitempty"`
}

type couponRuleRequest struct {
	RuleType  string                 `json:"rule_type"`
	RuleValue couponRuleValueRequest `json:"rule_value"`
}

type couponRuleValueRequest struct {
	ProductIDs  []string `json:"product_ids,omitempty"`
	CategoryIDs []string `json:"category_ids,omitempty"`
	SellerIDs   []string `json:"seller_ids,omitempty"`
}

type couponValidationRequest struct {
	CouponCode     string              `json:"coupon_code"`
	Code           string              `json:"code,omitempty"`
	CampaignID     string              `json:"campaign_id,omitempty"`
	UserID         string              `json:"user_id"`
	GuestSessionID string              `json:"guest_session_id,omitempty"`
	CartID         string              `json:"cart_id,omitempty"`
	OrderID        string              `json:"order_id,omitempty"`
	Currency       string              `json:"currency"`
	SubtotalAmount int64               `json:"subtotal_amount"`
	Subtotal       *amountValue        `json:"subtotal,omitempty"`
	CartSubtotal   *amountValue        `json:"cart_subtotal,omitempty"`
	Items          []couponItemRequest `json:"items,omitempty"`
}

type couponItemRequest struct {
	ProductID          string       `json:"product_id"`
	VariantID          string       `json:"variant_id,omitempty"`
	SellerID           string       `json:"seller_id"`
	CategoryID         string       `json:"category_id,omitempty"`
	CategoryIDs        []string     `json:"category_ids,omitempty"`
	Quantity           int64        `json:"quantity,omitempty"`
	UnitAmount         int64        `json:"unit_amount,omitempty"`
	UnitPrice          *amountValue `json:"unit_price,omitempty"`
	LineSubtotalAmount int64        `json:"line_subtotal_amount"`
	LineSubtotal       *amountValue `json:"line_subtotal,omitempty"`
}

type couponRedemptionRequest struct {
	CouponID       string       `json:"coupon_id"`
	CampaignID     string       `json:"campaign_id,omitempty"`
	OrderID        string       `json:"order_id"`
	UserID         string       `json:"user_id"`
	Discount       *amountValue `json:"discount,omitempty"`
	DiscountAmount int64        `json:"discount_amount,omitempty"`
	Currency       string       `json:"currency,omitempty"`
	RequestID      string       `json:"request_id,omitempty"`
}

type couponListResponse struct {
	Coupons []couponResponse `json:"coupons"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
}

type couponResponse struct {
	CouponID          string               `json:"coupon_id"`
	SellerID          string               `json:"seller_id,omitempty"`
	Code              string               `json:"code"`
	DiscountType      string               `json:"discount_type"`
	DiscountValue     int64                `json:"discount_value"`
	MaxDiscountAmount *moneyResponse       `json:"max_discount_amount,omitempty"`
	MinCartAmount     moneyResponse        `json:"min_cart_amount"`
	Currency          string               `json:"currency"`
	UsageLimit        *int64               `json:"usage_limit,omitempty"`
	PerUserLimit      *int64               `json:"per_user_limit,omitempty"`
	Status            string               `json:"status"`
	StartsAt          string               `json:"starts_at,omitempty"`
	EndsAt            string               `json:"ends_at,omitempty"`
	CreatedAt         string               `json:"created_at"`
	UpdatedAt         string               `json:"updated_at"`
	Rules             []couponRuleResponse `json:"rules,omitempty"`
}

type couponRuleResponse struct {
	RuleID    string                  `json:"rule_id"`
	RuleType  string                  `json:"rule_type"`
	RuleValue couponRuleValueResponse `json:"rule_value"`
}

type couponRuleValueResponse struct {
	ProductIDs  []string `json:"product_ids,omitempty"`
	CategoryIDs []string `json:"category_ids,omitempty"`
	SellerIDs   []string `json:"seller_ids,omitempty"`
}

type couponValidationResponse struct {
	Valid            bool          `json:"valid"`
	CouponID         string        `json:"coupon_id,omitempty"`
	CampaignID       string        `json:"campaign_id,omitempty"`
	Discount         moneyResponse `json:"discount"`
	Reason           string        `json:"reason,omitempty"`
	EligibleSubtotal int64         `json:"eligible_subtotal,omitempty"`
}

type couponRedemptionResponse struct {
	RedemptionID string        `json:"redemption_id"`
	CouponID     string        `json:"coupon_id"`
	CampaignID   string        `json:"campaign_id,omitempty"`
	OrderID      string        `json:"order_id"`
	UserID       string        `json:"user_id"`
	Discount     moneyResponse `json:"discount"`
	RedeemedAt   string        `json:"redeemed_at"`
}

type moneyResponse struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type amountValue struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency,omitempty"`
}

type campaignCreateRequest struct {
	Name     string                  `json:"name"`
	Budget   *amountValue            `json:"budget,omitempty"`
	Currency string                  `json:"currency,omitempty"`
	StartsAt time.Time               `json:"starts_at"`
	EndsAt   time.Time               `json:"ends_at"`
	Metadata campaignMetadataRequest `json:"metadata,omitempty"`
}

type campaignPatchRequest struct {
	Name     *string                  `json:"name,omitempty"`
	Budget   *amountValue             `json:"budget,omitempty"`
	Currency *string                  `json:"currency,omitempty"`
	StartsAt *time.Time               `json:"starts_at,omitempty"`
	EndsAt   *time.Time               `json:"ends_at,omitempty"`
	Metadata *campaignMetadataRequest `json:"metadata,omitempty"`
	Status   *string                  `json:"status,omitempty"`
}

type campaignMetadataRequest struct {
	CouponIDs                   []string `json:"coupon_ids,omitempty"`
	Channels                    []string `json:"channels,omitempty"`
	Description                 string   `json:"description,omitempty"`
	UsageLimit                  *int64   `json:"usage_limit,omitempty"`
	PerUserLimit                *int64   `json:"per_user_limit,omitempty"`
	BudgetAlertThresholdPercent *int64   `json:"budget_alert_threshold_percent,omitempty"`
}

type campaignListResponse struct {
	Campaigns []campaignResponse `json:"campaigns"`
	Limit     int                `json:"limit"`
	Offset    int                `json:"offset"`
}

type campaignResponse struct {
	CampaignID string                   `json:"campaign_id"`
	SellerID   string                   `json:"seller_id,omitempty"`
	Name       string                   `json:"name"`
	Status     string                   `json:"status"`
	Budget     *moneyResponse           `json:"budget,omitempty"`
	Currency   string                   `json:"currency"`
	StartsAt   string                   `json:"starts_at"`
	EndsAt     string                   `json:"ends_at"`
	Metadata   campaignMetadataResponse `json:"metadata,omitempty"`
	CreatedBy  string                   `json:"created_by,omitempty"`
	CreatedAt  string                   `json:"created_at"`
	UpdatedAt  string                   `json:"updated_at"`
}

type campaignMetadataResponse struct {
	CouponIDs                   []string `json:"coupon_ids,omitempty"`
	Channels                    []string `json:"channels,omitempty"`
	Description                 string   `json:"description,omitempty"`
	UsageLimit                  *int64   `json:"usage_limit,omitempty"`
	PerUserLimit                *int64   `json:"per_user_limit,omitempty"`
	BudgetAlertThresholdPercent *int64   `json:"budget_alert_threshold_percent,omitempty"`
}

type campaignValidationRequest struct {
	CampaignID     string `json:"campaign_id"`
	SellerID       string `json:"seller_id"`
	CouponID       string `json:"coupon_id,omitempty"`
	UserID         string `json:"user_id,omitempty"`
	Currency       string `json:"currency"`
	DiscountAmount int64  `json:"discount_amount,omitempty"`
}

type campaignValidationResponse struct {
	Valid           bool           `json:"valid"`
	CampaignID      string         `json:"campaign_id,omitempty"`
	CouponID        string         `json:"coupon_id,omitempty"`
	RemainingBudget *moneyResponse `json:"remaining_budget,omitempty"`
	BudgetUnlimited bool           `json:"budget_unlimited,omitempty"`
	Reason          string         `json:"reason,omitempty"`
}

type sellerAnalyticsResponse struct {
	Revenue        moneyResponse              `json:"revenue"`
	Orders         int64                      `json:"orders"`
	ConversionRate float64                    `json:"conversion_rate"`
	TopProducts    []topProductMetricResponse `json:"top_products"`
}

type topProductMetricResponse struct {
	ProductID string        `json:"product_id"`
	Title     string        `json:"title"`
	UnitsSold int64         `json:"units_sold"`
	Orders    int64         `json:"orders"`
	Revenue   moneyResponse `json:"revenue"`
}

type auditLogListResponse struct {
	AuditLogs  []auditLogResponse `json:"audit_logs"`
	NextCursor *string            `json:"next_cursor"`
}

type auditLogResponse struct {
	AuditID       string         `json:"audit_id"`
	SellerID      string         `json:"seller_id"`
	ActorUserID   string         `json:"actor_user_id"`
	ActorSellerID string         `json:"actor_seller_id,omitempty"`
	ActorRoles    []string       `json:"actor_roles,omitempty"`
	Action        string         `json:"action"`
	ResourceType  string         `json:"resource_type"`
	ResourceID    string         `json:"resource_id"`
	RequestID     string         `json:"request_id,omitempty"`
	TraceID       string         `json:"trace_id,omitempty"`
	Decision      string         `json:"decision"`
	Reason        string         `json:"reason,omitempty"`
	Before        map[string]any `json:"before,omitempty"`
	After         map[string]any `json:"after,omitempty"`
	CreatedAt     string         `json:"created_at"`
}

func (v *amountValue) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	if raw == "" || raw == "null" {
		return nil
	}
	if strings.HasPrefix(raw, "{") {
		type alias amountValue
		var value alias
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		v.Amount = value.Amount
		v.Currency = strings.TrimSpace(value.Currency)
		return nil
	}
	var amount int64
	if err := json.Unmarshal(data, &amount); err != nil {
		return fmt.Errorf("amount must be an integer minor-unit value or money object")
	}
	v.Amount = amount
	v.Currency = ""
	return nil
}

func (r couponValidationRequest) resolvedCouponCode() string {
	return firstNonEmpty(r.CouponCode, r.Code)
}

func (r couponValidationRequest) resolvedUserID() string {
	return firstNonEmpty(r.UserID, r.GuestSessionID, r.CartID)
}

func (r couponValidationRequest) resolvedCurrency() string {
	return firstNonEmpty(r.Currency, currencyFromAmount(r.Subtotal), currencyFromAmount(r.CartSubtotal))
}

func (r couponValidationRequest) resolvedSubtotalAmount() int64 {
	if r.SubtotalAmount != 0 {
		return r.SubtotalAmount
	}
	if r.Subtotal != nil {
		return r.Subtotal.Amount
	}
	if r.CartSubtotal != nil {
		return r.CartSubtotal.Amount
	}
	return 0
}

func (r couponItemRequest) resolvedCategoryIDs() []string {
	if strings.TrimSpace(r.CategoryID) == "" {
		return r.CategoryIDs
	}
	categoryIDs := make([]string, 0, len(r.CategoryIDs)+1)
	categoryIDs = append(categoryIDs, r.CategoryID)
	categoryIDs = append(categoryIDs, r.CategoryIDs...)
	return categoryIDs
}

func (r couponItemRequest) resolvedUnitAmount() int64 {
	if r.UnitAmount != 0 {
		return r.UnitAmount
	}
	if r.UnitPrice != nil {
		return r.UnitPrice.Amount
	}
	return 0
}

func (r couponItemRequest) resolvedLineSubtotalAmount() int64 {
	if r.LineSubtotalAmount != 0 {
		return r.LineSubtotalAmount
	}
	if r.LineSubtotal != nil {
		return r.LineSubtotal.Amount
	}
	return 0
}

func currencyFromAmount(value *amountValue) string {
	if value == nil {
		return ""
	}
	return value.Currency
}
