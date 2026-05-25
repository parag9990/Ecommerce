package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoCartRepository struct {
	collection *mongo.Collection
	logger     *slog.Logger
}

func NewMongoCartRepository(db *mongo.Database, logger *slog.Logger) (*MongoCartRepository, error) {
	if db == nil {
		return nil, errors.New("mongo database is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &MongoCartRepository{
		collection: db.Collection(domain.CartCollectionName),
		logger:     logger,
	}, nil
}

func (r *MongoCartRepository) FindCartByID(ctx context.Context, cartID string) (*domain.Cart, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cartID = strings.TrimSpace(cartID)
	if cartID == "" {
		return nil, domain.ErrGuestCartIDRequired
	}
	var cart domain.Cart
	if err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: cartID}}).Decode(&cart); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrCartNotFound
		}
		return nil, fmt.Errorf("%w: find cart by id: %v", domain.ErrCartStoreUnavailable, err)
	}
	return &cart, nil
}

func (r *MongoCartRepository) FindActiveByOwner(ctx context.Context, owner domain.CartOwner) (*domain.Cart, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := owner.Validate(); err != nil {
		return nil, err
	}
	var cart domain.Cart
	if err := r.collection.FindOne(ctx, activeOwnerFilter(owner)).Decode(&cart); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrCartNotFound
		}
		return nil, fmt.Errorf("%w: find active cart: %v", domain.ErrCartStoreUnavailable, err)
	}
	return &cart, nil
}

func (r *MongoCartRepository) FindActiveGuestCartByID(ctx context.Context, cartID string, guestSessionID string) (*domain.Cart, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cartID = strings.TrimSpace(cartID)
	guestSessionID = strings.TrimSpace(guestSessionID)
	if cartID == "" {
		return nil, domain.ErrGuestCartIDRequired
	}
	if guestSessionID == "" {
		return nil, domain.ErrGuestSessionIDRequired
	}
	filter := bson.D{
		{Key: "_id", Value: cartID},
		{Key: "status", Value: string(domain.CartStatusActive)},
		{Key: "guest_session_id", Value: guestSessionID},
	}
	var cart domain.Cart
	if err := r.collection.FindOne(ctx, filter).Decode(&cart); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrCartNotFound
		}
		return nil, fmt.Errorf("%w: find active guest cart: %v", domain.ErrCartStoreUnavailable, err)
	}
	return &cart, nil
}

func (r *MongoCartRepository) CreateActiveCart(ctx context.Context, cart *domain.Cart) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := r.insertActiveCart(ctx, cart); err != nil {
		return err
	}
	r.logger.Info("cart.mongo.active_created", slog.String("cart_id", cart.ID))
	return nil
}

func (r *MongoCartRepository) SaveWithVersion(ctx context.Context, cart *domain.Cart, expectedVersion int64) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if cart == nil {
		return fmt.Errorf("%w: cart is nil", domain.ErrInvalidCart)
	}
	if expectedVersion < 1 {
		return fmt.Errorf("%w: expected version must be greater than zero", domain.ErrInvalidCart)
	}
	if cart.Status != domain.CartStatusActive {
		return domain.ErrCartNotActive
	}
	if err := r.saveActiveCartWithVersion(ctx, cart, expectedVersion); err != nil {
		return err
	}
	cart.Version = expectedVersion + 1
	return nil
}

func (r *MongoCartRepository) FindExpiredActiveCarts(ctx context.Context, now time.Time, limit int64) ([]domain.ExpiredCartCandidate, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if limit <= 0 {
		return nil, fmt.Errorf("%w: expiry scan limit must be greater than zero", domain.ErrInvalidCart)
	}
	now = now.UTC()
	filter := bson.D{
		{Key: "status", Value: string(domain.CartStatusActive)},
		{Key: "expires_at", Value: bson.D{{Key: "$lte", Value: now}}},
	}
	findOptions := options.Find().
		SetLimit(limit).
		SetSort(bson.D{{Key: "expires_at", Value: 1}, {Key: "_id", Value: 1}}).
		SetProjection(bson.D{
			{Key: "_id", Value: 1},
			{Key: "user_id", Value: 1},
			{Key: "guest_session_id", Value: 1},
			{Key: "version", Value: 1},
			{Key: "expires_at", Value: 1},
		})

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("%w: find expired active carts: %v", domain.ErrCartStoreUnavailable, err)
	}
	defer cursor.Close(ctx)

	candidates := make([]domain.ExpiredCartCandidate, 0)
	for cursor.Next(ctx) {
		var candidate domain.ExpiredCartCandidate
		if err := cursor.Decode(&candidate); err != nil {
			return nil, fmt.Errorf("%w: decode expired active cart: %v", domain.ErrCartStoreUnavailable, err)
		}
		candidates = append(candidates, candidate)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("%w: iterate expired active carts: %v", domain.ErrCartStoreUnavailable, err)
	}
	return candidates, nil
}

func (r *MongoCartRepository) MarkCartsExpired(ctx context.Context, cartIDs []string, now time.Time) (int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ids := normalizeCartIDs(cartIDs)
	if len(ids) == 0 {
		return 0, nil
	}
	now = now.UTC()
	filter := bson.D{
		{Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}},
		{Key: "status", Value: string(domain.CartStatusActive)},
		{Key: "expires_at", Value: bson.D{{Key: "$lte", Value: now}}},
	}
	update := expireCartUpdate(now)
	result, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("%w: mark carts expired: %v", domain.ErrCartStoreUnavailable, err)
	}
	return result.ModifiedCount, nil
}

func (r *MongoCartRepository) ExpireActiveCart(ctx context.Context, cartID string, expectedVersion int64, now time.Time) (bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cartID = strings.TrimSpace(cartID)
	if cartID == "" {
		return false, fmt.Errorf("%w: cart id is required", domain.ErrInvalidCart)
	}
	if expectedVersion < 1 {
		return false, fmt.Errorf("%w: expected version must be greater than zero", domain.ErrInvalidCart)
	}
	now = now.UTC()
	filter := bson.D{
		{Key: "_id", Value: cartID},
		{Key: "status", Value: string(domain.CartStatusActive)},
		{Key: "version", Value: expectedVersion},
		{Key: "expires_at", Value: bson.D{{Key: "$lte", Value: now}}},
	}
	result, err := r.collection.UpdateOne(ctx, filter, expireCartUpdate(now))
	if err != nil {
		return false, fmt.Errorf("%w: expire active cart: %v", domain.ErrCartStoreUnavailable, err)
	}
	return result.ModifiedCount > 0, nil
}

func (r *MongoCartRepository) MergeGuestIntoUser(ctx context.Context, userCart *domain.Cart, guestCart *domain.Cart, expectedUserVersion int64, expectedGuestVersion int64) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if userCart == nil || guestCart == nil {
		return fmt.Errorf("%w: user and guest carts are required", domain.ErrInvalidCart)
	}
	if expectedUserVersion < 0 {
		return fmt.Errorf("%w: expected user version cannot be negative", domain.ErrInvalidCart)
	}
	if expectedGuestVersion < 1 {
		return fmt.Errorf("%w: expected guest version must be greater than zero", domain.ErrInvalidCart)
	}
	if userCart.Status != domain.CartStatusActive {
		return domain.ErrCartNotActive
	}
	if guestCart.Status != domain.CartStatusMerged {
		return domain.ErrCartNotActive
	}
	if guestCart.MergedIntoCartID == nil || strings.TrimSpace(*guestCart.MergedIntoCartID) != userCart.ID {
		return fmt.Errorf("%w: guest cart merge target does not match user cart", domain.ErrInvalidCart)
	}
	if err := userCart.Validate(); err != nil {
		return err
	}

	session, err := r.collection.Database().Client().StartSession()
	if err != nil {
		return fmt.Errorf("%w: start merge transaction: %v", domain.ErrCartStoreUnavailable, err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		if expectedUserVersion == 0 {
			if err := r.insertActiveCart(sc, userCart); err != nil {
				return nil, err
			}
		} else if err := r.saveActiveCartWithVersion(sc, userCart, expectedUserVersion); err != nil {
			return nil, err
		}
		if err := r.markGuestCartMergedWithVersion(sc, guestCart, expectedGuestVersion); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	if expectedUserVersion > 0 {
		userCart.Version = expectedUserVersion + 1
	}
	guestCart.Version = expectedGuestVersion + 1
	r.logger.Info(
		"cart.mongo.merge_committed",
		slog.String("guest_cart_id", guestCart.ID),
		slog.String("target_cart_id", userCart.ID),
	)
	return nil
}

func (r *MongoCartRepository) insertActiveCart(ctx context.Context, cart *domain.Cart) error {
	if cart == nil {
		return fmt.Errorf("%w: cart is nil", domain.ErrInvalidCart)
	}
	if err := cart.Validate(); err != nil {
		return err
	}
	if _, err := r.collection.InsertOne(ctx, cart); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrActiveCartExists
		}
		return fmt.Errorf("%w: create active cart: %v", domain.ErrCartStoreUnavailable, err)
	}
	return nil
}

func (r *MongoCartRepository) saveActiveCartWithVersion(ctx context.Context, cart *domain.Cart, expectedVersion int64) error {
	if err := cart.Validate(); err != nil {
		return err
	}
	filter := bson.D{
		{Key: "_id", Value: cart.ID},
		{Key: "status", Value: string(domain.CartStatusActive)},
		{Key: "version", Value: expectedVersion},
	}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "items", Value: cart.Items},
			{Key: "coupon_code", Value: cart.CouponCode},
			{Key: "coupon_preview", Value: cart.CouponPreview},
			{Key: "totals", Value: cart.Totals},
			{Key: "updated_at", Value: cart.UpdatedAt},
			{Key: "expires_at", Value: cart.ExpiresAt},
		}},
		{Key: "$inc", Value: bson.D{{Key: "version", Value: 1}}},
	}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("%w: save cart: %v", domain.ErrCartStoreUnavailable, err)
	}
	if result.MatchedCount == 0 {
		return domain.ErrCartVersionConflict
	}
	return nil
}

func (r *MongoCartRepository) markGuestCartMergedWithVersion(ctx context.Context, guestCart *domain.Cart, expectedVersion int64) error {
	filter := bson.D{
		{Key: "_id", Value: guestCart.ID},
		{Key: "status", Value: string(domain.CartStatusActive)},
		{Key: "version", Value: expectedVersion},
	}
	if guestCart.GuestSessionID != nil && strings.TrimSpace(*guestCart.GuestSessionID) != "" {
		filter = append(filter, bson.E{Key: "guest_session_id", Value: strings.TrimSpace(*guestCart.GuestSessionID)})
	}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "status", Value: string(domain.CartStatusMerged)},
			{Key: "merged_into_cart_id", Value: guestCart.MergedIntoCartID},
			{Key: "updated_at", Value: guestCart.UpdatedAt},
		}},
		{Key: "$inc", Value: bson.D{{Key: "version", Value: 1}}},
	}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("%w: mark guest cart merged: %v", domain.ErrCartStoreUnavailable, err)
	}
	if result.MatchedCount == 0 {
		return domain.ErrCartVersionConflict
	}
	return nil
}

func activeOwnerFilter(owner domain.CartOwner) bson.D {
	filter := bson.D{{Key: "status", Value: string(domain.CartStatusActive)}}
	if owner.IsUser() {
		return append(filter, bson.E{Key: "user_id", Value: owner.UserID})
	}
	return append(filter, bson.E{Key: "guest_session_id", Value: owner.GuestSessionID})
}

func expireCartUpdate(now time.Time) bson.D {
	return bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "status", Value: string(domain.CartStatusExpired)},
			{Key: "updated_at", Value: now.UTC()},
		}},
		{Key: "$inc", Value: bson.D{{Key: "version", Value: 1}}},
	}
}

func normalizeCartIDs(cartIDs []string) []string {
	if len(cartIDs) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(cartIDs))
	ids := make([]string, 0, len(cartIDs))
	for _, cartID := range cartIDs {
		cartID = strings.TrimSpace(cartID)
		if cartID == "" {
			continue
		}
		if _, ok := seen[cartID]; ok {
			continue
		}
		seen[cartID] = struct{}{}
		ids = append(ids, cartID)
	}
	return ids
}
