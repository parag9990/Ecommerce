# MongoDB Schema Design

This document covers MongoDB-backed services: Product, Cart, Wishlist, Recommendation, Session Management, and Notification.

## Product Service

Database: `product_db`

### Collection: `products`

Example document:

```json
{
  "_id": "prod_123",
  "seller_id": "seller_456",
  "title": "Running Shoes",
  "description": "Lightweight running shoes",
  "brand": "Acme",
  "category_id": "cat_shoes",
  "status": "published",
  "attributes": {
    "gender": "men",
    "material": "mesh"
  },
  "images": [
    {
      "url": "https://cdn.example.com/prod_123/main.jpg",
      "alt": "Running shoes side view",
      "position": 1
    }
  ],
  "variants": [
    {
      "variant_id": "var_1",
      "sku": "ACME-SHOE-9-BLK",
      "attributes": {
        "size": "9",
        "color": "black"
      },
      "price": {
        "amount": 299900,
        "currency": "INR"
      },
      "mrp": {
        "amount": 399900,
        "currency": "INR"
      },
      "stock_quantity": 120,
      "reserved_quantity": 5
    }
  ],
  "rating": {
    "average": 4.4,
    "count": 120
  },
  "created_at": "2026-05-18T00:00:00Z",
  "updated_at": "2026-05-18T00:00:00Z"
}
```

Indexes:

```javascript
db.products.createIndex({ seller_id: 1, status: 1, updated_at: -1 })
db.products.createIndex({ category_id: 1, status: 1, updated_at: -1 })
db.products.createIndex({ "variants.sku": 1 }, { unique: true })
db.products.createIndex({ title: "text", description: "text", brand: "text" })
```

Scalability:

- Use product id as stable shard key if sharding.
- Keep product detail document self-contained for fast reads.
- Search Service owns full-text search, Mongo text index is only fallback/dev.

### Collection: `categories`

Indexes:

```javascript
db.categories.createIndex({ parent_id: 1, sort_order: 1 })
db.categories.createIndex({ slug: 1 }, { unique: true })
```

### Collection: `inventory_snapshots`

Use for audit and periodic stock snapshots.

Indexes:

```javascript
db.inventory_snapshots.createIndex({ product_id: 1, variant_id: 1, created_at: -1 })
```

## Cart Service

Database: `cart_db`

### Collection: `carts`

Example document:

```json
{
  "_id": "cart_123",
  "user_id": "user_123",
  "guest_session_id": null,
  "status": "active",
  "items": [
    {
      "item_id": "item_1",
      "product_id": "prod_123",
      "variant_id": "var_1",
      "seller_id": "seller_456",
      "title_snapshot": "Running Shoes",
      "image_url_snapshot": "https://cdn.example.com/prod_123/main.jpg",
      "unit_price": {
        "amount": 299900,
        "currency": "INR"
      },
      "quantity": 2,
      "added_at": "2026-05-18T00:00:00Z"
    }
  ],
  "coupon_code": "SAVE10",
  "totals": {
    "subtotal": 599800,
    "discount": 59980,
    "total": 539820,
    "currency": "INR"
  },
  "created_at": "2026-05-18T00:00:00Z",
  "updated_at": "2026-05-18T00:00:00Z",
  "expires_at": "2026-06-17T00:00:00Z"
}
```

Indexes:

```javascript
db.carts.createIndex({ user_id: 1, status: 1 })
db.carts.createIndex({ guest_session_id: 1, status: 1 })
db.carts.createIndex({ expires_at: 1 }, { expireAfterSeconds: 0 })
```

Scalability:

- Active cart also cached in Redis.
- Cart writes are per user/session, natural partitioning.

## Wishlist Service

Database: `wishlist_db`

### Collection: `wishlists`

Example document:

```json
{
  "_id": "wish_123",
  "user_id": "user_123",
  "visibility": "private",
  "items": [
    {
      "product_id": "prod_123",
      "variant_id": "var_1",
      "added_at": "2026-05-18T00:00:00Z",
      "last_known_price": {
        "amount": 299900,
        "currency": "INR"
      },
      "availability": "in_stock"
    }
  ],
  "created_at": "2026-05-18T00:00:00Z",
  "updated_at": "2026-05-18T00:00:00Z"
}
```

Indexes:

```javascript
db.wishlists.createIndex({ user_id: 1 }, { unique: true })
db.wishlists.createIndex({ "items.product_id": 1 })
```

## Recommendation Service

Database: `recommendation_db`

### Collection: `user_interactions`

Example document:

```json
{
  "_id": "interaction_123",
  "user_id": "user_123",
  "anonymous_id": "anon_123",
  "product_id": "prod_123",
  "category_id": "cat_shoes",
  "event_type": "product_view",
  "weight": 1,
  "occurred_at": "2026-05-18T00:00:00Z"
}
```

Indexes:

```javascript
db.user_interactions.createIndex({ user_id: 1, occurred_at: -1 })
db.user_interactions.createIndex({ product_id: 1, event_type: 1, occurred_at: -1 })
db.user_interactions.createIndex({ occurred_at: 1 }, { expireAfterSeconds: 15552000 })
```

### Collection: `recommendation_sets`

Indexes:

```javascript
db.recommendation_sets.createIndex({ context_key: 1, strategy_id: 1 }, { unique: true })
db.recommendation_sets.createIndex({ expires_at: 1 }, { expireAfterSeconds: 0 })
```

## Session Management Service

Database: `session_db`

### Collection: `sessions`

Example document:

```json
{
  "_id": "sess_123",
  "anonymous_id": "anon_123",
  "user_id": "user_123",
  "status": "active",
  "entry_page": "/",
  "exit_page": null,
  "device": {
    "type": "mobile",
    "browser": "Chrome",
    "os": "Android"
  },
  "geo": {
    "country": "IN",
    "city": "Delhi"
  },
  "started_at": "2026-05-18T00:00:00Z",
  "last_seen_at": "2026-05-18T00:20:00Z",
  "ended_at": null
}
```

Indexes:

```javascript
db.sessions.createIndex({ anonymous_id: 1, started_at: -1 })
db.sessions.createIndex({ user_id: 1, started_at: -1 })
db.sessions.createIndex({ status: 1, last_seen_at: -1 })
```

### Collection: `session_events`

Example document:

```json
{
  "_id": "evt_123",
  "session_id": "sess_123",
  "anonymous_id": "anon_123",
  "user_id": "user_123",
  "event_type": "click",
  "path": "/products/prod_123",
  "properties": {
    "element_id": "add-to-cart",
    "x": 42.5,
    "y": 71.2,
    "viewport_width": 390,
    "viewport_height": 844
  },
  "occurred_at": "2026-05-18T00:10:00Z",
  "received_at": "2026-05-18T00:10:01Z"
}
```

Indexes:

```javascript
db.session_events.createIndex({ session_id: 1, occurred_at: 1 })
db.session_events.createIndex({ user_id: 1, occurred_at: -1 })
db.session_events.createIndex({ event_type: 1, occurred_at: -1 })
db.session_events.createIndex({ occurred_at: 1 }, { expireAfterSeconds: 7776000 })
```

### Collection: `heatmap_points`

Indexes:

```javascript
db.heatmap_points.createIndex({ path: 1, device_type: 1, day: 1 })
```

### Collection: `analytics_aggregates`

Indexes:

```javascript
db.analytics_aggregates.createIndex({ metric: 1, bucket: 1, segment: 1 }, { unique: true })
```

## Notification Service

Database: `notification_db`

### Collection: `notification_templates`

Example document:

```json
{
  "_id": "tpl_order_paid_email",
  "template_key": "order_paid",
  "channel": "email",
  "subject": "Your order is confirmed",
  "body": "Hi {{name}}, your order {{order_id}} is confirmed.",
  "status": "active",
  "version": 1,
  "created_at": "2026-05-18T00:00:00Z",
  "updated_at": "2026-05-18T00:00:00Z"
}
```

Indexes:

```javascript
db.notification_templates.createIndex({ template_key: 1, channel: 1, version: -1 })
```

### Collection: `notification_deliveries`

Example document:

```json
{
  "_id": "delivery_123",
  "user_id": "user_123",
  "channel": "email",
  "template_key": "order_paid",
  "status": "sent",
  "provider": "ses",
  "provider_message_id": "msg_123",
  "attempts": 1,
  "payload": {
    "order_id": "order_123"
  },
  "created_at": "2026-05-18T00:00:00Z",
  "updated_at": "2026-05-18T00:00:00Z"
}
```

Indexes:

```javascript
db.notification_deliveries.createIndex({ user_id: 1, created_at: -1 })
db.notification_deliveries.createIndex({ status: 1, created_at: 1 })
db.notification_deliveries.createIndex({ provider: 1, provider_message_id: 1 })
```

### Collection: `notification_preferences`

Indexes:

```javascript
db.notification_preferences.createIndex({ user_id: 1 }, { unique: true })
```

## MongoDB Operational Notes

- Use replica sets in production.
- Enable backups and point-in-time recovery where supported.
- Use TTL indexes for raw events and stale carts.
- Keep documents below MongoDB document size limit.
- Avoid unbounded arrays for high-volume events. Session events are separate documents, not embedded inside session.
- Use schema validation in production collections for core required fields.

