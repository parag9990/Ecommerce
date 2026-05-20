package sessionlink

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

var ErrPrivacyPepperRequired = errors.New("session event pepper is required")

type PrivacyHasher struct {
	pepper []byte
}

func NewPrivacyHasher(pepper string) (*PrivacyHasher, error) {
	pepper = strings.TrimSpace(pepper)
	if pepper == "" {
		return nil, ErrPrivacyPepperRequired
	}
	return &PrivacyHasher{pepper: []byte(pepper)}, nil
}

func (h *PrivacyHasher) Hash(value string) string {
	if h == nil {
		return ""
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	mac := hmac.New(sha256.New, h.pepper)
	_, _ = mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

func (h *PrivacyHasher) HashIP(value string) string {
	return h.Hash(value)
}

func (h *PrivacyHasher) HashDeviceFingerprint(value string) string {
	return h.Hash(value)
}

type DisabledPrivacyHasher struct{}

func NewDisabledPrivacyHasher() DisabledPrivacyHasher {
	return DisabledPrivacyHasher{}
}

func (DisabledPrivacyHasher) HashIP(string) string {
	return ""
}

func (DisabledPrivacyHasher) HashDeviceFingerprint(string) string {
	return ""
}
