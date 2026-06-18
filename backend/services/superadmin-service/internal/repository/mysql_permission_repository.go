package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"ecommerce/superadmin-service/internal/domain"
)

type MySQLPermissionRepository struct {
	db *sql.DB
}

func NewMySQLPermissionRepository(db *sql.DB) (*MySQLPermissionRepository, error) {
	if db == nil {
		return nil, errors.New("mysql permission repository requires db")
	}
	return &MySQLPermissionRepository{db: db}, nil
}

func (r *MySQLPermissionRepository) AdminHasPermission(ctx context.Context, adminID string, permission domain.Permission) (bool, error) {
	if strings.TrimSpace(adminID) == "" {
		return false, domain.NewValidationError("admin_id is required")
	}
	if !permission.Valid() {
		return false, domain.NewValidationError(fmt.Sprintf("unknown permission %q", permission))
	}

	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM admin_users au
			JOIN admin_role_permissions arp
			  ON arp.role = au.role
			JOIN admin_permissions ap
			  ON ap.permission_key = arp.permission_key
			WHERE au.admin_id = ?
			  AND au.status = 'active'
			  AND arp.permission_key = ?
		) AS allowed`

	var allowed int
	if err := r.db.QueryRowContext(ctx, query, adminID, string(permission)).Scan(&allowed); err != nil {
		return false, fmt.Errorf("check admin permission: %w", err)
	}
	return allowed == 1, nil
}

func (r *MySQLPermissionRepository) ListPermissionsByRole(ctx context.Context, role domain.AdminRole) ([]domain.Permission, error) {
	if role == "" {
		return nil, domain.NewValidationError("role is required")
	}

	const query = `
		SELECT arp.permission_key
		FROM admin_role_permissions arp
		JOIN admin_permissions ap
		  ON ap.permission_key = arp.permission_key
		WHERE arp.role = ?
		ORDER BY arp.permission_key`

	rows, err := r.db.QueryContext(ctx, query, string(role))
	if err != nil {
		return nil, fmt.Errorf("list role permissions: %w", err)
	}
	defer rows.Close()

	var permissions []domain.Permission
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, fmt.Errorf("scan role permission: %w", err)
		}
		permission := domain.Permission(raw)
		if !permission.Valid() {
			return nil, domain.NewInternal(fmt.Sprintf("database contains unknown permission %q", raw), nil)
		}
		permissions = append(permissions, permission)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate role permissions: %w", err)
	}

	return permissions, nil
}

func (r *MySQLPermissionRepository) ListPermissionDefinitions(ctx context.Context) ([]domain.PermissionDefinition, error) {
	const query = `
		SELECT permission_key, description
		FROM admin_permissions
		ORDER BY permission_key`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list permission definitions: %w", err)
	}
	defer rows.Close()

	var definitions []domain.PermissionDefinition
	for rows.Next() {
		var key string
		var description sql.NullString
		if err := rows.Scan(&key, &description); err != nil {
			return nil, fmt.Errorf("scan permission definition: %w", err)
		}

		permission := domain.Permission(key)
		if !permission.Valid() {
			return nil, domain.NewInternal(fmt.Sprintf("database contains unknown permission %q", key), nil)
		}
		definition, _ := domain.PermissionByKey(permission)
		if description.Valid {
			definition.Description = description.String
		}
		definitions = append(definitions, definition)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate permission definitions: %w", err)
	}

	return definitions, nil
}
