package domain

import (
	"fmt"
	"strings"
)

type SourceEventType string

const (
	EventProductViewed       SourceEventType = "ProductViewed"
	EventCartItemAdded       SourceEventType = "CartItemAdded"
	EventWishlistItemAdded   SourceEventType = "WishlistItemAdded"
	EventWishlistItemRemoved SourceEventType = "WishlistItemRemoved"
	EventOrderPaid           SourceEventType = "OrderPaid"
	EventPurchaseCompleted   SourceEventType = "PurchaseCompleted"
)

type InteractionType string

const (
	InteractionProductView    InteractionType = "product_view"
	InteractionAddToCart      InteractionType = "add_to_cart"
	InteractionWishlistAdd    InteractionType = "wishlist_add"
	InteractionWishlistRemove InteractionType = "wishlist_remove"
	InteractionPurchase       InteractionType = "purchase"
)

func SupportedSourceEventTypes() []SourceEventType {
	return []SourceEventType{
		EventProductViewed,
		EventCartItemAdded,
		EventWishlistItemAdded,
		EventWishlistItemRemoved,
		EventOrderPaid,
		EventPurchaseCompleted,
	}
}

func ParseSourceEventType(value string) (SourceEventType, error) {
	eventType := SourceEventType(strings.TrimSpace(value))
	if eventType.IsSupported() {
		return eventType, nil
	}
	return "", fmt.Errorf("%w: %s", ErrUnsupportedEventType, value)
}

func (t SourceEventType) IsSupported() bool {
	switch t {
	case EventProductViewed,
		EventCartItemAdded,
		EventWishlistItemAdded,
		EventWishlistItemRemoved,
		EventOrderPaid,
		EventPurchaseCompleted:
		return true
	default:
		return false
	}
}

func (t SourceEventType) NormalizedInteractionType() (InteractionType, bool) {
	switch t {
	case EventProductViewed:
		return InteractionProductView, true
	case EventCartItemAdded:
		return InteractionAddToCart, true
	case EventWishlistItemAdded:
		return InteractionWishlistAdd, true
	case EventWishlistItemRemoved:
		return InteractionWishlistRemove, true
	case EventOrderPaid, EventPurchaseCompleted:
		return InteractionPurchase, true
	default:
		return "", false
	}
}

func (t SourceEventType) InteractionWeight() (int, bool) {
	switch t {
	case EventProductViewed:
		return 1, true
	case EventWishlistItemAdded:
		return 3, true
	case EventWishlistItemRemoved:
		return -2, true
	case EventCartItemAdded:
		return 4, true
	case EventOrderPaid, EventPurchaseCompleted:
		return 8, true
	default:
		return 0, false
	}
}

func (t InteractionType) IsValid() bool {
	switch t {
	case InteractionProductView,
		InteractionAddToCart,
		InteractionWishlistAdd,
		InteractionWishlistRemove,
		InteractionPurchase:
		return true
	default:
		return false
	}
}
