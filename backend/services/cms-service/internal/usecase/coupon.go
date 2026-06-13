package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
)

const (
	defaultCouponListLimit = 50
	maxCouponListLimit     = 500
)

type CouponRepository interface {
	CreateCoupon(ctx context.Context, coupon domain.Coupon, rules []domain.CouponRule) (domain.Coupon, error)
	UpdateCoupon(ctx context.Context, coupon domain.Coupon, rules []domain.CouponRule) (domain.Coupon, error)
	UpdateCouponStatus(ctx context.Context, couponID string, status domain.CouponStatus, updatedAt time.Time) (domain.Coupon, error)
	GetCouponByID(ctx context.Context, couponID string) (domain.Coupon, error)
	FindCouponByCode(ctx context.Context, code string) (domain.Coupon, error)
	ListCoupons(ctx context.Context, filter CouponFilter) ([]domain.Coupon, error)
	ListCouponRules(ctx context.Context, couponID string) ([]domain.CouponRule, error)
	CountRedemptions(ctx context.Context, couponID string) (int64, error)
	CountUserRedemptions(ctx context.Context, couponID string, userID string) (int64, error)
	InsertRedemption(ctx context.Context, redemption domain.CouponRedemption, coupon domain.Coupon) error
	GetRedemption(ctx context.Context, couponID string, orderID string) (domain.CouponRedemption, error)
}

type CouponUsecase struct {
	authorizer    *Authorizer
	repo          CouponRepository
	campaignRepo  CampaignRepository
	auditRecorder AuditRecorder
	logger        *slog.Logger
	now           func() time.Time
	newID         func(prefix string) string
}

type CouponFilter struct {
	SellerID string
	Status   domain.CouponStatus
	Code     string
	Limit    int
	Offset   int
}

type CreateCouponInput struct {
	Actor             domain.ActorContext
	Code              string
	DiscountType      domain.DiscountType
	DiscountValue     int64
	MaxDiscountAmount *int64
	MinCartAmount     int64
	Currency          string
	UsageLimit        *int64
	PerUserLimit      *int64
	Status            domain.CouponStatus
	StartsAt          *time.Time
	EndsAt            *time.Time
	Rules             []domain.CouponRule
	RequestID         string
}

type UpdateCouponInput struct {
	Actor             domain.ActorContext
	CouponID          string
	Code              *string
	DiscountType      *domain.DiscountType
	DiscountValue     *int64
	MaxDiscountAmount *int64
	MinCartAmount     *int64
	Currency          *string
	UsageLimit        *int64
	PerUserLimit      *int64
	Status            *domain.CouponStatus
	StartsAt          *time.Time
	EndsAt            *time.Time
	Rules             *[]domain.CouponRule
	RequestID         string
}

type DisableCouponInput struct {
	Actor     domain.ActorContext
	CouponID  string
	RequestID string
}

type ListCouponsInput struct {
	Actor     domain.ActorContext
	Status    domain.CouponStatus
	Code      string
	Limit     int
	Offset    int
	RequestID string
}

type ValidateCouponInput struct {
	CouponCode     string
	CampaignID     string
	UserID         string
	CartID         string
	OrderID        string
	Currency       string
	SubtotalAmount int64
	Items          []domain.CouponCartItem
	RequestID      string
}

type RecordCouponRedemptionInput struct {
	CouponID       string
	CampaignID     string
	OrderID        string
	UserID         string
	DiscountAmount int64
	Currency       string
	RequestID      string
}

func (uc *CouponUsecase) SetCampaignRepository(repo CampaignRepository) {
	uc.campaignRepo = repo
}

func NewCouponUsecase(authorizer *Authorizer, repo CouponRepository, auditRecorder AuditRecorder, logger *slog.Logger) (*CouponUsecase, error) {
	if authorizer == nil {
		return nil, domain.ErrInvalidAuthorizationIn
	}
	if repo == nil {
		return nil, domain.ErrCouponRepositoryRequired
	}
	if auditRecorder == nil {
		return nil, domain.ErrAuditRecorderRequired
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &CouponUsecase{
		authorizer:    authorizer,
		repo:          repo,
		auditRecorder: auditRecorder,
		logger:        logger,
		now:           func() time.Time { return time.Now().UTC() },
		newID:         newModerationID,
	}, nil
}

func (uc *CouponUsecase) CreateCoupon(ctx context.Context, input CreateCouponInput) (domain.Coupon, error) {
	input = input.normalized()
	sellerID := input.Actor.SellerID
	if sellerID == "" {
		return domain.Coupon{}, domain.ErrUnauthenticated
	}
	if err := uc.authorizeCouponAction(ctx, input.Actor, sellerID, "", domain.PermissionCouponsCreate, input.requestID()); err != nil {
		return domain.Coupon{}, err
	}

	now := uc.now()
	coupon := domain.Coupon{
		CouponID:          uc.newID("coupon"),
		SellerID:          stringPtr(sellerID),
		Code:              input.Code,
		DiscountType:      input.DiscountType,
		DiscountValue:     input.DiscountValue,
		MaxDiscountAmount: input.MaxDiscountAmount,
		MinCartAmount:     input.MinCartAmount,
		Currency:          input.Currency,
		UsageLimit:        input.UsageLimit,
		PerUserLimit:      input.PerUserLimit,
		Status:            input.Status,
		StartsAt:          input.StartsAt,
		EndsAt:            input.EndsAt,
		CreatedBy:         input.Actor.UserID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if coupon.Status == "" {
		coupon.Status = domain.CouponStatusDraft
	}
	rules := uc.prepareRules(coupon.CouponID, input.Rules)
	if err := validateSellerOwnedRules(coupon.SellerID, rules); err != nil {
		return domain.Coupon{}, err
	}
	if err := domain.ValidateCouponDefinition(coupon, rules); err != nil {
		return domain.Coupon{}, err
	}

	created, err := uc.repo.CreateCoupon(ctx, coupon.Normalized(), rules)
	if err != nil {
		return domain.Coupon{}, err
	}
	created.Rules, err = uc.repo.ListCouponRules(ctx, created.CouponID)
	if err != nil {
		return domain.Coupon{}, err
	}

	if err := uc.recordCouponAudit(ctx, domain.AuditActionCouponCreated, input.Actor, sellerID, created.CouponID, input.requestID(), nil, couponAuditSnapshot(created)); err != nil {
		return domain.Coupon{}, err
	}

	uc.logger.InfoContext(ctx, "cms.coupon.created",
		slog.String("coupon_id", created.CouponID),
		slog.String("seller_id", sellerID),
		slog.String("code", created.Code),
		slog.String("request_id", input.requestID()),
	)
	return created, nil
}

func (uc *CouponUsecase) UpdateCoupon(ctx context.Context, input UpdateCouponInput) (domain.Coupon, error) {
	input = input.normalized()
	if input.CouponID == "" {
		return domain.Coupon{}, domain.NewValidationError("Invalid coupon update request.", []domain.FieldViolation{{Field: "coupon_id", Reason: "Coupon id is required."}})
	}

	existing, err := uc.repo.GetCouponByID(ctx, input.CouponID)
	if err != nil {
		return domain.Coupon{}, err
	}
	existing.Rules, err = uc.repo.ListCouponRules(ctx, existing.CouponID)
	if err != nil {
		return domain.Coupon{}, err
	}
	sellerID := couponSellerID(existing)
	if sellerID == "" || input.Actor.SellerID != sellerID {
		return domain.Coupon{}, domain.ErrCouponOwnershipMismatch
	}
	if err := uc.authorizeCouponAction(ctx, input.Actor, sellerID, existing.CouponID, domain.PermissionCouponsUpdate, input.requestID()); err != nil {
		return domain.Coupon{}, err
	}

	before := couponAuditSnapshot(existing)
	updated := existing
	if input.Code != nil {
		updated.Code = *input.Code
	}
	if input.DiscountType != nil {
		updated.DiscountType = *input.DiscountType
	}
	if input.DiscountValue != nil {
		updated.DiscountValue = *input.DiscountValue
	}
	if input.MaxDiscountAmount != nil {
		updated.MaxDiscountAmount = input.MaxDiscountAmount
	}
	if input.MinCartAmount != nil {
		updated.MinCartAmount = *input.MinCartAmount
	}
	if input.Currency != nil {
		updated.Currency = *input.Currency
	}
	if input.UsageLimit != nil {
		updated.UsageLimit = input.UsageLimit
	}
	if input.PerUserLimit != nil {
		updated.PerUserLimit = input.PerUserLimit
	}
	if input.Status != nil {
		updated.Status = *input.Status
	}
	if input.StartsAt != nil {
		updated.StartsAt = input.StartsAt
	}
	if input.EndsAt != nil {
		updated.EndsAt = input.EndsAt
	}
	updated.UpdatedAt = uc.now()

	rules := existing.Rules
	if input.Rules != nil {
		rules = uc.prepareRules(updated.CouponID, *input.Rules)
	}
	if err := validateSellerOwnedRules(updated.SellerID, rules); err != nil {
		return domain.Coupon{}, err
	}
	if err := domain.ValidateCouponDefinition(updated, rules); err != nil {
		return domain.Coupon{}, err
	}

	saved, err := uc.repo.UpdateCoupon(ctx, updated.Normalized(), rules)
	if err != nil {
		return domain.Coupon{}, err
	}
	saved.Rules, err = uc.repo.ListCouponRules(ctx, saved.CouponID)
	if err != nil {
		return domain.Coupon{}, err
	}
	if err := uc.recordCouponAudit(ctx, domain.AuditActionCouponUpdated, input.Actor, sellerID, saved.CouponID, input.requestID(), before, couponAuditSnapshot(saved)); err != nil {
		return domain.Coupon{}, err
	}

	uc.logger.InfoContext(ctx, "cms.coupon.updated",
		slog.String("coupon_id", saved.CouponID),
		slog.String("seller_id", sellerID),
		slog.String("code", saved.Code),
		slog.String("request_id", input.requestID()),
	)
	return saved, nil
}

func (uc *CouponUsecase) DisableCoupon(ctx context.Context, input DisableCouponInput) (domain.Coupon, error) {
	input = input.normalized()
	if input.CouponID == "" {
		return domain.Coupon{}, domain.NewValidationError("Invalid coupon disable request.", []domain.FieldViolation{{Field: "coupon_id", Reason: "Coupon id is required."}})
	}

	existing, err := uc.repo.GetCouponByID(ctx, input.CouponID)
	if err != nil {
		return domain.Coupon{}, err
	}
	existing.Rules, err = uc.repo.ListCouponRules(ctx, existing.CouponID)
	if err != nil {
		return domain.Coupon{}, err
	}
	sellerID := couponSellerID(existing)
	if sellerID == "" || input.Actor.SellerID != sellerID {
		return domain.Coupon{}, domain.ErrCouponOwnershipMismatch
	}
	if err := uc.authorizeCouponAction(ctx, input.Actor, sellerID, existing.CouponID, domain.PermissionCouponsDisable, input.requestID()); err != nil {
		return domain.Coupon{}, err
	}

	if existing.Status.Normalized() == domain.CouponStatusPaused {
		return existing, nil
	}

	before := couponAuditSnapshot(existing)
	disabled, err := uc.repo.UpdateCouponStatus(ctx, existing.CouponID, domain.CouponStatusPaused, uc.now())
	if err != nil {
		return domain.Coupon{}, err
	}
	disabled.Rules, err = uc.repo.ListCouponRules(ctx, disabled.CouponID)
	if err != nil {
		return domain.Coupon{}, err
	}
	if err := uc.recordCouponAudit(ctx, domain.AuditActionCouponDisabled, input.Actor, sellerID, disabled.CouponID, input.requestID(), before, couponAuditSnapshot(disabled)); err != nil {
		return domain.Coupon{}, err
	}

	uc.logger.InfoContext(ctx, "cms.coupon.disabled",
		slog.String("coupon_id", disabled.CouponID),
		slog.String("seller_id", sellerID),
		slog.String("request_id", input.requestID()),
	)
	return disabled, nil
}

func (uc *CouponUsecase) ListCoupons(ctx context.Context, input ListCouponsInput) ([]domain.Coupon, error) {
	input = input.normalized()
	sellerID := input.Actor.SellerID
	if sellerID == "" {
		return nil, domain.ErrUnauthenticated
	}
	if err := uc.authorizeCouponAction(ctx, input.Actor, sellerID, "", domain.PermissionCouponsRead, input.requestID()); err != nil {
		return nil, err
	}

	filter := CouponFilter{
		SellerID: sellerID,
		Status:   input.Status,
		Code:     input.Code,
		Limit:    normalizedLimit(input.Limit),
		Offset:   normalizedOffset(input.Offset),
	}
	coupons, err := uc.repo.ListCoupons(ctx, filter)
	if err != nil {
		return nil, err
	}
	for i := range coupons {
		coupons[i].Rules, err = uc.repo.ListCouponRules(ctx, coupons[i].CouponID)
		if err != nil {
			return nil, err
		}
	}
	return coupons, nil
}

func (uc *CouponUsecase) ValidateCoupon(ctx context.Context, input ValidateCouponInput) (domain.CouponValidationResult, error) {
	input = input.normalized()
	if input.CouponCode == "" {
		return domain.CouponValidationResult{}, domain.NewValidationError("Invalid coupon validation request.", []domain.FieldViolation{{Field: "coupon_code", Reason: "Coupon code is required."}})
	}
	if input.UserID == "" {
		return domain.CouponValidationResult{}, domain.NewValidationError("Invalid coupon validation request.", []domain.FieldViolation{{Field: "user_id", Reason: "User id is required."}})
	}
	if input.SubtotalAmount <= 0 {
		return domain.CouponValidationResult{}, domain.NewValidationError("Invalid coupon validation request.", []domain.FieldViolation{{Field: "subtotal_amount", Reason: "Subtotal amount must be greater than zero."}})
	}

	coupon, err := uc.repo.FindCouponByCode(ctx, input.CouponCode)
	if err != nil {
		if errors.Is(err, domain.ErrCouponNotFound) {
			return domain.InvalidCouponResult(input.Currency, "", domain.CouponInvalidReasonNotFound), nil
		}
		return domain.CouponValidationResult{}, err
	}
	rules, err := uc.repo.ListCouponRules(ctx, coupon.CouponID)
	if err != nil {
		return domain.CouponValidationResult{}, err
	}

	if reason := domain.ValidateCouponEligibility(coupon, input.Currency, input.SubtotalAmount, uc.now()); reason != "" {
		return domain.InvalidCouponResult(input.Currency, coupon.CouponID, reason), nil
	}

	var campaign domain.Campaign
	var campaignRemaining int64
	var campaignUnlimited bool
	if input.CampaignID != "" {
		var reason domain.CouponInvalidReason
		campaign, campaignRemaining, campaignUnlimited, reason, err = uc.couponCampaignGuard(ctx, input.CampaignID, coupon, input.UserID, input.Currency, 0)
		if err != nil {
			return domain.CouponValidationResult{}, err
		}
		if reason != "" {
			result := domain.InvalidCouponResult(input.Currency, coupon.CouponID, reason)
			result.CampaignID = input.CampaignID
			return result, nil
		}
	}

	eligibleSubtotal := domain.CalculateEligibleSubtotal(input.SubtotalAmount, input.Items, coupon, rules)
	if eligibleSubtotal <= 0 {
		return domain.InvalidCouponResult(input.Currency, coupon.CouponID, domain.CouponInvalidReasonScopeNotMatched), nil
	}

	if reason, err := uc.usageLimitReason(ctx, coupon, input.UserID); err != nil {
		return domain.CouponValidationResult{}, err
	} else if reason != "" {
		return domain.InvalidCouponResult(input.Currency, coupon.CouponID, reason), nil
	}

	discount := domain.CalculateCouponDiscount(coupon, eligibleSubtotal)
	if discount <= 0 {
		return domain.InvalidCouponResult(input.Currency, coupon.CouponID, domain.CouponInvalidReasonInvalidDiscountSetup), nil
	}
	if input.CampaignID != "" && !campaignUnlimited && discount > campaignRemaining {
		result := domain.InvalidCouponResult(input.Currency, coupon.CouponID, domain.CouponInvalidReasonCampaignBudgetExhausted)
		result.CampaignID = campaign.CampaignID
		return result, nil
	}

	result := domain.CouponValidationResult{
		Valid:            true,
		CouponID:         coupon.CouponID,
		CampaignID:       campaign.CampaignID,
		AppliedRuleCodes: appliedCouponRuleCodes(coupon, rules, campaign.CampaignID != ""),
		Discount: domain.Money{
			Amount:   discount,
			Currency: coupon.Normalized().Currency,
		},
		EligibleSubtotal: eligibleSubtotal,
	}
	uc.logger.InfoContext(ctx, "cms.coupon.validated",
		slog.String("coupon_id", result.CouponID),
		slog.String("campaign_id", result.CampaignID),
		slog.String("coupon_code", input.CouponCode),
		slog.Bool("valid", result.Valid),
		slog.Int64("discount_amount", result.Discount.Amount),
		slog.String("currency", result.Discount.Currency),
		slog.String("request_id", input.RequestID),
	)
	return result, nil
}

func (uc *CouponUsecase) RecordRedemption(ctx context.Context, input RecordCouponRedemptionInput) (domain.CouponRedemption, error) {
	input = input.normalized()
	if input.CouponID == "" || input.OrderID == "" || input.UserID == "" {
		return domain.CouponRedemption{}, domain.NewValidationError("Invalid coupon redemption request.", []domain.FieldViolation{{Field: "redemption", Reason: "Coupon id, order id, and user id are required."}})
	}
	if input.DiscountAmount <= 0 {
		return domain.CouponRedemption{}, domain.NewValidationError("Invalid coupon redemption request.", []domain.FieldViolation{{Field: "discount_amount", Reason: "Discount amount must be greater than zero."}})
	}

	coupon, err := uc.repo.GetCouponByID(ctx, input.CouponID)
	if err != nil {
		return domain.CouponRedemption{}, err
	}
	coupon = coupon.Normalized()
	if coupon.Currency != domain.NormalizeCurrency(input.Currency) {
		return domain.CouponRedemption{}, domain.ErrCouponCurrencyMismatch
	}
	if input.CampaignID != "" {
		if _, _, _, reason, err := uc.couponCampaignGuard(ctx, input.CampaignID, coupon, input.UserID, input.Currency, input.DiscountAmount); err != nil {
			return domain.CouponRedemption{}, err
		} else if reason != "" {
			return domain.CouponRedemption{}, campaignReasonError(reason)
		}
	}

	now := uc.now()
	redemption := domain.CouponRedemption{
		RedemptionID: uc.newID("redemption"),
		CouponID:     coupon.CouponID,
		CampaignID:   input.CampaignID,
		OrderID:      input.OrderID,
		UserID:       input.UserID,
		Discount: domain.Money{
			Amount:   input.DiscountAmount,
			Currency: domain.NormalizeCurrency(input.Currency),
		},
		RequestID:  input.RequestID,
		RedeemedAt: now,
		CreatedAt:  now,
	}

	err = uc.repo.InsertRedemption(ctx, redemption, coupon)
	if err != nil {
		if errors.Is(err, domain.ErrCouponRedemptionDuplicate) {
			existing, findErr := uc.repo.GetRedemption(ctx, redemption.CouponID, redemption.OrderID)
			if findErr != nil {
				return domain.CouponRedemption{}, findErr
			}
			if sameRedemption(existing, redemption) {
				return existing, nil
			}
			return domain.CouponRedemption{}, domain.ErrCouponRedemptionConflict
		}
		return domain.CouponRedemption{}, err
	}

	if sellerID := couponSellerID(coupon); sellerID != "" {
		systemActor := domain.ActorContext{
			UserID:    firstNonEmptyString(input.UserID, "system"),
			SellerID:  sellerID,
			RequestID: input.RequestID,
		}
		if auditErr := uc.recordCouponAudit(ctx, domain.AuditActionCouponRedeemed, systemActor, sellerID, coupon.CouponID, input.RequestID, nil, map[string]any{
			"coupon_id":       coupon.CouponID,
			"campaign_id":     redemption.CampaignID,
			"order_id":        redemption.OrderID,
			"user_id":         redemption.UserID,
			"discount_amount": redemption.Discount.Amount,
			"currency":        redemption.Discount.Currency,
		}); auditErr != nil {
			return domain.CouponRedemption{}, auditErr
		}
	}

	uc.logger.InfoContext(ctx, "cms.coupon.redemption_recorded",
		slog.String("coupon_id", redemption.CouponID),
		slog.String("campaign_id", redemption.CampaignID),
		slog.String("order_id", redemption.OrderID),
		slog.String("user_id", redemption.UserID),
		slog.Int64("discount_amount", redemption.Discount.Amount),
		slog.String("currency", redemption.Discount.Currency),
		slog.String("request_id", input.RequestID),
	)
	return redemption, nil
}

func (uc *CouponUsecase) usageLimitReason(ctx context.Context, coupon domain.Coupon, userID string) (domain.CouponInvalidReason, error) {
	if coupon.UsageLimit != nil {
		used, err := uc.repo.CountRedemptions(ctx, coupon.CouponID)
		if err != nil {
			return "", err
		}
		if used >= *coupon.UsageLimit {
			return domain.CouponInvalidReasonUsageLimitReached, nil
		}
	}
	if coupon.PerUserLimit != nil {
		used, err := uc.repo.CountUserRedemptions(ctx, coupon.CouponID, userID)
		if err != nil {
			return "", err
		}
		if used >= *coupon.PerUserLimit {
			return domain.CouponInvalidReasonPerUserLimitReached, nil
		}
	}
	return "", nil
}

func (uc *CouponUsecase) couponCampaignGuard(ctx context.Context, campaignID string, coupon domain.Coupon, userID string, currency string, discountAmount int64) (domain.Campaign, int64, bool, domain.CouponInvalidReason, error) {
	if uc.campaignRepo == nil {
		return domain.Campaign{}, 0, false, "", domain.ErrCampaignRepositoryRequired
	}
	campaign, err := uc.campaignRepo.GetCampaignByID(ctx, campaignID)
	if err != nil {
		if errors.Is(err, domain.ErrCampaignNotFound) {
			return domain.Campaign{}, 0, false, domain.CouponInvalidReasonCampaignNotFound, nil
		}
		return domain.Campaign{}, 0, false, "", err
	}
	coupon = coupon.Normalized()
	sellerID := couponSellerID(coupon)
	if reason := domain.ValidateCampaignEligibility(campaign, sellerID, coupon.CouponID, currency, uc.now()); reason != "" {
		return campaign, 0, false, couponReasonFromCampaignReason(reason), nil
	}
	campaign = campaign.Normalized()
	if campaign.Metadata.UsageLimit != nil {
		used, err := uc.campaignRepo.CountCampaignRedemptions(ctx, campaign.CampaignID)
		if err != nil {
			return domain.Campaign{}, 0, false, "", err
		}
		if used >= *campaign.Metadata.UsageLimit {
			return campaign, 0, false, domain.CouponInvalidReasonCampaignUsageLimitReached, nil
		}
	}
	if campaign.Metadata.PerUserLimit != nil && strings.TrimSpace(userID) != "" {
		used, err := uc.campaignRepo.CountCampaignUserRedemptions(ctx, campaign.CampaignID, userID)
		if err != nil {
			return domain.Campaign{}, 0, false, "", err
		}
		if used >= *campaign.Metadata.PerUserLimit {
			return campaign, 0, false, domain.CouponInvalidReasonCampaignUsageLimitReached, nil
		}
	}
	if campaign.BudgetAmount == nil {
		return campaign, 0, true, "", nil
	}
	spent, err := uc.campaignRepo.SumCampaignDiscounts(ctx, campaign.CampaignID)
	if err != nil {
		return domain.Campaign{}, 0, false, "", err
	}
	remaining := *campaign.BudgetAmount - spent
	if remaining <= 0 || (discountAmount > 0 && discountAmount > remaining) {
		return campaign, 0, false, domain.CouponInvalidReasonCampaignBudgetExhausted, nil
	}
	return campaign, remaining, false, "", nil
}

func (uc *CouponUsecase) prepareRules(couponID string, rules []domain.CouponRule) []domain.CouponRule {
	out := make([]domain.CouponRule, 0, len(rules))
	for _, rule := range rules {
		rule.CouponID = couponID
		if rule.RuleID == "" {
			rule.RuleID = uc.newID("rule")
		}
		if rule.CreatedAt.IsZero() {
			rule.CreatedAt = uc.now()
		}
		out = append(out, rule.Normalized())
	}
	return out
}

func (uc *CouponUsecase) authorizeCouponAction(ctx context.Context, actor domain.ActorContext, sellerID string, couponID string, permission domain.Permission, requestID string) error {
	_, err := uc.authorizer.Authorize(ctx, AuthorizeInput{
		Actor:              actor,
		ResourceSellerID:   sellerID,
		RequiredPermission: permission,
		ResourceType:       "coupon",
		ResourceID:         couponID,
		RequestID:          requestID,
	})
	return err
}

func (uc *CouponUsecase) recordCouponAudit(ctx context.Context, action domain.Permission, actor domain.ActorContext, sellerID string, couponID string, requestID string, before map[string]any, after map[string]any) error {
	actor = actor.Normalized()
	if actor.UserID == "" {
		actor.UserID = "system"
	}
	if actor.SellerID == "" {
		actor.SellerID = sellerID
	}
	if err := uc.auditRecorder.Record(ctx, domain.AuditEvent{
		AuditID:          uc.newID("audit"),
		ActorUserID:      actor.UserID,
		ActorSellerID:    actor.SellerID,
		ActorRoles:       actor.Roles,
		Action:           action,
		ResourceType:     "coupon",
		ResourceID:       couponID,
		ResourceSellerID: sellerID,
		RequestID:        requestID,
		Decision:         domain.AuditDecisionAllowed,
		Before:           before,
		After:            after,
		CreatedAt:        uc.now(),
	}); err != nil {
		uc.logger.ErrorContext(ctx, "cms.coupon.audit_failed",
			slog.String("action", action.String()),
			slog.String("coupon_id", couponID),
			slog.String("seller_id", sellerID),
			slog.String("request_id", requestID),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("%w: %v", domain.ErrAuditWriteFailed, err)
	}
	return nil
}

func validateSellerOwnedRules(sellerID *string, rules []domain.CouponRule) error {
	if sellerID == nil || strings.TrimSpace(*sellerID) == "" {
		return nil
	}
	ownerSellerID := strings.TrimSpace(*sellerID)
	for _, rule := range rules {
		if rule.Type.Normalized() != domain.CouponRuleTypeSellerScope {
			continue
		}
		for _, id := range rule.SellerIDs {
			if strings.TrimSpace(id) != ownerSellerID {
				return domain.NewValidationError("Invalid coupon seller scope.", []domain.FieldViolation{{Field: "rules.seller_ids", Reason: "Seller-owned coupons can only target the owning seller."}})
			}
		}
	}
	return nil
}

func couponAuditSnapshot(coupon domain.Coupon) map[string]any {
	coupon = coupon.Normalized()
	sellerID := ""
	if coupon.SellerID != nil {
		sellerID = *coupon.SellerID
	}
	return map[string]any{
		"coupon_id":           coupon.CouponID,
		"seller_id":           sellerID,
		"code":                coupon.Code,
		"discount_type":       coupon.DiscountType,
		"discount_value":      coupon.DiscountValue,
		"max_discount_amount": coupon.MaxDiscountAmount,
		"min_cart_amount":     coupon.MinCartAmount,
		"currency":            coupon.Currency,
		"usage_limit":         coupon.UsageLimit,
		"per_user_limit":      coupon.PerUserLimit,
		"status":              coupon.Status,
		"starts_at":           coupon.StartsAt,
		"ends_at":             coupon.EndsAt,
		"rules_count":         len(coupon.Rules),
	}
}

func sameRedemption(existing domain.CouponRedemption, incoming domain.CouponRedemption) bool {
	return existing.CouponID == incoming.CouponID &&
		existing.CampaignID == incoming.CampaignID &&
		existing.OrderID == incoming.OrderID &&
		existing.UserID == incoming.UserID &&
		existing.Discount.Amount == incoming.Discount.Amount &&
		existing.Discount.Currency == incoming.Discount.Currency
}

func appliedCouponRuleCodes(coupon domain.Coupon, rules []domain.CouponRule, campaignLinked bool) []string {
	coupon = coupon.Normalized()
	rules = domain.NormalizeCouponRules(rules)
	seen := make(map[string]struct{}, len(rules)+5)
	codes := make([]string, 0, len(rules)+5)
	add := func(code string) {
		code = strings.TrimSpace(code)
		if code == "" {
			return
		}
		if _, exists := seen[code]; exists {
			return
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}
	if coupon.MinCartAmount > 0 {
		add("min_cart_amount")
	}
	if coupon.UsageLimit != nil {
		add("global_usage_limit")
	}
	if coupon.PerUserLimit != nil {
		add("user_usage_limit")
	}
	for _, rule := range rules {
		add(string(rule.Type.Normalized()))
	}
	switch coupon.DiscountType.Normalized() {
	case domain.DiscountTypeFixed:
		add("fixed_discount")
	case domain.DiscountTypePercentage:
		add("percentage_discount")
	}
	if campaignLinked {
		add("campaign_scope")
	}
	return codes
}

func couponSellerID(coupon domain.Coupon) string {
	if coupon.SellerID == nil {
		return ""
	}
	return strings.TrimSpace(*coupon.SellerID)
}

func normalizedLimit(limit int) int {
	if limit <= 0 {
		return defaultCouponListLimit
	}
	if limit > maxCouponListLimit {
		return maxCouponListLimit
	}
	return limit
}

func normalizedOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}

func stringPtr(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func couponReasonFromCampaignReason(reason domain.CampaignInvalidReason) domain.CouponInvalidReason {
	return domain.CouponInvalidReason(reason)
}

func campaignReasonError(reason domain.CouponInvalidReason) error {
	switch reason {
	case domain.CouponInvalidReasonCampaignBudgetExhausted:
		return domain.ErrCampaignBudgetExhausted
	case domain.CouponInvalidReasonCampaignUsageLimitReached:
		return domain.ErrCampaignUsageLimitReached
	case domain.CouponInvalidReasonCouponNotInCampaign:
		return domain.ErrCampaignCouponNotLinked
	case domain.CouponInvalidReasonCampaignSellerMismatch:
		return domain.ErrCampaignOwnershipMismatch
	case domain.CouponInvalidReasonCampaignNotActive,
		domain.CouponInvalidReasonCampaignNotStarted,
		domain.CouponInvalidReasonCampaignExpired:
		return domain.ErrInvalidCampaignStatusTransition
	case domain.CouponInvalidReasonCampaignNotFound:
		return domain.ErrCampaignNotFound
	default:
		return domain.ErrValidationFailed
	}
}

func (i CreateCouponInput) normalized() CreateCouponInput {
	i.Actor = i.Actor.Normalized()
	i.Code = domain.NormalizeCouponCode(i.Code)
	i.DiscountType = i.DiscountType.Normalized()
	i.Currency = domain.NormalizeCurrency(i.Currency)
	i.Status = i.Status.Normalized()
	i.RequestID = strings.TrimSpace(i.RequestID)
	return i
}

func (i CreateCouponInput) requestID() string {
	if i.RequestID != "" {
		return i.RequestID
	}
	return i.Actor.RequestID
}

func (i UpdateCouponInput) normalized() UpdateCouponInput {
	i.Actor = i.Actor.Normalized()
	i.CouponID = strings.TrimSpace(i.CouponID)
	if i.Code != nil {
		code := domain.NormalizeCouponCode(*i.Code)
		i.Code = &code
	}
	if i.DiscountType != nil {
		discountType := i.DiscountType.Normalized()
		i.DiscountType = &discountType
	}
	if i.Currency != nil {
		currency := domain.NormalizeCurrency(*i.Currency)
		i.Currency = &currency
	}
	if i.Status != nil {
		status := i.Status.Normalized()
		i.Status = &status
	}
	i.RequestID = strings.TrimSpace(i.RequestID)
	return i
}

func (i UpdateCouponInput) requestID() string {
	if i.RequestID != "" {
		return i.RequestID
	}
	return i.Actor.RequestID
}

func (i DisableCouponInput) normalized() DisableCouponInput {
	i.Actor = i.Actor.Normalized()
	i.CouponID = strings.TrimSpace(i.CouponID)
	i.RequestID = strings.TrimSpace(i.RequestID)
	return i
}

func (i DisableCouponInput) requestID() string {
	if i.RequestID != "" {
		return i.RequestID
	}
	return i.Actor.RequestID
}

func (i ListCouponsInput) normalized() ListCouponsInput {
	i.Actor = i.Actor.Normalized()
	i.Status = i.Status.Normalized()
	i.Code = domain.NormalizeCouponCode(i.Code)
	i.RequestID = strings.TrimSpace(i.RequestID)
	return i
}

func (i ListCouponsInput) requestID() string {
	if i.RequestID != "" {
		return i.RequestID
	}
	return i.Actor.RequestID
}

func (i ValidateCouponInput) normalized() ValidateCouponInput {
	i.CouponCode = domain.NormalizeCouponCode(i.CouponCode)
	i.CampaignID = strings.TrimSpace(i.CampaignID)
	i.UserID = strings.TrimSpace(i.UserID)
	i.CartID = strings.TrimSpace(i.CartID)
	i.OrderID = strings.TrimSpace(i.OrderID)
	i.Currency = domain.NormalizeCurrency(i.Currency)
	i.RequestID = strings.TrimSpace(i.RequestID)
	for idx := range i.Items {
		i.Items[idx] = i.Items[idx].Normalized()
	}
	return i
}

func (i RecordCouponRedemptionInput) normalized() RecordCouponRedemptionInput {
	i.CouponID = strings.TrimSpace(i.CouponID)
	i.CampaignID = strings.TrimSpace(i.CampaignID)
	i.OrderID = strings.TrimSpace(i.OrderID)
	i.UserID = strings.TrimSpace(i.UserID)
	i.Currency = domain.NormalizeCurrency(i.Currency)
	i.RequestID = strings.TrimSpace(i.RequestID)
	return i
}
