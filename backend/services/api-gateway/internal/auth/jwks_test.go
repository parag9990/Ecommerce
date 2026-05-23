package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestRemoteJWKSKeyProviderFetchesKeyByKID(t *testing.T) {
	privateKey := newTestRSAKey(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if err := json.NewEncoder(w).Encode(testJWKS(privateKey.PublicKey, "test-key")); err != nil {
			t.Fatalf("encode jwks: %v", err)
		}
	}))
	defer server.Close()

	provider, err := NewRemoteJWKSKeyProvider(JWKSConfig{URL: server.URL, CacheTTL: time.Minute}, server.Client(), nil)
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}

	key, err := provider.Keyfunc(context.Background())(&jwt.Token{Header: map[string]any{"kid": "test-key"}})
	if err != nil {
		t.Fatalf("keyfunc: %v", err)
	}
	publicKey, ok := key.(*rsa.PublicKey)
	if !ok {
		t.Fatalf("expected rsa public key, got %T", key)
	}
	if publicKey.N.Cmp(privateKey.PublicKey.N) != 0 || publicKey.E != privateKey.PublicKey.E {
		t.Fatal("expected fetched public key to match jwks")
	}
}

func testJWKS(publicKey rsa.PublicKey, kid string) map[string]any {
	exponent := big.NewInt(int64(publicKey.E)).Bytes()
	return map[string]any{
		"keys": []map[string]string{
			{
				"kty": "RSA",
				"use": "sig",
				"kid": kid,
				"alg": "RS256",
				"n":   base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
				"e":   base64.RawURLEncoding.EncodeToString(exponent),
			},
		},
	}
}

func TestDecodeJWKSRejectsInvalidExponent(t *testing.T) {
	doc := `{"keys":[{"kty":"RSA","use":"sig","kid":"bad","n":"AQAB","e":"AA"}]}`
	keys, err := decodeJWKS(strings.NewReader(doc))
	if err == nil {
		t.Fatalf("expected invalid exponent error, got keys %+v", keys)
	}
}
