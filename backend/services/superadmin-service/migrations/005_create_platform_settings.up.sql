CREATE TABLE IF NOT EXISTS platform_settings (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  setting_key VARCHAR(128) NOT NULL,
  setting_type ENUM('maintenance', 'commission', 'feature_flags', 'search') NOT NULL,
  value_json JSON NOT NULL,
  risk_level ENUM('low', 'medium', 'high', 'critical') NOT NULL DEFAULT 'medium',
  version BIGINT UNSIGNED NOT NULL DEFAULT 1,
  updated_by_admin_id VARCHAR(64) NOT NULL,
  update_reason VARCHAR(512) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_platform_settings_key (setting_key),
  KEY idx_platform_settings_type (setting_type),
  KEY idx_platform_settings_updated (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO platform_settings
  (setting_key, setting_type, value_json, risk_level, version, updated_by_admin_id, update_reason)
VALUES
  ('maintenance_mode', 'maintenance', JSON_OBJECT(
    'enabled', false,
    'message', '',
    'allow_admins', true,
    'allow_health_checks', true
  ), 'critical', 1, 'system', 'initial default'),
  ('commission_rules', 'commission', JSON_OBJECT(
    'default_rate_bps', 1000,
    'currency', 'INR',
    'category_overrides', JSON_ARRAY(),
    'seller_overrides', JSON_ARRAY(),
    'effective_from', '2026-01-01T00:00:00Z'
  ), 'critical', 1, 'system', 'initial default'),
  ('feature_flags', 'feature_flags', JSON_OBJECT(
    'flags', JSON_OBJECT(
      'new_checkout', JSON_OBJECT(
        'enabled', false,
        'rollout_percent', 0,
        'allowed_roles', JSON_ARRAY('buyer'),
        'description', 'New checkout experience'
      )
    )
  ), 'high', 1, 'system', 'initial default'),
  ('search_synonyms', 'search', JSON_OBJECT(
    'synonyms', JSON_ARRAY()
  ), 'medium', 1, 'system', 'initial default')
ON DUPLICATE KEY UPDATE setting_key = setting_key;
