package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/parag/ecommerce/backend/services/user-service/internal/domain"
	"github.com/parag/ecommerce/backend/services/user-service/internal/usecase"
)

var _ usecase.AddressRepository = (*MySQLAddressRepository)(nil)

type MySQLAddressRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewMySQLAddressRepository(db *sql.DB, options ...Option) (*MySQLAddressRepository, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}

	configured := newRepositoryOptions(options)
	return &MySQLAddressRepository{db: db, logger: configured.logger}, nil
}

func (r *MySQLAddressRepository) ListAddresses(ctx context.Context, userID string, limit int, offset int) ([]domain.Address, error) {
	if limit <= 0 {
		return []domain.Address{}, nil
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			address_id,
			user_id,
			name,
			phone,
			line1,
			line2,
			city,
			state,
			postal_code,
			country,
			is_default,
			created_at,
			updated_at,
			deleted_at
		FROM user_addresses
		WHERE user_id = ?
		  AND deleted_at IS NULL
		ORDER BY is_default DESC, updated_at DESC
		LIMIT ? OFFSET ?
	`, userID, limit, offset)
	if err != nil {
		logRepositoryError(ctx, r.logger, "list_addresses", err)
		return nil, fmt.Errorf("query user addresses: %w", err)
	}
	defer rows.Close()

	addresses := make([]domain.Address, 0, limit)
	for rows.Next() {
		address, err := scanAddress(rows)
		if err != nil {
			logRepositoryError(ctx, r.logger, "scan_address", err)
			return nil, fmt.Errorf("scan address: %w", err)
		}
		addresses = append(addresses, address)
	}
	if err := rows.Err(); err != nil {
		logRepositoryError(ctx, r.logger, "iterate_addresses", err)
		return nil, fmt.Errorf("iterate addresses: %w", err)
	}

	return addresses, nil
}

func (r *MySQLAddressRepository) CreateAddress(ctx context.Context, address domain.Address) (domain.Address, error) {
	if address.IsDefault {
		return r.createDefaultAddress(ctx, address)
	}

	if err := r.insertAddress(ctx, r.db, address); err != nil {
		return domain.Address{}, err
	}
	return r.findAddress(ctx, address.UserID, address.AddressID)
}

func (r *MySQLAddressRepository) UpdateAddress(ctx context.Context, address domain.Address) (domain.Address, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE user_addresses
		SET
			name = ?,
			phone = ?,
			line1 = ?,
			line2 = ?,
			city = ?,
			state = ?,
			postal_code = ?,
			country = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ?
		  AND address_id = ?
		  AND deleted_at IS NULL
	`,
		address.Name,
		nullableCleanStringPtr(address.Phone),
		address.Line1,
		nullableCleanStringPtr(address.Line2),
		address.City,
		address.State,
		address.PostalCode,
		address.Country,
		address.UserID,
		address.AddressID,
	)
	if err != nil {
		if isDuplicateKey(err) {
			return domain.Address{}, domain.ErrDuplicateAddress
		}
		logRepositoryError(ctx, r.logger, "update_address", err)
		return domain.Address{}, fmt.Errorf("update address: %w", err)
	}

	affected, err := rowsAffected(result)
	if err != nil {
		return domain.Address{}, err
	}
	if affected == 0 {
		exists, err := r.activeAddressExists(ctx, address.UserID, address.AddressID)
		if err != nil {
			return domain.Address{}, err
		}
		if !exists {
			return domain.Address{}, domain.ErrAddressNotFound
		}
	}

	return r.findAddress(ctx, address.UserID, address.AddressID)
}

func (r *MySQLAddressRepository) DeleteAddress(ctx context.Context, userID string, addressID string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE user_addresses
		SET
			deleted_at = CURRENT_TIMESTAMP,
			is_default = FALSE,
			updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ?
		  AND address_id = ?
		  AND deleted_at IS NULL
	`, userID, addressID)
	if err != nil {
		logRepositoryError(ctx, r.logger, "delete_address", err)
		return fmt.Errorf("soft delete address: %w", err)
	}
	return ensureAffected(result, domain.ErrAddressNotFound)
}

func (r *MySQLAddressRepository) SetDefaultAddress(ctx context.Context, userID string, addressID string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: defaultAddressTransactionIsoLevel})
	if err != nil {
		logRepositoryError(ctx, r.logger, "begin_set_default_address", err)
		return fmt.Errorf("begin set default address: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var lockedAddressID string
	err = tx.QueryRowContext(ctx, `
		SELECT address_id
		FROM user_addresses
		WHERE user_id = ?
		  AND address_id = ?
		  AND deleted_at IS NULL
		LIMIT 1
		FOR UPDATE
	`, userID, addressID).Scan(&lockedAddressID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrAddressNotFound
		}
		logRepositoryError(ctx, r.logger, "lock_default_address", err)
		return fmt.Errorf("lock address: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE user_addresses
		SET is_default = FALSE,
		    updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ?
		  AND deleted_at IS NULL
	`, userID); err != nil {
		logRepositoryError(ctx, r.logger, "clear_default_addresses", err)
		return fmt.Errorf("clear default addresses: %w", err)
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE user_addresses
		SET is_default = TRUE,
		    updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ?
		  AND address_id = ?
		  AND deleted_at IS NULL
	`, userID, addressID)
	if err != nil {
		logRepositoryError(ctx, r.logger, "set_default_address", err)
		return fmt.Errorf("set default address: %w", err)
	}
	if err := ensureAffected(result, domain.ErrAddressNotFound); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		logRepositoryError(ctx, r.logger, "commit_set_default_address", err)
		return fmt.Errorf("commit set default address: %w", err)
	}
	return nil
}

func (r *MySQLAddressRepository) createDefaultAddress(ctx context.Context, address domain.Address) (domain.Address, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: defaultAddressTransactionIsoLevel})
	if err != nil {
		logRepositoryError(ctx, r.logger, "begin_create_default_address", err)
		return domain.Address{}, fmt.Errorf("begin create default address: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, `
		UPDATE user_addresses
		SET is_default = FALSE,
		    updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ?
		  AND deleted_at IS NULL
	`, address.UserID); err != nil {
		logRepositoryError(ctx, r.logger, "clear_default_addresses_for_create", err)
		return domain.Address{}, fmt.Errorf("clear default addresses: %w", err)
	}

	if err := r.insertAddress(ctx, tx, address); err != nil {
		return domain.Address{}, err
	}

	if err := tx.Commit(); err != nil {
		logRepositoryError(ctx, r.logger, "commit_create_default_address", err)
		return domain.Address{}, fmt.Errorf("commit create default address: %w", err)
	}

	return r.findAddress(ctx, address.UserID, address.AddressID)
}

func (r *MySQLAddressRepository) insertAddress(ctx context.Context, execer sqlExecer, address domain.Address) error {
	_, err := execer.ExecContext(ctx, `
		INSERT INTO user_addresses (
			address_id,
			user_id,
			name,
			phone,
			line1,
			line2,
			city,
			state,
			postal_code,
			country,
			is_default
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		address.AddressID,
		address.UserID,
		address.Name,
		nullableCleanStringPtr(address.Phone),
		address.Line1,
		nullableCleanStringPtr(address.Line2),
		address.City,
		address.State,
		address.PostalCode,
		address.Country,
		address.IsDefault,
	)
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrDuplicateAddress
		}
		if isForeignKeyConstraint(err) {
			return domain.ErrUserNotFound
		}
		logRepositoryError(ctx, r.logger, "insert_address", err)
		return fmt.Errorf("insert address: %w", err)
	}
	return nil
}

func (r *MySQLAddressRepository) findAddress(ctx context.Context, userID string, addressID string) (domain.Address, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			address_id,
			user_id,
			name,
			phone,
			line1,
			line2,
			city,
			state,
			postal_code,
			country,
			is_default,
			created_at,
			updated_at,
			deleted_at
		FROM user_addresses
		WHERE user_id = ?
		  AND address_id = ?
		  AND deleted_at IS NULL
		LIMIT 1
	`, userID, addressID)

	address, err := scanAddress(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Address{}, domain.ErrAddressNotFound
		}
		logRepositoryError(ctx, r.logger, "find_address", err)
		return domain.Address{}, fmt.Errorf("query address: %w", err)
	}
	return address, nil
}

func (r *MySQLAddressRepository) activeAddressExists(ctx context.Context, userID string, addressID string) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx, `
		SELECT 1
		FROM user_addresses
		WHERE user_id = ?
		  AND address_id = ?
		  AND deleted_at IS NULL
		LIMIT 1
	`, userID, addressID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		logRepositoryError(ctx, r.logger, "active_address_exists", err)
		return false, fmt.Errorf("query active address existence: %w", err)
	}
	return true, nil
}

func scanAddress(row sqlScanner) (domain.Address, error) {
	var (
		address   domain.Address
		phone     sql.NullString
		line2     sql.NullString
		deletedAt sql.NullTime
	)

	err := row.Scan(
		&address.AddressID,
		&address.UserID,
		&address.Name,
		&phone,
		&address.Line1,
		&line2,
		&address.City,
		&address.State,
		&address.PostalCode,
		&address.Country,
		&address.IsDefault,
		&address.CreatedAt,
		&address.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		return domain.Address{}, err
	}

	address.Phone = nullStringPtr(phone)
	address.Line2 = nullStringPtr(line2)
	address.CreatedAt = address.CreatedAt.UTC()
	address.UpdatedAt = address.UpdatedAt.UTC()
	address.DeletedAt = nullTimePtr(deletedAt)

	return address, nil
}
