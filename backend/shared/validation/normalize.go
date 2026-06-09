package validation

import (
	"strings"
)

const DefaultPhoneRegion = "IN"

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func NormalizeText(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func NormalizeOptionalText(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := NormalizeText(*value)
	return &normalized
}

func NormalizeTrimmed(value string) string {
	return strings.TrimSpace(value)
}

func NormalizeOptionalTrimmed(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := NormalizeTrimmed(*value)
	return &normalized
}

func NormalizeUpper(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func NormalizeOptionalUpper(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := NormalizeUpper(*value)
	return &normalized
}

func NormalizeCountry(country string) string {
	normalized := NormalizeText(country)
	if len(normalized) == 2 {
		return strings.ToUpper(normalized)
	}
	return normalized
}
