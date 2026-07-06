package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
)

const defaultUpdateItemSaveAttempts = 3

type UpdateItemUsecase struct {
	repository      CartRepository
	cache           CartCache
	productClient   ProductClient
	clock           Clock
	logger          *slog.Logger
	ttlPolicy       CartTTLPolicy
	defaultCurrency string
	maxSaveAttempts int
}

type UpdateItemDependencies struct {
	Repository      CartRepository
	Cache           CartCache
	ProductClient   ProductClient
	Clock           Clock
	Logger          *slog.Logger
	CartTTL         time.Duration
	UserCartTTL     time.Duration
	GuestCartTTL    time.Duration
	DefaultCurrency string
	MaxSaveAttempts int
}

type UpdateItemCommand struct {
	UserID         string
	GuestSessionID string
	ItemID         string
	Quantity       int
}

func NewUpdateItemUsecase(deps UpdateItemDependencies) (*UpdateItemUsecase, error) {
	if deps.Repository == nil {
		return nil, errors.New("cart repository is required")
	}
	if deps.Cache == nil {
		return nil, errors.New("cart cache is required")
	}
	if deps.ProductClient == nil {
		return nil, errors.New("product client is required")
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
		deps.MaxSaveAttempts = defaultUpdateItemSaveAttempts
	}
	return &UpdateItemUsecase{
		repository:      deps.Repository,
		cache:           deps.Cache,
		productClient:   deps.ProductClient,
		clock:           deps.Clock,
		logger:          deps.Logger,
		ttlPolicy:       ttlPolicy,
		defaultCurrency: deps.DefaultCurrency,
		maxSaveAttempts: deps.MaxSaveAttempts,
	}, nil
}

func (u *UpdateItemUsecase) UpdateItem(ctx context.Context, cmd UpdateItemCommand) (*domain.Cart, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cmd = normalizeUpdateItemCommand(cmd)
	if err := ValidateUpdateItemCommand(cmd); err != nil {
		return nil, err
	}
	owner, err := domain.ResolveCartOwner(cmd.UserID, cmd.GuestSessionID)
	if err != nil {
		return nil, err
	}

	var lastConflict error
	for attempt := 1; attempt <= u.maxSaveAttempts; attempt++ {
		cart, err := u.updateItemAttempt(ctx, owner, cmd.ItemID, cmd.Quantity)
		if err == nil {
			u.refreshCache(ctx, owner, cart)
			return cart, nil
		}
		if !errors.Is(err, domain.ErrCartVersionConflict) {
			return nil, err
		}
		lastConflict = err
		u.logger.Warn(
			"cart.update_item.version_conflict",
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

func ValidateUpdateItemCommand(cmd UpdateItemCommand) error {
	if strings.TrimSpace(cmd.ItemID) == "" {
		return domain.ErrItemIDRequired
	}
	if cmd.Quantity < domain.MinItemQuantity {
		return domain.ErrQuantityTooSmall
	}
	if cmd.Quantity > domain.MaxItemQuantity {
		return domain.ErrQuantityTooLarge
	}
	return nil
}

func (u *UpdateItemUsecase) updateItemAttempt(ctx context.Context, owner domain.CartOwner, itemID string, quantity int) (*domain.Cart, error) {
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
		if err := expireActiveCartForOwner(ctx, u.repository, u.cache, u.logger, owner, cart, now, "update_item"); err != nil {
			return nil, err
		}
		return nil, domain.ErrCartExpired
	}

	item, found := findCartItem(cart, itemID)
	if !found {
		return nil, domain.ErrCartItemNotFound
	}
	product, err := u.productClient.GetProductForCart(ctx, item.ProductID)
	if err != nil {
		return nil, err
	}
	if err := validateProductForCart(product); err != nil {
		return nil, err
	}
	variant, err := FindSellableVariant(product, item.VariantID)
	if err != nil {
		return nil, err
	}
	snapshot, err := BuildProductVariantSnapshot(product, variant)
	if err != nil {
		return nil, err
	}

	expectedVersion := cart.Version
	if err := cart.SetItemQuantity(itemID, quantity, snapshot, now); err != nil {
		return nil, err
	}
	cart.ClearCouponPreview()
	if err := cart.RecalculateTotals(recalculateCurrency(cart, u.defaultCurrency)); err != nil {
		return nil, err
	}
	cart.Touch(now, expiresAt)
	if err := u.repository.SaveWithVersion(ctx, cart, expectedVersion); err != nil {
		return nil, err
	}
	return cart, nil
}

func (u *UpdateItemUsecase) refreshCache(ctx context.Context, owner domain.CartOwner, cart *domain.Cart) {
	summary := domain.BuildCartSummary(cart)
	if err := u.cache.SetActiveCart(ctx, owner, cart); err != nil {
		u.logger.Warn("cart.update_item.cache_active_failed", slog.String("error", err.Error()), slog.String("cart_id", cart.ID))
	}
	if err := u.cache.SetCartSummary(ctx, owner, summary); err != nil {
		u.logger.Warn("cart.update_item.cache_summary_failed", slog.String("error", err.Error()), slog.String("cart_id", cart.ID))
	}
}

func normalizeUpdateItemCommand(cmd UpdateItemCommand) UpdateItemCommand {
	cmd.UserID = strings.TrimSpace(cmd.UserID)
	cmd.GuestSessionID = strings.TrimSpace(cmd.GuestSessionID)
	cmd.ItemID = strings.TrimSpace(cmd.ItemID)
	return cmd
}

func findCartItem(cart *domain.Cart, itemID string) (domain.CartItem, bool) {
	if cart == nil {
		return domain.CartItem{}, false
	}
	itemID = strings.TrimSpace(itemID)
	for _, item := range cart.Items {
		if strings.TrimSpace(item.ItemID) == itemID {
			return item, true
		}
	}
	return domain.CartItem{}, false
}
