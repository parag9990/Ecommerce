package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

// RecipientProtector encrypts retry-required recipient addresses before they
// are persisted in delivery instructions.
type RecipientProtector struct {
	aead cipher.AEAD
}

func NewRecipientProtector(encodedKey string) (*RecipientProtector, error) {
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encodedKey))
	if err != nil {
		return nil, errors.New("notification delivery encryption key must be base64 encoded")
	}
	if len(key) != 32 {
		return nil, errors.New("notification delivery encryption key must decode to 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create notification delivery cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create notification delivery AEAD: %w", err)
	}
	return &RecipientProtector{aead: aead}, nil
}

func (p *RecipientProtector) Protect(recipient, deliveryID string) (string, error) {
	if p == nil || p.aead == nil {
		return "", errors.New("recipient protector is not configured")
	}
	recipient = strings.TrimSpace(recipient)
	deliveryID = strings.TrimSpace(deliveryID)
	if recipient == "" || deliveryID == "" {
		return "", errors.New("recipient and delivery ID are required for encryption")
	}
	nonce := make([]byte, p.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate recipient encryption nonce: %w", err)
	}
	encrypted := p.aead.Seal(nonce, nonce, []byte(recipient), []byte(deliveryID))
	return base64.RawURLEncoding.EncodeToString(encrypted), nil
}

func (p *RecipientProtector) Reveal(encoded, deliveryID string) (string, error) {
	if p == nil || p.aead == nil {
		return "", errors.New("recipient protector is not configured")
	}
	ciphertext, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil || len(ciphertext) < p.aead.NonceSize() {
		return "", errors.New("stored notification recipient is invalid")
	}
	nonce, ciphertext := ciphertext[:p.aead.NonceSize()], ciphertext[p.aead.NonceSize():]
	plaintext, err := p.aead.Open(nil, nonce, ciphertext, []byte(strings.TrimSpace(deliveryID)))
	if err != nil {
		return "", errors.New("stored notification recipient cannot be decrypted")
	}
	return string(plaintext), nil
}
