-- Local/demo CMS staff access for Docker Compose only.

INSERT INTO seller_staff (
  staff_id,
  seller_id,
  user_id,
  role,
  status,
  invited_by
) VALUES (
  'staff_local_seller_owner',
  'seller_local_demo',
  'user_local_seller',
  'seller',
  'active',
  'system:local-seed'
)
ON DUPLICATE KEY UPDATE
  role = VALUES(role),
  status = VALUES(status),
  invited_by = VALUES(invited_by);
