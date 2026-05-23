package observability

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
)

const redactedValue = "[REDACTED]"

var sensitiveFieldFragments = []string{
	"authorization",
	"password",
	"passwd",
	"secret",
	"token",
	"refresh",
	"otp",
	"cvv",
	"card",
	"cookie",
	"set-cookie",
	"signature",
}

func HashUserID(userID, salt string) string {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(strings.TrimSpace(salt) + ":" + userID))
	return "u_" + hex.EncodeToString(sum[:])[:12]
}

func RedactField(key string, value any) any {
	if IsSensitiveField(key) {
		return redactedValue
	}
	return value
}

func RedactHeaders(headers http.Header) map[string]string {
	if len(headers) == 0 {
		return nil
	}
	out := make(map[string]string, len(headers))
	for key, values := range headers {
		if IsSensitiveField(key) {
			out[http.CanonicalHeaderKey(key)] = redactedValue
			continue
		}
		if len(values) == 0 {
			continue
		}
		out[http.CanonicalHeaderKey(key)] = values[0]
	}
	return out
}

func IsSensitiveField(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	if normalized == "" {
		return false
	}
	for _, fragment := range sensitiveFieldFragments {
		if strings.Contains(normalized, fragment) {
			return true
		}
	}
	return false
}
