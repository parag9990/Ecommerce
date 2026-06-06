# Project Dependency & Setup Guide

## 1. Project Overview

### Variables Used

```text
SERVICE_NAME = "Notification Service"
TASK_FILE_NAME = "task2.md"
INPUT_FILE_PATH = TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}
OUTPUT_FILE_NAME = task2_Dependency.md
OUTPUT_FILE_PATH = TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}
```

This guide is generated from `INPUT_FILE_PATH` and saved at `OUTPUT_FILE_PATH`.
Original implementation file ko modify nahi kiya gaya.

### Task Scope

`{TASK_FILE_NAME}` ka goal MongoDB choose karna hai because notification templates aur delivery logs ka data shape flexible hota hai. Simple words me: email template, SMS text, push metadata, provider message id, retry status, and delivery payload sab same relational table shape me naturally fit nahi hote.

Current repository me runnable Go service already present hai:

```text
backend/services/notification-service
```

Task-specific dependency impact:

| Area | Task 2 Decision |
|---|---|
| Database | MongoDB |
| Service-owned database | `notification_db` |
| Main collections | `notification_templates`, `notification_deliveries` |
| Required migration for Task 2 baseline | `backend/services/notification-service/migrations/001_create_notification_collections.up.js` |
| New install steps compared with previous dependency docs | None; MongoDB setup is already documented in `task1_Dependency.md` |

Important: shared setup already exists, so this file avoids repeating same installation, Docker, Go module, and generic troubleshooting content.

### Read This First

Before following this file, read:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

Use these existing sections from that file:

| Section / Topic | Why reused |
|---|---|
| Fresh Clone Setup | Same repository clone and root setup |
| Go Dependency System | Same Go module and workspace setup |
| Environment Variables | Same `.env` loading process |
| MongoDB Setup | Same MongoDB install, Docker run, and connection string format |
| MongoDB Migrations | Same migration execution process |
| Docker Compose Example | Same local MongoDB container pattern |
| Start The Service | Same gRPC service startup flow |
| Common Setup Issues | Same generic Mongo/Go/env errors |

## 2. Tech Stack

Only task-specific or changed items are explained here. Common stack details are already explained in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

| Technology | Required for Task 2? | Status | Beginner explanation |
|---|---:|---|---|
| MongoDB | Yes | Reused setup, task-specific database choice | MongoDB ek document database hai. Isme JSON-like BSON documents store hote hain, so templates aur delivery payload flexible shape me store ho sakte hain. |
| MongoDB Go Driver v2 | Yes for runnable code | Already present | Go service MongoDB se connect/query/insert karne ke liye official driver use karta hai. |
| `mongosh` | Recommended | Reused setup | Mongo shell CLI hai. Collections, indexes, validators verify karne ke liye useful hai. |
| Go 1.26.3 | Yes for runnable service | Reused setup | Service Go me written hai. Install and module commands previous dependency guide me already covered hain. |
| gRPC + Protobuf | Runtime service API | Reused setup | Service `:9090` par gRPC expose karta hai, but Task 2 koi new API port introduce nahi karta. |
| Docker | Optional | Reused setup | Local MongoDB container run karne ke liye useful hai. No new Task 2 Docker container added. |
| RabbitMQ | No for Task 2 | Later-task dependency | Event consumer/retry tasks ke liye hai, MongoDB selection ke liye required nahi. |
| SMTP/Mailpit/SMS gateway | No for Task 2 | Later-task dependency | Actual notification send testing ke liye hai, templates/delivery storage decision ke liye nahi. |
| Redis/Kafka/MySQL/PostgreSQL | No | Not used by this task | Is task me in services ka setup required nahi hai. |

## 3. Required Software

Task 2 ke liye software requirement ka short version:

| Software | Required? | Setup source |
|---|---:|---|
| Go 1.26.3 | Yes, only if running/testing backend code | `task1_Dependency.md` section `Go Dependency System` |
| MongoDB server | Yes, if verifying actual collections/migrations | `task1_Dependency.md` section `MongoDB Setup` |
| `mongosh` | Recommended | `task1_Dependency.md` section `MongoDB Setup` |
| Docker | Optional | `task1_Dependency.md` section `MongoDB With Docker` |
| Git | Yes for clone | `task1_Dependency.md` section `Fresh Clone Setup` |

No new operating-system package is introduced by `{TASK_FILE_NAME}` beyond MongoDB and `mongosh`, and both are already covered in the previous dependency file.

## 4. Dependency Management

Refer:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Go Dependency System
```

Task 2 does not require a new `go get` command now because the current service module already contains the MongoDB driver.

Current module:

```text
backend/services/notification-service/go.mod
```

Task 2 relevant dependency:

| Dependency | Version in current repo | Why it matters |
|---|---|---|
| `go.mongodb.org/mongo-driver/v2` | `v2.6.0` | Mongo client, collection handles, BSON filters, inserts, indexes, and query operations |

Task-specific code paths using this dependency:

| File | Purpose |
|---|---|
| `backend/services/notification-service/internal/repository/mongo_notification_repository.go` | Opens Mongo connection and implements template/delivery repository methods |
| `backend/services/notification-service/internal/domain/template.go` | Defines stored template shape and validation |
| `backend/services/notification-service/internal/domain/delivery.go` | Defines delivery log shape, status, payload safety, and validation |
| `backend/services/notification-service/internal/config/config.go` | Reads Mongo URI, database, and collection env variables |
| `backend/services/notification-service/migrations/001_create_notification_collections.up.js` | Creates Task 2 collections, validators, and indexes |

Use normal Go commands from the previous dependency guide if dependencies are missing:

```bash
cd backend/services/notification-service
go mod download
go test ./...
```

Only run `go mod tidy` when you intentionally changed Go imports. Beginner note: unnecessary `go mod tidy` can modify `go.mod`/`go.sum`, so always review diff before commit.

## 5. Database Setup

### Database Decision

MongoDB is mandatory for the current runnable service and is the main dependency selected by `{TASK_FILE_NAME}`.

Hinglish explanation: MongoDB flexible documents store karta hai. Notification templates and delivery logs me channel-specific aur provider-specific fields vary karte hain, isliye MongoDB relational table se zyada natural fit hai.

Detailed MongoDB installation, Docker command, native install notes, and generic verification are already explained in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: MongoDB Setup
```

### Task 2 Database Objects

| Object | Value | Required? | Purpose |
|---|---|---:|---|
| Database | `notification_db` | Yes | `{SERVICE_NAME}` owned Mongo database |
| Templates collection | `notification_templates` | Yes | Channel-specific message templates |
| Deliveries collection | `notification_deliveries` | Yes | Per-notification delivery records and provider metadata |

### Collection: `notification_templates`

This collection stores reusable notification content.

Task-specific stable fields:

| Field | Required? | Notes |
|---|---:|---|
| `_id` | Yes | Stable document id |
| `template_key` | Yes | Business key, example `order_paid` |
| `channel` | Yes | Must be one of Task 1 channel values: `email`, `sms`, `push`, `whatsapp_like` |
| `subject` | Required for email | Email subject; non-email channels can omit it |
| `body` | Yes | Template body/content |
| `status` | Yes | `active` or `inactive` |
| `version` | Yes | Starts from `1`; latest active version is selected by sort |
| `created_at`, `updated_at` | Yes | BSON date timestamps |

Task-specific index created by migration `001`:

```javascript
{ template_key: 1, channel: 1, version: -1 }
```

Current migration names this index:

```text
template_key_channel_version
```

### Collection: `notification_deliveries`

This collection stores delivery logs. Iska data provider aur channel ke hisab se thoda flexible ho sakta hai.

Task-specific baseline fields:

| Field | Required? | Notes |
|---|---:|---|
| `_id` | Yes | Stable delivery id |
| `user_id` | Yes for normal delivery | Later OTP migration allows optional user for OTP deliveries |
| `channel` | Yes | Must follow Task 1 channel values |
| `template_key` | Yes | Template traceability |
| `status` | Yes | Example: `pending`, `accepted`, `failed`, `sent` |
| `provider` | Optional | Provider name, example `smtp` or `http_gateway` |
| `provider_message_id` | Optional | Provider reference id |
| `attempts` | Yes | Starts from `0` or `1` depending on workflow |
| `payload` | Optional | Non-sensitive event metadata only |
| `created_at`, `updated_at` | Yes | BSON date timestamps |

Task-specific indexes created by migration `001`:

```javascript
{ user_id: 1, created_at: -1 }
{ status: 1, created_at: 1 }
{ provider: 1, provider_message_id: 1 }
```

Current migration names these indexes:

```text
user_id_created_at
status_created_at
provider_message_id
```

### Migration Requirement

Migration setup process is already explained in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: MongoDB Migrations
```

For Task 2 specifically, this migration must exist and run successfully:

```text
backend/services/notification-service/migrations/001_create_notification_collections.up.js
```

It creates or updates:

| Item | Created/updated by migration `001` |
|---|---|
| `notification_templates` validator | Yes |
| `notification_deliveries` validator | Yes |
| Template unique revision index | Yes |
| Delivery user/status/provider indexes | Yes |

Task-specific verify commands:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
mongosh "$NOTIFICATION_MONGO_URI" --quiet --eval 'const d = db.getSiblingDB(process.env.NOTIFICATION_MONGO_DATABASE || "notification_db"); printjson(d.getCollectionNames());'
```

Verify Task 2 indexes:

```bash
mongosh "$NOTIFICATION_MONGO_URI" --quiet --eval 'const d = db.getSiblingDB(process.env.NOTIFICATION_MONGO_DATABASE || "notification_db"); printjson(d.notification_templates.getIndexes()); printjson(d.notification_deliveries.getIndexes());'
```

Expected result: `notification_templates` and `notification_deliveries` should exist, and the index names listed above should be visible.

## 6. Redis / Queue / External Services

No new Redis, Kafka, RabbitMQ, SMTP, SMS gateway, Mailpit, MinIO, Elasticsearch, or third-party provider setup is introduced by `{TASK_FILE_NAME}`.

| External service | Task 2 status | What to do |
|---|---|---|
| MongoDB | Required | Follow `task1_Dependency.md`, then verify Task 2 collections/indexes |
| RabbitMQ | Not required for Task 2 | Only needed for later event/retry tasks |
| Mailpit/SMTP | Not required for Task 2 | Only needed for actual email sending tests |
| SMS gateway | Not required for Task 2 | Only needed for actual SMS sending tests |
| Redis | Not used by this service task | No setup |
| Kafka | Not used by current implementation | No setup |

Beginner note: agar sirf `{TASK_FILE_NAME}` ka MongoDB decision and schema verify karna hai, then MongoDB + `mongosh` enough hai. RabbitMQ/Mailpit enable karne ki zarurat nahi.

## 7. Environment Variables

Complete env loading process is already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Environment Variables
```

Task 2 introduces no new env variables beyond the Mongo variables already documented there.

Task-specific reused variables:

| Variable | Required? | Default/example | Purpose |
|---|---:|---|---|
| `NOTIFICATION_MONGO_URI` | Yes | `mongodb://username:password@localhost:27017/notification_db?authSource=admin` | Full MongoDB connection URL |
| `NOTIFICATION_MONGO_DATABASE` | Optional | `notification_db` | Service-owned database name |
| `NOTIFICATION_MONGO_TEMPLATES_COLLECTION` | Optional | `notification_templates` | Template collection name |
| `NOTIFICATION_MONGO_DELIVERIES_COLLECTION` | Optional | `notification_deliveries` | Delivery collection name |

Where credentials go:

```text
backend/services/notification-service/.env
```

Security note: local `.env` values are for development only. Real Mongo username/password production me secret manager, Kubernetes Secret, or platform env injection se aane chahiye. Source code, markdown docs, and logs me real passwords print mat karo.

Common mistakes specific to Task 2:

| Mistake | Result | Fix |
|---|---|---|
| `NOTIFICATION_MONGO_URI` missing | Service startup fails before gRPC starts | Load `.env` before running service |
| Template and delivery collection env vars set to same value | Config validation fails with distinct collection error | Keep collection names different |
| Collection env var changed after migrations ran | Service may read a collection without expected validators/indexes | Re-run migrations with same env values or revert names |
| Database env var changed from `notification_db` accidentally | Collections appear missing | Use same DB name in URI/env/migration verification |

## 8. Docker Setup

No service-specific `Dockerfile` or `docker-compose.yml` was found for this task in the repository.

Docker setup for MongoDB is already explained in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: MongoDB With Docker
Section: Docker Compose Example
```

Task 2 Docker impact:

| Docker item | Status |
|---|---|
| New container | None |
| New port | None |
| New volume | None beyond MongoDB data volume already documented |
| New network | None |
| New health check | None |
| New restart policy | None |

If using Docker locally, only the same MongoDB container is needed to verify Task 2.

### Ports & Networking

Detailed networking explanation is reused from `task1_Dependency.md`.

| Service | Port | Purpose | Status |
|---|---:|---|---|
| MongoDB | `27017` | Stores `notification_templates` and `notification_deliveries` | Reused |
| Backend gRPC | `9090` | Runnable service API | Reused |
| Analytics HTTP | `8081` | Metrics/webhooks for later tasks, only when enabled | Not Task 2 |
| RabbitMQ AMQP | `5672` | Event/retry later tasks | Not Task 2 |
| RabbitMQ UI | `15672` | RabbitMQ admin UI | Not Task 2 |
| Mailpit SMTP | `1025` | Local email testing | Not Task 2 |
| Mailpit UI | `8025` | View captured local emails | Not Task 2 |

Port conflict guidance:

| Symptom | Likely cause | Fix |
|---|---|---|
| Mongo cannot bind `27017` | Another Mongo instance/container running | Stop old instance or change Docker port mapping and URI |
| Service starts but cannot ping Mongo | URI points to wrong host/port | Check `NOTIFICATION_MONGO_URI` and `docker ps` |
| gRPC `:9090` already in use | Another service instance running | Stop old process or change `NOTIFICATION_GRPC_ADDRESS` |

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation

Follow this first:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
```

It already covers clone, Go install, Mongo install, `.env` loading, Docker, migrations, and generic errors.

### Step 2: Go To Project Directory

Documentation folder:

```bash
cd "TaskImplementation/{SERVICE_NAME}"
```

Runnable backend service folder:

```bash
cd backend/services/notification-service
```

Use the backend folder when running Go commands or Mongo migrations.

### Step 3: Install Only New Dependencies

No new dependency install is required for `{TASK_FILE_NAME}`.

If this is a fresh clone, use previous dependency guide and run:

```bash
cd backend/services/notification-service
go mod download
```

### Step 4: Setup Only New Databases/Services

No new service beyond MongoDB. Start MongoDB using the previous dependency guide.

Task 2 requires the MongoDB baseline collections:

```text
notification_templates
notification_deliveries
```

### Step 5: Add Only New Or Changed Env Variables

No new or changed env variables. Confirm the reused Mongo variables exist in:

```text
backend/services/notification-service/.env
```

### Step 6: Run Migrations If Needed

Use the migration process from the previous dependency guide.

Task 2 minimum migration:

```text
migrations/001_create_notification_collections.up.js
```

For current full service, run all `.up.js` migrations in order because later tasks added more collections/fields.

### Step 7: Start Backend Service

Start process is reused from:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Start The Service
```

Short reminder:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
go run ./cmd/server
```

Expected successful log:

```text
notification.grpc.started
```

### Step 8: Verify Task 2 Functionality

Task 2 does not add a separate public API. Verify database setup instead:

```bash
mongosh "$NOTIFICATION_MONGO_URI" --quiet --eval 'const d = db.getSiblingDB(process.env.NOTIFICATION_MONGO_DATABASE || "notification_db"); printjson(d.notification_templates.getIndexes()); printjson(d.notification_deliveries.getIndexes());'
```

Run focused Go tests:

```bash
cd backend/services/notification-service
go test ./internal/domain ./internal/config ./internal/repository
```

What this verifies:

| Check | Why important |
|---|---|
| Domain template validation | Invalid channel/status/template shape reject hota hai |
| Domain delivery validation | Unsafe payload/status/timestamp mistakes catch hote hain |
| Config validation | Mongo URI and distinct collection names validate hote hain |
| Repository package tests | Mongo repository behaviour and query assumptions covered hain |

## 10. Running The Project

For normal local startup, do not duplicate steps. Follow:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Beginner Runbook
Section: Start The Service
```

Task 2 runtime facts:

| Runtime fact | Detail |
|---|---|
| Service will load Mongo config first | `config.Load()` reads `NOTIFICATION_MONGO_*` variables |
| Service will ping Mongo on startup | `OpenMongoNotificationRepository(...)` calls `client.Ping(...)` |
| gRPC starts only after Mongo is reachable | Mongo misconfig blocks startup |
| Templates are read from Mongo | Repository method `FindLatestActive(...)` uses `template_key`, `channel`, `status`, and version sort |
| Deliveries are stored in Mongo | Repository method `InsertDelivery(...)` writes delivery records |

Beginner note: agar `notification.grpc.started` log nahi aaya, pehle Mongo connection and `.env` check karo.

## 11. Common Errors & Fixes

Generic errors like old Go version, missing `.env`, Mongo port unreachable, and Docker daemon issues are already covered in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Common Setup Issues
```

Task 2 specific errors:

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `Mongo notification collections must be distinct` | Two Mongo collection env vars have same value | Keep `NOTIFICATION_MONGO_TEMPLATES_COLLECTION` and `NOTIFICATION_MONGO_DELIVERIES_COLLECTION` different | Do not customize collection names unless needed |
| `notification template already exists` or duplicate key error | Same `_id` or same `template_key + channel + version` inserted again | Increment `version` or use different `_id` | Treat template versions as immutable revisions |
| Mongo validation rejects email template | Email template missing non-empty `subject` | Add `subject` for `channel: "email"` | Follow migration `001` validator rules |
| `notification template not found` | No active template for requested `template_key` and `channel` | Seed/insert active template or run seed migration if later task needs rendering | Verify `status: "active"` and correct channel |
| Collections exist but indexes missing | Migration `001` did not run or ran against different database | Re-run migrations with same `.env` loaded | Always load `.env` before migration commands |
| Delivery insert fails due unsafe payload | Payload key contains `otp`, `password`, `secret`, `token`, `authorization`, or API-key-like text | Remove sensitive field or store safe reference id only | Keep payload minimal and non-sensitive |
| Service starts against empty collection name after env customization | Collection name has invalid Mongo format or blank value | Remove custom value or use valid collection name | Prefer defaults unless there is a deployment reason |

## 12. Security & Best Practices

### Task 2 Security Rules

| Rule | Beginner-friendly explanation |
|---|---|
| Keep Mongo URI secret | URI me username/password ho sakta hai. Isko logs, docs, screenshots, or Git me expose mat karo. |
| Use service-owned database | Other services ko `notification_db` directly access nahi karna chahiye. `{SERVICE_NAME}` hi owner hai. |
| Store minimum payload | Delivery payload me only useful IDs/metadata rakho, full OTP, token, password, secret, or raw provider secrets nahi. |
| Keep channel values canonical | `email`, `sms`, `push`, `whatsapp_like` only. Random vendor strings mat store karo. |
| Keep validators enabled | Mongo flexible hai, but required fields and status/channel validation still important hai. |
| Keep indexes before production traffic | Template lookup and delivery status/history queries index-backed hone chahiye. |
| Do not expose MongoDB publicly | Production me Mongo private network, auth, TLS, backups, and least privilege user ke saath run karo. |
| Version templates instead of overwriting blindly | Old delivery debugging ke liye template history useful hoti hai. |

### Configuration Audit

| Audit item | Current observation | Recommendation |
|---|---|---|
| Mongo credentials | Local `.env` file has development-style connection examples | Treat as local only; inject real credentials through secret manager |
| Docker compose | No committed service-specific compose file found | For team onboarding, create approved infra compose later if needed |
| Migration runner | JS migrations exist, but no dedicated migration CLI wrapper was found | Keep manual `mongosh` process documented or add a small migration runner in future |
| Collection customization | Env vars allow custom collection names | Avoid changing names unless migrations, config, and dashboards are updated together |
| Mongo health checks | Service pings Mongo during startup | Add container/Kubernetes readiness checks in deployment manifests later |
| Sensitive delivery payload | Domain validation blocks obvious sensitive payload keys | Continue reviewing payload contents before adding new event types |

## 13. Missing or Misconfigured Things

Task 2 itself is documented and implemented through Mongo repository/schema pieces, but onboarding gaps remain:

| Gap | Impact | Suggested fix |
|---|---|---|
| No committed Docker Compose file for this service | Beginners may copy commands manually | Add an official local compose file when infra folder is finalized |
| No single migration command/script | Developers can forget migration order | Add a documented `make migrate-notifications` or small script later |
| Mongo install steps are in previous dependency file, not repeated here | This file depends on reference discipline | Always read `task1_Dependency.md` first |
| Later migrations extend Task 2 collections | Running only migration `001` is not enough for full current service | For current service startup/testing, run all `.up.js` migrations in order |
| No separate Task 2 API endpoint | Verification must happen through DB/index checks and repository tests | Use `mongosh` index checks and focused Go tests |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Fresh Clone Setup | Same repository clone flow |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Go Dependency System | Same Go module, `go.mod`, `go.sum`, and test/build commands |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Environment Variables | Same `.env` location and loading method |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | MongoDB Setup | Same MongoDB install, Docker setup, connection string, and verification |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | MongoDB Migrations | Same `mongosh` migration execution process |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Docker Compose Example | Same optional Mongo local container setup |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Start The Service | Same `go run ./cmd/server` startup flow |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Common Setup Issues | Same generic Go/Mongo/Docker/env troubleshooting |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Credentials Placement | Same local `.env` vs production secret-manager rule |

## 15. Final Checklist

Use this checklist for `{TASK_FILE_NAME}` setup review:

- [ ] Previous dependency documentation checked first
- [ ] No duplicate setup documentation copied unnecessarily
- [ ] MongoDB server started or available
- [ ] `.env` loaded before running service or migrations
- [ ] `NOTIFICATION_MONGO_URI` configured
- [ ] `NOTIFICATION_MONGO_DATABASE` points to `notification_db` unless intentionally changed
- [ ] `NOTIFICATION_MONGO_TEMPLATES_COLLECTION` and `NOTIFICATION_MONGO_DELIVERIES_COLLECTION` are distinct
- [ ] Migration `001_create_notification_collections.up.js` ran successfully
- [ ] `notification_templates` collection exists
- [ ] `notification_deliveries` collection exists
- [ ] Task 2 indexes are visible in `mongosh`
- [ ] No OTP/password/token/secret stored in delivery payload
- [ ] Backend service starts and logs `notification.grpc.started`
- [ ] Focused Go tests run for domain/config/repository packages
- [ ] Later-task dependencies like RabbitMQ/Mailpit are kept disabled unless needed
- [ ] Production secrets are not committed or printed in logs
