-- Scalable E-Commerce Platform - Relational Database Design
-- Target: MySQL 8+
-- Rule: each microservice owns its database. No cross-service foreign keys.

SET NAMES utf8mb4;
SET time_zone = '+00:00';

-- =========================================================
-- Auth Service Database
-- =========================================================

CREATE DATABASE IF NOT EXISTS auth_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE auth_db;

CREATE TABLE IF NOT EXISTS auth_accounts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  account_id VARCHAR(64) NOT NULL,
  email VARCHAR(255) NULL,
  phone VARCHAR(32) NULL,
  email_verified BOOLEAN NOT NULL DEFAULT FALSE,
  phone_verified BOOLEAN NOT NULL DEFAULT FALSE,
  status ENUM('active', 'blocked', 'deleted') NOT NULL DEFAULT 'active',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_auth_accounts_account_id (account_id),
  UNIQUE KEY uk_auth_accounts_email (email),
  UNIQUE KEY uk_auth_accounts_phone (phone),
  KEY idx_auth_accounts_status (status)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS credentials (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  account_id VARCHAR(64) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  password_algo VARCHAR(32) NOT NULL DEFAULT 'argon2id',
  password_changed_at TIMESTAMP NULL,
  failed_attempts INT NOT NULL DEFAULT 0,
  locked_until TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_credentials_account_id (account_id),
  CONSTRAINT fk_credentials_account FOREIGN KEY (account_id) REFERENCES auth_accounts(account_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS refresh_tokens (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  token_id VARCHAR(64) NOT NULL,
  account_id VARCHAR(64) NOT NULL,
  session_id VARCHAR(64) NOT NULL,
  token_hash CHAR(64) NOT NULL,
  parent_token_id VARCHAR(64) NULL,
  revoked_at TIMESTAMP NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_refresh_tokens_token_id (token_id),
  UNIQUE KEY uk_refresh_tokens_hash (token_hash),
  KEY idx_refresh_tokens_account_session (account_id, session_id),
  KEY idx_refresh_tokens_expires_at (expires_at),
  CONSTRAINT fk_refresh_tokens_account FOREIGN KEY (account_id) REFERENCES auth_accounts(account_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS otp_challenges (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  challenge_id VARCHAR(64) NOT NULL,
  account_id VARCHAR(64) NULL,
  target VARCHAR(255) NOT NULL,
  channel ENUM('email', 'phone') NOT NULL,
  purpose ENUM('signup', 'login', 'password_reset', 'phone_verify', 'email_verify') NOT NULL,
  otp_hash CHAR(64) NOT NULL,
  attempts INT NOT NULL DEFAULT 0,
  max_attempts INT NOT NULL DEFAULT 5,
  verified_at TIMESTAMP NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_otp_challenges_challenge_id (challenge_id),
  KEY idx_otp_target_purpose_expiry (target, purpose, expires_at),
  KEY idx_otp_account (account_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS role_assignments (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  account_id VARCHAR(64) NOT NULL,
  role VARCHAR(64) NOT NULL,
  scope_type VARCHAR(64) NULL,
  scope_id VARCHAR(64) NULL,
  assigned_by VARCHAR(64) NULL,
  assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  revoked_at TIMESTAMP NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_role_active (account_id, role, scope_type, scope_id, revoked_at),
  KEY idx_role_assignments_account (account_id),
  KEY idx_role_assignments_role (role),
  CONSTRAINT fk_role_assignments_account FOREIGN KEY (account_id) REFERENCES auth_accounts(account_id)
) ENGINE=InnoDB;

-- =========================================================
-- User Service Database
-- =========================================================

CREATE DATABASE IF NOT EXISTS user_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE user_db;

CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id VARCHAR(64) NOT NULL,
  auth_account_id VARCHAR(64) NOT NULL,
  email VARCHAR(255) NOT NULL,
  phone VARCHAR(32) NULL,
  full_name VARCHAR(255) NOT NULL,
  avatar_url VARCHAR(1024) NULL,
  status ENUM('active', 'blocked', 'deleted') NOT NULL DEFAULT 'active',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_users_user_id (user_id),
  UNIQUE KEY uk_users_auth_account_id (auth_account_id),
  UNIQUE KEY uk_users_email (email),
  KEY idx_users_status (status)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS user_addresses (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  address_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  name VARCHAR(255) NOT NULL,
  phone VARCHAR(32) NULL,
  line1 VARCHAR(255) NOT NULL,
  line2 VARCHAR(255) NULL,
  city VARCHAR(128) NOT NULL,
  state VARCHAR(128) NOT NULL,
  postal_code VARCHAR(32) NOT NULL,
  country VARCHAR(64) NOT NULL,
  is_default BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_user_addresses_address_id (address_id),
  KEY idx_user_addresses_user_default (user_id, is_default),
  CONSTRAINT fk_user_addresses_user FOREIGN KEY (user_id) REFERENCES users(user_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS seller_profiles (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  seller_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  store_name VARCHAR(255) NOT NULL,
  display_name VARCHAR(255) NULL,
  gst_number VARCHAR(64) NULL,
  support_email VARCHAR(255) NULL,
  status ENUM('draft', 'pending_review', 'active', 'suspended', 'rejected') NOT NULL DEFAULT 'draft',
  approved_by VARCHAR(64) NULL,
  approved_at TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_seller_profiles_seller_id (seller_id),
  UNIQUE KEY uk_seller_profiles_user_id (user_id),
  KEY idx_seller_profiles_status (status),
  CONSTRAINT fk_seller_profiles_user FOREIGN KEY (user_id) REFERENCES users(user_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS seller_kyc_documents (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  document_id VARCHAR(64) NOT NULL,
  seller_id VARCHAR(64) NOT NULL,
  document_type VARCHAR(64) NOT NULL,
  storage_url VARCHAR(1024) NOT NULL,
  status ENUM('pending', 'approved', 'rejected') NOT NULL DEFAULT 'pending',
  reviewed_by VARCHAR(64) NULL,
  reviewed_at TIMESTAMP NULL,
  rejection_reason VARCHAR(512) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_kyc_document_id (document_id),
  KEY idx_kyc_seller_status (seller_id, status),
  CONSTRAINT fk_kyc_seller FOREIGN KEY (seller_id) REFERENCES seller_profiles(seller_id)
) ENGINE=InnoDB;

-- =========================================================
-- Order Service Database
-- =========================================================

CREATE DATABASE IF NOT EXISTS order_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE order_db;

CREATE TABLE IF NOT EXISTS orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  order_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  cart_id VARCHAR(64) NULL,
  status ENUM('created', 'pending_payment', 'paid', 'packed', 'shipped', 'delivered', 'cancelled', 'refunded', 'payment_failed') NOT NULL DEFAULT 'created',
  currency CHAR(3) NOT NULL,
  subtotal_amount BIGINT NOT NULL DEFAULT 0,
  discount_amount BIGINT NOT NULL DEFAULT 0,
  shipping_amount BIGINT NOT NULL DEFAULT 0,
  tax_amount BIGINT NOT NULL DEFAULT 0,
  total_amount BIGINT NOT NULL DEFAULT 0,
  address_snapshot JSON NOT NULL,
  coupon_code VARCHAR(64) NULL,
  payment_id VARCHAR(64) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_orders_order_id (order_id),
  KEY idx_orders_user_created (user_id, created_at),
  KEY idx_orders_status_created (status, created_at),
  KEY idx_orders_payment_id (payment_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS order_items (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  order_item_id VARCHAR(64) NOT NULL,
  order_id VARCHAR(64) NOT NULL,
  seller_id VARCHAR(64) NOT NULL,
  product_id VARCHAR(64) NOT NULL,
  variant_id VARCHAR(64) NOT NULL,
  sku VARCHAR(128) NOT NULL,
  title_snapshot VARCHAR(512) NOT NULL,
  image_url_snapshot VARCHAR(1024) NULL,
  quantity INT NOT NULL,
  currency CHAR(3) NOT NULL,
  unit_amount BIGINT NOT NULL,
  discount_amount BIGINT NOT NULL DEFAULT 0,
  tax_amount BIGINT NOT NULL DEFAULT 0,
  total_amount BIGINT NOT NULL,
  fulfillment_status ENUM('pending', 'packed', 'shipped', 'delivered', 'cancelled', 'returned') NOT NULL DEFAULT 'pending',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_items_order_item_id (order_item_id),
  KEY idx_order_items_order (order_id),
  KEY idx_order_items_seller_created (seller_id, created_at),
  KEY idx_order_items_product (product_id),
  CONSTRAINT fk_order_items_order FOREIGN KEY (order_id) REFERENCES orders(order_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS order_status_history (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  history_id VARCHAR(64) NOT NULL,
  order_id VARCHAR(64) NOT NULL,
  from_status VARCHAR(64) NULL,
  to_status VARCHAR(64) NOT NULL,
  reason VARCHAR(512) NULL,
  changed_by VARCHAR(64) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_status_history_id (history_id),
  KEY idx_order_status_history_order_created (order_id, created_at),
  CONSTRAINT fk_order_status_history_order FOREIGN KEY (order_id) REFERENCES orders(order_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS shipments (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  shipment_id VARCHAR(64) NOT NULL,
  order_id VARCHAR(64) NOT NULL,
  seller_id VARCHAR(64) NOT NULL,
  carrier VARCHAR(128) NULL,
  tracking_number VARCHAR(128) NULL,
  status ENUM('pending', 'packed', 'shipped', 'delivered', 'failed') NOT NULL DEFAULT 'pending',
  shipped_at TIMESTAMP NULL,
  delivered_at TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_shipments_shipment_id (shipment_id),
  KEY idx_shipments_order (order_id),
  KEY idx_shipments_seller_status (seller_id, status),
  CONSTRAINT fk_shipments_order FOREIGN KEY (order_id) REFERENCES orders(order_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS order_idempotency_keys (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id VARCHAR(64) NOT NULL,
  idempotency_key VARCHAR(128) NOT NULL,
  order_id VARCHAR(64) NULL,
  request_hash CHAR(64) NOT NULL,
  status ENUM('processing', 'completed', 'failed') NOT NULL DEFAULT 'processing',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expires_at TIMESTAMP NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_idempotency_user_key (user_id, idempotency_key),
  KEY idx_order_idempotency_expires (expires_at)
) ENGINE=InnoDB;

-- =========================================================
-- Payment Service Database
-- =========================================================

CREATE DATABASE IF NOT EXISTS payment_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE payment_db;

CREATE TABLE IF NOT EXISTS payments (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  payment_id VARCHAR(64) NOT NULL,
  order_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  provider VARCHAR(64) NOT NULL,
  provider_payment_id VARCHAR(128) NULL,
  provider_intent_id VARCHAR(128) NULL,
  status ENUM('initiated', 'requires_action', 'authorized', 'captured', 'failed', 'refunded', 'partially_refunded') NOT NULL DEFAULT 'initiated',
  currency CHAR(3) NOT NULL,
  amount BIGINT NOT NULL,
  captured_amount BIGINT NOT NULL DEFAULT 0,
  refunded_amount BIGINT NOT NULL DEFAULT 0,
  idempotency_key VARCHAR(128) NOT NULL,
  failure_code VARCHAR(128) NULL,
  failure_message VARCHAR(512) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_payments_payment_id (payment_id),
  UNIQUE KEY uk_payments_idempotency (provider, idempotency_key),
  KEY idx_payments_order (order_id),
  KEY idx_payments_provider_payment (provider, provider_payment_id),
  KEY idx_payments_status_created (status, created_at)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS payment_attempts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  attempt_id VARCHAR(64) NOT NULL,
  payment_id VARCHAR(64) NOT NULL,
  provider_attempt_id VARCHAR(128) NULL,
  status ENUM('initiated', 'succeeded', 'failed') NOT NULL DEFAULT 'initiated',
  failure_code VARCHAR(128) NULL,
  failure_message VARCHAR(512) NULL,
  raw_provider_response JSON NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_payment_attempts_attempt_id (attempt_id),
  KEY idx_payment_attempts_payment (payment_id),
  CONSTRAINT fk_payment_attempts_payment FOREIGN KEY (payment_id) REFERENCES payments(payment_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS refunds (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  refund_id VARCHAR(64) NOT NULL,
  payment_id VARCHAR(64) NOT NULL,
  provider_refund_id VARCHAR(128) NULL,
  status ENUM('requested', 'approved', 'rejected', 'processing', 'succeeded', 'failed') NOT NULL DEFAULT 'requested',
  currency CHAR(3) NOT NULL,
  amount BIGINT NOT NULL,
  reason VARCHAR(512) NOT NULL,
  requested_by VARCHAR(64) NOT NULL,
  reviewed_by VARCHAR(64) NULL,
  reviewed_at TIMESTAMP NULL,
  idempotency_key VARCHAR(128) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_refunds_refund_id (refund_id),
  UNIQUE KEY uk_refunds_idempotency (payment_id, idempotency_key),
  KEY idx_refunds_payment_status (payment_id, status),
  CONSTRAINT fk_refunds_payment FOREIGN KEY (payment_id) REFERENCES payments(payment_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS payment_webhook_events (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  webhook_event_id VARCHAR(64) NOT NULL,
  provider VARCHAR(64) NOT NULL,
  provider_event_id VARCHAR(128) NOT NULL,
  event_type VARCHAR(128) NOT NULL,
  processed BOOLEAN NOT NULL DEFAULT FALSE,
  payload JSON NOT NULL,
  received_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  processed_at TIMESTAMP NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_webhook_provider_event (provider, provider_event_id),
  UNIQUE KEY uk_webhook_event_id (webhook_event_id),
  KEY idx_webhook_processed_received (processed, received_at)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS payment_reconciliations (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  reconciliation_id VARCHAR(64) NOT NULL,
  provider VARCHAR(64) NOT NULL,
  settlement_id VARCHAR(128) NULL,
  status ENUM('matched', 'mismatch', 'missing_local', 'missing_provider') NOT NULL,
  payment_id VARCHAR(64) NULL,
  provider_payment_id VARCHAR(128) NULL,
  details JSON NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_reconciliation_id (reconciliation_id),
  KEY idx_reconciliation_provider_status (provider, status)
) ENGINE=InnoDB;

-- =========================================================
-- CMS Service Database
-- =========================================================

CREATE DATABASE IF NOT EXISTS cms_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE cms_db;

CREATE TABLE IF NOT EXISTS seller_settings (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  seller_id VARCHAR(64) NOT NULL,
  return_policy TEXT NULL,
  shipping_policy TEXT NULL,
  support_email VARCHAR(255) NULL,
  settings_json JSON NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_seller_settings_seller_id (seller_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS seller_staff (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  staff_id VARCHAR(64) NOT NULL,
  seller_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  role VARCHAR(64) NOT NULL,
  status ENUM('invited', 'active', 'disabled') NOT NULL DEFAULT 'invited',
  invited_by VARCHAR(64) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_seller_staff_staff_id (staff_id),
  UNIQUE KEY uk_seller_staff_seller_user (seller_id, user_id),
  KEY idx_seller_staff_seller_status (seller_id, status)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS coupons (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  coupon_id VARCHAR(64) NOT NULL,
  seller_id VARCHAR(64) NULL,
  code VARCHAR(64) NOT NULL,
  discount_type ENUM('fixed', 'percentage') NOT NULL,
  discount_value BIGINT NOT NULL,
  max_discount_amount BIGINT NULL,
  min_cart_amount BIGINT NULL,
  currency CHAR(3) NOT NULL DEFAULT 'INR',
  usage_limit INT NULL,
  per_user_limit INT NULL,
  status ENUM('draft', 'active', 'paused', 'expired') NOT NULL DEFAULT 'draft',
  starts_at TIMESTAMP NULL,
  ends_at TIMESTAMP NULL,
  created_by VARCHAR(64) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_coupons_coupon_id (coupon_id),
  UNIQUE KEY uk_coupons_code (code),
  KEY idx_coupons_seller_status (seller_id, status),
  KEY idx_coupons_window (starts_at, ends_at)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS coupon_rules (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  rule_id VARCHAR(64) NOT NULL,
  coupon_id VARCHAR(64) NOT NULL,
  rule_type VARCHAR(64) NOT NULL,
  rule_value JSON NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_coupon_rules_rule_id (rule_id),
  KEY idx_coupon_rules_coupon (coupon_id),
  CONSTRAINT fk_coupon_rules_coupon FOREIGN KEY (coupon_id) REFERENCES coupons(coupon_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS coupon_redemptions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  redemption_id VARCHAR(64) NOT NULL,
  coupon_id VARCHAR(64) NOT NULL,
  order_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  discount_amount BIGINT NOT NULL,
  currency CHAR(3) NOT NULL,
  redeemed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_coupon_redemptions_redemption_id (redemption_id),
  UNIQUE KEY uk_coupon_redemptions_coupon_order (coupon_id, order_id),
  KEY idx_coupon_redemptions_user_coupon (user_id, coupon_id),
  CONSTRAINT fk_coupon_redemptions_coupon FOREIGN KEY (coupon_id) REFERENCES coupons(coupon_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS campaigns (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  campaign_id VARCHAR(64) NOT NULL,
  seller_id VARCHAR(64) NULL,
  name VARCHAR(255) NOT NULL,
  status ENUM('draft', 'active', 'paused', 'completed') NOT NULL DEFAULT 'draft',
  budget_amount BIGINT NULL,
  currency CHAR(3) NOT NULL DEFAULT 'INR',
  starts_at TIMESTAMP NOT NULL,
  ends_at TIMESTAMP NOT NULL,
  metadata JSON NULL,
  created_by VARCHAR(64) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_campaigns_campaign_id (campaign_id),
  KEY idx_campaigns_seller_window (seller_id, starts_at, ends_at),
  KEY idx_campaigns_status (status)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS cms_audit_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  audit_id VARCHAR(64) NOT NULL,
  seller_id VARCHAR(64) NOT NULL,
  actor_user_id VARCHAR(64) NOT NULL,
  actor_roles_json JSON NULL,
  action VARCHAR(128) NOT NULL,
  resource_type VARCHAR(64) NOT NULL,
  resource_id VARCHAR(64) NOT NULL,
  request_id VARCHAR(128) NULL,
  trace_id VARCHAR(128) NULL,
  decision ENUM('allowed', 'denied') NOT NULL DEFAULT 'allowed',
  reason VARCHAR(512) NULL,
  ip_hash CHAR(64) NULL,
  user_agent_hash CHAR(64) NULL,
  before_json JSON NULL,
  after_json JSON NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_cms_audit_logs_audit_id (audit_id),
  KEY idx_cms_audit_seller_created (seller_id, created_at),
  KEY idx_cms_audit_actor_created (actor_user_id, created_at),
  KEY idx_cms_audit_resource (resource_type, resource_id),
  KEY idx_cms_audit_request (request_id),
  KEY idx_cms_audit_action_created (action, created_at)
) ENGINE=InnoDB;

-- =========================================================
-- Superadmin Service Database
-- =========================================================

CREATE DATABASE IF NOT EXISTS superadmin_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE superadmin_db;

CREATE TABLE IF NOT EXISTS admin_users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  admin_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  role VARCHAR(64) NOT NULL,
  status ENUM('active', 'disabled') NOT NULL DEFAULT 'active',
  mfa_required BOOLEAN NOT NULL DEFAULT TRUE,
  created_by VARCHAR(64) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_users_admin_id (admin_id),
  UNIQUE KEY uk_admin_users_user_id (user_id),
  KEY idx_admin_users_role_status (role, status)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS admin_permissions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  permission_id VARCHAR(64) NOT NULL,
  permission_key VARCHAR(128) NOT NULL,
  description VARCHAR(512) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_permissions_id (permission_id),
  UNIQUE KEY uk_admin_permissions_key (permission_key)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS admin_role_permissions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  role VARCHAR(64) NOT NULL,
  permission_key VARCHAR(128) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_role_permission (role, permission_key)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS platform_settings (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  setting_key VARCHAR(128) NOT NULL,
  setting_value JSON NOT NULL,
  updated_by VARCHAR(64) NOT NULL,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_platform_settings_key (setting_key)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS admin_audit_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  audit_id VARCHAR(64) NOT NULL,
  actor_admin_id VARCHAR(64) NOT NULL,
  action VARCHAR(128) NOT NULL,
  resource_type VARCHAR(64) NOT NULL,
  resource_id VARCHAR(64) NOT NULL,
  request_id VARCHAR(64) NULL,
  ip_hash CHAR(64) NULL,
  reason VARCHAR(512) NULL,
  before_json JSON NULL,
  after_json JSON NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_audit_logs_audit_id (audit_id),
  KEY idx_admin_audit_actor_created (actor_admin_id, created_at),
  KEY idx_admin_audit_resource (resource_type, resource_id),
  KEY idx_admin_audit_action_created (action, created_at)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS admin_review_tasks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  task_id VARCHAR(64) NOT NULL,
  task_type VARCHAR(64) NOT NULL,
  resource_type VARCHAR(64) NOT NULL,
  resource_id VARCHAR(64) NOT NULL,
  status ENUM('open', 'approved', 'rejected', 'cancelled') NOT NULL DEFAULT 'open',
  assigned_to VARCHAR(64) NULL,
  created_by VARCHAR(64) NULL,
  reviewed_by VARCHAR(64) NULL,
  reviewed_at TIMESTAMP NULL,
  reason VARCHAR(512) NULL,
  metadata JSON NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_review_tasks_task_id (task_id),
  KEY idx_review_tasks_status_type (status, task_type),
  KEY idx_review_tasks_resource (resource_type, resource_id)
) ENGINE=InnoDB;
