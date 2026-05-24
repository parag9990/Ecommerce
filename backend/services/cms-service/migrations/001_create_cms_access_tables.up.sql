CREATE TABLE IF NOT EXISTS seller_staff (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  staff_id VARCHAR(64) NOT NULL,
  seller_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  role ENUM('seller', 'seller_manager', 'seller_catalog_editor', 'seller_order_manager') NOT NULL,
  status ENUM('invited', 'active', 'disabled', 'revoked') NOT NULL DEFAULT 'invited',
  invited_by VARCHAR(64) NULL,
  created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uk_seller_staff_id (staff_id),
  UNIQUE KEY uk_seller_staff_seller_user (seller_id, user_id),
  KEY idx_seller_staff_seller_status (seller_id, status),
  KEY idx_seller_staff_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS cms_audit_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  audit_id VARCHAR(64) NOT NULL,
  seller_id VARCHAR(64) NOT NULL,
  actor_user_id VARCHAR(64) NOT NULL,
  action VARCHAR(128) NOT NULL,
  resource_type VARCHAR(64) NOT NULL,
  resource_id VARCHAR(128) NOT NULL,
  before_json JSON NULL,
  after_json JSON NULL,
  created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uk_cms_audit_id (audit_id),
  KEY idx_cms_audit_seller_created (seller_id, created_at),
  KEY idx_cms_audit_actor_created (actor_user_id, created_at),
  KEY idx_cms_audit_resource (resource_type, resource_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
