package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
)

type MySQLSellerStaffRepository struct {
	db *sql.DB
}

func NewMySQLSellerStaffRepository(db *sql.DB) (*MySQLSellerStaffRepository, error) {
	if db == nil {
		return nil, errors.New("mysql db is required")
	}
	return &MySQLSellerStaffRepository{db: db}, nil
}

func (r *MySQLSellerStaffRepository) GetBySellerAndUser(ctx context.Context, sellerID string, userID string) (domain.SellerStaff, error) {
	const query = `
SELECT staff_id, seller_id, user_id, role, status, COALESCE(invited_by, ''), created_at, updated_at
FROM seller_staff
WHERE seller_id = ? AND user_id = ?
LIMIT 1`

	var staff domain.SellerStaff
	var role string
	var status string
	if err := r.db.QueryRowContext(ctx, query, strings.TrimSpace(sellerID), strings.TrimSpace(userID)).Scan(
		&staff.StaffID,
		&staff.SellerID,
		&staff.UserID,
		&role,
		&status,
		&staff.InvitedBy,
		&staff.CreatedAt,
		&staff.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.SellerStaff{}, domain.ErrSellerStaffNotFound
		}
		return domain.SellerStaff{}, err
	}
	staff.Role = domain.Role(strings.TrimSpace(role))
	staff.Status = domain.StaffStatus(status).Normalized()
	return staff, nil
}
