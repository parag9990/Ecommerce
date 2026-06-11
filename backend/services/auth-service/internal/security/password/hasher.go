package password

import (
	"context"
	"errors"
)

const (
	AlgorithmArgon2id = "argon2id"
	AlgorithmBcrypt   = "bcrypt"
)

var (
	ErrUnsupportedAlgorithm = errors.New("unsupported password algorithm")
	ErrInvalidHashFormat    = errors.New("invalid password hash format")
)

type HashResult struct {
	EncodedHash string
	Algorithm   string
}

type VerificationResult struct {
	Valid       bool
	NeedsRehash bool
	Algorithm   string
}

type Hasher interface {
	Hash(ctx context.Context, plain string) (HashResult, error)
	Verify(ctx context.Context, plain string, encodedHash string, algorithm string) (VerificationResult, error)
}
