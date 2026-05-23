package ratelimit

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func BuildKey(prefix string, dimension Dimension, policyName string, identity string) string {
	parts := []string{
		strings.Trim(strings.TrimSpace(prefix), ":"),
		sanitizeKeyPart(string(dimension)),
		sanitizeKeyPart(policyName),
		HashIdentity(identity),
	}
	return strings.Join(parts, ":")
}

func HashIdentity(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	if normalized == "" {
		normalized = "unknown"
	}
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])[:32]
}

func sanitizeKeyPart(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "unknown"
	}
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '.', r == '_', r == '-':
			return r
		default:
			return '_'
		}
	}, value)
}
