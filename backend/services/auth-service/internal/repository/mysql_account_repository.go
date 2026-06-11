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
