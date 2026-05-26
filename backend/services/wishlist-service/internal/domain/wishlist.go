package domain

import (
	"fmt"
	"strings"
	"time"
)

type Visibility string

const (
	VisibilityPrivate Visibility = "private"
)

func (v Visibility) IsValid() bool {
	return v == VisibilityPrivate
}

type Availability string

const (
	AvailabilityInStock    Availability = "in_stock"
	AvailabilityOutOfStock Availability = "out_of_stock"
	AvailabilityUnknown    Availability = "unknown"
	AvailabilityDeleted    Availability = "deleted"
)

type AvailabilityUpdateResult struct {
	MatchedCount  int64
	ModifiedCount int64
}

type PriceUpdateResult struct {
	MatchedCount  int64
	ModifiedCount int64
}

type PriceDropCandidate struct {
	UserID        string
	ProductID     string
	VariantID     string
	PreviousPrice Money
}

func (a Availability) Normalized() Availability {
	if a == "" {
		return AvailabilityUnknown
	}
	return a
}

func (a Availability) IsValid() bool {
	switch a.Normalized() {
	case AvailabilityInStock, AvailabilityOutOfStock, AvailabilityUnknown, AvailabilityDeleted:
		return true
	default:
		return false
	}
}

type Money struct {
	Amount   int64
	Currency string
}

func (m Money) Validate() error {
	if m.Amount < 0 {
		return invalidField("last_known_price.amount", "must be greater than or equal to zero")
	}
	if !isISO4217Like(m.Currency) {
		return invalidField("last_known_price.currency", "must be a three-letter uppercase currency code")
	}
	return nil
}

type AddWishlistItemInput struct {
	ProductID      string
	VariantID      string
	LastKnownPrice *Money
	Availability   Availability
}

type WishlistItem struct {
	ProductID      string
	VariantID      string
	AddedAt        time.Time
	LastKnownPrice *Money
	Availability   Availability
}

func NewWishlistItem(input AddWishlistItemInput, addedAt time.Time) (WishlistItem, error) {
	item := WishlistItem{
		ProductID:      normalizeID(input.ProductID),
		VariantID:      normalizeID(input.VariantID),
		AddedAt:        normalizeTime(addedAt),
		LastKnownPrice: copyMoney(input.LastKnownPrice),
		Availability:   input.Availability.Normalized(),
	}
	if err := item.Validate(); err != nil {
		return WishlistItem{}, err
	}
	return item, nil
}

func (i WishlistItem) Validate() error {
	if i.ProductID == "" {
		return invalidField("items.product_id", "is required")
	}
	if i.AddedAt.IsZero() {
		return invalidField("items.added_at", "is required")
	}
	if !i.Availability.IsValid() {
		return invalidField("items.availability", "must be in_stock, out_of_stock, unknown, or deleted")
	}
	if i.LastKnownPrice != nil {
		if err := i.LastKnownPrice.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type Wishlist struct {
	ID         string
	UserID     string
	Visibility Visibility
	Items      []WishlistItem
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewWishlist(id, userID string, createdAt time.Time) (*Wishlist, error) {
	now := normalizeTime(createdAt)
	wishlist := &Wishlist{
		ID:         normalizeID(id),
		UserID:     normalizeID(userID),
		Visibility: VisibilityPrivate,
		Items:      make([]WishlistItem, 0),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := wishlist.Validate(); err != nil {
		return nil, err
	}
	return wishlist, nil
}

func RehydrateWishlist(snapshot Wishlist) (*Wishlist, error) {
	wishlist := &Wishlist{
		ID:         normalizeID(snapshot.ID),
		UserID:     normalizeID(snapshot.UserID),
		Visibility: snapshot.Visibility,
		Items:      copyItems(snapshot.Items),
		CreatedAt:  normalizeTime(snapshot.CreatedAt),
		UpdatedAt:  normalizeTime(snapshot.UpdatedAt),
	}
	for idx := range wishlist.Items {
		wishlist.Items[idx].ProductID = normalizeID(wishlist.Items[idx].ProductID)
		wishlist.Items[idx].VariantID = normalizeID(wishlist.Items[idx].VariantID)
		wishlist.Items[idx].AddedAt = normalizeTime(wishlist.Items[idx].AddedAt)
		wishlist.Items[idx].Availability = wishlist.Items[idx].Availability.Normalized()
		wishlist.Items[idx].LastKnownPrice = copyMoney(wishlist.Items[idx].LastKnownPrice)
	}
	if wishlist.Items == nil {
		wishlist.Items = make([]WishlistItem, 0)
	}
	if err := wishlist.Validate(); err != nil {
		return nil, err
	}
	return wishlist, nil
}

func (w *Wishlist) AddItem(input AddWishlistItemInput, addedAt time.Time) error {
	if w == nil {
		return invalidField("wishlist", "is required")
	}
	if err := w.Validate(); err != nil {
		return err
	}
	item, err := NewWishlistItem(input, addedAt)
	if err != nil {
		return err
	}
	if item.AddedAt.Before(w.CreatedAt) {
		return invalidField("items.added_at", "must not be before created_at")
	}
	if w.HasProduct(item.ProductID) {
		return fmt.Errorf("%w: %s", ErrDuplicateProduct, item.ProductID)
	}

	next := w.Snapshot()
	next.Items = append(next.Items, item)
	next.UpdatedAt = item.AddedAt
	if err := next.Validate(); err != nil {
		return err
	}
	w.Items = next.Items
	w.UpdatedAt = item.AddedAt
	return nil
}

func (w *Wishlist) RemoveProduct(productID string, removedAt time.Time) error {
	if w == nil {
		return invalidField("wishlist", "is required")
	}
	if err := w.Validate(); err != nil {
		return err
	}
	productID = normalizeID(productID)
	if productID == "" {
		return invalidField("product_id", "is required")
	}
	removedAt = normalizeTime(removedAt)
	if removedAt.IsZero() {
		return invalidField("removed_at", "is required")
	}
	if removedAt.Before(w.CreatedAt) {
		return invalidField("removed_at", "must not be before created_at")
	}

	for idx, item := range w.Items {
		if item.ProductID == productID {
			next := w.Snapshot()
			next.Items = append(next.Items[:idx], next.Items[idx+1:]...)
			next.UpdatedAt = removedAt
			if err := next.Validate(); err != nil {
				return err
			}
			w.Items = next.Items
			w.UpdatedAt = removedAt
			return nil
		}
	}
	return fmt.Errorf("%w: %s", ErrWishlistItemNotFound, productID)
}

func (w *Wishlist) HasProduct(productID string) bool {
	if w == nil {
		return false
	}
	productID = normalizeID(productID)
	if productID == "" {
		return false
	}
	for _, item := range w.Items {
		if item.ProductID == productID {
			return true
		}
	}
	return false
}

func (w *Wishlist) FindItem(productID string) (WishlistItem, bool) {
	if w == nil {
		return WishlistItem{}, false
	}
	productID = normalizeID(productID)
	for _, item := range w.Items {
		if item.ProductID == productID {
			return copyItem(item), true
		}
	}
	return WishlistItem{}, false
}

func (w *Wishlist) Validate() error {
	if w == nil {
		return invalidField("wishlist", "is required")
	}
	if normalizeID(w.ID) == "" {
		return invalidField("wishlist_id", "is required")
	}
	if normalizeID(w.UserID) == "" {
		return invalidField("user_id", "is required")
	}
	if !w.Visibility.IsValid() {
		return invalidField("visibility", "must be private")
	}
	if w.CreatedAt.IsZero() {
		return invalidField("created_at", "is required")
	}
	if w.UpdatedAt.IsZero() {
		return invalidField("updated_at", "is required")
	}
	if w.UpdatedAt.Before(w.CreatedAt) {
		return invalidField("updated_at", "must not be before created_at")
	}
	seenProducts := make(map[string]struct{}, len(w.Items))
	for _, item := range w.Items {
		if err := item.Validate(); err != nil {
			return err
		}
		productID := normalizeID(item.ProductID)
		if _, exists := seenProducts[productID]; exists {
			return fmt.Errorf("%w: %s", ErrDuplicateProduct, productID)
		}
		seenProducts[productID] = struct{}{}
	}
	return nil
}

func (w *Wishlist) Snapshot() Wishlist {
	if w == nil {
		return Wishlist{}
	}
	return Wishlist{
		ID:         w.ID,
		UserID:     w.UserID,
		Visibility: w.Visibility,
		Items:      copyItems(w.Items),
		CreatedAt:  w.CreatedAt,
		UpdatedAt:  w.UpdatedAt,
	}
}

func normalizeID(value string) string {
	return strings.TrimSpace(value)
}

func normalizeTime(value time.Time) time.Time {
	if value.IsZero() {
		return value
	}
	return value.UTC()
}

func copyItems(items []WishlistItem) []WishlistItem {
	if items == nil {
		return nil
	}
	copied := make([]WishlistItem, len(items))
	for idx, item := range items {
		copied[idx] = copyItem(item)
	}
	return copied
}

func copyItem(item WishlistItem) WishlistItem {
	return WishlistItem{
		ProductID:      item.ProductID,
		VariantID:      item.VariantID,
		AddedAt:        item.AddedAt,
		LastKnownPrice: copyMoney(item.LastKnownPrice),
		Availability:   item.Availability,
	}
}

func copyMoney(money *Money) *Money {
	if money == nil {
		return nil
	}
	copied := *money
	copied.Currency = strings.ToUpper(strings.TrimSpace(copied.Currency))
	return &copied
}

func isISO4217Like(currency string) bool {
	currency = strings.TrimSpace(currency)
	if len(currency) != 3 {
		return false
	}
	for _, c := range currency {
		if c < 'A' || c > 'Z' {
			return false
		}
	}
	return true
}
