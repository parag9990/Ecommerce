package identifier

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const idRandomByteLength = 16

type CryptoGenerator struct{}

func NewCryptoGenerator() *CryptoGenerator {
	return &CryptoGenerator{}
}

func (CryptoGenerator) NewID(prefix string) (string, error) {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" || len(prefix) > 12 {
		return "", errors.New("id prefix must be between 1 and 12 characters")
	}
	for _, char := range prefix {
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') {
			return "", errors.New("id prefix must use lowercase alphanumeric characters")
		}
	}
	randomBytes := make([]byte, idRandomByteLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate secure random id: %w", err)
	}
	return prefix + "_" + hex.EncodeToString(randomBytes), nil
}
