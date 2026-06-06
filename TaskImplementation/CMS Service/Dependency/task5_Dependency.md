# Dependency + Setup Documentation

## Document Variables

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `CMS Service` |
| `TASK_FILE_NAME` | `task5.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task5_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}` |
| `BACKEND_SERVICE_PATH` | `backend/services/cms-service` |
| `CAMPAIGN_MIGRATION_PATH` | `${BACKEND_SERVICE_PATH}/migrations/004_create_offer_campaigns.up.sql` |

This file documents dependency, setup, environment, database, and DevOps requirements for `${INPUT_FILE_PATH}`.

Simple goal: beginner developer ko clear ho ki offer campaign flow run/test karne ke liye kya setup chahiye, kya previous dependency docs me already covered hai, aur `${TASK_FILE_NAME}` ke liye kaunsi campaign-specific cheezein verify karni hain.

Important: original implementation task file is not modified. This is a new setup/dependency document saved at `${OUTPUT_FILE_PATH}`.

## 1. Previous Dependency File Reuse

Previous dependency files inside `TaskImplementation/${SERVICE_NAME}/` were checked first.

| Previous file | Already explains | Reuse decision |
|---|---|---|
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | Go setup, Go modules, MySQL install, MySQL Docker, full `.env`, service run commands, ports, shared HTTP/gRPC notes, common setup errors | Follow this first. Do not repeat full setup here. |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | MySQL 8+ decision, `cms_db`, InnoDB, UTF-8, migration order, DB verification, schema ownership | Reuse DB foundation. This file only expands campaign table checks. |
| `TaskImplementation/${SERVICE_NAME}/task3_Dependency.md` | Product moderation setup, Product Service dependency, auth headers, product-specific troubleshooting | Product Service setup is not new for campaigns. Refer only if running moderation flows too. |
| `TaskImplementation/${SERVICE_NAME}/task4_Dependency.md` | Coupon engine setup, coupon tables, coupon validation, coupon redemption, gRPC coupon smoke test | Required background because campaigns depend on coupon/redemption records. |

Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`2. Tech Stack Analysis`
`3. Go Dependency System`
`4. Database Analysis`
`5. Environment Variables`
`6. External Services Analysis`
`7. Ports and Networking`
`8. Docker and DevOps Setup`
`9. Complete Project Run Instructions`
`10. HTTP and gRPC Endpoint Setup Notes`
`11. Common Errors and Fixes`
`15. Minimal Local Command Flow`

Refer:
`TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`

Sections:
`5. Database Setup`
`7. Environment Variables`
`9. Local Development Setup`
`11. Common Errors & Fixes`

Refer:
`TaskImplementation/${SERVICE_NAME}/task4_Dependency.md`

Sections:
`5. Database Setup`
`6. Redis / Queue / External Services`
`7. Environment Variables`
`9. Local Development Setup`
`10. Running the Project`
`11. Common Errors & Fixes`

## 2. What `${TASK_FILE_NAME}` Adds

`${TASK_FILE_NAME}` covers offer campaigns: start/end time, budget cap, usage limits, seller ownership, campaign lifecycle, and coupon engine linkage.

Current backend implementation under `${BACKEND_SERVICE_PATH}` contains campaign support already.

| Area | File/path | Setup impact |
|---|---|---|
| Campaign domain | `${BACKEND_SERVICE_PATH}/internal/domain/campaign.go` | No external install. Defines status, metadata, ownership, time-window, budget, usage, and coupon-link validation. |
| Campaign usecase | `${BACKEND_SERVICE_PATH}/internal/usecase/campaign.go` | Needs Authorizer, campaign repository, coupon reader, audit recorder, and `CMS_CAMPAIGN_MAX_DURATION_DAYS`. |
| Campaign repository | `${BACKEND_SERVICE_PATH}/internal/repository/mysql_campaign_repository.go` | Requires `campaigns` table and `coupon_redemptions.campaign_id`. |
| Campaign migration | `${CAMPAIGN_MIGRATION_PATH}` | Creates `campaigns`; alters `coupon_redemptions` to add campaign linkage and indexes. |
| HTTP routes | `${BACKEND_SERVICE_PATH}/internal/transport/http/handler.go` | Adds internal campaign list/create/update/disable/validate endpoints. |
| gRPC reads | `${BACKEND_SERVICE_PATH}/internal/transport/grpc/campaign_handler.go` | Adds `GetCampaign` and `ListCampaigns` RPC handlers. |
| Proto contract | `proto/ecommerce/cms/v1/cms.proto` | Defines campaign read messages. Current proto does not expose campaign mutation RPCs. |
| Coupon integration | `${BACKEND_SERVICE_PATH}/internal/usecase/coupon.go` and `${BACKEND_SERVICE_PATH}/internal/repository/mysql_coupon_repository.go` | Coupon validation/redemption can enforce campaign budget, usage, seller, currency, and linked coupon checks. |
| Audit logs | `${BACKEND_SERVICE_PATH}/internal/domain/audit.go` and MySQL audit repository | Campaign create/update/disable/resume/complete writes audit records. |

No new Go package, database engine, Docker container, Redis cache, Kafka/RabbitMQ queue, object storage, payment provider, email/SMS provider, or third-party credential is introduced by `${TASK_FILE_NAME}`.

## 3. Tech Stack Analysis

Most technology setup is already documented in previous files. This section only explains what matters for campaigns.

| Technology/service | Status | Why it matters for `${TASK_FILE_NAME}` | Beginner explanation |
|---|---|---|---|
| Go | Reused | Campaign domain, usecase, repository, HTTP, and gRPC code are implemented in Go. | Go backend language hai. Fast APIs and service logic ke liye use hota hai. |
| Go modules | Reused | `go.mod` has no new direct dependency for campaigns. | `go.mod` dependency list hai; `go.sum` checksum safety file hai. |
| `net/http` | Reused with campaign routes | Seller campaign management and campaign validation run on the existing HTTP server. | Ye Go ka built-in HTTP server hai. |
| gRPC | Reused with campaign reads | Internal services can call `GetCampaign` and `ListCampaigns`; `ValidateCoupon` also accepts `campaign_id`. | gRPC typed service-to-service communication hota hai. |
| Protocol Buffers | Reused | Campaign read request/response messages live in `proto/ecommerce/cms/v1/cms.proto`. | Proto contract hai jisse generated Go types bante hain. |
| MySQL 8+ | Reused with campaign migration | Campaigns, campaign metadata, redemption linkage, and audit logs are stored in MySQL. | MySQL table-based database hai. Structured records aur indexes ke liye useful hai. |
| `github.com/go-sql-driver/mysql` | Reused | Go `database/sql` ko MySQL se connect karata hai. | Driver bridge ki tarah kaam karta hai. |
| Coupon engine | Required dependency | Campaign budget/usage checks depend on coupon redemptions and linked coupons. | Coupon discount calculate karta hai; campaign us discount ko time, budget, and seller control deta hai. |
| Auth/API Gateway headers | Required for seller routes | Seller id, user id, roles, staff status, and request id come from trusted headers. | Gateway JWT validate karke service ko identity headers deta hai. |
| Docker | Optional, reused | Beginner local MySQL setup ke liye easiest path. | Docker se MySQL container run kar sakte ho without manual install. |
| Redis/Kafka/RabbitMQ | Not used | No campaign runtime code connects to these services. | Is task ke liye in containers ko start karna zaruri nahi hai. |

## 4. Required Software and Dependency Management

Do not reinstall tools if you already completed previous dependency guides.

| Software/tool | Required? | New for `${TASK_FILE_NAME}`? | Reference |
|---|---:|---:|---|
| Go compatible with `${BACKEND_SERVICE_PATH}/go.mod` | Yes | No | `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`, section `3. Go Dependency System` |
| MySQL 8+ server | Yes | No, but campaign table is task-specific | `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`, section `4. Database Analysis` |
| MySQL CLI client | Strongly recommended | No | Same previous section |
| Docker | Optional | No | `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`, section `8. Docker and DevOps Setup` |
| `grpcurl` | Optional | No | Useful only for manual gRPC read/coupon validation tests |
| Buf/protobuf plugins | Optional | No | Required only if proto files change |

No new Go dependency is introduced by `${TASK_FILE_NAME}`.

Current direct module dependencies remain:

| Dependency | Used for |
|---|---|
| `github.com/go-sql-driver/mysql` | MySQL connection and SQL repositories. |
| `google.golang.org/grpc` | Existing internal gRPC server. |
| `google.golang.org/protobuf` | Generated protobuf messages. |

Reuse dependency commands:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`3. Go Dependency System`
`9. Complete Project Run Instructions`
`15. Minimal Local Command Flow`
```

Optional verification from service folder:

```bash
cd backend/services/cms-service
go mod download
go test ./...
```

Beginner note: campaign task ke liye `go get` run karne ki zarurat nahi hai. Required packages already `go.mod` me present hain.

## 5. Database Setup

### 5.1 Database used

| Database | Required? | Default port | Status |
|---|---:|---:|---|
| MySQL 8+ | Yes | 3306 | Reused setup with task-specific `campaigns` table and redemption linkage. |

MySQL install, Docker command, DB user creation, and full migration flow are already documented.

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

Campaign support is created by:

```text
${BACKEND_SERVICE_PATH}/migrations/004_create_offer_campaigns.up.sql
```

This migration does two things:

| Change | Why it matters |
|---|---|
| Creates `campaigns` | Stores campaign name, seller owner, status, budget, currency, time window, metadata, creator, and timestamps. |
| Alters `coupon_redemptions` | Adds nullable `campaign_id`, campaign indexes, and a foreign key to `campaigns(campaign_id)`. |

Important migration order:

```text
001_create_cms_access_tables.up.sql
002_create_product_moderation_reviews.up.sql
003_create_coupon_engine_tables.up.sql
004_create_offer_campaigns.up.sql
005_create_seller_analytics_tables.up.sql
006_create_seller_settings.up.sql
007_harden_cms_audit_logs.up.sql
```

Why order matters: migration `004` alters `coupon_redemptions`, so migration `003` must run before it. Service startup initializes all repositories, so on a fresh local DB run the full migration set from the previous setup guide.

Do not blindly rerun migration `004` after it has already applied. The `ALTER TABLE coupon_redemptions ADD COLUMN campaign_id` part can fail with duplicate column/key errors. Use a migration tool/version table for repeated environments.

### 5.3 Campaign table fields

| Column | Purpose |
|---|---|
| `campaign_id` | Stable public/internal id returned by APIs. |
| `seller_id` | Seller owner. Seller campaign routes require this to match authenticated seller context. |
| `name` | Seller-facing campaign name, max 255 characters. |
| `status` | Lifecycle state: `draft`, `active`, `paused`, `completed`. |
| `budget_amount` | Optional total campaign discount budget in minor currency units. |
| `currency` | 3-letter uppercase currency, default `INR`. |
| `starts_at` | UTC start time. |
| `ends_at` | UTC end time. |
| `metadata` | JSON config for coupon ids, channels, description, usage limits, per-user limits, and budget alert threshold. |
| `created_by` | Actor user id that created the campaign. |
| `created_at`, `updated_at` | Audit/debug timestamps. |

Campaign metadata currently supports:

| Metadata key | Required? | Meaning |
|---|---:|---|
| `coupon_ids` | Optional | Coupon ids allowed under this campaign. If empty, current code allows any coupon for standalone campaign eligibility. |
| `channels` | Optional | Logical channels like seller dashboard, homepage slot, app banner. |
| `description` | Optional | Human-readable campaign note. |
| `usage_limit` | Optional | Total successful campaign redemptions allowed. |
| `per_user_limit` | Optional | Successful campaign redemptions allowed per user. |
| `budget_alert_threshold_percent` | Optional | Alert threshold between 1 and 100. Current code validates it but does not send alerts. |

Money note: store money in minor units. Example: INR 500.00 is `50000`, not `500.00`.

### 5.4 Important keys and constraints

| Table | Key/constraint | Why it matters |
|---|---|---|
| `campaigns` | `uk_campaigns_campaign_id` | Prevents duplicate campaign ids. |
| `campaigns` | `idx_campaigns_seller_window` | Speeds seller/time-window queries. |
| `campaigns` | `idx_campaigns_seller_status` | Speeds seller dashboard status filters. |
| `campaigns` | `idx_campaigns_status_window` | Speeds active-window scans. |
| `campaigns` | `chk_campaigns_budget` | Budget must be null or positive. |
| `campaigns` | `chk_campaigns_currency` | Currency must look like `INR`, `USD`, etc. |
| `campaigns` | `chk_campaigns_window` | `starts_at` must be before `ends_at`. |
| `campaigns` | `chk_campaigns_metadata_json` | Metadata must be valid JSON. |
| `coupon_redemptions` | `idx_coupon_redemptions_campaign_created` | Speeds campaign redemption count/spend lookups. |
| `coupon_redemptions` | `idx_coupon_redemptions_campaign_user` | Speeds per-user campaign limit checks. |
| `coupon_redemptions` | `fk_coupon_redemptions_campaign` | Blocks redemption rows pointing to unknown campaigns. |

### 5.5 Task-specific DB verification

After running migrations:

```bash
cd backend/services/cms-service
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW TABLES LIKE 'campaigns';" cms_db
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW COLUMNS FROM campaigns;" cms_db
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW INDEX FROM campaigns;" cms_db
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW COLUMNS FROM coupon_redemptions LIKE 'campaign_id';" cms_db
```

Expected:

```text
campaigns table exists
campaign_id, seller_id, status, budget_amount, currency, starts_at, ends_at, metadata columns exist
coupon_redemptions.campaign_id exists
uk_campaigns_campaign_id exists
idx_campaigns_seller_window exists
idx_coupon_redemptions_campaign_created exists
idx_coupon_redemptions_campaign_user exists
```

Deep debug command:

```bash
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW CREATE TABLE campaigns\\G" cms_db
```

### 5.6 Campaign spend and usage source

Campaign spend and usage are calculated from `coupon_redemptions`, not from a separate campaign counter table.

| Runtime check | Query idea |
|---|---|
| Total campaign usage | `COUNT(*) FROM coupon_redemptions WHERE campaign_id = ?` |
| Per-user campaign usage | `COUNT(*) FROM coupon_redemptions WHERE campaign_id = ? AND user_id = ?` |
| Campaign spend | `SUM(discount_amount) FROM coupon_redemptions WHERE campaign_id = ?` |

Important beginner explanation: campaign budget is not consumed when cart preview happens. Budget is enforced when final redemption is recorded after order/payment success.

## 6. Environment Variables

Most environment variables are already covered in previous dependency files. Do not duplicate the full `.env`.

Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Section:
`5. Environment Variables`

### 6.1 Task-specific or central variables

| Variable | Required? | Default | Why it matters for `${TASK_FILE_NAME}` |
|---|---:|---|---|
| `CMS_CAMPAIGN_MAX_DURATION_DAYS` | Optional | `90` | Maximum allowed campaign duration. Negative value fails config validation. Keep it positive for predictable seller campaigns. |
| `CMS_DB_TIMEZONE` | Recommended | `UTC` | Campaign windows must compare predictably in UTC. |
| `CMS_HTTP_ADDR` | Required by config default | `:8087` | HTTP campaign routes bind here. |
| `CMS_GRPC_ADDR` | Required by config default | `:9098` | gRPC campaign read/coupon validation methods bind here. |
| `CMS_GRPC_ALLOWED_INTERNAL_CALLERS` | Optional | `cart-service,order-service,api-gateway` | gRPC internal allowlist. Campaign reads allow `api-gateway`, `cart-service`, and `order-service` when using service metadata. |
| `CMS_INTERNAL_AUTH_HEADER` | Required | `X-Internal-Token` | Header name checked by internal HTTP and gRPC calls when token is configured. |
| `CMS_INTERNAL_AUTH_TOKEN` | Required in production, recommended locally | empty locally unless set | Protects internal campaign endpoints. |
| `CMS_STAFF_STATUS_SOURCE` | Required by config default | `gateway` | If set to `mysql`, staff records must be seeded in `seller_staff`. |
| `CMS_MYSQL_DSN` or `CMS_DB_*` | Required for startup | See previous docs | Must point to DB where migration `004` has been applied. |

No campaign-specific secret is introduced by `${TASK_FILE_NAME}`. Do not create a separate campaign API key, payment credential, Redis credential, Kafka credential, or third-party integration credential for this task.

### 6.2 Suggested local `.env` additions/checks

If your existing `.env` already has these, only verify the values:

```bash
CMS_CAMPAIGN_MAX_DURATION_DAYS=90
CMS_DB_TIMEZONE=UTC
CMS_INTERNAL_AUTH_TOKEN=local-cms-internal-token
CMS_GRPC_ALLOWED_INTERNAL_CALLERS=cart-service,order-service,api-gateway
```

Important: current code uses `os.Getenv`. It does not auto-load `.env`. Source the file before running:

```bash
cd backend/services/cms-service
set -a
source .env
set +a
go run ./cmd/server
```

## 7. Redis, Queues, and External Services

| Service | Required for campaign management? | Required for full platform flow? | Notes |
|---|---:|---:|---|
| MySQL | Yes | Yes | Source of truth for campaigns, coupons, redemptions, and audit logs. |
| Coupon engine tables/usecase | Yes | Yes | Campaigns depend on coupon ownership/currency checks and redemption history. |
| API Gateway/Auth | Simulated locally with headers | Yes | Seller identity and role context must be trusted. |
| Cart Service | No | Yes | Full checkout preview can call coupon validation with `campaign_id`. |
| Order Service | No | Yes | Full order flow records coupon redemption with `campaign_id` after payment success. |
| Product Service | No for campaigns | Only for product moderation flows | Campaign create/list/validate does not call Product Service. |
| Redis | No | No current code dependency | Future cache can be added, but MySQL remains source of truth. |
| Kafka/RabbitMQ/NATS | No | No current code dependency | No campaign queue consumer/producer exists right now. |
| Elasticsearch/Typesense | No | No current code dependency | Campaign search here is MySQL `LIKE` by name. |
| SMTP/SMS/OAuth/Payment provider | No | No current campaign dependency | Budget is discount governance, not payment settlement. |

Simple explanation: campaign task ka real external dependency coupon engine hai, because campaign budget and usage are enforced through coupon redemption records. Redis ya queue start karne ki zarurat nahi hai.

## 8. Docker, Ports, and DevOps Setup

No new Docker container, Dockerfile, volume, network, or compose service is introduced by `${TASK_FILE_NAME}`.

Reuse:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`4. Database Analysis` -> MySQL Docker setup
`7. Ports and Networking`
`8. Docker and DevOps Setup`
```

Task-specific reminders:

| Item | Value / behavior |
|---|---|
| HTTP default port | `8087` from `CMS_HTTP_ADDR=:8087` |
| gRPC default port | `9098` from `CMS_GRPC_ADDR=:9098` |
| MySQL default port | `3306` |
| Local host DB name | `cms_db` |
| Service binds to all interfaces when address is `:8087` or `:9098` | Use `127.0.0.1:8087` and `127.0.0.1:9098` for stricter local-only binding. |
| Migration runner | Not present in repo | Manual SQL commands are documented; production should use a versioned migration tool. |

## 9. Local Development Flow

### Step 1: Follow shared setup first

Follow these before campaign-specific testing:

```text
TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
TaskImplementation/${SERVICE_NAME}/task2_Dependency.md
TaskImplementation/${SERVICE_NAME}/task4_Dependency.md
```

### Step 2: Go to service folder

```bash
cd backend/services/cms-service
```

### Step 3: Verify migrations

If this is a fresh database, run the full migration flow from `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`.

At minimum for campaign/coupon flow, these must exist:

```text
seller_staff
cms_audit_logs
coupons
coupon_rules
coupon_redemptions
campaigns
```

Verification:

```bash
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW TABLES LIKE 'campaigns'; SHOW COLUMNS FROM coupon_redemptions LIKE 'campaign_id';" cms_db
```

### Step 4: Start the service

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

### Step 5: Health check

```bash
curl http://localhost:8087/healthz
```

Expected:

```json
{"success":true}
```

### Step 6: Create a campaign

This smoke test creates a campaign without linked `coupon_ids`. That keeps the first test simple. For strict coupon linkage, create a coupon first using `TaskImplementation/${SERVICE_NAME}/task4_Dependency.md` and then include the returned `coupon_id` in `metadata.coupon_ids`.

Linux/WSL bash example:

```bash
STARTS_AT="$(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%SZ)"
ENDS_AT="$(date -u -d '7 days' +%Y-%m-%dT%H:%M:%SZ)"

curl -X POST http://localhost:8087/internal/v1/cms/seller/campaigns \
  -H "Content-Type: application/json" \
  -H "X-Internal-Token: local-cms-internal-token" \
  -H "X-User-ID: user_seller_1" \
  -H "X-Seller-ID: seller_1" \
  -H "X-Roles: seller_catalog_editor" \
  -H "X-Staff-Status: active" \
  -H "X-Request-ID: req_campaign_create_1" \
  -d "{
    \"name\": \"Local Campaign Smoke Test\",
    \"currency\": \"INR\",
    \"budget\": {\"amount\": 100000, \"currency\": \"INR\"},
    \"starts_at\": \"${STARTS_AT}\",
    \"ends_at\": \"${ENDS_AT}\",
    \"metadata\": {
      \"channels\": [\"seller_dashboard\"],
      \"description\": \"Local smoke test campaign\",
      \"usage_limit\": 100,
      \"per_user_limit\": 2,
      \"budget_alert_threshold_percent\": 80
    }
  }"
```

If your local `.env` does not set `CMS_INTERNAL_AUTH_TOKEN`, remove the `X-Internal-Token` header from local curl commands. Production should always set the token.

Expected response contains:

```json
{
  "campaign_id": "camp_...",
  "seller_id": "seller_1",
  "status": "draft"
}
```

Save the returned id:

```bash
CAMPAIGN_ID=camp_replace_me
```

### Step 7: Activate campaign

Create always starts as `draft`. Campaign validation requires `active`.

```bash
curl -X PATCH "http://localhost:8087/internal/v1/cms/seller/campaigns/${CAMPAIGN_ID}" \
  -H "Content-Type: application/json" \
  -H "X-Internal-Token: local-cms-internal-token" \
  -H "X-User-ID: user_seller_1" \
  -H "X-Seller-ID: seller_1" \
  -H "X-Roles: seller_catalog_editor" \
  -H "X-Staff-Status: active" \
  -H "X-Request-ID: req_campaign_activate_1" \
  -d '{"status":"active"}'
```

### Step 8: List campaigns

```bash
curl "http://localhost:8087/internal/v1/cms/seller/campaigns?status=active&limit=10&offset=0" \
  -H "X-Internal-Token: local-cms-internal-token" \
  -H "X-User-ID: user_seller_1" \
  -H "X-Seller-ID: seller_1" \
  -H "X-Roles: seller_catalog_editor" \
  -H "X-Staff-Status: active"
```

Expected: only campaigns owned by `seller_1` appear.

### Step 9: Validate campaign directly

```bash
curl -X POST http://localhost:8087/internal/v1/cms/campaigns/validate \
  -H "Content-Type: application/json" \
  -H "X-Internal-Token: local-cms-internal-token" \
  -H "X-Request-ID: req_campaign_validate_1" \
  -d "{
    \"campaign_id\": \"${CAMPAIGN_ID}\",
    \"seller_id\": \"seller_1\",
    \"currency\": \"INR\",
    \"discount_amount\": 25000
  }"
```

Expected:

```json
{
  "valid": true,
  "campaign_id": "camp_...",
  "remaining_budget": {"amount": 100000, "currency": "INR"}
}
```

Note: direct campaign validation only checks coupon linkage strictly when `coupon_id` is sent. In the full coupon flow, the coupon id is known and checked automatically.

### Step 10: Full coupon + campaign smoke test

First create an active coupon using `TaskImplementation/${SERVICE_NAME}/task4_Dependency.md`. Then either create a campaign with that coupon id in `metadata.coupon_ids`, or patch metadata to include it.

Validate coupon with campaign:

```bash
curl -X POST http://localhost:8087/internal/v1/cms/coupons/validate \
  -H "Content-Type: application/json" \
  -H "X-Internal-Token: local-cms-internal-token" \
  -H "X-Request-ID: req_coupon_campaign_validate_1" \
  -d "{
    \"coupon_code\": \"SAVE250\",
    \"campaign_id\": \"${CAMPAIGN_ID}\",
    \"user_id\": \"buyer_1\",
    \"cart_id\": \"cart_1\",
    \"currency\": \"INR\",
    \"subtotal_amount\": 149900,
    \"items\": [
      {
        \"product_id\": \"prod_1\",
        \"seller_id\": \"seller_1\",
        \"category_ids\": [\"cat_shoes\"],
        \"quantity\": 1,
        \"unit_amount\": 149900,
        \"line_subtotal_amount\": 149900
      }
    ]
  }"
```

Record redemption after successful order:

```bash
curl -X POST http://localhost:8087/internal/v1/cms/coupons/redemptions \
  -H "Content-Type: application/json" \
  -H "X-Internal-Token: local-cms-internal-token" \
  -H "X-Request-ID: req_coupon_campaign_redeem_1" \
  -d "{
    \"coupon_id\": \"coupon_replace_me\",
    \"campaign_id\": \"${CAMPAIGN_ID}\",
    \"order_id\": \"order_campaign_1\",
    \"user_id\": \"buyer_1\",
    \"discount\": {\"amount\": 25000, \"currency\": \"INR\"}
  }"
```

This final redemption path locks the coupon and campaign rows inside a transaction before inserting `coupon_redemptions`.

## 10. HTTP and gRPC Endpoint Setup Notes

### 10.1 HTTP routes

Current implementation exposes these campaign HTTP endpoints:

| Route | Auth | Purpose |
|---|---|---|
| `GET /internal/v1/cms/seller/campaigns` | Internal token + seller actor headers | List seller campaigns. Supports `status`, `starts_after`, `ends_before`, `q`, `limit`, `offset`. |
| `POST /internal/v1/cms/seller/campaigns` | Internal token + seller actor headers | Create seller campaign as `draft`. |
| `PATCH /internal/v1/cms/seller/campaigns/{campaign_id}` | Internal token + seller actor headers | Update campaign fields or status. |
| `POST /internal/v1/cms/seller/campaigns/{campaign_id}/disable` | Internal token + seller actor headers | Pause campaign. |
| `POST /internal/v1/cms/campaigns/validate` | Internal token | Validate campaign eligibility for internal callers. |

Required local seller headers for management endpoints:

| Header | Example | Why |
|---|---|---|
| `X-User-ID` | `user_seller_1` | Actor user for auth/audit. |
| `X-Seller-ID` | `seller_1` | Seller ownership scope. |
| `X-Roles` | `seller_catalog_editor` | Role with campaign permissions. |
| `X-Staff-Status` | `active` | Required active staff context. |
| `X-Request-ID` | `req_campaign_1` | Trace/audit id. |
| `X-Internal-Token` | `local-cms-internal-token` | Required only if `CMS_INTERNAL_AUTH_TOKEN` is configured. |

Roles that include campaign permissions in current code: `seller`, `seller_manager`, and `seller_catalog_editor`.

### 10.2 gRPC methods

Current proto has campaign read methods only:

| Method | Current purpose | Notes |
|---|---|---|
| `CMSService.GetCampaign` | Read one campaign and its spent amount. | Authenticated actor metadata works. If no actor metadata, internal callers `api-gateway`, `cart-service`, or `order-service` can read with `x-service-name` and internal token. |
| `CMSService.ListCampaigns` | List seller campaigns. | Requires actor metadata with seller context. |
| `CMSService.ValidateCoupon` | Validate coupon and optionally campaign. | Send `campaign_id` in `ValidateCouponRequest`. |

Current proto does not expose `CreateCampaign`, `UpdateCampaign`, `DisableCampaign`, or standalone `ValidateCampaign` RPCs. Use HTTP endpoints for those flows unless proto/API is intentionally extended later.

Example campaign read with `grpcurl`:

```bash
grpcurl -plaintext \
  -H "x-user-id: user_seller_1" \
  -H "x-seller-id: seller_1" \
  -H "x-roles: seller_catalog_editor" \
  -H "x-staff-status: active" \
  -H "x-request-id: req_grpc_campaign_get_1" \
  -d "{\"campaign_id\":\"${CAMPAIGN_ID}\",\"seller_id\":\"seller_1\"}" \
  localhost:9098 ecommerce.cms.v1.CMSService/GetCampaign
```

If using internal service metadata instead of actor metadata for `GetCampaign`:

```bash
grpcurl -plaintext \
  -H "x-service-name: cart-service" \
  -H "x-internal-token: local-cms-internal-token" \
  -H "x-request-id: req_grpc_campaign_get_internal_1" \
  -d "{\"campaign_id\":\"${CAMPAIGN_ID}\",\"seller_id\":\"seller_1\"}" \
  localhost:9098 ecommerce.cms.v1.CMSService/GetCampaign
```

## 11. Campaign Runtime Rules

### 11.1 Ownership

Seller campaign rule:

```text
actor.seller_id must equal campaign.seller_id
```

Campaigns with `seller_id = NULL` are treated as platform/future campaigns. Current seller management APIs require seller-owned campaigns.

### 11.2 Lifecycle

| Status | Meaning | Allowed next states in current domain logic |
|---|---|---|
| `draft` | Created but not live | `active`, `paused`, `completed` |
| `active` | Eligible only when current UTC time is inside window | `paused`, `completed` |
| `paused` | Temporarily disabled | `active`, `completed` |
| `completed` | Finished/read-only | No further changes |

### 11.3 Create/update validation

| Field | Rule |
|---|---|
| `name` | Required, trimmed, max 255 characters. |
| `starts_at` | Required timestamp, normalized to UTC. |
| `ends_at` | Required timestamp, must be after `starts_at`. |
| Duration | Must not exceed `CMS_CAMPAIGN_MAX_DURATION_DAYS`; default is 90 days. |
| `budget.amount` | Optional, but if present must be positive. |
| `currency` | Must be a 3-letter uppercase code after normalization. |
| `metadata.usage_limit` | Optional positive integer. |
| `metadata.per_user_limit` | Optional positive integer. |
| `metadata.budget_alert_threshold_percent` | Optional integer between 1 and 100. |
| `metadata.coupon_ids` | If present, every coupon must exist, belong to the same seller, and use the same currency. |

### 11.4 Active campaign edit restrictions

When an existing campaign is `active`, current usecase blocks risky reductions:

| Edit | Current behavior |
|---|---|
| Move `starts_at` into future | Blocked. |
| Shorten `ends_at` | Blocked. Only extension is safe. |
| Reduce budget | Blocked. Budget can only increase. |
| Reduce `usage_limit` or `per_user_limit` | Blocked. |
| Remove existing linked coupon ids | Blocked. |
| Completed campaign edit | Blocked as read-only. |

### 11.5 Eligibility validation order

Campaign validation checks:

```text
1. Campaign exists
2. Seller ownership matches
3. Status is active
4. Current UTC time is inside starts_at/ends_at
5. Currency matches
6. Coupon id is allowed by campaign metadata
7. Usage limit is available
8. Per-user limit is available
9. Budget is available
```

Reason codes returned by campaign validation:

| Reason | Meaning |
|---|---|
| `campaign_not_found` | Campaign id does not exist. |
| `seller_scope_mismatch` | Campaign seller and request seller do not match. |
| `campaign_not_active` | Campaign status is not `active`. |
| `campaign_not_started` | Start time is in the future. |
| `campaign_expired` | End time is in the past. |
| `campaign_usage_limit_reached` | Global or per-user campaign usage cap is reached. |
| `campaign_budget_exhausted` | Remaining budget is zero or lower than requested discount. |
| `coupon_not_in_campaign` | Coupon id is not present in campaign metadata. |
| `currency_mismatch` | Campaign/coupon/request currency differs. |

## 12. Common Task-Specific Errors and Fixes

Generic Go, MySQL install, Docker daemon, `.env` loading, and port errors are already covered in previous dependency files. Only campaign-specific issues are listed here.

| Error | Likely cause | Fix | Prevention |
|---|---|---|---|
| `Table 'cms_db.campaigns' doesn't exist` | Migration `004` was not applied. | Run migrations in numeric order from the shared guide. | Verify `SHOW TABLES LIKE 'campaigns';` before starting service. |
| `Unknown table 'coupon_redemptions'` while applying migration `004` | Coupon migration `003` was skipped. | Apply migration `003` before `004`, or run full migration set. | Never run later migrations first on an empty DB. |
| Duplicate column/key error on `campaign_id` | Migration `004` was manually rerun. | Check schema; do not rerun applied ALTER migration. | Use a migration version table/tool. |
| `VALIDATION_FAILED` with duration message | Campaign duration exceeds `CMS_CAMPAIGN_MAX_DURATION_DAYS`. | Shorten `ends_at` or increase env value intentionally. | Keep campaign windows under default 90 days locally. |
| `VALIDATION_FAILED` on `metadata.coupon_ids` | Linked coupon does not exist, belongs to another seller, or has different currency. | Create/list coupon first using task4 setup, then use same seller/currency. | Copy coupon id from API response, not coupon code. |
| `AUTHENTICATION_REQUIRED` | Missing `X-User-ID` or seller context on seller route. | Send `X-User-ID`, `X-Seller-ID`, role, and staff headers. | Let Gateway inject headers in real environment. |
| `PERMISSION_DENIED` | Role lacks campaign permission or seller mismatch. | Use `seller`, `seller_manager`, or `seller_catalog_editor`; keep same seller id. | Do not trust body seller ids. |
| `campaign_not_active` | New campaign is still `draft` or has been paused/completed. | Patch status to `active` for smoke test. | Remember create returns `draft`. |
| `campaign_not_started` / `campaign_expired` | UTC window does not include current time. | Use current UTC timestamps in `starts_at` and `ends_at`. | Keep local smoke windows simple and recent. |
| `campaign_budget_exhausted` | Redemption spend has reached budget, or requested discount exceeds remaining budget. | Increase budget or use a fresh campaign. | Use minor units correctly. |
| `campaign_usage_limit_reached` | Global or per-user limit is reached from existing redemptions. | Use new buyer/order/campaign or raise limit. | Separate low-limit edge-case data from smoke-test data. |
| `COUPON_NOT_IN_CAMPAIGN` | Coupon id sent in validation/redemption is not allowed by campaign metadata. | Add coupon id to `metadata.coupon_ids` or leave list empty for broad campaign. | Keep campaign/coupon fixtures aligned. |
| `CURRENCY_MISMATCH` | Campaign, coupon, request, or redemption currency differs. | Use one currency, for example `INR`, across campaign/coupon/cart/redemption. | Include currency explicitly in request payloads. |
| `AUDIT_WRITE_FAILED` | Audit table missing or DB write failed. | Run migration `001` and audit hardening migration from shared setup. | Run full migration set before service startup. |
| gRPC `internal service metadata is required` | Missing `x-service-name` for internal service call. | Add `x-service-name: cart-service` or another allowed caller. | Configure gRPC client interceptor. |
| gRPC `internal authorization is required` | Missing/wrong lower-case internal token metadata. | Send `x-internal-token` matching `CMS_INTERNAL_AUTH_TOKEN`. | Keep token and header name aligned. |

## 13. Security and Best Practices

| Practice | Why |
|---|---|
| Keep campaign endpoints behind Gateway/internal network | Service trusts identity headers; direct public access can spoof seller context. |
| Set `CMS_INTERNAL_AUTH_TOKEN` locally and in production | Prevents unauthenticated internal endpoint access. |
| Use UTC for all campaign windows | Avoids timezone bugs in start/end validation. |
| Store budgets in minor currency units | Avoids floating-point money mistakes. |
| Avoid unlimited budgets in production unless explicitly approved | `budget_amount = NULL` means no campaign spend cap. |
| Validate linked coupon seller and currency | Prevents cross-seller promotions and currency mismatch. |
| Do final budget/usage checks in transaction | Prevents overspend under concurrent order redemptions. Current repository locks rows for redemption. |
| Treat completed campaigns as historical/read-only records | Audit and finance review need stable history. |
| Use migration tracking | Manual reruns of ALTER migrations are fragile. |
| Keep `CMS_CAMPAIGN_MAX_DURATION_DAYS` bounded | Prevents accidentally year-long seller campaigns. |
| Log `request_id`, `seller_id`, `campaign_id`, and reason codes | Makes debugging campaign disputes easier. |

## 14. Missing or Misconfigured Things to Watch

| Gap / assumption | Impact | Recommendation |
|---|---|---|
| No repo-level migration runner/version table | Manual migrations can be skipped or rerun accidentally. | Add a migration tool such as `golang-migrate` or a service migration runner. |
| No service Dockerfile/compose stack | Beginner setup depends on manual Go + MySQL commands. | Add local compose with MySQL healthcheck and service container later. |
| Campaign mutation RPCs are not in current proto | Task design mentions gRPC create/list, but current proto only supports campaign reads. | Use HTTP for mutations, or intentionally extend proto later. |
| Campaign usage limits live in `metadata` JSON | Flexible but less query-friendly than typed columns. | Add typed `usage_limit` / `per_user_limit` columns if reporting/query needs grow. |
| Budget alert threshold is validated but no alert worker exists | No alert is emitted when threshold is crossed. | Add metrics/event/notification worker in a future task if required. |
| No Redis cache for campaign validation | Fine for MVP; high traffic may hit MySQL more. | Add cache only as optimization, not source of truth. |
| No Kafka/RabbitMQ campaign events | Other services cannot asynchronously react to campaign changes yet. | Add events only when a consumer exists. |
| Product Service not used by campaign flow | Good for independence, but seller product targeting must come from coupon/cart snapshots. | Keep Product Service setup separate unless testing product moderation. |

## 15. References to Previous Dependency Files

Use these instead of duplicating shared setup:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`3. Go Dependency System`
`4. Database Analysis`
`5. Environment Variables`
`8. Docker and DevOps Setup`
`9. Complete Project Run Instructions`
`10. HTTP and gRPC Endpoint Setup Notes`
`11. Common Errors and Fixes`
```

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`

Sections:
`5. Database Setup`
`11. Common Errors & Fixes`
```

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task4_Dependency.md`

Sections:
`5. Database Setup`
`9. Local Development Setup`
`10. Running the Project`
`11. Common Errors & Fixes`
```

## 16. Final Checklist

- [ ] Previous dependency docs for tasks 1-4 checked before using this file.
- [ ] Go dependencies downloaded from `${BACKEND_SERVICE_PATH}`.
- [ ] MySQL 8+ running.
- [ ] Full migration set applied in numeric order.
- [ ] `campaigns` table exists.
- [ ] `coupon_redemptions.campaign_id` exists.
- [ ] `.env` loaded into shell before `go run`.
- [ ] `CMS_CAMPAIGN_MAX_DURATION_DAYS` reviewed.
- [ ] `CMS_DB_TIMEZONE=UTC` set or defaulted.
- [ ] `CMS_INTERNAL_AUTH_TOKEN` set for local/prod internal endpoint safety.
- [ ] HTTP service healthy on `CMS_HTTP_ADDR`.
- [ ] gRPC service listening on `CMS_GRPC_ADDR` if gRPC tests are needed.
- [ ] Seller headers include `X-User-ID`, `X-Seller-ID`, `X-Roles`, and `X-Staff-Status: active`.
- [ ] Campaign create returns `draft`.
- [ ] Campaign patched to `active` before validation.
- [ ] Linked coupon ids, if used, exist for the same seller and currency.
- [ ] Coupon validation with `campaign_id` tested if running full coupon/campaign flow.
- [ ] Redemption with `campaign_id` tested to verify budget/usage counts.
- [ ] No Redis/Kafka/RabbitMQ/Product Service setup assumed for campaign-only testing.

## Final Notes

`${TASK_FILE_NAME}` mainly adds campaign-specific MySQL schema, seller-scoped campaign APIs, campaign validation, and coupon redemption governance. Shared Go, MySQL, Docker, `.env`, auth headers, and coupon engine setup should be reused from the earlier dependency files.

For a beginner developer: first make the service run using task1/task2 docs, then make coupon flow work using task4 docs, then use this file to verify campaigns on top.
