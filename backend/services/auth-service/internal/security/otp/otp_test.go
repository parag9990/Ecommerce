package otp

import (
	"testing"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

func TestGeneratorProducesFixedLengthDigits(t *testing.T) {
	generator, err := NewGenerator(6)
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	for i := 0; i < 100; i++ {
		code, err := generator.Generate()
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		if !ValidateCode(code, 6) {
			t.Fatalf("generated invalid code %q", code)
		}
	}
}

func TestHasherUsesChallengeIDAndPepper(t *testing.T) {
	hasher, err := NewHasher("pepper-one")
	if err != nil {
		t.Fatalf("NewHasher() error = %v", err)
	}

	hashA, err := hasher.Hash("otp_chal_a", "123456")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	hashB, err := hasher.Hash("otp_chal_b", "123456")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if hashA == hashB {
		t.Fatal("same OTP under different challenges produced identical hash")
	}

	matched, err := hasher.Compare("otp_chal_a", "123456", hashA)
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	if !matched {
		t.Fatal("expected hash comparison to match")
	}
}

func TestNormalizeTarget(t *testing.T) {
	email, err := NormalizeTarget(domain.OTPChannelEmail, " User@Example.COM ")
	if err != nil {
		t.Fatalf("NormalizeTarget(email) error = %v", err)
	}
	if email != "user@example.com" {
		t.Fatalf("normalized email = %q", email)
	}

	phone, err := NormalizeTarget(domain.OTPChannelPhone, " +14155550100 ")
	if err != nil {
		t.Fatalf("NormalizeTarget(phone) error = %v", err)
	}
	if phone != "+14155550100" {
		t.Fatalf("normalized phone = %q", phone)
	}
}
