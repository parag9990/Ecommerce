package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

type MySQLOTPRepository struct {
	db *sql.DB
}

func NewMySQLOTPRepository(db *sql.DB) (*MySQLOTPRepository, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	return &MySQLOTPRepository{db: db}, nil
}

func (r *MySQLOTPRepository) CreateOTPChallenge(ctx context.Context, challenge domain.OTPChallenge) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO otp_challenges (
			challenge_id,
			account_id,
			target,
			channel,
			purpose,
			otp_hash,
			attempts,
			max_attempts,
			expires_at,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		challenge.ChallengeID,
		nullableStringPtr(challenge.AccountID),
		challenge.Target,
		string(challenge.Channel),
		string(challenge.Purpose),
		challenge.OTPHash,
		challenge.Attempts,
		challenge.MaxAttempts,
		challenge.ExpiresAt.UTC(),
		challenge.CreatedAt.UTC(),
	)
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrInvalidOTPRequest
		}
		return fmt.Errorf("insert otp challenge: %w", err)
	}
	return nil
}

func (r *MySQLOTPRepository) ExpireOlderActiveOTPChallenges(ctx context.Context, target string, purpose domain.OTPPurpose, now time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE otp_challenges
		SET expires_at = ?
		WHERE target = ?
		  AND purpose = ?
		  AND verified_at IS NULL
		  AND expires_at > ?
	`, now.UTC(), target, string(purpose), now.UTC())
	if err != nil {
		return fmt.Errorf("expire older otp challenges: %w", err)
	}
	return nil
}

func (r *MySQLOTPRepository) ExpireOTPChallenge(ctx context.Context, challengeID string, now time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE otp_challenges
		SET expires_at = ?
		WHERE challenge_id = ?
		  AND verified_at IS NULL
	`, now.UTC(), challengeID)
	if err != nil {
		return fmt.Errorf("expire otp challenge: %w", err)
	}
	return nil
}

func (r *MySQLOTPRepository) MutateLockedOTPChallenge(ctx context.Context, challengeID string, mutate func(domain.OTPChallenge) (domain.OTPChallengeMutation, error)) error {
	if mutate == nil {
		return errors.New("otp challenge mutation callback is required")
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("begin otp challenge transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	challenge, err := scanOTPChallenge(tx.QueryRowContext(ctx, `
		SELECT
			challenge_id,
			account_id,
			target,
			channel,
			purpose,
			otp_hash,
			attempts,
			max_attempts,
			verified_at,
			expires_at,
			created_at
		FROM otp_challenges
		WHERE challenge_id = ?
		LIMIT 1
		FOR UPDATE
	`, challengeID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOTPChallengeNotFound
		}
		return fmt.Errorf("lock otp challenge: %w", err)
	}

	mutation, businessErr := mutate(challenge)
	if !mutation.IsZero() {
		if err := applyOTPChallengeMutation(ctx, tx, challenge.ChallengeID, mutation); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit otp challenge transaction: %w", err)
	}
	return businessErr
}

func applyOTPChallengeMutation(ctx context.Context, tx *sql.Tx, challengeID string, mutation domain.OTPChallengeMutation) error {
	if mutation.IncrementAttempts {
		result, err := tx.ExecContext(ctx, `
			UPDATE otp_challenges
			SET attempts = attempts + 1
			WHERE challenge_id = ?
			  AND verified_at IS NULL
		`, challengeID)
		if err != nil {
			return fmt.Errorf("increment otp attempts: %w", err)
		}
		if err := ensureAffected(result, domain.ErrOTPChallengeNotFound); err != nil {
			return err
		}
	}

	if mutation.MarkVerifiedAt != nil {
		result, err := tx.ExecContext(ctx, `
			UPDATE otp_challenges
			SET verified_at = ?
			WHERE challenge_id = ?
			  AND verified_at IS NULL
		`, mutation.MarkVerifiedAt.UTC(), challengeID)
		if err != nil {
			return fmt.Errorf("mark otp verified: %w", err)
		}
		if err := ensureAffected(result, domain.ErrOTPAlreadyUsed); err != nil {
			return err
		}
	}

	if mutation.MarkEmailVerifiedAccount != "" {
		if err := markAccountVerified(ctx, tx, mutation.MarkEmailVerifiedAccount, "email_verified"); err != nil {
			return err
		}
	}
	if mutation.MarkPhoneVerifiedAccount != "" {
		if err := markAccountVerified(ctx, tx, mutation.MarkPhoneVerifiedAccount, "phone_verified"); err != nil {
			return err
		}
	}
	return nil
}

func markAccountVerified(ctx context.Context, tx *sql.Tx, accountID string, column string) error {
	if column != "email_verified" && column != "phone_verified" {
		return errors.New("unsupported account verification column")
	}
	query := fmt.Sprintf("UPDATE auth_accounts SET %s = TRUE WHERE account_id = ?", column)
	result, err := tx.ExecContext(ctx, query, accountID)
	if err != nil {
		return fmt.Errorf("mark account verified: %w", err)
	}
	return ensureAffected(result, domain.ErrAccountNotFound)
}

func scanOTPChallenge(row sqlScanner) (domain.OTPChallenge, error) {
	var (
		challenge  domain.OTPChallenge
		accountID  sql.NullString
		channel    string
		purpose    string
		verifiedAt sql.NullTime
	)
	if err := row.Scan(
		&challenge.ChallengeID,
		&accountID,
		&challenge.Target,
		&channel,
		&purpose,
		&challenge.OTPHash,
		&challenge.Attempts,
		&challenge.MaxAttempts,
		&verifiedAt,
		&challenge.ExpiresAt,
		&challenge.CreatedAt,
	); err != nil {
		return domain.OTPChallenge{}, err
	}
	if accountID.Valid {
		value := accountID.String
		challenge.AccountID = &value
	}
	challenge.Channel = domain.OTPChannel(channel)
	challenge.Purpose = domain.OTPPurpose(purpose)
	challenge.VerifiedAt = nullTimePtr(verifiedAt)
	challenge.ExpiresAt = challenge.ExpiresAt.UTC()
	challenge.CreatedAt = challenge.CreatedAt.UTC()
	return challenge, nil
}

func nullableStringPtr(value *string) any {
	if value == nil || *value == "" {
		return nil
	}
	return *value
}
