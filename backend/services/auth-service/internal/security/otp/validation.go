package otp

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

var e164Pattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

func NormalizeTarget(channel domain.OTPChannel, target string) (string, error) {
	switch channel {
	case domain.OTPChannelEmail:
		return NormalizeEmail(target)
	case domain.OTPChannelPhone:
		return NormalizePhone(target)
	default:
		return "", errors.New("unsupported otp channel")
	}
}

func NormalizeEmail(input string) (string, error) {
	target := strings.ToLower(strings.TrimSpace(input))
	if target == "" {
		return "", errors.New("email is required")
	}
	address, err := mail.ParseAddress(target)
	if err != nil || address.Address != target {
		return "", errors.New("invalid email")
	}
	return target, nil
}

func NormalizePhone(input string) (string, error) {
	target := strings.TrimSpace(input)
	if !e164Pattern.MatchString(target) {
		return "", errors.New("phone must be in E.164 format")
	}
	return target, nil
}

func ValidateCode(code string, length int) bool {
	code = strings.TrimSpace(code)
	if len(code) != length {
		return false
	}
	for _, ch := range code {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}
