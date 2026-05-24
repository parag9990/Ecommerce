CREATE TABLE IF NOT EXISTS campaigns (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  campaign_id VARCHAR(64) NOT NULL,
  seller_id VARCHAR(64) NULL,
  name VARCHAR(255) NOT NULL,
  status ENUM('draft', 'active', 'paused', 'completed') NOT NULL DEFAULT 'draft',
  budget_amount BIGINT NULL,
  currency CHAR(3) NOT NULL DEFAULT 'INR',
  starts_at TIMESTAMP(6) NOT NULL,
  ends_at TIMESTAMP(6) NOT NULL,
  metadata JSON NULL,
  created_by VARCHAR(64) NOT NULL,
  created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uk_campaigns_campaign_id (campaign_id),
  KEY idx_campaigns_seller_window (seller_id, starts_at, ends_at),
  KEY idx_campaigns_seller_status (seller_id, status, starts_at),
  KEY idx_campaigns_status_window (status, starts_at, ends_at),
  CONSTRAINT chk_campaigns_budget
    CHECK (budget_amount IS NULL OR budget_amount > 0),
  CONSTRAINT chk_campaigns_currency
    CHECK (currency REGEXP '^[A-Z]{3}$'),
  CONSTRAINT chk_campaigns_window
    CHECK (starts_at < ends_at),
  CONSTRAINT chk_campaigns_metadata_json
    CHECK (metadata IS NULL OR JSON_VALID(metadata))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE coupon_redemptions
  ADD COLUMN campaign_id VARCHAR(64) NULL AFTER coupon_id,
  ADD KEY idx_coupon_redemptions_campaign_created (campaign_id, created_at),
  ADD KEY idx_coupon_redemptions_campaign_user (campaign_id, user_id),
  ADD CONSTRAINT fk_coupon_redemptions_campaign
    FOREIGN KEY (campaign_id) REFERENCES campaigns (campaign_id)
    ON DELETE RESTRICT;
