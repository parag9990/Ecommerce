package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
)

const defaultMergeGuestCartSaveAttempts = 3

type MergeCartRepository interface {
	FindCartByID(ctx context.Context, cartID string) (*domain.Cart, error)
	FindActiveByOwner(ctx context.Context, owner domain.CartOwner) (*domain.Cart, error)
	MergeGuestIntoUser(ctx context.Context, userCart *domain.Cart, guestCart *domain.Cart, expectedUserVersion int64, expectedGuestVersion int64) error
	ExpireActiveCart(ctx context.Context, cartID string, expectedVersion int64, now time.Time) (bool, error)
}

type MergeCartCache interface {
	SetActiveCart(ctx context.Context, owner domain.CartOwner, cart *domain.Cart) error
	SetCartSummary(ctx context.Context, owner domain.CartOwner, summary domain.CartSummary) error
	DeleteActiveCart(ctx context.Context, owner domain.CartOwner) error
	DeleteCartSummary(ctx context.Context, owner domain.CartOwner) error
}

type MergeGuestCartUsecase struct {
	repository                MergeCartRepository
	cache                     MergeCartCache
	ids                       IDGenerator
	clock                     Clock
	logger                    *slog.Logger
	ttlPolicy                 CartTTLPolicy
	defaultCurrency           string
	maxSaveAttempts           int
	allowGuestCartIDOnlyMerge bool
}

type MergeGuestCartDependencies struct {
	Repository                MergeCartRepository
	Cache                     MergeCartCache
	IDGenerator               IDGenerator
	Clock                     Clock
	Logger                    *slog.Logger
	CartTTL                   time.Duration
	UserCartTTL               time.Duration
	GuestCartTTL              time.Duration
	DefaultCurrency           string
	MaxSaveAttempts           int
	AllowGuestCartIDOnlyMerge bool
}

type MergeGuestCartCommand struct {
	UserID         string
	GuestSessionID string
	GuestCartID    string
}

func NewMergeGuestCartUsecase(deps MergeGuestCartDependencies) (*MergeGuestCartUsecase, error) {
	if deps.Repository == nil {
		return nil, errors.New("cart repository is required")
	}
	if deps.Cache == nil {
		return nil, errors.New("cart cache is required")
	}
	if deps.IDGenerator == nil {
		deps.IDGenerator = CryptoIDGenerator{}
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
	deps.DefaultCurrency = strings.ToUpper(strings.TrimSpace(deps.DefaultCurrency))
	if deps.DefaultCurrency == "" {
		deps.DefaultCurrency = domain.CurrencyINR
	}
	if err := domain.NewMoney(0, deps.DefaultCurrency).Validate(); err != nil {
		return nil, err
	}
	if deps.MaxSaveAttempts <= 0 {
		deps.MaxSaveAttempts = defaultMergeGuestCartSaveAttempts
	}
	return &MergeGuestCartUsecase{
		repository:                deps.Repository,
		cache:                     deps.Cache,
		ids:                       deps.IDGenerator,
		clock:                     deps.Clock,
		logger:                    deps.Logger,
		ttlPolicy:                 ttlPolicy,
		defaultCurrency:           deps.DefaultCurrency,
		maxSaveAttempts:           deps.MaxSaveAttempts,
		allowGuestCartIDOnlyMerge: deps.AllowGuestCartIDOnlyMerge,
	}, nil
}

func (u *MergeGuestCartUsecase) MergeGuestCart(ctx context.Context, cmd MergeGuestCartCommand) (*domain.Cart, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cmd = normalizeMergeGuestCartCommand(cmd)
	if err := ValidateMergeGuestCartCommand(cmd); err != nil {
		return nil, err
	}
	userOwner := domain.CartOwner{Type: domain.CartOwnerTypeUser, UserID: cmd.UserID}
	if err := userOwner.Validate(); err != nil {
		return nil, err
	}

	u.logger.Info(
		"cart.merge.started",
		slog.String("user_id", cmd.UserID),
		slog.String("guest_cart_id", cmd.GuestCartID),
		slog.String("guest_session_id_hash", sessionIDHash(cmd.GuestSessionID)),
	)

	var lastConflict error
	for attempt := 1; attempt <= u.maxSaveAttempts; attempt++ {
		userCart, guestCart, warnings, err := u.mergeGuestCartAttempt(ctx, userOwner, cmd)
		if err == nil {
			u.logMergeWarnings(cmd, userCart, warnings)
			u.refreshMergeCache(ctx, userOwner, guestCart, userCart)
			u.logger.Info(
				"cart.merge.completed",
				slog.String("user_id", cmd.UserID),
				slog.String("guest_cart_id", cmd.GuestCartID),
				slog.String("target_cart_id", userCart.ID),
				slog.Int("item_count", userCart.Totals.ItemCount),
				slog.Int("warnings_count", len(warnings)),
			)
			return userCart, nil
		}
		if errors.Is(err, domain.ErrGuestCartAccessDenied) || errors.Is(err, domain.ErrGuestSessionIDRequired) {
			u.logger.Warn("cart.merge.guest_access_denied", slog.String("guest_cart_id", cmd.GuestCartID), slog.String("user_id", cmd.UserID))
		}
		if !isRetryableMergeConflict(err) {
			return nil, err
		}
		lastConflict = err
		u.logger.Warn(
			"cart.merge.version_conflict",
			slog.String("guest_cart_id", cmd.GuestCartID),
			slog.Int("attempt", attempt),
		)
	}
	if lastConflict == nil {
		lastConflict = domain.ErrCartVersionConflict
	}
	return nil, lastConflict
}

func ValidateMergeGuestCartCommand(cmd MergeGuestCartCommand) error {
	if strings.TrimSpace(cmd.UserID) == "" {
		return domain.ErrUserIDRequired
	}
	if strings.TrimSpace(cmd.GuestCartID) == "" {
		return domain.ErrGuestCartIDRequired
	}
	return nil
}

func (u *MergeGuestCartUsecase) mergeGuestCartAttempt(ctx context.Context, userOwner domain.CartOwner, cmd MergeGuestCartCommand) (*domain.Cart, *domain.Cart, []domain.CartMergeWarning, error) {
	guestCart, err := u.repository.FindCartByID(ctx, cmd.GuestCartID)
	if err != nil {
		return nil, nil, nil, err
	}
	if err := u.ensureGuestCartAccessible(guestCart, cmd); err != nil {
		return nil, nil, nil, err
	}
	if guestCart.Status == domain.CartStatusMerged {
		userCart, err := u.idempotentMergedCart(ctx, userOwner, guestCart)
		return userCart, guestCart, nil, err
	}
	if guestCart.Status != domain.CartStatusActive {
		return nil, nil, nil, domain.ErrCartNotActive
	}

	now := u.clock.Now().UTC()
	guestOwner := domain.CartOwner{Type: domain.CartOwnerTypeGuest, GuestSessionID: cartGuestSessionID(guestCart)}
	if domain.IsActiveCartExpired(guestCart, now) {
		if err := expireActiveCartForOwner(ctx, u.repository, u.cache, u.logger, guestOwner, guestCart, now, "merge_guest_cart"); err != nil {
			return nil, nil, nil, err
		}
		return nil, nil, nil, domain.ErrCartExpired
	}
	expiresAt := now.Add(u.ttlPolicy.TTLForOwner(userOwner))
	userCart, expectedUserVersion, err := u.loadOrBuildUserCart(ctx, userOwner, guestCart, now, expiresAt)
	if err != nil {
		return nil, nil, nil, err
	}
	if userCart.ID == guestCart.ID {
		return nil, nil, nil, fmt.Errorf("%w: source and target cart cannot be the same", domain.ErrInvalidCart)
	}
	expectedGuestVersion := guestCart.Version

	warnings, err := domain.MergeGuestItemsIntoUserCart(userCart, guestCart, now, u.ids.NewItemID)
	if err != nil {
		return nil, nil, nil, err
	}
	userCart.ClearCouponPreview()
	if err := userCart.RecalculateTotals(mergeRecalculateCurrency(userCart, u.defaultCurrency)); err != nil {
		return nil, nil, nil, err
	}
	userCart.Touch(now, expiresAt)
	if err := guestCart.MarkMerged(userCart.ID, now); err != nil {
		return nil, nil, nil, err
	}
	if err := u.repository.MergeGuestIntoUser(ctx, userCart, guestCart, expectedUserVersion, expectedGuestVersion); err != nil {
		return nil, nil, nil, err
	}
	return userCart, guestCart, warnings, nil
}

func (u *MergeGuestCartUsecase) ensureGuestCartAccessible(guestCart *domain.Cart, cmd MergeGuestCartCommand) error {
	if guestCart == nil {
		return domain.ErrCartNotFound
	}
	if guestCart.UserID != nil && strings.TrimSpace(*guestCart.UserID) != "" {
		return domain.ErrGuestCartAccessDenied
	}
	guestSessionID := cartGuestSessionID(guestCart)
	if guestSessionID == "" {
		return domain.ErrGuestCartAccessDenied
	}
	if cmd.GuestSessionID == "" && !u.allowGuestCartIDOnlyMerge {
		return domain.ErrGuestSessionIDRequired
	}
	if cmd.GuestSessionID != "" && cmd.GuestSessionID != guestSessionID {
		return domain.ErrGuestCartAccessDenied
	}
	return nil
}

func (u *MergeGuestCartUsecase) idempotentMergedCart(ctx context.Context, userOwner domain.CartOwner, guestCart *domain.Cart) (*domain.Cart, error) {
	if guestCart.MergedIntoCartID == nil || strings.TrimSpace(*guestCart.MergedIntoCartID) == "" {
		return nil, domain.ErrCartNotActive
	}
	userCart, err := u.repository.FindCartByID(ctx, *guestCart.MergedIntoCartID)
	if err != nil {
		return nil, err
	}
	if userCart.UserID == nil || strings.TrimSpace(*userCart.UserID) != userOwner.UserID {
		return nil, domain.ErrGuestCartAccessDenied
	}
	if userCart.Status != domain.CartStatusActive {
		return nil, domain.ErrCartNotActive
	}
	if domain.IsActiveCartExpired(userCart, u.clock.Now().UTC()) {
		return nil, domain.ErrCartExpired
	}
	return userCart, nil
}

func (u *MergeGuestCartUsecase) loadOrBuildUserCart(ctx context.Context, userOwner domain.CartOwner, guestCart *domain.Cart, now time.Time, expiresAt time.Time) (*domain.Cart, int64, error) {
	userCart, err := u.repository.FindActiveByOwner(ctx, userOwner)
	if err == nil {
		if domain.IsActiveCartExpired(userCart, now) {
			if err := expireActiveCartForOwner(ctx, u.repository, u.cache, u.logger, userOwner, userCart, now, "merge_user_cart"); err != nil {
				return nil, 0, err
			}
			return u.buildUserCart(userOwner, guestCart, now, expiresAt)
		}
		return userCart, userCart.Version, nil
	}
	if !errors.Is(err, domain.ErrCartNotFound) {
		return nil, 0, err
	}
	return u.buildUserCart(userOwner, guestCart, now, expiresAt)
}

func (u *MergeGuestCartUsecase) buildUserCart(userOwner domain.CartOwner, guestCart *domain.Cart, now time.Time, expiresAt time.Time) (*domain.Cart, int64, error) {
	cartID, err := u.ids.NewCartID()
	if err != nil {
		return nil, 0, err
	}
	currency := mergeInitialCurrency(guestCart, u.defaultCurrency)
	userCart, err := domain.NewActiveCart(cartID, userOwner, currency, now, expiresAt)
	if err != nil {
		return nil, 0, err
	}
	return userCart, 0, nil
}

func (u *MergeGuestCartUsecase) refreshMergeCache(ctx context.Context, userOwner domain.CartOwner, guestCart *domain.Cart, userCart *domain.Cart) {
	summary := domain.BuildCartSummary(userCart)
	if err := u.cache.SetActiveCart(ctx, userOwner, userCart); err != nil {
		u.logger.Warn("cart.merge.cache_refresh_failed", slog.String("cache_key_type", "active_user"), slog.String("target_cart_id", userCart.ID), slog.String("error", err.Error()))
	}
	if err := u.cache.SetCartSummary(ctx, userOwner, summary); err != nil {
		u.logger.Warn("cart.merge.cache_refresh_failed", slog.String("cache_key_type", "summary_user"), slog.String("target_cart_id", userCart.ID), slog.String("error", err.Error()))
	}

	guestSessionID := cartGuestSessionID(guestCart)
	if guestSessionID == "" {
		return
	}
	guestOwner := domain.CartOwner{Type: domain.CartOwnerTypeGuest, GuestSessionID: guestSessionID}
	if err := u.cache.DeleteActiveCart(ctx, guestOwner); err != nil {
		u.logger.Warn("cart.merge.cache_refresh_failed", slog.String("cache_key_type", "active_guest"), slog.String("target_cart_id", userCart.ID), slog.String("error", err.Error()))
	}
	if err := u.cache.DeleteCartSummary(ctx, guestOwner); err != nil {
		u.logger.Warn("cart.merge.cache_refresh_failed", slog.String("cache_key_type", "summary_guest"), slog.String("target_cart_id", userCart.ID), slog.String("error", err.Error()))
	}
}

func (u *MergeGuestCartUsecase) logMergeWarnings(cmd MergeGuestCartCommand, userCart *domain.Cart, warnings []domain.CartMergeWarning) {
	for _, warning := range warnings {
		u.logger.Warn(
			"cart.merge.warning",
			slog.String("code", warning.Code),
			slog.String("user_id", cmd.UserID),
			slog.String("guest_cart_id", cmd.GuestCartID),
			slog.String("target_cart_id", userCart.ID),
			slog.String("product_id", warning.ProductID),
			slog.String("variant_id", warning.VariantID),
			slog.String("message", warning.Message),
		)
	}
}

func normalizeMergeGuestCartCommand(cmd MergeGuestCartCommand) MergeGuestCartCommand {
	cmd.UserID = strings.TrimSpace(cmd.UserID)
	cmd.GuestSessionID = strings.TrimSpace(cmd.GuestSessionID)
	cmd.GuestCartID = strings.TrimSpace(cmd.GuestCartID)
	return cmd
}

func mergeInitialCurrency(guestCart *domain.Cart, defaultCurrency string) string {
	if guestCart != nil && strings.TrimSpace(guestCart.Totals.Currency) != "" {
		return guestCart.Totals.Currency
	}
	return defaultCurrency
}

func mergeRecalculateCurrency(cart *domain.Cart, defaultCurrency string) string {
	if cart != nil && len(cart.Items) > 0 {
		return cart.Items[0].UnitPrice.Currency
	}
	if cart != nil && strings.TrimSpace(cart.Totals.Currency) != "" {
		return cart.Totals.Currency
	}
	return defaultCurrency
}

func cartGuestSessionID(cart *domain.Cart) string {
	if cart == nil || cart.GuestSessionID == nil {
		return ""
	}
	return strings.TrimSpace(*cart.GuestSessionID)
}

func isRetryableMergeConflict(err error) bool {
	return errors.Is(err, domain.ErrCartVersionConflict) || errors.Is(err, domain.ErrActiveCartExists)
}

func sessionIDHash(sessionID string) string {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(sessionID))
	return hex.EncodeToString(sum[:8])
}
