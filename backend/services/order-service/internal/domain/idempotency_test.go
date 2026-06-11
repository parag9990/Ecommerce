package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestValidateIdempotencyKey(t *testing.T) {
	if !errors.Is(ValidateIdempotencyKey("  "), ErrIdempotencyKeyRequired) {
		t.Fatal("ValidateIdempotencyKey(blank) did not require a key")
	}
	if !errors.Is(ValidateIdempotencyKey(strings.Repeat("x", 129)), ErrIdempotencyKeyInvalid) {
		t.Fatal("ValidateIdempotencyKey(oversized) did not reject storage overflow")
	}
	if err := ValidateIdempotencyKey("checkout_01JABC123XYZ"); err != nil {
		t.Fatalf("ValidateIdempotencyKey(valid) error = %v", err)
	}
}

func TestIdempotencyClaimValidateForOrderBindingRequiresAcquiredMatchingClaim(t *testing.T) {
	claim := IdempotencyClaim{
		UserID: "user_1", Key: "key_1", RequestHash: strings.Repeat("a", 64),
		Status: IdempotencyStatusProcessing, Decision: ClaimAcquired, ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := claim.ValidateForOrderBinding("user_1", "key_1"); err != nil {
		t.Fatalf("ValidateForOrderBinding(valid) error = %v", err)
	}
	claim.Decision = ClaimReplay
	if !errors.Is(claim.ValidateForOrderBinding("user_1", "key_1"), ErrInvalidCheckoutCommand) {
		t.Fatal("ValidateForOrderBinding(replay) allowed non-winner order creation")
	}
}
