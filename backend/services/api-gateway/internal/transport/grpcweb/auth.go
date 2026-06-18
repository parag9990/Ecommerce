package grpcweb

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	gatewayauth "ecommerce/api-gateway/internal/auth"
	"ecommerce/api-gateway/internal/config"
)

type TokenVerifier interface {
	Verify(ctx context.Context, raw string) (gatewayauth.AccessClaims, error)
}

func NewTokenVerifier(cfg config.Config, logger *slog.Logger) (TokenVerifier, error) {
	if strings.TrimSpace(cfg.JWTJWKSURL) == "" {
		return nil, fmt.Errorf("JWT_JWKS_URL is required when protected gRPC-Web methods are configured")
	}
	keys, err := gatewayauth.NewRemoteJWKSKeyProvider(gatewayauth.JWKSConfig{
		URL:      cfg.JWTJWKSURL,
		CacheTTL: cfg.JWTJWKSCacheTTL,
	}, &http.Client{Timeout: cfg.JWTJWKSFetchTimeout}, logger)
	if err != nil {
		return nil, fmt.Errorf("configure grpc-web jwks key provider: %w", err)
	}
	verifier, err := gatewayauth.NewVerifier(cfg.JWTIssuer, cfg.JWTAudience, cfg.JWTAllowedAlgs, cfg.JWTClockSkew, keys)
	if err != nil {
		return nil, fmt.Errorf("configure grpc-web jwt verifier: %w", err)
	}
	return verifier, nil
}
