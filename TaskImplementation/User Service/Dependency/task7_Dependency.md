# Project Dependency & Setup Guide

## Variables Used In This Guide

```text
SERVICE_NAME = User Service
TASK_FILE_NAME = task7.md
INPUT_FILE_PATH = TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}
OUTPUT_FILE_NAME = task7_Dependency.md
OUTPUT_FILE_PATH = TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}
SERVICE_CODE_DIR = backend/services/user-service
MIGRATION_DIR = {SERVICE_CODE_DIR}/migrations
```

Beginner note: Is guide me `SERVICE_NAME` and `TASK_FILE_NAME` ko reusable variables ki tarah treat kiya gaya hai. Actual backend code folder `SERVICE_CODE_DIR` hai.

## 1. Project Overview

This document is the dependency and setup companion for `INPUT_FILE_PATH`.

Current task ka goal record-level audit metadata enable karna hai:

- `created_by`, `updated_by`, `status_changed_by`, `status_changed_at`
- `deleted_by`, `deleted_at`
- `status_reason`
- server-side audit actor resolution through gRPC metadata
- MySQL migration `002_add_user_audit_fields`

Simple Hinglish:

Audit fields ka matlab hai ki database record me "kisne action kiya", "kab kiya", aur "record ka current lifecycle status kya hai" clearly store ho.

Task-specific implementation files checked:

| File | Why checked |
|---|---|
| `INPUT_FILE_PATH` | Current task boundary and audit-field design |
| `{MIGRATION_DIR}/002_add_user_audit_fields.up.sql` | New audit columns, indexes, and backfill |
| `{MIGRATION_DIR}/002_add_user_audit_fields.down.sql` | Audit rollback behavior |
| `{SERVICE_CODE_DIR}/internal/audit/actor.go` | Actor types and stored audit id format |
| `{SERVICE_CODE_DIR}/internal/audit/context.go` | Actor storage in `context.Context` |
| `{SERVICE_CODE_DIR}/internal/domain/audit.go` | Domain audit structs and validation |
| `{SERVICE_CODE_DIR}/internal/transport/grpc/metadata.go` | Metadata keys used for actor resolution |
| `{SERVICE_CODE_DIR}/internal/transport/grpc/interceptors.go` | gRPC audit actor interceptor |
| `{SERVICE_CODE_DIR}/internal/repository/*.go` | SQL queries that read/write audit columns |
| `{SERVICE_CODE_DIR}/internal/config/config.go` | Runtime env and current repo caveat around events |

Important boundary:

- Task 7 does not add Redis.
- Task 7 does not require RabbitMQ or Kafka.
- Task 7 does not create a new Dockerfile or checked-in compose file.
- Task 7 does not require a new `.env` secret.
- Later event/outbox code exists in the current source tree, but event publishing is outside `TASK_FILE_NAME`.

## 2. Tech Stack

Most base technology setup is already documented. Reuse it first:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`1. Project Tech Stack Analysis`
`2. Language-Specific Dependency System: Go`
`3. Database Analysis`
`5. External Services Analysis`
```

Task 7-specific stack impact:

| Technology / Library | Required? | What it is | Current task usage |
|---|---:|---|---|
| Go 1.24+ | Yes | Go compiled backend language hai. | Audit structs, actor resolver, repositories, and tests Go me hain. |
| Go Modules | Yes | Go dependency manager. | No new Task 7 module install required. |
| MySQL 8+ | Yes | Relational database, tables and indexes ke saath. | Existing tables me audit columns add hote hain. |
| `database/sql` | Yes | Go standard DB package. | Audit columns insert/update/select queries me use hota hai. |
| `github.com/go-sql-driver/mysql` | Yes, reused | MySQL driver for Go. | Existing DSN se new audit columns read/write honge. |
| gRPC | Yes, reused | Internal service API framework. | Mutating calls metadata se audit actor resolve karte hain. |
| Protocol Buffers | Yes, reused | Typed gRPC contract. | `AuditInfo` already generated code me present hai. |
| `google.golang.org/grpc/metadata` | Yes, reused | gRPC metadata helper. | `x-actor-id`, `x-actor-type`, `x-user-id`, `x-service-name` read karta hai. |
| `google.golang.org/protobuf/types/known/timestamppb` | Yes, reused | Proto timestamp converter. | Audit timestamps responses me map hote hain. |
| `go-sqlmock` | Test only, reused | SQL unit test mock. | Repository audit SQL expectations test karne ke liye. |
| RabbitMQ / Kafka clients | Not required for Task 7 | Queue/event libraries. | Current module me later outbox code ke liye present hain; Task 7-only run me worker disabled rakho. |

No new framework installation is introduced by `TASK_FILE_NAME`.

## 3. Required Software

Base software setup is reused:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`3. Database Analysis`
`5. External Services Analysis`
`8. Project Run Instructions`
```

Task 7-specific requirement:

| Software | Required? | Why |
|---|---:|---|
| MySQL client or migration runner | Yes | `002_add_user_audit_fields.up.sql` apply karne ke liye. |
| MySQL user with `ALTER`, `UPDATE`, and `INDEX` privileges | Yes for migration | New columns/indexes add and old rows backfill hote hain. |
| grpcurl | Optional but useful | Metadata ke saath gRPC smoke test karne ke liye. |
| Docker | Optional | Sirf local MySQL container ke liye, same as previous setup. |

Runtime application user ko usually `ALTER` privilege nahi dena chahiye. Migration admin user se run karo, app runtime user ko normal `SELECT`, `INSERT`, `UPDATE`, `DELETE` enough hona chahiye.

## 4. Dependency Management

Go dependency basics already explained hain:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`2. Language-Specific Dependency System: Go`
```

Current task does not require `go get`.

Use these only for a fresh clone or module cache setup:

```bash
SERVICE_CODE_DIR=backend/services/user-service
cd "$SERVICE_CODE_DIR"
go mod download
```

Run focused tests:

```bash
SERVICE_CODE_DIR=backend/services/user-service
cd "$SERVICE_CODE_DIR"
go test ./internal/audit ./internal/domain ./internal/usecase ./internal/repository ./internal/transport/grpc
```

Use `go mod tidy` only if dependency files changed or Go reports module errors:

```bash
SERVICE_CODE_DIR=backend/services/user-service
cd "$SERVICE_CODE_DIR"
go mod tidy
```

Beginner note:

- `go mod download` dependencies download karta hai.
- `go mod tidy` unused dependencies remove bhi kar sakta hai.
- Random `go get` run mat karo, kyunki Task 7 ke audit code ke liye new package install required nahi hai.

## 5. Database Setup

### Reused MySQL Setup

MySQL installation, Docker command, Docker Compose example, DSN format, and base credential placement already documented hain:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`3. Database Analysis`
`7. Docker and DevOps Setup`
`8. Project Run Instructions`
```

Base schema setup already documented hai:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`

Section:
`5. Database Setup`
```

### Task 7 New Migration

Task 7 ka new migration:

```text
{MIGRATION_DIR}/002_add_user_audit_fields.up.sql
{MIGRATION_DIR}/002_add_user_audit_fields.down.sql
```

Prerequisite:

```text
{MIGRATION_DIR}/001_create_user_tables.up.sql
```

Migration order:

| Order | File | Required for Task 7? | Purpose |
|---:|---|---:|---|
| 1 | `001_create_user_tables.up.sql` | Yes | Base user/profile/address/KYC tables |
| 2 | `002_add_user_audit_fields.up.sql` | Yes | Audit columns, indexes, and backfill |
| 3 | `003_create_user_outbox_events.up.sql` | No for Task 7 | Later event/outbox table |

Task 7-only local rule:

- Apply `001` and `002`.
- Set `USER_EVENTS_ENABLED=false` if you do not apply `003`.
- Keep `OUTBOX_WORKER_ENABLED=false` unless you are working on later event/outbox setup.

### New Columns And Indexes

| Table | New audit setup | Why needed |
|---|---|---|
| `users` | `created_by`, `updated_by`, `status_changed_by`, `status_changed_at`, `deleted_by`, `deleted_at`, `idx_users_status_updated`, `idx_users_deleted_at` | User lifecycle and soft-delete traceability |
| `user_addresses` | `status`, `created_by`, `updated_by`, `deleted_by`, `idx_user_addresses_user_status` | Address update/delete traceability and active/deleted filtering |
| `seller_profiles` | `created_by`, `updated_by`, `status_changed_by`, `status_changed_at`, `status_reason`, `idx_seller_profiles_status_updated` | Seller approval/suspension/rejection audit |
| `seller_kyc_documents` | `created_by`, `updated_by`, `updated_at`, `status_changed_by`, `status_changed_at`, `idx_kyc_documents_status_updated` | KYC review audit and timestamp consistency |

Backfill behavior:

| Backfill value | Meaning |
|---|---|
| `system:backfill` | Historical rows existed before audit fields. Migration filled missing actor fields. |

Simple Hinglish:

Purane rows me actor data available nahi tha. Isliye migration `system:backfill` set karta hai, taaki reports me blank audit data na dikhe.

### Apply Task 7 Migration

From repository root:

```bash
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_ADMIN_USER=root
SERVICE_CODE_DIR=backend/services/user-service

mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_ADMIN_USER" -p \
  < "$SERVICE_CODE_DIR/migrations/002_add_user_audit_fields.up.sql"
```

If this is a fresh local DB, apply base migration first:

```bash
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_ADMIN_USER=root
SERVICE_CODE_DIR=backend/services/user-service

mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_ADMIN_USER" -p \
  < "$SERVICE_CODE_DIR/migrations/001_create_user_tables.up.sql"

mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_ADMIN_USER" -p \
  < "$SERVICE_CODE_DIR/migrations/002_add_user_audit_fields.up.sql"
```

### Verify Audit Columns

```bash
mysql -h 127.0.0.1 -P 3306 -u ecommerce_user -p user_db
```

Then run:

```sql
SELECT table_name, column_name
FROM information_schema.columns
WHERE table_schema = 'user_db'
  AND table_name IN ('users', 'user_addresses', 'seller_profiles', 'seller_kyc_documents')
  AND column_name IN (
    'created_by',
    'updated_by',
    'status_changed_by',
    'status_changed_at',
    'deleted_by',
    'deleted_at',
    'status_reason',
    'status'
  )
ORDER BY table_name, ordinal_position;
```

Check indexes:

```sql
SHOW INDEX FROM users;
SHOW INDEX FROM user_addresses;
SHOW INDEX FROM seller_profiles;
SHOW INDEX FROM seller_kyc_documents;
```

Check backfilled rows:

```sql
SELECT user_id, status, created_by, updated_by, status_changed_by, status_changed_at
FROM users
LIMIT 10;
```

Expected:

- New/historical rows should not look confusing.
- Old rows may show `system:backfill`.
- New writes should show the actual service/user/admin actor.

### Rollback Task 7 Migration

Rollback file:

```text
{MIGRATION_DIR}/002_add_user_audit_fields.down.sql
```

Command:

```bash
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_ADMIN_USER=root
SERVICE_CODE_DIR=backend/services/user-service

mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_ADMIN_USER" -p \
  < "$SERVICE_CODE_DIR/migrations/002_add_user_audit_fields.down.sql"
```

Warning:

Down migration audit columns drop karega. Isse audit data delete ho jayega. Production/staging me rollback se pehle DB snapshot/backup mandatory hai.

## 6. Redis / Queue / External Services

Task 7 external service status:

| Service | Required for Task 7? | Why |
|---|---:|---|
| MySQL | Yes | Audit columns same relational tables me store hote hain. |
| Docker | Optional | Local MySQL run karne ke liye convenient hai. |
| gRPC | Yes | Service API and metadata actor propagation ke liye. |
| Redis | No | Audit fields Redis use nahi karte. |
| RabbitMQ | No | Later outbox worker ke liye only. |
| Kafka | No | Later outbox worker ke liye only. |
| NATS | No | Current code me NATS setup nahi mila. |
| MinIO / S3 | No for Task 7 | KYC `storage_url` metadata exists, but audit task object storage configure nahi karta. |

Current repo caveat:

`{SERVICE_CODE_DIR}/internal/config/config.go` me event/outbox variables present hain. Ye Task 7 ka required setup nahi hai.

Task 7-only local smoke test ke liye:

```env
USER_EVENTS_ENABLED=false
OUTBOX_WORKER_ENABLED=false
```

If `OUTBOX_WORKER_ENABLED=true`, then RabbitMQ or Kafka configuration required hoga. That setup belongs to later event/outbox documentation, not `TASK_FILE_NAME`.

## 7. Environment Variables

Base `.env` setup already documented hai:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`4. Environment Variables`
```

Validation region already documented hai:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task6_Dependency.md`

Section:
`7. Environment Variables`
```

### Task 7 New Or Changed `.env` Values

Task 7 itself introduces no new secret and no new required `.env` variable.

For current repo Task 7-only local runs, add this delta only if you have not applied the later outbox migration:

```env
# Task 7-only local isolation.
# This avoids later event/outbox writes while verifying audit fields.
USER_EVENTS_ENABLED=false
OUTBOX_WORKER_ENABLED=false
```

| Variable | Required? | Example | Purpose | Security note |
|---|---:|---|---|---|
| `USER_EVENTS_ENABLED` | No for Task 7 | `false` | Prevents later outbox writes during Task 7-only smoke tests. | Not secret. Set intentionally per environment. |
| `OUTBOX_WORKER_ENABLED` | No for Task 7 | `false` | Keeps RabbitMQ/Kafka worker off. | Not secret. Do not enable without queue setup. |

Where `.env` should live:

```text
{SERVICE_CODE_DIR}/.env
```

How project loads `.env`:

- Code reads environment variables with `os.Getenv`.
- Code does not auto-load `.env`.
- Source the file before running.

```bash
SERVICE_CODE_DIR=backend/services/user-service
cd "$SERVICE_CODE_DIR"
set -a
. ./.env
set +a
go run ./cmd/server
```

Security note:

- Local `.env` may contain DB passwords.
- `.gitignore` ignores `.env` files.
- Do not copy real local passwords into docs, commits, screenshots, or logs.
- Prefer placeholders in docs, for example `change_me_local_password`.

### gRPC Metadata Required For Audit

These are not `.env` variables. Ye per-request metadata headers hain.

| Metadata key | Required? | Example | Purpose |
|---|---:|---|---|
| `x-actor-id` | Required for admin/system writes if no fallback exists | `admin_123` | Exact actor performing mutation |
| `x-actor-type` | Required with `x-actor-id` when type cannot be inferred | `admin` | Valid values: `user`, `admin`, `service`, `system` |
| `x-user-id` | Required/recommended for user-owned writes | `user_123` | User actor fallback and ownership checks |
| `x-service-name` | Required/recommended for service writes | `auth-service` | Stores as `service:auth-service` |
| `x-request-id` | Recommended | `local-task7-001` | Log/request traceability |
| `x-roles` | Optional | `seller,admin` | Future RBAC/context use |

Actor storage format:

| Actor type | Metadata example | Stored audit value |
|---|---|---|
| User | `x-user-id: user_123` | `user_123` |
| Admin | `x-actor-id: admin_456`, `x-actor-type: admin` | `admin_456` |
| Service | `x-service-name: auth-service` | `service:auth-service` |
| System | `x-actor-id: migration`, `x-actor-type: system` | `system:migration` |

Common mistake:

Do not send `created_by`, `updated_by`, `deleted_by`, or `status_changed_by` in request body. Audit fields server-side metadata se generate hote hain.

## 8. Docker Setup

No new Dockerfile, compose service, Docker network, volume, or exposed port is introduced by `TASK_FILE_NAME`.

Reuse Docker/MySQL setup:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`7. Docker and DevOps Setup`
```

Task 7 Docker-specific note:

- Same MySQL container can be reused.
- Migration `002` needs `ALTER TABLE`; use a DB admin/migration user.
- Do not run application runtime as root DB user in production.

If MySQL is running in a local Docker container, apply migration from host:

```bash
SERVICE_CODE_DIR=backend/services/user-service
mysql -h 127.0.0.1 -P 3306 -u root -p \
  < "$SERVICE_CODE_DIR/migrations/002_add_user_audit_fields.up.sql"
```

Or through `docker exec` if your local MySQL root password is the same disposable password from the previous Docker example:

```bash
SERVICE_CODE_DIR=backend/services/user-service
docker exec -i ecommerce-user-mysql mysql -uroot -proot user_db \
  < "$SERVICE_CODE_DIR/migrations/002_add_user_audit_fields.up.sql"
```

Beginner note: `docker compose down -v` local DB data delete karta hai. Task 7 audit verification ke baad accidentally data delete mat karo unless local disposable DB hai.

## 9. Local Development Setup

### Step 1: Read previous dependency documentation

Follow these first:

| Order | File | Why |
|---:|---|---|
| 1 | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Go, MySQL, Docker, `.env`, ports, grpcurl |
| 2 | `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | Base MySQL schema and migration order |
| 3 | `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | Repository DB assumptions and schema mismatch errors |
| 4 | `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | gRPC metadata/reflection/runtime setup |
| 5 | `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | Gateway and JWT context propagation |
| 6 | `TaskImplementation/{SERVICE_NAME}/task6_Dependency.md` | Validation region and request validation setup |

### Step 2: Go to project directory

From repository root:

```bash
SERVICE_CODE_DIR=backend/services/user-service
cd "$SERVICE_CODE_DIR"
```

### Step 3: Install only needed dependencies

No new Task 7 packages required.

Fresh clone only:

```bash
SERVICE_CODE_DIR=backend/services/user-service
cd "$SERVICE_CODE_DIR"
go mod download
```

### Step 4: Setup only reused databases/services

Use previous MySQL setup. Task 7 needs the same `user_db`.

### Step 5: Apply migrations

From repository root:

```bash
SERVICE_CODE_DIR=backend/services/user-service
mysql -h 127.0.0.1 -P 3306 -u root -p \
  < "$SERVICE_CODE_DIR/migrations/001_create_user_tables.up.sql"

mysql -h 127.0.0.1 -P 3306 -u root -p \
  < "$SERVICE_CODE_DIR/migrations/002_add_user_audit_fields.up.sql"
```

If `001` is already applied, do not re-run blindly in shared/staging DB. Verify tables first:

```sql
SHOW TABLES FROM user_db;
```

### Step 6: Add only new or changed environment values

Keep base `.env` from previous docs. For Task 7-only local verification, add:

```env
USER_EVENTS_ENABLED=false
OUTBOX_WORKER_ENABLED=false
```

### Step 7: Run focused tests

```bash
SERVICE_CODE_DIR=backend/services/user-service
cd "$SERVICE_CODE_DIR"
go test ./internal/audit ./internal/domain ./internal/usecase ./internal/repository ./internal/transport/grpc
```

### Step 8: Start backend service

```bash
SERVICE_CODE_DIR=backend/services/user-service
cd "$SERVICE_CODE_DIR"
set -a
. ./.env
set +a
go run ./cmd/server
```

Expected log:

```text
user_service_grpc_listening
```

### Step 9: Verify APIs related to `TASK_FILE_NAME`

List service:

```bash
grpcurl -plaintext localhost:50052 list
grpcurl -plaintext localhost:50052 describe ecommerce.user.v1.UserService
```

Create a user with service actor metadata:

```bash
grpcurl -plaintext \
  -H 'x-service-name: auth-service' \
  -H 'x-request-id: local-task7-create-user' \
  -d '{
    "auth_account_id": "auth_task7_001",
    "email": "task7.user@example.com",
    "phone": "+919999999999",
    "full_name": "Task Seven User"
  }' \
  localhost:50052 ecommerce.user.v1.UserService/CreateUser
```

Expected response should include `audit.created_by` similar to:

```text
service:auth-service
```

Update user status with admin actor metadata:

```bash
grpcurl -plaintext \
  -H 'x-actor-id: admin_123' \
  -H 'x-actor-type: admin' \
  -H 'x-request-id: local-task7-status-update' \
  -d '{
    "user_id": "<USER_ID_FROM_CREATE_RESPONSE>",
    "status": "blocked"
  }' \
  localhost:50052 ecommerce.user.v1.UserService/UpdateUserStatus
```

Expected response should include:

```text
audit.updated_by = admin_123
audit.status_changed_by = admin_123
```

Verify in MySQL:

```sql
SELECT user_id, status, created_by, updated_by, status_changed_by, status_changed_at
FROM users
WHERE email = 'task7.user@example.com';
```

## 10. Running the Project

Ports and networking:

| Service | Port | Purpose | Status |
|---|---:|---|---|
| MySQL | `3306` | Stores base tables and audit columns | Reused |
| Current service gRPC | `50052` | Internal gRPC API and audit metadata entry point | Reused |
| API Gateway HTTP | `8080` | REST-to-gRPC path if using Gateway tests | Reused |
| Redis | `6379` | Not used by Task 7 | Not required |
| RabbitMQ | `5672` | Later outbox worker only | Not required |
| Kafka | `9092` | Later outbox worker only | Not required |

How to change ports:

| Need | Config |
|---|---|
| Change MySQL host/port | Edit `USER_SERVICE_DATABASE_DSN` |
| Change current service gRPC listen address | Edit `USER_SERVICE_GRPC_ADDRESS` in `{SERVICE_CODE_DIR}/.env` |
| Change Gateway upstream target | See `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md`, section `7. Environment Variables` |

Task 7 network rule:

Do not expose the gRPC port directly to untrusted clients. Audit actor metadata should come from trusted Gateway/Auth/Superadmin callers. Agar public client direct gRPC metadata fake kar sake, audit trail unreliable ho jayega.

## 11. Common Errors & Fixes

Generic Go/MySQL/Docker errors are already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`9. Common Errors and Fixes`
```

Repository and schema mismatch errors already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`

Section:
`12. Repository-Specific Common Errors And Fixes`
```

Task 7-specific issues:

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `Unknown column 'created_by'` | Migration `002` not applied, but code expects audit columns | Apply `{MIGRATION_DIR}/002_add_user_audit_fields.up.sql` | Apply migrations in order: `001`, then `002` |
| `Duplicate column name 'created_by'` | Migration `002` re-run on same DB | Do not re-run manually; verify schema first | Use a migration tool/history table in real deployments |
| `audit actor is required` | Mutating gRPC call has no `x-user-id`, `x-service-name`, or `x-actor-id` metadata | Send correct metadata | Gateway/Auth/Superadmin must always propagate actor context |
| `audit actor is invalid` | `x-actor-type` missing/unsupported with `x-actor-id` | Use `user`, `admin`, `service`, or `system` | Validate caller metadata in Gateway before forwarding |
| `Access denied; you need ALTER privilege` | Runtime DB user used for migration | Run migration with admin/migration DB user | Separate migration credentials from app runtime credentials |
| `insert outbox event: Table 'user_db.user_outbox_events' doesn't exist` | Current repo event code enabled, but outbox migration `003` not applied | For Task 7-only set `USER_EVENTS_ENABLED=false`; or apply later outbox migration | Keep Task 7 smoke tests isolated from later event setup |
| Audit response fields are empty for old rows | Historical rows existed before audit columns | Check backfill query and `system:backfill` values | Run full `002` up migration, not partial SQL |
| Status changed but audit actor not updated | Caller path bypassed service usecase or used manual SQL | Update through service usecase/repository flow | Avoid manual DB status updates except controlled migrations |
| `deleted_by` set without `deleted_at` | Bad manual data patch or partial rollback | Repair row consistently | Keep soft-delete fields updated in same query |

## 12. Security & Best Practices

Task-specific best practices:

- Audit fields request body se accept mat karo. Always trusted context/metadata se set karo.
- `created_by`, `updated_by`, `deleted_by`, and `status_changed_by` ko server-side generate karo.
- Use UTC timestamps only. Current code UTC clock and `loc=UTC` DSN pattern use karta hai.
- Actor ids 64 characters ke andar rakho, because DB columns `VARCHAR(64)` hain.
- Service/system actors prefix ke saath store hon: `service:auth-service`, `system:backfill`.
- Migration se pehle production backup/snapshot lo.
- Migration admin credentials and app runtime credentials separate rakho.
- Runtime DB user ko unnecessary `ALTER`/`DROP` privileges mat do.
- Do not log DB DSN, JWT, raw phone numbers, email, full address, or KYC private URLs.
- Propagate `x-request-id` so audit debugging and logs match ho sake.
- Internal/admin APIs can expose audit metadata carefully; public user-facing APIs should avoid exposing sensitive actor ids unless product/security approve kare.
- Direct gRPC exposure restrict karo. Gateway/Auth layer actor identity verify kare before forwarding metadata.

## 13. Missing or Misconfigured Things

| Item | Current status | Risk | Suggested fix |
|---|---|---|---|
| No checked-in app Dockerfile/compose for current service | Previous docs also found no app container setup | Beginners may not know how to containerize service | Add service Dockerfile/compose later if deployment needs it |
| Migration runner not standardized | Raw SQL files exist, but no migration history tool wired in repo | Manual re-run can cause duplicate column/index errors | Use a migration tool and track applied versions |
| `002` migration is not idempotent | `ALTER TABLE ADD COLUMN` fails if repeated | Shared DB manual setup can break | Verify schema before manual run; prefer versioned migrations |
| Audit DB columns are nullable after migration | Old rows can migrate safely, but future missing data possible | Long-term audit completeness weaker | After backfill and confidence, consider `NOT NULL` for required actor fields |
| Current source includes later outbox hooks | `USER_EVENTS_ENABLED` defaults true | Task 7-only write tests may hit missing `user_outbox_events` table | Set `USER_EVENTS_ENABLED=false` for Task 7-only, or document/apply later event migration |
| Local `.env` can contain real DB password | `.gitignore` ignores it, but copying is still risky | Secret leakage in docs/screenshots | Use placeholders and add sanitized `.env.example` |
| No health endpoint for service | gRPC reflection helps only if enabled | Deployment health checks are harder | Add gRPC health service or platform health check later |
| Audit metadata trust depends on upstream | Direct untrusted gRPC callers can spoof metadata | Audit trail can become unreliable | Restrict gRPC network access and verify identity at Gateway/Auth |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `1. Project Tech Stack Analysis` | Base Go/MySQL/gRPC stack already explained |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `2. Language-Specific Dependency System: Go` | Go modules, `go.mod`, `go.sum`, `go.work`, and common commands already covered |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Database Analysis` | MySQL install, Docker setup, DSN, credentials, and default port reused |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Environment Variables` | Base `.env`, DB DSN, gRPC address, reflection, DB pool, shutdown, logging reused |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. Docker and DevOps Setup` | No new Docker setup added by Task 7 |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `9. Common Errors and Fixes` | Generic Go/MySQL/Docker issues already documented |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | Base tables and migration order are Task 7 prerequisites |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `11. Common Errors & Fixes` | Migration permission/base-schema issues reused |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `7. Database Setup Required For Task 3` | Repository layer depends on correct MySQL schema |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `12. Repository-Specific Common Errors And Fixes` | Schema mismatch, duplicate key, timestamp scan issues reused |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | `7. Environment Variables -> Task 4 Runtime Metadata` | gRPC metadata pattern reused and extended for audit actors |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | `8. Docker Setup` and `9. Local Development Setup` | gRPC reflection and local service startup reused |
| `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | `7. Environment Variables` | Gateway-to-service env and JWT context propagation reused |
| `TaskImplementation/{SERVICE_NAME}/task6_Dependency.md` | `7. Environment Variables` | Validation phone region setup reused |

## 15. Final Checklist

- [ ] Previous dependency documentation checked first
- [ ] No duplicate Go/MySQL/Docker setup copied unnecessarily
- [ ] Base migration `001_create_user_tables.up.sql` applied
- [ ] Task 7 migration `002_add_user_audit_fields.up.sql` applied
- [ ] Audit columns verified in all four tables
- [ ] Audit indexes verified
- [ ] Backfilled rows checked for `system:backfill`
- [ ] `.env` contains base DSN and gRPC settings from previous docs
- [ ] For Task 7-only local run, `USER_EVENTS_ENABLED=false` is set if migration `003` is not applied
- [ ] `OUTBOX_WORKER_ENABLED=false` unless later queue setup is intentionally configured
- [ ] Mutating gRPC calls include actor metadata
- [ ] `x-request-id` is propagated for debugging
- [ ] `CreateUser` smoke test shows `audit.created_by`
- [ ] Status update smoke test shows `audit.status_changed_by`
- [ ] MySQL verification query confirms audit values
- [ ] Logs checked for audit metadata errors
- [ ] Real `.env` secrets were not copied into docs or commits
- [ ] Rollback risk understood before using `002_add_user_audit_fields.down.sql`
