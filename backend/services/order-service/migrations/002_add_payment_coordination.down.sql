-- Roll back Task 4. Failed payment orders become cancelled in the older lifecycle.

USE order_db;

UPDATE order_status_history
SET from_status = 'cancelled'
WHERE from_status = 'payment_failed';

UPDATE order_status_history
SET to_status = 'cancelled',
    reason = 'payment_failed_downgraded_to_cancelled'
WHERE to_status = 'payment_failed';

UPDATE orders
SET status = 'cancelled'
WHERE status = 'payment_failed';

ALTER TABLE order_status_history
  MODIFY COLUMN from_status ENUM(
    'created',
    'pending_payment',
    'paid',
    'packed',
    'shipped',
    'delivered',
    'cancelled',
    'refunded'
  ) NULL,
  MODIFY COLUMN to_status ENUM(
    'created',
    'pending_payment',
    'paid',
    'packed',
    'shipped',
    'delivered',
    'cancelled',
    'refunded'
  ) NOT NULL;

ALTER TABLE orders
  DROP CHECK chk_orders_inventory_reservation_pair,
  DROP INDEX idx_orders_inventory_reservation,
  DROP COLUMN inventory_reserved_until,
  DROP COLUMN inventory_reservation_id,
  MODIFY COLUMN status ENUM(
    'created',
    'pending_payment',
    'paid',
    'packed',
    'shipped',
    'delivered',
    'cancelled',
    'refunded'
  ) NOT NULL DEFAULT 'created';
