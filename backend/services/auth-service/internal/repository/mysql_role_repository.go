package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

type MySQLRoleRepository struct {
	db *sql.DB
}

func NewMySQLRoleRepository(db *sql.DB) (*MySQLRoleRepository, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	return &MySQLRoleRepository{db: db}, nil
}

func (r *MySQLRoleRepository) ActiveRoles(ctx context.Context, accountID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT role
		FROM role_assignments
		WHERE account_id = ? AND revoked_at IS NULL
		ORDER BY role
	`, accountID)
	if err != nil {
		return nil, fmt.Errorf("query active roles: %w", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate roles: %w", err)
	}
	return roles, nil
}

func (r *MySQLRoleRepository) ResolveAccountID(ctx context.Context, userOrAccountID string) (string, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT account_id
		FROM auth_accounts
		WHERE account_id = ? OR user_id = ?
		LIMIT 1
	`, userOrAccountID, userOrAccountID)

	var accountID string
	if err := row.Scan(&accountID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrAccountNotFound
		}
		return "", fmt.Errorf("resolve account id: %w", err)
	}
	return accountID, nil
}

func (r *MySQLRoleRepository) AssignRole(ctx context.Context, assignment domain.RoleAssignment) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("begin role assignment: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := lockAccount(ctx, tx, assignment.AccountID); err != nil {
		return err
	}

	var existingID uint64
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM role_assignments
		WHERE account_id = ?
		  AND role = ?
		  AND COALESCE(scope_type, '') = COALESCE(?, '')
		  AND COALESCE(scope_id, '') = COALESCE(?, '')
		  AND revoked_at IS NULL
		LIMIT 1
		FOR UPDATE
	`, assignment.AccountID, assignment.Role.String(), nullableString(assignment.ScopeType), nullableString(assignment.ScopeID)).Scan(&existingID)
	if err == nil {
		return domain.ErrRoleAlreadyAssigned
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check active role assignment: %w", err)
	}

	assignedAt := assignment.AssignedAt
	if assignedAt.IsZero() {
		assignedAt = time.Now().UTC()
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO role_assignments (
			account_id,
			role,
			scope_type,
			scope_id,
			assigned_by,
			reason,
			assigned_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`, assignment.AccountID, assignment.Role.String(), nullableString(assignment.ScopeType), nullableString(assignment.ScopeID), nullableString(assignment.AssignedBy), nullableString(assignment.Reason), assignedAt.UTC())
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrRoleAlreadyAssigned
		}
		return fmt.Errorf("insert role assignment: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit role assignment: %w", err)
	}
	return nil
}

func (r *MySQLRoleRepository) RevokeRole(ctx context.Context, accountID string, role domain.Role, scopeType string, scopeID string, revokedAt time.Time) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("begin role revoke: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := lockAccount(ctx, tx, accountID); err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE role_assignments
		SET revoked_at = ?
		WHERE account_id = ?
		  AND role = ?
		  AND COALESCE(scope_type, '') = COALESCE(?, '')
		  AND COALESCE(scope_id, '') = COALESCE(?, '')
		  AND revoked_at IS NULL
	`, revokedAt.UTC(), accountID, role.String(), nullableString(scopeType), nullableString(scopeID))
	if err != nil {
		return fmt.Errorf("revoke role assignment: %w", err)
	}
	if err := ensureAffected(result, domain.ErrRoleAssignmentNotFound); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit role revoke: %w", err)
	}
	return nil
}

func (r *MySQLRoleRepository) GetRoleAssignments(ctx context.Context, accountID string, includeRevoked bool) ([]domain.RoleAssignment, error) {
	query := `
		SELECT
			account_id,
			role,
			COALESCE(scope_type, '') AS scope_type,
			COALESCE(scope_id, '') AS scope_id,
			COALESCE(assigned_by, '') AS assigned_by,
			COALESCE(reason, '') AS reason,
			assigned_at,
			revoked_at
		FROM role_assignments
		WHERE account_id = ?
	`
	if !includeRevoked {
		query += ` AND revoked_at IS NULL`
	}
	query += ` ORDER BY role, scope_type, scope_id, assigned_at`

	rows, err := r.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("query role assignments: %w", err)
	}
	defer rows.Close()

	var assignments []domain.RoleAssignment
	for rows.Next() {
		assignment, err := scanRoleAssignment(rows)
		if err != nil {
			return nil, fmt.Errorf("scan role assignment: %w", err)
		}
		assignments = append(assignments, assignment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate role assignments: %w", err)
	}
	return assignments, nil
}

func lockAccount(ctx context.Context, tx *sql.Tx, accountID string) error {
	var locked string
	if err := tx.QueryRowContext(ctx, `
		SELECT account_id
		FROM auth_accounts
		WHERE account_id = ?
		LIMIT 1
		FOR UPDATE
	`, accountID).Scan(&locked); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrAccountNotFound
		}
		return fmt.Errorf("lock account: %w", err)
	}
	return nil
}

func scanRoleAssignment(row sqlScanner) (domain.RoleAssignment, error) {
	var (
		assignment domain.RoleAssignment
		role       string
		revokedAt  sql.NullTime
	)
	if err := row.Scan(
		&assignment.AccountID,
		&role,
		&assignment.ScopeType,
		&assignment.ScopeID,
		&assignment.AssignedBy,
		&assignment.Reason,
		&assignment.AssignedAt,
		&revokedAt,
	); err != nil {
		return domain.RoleAssignment{}, err
	}
	assignment.Role = domain.Role(role)
	assignment.AssignedAt = assignment.AssignedAt.UTC()
	assignment.RevokedAt = nullTimePtr(revokedAt)
	return assignment, nil
}
