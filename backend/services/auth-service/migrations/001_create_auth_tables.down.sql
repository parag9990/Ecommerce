-- Rollback Auth Service tables.
-- Drop child tables before parent tables because of foreign keys.

USE auth_db;

DROP TABLE IF EXISTS role_assignments;
DROP TABLE IF EXISTS otp_challenges;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS credentials;
DROP TABLE IF EXISTS auth_accounts;
