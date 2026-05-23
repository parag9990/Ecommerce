package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid access token")
	ErrMissingClaim = errors.New("missing required access token claim")
)

type KeyProvider interface {
	Keyfunc(ctx context.Context) jwt.Keyfunc
}

type Verifier struct {
	parser      *jwt.Parser
	keyProvider KeyProvider
}

func NewVerifier(issuer string, audience string, allowedAlgs []string, clockSkew time.Duration, keys KeyProvider) (*Verifier, error) {
	issuer = strings.TrimSpace(issuer)
	audience = strings.TrimSpace(audience)
	allowedAlgs = normalizeAllowedAlgorithms(allowedAlgs)
	if issuer == "" {
		return nil, fmt.Errorf("jwt issuer is required")
	}
	if audience == "" {
		return nil, fmt.Errorf("jwt audience is required")
	}
	if len(allowedAlgs) == 0 {
		return nil, fmt.Errorf("at least one jwt signing algorithm is required")
	}
	if keys == nil {
		return nil, fmt.Errorf("jwt key provider is required")
	}
	if clockSkew < 0 {
		return nil, fmt.Errorf("jwt clock skew must not be negative")
	}

	return &Verifier{
		parser: jwt.NewParser(
			jwt.WithIssuer(issuer),
			jwt.WithAudience(audience),
			jwt.WithValidMethods(allowedAlgs),
			jwt.WithLeeway(clockSkew),
			jwt.WithExpirationRequired(),
			jwt.WithIssuedAt(),
		),
		keyProvider: keys,
	}, nil
}

func (v *Verifier) Verify(ctx context.Context, raw string) (AccessClaims, error) {
	if err := ctx.Err(); err != nil {
		return AccessClaims{}, err
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return AccessClaims{}, ErrInvalidToken
	}

	claims := AccessClaims{}
	token, err := v.parser.ParseWithClaims(raw, &claims, v.keyProvider.Keyfunc(ctx))
	if err != nil || token == nil || !token.Valid {
		return AccessClaims{}, ErrInvalidToken
	}
	if err := validateAccessClaims(claims); err != nil {
		return AccessClaims{}, err
	}
	claims.Roles = normalizeRoles(claims.Roles)
	return claims, nil
}

func validateAccessClaims(claims AccessClaims) error {
	if strings.TrimSpace(claims.UserID()) == "" || strings.TrimSpace(claims.SessionID) == "" {
		return ErrMissingClaim
	}
	if len(normalizeRoles(claims.Roles)) == 0 {
		return ErrMissingClaim
	}
	if claims.ExpiresAt == nil || claims.IssuedAt == nil {
		return ErrMissingClaim
	}
	if strings.TrimSpace(claims.TokenType) != AccessTokenType {
		return ErrInvalidToken
	}
	return nil
}

func normalizeAllowedAlgorithms(algorithms []string) []string {
	seen := make(map[string]struct{}, len(algorithms))
	out := make([]string, 0, len(algorithms))
	for _, algorithm := range algorithms {
		algorithm = strings.TrimSpace(algorithm)
		if algorithm == "" {
			continue
		}
		if _, exists := seen[algorithm]; exists {
			continue
		}
		seen[algorithm] = struct{}{}
		out = append(out, algorithm)
	}
	return out
}

func normalizeRoles(roles []string) []string {
	seen := make(map[string]struct{}, len(roles))
	out := make([]string, 0, len(roles))
	for _, role := range roles {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}
		if _, exists := seen[role]; exists {
			continue
		}
		seen[role] = struct{}{}
		out = append(out, role)
	}
	return out
}
