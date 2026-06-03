package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	sharedvalidation "github.com/parag/ecommerce/backend/shared/validation"
)

const (
	maxIDLength              = sharedvalidation.MaxIDLength
	maxEmailLength           = sharedvalidation.MaxEmailLength
	maxPhoneLength           = sharedvalidation.MaxE164PhoneLength
	maxNameLength            = sharedvalidation.MaxShortNameLength
	maxAvatarURLLength       = sharedvalidation.MaxAvatarURLLength
	maxAddressLineLength     = sharedvalidation.MaxAddressLineLength
	maxCityStateLength       = sharedvalidation.MaxCityStateLength
	maxPostalCodeLength      = sharedvalidation.MaxPostalCodeLength
	maxCountryLength         = sharedvalidation.MaxCountryLength
	maxGSTNumberLength       = 64
	maxDocumentTypeLength    = 64
	maxStorageURLLength      = sharedvalidation.MaxAvatarURLLength
	maxRejectionReasonLength = 512
)

func trim(value string) string {
	return strings.TrimSpace(value)
}

func cleanOptional(value *string) *string {
	if value == nil {
		return nil
	}

	cleaned := trim(*value)
	if cleaned == "" {
		return nil
	}

	return &cleaned
}

func validateRequiredString(v *validationCollector, field string, value string, maxLength int) {
	validateRequiredSafeText(v, field, value, 1, maxLength)
}

func validateOptionalString(v *validationCollector, field string, value *string, maxLength int) {
	if value == nil {
		return
	}

	validateOptionalSafeText(v, field, *value, 0, maxLength)
}

func validateRequiredSafeText(v *validationCollector, field string, value string, minLength int, maxLength int) {
	var validationErr sharedvalidation.Error
	sharedvalidation.ValidateRequiredSafeText(&validationErr, field, value, minLength, maxLength)
	v.addAll(validationErr.Fields)
}

func validateOptionalSafeText(v *validationCollector, field string, value string, minLength int, maxLength int) {
	var validationErr sharedvalidation.Error
	sharedvalidation.ValidateOptionalSafeText(&validationErr, field, value, minLength, maxLength)
	v.addAll(validationErr.Fields)
}

func validateMaxRunes(v *validationCollector, field string, value string, maxLength int) {
	if utf8.RuneCountInString(value) > maxLength {
		v.add(field, "exceeds maximum length")
	}
}

func validateNoControlChars(v *validationCollector, field string, value string) {
	if !sharedvalidation.IsSafeText(value) {
		v.add(field, "contains unsupported characters")
	}
}

func validateID(v *validationCollector, field string, value string) {
	var validationErr sharedvalidation.Error
	sharedvalidation.ValidateID(&validationErr, field, value)
	v.addAll(validationErr.Fields)
}

func validateOptionalID(v *validationCollector, field string, value *string) {
	if value == nil {
		return
	}

	validateID(v, field, *value)
}

func validateEmail(v *validationCollector, field string, value string) {
	var validationErr sharedvalidation.Error
	sharedvalidation.ValidateEmail(&validationErr, field, value)
	v.addAll(validationErr.Fields)
}

func validateOptionalEmail(v *validationCollector, field string, value *string) {
	if value == nil {
		return
	}

	var validationErr sharedvalidation.Error
	sharedvalidation.ValidateOptionalEmail(&validationErr, field, *value)
	v.addAll(validationErr.Fields)
}

func validateOptionalPhone(v *validationCollector, field string, value *string) {
	if value == nil {
		return
	}

	var validationErr sharedvalidation.Error
	sharedvalidation.ValidateOptionalPhone(&validationErr, field, *value, sharedvalidation.DefaultPhoneRegion)
	v.addAll(validationErr.Fields)
}

func validateOptionalHTTPURL(v *validationCollector, field string, value *string, maxLength int) {
	if value == nil {
		return
	}

	var validationErr sharedvalidation.Error
	sharedvalidation.ValidateOptionalHTTPSURL(&validationErr, field, *value, maxLength)
	v.addAll(validationErr.Fields)
}

func validateRequiredHTTPSURL(v *validationCollector, field string, value string, maxLength int) {
	if strings.TrimSpace(value) == "" {
		v.add(field, "is required")
		return
	}

	var validationErr sharedvalidation.Error
	sharedvalidation.ValidateOptionalHTTPSURL(&validationErr, field, value, maxLength)
	v.addAll(validationErr.Fields)
}

func validateOptionalGSTNumber(v *validationCollector, field string, value *string) {
	if value == nil {
		return
	}

	validateMaxRunes(v, field, *value, maxGSTNumberLength)
	var validationErr sharedvalidation.Error
	sharedvalidation.ValidateOptionalGSTIN(&validationErr, field, *value)
	v.addAll(validationErr.Fields)
}

func validatePostalCode(v *validationCollector, field string, postalCode string, country string) {
	var validationErr sharedvalidation.Error
	sharedvalidation.ValidatePostalCode(&validationErr, field, postalCode, country)
	v.addAll(validationErr.Fields)
}

func validateTimestamp(v *validationCollector, field string, value time.Time) {
	if value.IsZero() {
		v.add(field, "is required")
	}
}
