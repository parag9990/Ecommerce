# Project Dependency & Setup Guide

## Document Variables

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `CMS Service` |
| `TASK_FILE_NAME` | `task6.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task6_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}` |
| `BACKEND_SERVICE_PATH` | `backend/services/cms-service` |
| `ANALYTICS_MIGRATION_PATH` | `${BACKEND_SERVICE_PATH}/migrations/005_create_seller_analytics_tables.up.sql` |

This document explains setup, dependencies, environment variables, database requirements, and DevOps notes for `${INPUT_FILE_PATH}`.

Simple goal: beginner developer ko samajh aaye ki seller analytics dashboard summary run/test karne ke liye kya setup chahiye, kaunsi cheezein previous dependency files me already documented hain, aur `${TASK_FILE_NAME}` ke liye kaunsi new analytics-specific cheezein verify karni hain.

Important: original implementation file is not modified. This is a new documentation file saved at `${OUTPUT_FILE_PATH}`.

## 1. Project Overview

`${TASK_FILE_NAME}` covers seller analytics APIs for revenue, paid orders, conversion rate, and top products.

Current backend implementation under `${BACKEND_SERVICE_PATH}` has actual seller analytics support:

| Area | File/path | Setup impact |
|---|---|---|
| Domain model | `${BACKEND_SERVICE_PATH}/internal/domain/seller_analytics.go` | Defines date range, currency, revenue, conversion, and top-product response rules. |
| Usecase | `${BACKEND_SERVICE_PATH}/internal/usecase/seller_analytics.go` | Requires Authorizer, MySQL analytics repository, logger, and `CMS_ANALYTICS_*` config. |
| Repository | `${BACKEND_SERVICE_PATH}/internal/repository/mysql_seller_analytics_repository.go` | Reads aggregate data from MySQL analytics read-model tables. |
| Migration | `${ANALYTICS_MIGRATION_PATH}` | Creates seller analytics daily tables and event dedupe table. |
| HTTP route | `${BACKEND_SERVICE_PATH}/internal/transport/http/handler.go` | Exposes `GET /api/v1/seller/dashboard/summary` and internal mirror route. |
| Config | `${BACKEND_SERVICE_PATH}/internal/config/config.go` | Adds task-specific `CMS_ANALYTICS_*` environment variables. |

Very important beginner note: ye API raw Order DB ya Payment DB ko directly query nahi karti. API MySQL me already prepared aggregate/read-model tables se data read karti hai. Agar tables empty hain, API valid response de sakti hai but metrics zero/empty aayenge.

## 2. Tech Stack

Most tech stack setup is already explained in previous dependency docs. Is file me only task-specific usage explain kiya gaya hai.

| Technology/service | Status for `${TASK_FILE_NAME}` | Why used | Beginner explanation |
|---|---|---|---|
| Go | Reused | Analytics domain, usecase, repository, HTTP handler, and config Go me implemented hain. | Go backend language hai jo fast APIs and services ke liye use hoti hai. |
| Go modules | Reused | No new direct Go package was added for analytics. | `go.mod` dependency list hai, `go.sum` checksum safety file hai. |
| `net/http` | Reused with analytics route | Seller dashboard summary HTTP API existing CMS HTTP server par mount hai. | Ye Go ka built-in HTTP server hai. |
| MySQL 8+ | Reused with new analytics tables | Analytics read-model, conversion aggregates, top products, and dedupe records yahin store hote hain. | MySQL relational database hai jisme structured data tables me store hota hai. |
| `database/sql` + MySQL driver | Reused | Repository SQL queries execute karta hai. | Driver Go code ko MySQL se connect karata hai. |
| Auth/API Gateway headers | Reused and required | Seller identity, role, staff status, and request id trusted headers se aate hain. | Gateway normally JWT validate karke CMS ko identity headers bhejta hai. Local me headers manually dene padte hain. |
| Order Service / Payment Service | Upstream data dependency | Real revenue/orders data in services se aggregate pipeline me aana chahiye. | CMS direct DB read nahi karta; order/payment facts ko read-model me sync karna hota hai. |
| Session/Analytics Service | Upstream data dependency | Conversion rate ke denominator/session counts ke liye aggregate data chahiye. | Product views and paid-order sessions precomputed table me aane chahiye. |
| Redis/Kafka/RabbitMQ | Not implemented for this task | Current CMS code does not connect to these services for analytics. | Future event pipeline ho sakta hai, but local API run ke liye abhi start karna zaruri nahi. |
| gRPC/Protobuf | Reused by service generally, but analytics gap exists | Current proto/server does not expose `GetSellerAnalytics`. | gRPC setup previous docs me covered hai; analytics currently HTTP-only in code. |

## 3. Required Software

Follow previous setup first. Do not reinstall tools if they are already working.

| Software/tool | Required? | New for `${TASK_FILE_NAME}`? | Reference |
|---|---:|---:|---|
| Go compatible with `${BACKEND_SERVICE_PATH}/go.mod` | Yes | No | `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`, section `3. Go Dependency System` |
| MySQL 8+ | Yes | No engine change, but new tables required | `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`, section `4. Database Analysis` |
| MySQL CLI client | Strongly recommended | No | Same previous DB setup section |
| Docker | Optional | No | `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`, section `8. Docker and DevOps Setup` |
| `curl` | Recommended | No | Used for local analytics route smoke test |
| `grpcurl` | Optional | No analytics use yet | Useful for other CMS gRPC methods, not for this task's HTTP analytics route |
| Buf/protobuf plugins | Optional | No | Required only if proto contract is updated later |

## 4. Dependency Management

No new Go dependency is introduced by `${TASK_FILE_NAME}`.

Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`3. Go Dependency System`
`9. Complete Project Run Instructions`
`15. Minimal Local Command Flow`

Current direct dependencies remain:

| Dependency | Why it exists |
|---|---|
| `github.com/go-sql-driver/mysql` | MySQL connection driver. |
| `google.golang.org/grpc` | CMS service gRPC server for existing internal RPCs. |
| `google.golang.org/protobuf` | Generated protobuf message support. |

Task-specific command, only to verify analytics code compiles/tests after migration/config changes:

```bash
cd backend/services/cms-service
go test ./...
go build ./cmd/server
```

If Go dependency errors happen, do not duplicate troubleshooting here. Follow:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`, section `3. Go Dependency System`.

## 5. Database Setup

### 5.1 Database used

| Database | Required? | Default port | Status |
|---|---:|---:|---|
| MySQL 8+ | Yes | 3306 | Reused setup with new seller analytics read-model tables. |

MySQL installation, Docker command, DB user creation, connection string formats, and base migration flow are already documented.

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
`11. Common Errors & Fixes`

### 5.2 Task-specific migration

Analytics support is created by:

```text
${ANALYTICS_MIGRATION_PATH}
```

This migration creates four tables:

| Table | Purpose |
|---|---|
| `seller_analytics_daily` | Daily seller-level revenue, refund, net revenue, paid orders, paid items, and cancelled orders by currency. |
| `seller_product_analytics_daily` | Daily product-level units sold, order count, revenue, and title snapshot for top-products ranking. |
| `seller_conversion_daily` | Daily seller funnel counts: product views, add-to-cart sessions, checkout sessions, and paid-order sessions. |
| `seller_analytics_event_dedupe` | Event id dedupe table for future aggregate ingestion/idempotency. Current code does not include a queue consumer. |

Run order on a fresh database:

```text
001_create_cms_access_tables.up.sql
002_create_product_moderation_reviews.up.sql
003_create_coupon_engine_tables.up.sql
004_create_offer_campaigns.up.sql
005_create_seller_analytics_tables.up.sql
006_create_seller_settings.up.sql
007_harden_cms_audit_logs.up.sql
```

Why order matters: service startup initializes all repositories. Even if you only test analytics, CMS still needs earlier auth/audit tables and later code may expect the complete current schema.

If migrations `001` to `004` are already applied, apply only analytics migration:

```bash
cd backend/services/cms-service
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/005_create_seller_analytics_tables.up.sql
```

Do not blindly rerun this migration in the same DB. `CREATE TABLE IF NOT EXISTS` is safe for tables, but a real project should still use a migration version table/tool so environments stay predictable.

### 5.3 Analytics table fields

`seller_analytics_daily` important fields:

| Column | Meaning |
|---|---|
| `seller_id` | Seller whose dashboard metric this row belongs to. |
| `metric_date` | UTC day for the aggregate. |
| `currency` | 3-letter currency like `INR`. |
| `gross_revenue_amount` | Revenue before refunds, in minor units. |
| `refund_amount` | Refund amount, in minor units. |
| `net_revenue_amount` | Revenue shown by current API summary query. |
| `paid_order_count` | Paid orders for seller. |
| `paid_item_count` | Paid item count for seller. |
| `cancelled_order_count` | Cancelled orders count for seller. |

`seller_product_analytics_daily` important fields:

| Column | Meaning |
|---|---|
| `seller_id`, `product_id` | Seller product identity. |
| `metric_date`, `currency` | Date/currency bucket. |
| `title_snapshot` | Product title copied for dashboard display. |
| `units_sold` | Units sold during that day. |
| `order_count` | Orders containing that product. |
| `revenue_amount` | Product revenue in minor units. |

`seller_conversion_daily` important fields:

| Column | Meaning |
|---|---|
| `product_view_sessions` | Sessions/views used as conversion denominator. |
| `add_to_cart_sessions` | Sessions that reached add-to-cart. |
| `checkout_started_sessions` | Sessions that started checkout. |
| `paid_order_sessions` | Sessions that resulted in paid orders. |

Money note: amounts are stored in minor units. Example: INR 1,250.00 should be stored as `125000`, not `1250.00`.

### 5.4 Required aggregate data

The analytics API reads prepared aggregates. It does not populate them automatically.

For local testing, insert seed rows manually or create a small test fixture. Example:

```sql
INSERT INTO seller_analytics_daily
  (seller_id, metric_date, currency, gross_revenue_amount, refund_amount, net_revenue_amount, paid_order_count, paid_item_count, cancelled_order_count)
VALUES
  ('seller_123', '2026-06-01', 'INR', 150000, 10000, 140000, 7, 12, 1);

INSERT INTO seller_product_analytics_daily
  (seller_id, product_id, metric_date, currency, title_snapshot, units_sold, order_count, revenue_amount)
VALUES
  ('seller_123', 'prod_101', '2026-06-01', 'INR', 'Cotton T-Shirt', 12, 7, 140000);

INSERT INTO seller_conversion_daily
  (seller_id, metric_date, product_view_sessions, add_to_cart_sessions, checkout_started_sessions, paid_order_sessions)
VALUES
  ('seller_123', '2026-06-01', 200, 40, 20, 7);
```

In production, this data should come from an event/job pipeline fed by Order, Payment, and session analytics facts. Current CMS code has only the read path.

### 5.5 Task-specific DB verification

After migration:

```bash
cd backend/services/cms-service
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW TABLES LIKE 'seller_analytics_daily';" cms_db
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW TABLES LIKE 'seller_product_analytics_daily';" cms_db
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW TABLES LIKE 'seller_conversion_daily';" cms_db
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW INDEX FROM seller_analytics_daily;" cms_db
```

Expected:

```text
seller_analytics_daily exists
seller_product_analytics_daily exists
seller_conversion_daily exists
seller_analytics_event_dedupe exists
uk_seller_analytics_day_currency exists
uk_seller_product_day_currency exists
uk_seller_conversion_day exists
```

Quick data check for one seller:

```bash
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SELECT seller_id, metric_date, currency, net_revenue_amount, paid_order_count FROM seller_analytics_daily WHERE seller_id='seller_123';" cms_db
```

## 6. Redis / Queue / External Services

| Service | Required for local API run? | Required for real fresh analytics? | Notes |
|---|---:|---:|---|
| MySQL | Yes | Yes | Analytics API reads all metrics from MySQL read-model tables. |
| API Gateway/Auth | Simulated locally with headers | Yes | Seller identity and roles must be trusted. |
| Order Service | No direct runtime call | Yes as upstream fact source | Paid orders/items should eventually feed analytics aggregates. |
| Payment Service | No direct runtime call | Yes as upstream fact source | Captured/refunded payment facts should feed net revenue. |
| Session/Analytics Service | No direct runtime call | Yes as upstream fact source | Product-view and paid-order sessions should feed conversion table. |
| Redis | No | No current code dependency | No cache client is configured for this task. |
| Kafka/RabbitMQ/NATS | No | Future/optional | `seller_analytics_event_dedupe` hints idempotent event ingestion, but no consumer exists now. |
| Elasticsearch/Typesense | No | No | Top products are queried from MySQL aggregates. |
| SMTP/SMS/OAuth/Stripe/Twilio/S3 | No | No | Not used by current analytics code. |

Simple explanation: dashboard ko fast read chahiye, isliye CMS aggregate tables se read karta hai. Fresh data ke liye future ingestion pipeline chahiye, but local endpoint test ke liye MySQL seed rows enough hain.

## 7. Environment Variables

Full `.env` setup is already covered in previous dependency files. Do not duplicate the complete file.

Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Section:
`5. Environment Variables`

### 7.1 New task-specific variables

| Variable | Required? | Default | Validation | Purpose |
|---|---:|---|---|---|
| `CMS_ANALYTICS_DEFAULT_CURRENCY` | Optional | `INR` | Must be 3 uppercase letters | Currency used when request does not pass `currency`. |
| `CMS_ANALYTICS_DEFAULT_RANGE_DAYS` | Optional | `30` | Must be greater than 0 | Default date window when `from`/`to` are omitted. |
| `CMS_ANALYTICS_MAX_RANGE_DAYS` | Optional | `366` | 1 to 366 | Maximum allowed inclusive date range. |
| `CMS_ANALYTICS_DEFAULT_TOP_PRODUCTS_LIMIT` | Optional | `5` | Must be greater than 0 | Default number of top products. |
| `CMS_ANALYTICS_MAX_TOP_PRODUCTS_LIMIT` | Optional | `20` | 1 to 20 | Hard cap for `top_products_limit`. |

Suggested local additions/checks:

```bash
CMS_ANALYTICS_DEFAULT_CURRENCY=INR
CMS_ANALYTICS_DEFAULT_RANGE_DAYS=30
CMS_ANALYTICS_MAX_RANGE_DAYS=366
CMS_ANALYTICS_DEFAULT_TOP_PRODUCTS_LIMIT=5
CMS_ANALYTICS_MAX_TOP_PRODUCTS_LIMIT=20
```

### 7.2 Reused variables that analytics still needs

| Variable | Why it matters for analytics |
|---|---|
| `CMS_HTTP_ADDR` | HTTP analytics route binds here. Default is `:8087`. |
| `CMS_GRPC_ADDR` | Existing CMS gRPC server binds here. Default is `:9098`, but analytics gRPC is not implemented yet. |
| `CMS_INTERNAL_AUTH_HEADER` | Header name checked by internal auth middleware. Default is `X-Internal-Token`. |
| `CMS_INTERNAL_AUTH_TOKEN` | If set, every analytics HTTP request must include this token. Required in production. |
| `CMS_STAFF_STATUS_SOURCE` | If `gateway`, service trusts `X-Staff-Status`. If `mysql`, staff row must exist in `seller_staff`. |
| `CMS_MYSQL_DSN` or `CMS_DB_*` | Must point to the DB where analytics migration and previous migrations were applied. |
| `CMS_DB_TIMEZONE` | Keep `UTC` so metric dates and DB session timezone are predictable. |

Current code uses `os.Getenv`; it does not auto-load `.env`. Start locally like this:

```bash
cd backend/services/cms-service
set -a
source .env
set +a
go run ./cmd/server
```

Security note: never commit real DB passwords or internal tokens. Local sample values are fine; production secrets should come from a secret manager or deployment environment.

## 8. Docker Setup

No new Dockerfile, compose service, volume, network, or container is introduced by `${TASK_FILE_NAME}`.

Reuse MySQL Docker setup from:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`4. Database Analysis`
`8. Docker and DevOps Setup`

Docker-specific task note: after the MySQL container is running, apply `${ANALYTICS_MIGRATION_PATH}` to the same `cms_db` database used by previous CMS tasks.

Ports:

| Service | Port | Purpose | Status |
|---|---:|---|---|
| CMS HTTP | 8087 | `/api/v1/seller/dashboard/summary` and health route | Reused |
| CMS gRPC | 9098 | Existing CMS internal RPCs | Reused, analytics RPC missing |
| MySQL | 3306 | CMS schema and analytics tables | Reused |
| Redis/Kafka/RabbitMQ | N/A | Not used by current analytics code | Not required |

If port `8087` is busy, change `CMS_HTTP_ADDR`, for example:

```bash
CMS_HTTP_ADDR=:18087
```

If port `3306` is busy, change local MySQL container/host port and update `CMS_DB_PORT` or `CMS_MYSQL_DSN`.

## 9. Local Development Setup

### Step 1: Read previous dependency documentation

Follow these first:

```text
TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
TaskImplementation/${SERVICE_NAME}/task2_Dependency.md
TaskImplementation/${SERVICE_NAME}/task4_Dependency.md
TaskImplementation/${SERVICE_NAME}/task5_Dependency.md
```

They already explain clone, Go setup, MySQL setup, Docker setup, base `.env`, migrations, service startup, internal auth headers, and common setup issues.

### Step 2: Go to project directory

```bash
cd backend/services/cms-service
```

### Step 3: Install only new dependencies if any

No new dependencies for `${TASK_FILE_NAME}`.

Only verify current module state:

```bash
go mod download
go test ./...
```

### Step 4: Setup only new databases/services if any

No new database engine or external service.

Apply the analytics migration if not already applied:

```bash
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/005_create_seller_analytics_tables.up.sql
```

### Step 5: Add only new or changed environment variables

Add/check the `CMS_ANALYTICS_*` values listed in section `7.1`.

### Step 6: Seed analytics rows for local smoke test

If upstream ingestion is not running, insert sample rows from section `5.4`. Without seed data, API can return zeros.

### Step 7: Start backend service

```bash
set -a
source .env
set +a
go run ./cmd/server
```

Expected logs include:

```text
cms.mysql.connected
cms.http.started
cms.grpc.started
```

### Step 8: Verify health

```bash
curl -i http://localhost:8087/healthz
```

Expected: HTTP `200`.

## 10. Running the Project

Task-specific analytics smoke test:

```bash
curl -i "http://localhost:8087/api/v1/seller/dashboard/summary?from=2026-06-01&to=2026-06-01&currency=INR&top_products_limit=5" \
  -H "X-Internal-Token: local-cms-internal-token" \
  -H "X-User-ID: user_123" \
  -H "X-Seller-ID: seller_123" \
  -H "X-Roles: seller" \
  -H "X-Staff-Status: active" \
  -H "X-Request-ID: req_local_analytics_001"
```

If your `.env` uses a different `CMS_INTERNAL_AUTH_HEADER`, change the header name. If `CMS_INTERNAL_AUTH_TOKEN` is empty locally, the middleware will not require the token header, but production should always configure it.

Expected success shape:

```json
{
  "revenue": {
    "amount": 140000,
    "currency": "INR"
  },
  "orders": 7,
  "conversion_rate": 3.5,
  "top_products": [
    {
      "product_id": "prod_101",
      "title": "Cotton T-Shirt",
      "units_sold": 12,
      "orders": 7,
      "revenue": {
        "amount": 140000,
        "currency": "INR"
      }
    }
  ]
}
```

Supported query params:

| Query param | Required? | Example | Notes |
|---|---:|---|---|
| `from` | No | `2026-06-01` | Accepts `YYYY-MM-DD` or RFC3339. Defaults based on `CMS_ANALYTICS_DEFAULT_RANGE_DAYS`. |
| `to` | No | `2026-06-30` | Cannot be in the future. |
| `currency` | No | `INR` | Must normalize to 3-letter uppercase code. |
| `top_products_limit` | No | `5` | Must be positive; capped by `CMS_ANALYTICS_MAX_TOP_PRODUCTS_LIMIT`. |

Permission requirement:

| Requirement | Value |
|---|---|
| Required permission | `cms:analytics:read` |
| Accepted local roles with this permission | `seller`, `seller_manager`, `seller_catalog_editor`, `seller_order_manager` |
| Required seller context | `X-Seller-ID` must match the seller whose data is requested |
| Staff status source | Header-based when `CMS_STAFF_STATUS_SOURCE=gateway`; DB-based when set to `mysql` |

gRPC note: `${TASK_FILE_NAME}` and `api/master-api.json` mention `CMSService.GetSellerAnalytics`, but current `proto/ecommerce/cms/v1/cms.proto` and gRPC server do not expose that RPC. Use HTTP for this task until proto/server support is added.

## 11. Common Errors & Fixes

Generic Go/MySQL/Docker/internal-auth errors are already covered in previous dependency files. This section only lists analytics-specific issues.

| Error/symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `analytics_temporarily_unavailable` | MySQL query failed, analytics tables missing, DB down, or staff lookup failed. | Verify MySQL is running and migration `005` has been applied. Check service logs for the exact query failure. | Run DB verification from section `5.5` before testing API. |
| Response has `0` revenue/orders and empty `top_products` | Tables exist but no aggregate rows match seller/date/currency. | Insert seed rows or run the upstream aggregation job/pipeline. | Always test with known seller/date/currency fixture. |
| `invalid_date_range` | `from`/`to` format wrong, `from > to`, or `to` is future date. | Use `YYYY-MM-DD` or RFC3339 and keep range valid. | Use UTC date strings in smoke tests. |
| `date_range_too_large` | Requested range exceeds `CMS_ANALYTICS_MAX_RANGE_DAYS`. | Reduce date range or adjust env value within allowed max `366`. | Keep dashboard default windows small. |
| `invalid_top_products_limit` | Query value is `0`, negative, or non-numeric. | Use positive integer like `5`. | Keep UI/API clients bounded by env max. |
| `permission_denied` | Role does not include `cms:analytics:read` or staff row/status is inactive. | Use a seller role listed in section `10` and `X-Staff-Status: active`, or seed active `seller_staff` when using MySQL staff source. | Match local auth mode with `CMS_STAFF_STATUS_SOURCE`. |
| `seller_context_required` | Missing `X-Seller-ID`. | Add `X-Seller-ID` header. | Let API Gateway always forward seller context for seller dashboard routes. |
| `INTERNAL_AUTH_REQUIRED` | `CMS_INTERNAL_AUTH_TOKEN` is set but request header missing/wrong. | Send configured internal auth header and token. | Keep local curl snippets aligned with `.env`. |
| gRPC client cannot find `GetSellerAnalytics` | Current proto/server does not implement analytics RPC. | Use HTTP endpoint for now. | Add proto/server implementation before documenting gRPC smoke tests. |

## 12. Security & Best Practices

Task-specific best practices:

| Practice | Why |
|---|---|
| Keep analytics route behind internal auth/API Gateway | Seller metrics are sensitive business data. |
| Never accept `seller_id` as an analytics query param for seller dashboard | Current code correctly uses seller id from trusted actor context. |
| Store money in minor units | Avoid float rounding issues in revenue. |
| Keep `CMS_DB_TIMEZONE=UTC` | Date buckets should be predictable across services. |
| Hash seller ids in logs | Current usecase hashes seller id before logging, which is safer than raw ids. |
| Limit date ranges and top-product count | Prevent expensive dashboard queries. |
| Use aggregate/read-model tables | Avoid direct cross-service DB reads from Order/Payment/Session databases. |
| Seed/test with fixed date ranges | Dashboard smoke tests become repeatable. |
| Add indexes before increasing dashboard traffic | Existing migration adds seller/date/currency indexes; keep future filters indexed. |

## 13. Missing or Misconfigured Things

| Finding | Impact | Suggested fix |
|---|---|---|
| No ingestion worker/consumer exists for analytics aggregates | API can only read data already present in MySQL; fresh order/payment/session facts will not automatically update metrics. | Add event consumer or scheduled aggregator fed by Order, Payment, and session analytics facts. |
| `seller_analytics_event_dedupe` exists but no producer/consumer uses it | Dedupe design is ready, but runtime idempotency pipeline is incomplete. | Wire it when Kafka/RabbitMQ/NATS/job ingestion is implemented. |
| `api/master-api.json` mentions `CMSService.GetSellerAnalytics`, but proto/server do not implement it | gRPC clients cannot call analytics even though master API docs suggest it. | Add proto messages/RPC, regenerate Go code, inject analytics usecase into gRPC server, and add tests. |
| No Docker Compose file found for CMS local stack | Beginners must rely on previous docs/manual Docker commands. | Add a root or service-level compose file for MySQL and CMS when DevOps setup is formalized. |
| No migration runner/version table in service folder | Manual SQL apply can drift across environments. | Add a migration tool workflow such as goose/migrate or a platform migration runner. |
| No cache layer for dashboard summary | High-traffic dashboards may repeatedly hit MySQL aggregates. | Add Redis/cache later if performance requires it; not required for current task. |
| No freshness timestamp in API response | UI cannot show "last updated" status. | Add aggregate freshness fields or metadata in future API contract. |

## 14. References to Previous Dependency Files

Previous dependency files inside `TaskImplementation/${SERVICE_NAME}/` were checked first.

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `2. Tech Stack Analysis` | Same Go, `net/http`, gRPC, protobuf, MySQL, driver, auth-header foundation. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `3. Go Dependency System` | No new Go dependency was added for analytics. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `4. Database Analysis` | MySQL install, Docker, DB user, and connection string setup are unchanged. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `5. Environment Variables` | Full `.env` explanation already exists; this file only adds `CMS_ANALYTICS_*`. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `7. Ports and Networking` | CMS HTTP, gRPC, and MySQL ports are reused. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `8. Docker and DevOps Setup` | No new Docker service was introduced. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `11. Common Errors and Fixes` | Generic Go/MySQL/Docker/internal auth errors are unchanged. |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | Same `cms_db`, MySQL 8+, InnoDB, UTF-8, and migration-order foundation. |
| `TaskImplementation/${SERVICE_NAME}/task3_Dependency.md` | `7. External Services Analysis` | Product Service/auth gateway notes are reused; analytics does not add Product Service calls. |
| `TaskImplementation/${SERVICE_NAME}/task4_Dependency.md` | `6. Redis / Queue / External Services` | Confirms Redis/queue services are not currently required by CMS code. |
| `TaskImplementation/${SERVICE_NAME}/task5_Dependency.md` | `8. Docker, Ports, and DevOps Setup` | Same CMS HTTP/gRPC/MySQL runtime shape; no new container or port. |

## 15. Final Checklist

Use this checklist before saying `${TASK_FILE_NAME}` setup is ready:

* [ ] Previous dependency documentation checked first.
* [ ] No duplicate Go/MySQL/Docker baseline setup copied into this file.
* [ ] MySQL is running.
* [ ] Base CMS migrations are applied in numeric order.
* [ ] `${ANALYTICS_MIGRATION_PATH}` has been applied.
* [ ] Analytics tables exist: `seller_analytics_daily`, `seller_product_analytics_daily`, `seller_conversion_daily`, `seller_analytics_event_dedupe`.
* [ ] `.env` includes/accepts `CMS_ANALYTICS_DEFAULT_CURRENCY`.
* [ ] `.env` includes/accepts `CMS_ANALYTICS_DEFAULT_RANGE_DAYS`.
* [ ] `.env` includes/accepts `CMS_ANALYTICS_MAX_RANGE_DAYS`.
* [ ] `.env` includes/accepts `CMS_ANALYTICS_DEFAULT_TOP_PRODUCTS_LIMIT`.
* [ ] `.env` includes/accepts `CMS_ANALYTICS_MAX_TOP_PRODUCTS_LIMIT`.
* [ ] `CMS_INTERNAL_AUTH_TOKEN` and curl header match, if token is configured.
* [ ] Local seed data exists for the seller/date/currency being tested.
* [ ] Service starts and logs `cms.mysql.connected`, `cms.http.started`, and `cms.grpc.started`.
* [ ] `GET /healthz` returns HTTP `200`.
* [ ] `GET /api/v1/seller/dashboard/summary` returns expected revenue/orders/conversion/top products.
* [ ] Missing gRPC analytics implementation is understood and not tested as if it already exists.
* [ ] Logs checked for analytics request success/failure.
* [ ] No real secrets were added to markdown or committed files.

