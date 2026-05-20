-- Add token-claim metadata maintained by Auth Service as a read model.
-- User profile ownership remains in User Service; Auth stores only IDs needed
-- to mint and refresh access-token claims safely.

USE auth_db;

ALTER TABLE auth_accounts
  ADD COLUMN user_id VARCHAR(64) NULL AFTER account_id,
  ADD COLUMN seller_id VARCHAR(64) NULL AFTER phone,
  ADD COLUMN tenant_id VARCHAR(64) NULL AFTER seller_id,
  ADD UNIQUE KEY uk_auth_accounts_user_id (user_id),
  ADD KEY idx_auth_accounts_seller_id (seller_id),
  ADD KEY idx_auth_accounts_tenant_id (tenant_id);
