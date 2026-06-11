package otp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

type Hasher struct {
	pepper []byte
}

func NewHasher(pepper string) (*Hasher, error) {
	pepper = strings.TrimSpace(pepper)
	if pepper == "" {
		return nil, errors.New("otp hash pepper is required")
	}
	return &Hasher{pepper: []byte(pepper)}, nil
}

func (h *Hasher) Hash(challengeID string, code string) (string, error) {
	if h == nil || len(h.pepper) == 0 {
		return "", errors.New("otp hasher is not initialized")
	}
	challengeID = strings.TrimSpace(challengeID)
	code = strings.TrimSpace(code)
	if challengeID == "" || code == "" {
		return "", errors.New("challenge id and otp are required")
	}

	mac := hmac.New(sha256.New, h.pepper)
	_, _ = mac.Write([]byte(challengeID))
	_, _ = mac.Write([]byte(":"))
	_, _ = mac.Write([]byte(code))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func (h *Hasher) Compare(challengeID string, code string, expectedHash string) (bool, error) {
	actual, err := h.Hash(challengeID, code)
	if err != nil {
		return false, err
	}
	expectedHash = strings.TrimSpace(expectedHash)
	if expectedHash == "" {
		return false, nil
	}
	return hmac.Equal([]byte(actual), []byte(expectedHash)), nil
}

func HashLookupKey(value string, pepper string) (string, error) {
	value = strings.TrimSpace(value)
	pepper = strings.TrimSpace(pepper)
	if value == "" {
		return "", errors.New("lookup key value is required")
	}
	if pepper == "" {
		return "", errors.New("lookup key pepper is required")
	}

	mac := hmac.New(sha256.New, []byte(pepper))
	_, _ = mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
