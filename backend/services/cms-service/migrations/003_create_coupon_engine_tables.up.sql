CREATE TABLE IF NOT EXISTS coupons (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  coupon_id VARCHAR(64) NOT NULL,
  seller_id VARCHAR(64) NULL,
  code VARCHAR(64) NOT NULL,
  discount_type ENUM('fixed', 'percentage') NOT NULL,
  discount_value BIGINT NOT NULL,
  max_discount_amount BIGINT NULL,
  min_cart_amount BIGINT NOT NULL DEFAULT 0,
  currency CHAR(3) NOT NULL DEFAULT 'INR',
  usage_limit BIGINT NULL,
  per_user_limit BIGINT NULL,
  status ENUM('draft', 'active', 'paused', 'expired') NOT NULL DEFAULT 'draft',
  starts_at TIMESTAMP(6) NULL,
  ends_at TIMESTAMP(6) NULL,
  created_by VARCHAR(64) NULL,
  created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uk_coupons_coupon_id (coupon_id),
  UNIQUE KEY uk_coupons_code (code),
  KEY idx_coupons_seller_status (seller_id, status, created_at),
  KEY idx_coupons_status_window (status, starts_at, ends_at),
  CONSTRAINT chk_coupons_discount_value
    CHECK (discount_value > 0),
  CONSTRAINT chk_coupons_percentage_value
    CHECK (discount_type <> 'percentage' OR discount_value BETWEEN 1 AND 100),
  CONSTRAINT chk_coupons_money_amounts
    CHECK (
      min_cart_amount >= 0
      AND (max_discount_amount IS NULL OR max_discount_amount > 0)
    ),
  CONSTRAINT chk_coupons_usage_limits
    CHECK (
      (usage_limit IS NULL OR usage_limit > 0)
      AND (per_user_limit IS NULL OR per_user_limit > 0)
    ),
  CONSTRAINT chk_coupons_time_window
    CHECK (starts_at IS NULL OR ends_at IS NULL OR starts_at < ends_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS coupon_rules (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  rule_id VARCHAR(64) NOT NULL,
  coupon_id VARCHAR(64) NOT NULL,
  rule_type ENUM('product_scope', 'category_scope', 'seller_scope') NOT NULL,
  rule_value JSON NOT NULL,
  created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uk_coupon_rules_rule_id (rule_id),
  KEY idx_coupon_rules_coupon (coupon_id, rule_type),
  CONSTRAINT fk_coupon_rules_coupon
    FOREIGN KEY (coupon_id) REFERENCES coupons (coupon_id)
    ON DELETE CASCADE,
  CONSTRAINT chk_coupon_rules_value_json
    CHECK (JSON_VALID(rule_value))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS coupon_redemptions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  redemption_id VARCHAR(64) NOT NULL,
  coupon_id VARCHAR(64) NOT NULL,
  order_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  discount_amount BIGINT NOT NULL,
  currency CHAR(3) NOT NULL,
  request_id VARCHAR(128) NULL,
  redeemed_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uk_coupon_redemptions_redemption_id (redemption_id),
  UNIQUE KEY uk_coupon_redemptions_coupon_order (coupon_id, order_id),
  KEY idx_coupon_redemptions_user_coupon (user_id, coupon_id),
  KEY idx_coupon_redemptions_coupon_created (coupon_id, created_at),
  KEY idx_coupon_redemptions_order (order_id),
  CONSTRAINT fk_coupon_redemptions_coupon
    FOREIGN KEY (coupon_id) REFERENCES coupons (coupon_id)
    ON DELETE RESTRICT,
  CONSTRAINT chk_coupon_redemptions_discount
    CHECK (discount_amount > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
