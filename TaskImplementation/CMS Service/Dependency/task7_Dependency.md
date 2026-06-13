# Project Dependency & Setup Guide

## Document Variables

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `CMS Service` |
| `TASK_FILE_NAME` | `task7.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task7_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}` |
| `BACKEND_SERVICE_PATH` | `backend/services/cms-service` |
| `PROTO_FILE_PATH` | `proto/ecommerce/cms/v1/cms.proto` |
| `GEN_GO_PATH` | `${BACKEND_SERVICE_PATH}/internal/gen/ecommerce/cms/v1` |
| `GRPC_TRANSPORT_PATH` | `${BACKEND_SERVICE_PATH}/internal/transport/grpc` |

This guide explains dependency, setup, environment, database, gRPC, and DevOps requirements for `${INPUT_FILE_PATH}`.

Simple goal: beginner developer ko clear ho ki CMS gRPC methods run/test karne ke liye kya setup chahiye, kya previous dependency files me already covered hai, aur `${TASK_FILE_NAME}` ke liye kaunsi proto/gRPC-specific cheezein verify karni hain.

Important: original implementation task file is not modified. This is a new setup/dependency document saved at `${OUTPUT_FILE_PATH}`.

## 1. Project Overview

`${TASK_FILE_NAME}` CMS gRPC contract and service-to-service access define karta hai.

Main task methods:

| gRPC method | Primary caller | Runtime purpose | Current status |
|---|---|---|---|
| `ValidateCoupon` | `cart-service`, `order-service` | Coupon preview/final validation without redemption write | Implemented in `${GRPC_TRANSPORT_PATH}/coupon_handler.go` |
| `GetSellerSettings` | API Gateway / seller dashboard | Seller return/shipping/support settings read | Implemented in `${GRPC_TRANSPORT_PATH}/seller_settings_handler.go` |
| `GetCampaign` | API Gateway, selected internal callers | Single campaign detail read | Implemented in `${GRPC_TRANSPORT_PATH}/campaign_handler.go` |
| `ListCampaigns` | Seller dashboard | Seller campaign list with pagination/status filter | Implemented in `${GRPC_TRANSPORT_PATH}/campaign_handler.go` |

Current repo reality: `${INPUT_FILE_PATH}` describes this as a blueprint, but the repo now also contains the real proto file, generated Go files, and gRPC handlers. Use the current source files as the runtime truth.

Current task-related files:

| Area | Path | Setup impact |
|---|---|---|
| Proto source | `${PROTO_FILE_PATH}` | Contract source for `ecommerce.cms.v1.CMSService`. |
| Generated Go code | `${GEN_GO_PATH}/cms.pb.go`, `${GEN_GO_PATH}/cms_grpc.pb.go` | Required for compile/run; generated from proto. Do not edit manually. |
| gRPC server registration | `${GRPC_TRANSPORT_PATH}/server.go` | Registers `CMSService` and configures interceptors/message limits. |
| gRPC auth metadata | `${GRPC_TRANSPORT_PATH}/auth.go` | Reads `x-service-name`, seller/user metadata, request id, trace id, and internal token. |
| gRPC mapping | `${GRPC_TRANSPORT_PATH}/mapper.go` | Converts proto requests/responses to domain/usecase types. |
| Service startup | `${BACKEND_SERVICE_PATH}/cmd/server/main.go` | Starts both HTTP and gRPC servers. |
| Config | `${BACKEND_SERVICE_PATH}/internal/config/config.go` | Loads `CMS_GRPC_*` and reused security/database env variables. |

Very important beginner note: Browser frontend direct gRPC call nahi karega. Browser API Gateway/HTTP ko call karega. Gateway, Cart Service, or Order Service CMS gRPC ko internal network me call karenge.

## 2. Tech Stack

Most baseline tech stack setup is already documented in previous dependency files. This section only explains what matters for `${TASK_FILE_NAME}`.

| Technology/service | Status for `${TASK_FILE_NAME}` | Why used | Beginner explanation |
|---|---|---|---|
| Go | Reused | CMS server, usecases, repositories, and gRPC handlers Go me implemented hain. | Go backend language hai jo fast services and APIs ke liye use hoti hai. |
| Go modules | Reused | `${BACKEND_SERVICE_PATH}/go.mod` dependencies track karta hai. | `go.mod` dependency list hai; `go.sum` checksum safety file hai. |
| gRPC Go | Task focus, reused dependency | Internal typed RPC methods serve karne ke liye. | gRPC service-to-service communication hai. Function call jaisa feel hota hai, but network ke through. |
| Protocol Buffers | Task focus, reused dependency | Request/response contracts strongly typed banane ke liye. | `.proto` file schema hoti hai; usse Go code generate hota hai. |
| Buf CLI | Optional unless proto changes | `buf lint` and `buf generate` run karne ke liye. | Buf proto files ko lint/generate karne ka clean tool hai. |
| `grpcurl` | Optional debugging tool | Local terminal se gRPC call test karne ke liye. | `curl` jaisa tool, but gRPC ke liye. |
| MySQL 8+ | Reused | Coupons, campaigns, seller settings, staff, and audit data source of truth hain. | MySQL relational DB hai jisme data tables me store hota hai. |
| `github.com/go-sql-driver/mysql` | Reused | Go service ko MySQL se connect karata hai. | Driver Go code aur MySQL ke beech bridge hai. |
| `net/http` | Reused | Same binary HTTP server bhi start karta hai. | Health/internal routes ke liye Go built-in HTTP server use hota hai. |
| Redis | Optional future only | Task file cache strategy mention karta hai, but current code Redis use nahi karta. | Redis fast cache hai, but current local run ke liye required nahi. |
| Kafka/RabbitMQ/NATS | Not used | Current CMS gRPC methods queue consume/publish nahi karte. | Is task ke liye queue containers start karne ki zarurat nahi hai. |

Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`2. Tech Stack Analysis`
`3. Go Dependency System`
`10. HTTP and gRPC Endpoint Setup Notes`

## 3. Required Software

Do not reinstall tools if previous CMS dependency guides are already complete.

| Software/tool | Required? | New for `${TASK_FILE_NAME}`? | Reference / note |
|---|---:|---:|---|
| Go compatible with `${BACKEND_SERVICE_PATH}/go.mod` | Yes | No | Current module declares `go 1.26.3`. |
| MySQL 8+ | Yes | No | Reuse previous MySQL setup. |
| MySQL CLI client | Recommended | No | Useful for checking required rows/tables. |
| Docker | Optional | No | Recommended beginner path for MySQL only. |
| Buf CLI | Optional | Proto-specific | Needed only if editing/regenerating proto. |
| Go protobuf plugins | Optional | Proto-specific | Needed if `buf generate` uses local plugins. |
| `grpcurl` | Optional | Proto-specific | Useful for manual gRPC smoke tests. |
| Redis | No | No | Only future cache design; no current code dependency. |
| Product Service | No for task-7 gRPC smoke tests | Reused by moderation flows | Service startup validates Product Service URL, but task-7 gRPC reads do not call it. |

Required software already explained in detail:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`3. Go Dependency System`
`4. Database Analysis`
`8. Docker and DevOps Setup`
```

## 4. Dependency Management

No new Go module dependency is introduced by `${TASK_FILE_NAME}`.

Current direct dependencies in `${BACKEND_SERVICE_PATH}/go.mod`:

| Dependency | Used for |
|---|---|
| `github.com/go-sql-driver/mysql v1.9.3` | MySQL connection and repository queries. |
| `google.golang.org/grpc v1.77.0` | gRPC server, interceptors, status codes, and generated service plumbing. |
| `google.golang.org/protobuf v1.36.10` | Generated proto messages plus `Timestamp` and `Struct` support. |

Task-specific proto generation files:

| File | Meaning |
|---|---|
| `buf.yaml` | Proto module root is `proto`, lint uses `STANDARD`, breaking rule uses `FILE`. |
| `buf.gen.yaml` | Generates Go and Go gRPC code into `${BACKEND_SERVICE_PATH}` with module trimming. |
| `${PROTO_FILE_PATH}` | Source schema. Edit this file first when changing CMS gRPC contract. |
| `${GEN_GO_PATH}` | Generated output. Do not manually edit generated `.pb.go` files. |

If you only run the service, generated files are already present. If you change proto, run from repo root:

```bash
buf lint
buf generate
cd backend/services/cms-service
go test ./...
```

If `buf` or plugins are missing, install steps are already explained in:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Section:
`10. HTTP and gRPC Endpoint Setup Notes`
```

Beginner note: `go get` is not needed for this task unless you intentionally upgrade gRPC/protobuf versions. Random dependency upgrades can create generated-code mismatch, so avoid that during setup.

## 5. Database Setup

### 5.1 Database used

| Database | Required? | Default port | Status |
|---|---:|---:|---|
| MySQL 8+ | Yes | 3306 | Reused setup; no new database engine. |

MySQL install, Docker command, DB user creation, connection string formats, and base migration flow are already documented.

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

### 5.2 Tables needed by task-7 gRPC methods

`${TASK_FILE_NAME}` adds no new SQL migration. It reuses tables created by previous CMS tasks.

| gRPC method | Main tables required | Previous dependency file |
|---|---|---|
| `ValidateCoupon` | `coupons`, `coupon_rules`, `coupon_redemptions`, optionally `campaigns` | `task4_Dependency.md`, `task5_Dependency.md` |
| `GetSellerSettings` | `seller_settings`, `seller_staff`, `cms_audit_logs` | `task2_Dependency.md`, `task1_Dependency.md` |
| `GetCampaign` | `campaigns`, `coupon_redemptions`, `seller_staff`, `cms_audit_logs` | `task5_Dependency.md` |
| `ListCampaigns` | `campaigns`, `seller_staff`, `cms_audit_logs` | `task5_Dependency.md` |

Fresh local DB recommendation: apply all current CMS migrations in numeric order as documented earlier. The service initializes multiple repositories at startup, and a beginner setup is simplest when the complete current schema exists.

Current migration order:

```text
001_create_cms_access_tables.up.sql
002_create_product_moderation_reviews.up.sql
003_create_coupon_engine_tables.up.sql
004_create_offer_campaigns.up.sql
005_create_seller_analytics_tables.up.sql
006_create_seller_settings.up.sql
007_harden_cms_audit_logs.up.sql
```

Do not duplicate migration commands here; use the exact commands from:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Section:
`4. Database Analysis`
```

### 5.3 Task-specific verification queries

After migrations, verify task-7 tables:

```bash
cd backend/services/cms-service
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW TABLES LIKE 'coupons'; SHOW TABLES LIKE 'coupon_rules'; SHOW TABLES LIKE 'campaigns'; SHOW TABLES LIKE 'seller_settings';" cms_db
```

Check important indexes for gRPC read latency:

```bash
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW INDEX FROM coupons; SHOW INDEX FROM campaigns; SHOW INDEX FROM seller_settings;" cms_db
```

Expected task-specific idea:

| Table | Why it matters |
|---|---|
| `coupons` indexed by code/status | `ValidateCoupon` must find coupon fast. |
| `campaigns` indexed by seller/status/window | `GetCampaign` and `ListCampaigns` must avoid slow scans. |
| `seller_settings` keyed by seller | `GetSellerSettings` should be one-row lookup. |
| `seller_staff` and `cms_audit_logs` | Authorization/audit flow needs these when MySQL staff source is used. |

### 5.4 Seed data note

gRPC server can start without coupon/campaign/settings rows, but smoke tests need meaningful rows.

| Method | Without seed data |
|---|---|
| `ValidateCoupon` | Likely returns `valid=false` with reason like missing/inactive coupon, or a not-found business result depending on usecase data. |
| `GetSellerSettings` | May return `NotFound`/default behavior depending on repository/usecase state. |
| `GetCampaign` | Returns `NotFound` if campaign id does not exist or seller scope does not match. |
| `ListCampaigns` | Can return an empty list for a seller. |

Seed examples and table-specific details are already covered in previous dependency files. Follow coupon setup from `task4_Dependency.md`, campaign setup from `task5_Dependency.md`, and seller settings setup from `task2_Dependency.md`.

## 6. Redis / Queue / External Services

No new Redis, Kafka, RabbitMQ, NATS, MinIO, Elasticsearch, SMTP, Stripe, Twilio, Firebase, OAuth provider, or object storage setup is introduced by `${TASK_FILE_NAME}`.

| Service | Required for task-7 local gRPC? | Status | Notes |
|---|---:|---|---|
| MySQL | Yes | Reused | Source of truth for coupons, campaigns, settings, staff, and audit data. |
| API Gateway | Not required for direct local grpcurl | Reused platform dependency | In real flow it attaches trusted metadata before calling CMS. |
| Cart Service | Not required to start CMS | Real caller | Calls `ValidateCoupon` as `x-service-name: cart-service`. |
| Order Service | Not required to start CMS | Real caller | Calls `ValidateCoupon` as `x-service-name: order-service`. |
| Product Service | No for task-7 gRPC | Reused by moderation routes | `CMS_PRODUCT_SERVICE_BASE_URL` must be valid, but task-7 gRPC reads do not call it. |
| Redis | No | Optional future cache only | Task file suggests cache strategy, but current code has no Redis client/dependency. |
| Kafka/RabbitMQ/NATS | No | Not implemented | No queue worker is needed for these gRPC methods. |

Redis note: agar future me cache add ho, seller-scoped keys and short TTL use karo. Current setup me Redis container start karna unnecessary hai.

## 7. Environment Variables

No brand-new environment variable is introduced by `${TASK_FILE_NAME}`. It uses existing CMS config.

Full `.env` setup is already covered here:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Section:
`5. Environment Variables`
```

Task-7 gRPC variables to verify in `${BACKEND_SERVICE_PATH}/.env`:

| Variable | Required? | Example | Purpose | Security note |
|---|---:|---|---|---|
| `CMS_GRPC_ADDR` | Required by config default | `:9098` | gRPC server bind address. | Use `127.0.0.1:9098` for local-only binding if needed. |
| `CMS_GRPC_MAX_RECV_MSG_BYTES` | Optional | `1048576` | Max incoming gRPC message size. | Keep bounded to avoid memory abuse. |
| `CMS_GRPC_MAX_SEND_MSG_BYTES` | Optional | `1048576` | Max outgoing gRPC message size. | Keep bounded to avoid large responses. |
| `CMS_GRPC_ALLOWED_INTERNAL_CALLERS` | Optional | `cart-service,order-service,api-gateway` | Global internal caller allowlist. | Do not allow random caller names in production. |
| `CMS_INTERNAL_AUTH_HEADER` | Required by config default | `X-Internal-Token` | Metadata/header name checked for internal auth. | Keep consistent between CMS and callers. |
| `CMS_INTERNAL_AUTH_TOKEN` | Required in production | `replace-with-local-token` | Shared token for internal gRPC/HTTP auth. | Never commit real token; production should prefer mTLS/service identity plus secret rotation. |
| `CMS_STAFF_STATUS_SOURCE` | Optional | `gateway` or `mysql` | Seller auth source for active staff checks. | Use `mysql` only when `seller_staff` rows are seeded correctly. |
| `CMS_MYSQL_DSN` or `CMS_DB_*` | Required | See previous docs | DB connection used by all gRPC methods. | Keep DB password local-only. |

Not a full `.env`, only task-7 verification snippet:

```bash
CMS_GRPC_ADDR=:9098
CMS_GRPC_ALLOWED_INTERNAL_CALLERS=cart-service,order-service,api-gateway
CMS_GRPC_MAX_RECV_MSG_BYTES=1048576
CMS_GRPC_MAX_SEND_MSG_BYTES=1048576
CMS_INTERNAL_AUTH_HEADER=X-Internal-Token
CMS_INTERNAL_AUTH_TOKEN=replace-with-local-token
```

How env is loaded:

| Behavior | Meaning |
|---|---|
| Code uses `os.Getenv` | `.env` is not auto-loaded. |
| Linux/macOS local run | Use `set -a; source .env; set +a` before `go run` so variables are exported to the Go process. |
| Wrong duration/int values | Some helpers fall back to defaults, but validation still catches invalid required config. |
| `CMS_ENV=production` | `CMS_INTERNAL_AUTH_TOKEN` and DB password/DSN become mandatory. |

## 8. Docker Setup

No new Dockerfile, Docker Compose service, volume, network, or container is introduced by `${TASK_FILE_NAME}`.

Reuse MySQL Docker setup from:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Section:
`8. Docker and DevOps Setup`
```

Current Docker/DevOps status:

| Item | Current status | Task-7 impact |
|---|---|---|
| MySQL container | Optional and reused | Required if no local MySQL server exists. |
| CMS service Dockerfile | Not found in service folder | Run service directly with Go for local beginner setup. |
| Docker Compose stack | Not found for full CMS local stack | Follow manual MySQL + Go run flow from previous docs. |
| Redis container | Not required | Do not start unless implementing future cache. |
| gRPC reflection | Not registered | `grpcurl list` without proto/reflection will fail. Use `-proto` flags. |
| gRPC health service | Not registered | Use service logs and actual method calls for gRPC verification. |

Ports:

| Service | Port | Purpose | Status |
|---|---:|---|---|
| CMS HTTP | 8087 | Health/internal HTTP routes | Reused |
| CMS gRPC | 9098 | `CMSService` internal RPCs | Reused, task-7 primary |
| MySQL | 3306 | CMS database | Reused |
| Product Service | 8080 | Product moderation dependency | Reused, not task-7 gRPC critical |
| Redis | 6379 | Optional future cache | Not required |
| Kafka/RabbitMQ | N/A | Queues | Not used |

Port conflict checks are already documented in `task1_Dependency.md`. For task-7 specifically, if `9098` is busy, update `CMS_GRPC_ADDR` and make callers/grpcurl use the same address.

## 9. Local Development Setup

### Step 1: Read previous dependency documentation first

Follow these before this file:

1. `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
2. `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`
3. `TaskImplementation/${SERVICE_NAME}/task4_Dependency.md`
4. `TaskImplementation/${SERVICE_NAME}/task5_Dependency.md`
5. `TaskImplementation/${SERVICE_NAME}/task6_Dependency.md`

They already explain clone, Go setup, MySQL install/Docker, base `.env`, migrations, ports, Product Service notes, and generic troubleshooting.

### Step 2: Go to service directory

```bash
cd backend/services/cms-service
```

### Step 3: Install only new dependencies if any

No new Go dependency is needed. For normal local run:

```bash
go mod download
```

Only if proto changes are made, return to repo root and run:

```bash
buf lint
buf generate
```

### Step 4: Setup databases/services

No new database or external service is introduced. Ensure:

| Requirement | What to check |
|---|---|
| MySQL running | `mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SELECT 1;" cms_db` |
| CMS migrations applied | Full numeric migration sequence from previous docs. |
| Task tables exist | `coupons`, `coupon_rules`, `campaigns`, `seller_settings`. |
| Optional seed data | Add coupon/campaign/settings rows if you want non-empty gRPC responses. |

### Step 5: Add only reused gRPC env values

Create or update `${BACKEND_SERVICE_PATH}/.env` using previous docs. Verify the gRPC-related variables from section 7.

Load env:

```bash
set -a
source .env
set +a
```

### Step 6: Start backend service

```bash
go run ./cmd/server
```

Expected logs:

```text
cms.mysql.connected
cms.http.started
cms.grpc.started
```

### Step 7: Verify gRPC is reachable

Current server does not enable gRPC reflection, so use the proto file with `grpcurl`.

From repo root:

```bash
grpcurl -plaintext \
  -import-path proto \
  -proto ecommerce/cms/v1/cms.proto \
  localhost:9098 \
  list ecommerce.cms.v1.CMSService
```

Expected methods include:

```text
ecommerce.cms.v1.CMSService.GetCampaign
ecommerce.cms.v1.CMSService.GetSellerSettings
ecommerce.cms.v1.CMSService.ListAuditLogs
ecommerce.cms.v1.CMSService.ListCampaigns
ecommerce.cms.v1.CMSService.ValidateCoupon
```

### Step 8: Verify `${TASK_FILE_NAME}` functionality

Validate coupon as Cart Service:

```bash
grpcurl -plaintext \
  -import-path proto \
  -proto ecommerce/cms/v1/cms.proto \
  -H "x-service-name: cart-service" \
  -H "x-internal-token: replace-with-local-token" \
  -H "x-request-id: req_local_001" \
  -d '{
    "couponCode": "SAVE10",
    "userId": "user_123",
    "cartId": "cart_123",
    "currency": "INR",
    "cartSubtotal": {"amount": 100000, "currency": "INR"},
    "items": [
      {
        "productId": "prod_1",
        "sellerId": "seller_123",
        "categoryId": "cat_1",
        "quantity": 1,
        "lineTotal": {"amount": 100000, "currency": "INR"},
        "unitPrice": {"amount": 100000, "currency": "INR"}
      }
    ]
  }' \
  localhost:9098 \
  ecommerce.cms.v1.CMSService/ValidateCoupon
```

Read seller settings with seller metadata:

```bash
grpcurl -plaintext \
  -import-path proto \
  -proto ecommerce/cms/v1/cms.proto \
  -H "x-user-id: user_123" \
  -H "x-seller-id: seller_123" \
  -H "x-roles: seller" \
  -H "x-staff-status: active" \
  -H "x-request-id: req_local_002" \
  -d '{"sellerId":"seller_123"}' \
  localhost:9098 \
  ecommerce.cms.v1.CMSService/GetSellerSettings
```

List seller campaigns:

```bash
grpcurl -plaintext \
  -import-path proto \
  -proto ecommerce/cms/v1/cms.proto \
  -H "x-user-id: user_123" \
  -H "x-seller-id: seller_123" \
  -H "x-roles: seller" \
  -H "x-staff-status: active" \
  -H "x-request-id: req_local_003" \
  -d '{"sellerId":"seller_123","page":{"limit":10,"offset":0}}' \
  localhost:9098 \
  ecommerce.cms.v1.CMSService/ListCampaigns
```

If `CMS_INTERNAL_AUTH_TOKEN` is empty locally, the `x-internal-token` metadata is not required. Production should always configure it.

## 10. Running the Project

Recommended incremental flow:

| Step | Command/action | New or reused |
|---:|---|---|
| 1 | Read previous dependency docs listed in section 9 | Reused |
| 2 | Start MySQL using previous Docker/local setup | Reused |
| 3 | Apply all CMS migrations in numeric order | Reused |
| 4 | Verify task-7 tables exist | Task-specific check |
| 5 | Create/update `${BACKEND_SERVICE_PATH}/.env` | Reused with gRPC verification |
| 6 | `set -a; source .env; set +a` | Reused |
| 7 | `go mod download` | Reused |
| 8 | `go test ./...` | Reused verification |
| 9 | `go run ./cmd/server` | Reused |
| 10 | Call task-7 gRPC methods with `grpcurl -proto ...` | Task-specific |

Task-specific test commands from service folder:

```bash
go test ./internal/transport/grpc/... ./internal/usecase/...
go test ./...
```

Task-specific proto commands from repo root:

```bash
buf lint
buf generate
```

## 11. Common Errors & Fixes

Generic Go/MySQL/Docker/internal-auth errors are already covered in previous dependency files. This section only lists gRPC/proto task-7 issues.

| Error / symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `internal service metadata is required` | `ValidateCoupon` called without `x-service-name`. | Add `-H "x-service-name: cart-service"` or `order-service`. | Client interceptor should always attach service name. |
| `internal service is not allowed` | Caller not in `CMS_GRPC_ALLOWED_INTERNAL_CALLERS`. | Use allowed caller name or update env for real trusted service. | Keep allowlist synced with service names. |
| `internal service is not allowed for this method` | Method-specific allowlist failed. | `ValidateCoupon` only accepts `cart-service`/`order-service`; settings internal path accepts `api-gateway`; campaign reads accept `api-gateway`, `cart-service`, `order-service`. | Document caller ownership before wiring clients. |
| `internal authorization is required` | `CMS_INTERNAL_AUTH_TOKEN` is set but metadata header is missing/wrong. | Send `x-internal-token` with the configured token. | Use shared gRPC client interceptor. |
| `authentication required` on seller read | Missing `x-user-id`, `x-seller-id`, roles, or active staff metadata. | Add seller metadata headers or seed staff rows if using MySQL staff source. | Let API Gateway attach trusted identity metadata. |
| `permission denied` on seller read | Role lacks permission or seller scope mismatches request seller. | Use role `seller`/`seller_manager` for settings/campaign reads and matching seller id. | Do not trust seller id from request body alone. |
| `resource not found` for campaign/settings | Row missing or seller id mismatch. | Seed data for same `seller_id` used in metadata. | Use local fixtures with known ids. |
| `grpcurl` says server does not support reflection | Reflection is not registered. | Use `-import-path proto -proto ecommerce/cms/v1/cms.proto`. | Add reflection only for safe local/dev environments if desired. |
| `unknown service ecommerce.cms.v1.CMSService` | Wrong port, stale proto, or service not started. | Check `cms.grpc.started` log and use `localhost:9098`. | Keep `CMS_GRPC_ADDR` and client target aligned. |
| `buf: command not found` | Buf CLI not installed. | Install Buf or use previous docs for proto tooling. | Add proto tooling to developer setup checklist. |
| Generated Go compile errors after proto edit | `buf generate` not run or plugin versions mismatch. | Run `buf generate`, then `go test ./...`. | Never edit generated `.pb.go` manually. |
| `go` version error | Local Go is older than `go.mod`. | Install a Go version compatible with `${BACKEND_SERVICE_PATH}/go.mod`. | Check `go version` before debugging code. |
| `DeadlineExceeded` from client | DB slow/down, service overloaded, or deadline too aggressive. | Verify MySQL and increase local test deadline slightly. | Use sensible client deadlines and monitor DB latency. |

## 12. Security & Best Practices

Task-specific security notes:

| Area | Recommendation |
|---|---|
| Internal service identity | Use `x-service-name` plus `CMS_INTERNAL_AUTH_TOKEN` locally; production should prefer mTLS/service mesh identity as well. |
| Seller identity | Trust only API Gateway/service metadata, not seller id from request body alone. |
| Allowed callers | Keep `CMS_GRPC_ALLOWED_INTERNAL_CALLERS` tight: `cart-service`, `order-service`, `api-gateway`. |
| Method-specific authorization | Keep `ValidateCoupon` limited to Cart/Order. Keep seller reads scoped by seller metadata and permissions. |
| Message size | Keep `CMS_GRPC_MAX_RECV_MSG_BYTES` and `CMS_GRPC_MAX_SEND_MSG_BYTES` bounded. |
| Logging | Current interceptor logs method/code/request id/trace id/caller. Do not log raw tokens, full metadata, passwords, or raw customer PII. |
| Coupon code logs | Prefer redacted/hash if logging coupon-specific events later. |
| Proto compatibility | Add new fields with new numbers. Do not reuse deleted field numbers. Use `reserved` when removing fields. |
| Generated code | Do not manually edit `${GEN_GO_PATH}`. Change proto and regenerate. |
| Plaintext gRPC | `grpcurl -plaintext` is local-only. Production should use TLS/mTLS. |
| Redis cache future | Cache keys must include seller scope where relevant and use short TTLs for coupon/campaign data. |

Beginner best practices specific to `${TASK_FILE_NAME}`:

- Keep gRPC handlers thin: auth, validation, mapping only.
- Keep business rules in usecases, not in proto mappers.
- Use request id and trace id metadata for every internal call.
- Use client deadlines. For local manual tests, 1-2 seconds is okay; production callers should use tighter method-specific deadlines.
- Treat invalid coupons as business result (`valid=false`) instead of gRPC technical error.
- Keep Cart/Order away from CMS DB. They should send cart/order snapshot to CMS gRPC.

## 13. Missing or Misconfigured Things

Current audit after inspecting `${INPUT_FILE_PATH}` and current implementation:

| Item | Current state | Impact | Suggested fix |
|---|---|---|---|
| `${INPUT_FILE_PATH}` says no backend/proto code was created | Current repo now has proto/generated/handler files | Docs can confuse new developers | Treat current source as runtime truth and keep task file note historical. |
| gRPC reflection not registered | `grpcurl list` without proto fails | Beginner debugging friction | Add reflection in local/dev only, or keep using `grpcurl -proto`. |
| No gRPC health service registered | No standard `grpc.health.v1.Health/Check` | Load balancers/probes cannot use standard gRPC health | Add gRPC health service when deployment hardens. |
| No TLS/mTLS config in CMS gRPC server | Local plaintext only | Production internal traffic would be weak if deployed as-is | Add TLS/mTLS or service mesh policy for production. |
| `api/master-api.json` service registry is stale vs proto | Proto implements `GetCampaign`/`ListCampaigns`, registry section does not list them | Gateway/client generation may miss methods | Update API registry to match proto when gateway contract is finalized. |
| Proto uses local `Money`/`PageRequest` messages | Task blueprint mentioned shared common protos | Cross-service consistency can drift | Standardize common money/pagination protos in a backward-compatible proto update if platform requires it. |
| No service Dockerfile/compose stack found | Local onboarding relies on manual Go run + MySQL Docker | Repeated setup can drift | Add CMS Dockerfile and local compose stack later. |
| No Redis cache implementation | Task file documents optional cache strategy only | High read traffic may hit MySQL directly | Add Redis only when performance data justifies it. |
| No migration runner/version table | Manual migration tracking can drift | Duplicate alters/reruns can fail | Add a migration tool workflow such as goose/migrate. |

## 14. References to Previous Dependency Files

Previous dependency files inside `TaskImplementation/${SERVICE_NAME}/` were checked before generating this file.

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `2. Tech Stack Analysis` | Same Go, `net/http`, gRPC, protobuf, MySQL, and service runtime foundation. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `3. Go Dependency System` | Same Go modules, `go mod download`, `go test`, and dependency troubleshooting. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `4. Database Analysis` | MySQL install, Docker command, DB user, and full migration commands are unchanged. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `5. Environment Variables` | Full `.env` setup already exists; this file only highlights gRPC env values. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `7. Ports and Networking` | CMS HTTP, CMS gRPC, MySQL, and Product Service ports are reused. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `8. Docker and DevOps Setup` | No new Docker service/container is introduced. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `10. HTTP and gRPC Endpoint Setup Notes` | Baseline proto/gRPC setup is reused. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `11. Common Errors and Fixes` | Generic Go/MySQL/Docker/internal auth errors are unchanged. |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | Same `cms_db`, MySQL 8+, InnoDB, UTF-8, migration order, and seller settings table foundation. |
| `TaskImplementation/${SERVICE_NAME}/task4_Dependency.md` | `5. Database Setup` | Coupon tables and `ValidateCoupon` data requirements are reused. |
| `TaskImplementation/${SERVICE_NAME}/task4_Dependency.md` | `6. Redis / Queue / External Services` | Confirms no Redis/queue dependency for current coupon validation runtime. |
| `TaskImplementation/${SERVICE_NAME}/task5_Dependency.md` | `2. What task adds` and `5. Database Setup` | Campaign table, campaign gRPC reads, campaign-coupon linkage, and migration notes are reused. |
| `TaskImplementation/${SERVICE_NAME}/task5_Dependency.md` | `8. Docker, Ports, and DevOps Setup` | Same CMS HTTP/gRPC/MySQL runtime shape; no new container or port. |
| `TaskImplementation/${SERVICE_NAME}/task6_Dependency.md` | `7. Environment Variables` | Confirms current service uses direct `os.Getenv` and `.env` must be sourced manually. |
| `TaskImplementation/${SERVICE_NAME}/task6_Dependency.md` | `13. Missing or Misconfigured Things` | Reuses notes about missing Docker Compose and migration runner where still unchanged. |

## 15. Final Checklist

Use this checklist for `${TASK_FILE_NAME}` setup:

- [ ] Previous dependency documentation checked first.
- [ ] No duplicate Go/MySQL/Docker baseline setup copied into this file.
- [ ] Go version is compatible with `${BACKEND_SERVICE_PATH}/go.mod`.
- [ ] MySQL is running.
- [ ] Full CMS migrations are applied in numeric order.
- [ ] `coupons`, `coupon_rules`, `campaigns`, and `seller_settings` tables exist.
- [ ] `${BACKEND_SERVICE_PATH}/.env` exists and has reused base config.
- [ ] `CMS_GRPC_ADDR` is set or defaulting to `:9098`.
- [ ] `CMS_GRPC_ALLOWED_INTERNAL_CALLERS` includes only trusted callers.
- [ ] `CMS_INTERNAL_AUTH_TOKEN` is set for production-like testing.
- [ ] `.env` is loaded with `set -a; source .env; set +a` before running the service.
- [ ] `buf lint` passes if proto was changed.
- [ ] `buf generate` was run if proto was changed.
- [ ] Generated `.pb.go` files were not edited manually.
- [ ] `go test ./...` passes from `${BACKEND_SERVICE_PATH}`.
- [ ] Service starts and logs `cms.mysql.connected`, `cms.http.started`, and `cms.grpc.started`.
- [ ] `grpcurl` uses `-import-path proto -proto ecommerce/cms/v1/cms.proto` because reflection is not registered.
- [ ] `ValidateCoupon` tested with `x-service-name: cart-service` or `order-service`.
- [ ] Seller read methods tested with `x-user-id`, `x-seller-id`, `x-roles`, and `x-staff-status`.
- [ ] Logs checked for request id, trace id, caller service, gRPC method, and status code.
- [ ] No Redis/Kafka/RabbitMQ setup was added because current code does not use them.
- [ ] Missing production items are understood: TLS/mTLS, gRPC health, reflection decision, Dockerfile/compose, migration runner.

Final simple summary: `${TASK_FILE_NAME}` does not add a new database, Docker container, queue, or Go package. It mainly requires the existing CMS Go/MySQL setup plus proto/gRPC tooling awareness, correct `CMS_GRPC_*` configuration, trusted gRPC metadata, generated code consistency, and grpcurl smoke tests using the proto file.
