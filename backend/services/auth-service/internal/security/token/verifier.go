package token

import (
	"crypto"
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

type Verifier struct {
	issuer      string
	audience    string
	keys        *KeySet
	clockSkew   time.Duration
	now         func() time.Time
	signingAlgo string
}

type VerifierConfig struct {
	Issuer           string
	Audience         string
	ClockSkew        time.Duration
	SigningAlgorithm string
}

func NewVerifier(cfg VerifierConfig, keys *KeySet) (*Verifier, error) {
	if cfg.Issuer == "" {
		return nil, errors.New("jwt issuer is required")
	}
	if cfg.Audience == "" {
		return nil, errors.New("jwt audience is required")
	}
	if cfg.ClockSkew < 0 {
		return nil, errors.New("jwt clock skew cannot be negative")
	}
	if cfg.SigningAlgorithm != "" && cfg.SigningAlgorithm != SigningAlgorithmRS256 {
		return nil, fmt.Errorf("unsupported jwt signing algorithm %q", cfg.SigningAlgorithm)
	}
	if keys == nil {
		return nil, errors.New("jwt key set is required")
	}

	return &Verifier{
		issuer:      cfg.Issuer,
		audience:    cfg.Audience,
		keys:        keys,
		clockSkew:   cfg.ClockSkew,
		now:         time.Now,
		signingAlgo: SigningAlgorithmRS256,
	}, nil
}

func (v *Verifier) WithClock(now func() time.Time) {
	if now != nil {
		v.now = now
	}
}

func (v *Verifier) VerifyAccessToken(tokenString string) (domain.TokenClaims, error) {
	header, claims, signingInput, signature, err := parseJWT(tokenString)
	if err != nil {
		return domain.TokenClaims{}, domain.ErrInvalidAccessToken
	}
	if header.Algorithm != v.signingAlgo || header.KeyID == "" {
		return domain.TokenClaims{}, domain.ErrInvalidAccessToken
	}

	publicKey, err := v.keys.PublicKeyByID(header.KeyID)
	if err != nil {
		return domain.TokenClaims{}, domain.ErrInvalidAccessToken
	}
	if err := verifyRS256(signingInput, signature, publicKey); err != nil {
		return domain.TokenClaims{}, domain.ErrInvalidAccessToken
	}
	if err := claims.Validate(v.now().UTC(), v.issuer, v.audience, v.clockSkew); err != nil {
		return domain.TokenClaims{}, err
	}
	return claims.DomainClaims(), nil
}

func parseJWT(tokenString string) (jwtHeader, AccessClaims, []byte, []byte, error) {
	tokenString = strings.TrimSpace(tokenString)
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return jwtHeader{}, AccessClaims{}, nil, nil, ErrMalformedJWT
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return jwtHeader{}, AccessClaims{}, nil, nil, ErrMalformedJWT
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return jwtHeader{}, AccessClaims{}, nil, nil, ErrMalformedJWT
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return jwtHeader{}, AccessClaims{}, nil, nil, ErrMalformedJWT
	}

	var header jwtHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return jwtHeader{}, AccessClaims{}, nil, nil, ErrMalformedJWT
	}
	claims, err := decodeClaims(payloadBytes)
	if err != nil {
		return jwtHeader{}, AccessClaims{}, nil, nil, ErrMalformedJWT
	}

	signingInput := []byte(parts[0] + "." + parts[1])
	return header, claims, signingInput, signature, nil
}

func verifyRS256(signingInput []byte, signature []byte, publicKey *rsa.PublicKey) error {
	digest := sha256.Sum256(signingInput)
	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, digest[:], signature); err != nil {
		return fmt.Errorf("verify jwt signature: %w", err)
	}
	return nil
}
