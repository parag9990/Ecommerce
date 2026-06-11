-- Roll back the Task 5 query-supporting index.

USE order_db;

ALTER TABLE orders
  DROP INDEX idx_orders_user_created_order;
