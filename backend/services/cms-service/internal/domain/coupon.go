package domain

import (
	"strings"
	"time"
)

const (
	defaultCouponCurrency = "INR"
	minCouponCodeLength   = 3
	maxCouponCodeLength   = 64
)

type DiscountType string

const (
	DiscountTypeFixed      DiscountType = "fixed"
	DiscountTypePercentage DiscountType = "percentage"
)

type CouponStatus string

const (
	CouponStatusDraft   CouponStatus = "draft"
	CouponStatusActive  CouponStatus = "active"
	CouponStatusPaused  CouponStatus = "paused"
	CouponStatusExpired CouponStatus = "expired"
)

type CouponRuleType string

const (
	CouponRuleTypeProductScope  CouponRuleType = "product_scope"
	CouponRuleTypeCategoryScope CouponRuleType = "category_scope"
	CouponRuleTypeSellerScope   CouponRuleType = "seller_scope"
)

type CouponInvalidReason string

const (
	CouponInvalidReasonNotFound                  CouponInvalidReason = "coupon_not_found"
	CouponInvalidReasonNotActive                 CouponInvalidReason = "coupon_not_active"
	CouponInvalidReasonNotStarted                CouponInvalidReason = "coupon_not_started"
	CouponInvalidReasonExpired                   CouponInvalidReason = "coupon_expired"
	CouponInvalidReasonCurrencyMismatch          CouponInvalidReason = "currency_mismatch"
	CouponInvalidReasonMinCartNotMet             CouponInvalidReason = "min_cart_not_met"
	CouponInvalidReasonScopeNotMatched           CouponInvalidReason = "scope_not_matched"
	CouponInvalidReasonUsageLimitReached         CouponInvalidReason = "usage_limit_reached"
	CouponInvalidReasonPerUserLimitReached       CouponInvalidReason = "per_user_limit_reached"
	CouponInvalidReasonInvalidDiscountSetup      CouponInvalidReason = "invalid_discount_config"
	CouponInvalidReasonCampaignNotFound          CouponInvalidReason = CouponInvalidReason(CampaignInvalidReasonNotFound)
	CouponInvalidReasonCampaignSellerMismatch    CouponInvalidReason = CouponInvalidReason(CampaignInvalidReasonSellerMismatch)
	CouponInvalidReasonCampaignNotActive         CouponInvalidReason = CouponInvalidReason(CampaignInvalidReasonNotActive)
	CouponInvalidReasonCampaignNotStarted        CouponInvalidReason = CouponInvalidReason(CampaignInvalidReasonNotStarted)
	CouponInvalidReasonCampaignExpired           CouponInvalidReason = CouponInvalidReason(CampaignInvalidReasonExpired)
	CouponInvalidReasonCampaignUsageLimitReached CouponInvalidReason = CouponInvalidReason(CampaignInvalidReasonUsageLimitReached)
	CouponInvalidReasonCampaignBudgetExhausted   CouponInvalidReason = CouponInvalidReason(CampaignInvalidReasonBudgetExhausted)
	CouponInvalidReasonCouponNotInCampaign       CouponInvalidReason = CouponInvalidReason(CampaignInvalidReasonCouponNotLinked)
)

type Coupon struct {
	CouponID          string
	SellerID          *string
	Code              string
	DiscountType      DiscountType
	DiscountValue     int64
	MaxDiscountAmount *int64
	MinCartAmount     int64
	Currency          string
	UsageLimit        *int64
	PerUserLimit      *int64
	Status            CouponStatus
	StartsAt          *time.Time
	EndsAt            *time.Time
	CreatedBy         string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Rules             []CouponRule
}

type CouponRule struct {
	RuleID      string
	CouponID    string
	Type        CouponRuleType
	ProductIDs  []string
	CategoryIDs []string
	SellerIDs   []string
	CreatedAt   time.Time
}

type CouponCartItem struct {
	ProductID          string
	VariantID          string
	SellerID           string
	CategoryIDs        []string
	Quantity           int64
	UnitAmount         int64
	LineSubtotalAmount int64
}

type CouponValidationResult struct {
	Valid            bool
	CouponID         string
	CampaignID       string
	Discount         Money
	Reason           CouponInvalidReason
	AppliedRuleCodes []string
	EligibleSubtotal int64
}

type CouponRedemption struct {
	RedemptionID string
	CouponID     string
	CampaignID   string
	OrderID      string
	UserID       string
	Discount     Money
	RequestID    string
	RedeemedAt   time.Time
	CreatedAt    time.Time
}

func (t DiscountType) Normalized() DiscountType {
	return DiscountType(strings.ToLower(strings.TrimSpace(string(t))))
}

func (t DiscountType) Valid() bool {
	switch t.Normalized() {
	case DiscountTypeFixed, DiscountTypePercentage:
		return true
	default:
		return false
	}
}

func (s CouponStatus) Normalized() CouponStatus {
	return CouponStatus(strings.ToLower(strings.TrimSpace(string(s))))
}

func (s CouponStatus) Valid() bool {
	switch s.Normalized() {
	case CouponStatusDraft, CouponStatusActive, CouponStatusPaused, CouponStatusExpired:
		return true
	default:
		return false
	}
}

func (t CouponRuleType) Normalized() CouponRuleType {
	return CouponRuleType(strings.ToLower(strings.TrimSpace(string(t))))
}

func (t CouponRuleType) Valid() bool {
	switch t.Normalized() {
	case CouponRuleTypeProductScope, CouponRuleTypeCategoryScope, CouponRuleTypeSellerScope:
		return true
	default:
		return false
	}
}

func (c Coupon) Normalized() Coupon {
	c.CouponID = strings.TrimSpace(c.CouponID)
	c.Code = NormalizeCouponCode(c.Code)
	c.DiscountType = c.DiscountType.Normalized()
	c.Status = c.Status.Normalized()
	if c.Status == "" {
		c.Status = CouponStatusDraft
	}
	c.Currency = NormalizeCurrency(c.Currency)
	c.CreatedBy = strings.TrimSpace(c.CreatedBy)
	if c.SellerID != nil {
		sellerID := strings.TrimSpace(*c.SellerID)
		if sellerID == "" {
			c.SellerID = nil
		} else {
			c.SellerID = &sellerID
		}
	}
	c.Rules = NormalizeCouponRules(c.Rules)
	return c
}

func (r CouponRule) Normalized() CouponRule {
	r.RuleID = strings.TrimSpace(r.RuleID)
	r.CouponID = strings.TrimSpace(r.CouponID)
	r.Type = r.Type.Normalized()
	r.ProductIDs = normalizeIDList(r.ProductIDs)
	r.CategoryIDs = normalizeIDList(r.CategoryIDs)
	r.SellerIDs = normalizeIDList(r.SellerIDs)
	return r
}

func NormalizeCouponRules(rules []CouponRule) []CouponRule {
	out := make([]CouponRule, 0, len(rules))
	for _, rule := range rules {
		out = append(out, rule.Normalized())
	}
	return out
}

func NormalizeCouponCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func NormalizeCurrency(currency string) string {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		return defaultCouponCurrency
	}
	return currency
}

func ValidateCouponDefinition(coupon Coupon, rules []CouponRule) error {
	coupon = coupon.Normalized()
	rules = NormalizeCouponRules(rules)

	fields := make([]FieldViolation, 0)
	if !validCouponCode(coupon.Code) {
		fields = append(fields, FieldViolation{Field: "code", Reason: "Coupon code must be 3-64 characters using letters, numbers, underscores, or hyphens."})
	}
	if !coupon.DiscountType.Valid() {
		fields = append(fields, FieldViolation{Field: "discount_type", Reason: "Discount type must be fixed or percentage."})
	}
	if coupon.DiscountValue <= 0 {
		fields = append(fields, FieldViolation{Field: "discount_value", Reason: "Discount value must be greater than zero."})
	}
	if coupon.DiscountType.Normalized() == DiscountTypePercentage && (coupon.DiscountValue < 1 || coupon.DiscountValue > 100) {
		fields = append(fields, FieldViolation{Field: "discount_value", Reason: "Percentage discount value must be between 1 and 100."})
	}
	if coupon.MaxDiscountAmount != nil && *coupon.MaxDiscountAmount <= 0 {
		fields = append(fields, FieldViolation{Field: "max_discount_amount", Reason: "Max discount amount must be greater than zero when set."})
	}
	if coupon.MinCartAmount < 0 {
		fields = append(fields, FieldViolation{Field: "min_cart_amount", Reason: "Minimum cart amount cannot be negative."})
	}
	if !validCouponCurrency(coupon.Currency) {
		fields = append(fields, FieldViolation{Field: "currency", Reason: "Currency must be a 3-letter ISO-style code."})
	}
	if coupon.UsageLimit != nil && *coupon.UsageLimit <= 0 {
		fields = append(fields, FieldViolation{Field: "usage_limit", Reason: "Usage limit must be greater than zero when set."})
	}
	if coupon.PerUserLimit != nil && *coupon.PerUserLimit <= 0 {
		fields = append(fields, FieldViolation{Field: "per_user_limit", Reason: "Per-user limit must be greater than zero when set."})
	}
	if !coupon.Status.Valid() {
		fields = append(fields, FieldViolation{Field: "status", Reason: "Coupon status must be draft, active, paused, or expired."})
	}
	if coupon.StartsAt != nil && coupon.EndsAt != nil && !coupon.StartsAt.Before(*coupon.EndsAt) {
		fields = append(fields, FieldViolation{Field: "ends_at", Reason: "End time must be after start time."})
	}

	for _, rule := range rules {
		if !rule.Type.Valid() {
			fields = append(fields, FieldViolation{Field: "rules", Reason: "Rule type must be product_scope, category_scope, or seller_scope."})
			continue
		}
		switch rule.Type.Normalized() {
		case CouponRuleTypeProductScope:
			if len(rule.ProductIDs) == 0 {
				fields = append(fields, FieldViolation{Field: "rules.product_ids", Reason: "Product scope requires at least one product id."})
			}
		case CouponRuleTypeCategoryScope:
			if len(rule.CategoryIDs) == 0 {
				fields = append(fields, FieldViolation{Field: "rules.category_ids", Reason: "Category scope requires at least one category id."})
			}
		case CouponRuleTypeSellerScope:
			if len(rule.SellerIDs) == 0 {
				fields = append(fields, FieldViolation{Field: "rules.seller_ids", Reason: "Seller scope requires at least one seller id."})
			}
		}
	}

	if len(fields) > 0 {
		return NewValidationError("Invalid coupon.", fields)
	}
	return nil
}

func ValidateCouponEligibility(coupon Coupon, currency string, subtotalAmount int64, now time.Time) CouponInvalidReason {
	coupon = coupon.Normalized()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if coupon.Status != CouponStatusActive {
		return CouponInvalidReasonNotActive
	}
	if coupon.StartsAt != nil && now.Before(coupon.StartsAt.UTC()) {
		return CouponInvalidReasonNotStarted
	}
	if coupon.EndsAt != nil && now.After(coupon.EndsAt.UTC()) {
		return CouponInvalidReasonExpired
	}
	if coupon.Currency != NormalizeCurrency(currency) {
		return CouponInvalidReasonCurrencyMismatch
	}
	if subtotalAmount < coupon.MinCartAmount {
		return CouponInvalidReasonMinCartNotMet
	}
	return ""
}

func CalculateEligibleSubtotal(cartSubtotal int64, items []CouponCartItem, coupon Coupon, rules []CouponRule) int64 {
	if cartSubtotal < 0 {
		return 0
	}

	coupon = coupon.Normalized()
	rules = NormalizeCouponRules(rules)
	scope := newCouponScope(coupon, rules)
	if !scope.hasAnyScope() {
		return cartSubtotal
	}

	var total int64
	for _, item := range items {
		item = item.Normalized()
		if item.LineSubtotalAmount <= 0 {
			continue
		}
		if scope.matches(item) {
			total += item.LineSubtotalAmount
		}
	}
	if total > cartSubtotal {
		return cartSubtotal
	}
	return total
}

func CalculateCouponDiscount(coupon Coupon, eligibleSubtotal int64) int64 {
	coupon = coupon.Normalized()
	if eligibleSubtotal <= 0 || coupon.DiscountValue <= 0 {
		return 0
	}

	var discount int64
	switch coupon.DiscountType {
	case DiscountTypeFixed:
		discount = coupon.DiscountValue
	case DiscountTypePercentage:
		if coupon.DiscountValue > 100 {
			return 0
		}
		discount = eligibleSubtotal * coupon.DiscountValue / 100
		if coupon.MaxDiscountAmount != nil && discount > *coupon.MaxDiscountAmount {
			discount = *coupon.MaxDiscountAmount
		}
	default:
		return 0
	}
	if discount > eligibleSubtotal {
		return eligibleSubtotal
	}
	if discount < 0 {
		return 0
	}
	return discount
}

func InvalidCouponResult(currency string, couponID string, reason CouponInvalidReason) CouponValidationResult {
	return CouponValidationResult{
		Valid:    false,
		CouponID: strings.TrimSpace(couponID),
		Discount: Money{
			Amount:   0,
			Currency: NormalizeCurrency(currency),
		},
		Reason: reason,
	}
}

func (item CouponCartItem) Normalized() CouponCartItem {
	item.ProductID = strings.TrimSpace(item.ProductID)
	item.VariantID = strings.TrimSpace(item.VariantID)
	item.SellerID = strings.TrimSpace(item.SellerID)
	item.CategoryIDs = normalizeIDList(item.CategoryIDs)
	if item.LineSubtotalAmount <= 0 && item.Quantity > 0 && item.UnitAmount > 0 {
		item.LineSubtotalAmount = item.Quantity * item.UnitAmount
	}
	return item
}

type couponScope struct {
	productIDs  map[string]struct{}
	categoryIDs map[string]struct{}
	sellerIDs   map[string]struct{}
}

func newCouponScope(coupon Coupon, rules []CouponRule) couponScope {
	scope := couponScope{}
	if coupon.SellerID != nil && strings.TrimSpace(*coupon.SellerID) != "" {
		scope.addSellerID(*coupon.SellerID)
	}
	for _, rule := range rules {
		switch rule.Type.Normalized() {
		case CouponRuleTypeProductScope:
			for _, id := range rule.ProductIDs {
				scope.addProductID(id)
			}
		case CouponRuleTypeCategoryScope:
			for _, id := range rule.CategoryIDs {
				scope.addCategoryID(id)
			}
		case CouponRuleTypeSellerScope:
			for _, id := range rule.SellerIDs {
				scope.addSellerID(id)
			}
		}
	}
	return scope
}

func (s couponScope) hasAnyScope() bool {
	return len(s.productIDs) > 0 || len(s.categoryIDs) > 0 || len(s.sellerIDs) > 0
}

func (s couponScope) matches(item CouponCartItem) bool {
	if len(s.sellerIDs) > 0 {
		if _, ok := s.sellerIDs[item.SellerID]; !ok {
			return false
		}
	}
	if len(s.productIDs) > 0 {
		if _, ok := s.productIDs[item.ProductID]; !ok {
			return false
		}
	}
	if len(s.categoryIDs) > 0 {
		matched := false
		for _, categoryID := range item.CategoryIDs {
			if _, ok := s.categoryIDs[categoryID]; ok {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func (s *couponScope) addProductID(id string) {
	id = strings.TrimSpace(id)
	if id == "" {
		return
	}
	if s.productIDs == nil {
		s.productIDs = make(map[string]struct{})
	}
	s.productIDs[id] = struct{}{}
}

func (s *couponScope) addCategoryID(id string) {
	id = strings.TrimSpace(id)
	if id == "" {
		return
	}
	if s.categoryIDs == nil {
		s.categoryIDs = make(map[string]struct{})
	}
	s.categoryIDs[id] = struct{}{}
}

func (s *couponScope) addSellerID(id string) {
	id = strings.TrimSpace(id)
	if id == "" {
		return
	}
	if s.sellerIDs == nil {
		s.sellerIDs = make(map[string]struct{})
	}
	s.sellerIDs[id] = struct{}{}
}

func validCouponCode(code string) bool {
	if len(code) < minCouponCodeLength || len(code) > maxCouponCodeLength {
		return false
	}
	for _, r := range code {
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		if r == '_' || r == '-' {
			continue
		}
		return false
	}
	return true
}

func validCouponCurrency(currency string) bool {
	currency = NormalizeCurrency(currency)
	if len(currency) != 3 {
		return false
	}
	for _, r := range currency {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

func normalizeIDList(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
