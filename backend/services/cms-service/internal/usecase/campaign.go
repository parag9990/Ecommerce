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

const defaultCampaignMaxDuration = 90 * 24 * time.Hour

type CampaignRepository interface {
	CreateCampaign(ctx context.Context, campaign domain.Campaign) (domain.Campaign, error)
	UpdateCampaign(ctx context.Context, campaign domain.Campaign) (domain.Campaign, error)
	UpdateCampaignStatus(ctx context.Context, campaignID string, sellerID string, status domain.CampaignStatus, updatedAt time.Time) (domain.Campaign, error)
	GetCampaignByID(ctx context.Context, campaignID string) (domain.Campaign, error)
	ListCampaigns(ctx context.Context, filter CampaignFilter) ([]domain.Campaign, error)
	CountCampaignRedemptions(ctx context.Context, campaignID string) (int64, error)
	CountCampaignUserRedemptions(ctx context.Context, campaignID string, userID string) (int64, error)
	SumCampaignDiscounts(ctx context.Context, campaignID string) (int64, error)
}

type CampaignCouponReader interface {
	GetCouponByID(ctx context.Context, couponID string) (domain.Coupon, error)
}

type CampaignUsecase struct {
	authorizer    *Authorizer
	repo          CampaignRepository
	couponReader  CampaignCouponReader
	auditRecorder AuditRecorder
	logger        *slog.Logger
	now           func() time.Time
	newID         func(prefix string) string
	maxDuration   time.Duration
}

type CampaignOptions struct {
	MaxDuration time.Duration
}

type CampaignFilter struct {
	SellerID    string
	Status      domain.CampaignStatus
	StartsAfter *time.Time
	EndsBefore  *time.Time
	Query       string
	Limit       int
	Offset      int
}

type CreateCampaignInput struct {
	Actor        domain.ActorContext
	Name         string
	BudgetAmount *int64
	Currency     string
	StartsAt     time.Time
	EndsAt       time.Time
	Metadata     domain.CampaignMetadata
	RequestID    string
}

type UpdateCampaignInput struct {
	Actor        domain.ActorContext
	CampaignID   string
	Name         *string
	BudgetAmount *int64
	Currency     *string
	StartsAt     *time.Time
	EndsAt       *time.Time
	Metadata     *domain.CampaignMetadata
	Status       *domain.CampaignStatus
	RequestID    string
}

type DisableCampaignInput struct {
	Actor      domain.ActorContext
	CampaignID string
	RequestID  string
}

type ListCampaignsInput struct {
	Actor       domain.ActorContext
	Status      domain.CampaignStatus
	StartsAfter *time.Time
	EndsBefore  *time.Time
	Query       string
	Limit       int
	Offset      int
	RequestID   string
}

type GetCampaignInput struct {
	Actor          domain.ActorContext
	SellerID       string
	CampaignID     string
	InternalCaller string
	RequestID      string
}

type CampaignRead struct {
	Campaign domain.Campaign
	Spent    domain.Money
}

type ValidateCampaignInput struct {
	CampaignID     string
	SellerID       string
	CouponID       string
	UserID         string
	Currency       string
	DiscountAmount int64
	RequestID      string
}

func NewCampaignUsecase(authorizer *Authorizer, repo CampaignRepository, couponReader CampaignCouponReader, auditRecorder AuditRecorder, logger *slog.Logger, opts CampaignOptions) (*CampaignUsecase, error) {
	if authorizer == nil {
		return nil, domain.ErrInvalidAuthorizationIn
	}
	if repo == nil {
		return nil, domain.ErrCampaignRepositoryRequired
	}
	if couponReader == nil {
		return nil, domain.ErrCampaignCouponReaderRequired
	}
	if auditRecorder == nil {
		return nil, domain.ErrAuditRecorderRequired
	}
	if logger == nil {
		logger = slog.Default()
	}
	maxDuration := opts.MaxDuration
	if maxDuration <= 0 {
		maxDuration = defaultCampaignMaxDuration
	}
	return &CampaignUsecase{
		authorizer:    authorizer,
		repo:          repo,
		couponReader:  couponReader,
		auditRecorder: auditRecorder,
		logger:        logger,
		now:           func() time.Time { return time.Now().UTC() },
		newID:         newModerationID,
		maxDuration:   maxDuration,
	}, nil
}

func (uc *CampaignUsecase) CreateCampaign(ctx context.Context, input CreateCampaignInput) (domain.Campaign, error) {
	input = input.normalized()
	sellerID := input.Actor.SellerID
	if sellerID == "" {
		return domain.Campaign{}, domain.ErrUnauthenticated
	}
	if err := uc.authorizeCampaignAction(ctx, input.Actor, sellerID, "", domain.PermissionCampaignsCreate, input.requestID()); err != nil {
		return domain.Campaign{}, err
	}

	now := uc.now()
	campaign := domain.Campaign{
		CampaignID:   uc.newID("camp"),
		SellerID:     stringPtr(sellerID),
		Name:         input.Name,
		Status:       domain.CampaignStatusDraft,
		BudgetAmount: input.BudgetAmount,
		Currency:     input.Currency,
		StartsAt:     input.StartsAt,
		EndsAt:       input.EndsAt,
		Metadata:     input.Metadata,
		CreatedBy:    input.Actor.UserID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := uc.validateCampaign(ctx, campaign); err != nil {
		return domain.Campaign{}, err
	}

	created, err := uc.repo.CreateCampaign(ctx, campaign.Normalized())
	if err != nil {
		return domain.Campaign{}, err
	}
	if err := uc.recordCampaignAudit(ctx, domain.AuditActionCampaignCreated, input.Actor, sellerID, created.CampaignID, input.requestID(), nil, campaignAuditSnapshot(created)); err != nil {
		return domain.Campaign{}, err
	}

	uc.logger.InfoContext(ctx, "cms.campaign.created",
		slog.String("campaign_id", created.CampaignID),
		slog.String("seller_id", sellerID),
		slog.String("status", string(created.Status)),
		slog.String("request_id", input.requestID()),
	)
	return created, nil
}

func (uc *CampaignUsecase) UpdateCampaign(ctx context.Context, input UpdateCampaignInput) (domain.Campaign, error) {
	input = input.normalized()
	if input.CampaignID == "" {
		return domain.Campaign{}, domain.NewValidationError("Invalid campaign update request.", []domain.FieldViolation{{Field: "campaign_id", Reason: "Campaign id is required."}})
	}

	existing, err := uc.repo.GetCampaignByID(ctx, input.CampaignID)
	if err != nil {
		return domain.Campaign{}, err
	}
	existing = existing.Normalized()
	sellerID := campaignSellerID(existing)
	if sellerID == "" || input.Actor.SellerID != sellerID {
		return domain.Campaign{}, domain.ErrCampaignOwnershipMismatch
	}

	if err := uc.authorizeCampaignAction(ctx, input.Actor, sellerID, existing.CampaignID, domain.PermissionCampaignsUpdate, input.requestID()); err != nil {
		return domain.Campaign{}, err
	}
	if existing.Status == domain.CampaignStatusCompleted {
		return domain.Campaign{}, domain.ErrCampaignCompletedReadOnly
	}

	before := campaignAuditSnapshot(existing)
	updated := existing
	if input.Name != nil {
		updated.Name = *input.Name
	}
	if input.BudgetAmount != nil {
		updated.BudgetAmount = input.BudgetAmount
	}
	if input.Currency != nil {
		updated.Currency = *input.Currency
	}
	if input.StartsAt != nil {
		updated.StartsAt = input.StartsAt.UTC()
	}
	if input.EndsAt != nil {
		updated.EndsAt = input.EndsAt.UTC()
	}
	if input.Metadata != nil {
		updated.Metadata = *input.Metadata
	}
	if input.Status != nil {
		if err := domain.ValidateCampaignStatusTransition(existing.Status, *input.Status); err != nil {
			return domain.Campaign{}, err
		}
		updated.Status = input.Status.Normalized()
	}
	updated.UpdatedAt = uc.now()

	if err := enforceCampaignEditRules(existing, updated, uc.now()); err != nil {
		return domain.Campaign{}, err
	}
	if err := uc.validateCampaign(ctx, updated); err != nil {
		return domain.Campaign{}, err
	}

	saved, err := uc.repo.UpdateCampaign(ctx, updated.Normalized())
	if err != nil {
		return domain.Campaign{}, err
	}
	action := campaignAuditAction(existing.Status, saved.Status)
	if err := uc.recordCampaignAudit(ctx, action, input.Actor, sellerID, saved.CampaignID, input.requestID(), before, campaignAuditSnapshot(saved)); err != nil {
		return domain.Campaign{}, err
	}

	uc.logger.InfoContext(ctx, "cms.campaign.updated",
		slog.String("campaign_id", saved.CampaignID),
		slog.String("seller_id", sellerID),
		slog.String("status", string(saved.Status)),
		slog.String("request_id", input.requestID()),
	)
	return saved, nil
}

func (uc *CampaignUsecase) DisableCampaign(ctx context.Context, input DisableCampaignInput) (domain.Campaign, error) {
	input = input.normalized()
	if input.CampaignID == "" {
		return domain.Campaign{}, domain.NewValidationError("Invalid campaign disable request.", []domain.FieldViolation{{Field: "campaign_id", Reason: "Campaign id is required."}})
	}

	existing, err := uc.repo.GetCampaignByID(ctx, input.CampaignID)
	if err != nil {
		return domain.Campaign{}, err
	}
	existing = existing.Normalized()
	sellerID := campaignSellerID(existing)
	if sellerID == "" || input.Actor.SellerID != sellerID {
		return domain.Campaign{}, domain.ErrCampaignOwnershipMismatch
	}
	if err := uc.authorizeCampaignAction(ctx, input.Actor, sellerID, existing.CampaignID, domain.PermissionCampaignsDisable, input.requestID()); err != nil {
		return domain.Campaign{}, err
	}
	if existing.Status == domain.CampaignStatusCompleted {
		return domain.Campaign{}, domain.ErrCampaignCompletedReadOnly
	}
	if existing.Status == domain.CampaignStatusPaused {
		return existing, nil
	}

	before := campaignAuditSnapshot(existing)
	disabled, err := uc.repo.UpdateCampaignStatus(ctx, existing.CampaignID, sellerID, domain.CampaignStatusPaused, uc.now())
	if err != nil {
		return domain.Campaign{}, err
	}
	if err := uc.recordCampaignAudit(ctx, domain.AuditActionCampaignDisabled, input.Actor, sellerID, disabled.CampaignID, input.requestID(), before, campaignAuditSnapshot(disabled)); err != nil {
		return domain.Campaign{}, err
	}

	uc.logger.InfoContext(ctx, "cms.campaign.disabled",
		slog.String("campaign_id", disabled.CampaignID),
		slog.String("seller_id", sellerID),
		slog.String("request_id", input.requestID()),
	)
	return disabled, nil
}

func (uc *CampaignUsecase) ListCampaigns(ctx context.Context, input ListCampaignsInput) ([]domain.Campaign, error) {
	input = input.normalized()
	sellerID := input.Actor.SellerID
	if sellerID == "" {
		return nil, domain.ErrUnauthenticated
	}
	if err := uc.authorizeCampaignAction(ctx, input.Actor, sellerID, "", domain.PermissionCampaignsRead, input.requestID()); err != nil {
		return nil, err
	}
	return uc.repo.ListCampaigns(ctx, CampaignFilter{
		SellerID:    sellerID,
		Status:      input.Status,
		StartsAfter: input.StartsAfter,
		EndsBefore:  input.EndsBefore,
		Query:       input.Query,
		Limit:       normalizedLimit(input.Limit),
		Offset:      normalizedOffset(input.Offset),
	})
}

func (uc *CampaignUsecase) GetCampaign(ctx context.Context, input GetCampaignInput) (CampaignRead, error) {
	input = input.normalized()
	if input.CampaignID == "" {
		return CampaignRead{}, domain.NewValidationError("Invalid campaign read request.", []domain.FieldViolation{{Field: "campaign_id", Reason: "Campaign id is required."}})
	}
	sellerID := input.SellerID
	if sellerID == "" {
		sellerID = input.Actor.SellerID
	}
	if sellerID == "" {
		return CampaignRead{}, domain.ErrSellerContextRequired
	}
	if input.InternalCaller == "" {
		if err := uc.authorizeCampaignAction(ctx, input.Actor, sellerID, input.CampaignID, domain.PermissionCampaignsRead, input.requestID()); err != nil {
			return CampaignRead{}, err
		}
	}

	campaign, err := uc.repo.GetCampaignByID(ctx, input.CampaignID)
	if err != nil {
		return CampaignRead{}, err
	}
	campaign = campaign.Normalized()
	if campaignSellerID(campaign) != sellerID {
		return CampaignRead{}, domain.ErrCampaignNotFound
	}
	spent, err := uc.repo.SumCampaignDiscounts(ctx, campaign.CampaignID)
	if err != nil {
		return CampaignRead{}, err
	}
	return CampaignRead{
		Campaign: campaign,
		Spent:    domain.Money{Amount: spent, Currency: campaign.Currency},
	}, nil
}

func (uc *CampaignUsecase) ListCampaignReads(ctx context.Context, input ListCampaignsInput) ([]CampaignRead, error) {
	campaigns, err := uc.ListCampaigns(ctx, input)
	if err != nil {
		return nil, err
	}
	reads := make([]CampaignRead, 0, len(campaigns))
	for _, campaign := range campaigns {
		campaign = campaign.Normalized()
		spent, err := uc.repo.SumCampaignDiscounts(ctx, campaign.CampaignID)
		if err != nil {
			return nil, err
		}
		reads = append(reads, CampaignRead{
			Campaign: campaign,
			Spent:    domain.Money{Amount: spent, Currency: campaign.Currency},
		})
	}
	return reads, nil
}

func (uc *CampaignUsecase) ValidateCampaign(ctx context.Context, input ValidateCampaignInput) (domain.CampaignValidationResult, error) {
	input = input.normalized()
	if input.CampaignID == "" || input.SellerID == "" {
		return domain.CampaignValidationResult{}, domain.NewValidationError("Invalid campaign validation request.", []domain.FieldViolation{{Field: "campaign", Reason: "Campaign id and seller id are required."}})
	}

	campaign, err := uc.repo.GetCampaignByID(ctx, input.CampaignID)
	if err != nil {
		if errors.Is(err, domain.ErrCampaignNotFound) {
			return domain.InvalidCampaignResult(input.Currency, input.CampaignID, input.CouponID, domain.CampaignInvalidReasonNotFound), nil
		}
		return domain.CampaignValidationResult{}, err
	}
	reason, remaining, unlimited, err := uc.campaignInvalidReason(ctx, campaign, input.SellerID, input.CouponID, input.UserID, input.Currency, input.DiscountAmount)
	if err != nil {
		return domain.CampaignValidationResult{}, err
	}
	if reason != "" {
		return domain.InvalidCampaignResult(input.Currency, campaign.CampaignID, input.CouponID, reason), nil
	}
	return domain.CampaignValidationResult{
		Valid:           true,
		CampaignID:      campaign.CampaignID,
		CouponID:        input.CouponID,
		RemainingBudget: domain.Money{Amount: remaining, Currency: campaign.Normalized().Currency},
		BudgetUnlimited: unlimited,
	}, nil
}

func (uc *CampaignUsecase) campaignInvalidReason(ctx context.Context, campaign domain.Campaign, sellerID string, couponID string, userID string, currency string, discountAmount int64) (domain.CampaignInvalidReason, int64, bool, error) {
	campaign = campaign.Normalized()
	if reason := domain.ValidateCampaignEligibility(campaign, sellerID, couponID, currency, uc.now()); reason != "" {
		return reason, 0, false, nil
	}
	if campaign.Metadata.UsageLimit != nil {
		used, err := uc.repo.CountCampaignRedemptions(ctx, campaign.CampaignID)
		if err != nil {
			return "", 0, false, err
		}
		if used >= *campaign.Metadata.UsageLimit {
			return domain.CampaignInvalidReasonUsageLimitReached, 0, false, nil
		}
	}
	if campaign.Metadata.PerUserLimit != nil && strings.TrimSpace(userID) != "" {
		used, err := uc.repo.CountCampaignUserRedemptions(ctx, campaign.CampaignID, userID)
		if err != nil {
			return "", 0, false, err
		}
		if used >= *campaign.Metadata.PerUserLimit {
			return domain.CampaignInvalidReasonUsageLimitReached, 0, false, nil
		}
	}
	if campaign.BudgetAmount == nil {
		return "", 0, true, nil
	}
	spent, err := uc.repo.SumCampaignDiscounts(ctx, campaign.CampaignID)
	if err != nil {
		return "", 0, false, err
	}
	remaining := *campaign.BudgetAmount - spent
	if remaining <= 0 || (discountAmount > 0 && discountAmount > remaining) {
		return domain.CampaignInvalidReasonBudgetExhausted, 0, false, nil
	}
	return "", remaining, false, nil
}

func (uc *CampaignUsecase) validateCampaign(ctx context.Context, campaign domain.Campaign) error {
	campaign = campaign.Normalized()
	if err := domain.ValidateCampaignDefinition(campaign, uc.maxDuration); err != nil {
		return err
	}
	sellerID := campaignSellerID(campaign)
	if sellerID == "" {
		return domain.NewValidationError("Invalid campaign seller.", []domain.FieldViolation{{Field: "seller_id", Reason: "Seller campaign must belong to an authenticated seller."}})
	}
	for _, couponID := range campaign.Metadata.CouponIDs {
		coupon, err := uc.couponReader.GetCouponByID(ctx, couponID)
		if err != nil {
			return domain.NewValidationError("Invalid campaign coupon scope.", []domain.FieldViolation{{Field: "metadata.coupon_ids", Reason: "Linked coupon must exist and belong to the seller."}})
		}
		coupon = coupon.Normalized()
		if couponSellerID(coupon) != sellerID {
			return domain.NewValidationError("Invalid campaign coupon scope.", []domain.FieldViolation{{Field: "metadata.coupon_ids", Reason: "Linked coupon must belong to the same seller."}})
		}
		if coupon.Currency != campaign.Currency {
			return domain.NewValidationError("Invalid campaign coupon scope.", []domain.FieldViolation{{Field: "metadata.coupon_ids", Reason: "Linked coupon currency must match the campaign currency."}})
		}
	}
	return nil
}

func (uc *CampaignUsecase) authorizeCampaignAction(ctx context.Context, actor domain.ActorContext, sellerID string, campaignID string, permission domain.Permission, requestID string) error {
	_, err := uc.authorizer.Authorize(ctx, AuthorizeInput{
		Actor:              actor,
		ResourceSellerID:   sellerID,
		RequiredPermission: permission,
		ResourceType:       "campaign",
		ResourceID:         campaignID,
		RequestID:          requestID,
	})
	return err
}

func (uc *CampaignUsecase) recordCampaignAudit(ctx context.Context, action domain.Permission, actor domain.ActorContext, sellerID string, campaignID string, requestID string, before map[string]any, after map[string]any) error {
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
		ResourceType:     "campaign",
		ResourceID:       campaignID,
		ResourceSellerID: sellerID,
		RequestID:        requestID,
		Decision:         domain.AuditDecisionAllowed,
		Before:           before,
		After:            after,
		CreatedAt:        uc.now(),
	}); err != nil {
		uc.logger.ErrorContext(ctx, "cms.campaign.audit_failed",
			slog.String("action", action.String()),
			slog.String("campaign_id", campaignID),
			slog.String("seller_id", sellerID),
			slog.String("request_id", requestID),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("%w: %v", domain.ErrAuditWriteFailed, err)
	}
	return nil
}

func enforceCampaignEditRules(existing domain.Campaign, updated domain.Campaign, now time.Time) error {
	existing = existing.Normalized()
	updated = updated.Normalized()
	if existing.Status != domain.CampaignStatusActive {
		return nil
	}
	now = now.UTC()
	if updated.StartsAt.After(now) {
		return domain.NewValidationError("Invalid active campaign update.", []domain.FieldViolation{{Field: "starts_at", Reason: "Active campaign start time cannot be moved into the future."}})
	}
	if updated.EndsAt.Before(existing.EndsAt) {
		return domain.NewValidationError("Invalid active campaign update.", []domain.FieldViolation{{Field: "ends_at", Reason: "Active campaign end time can only be extended."}})
	}
	if budgetReduced(existing.BudgetAmount, updated.BudgetAmount) {
		return domain.NewValidationError("Invalid active campaign update.", []domain.FieldViolation{{Field: "budget.amount", Reason: "Active campaign budget can only be increased."}})
	}
	if usageLimitReduced(existing.Metadata.UsageLimit, updated.Metadata.UsageLimit) {
		return domain.NewValidationError("Invalid active campaign update.", []domain.FieldViolation{{Field: "metadata.usage_limit", Reason: "Active campaign usage limit cannot be reduced."}})
	}
	if usageLimitReduced(existing.Metadata.PerUserLimit, updated.Metadata.PerUserLimit) {
		return domain.NewValidationError("Invalid active campaign update.", []domain.FieldViolation{{Field: "metadata.per_user_limit", Reason: "Active campaign per-user limit cannot be reduced."}})
	}
	if !containsAll(updated.Metadata.CouponIDs, existing.Metadata.CouponIDs) {
		return domain.NewValidationError("Invalid active campaign update.", []domain.FieldViolation{{Field: "metadata.coupon_ids", Reason: "Active campaign linked coupons cannot be removed."}})
	}
	return nil
}

func budgetReduced(existing *int64, updated *int64) bool {
	switch {
	case existing == nil && updated != nil:
		return true
	case existing != nil && updated != nil:
		return *updated < *existing
	default:
		return false
	}
}

func usageLimitReduced(existing *int64, updated *int64) bool {
	switch {
	case existing == nil:
		return false
	case updated == nil:
		return false
	default:
		return *updated < *existing
	}
}

func containsAll(values []string, required []string) bool {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[strings.TrimSpace(value)] = struct{}{}
	}
	for _, value := range required {
		if _, ok := set[strings.TrimSpace(value)]; !ok {
			return false
		}
	}
	return true
}

func campaignAuditAction(from domain.CampaignStatus, to domain.CampaignStatus) domain.Permission {
	if from != to {
		switch to {
		case domain.CampaignStatusPaused:
			return domain.AuditActionCampaignDisabled
		case domain.CampaignStatusActive:
			return domain.AuditActionCampaignResumed
		case domain.CampaignStatusCompleted:
			return domain.AuditActionCampaignCompleted
		}
	}
	return domain.AuditActionCampaignUpdated
}

func campaignAuditSnapshot(campaign domain.Campaign) map[string]any {
	campaign = campaign.Normalized()
	return map[string]any{
		"campaign_id":   campaign.CampaignID,
		"seller_id":     campaignSellerID(campaign),
		"name":          campaign.Name,
		"status":        campaign.Status,
		"budget_amount": campaign.BudgetAmount,
		"currency":      campaign.Currency,
		"starts_at":     campaign.StartsAt,
		"ends_at":       campaign.EndsAt,
		"metadata":      campaign.Metadata,
	}
}

func campaignSellerID(campaign domain.Campaign) string {
	if campaign.SellerID == nil {
		return ""
	}
	return strings.TrimSpace(*campaign.SellerID)
}

func (i CreateCampaignInput) normalized() CreateCampaignInput {
	i.Actor = i.Actor.Normalized()
	i.Name = strings.TrimSpace(i.Name)
	i.Currency = domain.NormalizeCurrency(i.Currency)
	i.StartsAt = i.StartsAt.UTC()
	i.EndsAt = i.EndsAt.UTC()
	i.Metadata = i.Metadata.Normalized()
	i.RequestID = strings.TrimSpace(i.RequestID)
	return i
}

func (i CreateCampaignInput) requestID() string {
	if i.RequestID != "" {
		return i.RequestID
	}
	return i.Actor.RequestID
}

func (i UpdateCampaignInput) normalized() UpdateCampaignInput {
	i.Actor = i.Actor.Normalized()
	i.CampaignID = strings.TrimSpace(i.CampaignID)
	if i.Name != nil {
		name := strings.TrimSpace(*i.Name)
		i.Name = &name
	}
	if i.Currency != nil {
		currency := domain.NormalizeCurrency(*i.Currency)
		i.Currency = &currency
	}
	if i.Status != nil {
		status := i.Status.Normalized()
		i.Status = &status
	}
	if i.Metadata != nil {
		metadata := i.Metadata.Normalized()
		i.Metadata = &metadata
	}
	i.RequestID = strings.TrimSpace(i.RequestID)
	return i
}

func (i UpdateCampaignInput) requestID() string {
	if i.RequestID != "" {
		return i.RequestID
	}
	return i.Actor.RequestID
}

func (i DisableCampaignInput) normalized() DisableCampaignInput {
	i.Actor = i.Actor.Normalized()
	i.CampaignID = strings.TrimSpace(i.CampaignID)
	i.RequestID = strings.TrimSpace(i.RequestID)
	return i
}

func (i DisableCampaignInput) requestID() string {
	if i.RequestID != "" {
		return i.RequestID
	}
	return i.Actor.RequestID
}

func (i ListCampaignsInput) normalized() ListCampaignsInput {
	i.Actor = i.Actor.Normalized()
	i.Status = i.Status.Normalized()
	i.Query = strings.TrimSpace(i.Query)
	i.RequestID = strings.TrimSpace(i.RequestID)
	return i
}

func (i ListCampaignsInput) requestID() string {
	if i.RequestID != "" {
		return i.RequestID
	}
	return i.Actor.RequestID
}

func (i GetCampaignInput) normalized() GetCampaignInput {
	i.Actor = i.Actor.Normalized()
	i.SellerID = strings.TrimSpace(i.SellerID)
	i.CampaignID = strings.TrimSpace(i.CampaignID)
	i.InternalCaller = strings.TrimSpace(i.InternalCaller)
	i.RequestID = strings.TrimSpace(i.RequestID)
	return i
}

func (i GetCampaignInput) requestID() string {
	if i.RequestID != "" {
		return i.RequestID
	}
	return i.Actor.RequestID
}

func (i ValidateCampaignInput) normalized() ValidateCampaignInput {
	i.CampaignID = strings.TrimSpace(i.CampaignID)
	i.SellerID = strings.TrimSpace(i.SellerID)
	i.CouponID = strings.TrimSpace(i.CouponID)
	i.UserID = strings.TrimSpace(i.UserID)
	i.Currency = domain.NormalizeCurrency(i.Currency)
	i.RequestID = strings.TrimSpace(i.RequestID)
	return i
}
