# Step 6 - Database Design

## Database Strategy

Har microservice ka separate database hoga. Isse ownership clear rahegi, scaling independent rahegi, aur accidental coupling kam hoga.

| Service | DB | Reason |
|---|---|---|
| Auth | MySQL + Redis | Credentials, tokens, OTP audit strongly consistent. Redis short-lived counters. |
| User | MySQL | Structured profile, address, seller KYC. |
| Product | MongoDB | Dynamic product attributes and variants. |
| Cart | MongoDB + Redis | Flexible cart document, hot cache. |
| Wishlist | MongoDB | User list document simple and flexible. |
| Order | MySQL | Transactional order lifecycle. |
| Payment | MySQL | Financial audit and reconciliation. |
| Recommendation | MongoDB + Redis | Flexible feature docs and low-latency cache. |
| Search | Typesense | Search index optimized for text/facets. |
| CMS | MySQL | Coupons, campaigns, seller permissions. |
| Session | MongoDB + Redis | High-volume flexible events and active sessions. |
| Notification | MongoDB | Templates and provider payloads vary. |
| Superadmin | MySQL | Admin permissions and audit logs. |

## MySQL Schemas

Detailed SQL DDL is available in [draw.sql](../database/draw.sql).

### Auth DB

Tables:

- `auth_accounts`: login identity and verification status.
- `credentials`: password hash and password policy metadata.
- `refresh_tokens`: rotating refresh tokens.
- `otp_challenges`: email/phone OTP challenges.
- `role_assignments`: roles per user.

Relationships:

- One `auth_account` has one `credential`.
- One `auth_account` has many `refresh_tokens`.
- One `auth_account` has many `role_assignments`.

Indexes:

- Unique `email`, unique `phone`.
- `refresh_tokens(token_hash)`.
- `otp_challenges(target, purpose, expires_at)`.

Scalability:

- Partition refresh tokens by `created_at` if token volume becomes large.
- Redis rate limits prevent brute force.
- Auth DB read replicas can serve role lookup if needed.

### User DB

Tables:

- `users`
- `user_addresses`
- `seller_profiles`
- `seller_kyc_documents`

Relationships:

- One user has many addresses.
- One user may have one seller profile.
- One seller profile has many KYC documents.

Indexes:

- `users(auth_account_id)`
- `user_addresses(user_id, is_default)`
- `seller_profiles(user_id)`
- `seller_profiles(status)`

Scalability:

- User profile cache can sit in Redis for hot users.
- Address writes stay in primary DB.
- Seller read models can be denormalized for admin search.

### Order DB

Tables:

- `orders`
- `order_items`
- `order_status_history`
- `shipments`
- `order_idempotency_keys`

Relationships:

- One order has many order items.
- One order has many status history records.
- One order can have many shipments for split packages.

Indexes:

- `orders(user_id, created_at)`
- `orders(status, created_at)`
- `order_items(seller_id, created_at)`
- `order_idempotency_keys(user_id, idempotency_key)` unique.

Scalability:

- Orders can be partitioned by month or hash of user id at high scale.
- Seller order list can use read replica or materialized read model.
- Immutable order item snapshots avoid cross-service read dependency.

### Payment DB

Tables:

- `payments`
- `payment_attempts`
- `refunds`
- `payment_webhook_events`
- `payment_reconciliations`

Relationships:

- One order can have multiple payment attempts.
- One payment can have multiple refunds.
- One provider webhook event maps to one stored event id.

Indexes:

- `payments(order_id)`
- `payments(provider, provider_payment_id)`
- `payment_webhook_events(provider, provider_event_id)` unique.
- `refunds(payment_id, status)`.

Scalability:

- Payment rows immutable where possible.
- Webhook processing idempotent.
- Reconciliation jobs process provider reports in batches.

### CMS DB

Tables:

- `seller_settings`
- `seller_staff`
- `coupons`
- `coupon_rules`
- `coupon_redemptions`
- `campaigns`
- `cms_audit_logs`

Relationships:

- One seller has many staff.
- One coupon has many rules.
- One coupon has many redemptions.

Indexes:

- `coupons(code)` unique.
- `coupons(seller_id, status)`.
- `coupon_redemptions(user_id, coupon_id)`.
- `campaigns(seller_id, starts_at, ends_at)`.

Scalability:

- Coupon validation results can be cached for short TTL.
- Coupon redemption final write should be transactional and idempotent.
- Analytics should use aggregates, not scan raw orders.

### Superadmin DB

Tables:

- `admin_users`
- `admin_permissions`
- `admin_role_permissions`
- `platform_settings`
- `admin_audit_logs`
- `admin_review_tasks`

Relationships:

- Admin users have roles.
- Roles map to permissions.
- Review tasks connect admin workflows to service resources.

Indexes:

- `admin_users(user_id)` unique.
- `platform_settings(setting_key)` unique.
- `admin_audit_logs(actor_admin_id, created_at)`.
- `admin_audit_logs(resource_type, resource_id)`.

Scalability:

- Audit logs can be archived to object storage.
- Admin search can use read replicas or OpenSearch later.
- Settings should be cached with invalidation events.

## MongoDB Schemas

Detailed collection design is available in [mongodb-schema-design.md](../database/mongodb-schema-design.md).

## Typesense Search Schema

Collection: `products`

Fields:

| Field | Type | Facet | Sort | Detail |
|---|---|---|---|---|
| `id` | string | no | no | Product id. |
| `title` | string | no | no | Searchable title. |
| `description` | string | no | no | Searchable description. |
| `brand` | string | yes | no | Brand facet. |
| `category_ids` | string[] | yes | no | Category filters. |
| `seller_id` | string | yes | no | Seller filter. |
| `price` | float | yes | yes | Current sale price. |
| `rating` | float | yes | yes | Average rating. |
| `popularity_score` | int32 | yes | yes | Ranking signal. |
| `in_stock` | bool | yes | no | Stock filter. |
| `created_at` | int64 | no | yes | New arrivals sort. |

Indexing strategy:

- Facets on brand, category, seller, price, rating, stock.
- Sort by price, rating, popularity, created_at.
- Synonyms for common product terms.
- Zero-result query tracking for improving synonyms.

## Cross-Service Data Consistency

Use saga-style orchestration for checkout:

1. Order Service creates pending order with idempotency key.
2. Product Service reserves inventory.
3. Payment Service creates payment intent.
4. Payment webhook confirms success.
5. Order Service marks order paid.
6. Product Service commits inventory.
7. Notification Service sends confirmation.

Failure handling:

- Inventory reserve fails: order not created or marked failed.
- Payment intent fails: release inventory.
- Payment webhook failure retry: webhook event idempotency handles duplicates.
- Order cancel after payment: trigger refund flow.

## Indexing Strategy Summary

| Service | High-Value Indexes |
|---|---|
| Auth | email, phone, token hash, OTP target and expiry |
| User | auth account id, seller status, user address default |
| Product | seller id, category id, status, variant sku, inventory quantity |
| Cart | user id, session id, updated at TTL |
| Wishlist | user id, product id |
| Order | user id + created at, seller id + created at, status |
| Payment | order id, provider event id, provider payment id |
| CMS | coupon code, seller id + status, campaign time window |
| Session | session id, user id, anonymous id, event time TTL |
| Notification | user id, status, template key |
| Superadmin | actor + created at, resource type + resource id |

## Scalability Approach

- Use read replicas for MySQL read-heavy admin and order queries.
- Use Mongo sharding for high-volume session events and product catalog if needed.
- Use Redis for hot reads: cart count, active sessions, rate limits, product summaries.
- Use async events to avoid synchronous fan-out.
- Use batch consumers for search indexing and recommendation features.
- Use TTL indexes for OTPs, sessions, raw analytics, and stale carts.
- Use outbox pattern for important service events from MySQL-backed services.

