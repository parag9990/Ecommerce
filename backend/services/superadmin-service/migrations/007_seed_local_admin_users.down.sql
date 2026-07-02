DELETE FROM admin_audit_logs
WHERE actor_admin_id IN (
  'user_local_analytics_admin',
  'user_local_superadmin'
);

DELETE FROM admin_users
WHERE admin_id IN (
  'user_local_analytics_admin',
  'user_local_superadmin'
);
