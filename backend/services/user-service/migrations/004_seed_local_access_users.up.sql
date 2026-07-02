-- Local/demo users for Docker Compose access only.

USE user_db;

INSERT INTO users (
  user_id,
  auth_account_id,
  email,
  phone,
  full_name,
  avatar_url,
  status,
  created_by,
  updated_by,
  status_changed_by,
  status_changed_at
) VALUES
  (
    'user_local_seller',
    'auth_local_seller',
    'seller.local@example.com',
    NULL,
    'Local Demo Seller',
    NULL,
    'active',
    'system:local-seed',
    'system:local-seed',
    'system:local-seed',
    CURRENT_TIMESTAMP
  ),
  (
    'user_local_analytics_admin',
    'auth_local_analytics_admin',
    'analytics.admin.local@example.com',
    NULL,
    'Local Analytics Admin',
    NULL,
    'active',
    'system:local-seed',
    'system:local-seed',
    'system:local-seed',
    CURRENT_TIMESTAMP
  ),
  (
    'user_local_superadmin',
    'auth_local_superadmin',
    'superadmin.local@example.com',
    NULL,
    'Local Superadmin',
    NULL,
    'active',
    'system:local-seed',
    'system:local-seed',
    'system:local-seed',
    CURRENT_TIMESTAMP
  )
ON DUPLICATE KEY UPDATE
  auth_account_id = VALUES(auth_account_id),
  email = VALUES(email),
  full_name = VALUES(full_name),
  status = VALUES(status),
  updated_by = VALUES(updated_by),
  status_changed_by = VALUES(status_changed_by),
  status_changed_at = VALUES(status_changed_at);

INSERT INTO seller_profiles (
  seller_id,
  user_id,
  store_name,
  display_name,
  gst_number,
  support_email,
  status,
  approved_by,
  approved_at,
  created_by,
  updated_by,
  status_changed_by,
  status_changed_at,
  status_reason
) VALUES (
  'seller_local_demo',
  'user_local_seller',
  'Local Demo Store',
  'Local Demo Store',
  '29ABCDE1234F1Z5',
  'seller.local@example.com',
  'active',
  'system:local-seed',
  CURRENT_TIMESTAMP,
  'system:local-seed',
  'system:local-seed',
  'system:local-seed',
  CURRENT_TIMESTAMP,
  'Local Docker demo seller profile'
)
ON DUPLICATE KEY UPDATE
  store_name = VALUES(store_name),
  display_name = VALUES(display_name),
  gst_number = VALUES(gst_number),
  support_email = VALUES(support_email),
  status = VALUES(status),
  approved_by = VALUES(approved_by),
  approved_at = VALUES(approved_at),
  updated_by = VALUES(updated_by),
  status_changed_by = VALUES(status_changed_by),
  status_changed_at = VALUES(status_changed_at),
  status_reason = VALUES(status_reason);
