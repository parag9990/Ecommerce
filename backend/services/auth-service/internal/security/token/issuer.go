package token

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

type AccessTokenInput struct {
	UserID    string
	SessionID string
	Roles     []string
	SellerID  string
	TenantID  string
}

type AccessTokenResult struct {
	Token     string
	ExpiresIn int64
	Claims    AccessClaims
}

type Issuer struct {
	issuer      string
	audience    string
	ttl         time.Duration
	keyID       string
	privateKey  *rsa.PrivateKey
	now         func() time.Time
	signingAlgo string
}

type IssuerConfig struct {
	Issuer           string
	Audience         string
	AccessTokenTTL   time.Duration
	SigningAlgorithm string
}

func NewIssuer(cfg IssuerConfig, keys *KeySet) (*Issuer, error) {
	if cfg.Issuer == "" {
		return nil, errors.New("jwt issuer is required")
	}
	if cfg.Audience == "" {
		return nil, errors.New("jwt audience is required")
	}
	if cfg.AccessTokenTTL <= 0 {
		return nil, errors.New("jwt access token ttl must be greater than zero")
	}
	if cfg.SigningAlgorithm != "" && cfg.SigningAlgorithm != SigningAlgorithmRS256 {
		return nil, fmt.Errorf("unsupported jwt signing algorithm %q", cfg.SigningAlgorithm)
	}
	if keys == nil || keys.PrivateKey() == nil {
		return nil, errors.New("rsa signing key is required")
	}
	if keys.KeyID() == "" {
		return nil, errors.New("jwt key id is required")
	}

	return &Issuer{
		issuer:      cfg.Issuer,
		audience:    cfg.Audience,
		ttl:         cfg.AccessTokenTTL,
		keyID:       keys.KeyID(),
		privateKey:  keys.PrivateKey(),
		now:         time.Now,
		signingAlgo: SigningAlgorithmRS256,
	}, nil
}

func (i *Issuer) WithClock(now func() time.Time) {
	if now != nil {
		i.now = now
	}
}

func (i *Issuer) IssueAccessToken(input AccessTokenInput) (AccessTokenResult, error) {
	input.UserID = strings.TrimSpace(input.UserID)
	input.SessionID = strings.TrimSpace(input.SessionID)
	if input.UserID == "" || input.SessionID == "" {
		return AccessTokenResult{}, domain.ErrTokenSubjectMissing
	}
	if len(input.Roles) == 0 {
		return AccessTokenResult{}, domain.ErrRoleRequired
	}

	jti, err := NewJWTID()
	if err != nil {
		return AccessTokenResult{}, err
	}

	now := i.now().UTC()
	expiresAt := now.Add(i.ttl)
	claims := AccessClaims{
		Subject:   input.UserID,
		SessionID: input.SessionID,
		Roles:     uniqueNonBlank(input.Roles),
		SellerID:  strings.TrimSpace(input.SellerID),
		TenantID:  strings.TrimSpace(input.TenantID),
		TokenType: domain.TokenTypeAccess,
		ID:        jti,
		Issuer:    i.issuer,
		Audience:  []string{i.audience},
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		ExpiresAt: expiresAt.Unix(),
	}
	if len(claims.Roles) == 0 {
		return AccessTokenResult{}, domain.ErrRoleRequired
	}

	signed, err := signJWT(i.keyID, claims, i.privateKey)
	if err != nil {
		return AccessTokenResult{}, err
	}

	return AccessTokenResult{
		Token:     signed,
		ExpiresIn: int64(i.ttl.Seconds()),
		Claims:    claims,
	}, nil
}

type jwtHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
	KeyID     string `json:"kid"`
}

func signJWT(keyID string, claims AccessClaims, privateKey *rsa.PrivateKey) (string, error) {
	header := jwtHeader{Algorithm: SigningAlgorithmRS256, Type: "JWT", KeyID: keyID}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("encode jwt header: %w", err)
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("encode jwt claims: %w", err)
	}

	unsigned := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(claimsJSON)
	digest := sha256.Sum256([]byte(unsigned))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}

	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func uniqueNonBlank(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
