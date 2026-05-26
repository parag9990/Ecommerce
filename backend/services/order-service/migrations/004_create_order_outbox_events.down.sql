-- Roll back Task 8 order outbox storage.

USE order_db;

DROP TABLE IF EXISTS order_outbox_events;
