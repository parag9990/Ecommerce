package password

import (
	"context"
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func testArgonParams() Argon2idParams {
	return Argon2idParams{
		MemoryKiB:   1024,
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}
}

func TestArgon2idHashUsesUniqueSaltAndVerifies(t *testing.T) {
	ctx := context.Background()
	hasher, err := NewArgon2idHasher(testArgonParams(), DefaultPolicy())
	if err != nil {
		t.Fatalf("NewArgon2idHasher() error = %v", err)
	}

	first, err := hasher.Hash(ctx, "correct-horse-battery")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	second, err := hasher.Hash(ctx, "correct-horse-battery")
	if err != nil {
		t.Fatalf("Hash() second error = %v", err)
	}
	if first.EncodedHash == second.EncodedHash {
		t.Fatal("expected unique hashes for the same password")
	}
	if !strings.HasPrefix(first.EncodedHash, "$argon2id$v=19$") {
		t.Fatalf("hash format = %q", first.EncodedHash)
	}

	result, err := hasher.Verify(ctx, "correct-horse-battery", first.EncodedHash)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !result.Valid {
		t.Fatal("expected correct password to verify")
	}

	result, err = hasher.Verify(ctx, "wrong-password", first.EncodedHash)
	if err != nil {
		t.Fatalf("Verify() wrong password error = %v", err)
	}
	if result.Valid {
		t.Fatal("expected wrong password to fail")
	}
}

func TestArgon2idVerifyMalformedHash(t *testing.T) {
	hasher, err := NewArgon2idHasher(testArgonParams(), DefaultPolicy())
	if err != nil {
		t.Fatalf("NewArgon2idHasher() error = %v", err)
	}

	result, err := hasher.Verify(context.Background(), "password", "not-a-hash")
	if !errors.Is(err, ErrInvalidHashFormat) {
		t.Fatalf("Verify() error = %v, want ErrInvalidHashFormat", err)
	}
	if result.Valid {
		t.Fatal("malformed hash must not verify")
	}
}

func TestArgon2idNeedsRehashForOlderParams(t *testing.T) {
	ctx := context.Background()
	oldParams := testArgonParams()
	oldParams.MemoryKiB = 512
	oldHasher, err := NewArgon2idHasher(oldParams, DefaultPolicy())
	if err != nil {
		t.Fatalf("old hasher error = %v", err)
	}
	currentHasher, err := NewArgon2idHasher(testArgonParams(), DefaultPolicy())
	if err != nil {
		t.Fatalf("current hasher error = %v", err)
	}

	oldHash, err := oldHasher.Hash(ctx, "correct-horse-battery")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	result, err := currentHasher.Verify(ctx, "correct-horse-battery", oldHash.EncodedHash)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !result.Valid || !result.NeedsRehash {
		t.Fatalf("Verify() = %+v, want valid with rehash", result)
	}
}

func TestRouterVerifiesBcryptAndRequiresArgon2idRehash(t *testing.T) {
	ctx := context.Background()
	router, err := NewRouter(RouterConfig{
		PreferredAlgorithm: AlgorithmArgon2id,
		Policy:             DefaultPolicy(),
		Argon2id:           testArgonParams(),
		Bcrypt:             BcryptParams{Cost: bcrypt.MinCost},
	})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	bcryptHash, err := bcrypt.GenerateFromPassword([]byte("correct-horse-battery"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}

	result, err := router.Verify(ctx, "correct-horse-battery", string(bcryptHash), AlgorithmBcrypt)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !result.Valid || !result.NeedsRehash || result.Algorithm != AlgorithmBcrypt {
		t.Fatalf("Verify() = %+v, want bcrypt valid with rehash", result)
	}
}

func TestPolicyValidation(t *testing.T) {
	policy := Policy{MinLength: 8, MaxLength: 12}

	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{name: "blank", password: "   ", wantErr: ErrPasswordBlank},
		{name: "short", password: "short", wantErr: ErrPasswordTooShort},
		{name: "long", password: "this-password-is-too-long", wantErr: ErrPasswordTooLong},
		{name: "valid", password: "long-pass", wantErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.ValidatePassword(tt.password)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidatePassword() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
