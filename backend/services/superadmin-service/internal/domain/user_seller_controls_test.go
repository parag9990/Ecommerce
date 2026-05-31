package domain_test

import (
	"testing"

	"ecommerce/superadmin-service/internal/domain"
)

func TestUserStatusTransitions(t *testing.T) {
	cases := []struct {
		name    string
		current domain.UserStatus
		next    domain.UserStatus
		allowed bool
	}{
		{name: "active can be blocked", current: domain.UserStatusActive, next: domain.UserStatusBlocked, allowed: true},
		{name: "blocked can be active", current: domain.UserStatusBlocked, next: domain.UserStatusActive, allowed: true},
		{name: "deleted cannot be restored", current: domain.UserStatusDeleted, next: domain.UserStatusActive, allowed: false},
		{name: "active cannot remain active", current: domain.UserStatusActive, next: domain.UserStatusActive, allowed: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := domain.CanUpdateUserStatus(tc.current, tc.next)
			if got != tc.allowed {
				t.Fatalf("CanUpdateUserStatus(%s, %s) = %v, want %v", tc.current, tc.next, got, tc.allowed)
			}
		})
	}
}

func TestSellerStatusTransitions(t *testing.T) {
	cases := []struct {
		name    string
		current domain.SellerStatus
		next    domain.SellerStatus
		allowed bool
	}{
		{name: "pending review can be approved", current: domain.SellerStatusPendingReview, next: domain.SellerStatusActive, allowed: true},
		{name: "pending review can be rejected", current: domain.SellerStatusPendingReview, next: domain.SellerStatusRejected, allowed: true},
		{name: "active can be suspended", current: domain.SellerStatusActive, next: domain.SellerStatusSuspended, allowed: true},
		{name: "suspended can be reinstated", current: domain.SellerStatusSuspended, next: domain.SellerStatusActive, allowed: true},
		{name: "draft cannot be approved by admin", current: domain.SellerStatusDraft, next: domain.SellerStatusActive, allowed: false},
		{name: "rejected cannot be directly approved", current: domain.SellerStatusRejected, next: domain.SellerStatusActive, allowed: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := domain.CanUpdateSellerStatus(tc.current, tc.next)
			if got != tc.allowed {
				t.Fatalf("CanUpdateSellerStatus(%s, %s) = %v, want %v", tc.current, tc.next, got, tc.allowed)
			}
		})
	}
}

func TestNormalizeMutationReason(t *testing.T) {
	if _, err := domain.NormalizeMutationReason("too short"); err == nil {
		t.Fatal("expected short reason to fail")
	}

	reason, err := domain.NormalizeMutationReason("  confirmed support investigation  ")
	if err != nil {
		t.Fatalf("NormalizeMutationReason returned error: %v", err)
	}
	if reason != "confirmed support investigation" {
		t.Fatalf("reason = %q", reason)
	}
}
