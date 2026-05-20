package token

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

type AccessClaims struct {
	Subject   string   `json:"sub"`
	SessionID string   `json:"sid"`
	Roles     []string `json:"roles"`
	SellerID  string   `json:"seller_id,omitempty"`
	TenantID  string   `json:"tenant_id,omitempty"`
	TokenType string   `json:"token_type"`
	ID        string   `json:"jti"`
	Issuer    string   `json:"iss"`
	Audience  []string `json:"aud"`
	IssuedAt  int64    `json:"iat"`
	NotBefore int64    `json:"nbf"`
	ExpiresAt int64    `json:"exp"`
}

func (c AccessClaims) Validate(now time.Time, issuer string, audience string, leeway time.Duration) error {
	if c.Subject == "" || c.SessionID == "" || c.ID == "" {
		return domain.ErrInvalidAccessToken
	}
	if c.TokenType != domain.TokenTypeAccess {
		return domain.ErrInvalidAccessToken
	}
	if len(c.Roles) == 0 {
		return domain.ErrInvalidAccessToken
	}
	if c.Issuer != issuer {
		return domain.ErrInvalidAccessToken
	}
	if !hasAudience(c.Audience, audience) {
		return domain.ErrInvalidAccessToken
	}

	now = now.UTC()
	if c.ExpiresAt <= 0 || now.After(time.Unix(c.ExpiresAt, 0).Add(leeway)) {
		return domain.ErrInvalidAccessToken
	}
	if c.NotBefore > 0 && now.Add(leeway).Before(time.Unix(c.NotBefore, 0)) {
		return domain.ErrInvalidAccessToken
	}
	if c.IssuedAt <= 0 {
		return domain.ErrInvalidAccessToken
	}
	return nil
}

func (c AccessClaims) DomainClaims() domain.TokenClaims {
	return domain.TokenClaims{
		UserID:    c.Subject,
		SessionID: c.SessionID,
		Roles:     append([]string(nil), c.Roles...),
		SellerID:  c.SellerID,
		TenantID:  c.TenantID,
		TokenID:   c.ID,
		Issuer:    c.Issuer,
		Audience:  append([]string(nil), c.Audience...),
		ExpiresAt: time.Unix(c.ExpiresAt, 0).UTC(),
		IssuedAt:  time.Unix(c.IssuedAt, 0).UTC(),
		NotBefore: time.Unix(c.NotBefore, 0).UTC(),
	}
}

func hasAudience(audiences []string, expected string) bool {
	for _, audience := range audiences {
		if audience == expected {
			return true
		}
	}
	return false
}

func decodeClaims(payload []byte) (AccessClaims, error) {
	var raw struct {
		Subject   string          `json:"sub"`
		SessionID string          `json:"sid"`
		Roles     []string        `json:"roles"`
		SellerID  string          `json:"seller_id,omitempty"`
		TenantID  string          `json:"tenant_id,omitempty"`
		TokenType string          `json:"token_type"`
		ID        string          `json:"jti"`
		Issuer    string          `json:"iss"`
		Audience  json.RawMessage `json:"aud"`
		IssuedAt  int64           `json:"iat"`
		NotBefore int64           `json:"nbf"`
		ExpiresAt int64           `json:"exp"`
	}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return AccessClaims{}, fmt.Errorf("decode jwt claims: %w", err)
	}
	audience, err := decodeAudience(raw.Audience)
	if err != nil {
		return AccessClaims{}, err
	}

	return AccessClaims{
		Subject:   raw.Subject,
		SessionID: raw.SessionID,
		Roles:     raw.Roles,
		SellerID:  raw.SellerID,
		TenantID:  raw.TenantID,
		TokenType: raw.TokenType,
		ID:        raw.ID,
		Issuer:    raw.Issuer,
		Audience:  audience,
		IssuedAt:  raw.IssuedAt,
		NotBefore: raw.NotBefore,
		ExpiresAt: raw.ExpiresAt,
	}, nil
}

func decodeAudience(raw json.RawMessage) ([]string, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, nil
	}

	var audiences []string
	if err := json.Unmarshal(raw, &audiences); err == nil {
		return audiences, nil
	}

	var audience string
	if err := json.Unmarshal(raw, &audience); err != nil {
		return nil, fmt.Errorf("decode jwt audience: %w", err)
	}
	if audience == "" {
		return nil, nil
	}
	return []string{audience}, nil
}

var ErrMalformedJWT = errors.New("malformed jwt")
