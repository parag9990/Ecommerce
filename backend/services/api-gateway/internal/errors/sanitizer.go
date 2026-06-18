package gatewayerrors

import "strings"

const maxPublicMessageLength = 300

var unsafeFragments = []string{
	"panic:",
	"stack trace",
	"goroutine ",
	"rpc error:",
	"sql:",
	"mysql",
	"postgres",
	"mongo",
	"redis",
	"password",
	"passwd",
	"token",
	"secret",
	"authorization",
	"api key",
	"apikey",
	"connection refused",
	"dial tcp",
	".svc.",
	"localhost:",
}

func sanitizePublicMessage(msg string, fallback string) string {
	trimmed := strings.TrimSpace(msg)
	if trimmed == "" {
		return fallback
	}
	if len(trimmed) > maxPublicMessageLength || containsUnsafeFragment(trimmed) {
		return fallback
	}
	return trimmed
}

func sanitizeIdentifier(value string, maxLength int) string {
	value = strings.TrimSpace(value)
	if value == "" || maxLength <= 0 || len(value) > maxLength || containsUnsafeFragment(value) {
		return ""
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			continue
		}
		switch r {
		case '_', '-', '.', '/', '[', ']', ':':
			continue
		default:
			return ""
		}
	}
	return value
}

func normalizeReason(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		value = fallback
	}
	value = strings.ToLower(strings.ReplaceAll(value, " ", "_"))
	value = sanitizeIdentifier(value, 63)
	if value == "" {
		return fallback
	}
	return value
}

func unsafeKey(key string) bool {
	lower := strings.ToLower(key)
	return strings.Contains(lower, "password") ||
		strings.Contains(lower, "token") ||
		strings.Contains(lower, "secret") ||
		strings.Contains(lower, "authorization") ||
		strings.Contains(lower, "credential") ||
		strings.Contains(lower, "cookie") ||
		strings.Contains(lower, "otp") ||
		strings.Contains(lower, "card")
}

func containsUnsafeFragment(value string) bool {
	lower := strings.ToLower(value)
	for _, fragment := range unsafeFragments {
		if strings.Contains(lower, fragment) {
			return true
		}
	}
	return false
}
