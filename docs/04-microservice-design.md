# Step 4 - Microservice Design

Golden rule: har service apni database own karegi. Dusri service ke DB ko directly read/write nahi karna. Data chahiye to gRPC call ya event projection use karo.

## Service Communication Pattern

| Pattern | Use Case | Example |
|---|---|---|
| REST via API Gateway | Browser/public clients | `GET /api/v1/products/{id}` |
| gRPC internal | Service-to-service sync calls | Order to Product `ReserveInventory` |
| Events | Async workflows and projections | Product emits `ProductUpdated`, Search consumes |
| Webhook HTTP | External providers | Payment provider webhook to Payment Service |

## User Service

### Purpose

User profile, addresses, seller profile, and KYC metadata manage karna.

### Responsibilities

- Buyer profile create/update.
- Seller profile and KYC status maintain.
- Address CRUD.
- User status: active, blocked, deleted.
- Publish user profile events.

### Database Choice

MySQL. Reason: user, address, seller profile, and KYC records structured hain. Strong relational constraints and transactions useful honge.

### Tables

- `users`
- `user_addresses`
- `seller_profiles`
- `seller_kyc_documents`

### gRPC Services

- `CreateUser`
- `GetUser`
- `BatchGetUsers`
- `UpdateUserProfile`
- `ListUserAddresses`
- `CreateAddress`
- `UpdateAddress`
- `DeleteAddress`
- `GetSellerProfile`
- `UpdateSellerProfile`
- `UpdateUserStatus`

### REST APIs

- `GET /api/v1/me`
- `PATCH /api/v1/me`
- `GET /api/v1/me/addresses`
- `POST /api/v1/me/addresses`
- `PATCH /api/v1/me/addresses/{address_id}`
- `DELETE /api/v1/me/addresses/{address_id}`
- `GET /api/v1/sellers/me`
- `PATCH /api/v1/sellers/me`

### Internal Logic

- Auth Service signup ke baad User Service me profile create karega.
- Address limit enforce hoga, for example max 20 addresses per user.
- Seller profile create hone ke baad Superadmin approval required ho sakta hai.
- KYC files object storage/CDN me rahenge, DB me metadata and verification status store hoga.

## Auth Service

### Purpose

Authentication, authorization, tokens, OTP, and role assignments manage karna.

### Responsibilities

- Signup and login.
- Password hashing and credential validation.
- JWT access token issue.
- Refresh token rotation.
- Email and phone OTP.
- RBAC role mapping.
- Logout and token revoke.

### Database Choice

MySQL plus Redis. MySQL for credentials and token audit. Redis for rate limits, OTP retry counters, and short-lived auth state.

### Tables

- `auth_accounts`
- `credentials`
- `refresh_tokens`
- `otp_challenges`
- `role_assignments`

### gRPC Services

- `Register`
- `Login`
- `RefreshToken`
- `Logout`
- `VerifyAccessToken`
- `CreateOTPChallenge`
- `VerifyOTP`
- `AssignRole`
- `RevokeRole`
- `GetUserRoles`

### REST APIs

- `POST /api/v1/auth/signup`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/otp/send`
- `POST /api/v1/auth/otp/verify`
- `POST /api/v1/auth/password/forgot`
- `POST /api/v1/auth/password/reset`

### Internal Logic

- Passwords Argon2id or bcrypt se hash honge.
- Access token TTL short, for example 15 minutes.
- Refresh token rotation enabled. Old token reuse fraud signal hoga.
- OTP challenge hash store hoga, plain OTP DB me nahi.
- Role claims JWT me honge, but critical permission checks service side role lookup se verify ho sakte hain.

## Product Service

### Purpose

Catalog, categories, variants, pricing snapshot, inventory, and product publication manage karna.

### Responsibilities

- Seller product CRUD.
- Product variants and attributes.
- Category tree.
- Inventory reserve/release/decrement.
- Product publish/unpublish workflow.
- Emit product index events.

### Database Choice

MongoDB. Reason: product attributes category-wise dynamic hote hain. Fashion, electronics, grocery sabka schema alag ho sakta hai.

### Collections

- `products`
- `categories`
- `brands`
- `inventory_snapshots`
- `price_books`

### gRPC Services

- `CreateProduct`
- `UpdateProduct`
- `PublishProduct`
- `UnpublishProduct`
- `GetProduct`
- `BatchGetProducts`
- `ListProducts`
- `ReserveInventory`
- `ReleaseInventory`
- `CommitInventory`
- `UpdatePrice`

### REST APIs

- `GET /api/v1/products`
- `GET /api/v1/products/{product_id}`
- `GET /api/v1/categories`
- `POST /api/v1/seller/products`
- `PATCH /api/v1/seller/products/{product_id}`
- `POST /api/v1/seller/products/{product_id}/publish`

### Internal Logic

- Product writes seller ownership validate karenge.
- Published product immutable snapshot search index me jayega.
- Inventory reserve checkout ke time short TTL ke saath hoga.
- Product detail read path cache-friendly hoga.

## Cart Service

### Purpose

User and guest carts manage karna.

### Responsibilities

- Add/update/remove cart items.
- Guest cart and logged-in cart merge.
- Cart totals and coupon preview.
- Cart expiration.

### Database Choice

MongoDB plus Redis. Mongo durable cart snapshot ke liye, Redis hot cart cache and fast badge count ke liye.

### Collections

- `carts`
- `cart_audit_events`

### gRPC Services

- `GetCart`
- `AddItem`
- `UpdateItemQuantity`
- `RemoveItem`
- `ClearCart`
- `MergeGuestCart`
- `ApplyCouponPreview`
- `GetCartSummary`

### REST APIs

- `GET /api/v1/cart`
- `POST /api/v1/cart/items`
- `PATCH /api/v1/cart/items/{item_id}`
- `DELETE /api/v1/cart/items/{item_id}`
- `POST /api/v1/cart/merge`
- `POST /api/v1/cart/coupons/preview`

### Internal Logic

- Cart item me product title, image, price snapshot store hoga for stable UX.
- Checkout se pehle Order Service fresh price and stock validate karega.
- Guest cart session id based hoga.

## Wishlist Service

### Purpose

User wishlist and save-for-later flows manage karna.

### Responsibilities

- Add/remove wishlist items.
- Move wishlist item to cart.
- Track price drop and availability.
- Emit wishlist analytics events.

### Database Choice

MongoDB. Reason: wishlist document simple, flexible, and read-heavy hai.

### Collections

- `wishlists`
- `wishlist_events`

### gRPC Services

- `GetWishlist`
- `AddWishlistItem`
- `RemoveWishlistItem`
- `MoveToCart`
- `CheckWishlistStatus`

### REST APIs

- `GET /api/v1/wishlist`
- `POST /api/v1/wishlist/items`
- `DELETE /api/v1/wishlist/items/{product_id}`
- `POST /api/v1/wishlist/items/{product_id}/move-to-cart`

### Internal Logic

- Same product duplicate add nahi hoga.
- Product deleted/out-of-stock events consume karke item status update hoga.
- Price drop notification opt-in respect karega.

## Order Service

### Purpose

Order lifecycle, order items, fulfillment, cancellation, and order history manage karna.

### Responsibilities

- Cart to order conversion.
- Order status machine.
- Multi-seller item grouping.
- Inventory reservation coordination.
- Payment coordination.
- Cancellation and refund trigger.

### Database Choice

MySQL. Reason: order data transactional, relational, auditable, and financial workflows ke saath linked hai.

### Tables

- `orders`
- `order_items`
- `order_status_history`
- `shipments`
- `order_idempotency_keys`

### gRPC Services

- `CreateOrderFromCart`
- `GetOrder`
- `ListOrders`
- `ListSellerOrders`
- `CancelOrder`
- `MarkOrderPaid`
- `UpdateFulfillment`
- `GetOrderForPayment`

### REST APIs

- `POST /api/v1/orders/checkout`
- `GET /api/v1/orders`
- `GET /api/v1/orders/{order_id}`
- `POST /api/v1/orders/{order_id}/cancel`
- `GET /api/v1/seller/orders`
- `PATCH /api/v1/seller/orders/{order_id}/fulfillment`

### Internal Logic

- Order create idempotency key required.
- Order item price snapshot immutable rahega.
- Payment success webhook ke baad `paid` status set hoga.
- Cancellation status and payment status ke basis pe allowed hoga.

## Payment Service

### Purpose

Payment provider abstraction, payment intents, webhooks, refunds, and reconciliation manage karna.

### Responsibilities

- Payment intent create.
- External provider integration.
- Webhook signature verification.
- Payment status update.
- Refunds and partial refunds.
- Reconciliation reports.

### Database Choice

MySQL. Reason: financial records transactional, auditable, and strongly consistent hone chahiye.

### Tables

- `payments`
- `payment_attempts`
- `refunds`
- `payment_webhook_events`
- `payment_reconciliations`

### gRPC Services

- `CreatePaymentIntent`
- `GetPayment`
- `ListPaymentsForOrder`
- `RefundPayment`
- `GetRefund`
- `MarkPaymentFailed`

### REST APIs

- `POST /api/v1/payments/{payment_id}/retry`
- `POST /api/v1/payments/{payment_id}/refund`
- `POST /api/v1/webhooks/payments/{provider}`
- `GET /api/v1/admin/payments`

### Internal Logic

- Provider-specific implementation interface ke peeche rahega.
- Webhook idempotency provider event id se enforce hoga.
- Payment status frontend callback se final nahi maana jayega. Provider webhook source of truth hoga.
- Refund creates `RefundRequested`, then provider call, then webhook/response update.

## Recommendation Service

### Purpose

Trending, similar, personalized, and campaign-based recommendations serve karna.

### Responsibilities

- User behavior event ingestion.
- Product interaction scoring.
- Trending lists.
- Similar product suggestions.
- Personalized product ranking.

### Database Choice

MongoDB plus Redis. Mongo flexible interaction features and generated recommendation sets ke liye. Redis low-latency cached recommendations ke liye.

### Collections

- `user_interactions`
- `product_features`
- `recommendation_sets`
- `ab_test_assignments`

### gRPC Services

- `GetRecommendations`
- `GetSimilarProducts`
- `GetTrendingProducts`
- `TrackRecommendationImpression`
- `TrackRecommendationClick`

### REST APIs

- `GET /api/v1/recommendations`
- `GET /api/v1/products/{product_id}/similar`

### Internal Logic

- MVP rule-based: trending by category, recent purchases, similar attributes.
- Later ML can add embeddings/vector scoring.
- Cold-start fallback: category popular and global trending.

## Search Service (Typesense)

### Purpose

Fast product search, facets, sorting, autocomplete, and zero-result analytics manage karna.

### Responsibilities

- Typesense schema create/update.
- Product index updates.
- Full reindex.
- Search query execution.
- Autocomplete and facets.
- Synonyms and merchandising rules.

### Database Choice

Typesense as search index. Service may use small MongoDB/MySQL config store if synonyms and rules become complex, but source of truth for product remains Product Service.

### Typesense Collections

- `products`
- `popular_queries`

### gRPC Services

- `SearchProducts`
- `Autocomplete`
- `ReindexProduct`
- `DeleteProductFromIndex`
- `CreateSynonym`
- `ListSynonyms`

### REST APIs

- `GET /api/v1/search`
- `GET /api/v1/search/autocomplete`
- `POST /api/v1/admin/search/reindex`
- `POST /api/v1/admin/search/synonyms`

### Internal Logic

- Search result product ids and ranking Typesense se aayenge.
- Canonical product availability Product Service se verify ho sakti hai.
- Product update events eventually consistent index maintain karenge.

## CMS Service

### Purpose

Seller dashboard backend: product workflow, coupons, campaigns, seller settings, seller analytics.

### Responsibilities

- Seller product moderation workflow.
- Coupon validation.
- Offer and campaign management.
- Seller dashboard metrics.
- Seller staff permissions.
- Audit logs.

### Database Choice

MySQL. Reason: coupons, campaigns, permissions, and audit logs structured relational data hain.

### Tables

- `seller_settings`
- `seller_staff`
- `coupons`
- `coupon_rules`
- `coupon_redemptions`
- `campaigns`
- `cms_audit_logs`

### gRPC Services

- `ValidateCoupon`
- `CreateCoupon`
- `UpdateCoupon`
- `ListCoupons`
- `GetSellerSettings`
- `UpdateSellerSettings`
- `GetSellerAnalytics`
- `RecordCouponRedemption`

### REST APIs

- `GET /api/v1/seller/dashboard/summary`
- `GET /api/v1/seller/coupons`
- `POST /api/v1/seller/coupons`
- `PATCH /api/v1/seller/coupons/{coupon_id}`
- `GET /api/v1/seller/campaigns`
- `POST /api/v1/seller/campaigns`

### Internal Logic

- Coupon validation side-effect free hoga.
- Coupon redemption final Order Service payment success ke baad record karega.
- Seller analytics can query read model or pre-aggregated data.

## Session Management Service (Independent)

### Purpose

User activity, journey, session, device tracking, analytics, and heatmap data independent system ke roop me manage karna.

### Responsibilities

- Anonymous and logged-in sessions.
- Page view, click, scroll, add-to-cart, checkout events.
- Journey reconstruction.
- Live active session tracking.
- Funnel and cohort analytics.
- Heatmap aggregation.
- Privacy and retention.

### Database Choice

MongoDB plus Redis. Events high-volume and flexible hain. Redis active sessions and recent counters ke liye.

### Collections

- `sessions`
- `session_events`
- `journey_summaries`
- `heatmap_points`
- `analytics_aggregates`

### gRPC Services

- `GetSession`
- `ListSessions`
- `GetJourney`
- `GetLiveMetrics`
- `GetFunnelReport`
- `GetHeatmap`
- `DeleteUserSessionData`

### REST APIs

- `POST /api/v1/sessions/events`
- `GET /api/v1/analytics/live`
- `GET /api/v1/analytics/sessions`
- `GET /api/v1/analytics/sessions/{session_id}/journey`
- `GET /api/v1/analytics/funnels`
- `GET /api/v1/analytics/heatmaps`

### Internal Logic

- Client SDK assigns anonymous id.
- Login ke baad anonymous id user id se link hota hai.
- Raw events TTL expire ho sakte hain, aggregates long-term store honge.
- PII masking and deletion support mandatory hai.

## Notification Service

### Purpose

Email, SMS, push notifications, templates, delivery logs, retries manage karna.

### Responsibilities

- OTP email/SMS.
- Order status notification.
- Payment success/failure notification.
- Price drop notification.
- Template management.
- Delivery retry and DLQ.

### Database Choice

MongoDB. Reason: templates, provider payloads, delivery metadata flexible format me aate hain.

### Collections

- `notification_templates`
- `notification_deliveries`
- `notification_preferences`
- `provider_events`

### gRPC Services

- `SendOTP`
- `SendNotification`
- `GetDeliveryStatus`
- `UpdateNotificationPreference`
- `RenderTemplate`

### REST APIs

- `GET /api/v1/me/notification-preferences`
- `PATCH /api/v1/me/notification-preferences`
- `POST /api/v1/admin/notifications/templates`

### Internal Logic

- Events consume karke async notification send hoga.
- Retry exponential backoff ke saath.
- Provider response raw payload secure log/store hoga.

## API Gateway

### Purpose

Public API entry point, REST to gRPC bridge, auth enforcement, rate limiting, and response shaping.

### Responsibilities

- REST route handling.
- gRPC client orchestration.
- JWT validation and RBAC.
- Request validation.
- Rate limiting.
- CORS.
- Error mapping.
- Aggregation for frontend convenience.

### Database Choice

No primary DB. Redis for rate limiting and temporary request metadata.

### gRPC Clients

- Auth
- User
- Product
- Cart
- Wishlist
- Order
- Payment
- Search
- CMS
- Session
- Notification
- Superadmin

### REST APIs

Gateway exposes almost all public REST APIs from `api/master-api.json`.

### Internal Logic

- Gateway does not own business rules.
- Gateway can aggregate simple read responses, for example product detail plus recommendations.
- All protected requests receive user context.
- All responses follow common envelope.

## Superadmin Service

### Purpose

Platform-level control plane for users, sellers, orders, payments, sessions, settings, and audits.

### Responsibilities

- Admin user and permission management.
- User block/unblock.
- Seller approval/suspension.
- Refund review.
- Platform settings.
- Cross-service admin workflows.
- Immutable admin audit logs.

### Database Choice

MySQL. Reason: admin permissions, settings, and audit trails structured and compliance-sensitive hain.

### Tables

- `admin_users`
- `admin_permissions`
- `admin_role_permissions`
- `platform_settings`
- `admin_audit_logs`
- `admin_review_tasks`

### gRPC Services

- `ListUsersForAdmin`
- `UpdateUserStatus`
- `ListSellersForAdmin`
- `UpdateSellerStatus`
- `ReviewRefund`
- `GetPlatformSettings`
- `UpdatePlatformSetting`
- `ListAuditLogs`

### REST APIs

- `GET /api/v1/admin/users`
- `PATCH /api/v1/admin/users/{user_id}/status`
- `GET /api/v1/admin/sellers`
- `PATCH /api/v1/admin/sellers/{seller_id}/status`
- `GET /api/v1/admin/orders`
- `GET /api/v1/admin/payments`
- `POST /api/v1/admin/refunds/{refund_id}/review`
- `GET /api/v1/admin/audit-logs`
- `GET /api/v1/admin/settings`
- `PATCH /api/v1/admin/settings/{key}`

### Internal Logic

- Every admin mutation writes audit log in same transaction where possible.
- Superadmin APIs call downstream services through gRPC with admin context.
- High-risk actions may require maker-checker approval.

