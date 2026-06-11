package token

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

func TestIssueAndVerifyAccessToken(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	keys := NewKeySet("auth-key-test", privateKey, &privateKey.PublicKey)
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	issuer, err := NewIssuer(IssuerConfig{
		Issuer:         "ecommerce-auth",
		Audience:       "ecommerce-api",
		AccessTokenTTL: 15 * time.Minute,
	}, keys)
	if err != nil {
		t.Fatalf("NewIssuer() error = %v", err)
	}
	issuer.WithClock(func() time.Time { return now })

	verifier, err := NewVerifier(VerifierConfig{
		Issuer:    "ecommerce-auth",
		Audience:  "ecommerce-api",
		ClockSkew: 30 * time.Second,
	}, keys)
	if err != nil {
		t.Fatalf("NewVerifier() error = %v", err)
	}
	verifier.WithClock(func() time.Time { return now.Add(time.Minute) })

	result, err := issuer.IssueAccessToken(AccessTokenInput{
		UserID:    "user_123",
		SessionID: "sess_123",
		Roles:     []string{"buyer", "buyer"},
		SellerID:  "seller_456",
		TenantID:  "tenant_default",
	})
	if err != nil {
		t.Fatalf("IssueAccessToken() error = %v", err)
	}
	if result.ExpiresIn != int64((15 * time.Minute).Seconds()) {
		t.Fatalf("ExpiresIn = %d", result.ExpiresIn)
	}

	header, _, _, _, err := parseJWT(result.Token)
	if err != nil {
		t.Fatalf("parseJWT() error = %v", err)
	}
	if header.Algorithm != SigningAlgorithmRS256 || header.KeyID != "auth-key-test" {
		t.Fatalf("header = %+v", header)
	}

	claims, err := verifier.VerifyAccessToken(result.Token)
	if err != nil {
		t.Fatalf("VerifyAccessToken() error = %v", err)
	}
	if claims.UserID != "user_123" || claims.SessionID != "sess_123" || claims.SellerID != "seller_456" {
		t.Fatalf("claims = %+v", claims)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != "buyer" {
		t.Fatalf("roles = %+v", claims.Roles)
	}
}

func TestVerifyAccessTokenRejectsWrongAudience(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	keys := NewKeySet("auth-key-test", privateKey, &privateKey.PublicKey)
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	issuer, err := NewIssuer(IssuerConfig{
		Issuer:         "ecommerce-auth",
		Audience:       "ecommerce-api",
		AccessTokenTTL: 15 * time.Minute,
	}, keys)
	if err != nil {
		t.Fatalf("NewIssuer() error = %v", err)
	}
	issuer.WithClock(func() time.Time { return now })

	verifier, err := NewVerifier(VerifierConfig{
		Issuer:   "ecommerce-auth",
		Audience: "another-api",
	}, keys)
	if err != nil {
		t.Fatalf("NewVerifier() error = %v", err)
	}
	verifier.WithClock(func() time.Time { return now })

	result, err := issuer.IssueAccessToken(AccessTokenInput{
		UserID:    "user_123",
		SessionID: "sess_123",
		Roles:     []string{"buyer"},
	})
	if err != nil {
		t.Fatalf("IssueAccessToken() error = %v", err)
	}

	_, err = verifier.VerifyAccessToken(result.Token)
	if !errors.Is(err, domain.ErrInvalidAccessToken) {
		t.Fatalf("VerifyAccessToken() error = %v, want invalid access token", err)
	}
}

func TestVerifyAccessTokenSecurityClaimsRejects(t *testing.T) {
	privateKey := generateTestRSAKey(t)
	keys := NewKeySet("auth-key-test", privateKey, &privateKey.PublicKey)
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	verifier, err := NewVerifier(VerifierConfig{
		Issuer:    "ecommerce-auth",
		Audience:  "ecommerce-api",
		ClockSkew: 30 * time.Second,
	}, keys)
	if err != nil {
		t.Fatalf("NewVerifier() error = %v", err)
	}
	verifier.WithClock(func() time.Time { return now })

	tests := []struct {
		name   string
		mutate func(*AccessClaims)
	}{
		{
			name: "expired access token",
			mutate: func(claims *AccessClaims) {
				claims.ExpiresAt = now.Add(-time.Minute).Unix()
			},
		},
		{
			name: "missing exp",
			mutate: func(claims *AccessClaims) {
				claims.ExpiresAt = 0
			},
		},
		{
			name: "wrong issuer",
			mutate: func(claims *AccessClaims) {
				claims.Issuer = "staging-auth"
			},
		},
		{
			name: "wrong audience",
			mutate: func(claims *AccessClaims) {
				claims.Audience = []string{"checkout-api"}
			},
		},
		{
			name: "missing subject",
			mutate: func(claims *AccessClaims) {
				claims.Subject = ""
			},
		},
		{
			name: "missing session id",
			mutate: func(claims *AccessClaims) {
				claims.SessionID = ""
			},
		},
		{
			name: "missing roles",
			mutate: func(claims *AccessClaims) {
				claims.Roles = nil
			},
		},
		{
			name: "refresh token type used as access",
			mutate: func(claims *AccessClaims) {
				claims.TokenType = "refresh"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := validTestAccessClaims(now)
			tt.mutate(&claims)
			tokenString, err := signJWT("auth-key-test", claims, privateKey)
			if err != nil {
				t.Fatalf("signJWT() error = %v", err)
			}

			_, err = verifier.VerifyAccessToken(tokenString)
			if !errors.Is(err, domain.ErrInvalidAccessToken) {
				t.Fatalf("VerifyAccessToken() error = %v, want invalid access token", err)
			}
		})
	}
}

func TestVerifyAccessTokenRejectsForgedSignature(t *testing.T) {
	privateKey := generateTestRSAKey(t)
	keys := NewKeySet("auth-key-test", privateKey, &privateKey.PublicKey)
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	verifier, err := NewVerifier(VerifierConfig{
		Issuer:   "ecommerce-auth",
		Audience: "ecommerce-api",
	}, keys)
	if err != nil {
		t.Fatalf("NewVerifier() error = %v", err)
	}
	verifier.WithClock(func() time.Time { return now })

	tokenString, err := signJWT("auth-key-test", validTestAccessClaims(now), privateKey)
	if err != nil {
		t.Fatalf("signJWT() error = %v", err)
	}

	parts := strings.Split(tokenString, ".")
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatalf("DecodeString(signature) error = %v", err)
	}
	signature[len(signature)-1] ^= 0xff
	parts[2] = base64.RawURLEncoding.EncodeToString(signature)
	forged := strings.Join(parts, ".")
	if forged == tokenString {
		t.Fatal("forged token unexpectedly matches original token")
	}

	_, err = verifier.VerifyAccessToken(forged)
	if !errors.Is(err, domain.ErrInvalidAccessToken) {
		t.Fatalf("VerifyAccessToken() error = %v, want invalid access token", err)
	}
}

func TestVerifyAccessTokenRejectsUnknownKeyID(t *testing.T) {
	privateKey := generateTestRSAKey(t)
	keys := NewKeySet("auth-key-test", privateKey, &privateKey.PublicKey)
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	verifier, err := NewVerifier(VerifierConfig{
		Issuer:   "ecommerce-auth",
		Audience: "ecommerce-api",
	}, keys)
	if err != nil {
		t.Fatalf("NewVerifier() error = %v", err)
	}
	verifier.WithClock(func() time.Time { return now })

	tokenString, err := signJWT("unknown-key-id", validTestAccessClaims(now), privateKey)
	if err != nil {
		t.Fatalf("signJWT() error = %v", err)
	}

	_, err = verifier.VerifyAccessToken(tokenString)
	if !errors.Is(err, domain.ErrInvalidAccessToken) {
		t.Fatalf("VerifyAccessToken() error = %v, want invalid access token", err)
	}
}

func TestVerifyAccessTokenRejectsUnsignedAlgorithmHeader(t *testing.T) {
	privateKey := generateTestRSAKey(t)
	keys := NewKeySet("auth-key-test", privateKey, &privateKey.PublicKey)
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	verifier, err := NewVerifier(VerifierConfig{
		Issuer:   "ecommerce-auth",
		Audience: "ecommerce-api",
	}, keys)
	if err != nil {
		t.Fatalf("NewVerifier() error = %v", err)
	}
	verifier.WithClock(func() time.Time { return now })

	claimsJSON, err := json.Marshal(validTestAccessClaims(now))
	if err != nil {
		t.Fatalf("Marshal(claims) error = %v", err)
	}
	headerJSON, err := json.Marshal(jwtHeader{Algorithm: "none", Type: "JWT", KeyID: "auth-key-test"})
	if err != nil {
		t.Fatalf("Marshal(header) error = %v", err)
	}
	tokenString := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(claimsJSON) + "."

	_, err = verifier.VerifyAccessToken(tokenString)
	if !errors.Is(err, domain.ErrInvalidAccessToken) {
		t.Fatalf("VerifyAccessToken() error = %v, want invalid access token", err)
	}
}

func TestHashRefreshTokenUsesPepper(t *testing.T) {
	first, err := HashRefreshToken("refresh-token", "pepper-one")
	if err != nil {
		t.Fatalf("HashRefreshToken() error = %v", err)
	}
	second, err := HashRefreshToken("refresh-token", "pepper-one")
	if err != nil {
		t.Fatalf("HashRefreshToken() error = %v", err)
	}
	third, err := HashRefreshToken("refresh-token", "pepper-two")
	if err != nil {
		t.Fatalf("HashRefreshToken() error = %v", err)
	}

	if first != second {
		t.Fatal("same token and pepper must hash consistently")
	}
	if first == third {
		t.Fatal("different pepper must produce different refresh token hash")
	}
}

func generateTestRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	return privateKey
}

func validTestAccessClaims(now time.Time) AccessClaims {
	return AccessClaims{
		Subject:   "user_123",
		SessionID: "sess_123",
		Roles:     []string{"buyer"},
		TokenType: domain.TokenTypeAccess,
		ID:        "jti_123",
		Issuer:    "ecommerce-auth",
		Audience:  []string{"ecommerce-api"},
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		ExpiresAt: now.Add(15 * time.Minute).Unix(),
	}
}
