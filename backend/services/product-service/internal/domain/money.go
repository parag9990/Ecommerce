package domain

import "strings"

type Money struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

func NewMoney(amount int64, currency string) Money {
	return Money{
		Amount:   amount,
		Currency: normalizeCurrency(currency),
	}
}

func (m Money) Validate(field string) ValidationReport {
	var report ValidationReport
	if m.Amount <= 0 {
		report.AddError(CodeInvalidPrice, fieldPath(field, "amount"), "money amount must be greater than zero and stored in minor units")
	}
	if !isCurrencyCode(m.Currency) {
		report.AddError(CodeInvalidCurrency, fieldPath(field, "currency"), "currency must be a three-letter ISO-style code")
	}
	return report
}

func normalizeCurrency(currency string) string {
	return strings.ToUpper(strings.TrimSpace(currency))
}

func isCurrencyCode(currency string) bool {
	currency = normalizeCurrency(currency)
	if len(currency) != 3 {
		return false
	}
	for _, ch := range currency {
		if ch < 'A' || ch > 'Z' {
			return false
		}
	}
	return true
}
