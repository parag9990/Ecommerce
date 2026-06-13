# Project Dependency & Setup Guide

## Document Variables

| Variable | Value |
|---|---|
| `SERVICE_NAME` | `CMS Service` |
| `TASK_FILE_NAME` | `task8.md` |
| `INPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | `task8_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}` |
| `BACKEND_SERVICE_PATH` | `backend/services/cms-service` |
| `AUDIT_BASE_MIGRATION_PATH` | `${BACKEND_SERVICE_PATH}/migrations/001_create_cms_access_tables.up.sql` |
| `AUDIT_HARDENING_MIGRATION_PATH` | `${BACKEND_SERVICE_PATH}/migrations/007_harden_cms_audit_logs.up.sql` |
| `PROTO_FILE_PATH` | `proto/ecommerce/cms/v1/cms.proto` |

This guide explains dependency, setup, environment, database, and DevOps requirements for `${INPUT_FILE_PATH}`.

Simple goal: beginner developer ko clear ho ki seller dashboard audit logs run/test karne ke liye kya setup chahiye, kya previous dependency docs me already covered hai, aur `${TASK_FILE_NAME}` ke liye kaunsi audit-specific cheezein verify karni hain.

Important: original implementation task file is not modified. This is a new setup/dependency document saved at `${OUTPUT_FILE_PATH}`.

## 1. Project Overview

`${TASK_FILE_NAME}` seller dashboard actions ke audit logs cover karta hai. Audit log ka kaam hai: kaun actor tha, kis seller workspace me action hua, kya action hua, kis resource par hua, request/trace id kya tha, before/after snapshot kya tha, aur action kab hua.

Current repo reality: `${INPUT_FILE_PATH}` documentation-style implementation guide hai, but current backend source me audit logging ka runtime support already present hai.

| Area | Path | Setup impact |
|---|---|---|
| Domain model and redaction | `${BACKEND_SERVICE_PATH}/internal/domain/audit.go` | `AuditEvent`, `AuditEntry`, filters, cursor page, and snapshot redaction rules. |
| Audit write hook | `${BACKEND_SERVICE_PATH}/internal/usecase/authorize_seller_action.go` | Mutation permissions audit recorder ko call karte hain. |
| Audit timeline usecase | `${BACKEND_SERVICE_PATH}/internal/usecase/audit_logs.go` | Default range, page size, filters, and `cms:audit_logs:read` authorization. |
| Persistent repository | `${BACKEND_SERVICE_PATH}/internal/repository/mysql_audit_repository.go` | `cms_audit_logs` me insert/list karta hai. |
| Logger fallback/helper | `${BACKEND_SERVICE_PATH}/internal/repository/logger_audit_repository.go` | Structured log sink available hai, but server currently MySQL recorder wire karta hai. |
| HTTP route | `${BACKEND_SERVICE_PATH}/internal/transport/http/handler.go` | `GET /api/v1/seller/audit-logs` and internal mirror route. |
| gRPC route | `${BACKEND_SERVICE_PATH}/internal/transport/grpc/audit_handler.go` | `ListAuditLogs` gRPC method. |
| Proto contract | `${PROTO_FILE_PATH}` | `ListAuditLogsRequest`, `AuditLog`, `ListAuditLogsResponse`. |
| Base audit table | `${AUDIT_BASE_MIGRATION_PATH}` | Creates `cms_audit_logs` with base columns. |
| Audit hardening migration | `${AUDIT_HARDENING_MIGRATION_PATH}` | Adds request/trace/decision/role/hash fields and indexes. |
| Config | `${BACKEND_SERVICE_PATH}/internal/config/config.go` | Loads `CMS_AUDIT_*` guardrail env variables. |

Very important beginner note: Audit logs normal application logs jaise temporary debugging logs nahi hain. Ye dispute, security review, and traceability ke liye persistent evidence hote hain. Isliye MySQL table and migration setup mandatory hai.

## 2. Tech Stack

Most baseline technology setup is already documented in previous dependency files. Is section me only audit-specific usage explain kiya gaya hai.

| Technology/service | Status for `${TASK_FILE_NAME}` | Why used | Beginner explanation |
|---|---|---|---|
| Go | Reused | Audit domain, usecase, repository, HTTP/gRPC handlers Go me implemented hain. | Go backend language hai jo fast APIs and services ke liye use hoti hai. |
| Go modules | Reused | No new external package was added only for audit logs. | `go.mod` dependency list hai, `go.sum` checksum safety file hai. |
| `net/http` | Reused with audit route | Seller dashboard audit timeline HTTP endpoint serve karta hai. | Ye Go ka built-in HTTP server hai; Fiber/Gin yahan use nahi hua. |
| gRPC | Reused with audit method | Internal service-to-service audit timeline read ke liye `ListAuditLogs` method available hai. | gRPC typed service call hai jo internal services ke beech use hota hai. |
| Protocol Buffers | Reused with audit messages | Audit log request/response schema strongly typed banata hai. | `.proto` contract file hoti hai jisse generated Go code banta hai. |
| MySQL 8+ | Required and reused | `cms_audit_logs` persistent source of truth hai. | MySQL table-based relational DB hai. Audit logs indexed rows me store hote hain. |
| `github.com/go-sql-driver/mysql` | Reused | Go service ko MySQL se connect karata hai. | Driver Go code aur MySQL ke beech bridge hai. |
| `database/sql` | Reused | Repository insert/list queries execute karta hai. | Go ka standard DB layer hai. |
| `log/slog` | Reused | Audit failures and list activity structured JSON logs me aati hai. | Structured logs debugging and monitoring easy banate hain. |
| API Gateway/Auth headers | Required | Seller identity, roles, staff status, request id, trace id trusted context se aate hain. | Gateway JWT verify karke backend ko safe identity headers bhejta hai. |
| Redis | Not used | Current code audit logs ke liye Redis client use nahi karta. | Redis cache hai, but audit source of truth ke liye required nahi. |
| Kafka/RabbitMQ/NATS | Not used | Current audit write/list flow queue publish/consume nahi karta. | Queue container start karne ki zarurat nahi hai. |
| Docker | Optional | MySQL local container ke liye useful hai. | Beginner setup me Docker se DB start karna easy hota hai. |

Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`2. Tech Stack Analysis`
`3. Go Dependency System`
`4. Database Analysis`

## 3. Required Software

Do not reinstall tools if previous setup already works.

| Software/tool | Required? | New for `${TASK_FILE_NAME}`? | Reference / note |
|---|---:|---:|---|
| Go compatible with `${BACKEND_SERVICE_PATH}/go.mod` | Yes | No | Current module declares `go 1.26.3`. |
| MySQL 8+ | Yes | No new engine, but audit hardening migration is required | Reuse previous MySQL setup. |
| MySQL CLI client | Strongly recommended | No | Needed for migration and verification queries. |
| Docker | Optional | No | Use only if running MySQL in container. |
| `curl` | Recommended | Audit HTTP smoke test | Used to call audit timeline route locally. |
| `grpcurl` | Optional | Audit gRPC smoke test | Useful because gRPC reflection is not registered. |
| Buf CLI | Optional | Only if proto changes | Not needed for normal run because generated files already exist. |
| Redis/Kafka/RabbitMQ | No | No | Current audit implementation does not connect to them. |

Full install/setup steps for Go, MySQL, Docker, and `.env` are already documented in:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`3. Go Dependency System`
`4. Database Analysis`
`5. Environment Variables`
`8. Docker and DevOps Setup`
```

## 4. Dependency Management

No new Go module dependency is introduced by `${TASK_FILE_NAME}`.

Current direct dependencies in `${BACKEND_SERVICE_PATH}/go.mod`:

| Dependency | Used for audit logs |
|---|---|
| `github.com/go-sql-driver/mysql v1.9.3` | MySQL connection and audit repository queries. |
| `google.golang.org/grpc v1.77.0` | `ListAuditLogs` gRPC server method and generated plumbing. |
| `google.golang.org/protobuf v1.36.10` | Audit proto messages, `Timestamp`, and `Struct` fields. |

Baseline Go dependency commands are reused:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Section:
`3. Go Dependency System`
```

Task-specific verification commands from service folder:

```bash
cd backend/services/cms-service
go test ./internal/domain ./internal/usecase ./internal/repository ./internal/transport/http ./internal/transport/grpc
go test ./...
go build ./cmd/server
```

Proto commands are only needed if audit proto contract changes:

```bash
buf lint
buf generate
cd backend/services/cms-service
go test ./...
```

Beginner note: Generated `.pb.go` files manually edit mat karo. Proto change karo, then `buf generate` run karo.

## 5. Database Setup

### 5.1 Database used

| Database | Required? | Default port | Status |
|---|---:|---:|---|
| MySQL 8+ | Yes | 3306 | Reused database engine with audit-specific table/migration. |

MySQL ek relational database hai jisme rows and columns wali tables hoti hain. Is task me `cms_audit_logs` table seller actions ka permanent, searchable timeline store karti hai.

MySQL installation, Docker command, DB user creation, connection string formats, and base migration commands are already explained here:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`4. Database Analysis`
`8. Docker and DevOps Setup`
`9. Complete Project Run Instructions`
```

### 5.2 Audit table and migration requirement

Audit logs require the table created by:

```text
${AUDIT_BASE_MIGRATION_PATH}
```

Then `${TASK_FILE_NAME}` production-ready columns/indexes require:

```text
${AUDIT_HARDENING_MIGRATION_PATH}
```

Current migration order for a fresh local DB:

```text
001_create_cms_access_tables.up.sql
002_create_product_moderation_reviews.up.sql
003_create_coupon_engine_tables.up.sql
004_create_offer_campaigns.up.sql
005_create_seller_analytics_tables.up.sql
006_create_seller_settings.up.sql
007_harden_cms_audit_logs.up.sql
```

Why all migrations matter: service startup initializes repositories for audit, product moderation, coupons, campaigns, analytics, settings, and optionally staff. Fresh beginner setup sab migrations numeric order me apply kare to startup predictable rahega.

### 5.3 Audit table fields

Base columns from migration `001`:

| Column | Meaning |
|---|---|
| `id` | Internal auto-increment sequence. Cursor pagination me use hota hai. |
| `audit_id` | Public/stable unique audit identifier. |
| `seller_id` | Seller workspace scope. Every list query must filter by this. |
| `actor_user_id` | User/staff/system actor jisne action kiya. |
| `action` | Permission/action string, e.g. `cms:campaigns:update`. |
| `resource_type` | Business object type, e.g. `campaign`, `coupon`, `product_review`. |
| `resource_id` | Business object id. |
| `before_json` | Change se pehle compact snapshot. |
| `after_json` | Change ke baad compact snapshot. |
| `created_at` | UTC timestamp for timeline order. |

Hardening columns from migration `007`:

| Column | Meaning |
|---|---|
| `actor_roles_json` | Action ke time actor ke effective roles. |
| `request_id` | Gateway/service log correlation id. |
| `trace_id` | Distributed trace correlation id. |
| `decision` | `allowed` or `denied`. Current writes mainly allowed actions store karte hain. |
| `reason` | Manual reason or denial reason when available. |
| `ip_hash` | Privacy-safe network clue; raw IP nahi. |
| `user_agent_hash` | Privacy-safe device/browser clue; raw UA nahi. |

### 5.4 Audit indexes

| Index | Why it matters |
|---|---|
| `uk_cms_audit_id` | Duplicate audit ids block karta hai. |
| `idx_cms_audit_seller_created` | Seller activity timeline fast banata hai. |
| `idx_cms_audit_actor_created` | Actor-specific activity filter fast banata hai. |
| `idx_cms_audit_resource` | Resource history lookup fast banata hai. |
| `idx_cms_audit_request` | Request id se app/gateway logs join karna easy hota hai. |
| `idx_cms_audit_action_created` | Action filter with time ordering improve hota hai. |

### 5.5 Task-specific DB verification

After migrations, verify audit columns and indexes:

```bash
cd backend/services/cms-service
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW COLUMNS FROM cms_audit_logs;" cms_db
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW INDEX FROM cms_audit_logs;" cms_db
```

Quick row count check:

```bash
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SELECT COUNT(*) AS audit_log_count FROM cms_audit_logs;" cms_db
```

Optional local seed for read testing:

```sql
INSERT INTO cms_audit_logs (
  audit_id,
  seller_id,
  actor_user_id,
  actor_roles_json,
  action,
  resource_type,
  resource_id,
  request_id,
  decision,
  before_json,
  after_json,
  created_at
) VALUES (
  'audit_local_001',
  'seller_123',
  'user_123',
  JSON_ARRAY('seller'),
  'cms:campaigns:update',
  'campaign',
  'camp_123',
  'req_local_audit_001',
  'allowed',
  JSON_OBJECT('status', 'draft'),
  JSON_OBJECT('status', 'active'),
  UTC_TIMESTAMP(6)
);
```

Do not insert real passwords, tokens, raw IPs, full addresses, or payment data into audit snapshots.

### 5.6 Connection string and credentials

No new DB credential variable is introduced by `${TASK_FILE_NAME}`. The audit repository uses the same CMS MySQL connection.

Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Sections:
`4. Database Analysis`
`5. Environment Variables`

Credential placement remains:

| Variable/source | Purpose |
|---|---|
| `${BACKEND_SERVICE_PATH}/.env` | Local env file to source before `go run`. |
| `CMS_MYSQL_DSN` | Direct MySQL DSN if you prefer one string. |
| `CMS_DB_HOST`, `CMS_DB_PORT`, `CMS_DB_NAME`, `CMS_DB_USER`, `CMS_DB_PASSWORD` | Component DB config if DSN is empty. |

## 6. Redis / Queue / External Services

No new Redis, Kafka, RabbitMQ, NATS, MinIO, Elasticsearch, SMTP, Stripe, Twilio, Firebase, OAuth provider, Kubernetes manifest, or object storage setup is introduced by `${TASK_FILE_NAME}`.

| Service | Required for audit logs? | Status | Notes |
|---|---:|---|---|
| MySQL | Yes | Reused with audit table | Persistent source of truth for audit entries. |
| API Gateway/Auth | Required in full platform | Reused | Supplies trusted identity headers/metadata. |
| Product Service | Not required for audit list route | Reused elsewhere | Product moderation mutations can create audit logs and may call Product Service. |
| Redis | No | Not implemented | Do not start a Redis container for this task. |
| Kafka/RabbitMQ/NATS | No | Not implemented | Current audit writes are synchronous DB inserts, not queued events. |
| Log aggregation | Optional operational dependency | Not in local setup | Useful for `cms.audit_logs.*` and audit write failure logs. |

Beginner note: Audit logs ko cache ya queue me rely mat karo as source of truth. Cache optional hota hai; audit evidence permanent DB table me hona chahiye.

## 7. Environment Variables

Full `.env` setup is already documented in:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Section:
`5. Environment Variables`
```

`${TASK_FILE_NAME}` introduces/uses these audit-specific guardrails:

| Variable | Required? | Example | Purpose | Security/config note |
|---|---:|---|---|---|
| `CMS_AUDIT_DEFAULT_RANGE_DAYS` | Optional | `30` | `from` absent ho to last N days default range. | Must be positive. Large default range slow queries kar sakta hai. |
| `CMS_AUDIT_DEFAULT_PAGE_SIZE` | Optional | `20` | `page_size` absent ho to default rows. | Must be positive and not bigger than max page size. |
| `CMS_AUDIT_MAX_PAGE_SIZE` | Optional | `100` | Max audit rows per page. | Must be between 1 and 100. Keep bounded to avoid heavy responses. |

Task-specific `.env` snippet only:

```env
# Seller dashboard audit log timeline guardrails.
CMS_AUDIT_DEFAULT_RANGE_DAYS=30
CMS_AUDIT_DEFAULT_PAGE_SIZE=20
CMS_AUDIT_MAX_PAGE_SIZE=100
```

Reused variables still required for audit log runtime:

| Variable | Why audit logs need it |
|---|---|
| `CMS_HTTP_ADDR` | HTTP audit route bind address. |
| `CMS_GRPC_ADDR` | gRPC `ListAuditLogs` bind address. |
| `CMS_INTERNAL_AUTH_HEADER` | Internal HTTP/gRPC auth header/metadata name. |
| `CMS_INTERNAL_AUTH_TOKEN` | Protects internal routes in production-like setup. |
| `CMS_STAFF_STATUS_SOURCE` | Controls whether staff active status comes from gateway header or MySQL. |
| `CMS_MYSQL_DSN` or `CMS_DB_*` | Audit repository DB connection. |

How `.env` is loaded:

| Behavior | Meaning |
|---|---|
| Code uses `os.Getenv` | `.env` auto-load nahi hota. |
| Local shell run | `set -a; source .env; set +a` required before `go run`. |
| Invalid audit integers | `envInt` may fallback for parse errors; validation catches zero/negative/out-of-range values. |
| `CMS_ENV=production` | Internal token and DB password/DSN validation stricter ho jati hai. |

Common audit env mistakes:

| Mistake | Result | Fix |
|---|---|---|
| `CMS_AUDIT_MAX_PAGE_SIZE=1000` | Config validation fails because max allowed is 100. | Use `100` or less. |
| `CMS_AUDIT_DEFAULT_PAGE_SIZE` bigger than max | Config validation fails. | Keep default <= max. |
| `.env` not sourced | Defaults may apply, DB/token may be missing. | Source `${BACKEND_SERVICE_PATH}/.env` before service start. |
| Weak `CMS_INTERNAL_AUTH_TOKEN` | Internal audit route easier to misuse. | Use a long random token outside local throwaway setup. |

## 8. Docker Setup

No new Dockerfile, Docker Compose service, volume, network, restart policy, healthcheck, or container is added by `${TASK_FILE_NAME}`.

Reuse MySQL Docker setup from:

```md
Refer:
`TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`

Section:
`8. Docker and DevOps Setup`
```

Current Docker/DevOps status:

| Item | Current status | Audit task impact |
|---|---|---|
| MySQL container | Optional and reused | Required only if no local MySQL server exists. |
| CMS service Dockerfile | Not found under `${BACKEND_SERVICE_PATH}` | Run service directly with Go for local beginner setup. |
| Docker Compose stack | Not found for full local service | Previous docs provide MySQL-only compose pattern. |
| Redis container | Not required | Do not start for audit logs. |
| Queue containers | Not required | Audit implementation is synchronous DB write/list. |
| Migration runner/version table | Not found | Apply SQL files carefully in numeric order. |

Ports:

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Backend HTTP | 8087 | `/healthz` and audit HTTP route | Reused |
| Backend gRPC | 9098 | `ListAuditLogs` internal RPC | Reused |
| MySQL | 3306 | `cms_audit_logs` storage | Reused |
| Product Service | 8080 | Product moderation dependency | Reused, not required for audit list smoke test |
| Redis | 6379 | Cache | Not used |
| Kafka/RabbitMQ | N/A | Queue | Not used |

If port `9098` or `8087` is busy, update `CMS_GRPC_ADDR` or `CMS_HTTP_ADDR` and use the same port in smoke-test commands.

## 9. Local Development Setup

### Step 1: Read previous dependency documentation

Follow these first:

1. `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md`
2. `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md`
3. `TaskImplementation/${SERVICE_NAME}/task4_Dependency.md`
4. `TaskImplementation/${SERVICE_NAME}/task5_Dependency.md`
5. `TaskImplementation/${SERVICE_NAME}/task6_Dependency.md`
6. `TaskImplementation/${SERVICE_NAME}/task7_Dependency.md`

They already explain clone flow, Go modules, MySQL install/Docker, base `.env`, migrations, ports, gRPC tooling, and generic troubleshooting.

### Step 2: Go to project directory

```bash
cd backend/services/cms-service
```

### Step 3: Install only new dependencies if any

No new dependency is required for `${TASK_FILE_NAME}`. Normal local command:

```bash
go mod download
```

### Step 4: Setup only new databases/services if any

No new database engine or external service is needed. Make sure MySQL is running and migrations include `001` and `007`.

Task-specific verification:

```bash
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SHOW COLUMNS FROM cms_audit_logs;" cms_db
```

### Step 5: Add only new or changed environment variables

Add/verify only the audit guardrails in `${BACKEND_SERVICE_PATH}/.env`:

```env
CMS_AUDIT_DEFAULT_RANGE_DAYS=30
CMS_AUDIT_DEFAULT_PAGE_SIZE=20
CMS_AUDIT_MAX_PAGE_SIZE=100
```

Then load env:

```bash
set -a
source .env
set +a
```

### Step 6: Run migrations if needed

If fresh DB, apply all migrations in numeric order using the exact commands from previous dependency docs.

If only audit hardening is missing and base table already exists:

```bash
mysql -h 127.0.0.1 -P 3306 -u cms_user -p cms_db < migrations/007_harden_cms_audit_logs.up.sql
```

Do not blindly rerun `007` if it was already applied; duplicate column/index errors aa sakte hain because there is no migration version table in this service yet.

### Step 7: Start backend service

```bash
go run ./cmd/server
```

Expected logs:

```text
cms.mysql.connected
cms.http.started
cms.grpc.started
```

### Step 8: Verify APIs or functionality related to `${TASK_FILE_NAME}`

Health check:

```bash
curl http://localhost:8087/healthz
```

HTTP audit list smoke test:

```bash
curl -s "http://localhost:8087/api/v1/seller/audit-logs?page_size=10" \
  -H "X-Internal-Token: replace-with-local-token" \
  -H "X-User-ID: user_123" \
  -H "X-Seller-ID: seller_123" \
  -H "X-Roles: seller" \
  -H "X-Staff-Status: active" \
  -H "X-Request-ID: req_local_audit_001"
```

Useful query params:

| Param | Example | Meaning |
|---|---|---|
| `from` | `2026-06-01` or `2026-06-01T00:00:00Z` | Start time. |
| `to` | `2026-06-05T23:59:59Z` | End time. |
| `actor_user_id` | `user_123` | Actor filter. |
| `action` | `cms:campaigns:update` | Action filter. |
| `resource_type` | `campaign` | Resource type filter. |
| `resource_id` | `camp_123` | Resource id filter. |
| `page_size` | `20` | Rows per page, capped by config. |
| `cursor` | returned `next_cursor` | Next page cursor. |

gRPC audit list smoke test from repo root:

```bash
grpcurl -plaintext \
  -import-path proto \
  -proto ecommerce/cms/v1/cms.proto \
  -H "x-user-id: user_123" \
  -H "x-seller-id: seller_123" \
  -H "x-roles: seller" \
  -H "x-staff-status: active" \
  -H "x-request-id: req_local_audit_002" \
  -H "x-internal-token: replace-with-local-token" \
  -d '{"sellerId":"seller_123","pageSize":10}' \
  localhost:9098 \
  ecommerce.cms.v1.CMSService/ListAuditLogs
```

If `CMS_INTERNAL_AUTH_TOKEN` is empty in local env, token metadata/header may not be required. Production-like testing should always set and send it.

## 10. Running the Project

Recommended incremental flow:

| Step | Command/action | New or reused |
|---:|---|---|
| 1 | Read previous dependency docs listed in section 9 | Reused |
| 2 | Start MySQL using previous local/Docker setup | Reused |
| 3 | Apply all migrations in numeric order, including `007` | Audit-specific verification |
| 4 | Verify `cms_audit_logs` columns and indexes | Task-specific |
| 5 | Create/update `${BACKEND_SERVICE_PATH}/.env` | Reused with `CMS_AUDIT_*` checks |
| 6 | `set -a; source .env; set +a` | Reused |
| 7 | `go mod download` | Reused |
| 8 | `go test ./...` | Reused verification |
| 9 | `go run ./cmd/server` | Reused |
| 10 | Call HTTP or gRPC audit list route | Task-specific |
| 11 | Perform a coupon/campaign/product moderation mutation, then list audit logs | Task-specific end-to-end check |

Example audit row verification after a mutation:

```bash
mysql -h 127.0.0.1 -P 3306 -u cms_user -p -e "SELECT audit_id, seller_id, actor_user_id, action, resource_type, resource_id, request_id, decision, created_at FROM cms_audit_logs ORDER BY created_at DESC LIMIT 5;" cms_db
```

Expected behavior:

| Scenario | Expected result |
|---|---|
| No audit rows yet | API returns empty `audit_logs` list. |
| Seed row for matching seller exists | API returns that row. |
| Request seller differs from auth seller | HTTP/gRPC returns forbidden. |
| Page size over max | Usecase caps to max instead of returning huge page. |
| Invalid date range | API returns invalid audit filter/date range error. |

## 11. Common Errors & Fixes

Generic Go/MySQL/Docker/internal-auth errors are already documented in previous dependency files. This section only adds audit-specific issues.

| Error / symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `cms.audit_repository.init_failed` on startup | DB connection exists but audit repository initialization got nil/invalid DB, or startup wiring failed. | Check `cms.mysql.connected` log and DB config. | Keep one valid MySQL connection before initializing repositories. |
| `Error 1146: Table 'cms_db.cms_audit_logs' doesn't exist` | Migration `001` not applied. | Apply migrations from `001` onward. | Fresh setup checklist must include all migrations. |
| `Unknown column 'actor_roles_json'` or `request_id` | Migration `007` not applied. | Apply `${AUDIT_HARDENING_MIGRATION_PATH}` once. | Verify columns after migration. |
| Duplicate column/index error while applying `007` | Migration already applied manually. | Do not rerun blindly; inspect `SHOW COLUMNS` and `SHOW INDEX`. | Add migration versioning tool later. |
| `AUDIT_WRITE_FAILED` on mutation API | MySQL insert failed, missing required audit fields, JSON marshal error, or DB issue. | Check service logs and `cms_audit_logs` schema. | Keep audit table migrated and snapshots JSON-safe. |
| `audit_logs_temporarily_unavailable` | List query failed or staff lookup failed. | Check MySQL connectivity, table schema, and `CMS_STAFF_STATUS_SOURCE`. | Monitor DB latency/errors and seed staff if using MySQL staff source. |
| `permission_denied` on audit list | Actor lacks `cms:audit_logs:read` or role is not owner/manager. | Use role `seller` or `seller_manager` for local smoke test. | Keep audit read access limited. |
| `seller_context_required` | Missing trusted seller id header/metadata. | Send `X-Seller-ID` for HTTP or `x-seller-id` for gRPC. | Gateway should always attach seller context for seller dashboard routes. |
| `invalid_audit_log_filter` | `from >= to`, negative page size, or invalid cursor. | Use valid date range and cursor returned by API. | UI should validate filters before request. |
| Empty `audit_logs` but no error | No matching rows for seller/filter/date range. | Insert seed row or perform audited mutation. | Default range is last `CMS_AUDIT_DEFAULT_RANGE_DAYS`; old rows need explicit `from`. |
| `grpcurl` reflection error | gRPC reflection is not registered. | Use `-import-path proto -proto ecommerce/cms/v1/cms.proto`. | Document proto-based grpcurl command for local tests. |
| Raw secret appears in snapshot | Usecase passed sensitive field name not caught by review, or key naming bypassed expectations. | Fix snapshot builder and scrub DB if needed. | Use `RedactAuditSnapshot` and never add secrets to snapshots. |

## 12. Security & Best Practices

Task-specific security notes:

| Area | Recommendation |
|---|---|
| Source of truth | MySQL `cms_audit_logs` should be persistent audit source, not only stdout logs. |
| Seller scoping | Every list query must include seller scope from trusted auth context. |
| Read permission | Keep audit timeline behind `cms:audit_logs:read`. |
| Role access | `seller` and `seller_manager` can read; catalog/order staff should not read full audit history. |
| Fail-closed writes | High-risk seller mutations should fail if audit write fails. Current usecases wrap failures as `ErrAuditWriteFailed`. |
| Snapshot minimization | Store compact before/after summaries, not full DB rows or large arrays. |
| Secret redaction | Passwords, tokens, OTPs, card data, API keys, and raw auth headers must never enter audit JSON. |
| PII minimization | Prefer hashes for IP/user-agent; avoid raw IP, full address, and unnecessary personal data. |
| Cursor safety | Use returned cursor only; do not let clients construct DB ids manually. |
| Retention | Define production retention/archival policy before audit table grows too large. |
| Index hygiene | Keep seller/time/action/request indexes; audit timeline can become high-volume. |
| Internal auth | Use strong `CMS_INTERNAL_AUTH_TOKEN` locally; production should add mTLS/service identity. |

Beginner best practices:

- Audit log me "who, what, where, when, request id" always store hona chahiye.
- `before_json` and `after_json` me sirf changed/important fields rakho.
- Money values minor units me store karo, e.g. INR 500.00 as `50000`.
- Time always UTC rakho.
- Audit records update/delete mat karo; correction ke liye new compensating audit event create karo.
- UI filters page-size bounded rakho, infinite unbounded export mat banao.
- Dispute investigation ke liye `request_id` and `trace_id` preserve karo.

## 13. Missing or Misconfigured Things

Current audit after inspecting `${INPUT_FILE_PATH}` and source implementation:

| Item | Current state | Impact | Suggested fix |
|---|---|---|---|
| Migration versioning | Raw SQL files exist, but no migration runner/version table found | Manual reruns can cause duplicate column/index errors | Add goose/migrate-style workflow or platform migration runner. |
| Transaction coupling | Audit recorder uses `*sql.DB`, not `*sql.Tx` | Some business mutations may commit separately from audit insert | For high-risk mutations, add transaction-aware repository methods. |
| Denied attempt persistence | Authorizer can record mutation decisions, but current read focus mainly shows stored allowed audit rows | Denied permission attempts may be less complete depending on flow | Decide if denied attempts should always persist with reason. |
| Retention/archival automation | No retention job found | Audit table can grow indefinitely | Add retention policy, partitioning, or archive job. |
| Export/reporting controls | No audit export endpoint found | Large manual exports may be ad hoc | Add controlled export with permissions, rate limits, and masking if required. |
| gRPC reflection | Not registered | `grpcurl list` without proto fails | Use proto flags locally or enable reflection only in dev. |
| gRPC health service | Not registered | Standard gRPC health probes unavailable | Add `grpc.health.v1.Health` when deployment hardens. |
| Service Dockerfile/compose | Not found for this service | Beginner setup relies on manual Go run | Add Dockerfile and local compose later if team wants one-command startup. |
| Secrets in local `.env` | Placeholder `change-me` values exist | Unsafe if copied to shared env | Replace with real local-only random tokens; never commit real secrets. |
| Raw IP/user-agent hashing salt | No explicit salt env found | Hashing strategy may be incomplete if added later | Add secret salt env if IP/UA hashing becomes required. |

## 14. References to Previous Dependency Files

Previous dependency files inside `TaskImplementation/${SERVICE_NAME}/` were checked before generating this file.

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `2. Tech Stack Analysis` | Same Go, `net/http`, gRPC, protobuf, MySQL, and logging foundation. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `3. Go Dependency System` | Same Go modules, `go mod download`, `go test`, `go build`, and dependency troubleshooting. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `4. Database Analysis` | MySQL install, Docker command, DB/user creation, connection strings, and full migration flow are unchanged. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `5. Environment Variables` | Full `.env` setup already exists; this file only highlights audit-specific variables. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `7. Ports and Networking` | HTTP `8087`, gRPC `9098`, MySQL `3306`, and port conflict guidance are reused. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `8. Docker and DevOps Setup` | No new Docker service/container is introduced by audit logs. |
| `TaskImplementation/${SERVICE_NAME}/task1_Dependency.md` | `11. Common Errors and Fixes` | Generic Go/MySQL/Docker/internal auth errors are unchanged. |
| `TaskImplementation/${SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | Same `cms_db`, MySQL 8+, migration order, and seller settings/staff foundation. |
| `TaskImplementation/${SERVICE_NAME}/task3_Dependency.md` | `5. Database Analysis` and `13. Security and Best Practices` | Product moderation mutations already describe audit-worthy workflow and product review data. |
| `TaskImplementation/${SERVICE_NAME}/task4_Dependency.md` | `5. Database Setup` and `12. Security & Best Practices` | Coupon create/update/disable flows reuse audit write behavior and same DB/runtime setup. |
| `TaskImplementation/${SERVICE_NAME}/task5_Dependency.md` | `5. Database Setup` and `13. Security and Best Practices` | Campaign create/update/disable flows already generate audit entries and reuse same MySQL setup. |
| `TaskImplementation/${SERVICE_NAME}/task6_Dependency.md` | `7. Environment Variables` | Confirms `.env` is loaded through `os.Getenv` and must be sourced manually. |
| `TaskImplementation/${SERVICE_NAME}/task6_Dependency.md` | `13. Missing or Misconfigured Things` | Missing Docker Compose and migration runner notes are still relevant. |
| `TaskImplementation/${SERVICE_NAME}/task7_Dependency.md` | `2. Tech Stack`, `7. Environment Variables`, and `8. Docker Setup` | Audit gRPC method reuses existing gRPC/proto/env/port setup. |
| `TaskImplementation/${SERVICE_NAME}/task7_Dependency.md` | `11. Common Errors & Fixes` | `grpcurl` reflection, internal token, metadata, and seller auth troubleshooting are reused. |

## 15. Final Checklist

Use this checklist for `${TASK_FILE_NAME}` setup:

- [ ] Previous dependency documentation checked first.
- [ ] No duplicate Go/MySQL/Docker baseline setup copied into this file.
- [ ] Go version is compatible with `${BACKEND_SERVICE_PATH}/go.mod`.
- [ ] MySQL is running.
- [ ] Full CMS migrations are applied in numeric order.
- [ ] `cms_audit_logs` table exists.
- [ ] Migration `007_harden_cms_audit_logs.up.sql` is applied exactly once.
- [ ] Audit columns `actor_roles_json`, `request_id`, `trace_id`, `decision`, `reason`, `ip_hash`, and `user_agent_hash` exist.
- [ ] Audit indexes `idx_cms_audit_request` and `idx_cms_audit_action_created` exist.
- [ ] `${BACKEND_SERVICE_PATH}/.env` exists.
- [ ] `CMS_AUDIT_DEFAULT_RANGE_DAYS` is positive.
- [ ] `CMS_AUDIT_DEFAULT_PAGE_SIZE` is positive and <= `CMS_AUDIT_MAX_PAGE_SIZE`.
- [ ] `CMS_AUDIT_MAX_PAGE_SIZE` is between 1 and 100.
- [ ] DB credentials are configured through `CMS_MYSQL_DSN` or `CMS_DB_*`.
- [ ] `CMS_INTERNAL_AUTH_TOKEN` is not a real shared secret in committed files.
- [ ] `.env` is loaded with `set -a; source .env; set +a` before running the service.
- [ ] `go test ./...` passes from `${BACKEND_SERVICE_PATH}`.
- [ ] Service starts and logs `cms.mysql.connected`, `cms.http.started`, and `cms.grpc.started`.
- [ ] `GET /healthz` returns success.
- [ ] Audit list HTTP route is tested with trusted seller/user/role headers.
- [ ] Audit gRPC route is tested with `grpcurl -proto` if needed.
- [ ] At least one audit row is inserted by seed or audited mutation for local read testing.
- [ ] Audit response only returns rows for the authenticated seller scope.
- [ ] `before_json` and `after_json` do not contain secrets or raw sensitive PII.
- [ ] Logs checked for audit write/list failures.
- [ ] No Redis/Kafka/RabbitMQ setup was added because current code does not use them.
- [ ] Missing production items are understood: migration runner, retention/archival policy, transaction-aware audit writes, gRPC health/reflection decision, Dockerfile/compose.

Final simple summary: `${TASK_FILE_NAME}` does not add a new language runtime, database engine, queue, cache, or Docker container. It reuses the existing Go/MySQL/HTTP/gRPC setup and adds audit-specific requirements: `cms_audit_logs` schema hardening, `CMS_AUDIT_*` guardrails, seller-scoped audit read verification, secure snapshot handling, and careful retention/DevOps planning.
