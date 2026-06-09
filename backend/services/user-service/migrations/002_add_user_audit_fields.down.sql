SET NAMES utf8mb4;
SET time_zone = '+00:00';

USE user_db;

ALTER TABLE seller_kyc_documents
  DROP KEY idx_kyc_documents_status_updated,
  DROP COLUMN status_changed_at,
  DROP COLUMN status_changed_by,
  DROP COLUMN updated_at,
  DROP COLUMN updated_by,
  DROP COLUMN created_by;

ALTER TABLE seller_profiles
  DROP KEY idx_seller_profiles_status_updated,
  DROP COLUMN status_reason,
  DROP COLUMN status_changed_at,
  DROP COLUMN status_changed_by,
  DROP COLUMN updated_by,
  DROP COLUMN created_by;

ALTER TABLE user_addresses
  DROP KEY idx_user_addresses_user_status,
  DROP COLUMN deleted_by,
  DROP COLUMN updated_by,
  DROP COLUMN created_by,
  DROP COLUMN status;

ALTER TABLE users
  DROP KEY idx_users_deleted_at,
  DROP KEY idx_users_status_updated,
  DROP COLUMN deleted_at,
  DROP COLUMN deleted_by,
  DROP COLUMN status_changed_at,
  DROP COLUMN status_changed_by,
  DROP COLUMN updated_by,
  DROP COLUMN created_by;
