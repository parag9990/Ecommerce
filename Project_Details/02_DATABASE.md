# Database

## Databases Used

| Database | Local service | Used by |
| --- | --- | --- |
| MySQL 8.4 | `mysql` in `docker-compose.yml` | auth, user, order, payment, CMS, superadmin |
| MongoDB 8.0 replica set | `mongodb` in `docker-compose.yml` | product, cart, wishlist, recommendation, session, notification |
| Redis 7.4 | `redis` in `docker-compose.yml` | auth, API Gateway rate limiting/session-related integration |
| Typesense | `typesense` in `docker-compose.yml` | search service |
| RabbitMQ 4.1 | `rabbitmq` in `docker-compose.yml` | notification and optional service event publishing |
| Kafka 3.9.1 | `kafka` in `docker-compose.yml` | optional event publishing/consumption paths |

## MySQL Databases

`infra/mysql/init/00-create-databases.sql` creates:

| Database | Owner service |
| --- | --- |
| `auth_db` | Auth service |
| `user_db` | User service |
| `order_db` | Order service |
| `payment_db` | Payment service |
| `cms_db` | CMS service |
| `superadmin_db` | Superadmin service |

## Auth Service Schema

Migrations: `backend/services/auth-service/migrations`.

| Table | Purpose |
| --- | --- |
| `auth_accounts` | Authentication account records and account metadata |
| `credentials` | Password credential hashes and metadata |
| `refresh_tokens` | Refresh-token hashes, session metadata, expiry/revocation |
| `otp_challenges` | OTP challenge records |
| `role_assignments` | Active/scoped role assignments |
| `auth_outbox_events` | Auth/session-link/fraud-signal event outbox |

Indexes and constraints found include refresh-token hash uniqueness, account metadata indexes, OTP account FK, active role assignment uniqueness, scope indexes, and outbox status/lock indexes.

## User Service Schema

Migrations: `backend/services/user-service/migrations`.

| Table | Purpose |
| --- | --- |
| `users` | User profile and status data |
| `user_addresses` | User-owned addresses |
| `seller_profiles` | Seller profile/status data |
| `seller_kyc_documents` | Seller KYC document metadata |
| `user_outbox_events` | User event outbox |

Indexes and constraints found include user/seller lookup indexes, address ownership indexes, audit/status/deleted fields, and outbox status/retry indexes.

## Order Service Schema

Migrations: `backend/services/order-service/migrations`.

| Table | Purpose |
| --- | --- |
| `orders` | Order aggregate and buyer/seller/payment state |
| `order_items` | Product line items for an order |
| `order_status_history` | Status transition history |
| `shipments` | Shipment metadata |
| `order_idempotency_keys` | Checkout idempotency |
| `order_outbox_events` | Order event outbox |

Relationships found are internal to the order database, for example order items/status/shipment rows reference orders. Cross-service foreign keys to user/product/payment databases were not found.

## Payment Service Schema

Migrations: `backend/services/payment-service/migrations`.

| Table | Purpose |
| --- | --- |
| `payments` | Payment records, provider references, and state |
| `payment_attempts` | Payment attempt history |
| `refunds` | Refund records and review status |
| `payment_webhook_events` | Provider webhook idempotency/audit records |
| `payment_reconciliations` | Settlement/reconciliation records |

Indexes found include provider intent lookup, provider refund lookup, retry lineage, webhook idempotency, and reconciliation lookup indexes.

## Product Service MongoDB Schema

Migrations: `backend/services/product-service/migrations`.

| Collection | Purpose |
| --- | --- |
| `products` | Product documents, seller workflow, prices, media, inventory-related fields |
| `categories` | Category hierarchy and catalog navigation |
| `brands` | Brand metadata |
| `inventory_snapshots` | Inventory state snapshots |
| `price_books` | Price book data |
| `inventory_reservations` | Reservation/commit/release workflow |
| `product_event_outbox` | Product event outbox |

Indexes found include product status/published/category/price/rating reads, category reads, reservation idempotency/order/status/expiry indexes, TTL cleanup, and product outbox status/published TTL indexes.

## Notification Service MongoDB Schema

Migrations: `backend/services/notification-service/migrations`.

| Collection | Purpose |
| --- | --- |
| `notification_templates` | Renderable notification templates by key/channel/version |
| `notification_deliveries` | Delivery attempts and provider status |
| `notification_preferences` | User notification preferences and suppression |
| `provider_events` | Provider delivery/open/failure events |

Seeded templates include OTP verification, order status update, payment status update, promotional offer, welcome user, seller approved, and address security notice templates.

## Other MongoDB-Backed Services

Mongo migration jobs exist in `docker-compose.yml` for:

| Service | Migration location |
| --- | --- |
| Cart | `backend/services/cart-service/migrations` |
| Wishlist | `backend/services/wishlist-service/migrations` |
| Recommendation | `backend/services/recommendation-service/migrations` |
| Session | `backend/services/session-service/migrations` |

Detailed collection documentation for these services is outside the requested service-specific files. They are present in the repository and should be reviewed before changing those services.

## Shared Database Design Docs

| File | Notes |
| --- | --- |
| `database/mongodb-schema-design.md` | Design notes for MongoDB-backed services |
| `database/draw.sql` | SQL drawing/design file |

## Service Data Ownership

| Service | Owned data |
| --- | --- |
| Auth | Auth accounts, credentials, tokens, OTP, roles, auth outbox |
| User | Users, addresses, seller profiles, seller KYC, user outbox |
| Product | Products, categories, brands, inventory reservations, product outbox |
| Order | Orders, items, status history, shipments, order idempotency, order outbox |
| Payment | Payments, attempts, refunds, webhook events, reconciliations |
| Notification | Templates, deliveries, preferences, provider events |
| CMS | MySQL database exists; detailed schema not covered by requested service docs |
| Superadmin | MySQL database exists; detailed schema not covered by requested service docs |

## Missing Database Information

- ERD diagrams: Not found in current codebase.
- Seeder framework for MySQL services: Not found in current codebase.
- Documented backup/restore procedures: Not found in current codebase.
- Production database sizing and retention policy: Not found in current codebase.
