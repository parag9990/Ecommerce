package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/auth-service/internal/domain"
)

const maxOutboxLastErrorLength = 1024

type MySQLOutboxRepository struct {
	db *sql.DB
}

func NewMySQLOutboxRepository(db *sql.DB) (*MySQLOutboxRepository, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	return &MySQLOutboxRepository{db: db}, nil
}

func (r *MySQLOutboxRepository) InsertOutboxEvent(ctx context.Context, event domain.OutboxEvent) error {
	if event.Status == "" {
		event.Status = domain.OutboxStatusPending
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO auth_outbox_events (
			event_id,
			event_type,
			version,
			aggregate_type,
			aggregate_id,
			routing_key,
			payload,
			trace_id,
			status,
			next_attempt_at,
			created_at,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, event.EventID, event.EventType, event.Version, event.AggregateType, event.AggregateID, event.RoutingKey, string(event.Payload), nullableString(event.TraceID), event.Status, nullableTime(event.NextAttemptAt), event.CreatedAt.UTC(), event.UpdatedAt.UTC())
	if err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}
	return nil
}

func (r *MySQLOutboxRepository) LockPendingOutboxEvents(ctx context.Context, limit int, now time.Time, staleBefore time.Time, workerID string) ([]domain.OutboxEvent, error) {
	if limit <= 0 {
		return nil, nil
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("begin outbox lock: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	rows, err := tx.QueryContext(ctx, `
		SELECT
			event_id,
			event_type,
			version,
			aggregate_type,
			aggregate_id,
			routing_key,
			payload,
			COALESCE(trace_id, '') AS trace_id,
			status,
			attempts,
			next_attempt_at,
			published_at,
			locked_at,
			COALESCE(locked_by, '') AS locked_by,
			COALESCE(last_error, '') AS last_error,
			created_at,
			updated_at
		FROM auth_outbox_events
		WHERE (
				status IN ('pending', 'failed')
				AND (next_attempt_at IS NULL OR next_attempt_at <= ?)
			)
			OR (
				status = 'processing'
				AND locked_at IS NOT NULL
				AND locked_at <= ?
			)
		ORDER BY created_at ASC
		LIMIT ?
		FOR UPDATE SKIP LOCKED
	`, now.UTC(), staleBefore.UTC(), limit)
	if err != nil {
		return nil, fmt.Errorf("query pending outbox events: %w", err)
	}
	defer rows.Close()

	events := make([]domain.OutboxEvent, 0, limit)
	for rows.Next() {
		event, err := scanOutboxEvent(rows)
		if err != nil {
			return nil, fmt.Errorf("scan outbox event: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate outbox events: %w", err)
	}

	for _, event := range events {
		result, err := tx.ExecContext(ctx, `
			UPDATE auth_outbox_events
			SET
				status = 'processing',
				locked_at = ?,
				locked_by = ?,
				updated_at = ?
			WHERE event_id = ?
		`, now.UTC(), nullableString(workerID), now.UTC(), event.EventID)
		if err != nil {
			return nil, fmt.Errorf("mark outbox event processing: %w", err)
		}
		if err := ensureAffected(result, sql.ErrNoRows); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit outbox lock: %w", err)
	}
	return events, nil
}

func (r *MySQLOutboxRepository) MarkOutboxEventPublished(ctx context.Context, eventID string, publishedAt time.Time) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE auth_outbox_events
		SET
			status = 'published',
			published_at = ?,
			locked_at = NULL,
			locked_by = NULL,
			last_error = NULL,
			updated_at = ?
		WHERE event_id = ?
	`, publishedAt.UTC(), publishedAt.UTC(), eventID)
	if err != nil {
		return fmt.Errorf("mark outbox event published: %w", err)
	}
	return ensureAffected(result, sql.ErrNoRows)
}

func (r *MySQLOutboxRepository) MarkOutboxEventFailed(ctx context.Context, eventID string, nextAttemptAt time.Time, lastError string) error {
	now := time.Now().UTC()
	result, err := r.db.ExecContext(ctx, `
		UPDATE auth_outbox_events
		SET
			status = 'failed',
			attempts = attempts + 1,
			next_attempt_at = ?,
			locked_at = NULL,
			locked_by = NULL,
			last_error = ?,
			updated_at = ?
		WHERE event_id = ?
	`, nextAttemptAt.UTC(), truncateOutboxError(lastError), now, eventID)
	if err != nil {
		return fmt.Errorf("mark outbox event failed: %w", err)
	}
	return ensureAffected(result, sql.ErrNoRows)
}

func (r *MySQLOutboxRepository) MarkOutboxEventDeadLetter(ctx context.Context, eventID string, lastError string) error {
	now := time.Now().UTC()
	result, err := r.db.ExecContext(ctx, `
		UPDATE auth_outbox_events
		SET
			status = 'dead_letter',
			attempts = attempts + 1,
			locked_at = NULL,
			locked_by = NULL,
			last_error = ?,
			updated_at = ?
		WHERE event_id = ?
	`, truncateOutboxError(lastError), now, eventID)
	if err != nil {
		return fmt.Errorf("mark outbox event dead-letter: %w", err)
	}
	return ensureAffected(result, sql.ErrNoRows)
}

func scanOutboxEvent(row sqlScanner) (domain.OutboxEvent, error) {
	var (
		event         domain.OutboxEvent
		status        string
		nextAttemptAt sql.NullTime
		publishedAt   sql.NullTime
		lockedAt      sql.NullTime
	)
	if err := row.Scan(
		&event.EventID,
		&event.EventType,
		&event.Version,
		&event.AggregateType,
		&event.AggregateID,
		&event.RoutingKey,
		&event.Payload,
		&event.TraceID,
		&status,
		&event.Attempts,
		&nextAttemptAt,
		&publishedAt,
		&lockedAt,
		&event.LockedBy,
		&event.LastError,
		&event.CreatedAt,
		&event.UpdatedAt,
	); err != nil {
		return domain.OutboxEvent{}, err
	}
	event.Status = domain.OutboxStatus(status)
	event.NextAttemptAt = nullTimePtr(nextAttemptAt)
	event.PublishedAt = nullTimePtr(publishedAt)
	event.LockedAt = nullTimePtr(lockedAt)
	event.CreatedAt = event.CreatedAt.UTC()
	event.UpdatedAt = event.UpdatedAt.UTC()
	return event, nil
}

func truncateOutboxError(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= maxOutboxLastErrorLength {
		return value
	}
	return value[:maxOutboxLastErrorLength]
}
