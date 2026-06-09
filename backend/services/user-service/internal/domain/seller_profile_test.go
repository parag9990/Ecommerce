package domain

import (
	"errors"
	"testing"
	"time"
)

func TestSellerProfileLifecycle(t *testing.T) {
	profile, err := NewSellerProfile(NewSellerProfileParams{
		SellerID:     "seller_123",
		UserID:       "user_123",
		StoreName:    "Aarav Retail Pvt Ltd",
		DisplayName:  strptr("Aarav Retail"),
		GSTNumber:    strptr("27abcde1234f1z5"),
		SupportEmail: strptr("support@aaravretail.example"),
		CreatedBy:    "user_123",
		CreatedAt:    fixedTime(),
	})
	if err != nil {
		t.Fatalf("NewSellerProfile returned error: %v", err)
	}

	if profile.GSTNumber == nil || *profile.GSTNumber != "27ABCDE1234F1Z5" {
		t.Fatalf("expected uppercase GST number, got %#v", profile.GSTNumber)
	}
	if err := profile.SubmitForReview("user_123", fixedTime().Add(time.Minute)); err != nil {
		t.Fatalf("SubmitForReview returned error: %v", err)
	}
	if err := profile.Approve("admin_123", fixedTime().Add(2*time.Minute)); err != nil {
		t.Fatalf("Approve returned error: %v", err)
	}
	if err := profile.Suspend("admin_123", "policy violation", fixedTime().Add(3*time.Minute)); err != nil {
		t.Fatalf("Suspend returned error: %v", err)
	}
	if err := profile.Reactivate("admin_123", "resolved", fixedTime().Add(4*time.Minute)); err != nil {
		t.Fatalf("Reactivate returned error: %v", err)
	}

	if profile.Status != SellerStatusActive {
		t.Fatalf("expected active seller, got %q", profile.Status)
	}
	if profile.ApprovedBy == nil || *profile.ApprovedBy != "admin_123" {
		t.Fatalf("expected approval metadata to be preserved, got %#v", profile.ApprovedBy)
	}
}

func TestSellerProfileRejectsInvalidTransition(t *testing.T) {
	profile, err := NewSellerProfile(NewSellerProfileParams{
		SellerID:  "seller_123",
		UserID:    "user_123",
		StoreName: "Aarav Retail Pvt Ltd",
		CreatedBy: "user_123",
		CreatedAt: fixedTime(),
	})
	if err != nil {
		t.Fatalf("NewSellerProfile returned error: %v", err)
	}

	err = profile.Approve("admin_123", fixedTime().Add(time.Minute))
	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("expected invalid transition error, got %v", err)
	}
}
