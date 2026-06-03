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
	AuditFields
	StatusAuditFields
	SoftDeleteFields
}

type NewUserParams struct {
	UserID        string
	AuthAccountID string
	Email         string
	Phone         *string
	FullName      string
	AvatarURL     *string
	CreatedBy     string
	CreatedAt     time.Time
}

type UserProfilePatch struct {
	Phone     *string
	FullName  *string
	AvatarURL *string
	UpdatedBy string
	UpdatedAt time.Time
}

func NewUser(params NewUserParams) (User, error) {
	createdAt := params.CreatedAt.UTC()
	user := User{
		UserID:        trim(params.UserID),
		AuthAccountID: trim(params.AuthAccountID),
		Email:         sharedNormalizeEmail(params.Email),
		Phone:         cleanOptionalPhone(params.Phone),
		FullName:      cleanText(params.FullName),
		AvatarURL:     cleanOptional(params.AvatarURL),
		Status:        UserStatusActive,
		AuditFields: AuditFields{
			CreatedBy: trim(params.CreatedBy),
			UpdatedBy: trim(params.CreatedBy),
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		},
		StatusAuditFields: StatusAuditFields{
			StatusChangedBy: cleanOptional(&params.CreatedBy),
			StatusChangedAt: &createdAt,
		},
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
	validateRequiredSafeText(&v, "full_name", u.FullName, 2, maxNameLength)
	validateOptionalHTTPURL(&v, "avatar_url", u.AvatarURL, maxAvatarURLLength)
	if !u.Status.Valid() {
		v.add("status", "is not supported")
	}
	validateAuditFields(&v, u.AuditFields)
	validateStatusAuditFields(&v, u.StatusAuditFields, u.CreatedAt)
	validateSoftDeleteFields(&v, u.SoftDeleteFields, u.CreatedAt)
	if u.IsDeleted() && u.DeletedAt == nil {
		v.add("deleted_at", "is required for deleted user")
	}

	return v.err()
}

func (u User) IsActive() bool {
	return u.Status == UserStatusActive
}

func (u User) IsDeleted() bool {
	return u.Status == UserStatusDeleted || u.DeletedAt != nil
}

func (u *User) ApplyProfilePatch(patch UserProfilePatch) error {
	if u.IsDeleted() {
		return fmt.Errorf("%w: user %q", ErrDeletedResource, u.UserID)
	}

	if trim(patch.UpdatedBy) == "" {
		return ValidationError{Fields: []FieldError{{Field: "updated_by", Message: "is required"}}}
	}
	if patch.UpdatedAt.IsZero() {
		return ValidationError{Fields: []FieldError{{Field: "updated_at", Message: "is required"}}}
	}

	next := *u
	if patch.Phone != nil {
		next.Phone = cleanOptionalPhone(patch.Phone)
	}
	if patch.FullName != nil {
		next.FullName = cleanText(*patch.FullName)
	}
	if patch.AvatarURL != nil {
		next.AvatarURL = cleanOptional(patch.AvatarURL)
	}
	next.UpdatedBy = trim(patch.UpdatedBy)
	next.UpdatedAt = patch.UpdatedAt.UTC()

	if err := next.Validate(); err != nil {
		return err
	}

	*u = next
	return nil
}

func (u *User) TransitionStatus(to UserStatus, actorID string, at time.Time) error {
	actorID = trim(actorID)
	if actorID == "" {
		return ValidationError{Fields: []FieldError{{Field: "updated_by", Message: "is required"}}}
	}
	if at.IsZero() {
		return ValidationError{Fields: []FieldError{{Field: "updated_at", Message: "is required"}}}
	}
	if !to.Valid() {
		return ValidationError{Fields: []FieldError{{Field: "status", Message: "is not supported"}}}
	}
	if u.Status == to {
		u.UpdatedBy = actorID
		u.UpdatedAt = at.UTC()
		return u.Validate()
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

	changedAt := at.UTC()
	u.Status = to
	u.UpdatedBy = actorID
	u.UpdatedAt = changedAt
	u.StatusChangedBy = &actorID
	u.StatusChangedAt = &changedAt
	if to == UserStatusDeleted {
		u.DeletedBy = &actorID
		u.DeletedAt = &changedAt
	}
	return u.Validate()
}
