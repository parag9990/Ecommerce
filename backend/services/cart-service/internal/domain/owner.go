package domain

import (
	"fmt"
	"strings"
)

type CartOwnerType string

const (
	CartOwnerTypeUser  CartOwnerType = "user"
	CartOwnerTypeGuest CartOwnerType = "guest"
)

type CartOwner struct {
	Type           CartOwnerType
	UserID         string
	GuestSessionID string
}

func ResolveCartOwner(userID string, guestSessionID string) (CartOwner, error) {
	userID = strings.TrimSpace(userID)
	guestSessionID = strings.TrimSpace(guestSessionID)
	if userID != "" {
		return CartOwner{Type: CartOwnerTypeUser, UserID: userID}, nil
	}
	if guestSessionID != "" {
		return CartOwner{Type: CartOwnerTypeGuest, GuestSessionID: guestSessionID}, nil
	}
	return CartOwner{}, ErrCartOwnerMissing
}

func (o CartOwner) Validate() error {
	switch o.Type {
	case CartOwnerTypeUser:
		if strings.TrimSpace(o.UserID) == "" || strings.TrimSpace(o.GuestSessionID) != "" {
			return fmt.Errorf("%w: user owner requires only user_id", ErrInvalidCartOwner)
		}
	case CartOwnerTypeGuest:
		if strings.TrimSpace(o.GuestSessionID) == "" || strings.TrimSpace(o.UserID) != "" {
			return fmt.Errorf("%w: guest owner requires only guest_session_id", ErrInvalidCartOwner)
		}
	default:
		return fmt.Errorf("%w: unsupported owner type %q", ErrInvalidCartOwner, o.Type)
	}
	return nil
}

func (o CartOwner) IsUser() bool {
	return o.Type == CartOwnerTypeUser
}

func (o CartOwner) IsGuest() bool {
	return o.Type == CartOwnerTypeGuest
}
