# Project Dependency & Setup Guide

## Document Variables

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `CMS Service` |
| `TASK_FILE_NAME` | `task4.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task4_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}` |

This guide documents dependency, setup, environment, database, and DevOps requirements for `INPUT_FILE_PATH`.

Simple goal: beginner developer ko clear ho ki coupon engine MVP run/test karne ke liye kya setup chahiye, kya already previous dependency docs me covered hai, aur `TASK_FILE_NAME` ke liye kaunsi extra coupon-specific checks karni hain.

Important: original implementation task file is not modified. This is a new dependency/setup document saved at `OUTPUT_FILE_PATH`.

## 1. Project Overview

`TASK_FILE_NAME` coupon engine MVP define karta hai. Current backend implementation under `backend/services/cms-service` me coupon domain, usecase, MySQL repository, HTTP handlers, gRPC validation handler, tests, and migration already present hain.

| Area | Decision |
|---|---|
| Source of truth | MySQL tables: `coupons`, `coupon_rules`, `coupon_redemptions` |
| Main workflows | Seller coupon manage, cart/order coupon validate, order redemption record |
| Discount types | `fixed`, `percentage` |
| Scope rules | `product_scope`, `category_scope`, `seller_scope` |
| Runtime API | HTTP internal routes plus gRPC `ValidateCoupon` |
| Preview behavior | Side-effect free; preview does not insert redemption |
| Final redemption | Idempotent insert using `(coupon_id, order_id)` unique key |

Hinglish explanation: Coupon engine ka kaam hai coupon code ko validate karna, eligible cart items par discount calculate karna, aur final paid order ke baad redemption history save karna. Cart preview ke time usage count badhna nahi chahiye.

## 2. Tech Stack

Most tech stack setup is already explained in previous dependency files. Do not duplicate installation.

| Technology/service | Status for `TASK_FILE_NAME` | What to do |
|---|---|---|
| Go | Reused | Refer `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`, section `2. Tech Stack Analysis`. |
| Go modules | Reused | Same `go.mod` and `go.sum`; no new `go get` needed for this task. |
| `net/http` | Reused with task-specific routes | Coupon seller/internal HTTP routes run on existing HTTP server. |
| gRPC | Reused with task-specific method | `ValidateCoupon` is exposed for internal `cart-service` and `order-service` callers. |
| Protocol Buffers | Reused | `proto/ecommerce/cms/v1/cms.proto` defines `ValidateCouponRequest` and response messages. |
| MySQL 8+ | Reused with task-specific tables | Coupon tables come from `003_create_coupon_engine_tables.up.sql`. |
| `github.com/go-sql-driver/mysql` | Reused | Existing MySQL driver; no new package install. |
| API Gateway/Auth headers | Reused | Seller coupon management requires actor headers and coupon permissions. |
| Redis/Kafka/RabbitMQ | Not used | No local setup required for this task. Future cache/queue ideas are not current runtime dependencies. |

Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`2. Tech Stack Analysis`
`3. Go Dependency System`
`10. HTTP and gRPC Endpoint Setup Notes`

## 3. Required Software

Follow previous setup first. Is task ke liye fresh install list repeat karne ki zarurat nahi hai.

| Software | Required? | New for `TASK_FILE_NAME`? | Reference |
|---|---:|---:|---|
| Go compatible with `go.mod` | Yes | No | `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`, section `3. Go Dependency System` |
| MySQL 8+ | Yes | No, but coupon tables are task-specific | `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`, section `4. Database Analysis` |
| MySQL CLI client | Strongly recommended | No | Same previous section |
| Docker | Optional | No | `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`, section `8. Docker and DevOps Setup` |
| `grpcurl` | Optional | Optional task-specific testing helper | Useful only if you want to manually call gRPC `ValidateCoupon`. |
| Cart Service / Order Service | Not required to start this service | Required for full platform flow | They call this service; this service does not need their databases for local coupon smoke tests. |

## 4. Dependency Management

No new Go dependency is introduced by `TASK_FILE_NAME`.

Current direct dependencies remain:

| Dependency | Why it matters |
|---|---|
| `github.com/go-sql-driver/mysql v1.9.3` | MySQL connection and coupon repository queries. |
| `google.golang.org/grpc v1.77.0` | Internal gRPC server for `ValidateCoupon`. |
| `google.golang.org/protobuf v1.36.10` | Generated protobuf messages for gRPC requests/responses. |

Reuse dependency commands:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`3. Go Dependency System`
`9. Complete Project Run Instructions`
`15. Minimal Local Command Flow`
```

Beginner note: `go.mod` dependency menu hai, `go.sum` checksum safety file hai. Coupon task ke liye `go mod tidy`, `go mod download`, `go test ./...`, and `go run ./cmd/server` ka same flow use hoga.

## 5. Database Setup

### 5.1 Database used

| Database | Required? | Default port | Status |
|---|---:|---:|---|
| MySQL 8+ | Yes | 3306 | Reused setup with task-specific coupon tables |

MySQL installation, Docker command, DB user creation, and full migration flow are already documented.

Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`4. Database Analysis`
`8. Docker and DevOps Setup`
`9. Complete Project Run Instructions`

Refer:
`TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`

Sections:
`5. Database Setup`
`7. Environment Variables`

### 5.2 Task-specific migration

Coupon engine tables are created by:

```text
backend/services/cms-service/migrations/003_create_coupon_engine_tables.up.sql
```

This migration creates:

| Table | Purpose |
|---|---|
| `coupons` | Master coupon data: code, seller, discount type, amount, status, usage limits, validity window. |
| `coupon_rules` | Flexible JSON rules for product/category/seller scope. |
| `coupon_redemptions` | Final paid-order usage history and idempotency record. |

Important: local service startup initializes many repositories, not only coupons. For a clean run, apply all service migrations in numeric order as already shown in `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`.

### 5.3 Important coupon keys and constraints

| Table | Key/constraint | Why it matters |
|---|---|---|
| `coupons` | `uk_coupons_coupon_id` | Stable internal coupon id duplicate nahi hona chahiye. |
| `coupons` | `uk_coupons_code` | Same coupon code duplicate nahi ho sakta. |
| `coupons` | `idx_coupons_seller_status` | Seller dashboard list/filter fast hota hai. |
| `coupons` | `idx_coupons_status_window` | Active/time-window filtering fast hota hai. |
| `coupon_rules` | `fk_coupon_rules_coupon` | Coupon delete/update ke saath rules consistent rehte hain. |
| `coupon_rules` | `chk_coupon_rules_value_json` | Rule value valid JSON hona chahiye. |
| `coupon_redemptions` | `uk_coupon_redemptions_coupon_order` | Same coupon same order ke liye double redeem nahi hota. |
| `coupon_redemptions` | `idx_coupon_redemptions_user_coupon` | Per-user usage limit check fast hota hai. |

### 5.4 Task-specific DB verification

After running migrations:

```bash
cd backend/services/cms-service
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW TABLES LIKE 'coupon%';" cms_db
```

Expected:

```text
coupon_redemptions
coupon_rules
coupons
```

Check indexes:

```bash
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW INDEX FROM coupons; SHOW INDEX FROM coupon_redemptions;" cms_db
```

Look for:

```text
uk_coupons_code
idx_coupons_seller_status
uk_coupon_redemptions_coupon_order
idx_coupon_redemptions_user_coupon
```

### 5.5 Credentials placement

No new DB credentials are introduced by `TASK_FILE_NAME`.

Use the existing local env location:

```text
backend/services/cms-service/.env
```

Important DB variables are already explained in previous files:

| Variable | Why it matters for this task |
|---|---|
| `CMS_MYSQL_DSN` | If set, it must point to DB where coupon tables exist. |
| `CMS_DB_HOST` | MySQL host if DSN is not set. |
| `CMS_DB_PORT` | MySQL port, usually `3306`. |
| `CMS_DB_NAME` | Should be `cms_db` for documented local setup. |
| `CMS_DB_USER` | App DB user. Avoid root for runtime. |
| `CMS_DB_PASSWORD` | App DB password. Keep local only, do not commit. |
| `CMS_DB_TIMEZONE` | Use `UTC` so coupon start/end windows remain predictable. |

## 6. Redis / Queue / External Services

No Redis, Kafka, RabbitMQ, NATS, MinIO, Elasticsearch, SMTP, Stripe, Twilio, Firebase, or OAuth provider is introduced by `TASK_FILE_NAME`.

| Service | Required for local coupon smoke test? | Notes |
|---|---:|---|
| MySQL | Yes | Coupon source of truth. |
| API Gateway/Auth layer | Simulated locally with headers | Seller routes need `X-User-ID`, `X-Seller-ID`, `X-Roles`, and internal token when configured. |
| Cart Service | No | In full platform, Cart calls gRPC/HTTP validation with cart snapshot. |
| Order Service | No | In full platform, Order records redemption after payment success. |
| Product Service | No for coupon validation | Cart item snapshots include product/category/seller ids, so this service does not query Product Service for coupon validation. |
| Campaign table/repository | Optional coupling | Current code accepts `campaign_id`; if used, campaign migration/table must also exist. |

Beginner note: Redis cache future me coupon validation speed improve kar sakta hai, but current code MySQL se directly validate karta hai. Redis ko source of truth mat banao.

## 7. Environment Variables

No brand-new coupon-only environment variable is introduced by `TASK_FILE_NAME`.

Do not duplicate the full `.env`. Use the complete `.env` from:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Section:
`5. Environment Variables`
```

Task-specific variables to verify:

| Variable | Required? | Example | Purpose |
|---|---:|---|---|
| `CMS_HTTP_ADDR` | Yes | `:8087` | Internal HTTP coupon management/validation routes. |
| `CMS_GRPC_ADDR` | Yes | `:9098` | gRPC `ValidateCoupon` endpoint for Cart/Order. |
| `CMS_GRPC_ALLOWED_INTERNAL_CALLERS` | Yes for gRPC allowlist | `cart-service,order-service,api-gateway` | Only trusted internal callers should call gRPC methods. |
| `CMS_INTERNAL_AUTH_HEADER` | Yes | `X-Internal-Token` | Header name checked by internal HTTP and gRPC calls. |
| `CMS_INTERNAL_AUTH_TOKEN` | Required in production, recommended locally | `local-cms-internal-token` | Shared internal token. Do not commit real value. |
| `CMS_MAX_BODY_BYTES` | Optional | `1048576` | Protects coupon create/validate JSON body size. |
| `CMS_DB_*` / `CMS_MYSQL_DSN` | Yes | See previous file | DB connection where coupon tables exist. |

How project loads env: current code uses `os.Getenv`. It does not auto-load `.env`, so source it before running:

```bash
cd backend/services/cms-service
set -a
source .env
set +a
go run ./cmd/server
```

Common mistake: `.env` file create kar dena enough nahi hai. Shell me `source .env` karna zaruri hai unless your process manager loads it.

## 8. Docker Setup

No new Dockerfile, Docker Compose service, volume, network, healthcheck, or restart policy is introduced by `TASK_FILE_NAME`.

Reuse:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`4. Database Analysis` -> MySQL Docker setup
`8. Docker and DevOps Setup`
```

Task-specific Docker reminder:

| Scenario | Correct value |
|---|---|
| Service runs on host, MySQL runs in Docker | `CMS_DB_HOST=localhost` |
| Service and MySQL run in same Compose network | `CMS_DB_HOST=<mysql-service-name>` |
| Cart/Order call this service in Compose | They should target the actual Compose service/container DNS name and the port configured by `CMS_GRPC_ADDR`. |

## 9. Local Development Setup

### Step 1: Read previous dependency documentation

Follow these first:

```text
TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
TaskImplementation/${SERVICE_NAME}/task2_Dependency.md
TaskImplementation/${SERVICE_NAME}/task3_Dependency.md
```

### Step 2: Go to project directory

```bash
cd backend/services/cms-service
```

### Step 3: Install only new dependencies

No new dependency install is required for `TASK_FILE_NAME`.

Optional verification:

```bash
go mod download
go test ./...
```

### Step 4: Setup only new database/service changes

If all migrations from previous dependency docs are already applied, do not rerun them.

If coupon tables are missing, run the full migration flow from `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`. At minimum, coupon MVP needs:

```bash
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/003_create_coupon_engine_tables.up.sql
```

Warning: run migrations in numeric order on a fresh DB. Running only migration `003` on an empty DB can miss shared tables like `cms_audit_logs`, which coupon create/update audit writes need.

### Step 5: Add only new or changed env variables

No new coupon-only env variable. Verify `CMS_HTTP_ADDR`, `CMS_GRPC_ADDR`, `CMS_INTERNAL_AUTH_TOKEN`, and DB variables from section `7`.

### Step 6: Start backend service

```bash
set -a
source .env
set +a
go run ./cmd/server
```

Expected logs:

```text
cms.mysql.connected
cms.http.started
cms.grpc.started
```

### Step 7: Verify service health

```bash
curl http://localhost:8087/healthz
```

Expected:

```json
{"success":true}
```

### Step 8: Verify coupon create/list HTTP flow

Use seller headers because seller coupon management is permission protected.

```bash
curl -X POST http://localhost:8087/internal/v1/cms/seller/coupons \
  -H "Content-Type: application/json" \
  -H "X-Internal-Token: local-cms-internal-token" \
  -H "X-User-ID: user_seller_1" \
  -H "X-Seller-ID: seller_1" \
  -H "X-Roles: seller_catalog_editor" \
  -H "X-Staff-Status: active" \
  -H "X-Request-ID: req_coupon_create_1" \
  -d '{
    "code": "SAVE250",
    "discount_type": "fixed",
    "discount_value": 25000,
    "min_cart_amount": {"amount": 99900, "currency": "INR"},
    "currency": "INR",
    "status": "active",
    "rules": [
      {
        "rule_type": "seller_scope",
        "rule_value": {"seller_ids": ["seller_1"]}
      }
    ]
  }'
```

Then list:

```bash
curl "http://localhost:8087/internal/v1/cms/seller/coupons?limit=10&offset=0" \
  -H "X-Internal-Token: local-cms-internal-token" \
  -H "X-User-ID: user_seller_1" \
  -H "X-Seller-ID: seller_1" \
  -H "X-Roles: seller_catalog_editor" \
  -H "X-Staff-Status: active"
```

If your local `.env` does not set `CMS_INTERNAL_AUTH_TOKEN`, remove the `X-Internal-Token` header from local curl commands. Production should set the token.

### Step 9: Verify coupon validation HTTP flow

```bash
curl -X POST http://localhost:8087/internal/v1/cms/coupons/validate \
  -H "Content-Type: application/json" \
  -H "X-Internal-Token: local-cms-internal-token" \
  -H "X-Request-ID: req_coupon_validate_1" \
  -d '{
    "coupon_code": "SAVE250",
    "user_id": "buyer_1",
    "cart_id": "cart_1",
    "currency": "INR",
    "subtotal_amount": 149900,
    "items": [
      {
        "product_id": "prod_1",
        "seller_id": "seller_1",
        "category_ids": ["cat_shoes"],
        "quantity": 1,
        "unit_amount": 149900,
        "line_subtotal_amount": 149900
      }
    ]
  }'
```

Expected valid response contains:

```json
{
  "valid": true,
  "coupon_id": "...",
  "discount": {"amount": 25000, "currency": "INR"}
}
```

### Step 10: Verify redemption after successful order

Use the `coupon_id` returned by create/list/validate.

```bash
curl -X POST http://localhost:8087/internal/v1/cms/coupons/redemptions \
  -H "Content-Type: application/json" \
  -H "X-Internal-Token: local-cms-internal-token" \
  -H "X-Request-ID: req_coupon_redeem_1" \
  -d '{
    "coupon_id": "coupon_replace_me",
    "order_id": "order_1",
    "user_id": "buyer_1",
    "discount": {"amount": 25000, "currency": "INR"}
  }'
```

Same request repeat karne par same redemption idempotently return ho sakti hai. Same `(coupon_id, order_id)` with different user/amount conflict dega.

## 10. Running the Project

Complete shared run command flow is already documented in `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`.

Task-specific run checklist:

| Check | Command / place |
|---|---|
| Coupon tables exist | `SHOW TABLES LIKE 'coupon%';` |
| Audit table exists | `SHOW TABLES LIKE 'cms_audit_logs';` |
| HTTP server running | `curl http://localhost:8087/healthz` |
| gRPC server listening | `CMS_GRPC_ADDR`, default `:9098` |
| Internal caller allowlist correct | `CMS_GRPC_ALLOWED_INTERNAL_CALLERS` includes `cart-service` and `order-service` |
| Seller role can manage coupon | Use `X-Roles: seller_catalog_editor`, `seller_manager`, or `seller` |
| Cart/Order snapshot has required fields | `coupon_code`, `user_id`, `currency`, `subtotal_amount` |

Optional gRPC smoke test with `grpcurl`:

```bash
grpcurl -plaintext \
  -H "x-service-name: cart-service" \
  -H "x-internal-token: local-cms-internal-token" \
  -H "x-request-id: req_grpc_coupon_1" \
  -d '{
    "coupon_code": "SAVE250",
    "user_id": "buyer_1",
    "cart_id": "cart_1",
    "currency": "INR",
    "cart_subtotal": {"amount": 149900, "currency": "INR"},
    "items": [
      {
        "product_id": "prod_1",
        "seller_id": "seller_1",
        "category_id": "cat_shoes",
        "quantity": 1,
        "line_total": {"amount": 149900, "currency": "INR"},
        "unit_price": {"amount": 149900, "currency": "INR"}
      }
    ]
  }' \
  localhost:9098 ecommerce.cms.v1.CMSService/ValidateCoupon
```

If `CMS_INTERNAL_AUTH_TOKEN` is empty locally, remove the token metadata. Keep `x-service-name: cart-service`; gRPC requires internal service metadata.

## 11. Common Errors & Fixes

Generic errors like Go install, MySQL install, Docker daemon, and `.env` loading are already covered in previous docs. Only task-specific issues are listed here.

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `Coupon code already exists` / `COUPON_CODE_EXISTS` | `coupons.code` unique key already has same normalized code. | Use a new code or update existing coupon. | Normalize code mentally: `save250` and `SAVE250` are same. |
| `Coupon not found` | Validation code does not exist in DB. | Create coupon first or check code spelling. | Use list endpoint before validation. |
| `coupon_not_active` | Coupon status is `draft`, `paused`, or `expired`. | Set status to `active` when testing. | Test fixtures should explicitly set active status. |
| `coupon_not_started` / `coupon_expired` | `starts_at` or `ends_at` window does not include current UTC time. | Fix date window in RFC3339/UTC. | Keep local test coupons without a narrow time window. |
| `currency_mismatch` | Request currency differs from coupon currency. | Send `INR` if coupon is `INR`, or create matching coupon. | Keep all local money values in one currency. |
| `min_cart_not_met` | `subtotal_amount` is lower than `min_cart_amount`. | Increase subtotal or lower coupon min cart. | Remember amounts are minor units: INR 999.00 is `99900`. |
| `scope_not_matched` | Cart item seller/product/category does not match coupon rules. | Send matching `seller_id`, `product_id`, or category. | Add realistic cart snapshots in tests. |
| `USAGE_LIMIT_REACHED` / `PER_USER_LIMIT_REACHED` | Existing redemption counts reached configured limits. | Use a new coupon/user/order in local test. | Keep low-limit coupons separate from reusable smoke-test coupons. |
| `REDEMPTION_CONFLICT` | Same coupon/order already redeemed with different user, campaign, amount, or currency. | Reuse exact same payload or create a new order id. | Treat `(coupon_id, order_id)` as idempotency key. |
| gRPC `internal service metadata is required` | Missing `x-service-name`. | Add `x-service-name: cart-service` or `order-service`. | Put metadata in grpc client interceptor. |
| gRPC `internal service is not allowed for this method` | Caller not in method allowlist. | Use `cart-service` or `order-service` for `ValidateCoupon`, and verify `CMS_GRPC_ALLOWED_INTERNAL_CALLERS`. | Keep allowlist aligned with platform service names. |
| `AUDIT_WRITE_FAILED` during create/update/redemption | Audit table missing or DB write failed. | Run shared migration `001` and audit hardening migration if needed. | Always run migrations in numeric order on fresh DB. |

## 12. Security & Best Practices

Task-specific security notes:

| Area | Recommendation |
|---|---|
| Internal auth | Set `CMS_INTERNAL_AUTH_TOKEN` outside local quick tests. Production me empty token unsafe hai. |
| gRPC metadata | Only trusted services should send `x-service-name`. Do not expose gRPC port publicly. |
| Seller headers | `X-User-ID`, `X-Seller-ID`, `X-Roles`, and `X-Staff-Status` should come from trusted gateway, not public clients. |
| Money values | Always use minor units as integers. Floating point discount math avoid karo. |
| Coupon preview | Preview should stay side-effect free. Redemption sirf successful payment/order ke baad record karo. |
| Idempotency | Use `(coupon_id, order_id)` as duplicate guard. Retry same payload only. |
| Scope rules | Cart/Order should send trusted product/category/seller snapshot. Buyer input directly trust mat karo. |
| Secrets | `.env` local-only rakho; production me secret manager use karo. |
| Audit logs | Coupon create/update/disable/redemption actions audit me preserve hone chahiye. |

## 13. Missing or Misconfigured Things

| Observation | Impact | Suggested fix |
|---|---|---|
| No dedicated Docker Compose file for this service | Beginners must use previous MySQL Docker command manually. | Add a local infra compose later if team wants one-command startup. |
| `.env` is not auto-loaded by code | Service may start with defaults or missing secrets. | Source `.env` before `go run`, or use a process manager that loads env. |
| Coupon HTTP validate/redemption endpoints are internal but still exposed on HTTP server | Public exposure would be risky without gateway/network controls. | Keep service behind internal network/API Gateway and set internal token. |
| gRPC `ValidateCoupon` only allows `cart-service` and `order-service` at method level | Manual grpc tests fail with random service names. | Use expected metadata or update method allowlist intentionally. |
| Product/category/seller snapshot correctness depends on caller | Wrong snapshots can produce wrong eligibility. | Cart/Order should build snapshot from trusted cart/product data, not raw buyer fields. |
| Campaign coupling exists when `campaign_id` is sent | Missing campaign tables can break campaign-linked validation. | Run full migration set and avoid `campaign_id` unless campaign flow is configured. |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `2. Tech Stack Analysis` | Same Go, HTTP, gRPC, MySQL, logging, and gateway concepts. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `3. Go Dependency System` | No new Go dependency for `TASK_FILE_NAME`. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `4. Database Analysis` | MySQL install, Docker run, DB user creation, and migration command pattern already documented. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `5. Environment Variables` | Full `.env` is already documented; this file only lists coupon-specific variables to verify. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `7. Ports and Networking` | HTTP/gRPC/MySQL port basics are reused. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `8. Docker and DevOps Setup` | No new container/volume/network for coupon MVP. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `11. Common Errors and Fixes` | Generic Go/MySQL/Docker/env issues already explained. |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | Same `cms_db`, InnoDB, UTF-8, migration order, and DB verification style. |
| `TaskImplementation/${SERVICE_NAME}/task3_Dependency.md` | `6. Environment Variables` and `7. External Services Analysis` | Same internal auth header pattern and trusted gateway/header behavior. |

## 15. Final Checklist

* [ ] Previous dependency documentation checked first.
* [ ] No duplicate Go/MySQL/Docker setup copied into this file.
* [ ] MySQL server running.
* [ ] Full service migrations applied in numeric order.
* [ ] `coupons`, `coupon_rules`, and `coupon_redemptions` tables verified.
* [ ] `.env` created at `backend/services/cms-service/.env`.
* [ ] `.env` sourced before running service.
* [ ] `CMS_HTTP_ADDR` and `CMS_GRPC_ADDR` verified.
* [ ] `CMS_INTERNAL_AUTH_TOKEN` set for realistic internal-route testing.
* [ ] Service starts and `/healthz` returns success.
* [ ] Seller coupon create/list route tested with seller actor headers.
* [ ] Coupon validation tested with cart snapshot.
* [ ] Redemption tested only after validation using a unique `order_id`.
* [ ] Logs checked for `cms.coupon.*` messages.
* [ ] No real credentials committed.
