package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

const RefreshTokenEntropyBytes = 32

var (
	ErrRefreshTokenBlank  = errors.New("refresh token is required")
	ErrRefreshPepperBlank = errors.New("refresh token pepper is required")
)

func NewRefreshToken() (string, error) {
	return randomBase64URL(RefreshTokenEntropyBytes)
}

func HashRefreshToken(plainToken string, pepper string) (string, error) {
	plainToken = strings.TrimSpace(plainToken)
	if plainToken == "" {
		return "", ErrRefreshTokenBlank
	}
	if pepper == "" {
		return "", ErrRefreshPepperBlank
	}

	mac := hmac.New(sha256.New, []byte(pepper))
	_, _ = mac.Write([]byte(plainToken))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
