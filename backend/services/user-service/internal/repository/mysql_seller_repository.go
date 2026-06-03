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

var _ usecase.SellerRepository = (*MySQLSellerRepository)(nil)

type MySQLSellerRepository struct {
	executor sqlExecutor
	logger   *slog.Logger
}

func NewMySQLSellerRepository(db *sql.DB, options ...Option) (*MySQLSellerRepository, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}

	return newMySQLSellerRepository(db, options...), nil
}

func newMySQLSellerRepository(executor sqlExecutor, options ...Option) *MySQLSellerRepository {
	configured := newRepositoryOptions(options)
	return &MySQLSellerRepository{executor: executor, logger: configured.logger}
}

func (r *MySQLSellerRepository) CreateSellerProfile(ctx context.Context, seller domain.SellerProfile) (domain.SellerProfile, error) {
	status := seller.Status
	if status == "" {
		status = domain.SellerStatusDraft
	}

	_, err := r.executor.ExecContext(ctx, `
		INSERT INTO seller_profiles (
			seller_id,
			user_id,
			store_name,
			display_name,
			gst_number,
			support_email,
			status,
			approved_by,
			approved_at,
			created_by,
			updated_by,
			status_changed_by,
			status_changed_at,
			status_reason,
			created_at,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		seller.SellerID,
		seller.UserID,
		seller.StoreName,
		nullableCleanStringPtr(seller.DisplayName),
		nullableCleanStringPtr(seller.GSTNumber),
		nullableCleanStringPtr(seller.SupportEmail),
		status,
		nullableCleanStringPtr(seller.ApprovedBy),
		nullableTimePtr(seller.ApprovedAt),
		seller.CreatedBy,
		seller.UpdatedBy,
		nullableCleanStringPtr(seller.StatusChangedBy),
		nullableTimePtr(seller.StatusChangedAt),
		nullableCleanStringPtr(seller.StatusReason),
		seller.CreatedAt,
		seller.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKey(err) {
			return domain.SellerProfile{}, domain.ErrDuplicateSeller
		}
		if isForeignKeyConstraint(err) {
			return domain.SellerProfile{}, domain.ErrUserNotFound
		}
		logRepositoryError(ctx, r.logger, "create_seller_profile", err)
		return domain.SellerProfile{}, fmt.Errorf("insert seller profile: %w", err)
	}

	return r.GetSellerProfileBySellerID(ctx, seller.SellerID)
}

func (r *MySQLSellerRepository) GetSellerProfileByUserID(ctx context.Context, userID string) (domain.SellerProfile, error) {
	row := r.executor.QueryRowContext(ctx, `
		SELECT
			seller_id,
			user_id,
			store_name,
			display_name,
			gst_number,
			support_email,
			status,
			approved_by,
			approved_at,
			created_by,
			updated_by,
			status_changed_by,
			status_changed_at,
			status_reason,
			created_at,
			updated_at
		FROM seller_profiles
		WHERE user_id = ?
		LIMIT 1
	`, userID)

	seller, err := scanSellerProfile(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.SellerProfile{}, domain.ErrSellerNotFound
		}
		logRepositoryError(ctx, r.logger, "get_seller_profile_by_user_id", err)
		return domain.SellerProfile{}, fmt.Errorf("query seller by user id: %w", err)
	}
	return seller, nil
}

func (r *MySQLSellerRepository) GetSellerProfileBySellerID(ctx context.Context, sellerID string) (domain.SellerProfile, error) {
	row := r.executor.QueryRowContext(ctx, `
		SELECT
			seller_id,
			user_id,
			store_name,
			display_name,
			gst_number,
			support_email,
			status,
			approved_by,
			approved_at,
			created_by,
			updated_by,
			status_changed_by,
			status_changed_at,
			status_reason,
			created_at,
			updated_at
		FROM seller_profiles
		WHERE seller_id = ?
		LIMIT 1
	`, sellerID)

	seller, err := scanSellerProfile(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.SellerProfile{}, domain.ErrSellerNotFound
		}
		logRepositoryError(ctx, r.logger, "get_seller_profile_by_seller_id", err)
		return domain.SellerProfile{}, fmt.Errorf("query seller by seller id: %w", err)
	}
	return seller, nil
}

func (r *MySQLSellerRepository) UpdateSellerProfile(ctx context.Context, sellerID string, patch domain.SellerProfilePatch) (domain.SellerProfile, error) {
	sets := make([]string, 0, 5)
	args := make([]any, 0, 6)

	if patch.StoreName != nil {
		sets = append(sets, "store_name = ?")
		args = append(args, *patch.StoreName)
	}
	if patch.DisplayName != nil {
		sets = append(sets, "display_name = ?")
		args = append(args, nullableCleanStringPtr(patch.DisplayName))
	}
	if patch.GSTNumber != nil {
		sets = append(sets, "gst_number = ?")
		args = append(args, nullableCleanStringPtr(patch.GSTNumber))
	}
	if patch.SupportEmail != nil {
		sets = append(sets, "support_email = ?")
		args = append(args, nullableCleanStringPtr(patch.SupportEmail))
	}

	if len(sets) == 0 {
		return r.GetSellerProfileBySellerID(ctx, sellerID)
	}

	sets = append(sets, "updated_by = ?", "updated_at = ?")
	args = append(args, patch.UpdatedBy, patch.UpdatedAt)
	args = append(args, sellerID)

	result, err := r.executor.ExecContext(ctx, `
		UPDATE seller_profiles
		SET `+strings.Join(sets, ", ")+`
		WHERE seller_id = ?
	`, args...)
	if err != nil {
		if isDuplicateKey(err) {
			return domain.SellerProfile{}, domain.ErrDuplicateSeller
		}
		logRepositoryError(ctx, r.logger, "update_seller_profile", err)
		return domain.SellerProfile{}, fmt.Errorf("update seller profile: %w", err)
	}

	affected, err := rowsAffected(result)
	if err != nil {
		return domain.SellerProfile{}, err
	}
	if affected == 0 {
		exists, err := r.sellerProfileExists(ctx, sellerID)
		if err != nil {
			return domain.SellerProfile{}, err
		}
		if !exists {
			return domain.SellerProfile{}, domain.ErrSellerNotFound
		}
	}

	return r.GetSellerProfileBySellerID(ctx, sellerID)
}

func (r *MySQLSellerRepository) UpdateSellerStatus(ctx context.Context, seller domain.SellerProfile) (domain.SellerProfile, error) {
	result, err := r.executor.ExecContext(ctx, `
		UPDATE seller_profiles
		SET
			status = ?,
			approved_by = ?,
			approved_at = ?,
			status_changed_by = ?,
			status_changed_at = ?,
			status_reason = ?,
			updated_by = ?,
			updated_at = ?
		WHERE seller_id = ?
	`,
		seller.Status,
		nullableCleanStringPtr(seller.ApprovedBy),
		nullableTimePtr(seller.ApprovedAt),
		nullableCleanStringPtr(seller.StatusChangedBy),
		nullableTimePtr(seller.StatusChangedAt),
		nullableCleanStringPtr(seller.StatusReason),
		seller.UpdatedBy,
		seller.UpdatedAt,
		seller.SellerID,
	)
	if err != nil {
		logRepositoryError(ctx, r.logger, "update_seller_status", err)
		return domain.SellerProfile{}, fmt.Errorf("update seller status: %w", err)
	}
	if err := ensureAffected(result, domain.ErrSellerNotFound); err != nil {
		return domain.SellerProfile{}, err
	}
	return r.GetSellerProfileBySellerID(ctx, seller.SellerID)
}

func (r *MySQLSellerRepository) AddKYCDocument(ctx context.Context, doc domain.KYCDocument) (domain.KYCDocument, error) {
	status := doc.Status
	if status == "" {
		status = domain.KYCStatusPending
	}

	_, err := r.executor.ExecContext(ctx, `
		INSERT INTO seller_kyc_documents (
			document_id,
			seller_id,
			document_type,
			storage_url,
			status,
			reviewed_by,
			reviewed_at,
			rejection_reason,
			created_by,
			updated_by,
			created_at,
			updated_at,
			status_changed_by,
			status_changed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		doc.DocumentID,
		doc.SellerID,
		doc.DocumentType,
		doc.StorageURL,
		status,
		nullableCleanStringPtr(doc.ReviewedBy),
		nullableTimePtr(doc.ReviewedAt),
		nullableCleanStringPtr(doc.RejectionReason),
		doc.CreatedBy,
		doc.UpdatedBy,
		doc.CreatedAt,
		doc.UpdatedAt,
		nullableCleanStringPtr(doc.StatusChangedBy),
		nullableTimePtr(doc.StatusChangedAt),
	)
	if err != nil {
		if isDuplicateKey(err) {
			return domain.KYCDocument{}, domain.ErrDuplicateKYCDocument
		}
		if isForeignKeyConstraint(err) {
			return domain.KYCDocument{}, domain.ErrSellerNotFound
		}
		logRepositoryError(ctx, r.logger, "add_kyc_document", err)
		return domain.KYCDocument{}, fmt.Errorf("insert kyc document: %w", err)
	}

	return r.GetKYCDocument(ctx, doc.SellerID, doc.DocumentID)
}

func (r *MySQLSellerRepository) ListKYCDocuments(ctx context.Context, sellerID string) ([]domain.KYCDocument, error) {
	rows, err := r.executor.QueryContext(ctx, `
		SELECT
			document_id,
			seller_id,
			document_type,
			storage_url,
			status,
			reviewed_by,
			reviewed_at,
			rejection_reason,
			created_by,
			updated_by,
			created_at,
			updated_at,
			status_changed_by,
			status_changed_at
		FROM seller_kyc_documents
		WHERE seller_id = ?
		ORDER BY created_at DESC
	`, sellerID)
	if err != nil {
		logRepositoryError(ctx, r.logger, "list_kyc_documents", err)
		return nil, fmt.Errorf("query kyc documents: %w", err)
	}
	defer rows.Close()

	documents := make([]domain.KYCDocument, 0)
	for rows.Next() {
		document, err := scanKYCDocument(rows)
		if err != nil {
			logRepositoryError(ctx, r.logger, "scan_kyc_document", err)
			return nil, fmt.Errorf("scan kyc document: %w", err)
		}
		documents = append(documents, document)
	}
	if err := rows.Err(); err != nil {
		logRepositoryError(ctx, r.logger, "iterate_kyc_documents", err)
		return nil, fmt.Errorf("iterate kyc documents: %w", err)
	}

	return documents, nil
}

func (r *MySQLSellerRepository) GetKYCDocument(ctx context.Context, sellerID string, documentID string) (domain.KYCDocument, error) {
	row := r.executor.QueryRowContext(ctx, `
		SELECT
			document_id,
			seller_id,
			document_type,
			storage_url,
			status,
			reviewed_by,
			reviewed_at,
			rejection_reason,
			created_by,
			updated_by,
			created_at,
			updated_at,
			status_changed_by,
			status_changed_at
		FROM seller_kyc_documents
		WHERE document_id = ?
		  AND seller_id = ?
		LIMIT 1
	`, documentID, sellerID)

	document, err := scanKYCDocument(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.KYCDocument{}, domain.ErrKYCDocumentNotFound
		}
		logRepositoryError(ctx, r.logger, "get_kyc_document", err)
		return domain.KYCDocument{}, fmt.Errorf("query kyc document: %w", err)
	}
	return document, nil
}

func (r *MySQLSellerRepository) ReviewKYCDocument(ctx context.Context, document domain.KYCDocument) (domain.KYCDocument, error) {
	result, err := r.executor.ExecContext(ctx, `
		UPDATE seller_kyc_documents
		SET
			status = ?,
			reviewed_by = ?,
			reviewed_at = ?,
			rejection_reason = ?,
			status_changed_by = ?,
			status_changed_at = ?,
			updated_by = ?,
			updated_at = ?
		WHERE seller_id = ?
		  AND document_id = ?
	`,
		document.Status,
		nullableCleanStringPtr(document.ReviewedBy),
		nullableTimePtr(document.ReviewedAt),
		nullableCleanStringPtr(document.RejectionReason),
		nullableCleanStringPtr(document.StatusChangedBy),
		nullableTimePtr(document.StatusChangedAt),
		document.UpdatedBy,
		document.UpdatedAt,
		document.SellerID,
		document.DocumentID,
	)
	if err != nil {
		logRepositoryError(ctx, r.logger, "review_kyc_document", err)
		return domain.KYCDocument{}, fmt.Errorf("review kyc document: %w", err)
	}
	if err := ensureAffected(result, domain.ErrKYCDocumentNotFound); err != nil {
		return domain.KYCDocument{}, err
	}
	return r.GetKYCDocument(ctx, document.SellerID, document.DocumentID)
}

func (r *MySQLSellerRepository) sellerProfileExists(ctx context.Context, sellerID string) (bool, error) {
	var exists int
	err := r.executor.QueryRowContext(ctx, `
		SELECT 1
		FROM seller_profiles
		WHERE seller_id = ?
		LIMIT 1
	`, sellerID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		logRepositoryError(ctx, r.logger, "seller_profile_exists", err)
		return false, fmt.Errorf("query seller profile existence: %w", err)
	}
	return true, nil
}

func scanSellerProfile(row sqlScanner) (domain.SellerProfile, error) {
	var (
		seller          domain.SellerProfile
		displayName     sql.NullString
		gstNumber       sql.NullString
		supportEmail    sql.NullString
		status          string
		approvedBy      sql.NullString
		approvedAt      sql.NullTime
		statusChangedBy sql.NullString
		statusChangedAt sql.NullTime
		statusReason    sql.NullString
	)

	err := row.Scan(
		&seller.SellerID,
		&seller.UserID,
		&seller.StoreName,
		&displayName,
		&gstNumber,
		&supportEmail,
		&status,
		&approvedBy,
		&approvedAt,
		&seller.CreatedBy,
		&seller.UpdatedBy,
		&statusChangedBy,
		&statusChangedAt,
		&statusReason,
		&seller.CreatedAt,
		&seller.UpdatedAt,
	)
	if err != nil {
		return domain.SellerProfile{}, err
	}

	seller.DisplayName = nullStringPtr(displayName)
	seller.GSTNumber = nullStringPtr(gstNumber)
	seller.SupportEmail = nullStringPtr(supportEmail)
	seller.Status = domain.SellerStatus(status)
	seller.ApprovedBy = nullStringPtr(approvedBy)
	seller.ApprovedAt = nullTimePtr(approvedAt)
	seller.StatusChangedBy = nullStringPtr(statusChangedBy)
	seller.StatusChangedAt = nullTimePtr(statusChangedAt)
	seller.StatusReason = nullStringPtr(statusReason)
	seller.CreatedAt = seller.CreatedAt.UTC()
	seller.UpdatedAt = seller.UpdatedAt.UTC()

	return seller, nil
}

func scanKYCDocument(row sqlScanner) (domain.KYCDocument, error) {
	var (
		document        domain.KYCDocument
		documentType    string
		status          string
		reviewedBy      sql.NullString
		reviewedAt      sql.NullTime
		rejectionReason sql.NullString
		statusChangedBy sql.NullString
		statusChangedAt sql.NullTime
	)

	err := row.Scan(
		&document.DocumentID,
		&document.SellerID,
		&documentType,
		&document.StorageURL,
		&status,
		&reviewedBy,
		&reviewedAt,
		&rejectionReason,
		&document.CreatedBy,
		&document.UpdatedBy,
		&document.CreatedAt,
		&document.UpdatedAt,
		&statusChangedBy,
		&statusChangedAt,
	)
	if err != nil {
		return domain.KYCDocument{}, err
	}

	document.DocumentType = domain.KYCDocumentType(documentType)
	document.Status = domain.KYCStatus(status)
	document.ReviewedBy = nullStringPtr(reviewedBy)
	document.ReviewedAt = nullTimePtr(reviewedAt)
	document.RejectionReason = nullStringPtr(rejectionReason)
	document.StatusChangedBy = nullStringPtr(statusChangedBy)
	document.StatusChangedAt = nullTimePtr(statusChangedAt)
	document.CreatedAt = document.CreatedAt.UTC()
	document.UpdatedAt = document.UpdatedAt.UTC()

	return document, nil
}
