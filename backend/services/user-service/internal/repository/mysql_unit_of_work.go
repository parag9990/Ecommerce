package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/parag/ecommerce/backend/services/user-service/internal/events"
	"github.com/parag/ecommerce/backend/services/user-service/internal/usecase"
)

var _ usecase.UnitOfWork = (*MySQLUnitOfWork)(nil)

type MySQLUnitOfWork struct {
	db             *sql.DB
	logger         *slog.Logger
	recorderConfig events.RecorderConfig
}

func NewMySQLUnitOfWork(db *sql.DB, recorderConfig events.RecorderConfig, options ...Option) (*MySQLUnitOfWork, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	configured := newRepositoryOptions(options)
	if recorderConfig.Logger == nil {
		recorderConfig.Logger = configured.logger
	}
	return &MySQLUnitOfWork{db: db, logger: configured.logger, recorderConfig: recorderConfig}, nil
}

func (u *MySQLUnitOfWork) WithinTx(ctx context.Context, fn func(context.Context, usecase.TransactionRepositories) error) error {
	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		logRepositoryError(ctx, u.logger, "begin_user_service_unit_of_work", err)
		return fmt.Errorf("begin user service transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	options := []Option{WithLogger(u.logger)}
	outboxRepo := newMySQLOutboxRepository(tx, options...)
	recorder, err := events.NewOutboxRecorder(outboxRepo, u.recorderConfig)
	if err != nil {
		return err
	}

	repositories := usecase.TransactionRepositories{
		Users:     newMySQLUserRepository(tx, options...),
		Addresses: newMySQLAddressRepository(tx, nil, options...),
		Sellers:   newMySQLSellerRepository(tx, options...),
		Events:    recorder,
	}

	if err := fn(ctx, repositories); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		logRepositoryError(ctx, u.logger, "commit_user_service_unit_of_work", err)
		return fmt.Errorf("commit user service transaction: %w", err)
	}
	return nil
}
