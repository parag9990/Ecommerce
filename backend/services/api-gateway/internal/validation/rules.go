package validation

import (
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode"
)

var (
	publicIDPattern    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_:-]{2,127}$`)
	providerPattern    = regexp.MustCompile(`^[a-z][a-z0-9_-]{1,39}$`)
	settingKeyPattern  = regexp.MustCompile(`^[A-Za-z0-9_.:-]{1,120}$`)
	couponCodePattern  = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
	skuPattern         = regexp.MustCompile(`^[A-Za-z0-9_.:-]{2,80}$`)
	currencyPattern    = regexp.MustCompile(`^[A-Z]{3}$`)
	statusValuePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{1,40}$`)
)

func IsValidPublicID(value string) bool {
	return publicIDPattern.MatchString(strings.TrimSpace(value))
}

func IsValidProvider(value string) bool {
	return providerPattern.MatchString(strings.TrimSpace(value))
}

func IsValidSettingKey(value string) bool {
	return settingKeyPattern.MatchString(strings.TrimSpace(value))
}

func IsValidIdempotencyKey(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) < 16 || len(value) > 128 {
		return false
	}
	for _, r := range value {
		if r <= 32 || r == 127 || unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func IsValidEmail(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) > 255 {
		return false
	}
	_, err := mail.ParseAddress(value)
	return err == nil && !strings.Contains(value, "\n")
}

func IsValidDateTime(value string) bool {
	_, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	return err == nil
}

func IsValidCurrency(value string) bool {
	return currencyPattern.MatchString(strings.TrimSpace(value))
}

func safeStringValue(value any) string {
	text, _ := value.(string)
	return strings.TrimSpace(text)
}
