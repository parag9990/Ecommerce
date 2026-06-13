package domain

import (
	"errors"
	"testing"
	"time"
)

func TestValidateCampaignDefinitionRejectsInvalidWindowAndBudget(t *testing.T) {
	budget := int64(-1)
	err := ValidateCampaignDefinition(Campaign{
		Name:         "Weekend sale",
		Status:       CampaignStatusDraft,
		BudgetAmount: &budget,
		Currency:     "INR",
		StartsAt:     time.Unix(2000, 0).UTC(),
		EndsAt:       time.Unix(1000, 0).UTC(),
	}, 90*24*time.Hour)

	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestValidateCampaignEligibilityChecksOwnershipStatusWindowAndCoupon(t *testing.T) {
	sellerID := "seller_1"
	campaign := Campaign{
		CampaignID: "camp_1",
		SellerID:   &sellerID,
		Name:       "Festive sale",
		Status:     CampaignStatusActive,
		Currency:   "INR",
		StartsAt:   time.Unix(1000, 0).UTC(),
		EndsAt:     time.Unix(3000, 0).UTC(),
		Metadata: CampaignMetadata{
			CouponIDs: []string{"coupon_1"},
		},
	}

	if reason := ValidateCampaignEligibility(campaign, "seller_2", "coupon_1", "INR", time.Unix(2000, 0).UTC()); reason != CampaignInvalidReasonSellerMismatch {
		t.Fatalf("expected seller mismatch, got %q", reason)
	}
	if reason := ValidateCampaignEligibility(campaign, "seller_1", "coupon_2", "INR", time.Unix(2000, 0).UTC()); reason != CampaignInvalidReasonCouponNotLinked {
		t.Fatalf("expected coupon not linked, got %q", reason)
	}
	if reason := ValidateCampaignEligibility(campaign, "seller_1", "coupon_1", "INR", time.Unix(4000, 0).UTC()); reason != CampaignInvalidReasonExpired {
		t.Fatalf("expected expired, got %q", reason)
	}
	if reason := ValidateCampaignEligibility(campaign, "seller_1", "coupon_1", "INR", time.Unix(2000, 0).UTC()); reason != "" {
		t.Fatalf("expected valid campaign, got %q", reason)
	}
}

func TestValidateCampaignStatusTransitionKeepsCompletedTerminal(t *testing.T) {
	err := ValidateCampaignStatusTransition(CampaignStatusCompleted, CampaignStatusActive)
	if !errors.Is(err, ErrCampaignCompletedReadOnly) {
		t.Fatalf("expected completed read-only error, got %v", err)
	}
}
