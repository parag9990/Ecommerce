package domain

import (
	"errors"
	"testing"
	"time"
)

func strptr(value string) *string {
	return &value
}

func fixedTime() time.Time {
	return time.Date(2026, 5, 21, 10, 30, 0, 0, time.UTC)
}

func TestNewUserNormalizesProfileFields(t *testing.T) {
	user, err := NewUser(NewUserParams{
		UserID:        " user_123 ",
		AuthAccountID: " auth_123 ",
		Email:         "buyer@example.com",
		Phone:         strptr(" +919999999999 "),
		FullName:      " Aarav Sharma ",
		AvatarURL:     strptr("https://cdn.example.com/avatars/user_123.png"),
		CreatedBy:     "service:auth-service",
		CreatedAt:     fixedTime(),
	})
	if err != nil {
		t.Fatalf("NewUser returned error: %v", err)
	}

	if user.UserID != "user_123" {
		t.Fatalf("expected trimmed user id, got %q", user.UserID)
	}
	if user.Phone == nil || *user.Phone != "+919999999999" {
		t.Fatalf("expected trimmed phone, got %#v", user.Phone)
	}
	if user.Status != UserStatusActive {
		t.Fatalf("expected active status, got %q", user.Status)
	}
}

func TestUserRejectsInvalidEmail(t *testing.T) {
	_, err := NewUser(NewUserParams{
		UserID:        "user_123",
		AuthAccountID: "auth_123",
		Email:         "not-an-email",
		FullName:      "Aarav Sharma",
		CreatedBy:     "service:auth-service",
		CreatedAt:     fixedTime(),
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestDeletedUserCannotBePatched(t *testing.T) {
	user, err := NewUser(NewUserParams{
		UserID:        "user_123",
		AuthAccountID: "auth_123",
		Email:         "buyer@example.com",
		FullName:      "Aarav Sharma",
		CreatedBy:     "service:auth-service",
		CreatedAt:     fixedTime(),
	})
	if err != nil {
		t.Fatalf("NewUser returned error: %v", err)
	}

	if err := user.TransitionStatus(UserStatusDeleted, "admin_123", fixedTime().Add(time.Minute)); err != nil {
		t.Fatalf("TransitionStatus returned error: %v", err)
	}

	err = user.ApplyProfilePatch(UserProfilePatch{
		FullName:  strptr("New Name"),
		UpdatedBy: "user_123",
		UpdatedAt: fixedTime().Add(2 * time.Minute),
	})
	if !errors.Is(err, ErrDeletedResource) {
		t.Fatalf("expected deleted resource error, got %v", err)
	}
}
