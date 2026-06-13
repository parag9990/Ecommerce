package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/example/ecommerce-platform/backend/services/cms-service/internal/domain"
)

type MySQLSellerAnalyticsRepository struct {
	db *sql.DB
}

func NewMySQLSellerAnalyticsRepository(db *sql.DB) (*MySQLSellerAnalyticsRepository, error) {
	if db == nil {
		return nil, errors.New("mysql db is required")
	}
	return &MySQLSellerAnalyticsRepository{db: db}, nil
}

func (r *MySQLSellerAnalyticsRepository) GetSellerSummary(ctx context.Context, query domain.AnalyticsQuery) (domain.SellerAnalyticsSummary, error) {
	query = query.Normalized()
	const statement = `
SELECT
  COALESCE(SUM(net_revenue_amount), 0) AS revenue_amount,
  COALESCE(SUM(paid_order_count), 0) AS paid_orders,
  COALESCE(SUM(paid_item_count), 0) AS paid_items
FROM seller_analytics_daily
WHERE seller_id = ?
  AND metric_date >= ?
  AND metric_date <= ?
  AND currency = ?`

	var summary domain.SellerAnalyticsSummary
	summary.Currency = query.Currency
	err := r.db.QueryRowContext(ctx, statement,
		query.SellerID,
		domain.FormatAnalyticsDate(query.From),
		domain.FormatAnalyticsDate(query.To),
		query.Currency,
	).Scan(&summary.RevenueAmount, &summary.PaidOrders, &summary.PaidItems)
	if err != nil {
		return domain.SellerAnalyticsSummary{}, err
	}
	return summary, nil
}

func (r *MySQLSellerAnalyticsRepository) GetSellerConversion(ctx context.Context, query domain.AnalyticsQuery) (domain.SellerConversionSummary, error) {
	query = query.Normalized()
	const statement = `
SELECT
  COALESCE(SUM(product_view_sessions), 0) AS product_view_sessions,
  COALESCE(SUM(add_to_cart_sessions), 0) AS add_to_cart_sessions,
  COALESCE(SUM(checkout_started_sessions), 0) AS checkout_started_sessions,
  COALESCE(SUM(paid_order_sessions), 0) AS paid_order_sessions
FROM seller_conversion_daily
WHERE seller_id = ?
  AND metric_date >= ?
  AND metric_date <= ?`

	var summary domain.SellerConversionSummary
	err := r.db.QueryRowContext(ctx, statement,
		query.SellerID,
		domain.FormatAnalyticsDate(query.From),
		domain.FormatAnalyticsDate(query.To),
	).Scan(
		&summary.ProductViewSessions,
		&summary.AddToCartSessions,
		&summary.CheckoutStartedSessions,
		&summary.PaidOrderSessions,
	)
	if err != nil {
		return domain.SellerConversionSummary{}, err
	}
	return summary, nil
}

func (r *MySQLSellerAnalyticsRepository) ListTopProducts(ctx context.Context, query domain.AnalyticsQuery) ([]domain.TopProductMetric, error) {
	query = query.Normalized()
	limit := query.TopProductsLimit
	if limit <= 0 {
		limit = domain.AnalyticsDefaultTopProductsLimit
	}
	if limit > domain.AnalyticsMaxTopProductsLimit {
		limit = domain.AnalyticsMaxTopProductsLimit
	}

	const statement = `
SELECT
  product_id,
  COALESCE(MAX(NULLIF(title_snapshot, '')), '') AS title,
  COALESCE(SUM(units_sold), 0) AS units_sold,
  COALESCE(SUM(order_count), 0) AS orders,
  COALESCE(SUM(revenue_amount), 0) AS revenue_amount
FROM seller_product_analytics_daily
WHERE seller_id = ?
  AND metric_date >= ?
  AND metric_date <= ?
  AND currency = ?
GROUP BY product_id
ORDER BY revenue_amount DESC, units_sold DESC, product_id ASC
LIMIT ?`

	rows, err := r.db.QueryContext(ctx, statement,
		query.SellerID,
		domain.FormatAnalyticsDate(query.From),
		domain.FormatAnalyticsDate(query.To),
		query.Currency,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]domain.TopProductMetric, 0, limit)
	for rows.Next() {
		var product domain.TopProductMetric
		product.Revenue.Currency = query.Currency
		if err := rows.Scan(
			&product.ProductID,
			&product.Title,
			&product.UnitsSold,
			&product.Orders,
			&product.Revenue.Amount,
		); err != nil {
			return nil, err
		}
		products = append(products, product.Normalized(query.Currency))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *MySQLSellerAnalyticsRepository) String() string {
	return fmt.Sprintf("%T", r)
}
