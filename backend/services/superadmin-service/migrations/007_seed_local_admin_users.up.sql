-- Local/demo admin users for Docker Compose only.
-- admin_id intentionally matches the JWT subject user_id forwarded by the local gateway.

INSERT INTO admin_users (
  admin_id,
  user_id,
  role,
  status,
  mfa_required,
  created_by
) VALUES
  (
    'user_local_analytics_admin',
    'user_local_analytics_admin',
    'operations_admin',
    'active',
    FALSE,
    'system:local-seed'
  ),
  (
    'user_local_superadmin',
    'user_local_superadmin',
    'superadmin',
    'active',
    FALSE,
    'system:local-seed'
  )
ON DUPLICATE KEY UPDATE
  role = VALUES(role),
  status = VALUES(status),
  mfa_required = VALUES(mfa_required),
  created_by = VALUES(created_by);
