-- Auth Service Database
-- Target: MySQL 8+
-- Rule: Auth Service owns this schema. Other services must use service APIs/events.

SET NAMES utf8mb4;
SET time_zone = '+00:00';

CREATE DATABASE IF NOT EXISTS auth_db
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

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
  CONSTRAINT fk_credentials_account
    FOREIGN KEY (account_id) REFERENCES auth_accounts(account_id)
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
  CONSTRAINT fk_refresh_tokens_account
    FOREIGN KEY (account_id) REFERENCES auth_accounts(account_id)
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
  CONSTRAINT fk_role_assignments_account
    FOREIGN KEY (account_id) REFERENCES auth_accounts(account_id)
) ENGINE=InnoDB;
