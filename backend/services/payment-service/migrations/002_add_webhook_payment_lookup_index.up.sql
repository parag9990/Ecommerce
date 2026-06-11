USE payment_db;

CREATE INDEX idx_payments_provider_intent ON payments (provider, provider_intent_id);
