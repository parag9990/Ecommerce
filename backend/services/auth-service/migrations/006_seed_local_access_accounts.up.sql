-- Local/demo credentials for Docker Compose access only.
-- Password for all accounts: LocalDemo#2026!

USE auth_db;

INSERT INTO auth_accounts (
  account_id,
  user_id,
  email,
  phone,
  seller_id,
  tenant_id,
  email_verified,
  phone_verified,
  status
) VALUES
  (
    'auth_local_seller',
    'user_local_seller',
    'seller.local@example.com',
    NULL,
    'seller_local_demo',
    'tenant_local_demo',
    TRUE,
    FALSE,
    'active'
  ),
  (
    'auth_local_analytics_admin',
    'user_local_analytics_admin',
    'analytics.admin.local@example.com',
    NULL,
    NULL,
    'tenant_local_demo',
    TRUE,
    FALSE,
    'active'
  ),
  (
    'auth_local_superadmin',
    'user_local_superadmin',
    'superadmin.local@example.com',
    NULL,
    NULL,
    'tenant_local_demo',
    TRUE,
    FALSE,
    'active'
  )
ON DUPLICATE KEY UPDATE
  user_id = VALUES(user_id),
  seller_id = VALUES(seller_id),
  tenant_id = VALUES(tenant_id),
  email_verified = VALUES(email_verified),
  phone_verified = VALUES(phone_verified),
  status = VALUES(status),
  updated_at = CURRENT_TIMESTAMP;

INSERT INTO credentials (
  account_id,
  password_hash,
  password_algo,
  password_changed_at,
  failed_attempts,
  locked_until
) VALUES
  (
    'auth_local_seller',
    '$argon2id$v=19$m=65536,t=3,p=2$bPyNfwh1vJc2WMljYPMxnA$OsCdbuoPJR4IUZzmHSGM88L+AiX5NOiquQyRVPL604Y',
    'argon2id',
    CURRENT_TIMESTAMP,
    0,
    NULL
  ),
  (
    'auth_local_analytics_admin',
    '$argon2id$v=19$m=65536,t=3,p=2$6ZaTnIjjtebwTMMV/JrFCA$RqLz1bHTW0BKtLAIcUqb8wiWBIccTR7oU1NAvgZsHxg',
    'argon2id',
    CURRENT_TIMESTAMP,
    0,
    NULL
  ),
  (
    'auth_local_superadmin',
    '$argon2id$v=19$m=65536,t=3,p=2$/xvfeRjIrVvcf1zrI8GPpA$fJyupup2Dp2BsSn6g9jvl6TT03PRRfUolmIytJy3ie8',
    'argon2id',
    CURRENT_TIMESTAMP,
    0,
    NULL
  )
ON DUPLICATE KEY UPDATE
  password_hash = VALUES(password_hash),
  password_algo = VALUES(password_algo),
  password_changed_at = VALUES(password_changed_at),
  failed_attempts = 0,
  locked_until = NULL;

INSERT INTO role_assignments (
  account_id,
  role,
  scope_type,
  scope_id,
  assigned_by,
  reason
) VALUES
  (
    'auth_local_seller',
    'seller',
    'seller',
    'seller_local_demo',
    'system:local-seed',
    'Local Docker demo seller dashboard access'
  ),
  (
    'auth_local_analytics_admin',
    'operations_admin',
    NULL,
    NULL,
    'system:local-seed',
    'Local Docker demo session analytics access'
  ),
  (
    'auth_local_superadmin',
    'superadmin',
    NULL,
    NULL,
    'system:local-seed',
    'Local Docker demo superadmin panel access'
  )
ON DUPLICATE KEY UPDATE
  assigned_by = VALUES(assigned_by),
  reason = VALUES(reason),
  revoked_at = NULL;
