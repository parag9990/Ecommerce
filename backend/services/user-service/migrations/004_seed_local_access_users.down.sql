USE user_db;

DELETE FROM seller_profiles
WHERE seller_id = 'seller_local_demo';

DELETE FROM users
WHERE user_id IN (
  'user_local_seller',
  'user_local_analytics_admin',
  'user_local_superadmin'
);
