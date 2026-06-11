package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

type MySQLTokenRepository struct {
	db *sql.DB
}

func NewMySQLTokenRepository(db *sql.DB) (*MySQLTokenRepository, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	return &MySQLTokenRepository{db: db}, nil
}

func (r *MySQLTokenRepository) InsertRefreshToken(ctx context.Context, refreshToken domain.RefreshToken) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO refresh_tokens (
			token_id,
			account_id,
			session_id,
			token_hash,
			parent_token_id,
			expires_at,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`, refreshToken.TokenID, refreshToken.AccountID, refreshToken.SessionID, refreshToken.TokenHash, nullableString(refreshToken.ParentTokenID), refreshToken.ExpiresAt.UTC(), refreshToken.CreatedAt.UTC())
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrInvalidRefreshToken
		}
		return fmt.Errorf("insert refresh token: %w", err)
	}
	return nil
}

func (r *MySQLTokenRepository) FindRefreshTokenByHash(ctx context.Context, tokenHash string) (domain.RefreshToken, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			token_id,
			account_id,
			session_id,
			token_hash,
			parent_token_id,
			revoked_at,
			expires_at,
			created_at
		FROM refresh_tokens
		WHERE token_hash = ?
		LIMIT 1
	`, tokenHash)
	refreshToken, err := scanRefreshToken(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.RefreshToken{}, domain.ErrInvalidRefreshToken
		}
		return domain.RefreshToken{}, fmt.Errorf("query refresh token: %w", err)
	}
	return refreshToken, nil
}

func (r *MySQLTokenRepository) RotateRefreshToken(ctx context.Context, oldTokenID string, next domain.RefreshToken, revokedAt time.Time) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("begin refresh rotation: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	old, err := scanRefreshToken(tx.QueryRowContext(ctx, `
		SELECT
			token_id,
			account_id,
			session_id,
			token_hash,
			parent_token_id,
			revoked_at,
			expires_at,
			created_at
		FROM refresh_tokens
		WHERE token_id = ?
		LIMIT 1
		FOR UPDATE
	`, oldTokenID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrInvalidRefreshToken
		}
		return fmt.Errorf("lock refresh token: %w", err)
	}

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

	result, err := tx.ExecContext(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = ?
		WHERE token_id = ? AND revoked_at IS NULL
	`, now, oldTokenID)
	if err != nil {
		return fmt.Errorf("revoke old refresh token: %w", err)
	}
	if err := ensureAffected(result, domain.ErrRefreshTokenReuse); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO refresh_tokens (
			token_id,
			account_id,
			session_id,
			token_hash,
			parent_token_id,
			expires_at,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`, next.TokenID, next.AccountID, next.SessionID, next.TokenHash, nullableString(next.ParentTokenID), next.ExpiresAt.UTC(), next.CreatedAt.UTC())
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrInvalidRefreshToken
		}
		return fmt.Errorf("insert rotated refresh token: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit refresh rotation: %w", err)
	}
	return nil
}

func (r *MySQLTokenRepository) RevokeSessionTokens(ctx context.Context, sessionID string, revokedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = ?
		WHERE session_id = ? AND revoked_at IS NULL
	`, revokedAt.UTC(), sessionID)
	if err != nil {
		return fmt.Errorf("revoke session refresh tokens: %w", err)
	}
	return nil
}

func (r *MySQLTokenRepository) RevokeAccountTokens(ctx context.Context, accountID string, revokedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = ?
		WHERE account_id = ? AND revoked_at IS NULL
	`, revokedAt.UTC(), accountID)
	if err != nil {
		return fmt.Errorf("revoke account refresh tokens: %w", err)
	}
	return nil
}

type sqlScanner interface {
	Scan(dest ...any) error
}

func scanRefreshToken(row sqlScanner) (domain.RefreshToken, error) {
	var (
		refreshToken domain.RefreshToken
		parentToken  sql.NullString
		revokedAt    sql.NullTime
	)
	if err := row.Scan(
		&refreshToken.TokenID,
		&refreshToken.AccountID,
		&refreshToken.SessionID,
		&refreshToken.TokenHash,
		&parentToken,
		&revokedAt,
		&refreshToken.ExpiresAt,
		&refreshToken.CreatedAt,
	); err != nil {
		return domain.RefreshToken{}, err
	}
	if parentToken.Valid {
		refreshToken.ParentTokenID = parentToken.String
	}
	refreshToken.RevokedAt = nullTimePtr(revokedAt)
	refreshToken.ExpiresAt = refreshToken.ExpiresAt.UTC()
	refreshToken.CreatedAt = refreshToken.CreatedAt.UTC()
	return refreshToken, nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
