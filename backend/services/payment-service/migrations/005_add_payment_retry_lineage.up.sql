USE payment_db;

ALTER TABLE payments
  ADD COLUMN retry_of_payment_id VARCHAR(64) NULL AFTER idempotency_key,
  ADD COLUMN root_payment_id VARCHAR(64) NULL AFTER retry_of_payment_id,
  ADD COLUMN attempt_no INT UNSIGNED NOT NULL DEFAULT 1 AFTER root_payment_id,
  ADD COLUMN retry_request_key VARCHAR(128) NULL AFTER attempt_no,
  ADD CONSTRAINT fk_payments_retry_of
    FOREIGN KEY (retry_of_payment_id) REFERENCES payments(payment_id),
  ADD UNIQUE KEY uk_payments_retry_request (retry_of_payment_id, retry_request_key),
  ADD UNIQUE KEY uk_payments_root_attempt (root_payment_id, attempt_no),
  ADD KEY idx_payments_order_status (order_id, status),
  ADD KEY idx_payments_root_status (root_payment_id, status);
