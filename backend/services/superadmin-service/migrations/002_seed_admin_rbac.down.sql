DELETE FROM admin_role_permissions
WHERE role IN (
  'superadmin',
  'operations_admin',
  'finance_admin',
  'catalog_admin',
  'readonly_admin'
);

DELETE FROM admin_permissions
WHERE permission_key IN (
  'admin:users:read',
  'admin:users:write',
  'users:read',
  'users:status:update',
  'sellers:read',
  'sellers:status:update',
  'orders:read',
  'orders:manual_review:write',
  'payments:read',
  'payments:refund:review',
  'sessions:read',
  'sessions:risk:read',
  'catalog:moderation:read',
  'catalog:moderation:write',
  'search:synonyms:read',
  'search:synonyms:write',
  'platform:settings:read',
  'platform:settings:write',
  'audit:logs:read',
  'audit:logs:export'
);
