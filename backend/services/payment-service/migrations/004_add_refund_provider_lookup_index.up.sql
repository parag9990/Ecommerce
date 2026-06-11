USE payment_db;

CREATE INDEX idx_refunds_provider_refund ON refunds (provider_refund_id);
