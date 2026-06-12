package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/domain"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	DefaultCollectionName        = "wishlists"
	UniqueUserIDIndexName        = "uniq_wishlists_user_id"
	ItemsProductIDIndexName      = "idx_wishlists_items_product_id"
	ItemsProductVariantIndexName = "idx_wishlists_items_product_variant"
	PriceDropCandidateIndexName  = "idx_wishlists_price_drop_candidates"
	defaultPriceDropBatchSize    = int32(500)
	validationLevelStrict        = "strict"
	validationActionError        = "error"
	namespaceExistsErrorCode     = 48
	namespaceExistsErrorName     = "NamespaceExists"
	wishlistCollectionFieldName  = "name"
)

type MongoWishlistRepository struct {
	database       *mongo.Database
	collection     *mongo.Collection
	collectionName string
	idFactory      func() string
	priceDropBatch int32
	logger         *slog.Logger
}

func NewMongoWishlistRepository(database *mongo.Database, collectionName string, logger *slog.Logger) (*MongoWishlistRepository, error) {
	if database == nil {
		return nil, errors.New("wishlist mongo database is required")
	}
	collectionName = strings.TrimSpace(collectionName)
	if collectionName == "" {
		collectionName = DefaultCollectionName
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &MongoWishlistRepository{
		database:       database,
		collection:     database.Collection(collectionName),
		collectionName: collectionName,
		idFactory:      defaultWishlistID,
		priceDropBatch: defaultPriceDropBatchSize,
		logger:         logger,
	}, nil
}

func (r *MongoWishlistRepository) CollectionName() string {
	if r == nil {
		return ""
	}
	return r.collectionName
}

func (r *MongoWishlistRepository) SetPriceDropBatchSize(batchSize int) {
	if r == nil || batchSize <= 0 {
		return
	}
	r.priceDropBatch = int32(batchSize)
}

func (r *MongoWishlistRepository) Ping(ctx context.Context) error {
	if r == nil || r.database == nil {
		return errors.New("wishlist repository is not initialized")
	}
	ctx = contextOrBackground(ctx)
	if err := r.database.RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		return fmt.Errorf("ping wishlist mongo database: %w", err)
	}
	return nil
}

func (r *MongoWishlistRepository) EnsureCollection(ctx context.Context) error {
	if r == nil || r.database == nil || r.collection == nil {
		return errors.New("wishlist repository is not initialized")
	}
	ctx = contextOrBackground(ctx)

	if err := r.ensureCollectionValidator(ctx); err != nil {
		return err
	}
	if err := r.ensureIndexes(ctx); err != nil {
		return err
	}
	return nil
}

func (r *MongoWishlistRepository) EnsureWishlist(ctx context.Context, userID string, now time.Time) error {
	if r == nil || r.collection == nil {
		return errors.New("wishlist repository is not initialized")
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return fmt.Errorf("%w: user_id is required", domain.ErrInvalidWishlist)
	}
	now = normalizeTime(now)
	wishlistID := r.idFactory()
	if strings.TrimSpace(wishlistID) == "" {
		return errors.New("wishlist id factory returned an empty id")
	}

	_, err := r.collection.UpdateOne(
		contextOrBackground(ctx),
		bson.M{"user_id": userID},
		bson.M{"$setOnInsert": bson.M{
			"_id":        wishlistID,
			"user_id":    userID,
			"visibility": domain.VisibilityPrivate,
			"items":      bson.A{},
			"created_at": now,
			"updated_at": now,
		}},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil
		}
		return fmt.Errorf("ensure wishlist for user %q: %w", userID, err)
	}
	return nil
}

func (r *MongoWishlistRepository) AddItemIfNotExists(ctx context.Context, userID string, item domain.WishlistItem, now time.Time) error {
	if r == nil || r.collection == nil {
		return errors.New("wishlist repository is not initialized")
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return fmt.Errorf("%w: user_id is required", domain.ErrInvalidWishlist)
	}
	if err := item.Validate(); err != nil {
		return err
	}
	now = normalizeTime(now)
	itemDocument := newWishlistItemDocument(item)

	result, err := r.collection.UpdateOne(
		contextOrBackground(ctx),
		bson.M{
			"user_id": userID,
			"items.product_id": bson.M{
				"$ne": item.ProductID,
			},
		},
		bson.M{
			"$push": bson.M{
				"items": itemDocument,
			},
			"$set": bson.M{
				"updated_at": now,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("add wishlist item for user %q product %q: %w", userID, item.ProductID, err)
	}
	if result.ModifiedCount == 0 {
		return fmt.Errorf("%w: %s", domain.ErrDuplicateProduct, item.ProductID)
	}
	return nil
}

func (r *MongoWishlistRepository) RemoveItem(ctx context.Context, userID, productID string, now time.Time) error {
	if r == nil || r.collection == nil {
		return errors.New("wishlist repository is not initialized")
	}
	userID = strings.TrimSpace(userID)
	productID = strings.TrimSpace(productID)
	if userID == "" {
		return fmt.Errorf("%w: user_id is required", domain.ErrInvalidWishlist)
	}
	if productID == "" {
		return fmt.Errorf("%w: product_id is required", domain.ErrInvalidWishlist)
	}
	now = normalizeTime(now)

	_, err := r.collection.UpdateOne(
		contextOrBackground(ctx),
		bson.M{
			"user_id":          userID,
			"items.product_id": productID,
		},
		bson.M{
			"$pull": bson.M{
				"items": bson.M{
					"product_id": productID,
				},
			},
			"$set": bson.M{
				"updated_at": now,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("remove wishlist item for user %q product %q: %w", userID, productID, err)
	}
	return nil
}

func (r *MongoWishlistRepository) FindByUserID(ctx context.Context, userID string) (*domain.Wishlist, error) {
	if r == nil || r.collection == nil {
		return nil, errors.New("wishlist repository is not initialized")
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("%w: user_id is required", domain.ErrInvalidWishlist)
	}

	var document WishlistDocument
	err := r.collection.FindOne(contextOrBackground(ctx), bson.M{"user_id": userID}).Decode(&document)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("%w: %s", domain.ErrWishlistNotFound, userID)
		}
		return nil, fmt.Errorf("find wishlist by user %q: %w", userID, err)
	}
	wishlist, err := document.ToDomain()
	if err != nil {
		return nil, fmt.Errorf("decode wishlist for user %q: %w", userID, err)
	}
	return wishlist, nil
}

func (r *MongoWishlistRepository) UpdateProductAvailability(ctx context.Context, productID, variantID string, availability domain.Availability, updatedAt time.Time) (domain.AvailabilityUpdateResult, error) {
	if r == nil || r.collection == nil {
		return domain.AvailabilityUpdateResult{}, errors.New("wishlist repository is not initialized")
	}

	filter, update, arrayFilters, err := availabilityUpdateDocuments(productID, variantID, availability, updatedAt)
	if err != nil {
		return domain.AvailabilityUpdateResult{}, err
	}

	result, err := r.collection.UpdateMany(
		contextOrBackground(ctx),
		filter,
		update,
		options.UpdateMany().SetArrayFilters(arrayFilters),
	)
	if err != nil {
		return domain.AvailabilityUpdateResult{}, fmt.Errorf("update wishlist availability for product %q variant %q: %w", strings.TrimSpace(productID), strings.TrimSpace(variantID), err)
	}

	return domain.AvailabilityUpdateResult{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
	}, nil
}

func (r *MongoWishlistRepository) StreamPriceDropCandidates(
	ctx context.Context,
	productID string,
	variantID string,
	newPrice domain.Money,
	handle func(domain.PriceDropCandidate) error,
) error {
	if r == nil || r.collection == nil {
		return errors.New("wishlist repository is not initialized")
	}
	if handle == nil {
		return errors.New("price drop candidate handler is required")
	}

	filter, projection, err := priceDropCandidateDocuments(productID, variantID, newPrice)
	if err != nil {
		return err
	}
	cursor, err := r.collection.Find(
		contextOrBackground(ctx),
		filter,
		options.Find().
			SetProjection(projection).
			SetBatchSize(r.priceDropBatch),
	)
	if err != nil {
		return fmt.Errorf("find wishlist price-drop candidates for product %q variant %q: %w", strings.TrimSpace(productID), strings.TrimSpace(variantID), err)
	}
	defer cursor.Close(contextOrBackground(ctx))

	for cursor.Next(contextOrBackground(ctx)) {
		var document WishlistDocument
		if err := cursor.Decode(&document); err != nil {
			return fmt.Errorf("decode wishlist price-drop candidate: %w", err)
		}
		for _, item := range document.Items {
			if item.LastKnownPrice == nil {
				continue
			}
			candidate := domain.PriceDropCandidate{
				UserID:    document.UserID,
				ProductID: item.ProductID,
				VariantID: item.VariantID,
				PreviousPrice: domain.Money{
					Amount:   item.LastKnownPrice.Amount,
					Currency: item.LastKnownPrice.Currency,
				},
			}
			if err := handle(candidate); err != nil {
				return err
			}
		}
	}
	if err := cursor.Err(); err != nil {
		return fmt.Errorf("iterate wishlist price-drop candidates: %w", err)
	}
	return nil
}

func (r *MongoWishlistRepository) UpdateLastKnownPriceForProduct(
	ctx context.Context,
	productID string,
	variantID string,
	newPrice domain.Money,
	updatedAt time.Time,
) (domain.PriceUpdateResult, error) {
	if r == nil || r.collection == nil {
		return domain.PriceUpdateResult{}, errors.New("wishlist repository is not initialized")
	}

	filter, update, arrayFilters, err := priceSnapshotUpdateDocuments(productID, variantID, newPrice, updatedAt)
	if err != nil {
		return domain.PriceUpdateResult{}, err
	}

	result, err := r.collection.UpdateMany(
		contextOrBackground(ctx),
		filter,
		update,
		options.UpdateMany().SetArrayFilters(arrayFilters),
	)
	if err != nil {
		return domain.PriceUpdateResult{}, fmt.Errorf("update wishlist price snapshot for product %q variant %q: %w", strings.TrimSpace(productID), strings.TrimSpace(variantID), err)
	}

	return domain.PriceUpdateResult{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
	}, nil
}

func availabilityUpdateDocuments(productID, variantID string, availability domain.Availability, updatedAt time.Time) (bson.M, bson.M, []any, error) {
	productID = strings.TrimSpace(productID)
	variantID = strings.TrimSpace(variantID)
	availability = availability.Normalized()
	if productID == "" {
		return nil, nil, nil, fmt.Errorf("%w: product_id is required", domain.ErrInvalidWishlist)
	}
	if !availability.IsValid() {
		return nil, nil, nil, fmt.Errorf("%w: invalid availability", domain.ErrInvalidWishlist)
	}

	updatedAt = normalizeTime(updatedAt)
	itemFilter := bson.M{
		"item.product_id": productID,
		"item.added_at": bson.M{
			"$lte": updatedAt,
		},
	}
	elemMatch := bson.M{
		"product_id": productID,
		"added_at": bson.M{
			"$lte": updatedAt,
		},
	}
	if variantID != "" {
		itemFilter["item.variant_id"] = variantID
		elemMatch["variant_id"] = variantID
	}
	filter := bson.M{
		"items": bson.M{
			"$elemMatch": elemMatch,
		},
	}

	update := bson.M{
		"$set": bson.M{
			"items.$[item].availability": availability,
			"updated_at":                 updatedAt,
		},
	}
	return filter, update, []any{itemFilter}, nil
}

func priceDropCandidateDocuments(productID, variantID string, newPrice domain.Money) (bson.M, bson.M, error) {
	productID = strings.TrimSpace(productID)
	variantID = strings.TrimSpace(variantID)
	newPrice.Currency = strings.ToUpper(strings.TrimSpace(newPrice.Currency))
	if productID == "" {
		return nil, nil, fmt.Errorf("%w: product_id is required", domain.ErrInvalidWishlist)
	}
	if err := newPrice.Validate(); err != nil {
		return nil, nil, err
	}

	elemMatch := bson.M{
		"product_id":                productID,
		"last_known_price.currency": newPrice.Currency,
		"last_known_price.amount":   bson.M{"$gt": newPrice.Amount},
		"availability":              bson.M{"$ne": domain.AvailabilityDeleted},
	}
	if variantID != "" {
		elemMatch["variant_id"] = variantID
	}

	filter := bson.M{
		"items": bson.M{
			"$elemMatch": elemMatch,
		},
	}
	projection := bson.M{
		"user_id": 1,
		"items": bson.M{
			"$elemMatch": elemMatch,
		},
	}
	return filter, projection, nil
}

func priceSnapshotUpdateDocuments(productID, variantID string, newPrice domain.Money, updatedAt time.Time) (bson.M, bson.M, []any, error) {
	productID = strings.TrimSpace(productID)
	variantID = strings.TrimSpace(variantID)
	newPrice.Currency = strings.ToUpper(strings.TrimSpace(newPrice.Currency))
	if productID == "" {
		return nil, nil, nil, fmt.Errorf("%w: product_id is required", domain.ErrInvalidWishlist)
	}
	if err := newPrice.Validate(); err != nil {
		return nil, nil, nil, err
	}

	updatedAt = normalizeTime(updatedAt)
	itemFilter := bson.M{
		"item.product_id": productID,
	}
	elemMatch := bson.M{
		"product_id": productID,
	}
	if variantID != "" {
		itemFilter["item.variant_id"] = variantID
		elemMatch["variant_id"] = variantID
	}
	filter := bson.M{
		"items": bson.M{
			"$elemMatch": elemMatch,
		},
	}
	update := bson.M{
		"$set": bson.M{
			"items.$[item].last_known_price": bson.M{
				"amount":   newPrice.Amount,
				"currency": newPrice.Currency,
			},
			"updated_at": updatedAt,
		},
	}
	return filter, update, []any{itemFilter}, nil
}

func (r *MongoWishlistRepository) ensureCollectionValidator(ctx context.Context) error {
	exists, err := r.collectionExists(ctx)
	if err != nil {
		return err
	}
	if !exists {
		createOptions := options.CreateCollection().
			SetValidator(WishlistCollectionValidator()).
			SetValidationLevel(validationLevelStrict).
			SetValidationAction(validationActionError)

		err = r.database.CreateCollection(ctx, r.collectionName, createOptions)
		if err == nil {
			r.logger.Info("wishlist collection created", "collection", r.collectionName)
			return nil
		}
		if !isNamespaceExists(err) {
			return fmt.Errorf("create wishlist collection %q: %w", r.collectionName, err)
		}
	}

	if err := r.applyCollectionValidator(ctx); err != nil {
		return err
	}
	r.logger.Info("wishlist collection validator ensured", "collection", r.collectionName)
	return nil
}

func (r *MongoWishlistRepository) collectionExists(ctx context.Context) (bool, error) {
	names, err := r.database.ListCollectionNames(ctx, bson.D{{Key: wishlistCollectionFieldName, Value: r.collectionName}})
	if err != nil {
		return false, fmt.Errorf("list wishlist collections: %w", err)
	}
	return len(names) > 0, nil
}

func (r *MongoWishlistRepository) applyCollectionValidator(ctx context.Context) error {
	command := bson.D{
		{Key: "collMod", Value: r.collectionName},
		{Key: "validator", Value: WishlistCollectionValidator()},
		{Key: "validationLevel", Value: validationLevelStrict},
		{Key: "validationAction", Value: validationActionError},
	}
	if err := r.database.RunCommand(ctx, command).Err(); err != nil {
		return fmt.Errorf("update wishlist collection validator %q: %w", r.collectionName, err)
	}
	return nil
}

func (r *MongoWishlistRepository) ensureIndexes(ctx context.Context) error {
	names, err := r.collection.Indexes().CreateMany(ctx, WishlistIndexModels())
	if err != nil {
		return fmt.Errorf("ensure wishlist indexes: %w", err)
	}
	r.logger.Info("wishlist indexes ensured", "collection", r.collectionName, "indexes", names)
	return nil
}

func WishlistIndexModels() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().
				SetName(UniqueUserIDIndexName).
				SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "items.product_id", Value: 1}},
			Options: options.Index().
				SetName(ItemsProductIDIndexName),
		},
		{
			Keys: bson.D{
				{Key: "items.product_id", Value: 1},
				{Key: "items.variant_id", Value: 1},
			},
			Options: options.Index().
				SetName(ItemsProductVariantIndexName),
		},
		{
			Keys: bson.D{
				{Key: "items.product_id", Value: 1},
				{Key: "items.variant_id", Value: 1},
				{Key: "items.last_known_price.currency", Value: 1},
				{Key: "items.last_known_price.amount", Value: 1},
				{Key: "items.availability", Value: 1},
			},
			Options: options.Index().
				SetName(PriceDropCandidateIndexName),
		},
	}
}

func WishlistCollectionValidator() bson.D {
	return bson.D{
		{Key: "$jsonSchema", Value: bson.D{
			{Key: "bsonType", Value: "object"},
			{Key: "title", Value: "Wishlist"},
			{Key: "required", Value: bson.A{"_id", "user_id", "visibility", "items", "created_at", "updated_at"}},
			{Key: "properties", Value: bson.D{
				{Key: "_id", Value: bson.D{
					{Key: "bsonType", Value: "string"},
					{Key: "description", Value: "Wishlist primary id, exposed as wishlist_id in APIs"},
				}},
				{Key: "user_id", Value: bson.D{
					{Key: "bsonType", Value: "string"},
					{Key: "description", Value: "Authenticated buyer user id"},
				}},
				{Key: "visibility", Value: bson.D{
					{Key: "enum", Value: bson.A{"private"}},
					{Key: "description", Value: "MVP supports private wishlists only"},
				}},
				{Key: "items", Value: bson.D{
					{Key: "bsonType", Value: "array"},
					{Key: "description", Value: "Embedded wishlist items"},
					{Key: "items", Value: bson.D{
						{Key: "bsonType", Value: "object"},
						{Key: "required", Value: bson.A{"product_id", "added_at"}},
						{Key: "properties", Value: bson.D{
							{Key: "product_id", Value: bson.D{
								{Key: "bsonType", Value: "string"},
								{Key: "description", Value: "Saved product id"},
							}},
							{Key: "variant_id", Value: bson.D{
								{Key: "bsonType", Value: "string"},
								{Key: "description", Value: "Optional selected product variant id"},
							}},
							{Key: "added_at", Value: bson.D{
								{Key: "bsonType", Value: "date"},
								{Key: "description", Value: "When product was added to wishlist"},
							}},
							{Key: "last_known_price", Value: bson.D{
								{Key: "bsonType", Value: "object"},
								{Key: "required", Value: bson.A{"amount", "currency"}},
								{Key: "properties", Value: bson.D{
									{Key: "amount", Value: bson.D{
										{Key: "bsonType", Value: "long"},
										{Key: "description", Value: "Price in minor unit"},
									}},
									{Key: "currency", Value: bson.D{
										{Key: "bsonType", Value: "string"},
										{Key: "pattern", Value: "^[A-Z]{3}$"},
										{Key: "description", Value: "3-letter currency code"},
									}},
								}},
							}},
							{Key: "availability", Value: bson.D{
								{Key: "enum", Value: bson.A{"unknown", "in_stock", "out_of_stock", "deleted"}},
								{Key: "description", Value: "Product availability snapshot"},
							}},
						}},
					}},
				}},
				{Key: "created_at", Value: bson.D{
					{Key: "bsonType", Value: "date"},
					{Key: "description", Value: "Wishlist creation timestamp"},
				}},
				{Key: "updated_at", Value: bson.D{
					{Key: "bsonType", Value: "date"},
					{Key: "description", Value: "Wishlist last update timestamp"},
				}},
			}},
		}},
	}
}

func isNamespaceExists(err error) bool {
	var commandError mongo.CommandError
	if errors.As(err, &commandError) {
		return commandError.Code == namespaceExistsErrorCode || commandError.Name == namespaceExistsErrorName
	}
	return strings.Contains(err.Error(), namespaceExistsErrorName)
}

func contextOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func normalizeTime(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}
	return value.UTC()
}

func defaultWishlistID() string {
	var randomBytes [12]byte
	if _, err := rand.Read(randomBytes[:]); err == nil {
		return "wish_" + hex.EncodeToString(randomBytes[:])
	}
	return fmt.Sprintf("wish_%d", time.Now().UTC().UnixNano())
}
