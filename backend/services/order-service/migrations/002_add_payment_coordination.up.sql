-- Task 4: durable inventory reservation handoff and payment failure lifecycle.

USE order_db;

ALTER TABLE orders
  MODIFY COLUMN status ENUM(
    'created',
    'pending_payment',
    'paid',
    'payment_failed',
    'packed',
    'shipped',
    'delivered',
    'cancelled',
    'refunded'
  ) NOT NULL DEFAULT 'created',
  ADD COLUMN inventory_reservation_id VARCHAR(64) NULL AFTER payment_id,
  ADD COLUMN inventory_reserved_until TIMESTAMP NULL AFTER inventory_reservation_id,
  ADD KEY idx_orders_inventory_reservation (inventory_reservation_id),
  ADD CONSTRAINT chk_orders_inventory_reservation_pair CHECK (
    (inventory_reservation_id IS NULL AND inventory_reserved_until IS NULL)
    OR (inventory_reservation_id IS NOT NULL AND inventory_reserved_until IS NOT NULL)
  );

ALTER TABLE order_status_history
  MODIFY COLUMN from_status ENUM(
    'created',
    'pending_payment',
    'paid',
    'payment_failed',
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
    'payment_failed',
    'packed',
    'shipped',
    'delivered',
    'cancelled',
    'refunded'
  ) NOT NULL;
