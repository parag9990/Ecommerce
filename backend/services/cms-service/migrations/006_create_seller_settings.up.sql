CREATE TABLE IF NOT EXISTS seller_settings (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  seller_id VARCHAR(64) NOT NULL,
  return_policy TEXT NULL,
  shipping_policy TEXT NULL,
  support_email VARCHAR(255) NULL,
  settings_json JSON NULL,
  created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uk_seller_settings_seller_id (seller_id),
  CONSTRAINT chk_seller_settings_json
    CHECK (settings_json IS NULL OR JSON_VALID(settings_json))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
