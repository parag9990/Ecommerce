SET NAMES utf8mb4;
SET time_zone = '+00:00';

USE user_db;

ALTER TABLE users
  ADD COLUMN created_by VARCHAR(64) NULL AFTER status,
  ADD COLUMN updated_by VARCHAR(64) NULL AFTER created_by,
  ADD COLUMN status_changed_by VARCHAR(64) NULL AFTER updated_by,
  ADD COLUMN status_changed_at TIMESTAMP NULL AFTER status_changed_by,
  ADD COLUMN deleted_by VARCHAR(64) NULL AFTER status_changed_at,
  ADD COLUMN deleted_at TIMESTAMP NULL AFTER deleted_by,
  ADD KEY idx_users_status_updated (status, updated_at),
  ADD KEY idx_users_deleted_at (deleted_at);

ALTER TABLE user_addresses
  ADD COLUMN status ENUM('active', 'deleted') NOT NULL DEFAULT 'active' AFTER is_default,
  ADD COLUMN created_by VARCHAR(64) NULL AFTER status,
  ADD COLUMN updated_by VARCHAR(64) NULL AFTER created_by,
  ADD COLUMN deleted_by VARCHAR(64) NULL AFTER updated_by,
  ADD KEY idx_user_addresses_user_status (user_id, status, updated_at);

ALTER TABLE seller_profiles
  ADD COLUMN created_by VARCHAR(64) NULL AFTER approved_at,
  ADD COLUMN updated_by VARCHAR(64) NULL AFTER created_by,
  ADD COLUMN status_changed_by VARCHAR(64) NULL AFTER updated_by,
  ADD COLUMN status_changed_at TIMESTAMP NULL AFTER status_changed_by,
  ADD COLUMN status_reason VARCHAR(512) NULL AFTER status_changed_at,
  ADD KEY idx_seller_profiles_status_updated (status, updated_at);

ALTER TABLE seller_kyc_documents
  ADD COLUMN created_by VARCHAR(64) NULL AFTER rejection_reason,
  ADD COLUMN updated_by VARCHAR(64) NULL AFTER created_by,
  ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP AFTER created_at,
  ADD COLUMN status_changed_by VARCHAR(64) NULL AFTER updated_at,
  ADD COLUMN status_changed_at TIMESTAMP NULL AFTER status_changed_by,
  ADD KEY idx_kyc_documents_status_updated (status, updated_at);

UPDATE users
SET created_by = COALESCE(created_by, 'system:backfill'),
    updated_by = COALESCE(updated_by, 'system:backfill'),
    status_changed_by = COALESCE(status_changed_by, 'system:backfill'),
    status_changed_at = COALESCE(status_changed_at, updated_at),
    deleted_by = CASE
      WHEN status = 'deleted' THEN COALESCE(deleted_by, 'system:backfill')
      ELSE deleted_by
    END,
    deleted_at = CASE
      WHEN status = 'deleted' THEN COALESCE(deleted_at, updated_at)
      ELSE deleted_at
    END
WHERE created_by IS NULL
   OR updated_by IS NULL
   OR status_changed_by IS NULL
   OR status_changed_at IS NULL
   OR (status = 'deleted' AND (deleted_by IS NULL OR deleted_at IS NULL));

UPDATE user_addresses
SET created_by = COALESCE(created_by, 'system:backfill'),
    updated_by = COALESCE(updated_by, 'system:backfill'),
    status = CASE
      WHEN deleted_at IS NULL THEN 'active'
      ELSE 'deleted'
    END,
    deleted_by = CASE
      WHEN deleted_at IS NULL THEN deleted_by
      ELSE COALESCE(deleted_by, 'system:backfill')
    END
WHERE created_by IS NULL
   OR updated_by IS NULL
   OR deleted_by IS NULL;

UPDATE seller_profiles
SET created_by = COALESCE(created_by, 'system:backfill'),
    updated_by = COALESCE(updated_by, 'system:backfill'),
    status_changed_by = COALESCE(status_changed_by, 'system:backfill'),
    status_changed_at = COALESCE(status_changed_at, updated_at)
WHERE created_by IS NULL
   OR updated_by IS NULL
   OR status_changed_by IS NULL
   OR status_changed_at IS NULL;

UPDATE seller_kyc_documents
SET created_by = COALESCE(created_by, 'system:backfill'),
    updated_by = COALESCE(updated_by, 'system:backfill'),
    status_changed_by = COALESCE(status_changed_by, 'system:backfill'),
    status_changed_at = COALESCE(status_changed_at, created_at)
WHERE created_by IS NULL
   OR updated_by IS NULL
   OR status_changed_by IS NULL
   OR status_changed_at IS NULL;
