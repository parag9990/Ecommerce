-- Rollback Order Service tables.
-- Drop child tables before parent tables because of foreign keys.

USE order_db;

DROP TABLE IF EXISTS order_idempotency_keys;
DROP TABLE IF EXISTS shipments;
DROP TABLE IF EXISTS order_status_history;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
