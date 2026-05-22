package device

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

type PrivacyHasher struct {
	pepper []byte
}

func NewPrivacyHasher(pepper string) PrivacyHasher {
	return PrivacyHasher{pepper: []byte(strings.TrimSpace(pepper))}
}

func (h PrivacyHasher) Hash(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || len(h.pepper) == 0 {
		return ""
	}
	mac := hmac.New(sha256.New, h.pepper)
	_, _ = mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}
