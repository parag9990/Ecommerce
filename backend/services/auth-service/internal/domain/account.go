package domain

import "time"

type AccountStatus string

const (
	AccountStatusActive  AccountStatus = "active"
	AccountStatusBlocked AccountStatus = "blocked"
	AccountStatusDeleted AccountStatus = "deleted"
)

func (s AccountStatus) CanAuthenticate() bool {
	return s == AccountStatusActive
}

type AuthAccount struct {
	AccountID     string
	Email         *string
	Phone         *string
	EmailVerified bool
	PhoneVerified bool
	Status        AccountStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type AccountCredential struct {
	AccountID     string
	AccountStatus AccountStatus
	EmailVerified bool
	PhoneVerified bool
	Credential    Credential
}
