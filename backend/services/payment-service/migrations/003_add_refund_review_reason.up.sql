USE payment_db;

ALTER TABLE refunds
  ADD COLUMN review_reason VARCHAR(512) NULL AFTER reviewed_by;
