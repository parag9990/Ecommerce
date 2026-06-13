CREATE TABLE IF NOT EXISTS product_moderation_reviews (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  review_id VARCHAR(64) NOT NULL,
  product_id VARCHAR(64) NOT NULL,
  seller_id VARCHAR(64) NOT NULL,
  status ENUM('submitted', 'approved', 'rejected', 'cancelled') NOT NULL,
  submitted_by VARCHAR(64) NOT NULL,
  reviewed_by VARCHAR(64) NULL,
  rejection_reason VARCHAR(512) NULL,
  submitted_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  reviewed_at TIMESTAMP(6) NULL,
  created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  active_submitted_product_id VARCHAR(64)
    GENERATED ALWAYS AS (
      CASE WHEN status = 'submitted' THEN product_id ELSE NULL END
    ) STORED,
  PRIMARY KEY (id),
  UNIQUE KEY uk_product_moderation_review_id (review_id),
  UNIQUE KEY uk_product_moderation_active_submitted (active_submitted_product_id),
  KEY idx_product_moderation_seller_status (seller_id, status, submitted_at),
  KEY idx_product_moderation_product (product_id),
  KEY idx_product_moderation_status_submitted (status, submitted_at),
  CONSTRAINT chk_product_moderation_rejection_reason
    CHECK (status <> 'rejected' OR rejection_reason IS NOT NULL)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
