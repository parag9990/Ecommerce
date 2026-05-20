package otp

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

func randomBase64URL(byteLength int) (string, error) {
	if byteLength <= 0 {
		return "", fmt.Errorf("random byte length must be greater than zero")
	}

	raw := make([]byte, byteLength)
	if _, err := io.ReadFull(rand.Reader, raw); err != nil {
		return "", fmt.Errorf("generate secure random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
