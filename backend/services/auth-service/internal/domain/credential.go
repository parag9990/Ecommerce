package domain

import "time"

type Credential struct {
	AccountID         string
	PasswordHash      string
	PasswordAlgo      string
	PasswordChangedAt *time.Time
	FailedAttempts    int
	LockedUntil       *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (c Credential) IsLocked(now time.Time) bool {
	return c.LockedUntil != nil && c.LockedUntil.After(now.UTC())
}
