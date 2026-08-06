-- Local/demo buyer user for Docker Compose user app access only.

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
) VALUES (
  'user_local_buyer',
  'auth_local_buyer',
  'buyer.local@example.com',
  NULL,
  'Local Demo Buyer',
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
