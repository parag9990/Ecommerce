package domain

import (
	"fmt"
	"time"
)

type UserStatus string

const (
	UserStatusActive  UserStatus = "active"
	UserStatusBlocked UserStatus = "blocked"
	UserStatusDeleted UserStatus = "deleted"
)

type User struct {
	UserID        string
	AuthAccountID string
	Email         string
	Phone         *string
	FullName      string
	AvatarURL     *string
	Status        UserStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type NewUserParams struct {
	UserID        string
	AuthAccountID string
	Email         string
	Phone         *string
	FullName      string
	AvatarURL     *string
	CreatedAt     time.Time
}

type UserProfilePatch struct {
	Phone     *string
	FullName  *string
	AvatarURL *string
	UpdatedAt time.Time
}

func NewUser(params NewUserParams) (User, error) {
	createdAt := params.CreatedAt.UTC()
	user := User{
		UserID:        trim(params.UserID),
		AuthAccountID: trim(params.AuthAccountID),
		Email:         trim(params.Email),
		Phone:         cleanOptional(params.Phone),
		FullName:      trim(params.FullName),
		AvatarURL:     cleanOptional(params.AvatarURL),
		Status:        UserStatusActive,
		CreatedAt:     createdAt,
		UpdatedAt:     createdAt,
	}

	if err := user.Validate(); err != nil {
		return User{}, err
	}

	return user, nil
}

func (s UserStatus) Valid() bool {
	switch s {
	case UserStatusActive, UserStatusBlocked, UserStatusDeleted:
		return true
	default:
		return false
	}
}

func (u User) Validate() error {
	var v validationCollector

	validateID(&v, "user_id", u.UserID)
	validateID(&v, "auth_account_id", u.AuthAccountID)
	validateEmail(&v, "email", u.Email)
	validateOptionalPhone(&v, "phone", u.Phone)
	validateRequiredString(&v, "full_name", u.FullName, maxNameLength)
	validateOptionalHTTPURL(&v, "avatar_url", u.AvatarURL, maxAvatarURLLength)
	if !u.Status.Valid() {
		v.add("status", "is not supported")
	}
	validateTimestamp(&v, "created_at", u.CreatedAt)
	validateTimestamp(&v, "updated_at", u.UpdatedAt)
	if !u.CreatedAt.IsZero() && !u.UpdatedAt.IsZero() && u.UpdatedAt.Before(u.CreatedAt) {
		v.add("updated_at", "cannot be before created_at")
	}

	return v.err()
}

func (u User) IsActive() bool {
	return u.Status == UserStatusActive
}

func (u User) IsDeleted() bool {
	return u.Status == UserStatusDeleted
}

func (u *User) ApplyProfilePatch(patch UserProfilePatch) error {
	if u.IsDeleted() {
		return fmt.Errorf("%w: user %q", ErrDeletedResource, u.UserID)
	}

	if patch.UpdatedAt.IsZero() {
		return ValidationError{Fields: []FieldError{{Field: "updated_at", Message: "is required"}}}
	}

	next := *u
	if patch.Phone != nil {
		next.Phone = cleanOptional(patch.Phone)
	}
	if patch.FullName != nil {
		next.FullName = trim(*patch.FullName)
	}
	if patch.AvatarURL != nil {
		next.AvatarURL = cleanOptional(patch.AvatarURL)
	}
	next.UpdatedAt = patch.UpdatedAt.UTC()

	if err := next.Validate(); err != nil {
		return err
	}

	*u = next
	return nil
}

func (u *User) TransitionStatus(to UserStatus, at time.Time) error {
	if at.IsZero() {
		return ValidationError{Fields: []FieldError{{Field: "updated_at", Message: "is required"}}}
	}
	if !to.Valid() {
		return ValidationError{Fields: []FieldError{{Field: "status", Message: "is not supported"}}}
	}
	if u.Status == to {
		u.UpdatedAt = at.UTC()
		return nil
	}
	if u.Status == UserStatusDeleted {
		return invalidTransition("user", string(u.Status), string(to))
	}

	switch u.Status {
	case UserStatusActive:
		if to != UserStatusBlocked && to != UserStatusDeleted {
			return invalidTransition("user", string(u.Status), string(to))
		}
	case UserStatusBlocked:
		if to != UserStatusActive && to != UserStatusDeleted {
			return invalidTransition("user", string(u.Status), string(to))
		}
	default:
		return invalidTransition("user", string(u.Status), string(to))
	}

	u.Status = to
	u.UpdatedAt = at.UTC()
	return u.Validate()
}
