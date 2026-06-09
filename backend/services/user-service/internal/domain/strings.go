package domain

import (
	"strings"

	sharedvalidation "github.com/parag/ecommerce/backend/shared/validation"
)

func cleanText(value string) string {
	return sharedvalidation.NormalizeText(value)
}

func cleanCountry(value string) string {
	return sharedvalidation.NormalizeCountry(value)
}

func sharedNormalizeEmail(value string) string {
	return sharedvalidation.NormalizeEmail(value)
}

func cleanOptionalText(value *string) *string {
	if value == nil {
		return nil
	}
	cleaned := cleanText(*value)
	if cleaned == "" {
		return nil
	}
	return &cleaned
}

func cleanOptionalUpper(value *string) *string {
	cleaned := cleanOptional(value)
	if cleaned == nil {
		return nil
	}

	upper := strings.ToUpper(*cleaned)
	return &upper
}

func cleanOptionalPhone(value *string) *string {
	cleaned := cleanOptional(value)
	if cleaned == nil {
		return nil
	}
	normalized, ok := sharedvalidation.NormalizePhone(*cleaned, sharedvalidation.DefaultPhoneRegion)
	if !ok {
		return cleaned
	}
	return &normalized
}

func cleanOptionalEmail(value *string) *string {
	cleaned := cleanOptional(value)
	if cleaned == nil {
		return nil
	}
	normalized := sharedvalidation.NormalizeEmail(*cleaned)
	return &normalized
}
