package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
	"github.com/go-sql-driver/mysql"
)

type MySQLCredentialRepository struct {
	db *sql.DB
}

func NewMySQLCredentialRepository(db *sql.DB) (*MySQLCredentialRepository, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	return &MySQLCredentialRepository{db: db}, nil
}

func (r *MySQLCredentialRepository) CreateCredential(ctx context.Context, credential domain.Credential) error {
	changedAt := nullableTime(credential.PasswordChangedAt)
	lockedUntil := nullableTime(credential.LockedUntil)

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO credentials (
			account_id,
			password_hash,
			password_algo,
			password_changed_at,
			failed_attempts,
			locked_until
		) VALUES (?, ?, ?, ?, ?, ?)
	`, credential.AccountID, credential.PasswordHash, credential.PasswordAlgo, changedAt, credential.FailedAttempts, lockedUntil)
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrDuplicateCredential
		}
		return fmt.Errorf("insert credential: %w", err)
	}
	return nil
}

func (r *MySQLCredentialRepository) FindByIdentifier(ctx context.Context, identifier string) (domain.AccountCredential, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			a.account_id,
			a.status,
			a.email_verified,
			a.phone_verified,
			c.password_hash,
			c.password_algo,
			c.password_changed_at,
			c.failed_attempts,
			c.locked_until,
			c.created_at,
			c.updated_at
		FROM auth_accounts a
		JOIN credentials c ON c.account_id = a.account_id
		WHERE a.email = ? OR a.phone = ?
		LIMIT 1
	`, identifier, identifier)

	var (
		accountCredential domain.AccountCredential
		passwordChangedAt sql.NullTime
		lockedUntil       sql.NullTime
		createdAt         time.Time
		updatedAt         time.Time
		status            string
	)

	err := row.Scan(
		&accountCredential.AccountID,
		&status,
		&accountCredential.EmailVerified,
		&accountCredential.PhoneVerified,
		&accountCredential.Credential.PasswordHash,
		&accountCredential.Credential.PasswordAlgo,
		&passwordChangedAt,
		&accountCredential.Credential.FailedAttempts,
		&lockedUntil,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.AccountCredential{}, domain.ErrCredentialNotFound
		}
		return domain.AccountCredential{}, fmt.Errorf("query credential by identifier: %w", err)
	}

	accountCredential.AccountStatus = domain.AccountStatus(status)
	accountCredential.Credential.AccountID = accountCredential.AccountID
	accountCredential.Credential.PasswordChangedAt = nullTimePtr(passwordChangedAt)
	accountCredential.Credential.LockedUntil = nullTimePtr(lockedUntil)
	accountCredential.Credential.CreatedAt = createdAt.UTC()
	accountCredential.Credential.UpdatedAt = updatedAt.UTC()

	return accountCredential, nil
}

func (r *MySQLCredentialRepository) UpdatePassword(ctx context.Context, accountID string, passwordHash string, algorithm string, changedAt time.Time) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE credentials
		SET
			password_hash = ?,
			password_algo = ?,
			password_changed_at = ?,
			failed_attempts = 0,
			locked_until = NULL
		WHERE account_id = ?
	`, passwordHash, algorithm, changedAt.UTC(), accountID)
	if err != nil {
		return fmt.Errorf("update password hash: %w", err)
	}
	return ensureAffected(result, domain.ErrCredentialNotFound)
}

func (r *MySQLCredentialRepository) RehashPassword(ctx context.Context, accountID string, passwordHash string, algorithm string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE credentials
		SET
			password_hash = ?,
			password_algo = ?,
			password_changed_at = password_changed_at
		WHERE account_id = ?
	`, passwordHash, algorithm, accountID)
	if err != nil {
		return fmt.Errorf("rehash password: %w", err)
	}
	return ensureAffected(result, domain.ErrCredentialNotFound)
}

func (r *MySQLCredentialRepository) IncrementFailedAttempts(ctx context.Context, accountID string, maxAttempts int, lockUntil time.Time) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE credentials
		SET
			failed_attempts = failed_attempts + 1,
			locked_until = CASE
				WHEN failed_attempts + 1 >= ? THEN ?
				ELSE locked_until
			END
		WHERE account_id = ?
	`, maxAttempts, lockUntil.UTC(), accountID)
	if err != nil {
		return fmt.Errorf("increment failed attempts: %w", err)
	}
	return ensureAffected(result, domain.ErrCredentialNotFound)
}

func (r *MySQLCredentialRepository) ResetFailedAttempts(ctx context.Context, accountID string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE credentials
		SET
			failed_attempts = 0,
			locked_until = NULL
		WHERE account_id = ?
	`, accountID)
	if err != nil {
		return fmt.Errorf("reset failed attempts: %w", err)
	}
	return ensureAffected(result, domain.ErrCredentialNotFound)
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC()
}

func nullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time.UTC()
	return &t
}

func ensureAffected(result sql.Result, notFound error) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read rows affected: %w", err)
	}
	if affected == 0 {
		return notFound
	}
	return nil
}

func isDuplicateKey(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
