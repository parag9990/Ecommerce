package domain

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"strings"
)

const sessionReferencePrefix = "ref_"

type SessionReferenceCodec struct {
	aead cipher.AEAD
}

func NewSessionReferenceCodec(secret string) (*SessionReferenceCodec, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("session reference secret is required")
	}
	key := sha256.Sum256([]byte("session-analytics-reference:" + secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &SessionReferenceCodec{aead: aead}, nil
}

func (c *SessionReferenceCodec) Protect(sessionID string) (string, error) {
	if c == nil || c.aead == nil {
		return "", errors.New("session reference codec is not configured")
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return "", errors.New("session id is required")
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := c.aead.Seal(nonce, nonce, []byte(sessionID), []byte(sessionReferencePrefix))
	return sessionReferencePrefix + base64.RawURLEncoding.EncodeToString(sealed), nil
}

func (c *SessionReferenceCodec) Unprotect(reference string) (string, error) {
	if c == nil || c.aead == nil {
		return "", errors.New("session reference codec is not configured")
	}
	if !strings.HasPrefix(reference, sessionReferencePrefix) {
		return "", errors.New("invalid session reference")
	}
	encoded := strings.TrimPrefix(reference, sessionReferencePrefix)
	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil || len(payload) <= c.aead.NonceSize() {
		return "", errors.New("invalid session reference")
	}
	nonce, ciphertext := payload[:c.aead.NonceSize()], payload[c.aead.NonceSize():]
	plain, err := c.aead.Open(nil, nonce, ciphertext, []byte(sessionReferencePrefix))
	if err != nil {
		return "", errors.New("invalid session reference")
	}
	value := strings.TrimSpace(string(plain))
	if value == "" {
		return "", errors.New("invalid session reference")
	}
	return value, nil
}

func IsSessionReference(value string) bool {
	return strings.HasPrefix(strings.TrimSpace(value), sessionReferencePrefix)
}
