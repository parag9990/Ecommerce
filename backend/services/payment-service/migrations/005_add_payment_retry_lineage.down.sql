USE payment_db;

ALTER TABLE payments
  DROP FOREIGN KEY fk_payments_retry_of,
  DROP INDEX uk_payments_retry_request,
  DROP INDEX uk_payments_root_attempt,
  DROP INDEX idx_payments_order_status,
  DROP INDEX idx_payments_root_status,
  DROP COLUMN retry_request_key,
  DROP COLUMN attempt_no,
  DROP COLUMN root_payment_id,
  DROP COLUMN retry_of_payment_id;
