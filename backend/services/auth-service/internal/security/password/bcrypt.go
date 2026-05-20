package password

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const (
	DefaultBcryptCost     = 12
	bcryptMaxPasswordSize = 72
)

var ErrPasswordTooLongForBcrypt = errors.New("password exceeds bcrypt 72 byte limit")

type BcryptParams struct {
	Cost int
}

func DefaultBcryptParams() BcryptParams {
	return BcryptParams{Cost: DefaultBcryptCost}
}

func (p BcryptParams) Validate() error {
	if p.Cost < bcrypt.MinCost || p.Cost > bcrypt.MaxCost {
		return errors.New("bcrypt cost must be within bcrypt supported range")
	}
	return nil
}

type BcryptHasher struct {
	params BcryptParams
	policy Policy
}

func NewBcryptHasher(params BcryptParams, policy Policy) (*BcryptHasher, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	return &BcryptHasher{params: params, policy: policy}, nil
}

func (h *BcryptHasher) Hash(ctx context.Context, plain string) (HashResult, error) {
	if err := ctx.Err(); err != nil {
		return HashResult{}, err
	}
	if err := h.policy.ValidatePassword(plain); err != nil {
		return HashResult{}, err
	}
	if len([]byte(plain)) > bcryptMaxPasswordSize {
		return HashResult{}, ErrPasswordTooLongForBcrypt
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plain), h.params.Cost)
	if err != nil {
		return HashResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return HashResult{}, err
	}

	return HashResult{EncodedHash: string(hash), Algorithm: AlgorithmBcrypt}, nil
}

func (h *BcryptHasher) Verify(ctx context.Context, plain string, encodedHash string) (VerificationResult, error) {
	if err := ctx.Err(); err != nil {
		return VerificationResult{}, err
	}
	if len([]byte(plain)) > bcryptMaxPasswordSize {
		return VerificationResult{Algorithm: AlgorithmBcrypt}, nil
	}

	err := bcrypt.CompareHashAndPassword([]byte(encodedHash), []byte(plain))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return VerificationResult{Algorithm: AlgorithmBcrypt}, nil
		}
		return VerificationResult{Algorithm: AlgorithmBcrypt}, err
	}

	cost, err := bcrypt.Cost([]byte(encodedHash))
	if err != nil {
		return VerificationResult{Valid: true, Algorithm: AlgorithmBcrypt}, nil
	}

	return VerificationResult{
		Valid:       true,
		NeedsRehash: cost < h.params.Cost,
		Algorithm:   AlgorithmBcrypt,
	}, nil
}
