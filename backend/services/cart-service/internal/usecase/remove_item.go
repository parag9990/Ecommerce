package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
)

const defaultRemoveItemSaveAttempts = 3

type RemoveItemUsecase struct {
	repository      CartRepository
	cache           CartCache
	clock           Clock
	logger          *slog.Logger
	ttlPolicy       CartTTLPolicy
	defaultCurrency string
	maxSaveAttempts int
}

type RemoveItemDependencies struct {
	Repository      CartRepository
	Cache           CartCache
	Clock           Clock
	Logger          *slog.Logger
	CartTTL         time.Duration
	UserCartTTL     time.Duration
	GuestCartTTL    time.Duration
	DefaultCurrency string
	MaxSaveAttempts int
}

type RemoveItemCommand struct {
	UserID         string
	GuestSessionID string
	ItemID         string
}

func NewRemoveItemUsecase(deps RemoveItemDependencies) (*RemoveItemUsecase, error) {
	if deps.Repository == nil {
		return nil, errors.New("cart repository is required")
	}
	if deps.Cache == nil {
		return nil, errors.New("cart cache is required")
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
		deps.MaxSaveAttempts = defaultRemoveItemSaveAttempts
	}
	return &RemoveItemUsecase{
		repository:      deps.Repository,
		cache:           deps.Cache,
		clock:           deps.Clock,
		logger:          deps.Logger,
		ttlPolicy:       ttlPolicy,
		defaultCurrency: deps.DefaultCurrency,
		maxSaveAttempts: deps.MaxSaveAttempts,
	}, nil
}

func (u *RemoveItemUsecase) RemoveItem(ctx context.Context, cmd RemoveItemCommand) (*domain.Cart, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cmd = normalizeRemoveItemCommand(cmd)
	if err := ValidateRemoveItemCommand(cmd); err != nil {
		return nil, err
	}
	owner, err := domain.ResolveCartOwner(cmd.UserID, cmd.GuestSessionID)
	if err != nil {
		return nil, err
	}

	var lastConflict error
	for attempt := 1; attempt <= u.maxSaveAttempts; attempt++ {
		cart, err := u.removeItemAttempt(ctx, owner, cmd.ItemID)
		if err == nil {
			u.refreshCache(ctx, owner, cart)
			return cart, nil
		}
		if !errors.Is(err, domain.ErrCartVersionConflict) {
			return nil, err
		}
		lastConflict = err
		u.logger.Warn(
			"cart.remove_item.version_conflict",
			slog.String("owner_type", string(owner.Type)),
			slog.String("item_id", cmd.ItemID),
			slog.Int("attempt", attempt),
		)
	}
	if lastConflict == nil {
		lastConflict = domain.ErrCartVersionConflict
	}
	return nil, lastConflict
}

func ValidateRemoveItemCommand(cmd RemoveItemCommand) error {
	if strings.TrimSpace(cmd.ItemID) == "" {
		return domain.ErrItemIDRequired
	}
	return nil
}

func (u *RemoveItemUsecase) removeItemAttempt(ctx context.Context, owner domain.CartOwner, itemID string) (*domain.Cart, error) {
	now := u.clock.Now().UTC()
	expiresAt := now.Add(u.ttlPolicy.TTLForOwner(owner))

	cart, err := u.repository.FindActiveByOwner(ctx, owner)
	if err != nil {
		return nil, err
	}
	if cart.Status != domain.CartStatusActive {
		return nil, domain.ErrCartNotActive
	}
	if domain.IsActiveCartExpired(cart, now) {
		if err := expireActiveCartForOwner(ctx, u.repository, u.cache, u.logger, owner, cart, now, "remove_item"); err != nil {
			return nil, err
		}
		return nil, domain.ErrCartExpired
	}

	expectedVersion := cart.Version
	removed, err := cart.RemoveItem(itemID, now)
	if err != nil {
		return nil, err
	}
	cleaned := cart.CleanupZeroQuantityItems(now)
	if cleaned > 0 {
		u.logger.Warn(
			"cart.remove_item.cleaned_invalid_quantities",
			slog.String("cart_id", cart.ID),
			slog.Int("cleaned_count", cleaned),
		)
	}

	if !removed && cleaned == 0 {
		u.logger.Info(
			"cart.remove_item.idempotent",
			slog.String("cart_id", cart.ID),
			slog.String("item_id", itemID),
		)
		return cart, nil
	}

	cart.ClearCouponPreview()
	if err := cart.RecalculateTotals(recalculateCurrency(cart, u.defaultCurrency)); err != nil {
		return nil, err
	}
	cart.Touch(now, expiresAt)
	if err := u.repository.SaveWithVersion(ctx, cart, expectedVersion); err != nil {
		return nil, err
	}

	u.logger.Info(
		"cart.remove_item.saved",
		slog.String("cart_id", cart.ID),
		slog.String("item_id", itemID),
		slog.Bool("removed", removed),
		slog.Int("cleaned_count", cleaned),
		slog.Int("remaining_unique_items", len(cart.Items)),
	)
	return cart, nil
}

func (u *RemoveItemUsecase) refreshCache(ctx context.Context, owner domain.CartOwner, cart *domain.Cart) {
	summary := domain.BuildCartSummary(cart)
	if err := u.cache.SetActiveCart(ctx, owner, cart); err != nil {
		u.logger.Warn("cart.remove_item.cache_active_failed", slog.String("error", err.Error()), slog.String("cart_id", cart.ID))
	}
	if err := u.cache.SetCartSummary(ctx, owner, summary); err != nil {
		u.logger.Warn("cart.remove_item.cache_summary_failed", slog.String("error", err.Error()), slog.String("cart_id", cart.ID))
	}
}

func normalizeRemoveItemCommand(cmd RemoveItemCommand) RemoveItemCommand {
	cmd.UserID = strings.TrimSpace(cmd.UserID)
	cmd.GuestSessionID = strings.TrimSpace(cmd.GuestSessionID)
	cmd.ItemID = strings.TrimSpace(cmd.ItemID)
	return cmd
}

func recalculateCurrency(cart *domain.Cart, defaultCurrency string) string {
	if cart != nil && len(cart.Items) > 0 {
		return cart.Items[0].UnitPrice.Currency
	}
	return defaultCurrency
}
