-- Order Service Database
-- Target: MySQL 8+
-- Rule: Order Service owns this schema. Other services must use service APIs/events.
-- Cross-service foreign keys are intentionally avoided.

SET NAMES utf8mb4;
SET time_zone = '+00:00';

CREATE DATABASE IF NOT EXISTS order_db
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE order_db;

CREATE TABLE IF NOT EXISTS orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  order_id VARCHAR(64) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  cart_id VARCHAR(64) NULL,
  status ENUM(
    'created',
    'pending_payment',
    'paid',
    'packed',
    'shipped',
    'delivered',
    'cancelled',
    'refunded'
  ) NOT NULL DEFAULT 'created',
  currency CHAR(3) NOT NULL,
  subtotal_amount BIGINT NOT NULL DEFAULT 0,
  discount_amount BIGINT NOT NULL DEFAULT 0,
  shipping_amount BIGINT NOT NULL DEFAULT 0,
  tax_amount BIGINT NOT NULL DEFAULT 0,
  total_amount BIGINT NOT NULL DEFAULT 0,
  address_snapshot JSON NOT NULL,
  coupon_code VARCHAR(64) NULL,
  payment_id VARCHAR(64) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_orders_order_id (order_id),
  KEY idx_orders_user_created (user_id, created_at),
  KEY idx_orders_status_created (status, created_at),
  KEY idx_orders_payment_id (payment_id),
  CONSTRAINT chk_orders_amounts_non_negative CHECK (
    subtotal_amount >= 0
    AND discount_amount >= 0
    AND shipping_amount >= 0
    AND tax_amount >= 0
    AND total_amount >= 0
  )
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS order_items (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  order_item_id VARCHAR(64) NOT NULL,
  order_id VARCHAR(64) NOT NULL,
  seller_id VARCHAR(64) NOT NULL,
  product_id VARCHAR(64) NOT NULL,
  variant_id VARCHAR(64) NOT NULL,
  sku VARCHAR(128) NOT NULL,
  title_snapshot VARCHAR(512) NOT NULL,
  image_url_snapshot VARCHAR(1024) NULL,
  quantity INT NOT NULL,
  currency CHAR(3) NOT NULL,
  unit_amount BIGINT NOT NULL,
  discount_amount BIGINT NOT NULL DEFAULT 0,
  tax_amount BIGINT NOT NULL DEFAULT 0,
  total_amount BIGINT NOT NULL,
  fulfillment_status ENUM(
    'pending',
    'packed',
    'shipped',
    'delivered',
    'cancelled',
    'returned'
  ) NOT NULL DEFAULT 'pending',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_items_order_item_id (order_item_id),
  KEY idx_order_items_order (order_id),
  KEY idx_order_items_seller_created (seller_id, created_at),
  KEY idx_order_items_product (product_id),
  CONSTRAINT fk_order_items_order
    FOREIGN KEY (order_id) REFERENCES orders(order_id),
  CONSTRAINT chk_order_items_quantity_positive CHECK (quantity > 0),
  CONSTRAINT chk_order_items_amounts_non_negative CHECK (
    unit_amount >= 0
    AND discount_amount >= 0
    AND tax_amount >= 0
    AND total_amount >= 0
  )
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS order_status_history (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  history_id VARCHAR(64) NOT NULL,
  order_id VARCHAR(64) NOT NULL,
  from_status ENUM(
    'created',
    'pending_payment',
    'paid',
    'packed',
    'shipped',
    'delivered',
    'cancelled',
    'refunded'
  ) NULL,
  to_status ENUM(
    'created',
    'pending_payment',
    'paid',
    'packed',
    'shipped',
    'delivered',
    'cancelled',
    'refunded'
  ) NOT NULL,
  reason VARCHAR(512) NULL,
  actor_type ENUM(
    'system',
    'buyer',
    'seller',
    'admin',
    'payment_service',
    'logistics'
  ) NOT NULL,
  changed_by VARCHAR(64) NULL,
  metadata JSON NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_status_history_id (history_id),
  KEY idx_order_status_history_order_created (order_id, created_at),
  CONSTRAINT fk_order_status_history_order
    FOREIGN KEY (order_id) REFERENCES orders(order_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS shipments (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  shipment_id VARCHAR(64) NOT NULL,
  order_id VARCHAR(64) NOT NULL,
  seller_id VARCHAR(64) NOT NULL,
  carrier VARCHAR(128) NULL,
  tracking_number VARCHAR(128) NULL,
  status ENUM(
    'pending',
    'packed',
    'shipped',
    'delivered',
    'failed'
  ) NOT NULL DEFAULT 'pending',
  shipped_at TIMESTAMP NULL,
  delivered_at TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_shipments_shipment_id (shipment_id),
  KEY idx_shipments_order (order_id),
  KEY idx_shipments_seller_status (seller_id, status),
  KEY idx_shipments_tracking (carrier, tracking_number),
  CONSTRAINT fk_shipments_order
    FOREIGN KEY (order_id) REFERENCES orders(order_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS order_idempotency_keys (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id VARCHAR(64) NOT NULL,
  idempotency_key VARCHAR(128) NOT NULL,
  order_id VARCHAR(64) NULL,
  request_hash CHAR(64) NOT NULL,
  status ENUM('processing', 'completed', 'failed') NOT NULL DEFAULT 'processing',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expires_at TIMESTAMP NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_idempotency_user_key (user_id, idempotency_key),
  KEY idx_order_idempotency_expires (expires_at),
  KEY idx_order_idempotency_order (order_id),
  CONSTRAINT fk_order_idempotency_order
    FOREIGN KEY (order_id) REFERENCES orders(order_id)
) ENGINE=InnoDB;
