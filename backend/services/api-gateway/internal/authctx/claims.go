package authctx

import "time"

type Claims struct {
	UserID    string
	SessionID string
	SellerID  string
	Roles     []string
	TokenType string
	Issuer    string
	Audience  []string
	ExpiresAt time.Time
	NotBefore time.Time
	IssuedAt  time.Time
}

func (c Claims) HasRole(role string) bool {
	for _, existing := range c.Roles {
		if existing == role {
			return true
		}
	}
	return false
}
