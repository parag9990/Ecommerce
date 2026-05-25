package usecase

import (
	"errors"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cart-service/internal/domain"
)

type CartTTLPolicy struct {
	UserTTL  time.Duration
	GuestTTL time.Duration
}

func NewCartTTLPolicy(legacyTTL time.Duration, userTTL time.Duration, guestTTL time.Duration) (CartTTLPolicy, error) {
	if userTTL <= 0 {
		userTTL = legacyTTL
	}
	if guestTTL <= 0 {
		guestTTL = legacyTTL
	}
	if userTTL <= 0 || guestTTL <= 0 {
		return CartTTLPolicy{}, errors.New("cart ttl values must be greater than zero")
	}
	return CartTTLPolicy{UserTTL: userTTL, GuestTTL: guestTTL}, nil
}

func (p CartTTLPolicy) TTLForOwner(owner domain.CartOwner) time.Duration {
	if owner.IsUser() {
		return p.UserTTL
	}
	return p.GuestTTL
}
