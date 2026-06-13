# Project Dependency & Setup Guide

## Variables Used In This Guide

```text
SERVICE_NAME = "Session Management Service"
TASK_FILE_NAME = "task8.md"
INPUT_FILE_PATH = TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}
OUTPUT_FILE_NAME = {TASK_FILE_NAME without ".md"}_Dependency.md
OUTPUT_FILE_PATH = TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}
```

This document is generated for `INPUT_FILE_PATH` and saved at `OUTPUT_FILE_PATH`.

## 1. Project Overview

`TASK_FILE_NAME` ka focus data retention policy hai: raw events kitne din rakhne hain, session metadata kab anonymize karna hai, heatmap/analytics aggregates kab purge karne hain, aur user deletion request aane par MongoDB + Redis se user session data kaise clean karna hai.

Current repo me Task 8 ke liye actual backend pieces available hain:

| Area | Current repo path | Purpose |
|---|---|---|
| Retention domain | `backend/services/session-service/internal/domain/retention.go` | Retention classes, defaults, validation, deletion request models |
| Retention usecase | `backend/services/session-service/internal/usecase/retention.go` | Cleanup, user deletion, legal hold flow |
| Mongo retention repository | `backend/services/session-service/internal/repository/mongo_retention_repository.go` | MongoDB anonymize/delete/purge/audit operations |
| HTTP admin routes | `backend/services/session-service/internal/transport/http/handler.go` | Admin cleanup, deletion, legal-hold endpoints |
| Worker command | `backend/services/session-service/cmd/retention-worker/main.go` | Standalone retention cleanup worker |
| Migration | `backend/services/session-service/migrations/006_retention_policy.up.js` | Retention indexes, audit collection, backfill fields |

Important boundary: ye guide setup/dependency/DevOps documentation hai. Business logic rewrite nahi ki gayi.

## 2. Tech Stack

Most technology setup is reused from previous dependency files. Task 8 adds retention-specific runtime behavior but does not add a new programming language, database server, queue, or third-party SaaS.

| Technology | Required? | Task 8 role | Setup status |
|---|---|---|---|
| Go | Yes | Server and retention worker dono Go binaries hain | Reused from `TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md`, section `4. Dependency Management` |
| MongoDB | Yes | Retention indexes, raw event TTL, anonymization/purge, deletion audit collection | Base setup reused; Task 8 migration is new |
| Redis | Yes | Active session keys delete karne ke liye user deletion flow me required | Reused from `Dependency/task1_Dependency.md`, section `5.2 Redis Setup` |
| MongoDB Go Driver | Yes | Retention repository MongoDB operations run karta hai | Reused dependency |
| go-redis/v9 | Yes | User deletion active session cleanup ke liye Redis access | Reused dependency |
| `mongosh` | Recommended | `006_retention_policy.up.js` migration run/verify karne ke liye | Reused tool; Task 8 command is new |
| Docker | Optional but recommended | Local MongoDB/Redis containers | Reused from `Dependency/task1_Dependency.md`, section `8. Docker Setup` |
| Kafka/RabbitMQ/NATS | No | Task 8 current implementation me queue nahi hai | No setup needed |

Simple explanation: Retention worker ek background process hai jo old data cleanup/anonymize karta hai. Ye same Go project ka part hai, isliye separate framework install nahi hota.

## 3. Required Software

Base required software already documented hai:

```text
TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md
Sections:
3. Required Software
4. Dependency Management
5. Database Setup
8. Docker Setup
9. Local Development Setup
```

Task 8 ke liye new OS-level software install nahi hota. Ensure these existing tools are available:

| Software | Why needed for Task 8 |
|---|---|
| Go | `go run ./cmd/server` and `go run ./cmd/retention-worker` |
| MongoDB | Retention collections/indexes and cleanup target |
| `mongosh` | Run `006_retention_policy.up.js` |
| Redis | Active session keys cleanup during user deletion |
| Docker | Optional local MongoDB/Redis startup |

## 4. Dependency Management

This is a Go project. Full beginner explanation for `go.mod`, `go.sum`, `go mod download`, `go mod tidy`, build, run, module proxy, and version mismatch issues already exists in:

```text
TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md
Section:
4. Dependency Management
```

Task 8 does not add a new external Go module. Current direct dependencies remain:

| Dependency | Purpose |
|---|---|
| `go.mongodb.org/mongo-driver/v2` | MongoDB retention repository, indexes, cleanup queries |
| `github.com/redis/go-redis/v9` | Redis active-session cleanup |
| `github.com/mileusna/useragent` | Existing device tracking from previous tasks |
| `github.com/oschwald/geoip2-golang` | Existing GeoIP support from previous tasks |

If dependencies are already downloaded, no new install command is needed. If this is a fresh clone, follow the base command from `Dependency/task1_Dependency.md`:

```bash
cd backend/services/session-service
go mod download
```

## 5. Database Setup

### 5.1 MongoDB

MongoDB setup itself is reused:

```text
TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md
Section:
5.1 MongoDB Setup
```

Task 8 uses MongoDB for:

| Collection | Task 8 use |
|---|---|
| `session_events` | Raw event TTL and user deletion |
| `sessions` | Metadata anonymization, legal hold |
| `journey_summaries` | Derived journey anonymization/delete |
| `heatmap_points` | Old fine-grain heatmap purge |
| `heatmap_bucket_sessions` | Old marker purge |
| `analytics_aggregates` | Aggregate retention and PII policy check |
| `retention_deletion_audits` | New audit collection for deletion requests |

### 5.2 New Task 8 Migration

Task 8 introduces this migration:

```text
backend/services/session-service/migrations/006_retention_policy.up.js
backend/services/session-service/migrations/006_retention_policy.down.js
```

`006_retention_policy.up.js` creates/updates:

| Object | Why needed |
|---|---|
| `retention_deletion_audits` collection | Deletion request audit proof without storing raw analytics payload |
| `idx_sessions_retention_due` | Fast expired session metadata cleanup |
| `idx_sessions_retention_class_last_seen` | Retention class + last seen lookup |
| `idx_journey_retention_due` | Fast expired journey summary cleanup |
| `idx_journey_retention_class_recent` | Recent derived summary lookup |
| `idx_heatmap_retention_due` | Old heatmap point purge |
| `idx_heatmap_markers_first_seen` | Old heatmap session marker purge |
| `idx_aggregates_pii_policy` | Find aggregate docs that incorrectly contain PII |
| `idx_aggregates_retention_due` | Old aggregate purge |
| `uniq_retention_deletion_request` | Prevent duplicate deletion request IDs |
| `idx_retention_deletion_completed` | Audit query by completion time |
| `idx_retention_deletion_user_hash` | Audit lookup by hashed user ID |

Run migrations in order from the service folder:

```bash
cd backend/services/session-service
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/001_session_storage_indexes.up.js
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/002_journey_summaries.up.js
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/003_device_tracking_indexes.up.js
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/004_heatmap_points.up.js
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/005_analytics_aggregates.up.js
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/006_retention_policy.up.js
```

If `mongosh` is not installed on host and MongoDB is running in Docker, reuse the Docker exec migration style from:

```text
TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md
Section:
9. Local Development Setup -> Step 8: Run MongoDB migrations
```

Verify Task 8 migration:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.retention_deletion_audits.getIndexes().map(i => i.name)"
mongosh "mongodb://localhost:27017/session_db" --eval "db.sessions.getIndexes().map(i => i.name).filter(n => n.includes('retention'))"
mongosh "mongodb://localhost:27017/session_db" --eval "db.analytics_aggregates.getIndexes().map(i => i.name).filter(n => n.includes('retention') || n.includes('pii'))"
```

Expected names include:

```text
uniq_retention_deletion_request
idx_retention_deletion_completed
idx_sessions_retention_due
idx_aggregates_pii_policy
idx_aggregates_retention_due
```

Note: app startup also calls `EnsureIndexes()`, but migration is still recommended because it creates the audit collection validator and backfills missing retention fields.

## 6. Redis / Queue / External Services

### 6.1 Redis

Redis installation, Docker setup, default port, credential placement, and common errors are already explained here:

```text
TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md
Section:
5.2 Redis Setup
```

Task 8 uses existing Redis config only:

| Env var | Reused purpose |
|---|---|
| `SESSION_REDIS_ADDR` | Redis host/port |
| `SESSION_REDIS_PASSWORD` | Redis password if enabled |
| `SESSION_REDIS_DB` | Redis DB index |
| `SESSION_REDIS_KEY_PREFIX` | Key prefix, default `session` |

New Task 8 behavior: `DeleteUserSessionData` lists active session IDs by `user_id` or `anonymous_id`, then deletes matching active session keys. This is why retention dependencies ping Redis during startup.

### 6.2 Queue / Kafka / RabbitMQ

Task 8 current implementation does not use Kafka, RabbitMQ, NATS, or a message queue. No queue container, broker port, topic, exchange, or credential is required.

## 7. Environment Variables

### 7.1 Where To Put Values

Use the same env file location already documented:

```text
backend/services/session-service/.env
```

Important: Go code reads process environment variables. It does not automatically load `.env`. For local shell:

```bash
cd backend/services/session-service
set -a
. ./.env
set +a
```

Full `.env` loading explanation is reused from:

```text
TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md
Section:
7. Environment Variables
```

### 7.2 Task 8 `.env` Additions Only

Add these only if they are not already present:

```env
# Task 8 retention policy
SESSION_RAW_EVENT_TTL_DAYS=90
SESSION_METADATA_RETENTION_DAYS=365
SESSION_JOURNEY_RETENTION_DAYS=365
SESSION_HEATMAP_RETENTION_DAYS=730
SESSION_AGGREGATE_RETENTION_YEARS=5
SESSION_RETENTION_WORKER_BATCH_SIZE=1000
SESSION_RETENTION_WORKER_INTERVAL_MINUTES=60
SESSION_RETENTION_DRY_RUN=false
SESSION_RETENTION_WORKER_RUN_ONCE=false
SESSION_SMALL_COHORT_THRESHOLD=5
```

### 7.3 Variable Details

| Variable | Required? | Default | Valid range / note | Purpose |
|---|---|---:|---|---|
| `SESSION_RAW_EVENT_TTL_DAYS` | Yes | `90` | `30` to `180` | MongoDB TTL for `session_events` |
| `SESSION_METADATA_RETENTION_DAYS` | Yes | `365` | Must be >= raw event TTL | Session metadata anonymization window |
| `SESSION_JOURNEY_RETENTION_DAYS` | Yes | `365` | Must be >= raw event TTL | Journey summary anonymization window |
| `SESSION_HEATMAP_RETENTION_DAYS` | Yes | `730` | Must be >= journey retention | Heatmap point/marker purge window |
| `SESSION_AGGREGATE_RETENTION_YEARS` | Yes | `5` | `1` to `10` | Aggregate purge window |
| `SESSION_RETENTION_WORKER_BATCH_SIZE` | Yes | `1000` | `1` to `10000` | Cleanup batch size per run |
| `SESSION_RETENTION_WORKER_INTERVAL_MINUTES` | Yes for worker loop | `60` | Greater than `0` | Worker loop interval |
| `SESSION_RETENTION_WORKER_INTERVAL` | Optional | `1h` | Go duration like `30m`, overrides minutes form | Same interval with duration syntax |
| `SESSION_RETENTION_DRY_RUN` | Optional | `false` | `true`/`false` | Count what would change without modifying records |
| `SESSION_RETENTION_WORKER_RUN_ONCE` | Optional | `false` | `true`/`false` | Worker exits after one cleanup pass |
| `SESSION_SMALL_COHORT_THRESHOLD` | Yes | `5` | `2` to `100` | Flags/suppresses too-small aggregate cohorts |

Security note: These are policy values, not secrets. MongoDB and Redis credentials are still sensitive and must stay in `.env`, secret manager, Kubernetes Secret, or CI/CD secret variables.

## 8. Docker Setup

No new Docker container is introduced by Task 8.

Reuse MongoDB and Redis Docker setup:

```text
TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md
Section:
8. Docker Setup
```

Task 8 Docker-specific notes:

| Component | Port | Status |
|---|---:|---|
| Session HTTP API | `8086` | Reused |
| MongoDB | `27017` | Reused |
| Redis | `6379` | Reused |
| Retention worker | N/A | No HTTP port; background process |
| Kafka/RabbitMQ | N/A | Not used |

If the retention worker runs inside Docker Compose in future, it should use the same service network values:

```env
SESSION_MONGO_URI=mongodb://mongo:27017
SESSION_MONGO_DATABASE=session_db
SESSION_REDIS_ADDR=redis:6379
```

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation

Follow these first:

| File | Sections to follow |
|---|---|
| `TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md` | Required software, Go modules, MongoDB, Redis, Docker, `.env`, base run |
| `TaskImplementation/{SERVICE_NAME}/Dependency/task4_Dependency.md` | Admin auth header behavior |
| `TaskImplementation/{SERVICE_NAME}/Dependency/task6_Dependency.md` | Heatmap collections and migration `004` |
| `TaskImplementation/{SERVICE_NAME}/Dependency/task7_Dependency.md` | Analytics aggregates and migration `005` |

### Step 2: Go To Project Directory

```bash
cd backend/services/session-service
```

### Step 3: Install Dependencies

No new Task 8 module is added. For fresh clone only:

```bash
go mod download
```

### Step 4: Start Existing Databases/Services

Use previous MongoDB + Redis setup. If using Docker containers from `Dependency/task1_Dependency.md`:

```bash
docker start ecommerce-session-mongo ecommerce-session-redis
```

### Step 5: Load Environment Variables

```bash
set -a
. ./.env
set +a
```

### Step 6: Add Task 8 Retention Variables

Add the variables from section `7.2 Task 8 .env Additions Only`.

### Step 7: Run Migrations Through Task 8

```bash
SESSION_MONGO_DATABASE=session_db mongosh "mongodb://localhost:27017" migrations/006_retention_policy.up.js
```

If this is a fresh database, run migrations `001` through `006` in order as shown in section `5.2`.

### Step 8: Start Backend Server

```bash
go run ./cmd/server
```

Verify readiness:

```bash
curl "http://localhost:8086/healthz"
curl "http://localhost:8086/readyz"
```

`/readyz` should show MongoDB and Redis as OK.

### Step 9: Run Retention Worker Once For Local Testing

Dry run first:

```bash
SESSION_RETENTION_DRY_RUN=true SESSION_RETENTION_WORKER_RUN_ONCE=true go run ./cmd/retention-worker
```

Real one-shot cleanup:

```bash
SESSION_RETENTION_DRY_RUN=false SESSION_RETENTION_WORKER_RUN_ONCE=true go run ./cmd/retention-worker
```

Continuous worker loop:

```bash
SESSION_RETENTION_WORKER_INTERVAL_MINUTES=60 go run ./cmd/retention-worker
```

Simple explanation: dry run beginner-safe hai. Pehle output/logs check karo, phir `SESSION_RETENTION_DRY_RUN=false` use karo.

## 10. Running the Project

### 10.1 Backend API

```bash
cd backend/services/session-service
set -a
. ./.env
set +a
go run ./cmd/server
```

### 10.2 Verify Task 8 Admin APIs

Admin header behavior is reused from:

```text
TaskImplementation/{SERVICE_NAME}/Dependency/task4_Dependency.md
Section:
6.3 API Gateway / Admin Auth Context
```

Allowed local roles include `admin`, `superadmin`, and `operations_admin`.

Run retention cleanup via HTTP dry run:

```bash
curl -i -X POST "http://localhost:8086/api/v1/admin/sessions/retention/run" \
  -H "Content-Type: application/json" \
  -H "X-User-Roles: admin" \
  -H "X-User-ID: local-admin" \
  -d '{"request_id":"ret_local_001","dry_run":true}'
```

Delete/anonymize user session data:

```bash
curl -i -X POST "http://localhost:8086/api/v1/admin/sessions/delete-user-data" \
  -H "Content-Type: application/json" \
  -H "X-User-Roles: admin" \
  -H "X-User-ID: local-admin" \
  -d '{
    "request_id": "del_local_001",
    "user_id": "user_123",
    "anonymous_ids": ["anon_123"],
    "reason": "local privacy deletion test",
    "hard_delete": false
  }'
```

Set legal hold for a session:

```bash
curl -i -X POST "http://localhost:8086/api/v1/admin/sessions/sess_123/legal-hold" \
  -H "Content-Type: application/json" \
  -H "X-User-Roles: admin" \
  -H "X-User-ID: local-admin" \
  -d '{"hold":true,"reason":"local investigation test"}'
```

Verify deletion audit:

```bash
mongosh "mongodb://localhost:27017/session_db" --eval "db.retention_deletion_audits.findOne({request_id:'del_local_001'})"
```

## 11. Common Errors & Fixes

Generic MongoDB, Redis, Docker, `.env`, `go mod download`, and admin auth errors are already documented in:

```text
TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md
Section:
12. Common Errors & Fixes

TaskImplementation/{SERVICE_NAME}/Dependency/task4_Dependency.md
Section:
11. Common Errors & Fixes
```

Task 8-specific issues:

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `invalid retention config: ... raw event TTL must be between 30 and 180 days` | `SESSION_RAW_EVENT_TTL_DAYS` too low/high | Set value between `30` and `180` | Keep policy values reviewed before deploy |
| `session metadata retention cannot be shorter than raw event TTL` | Metadata days < raw event days | Increase `SESSION_METADATA_RETENTION_DAYS` | Keep retention windows ordered |
| `heatmap retention cannot be shorter than journey summary retention` | Heatmap retention days too small | Increase `SESSION_HEATMAP_RETENTION_DAYS` | Use default `730` unless policy says otherwise |
| `aggregate documents violate pii policy` | Some `analytics_aggregates.contains_pii=true` | Inspect and remove/rebuild unsafe aggregate docs | Never store user IDs, anonymous IDs, IP hashes in aggregate docs |
| `RETENTION_UNAVAILABLE` | Handler has no retention usecase configured | Start current server binary from this repo | Do not use stale binary |
| `RETENTION_STORAGE_UNAVAILABLE` | MongoDB or Redis unavailable during retention action | Check `/readyz`, MongoDB, Redis | Keep worker and API behind readiness checks |
| Duplicate deletion request ID | `uniq_retention_deletion_request` index blocks duplicate `request_id` | Use a new request ID or inspect existing audit | Make deletion request IDs idempotency-safe |
| Worker exits immediately | `SESSION_RETENTION_WORKER_RUN_ONCE=true` | Set it to `false` for continuous loop | Use run-once only for local/manual jobs |
| Worker changes data unexpectedly | Dry run was disabled | Use `SESSION_RETENTION_DRY_RUN=true` first | Make dry run default in staging/manual tests |

## 12. Security & Best Practices

Task 8 security is privacy-sensitive because deletion and retention change stored user activity data.

| Practice | Why |
|---|---|
| Keep retention/admin endpoints private behind API Gateway | Local header-based admin auth can be spoofed if service is public |
| Use `SESSION_RETENTION_DRY_RUN=true` before production cleanup | Shows impact before changing data |
| Keep `request_id` unique for deletion requests | Avoid duplicate/audit confusion |
| Store deletion audit metadata only, not raw analytics payload | Audit proof without retaining sensitive user behavior |
| Do not store PII in `analytics_aggregates` | Aggregates are allowed to live longer only if PII-free |
| Keep legal hold reason and actor available | Compliance/audit team ko context chahiye |
| Monitor cleanup counts | Sudden huge delete/anonymize count policy/config bug indicate kar sakta hai |
| Back up MongoDB before changing retention windows | Retention cleanup may permanently remove data |

Production warning: `X-User-Roles: admin` should only be trusted after a gateway/auth service validates the user and overwrites inbound identity headers.

## 13. Missing or Misconfigured Things

| Item | Current status | Risk | Suggested fix |
|---|---|---|---|
| No committed Docker Compose worker service | Worker command exists, compose service not visible | Ops may forget to run retention loop | Add `retention-worker` service in local/prod compose |
| No scheduler/Kubernetes CronJob manifest | Worker can run manually/loop, but no deployment schedule docs | Retention may not run in prod | Add CronJob or long-running worker deployment |
| Header-based admin auth | Local implementation checks trusted headers | Public spoofing risk | Put service behind trusted gateway; strip inbound auth headers |
| Retention migration must be run manually | `006_retention_policy.up.js` exists | Fresh DB may miss validators/backfill | Add migration runner/version tracking |
| Aggregate PII check fails closed | Good safety behavior, but can stop cleanup | Unsafe aggregate docs block job | Fix aggregate generator to always set `contains_pii=false` only when verified |
| No dedicated retention metrics dashboard | Logs exist, dashboard not visible | Cleanup failures may be missed | Export cleanup counts/errors to monitoring |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md` | Required software, Go modules, MongoDB, Redis, Docker, `.env`, base run, troubleshooting | Same local runtime stack |
| `TaskImplementation/{SERVICE_NAME}/Dependency/task2_Dependency.md` | MongoDB/Redis storage concepts | Same session storage layer |
| `TaskImplementation/{SERVICE_NAME}/Dependency/task3_Dependency.md` | Event ingestion, raw events, queue not required | Retention acts on existing raw events |
| `TaskImplementation/{SERVICE_NAME}/Dependency/task4_Dependency.md` | Journey summaries and admin auth header behavior | Retention anonymizes journeys and uses admin routes |
| `TaskImplementation/{SERVICE_NAME}/Dependency/task5_Dependency.md` | Device/GeoIP/privacy fields | Retention protects device, geo, IP hash related data |
| `TaskImplementation/{SERVICE_NAME}/Dependency/task6_Dependency.md` | Heatmap setup and migration `004` | Retention purges heatmap points/markers |
| `TaskImplementation/{SERVICE_NAME}/Dependency/task7_Dependency.md` | Analytics aggregate setup and migration `005` | Retention checks/purges aggregate documents |

## 15. Final Checklist

- [ ] Previous dependency documentation checked.
- [ ] Base Go, MongoDB, Redis, Docker setup followed from `Dependency/task1_Dependency.md`.
- [ ] `.env` loaded into shell before running server/worker.
- [ ] Task 8 retention env variables reviewed and added.
- [ ] Retention windows are valid: raw <= metadata/journey <= heatmap.
- [ ] MongoDB migrations `001` through `006` run in order.
- [ ] `retention_deletion_audits` collection exists.
- [ ] Retention indexes verified in MongoDB.
- [ ] Backend `/readyz` shows MongoDB and Redis OK.
- [ ] Admin retention dry-run endpoint tested.
- [ ] Standalone retention worker tested with `SESSION_RETENTION_DRY_RUN=true`.
- [ ] User deletion audit record verified.
- [ ] Production admin endpoints protected behind trusted API Gateway.
- [ ] No duplicate setup documentation added.

Final Hinglish summary: Task 8 me new install ka load nahi hai. Main kaam hai existing MongoDB/Redis setup reuse karna, retention env values set karna, migration `006` run karna, dry-run cleanup verify karna, aur production me retention/admin endpoints ko strongly protect karna.
