package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
)

func TestCreateCampaignRejectsCrossSellerCoupon(t *testing.T) {
	repo := newMemoryCampaignRepository()
	coupons := newMemoryCouponRepository()
	otherSeller := "seller_2"
	coupons.coupons["coupon_1"] = domain.Coupon{
		CouponID:      "coupon_1",
		SellerID:      &otherSeller,
		Code:          "SAVE10",
		DiscountType:  domain.DiscountTypeFixed,
		DiscountValue: 1000,
		Currency:      "INR",
		Status:        domain.CouponStatusActive,
	}
	uc := newTestCampaignUsecase(t, repo, coupons)

	_, err := uc.CreateCampaign(context.Background(), CreateCampaignInput{
		Actor:    activeActor(domain.RoleSellerCatalogEditor),
		Name:     "Festive sale",
		Currency: "INR",
		StartsAt: time.Unix(1700000100, 0).UTC(),
		EndsAt:   time.Unix(1700000200, 0).UTC(),
		Metadata: domain.CampaignMetadata{
			CouponIDs: []string{"coupon_1"},
		},
	})

	var validationErr domain.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestValidateCampaignReturnsBudgetExhausted(t *testing.T) {
	repo := newMemoryCampaignRepository()
	sellerID := "seller_1"
	budget := int64(1000)
	repo.campaigns["camp_1"] = domain.Campaign{
		CampaignID:   "camp_1",
		SellerID:     &sellerID,
		Name:         "Festive sale",
		Status:       domain.CampaignStatusActive,
		BudgetAmount: &budget,
		Currency:     "INR",
		StartsAt:     time.Unix(1699999900, 0).UTC(),
		EndsAt:       time.Unix(1700001000, 0).UTC(),
		Metadata:     domain.CampaignMetadata{CouponIDs: []string{"coupon_1"}},
	}
	repo.redemptions = append(repo.redemptions, domain.CouponRedemption{
		CampaignID: "camp_1",
		CouponID:   "coupon_1",
		UserID:     "user_1",
		Discount:   domain.Money{Amount: 1000, Currency: "INR"},
	})
	uc := newTestCampaignUsecase(t, repo, newMemoryCouponRepository())

	result, err := uc.ValidateCampaign(context.Background(), ValidateCampaignInput{
		CampaignID:     "camp_1",
		SellerID:       "seller_1",
		CouponID:       "coupon_1",
		UserID:         "user_2",
		Currency:       "INR",
		DiscountAmount: 1,
	})
	if err != nil {
		t.Fatalf("ValidateCampaign: %v", err)
	}
	if result.Valid || result.Reason != domain.CampaignInvalidReasonBudgetExhausted {
		t.Fatalf("expected budget exhausted, got %+v", result)
	}
}

func TestUpdateActiveCampaignRejectsBudgetDecrease(t *testing.T) {
	repo := newMemoryCampaignRepository()
	sellerID := "seller_1"
	budget := int64(1000)
	repo.campaigns["camp_1"] = domain.Campaign{
		CampaignID:   "camp_1",
		SellerID:     &sellerID,
		Name:         "Festive sale",
		Status:       domain.CampaignStatusActive,
		BudgetAmount: &budget,
		Currency:     "INR",
		StartsAt:     time.Unix(1699999900, 0).UTC(),
		EndsAt:       time.Unix(1700001000, 0).UTC(),
	}
	uc := newTestCampaignUsecase(t, repo, newMemoryCouponRepository())

	smallerBudget := int64(500)
	_, err := uc.UpdateCampaign(context.Background(), UpdateCampaignInput{
		Actor:        activeActor(domain.RoleSellerCatalogEditor),
		CampaignID:   "camp_1",
		BudgetAmount: &smallerBudget,
	})

	var validationErr domain.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func newTestCampaignUsecase(t *testing.T, repo CampaignRepository, coupons CampaignCouponReader) *CampaignUsecase {
	t.Helper()
	uc, err := NewCampaignUsecase(
		newTestAuthorizer(t, nil, nil),
		repo,
		coupons,
		auditRecorderFunc(func(context.Context, domain.AuditEvent) error { return nil }),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		CampaignOptions{MaxDuration: 90 * 24 * time.Hour},
	)
	if err != nil {
		t.Fatalf("NewCampaignUsecase: %v", err)
	}
	uc.now = func() time.Time { return time.Unix(1700000000, 0).UTC() }
	uc.newID = func(prefix string) string { return prefix + "_test" }
	return uc
}

type memoryCampaignRepository struct {
	campaigns   map[string]domain.Campaign
	redemptions []domain.CouponRedemption
}

func newMemoryCampaignRepository() *memoryCampaignRepository {
	return &memoryCampaignRepository{
		campaigns: make(map[string]domain.Campaign),
	}
}

func (r *memoryCampaignRepository) CreateCampaign(_ context.Context, campaign domain.Campaign) (domain.Campaign, error) {
	r.campaigns[campaign.CampaignID] = campaign.Normalized()
	return r.GetCampaignByID(context.Background(), campaign.CampaignID)
}

func (r *memoryCampaignRepository) UpdateCampaign(_ context.Context, campaign domain.Campaign) (domain.Campaign, error) {
	if _, ok := r.campaigns[campaign.CampaignID]; !ok {
		return domain.Campaign{}, domain.ErrCampaignNotFound
	}
	r.campaigns[campaign.CampaignID] = campaign.Normalized()
	return r.GetCampaignByID(context.Background(), campaign.CampaignID)
}

func (r *memoryCampaignRepository) UpdateCampaignStatus(_ context.Context, campaignID string, sellerID string, status domain.CampaignStatus, updatedAt time.Time) (domain.Campaign, error) {
	campaign, ok := r.campaigns[campaignID]
	if !ok || campaign.SellerID == nil || *campaign.SellerID != sellerID {
		return domain.Campaign{}, domain.ErrCampaignNotFound
	}
	campaign.Status = status
	campaign.UpdatedAt = updatedAt
	r.campaigns[campaignID] = campaign
	return campaign.Normalized(), nil
}

func (r *memoryCampaignRepository) GetCampaignByID(_ context.Context, campaignID string) (domain.Campaign, error) {
	campaign, ok := r.campaigns[campaignID]
	if !ok {
		return domain.Campaign{}, domain.ErrCampaignNotFound
	}
	return campaign.Normalized(), nil
}

func (r *memoryCampaignRepository) ListCampaigns(_ context.Context, filter CampaignFilter) ([]domain.Campaign, error) {
	out := make([]domain.Campaign, 0, len(r.campaigns))
	for _, campaign := range r.campaigns {
		campaign = campaign.Normalized()
		if filter.SellerID != "" && (campaign.SellerID == nil || *campaign.SellerID != filter.SellerID) {
			continue
		}
		if filter.Status.Normalized() != "" && campaign.Status != filter.Status.Normalized() {
			continue
		}
		out = append(out, campaign)
	}
	return out, nil
}

func (r *memoryCampaignRepository) CountCampaignRedemptions(_ context.Context, campaignID string) (int64, error) {
	var count int64
	for _, redemption := range r.redemptions {
		if redemption.CampaignID == campaignID {
			count++
		}
	}
	return count, nil
}

func (r *memoryCampaignRepository) CountCampaignUserRedemptions(_ context.Context, campaignID string, userID string) (int64, error) {
	var count int64
	for _, redemption := range r.redemptions {
		if redemption.CampaignID == campaignID && redemption.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (r *memoryCampaignRepository) SumCampaignDiscounts(_ context.Context, campaignID string) (int64, error) {
	var total int64
	for _, redemption := range r.redemptions {
		if redemption.CampaignID == campaignID {
			total += redemption.Discount.Amount
		}
	}
	return total, nil
}
