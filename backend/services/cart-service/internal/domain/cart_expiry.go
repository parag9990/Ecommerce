package domain

import (
	"strings"
	"time"
)

type ExpiredCartCandidate struct {
	CartID         string    `bson:"_id" json:"cart_id"`
	UserID         *string   `bson:"user_id,omitempty" json:"user_id,omitempty"`
	GuestSessionID *string   `bson:"guest_session_id,omitempty" json:"guest_session_id,omitempty"`
	Version        int64     `bson:"version" json:"version"`
	ExpiresAt      time.Time `bson:"expires_at" json:"expires_at"`
}

func IsActiveCartExpired(cart *Cart, now time.Time) bool {
	if cart == nil || cart.Status != CartStatusActive {
		return false
	}
	return !cart.ExpiresAt.After(now.UTC())
}

func (c *Cart) MarkExpired(now time.Time) error {
	if c == nil {
		return ErrInvalidCart
	}
	if c.Status != CartStatusActive {
		return ErrCartNotActive
	}
	now = now.UTC()
	if now.IsZero() {
		return ErrInvalidCart
	}
	c.Status = CartStatusExpired
	c.UpdatedAt = now
	return nil
}

func (c ExpiredCartCandidate) Owner() (CartOwner, bool) {
	if c.UserID != nil {
		userID := strings.TrimSpace(*c.UserID)
		if userID != "" {
			return CartOwner{Type: CartOwnerTypeUser, UserID: userID}, true
		}
	}
	if c.GuestSessionID != nil {
		guestSessionID := strings.TrimSpace(*c.GuestSessionID)
		if guestSessionID != "" {
			return CartOwner{Type: CartOwnerTypeGuest, GuestSessionID: guestSessionID}, true
		}
	}
	return CartOwner{}, false
}
