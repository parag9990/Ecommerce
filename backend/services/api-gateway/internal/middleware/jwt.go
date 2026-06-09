package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/parag/ecommerce/backend/services/api-gateway/internal/authctx"
)

var (
	errMissingToken = errors.New("missing bearer token")
	errInvalidToken = errors.New("invalid token")
)

type TokenVerifier interface {
	Verify(ctx context.Context, token string) (authctx.Claims, error)
}

type AuthMiddleware struct {
	verifier TokenVerifier
}

func NewAuthMiddleware(verifier TokenVerifier) AuthMiddleware {
	if verifier == nil {
		panic("token verifier is required")
	}
	return AuthMiddleware{verifier: verifier}
}

func (m AuthMiddleware) Required(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := bearerToken(r)
		if err != nil {
			writeAPIError(w, r, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "Authentication required")
			return
		}

		claims, err := m.verifier.Verify(r.Context(), token)
		if err != nil {
			writeAPIError(w, r, http.StatusUnauthorized, "AUTHENTICATION_REQUIRED", "Authentication required")
			return
		}

		next.ServeHTTP(w, r.WithContext(authctx.WithClaims(r.Context(), claims)))
	})
}

type JWTVerifier struct {
	secret   []byte
	issuer   string
	audience string
	skew     time.Duration
	clock    Clock
}

type Clock interface {
	Now() time.Time
}

func NewJWTVerifier(secret string, issuer string, audience string, skew time.Duration, clock Clock) *JWTVerifier {
	if strings.TrimSpace(secret) == "" {
		panic("jwt secret is required")
	}
	if clock == nil {
		clock = systemClock{}
	}
	return &JWTVerifier{
		secret:   []byte(secret),
		issuer:   strings.TrimSpace(issuer),
		audience: strings.TrimSpace(audience),
		skew:     skew,
		clock:    clock,
	}
}

func (v *JWTVerifier) Verify(ctx context.Context, token string) (authctx.Claims, error) {
	select {
	case <-ctx.Done():
		return authctx.Claims{}, ctx.Err()
	default:
	}

	headerPayload, signature, err := splitToken(token)
	if err != nil {
		return authctx.Claims{}, err
	}

	headerBytes, err := decodeJWTPart(headerPayload[0])
	if err != nil {
		return authctx.Claims{}, fmt.Errorf("%w: header", errInvalidToken)
	}
	var header jwtHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return authctx.Claims{}, fmt.Errorf("%w: header json", errInvalidToken)
	}
	if header.Alg != "HS256" {
		return authctx.Claims{}, fmt.Errorf("%w: unsupported alg", errInvalidToken)
	}

	if !v.validSignature(strings.Join(headerPayload, "."), signature) {
		return authctx.Claims{}, fmt.Errorf("%w: signature", errInvalidToken)
	}

	payloadBytes, err := decodeJWTPart(headerPayload[1])
	if err != nil {
		return authctx.Claims{}, fmt.Errorf("%w: payload", errInvalidToken)
	}
	var payload jwtPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return authctx.Claims{}, fmt.Errorf("%w: payload json", errInvalidToken)
	}

	claims, err := v.claimsFromPayload(payload)
	if err != nil {
		return authctx.Claims{}, err
	}
	return claims, nil
}

func (v *JWTVerifier) claimsFromPayload(payload jwtPayload) (authctx.Claims, error) {
	userID := strings.TrimSpace(payload.UserID)
	if userID == "" {
		userID = strings.TrimSpace(payload.Subject)
	}
	if userID == "" {
		return authctx.Claims{}, fmt.Errorf("%w: sub", errInvalidToken)
	}
	if payload.TokenType != "access" {
		return authctx.Claims{}, fmt.Errorf("%w: token_type", errInvalidToken)
	}

	now := v.clock.Now()
	expiresAt := unixTime(payload.ExpiresAt)
	if expiresAt.IsZero() || now.After(expiresAt.Add(v.skew)) {
		return authctx.Claims{}, fmt.Errorf("%w: exp", errInvalidToken)
	}
	notBefore := unixTime(payload.NotBefore)
	if !notBefore.IsZero() && now.Add(v.skew).Before(notBefore) {
		return authctx.Claims{}, fmt.Errorf("%w: nbf", errInvalidToken)
	}
	issuedAt := unixTime(payload.IssuedAt)
	if !issuedAt.IsZero() && now.Add(v.skew).Before(issuedAt) {
		return authctx.Claims{}, fmt.Errorf("%w: iat", errInvalidToken)
	}
	if v.issuer != "" && payload.Issuer != v.issuer {
		return authctx.Claims{}, fmt.Errorf("%w: iss", errInvalidToken)
	}
	if v.audience != "" && !payload.Audience.Contains(v.audience) {
		return authctx.Claims{}, fmt.Errorf("%w: aud", errInvalidToken)
	}

	return authctx.Claims{
		UserID:    userID,
		SessionID: strings.TrimSpace(payload.SessionID),
		SellerID:  strings.TrimSpace(payload.SellerID),
		Roles:     cleanRoles(payload.Roles),
		TokenType: payload.TokenType,
		Issuer:    payload.Issuer,
		Audience:  []string(payload.Audience),
		ExpiresAt: expiresAt,
		NotBefore: notBefore,
		IssuedAt:  issuedAt,
	}, nil
}

func bearerToken(r *http.Request) (string, error) {
	value := strings.TrimSpace(r.Header.Get("Authorization"))
	if value == "" {
		return "", errMissingToken
	}
	scheme, token, ok := strings.Cut(value, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
		return "", errMissingToken
	}
	return strings.TrimSpace(token), nil
}

func splitToken(token string) ([]string, []byte, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, nil, fmt.Errorf("%w: segments", errInvalidToken)
	}
	signature, err := decodeJWTPart(parts[2])
	if err != nil {
		return nil, nil, fmt.Errorf("%w: signature", errInvalidToken)
	}
	return parts[:2], signature, nil
}

func (v *JWTVerifier) validSignature(signingInput string, signature []byte) bool {
	mac := hmac.New(sha256.New, v.secret)
	_, _ = mac.Write([]byte(signingInput))
	return hmac.Equal(signature, mac.Sum(nil))
}

func decodeJWTPart(value string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(value)
}

func unixTime(value int64) time.Time {
	if value <= 0 {
		return time.Time{}
	}
	return time.Unix(value, 0).UTC()
}

func cleanRoles(roles []string) []string {
	cleaned := make([]string, 0, len(roles))
	for _, role := range roles {
		role = strings.TrimSpace(role)
		if role != "" {
			cleaned = append(cleaned, role)
		}
	}
	return cleaned
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
	KID string `json:"kid"`
}

type jwtPayload struct {
	Subject   string        `json:"sub"`
	UserID    string        `json:"user_id"`
	SessionID string        `json:"sid"`
	SellerID  string        `json:"seller_id"`
	Roles     []string      `json:"roles"`
	TokenType string        `json:"token_type"`
	ExpiresAt int64         `json:"exp"`
	NotBefore int64         `json:"nbf"`
	IssuedAt  int64         `json:"iat"`
	Issuer    string        `json:"iss"`
	Audience  audienceClaim `json:"aud"`
}

type audienceClaim []string

func (c *audienceClaim) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*c = []string{single}
		return nil
	}

	var multiple []string
	if err := json.Unmarshal(data, &multiple); err != nil {
		return err
	}
	*c = multiple
	return nil
}

func (c audienceClaim) Contains(value string) bool {
	for _, existing := range c {
		if existing == value {
			return true
		}
	}
	return false
}

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now().UTC()
}
