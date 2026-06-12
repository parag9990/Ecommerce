package domain

import (
	"fmt"
	"strings"
	"unicode"
)

const CurrencyINR = "INR"

type Money struct {
	Amount   int64  `bson:"amount" json:"amount"`
	Currency string `bson:"currency" json:"currency"`
}

func NewMoney(amount int64, currency string) Money {
	return Money{
		Amount:   amount,
		Currency: strings.ToUpper(strings.TrimSpace(currency)),
	}
}

func (m Money) Validate() error {
	if m.Amount < 0 {
		return fmt.Errorf("%w: amount must be non-negative", ErrInvalidMoney)
	}
	if !validCurrency(m.Currency) {
		return fmt.Errorf("%w: currency must be a 3-letter ISO code", ErrInvalidMoney)
	}
	return nil
}

func (m Money) SameCurrency(other Money) bool {
	return strings.EqualFold(m.Currency, other.Currency)
}

func validCurrency(currency string) bool {
	if len(currency) != 3 {
		return false
	}
	for _, r := range currency {
		if !unicode.IsUpper(r) || !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}
