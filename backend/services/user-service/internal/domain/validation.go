package domain

import (
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	maxIDLength              = 64
	maxEmailLength           = 255
	maxPhoneLength           = 32
	maxNameLength            = 255
	maxAvatarURLLength       = 1024
	maxAddressLineLength     = 255
	maxCityStateLength       = 128
	maxPostalCodeLength      = 32
	maxCountryLength         = 64
	maxGSTNumberLength       = 64
	maxDocumentTypeLength    = 64
	maxStorageURLLength      = 1024
	maxRejectionReasonLength = 512
)

var (
	idPattern      = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]*$`)
	phonePattern   = regexp.MustCompile(`^\+?[0-9][0-9 .()\-]{6,31}$`)
	gstPattern     = regexp.MustCompile(`^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z][1-9A-Z]Z[0-9A-Z]$`)
	controlPattern = regexp.MustCompile(`[\x00-\x1F\x7F]`)
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
	if value == "" {
		v.add(field, "is required")
		return
	}

	validateMaxRunes(v, field, value, maxLength)
	validateNoControlChars(v, field, value)
}

func validateOptionalString(v *validationCollector, field string, value *string, maxLength int) {
	if value == nil {
		return
	}

	validateMaxRunes(v, field, *value, maxLength)
	validateNoControlChars(v, field, *value)
}

func validateMaxRunes(v *validationCollector, field string, value string, maxLength int) {
	if utf8.RuneCountInString(value) > maxLength {
		v.add(field, "exceeds maximum length")
	}
}

func validateNoControlChars(v *validationCollector, field string, value string) {
	if controlPattern.MatchString(value) {
		v.add(field, "contains control characters")
	}
}

func validateID(v *validationCollector, field string, value string) {
	validateRequiredString(v, field, value, maxIDLength)
	if value != "" && !idPattern.MatchString(value) {
		v.add(field, "contains unsupported characters")
	}
}

func validateOptionalID(v *validationCollector, field string, value *string) {
	if value == nil {
		return
	}

	validateID(v, field, *value)
}

func validateEmail(v *validationCollector, field string, value string) {
	validateRequiredString(v, field, value, maxEmailLength)
	if value == "" {
		return
	}

	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value {
		v.add(field, "must be a valid email address")
	}
}

func validateOptionalEmail(v *validationCollector, field string, value *string) {
	if value == nil {
		return
	}

	validateEmail(v, field, *value)
}

func validateOptionalPhone(v *validationCollector, field string, value *string) {
	if value == nil {
		return
	}

	validateOptionalString(v, field, value, maxPhoneLength)
	if *value != "" && !phonePattern.MatchString(*value) {
		v.add(field, "must be a valid phone number")
	}
}

func validateOptionalHTTPURL(v *validationCollector, field string, value *string, maxLength int) {
	if value == nil {
		return
	}

	validateOptionalString(v, field, value, maxLength)
	parsed, err := url.ParseRequestURI(*value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		v.add(field, "must be an absolute URL")
		return
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		v.add(field, "must use http or https")
	}
}

func validateOptionalGSTNumber(v *validationCollector, field string, value *string) {
	if value == nil {
		return
	}

	validateOptionalString(v, field, value, maxGSTNumberLength)
	if *value != strings.ToUpper(*value) || !gstPattern.MatchString(*value) {
		v.add(field, "must be a valid GST number")
	}
}

func validateTimestamp(v *validationCollector, field string, value time.Time) {
	if value.IsZero() {
		v.add(field, "is required")
	}
}
