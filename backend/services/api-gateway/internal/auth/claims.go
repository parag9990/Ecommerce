package auth

import (
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

const AccessTokenType = "access"

type AccessClaims struct {
	SessionID string   `json:"sid"`
	Roles     []string `json:"roles"`
	SellerID  string   `json:"seller_id,omitempty"`
	TokenType string   `json:"token_type"`
	MFA       bool     `json:"mfa_verified,omitempty"`
	AMR       []string `json:"amr,omitempty"`

	jwt.RegisteredClaims
}

func (c AccessClaims) MFAVerified() bool {
	if c.MFA {
		return true
	}
	for _, method := range c.AMR {
		if strings.EqualFold(strings.TrimSpace(method), "mfa") {
			return true
		}
	}
	return false
}

func (c AccessClaims) UserID() string {
	return c.Subject
}

func (c AccessClaims) HasRole(role string) bool {
	role = strings.TrimSpace(role)
	if role == "" {
		return false
	}
	for _, current := range c.Roles {
		if strings.TrimSpace(current) == role {
			return true
		}
	}
	return false
}

func (c AccessClaims) HasAnyRole(allowed []string) bool {
	for _, role := range allowed {
		if c.HasRole(role) {
			return true
		}
	}
	return false
}

func cloneRoles(roles []string) []string {
	out := make([]string, len(roles))
	copy(out, roles)
	return out
}
