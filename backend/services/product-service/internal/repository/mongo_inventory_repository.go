package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"product-service/internal/domain"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (r *MongoProductRepository) ReserveInventoryAtomic(
	ctx context.Context,
	reservation *domain.InventoryReservation,
) ([]InventoryStockMutation, error) {
	if reservation == nil {
		return nil, fmt.Errorf("inventory reservation is required")
	}
	mutations := make([]InventoryStockMutation, 0, len(reservation.Items))
	err := r.withTransaction(ctx, func(tx context.Context) error {
		for index := range reservation.Items {
			item := &reservation.Items[index]
			mutation, err := r.ReserveVariant(tx, item.ProductID, item.VariantID, item.Quantity, reservation.CreatedAt)
			if err != nil {
				return err
			}
			item.SKU = mutation.SKU
			item.SellerID = mutation.SellerID
			mutations = append(mutations, mutation)
		}
		if report := reservation.Validate(); report.HasErrors() {
			return domain.ValidationError{Report: report}
		}
		return r.CreateInventoryReservation(tx, reservation)
	})
	if err != nil {
		return nil, err
	}
	return mutations, nil
}

func (r *MongoProductRepository) FinalizeInventoryReservationAtomic(
	ctx context.Context,
	reservation domain.InventoryReservation,
	terminalStatus domain.InventoryReservationStatus,
	reason string,
	at time.Time,
) ([]InventoryStockMutation, error) {
	mutations := make([]InventoryStockMutation, 0, len(reservation.Items))
	err := r.withTransaction(ctx, func(tx context.Context) error {
		for _, item := range reservation.Items {
			var mutation InventoryStockMutation
			var err error
			switch terminalStatus {
			case domain.ReservationStatusCommitted:
				mutation, err = r.CommitVariant(tx, item.ProductID, item.VariantID, item.Quantity, at)
			case domain.ReservationStatusReleased, domain.ReservationStatusExpired:
				mutation, err = r.ReleaseVariant(tx, item.ProductID, item.VariantID, item.Quantity, at)
			default:
				return fmt.Errorf("unsupported terminal reservation status %q", terminalStatus)
			}
			if err != nil {
				return err
			}
			mutations = append(mutations, mutation)
		}
		switch terminalStatus {
		case domain.ReservationStatusCommitted:
			return r.MarkReservationCommitted(tx, reservation.ID, at)
		case domain.ReservationStatusReleased:
			return r.MarkReservationReleased(tx, reservation.ID, reason, at)
		case domain.ReservationStatusExpired:
			return r.MarkReservationExpired(tx, reservation.ID, reason, at)
		default:
			return fmt.Errorf("unsupported terminal reservation status %q", terminalStatus)
		}
	})
	if err != nil {
		return nil, err
	}
	return mutations, nil
}

func (r *MongoProductRepository) ReserveVariant(
	ctx context.Context,
	productID string,
	variantID string,
	quantity int64,
	now time.Time,
) (InventoryStockMutation, error) {
	if err := ctx.Err(); err != nil {
		return InventoryStockMutation{}, err
	}
	productID = strings.TrimSpace(productID)
	variantID = strings.TrimSpace(variantID)
	if productID == "" || variantID == "" || quantity <= 0 {
		return InventoryStockMutation{}, ErrWriteConflict
	}
	previous, err := r.previousInventoryState(ctx, productID, variantID)
	if err != nil {
		return InventoryStockMutation{}, err
	}

	result, err := r.products.UpdateOne(
		ctx,
		reserveVariantFilter(productID, variantID, quantity),
		inventoryUpdate(
			bson.D{e("variants.$.reserved_quantity", quantity)},
			now,
		),
	)
	if err != nil {
		return InventoryStockMutation{}, fmt.Errorf("reserve inventory product %q variant %q: %w", productID, variantID, err)
	}
	if result.MatchedCount == 0 {
		return InventoryStockMutation{}, r.classifyReserveFailure(ctx, productID, variantID, quantity)
	}
	mutation, err := r.currentInventoryMutation(ctx, productID, variantID)
	if err != nil {
		return InventoryStockMutation{}, err
	}
	return mutationWithPreviousAvailability(mutation, previous), nil
}

func (r *MongoProductRepository) ReleaseVariant(
	ctx context.Context,
	productID string,
	variantID string,
	quantity int64,
	now time.Time,
) (InventoryStockMutation, error) {
	if err := ctx.Err(); err != nil {
		return InventoryStockMutation{}, err
	}
	productID = strings.TrimSpace(productID)
	variantID = strings.TrimSpace(variantID)
	if productID == "" || variantID == "" || quantity <= 0 {
		return InventoryStockMutation{}, ErrWriteConflict
	}
	previous, err := r.previousInventoryState(ctx, productID, variantID)
	if err != nil {
		return InventoryStockMutation{}, err
	}

	result, err := r.products.UpdateOne(
		ctx,
		bson.D{
			e("_id", productID),
			e("variants", bson.D{e("$elemMatch", bson.D{
				e("variant_id", variantID),
				e("reserved_quantity", bson.D{e("$gte", quantity)}),
			})}),
		},
		inventoryUpdate(
			bson.D{e("variants.$.reserved_quantity", -quantity)},
			now,
		),
	)
	if err != nil {
		return InventoryStockMutation{}, fmt.Errorf("release inventory product %q variant %q: %w", productID, variantID, err)
	}
	if result.MatchedCount == 0 {
		return InventoryStockMutation{}, r.classifyReleaseFailure(ctx, productID, variantID, quantity)
	}
	mutation, err := r.currentInventoryMutation(ctx, productID, variantID)
	if err != nil {
		return InventoryStockMutation{}, err
	}
	return mutationWithPreviousAvailability(mutation, previous), nil
}

func (r *MongoProductRepository) CommitVariant(
	ctx context.Context,
	productID string,
	variantID string,
	quantity int64,
	now time.Time,
) (InventoryStockMutation, error) {
	if err := ctx.Err(); err != nil {
		return InventoryStockMutation{}, err
	}
	productID = strings.TrimSpace(productID)
	variantID = strings.TrimSpace(variantID)
	if productID == "" || variantID == "" || quantity <= 0 {
		return InventoryStockMutation{}, ErrWriteConflict
	}
	previous, err := r.previousInventoryState(ctx, productID, variantID)
	if err != nil {
		return InventoryStockMutation{}, err
	}

	result, err := r.products.UpdateOne(
		ctx,
		bson.D{
			e("_id", productID),
			e("variants", bson.D{e("$elemMatch", bson.D{
				e("variant_id", variantID),
				e("stock_quantity", bson.D{e("$gte", quantity)}),
				e("reserved_quantity", bson.D{e("$gte", quantity)}),
			})}),
		},
		inventoryUpdate(
			bson.D{
				e("variants.$.stock_quantity", -quantity),
				e("variants.$.reserved_quantity", -quantity),
			},
			now,
		),
	)
	if err != nil {
		return InventoryStockMutation{}, fmt.Errorf("commit inventory product %q variant %q: %w", productID, variantID, err)
	}
	if result.MatchedCount == 0 {
		return InventoryStockMutation{}, r.classifyCommitFailure(ctx, productID, variantID, quantity)
	}
	mutation, err := r.currentInventoryMutation(ctx, productID, variantID)
	if err != nil {
		return InventoryStockMutation{}, err
	}
	return mutationWithPreviousAvailability(mutation, previous), nil
}

func (r *MongoProductRepository) FindReservationByID(ctx context.Context, reservationID string) (*domain.InventoryReservation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var doc inventoryReservationDocument
	err := r.reservations.FindOne(ctx, bson.D{e("_id", strings.TrimSpace(reservationID))}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find inventory reservation %q: %w", reservationID, err)
	}
	reservation := doc.toDomain()
	return &reservation, nil
}

func (r *MongoProductRepository) FindReservationByOrderID(ctx context.Context, orderID string) (*domain.InventoryReservation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var doc inventoryReservationDocument
	err := r.reservations.FindOne(ctx, bson.D{e("order_id", strings.TrimSpace(orderID))}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find inventory reservation by order %q: %w", orderID, err)
	}
	reservation := doc.toDomain()
	return &reservation, nil
}

func (r *MongoProductRepository) CreateInventoryReservation(ctx context.Context, reservation *domain.InventoryReservation) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if reservation == nil {
		return fmt.Errorf("inventory reservation is required")
	}
	_, err := r.reservations.InsertOne(ctx, inventoryReservationDocumentFromDomain(*reservation))
	if mongo.IsDuplicateKeyError(err) {
		return ErrDuplicateKey
	}
	if err != nil {
		return fmt.Errorf("insert inventory reservation %q: %w", reservation.ID, err)
	}
	return nil
}

func (r *MongoProductRepository) MarkReservationReleased(
	ctx context.Context,
	reservationID string,
	reason string,
	releasedAt time.Time,
) error {
	return r.markReservationTerminal(ctx, reservationID, domain.ReservationStatusReleased, reason, "released_at", releasedAt)
}

func (r *MongoProductRepository) MarkReservationCommitted(ctx context.Context, reservationID string, committedAt time.Time) error {
	return r.markReservationTerminal(ctx, reservationID, domain.ReservationStatusCommitted, "", "committed_at", committedAt)
}

func (r *MongoProductRepository) MarkReservationExpired(
	ctx context.Context,
	reservationID string,
	reason string,
	expiredAt time.Time,
) error {
	return r.markReservationTerminal(ctx, reservationID, domain.ReservationStatusExpired, reason, "expired_at", expiredAt)
}

func (r *MongoProductRepository) ListExpiredReservations(
	ctx context.Context,
	now time.Time,
	limit int64,
) ([]domain.InventoryReservation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 100
	}

	cursor, err := r.reservations.Find(
		ctx,
		bson.D{
			e("status", string(domain.ReservationStatusReserved)),
			e("expires_at", bson.D{e("$lte", now.UTC())}),
		},
		options.Find().
			SetSort(bson.D{e("expires_at", 1), e("_id", 1)}).
			SetLimit(limit),
	)
	if err != nil {
		return nil, fmt.Errorf("list expired inventory reservations: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []inventoryReservationDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("decode expired inventory reservations: %w", err)
	}

	reservations := make([]domain.InventoryReservation, 0, len(docs))
	for _, doc := range docs {
		reservations = append(reservations, doc.toDomain())
	}
	return reservations, nil
}

func (r *MongoProductRepository) CreateInventorySnapshot(ctx context.Context, snapshot domain.InventorySnapshot) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	_, err := r.snapshots.InsertOne(ctx, inventorySnapshotDocumentFromDomain(snapshot))
	if mongo.IsDuplicateKeyError(err) {
		return ErrDuplicateKey
	}
	if err != nil {
		return fmt.Errorf("insert inventory snapshot %q: %w", snapshot.ID, err)
	}
	return nil
}

func (r *MongoProductRepository) markReservationTerminal(
	ctx context.Context,
	reservationID string,
	status domain.InventoryReservationStatus,
	reason string,
	timestampField string,
	timestamp time.Time,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	now := timestamp.UTC()
	set := bson.D{
		e("status", string(status)),
		e("updated_at", now),
		e(timestampField, now),
	}
	if strings.TrimSpace(reason) != "" {
		set = append(set, e("reason", strings.TrimSpace(reason)))
	}
	result, err := r.reservations.UpdateOne(
		ctx,
		bson.D{
			e("_id", strings.TrimSpace(reservationID)),
			e("status", string(domain.ReservationStatusReserved)),
		},
		bson.D{e("$set", set)},
	)
	if err != nil {
		return fmt.Errorf("mark inventory reservation %q %s: %w", reservationID, status, err)
	}
	if result.MatchedCount == 0 {
		return ErrWriteConflict
	}
	return nil
}

func reserveVariantFilter(productID string, variantID string, quantity int64) bson.D {
	return bson.D{
		e("_id", productID),
		e("status", string(domain.ProductStatusPublished)),
		e("variants", bson.D{e("$elemMatch", bson.D{
			e("variant_id", variantID),
			e("status", string(domain.VariantStatusActive)),
		})}),
		e("$expr", availableQuantityAtLeastExpression(variantID, quantity)),
	}
}

func availableQuantityAtLeastExpression(variantID string, quantity int64) bson.D {
	selectedVariant := bson.D{e("$first", bson.D{e("$filter", bson.D{
		e("input", "$variants"),
		e("as", "variant"),
		e("cond", bson.D{e("$eq", bson.A{"$$variant.variant_id", variantID})}),
	})})}
	availableQuantity := bson.D{e("$let", bson.D{
		e("vars", bson.D{e("selected_variant", selectedVariant)}),
		e("in", bson.D{e("$subtract", bson.A{
			bson.D{e("$subtract", bson.A{"$$selected_variant.stock_quantity", "$$selected_variant.reserved_quantity"})},
			bson.D{e("$ifNull", bson.A{"$$selected_variant.safety_stock", 0})},
		})}),
	})}
	return bson.D{e("$gte", bson.A{availableQuantity, quantity})}
}

func inventoryUpdate(inc bson.D, now time.Time) bson.D {
	return bson.D{
		e("$inc", inc),
		e("$set", bson.D{e("updated_at", now.UTC())}),
	}
}

func (r *MongoProductRepository) classifyReserveFailure(
	ctx context.Context,
	productID string,
	variantID string,
	quantity int64,
) error {
	product, variant, err := r.findInventoryVariant(ctx, productID, variantID)
	if err != nil {
		return err
	}
	if product.Status != domain.ProductStatusPublished {
		return ErrInventoryUnavailable
	}
	if !variant.Active() {
		return ErrInventoryUnavailable
	}
	if !variant.InventoryState().CanFulfill(quantity) {
		return ErrInsufficientStock
	}
	return ErrWriteConflict
}

func (r *MongoProductRepository) classifyReleaseFailure(
	ctx context.Context,
	productID string,
	variantID string,
	quantity int64,
) error {
	_, variant, err := r.findInventoryVariant(ctx, productID, variantID)
	if err != nil {
		return err
	}
	if variant.ReservedQuantity < quantity {
		return ErrWriteConflict
	}
	return ErrWriteConflict
}

func (r *MongoProductRepository) classifyCommitFailure(
	ctx context.Context,
	productID string,
	variantID string,
	quantity int64,
) error {
	_, variant, err := r.findInventoryVariant(ctx, productID, variantID)
	if err != nil {
		return err
	}
	if variant.StockQuantity < quantity || variant.ReservedQuantity < quantity {
		return ErrWriteConflict
	}
	return ErrWriteConflict
}

func (r *MongoProductRepository) currentInventoryMutation(
	ctx context.Context,
	productID string,
	variantID string,
) (InventoryStockMutation, error) {
	product, variant, err := r.findInventoryVariant(ctx, productID, variantID)
	if err != nil {
		return InventoryStockMutation{}, err
	}
	state := variant.InventoryState()
	return InventoryStockMutation{
		ProductID:         product.ID,
		VariantID:         variant.ID,
		SKU:               variant.SKU,
		SellerID:          product.SellerID,
		StockQuantity:     state.StockQuantity,
		ReservedQuantity:  state.ReservedQuantity,
		SafetyStock:       state.SafetyStock,
		AvailableQuantity: state.AvailableQuantity(),
	}, nil
}

func (r *MongoProductRepository) previousInventoryState(
	ctx context.Context,
	productID string,
	variantID string,
) (domain.InventoryState, error) {
	_, variant, err := r.findInventoryVariant(ctx, productID, variantID)
	if err != nil {
		return domain.InventoryState{}, err
	}
	return variant.InventoryState(), nil
}

func mutationWithPreviousAvailability(
	mutation InventoryStockMutation,
	previous domain.InventoryState,
) InventoryStockMutation {
	mutation.PreviousAvailableQuantity = previous.AvailableQuantity()
	mutation.PreviousInStock = previous.AvailableQuantity() > 0
	mutation.HasPreviousAvailability = true
	return mutation
}

func (r *MongoProductRepository) findInventoryVariant(
	ctx context.Context,
	productID string,
	variantID string,
) (*domain.Product, domain.Variant, error) {
	product, err := r.FindProductByID(ctx, productID)
	if err != nil {
		return nil, domain.Variant{}, err
	}
	for _, variant := range product.Variants {
		if variant.ID == variantID {
			return product, variant, nil
		}
	}
	return nil, domain.Variant{}, ErrNotFound
}

type inventoryReservationDocument struct {
	ID             string                             `bson:"_id"`
	OrderID        string                             `bson:"order_id"`
	IdempotencyKey string                             `bson:"idempotency_key,omitempty"`
	Status         string                             `bson:"status"`
	Items          []inventoryReservationItemDocument `bson:"items"`
	ExpiresAt      time.Time                          `bson:"expires_at"`
	CreatedAt      time.Time                          `bson:"created_at"`
	UpdatedAt      time.Time                          `bson:"updated_at"`
	CommittedAt    *time.Time                         `bson:"committed_at,omitempty"`
	ReleasedAt     *time.Time                         `bson:"released_at,omitempty"`
	ExpiredAt      *time.Time                         `bson:"expired_at,omitempty"`
	Reason         string                             `bson:"reason,omitempty"`
}

type inventoryReservationItemDocument struct {
	ProductID string `bson:"product_id"`
	VariantID string `bson:"variant_id"`
	SKU       string `bson:"sku"`
	SellerID  string `bson:"seller_id"`
	Quantity  int64  `bson:"quantity"`
}

func inventoryReservationDocumentFromDomain(reservation domain.InventoryReservation) inventoryReservationDocument {
	items := make([]inventoryReservationItemDocument, 0, len(reservation.Items))
	for _, item := range reservation.Items {
		items = append(items, inventoryReservationItemDocument{
			ProductID: item.ProductID,
			VariantID: item.VariantID,
			SKU:       normalizeRepositorySKU(item.SKU),
			SellerID:  item.SellerID,
			Quantity:  item.Quantity,
		})
	}
	return inventoryReservationDocument{
		ID:             reservation.ID,
		OrderID:        reservation.OrderID,
		IdempotencyKey: reservation.IdempotencyKey,
		Status:         string(reservation.Status),
		Items:          items,
		ExpiresAt:      reservation.ExpiresAt,
		CreatedAt:      reservation.CreatedAt,
		UpdatedAt:      reservation.UpdatedAt,
		CommittedAt:    reservation.CommittedAt,
		ReleasedAt:     reservation.ReleasedAt,
		ExpiredAt:      reservation.ExpiredAt,
		Reason:         reservation.Reason,
	}
}

func (d inventoryReservationDocument) toDomain() domain.InventoryReservation {
	items := make([]domain.InventoryReservationItem, 0, len(d.Items))
	for _, item := range d.Items {
		items = append(items, domain.InventoryReservationItem{
			ProductID: item.ProductID,
			VariantID: item.VariantID,
			SKU:       item.SKU,
			SellerID:  item.SellerID,
			Quantity:  item.Quantity,
		})
	}
	return domain.InventoryReservation{
		ID:             d.ID,
		OrderID:        d.OrderID,
		IdempotencyKey: d.IdempotencyKey,
		Status:         domain.InventoryReservationStatus(d.Status),
		Items:          items,
		ExpiresAt:      d.ExpiresAt,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
		CommittedAt:    d.CommittedAt,
		ReleasedAt:     d.ReleasedAt,
		ExpiredAt:      d.ExpiredAt,
		Reason:         d.Reason,
	}
}

type inventorySnapshotDocument struct {
	ID                string                      `bson:"_id"`
	ProductID         string                      `bson:"product_id"`
	VariantID         string                      `bson:"variant_id"`
	SKU               string                      `bson:"sku"`
	SellerID          string                      `bson:"seller_id"`
	SnapshotType      string                      `bson:"snapshot_type"`
	StockQuantity     int64                       `bson:"stock_quantity"`
	ReservedQuantity  int64                       `bson:"reserved_quantity"`
	SafetyStock       int64                       `bson:"safety_stock,omitempty"`
	AvailableQuantity int64                       `bson:"available_quantity"`
	Reason            string                      `bson:"reason,omitempty"`
	Reference         *inventoryReferenceDocument `bson:"reference,omitempty"`
	CreatedAt         time.Time                   `bson:"created_at"`
}

type inventoryReferenceDocument struct {
	Type    string `bson:"type,omitempty"`
	ID      string `bson:"id,omitempty"`
	OrderID string `bson:"order_id,omitempty"`
}

func inventorySnapshotDocumentFromDomain(snapshot domain.InventorySnapshot) inventorySnapshotDocument {
	var reference *inventoryReferenceDocument
	if snapshot.Reference != nil {
		reference = &inventoryReferenceDocument{
			Type:    snapshot.Reference.Type,
			ID:      snapshot.Reference.ID,
			OrderID: snapshot.Reference.OrderID,
		}
	}
	return inventorySnapshotDocument{
		ID:                snapshot.ID,
		ProductID:         snapshot.ProductID,
		VariantID:         snapshot.VariantID,
		SKU:               normalizeRepositorySKU(snapshot.SKU),
		SellerID:          snapshot.SellerID,
		SnapshotType:      string(snapshot.SnapshotType),
		StockQuantity:     snapshot.StockQuantity,
		ReservedQuantity:  snapshot.ReservedQuantity,
		SafetyStock:       snapshot.SafetyStock,
		AvailableQuantity: snapshot.AvailableQuantity,
		Reason:            snapshot.Reason,
		Reference:         reference,
		CreatedAt:         snapshot.CreatedAt,
	}
}
