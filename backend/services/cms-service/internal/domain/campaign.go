package domain

import (
	"strings"
	"time"
)

const maxCampaignNameLength = 255

type CampaignStatus string

const (
	CampaignStatusDraft     CampaignStatus = "draft"
	CampaignStatusActive    CampaignStatus = "active"
	CampaignStatusPaused    CampaignStatus = "paused"
	CampaignStatusCompleted CampaignStatus = "completed"
)

type CampaignInvalidReason string

const (
	CampaignInvalidReasonNotFound          CampaignInvalidReason = "campaign_not_found"
	CampaignInvalidReasonSellerMismatch    CampaignInvalidReason = "seller_scope_mismatch"
	CampaignInvalidReasonNotActive         CampaignInvalidReason = "campaign_not_active"
	CampaignInvalidReasonNotStarted        CampaignInvalidReason = "campaign_not_started"
	CampaignInvalidReasonExpired           CampaignInvalidReason = "campaign_expired"
	CampaignInvalidReasonUsageLimitReached CampaignInvalidReason = "campaign_usage_limit_reached"
	CampaignInvalidReasonBudgetExhausted   CampaignInvalidReason = "campaign_budget_exhausted"
	CampaignInvalidReasonCouponNotLinked   CampaignInvalidReason = "coupon_not_in_campaign"
	CampaignInvalidReasonCurrencyMismatch  CampaignInvalidReason = "currency_mismatch"
)

type Campaign struct {
	CampaignID   string
	SellerID     *string
	Name         string
	Status       CampaignStatus
	BudgetAmount *int64
	Currency     string
	StartsAt     time.Time
	EndsAt       time.Time
	Metadata     CampaignMetadata
	CreatedBy    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CampaignMetadata struct {
	CouponIDs                   []string `json:"coupon_ids,omitempty"`
	Channels                    []string `json:"channels,omitempty"`
	Description                 string   `json:"description,omitempty"`
	UsageLimit                  *int64   `json:"usage_limit,omitempty"`
	PerUserLimit                *int64   `json:"per_user_limit,omitempty"`
	BudgetAlertThresholdPercent *int64   `json:"budget_alert_threshold_percent,omitempty"`
}

type CampaignValidationResult struct {
	Valid           bool
	CampaignID      string
	CouponID        string
	RemainingBudget Money
	BudgetUnlimited bool
	Reason          CampaignInvalidReason
}

func (s CampaignStatus) Normalized() CampaignStatus {
	return CampaignStatus(strings.ToLower(strings.TrimSpace(string(s))))
}

func (s CampaignStatus) Valid() bool {
	switch s.Normalized() {
	case CampaignStatusDraft, CampaignStatusActive, CampaignStatusPaused, CampaignStatusCompleted:
		return true
	default:
		return false
	}
}

func (c Campaign) Normalized() Campaign {
	c.CampaignID = strings.TrimSpace(c.CampaignID)
	c.Name = strings.TrimSpace(c.Name)
	c.Status = c.Status.Normalized()
	if c.Status == "" {
		c.Status = CampaignStatusDraft
	}
	c.Currency = NormalizeCurrency(c.Currency)
	c.StartsAt = c.StartsAt.UTC()
	c.EndsAt = c.EndsAt.UTC()
	c.CreatedBy = strings.TrimSpace(c.CreatedBy)
	if c.SellerID != nil {
		sellerID := strings.TrimSpace(*c.SellerID)
		if sellerID == "" {
			c.SellerID = nil
		} else {
			c.SellerID = &sellerID
		}
	}
	c.Metadata = c.Metadata.Normalized()
	return c
}

func (m CampaignMetadata) Normalized() CampaignMetadata {
	m.CouponIDs = normalizeIDList(m.CouponIDs)
	m.Channels = normalizeIDList(m.Channels)
	m.Description = strings.TrimSpace(m.Description)
	return m
}

func (c Campaign) IsInsideWindow(now time.Time) bool {
	c = c.Normalized()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()
	return !now.Before(c.StartsAt) && !now.After(c.EndsAt)
}

func (c Campaign) IsSellerOwnedBy(sellerID string) bool {
	c = c.Normalized()
	return c.SellerID != nil && *c.SellerID == strings.TrimSpace(sellerID)
}

func (c Campaign) CouponAllowed(couponID string) bool {
	c = c.Normalized()
	couponID = strings.TrimSpace(couponID)
	if couponID == "" || len(c.Metadata.CouponIDs) == 0 {
		return true
	}
	for _, id := range c.Metadata.CouponIDs {
		if id == couponID {
			return true
		}
	}
	return false
}

func ValidateCampaignDefinition(campaign Campaign, maxDuration time.Duration) error {
	campaign = campaign.Normalized()
	fields := make([]FieldViolation, 0)

	if campaign.Name == "" || len(campaign.Name) > maxCampaignNameLength {
		fields = append(fields, FieldViolation{Field: "name", Reason: "Campaign name is required and must be 255 characters or fewer."})
	}
	if !campaign.Status.Valid() {
		fields = append(fields, FieldViolation{Field: "status", Reason: "Campaign status must be draft, active, paused, or completed."})
	}
	if campaign.StartsAt.IsZero() {
		fields = append(fields, FieldViolation{Field: "starts_at", Reason: "Start time is required."})
	}
	if campaign.EndsAt.IsZero() {
		fields = append(fields, FieldViolation{Field: "ends_at", Reason: "End time is required."})
	}
	if !campaign.StartsAt.IsZero() && !campaign.EndsAt.IsZero() && !campaign.StartsAt.Before(campaign.EndsAt) {
		fields = append(fields, FieldViolation{Field: "ends_at", Reason: "End time must be after start time."})
	}
	if maxDuration > 0 && !campaign.StartsAt.IsZero() && !campaign.EndsAt.IsZero() && campaign.EndsAt.Sub(campaign.StartsAt) > maxDuration {
		fields = append(fields, FieldViolation{Field: "ends_at", Reason: "Campaign duration exceeds the configured maximum."})
	}
	if campaign.BudgetAmount != nil && *campaign.BudgetAmount <= 0 {
		fields = append(fields, FieldViolation{Field: "budget.amount", Reason: "Budget amount must be greater than zero when set."})
	}
	if !validCampaignCurrency(campaign.Currency) {
		fields = append(fields, FieldViolation{Field: "currency", Reason: "Currency must be a 3-letter ISO-style code."})
	}
	if campaign.Metadata.UsageLimit != nil && *campaign.Metadata.UsageLimit <= 0 {
		fields = append(fields, FieldViolation{Field: "metadata.usage_limit", Reason: "Usage limit must be greater than zero when set."})
	}
	if campaign.Metadata.PerUserLimit != nil && *campaign.Metadata.PerUserLimit <= 0 {
		fields = append(fields, FieldViolation{Field: "metadata.per_user_limit", Reason: "Per-user limit must be greater than zero when set."})
	}
	if campaign.Metadata.BudgetAlertThresholdPercent != nil {
		value := *campaign.Metadata.BudgetAlertThresholdPercent
		if value <= 0 || value > 100 {
			fields = append(fields, FieldViolation{Field: "metadata.budget_alert_threshold_percent", Reason: "Budget alert threshold must be between 1 and 100."})
		}
	}

	if len(fields) > 0 {
		return NewValidationError("Invalid campaign.", fields)
	}
	return nil
}

func ValidateCampaignStatusTransition(from CampaignStatus, to CampaignStatus) error {
	from = from.Normalized()
	to = to.Normalized()
	if !from.Valid() || !to.Valid() {
		return ErrInvalidCampaignStatus
	}
	if from == to {
		return nil
	}
	if from == CampaignStatusCompleted {
		return ErrCampaignCompletedReadOnly
	}
	switch from {
	case CampaignStatusDraft:
		if to == CampaignStatusActive || to == CampaignStatusPaused || to == CampaignStatusCompleted {
			return nil
		}
	case CampaignStatusActive:
		if to == CampaignStatusPaused || to == CampaignStatusCompleted {
			return nil
		}
	case CampaignStatusPaused:
		if to == CampaignStatusActive || to == CampaignStatusCompleted {
			return nil
		}
	}
	return ErrInvalidCampaignStatusTransition
}

func ValidateCampaignEligibility(campaign Campaign, sellerID string, couponID string, currency string, now time.Time) CampaignInvalidReason {
	campaign = campaign.Normalized()
	if !campaign.IsSellerOwnedBy(sellerID) {
		return CampaignInvalidReasonSellerMismatch
	}
	if campaign.Status != CampaignStatusActive {
		return CampaignInvalidReasonNotActive
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()
	if now.Before(campaign.StartsAt) {
		return CampaignInvalidReasonNotStarted
	}
	if now.After(campaign.EndsAt) {
		return CampaignInvalidReasonExpired
	}
	if campaign.Currency != NormalizeCurrency(currency) {
		return CampaignInvalidReasonCurrencyMismatch
	}
	if !campaign.CouponAllowed(couponID) {
		return CampaignInvalidReasonCouponNotLinked
	}
	return ""
}

func InvalidCampaignResult(currency string, campaignID string, couponID string, reason CampaignInvalidReason) CampaignValidationResult {
	return CampaignValidationResult{
		Valid:      false,
		CampaignID: strings.TrimSpace(campaignID),
		CouponID:   strings.TrimSpace(couponID),
		RemainingBudget: Money{
			Amount:   0,
			Currency: NormalizeCurrency(currency),
		},
		Reason: reason,
	}
}

func validCampaignCurrency(currency string) bool {
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
