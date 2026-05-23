package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestVerifierAcceptsValidAccessToken(t *testing.T) {
	privateKey := newTestRSAKey(t)
	verifier := newTestVerifier(t, &privateKey.PublicKey)
	raw := signTestToken(t, privateKey, AccessClaims{
		SessionID: "sess_123",
		Roles:     []string{"buyer"},
		SellerID:  "seller_123",
		TokenType: AccessTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user_123",
			Issuer:    "ecommerce-auth",
			Audience:  []string{"ecommerce-api"},
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Minute)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	}, jwt.SigningMethodRS256)

	claims, err := verifier.Verify(context.Background(), raw)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	if claims.UserID() != "user_123" || claims.SessionID != "sess_123" || !claims.HasRole("buyer") || claims.SellerID != "seller_123" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestVerifierRejectsInvalidTokens(t *testing.T) {
	privateKey := newTestRSAKey(t)
	verifier := newTestVerifier(t, &privateKey.PublicKey)

	tests := []struct {
		name   string
		claims AccessClaims
		method jwt.SigningMethod
		want   error
	}{
		{
			name: "expired",
			claims: AccessClaims{
				SessionID: "sess_123",
				Roles:     []string{"buyer"},
				TokenType: AccessTokenType,
				RegisteredClaims: jwt.RegisteredClaims{
					Subject:   "user_123",
					Issuer:    "ecommerce-auth",
					Audience:  []string{"ecommerce-api"},
					IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Minute)),
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
				},
			},
			method: jwt.SigningMethodRS256,
			want:   ErrInvalidToken,
		},
		{
			name: "wrong issuer",
			claims: AccessClaims{
				SessionID: "sess_123",
				Roles:     []string{"buyer"},
				TokenType: AccessTokenType,
				RegisteredClaims: jwt.RegisteredClaims{
					Subject:   "user_123",
					Issuer:    "other-auth",
					Audience:  []string{"ecommerce-api"},
					IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Minute)),
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
				},
			},
			method: jwt.SigningMethodRS256,
			want:   ErrInvalidToken,
		},
		{
			name: "wrong audience",
			claims: AccessClaims{
				SessionID: "sess_123",
				Roles:     []string{"buyer"},
				TokenType: AccessTokenType,
				RegisteredClaims: jwt.RegisteredClaims{
					Subject:   "user_123",
					Issuer:    "ecommerce-auth",
					Audience:  []string{"other-api"},
					IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Minute)),
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
				},
			},
			method: jwt.SigningMethodRS256,
			want:   ErrInvalidToken,
		},
		{
			name: "refresh token",
			claims: AccessClaims{
				SessionID: "sess_123",
				Roles:     []string{"buyer"},
				TokenType: "refresh",
				RegisteredClaims: jwt.RegisteredClaims{
					Subject:   "user_123",
					Issuer:    "ecommerce-auth",
					Audience:  []string{"ecommerce-api"},
					IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Minute)),
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
				},
			},
			method: jwt.SigningMethodRS256,
			want:   ErrInvalidToken,
		},
		{
			name: "missing roles",
			claims: AccessClaims{
				SessionID: "sess_123",
				TokenType: AccessTokenType,
				RegisteredClaims: jwt.RegisteredClaims{
					Subject:   "user_123",
					Issuer:    "ecommerce-auth",
					Audience:  []string{"ecommerce-api"},
					IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Minute)),
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
				},
			},
			method: jwt.SigningMethodRS256,
			want:   ErrMissingClaim,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := signTestToken(t, privateKey, tt.claims, tt.method)
			_, err := verifier.Verify(context.Background(), raw)
			if !errors.Is(err, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, err)
			}
		})
	}
}

func TestVerifierRejectsWrongAlgorithm(t *testing.T) {
	privateKey := newTestRSAKey(t)
	verifier := newTestVerifier(t, &privateKey.PublicKey)
	raw := signTestToken(t, privateKey, AccessClaims{
		SessionID: "sess_123",
		Roles:     []string{"buyer"},
		TokenType: AccessTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user_123",
			Issuer:    "ecommerce-auth",
			Audience:  []string{"ecommerce-api"},
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Minute)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	}, jwt.SigningMethodPS256)

	_, err := verifier.Verify(context.Background(), raw)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected invalid token, got %v", err)
	}
}

func newTestRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	return key
}

func newTestVerifier(t *testing.T, publicKey *rsa.PublicKey) *Verifier {
	t.Helper()
	verifier, err := NewVerifier(
		"ecommerce-auth",
		"ecommerce-api",
		[]string{"RS256"},
		0,
		staticKeyProvider{key: publicKey},
	)
	if err != nil {
		t.Fatalf("new verifier: %v", err)
	}
	return verifier
}

func signTestToken(t *testing.T, privateKey *rsa.PrivateKey, claims AccessClaims, method jwt.SigningMethod) string {
	t.Helper()
	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = "test-key"
	raw, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return raw
}

type staticKeyProvider struct {
	key *rsa.PublicKey
}

func (p staticKeyProvider) Keyfunc(context.Context) jwt.Keyfunc {
	return func(*jwt.Token) (any, error) {
		return p.key, nil
	}
}
