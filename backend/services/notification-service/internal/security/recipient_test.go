package security

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestRecipientProtectorRoundTripAndBinding(t *testing.T) {
	t.Parallel()

	key := base64.StdEncoding.EncodeToString([]byte("01234567890123456789012345678901"))
	protector, err := NewRecipientProtector(key)
	if err != nil {
		t.Fatalf("NewRecipientProtector() error = %v", err)
	}
	ciphertext, err := protector.Protect("buyer@example.com", "delivery_1")
	if err != nil {
		t.Fatalf("Protect() error = %v", err)
	}
	if strings.Contains(ciphertext, "buyer") {
		t.Fatal("Protect() exposed plaintext recipient")
	}
	got, err := protector.Reveal(ciphertext, "delivery_1")
	if err != nil || got != "buyer@example.com" {
		t.Fatalf("Reveal() = %q, %v", got, err)
	}
	if _, err := protector.Reveal(ciphertext, "delivery_2"); err == nil {
		t.Fatal("Reveal() succeeded with another delivery binding")
	}
}

func TestNewRecipientProtectorRejectsInvalidKey(t *testing.T) {
	t.Parallel()

	if _, err := NewRecipientProtector(base64.StdEncoding.EncodeToString([]byte("short"))); err == nil {
		t.Fatal("NewRecipientProtector() returned nil error for short key")
	}
}
