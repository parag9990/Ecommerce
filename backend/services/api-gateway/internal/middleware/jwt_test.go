package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestJWTVerifierAcceptsValidAccessToken(t *testing.T) {
	now := time.Date(2026, 6, 3, 10, 0, 0, 0, time.UTC)
	verifier := NewJWTVerifier("secret", "ecommerce-auth", "ecommerce-api", 30*time.Second, fixedClock{now: now})
	token := signJWT(t, "secret", map[string]any{
		"sub":        "user_123",
		"sid":        "sess_123",
		"seller_id":  "seller_123",
		"roles":      []string{"buyer", "seller"},
		"token_type": "access",
		"iss":        "ecommerce-auth",
		"aud":        "ecommerce-api",
		"iat":        now.Unix(),
		"exp":        now.Add(time.Minute).Unix(),
	})

	claims, err := verifier.Verify(context.Background(), token)
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	if claims.UserID != "user_123" || claims.SellerID != "seller_123" {
		t.Fatalf("claims = %#v", claims)
	}
}

func TestJWTVerifierRejectsExpiredToken(t *testing.T) {
	now := time.Date(2026, 6, 3, 10, 0, 0, 0, time.UTC)
	verifier := NewJWTVerifier("secret", "ecommerce-auth", "ecommerce-api", 0, fixedClock{now: now})
	token := signJWT(t, "secret", map[string]any{
		"sub":        "user_123",
		"roles":      []string{"buyer"},
		"token_type": "access",
		"iss":        "ecommerce-auth",
		"aud":        "ecommerce-api",
		"exp":        now.Add(-time.Second).Unix(),
	})

	if _, err := verifier.Verify(context.Background(), token); err == nil {
		t.Fatal("expected expired token error")
	}
}

func signJWT(t *testing.T, secret string, payload map[string]any) string {
	t.Helper()

	header := encodeJSONPart(t, map[string]any{"alg": "HS256", "typ": "JWT"})
	body := encodeJSONPart(t, payload)
	signingInput := header + "." + body

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signingInput))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return strings.Join([]string{header, body, signature}, ".")
}

func encodeJSONPart(t *testing.T, value any) string {
	t.Helper()

	bytes, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}
