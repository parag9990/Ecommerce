package password

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	DefaultArgon2idMemoryKiB   = 64 * 1024
	DefaultArgon2idIterations  = 3
	DefaultArgon2idParallelism = 2
	DefaultArgon2idSaltLength  = 16
	DefaultArgon2idKeyLength   = 32
)

type Argon2idParams struct {
	MemoryKiB   uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

func DefaultArgon2idParams() Argon2idParams {
	return Argon2idParams{
		MemoryKiB:   DefaultArgon2idMemoryKiB,
		Iterations:  DefaultArgon2idIterations,
		Parallelism: DefaultArgon2idParallelism,
		SaltLength:  DefaultArgon2idSaltLength,
		KeyLength:   DefaultArgon2idKeyLength,
	}
}

func (p Argon2idParams) Validate() error {
	if p.MemoryKiB == 0 {
		return errors.New("memory must be greater than zero")
	}
	if p.Iterations == 0 {
		return errors.New("iterations must be greater than zero")
	}
	if p.Parallelism == 0 {
		return errors.New("parallelism must be greater than zero")
	}
	if p.SaltLength == 0 {
		return errors.New("salt length must be greater than zero")
	}
	if p.KeyLength == 0 {
		return errors.New("key length must be greater than zero")
	}
	return nil
}

type Argon2idHasher struct {
	params Argon2idParams
	policy Policy
}

func NewArgon2idHasher(params Argon2idParams, policy Policy) (*Argon2idHasher, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	return &Argon2idHasher{params: params, policy: policy}, nil
}

func (h *Argon2idHasher) Hash(ctx context.Context, plain string) (HashResult, error) {
	if err := ctx.Err(); err != nil {
		return HashResult{}, err
	}
	if err := h.policy.ValidatePassword(plain); err != nil {
		return HashResult{}, err
	}

	salt, err := randomBytes(h.params.SaltLength)
	if err != nil {
		return HashResult{}, err
	}

	key := argon2.IDKey(
		[]byte(plain),
		salt,
		h.params.Iterations,
		h.params.MemoryKiB,
		h.params.Parallelism,
		h.params.KeyLength,
	)
	if err := ctx.Err(); err != nil {
		return HashResult{}, err
	}

	encodedHash := fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		h.params.MemoryKiB,
		h.params.Iterations,
		h.params.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	)

	return HashResult{EncodedHash: encodedHash, Algorithm: AlgorithmArgon2id}, nil
}

func (h *Argon2idHasher) Verify(ctx context.Context, plain string, encodedHash string) (VerificationResult, error) {
	if err := ctx.Err(); err != nil {
		return VerificationResult{}, err
	}

	params, salt, expectedKey, err := decodeArgon2idHash(encodedHash)
	if err != nil {
		return VerificationResult{Algorithm: AlgorithmArgon2id}, err
	}

	actualKey := argon2.IDKey(
		[]byte(plain),
		salt,
		params.Iterations,
		params.MemoryKiB,
		params.Parallelism,
		uint32(len(expectedKey)),
	)
	if err := ctx.Err(); err != nil {
		return VerificationResult{}, err
	}

	valid := subtle.ConstantTimeCompare(actualKey, expectedKey) == 1
	return VerificationResult{
		Valid:       valid,
		NeedsRehash: valid && h.needsRehash(params),
		Algorithm:   AlgorithmArgon2id,
	}, nil
}

func (h *Argon2idHasher) needsRehash(params Argon2idParams) bool {
	return params.MemoryKiB != h.params.MemoryKiB ||
		params.Iterations != h.params.Iterations ||
		params.Parallelism != h.params.Parallelism ||
		params.SaltLength != h.params.SaltLength ||
		params.KeyLength != h.params.KeyLength
}

func decodeArgon2idHash(encodedHash string) (Argon2idParams, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[0] != "" {
		return Argon2idParams{}, nil, nil, ErrInvalidHashFormat
	}
	if parts[1] != AlgorithmArgon2id {
		return Argon2idParams{}, nil, nil, ErrUnsupportedAlgorithm
	}
	if parts[2] != "v=19" {
		return Argon2idParams{}, nil, nil, ErrInvalidHashFormat
	}

	params, err := parseArgon2idParams(parts[3])
	if err != nil {
		return Argon2idParams{}, nil, nil, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return Argon2idParams{}, nil, nil, fmt.Errorf("%w: invalid salt encoding", ErrInvalidHashFormat)
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return Argon2idParams{}, nil, nil, fmt.Errorf("%w: invalid key encoding", ErrInvalidHashFormat)
	}
	if len(salt) == 0 || len(key) == 0 {
		return Argon2idParams{}, nil, nil, ErrInvalidHashFormat
	}

	params.SaltLength = uint32(len(salt))
	params.KeyLength = uint32(len(key))

	return params, salt, key, nil
}

func parseArgon2idParams(raw string) (Argon2idParams, error) {
	values := make(map[string]string, 3)
	for _, part := range strings.Split(raw, ",") {
		keyValue := strings.SplitN(part, "=", 2)
		if len(keyValue) != 2 {
			return Argon2idParams{}, ErrInvalidHashFormat
		}
		values[keyValue[0]] = keyValue[1]
	}

	memory, err := parseUint32(values, "m")
	if err != nil {
		return Argon2idParams{}, err
	}
	iterations, err := parseUint32(values, "t")
	if err != nil {
		return Argon2idParams{}, err
	}
	parallelism, err := parseUint8(values, "p")
	if err != nil {
		return Argon2idParams{}, err
	}

	params := Argon2idParams{
		MemoryKiB:   memory,
		Iterations:  iterations,
		Parallelism: parallelism,
	}
	if params.MemoryKiB == 0 || params.Iterations == 0 || params.Parallelism == 0 {
		return Argon2idParams{}, ErrInvalidHashFormat
	}
	return params, nil
}

func parseUint32(values map[string]string, key string) (uint32, error) {
	raw, ok := values[key]
	if !ok {
		return 0, ErrInvalidHashFormat
	}
	parsed, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid %s", ErrInvalidHashFormat, key)
	}
	return uint32(parsed), nil
}

func parseUint8(values map[string]string, key string) (uint8, error) {
	raw, ok := values[key]
	if !ok {
		return 0, ErrInvalidHashFormat
	}
	parsed, err := strconv.ParseUint(raw, 10, 8)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid %s", ErrInvalidHashFormat, key)
	}
	return uint8(parsed), nil
}
