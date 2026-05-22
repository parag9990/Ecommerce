package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/parag/ecommerce/backend/services/user-service/internal/domain"
	"github.com/parag/ecommerce/backend/services/user-service/internal/usecase"
)

var _ usecase.UserRepository = (*MySQLUserRepository)(nil)

type MySQLUserRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewMySQLUserRepository(db *sql.DB, options ...Option) (*MySQLUserRepository, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}

	configured := newRepositoryOptions(options)
	return &MySQLUserRepository{db: db, logger: configured.logger}, nil
}

func (r *MySQLUserRepository) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	status := user.Status
	if status == "" {
		status = domain.UserStatusActive
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (
			user_id,
			auth_account_id,
			email,
			phone,
			full_name,
			avatar_url,
			status
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		user.UserID,
		user.AuthAccountID,
		user.Email,
		nullableCleanStringPtr(user.Phone),
		user.FullName,
		nullableCleanStringPtr(user.AvatarURL),
		status,
	)
	if err != nil {
		if isDuplicateKey(err) {
			return domain.User{}, domain.ErrDuplicateUser
		}
		logRepositoryError(ctx, r.logger, "create_user", err)
		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}

	return r.FindUserByID(ctx, user.UserID)
}

func (r *MySQLUserRepository) FindUserByID(ctx context.Context, userID string) (domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			user_id,
			auth_account_id,
			email,
			phone,
			full_name,
			avatar_url,
			status,
			created_at,
			updated_at
		FROM users
		WHERE user_id = ?
		LIMIT 1
	`, userID)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		logRepositoryError(ctx, r.logger, "find_user_by_id", err)
		return domain.User{}, fmt.Errorf("query user by id: %w", err)
	}
	return user, nil
}

func (r *MySQLUserRepository) FindUserByAuthAccountID(ctx context.Context, authAccountID string) (domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			user_id,
			auth_account_id,
			email,
			phone,
			full_name,
			avatar_url,
			status,
			created_at,
			updated_at
		FROM users
		WHERE auth_account_id = ?
		LIMIT 1
	`, authAccountID)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		logRepositoryError(ctx, r.logger, "find_user_by_auth_account_id", err)
		return domain.User{}, fmt.Errorf("query user by auth account id: %w", err)
	}
	return user, nil
}

func (r *MySQLUserRepository) BatchFindUsers(ctx context.Context, userIDs []string) ([]domain.User, error) {
	if len(userIDs) == 0 {
		return []domain.User{}, nil
	}

	args := make([]any, 0, len(userIDs))
	for _, id := range userIDs {
		args = append(args, id)
	}

	query := `
		SELECT
			user_id,
			auth_account_id,
			email,
			phone,
			full_name,
			avatar_url,
			status,
			created_at,
			updated_at
		FROM users
		WHERE user_id IN (` + placeholders(len(userIDs)) + `)
		ORDER BY user_id
	`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		logRepositoryError(ctx, r.logger, "batch_find_users", err)
		return nil, fmt.Errorf("query batch users: %w", err)
	}
	defer rows.Close()

	users := make([]domain.User, 0, len(userIDs))
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			logRepositoryError(ctx, r.logger, "scan_batch_user", err)
			return nil, fmt.Errorf("scan batch user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		logRepositoryError(ctx, r.logger, "iterate_batch_users", err)
		return nil, fmt.Errorf("iterate batch users: %w", err)
	}

	return users, nil
}

func (r *MySQLUserRepository) UpdateUserProfile(ctx context.Context, userID string, patch domain.UserProfilePatch) (domain.User, error) {
	sets := make([]string, 0, 4)
	args := make([]any, 0, 5)

	if patch.FullName != nil {
		sets = append(sets, "full_name = ?")
		args = append(args, *patch.FullName)
	}
	if patch.Phone != nil {
		sets = append(sets, "phone = ?")
		args = append(args, nullableCleanStringPtr(patch.Phone))
	}
	if patch.AvatarURL != nil {
		sets = append(sets, "avatar_url = ?")
		args = append(args, nullableCleanStringPtr(patch.AvatarURL))
	}

	if len(sets) == 0 {
		return r.FindUserByID(ctx, userID)
	}

	sets = append(sets, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, userID)

	result, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET `+strings.Join(sets, ", ")+`
		WHERE user_id = ?
		  AND status <> 'deleted'
	`, args...)
	if err != nil {
		logRepositoryError(ctx, r.logger, "update_user_profile", err)
		return domain.User{}, fmt.Errorf("update user profile: %w", err)
	}

	affected, err := rowsAffected(result)
	if err != nil {
		return domain.User{}, err
	}
	if affected == 0 {
		exists, err := r.activeUserExists(ctx, userID)
		if err != nil {
			return domain.User{}, err
		}
		if !exists {
			return domain.User{}, domain.ErrUserNotFound
		}
	}

	return r.FindUserByID(ctx, userID)
}

func (r *MySQLUserRepository) activeUserExists(ctx context.Context, userID string) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx, `
		SELECT 1
		FROM users
		WHERE user_id = ?
		  AND status <> 'deleted'
		LIMIT 1
	`, userID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		logRepositoryError(ctx, r.logger, "active_user_exists", err)
		return false, fmt.Errorf("query active user existence: %w", err)
	}
	return true, nil
}

func scanUser(row sqlScanner) (domain.User, error) {
	var (
		user      domain.User
		phone     sql.NullString
		avatarURL sql.NullString
		status    string
	)

	err := row.Scan(
		&user.UserID,
		&user.AuthAccountID,
		&user.Email,
		&phone,
		&user.FullName,
		&avatarURL,
		&status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return domain.User{}, err
	}

	user.Phone = nullStringPtr(phone)
	user.AvatarURL = nullStringPtr(avatarURL)
	user.Status = domain.UserStatus(status)
	user.CreatedAt = user.CreatedAt.UTC()
	user.UpdatedAt = user.UpdatedAt.UTC()

	return user, nil
}
