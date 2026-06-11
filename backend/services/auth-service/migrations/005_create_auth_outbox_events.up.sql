-- Durable Auth Service outbox for session-link and fraud-signal events.
-- Events are published asynchronously to auth.events and consumed by Session Service.

USE auth_db;

CREATE TABLE IF NOT EXISTS auth_outbox_events (
  event_id VARCHAR(64) NOT NULL,
  event_type VARCHAR(80) NOT NULL,
  version INT NOT NULL,
  aggregate_type VARCHAR(80) NOT NULL,
  aggregate_id VARCHAR(64) NOT NULL,
  routing_key VARCHAR(120) NOT NULL,
  payload JSON NOT NULL,
  trace_id VARCHAR(128) NULL,
  status ENUM('pending', 'processing', 'published', 'failed', 'dead_letter') NOT NULL DEFAULT 'pending',
  attempts INT NOT NULL DEFAULT 0,
  next_attempt_at TIMESTAMP NULL,
  published_at TIMESTAMP NULL,
  locked_at TIMESTAMP NULL,
  locked_by VARCHAR(128) NULL,
  last_error VARCHAR(1024) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (event_id),
  KEY idx_auth_outbox_pending (status, next_attempt_at, created_at),
  KEY idx_auth_outbox_locked (status, locked_at),
  KEY idx_auth_outbox_aggregate (aggregate_type, aggregate_id),
  KEY idx_auth_outbox_type_created (event_type, created_at)
) ENGINE=InnoDB;
