package repository

import (
	"fmt"
	"time"

	"ecommerce/backend/services/wishlist-service/internal/domain"
)

type MoneySnapshotDocument struct {
	Amount   int64  `bson:"amount" json:"amount"`
	Currency string `bson:"currency" json:"currency"`
}

type WishlistItemDocument struct {
	ProductID      string                 `bson:"product_id" json:"product_id"`
	VariantID      string                 `bson:"variant_id,omitempty" json:"variant_id,omitempty"`
	AddedAt        time.Time              `bson:"added_at" json:"added_at"`
	LastKnownPrice *MoneySnapshotDocument `bson:"last_known_price,omitempty" json:"last_known_price,omitempty"`
	Availability   domain.Availability    `bson:"availability,omitempty" json:"availability,omitempty"`
}

type WishlistDocument struct {
	ID         string                 `bson:"_id" json:"wishlist_id"`
	UserID     string                 `bson:"user_id" json:"user_id"`
	Visibility domain.Visibility      `bson:"visibility" json:"visibility"`
	Items      []WishlistItemDocument `bson:"items" json:"items"`
	CreatedAt  time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time              `bson:"updated_at" json:"updated_at"`
}

func NewWishlistDocument(wishlist *domain.Wishlist) (WishlistDocument, error) {
	if wishlist == nil {
		return WishlistDocument{}, fmt.Errorf("%w: wishlist is required", domain.ErrInvalidWishlist)
	}
	if err := wishlist.Validate(); err != nil {
		return WishlistDocument{}, err
	}

	snapshot := wishlist.Snapshot()
	items := make([]WishlistItemDocument, 0, len(snapshot.Items))
	for _, item := range snapshot.Items {
		items = append(items, newWishlistItemDocument(item))
	}

	return WishlistDocument{
		ID:         snapshot.ID,
		UserID:     snapshot.UserID,
		Visibility: snapshot.Visibility,
		Items:      items,
		CreatedAt:  snapshot.CreatedAt,
		UpdatedAt:  snapshot.UpdatedAt,
	}, nil
}

func (d WishlistDocument) ToDomain() (*domain.Wishlist, error) {
	items := make([]domain.WishlistItem, 0, len(d.Items))
	for _, item := range d.Items {
		items = append(items, item.toDomain())
	}

	return domain.RehydrateWishlist(domain.Wishlist{
		ID:         d.ID,
		UserID:     d.UserID,
		Visibility: d.Visibility,
		Items:      items,
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
	})
}

func newWishlistItemDocument(item domain.WishlistItem) WishlistItemDocument {
	return WishlistItemDocument{
		ProductID:      item.ProductID,
		VariantID:      item.VariantID,
		AddedAt:        item.AddedAt,
		LastKnownPrice: newMoneySnapshotDocument(item.LastKnownPrice),
		Availability:   item.Availability,
	}
}

func (d WishlistItemDocument) toDomain() domain.WishlistItem {
	return domain.WishlistItem{
		ProductID:      d.ProductID,
		VariantID:      d.VariantID,
		AddedAt:        d.AddedAt,
		LastKnownPrice: d.LastKnownPrice.toDomain(),
		Availability:   d.Availability,
	}
}

func newMoneySnapshotDocument(money *domain.Money) *MoneySnapshotDocument {
	if money == nil {
		return nil
	}
	return &MoneySnapshotDocument{
		Amount:   money.Amount,
		Currency: money.Currency,
	}
}

func (d *MoneySnapshotDocument) toDomain() *domain.Money {
	if d == nil {
		return nil
	}
	return &domain.Money{
		Amount:   d.Amount,
		Currency: d.Currency,
	}
}
