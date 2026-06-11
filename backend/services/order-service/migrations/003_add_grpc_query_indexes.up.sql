-- Task 5: buyer order cursor listing reads newest orders by user and stable order ID.

USE order_db;

ALTER TABLE orders
  ADD KEY idx_orders_user_created_order (user_id, created_at, order_id);
