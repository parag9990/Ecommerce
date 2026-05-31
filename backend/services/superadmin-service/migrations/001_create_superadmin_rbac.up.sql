CREATE TABLE IF NOT EXISTS admin_users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  admin_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  role VARCHAR(64) NOT NULL,
  status ENUM('active', 'disabled') NOT NULL DEFAULT 'active',
  mfa_required BOOLEAN NOT NULL DEFAULT TRUE,
  created_by VARCHAR(64) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_users_admin_id (admin_id),
  UNIQUE KEY uk_admin_users_user_id (user_id),
  KEY idx_admin_users_role_status (role, status),
  KEY idx_admin_users_status_updated (status, updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS admin_permissions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  permission_id VARCHAR(64) NOT NULL,
  permission_key VARCHAR(128) NOT NULL,
  description VARCHAR(512) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_permissions_id (permission_id),
  UNIQUE KEY uk_admin_permissions_key (permission_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS admin_role_permissions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  role VARCHAR(64) NOT NULL,
  permission_key VARCHAR(128) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_role_permission (role, permission_key),
  KEY idx_admin_role_permissions_permission (permission_key),
  CONSTRAINT fk_admin_role_permissions_permission
    FOREIGN KEY (permission_key)
    REFERENCES admin_permissions(permission_key)
    ON UPDATE CASCADE
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
