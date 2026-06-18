package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"ecommerce/superadmin-service/internal/domain"

	"github.com/go-sql-driver/mysql"
)

type MySQLReviewTaskRepository struct {
	db *sql.DB
}

func NewMySQLReviewTaskRepository(db *sql.DB) (*MySQLReviewTaskRepository, error) {
	if db == nil {
		return nil, errors.New("mysql review task repository requires db")
	}
	return &MySQLReviewTaskRepository{db: db}, nil
}

func (r *MySQLReviewTaskRepository) Create(ctx context.Context, task domain.ReviewTask) (*domain.ReviewTask, error) {
	if strings.TrimSpace(task.TaskID) == "" {
		return nil, domain.NewValidationError("task_id is required")
	}
	if strings.TrimSpace(task.TaskType) == "" {
		return nil, domain.NewValidationError("task_type is required")
	}
	if strings.TrimSpace(task.ResourceType) == "" {
		return nil, domain.NewValidationError("resource_type is required")
	}
	if strings.TrimSpace(task.ResourceID) == "" {
		return nil, domain.NewValidationError("resource_id is required")
	}
	if task.Status == "" {
		task.Status = domain.ReviewTaskStatusOpen
	}
	if task.Status != domain.ReviewTaskStatusOpen {
		return nil, domain.NewValidationError("new review tasks must be open")
	}

	metadata, err := json.Marshal(task.Metadata)
	if err != nil {
		return nil, domain.NewValidationError("metadata must be JSON serializable")
	}
	if task.Metadata == nil {
		metadata = nil
	}

	const query = `
		INSERT INTO admin_review_tasks (
		  task_id,
		  task_type,
		  resource_type,
		  resource_id,
		  status,
		  assigned_to,
		  created_by,
		  reason,
		  metadata
		) VALUES (?, ?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, ?)`

	if _, err := r.db.ExecContext(
		ctx,
		query,
		task.TaskID,
		task.TaskType,
		task.ResourceType,
		task.ResourceID,
		string(task.Status),
		task.AssignedTo,
		task.CreatedBy,
		task.Reason,
		nullableJSON(metadata),
	); err != nil {
		if isDuplicateKey(err) {
			return nil, domain.NewReviewTaskAlreadyExists(task.ResourceType, task.ResourceID)
		}
		return nil, fmt.Errorf("create admin review task: %w", err)
	}

	created, err := r.getByTaskID(ctx, task.TaskID)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (r *MySQLReviewTaskRepository) ListByResource(ctx context.Context, resourceType string, resourceID string) ([]domain.ReviewTask, error) {
	resourceType = strings.TrimSpace(resourceType)
	resourceID = strings.TrimSpace(resourceID)
	if resourceType == "" {
		return nil, domain.NewValidationError("resource_type is required")
	}
	if resourceID == "" {
		return nil, domain.NewValidationError("resource_id is required")
	}

	const query = `
		SELECT task_id,
		       task_type,
		       resource_type,
		       resource_id,
		       status,
		       assigned_to,
		       created_by,
		       reviewed_by,
		       reviewed_at,
		       reason,
		       metadata,
		       created_at,
		       updated_at
		FROM admin_review_tasks
		WHERE resource_type = ?
		  AND resource_id = ?
		ORDER BY created_at DESC, id DESC`

	rows, err := r.db.QueryContext(ctx, query, resourceType, resourceID)
	if err != nil {
		return nil, fmt.Errorf("list admin review tasks by resource: %w", err)
	}
	defer rows.Close()

	tasks := make([]domain.ReviewTask, 0)
	for rows.Next() {
		task, err := scanReviewTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate admin review tasks: %w", err)
	}
	return tasks, nil
}

func (r *MySQLReviewTaskRepository) CloseForResource(ctx context.Context, req domain.CloseReviewTaskRequest) error {
	if strings.TrimSpace(req.TaskType) == "" {
		return domain.NewValidationError("task_type is required")
	}
	if strings.TrimSpace(req.ResourceType) == "" {
		return domain.NewValidationError("resource_type is required")
	}
	if strings.TrimSpace(req.ResourceID) == "" {
		return domain.NewValidationError("resource_id is required")
	}
	if strings.TrimSpace(string(req.Status)) == "" {
		return domain.NewValidationError("review task status is required")
	}
	if strings.TrimSpace(req.ReviewedBy) == "" {
		return domain.NewValidationError("reviewed_by is required")
	}

	const query = `
		UPDATE admin_review_tasks
		SET status = ?,
		    reviewed_by = ?,
		    reviewed_at = CURRENT_TIMESTAMP,
		    reason = ?
		WHERE task_type = ?
		  AND resource_type = ?
		  AND resource_id = ?
		  AND status = 'open'`

	if _, err := r.db.ExecContext(
		ctx,
		query,
		string(req.Status),
		req.ReviewedBy,
		req.Reason,
		req.TaskType,
		req.ResourceType,
		req.ResourceID,
	); err != nil {
		return fmt.Errorf("close admin review task: %w", err)
	}
	return nil
}

func (r *MySQLReviewTaskRepository) getByTaskID(ctx context.Context, taskID string) (*domain.ReviewTask, error) {
	const query = `
		SELECT task_id,
		       task_type,
		       resource_type,
		       resource_id,
		       status,
		       assigned_to,
		       created_by,
		       reviewed_by,
		       reviewed_at,
		       reason,
		       metadata,
		       created_at,
		       updated_at
		FROM admin_review_tasks
		WHERE task_id = ?`

	row := r.db.QueryRowContext(ctx, query, taskID)
	task, err := scanReviewTask(row)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

type reviewTaskScanner interface {
	Scan(dest ...any) error
}

func scanReviewTask(scanner reviewTaskScanner) (domain.ReviewTask, error) {
	var task domain.ReviewTask
	var status string
	var assignedTo sql.NullString
	var createdBy sql.NullString
	var reviewedBy sql.NullString
	var reviewedAt sql.NullTime
	var reason sql.NullString
	var metadata sql.NullString
	var createdAt sql.NullTime
	var updatedAt sql.NullTime

	if err := scanner.Scan(
		&task.TaskID,
		&task.TaskType,
		&task.ResourceType,
		&task.ResourceID,
		&status,
		&assignedTo,
		&createdBy,
		&reviewedBy,
		&reviewedAt,
		&reason,
		&metadata,
		&createdAt,
		&updatedAt,
	); err != nil {
		return domain.ReviewTask{}, fmt.Errorf("scan admin review task: %w", err)
	}

	task.Status = domain.ReviewTaskStatus(status)
	task.AssignedTo = assignedTo.String
	task.CreatedBy = createdBy.String
	task.ReviewedBy = reviewedBy.String
	task.Reason = reason.String
	if reviewedAt.Valid {
		task.ReviewedAt = &reviewedAt.Time
	}
	if createdAt.Valid {
		task.CreatedAt = &createdAt.Time
	}
	if updatedAt.Valid {
		task.UpdatedAt = &updatedAt.Time
	}
	if metadata.Valid && strings.TrimSpace(metadata.String) != "" {
		var decoded map[string]any
		if err := json.Unmarshal([]byte(metadata.String), &decoded); err != nil {
			return domain.ReviewTask{}, domain.NewInternal("admin review task metadata is invalid JSON", err)
		}
		task.Metadata = decoded
	}

	return task, nil
}

func nullableJSON(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return string(value)
}

func isDuplicateKey(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
