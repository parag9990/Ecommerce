package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/usecase"
)

type MySQLCampaignRepository struct {
	db *sql.DB
}

func NewMySQLCampaignRepository(db *sql.DB) (*MySQLCampaignRepository, error) {
	if db == nil {
		return nil, errors.New("mysql db is required")
	}
	return &MySQLCampaignRepository{db: db}, nil
}

func (r *MySQLCampaignRepository) CreateCampaign(ctx context.Context, campaign domain.Campaign) (domain.Campaign, error) {
	campaign = campaign.Normalized()
	metadataJSON, err := encodeCampaignMetadata(campaign.Metadata)
	if err != nil {
		return domain.Campaign{}, err
	}

	const query = `
INSERT INTO campaigns (
  campaign_id,
  seller_id,
  name,
  status,
  budget_amount,
  currency,
  starts_at,
  ends_at,
  metadata,
  created_by,
  created_at,
  updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = r.db.ExecContext(ctx, query,
		campaign.CampaignID,
		nullStringFromPtr(campaign.SellerID),
		campaign.Name,
		campaign.Status,
		nullInt64FromPtr(campaign.BudgetAmount),
		campaign.Currency,
		utcOrNow(campaign.StartsAt),
		utcOrNow(campaign.EndsAt),
		metadataJSON,
		campaign.CreatedBy,
		utcOrNow(campaign.CreatedAt),
		utcOrNow(campaign.UpdatedAt),
	)
	if err != nil {
		return domain.Campaign{}, err
	}
	return r.GetCampaignByID(ctx, campaign.CampaignID)
}

func (r *MySQLCampaignRepository) UpdateCampaign(ctx context.Context, campaign domain.Campaign) (domain.Campaign, error) {
	campaign = campaign.Normalized()
	metadataJSON, err := encodeCampaignMetadata(campaign.Metadata)
	if err != nil {
		return domain.Campaign{}, err
	}

	const query = `
UPDATE campaigns
SET name = ?,
    status = ?,
    budget_amount = ?,
    currency = ?,
    starts_at = ?,
    ends_at = ?,
    metadata = ?,
    updated_at = ?
WHERE campaign_id = ?`
	result, err := r.db.ExecContext(ctx, query,
		campaign.Name,
		campaign.Status,
		nullInt64FromPtr(campaign.BudgetAmount),
		campaign.Currency,
		utcOrNow(campaign.StartsAt),
		utcOrNow(campaign.EndsAt),
		metadataJSON,
		utcOrNow(campaign.UpdatedAt),
		campaign.CampaignID,
	)
	if err != nil {
		return domain.Campaign{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return domain.Campaign{}, err
	}
	if affected == 0 {
		return domain.Campaign{}, domain.ErrCampaignNotFound
	}
	return r.GetCampaignByID(ctx, campaign.CampaignID)
}

func (r *MySQLCampaignRepository) UpdateCampaignStatus(ctx context.Context, campaignID string, sellerID string, status domain.CampaignStatus, updatedAt time.Time) (domain.Campaign, error) {
	status = status.Normalized()
	if !status.Valid() {
		return domain.Campaign{}, domain.ErrInvalidCampaignStatus
	}
	const query = `
UPDATE campaigns
SET status = ?, updated_at = ?
WHERE campaign_id = ? AND seller_id = ?`
	result, err := r.db.ExecContext(ctx, query, status, utcOrNow(updatedAt), strings.TrimSpace(campaignID), strings.TrimSpace(sellerID))
	if err != nil {
		return domain.Campaign{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return domain.Campaign{}, err
	}
	if affected == 0 {
		return domain.Campaign{}, domain.ErrCampaignNotFound
	}
	return r.GetCampaignByID(ctx, campaignID)
}

func (r *MySQLCampaignRepository) GetCampaignByID(ctx context.Context, campaignID string) (domain.Campaign, error) {
	const query = `
SELECT campaign_id, seller_id, name, status, budget_amount, currency, starts_at,
       ends_at, metadata, created_by, created_at, updated_at
FROM campaigns
WHERE campaign_id = ?
LIMIT 1`
	return scanCampaign(r.db.QueryRowContext(ctx, query, strings.TrimSpace(campaignID)))
}

func (r *MySQLCampaignRepository) ListCampaigns(ctx context.Context, filter usecase.CampaignFilter) ([]domain.Campaign, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := strings.Builder{}
	query.WriteString(`
SELECT campaign_id, seller_id, name, status, budget_amount, currency, starts_at,
       ends_at, metadata, created_by, created_at, updated_at
FROM campaigns
WHERE seller_id = ?`)
	args := []any{strings.TrimSpace(filter.SellerID)}
	if filter.Status.Normalized() != "" {
		query.WriteString(" AND status = ?")
		args = append(args, filter.Status.Normalized())
	}
	if filter.StartsAfter != nil {
		query.WriteString(" AND starts_at >= ?")
		args = append(args, filter.StartsAfter.UTC())
	}
	if filter.EndsBefore != nil {
		query.WriteString(" AND ends_at <= ?")
		args = append(args, filter.EndsBefore.UTC())
	}
	if strings.TrimSpace(filter.Query) != "" {
		query.WriteString(" AND name LIKE ?")
		args = append(args, "%"+strings.TrimSpace(filter.Query)+"%")
	}
	query.WriteString(" ORDER BY starts_at DESC, id DESC LIMIT ? OFFSET ?")
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	campaigns := make([]domain.Campaign, 0, limit)
	for rows.Next() {
		campaign, err := scanCampaign(rows)
		if err != nil {
			return nil, err
		}
		campaigns = append(campaigns, campaign)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return campaigns, nil
}

func (r *MySQLCampaignRepository) CountCampaignRedemptions(ctx context.Context, campaignID string) (int64, error) {
	return countCampaignRedemptions(ctx, r.db, campaignID)
}

func (r *MySQLCampaignRepository) CountCampaignUserRedemptions(ctx context.Context, campaignID string, userID string) (int64, error) {
	return countCampaignUserRedemptions(ctx, r.db, campaignID, userID)
}

func (r *MySQLCampaignRepository) SumCampaignDiscounts(ctx context.Context, campaignID string) (int64, error) {
	return sumCampaignDiscounts(ctx, r.db, campaignID)
}

type campaignScanner interface {
	Scan(dest ...any) error
}

func scanCampaign(scanner campaignScanner) (domain.Campaign, error) {
	var campaign domain.Campaign
	var sellerID sql.NullString
	var status string
	var budgetAmount sql.NullInt64
	var rawMetadata []byte

	err := scanner.Scan(
		&campaign.CampaignID,
		&sellerID,
		&campaign.Name,
		&status,
		&budgetAmount,
		&campaign.Currency,
		&campaign.StartsAt,
		&campaign.EndsAt,
		&rawMetadata,
		&campaign.CreatedBy,
		&campaign.CreatedAt,
		&campaign.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Campaign{}, domain.ErrCampaignNotFound
		}
		return domain.Campaign{}, err
	}
	if sellerID.Valid {
		campaign.SellerID = &sellerID.String
	}
	campaign.Status = domain.CampaignStatus(status).Normalized()
	campaign.BudgetAmount = int64PtrFromNull(budgetAmount)
	if len(rawMetadata) > 0 {
		if err := json.Unmarshal(rawMetadata, &campaign.Metadata); err != nil {
			return domain.Campaign{}, err
		}
	}
	return campaign.Normalized(), nil
}

func encodeCampaignMetadata(metadata domain.CampaignMetadata) ([]byte, error) {
	return json.Marshal(metadata.Normalized())
}

type sqlQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func countCampaignRedemptions(ctx context.Context, q sqlQuerier, campaignID string) (int64, error) {
	var count int64
	err := q.QueryRowContext(ctx, `SELECT COUNT(*) FROM coupon_redemptions WHERE campaign_id = ?`, strings.TrimSpace(campaignID)).Scan(&count)
	return count, err
}

func countCampaignUserRedemptions(ctx context.Context, q sqlQuerier, campaignID string, userID string) (int64, error) {
	var count int64
	err := q.QueryRowContext(ctx, `SELECT COUNT(*) FROM coupon_redemptions WHERE campaign_id = ? AND user_id = ?`, strings.TrimSpace(campaignID), strings.TrimSpace(userID)).Scan(&count)
	return count, err
}

func sumCampaignDiscounts(ctx context.Context, q sqlQuerier, campaignID string) (int64, error) {
	var sum sql.NullInt64
	err := q.QueryRowContext(ctx, `SELECT COALESCE(SUM(discount_amount), 0) FROM coupon_redemptions WHERE campaign_id = ?`, strings.TrimSpace(campaignID)).Scan(&sum)
	if err != nil {
		return 0, err
	}
	if !sum.Valid {
		return 0, nil
	}
	return sum.Int64, nil
}
