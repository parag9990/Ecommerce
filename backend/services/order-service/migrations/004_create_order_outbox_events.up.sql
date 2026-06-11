-- Task 8: durable Order Service outbox for lifecycle events.

USE order_db;

CREATE TABLE IF NOT EXISTS order_outbox_events (
  event_id VARCHAR(64) NOT NULL,
  event_type VARCHAR(80) NOT NULL,
  version INT NOT NULL,
  aggregate_type VARCHAR(80) NOT NULL,
  aggregate_id VARCHAR(64) NOT NULL,
  routing_key VARCHAR(120) NOT NULL,
  deduplication_key VARCHAR(160) NOT NULL,
  source_history_id VARCHAR(64) NULL,
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
  UNIQUE KEY uk_order_outbox_deduplication (deduplication_key),
  KEY idx_order_outbox_pending (status, next_attempt_at, created_at),
  KEY idx_order_outbox_locked (status, locked_at),
  KEY idx_order_outbox_aggregate (aggregate_type, aggregate_id, created_at),
  KEY idx_order_outbox_type_created (event_type, created_at),
  CONSTRAINT fk_order_outbox_source_history
    FOREIGN KEY (source_history_id) REFERENCES order_status_history(history_id)
) ENGINE=InnoDB;
