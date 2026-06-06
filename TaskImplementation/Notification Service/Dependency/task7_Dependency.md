# Project Dependency & Setup Guide

## 1. Project Overview

### Variables Used

```text
SERVICE_NAME = "Notification Service"
TASK_FILE_NAME = "task7.md"
INPUT_FILE_PATH = TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}
OUTPUT_FILE_NAME = {TASK_FILE_NAME without ".md"}_Dependency.md
OUTPUT_FILE_PATH = TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}
```

This file documents only the dependency, setup, environment, database, and operational requirements for the user-preference implementation described by `{TASK_FILE_NAME}`. The original implementation file is not modified.

Simple Hinglish summary: current task user ke email, SMS, push, aur marketing choices MongoDB me store karta hai. Notification provider ko call karne se pehle latest preference check hoti hai. Opted-out event notification ko `suppressed` status ke saath audit kiya jata hai, retry ya DLQ me unnecessarily nahi bheja jata.

### Current Implementation Scope

The runnable implementation is present in:

```text
backend/services/notification-service/internal/domain/preference.go
backend/services/notification-service/internal/usecase/manage_preference.go
backend/services/notification-service/internal/usecase/consent_gate.go
backend/services/notification-service/internal/usecase/send_delivery_attempt.go
backend/services/notification-service/internal/repository/mongo_notification_repository.go
backend/services/notification-service/internal/transport/grpc/notification_handler.go
backend/services/notification-service/internal/config/config.go
backend/services/notification-service/cmd/server/main.go
backend/services/notification-service/migrations/006_add_notification_preferences_and_suppression.up.js
proto/ecommerce/notification/v1/notification.proto
```

Current runtime facts:

- MongoDB is mandatory because preference RPCs and send-time consent checks read the service-owned preference collection.
- No new Go module, container, queue, port, secret, or third-party provider is introduced.
- `NOTIFICATION_MONGO_PREFERENCES_COLLECTION` is the only task-specific environment variable.
- The task lists User Service as a dependency, but the current runtime makes no direct User Service network call. Authenticated identity arrives through trusted gRPC metadata.
- RabbitMQ is required only when testing/enabling event consumption and queued retry enforcement. Basic preference GET/PATCH RPCs need MongoDB but not RabbitMQ.
- Missing preference records return safe defaults and are persisted only after an update.

### Read Previous Dependency Docs First

Shared setup is intentionally not repeated. Start with:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
TaskImplementation/{SERVICE_NAME}/task6_Dependency.md
```

Most important reused topics:

| Previous file | Reused topic |
|---|---|
| `task1_Dependency.md` | Clone, Go modules, `.env` loading, MongoDB installation, Docker examples, startup, credentials |
| `task2_Dependency.md` | Base MongoDB collections, delivery storage, schema safety |
| `task3_Dependency.md` | Rendered notification templates |
| `task4_Dependency.md` | OTP provider and OTP security boundary |
| `task5_Dependency.md` | RabbitMQ event runtime and producer contracts |
| `task6_Dependency.md` | Retry worker, DLQ, encryption key, migration-order warning |

## 2. Tech Stack

| Technology | Required? | Current-task use | Beginner-friendly explanation |
|---|---:|---|---|
| Go `1.26.3` | Yes | Preference domain rules, use cases, gRPC handlers, repository, and tests | Go compiled backend language hai. Existing install and module setup `task1_Dependency.md`, section `Go Dependency System` me explained hai. |
| Go modules | Yes | Existing dependency versions lock karta hai | `go.mod` aur `go.sum` dependency list/checksums maintain karte hain. Current task koi new module add nahi karta. |
| MongoDB | Yes | Stores one preference document per user and suppression audit state | MongoDB document database hai. Existing server reuse hota hai; migration `006` current task ka new database change hai. |
| MongoDB Go Driver `v2.6.0` | Yes | Reads and atomically upserts preferences | Already present in `go.mod`; separate `go get` required nahi hai. |
| gRPC + Protobuf | Yes | Exposes `GetNotificationPreference` and `UpdateNotificationPreference` | gRPC internal service API hai; Protobuf typed request/response contract define karta hai. |
| API Gateway auth metadata | Required for preference RPC access | Supplies `x-user-id` and `x-roles` | Gateway ko verified user identity forward karni hai. Client-provided spoofed headers trust nahi karne chahiye. |
| RabbitMQ | Conditional | Queued notifications and retries par latest consent re-check | Same broker from earlier tasks; preference RPC ke liye required nahi hai. |
| SMTP/Mailpit | Conditional | Allowed/suppressed email end-to-end comparison | Same provider setup from earlier tasks; suppression verify karne ke liye useful hai. |
| Redis, Kafka, NATS, SQL database | No | No adapter/config exists for this task | In services ko current task ke liye start ya install mat karo. |

### Preference Behaviour Relevant To Setup

| Situation | Current result |
|---|---|
| No preference document exists | Email enabled; SMS, push, and marketing disabled |
| Security OTP through trusted path | Email/SMS allowed independently of stored marketing preference |
| Transactional notification | Requested channel must be enabled |
| Marketing notification | Requested channel and global marketing preference must both be enabled |
| WhatsApp-like transactional/marketing message | Suppressed because no public preference toggle exists |
| Preference changes while delivery waits in retry queue | Latest preference is checked immediately before the next provider call |

## 3. Required Software

| Software | Required? | Setup source / note |
|---|---:|---|
| Git | Yes for fresh clone | Refer `task1_Dependency.md`, section `Fresh Clone Setup` |
| Go `1.26.3` | Yes | Refer `task1_Dependency.md`, section `Go Dependency System` |
| MongoDB server | Yes | Reused existing database installation |
| `mongosh` | Yes for migration and manual verification | Runs migration `006` and inspects preferences/suppressed deliveries |
| `grpcurl` | Optional but recommended | Manually invokes preference RPCs because server reflection is not enabled |
| RabbitMQ | Conditional | Required only for queued event/retry consent verification |
| Mailpit or working provider | Conditional | Required only for allowed-delivery smoke testing |
| Docker | Optional | Reused local infrastructure option |

No new native package or external account is required for current-task functionality.

## 4. Dependency Management

Go installation, `go mod download`, `go mod tidy`, proxy/cache troubleshooting, build, and generic test commands are already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Go Dependency System
```

The current task reuses these already-locked direct dependencies:

```text
go.mongodb.org/mongo-driver/v2 v2.6.0
google.golang.org/grpc v1.81.1
google.golang.org/protobuf v1.36.11
```

Fresh-clone dependency command:

```bash
cd backend/services/notification-service
GOWORK=off go mod download
```

Do not run `go get` only for this task. No dependency addition is needed.

Verification commands:

```bash
cd backend/services/notification-service
GOWORK=off go build -o /tmp/notification-service-server ./cmd/server
GOWORK=off go test ./...
```

Current repository verification result:

- The server build passes.
- `GOWORK=off go test ./...` currently fails only at `TestManagePreferencePatchPreservesExplicitFalseAndOmittedFields`.
- The failing test fixture uses a stored `created_at` later than its injected update clock, so validation rejects `updated_at` before `created_at`. Fix this before treating full CI as green.

## 5. Database Setup

### Reused MongoDB Setup

MongoDB installation, Docker container, persistent volume, connection URI, credentials, health checks, and generic migration execution are already explained in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: MongoDB Setup
Section: MongoDB Migrations
```

The base delivery collection and payload-security rules are already explained in:

```text
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
Section: Database Setup
Section: Security & Best Practices
```

### New Database Change: Migration `006`

Task-specific files:

```text
backend/services/notification-service/migrations/006_add_notification_preferences_and_suppression.up.js
backend/services/notification-service/migrations/006_add_notification_preferences_and_suppression.down.js
```

Migration `006` performs two required changes:

| Change | Purpose |
|---|---|
| Creates `notification_preferences` with strict validation | Stores safe per-user channel and marketing choices |
| Creates unique index `unique_notification_preference_user` on `user_id` | Guarantees one preference document per user |
| Adds delivery status `suppressed` | Records intentional non-delivery without provider call |
| Adds `suppression_reason` validation | Keeps a safe audit reason such as `marketing_opted_out` |
| Updates delivery validator | Allows suppressed event records while blocking recipient/retry fields on them |

Preference document fields:

| Field | Meaning |
|---|---|
| `_id` | Stable preference record ID |
| `user_id` | Authenticated owner; unique |
| `email_enabled`, `sms_enabled`, `push_enabled` | Per-channel choices |
| `marketing_enabled` | Global promotional opt-in |
| `consent_source` | `profile_settings`, `onboarding`, or `support_request` |
| `created_at`, `updated_at` | UTC audit timestamps |

### Migration Order Warning

> Warning: migration `006` uses MongoDB `collMod` and replaces the `notification_deliveries` validator. Run it after migrations `001` through `005`. Do not re-run it after migration `007`, because that would overwrite later analytics-aware validator rules.

Use the correct flow:

| Database state | Correct action |
|---|---|
| Fresh database | Run every current `.up.js` file once, filename order me |
| Database has `001` through `005` only | Run migration `006`, then later migrations when needed |
| Database already has `007` applied | Do not re-run `006`; verify collection/index/validator instead |

The repository has migration scripts but no migration-history collection or automated migration runner. Applied versions must be tracked carefully outside the service.

### Run Migration `006`

For a database known to be at migration `005`:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
mongosh "$NOTIFICATION_MONGO_URI" migrations/006_add_notification_preferences_and_suppression.up.js
```

For a fresh current-repository database:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
for file in migrations/*.up.js; do
  mongosh "$NOTIFICATION_MONGO_URI" "$file"
done
```

Run the fresh-database loop only once. Re-running old `collMod` migrations over a live newer database can regress the active validator.

### Verify Preference Collection And Index

```bash
mongosh "$NOTIFICATION_MONGO_URI" --quiet --eval '
const d = db.getSiblingDB(process.env.NOTIFICATION_MONGO_DATABASE || "notification_db");
const name = process.env.NOTIFICATION_MONGO_PREFERENCES_COLLECTION || "notification_preferences";
printjson(d.getCollectionInfos({ name }));
printjson(d.getCollection(name).getIndexes());
'
```

Expected:

- Collection uses strict schema validation.
- Index `unique_notification_preference_user` exists and is unique.

### Verify Stored Preferences

After an update RPC:

```bash
mongosh "$NOTIFICATION_MONGO_URI" --quiet --eval '
const d = db.getSiblingDB(process.env.NOTIFICATION_MONGO_DATABASE || "notification_db");
const name = process.env.NOTIFICATION_MONGO_PREFERENCES_COLLECTION || "notification_preferences";
printjson(d.getCollection(name).find(
  {},
  { _id: 1, user_id: 1, email_enabled: 1, sms_enabled: 1, push_enabled: 1, marketing_enabled: 1, consent_source: 1, updated_at: 1 }
).limit(10).toArray());
'
```

### Verify Suppressed Deliveries

```bash
mongosh "$NOTIFICATION_MONGO_URI" --quiet --eval '
const d = db.getSiblingDB(process.env.NOTIFICATION_MONGO_DATABASE || "notification_db");
const name = process.env.NOTIFICATION_MONGO_DELIVERIES_COLLECTION || "notification_deliveries";
printjson(d.getCollection(name).find(
  { status: "suppressed" },
  { _id: 1, user_id: 1, channel: 1, template_key: 1, source_event_id: 1, suppression_reason: 1, updated_at: 1 }
).sort({ updated_at: -1 }).limit(10).toArray());
'
```

Routine debugging output should not include recipient ciphertext, rendered content, or secrets.

### Rollback Caution

The down migration converts `suppressed` deliveries to `failed`, removes suppression reasons, restores the earlier validator, and drops the preference collection. Production rollback deletes user choices and changes compliance audit meaning. Backup data and obtain explicit approval before rollback.

## 6. Redis / Queue / External Services

### User Service Boundary

The requirement lists User Service as a dependency, but the implemented service does not need a User Service URL, port, token, client, or database credential.

Current boundary:

1. API Gateway authenticates the request.
2. Gateway forwards exactly one trusted `x-user-id` and an allowed `x-roles` value.
3. The gRPC handler reads that identity and never accepts `user_id` from the preference request body.
4. Preferences remain owned by this service's MongoDB database.

Do not read or write the User Service database directly.

### RabbitMQ And Retry Enforcement

RabbitMQ setup, topology, environment, and troubleshooting are already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
Section: RabbitMQ Is Newly Mandatory For This Task

TaskImplementation/{SERVICE_NAME}/task6_Dependency.md
Section: RabbitMQ Setup Is Reused
Section: New Retry/DLQ Topology
```

Current-task impact:

- No new exchange, queue, binding, vhost, or RabbitMQ plugin is added.
- The retry worker loads the latest preference immediately before rendering/provider send.
- An opted-out queued delivery becomes `suppressed`.
- A suppressed delivery is completed without retry publication or DLQ publication.
- Event consumption must still have valid MongoDB, RabbitMQ, email-provider, retry, and encryption settings from earlier tasks.

### External Service Summary

| Service | Required? | Current-task note |
|---|---:|---|
| MongoDB | Yes | New collection and validator migration |
| User Service | Logical dependency only | No direct runtime call/config currently exists |
| API Gateway | Required for public authenticated flow | Must provide trusted user/role metadata; current gateway implementation is absent |
| RabbitMQ | Conditional | Reused for queued/retry consent enforcement |
| SMTP/Mailpit | Conditional | Reused to compare allowed versus suppressed delivery |
| Redis/Kafka/NATS | No | No current-task adapter/config |

## 7. Environment Variables

The `.env` location, shell loading method, credentials placement, and generic mistakes are already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Environment Variables
Section: Credentials Placement
```

Create/update the Git-ignored local file:

```text
backend/services/notification-service/.env
```

The service reads process environment with `os.LookupEnv`; it does not automatically parse `.env`:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
```

### New Task-Specific Variable

| Variable | Required? | Default | Purpose | Security note |
|---|---:|---|---|---|
| `NOTIFICATION_MONGO_PREFERENCES_COLLECTION` | Optional | `notification_preferences` | Selects the service-owned preference collection | Not a secret; keep it distinct from other collection names |

Incremental `.env` example:

```env
# Current-task preference storage. Default shown explicitly for clarity.
NOTIFICATION_MONGO_PREFERENCES_COLLECTION=notification_preferences
```

No new credential or secret is introduced.

Reused variables:

- `NOTIFICATION_MONGO_URI` and `NOTIFICATION_MONGO_DATABASE` are mandatory for service startup.
- `NOTIFICATION_GRPC_ADDRESS` controls the existing internal RPC listener.
- RabbitMQ, retry, delivery-encryption, and provider variables are required only when event/retry delivery is enabled.

Common current-task mistakes:

- Setting the preference collection equal to templates, deliveries, or provider-events collection causes startup validation failure.
- Expecting `.env` to load automatically causes missing MongoDB configuration.
- Changing the collection name without applying migration `006` to that same name causes read/update failures.

## 8. Docker Setup

No committed service-specific Dockerfile or Docker Compose file was found. Reuse:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: MongoDB With Docker
Section: RabbitMQ Setup
Section: Email Provider Setup With Mailpit
Section: Docker Compose Example
```

Current-task Docker impact:

| Docker item | Status |
|---|---|
| MongoDB container and volume | Reused; migration `006` required |
| RabbitMQ container and volume | Reused only for event/retry verification |
| Mailpit container | Reused only for allowed email smoke |
| New container, image, network, or volume | None |
| New health check or restart policy | None |

### Ports & Networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Backend gRPC | `9090` | Preference RPCs and existing internal API | Reused |
| MongoDB | `27017` | Preference and delivery persistence | Reused |
| RabbitMQ AMQP | `5672` | Optional event/retry delivery path | Reused, conditional |
| RabbitMQ Management UI | `15672` | Optional broker debugging | Reused, conditional |
| Mailpit SMTP | `1025` | Optional local allowed-email smoke | Reused, conditional |
| Mailpit UI | `8025` | Optional captured-email inspection | Reused, conditional |
| Analytics HTTP | `8081` | Later optional metrics/webhooks | Reused, not required |
| User Service | None introduced | No direct call in current implementation | Not applicable |

No new port is introduced. Keep gRPC internal; the preference handler trusts gateway-supplied metadata.

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation

Follow the previous files listed in section 1. Complete the reused Go and MongoDB setup first. Complete RabbitMQ/Mailpit setup only when testing queued delivery enforcement.

### Step 2: Go To The Runnable Service

```bash
cd backend/services/notification-service
```

### Step 3: Install Existing Dependencies

```bash
GOWORK=off go mod download
```

No new `go get` is needed.

### Step 4: Start Required Services

For preference GET/PATCH:

1. Start MongoDB.

For queued/event suppression verification:

1. Start MongoDB.
2. Start RabbitMQ.
3. Start Mailpit or another configured provider.

Use the reused commands from earlier dependency files.

### Step 5: Add And Load New Environment Configuration

Add the incremental variable from section 7, keep reused MongoDB values valid, then:

```bash
set -a
. ./.env
set +a
```

### Step 6: Run Migrations Safely

Apply migration `006` only when earlier migrations are already present. For a fresh database, apply all current migrations once in order. Follow the migration-order warning in section 5.

### Step 7: Run Tests And Build

```bash
GOWORK=off go build -o /tmp/notification-service-server ./cmd/server
GOWORK=off go test ./...
```

The full suite currently has the single known preference timestamp-fixture failure documented in section 4.

### Step 8: Start The Backend Service

```bash
GOWORK=off go run ./cmd/server
```

Expected minimum startup log:

```text
notification.grpc.started
```

With event/retry runtime enabled, also expect:

```text
notification.events.rabbitmq.started
notification.retry.rabbitmq.started
```

## 10. Running The Project

### Verify Preference RPCs With `grpcurl`

The gRPC server does not register reflection, so provide the source proto explicitly.

From `backend/services/notification-service`:

```bash
grpcurl -plaintext \
  -import-path ../../../proto \
  -proto ecommerce/notification/v1/notification.proto \
  -H 'x-user-id: user_local_001' \
  -H 'x-roles: buyer' \
  -d '{}' \
  localhost:9090 \
  ecommerce.notification.v1.NotificationService/GetNotificationPreference
```

Expected first-read defaults:

```json
{
  "emailEnabled": true
}
```

Fields with `false` may be omitted by Proto JSON output. This first GET does not create a MongoDB document.

Update selected fields:

```bash
grpcurl -plaintext \
  -import-path ../../../proto \
  -proto ecommerce/notification/v1/notification.proto \
  -H 'x-user-id: user_local_001' \
  -H 'x-roles: buyer' \
  -d '{"email_enabled":false,"marketing_enabled":true}' \
  localhost:9090 \
  ecommerce.notification.v1.NotificationService/UpdateNotificationPreference
```

Verify:

1. MongoDB now contains one preference document for `user_local_001`.
2. Omitted fields preserve their previous/default values.
3. Explicit `false` is preserved.
4. Repeated updates do not create a second document.
5. A request without `x-user-id` and permitted `x-roles` returns `Unauthenticated`.

### Verify Queued Opt-Out Enforcement

This flow reuses event and retry setup from `task5_Dependency.md` and `task6_Dependency.md`.

Currently reachable event-consumer smoke:

1. Set `email_enabled=false` for a test user.
2. Publish one of the currently supported transactional email events, such as `OrderPaid`, with that `user_id`.
3. Verify no provider message is sent.
4. Verify MongoDB delivery status is `suppressed` with reason `channel_opted_out`.
5. Verify no retry or DLQ message is created for that suppression.

Latest-preference retry smoke:

1. Start with `email_enabled=true`.
2. Make the provider temporarily unavailable and publish a supported transactional email event.
3. Wait until the delivery becomes `retry_scheduled`.
4. Set `email_enabled=false` before the next attempt.
5. Verify the next attempt becomes `suppressed` instead of calling the provider or entering the DLQ.

Marketing policy is implemented and unit-tested, but the current `internal/events/rules.go` file has no marketing event rule. End-to-end `marketing_opted_out` verification needs a future approved marketing producer/event path or a dedicated integration harness.

Do not use the OTP flow to test marketing opt-out. Trusted email/SMS OTP intentionally bypasses stored marketing preference.

### Public REST Route Status

The route contract exists in `api/master-api.json`:

```text
GET /api/v1/me/notification-preferences
PATCH /api/v1/me/notification-preferences
```

However, the current API Gateway folder contains only an `.env` file and no runnable route implementation. Use direct authenticated gRPC smoke tests until the gateway proxy/auth implementation exists.

## 11. Common Errors & Fixes

Generic Go, MongoDB, Docker, `.env`, RabbitMQ, and provider failures are already covered in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Common Setup Issues

TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
Section: Common Errors & Fixes

TaskImplementation/{SERVICE_NAME}/task6_Dependency.md
Section: Common Errors & Fixes
```

Current-task-specific issues:

| Error / symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `NOTIFICATION_MONGO_PREFERENCES_COLLECTION` validation error | Blank, invalid, or duplicate collection name | Restore `notification_preferences` or another valid distinct name | Keep collection names stable and distinct |
| Preference update fails with collection/validator error | Migration `006` not applied to configured collection | Apply migration in correct order and verify validator | Run migration before deploying dependent code |
| Duplicate key on `user_id` | Data was inserted before unique index or through unsafe manual writes | Remove/merge duplicates, then create unique index | Use service upsert path and keep index enabled |
| GET returns defaults but MongoDB has no row | Expected behavior; defaults are not persisted on read | Send a valid update if persistence is needed | Understand read-default versus saved preference |
| Update returns `InvalidArgument` | PATCH contains no fields | Send at least one optional boolean field | Validate gateway/client patch body |
| RPC returns `Unauthenticated` | Missing/invalid `x-user-id` or no permitted role | Send trusted metadata through gateway or local `grpcurl` test | Keep auth propagation contract tested |
| User can spoof another identity | Gateway forwards client-supplied identity headers or gRPC is publicly reachable | Strip external identity headers and inject verified identity only | Keep gRPC internal and add authenticated gateway/interceptor |
| Marketing email still sends after opt-out | Consent gate not used in a delivery path, stale deployment, or wrong user ID | Verify current server wiring, delivery user ID, and latest code | Contract-test every send path through consent gate |
| Queued opted-out delivery retries/DLQs | Worker deployment predates suppression handling or migration `006` is absent | Deploy current worker and apply migration `006` | Test suppression before production rollout |
| `collMod` fails | Earlier delivery collection migration missing | Apply migrations `001` through `005` first | Track migration history |
| Later analytics fields stop validating | Migration `006` was re-run after migration `007` | Re-apply latest approved validator migration and inspect data | Never replay old validator migrations |
| Full `go test ./...` fails in preference patch test | Test clock is earlier than fixture `created_at` | Align test fixture timestamps/clock | Keep deterministic valid timestamps in tests |
| REST route returns unavailable/not found | API Gateway implementation is absent | Use direct gRPC locally; implement gateway route separately | Include gateway in end-to-end deployment checklist |

## 12. Security & Best Practices

### Current-Task Security Rules

| Rule | Why |
|---|---|
| Default marketing to disabled | Missing consent must not be interpreted as opt-in |
| Derive user identity from trusted auth context | Prevents User A from editing User B's preferences |
| Keep gRPC listener internal | Handler currently trusts incoming identity metadata |
| Strip client-supplied `x-user-id` and `x-roles` at gateway | Prevents header spoofing |
| Preserve the unique `user_id` index | Prevents conflicting consent records |
| Re-check consent immediately before provider send | Handles opt-out after a message was queued |
| Record only safe suppression reason | Provides auditability without leaking content/recipient |
| Do not automatically switch channels | Consent for one channel does not grant consent for another |
| Keep WhatsApp-like marketing denied until explicit contract exists | Missing preference field is not consent |
| Never log provider secrets, recipient details, or message body | Preference audit should not create a privacy leak |
| Review legal/product consent policy before production | Engineering enforcement does not replace legal approval |

### Task-Specific Best Practices

- Add audit history if compliance requires proof of every consent transition; the current document stores only latest state and timestamps.
- Add a policy/version field if consent wording or regional rules may change.
- Add explicit channel consent before enabling WhatsApp-like delivery.
- Keep suppression metrics and alerts separate from provider failures; suppression is an intentional outcome.
- Test every new template key against a defined purpose. Unknown templates currently fail closed.
- Ensure all future direct-send and queue-send paths use the same consent gate.

## 13. Missing or Misconfigured Things

| Audit item | Current observation | Recommendation |
|---|---|---|
| Migration history | Scripts exist, but no applied-migration tracker/runner was found | Add ordered one-time migration tracking before production |
| Task blueprint migration name | `{TASK_FILE_NAME}` references a preference migration numbered `002`, but runnable repository uses migration `006` | Treat the runnable migration filename as authoritative and update blueprint references later |
| API Gateway implementation | Public routes are specified, but gateway folder has no runnable code | Implement authenticated REST-to-gRPC proxy and strip spoofable identity headers |
| gRPC trust boundary | Handler trusts `x-user-id` and `x-roles`; no service-side auth interceptor/signature verification was found | Keep private network access and add authenticated service-to-service identity |
| Direct User Service validation | No runtime check confirms user is active/not deleted | Decide whether gateway/auth context is sufficient or add an internal status contract |
| Consent history | Only latest preference record is stored | Add append-only consent audit if policy/compliance requires it |
| Preference policy version | No policy/terms version is stored | Add version/timestamp when legal wording changes |
| Public contract mismatch risk | Proto PATCH fields are optional, while master API schema references a full preference object | Define PATCH partial-field semantics explicitly in gateway/OpenAPI contract |
| Marketing event wiring | Consent policy supports promotional/price-drop templates, but current RabbitMQ event rules expose only transactional email events | Add an approved marketing producer contract before end-to-end marketing smoke |
| WhatsApp-like preference | No public opt-in field exists | Keep fail-closed behavior or add an approved additive contract |
| Health/readiness | No dedicated endpoint reports MongoDB or preference-store readiness | Add internal readiness before orchestrated deployment |
| Official local container setup | No committed Dockerfile/Compose file was found | Add approved one-command local infrastructure and service image |
| Live integration tests | No automated MongoDB/gateway/RabbitMQ end-to-end consent test was found | Add container-backed preference, suppression, and retry re-check tests |
| Full test suite | One current-task use-case test fails because of invalid fixture timestamps | Fix fixture before considering CI green |
| Local `.env` credentials | Development-style credentials exist in a Git-ignored file | Never reuse them in shared or production environments |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Fresh Clone Setup | Same repository clone process |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Go Dependency System | Same Go version, module download, build, and generic troubleshooting |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Environment Variables / Credentials Placement | Same `.env` path, loading method, and secret handling |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | MongoDB Setup / MongoDB Migrations | Same MongoDB install, Docker, credentials, URI, and `mongosh` pattern |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | RabbitMQ Setup / Email Provider Setup With Mailpit | Same optional event/provider infrastructure |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Docker Compose Example / Start The Service | Same local containers and service run command |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | Database Setup | Same base template/delivery collections |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | Security & Best Practices | Same MongoDB schema and safe-delivery rules |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | Database Setup / New Seed Migration | Same templates evaluated by consent policy |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | Environment Variables / Security & Best Practices | Same provider setup and trusted OTP security boundary |
| `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | RabbitMQ / Manual End-To-End Event Verification | Same event broker and producer flow |
| `TaskImplementation/{SERVICE_NAME}/task6_Dependency.md` | Retry/DLQ Topology / Environment Variables | Same queued delivery worker and latest-consent re-check runtime |
| `TaskImplementation/{SERVICE_NAME}/task6_Dependency.md` | Migration Order Warning | Same risk from re-running validator migrations |

## 15. Final Checklist

- [ ] Previous dependency documentation reviewed
- [ ] Go `1.26.3` installed and existing module dependencies downloaded
- [ ] No unnecessary new dependency or external service added
- [ ] MongoDB running with valid reused credentials
- [ ] `.env` loaded into the shell
- [ ] `NOTIFICATION_MONGO_PREFERENCES_COLLECTION` is valid and distinct
- [ ] Migration `006` applied after migrations `001` through `005`
- [ ] Old validator migrations not re-run over later migrations
- [ ] Preference collection validator verified
- [ ] Unique `user_id` index verified
- [ ] Service build passes
- [ ] Known full-suite timestamp-fixture failure reviewed/fixed
- [ ] Service starts with `notification.grpc.started`
- [ ] GET returns safe defaults for a user without a saved row
- [ ] PATCH persists explicit `false` and preserves omitted fields
- [ ] Requests without trusted user/role metadata are rejected
- [ ] Public gateway route/auth propagation implemented before public exposure
- [ ] RabbitMQ/provider started only if queued suppression is being tested
- [ ] Opted-out supported transactional event becomes `suppressed` with `channel_opted_out`
- [ ] Marketing opt-out policy verified through tests or an approved future producer path
- [ ] Suppressed delivery is not retried or dead-lettered
- [ ] Latest preference is re-checked before queued retry send
- [ ] OTP trusted-flow behavior remains unchanged
- [ ] No recipient/content/secrets exposed in suppression audit or logs
- [ ] No duplicate shared setup documentation added
