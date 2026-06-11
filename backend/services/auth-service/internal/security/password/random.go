package password

import (
	"crypto/rand"
	"fmt"
	"io"
)

func randomBytes(length uint32) ([]byte, error) {
	if length == 0 {
		return nil, fmt.Errorf("random byte length must be greater than zero")
	}

	b := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return nil, fmt.Errorf("generate secure random bytes: %w", err)
	}
	return b, nil
}
