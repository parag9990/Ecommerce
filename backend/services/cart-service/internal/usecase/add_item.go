package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
)

const defaultAddItemSaveAttempts = 3

type CartRepository interface {
	FindActiveByOwner(ctx context.Context, owner domain.CartOwner) (*domain.Cart, error)
	CreateActiveCart(ctx context.Context, cart *domain.Cart) error
	SaveWithVersion(ctx context.Context, cart *domain.Cart, expectedVersion int64) error
	ExpireActiveCart(ctx context.Context, cartID string, expectedVersion int64, now time.Time) (bool, error)
}

type CartCache interface {
	SetActiveCart(ctx context.Context, owner domain.CartOwner, cart *domain.Cart) error
	SetCartSummary(ctx context.Context, owner domain.CartOwner, summary domain.CartSummary) error
	DeleteActiveCart(ctx context.Context, owner domain.CartOwner) error
	DeleteCartSummary(ctx context.Context, owner domain.CartOwner) error
}

type ProductClient interface {
	GetProductForCart(ctx context.Context, productID string) (*ProductForCart, error)
}

type IDGenerator interface {
	NewCartID() (string, error)
	NewItemID() (string, error)
}

type Clock interface {
	Now() time.Time
}

type AddItemUsecase struct {
	repository      CartRepository
	cache           CartCache
	productClient   ProductClient
	ids             IDGenerator
	clock           Clock
	logger          *slog.Logger
	ttlPolicy       CartTTLPolicy
	maxSaveAttempts int
}

type AddItemDependencies struct {
	Repository      CartRepository
	Cache           CartCache
	ProductClient   ProductClient
	IDGenerator     IDGenerator
	Clock           Clock
	Logger          *slog.Logger
	CartTTL         time.Duration
	UserCartTTL     time.Duration
	GuestCartTTL    time.Duration
	MaxSaveAttempts int
}

type AddItemCommand struct {
	UserID         string
	GuestSessionID string
	ProductID      string
	VariantID      string
	Quantity       int
}

type ProductForCart struct {
	ID       string
	SellerID string
	Title    string
	ImageURL string
	Status   string
	Variants []ProductVariantForCart
}

type ProductVariantForCart struct {
	ID            string
	SKU           string
	Attributes    map[string]string
	PriceAmount   int64
	Currency      string
	StockQuantity int
	Available     bool
	Status        string
}

func NewAddItemUsecase(deps AddItemDependencies) (*AddItemUsecase, error) {
	if deps.Repository == nil {
		return nil, errors.New("cart repository is required")
	}
	if deps.Cache == nil {
		return nil, errors.New("cart cache is required")
	}
	if deps.ProductClient == nil {
		return nil, errors.New("product client is required")
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
	if deps.MaxSaveAttempts <= 0 {
		deps.MaxSaveAttempts = defaultAddItemSaveAttempts
	}
	return &AddItemUsecase{
		repository:      deps.Repository,
		cache:           deps.Cache,
		productClient:   deps.ProductClient,
		ids:             deps.IDGenerator,
		clock:           deps.Clock,
		logger:          deps.Logger,
		ttlPolicy:       ttlPolicy,
		maxSaveAttempts: deps.MaxSaveAttempts,
	}, nil
}

func (u *AddItemUsecase) AddItem(ctx context.Context, cmd AddItemCommand) (*domain.Cart, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cmd = normalizeAddItemCommand(cmd)
	if err := ValidateAddItemCommand(cmd); err != nil {
		return nil, err
	}
	owner, err := domain.ResolveCartOwner(cmd.UserID, cmd.GuestSessionID)
	if err != nil {
		return nil, err
	}

	product, err := u.productClient.GetProductForCart(ctx, cmd.ProductID)
	if err != nil {
		return nil, err
	}
	if err := validateProductForCart(product); err != nil {
		return nil, err
	}
	variant, err := FindSellableVariant(product, cmd.VariantID)
	if err != nil {
		return nil, err
	}
	snapshot, err := BuildProductVariantSnapshot(product, variant)
	if err != nil {
		return nil, err
	}

	var lastConflict error
	for attempt := 1; attempt <= u.maxSaveAttempts; attempt++ {
		cart, err := u.addItemAttempt(ctx, owner, snapshot, cmd.Quantity)
		if err == nil {
			u.refreshCache(ctx, owner, cart)
			return cart, nil
		}
		if !errors.Is(err, domain.ErrCartVersionConflict) {
			return nil, err
		}
		lastConflict = err
		u.logger.Warn(
			"cart.add_item.version_conflict",
			slog.String("owner_type", string(owner.Type)),
			slog.String("product_id", cmd.ProductID),
			slog.String("variant_id", cmd.VariantID),
			slog.Int("attempt", attempt),
		)
	}
	if lastConflict == nil {
		lastConflict = domain.ErrCartVersionConflict
	}
	return nil, lastConflict
}

func ValidateAddItemCommand(cmd AddItemCommand) error {
	if strings.TrimSpace(cmd.ProductID) == "" {
		return domain.ErrProductIDRequired
	}
	if strings.TrimSpace(cmd.VariantID) == "" {
		return domain.ErrVariantIDRequired
	}
	if cmd.Quantity < domain.MinItemQuantity {
		return domain.ErrQuantityTooSmall
	}
	if cmd.Quantity > domain.MaxItemQuantity {
		return domain.ErrQuantityTooLarge
	}
	return nil
}

func FindSellableVariant(product *ProductForCart, variantID string) (*ProductVariantForCart, error) {
	if product == nil {
		return nil, domain.ErrProductNotFound
	}
	variantID = strings.TrimSpace(variantID)
	for idx := range product.Variants {
		variant := &product.Variants[idx]
		if strings.TrimSpace(variant.ID) != variantID {
			continue
		}
		if !variantSellable(*variant) {
			return nil, domain.ErrVariantNotSellable
		}
		if strings.TrimSpace(variant.Currency) == "" {
			return nil, domain.ErrProductPriceInvalid
		}
		if err := domain.NewMoney(variant.PriceAmount, variant.Currency).Validate(); err != nil {
			return nil, fmt.Errorf("%w: %v", domain.ErrProductPriceInvalid, err)
		}
		return variant, nil
	}
	return nil, domain.ErrVariantNotFound
}

func BuildProductVariantSnapshot(product *ProductForCart, variant *ProductVariantForCart) (domain.ProductVariantSnapshot, error) {
	if product == nil {
		return domain.ProductVariantSnapshot{}, domain.ErrProductNotFound
	}
	if variant == nil {
		return domain.ProductVariantSnapshot{}, domain.ErrVariantNotFound
	}
	snapshot := domain.ProductVariantSnapshot{
		ProductID:     strings.TrimSpace(product.ID),
		VariantID:     strings.TrimSpace(variant.ID),
		SellerID:      strings.TrimSpace(product.SellerID),
		SKU:           strings.TrimSpace(variant.SKU),
		Title:         strings.TrimSpace(product.Title),
		ImageURL:      strings.TrimSpace(product.ImageURL),
		Attributes:    cloneAttributes(variant.Attributes),
		UnitPrice:     domain.NewMoney(variant.PriceAmount, variant.Currency),
		StockQuantity: variant.StockQuantity,
	}
	if err := snapshot.UnitPrice.Validate(); err != nil {
		return domain.ProductVariantSnapshot{}, fmt.Errorf("%w: %v", domain.ErrProductPriceInvalid, err)
	}
	return snapshot, nil
}

func (u *AddItemUsecase) addItemAttempt(ctx context.Context, owner domain.CartOwner, snapshot domain.ProductVariantSnapshot, qty int) (*domain.Cart, error) {
	now := u.clock.Now().UTC()
	expiresAt := now.Add(u.ttlPolicy.TTLForOwner(owner))
	cart, err := u.repository.FindActiveByOwner(ctx, owner)
	if errors.Is(err, domain.ErrCartNotFound) {
		cartID, idErr := u.ids.NewCartID()
		if idErr != nil {
			return nil, idErr
		}
		cart, err = domain.NewActiveCart(cartID, owner, snapshot.UnitPrice.Currency, now, expiresAt)
		if err != nil {
			return nil, err
		}
		if err := u.repository.CreateActiveCart(ctx, cart); err != nil {
			if errors.Is(err, domain.ErrActiveCartExists) {
				cart, err = u.repository.FindActiveByOwner(ctx, owner)
				if err != nil {
					return nil, err
				}
				if domain.IsActiveCartExpired(cart, now) {
					return nil, domain.ErrCartVersionConflict
				}
			} else {
				return nil, err
			}
		}
	} else if err != nil {
		return nil, err
	} else if domain.IsActiveCartExpired(cart, now) {
		if err := expireActiveCartForOwner(ctx, u.repository, u.cache, u.logger, owner, cart, now, "add_item"); err != nil {
			return nil, err
		}
		cartID, idErr := u.ids.NewCartID()
		if idErr != nil {
			return nil, idErr
		}
		cart, err = domain.NewActiveCart(cartID, owner, snapshot.UnitPrice.Currency, now, expiresAt)
		if err != nil {
			return nil, err
		}
		if err := u.repository.CreateActiveCart(ctx, cart); err != nil {
			if errors.Is(err, domain.ErrActiveCartExists) {
				return nil, domain.ErrCartVersionConflict
			}
			return nil, err
		}
	}

	expectedVersion := cart.Version
	itemID, err := u.ids.NewItemID()
	if err != nil {
		return nil, err
	}
	if err := cart.AddOrIncrementItem(snapshot, qty, now, itemID); err != nil {
		return nil, err
	}
	cart.ClearCouponPreview()
	if err := cart.RecalculateTotals(snapshot.UnitPrice.Currency); err != nil {
		return nil, err
	}
	cart.Touch(now, expiresAt)
	if err := u.repository.SaveWithVersion(ctx, cart, expectedVersion); err != nil {
		return nil, err
	}
	return cart, nil
}

func (u *AddItemUsecase) refreshCache(ctx context.Context, owner domain.CartOwner, cart *domain.Cart) {
	summary := domain.BuildCartSummary(cart)
	if err := u.cache.SetActiveCart(ctx, owner, cart); err != nil {
		u.logger.Warn("cart.add_item.cache_active_failed", slog.String("error", err.Error()), slog.String("cart_id", cart.ID))
	}
	if err := u.cache.SetCartSummary(ctx, owner, summary); err != nil {
		u.logger.Warn("cart.add_item.cache_summary_failed", slog.String("error", err.Error()), slog.String("cart_id", cart.ID))
	}
}

func normalizeAddItemCommand(cmd AddItemCommand) AddItemCommand {
	cmd.UserID = strings.TrimSpace(cmd.UserID)
	cmd.GuestSessionID = strings.TrimSpace(cmd.GuestSessionID)
	cmd.ProductID = strings.TrimSpace(cmd.ProductID)
	cmd.VariantID = strings.TrimSpace(cmd.VariantID)
	return cmd
}

func validateProductForCart(product *ProductForCart) error {
	if product == nil {
		return domain.ErrProductNotFound
	}
	if strings.TrimSpace(product.ID) == "" {
		return domain.ErrProductNotFound
	}
	if strings.TrimSpace(product.SellerID) == "" || strings.TrimSpace(product.Title) == "" {
		return domain.ErrProductNotSellable
	}
	if !productSellable(product.Status) {
		return domain.ErrProductNotSellable
	}
	return nil
}

func productSellable(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "published", "active", "sellable":
		return true
	default:
		return false
	}
}

func variantSellable(variant ProductVariantForCart) bool {
	if !variant.Available || variant.StockQuantity < 0 {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(variant.Status)) {
	case "", "active", "available", "enabled", "published":
		return true
	default:
		return false
	}
}

func cloneAttributes(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		output[key] = strings.TrimSpace(value)
	}
	if len(output) == 0 {
		return nil
	}
	return output
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}

type CryptoIDGenerator struct{}

func (CryptoIDGenerator) NewCartID() (string, error) {
	id, err := randomHexID("cart")
	if err != nil {
		return "", err
	}
	return id, nil
}

func (CryptoIDGenerator) NewItemID() (string, error) {
	id, err := randomHexID("item")
	if err != nil {
		return "", err
	}
	return id, nil
}

func randomHexID(prefix string) (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("generate %s id: %w", prefix, err)
	}
	return prefix + "_" + hex.EncodeToString(bytes[:]), nil
}
