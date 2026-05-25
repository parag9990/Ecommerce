package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
)

const (
	defaultCouponPreviewSaveAttempts = 3
	defaultCouponPreviewTimeout      = 700 * time.Millisecond
)

type CouponValidator interface {
	ValidateCoupon(ctx context.Context, req CouponValidationRequest) (*CouponValidationResult, error)
}

type CouponValidationRequest struct {
	Code           string
	UserID         string
	GuestSessionID string
	CartID         string
	Currency       string
	Subtotal       domain.Money
	Items          []CouponCartItem
	RequestedAt    time.Time
}

type CouponCartItem struct {
	ProductID    string
	VariantID    string
	SellerID     string
	UnitPrice    domain.Money
	Quantity     int
	LineSubtotal domain.Money
}

type CouponValidationResult struct {
	Valid    bool
	CouponID string
	Code     string
	Discount domain.Money
	Reason   string
}

type ApplyCouponPreviewUsecase struct {
	repository      CartRepository
	cache           CartCache
	couponValidator CouponValidator
	clock           Clock
	logger          *slog.Logger
	ttlPolicy       CartTTLPolicy
	requestTimeout  time.Duration
	maxSaveAttempts int
}

type ApplyCouponPreviewDependencies struct {
	Repository      CartRepository
	Cache           CartCache
	CouponValidator CouponValidator
	Clock           Clock
	Logger          *slog.Logger
	CartTTL         time.Duration
	UserCartTTL     time.Duration
	GuestCartTTL    time.Duration
	RequestTimeout  time.Duration
	MaxSaveAttempts int
}

type ApplyCouponPreviewCommand struct {
	UserID         string
	GuestSessionID string
	CartID         string
	CouponCode     string
}

type CouponPreviewResult struct {
	Valid    bool
	CouponID string
	Discount domain.Money
	Reason   string
}

func NewApplyCouponPreviewUsecase(deps ApplyCouponPreviewDependencies) (*ApplyCouponPreviewUsecase, error) {
	if deps.Repository == nil {
		return nil, errors.New("cart repository is required")
	}
	if deps.Cache == nil {
		return nil, errors.New("cart cache is required")
	}
	if deps.CouponValidator == nil {
		return nil, errors.New("coupon validator is required")
	}
	if deps.Clock == nil {
		deps.Clock = SystemClock{}
	}
	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}
	ttlPolicy, err := NewCartTTLPolicy(deps.CartTTL, deps.UserCartTTL, deps.GuestCartTTL)
	if err != nil {
		return nil, err
	}
	if deps.RequestTimeout <= 0 {
		deps.RequestTimeout = defaultCouponPreviewTimeout
	}
	if deps.MaxSaveAttempts <= 0 {
		deps.MaxSaveAttempts = defaultCouponPreviewSaveAttempts
	}
	return &ApplyCouponPreviewUsecase{
		repository:      deps.Repository,
		cache:           deps.Cache,
		couponValidator: deps.CouponValidator,
		clock:           deps.Clock,
		logger:          deps.Logger,
		ttlPolicy:       ttlPolicy,
		requestTimeout:  deps.RequestTimeout,
		maxSaveAttempts: deps.MaxSaveAttempts,
	}, nil
}

func (u *ApplyCouponPreviewUsecase) ApplyCouponPreview(ctx context.Context, cmd ApplyCouponPreviewCommand) (*CouponPreviewResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cmd = normalizeApplyCouponPreviewCommand(cmd)
	owner, err := domain.ResolveCartOwner(cmd.UserID, cmd.GuestSessionID)
	if err != nil {
		return nil, err
	}
	code, err := domain.NormalizeCouponCode(cmd.CouponCode)
	if err != nil {
		return nil, err
	}

	var lastConflict error
	for attempt := 1; attempt <= u.maxSaveAttempts; attempt++ {
		cart, preview, err := u.applyCouponPreviewAttempt(ctx, owner, cmd.CartID, code)
		if err == nil {
			u.refreshCache(ctx, owner, cart)
			u.logger.Info(
				"cart.coupon_preview.saved",
				slog.String("cart_id", cart.ID),
				slog.String("owner_type", string(owner.Type)),
				slog.Bool("valid", preview.Valid),
				slog.String("reason", preview.Reason),
				slog.Int64("discount_amount", preview.Discount.Amount),
				slog.String("currency", preview.Discount.Currency),
			)
			return preview, nil
		}
		if !errors.Is(err, domain.ErrCartVersionConflict) {
			return nil, err
		}
		lastConflict = err
		u.logger.Warn(
			"cart.coupon_preview.version_conflict",
			slog.String("owner_type", string(owner.Type)),
			slog.Int("attempt", attempt),
		)
	}
	if lastConflict == nil {
		lastConflict = domain.ErrCartVersionConflict
	}
	return nil, lastConflict
}

func (u *ApplyCouponPreviewUsecase) applyCouponPreviewAttempt(ctx context.Context, owner domain.CartOwner, cartID string, code string) (*domain.Cart, *CouponPreviewResult, error) {
	now := u.clock.Now().UTC()
	expiresAt := now.Add(u.ttlPolicy.TTLForOwner(owner))

	cart, err := u.repository.FindActiveByOwner(ctx, owner)
	if err != nil {
		return nil, nil, err
	}
	if domain.IsActiveCartExpired(cart, now) {
		if err := expireActiveCartForOwner(ctx, u.repository, u.cache, u.logger, owner, cart, now, "coupon_preview"); err != nil {
			return nil, nil, err
		}
		return nil, nil, domain.ErrCartExpired
	}
	if cartID != "" && cart.ID != cartID {
		return nil, nil, domain.ErrCartOwnershipMismatch
	}
	if err := domain.EnsureCouponPreviewAllowed(cart); err != nil {
		return nil, nil, err
	}

	expectedVersion := cart.Version
	cmsResult, err := u.validateCoupon(ctx, buildCouponValidationRequest(cart, owner, code, now))
	if err != nil {
		return nil, nil, err
	}
	if cmsResult == nil {
		return nil, nil, domain.ErrCouponPreviewUnavailable
	}
	cmsResult.Code = firstNonBlank(cmsResult.Code, code)

	if cmsResult.Valid {
		err = cart.ApplyValidCouponPreview(domain.CouponPreviewDecision{
			CouponID: cmsResult.CouponID,
			Code:     cmsResult.Code,
			Valid:    true,
			Discount: cmsResult.Discount,
		}, now)
	} else {
		err = cart.ApplyInvalidCouponPreview(code, cmsResult.Reason, now)
	}
	if err != nil {
		return nil, nil, err
	}
	cart.Touch(now, expiresAt)

	if err := u.repository.SaveWithVersion(ctx, cart, expectedVersion); err != nil {
		return nil, nil, err
	}
	return cart, couponPreviewResultFromCart(cart), nil
}

func (u *ApplyCouponPreviewUsecase) validateCoupon(ctx context.Context, req CouponValidationRequest) (*CouponValidationResult, error) {
	validateCtx := ctx
	cancel := func() {}
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && u.requestTimeout > 0 {
		validateCtx, cancel = context.WithTimeout(ctx, u.requestTimeout)
	}
	defer cancel()

	result, err := u.couponValidator.ValidateCoupon(validateCtx, req)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, err
		}
		return nil, fmt.Errorf("%w: %v", domain.ErrCouponPreviewUnavailable, err)
	}
	return result, nil
}

func (u *ApplyCouponPreviewUsecase) refreshCache(ctx context.Context, owner domain.CartOwner, cart *domain.Cart) {
	summary := domain.BuildCartSummary(cart)
	if err := u.cache.SetActiveCart(ctx, owner, cart); err != nil {
		u.logger.Warn("cart.coupon_preview.cache_active_failed", slog.String("error", err.Error()), slog.String("cart_id", cart.ID))
	}
	if err := u.cache.SetCartSummary(ctx, owner, summary); err != nil {
		u.logger.Warn("cart.coupon_preview.cache_summary_failed", slog.String("error", err.Error()), slog.String("cart_id", cart.ID))
	}
}

func buildCouponValidationRequest(cart *domain.Cart, owner domain.CartOwner, code string, now time.Time) CouponValidationRequest {
	items := make([]CouponCartItem, 0, len(cart.Items))
	for _, item := range cart.Items {
		items = append(items, CouponCartItem{
			ProductID:    strings.TrimSpace(item.ProductID),
			VariantID:    strings.TrimSpace(item.VariantID),
			SellerID:     strings.TrimSpace(item.SellerID),
			UnitPrice:    item.UnitPrice,
			Quantity:     item.Quantity,
			LineSubtotal: item.LineSubtotal,
		})
	}
	return CouponValidationRequest{
		Code:           code,
		UserID:         owner.UserID,
		GuestSessionID: owner.GuestSessionID,
		CartID:         cart.ID,
		Currency:       cart.Totals.Currency,
		Subtotal:       cart.Totals.Subtotal,
		Items:          items,
		RequestedAt:    now.UTC(),
	}
}

func couponPreviewResultFromCart(cart *domain.Cart) *CouponPreviewResult {
	if cart == nil || cart.CouponPreview == nil {
		return nil
	}
	return &CouponPreviewResult{
		Valid:    cart.CouponPreview.Valid,
		CouponID: cart.CouponPreview.CouponID,
		Discount: cart.CouponPreview.Discount,
		Reason:   cart.CouponPreview.Reason,
	}
}

func normalizeApplyCouponPreviewCommand(cmd ApplyCouponPreviewCommand) ApplyCouponPreviewCommand {
	cmd.UserID = strings.TrimSpace(cmd.UserID)
	cmd.GuestSessionID = strings.TrimSpace(cmd.GuestSessionID)
	cmd.CartID = strings.TrimSpace(cmd.CartID)
	cmd.CouponCode = strings.TrimSpace(cmd.CouponCode)
	return cmd
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
