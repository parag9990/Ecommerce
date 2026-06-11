package otp

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

const (
	defaultChallengeIDEntropyBytes = 18
	challengeIDPrefix              = "otp_chal_"
)

type Generator struct {
	length int
	max    *big.Int
	format string
}

func NewGenerator(length int) (*Generator, error) {
	if length < 6 || length > 10 {
		return nil, fmt.Errorf("otp length must be between 6 and 10 digits")
	}

	max := big.NewInt(1)
	for i := 0; i < length; i++ {
		max.Mul(max, big.NewInt(10))
	}

	return &Generator{
		length: length,
		max:    max,
		format: fmt.Sprintf("%%0%dd", length),
	}, nil
}

func (g *Generator) Generate() (string, error) {
	if g == nil || g.max == nil {
		return "", fmt.Errorf("otp generator is not initialized")
	}

	n, err := rand.Int(rand.Reader, g.max)
	if err != nil {
		return "", fmt.Errorf("generate otp: %w", err)
	}
	return fmt.Sprintf(g.format, n.Int64()), nil
}

func NewChallengeID() (string, error) {
	value, err := randomBase64URL(defaultChallengeIDEntropyBytes)
	if err != nil {
		return "", err
	}
	return challengeIDPrefix + value, nil
}

func ValidateChallengeID(challengeID string) bool {
	challengeID = strings.TrimSpace(challengeID)
	if len(challengeID) <= len(challengeIDPrefix) || len(challengeID) > 64 {
		return false
	}
	if !strings.HasPrefix(challengeID, challengeIDPrefix) {
		return false
	}
	for _, ch := range challengeID {
		if ch >= 'a' && ch <= 'z' {
			continue
		}
		if ch >= 'A' && ch <= 'Z' {
			continue
		}
		if ch >= '0' && ch <= '9' {
			continue
		}
		if ch == '_' || ch == '-' {
			continue
		}
		return false
	}
	return true
}
