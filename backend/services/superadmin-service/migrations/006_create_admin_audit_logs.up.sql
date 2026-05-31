CREATE TABLE IF NOT EXISTS admin_audit_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  audit_id VARCHAR(64) NOT NULL,
  actor_admin_id VARCHAR(64) NOT NULL,
  action VARCHAR(128) NOT NULL,
  resource_type VARCHAR(64) NOT NULL,
  resource_id VARCHAR(128) NOT NULL,
  request_id VARCHAR(64) NOT NULL,
  session_id VARCHAR(64) NULL,
  ip_hash VARCHAR(128) NULL,
  reason VARCHAR(512) NOT NULL,
  before_json JSON NULL,
  after_json JSON NULL,
  metadata_json JSON NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_audit_logs_audit_id (audit_id),
  KEY idx_admin_audit_actor_created (actor_admin_id, created_at),
  KEY idx_admin_audit_resource (resource_type, resource_id),
  KEY idx_admin_audit_action_created (action, created_at),
  KEY idx_admin_audit_request (request_id),
  KEY idx_admin_audit_created (created_at),
  CONSTRAINT fk_admin_audit_logs_actor
    FOREIGN KEY (actor_admin_id)
    REFERENCES admin_users(admin_id)
    ON UPDATE CASCADE
    ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
