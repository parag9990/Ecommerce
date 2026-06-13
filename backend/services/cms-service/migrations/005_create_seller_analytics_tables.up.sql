CREATE TABLE IF NOT EXISTS seller_analytics_daily (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  seller_id VARCHAR(64) NOT NULL,
  metric_date DATE NOT NULL,
  currency CHAR(3) NOT NULL DEFAULT 'INR',
  gross_revenue_amount BIGINT NOT NULL DEFAULT 0,
  refund_amount BIGINT NOT NULL DEFAULT 0,
  net_revenue_amount BIGINT NOT NULL DEFAULT 0,
  paid_order_count BIGINT NOT NULL DEFAULT 0,
  paid_item_count BIGINT NOT NULL DEFAULT 0,
  cancelled_order_count BIGINT NOT NULL DEFAULT 0,
  updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uk_seller_analytics_day_currency (seller_id, metric_date, currency),
  KEY idx_seller_analytics_seller_date (seller_id, metric_date),
  KEY idx_seller_analytics_seller_currency_date (seller_id, currency, metric_date),
  CONSTRAINT chk_seller_analytics_currency
    CHECK (currency REGEXP '^[A-Z]{3}$'),
  CONSTRAINT chk_seller_analytics_amounts
    CHECK (
      gross_revenue_amount >= 0
      AND refund_amount >= 0
      AND net_revenue_amount >= 0
      AND paid_order_count >= 0
      AND paid_item_count >= 0
      AND cancelled_order_count >= 0
    )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS seller_product_analytics_daily (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  seller_id VARCHAR(64) NOT NULL,
  product_id VARCHAR(64) NOT NULL,
  metric_date DATE NOT NULL,
  currency CHAR(3) NOT NULL DEFAULT 'INR',
  title_snapshot VARCHAR(512) NULL,
  units_sold BIGINT NOT NULL DEFAULT 0,
  order_count BIGINT NOT NULL DEFAULT 0,
  revenue_amount BIGINT NOT NULL DEFAULT 0,
  updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uk_seller_product_day_currency (seller_id, product_id, metric_date, currency),
  KEY idx_seller_product_seller_date (seller_id, metric_date),
  KEY idx_seller_product_top_revenue (seller_id, currency, metric_date, revenue_amount),
  KEY idx_seller_product_top_units (seller_id, currency, metric_date, units_sold),
  CONSTRAINT chk_seller_product_analytics_currency
    CHECK (currency REGEXP '^[A-Z]{3}$'),
  CONSTRAINT chk_seller_product_analytics_counts
    CHECK (units_sold >= 0 AND order_count >= 0 AND revenue_amount >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS seller_conversion_daily (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  seller_id VARCHAR(64) NOT NULL,
  metric_date DATE NOT NULL,
  product_view_sessions BIGINT NOT NULL DEFAULT 0,
  add_to_cart_sessions BIGINT NOT NULL DEFAULT 0,
  checkout_started_sessions BIGINT NOT NULL DEFAULT 0,
  paid_order_sessions BIGINT NOT NULL DEFAULT 0,
  updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uk_seller_conversion_day (seller_id, metric_date),
  KEY idx_seller_conversion_seller_date (seller_id, metric_date),
  CONSTRAINT chk_seller_conversion_counts
    CHECK (
      product_view_sessions >= 0
      AND add_to_cart_sessions >= 0
      AND checkout_started_sessions >= 0
      AND paid_order_sessions >= 0
    )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS seller_analytics_event_dedupe (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  event_id VARCHAR(128) NOT NULL,
  event_type VARCHAR(128) NOT NULL,
  source_service VARCHAR(64) NOT NULL,
  processed_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uk_seller_analytics_event_id (event_id),
  KEY idx_seller_analytics_event_source (source_service, event_type, processed_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
