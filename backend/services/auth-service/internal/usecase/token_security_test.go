package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
	tokensecurity "github.com/example/ecommerce-platform/backend/services/auth-service/internal/security/token"
	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/sessionlink"
)

const testRefreshPepper = "test-refresh-pepper"

func TestRefreshTokenRotationStoresOnlyHashAndRevokesOldToken(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	repo := newSecurityRefreshTokenRepository()
	uc := newSecurityTokenUsecase(t, repo, nil, now, time.Hour, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))

	issued, err := uc.IssueTokenPair(context.Background(), IssueTokenPairInput{
		AccountID: "auth_123",
		UserID:    "user_123",
		SessionID: "sess_123",
		Roles:     []string{"buyer"},
	})
	if err != nil {
		t.Fatalf("IssueTokenPair() error = %v", err)
	}

	initialRows := repo.listBySession("sess_123")
	if len(initialRows) != 1 {
		t.Fatalf("initial refresh rows = %d, want 1", len(initialRows))
	}
	if initialRows[0].TokenHash == "" || initialRows[0].TokenHash == issued.RefreshToken {
		t.Fatalf("initial refresh token stored unsafely: row=%+v plain=%q", initialRows[0], issued.RefreshToken)
	}

	refreshed, err := uc.RefreshToken(context.Background(), RefreshTokenInput{
		RefreshToken: issued.RefreshToken,
	})
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}
	if refreshed.RefreshToken == "" || refreshed.RefreshToken == issued.RefreshToken {
		t.Fatalf("rotated refresh token = %q, original = %q", refreshed.RefreshToken, issued.RefreshToken)
	}

	rows := repo.listBySession("sess_123")
	if len(rows) != 2 {
		t.Fatalf("refresh rows = %d, want 2", len(rows))
	}

	oldRow, newRow := rows[0], rows[1]
	if oldRow.RevokedAt == nil || !oldRow.RevokedAt.Equal(now) {
		t.Fatalf("old refresh token revoked_at = %v, want %v", oldRow.RevokedAt, now)
	}
	if newRow.RevokedAt != nil {
		t.Fatalf("new refresh token unexpectedly revoked: %+v", newRow)
	}
	if newRow.ParentTokenID != oldRow.TokenID {
		t.Fatalf("new parent token id = %q, want %q", newRow.ParentTokenID, oldRow.TokenID)
	}
	for _, row := range rows {
		if row.TokenHash == issued.RefreshToken || row.TokenHash == refreshed.RefreshToken {
			t.Fatalf("plain refresh token was stored in row: %+v", row)
		}
	}
}

func TestRefreshTokenReuseRevokesSessionFamilyAndRedactsToken(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	repo := newSecurityRefreshTokenRepository()
	linker := &recordingTokenSessionLinker{}
	logs := &bytes.Buffer{}
	uc := newSecurityTokenUsecase(t, repo, linker, now, time.Hour, slog.New(slog.NewTextHandler(logs, nil)))

	issued, err := uc.IssueTokenPair(context.Background(), IssueTokenPairInput{
		AccountID: "auth_123",
		UserID:    "user_123",
		SessionID: "sess_123",
		Roles:     []string{"buyer"},
	})
	if err != nil {
		t.Fatalf("IssueTokenPair() error = %v", err)
	}

	refreshed, err := uc.RefreshToken(context.Background(), RefreshTokenInput{
		RefreshToken: issued.RefreshToken,
		TraceID:      "trace_rotate",
	})
	if err != nil {
		t.Fatalf("first RefreshToken() error = %v", err)
	}

	_, err = uc.RefreshToken(context.Background(), RefreshTokenInput{
		RefreshToken: issued.RefreshToken,
		TraceID:      "trace_reuse",
		Device:       SessionDeviceInput{Fingerprint: "raw-device-fingerprint"},
		Network:      SessionNetworkInput{IPAddress: "203.0.113.10"},
	})
	if !errors.Is(err, domain.ErrRefreshTokenReuse) {
		t.Fatalf("reused RefreshToken() error = %v, want refresh token reuse", err)
	}

	for _, row := range repo.listBySession("sess_123") {
		if row.RevokedAt == nil {
			t.Fatalf("session token family member was not revoked: %+v", row)
		}
	}
	if linker.reuse.SessionID != "sess_123" || linker.reuse.TokenID == "" || linker.reuse.ActionTaken != sessionlink.ActionRevokedSessionFamily {
		t.Fatalf("reuse event = %+v", linker.reuse)
	}

	output := logs.String()
	if strings.Contains(output, issued.RefreshToken) || strings.Contains(output, refreshed.RefreshToken) {
		t.Fatalf("refresh token leaked into logs: %s", output)
	}
	if strings.Contains(output, "raw-device-fingerprint") || strings.Contains(output, "203.0.113.10") {
		t.Fatalf("raw session metadata leaked into logs: %s", output)
	}
}

func TestRefreshTokenExpiredRejectsWithoutRotating(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	repo := newSecurityRefreshTokenRepository()
	uc := newSecurityTokenUsecase(t, repo, nil, now, time.Minute, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))

	issued, err := uc.IssueTokenPair(context.Background(), IssueTokenPairInput{
		AccountID: "auth_123",
		UserID:    "user_123",
		SessionID: "sess_123",
		Roles:     []string{"buyer"},
	})
	if err != nil {
		t.Fatalf("IssueTokenPair() error = %v", err)
	}

	uc.WithClock(fixedClock{now: now.Add(2 * time.Minute)})
	_, err = uc.RefreshToken(context.Background(), RefreshTokenInput{
		RefreshToken: issued.RefreshToken,
	})
	if !errors.Is(err, domain.ErrInvalidRefreshToken) {
		t.Fatalf("expired RefreshToken() error = %v, want invalid refresh token", err)
	}
	if rows := repo.listBySession("sess_123"); len(rows) != 1 {
		t.Fatalf("expired refresh token rotated unexpectedly, rows = %d", len(rows))
	}
}

func TestRefreshTokenConcurrentReuseOnlyOneSucceeds(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	repo := newSecurityRefreshTokenRepository()
	uc := newSecurityTokenUsecase(t, repo, nil, now, time.Hour, slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))

	issued, err := uc.IssueTokenPair(context.Background(), IssueTokenPairInput{
		AccountID: "auth_123",
		UserID:    "user_123",
		SessionID: "sess_123",
		Roles:     []string{"buyer"},
	})
	if err != nil {
		t.Fatalf("IssueTokenPair() error = %v", err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := uc.RefreshToken(context.Background(), RefreshTokenInput{RefreshToken: issued.RefreshToken})
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	successes := 0
	rejections := 0
	for err := range errs {
		if err == nil {
			successes++
			continue
		}
		if errors.Is(err, domain.ErrRefreshTokenReuse) {
			rejections++
			continue
		}
		t.Fatalf("RefreshToken() unexpected concurrent error = %v", err)
	}
	if successes != 1 || rejections != 1 {
		t.Fatalf("concurrent refresh results: successes=%d rejections=%d, want 1/1", successes, rejections)
	}
}

func TestLogoutRevokedRefreshTokenCannotRefreshAndRedactsToken(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	repo := newSecurityRefreshTokenRepository()
	linker := &recordingTokenSessionLinker{}
	logs := &bytes.Buffer{}
	uc := newSecurityTokenUsecase(t, repo, linker, now, time.Hour, slog.New(slog.NewTextHandler(logs, nil)))

	issued, err := uc.IssueTokenPair(context.Background(), IssueTokenPairInput{
		AccountID: "auth_123",
		UserID:    "user_123",
		SessionID: "sess_123",
		Roles:     []string{"buyer"},
	})
	if err != nil {
		t.Fatalf("IssueTokenPair() error = %v", err)
	}

	err = uc.Logout(context.Background(), LogoutInput{
		RefreshToken: issued.RefreshToken,
		Reason:       "user_requested",
	})
	if err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	_, err = uc.RefreshToken(context.Background(), RefreshTokenInput{
		RefreshToken: issued.RefreshToken,
	})
	if !errors.Is(err, domain.ErrRefreshTokenReuse) {
		t.Fatalf("post-logout RefreshToken() error = %v, want refresh token reuse", err)
	}
	if linker.logout.SessionID != "sess_123" || linker.logout.Reason != "user_requested" {
		t.Fatalf("logout event = %+v", linker.logout)
	}
	if output := logs.String(); strings.Contains(output, issued.RefreshToken) {
		t.Fatalf("refresh token leaked into logs: %s", output)
	}
}

func newSecurityTokenUsecase(t *testing.T, repo *securityRefreshTokenRepository, linker SessionLinker, now time.Time, refreshTTL time.Duration, logger *slog.Logger) *TokenUsecase {
	t.Helper()

	if linker == nil {
		linker = &recordingTokenSessionLinker{}
	}
	uc, err := NewTokenUsecase(
		repo,
		securityTokenSubjectRepository{
			subjects: map[string]domain.TokenSubject{
				"auth_123": {
					AccountID: "auth_123",
					UserID:    "user_123",
					Status:    domain.AccountStatusActive,
				},
			},
		},
		securityRoleRepository{
			roles: map[string][]string{
				"auth_123": {"buyer"},
			},
		},
		&recordingAccessTokenIssuer{},
		stubAccessTokenVerifier{},
		linker,
		securityPrivacyHasher{},
		tokensecurity.JWKSet{},
		TokenUsecaseConfig{
			RefreshTokenTTL:    refreshTTL,
			RefreshTokenPepper: testRefreshPepper,
		},
		logger,
	)
	if err != nil {
		t.Fatalf("NewTokenUsecase() error = %v", err)
	}
	uc.WithClock(fixedClock{now: now})
	return uc
}

type securityRefreshTokenRepository struct {
	mu       sync.Mutex
	byHash   map[string]domain.RefreshToken
	hashByID map[string]string
}

func newSecurityRefreshTokenRepository() *securityRefreshTokenRepository {
	return &securityRefreshTokenRepository{
		byHash:   map[string]domain.RefreshToken{},
		hashByID: map[string]string{},
	}
}

func (r *securityRefreshTokenRepository) InsertRefreshToken(ctx context.Context, refreshToken domain.RefreshToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byHash[refreshToken.TokenHash]; exists {
		return domain.ErrInvalidRefreshToken
	}
	r.byHash[refreshToken.TokenHash] = refreshToken
	r.hashByID[refreshToken.TokenID] = refreshToken.TokenHash
	return nil
}

func (r *securityRefreshTokenRepository) FindRefreshTokenByHash(ctx context.Context, tokenHash string) (domain.RefreshToken, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	refreshToken, ok := r.byHash[tokenHash]
	if !ok {
		return domain.RefreshToken{}, domain.ErrInvalidRefreshToken
	}
	return refreshToken, nil
}

func (r *securityRefreshTokenRepository) RotateRefreshToken(ctx context.Context, oldTokenID string, next domain.RefreshToken, revokedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	oldHash, ok := r.hashByID[oldTokenID]
	if !ok {
		return domain.ErrInvalidRefreshToken
	}
	old := r.byHash[oldHash]
	now := revokedAt.UTC()
	if old.RevokedAt != nil {
		return domain.ErrRefreshTokenReuse
	}
	if !old.IsActive(now) {
		return domain.ErrInvalidRefreshToken
	}
	if old.AccountID != next.AccountID || old.SessionID != next.SessionID {
		return domain.ErrInvalidRefreshToken
	}

	old.RevokedAt = &now
	r.byHash[oldHash] = old
	r.byHash[next.TokenHash] = next
	r.hashByID[next.TokenID] = next.TokenHash
	return nil
}

func (r *securityRefreshTokenRepository) RevokeSessionTokens(ctx context.Context, sessionID string, revokedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := revokedAt.UTC()
	for hash, token := range r.byHash {
		if token.SessionID == sessionID && token.RevokedAt == nil {
			token.RevokedAt = &now
			r.byHash[hash] = token
		}
	}
	return nil
}

func (r *securityRefreshTokenRepository) RevokeAccountTokens(ctx context.Context, accountID string, revokedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := revokedAt.UTC()
	for hash, token := range r.byHash {
		if token.AccountID == accountID && token.RevokedAt == nil {
			token.RevokedAt = &now
			r.byHash[hash] = token
		}
	}
	return nil
}

func (r *securityRefreshTokenRepository) listBySession(sessionID string) []domain.RefreshToken {
	r.mu.Lock()
	defer r.mu.Unlock()

	rows := make([]domain.RefreshToken, 0)
	for _, token := range r.byHash {
		if token.SessionID == sessionID {
			rows = append(rows, token)
		}
	}
	for i := 0; i < len(rows); i++ {
		for j := i + 1; j < len(rows); j++ {
			if rows[j].ParentTokenID == "" && rows[i].ParentTokenID != "" {
				rows[i], rows[j] = rows[j], rows[i]
				continue
			}
			if rows[i].ParentTokenID == "" && rows[j].ParentTokenID != "" {
				continue
			}
			if rows[j].CreatedAt.Before(rows[i].CreatedAt) || (rows[j].CreatedAt.Equal(rows[i].CreatedAt) && rows[j].TokenID < rows[i].TokenID) {
				rows[i], rows[j] = rows[j], rows[i]
			}
		}
	}
	return rows
}

type securityTokenSubjectRepository struct {
	subjects map[string]domain.TokenSubject
}

func (r securityTokenSubjectRepository) FindTokenSubject(ctx context.Context, accountID string) (domain.TokenSubject, error) {
	subject, ok := r.subjects[accountID]
	if !ok {
		return domain.TokenSubject{}, domain.ErrAccountNotFound
	}
	return subject, nil
}

type securityRoleRepository struct {
	roles map[string][]string
}

func (r securityRoleRepository) ActiveRoles(ctx context.Context, accountID string) ([]string, error) {
	roles, ok := r.roles[accountID]
	if !ok {
		return nil, domain.ErrRoleRequired
	}
	return append([]string(nil), roles...), nil
}

type recordingAccessTokenIssuer struct {
	mu     sync.Mutex
	issued int
}

func (i *recordingAccessTokenIssuer) IssueAccessToken(input tokensecurity.AccessTokenInput) (tokensecurity.AccessTokenResult, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.issued++
	return tokensecurity.AccessTokenResult{
		Token:     fmt.Sprintf("access-token-%d", i.issued),
		ExpiresIn: int64((15 * time.Minute).Seconds()),
	}, nil
}

type stubAccessTokenVerifier struct{}

func (stubAccessTokenVerifier) VerifyAccessToken(tokenString string) (domain.TokenClaims, error) {
	return domain.TokenClaims{}, nil
}

type recordingTokenSessionLinker struct {
	mu     sync.Mutex
	logout sessionlink.LogoutSucceededEvent
	reuse  sessionlink.RefreshReuseDetectedEvent
}

func (l *recordingTokenSessionLinker) RecordLoginSucceeded(ctx context.Context, event sessionlink.LoginSucceededEvent) error {
	return nil
}

func (l *recordingTokenSessionLinker) RecordSignupSucceeded(ctx context.Context, event sessionlink.SignupSucceededEvent) error {
	return nil
}

func (l *recordingTokenSessionLinker) RecordLogoutSucceeded(ctx context.Context, event sessionlink.LogoutSucceededEvent) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.logout = event
	return nil
}

func (l *recordingTokenSessionLinker) RecordRefreshReuseDetected(ctx context.Context, event sessionlink.RefreshReuseDetectedEvent) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.reuse = event
	return nil
}

type securityPrivacyHasher struct{}

func (securityPrivacyHasher) HashIP(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return "hash:" + value
}

func (securityPrivacyHasher) HashDeviceFingerprint(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return "hash:" + value
}
