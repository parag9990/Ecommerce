SET NAMES utf8mb4;
SET time_zone = '+00:00';

USE user_db;

CREATE TABLE IF NOT EXISTS user_outbox_events (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  event_id VARCHAR(64) NOT NULL,
  event_type VARCHAR(80) NOT NULL,
  event_version INT NOT NULL,
  topic VARCHAR(120) NOT NULL,
  aggregate_type VARCHAR(80) NOT NULL,
  aggregate_id VARCHAR(64) NOT NULL,
  payload JSON NOT NULL,
  request_id VARCHAR(128) NULL,
  trace_id VARCHAR(128) NULL,
  status ENUM('pending', 'published', 'failed', 'dead') NOT NULL DEFAULT 'pending',
  attempts INT NOT NULL DEFAULT 0,
  next_attempt_at TIMESTAMP(6) NULL,
  locked_until TIMESTAMP(6) NULL,
  last_error VARCHAR(1024) NULL,
  occurred_at TIMESTAMP(6) NOT NULL,
  published_at TIMESTAMP(6) NULL,
  created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uk_user_outbox_event_id (event_id),
  KEY idx_user_outbox_pending (status, next_attempt_at, created_at),
  KEY idx_user_outbox_lock (locked_until),
  KEY idx_user_outbox_aggregate (aggregate_type, aggregate_id),
  KEY idx_user_outbox_type_time (event_type, occurred_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
