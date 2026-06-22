package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/parag/ecommerce/backend/services/user-service/internal/domain"
)

type AdminUserPage struct {
	Users []domain.User
	Total int64
}

type AdminSellerPage struct {
	Sellers      []domain.SellerProfile
	KYCDocuments map[string][]domain.KYCDocument
	Total        int64
}

func (r *MySQLUserRepository) ListUsersForAdmin(ctx context.Context, query, status string, limit, offset int) (AdminUserPage, error) {
	where, args := adminUserWhere(query, status)
	var total int64
	if err := r.executor.QueryRowContext(ctx, "SELECT COUNT(*) FROM users "+where, args...).Scan(&total); err != nil {
		return AdminUserPage{}, fmt.Errorf("count admin users: %w", err)
	}
	rows, err := r.executor.QueryContext(ctx, `
		SELECT user_id, auth_account_id, email, phone, full_name, avatar_url, status,
		       created_by, updated_by, status_changed_by, status_changed_at,
		       deleted_by, deleted_at, created_at, updated_at
		FROM users `+where+`
		ORDER BY created_at DESC, user_id DESC
		LIMIT ? OFFSET ?`, append(args, limit, offset)...)
	if err != nil {
		return AdminUserPage{}, fmt.Errorf("list admin users: %w", err)
	}
	defer rows.Close()
	users := make([]domain.User, 0, limit)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return AdminUserPage{}, fmt.Errorf("scan admin user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return AdminUserPage{}, fmt.Errorf("iterate admin users: %w", err)
	}
	return AdminUserPage{Users: users, Total: total}, nil
}

func (r *MySQLSellerRepository) ListSellersForAdmin(ctx context.Context, query, status string, limit, offset int) (AdminSellerPage, error) {
	where, args := adminSellerWhere(query, status)
	var total int64
	if err := r.executor.QueryRowContext(ctx, "SELECT COUNT(*) FROM seller_profiles "+where, args...).Scan(&total); err != nil {
		return AdminSellerPage{}, fmt.Errorf("count admin sellers: %w", err)
	}
	rows, err := r.executor.QueryContext(ctx, `
		SELECT seller_id, user_id, store_name, display_name, gst_number, support_email,
		       status, approved_by, approved_at, created_by, updated_by,
		       status_changed_by, status_changed_at, status_reason, created_at, updated_at
		FROM seller_profiles `+where+`
		ORDER BY created_at DESC, seller_id DESC
		LIMIT ? OFFSET ?`, append(args, limit, offset)...)
	if err != nil {
		return AdminSellerPage{}, fmt.Errorf("list admin sellers: %w", err)
	}
	defer rows.Close()
	sellers := make([]domain.SellerProfile, 0, limit)
	for rows.Next() {
		seller, err := scanSellerProfile(rows)
		if err != nil {
			return AdminSellerPage{}, fmt.Errorf("scan admin seller: %w", err)
		}
		sellers = append(sellers, seller)
	}
	if err := rows.Err(); err != nil {
		return AdminSellerPage{}, fmt.Errorf("iterate admin sellers: %w", err)
	}
	documents := make(map[string][]domain.KYCDocument, len(sellers))
	for _, seller := range sellers {
		docs, err := r.ListKYCDocuments(ctx, seller.SellerID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return AdminSellerPage{}, fmt.Errorf("list KYC documents for %s: %w", seller.SellerID, err)
		}
		documents[seller.SellerID] = docs
	}
	return AdminSellerPage{Sellers: sellers, KYCDocuments: documents, Total: total}, nil
}

func adminUserWhere(query, status string) (string, []any) {
	conditions := []string{"deleted_at IS NULL"}
	args := make([]any, 0, 5)
	if status = strings.TrimSpace(status); status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, status)
	}
	if query = strings.TrimSpace(query); query != "" {
		like := "%" + query + "%"
		conditions = append(conditions, "(user_id = ? OR email LIKE ? OR full_name LIKE ? OR phone LIKE ?)")
		args = append(args, query, like, like, like)
	}
	return "WHERE " + strings.Join(conditions, " AND "), args
}

func adminSellerWhere(query, status string) (string, []any) {
	conditions := []string{"1 = 1"}
	args := make([]any, 0, 6)
	if status = strings.TrimSpace(status); status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, status)
	}
	if query = strings.TrimSpace(query); query != "" {
		like := "%" + query + "%"
		conditions = append(conditions, "(seller_id = ? OR user_id = ? OR store_name LIKE ? OR display_name LIKE ? OR support_email LIKE ?)")
		args = append(args, query, query, like, like, like)
	}
	return "WHERE " + strings.Join(conditions, " AND "), args
}
