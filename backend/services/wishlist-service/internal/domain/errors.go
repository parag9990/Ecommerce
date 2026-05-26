package domain

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidWishlist       = errors.New("invalid wishlist")
	ErrWishlistNotFound      = errors.New("wishlist not found")
	ErrDuplicateProduct      = errors.New("product already exists in wishlist")
	ErrWishlistItemNotFound  = errors.New("wishlist item not found")
	ErrWishlistEventNotFound = errors.New("wishlist event not found")
)

type FieldError struct {
	Field   string
	Message string
}

func (e FieldError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func invalidField(field, message string) error {
	return fmt.Errorf("%w: %w", ErrInvalidWishlist, FieldError{
		Field:   field,
		Message: message,
	})
}
