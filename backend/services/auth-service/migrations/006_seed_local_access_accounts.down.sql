USE auth_db;

DELETE FROM role_assignments
WHERE account_id IN (
  'auth_local_seller',
  'auth_local_analytics_admin',
  'auth_local_superadmin'
);

DELETE FROM refresh_tokens
WHERE account_id IN (
  'auth_local_seller',
  'auth_local_analytics_admin',
  'auth_local_superadmin'
);

DELETE FROM otp_challenges
WHERE account_id IN (
  'auth_local_seller',
  'auth_local_analytics_admin',
  'auth_local_superadmin'
);

DELETE FROM auth_outbox_events
WHERE aggregate_type = 'auth_account'
  AND aggregate_id IN (
    'auth_local_seller',
    'auth_local_analytics_admin',
    'auth_local_superadmin'
  );

DELETE FROM credentials
WHERE account_id IN (
  'auth_local_seller',
  'auth_local_analytics_admin',
  'auth_local_superadmin'
);

DELETE FROM auth_accounts
WHERE account_id IN (
  'auth_local_seller',
  'auth_local_analytics_admin',
  'auth_local_superadmin'
);
