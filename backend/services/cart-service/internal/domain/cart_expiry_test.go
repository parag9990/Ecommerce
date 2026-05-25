package domain

import (
	"errors"
	"testing"
	"time"
)

func TestIsActiveCartExpiredUsesExpiresAtCutoff(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	cart := &Cart{
		ID:        "cart_123",
		UserID:    &userID,
		Status:    CartStatusActive,
		ExpiresAt: now,
	}

	if !IsActiveCartExpired(cart, now) {
		t.Fatal("IsActiveCartExpired() = false, want true when expires_at == now")
	}
	cart.ExpiresAt = now.Add(time.Second)
	if IsActiveCartExpired(cart, now) {
		t.Fatal("IsActiveCartExpired() = true, want false for future expires_at")
	}
	cart.Status = CartStatusCheckedOut
	cart.ExpiresAt = now.Add(-time.Second)
	if IsActiveCartExpired(cart, now) {
		t.Fatal("IsActiveCartExpired() = true, want false for non-active cart")
	}
}

func TestCartMarkExpired(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	userID := "user_123"
	cart := &Cart{
		ID:        "cart_123",
		UserID:    &userID,
		Status:    CartStatusActive,
		UpdatedAt: now.Add(-time.Hour),
		ExpiresAt: now.Add(-time.Second),
	}

	if err := cart.MarkExpired(now); err != nil {
		t.Fatalf("MarkExpired() error = %v", err)
	}
	if cart.Status != CartStatusExpired || !cart.UpdatedAt.Equal(now) {
		t.Fatalf("cart = %#v, want expired with updated_at", cart)
	}
	if err := cart.MarkExpired(now); !errors.Is(err, ErrCartNotActive) {
		t.Fatalf("MarkExpired() error = %v, want ErrCartNotActive", err)
	}
}

func TestExpiredCartCandidateOwner(t *testing.T) {
	userID := " user_123 "
	candidate := ExpiredCartCandidate{CartID: "cart_user", UserID: &userID}
	owner, ok := candidate.Owner()
	if !ok || !owner.IsUser() || owner.UserID != "user_123" {
		t.Fatalf("Owner() = %#v %v, want user owner", owner, ok)
	}

	guestSessionID := " sess_123 "
	candidate = ExpiredCartCandidate{CartID: "cart_guest", GuestSessionID: &guestSessionID}
	owner, ok = candidate.Owner()
	if !ok || !owner.IsGuest() || owner.GuestSessionID != "sess_123" {
		t.Fatalf("Owner() = %#v %v, want guest owner", owner, ok)
	}
}
