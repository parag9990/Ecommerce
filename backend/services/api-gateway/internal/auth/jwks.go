package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const maxJWKSResponseBytes = 1 << 20

var ErrKeyNotFound = errors.New("jwt signing key not found")

type JWKSConfig struct {
	URL      string
	CacheTTL time.Duration
}

type RemoteJWKSKeyProvider struct {
	url      string
	cacheTTL time.Duration
	client   *http.Client
	logger   *slog.Logger

	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	expiresAt time.Time
}

func NewRemoteJWKSKeyProvider(cfg JWKSConfig, client *http.Client, logger *slog.Logger) (*RemoteJWKSKeyProvider, error) {
	cfg.URL = strings.TrimSpace(cfg.URL)
	if cfg.URL == "" {
		return nil, fmt.Errorf("jwks url is required")
	}
	if cfg.CacheTTL <= 0 {
		return nil, fmt.Errorf("jwks cache ttl must be positive")
	}
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second}
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &RemoteJWKSKeyProvider{
		url:      cfg.URL,
		cacheTTL: cfg.CacheTTL,
		client:   client,
		logger:   logger,
		keys:     make(map[string]*rsa.PublicKey),
	}, nil
}

func (p *RemoteJWKSKeyProvider) Keyfunc(ctx context.Context) jwt.Keyfunc {
	return func(token *jwt.Token) (any, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		kid, _ := token.Header["kid"].(string)
		kid = strings.TrimSpace(kid)
		if kid == "" {
			return nil, fmt.Errorf("%w: missing kid", ErrKeyNotFound)
		}
		if key, ok := p.cachedKey(kid, time.Now()); ok {
			return key, nil
		}
		if err := p.refresh(ctx); err != nil {
			p.logger.WarnContext(ctx, "jwks_refresh_failed", "error", err)
			return nil, ErrKeyNotFound
		}
		if key, ok := p.cachedKey(kid, time.Now()); ok {
			return key, nil
		}
		return nil, fmt.Errorf("%w: unknown kid", ErrKeyNotFound)
	}
}

func (p *RemoteJWKSKeyProvider) cachedKey(kid string, now time.Time) (*rsa.PublicKey, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if now.After(p.expiresAt) {
		return nil, false
	}
	key, ok := p.keys[kid]
	return key, ok
}

func (p *RemoteJWKSKeyProvider) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.url, nil)
	if err != nil {
		return fmt.Errorf("build jwks request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch jwks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("fetch jwks: unexpected status %d", resp.StatusCode)
	}
	keys, err := decodeJWKS(io.LimitReader(resp.Body, maxJWKSResponseBytes))
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return errors.New("jwks response did not contain usable rsa signing keys")
	}

	p.mu.Lock()
	p.keys = keys
	p.expiresAt = time.Now().Add(p.cacheTTL)
	p.mu.Unlock()

	p.logger.InfoContext(ctx, "jwks_cache_refreshed", "key_count", len(keys), "ttl", p.cacheTTL.String())
	return nil
}

func decodeJWKS(reader io.Reader) (map[string]*rsa.PublicKey, error) {
	var set jwksDocument
	if err := json.NewDecoder(reader).Decode(&set); err != nil {
		return nil, fmt.Errorf("decode jwks: %w", err)
	}
	keys := make(map[string]*rsa.PublicKey, len(set.Keys))
	for _, jwk := range set.Keys {
		if !jwk.usableSigningRSAKey() {
			continue
		}
		key, err := jwk.rsaPublicKey()
		if err != nil {
			return nil, fmt.Errorf("decode jwk %q: %w", jwk.KID, err)
		}
		keys[jwk.KID] = key
	}
	return keys, nil
}

type jwksDocument struct {
	Keys []jwkDocument `json:"keys"`
}

type jwkDocument struct {
	KTY string `json:"kty"`
	Use string `json:"use,omitempty"`
	KID string `json:"kid"`
	Alg string `json:"alg,omitempty"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func (j jwkDocument) usableSigningRSAKey() bool {
	if strings.TrimSpace(j.KID) == "" || j.KTY != "RSA" || j.N == "" || j.E == "" {
		return false
	}
	if j.Use != "" && j.Use != "sig" {
		return false
	}
	return true
}

func (j jwkDocument) rsaPublicKey() (*rsa.PublicKey, error) {
	modulusBytes, err := base64.RawURLEncoding.DecodeString(j.N)
	if err != nil {
		return nil, fmt.Errorf("decode modulus: %w", err)
	}
	exponentBytes, err := base64.RawURLEncoding.DecodeString(j.E)
	if err != nil {
		return nil, fmt.Errorf("decode exponent: %w", err)
	}
	exponentValue := new(big.Int).SetBytes(exponentBytes)
	if !exponentValue.IsInt64() {
		return nil, fmt.Errorf("invalid exponent")
	}
	exponent := exponentValue.Int64()
	if exponent <= 1 || exponent > int64(^uint(0)>>1) {
		return nil, fmt.Errorf("invalid exponent")
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(modulusBytes),
		E: int(exponent),
	}, nil
}
