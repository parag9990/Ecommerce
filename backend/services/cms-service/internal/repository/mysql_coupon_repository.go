package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/usecase"
)

type MySQLCouponRepository struct {
	db *sql.DB
}

func NewMySQLCouponRepository(db *sql.DB) (*MySQLCouponRepository, error) {
	if db == nil {
		return nil, errors.New("mysql db is required")
	}
	return &MySQLCouponRepository{db: db}, nil
}

func (r *MySQLCouponRepository) CreateCoupon(ctx context.Context, coupon domain.Coupon, rules []domain.CouponRule) (domain.Coupon, error) {
	coupon = coupon.Normalized()
	rules = domain.NormalizeCouponRules(rules)

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return domain.Coupon{}, err
	}
	defer rollbackIfOpen(tx)

	const query = `
INSERT INTO coupons (
  coupon_id,
  seller_id,
  code,
  discount_type,
  discount_value,
  max_discount_amount,
  min_cart_amount,
  currency,
  usage_limit,
  per_user_limit,
  status,
  starts_at,
  ends_at,
  created_by,
  created_at,
  updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.ExecContext(ctx, query,
		coupon.CouponID,
		nullStringFromPtr(coupon.SellerID),
		coupon.Code,
		coupon.DiscountType,
		coupon.DiscountValue,
		nullInt64FromPtr(coupon.MaxDiscountAmount),
		coupon.MinCartAmount,
		coupon.Currency,
		nullInt64FromPtr(coupon.UsageLimit),
		nullInt64FromPtr(coupon.PerUserLimit),
		coupon.Status,
		nullTimeFromPtr(coupon.StartsAt),
		nullTimeFromPtr(coupon.EndsAt),
		nullString(coupon.CreatedBy),
		utcOrNow(coupon.CreatedAt),
		utcOrNow(coupon.UpdatedAt),
	)
	if err != nil {
		if isDuplicateKey(err) {
			return domain.Coupon{}, domain.ErrCouponCodeAlreadyExists
		}
		return domain.Coupon{}, err
	}

	if err := insertCouponRules(ctx, tx, coupon.CouponID, rules); err != nil {
		return domain.Coupon{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Coupon{}, err
	}
	tx = nil
	return r.GetCouponByID(ctx, coupon.CouponID)
}

func (r *MySQLCouponRepository) UpdateCoupon(ctx context.Context, coupon domain.Coupon, rules []domain.CouponRule) (domain.Coupon, error) {
	coupon = coupon.Normalized()
	rules = domain.NormalizeCouponRules(rules)

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return domain.Coupon{}, err
	}
	defer rollbackIfOpen(tx)

	const updateQuery = `
UPDATE coupons
SET code = ?,
    discount_type = ?,
    discount_value = ?,
    max_discount_amount = ?,
    min_cart_amount = ?,
    currency = ?,
    usage_limit = ?,
    per_user_limit = ?,
    status = ?,
    starts_at = ?,
    ends_at = ?,
    updated_at = ?
WHERE coupon_id = ?`
	result, err := tx.ExecContext(ctx, updateQuery,
		coupon.Code,
		coupon.DiscountType,
		coupon.DiscountValue,
		nullInt64FromPtr(coupon.MaxDiscountAmount),
		coupon.MinCartAmount,
		coupon.Currency,
		nullInt64FromPtr(coupon.UsageLimit),
		nullInt64FromPtr(coupon.PerUserLimit),
		coupon.Status,
		nullTimeFromPtr(coupon.StartsAt),
		nullTimeFromPtr(coupon.EndsAt),
		utcOrNow(coupon.UpdatedAt),
		coupon.CouponID,
	)
	if err != nil {
		if isDuplicateKey(err) {
			return domain.Coupon{}, domain.ErrCouponCodeAlreadyExists
		}
		return domain.Coupon{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return domain.Coupon{}, err
	}
	if affected == 0 {
		return domain.Coupon{}, domain.ErrCouponNotFound
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM coupon_rules WHERE coupon_id = ?`, coupon.CouponID); err != nil {
		return domain.Coupon{}, err
	}
	if err := insertCouponRules(ctx, tx, coupon.CouponID, rules); err != nil {
		return domain.Coupon{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Coupon{}, err
	}
	tx = nil
	return r.GetCouponByID(ctx, coupon.CouponID)
}

func (r *MySQLCouponRepository) UpdateCouponStatus(ctx context.Context, couponID string, status domain.CouponStatus, updatedAt time.Time) (domain.Coupon, error) {
	status = status.Normalized()
	if !status.Valid() {
		return domain.Coupon{}, domain.ErrInvalidCouponStatus
	}
	const query = `
UPDATE coupons
SET status = ?, updated_at = ?
WHERE coupon_id = ?`
	result, err := r.db.ExecContext(ctx, query, status, utcOrNow(updatedAt), strings.TrimSpace(couponID))
	if err != nil {
		return domain.Coupon{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return domain.Coupon{}, err
	}
	if affected == 0 {
		return domain.Coupon{}, domain.ErrCouponNotFound
	}
	return r.GetCouponByID(ctx, couponID)
}

func (r *MySQLCouponRepository) GetCouponByID(ctx context.Context, couponID string) (domain.Coupon, error) {
	const query = `
SELECT coupon_id, seller_id, code, discount_type, discount_value, max_discount_amount,
       min_cart_amount, currency, usage_limit, per_user_limit, status, starts_at,
       ends_at, created_by, created_at, updated_at
FROM coupons
WHERE coupon_id = ?
LIMIT 1`
	return scanCoupon(r.db.QueryRowContext(ctx, query, strings.TrimSpace(couponID)))
}

func (r *MySQLCouponRepository) FindCouponByCode(ctx context.Context, code string) (domain.Coupon, error) {
	const query = `
SELECT coupon_id, seller_id, code, discount_type, discount_value, max_discount_amount,
       min_cart_amount, currency, usage_limit, per_user_limit, status, starts_at,
       ends_at, created_by, created_at, updated_at
FROM coupons
WHERE code = ?
LIMIT 1`
	return scanCoupon(r.db.QueryRowContext(ctx, query, domain.NormalizeCouponCode(code)))
}

func (r *MySQLCouponRepository) ListCoupons(ctx context.Context, filter usecase.CouponFilter) ([]domain.Coupon, error) {
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
SELECT coupon_id, seller_id, code, discount_type, discount_value, max_discount_amount,
       min_cart_amount, currency, usage_limit, per_user_limit, status, starts_at,
       ends_at, created_by, created_at, updated_at
FROM coupons
WHERE 1 = 1`)
	args := make([]any, 0, 5)
	if strings.TrimSpace(filter.SellerID) != "" {
		query.WriteString(" AND seller_id = ?")
		args = append(args, strings.TrimSpace(filter.SellerID))
	}
	if filter.Status.Normalized() != "" {
		query.WriteString(" AND status = ?")
		args = append(args, filter.Status.Normalized())
	}
	if strings.TrimSpace(filter.Code) != "" {
		query.WriteString(" AND code = ?")
		args = append(args, domain.NormalizeCouponCode(filter.Code))
	}
	query.WriteString(" ORDER BY created_at DESC, id DESC LIMIT ? OFFSET ?")
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	coupons := make([]domain.Coupon, 0, limit)
	for rows.Next() {
		coupon, err := scanCoupon(rows)
		if err != nil {
			return nil, err
		}
		coupons = append(coupons, coupon)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return coupons, nil
}

func (r *MySQLCouponRepository) ListCouponRules(ctx context.Context, couponID string) ([]domain.CouponRule, error) {
	const query = `
SELECT rule_id, coupon_id, rule_type, rule_value, created_at
FROM coupon_rules
WHERE coupon_id = ?
ORDER BY id ASC`
	rows, err := r.db.QueryContext(ctx, query, strings.TrimSpace(couponID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := make([]domain.CouponRule, 0)
	for rows.Next() {
		rule, err := scanCouponRule(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *MySQLCouponRepository) CountRedemptions(ctx context.Context, couponID string) (int64, error) {
	const query = `SELECT COUNT(*) FROM coupon_redemptions WHERE coupon_id = ?`
	var count int64
	err := r.db.QueryRowContext(ctx, query, strings.TrimSpace(couponID)).Scan(&count)
	return count, err
}

func (r *MySQLCouponRepository) CountUserRedemptions(ctx context.Context, couponID string, userID string) (int64, error) {
	const query = `SELECT COUNT(*) FROM coupon_redemptions WHERE coupon_id = ? AND user_id = ?`
	var count int64
	err := r.db.QueryRowContext(ctx, query, strings.TrimSpace(couponID), strings.TrimSpace(userID)).Scan(&count)
	return count, err
}

func (r *MySQLCouponRepository) InsertRedemption(ctx context.Context, redemption domain.CouponRedemption, coupon domain.Coupon) error {
	coupon = coupon.Normalized()
	redemption.CouponID = strings.TrimSpace(redemption.CouponID)
	redemption.OrderID = strings.TrimSpace(redemption.OrderID)
	redemption.UserID = strings.TrimSpace(redemption.UserID)
	redemption.Discount.Currency = domain.NormalizeCurrency(redemption.Discount.Currency)

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer rollbackIfOpen(tx)

	lockedCoupon, err := scanCoupon(tx.QueryRowContext(ctx, `
SELECT coupon_id, seller_id, code, discount_type, discount_value, max_discount_amount,
       min_cart_amount, currency, usage_limit, per_user_limit, status, starts_at,
       ends_at, created_by, created_at, updated_at
FROM coupons
WHERE coupon_id = ?
LIMIT 1
FOR UPDATE`, redemption.CouponID))
	if err != nil {
		return err
	}
	if lockedCoupon.Currency != redemption.Discount.Currency {
		return domain.ErrCouponCurrencyMismatch
	}
	if redemption.CampaignID != "" {
		if err := validateCampaignRedemptionTx(ctx, tx, redemption, lockedCoupon); err != nil {
			return err
		}
	}

	const query = `
INSERT INTO coupon_redemptions (
  redemption_id,
  coupon_id,
  campaign_id,
  order_id,
  user_id,
  discount_amount,
  currency,
  request_id,
  redeemed_at,
  created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = tx.ExecContext(ctx, query,
		redemption.RedemptionID,
		redemption.CouponID,
		nullString(redemption.CampaignID),
		redemption.OrderID,
		redemption.UserID,
		redemption.Discount.Amount,
		redemption.Discount.Currency,
		nullString(redemption.RequestID),
		utcOrNow(redemption.RedeemedAt),
		utcOrNow(redemption.CreatedAt),
	)
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrCouponRedemptionDuplicate
		}
		return err
	}

	if lockedCoupon.UsageLimit != nil {
		count, err := countRedemptionsTx(ctx, tx, redemption.CouponID)
		if err != nil {
			return err
		}
		if count > *lockedCoupon.UsageLimit {
			return domain.ErrCouponUsageLimitReached
		}
	}
	if lockedCoupon.PerUserLimit != nil {
		count, err := countUserRedemptionsTx(ctx, tx, redemption.CouponID, redemption.UserID)
		if err != nil {
			return err
		}
		if count > *lockedCoupon.PerUserLimit {
			return domain.ErrCouponPerUserLimitReached
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	tx = nil
	return nil
}

func (r *MySQLCouponRepository) GetRedemption(ctx context.Context, couponID string, orderID string) (domain.CouponRedemption, error) {
	const query = `
SELECT redemption_id, coupon_id, campaign_id, order_id, user_id, discount_amount, currency, request_id, redeemed_at, created_at
FROM coupon_redemptions
WHERE coupon_id = ? AND order_id = ?
LIMIT 1`
	return scanCouponRedemption(r.db.QueryRowContext(ctx, query, strings.TrimSpace(couponID), strings.TrimSpace(orderID)))
}

type couponScanner interface {
	Scan(dest ...any) error
}

func scanCoupon(scanner couponScanner) (domain.Coupon, error) {
	var coupon domain.Coupon
	var sellerID sql.NullString
	var discountType string
	var maxDiscountAmount sql.NullInt64
	var usageLimit sql.NullInt64
	var perUserLimit sql.NullInt64
	var status string
	var startsAt sql.NullTime
	var endsAt sql.NullTime
	var createdBy sql.NullString

	err := scanner.Scan(
		&coupon.CouponID,
		&sellerID,
		&coupon.Code,
		&discountType,
		&coupon.DiscountValue,
		&maxDiscountAmount,
		&coupon.MinCartAmount,
		&coupon.Currency,
		&usageLimit,
		&perUserLimit,
		&status,
		&startsAt,
		&endsAt,
		&createdBy,
		&coupon.CreatedAt,
		&coupon.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Coupon{}, domain.ErrCouponNotFound
		}
		return domain.Coupon{}, err
	}
	if sellerID.Valid {
		coupon.SellerID = &sellerID.String
	}
	coupon.DiscountType = domain.DiscountType(discountType).Normalized()
	coupon.MaxDiscountAmount = int64PtrFromNull(maxDiscountAmount)
	coupon.UsageLimit = int64PtrFromNull(usageLimit)
	coupon.PerUserLimit = int64PtrFromNull(perUserLimit)
	coupon.Status = domain.CouponStatus(status).Normalized()
	coupon.StartsAt = timePtrFromNull(startsAt)
	coupon.EndsAt = timePtrFromNull(endsAt)
	coupon.CreatedBy = createdBy.String
	return coupon.Normalized(), nil
}

func scanCouponRule(scanner couponScanner) (domain.CouponRule, error) {
	var rule domain.CouponRule
	var ruleType string
	var rawValue []byte
	err := scanner.Scan(
		&rule.RuleID,
		&rule.CouponID,
		&ruleType,
		&rawValue,
		&rule.CreatedAt,
	)
	if err != nil {
		return domain.CouponRule{}, err
	}
	rule.Type = domain.CouponRuleType(ruleType).Normalized()
	if err := decodeRuleValue(rawValue, &rule); err != nil {
		return domain.CouponRule{}, err
	}
	return rule.Normalized(), nil
}

func scanCouponRedemption(scanner couponScanner) (domain.CouponRedemption, error) {
	var redemption domain.CouponRedemption
	var campaignID sql.NullString
	var currency string
	var requestID sql.NullString
	err := scanner.Scan(
		&redemption.RedemptionID,
		&redemption.CouponID,
		&campaignID,
		&redemption.OrderID,
		&redemption.UserID,
		&redemption.Discount.Amount,
		&currency,
		&requestID,
		&redemption.RedeemedAt,
		&redemption.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.CouponRedemption{}, domain.ErrCouponRedemptionNotFound
		}
		return domain.CouponRedemption{}, err
	}
	redemption.CampaignID = campaignID.String
	redemption.Discount.Currency = domain.NormalizeCurrency(currency)
	redemption.RequestID = requestID.String
	return redemption, nil
}

func insertCouponRules(ctx context.Context, tx *sql.Tx, couponID string, rules []domain.CouponRule) error {
	const query = `
INSERT INTO coupon_rules (
  rule_id,
  coupon_id,
  rule_type,
  rule_value,
  created_at
) VALUES (?, ?, ?, ?, ?)`
	for _, rule := range rules {
		rule = rule.Normalized()
		valueJSON, err := encodeRuleValue(rule)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, query, rule.RuleID, couponID, rule.Type, valueJSON, utcOrNow(rule.CreatedAt)); err != nil {
			return err
		}
	}
	return nil
}

func encodeRuleValue(rule domain.CouponRule) ([]byte, error) {
	rule = rule.Normalized()
	switch rule.Type {
	case domain.CouponRuleTypeProductScope:
		return json.Marshal(map[string][]string{"product_ids": rule.ProductIDs})
	case domain.CouponRuleTypeCategoryScope:
		return json.Marshal(map[string][]string{"category_ids": rule.CategoryIDs})
	case domain.CouponRuleTypeSellerScope:
		return json.Marshal(map[string][]string{"seller_ids": rule.SellerIDs})
	default:
		return nil, domain.ErrInvalidCouponRule
	}
}

func decodeRuleValue(raw []byte, rule *domain.CouponRule) error {
	var value struct {
		ProductIDs  []string `json:"product_ids"`
		CategoryIDs []string `json:"category_ids"`
		SellerIDs   []string `json:"seller_ids"`
	}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &value); err != nil {
			return err
		}
	}
	rule.ProductIDs = value.ProductIDs
	rule.CategoryIDs = value.CategoryIDs
	rule.SellerIDs = value.SellerIDs
	return nil
}

func countRedemptionsTx(ctx context.Context, tx *sql.Tx, couponID string) (int64, error) {
	var count int64
	err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM coupon_redemptions WHERE coupon_id = ?`, strings.TrimSpace(couponID)).Scan(&count)
	return count, err
}

func countUserRedemptionsTx(ctx context.Context, tx *sql.Tx, couponID string, userID string) (int64, error) {
	var count int64
	err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM coupon_redemptions WHERE coupon_id = ? AND user_id = ?`, strings.TrimSpace(couponID), strings.TrimSpace(userID)).Scan(&count)
	return count, err
}

func validateCampaignRedemptionTx(ctx context.Context, tx *sql.Tx, redemption domain.CouponRedemption, coupon domain.Coupon) error {
	campaign, err := scanCampaign(tx.QueryRowContext(ctx, `
SELECT campaign_id, seller_id, name, status, budget_amount, currency, starts_at,
       ends_at, metadata, created_by, created_at, updated_at
FROM campaigns
WHERE campaign_id = ?
LIMIT 1
FOR UPDATE`, strings.TrimSpace(redemption.CampaignID)))
	if err != nil {
		return err
	}
	if reason := domain.ValidateCampaignEligibility(campaign, sellerIDFromCoupon(coupon), coupon.CouponID, redemption.Discount.Currency, time.Now().UTC()); reason != "" {
		return campaignReasonToError(reason)
	}
	campaign = campaign.Normalized()
	if campaign.Metadata.UsageLimit != nil {
		count, err := countCampaignRedemptions(ctx, tx, campaign.CampaignID)
		if err != nil {
			return err
		}
		if count >= *campaign.Metadata.UsageLimit {
			return domain.ErrCampaignUsageLimitReached
		}
	}
	if campaign.Metadata.PerUserLimit != nil {
		count, err := countCampaignUserRedemptions(ctx, tx, campaign.CampaignID, redemption.UserID)
		if err != nil {
			return err
		}
		if count >= *campaign.Metadata.PerUserLimit {
			return domain.ErrCampaignPerUserLimitReached
		}
	}
	if campaign.BudgetAmount != nil {
		spent, err := sumCampaignDiscounts(ctx, tx, campaign.CampaignID)
		if err != nil {
			return err
		}
		if redemption.Discount.Amount > *campaign.BudgetAmount-spent {
			return domain.ErrCampaignBudgetExhausted
		}
	}
	return nil
}

func sellerIDFromCoupon(coupon domain.Coupon) string {
	coupon = coupon.Normalized()
	if coupon.SellerID == nil {
		return ""
	}
	return strings.TrimSpace(*coupon.SellerID)
}

func campaignReasonToError(reason domain.CampaignInvalidReason) error {
	switch reason {
	case domain.CampaignInvalidReasonSellerMismatch:
		return domain.ErrCampaignOwnershipMismatch
	case domain.CampaignInvalidReasonNotActive,
		domain.CampaignInvalidReasonNotStarted,
		domain.CampaignInvalidReasonExpired:
		return domain.ErrInvalidCampaignStatusTransition
	case domain.CampaignInvalidReasonCurrencyMismatch:
		return domain.ErrCampaignCurrencyMismatch
	case domain.CampaignInvalidReasonCouponNotLinked:
		return domain.ErrCampaignCouponNotLinked
	case domain.CampaignInvalidReasonBudgetExhausted:
		return domain.ErrCampaignBudgetExhausted
	case domain.CampaignInvalidReasonUsageLimitReached:
		return domain.ErrCampaignUsageLimitReached
	default:
		return domain.ErrValidationFailed
	}
}

func rollbackIfOpen(tx *sql.Tx) {
	if tx != nil {
		_ = tx.Rollback()
	}
}

func nullStringFromPtr(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return nullString(*value)
}

func nullInt64FromPtr(value *int64) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *value, Valid: true}
}

func nullTimeFromPtr(value *time.Time) sql.NullTime {
	if value == nil || value.IsZero() {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: value.UTC(), Valid: true}
}

func int64PtrFromNull(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	out := value.Int64
	return &out
}

func timePtrFromNull(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	out := value.Time.UTC()
	return &out
}

func utcOrNow(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}
	return value.UTC()
}

func (r *MySQLCouponRepository) String() string {
	return fmt.Sprintf("%T", r)
}
