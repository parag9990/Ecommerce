package password

import (
	"context"
	"strings"
)

type RouterConfig struct {
	PreferredAlgorithm string
	Policy             Policy
	Argon2id           Argon2idParams
	Bcrypt             BcryptParams
}

func DefaultRouterConfig() RouterConfig {
	return RouterConfig{
		PreferredAlgorithm: AlgorithmArgon2id,
		Policy:             DefaultPolicy(),
		Argon2id:           DefaultArgon2idParams(),
		Bcrypt:             DefaultBcryptParams(),
	}
}

type Router struct {
	preferredAlgorithm string
	argon2id           *Argon2idHasher
	bcrypt             *BcryptHasher
}

func NewRouter(cfg RouterConfig) (*Router, error) {
	if cfg.PreferredAlgorithm == "" {
		cfg.PreferredAlgorithm = AlgorithmArgon2id
	}

	argonHasher, err := NewArgon2idHasher(cfg.Argon2id, cfg.Policy)
	if err != nil {
		return nil, err
	}
	bcryptHasher, err := NewBcryptHasher(cfg.Bcrypt, cfg.Policy)
	if err != nil {
		return nil, err
	}
	if cfg.PreferredAlgorithm != AlgorithmArgon2id && cfg.PreferredAlgorithm != AlgorithmBcrypt {
		return nil, ErrUnsupportedAlgorithm
	}

	return &Router{
		preferredAlgorithm: cfg.PreferredAlgorithm,
		argon2id:           argonHasher,
		bcrypt:             bcryptHasher,
	}, nil
}

func (r *Router) Hash(ctx context.Context, plain string) (HashResult, error) {
	switch r.preferredAlgorithm {
	case AlgorithmArgon2id:
		return r.argon2id.Hash(ctx, plain)
	case AlgorithmBcrypt:
		return r.bcrypt.Hash(ctx, plain)
	default:
		return HashResult{}, ErrUnsupportedAlgorithm
	}
}

func (r *Router) Verify(ctx context.Context, plain string, encodedHash string, algorithm string) (VerificationResult, error) {
	switch {
	case algorithm == AlgorithmArgon2id || strings.HasPrefix(encodedHash, "$argon2id$"):
		return r.argon2id.Verify(ctx, plain, encodedHash)
	case algorithm == AlgorithmBcrypt || isBcryptHash(encodedHash):
		result, err := r.bcrypt.Verify(ctx, plain, encodedHash)
		if result.Valid && r.preferredAlgorithm != AlgorithmBcrypt {
			result.NeedsRehash = true
		}
		return result, err
	default:
		return VerificationResult{}, ErrUnsupportedAlgorithm
	}
}

func isBcryptHash(encodedHash string) bool {
	return strings.HasPrefix(encodedHash, "$2a$") ||
		strings.HasPrefix(encodedHash, "$2b$") ||
		strings.HasPrefix(encodedHash, "$2x$") ||
		strings.HasPrefix(encodedHash, "$2y$")
}

var _ Hasher = (*Router)(nil)
