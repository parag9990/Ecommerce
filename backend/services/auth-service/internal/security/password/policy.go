package password

import (
	"errors"
	"strings"
	"unicode/utf8"
)

const (
	DefaultMinLength = 8
	DefaultMaxLength = 128
)

var (
	ErrPasswordBlank    = errors.New("password cannot be blank")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong  = errors.New("password must be at most 128 characters")
)

type Policy struct {
	MinLength int
	MaxLength int
}

func DefaultPolicy() Policy {
	return Policy{
		MinLength: DefaultMinLength,
		MaxLength: DefaultMaxLength,
	}
}

func (p Policy) Validate() error {
	if p.MinLength <= 0 {
		return errors.New("minimum password length must be greater than zero")
	}
	if p.MaxLength < p.MinLength {
		return errors.New("maximum password length must be greater than or equal to minimum length")
	}
	return nil
}

func (p Policy) ValidatePassword(plain string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(plain) == "" {
		return ErrPasswordBlank
	}
	length := utf8.RuneCountInString(plain)
	if length < p.MinLength {
		return ErrPasswordTooShort
	}
	if length > p.MaxLength {
		return ErrPasswordTooLong
	}
	return nil
}
