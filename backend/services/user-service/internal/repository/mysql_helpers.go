package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

const (
	mysqlDuplicateEntryNumber         = 1062
	mysqlForeignKeyConstraintNumber   = 1452
	defaultAddressTransactionIsoLevel = sql.LevelReadCommitted
)

type sqlScanner interface {
	Scan(dest ...any) error
}

type sqlExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type sqlQueryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type sqlExecutor interface {
	sqlExecer
	sqlQueryer
}

type txBeginner interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type repositoryOptions struct {
	logger *slog.Logger
}

type Option func(*repositoryOptions)

func WithLogger(logger *slog.Logger) Option {
	return func(options *repositoryOptions) {
		if logger != nil {
			options.logger = logger
		}
	}
}

func newRepositoryOptions(options []Option) repositoryOptions {
	configured := repositoryOptions{logger: slog.Default()}
	for _, option := range options {
		option(&configured)
	}
	return configured
}

func nullableCleanStringPtr(value *string) any {
	if value == nil {
		return nil
	}
	cleaned := strings.TrimSpace(*value)
	if cleaned == "" {
		return nil
	}
	return cleaned
}

func nullStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func nullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time.UTC()
	return &t
}

func nullableTimePtr(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC()
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	parts := make([]string, n)
	for i := range parts {
		parts[i] = "?"
	}
	return strings.Join(parts, ",")
}

func rowsAffected(result sql.Result) (int64, error) {
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("read rows affected: %w", err)
	}
	return affected, nil
}

func ensureAffected(result sql.Result, notFound error) error {
	affected, err := rowsAffected(result)
	if err != nil {
		return err
	}
	if affected == 0 {
		return notFound
	}
	return nil
}

func isDuplicateKey(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlDuplicateEntryNumber
}

func isForeignKeyConstraint(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlForeignKeyConstraintNumber
}

func logRepositoryError(ctx context.Context, logger *slog.Logger, operation string, err error) {
	if logger == nil || err == nil {
		return
	}
	logger.ErrorContext(ctx, "user_repository_error",
		slog.String("operation", operation),
		slog.String("error_type", fmt.Sprintf("%T", err)),
	)
}
