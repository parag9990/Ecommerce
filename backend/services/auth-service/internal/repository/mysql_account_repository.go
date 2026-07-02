package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

type MySQLAccountRepository struct {
	db *sql.DB
}

func NewMySQLAccountRepository(db *sql.DB) (*MySQLAccountRepository, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	return &MySQLAccountRepository{db: db}, nil
}

func (r *MySQLAccountRepository) CreateAccount(ctx context.Context, account domain.AuthAccount) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO auth_accounts (
			account_id,
			user_id,
			email,
			phone,
			seller_id,
			tenant_id,
			email_verified,
			phone_verified,
			status,
			created_at,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, account.AccountID, nullableString(account.UserID), nullableStringPtr(account.Email), nullableStringPtr(account.Phone), nullableString(account.SellerID), nullableString(account.TenantID), account.EmailVerified, account.PhoneVerified, account.Status, account.CreatedAt.UTC(), account.UpdatedAt.UTC())
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrDuplicateAccount
		}
		return fmt.Errorf("insert auth account: %w", err)
	}
	return nil
}

func (r *MySQLAccountRepository) LinkUser(ctx context.Context, accountID string, userID string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE auth_accounts
		SET user_id = ?
		WHERE account_id = ?
	`, userID, accountID)
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrDuplicateAccount
		}
		return fmt.Errorf("link auth account user: %w", err)
	}
	return ensureAffected(result, domain.ErrAccountNotFound)
}

func (r *MySQLAccountRepository) FindTokenSubject(ctx context.Context, accountID string) (domain.TokenSubject, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			account_id,
			COALESCE(NULLIF(user_id, ''), account_id) AS user_id,
			COALESCE(seller_id, '') AS seller_id,
			COALESCE(tenant_id, '') AS tenant_id,
			status
		FROM auth_accounts
		WHERE account_id = ?
		LIMIT 1
	`, accountID)

	var subject domain.TokenSubject
	var status string
	if err := row.Scan(&subject.AccountID, &subject.UserID, &subject.SellerID, &subject.TenantID, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.TokenSubject{}, domain.ErrTokenSubjectMissing
		}
		return domain.TokenSubject{}, fmt.Errorf("query token subject: %w", err)
	}
	subject.Status = domain.AccountStatus(status)
	return subject, nil
}
