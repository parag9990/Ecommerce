package domain

import "time"

const TokenTypeAccess = "access"

type RefreshToken struct {
	TokenID       string
	AccountID     string
	SessionID     string
	TokenHash     string
	ParentTokenID string
	RevokedAt     *time.Time
	ExpiresAt     time.Time
	CreatedAt     time.Time
}

func (t RefreshToken) IsActive(now time.Time) bool {
	return t.RevokedAt == nil && now.UTC().Before(t.ExpiresAt.UTC())
}

type TokenSubject struct {
	AccountID string
	UserID    string
	SellerID  string
	TenantID  string
	Status    AccountStatus
}

func (s TokenSubject) CanIssueToken() bool {
	return s.AccountID != "" && s.UserID != "" && s.Status.CanAuthenticate()
}

type TokenClaims struct {
	UserID    string
	SessionID string
	Roles     []string
	SellerID  string
	TenantID  string
	TokenID   string
	Issuer    string
	Audience  []string
	ExpiresAt time.Time
	IssuedAt  time.Time
	NotBefore time.Time
}
