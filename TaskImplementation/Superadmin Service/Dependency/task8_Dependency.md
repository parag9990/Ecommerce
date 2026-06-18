# Project Dependency & Setup Guide

> Incremental beginner-friendly dependency handbook for `TASK_FILE_NAME`. Common Go, MySQL, Docker, `.env`, migration, and HTTP setup is reused from previous dependency guides. Is file ka focus sirf current task ke new setup points par hai: persistent admin audit logs.

## Document Variables

| Variable | Meaning |
|---|---|
| `SERVICE_NAME` | Service folder value supplied by the task request |
| `TASK_FILE_NAME` | Current implementation task filename supplied by the task request |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | Replace `.md` in `TASK_FILE_NAME` with `_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

## 1. Project Overview

`INPUT_FILE_PATH` defines immutable admin audit logs. Simple meaning: jab admin koi sensitive action karta hai, system ko durable record rakhna chahiye ki kis admin ne kya action kiya, kis resource par kiya, reason kya tha, request/session context kya tha, aur before/after safe summary kya thi.

### Current implementation reality

The current repository has a runnable Go HTTP implementation for audit logs:

```text
backend/services/superadmin-service/
|-- internal/domain/admin_audit.go
|-- internal/usecase/audit_log.go
|-- internal/usecase/audit_recorder.go
|-- internal/repository/mysql_audit_log_repository.go
|-- internal/transport/http/audit_log_handler.go
|-- internal/config/config.go
`-- migrations/006_create_admin_audit_logs.up.sql
```

Current flow:

```text
Admin client / API Gateway
        |
        | trusted admin headers
        v
GET /api/v1/admin/audit-logs
        |
        v
RBAC permission check: audit:logs:read
        |
        v
MySQL admin_audit_logs filtered query
```

Mutation audit write flow:

```text
Admin mutation usecase
        |
        v
AuditRecorder interface
        |
        v
MySQLAuditRecorder when DB is configured
        |
        v
INSERT into admin_audit_logs
```

Important current-state note:

- If `SUPERADMIN_DATABASE_DSN` is empty, the server can start in local mode, but audit writes only go to `LoggingAuditRecorder` and the audit list API returns storage-unavailable behavior.
- For `TASK_FILE_NAME`, functional verification requires MySQL, migrations, and `SUPERADMIN_DATABASE_DSN`.
- `INPUT_FILE_PATH` mentions an audit migration named `005_create_admin_audit_logs`. The actual repository uses `006_create_admin_audit_logs` because migration `005` already belongs to platform settings.
- Current code validates `request_id` and `reason` as required audit fields before DB insert.

## 2. Tech Stack

### Task-specific technologies

| Technology / component | Required? | Why used | Simple Hinglish explanation |
|---|---:|---|---|
| Go `net/http` audit route | Yes | Exposes `GET /api/v1/admin/audit-logs` | External web framework nahi hai; Go standard HTTP server se API route register hota hai. |
| MySQL `admin_audit_logs` table | Yes for functional audit | Stores durable admin action records | MySQL relational DB hai; audit rows structured and filterable form me save hote hain. |
| MySQL JSON columns | Yes | Stores safe `before_json`, `after_json`, and `metadata_json` snapshots | JSON column flexible summary store karta hai, but secrets sanitize hote hain. |
| `github.com/go-sql-driver/mysql` | Yes, reused | Lets Go connect to MySQL | Go app ko MySQL se baat karne ke liye driver chahiye. Ye already `go.mod` me present hai. |
| `MySQLAuditRecorder` | Yes for task success | Replaces temporary log-only recorder when DB exists | Mutation ke baad audit row DB me insert karne wala recorder. |
| `LoggingAuditRecorder` | Fallback only | Keeps local/no-DB mode from crashing | Ye sirf log me event print karta hai. Compliance-grade durable audit nahi hai. |
| RBAC permission `audit:logs:read` | Yes | Protects audit list API | Audit logs sensitive hain; sirf permitted admin read kar sakta hai. |
| Permission `audit:logs:export` | Future-facing | Critical export permission already exists | Export API current code me nahi hai, but permission future ke liye seeded hai. |

### Already documented shared technologies

Go installation, Go modules, MySQL installation, MySQL Docker setup, `.env` loading, HTTP header model, migrations, ports, and generic troubleshooting are already explained.

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`2. Tech Stack`, `3. Required Software`, `4. Dependency Management`, `5. Database Setup`, `7. Environment Variables`, `8. Ports & Networking`, `9. Docker Setup`, and `12. Common Errors & Fixes`

Also refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/Go_Modules.md`
`TaskImplementation/{SERVICE_NAME}/Dependency/MySQL.md`
`TaskImplementation/{SERVICE_NAME}/Dependency/Migrations.md`
`TaskImplementation/{SERVICE_NAME}/Dependency/Environment.md`
`TaskImplementation/{SERVICE_NAME}/Dependency/HTTP_API.md`

## 3. Required Software

No new local software is introduced by `TASK_FILE_NAME`.

| Software | Required? | Why | Status |
|---|---:|---|---|
| Go `1.26.3` compatible toolchain | Yes | Build, test, run the service | Reused |
| MySQL `8.x` | Yes for persistent audit logs | Stores RBAC and `admin_audit_logs` | Reused, now mandatory for task verification |
| MySQL client CLI | Recommended | Run migrations and verify rows | Reused |
| curl or API client | Recommended | Verify HTTP audit list API | Reused |
| Docker | Optional | Easiest local MySQL container setup | Reused |

For installation commands, do not duplicate here.

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/MySQL.md`

Sections:
`5. Installation Steps`, `6. Docker Setup, If Possible`, and `10. Verify Running Commands`

## 4. Dependency Management

### New dependency result

No new Go package is required for `TASK_FILE_NAME`.

Current module file:

```text
backend/services/superadmin-service/go.mod
```

Current direct dependency:

```text
github.com/go-sql-driver/mysql v1.10.0
```

Beginner note: `go.mod` dependency already present hai, isliye `go get` run karne ki zarurat nahi hai. Agar fresh clone hai, normal module download enough hai.

```bash
cd backend/services/superadmin-service
go mod download
go mod verify
```

### Task-specific verification

Run focused tests for the new audit domain/usecase/HTTP behavior:

```bash
cd backend/services/superadmin-service
go test ./internal/domain ./internal/usecase ./internal/transport/http
```

Or run all service tests:

```bash
go test ./...
```

If Go module setup fails, reuse the existing guide instead of debugging from scratch.

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/Go_Modules.md`

Sections:
`4. Dependency Management` and `11. Common Errors And Fixes`

## 5. Database Setup

### Reused MySQL setup

MySQL installation, Docker run command, database creation, local app user, DSN format, and generic migration flow are already documented.

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/MySQL.md`

Sections:
`5. Installation Steps`, `6. Docker Setup, If Possible`, `8. Required Environment Variables`, and `11. Common Errors And Fixes`

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/Migrations.md`

Sections:
`7. Local Setup Without Docker`, `9. Health Checks / Verification`, and `11. Common Errors And Fixes`

### Task-specific database change

`TASK_FILE_NAME` requires this actual migration:

```text
backend/services/superadmin-service/migrations/006_create_admin_audit_logs.up.sql
```

Rollback file:

```text
backend/services/superadmin-service/migrations/006_create_admin_audit_logs.down.sql
```

### What the table stores

| Column / concept | Purpose |
|---|---|
| `id` | Internal increasing sequence used for newest-first pagination and cursor generation |
| `audit_id` | Public-safe random audit id, unique per row |
| `actor_admin_id` | Admin who performed the action; foreign key to `admin_users.admin_id` |
| `action` | Machine-readable action like `seller.status.suspended` |
| `resource_type` | Resource category such as `user`, `seller`, `refund`, `platform_setting` |
| `resource_id` | Target resource id |
| `request_id` | Request trace id; required by current code |
| `session_id` | Admin session id if available |
| `ip_hash` | PII-safer hashed IP value, not raw IP |
| `reason` | Human-readable reason; required by current code |
| `before_json` | Safe summary before mutation |
| `after_json` | Safe summary after mutation |
| `metadata_json` | Extra safe metadata |
| `created_at` | Insert timestamp |

### Required migration order

Run the complete migration sequence in filename order for a clean local DB:

```text
001_create_superadmin_rbac.up.sql
002_seed_admin_rbac.up.sql
003_create_admin_review_tasks.up.sql
004_add_open_review_task_uniqueness.up.sql
005_create_platform_settings.up.sql
006_create_admin_audit_logs.up.sql
```

Why order matters:

| Migration | Why it matters for audit logs |
|---|---|
| `001` | Creates `admin_users`, which `admin_audit_logs.actor_admin_id` references |
| `002` | Seeds `audit:logs:read` and `audit:logs:export` permissions |
| `003` and `004` | Required by earlier admin review workflows that may write audit rows |
| `005` | Required by platform settings mutation flow that writes audit rows |
| `006` | Creates `admin_audit_logs` itself |

### Verify migration `006`

```bash
mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
"SHOW TABLES LIKE 'admin_audit_logs';"
```

Check important columns and indexes:

```bash
mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
"SHOW CREATE TABLE admin_audit_logs\\G"
```

### Credentials placement

No new database credential variable was introduced. Use the existing DSN:

```text
SUPERADMIN_DATABASE_DSN
```

Where to place it:

```text
backend/services/superadmin-service/.env
```

Important: Code reads OS environment variables with `os.Getenv`; it does not auto-load `.env`. Source the file before running.

## 6. Redis / Queue / External Services

### Current direct external services

| Service | Required for audit list API? | Required for mutation-generated audit rows? | Status |
|---|---:|---:|---|
| MySQL | Yes | Yes | Reused, mandatory for task verification |
| API Gateway / trusted admin context | Production yes | Production yes | Reused security boundary |
| User Service | No | Yes for user/seller mutation audit rows | Reused from earlier tasks |
| Order Service | No | Yes for order review audit rows | Reused from earlier tasks |
| Payment Service | No | Yes for refund audit rows | Reused from earlier tasks |

### Services not introduced by this task

| Service | Current status | Setup action |
|---|---|---|
| Redis | No Redis client/env is used for audit logs | Do not start Redis for `TASK_FILE_NAME` |
| Kafka / RabbitMQ / NATS | No broker producer is implemented for audit logs | No queue/broker setup needed |
| OpenSearch / Elasticsearch | Not used by current audit list implementation | No search cluster setup needed |
| Object storage / S3 / MinIO | Archive/export storage is future scope | No bucket/container setup needed |
| SIEM integration | Future production hardening only | No local setup needed |

Beginner note: Audit logs abhi directly MySQL me store hote hain. Redis fast cache ke liye hota hai, Kafka/RabbitMQ events ke liye hote hain, but current code me inka audit dependency nahi hai.

## 7. Environment Variables

### Reused environment setup

Shared `.env` location, source/loading behavior, duration format, and generic environment mistakes are already documented.

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/Environment.md`

Sections:
`7. Local Setup Without Docker`, `8. Required Environment Variables`, and `11. Common Errors And Fixes`

### New or changed variables

No brand-new environment variable was added for `TASK_FILE_NAME`.

However, for this task's functional verification, these existing variables become important:

| Variable | Required? | Example | Why it matters | Security note |
|---|---:|---|---|---|
| `SUPERADMIN_DATABASE_DSN` | Yes for persistent audit | `superadmin:devpassword@tcp(127.0.0.1:3306)/superadmin_db?parseTime=true&charset=utf8mb4&loc=UTC` | Without this, audit records are not persisted in MySQL | Contains DB password; do not commit or paste publicly |
| `SUPERADMIN_REQUIRE_DATABASE` | Recommended `true` | `true` | Forces startup failure if DB is missing even in local mode | Prevents accidentally testing log-only audit |
| `HTTP_ADDR` | Reused | `127.0.0.1:8088` | Local audit API listen address | Bind to localhost for local-only direct runs |

Task-specific `.env` snippet only:

```bash
SUPERADMIN_DATABASE_DSN='superadmin:devpassword@tcp(127.0.0.1:3306)/superadmin_db?parseTime=true&charset=utf8mb4&loc=UTC'
SUPERADMIN_REQUIRE_DATABASE=true
HTTP_ADDR=127.0.0.1:8088
```

Do not duplicate the full `.env` here. For complete variable list, refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`7. Environment Variables`

### Incoming request headers for audit list API

These are not `.env` variables; they are trusted request headers normally injected by API Gateway.

| Header | Required? | Why |
|---|---:|---|
| `X-Admin-Id` | Yes | DB RBAC checks use this admin id |
| `X-User-Id` or `X-Subject-Id` | Yes | Actor context validation |
| `X-Admin-Roles` or `X-Roles` | Yes | Actor context, though DB permission remains final |
| `X-Session-Id` | Yes | Admin route actor validation |
| `X-Request-Id` | Recommended for reads, required for mutation audit inserts | Trace/debug context |
| `X-IP-Hash` | Optional | PII-safer source IP trace |

For full header explanation, refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/HTTP_API.md`

## 8. Docker Setup

No new Docker container is required for `TASK_FILE_NAME`.

Reuse existing MySQL Docker setup:

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/MySQL.md`

Section:
`6. Docker Setup, If Possible`

Current Docker/DevOps status:

| Item | Status | Task-specific note |
|---|---|---|
| MySQL container | Reused | Must have migration `006` applied |
| Service Dockerfile | Missing | No new Dockerfile added by this task |
| docker-compose service | Missing | No compose update exists for audit logs |
| Redis container | Not required | No Redis code/env for audit logs |
| Message broker container | Not required | No Kafka/RabbitMQ/NATS code for audit logs |
| OpenSearch/SIEM container | Not required | Future audit search/security integration only |

If a future Compose stack is added, make sure service env includes:

```yaml
environment:
  SUPERADMIN_DATABASE_DSN: superadmin:devpassword@tcp(superadmin-mysql:3306)/superadmin_db?parseTime=true&charset=utf8mb4&loc=UTC
  SUPERADMIN_REQUIRE_DATABASE: "true"
```

Container note: inside Docker networking, `127.0.0.1` points to the same container, not the MySQL container. Use the Compose service DNS name such as `superadmin-mysql`.

## 9. Ports & Networking

No new port is introduced by `TASK_FILE_NAME`.

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Backend HTTP API | `8088` | `GET /api/v1/admin/audit-logs`, health, readiness | Reused |
| MySQL | `3306` | RBAC and `admin_audit_logs` storage | Reused |
| User Service | `8081` in current `.env` | Generates user/seller audit rows through mutation flows | Reused optional downstream |
| Order Service | `8084` in current `.env` | Generates order review audit rows through mutation flows | Reused optional downstream |
| Payment Service | `8085` in current `.env` | Generates refund review audit rows through mutation flows | Reused optional downstream |
| Redis | `6379` | Not used by current audit implementation | Not required |
| Kafka / RabbitMQ / NATS | N/A | No current audit broker publisher | Not required |

Port conflict and firewall/network basics are already explained.

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`8. Ports & Networking`

## 10. Local Development Setup

### Step 1: Read reused setup first

Read these first to avoid duplicate setup confusion:

```text
TaskImplementation/{SERVICE_NAME}/Dependency/Go_Modules.md
TaskImplementation/{SERVICE_NAME}/Dependency/MySQL.md
TaskImplementation/{SERVICE_NAME}/Dependency/Migrations.md
TaskImplementation/{SERVICE_NAME}/Dependency/Environment.md
TaskImplementation/{SERVICE_NAME}/Dependency/HTTP_API.md
```

### Step 2: Enter the backend service directory

```bash
cd backend/services/superadmin-service
```

### Step 3: Install dependencies

No new dependency install is required.

```bash
go mod download
go mod verify
```

### Step 4: Start MySQL and apply migrations

Use the reused MySQL setup and apply all migrations in order. Confirm audit table exists:

```bash
mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
"SHOW TABLES LIKE 'admin_audit_logs';"
```

### Step 5: Provision an active local admin

Audit rows have a foreign key to `admin_users.admin_id`, and the list API uses DB RBAC. Use the existing admin provisioning guidance.

For a local superadmin test actor, the admin row should exist:

```sql
INSERT INTO admin_users
  (admin_id, user_id, role, status, mfa_required, created_by)
VALUES
  ('admin_local', 'user_local', 'superadmin', 'active', TRUE, 'local_setup')
ON DUPLICATE KEY UPDATE
  role = VALUES(role),
  status = 'active',
  mfa_required = TRUE;
```

### Step 6: Configure task-important env values

Use the full previous `.env` guide, then confirm these values:

```bash
SUPERADMIN_DATABASE_DSN='superadmin:devpassword@tcp(127.0.0.1:3306)/superadmin_db?parseTime=true&charset=utf8mb4&loc=UTC'
SUPERADMIN_REQUIRE_DATABASE=true
HTTP_ADDR=127.0.0.1:8088
```

Load environment before running:

```bash
set -a
source .env
set +a
```

### Step 7: Run tests

```bash
go test ./internal/domain ./internal/usecase ./internal/transport/http
```

For full service confidence:

```bash
go test ./...
```

### Step 8: Start backend service

```bash
go run ./cmd/server
```

Expected log includes:

```text
superadmin service listening
```

### Step 9: Verify health and readiness

In a second terminal:

```bash
curl -i http://127.0.0.1:8088/healthz
curl -i http://127.0.0.1:8088/readyz
```

Expected:

```json
{"status":"ok"}
```

and:

```json
{"status":"ready"}
```

### Step 10: Create one local audit row for API verification

Best real verification is to perform a mutation endpoint that records audit automatically. For a beginner local smoke test, you can insert one local-only row after `admin_local` exists:

```sql
INSERT INTO admin_audit_logs (
  audit_id,
  actor_admin_id,
  action,
  resource_type,
  resource_id,
  request_id,
  session_id,
  ip_hash,
  reason,
  before_json,
  after_json,
  metadata_json
) VALUES (
  'aud_local_audit_001',
  'admin_local',
  'platform_setting.update',
  'platform_setting',
  'maintenance_mode',
  'req_audit_local_001',
  'sess_local',
  'sha256:local',
  'local audit log API verification',
  JSON_OBJECT('version', '1'),
  JSON_OBJECT('version', '2'),
  JSON_OBJECT('source', 'manual_local_verification')
);
```

If duplicate audit id error appears, change the `audit_id` suffix. Production me audit rows manually insert mat karo; API/usecase recorder se hi write hona chahiye.

### Step 11: Verify audit list API

```bash
curl -s "http://127.0.0.1:8088/api/v1/admin/audit-logs?page_size=10&resource_type=platform_setting" \
  -H "X-Admin-Id: admin_local" \
  -H "X-User-Id: user_local" \
  -H "X-Admin-Roles: superadmin" \
  -H "X-Session-Id: sess_local" \
  -H "X-Request-Id: req_audit_read_001"
```

Expected response shape:

```json
{
  "logs": [
    {
      "audit_id": "aud_local_audit_001",
      "actor_admin_id": "admin_local",
      "action": "platform_setting.update",
      "resource_type": "platform_setting",
      "resource_id": "maintenance_mode"
    }
  ]
}
```

Also verify directly in MySQL:

```bash
mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
"SELECT audit_id, actor_admin_id, action, resource_type, resource_id, created_at FROM admin_audit_logs ORDER BY id DESC LIMIT 5;"
```

## 11. Running the Project

Use the same run flow from previous dependency docs. Current task adds these success criteria:

| Check | Expected result |
|---|---|
| `SUPERADMIN_DATABASE_DSN` | Non-empty and points to migrated DB |
| `SUPERADMIN_REQUIRE_DATABASE` | `true` for local audit verification |
| `/healthz` | Returns `{"status":"ok"}` |
| `/readyz` | Returns `{"status":"ready"}` when DB is reachable |
| `admin_audit_logs` | Table exists after migration `006` |
| `admin_users` | Test admin exists and is active |
| `admin_role_permissions` | Test admin role has `audit:logs:read` |
| `GET /api/v1/admin/audit-logs` | Returns filtered audit rows |
| Mutation endpoints | Return success only after audit recorder succeeds |

Supported query params:

| Query param | Example | Purpose |
|---|---|---|
| `actor_id` | `admin_local` | Filter by admin actor |
| `action` | `seller.status.suspended` | Filter by action key |
| `resource_type` | `seller` | Filter by resource type |
| `resource_id` | `seller_1` | Filter by resource id |
| `request_id` | `req_123` | Find rows for one request |
| `from` | `2026-05-01T00:00:00Z` | Start time, RFC3339 |
| `to` | `2026-05-31T23:59:59Z` | End time, RFC3339 |
| `page` | `1` | Page number when cursor is absent |
| `page_size` | `50` | Defaults to `50`, max `100` |
| `cursor` | Encoded sequence id | Next-page cursor |

## 12. Common Errors & Fixes

Generic Go, Docker, MySQL connection, `.env`, readiness, and admin header errors are already documented in previous guides. Task-specific errors are below.

| Error / symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `admin audit log storage is not configured` | `SUPERADMIN_DATABASE_DSN` empty, so unavailable audit repository is wired | Set `SUPERADMIN_DATABASE_DSN`, set `SUPERADMIN_REQUIRE_DATABASE=true`, restart | Treat DB as mandatory for audit verification |
| `Table '...admin_audit_logs' doesn't exist` | Migration `006` not applied | Apply migrations in order through `006_create_admin_audit_logs.up.sql` | Use migration runner/history table in future |
| Foreign-key error on audit insert | `actor_admin_id` is not present in `admin_users` | Provision active admin before mutation/local seed | Bootstrap admins before enabling admin APIs |
| `FORBIDDEN` for audit list | Admin is inactive or role lacks `audit:logs:read` in DB | Check `admin_users` and `admin_role_permissions` | Seed migration `002` and use active allowed admin |
| `admin_id is required for admin routes` | Missing `X-Admin-Id` header | Send trusted admin headers through gateway/local curl | Gateway should inject validated admin context |
| `session_id is required for admin routes` | Missing `X-Session-Id` header | Add session header | Admin requests should always carry session context |
| `request_id is required` during mutation audit | Mutation call missing `X-Request-Id` | Send unique request ID | Gateway should generate request IDs |
| `reason is required` during mutation audit | Mutation reason missing or blank | Send meaningful reason | UI should require reason before risky actions |
| `cursor is invalid` | Cursor is malformed or zero | Use `next_cursor` returned by API, or remove cursor | Do not hand-edit cursors |
| `from must be an RFC3339 timestamp` | Time filter like `yesterday` was sent | Use `2026-05-01T00:00:00Z` format | Date picker/API client should send RFC3339 UTC |
| Duplicate `audit_id` in local SQL seed | Same smoke-test row inserted twice | Change local `audit_id` suffix | For real writes, app generates random ids |
| Mutation succeeds downstream but audit insert fails | Cross-service mutation cannot share one DB transaction | Investigate immediately; add retry/outbox in future hardening | Monitor audit insert failures and request IDs |

## 13. Security & Best Practices

### Task-specific security audit

| Severity | Finding | Suggested fix |
|---|---|---|
| Critical for public deployment | Service trusts admin headers; it does not verify JWT/session directly | Keep service private behind trusted API Gateway, strip user-supplied admin headers, use service-to-service auth |
| Critical | Audit logs are compliance/security records, but MySQL is technically mutable for privileged users | Production DB user should have `SELECT` and `INSERT` on `admin_audit_logs`, not `UPDATE` or `DELETE`; monitor privileged access |
| High | Local setup docs use broad DB grants for beginner simplicity | Use least-privilege production grants and separate migration/admin users |
| High | `SUPERADMIN_DATABASE_DSN` can be empty in local mode | Use `SUPERADMIN_REQUIRE_DATABASE=true` whenever testing audit functionality |
| High | Platform setting update currently writes business row then audit row without one DB transaction | Add transactional repository/usecase flow where local DB mutation and audit insert must commit together |
| High | Downstream mutation plus audit write cannot be atomic across services | Add idempotency, reconciliation, retry queue, or outbox pattern for production hardening |
| High | `before_json`, `after_json`, and `metadata_json` can contain sensitive summaries if callers pass unsafe keys | Keep using `SafeAuditSnapshot`; review new audit keys before adding them |
| Medium | No export API exists, but export permission exists | When export is implemented, require reason, request id, strict permission, and audit the export action |
| Medium | No archive/object-storage/SIEM integration exists | Add retention/archive/security forwarding when compliance requires it |
| Medium | No service Dockerfile or Compose stack exists | Add reproducible container setup before team-wide deployment |

### Task-specific best practices

- Always write audit through `AuditRecorder`, not direct SQL, except disposable local smoke tests.
- Keep audit rows append-only. Correction chahiye to new audit row create karo, old row edit mat karo.
- Store safe summaries only. Password, OTP, token, cookie, card number, CVV, API key, private key, and raw secret never store karo.
- Prefer hashed IP through `X-IP-Hash`; raw IP avoid karo unless policy explicitly allows it.
- Keep `request_id`, `session_id`, and `reason` present for every mutation. Investigation me ye fields sabse useful hote hain.
- Use `page_size` and filters while reading audit logs. Unbounded audit queries expensive ho sakti hain.
- Monitor audit insert failures as high severity.
- Keep audit read access limited. `readonly_admin` can read; export should stay stricter.

## 14. Missing or Misconfigured Things

Task-specific current gaps:

- [ ] `INPUT_FILE_PATH` mentions audit migration `005`, but actual repository file is `006_create_admin_audit_logs.up.sql`.
- [ ] No migration runner or migration history table, so migration `006` must be applied manually/in order.
- [ ] No service Dockerfile or full local Compose stack.
- [ ] No safe committed `.env.example` for task-important DB settings.
- [ ] Local example grants are broad; production append-only immutability needs restricted DB permissions.
- [ ] Platform setting update and audit insert are not currently wrapped in one DB transaction.
- [ ] Downstream mutation audit failure needs retry/reconciliation hardening.
- [ ] No audit export endpoint yet, only list API and export permission.
- [ ] No audit archive, retention job, object storage, OpenSearch, or SIEM integration.
- [ ] No direct JWT/JWKS/Auth Service verification in this service.
- [ ] Readiness only checks DB ping; it does not verify the audit table/migration is present.

These are not blockers for local `GET /api/v1/admin/audit-logs` verification, but they matter before production rollout.

## 15. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `task1_Dependency.md` | Tech stack, required software, MySQL, `.env`, ports, Docker, run flow, generic errors | Same Go HTTP service and base MySQL setup |
| `task2_Dependency.md` | MySQL decision and structured admin data ownership | Audit logs use same MySQL-backed admin data model |
| `task3_Dependency.md` | RBAC permissions and admin header behavior | Audit list depends on `audit:logs:read` and admin context |
| `task4_Dependency.md` | User/seller mutation audit direction and User Service boundary | User/seller flows now persist audit through MySQL recorder |
| `task5_Dependency.md` | Order/payment high-risk controls and downstream boundary | Refund and manual review flows write audit rows |
| `task6_Dependency.md` | Session visibility access audit and no Redis/queue direct setup | Session analytics access uses audit recorder; no new broker required |
| `task7_Dependency.md` | Platform setting mutation audit and migration `005` context | Shows why audit migration is actually `006` now |
| `Dependency/Go_Modules.md` | Go module setup and common Go dependency errors | No new Go package added |
| `Dependency/MySQL.md` | MySQL install, Docker, DSN, credentials, verification | Same database and credentials are reused |
| `Dependency/Migrations.md` | Ordered migration flow and migration troubleshooting | Current task adds migration `006` details only |
| `Dependency/Environment.md` | `.env` loading and common env mistakes | Current task only changes importance of existing DB variables |
| `Dependency/HTTP_API.md` | HTTP server and admin request headers | Audit list API uses same trusted header contract |

Full setup references:

```text
TaskImplementation/{SERVICE_NAME}/Dependency/Go_Modules.md
TaskImplementation/{SERVICE_NAME}/Dependency/MySQL.md
TaskImplementation/{SERVICE_NAME}/Dependency/Migrations.md
TaskImplementation/{SERVICE_NAME}/Dependency/Environment.md
TaskImplementation/{SERVICE_NAME}/Dependency/HTTP_API.md
```

## 16. Final Checklist

- [ ] Previous dependency documentation checked before using this file
- [ ] No duplicate Go/MySQL/Docker installation steps copied into this file
- [ ] MySQL is running
- [ ] `SUPERADMIN_DATABASE_DSN` is set and loaded into shell
- [ ] `SUPERADMIN_REQUIRE_DATABASE=true` is set for audit verification
- [ ] Migrations `001` through `006` applied in order
- [ ] `admin_audit_logs` table exists
- [ ] Active local admin exists in `admin_users`
- [ ] Admin role has `audit:logs:read`
- [ ] Focused audit tests pass
- [ ] Backend service starts on expected HTTP port
- [ ] `/healthz` and `/readyz` return expected responses
- [ ] One audit row exists through real mutation or local smoke-test seed
- [ ] `GET /api/v1/admin/audit-logs` returns audit data
- [ ] Logs checked for audit insert/list errors
- [ ] No Redis/Kafka/RabbitMQ setup added unnecessarily
- [ ] No secrets committed to Git
- [ ] Task-specific missing/misconfigured items reviewed before production
