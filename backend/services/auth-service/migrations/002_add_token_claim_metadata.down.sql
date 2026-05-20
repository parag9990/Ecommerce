-- Roll back token-claim metadata columns.

USE auth_db;

ALTER TABLE auth_accounts
  DROP KEY idx_auth_accounts_tenant_id,
  DROP KEY idx_auth_accounts_seller_id,
  DROP KEY uk_auth_accounts_user_id,
  DROP COLUMN tenant_id,
  DROP COLUMN seller_id,
  DROP COLUMN user_id;
