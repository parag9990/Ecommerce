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
