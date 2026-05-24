package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/usecase"
	"github.com/go-sql-driver/mysql"
)

const (
	defaultReviewListLimit = 50
	maxReviewListLimit     = 500
)

type MySQLProductModerationRepository struct {
	db *sql.DB
}

func NewMySQLProductModerationRepository(db *sql.DB) (*MySQLProductModerationRepository, error) {
	if db == nil {
		return nil, errors.New("mysql db is required")
	}
	return &MySQLProductModerationRepository{db: db}, nil
}

func (r *MySQLProductModerationRepository) CreateSubmittedReview(ctx context.Context, review domain.ProductModerationReview) (domain.ProductModerationReview, error) {
	review = normalizeReview(review)
	if review.ReviewID == "" || review.ProductID == "" || review.SellerID == "" || review.SubmittedBy == "" {
		return domain.ProductModerationReview{}, domain.NewValidationError("Invalid product review.", []domain.FieldViolation{
			{Field: "review", Reason: "Review id, product id, seller id, and submitted by are required."},
		})
	}

	const query = `
INSERT INTO product_moderation_reviews (
  review_id,
  product_id,
  seller_id,
  status,
  submitted_by,
  submitted_at,
  created_at,
  updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		review.ReviewID,
		review.ProductID,
		review.SellerID,
		domain.ProductReviewStatusSubmitted,
		review.SubmittedBy,
		review.SubmittedAt,
		review.CreatedAt,
		review.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKey(err) {
			return r.getSubmittedByProduct(ctx, review.ProductID)
		}
		return domain.ProductModerationReview{}, err
	}
	return r.GetReview(ctx, review.ReviewID)
}

func (r *MySQLProductModerationRepository) GetReview(ctx context.Context, reviewID string) (domain.ProductModerationReview, error) {
	const query = `
SELECT review_id, product_id, seller_id, status, submitted_by, reviewed_by, rejection_reason, submitted_at, reviewed_at, created_at, updated_at
FROM product_moderation_reviews
WHERE review_id = ?
LIMIT 1`

	return scanReview(r.db.QueryRowContext(ctx, query, strings.TrimSpace(reviewID)))
}

func (r *MySQLProductModerationRepository) MarkDecision(ctx context.Context, reviewID string, decision domain.ModerationDecision, reviewedBy string, reason string, reviewedAt time.Time) (domain.ProductModerationReview, error) {
	reviewID = strings.TrimSpace(reviewID)
	reviewedBy = strings.TrimSpace(reviewedBy)
	reason = domain.NormalizeModerationReason(reason)
	decision = decision.Normalized()
	if reviewID == "" || reviewedBy == "" || !decision.Valid() {
		return domain.ProductModerationReview{}, domain.ErrInvalidModerationDecision
	}
	if decision == domain.ModerationDecisionReject && reason == "" {
		return domain.ProductModerationReview{}, domain.ErrRejectionReasonRequired
	}
	if reviewedAt.IsZero() {
		reviewedAt = time.Now().UTC()
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return domain.ProductModerationReview{}, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	review, err := scanReview(tx.QueryRowContext(ctx, `
SELECT review_id, product_id, seller_id, status, submitted_by, reviewed_by, rejection_reason, submitted_at, reviewed_at, created_at, updated_at
FROM product_moderation_reviews
WHERE review_id = ?
LIMIT 1
FOR UPDATE`, reviewID))
	if err != nil {
		return domain.ProductModerationReview{}, err
	}
	if review.Status.Normalized() != domain.ProductReviewStatusSubmitted {
		return domain.ProductModerationReview{}, domain.ErrReviewAlreadyDecided
	}

	const updateQuery = `
UPDATE product_moderation_reviews
SET status = ?, reviewed_by = ?, rejection_reason = ?, reviewed_at = ?, updated_at = ?
WHERE review_id = ? AND status = ?`
	result, err := tx.ExecContext(ctx, updateQuery,
		decision.ReviewStatus(),
		reviewedBy,
		nullString(reason),
		reviewedAt.UTC(),
		reviewedAt.UTC(),
		reviewID,
		domain.ProductReviewStatusSubmitted,
	)
	if err != nil {
		return domain.ProductModerationReview{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return domain.ProductModerationReview{}, err
	}
	if affected != 1 {
		return domain.ProductModerationReview{}, domain.ErrReviewAlreadyDecided
	}
	if err := tx.Commit(); err != nil {
		return domain.ProductModerationReview{}, err
	}
	tx = nil
	return r.GetReview(ctx, reviewID)
}

func (r *MySQLProductModerationRepository) CancelSubmittedReview(ctx context.Context, reviewID string, reason string, cancelledAt time.Time) error {
	reviewID = strings.TrimSpace(reviewID)
	if reviewID == "" {
		return nil
	}
	if cancelledAt.IsZero() {
		cancelledAt = time.Now().UTC()
	}
	const query = `
UPDATE product_moderation_reviews
SET status = ?, rejection_reason = ?, reviewed_at = ?, updated_at = ?
WHERE review_id = ? AND status = ?`
	_, err := r.db.ExecContext(ctx, query,
		domain.ProductReviewStatusCancelled,
		nullString(domain.NormalizeModerationReason(reason)),
		cancelledAt.UTC(),
		cancelledAt.UTC(),
		reviewID,
		domain.ProductReviewStatusSubmitted,
	)
	return err
}

func (r *MySQLProductModerationRepository) ListReviews(ctx context.Context, filter usecase.ProductReviewFilter) ([]domain.ProductModerationReview, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = defaultReviewListLimit
	}
	if limit > maxReviewListLimit {
		limit = maxReviewListLimit
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := strings.Builder{}
	query.WriteString(`
SELECT review_id, product_id, seller_id, status, submitted_by, reviewed_by, rejection_reason, submitted_at, reviewed_at, created_at, updated_at
FROM product_moderation_reviews
WHERE 1 = 1`)
	args := make([]any, 0, 4)
	if filter.Status.Normalized() != "" {
		query.WriteString(" AND status = ?")
		args = append(args, filter.Status.Normalized())
	}
	if strings.TrimSpace(filter.SellerID) != "" {
		query.WriteString(" AND seller_id = ?")
		args = append(args, strings.TrimSpace(filter.SellerID))
	}
	query.WriteString(" ORDER BY submitted_at ASC, id ASC LIMIT ? OFFSET ?")
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := make([]domain.ProductModerationReview, 0, limit)
	for rows.Next() {
		review, err := scanReview(rows)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *MySQLProductModerationRepository) getSubmittedByProduct(ctx context.Context, productID string) (domain.ProductModerationReview, error) {
	const query = `
SELECT review_id, product_id, seller_id, status, submitted_by, reviewed_by, rejection_reason, submitted_at, reviewed_at, created_at, updated_at
FROM product_moderation_reviews
WHERE product_id = ? AND status = ?
LIMIT 1`
	return scanReview(r.db.QueryRowContext(ctx, query, strings.TrimSpace(productID), domain.ProductReviewStatusSubmitted))
}

type reviewScanner interface {
	Scan(dest ...any) error
}

func scanReview(scanner reviewScanner) (domain.ProductModerationReview, error) {
	var review domain.ProductModerationReview
	var status string
	var reviewedBy sql.NullString
	var rejectionReason sql.NullString
	var reviewedAt sql.NullTime
	err := scanner.Scan(
		&review.ReviewID,
		&review.ProductID,
		&review.SellerID,
		&status,
		&review.SubmittedBy,
		&reviewedBy,
		&rejectionReason,
		&review.SubmittedAt,
		&reviewedAt,
		&review.CreatedAt,
		&review.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ProductModerationReview{}, domain.ErrReviewNotFound
		}
		return domain.ProductModerationReview{}, err
	}
	review.Status = domain.ProductReviewStatus(status).Normalized()
	review.ReviewedBy = reviewedBy.String
	review.RejectionReason = rejectionReason.String
	if reviewedAt.Valid {
		review.ReviewedAt = reviewedAt.Time
	}
	return review, nil
}

func normalizeReview(review domain.ProductModerationReview) domain.ProductModerationReview {
	now := time.Now().UTC()
	review.ReviewID = strings.TrimSpace(review.ReviewID)
	review.ProductID = strings.TrimSpace(review.ProductID)
	review.SellerID = strings.TrimSpace(review.SellerID)
	review.Status = review.Status.Normalized()
	if review.Status == "" {
		review.Status = domain.ProductReviewStatusSubmitted
	}
	review.SubmittedBy = strings.TrimSpace(review.SubmittedBy)
	review.ReviewedBy = strings.TrimSpace(review.ReviewedBy)
	review.RejectionReason = domain.NormalizeModerationReason(review.RejectionReason)
	if review.SubmittedAt.IsZero() {
		review.SubmittedAt = now
	}
	if review.CreatedAt.IsZero() {
		review.CreatedAt = now
	}
	if review.UpdatedAt.IsZero() {
		review.UpdatedAt = review.CreatedAt
	}
	return review
}

func isDuplicateKey(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

func nullString(value string) sql.NullString {
	value = strings.TrimSpace(value)
	return sql.NullString{String: value, Valid: value != ""}
}

func (r *MySQLProductModerationRepository) String() string {
	return fmt.Sprintf("%T", r)
}
