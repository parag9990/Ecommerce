# Project Dependency & Setup Guide

## 1. Project Overview

### Variables Used

```text
SERVICE_NAME = "Notification Service"
TASK_FILE_NAME = "task6.md"
INPUT_FILE_PATH = TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}
OUTPUT_FILE_NAME = {TASK_FILE_NAME without ".md"}_Dependency.md
OUTPUT_FILE_PATH = TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}
```

This file documents only the dependency, environment, database, RabbitMQ, and operational setup needed by the retry and dead-letter queue implementation described in `{TASK_FILE_NAME}`.

Simple Hinglish summary: provider temporarily fail ho to notification bounded delay ke baad retry hoti hai. Permanent failure ya exhausted retries ke baad metadata-only job durable DLQ me jati hai. MongoDB delivery state ko track karta hai, aur RabbitMQ attempt, delayed-retry, aur dead-letter queues handle karta hai.

### Current Implementation Scope

The runnable implementation is present in:

```text
backend/services/notification-service/internal/retry/
backend/services/notification-service/internal/domain/retry_job.go
backend/services/notification-service/internal/usecase/queue_event.go
backend/services/notification-service/internal/usecase/send_delivery_attempt.go
backend/services/notification-service/internal/repository/mongo_notification_repository.go
backend/services/notification-service/internal/security/recipient.go
backend/services/notification-service/internal/config/config.go
backend/services/notification-service/cmd/server/main.go
backend/services/notification-service/migrations/005_add_retry_dlq_support.up.js
```

Important runtime coupling:

- Retry worker ka separate enable flag nahi hai.
- `NOTIFICATION_EVENT_CONSUMER_ENABLED=true` karne par event consumers aur retry worker dono start hote hain.
- RabbitMQ, MongoDB, email provider, retry configuration, and delivery encryption key all must be valid before startup succeeds.
- Durable retry is for event-driven non-OTP notifications. OTP plaintext durable queue/DB me store ya retry nahi hota.

### Read Previous Dependency Docs First

Shared setup ko repeat nahi kiya gaya. Begin with:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
```

Most important reused sections:

| Previous file | Reused topic |
|---|---|
| `task1_Dependency.md` | Fresh clone, Go modules, `.env` loading, MongoDB/RabbitMQ/Mailpit installation, Docker example, service startup |
| `task2_Dependency.md` | MongoDB collections, delivery storage, and database security |
| `task3_Dependency.md` | Template rendering and seeded notification templates |
| `task4_Dependency.md` | Provider configuration and OTP non-persistence rules |
| `task5_Dependency.md` | Mandatory RabbitMQ event runtime, delivery encryption key, event-to-attempt flow, and broker troubleshooting |

## 2. Tech Stack

| Technology | Required? | Current-task use | Beginner-friendly explanation |
|---|---:|---|---|
| Go `1.26.3` | Yes | Retry policy, worker, queue adapter, config validation, and tests | Go compiled backend language hai. Installation and commands already documented in `task1_Dependency.md`, section `Go Dependency System`. |
| RabbitMQ | Yes when retry flow is enabled | Durable attempt queue, TTL-based delay queues, publisher confirms, and DLQ | RabbitMQ message broker hai. Is task me failed delivery ko lose hone se bachata hai aur delayed retry route karta hai. |
| `github.com/rabbitmq/amqp091-go` `v1.10.0` | Yes | Declares topology, consumes attempts, publishes persistent jobs, and waits for confirms | Go ka RabbitMQ client hai. Dependency already locked in `go.mod`; new `go get` required nahi hai. |
| MongoDB | Yes | Stores retry-safe delivery instruction, encrypted recipient, attempt state, failure code, and timestamps | MongoDB document database hai. Existing installation reuse hoti hai; migration `005` is task-specific. |
| MongoDB Go Driver `v2.6.0` | Yes | Atomic attempt claims and delivery state transitions | Existing dependency hai; separate install required nahi hai. |
| SMTP/Mailpit | Yes for current local event-delivery smoke test | Produces a real transient failure or accepted delivery for retry verification | Mailpit local fake inbox hai. Setup reused from `task1_Dependency.md`. |
| Go standard library crypto/time/json packages | Yes | Recipient protection support, metadata-only JSON jobs, durations, and timeouts | Go ke saath built-in aate hain; separate package install nahi chahiye. |
| Prometheus/Grafana | Optional | Recommended for retry backlog and DLQ alerts | Current task ke liye install mandatory nahi; production operations ke liye useful hai. |
| Redis, Kafka, NATS, RabbitMQ delayed-message plugin | No | Not used by the implementation | Delay queues RabbitMQ ke built-in TTL + dead-letter routing se work karti hain. |

## 3. Required Software

| Software | Required? | Setup source / note |
|---|---:|---|
| Git | Yes for fresh clone | `task1_Dependency.md`, section `Fresh Clone Setup` |
| Go `1.26.3` | Yes | `task1_Dependency.md`, section `Go Dependency System` |
| MongoDB server | Yes | Reused existing database |
| `mongosh` | Yes for migration and verification | Migration `005` JavaScript file run karne ke liye |
| RabbitMQ server | Yes when retry flow is enabled | Same broker introduced by the previous task |
| RabbitMQ Management UI | Strongly recommended | Retry tiers, bindings, consumers, and DLQ inspect karne ke liye |
| Mailpit or working SMTP provider | Yes for end-to-end email retry smoke | Same provider setup as previous tasks |
| OpenSSL | Recommended | Base64-encoded 32-byte encryption key generate karne ke liye |
| Docker | Optional | Existing MongoDB, RabbitMQ, and Mailpit containers run karne ka easiest local option |

No new OS service or RabbitMQ plugin is introduced. Existing RabbitMQ must include the management UI only if visual debugging is desired.

## 4. Dependency Management

Go module installation, `go mod download`, `go mod tidy`, build/run commands, version mismatch, proxy/cache issues, and generic dependency troubleshooting are already explained in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Go Dependency System
```

The task-specific external library is already present:

```text
github.com/rabbitmq/amqp091-go v1.10.0
```

Fresh clone dependency command:

```bash
cd backend/services/notification-service
GOWORK=off go mod download
```

Do not run `go get` only for this task. `go.mod` and `go.sum` already lock the required version.

Task-focused unit tests do not require MongoDB, RabbitMQ, or Mailpit:

```bash
cd backend/services/notification-service
GOWORK=off go test ./internal/retry
GOWORK=off go test ./internal/config -run 'TestLoadFromEnv.*Retry'
GOWORK=off go test ./internal/repository -run 'TestRetryDelivery'
GOWORK=off go test ./internal/usecase -run 'Test.*DeliveryAttempt|Test.*QueueEvent'
```

`GOWORK=off` keeps commands scoped to this service module and avoids unrelated workspace-module failures.

## 5. Database Setup

### Reused MongoDB Setup

MongoDB installation, Docker container, persistent volume, URI format, credentials, health verification, and generic migration execution are already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: MongoDB Setup
Section: MongoDB Migrations
```

Existing collection design and safe-payload rules are documented in:

```text
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
Section: Database Setup
Section: Security & Best Practices
```

### New Database Change: Migration `005`

Task-specific files:

```text
backend/services/notification-service/migrations/005_add_retry_dlq_support.up.js
backend/services/notification-service/migrations/005_add_retry_dlq_support.down.js
```

Migration `005` changes `notification_deliveries`:

| Change | Purpose |
|---|---|
| Adds statuses `retry_scheduled` and `dead_lettered` | Retry wait and final failure clearly represent karne ke liye |
| Adds `max_attempts` | Har delivery ka bounded attempt limit persist karta hai |
| Adds `recipient_ciphertext` | Retry ke liye recipient encrypted form me store hota hai |
| Adds `last_failure_code` | Raw provider error ke badle sanitized operational reason store karta hai |
| Adds `next_retry_at` | Expected next attempt time show karta hai |
| Adds `dead_lettered_at` | Final failure timestamp store karta hai |
| Adds `processing_attempt` and `processing_lease_until` | Duplicate workers ko same attempt concurrently send karne se reduce karta hai |
| Adds index `delivery_retry_schedule_lookup` | Scheduled retries ko status/time se inspect karne me help karta hai |
| Adds index `delivery_failed_operations_view` | Recent failed/dead-lettered operations inspect karne me help karta hai |
| Strengthens validator | OTP durable retry fields block karta hai and retry-safe event records validate karta hai |

### Migration Order Warning

> Warning: migration `005` uses MongoDB `collMod` to replace the collection validator. Isko migration `006` or `007` ke baad casually re-run mat karo, warna later preference/analytics validator rules overwrite ho sakte hain.

Use one of these flows:

| Database state | Correct action |
|---|---|
| Fresh database | Run every `.up.js` file once, filename order me |
| Database has migrations `001` through `004` only | Run migration `005`, then later migrations as required |
| Database already has `006` or `007` applied | Do not re-run `005`; verify current validator/indexes instead |

The repository has migration files but no migration-history runner/table was found. Team ko applied migration numbers externally track karne chahiye.

### Run Migration `005`

For a database known to be at migration `004`:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
mongosh "$NOTIFICATION_MONGO_URI" migrations/005_add_retry_dlq_support.up.js
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

### Verify Retry Indexes

```bash
mongosh "$NOTIFICATION_MONGO_URI" --quiet --eval '
const d = db.getSiblingDB(process.env.NOTIFICATION_MONGO_DATABASE || "notification_db");
const name = process.env.NOTIFICATION_MONGO_DELIVERIES_COLLECTION || "notification_deliveries";
printjson(d.getCollection(name).getIndexes().filter(i =>
  ["delivery_retry_schedule_lookup", "delivery_failed_operations_view"].includes(i.name)
));
'
```

Expected result: both named indexes are present.

### Verify Retry Delivery State

After a transient provider failure:

```bash
mongosh "$NOTIFICATION_MONGO_URI" --quiet --eval '
const d = db.getSiblingDB(process.env.NOTIFICATION_MONGO_DATABASE || "notification_db");
const name = process.env.NOTIFICATION_MONGO_DELIVERIES_COLLECTION || "notification_deliveries";
printjson(d.getCollection(name).find(
  { status: { $in: ["retry_scheduled", "dead_lettered"] } },
  {
    _id: 1, status: 1, attempts: 1, max_attempts: 1,
    last_failure_code: 1, next_retry_at: 1, dead_lettered_at: 1, updated_at: 1
  }
).sort({ updated_at: -1 }).limit(10).toArray());
'
```

Do not print `recipient_ciphertext` in routine debugging output.

### Rollback Caution

`005_add_retry_dlq_support.down.js` converts retry/dead-letter statuses to `failed` and removes retry-specific fields/indexes. Production rollback se retry investigation data lose ho sakta hai. Backup and drain/stop consumers before any approved rollback.

## 6. Redis / Queue / External Services

### RabbitMQ Setup Is Reused

RabbitMQ installation, Docker command, volume, credentials, vhost, ports, management UI, and base health checks are already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: RabbitMQ Setup
```

The previous task already made RabbitMQ mandatory for event consumption:

```text
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
Section: RabbitMQ Is Newly Mandatory For This Task
```

This task does not add a second broker. It adds the retry/DLQ topology on the same RabbitMQ connection and vhost.

### New Retry/DLQ Topology

The service declares topology at startup. All exchanges and queues below are durable; published jobs are persistent.

| Type | Name | Route / arguments | Purpose |
|---|---|---|---|
| Direct exchange | `notification.delivery.attempt` | Route `send` | Active delivery attempts |
| Queue | `notification.delivery.attempt.v1` | DLX `notification.delivery.dlx`, route `rejected` | Worker consumes actual provider attempts here |
| Direct exchange | `notification.delivery.retry` | Delay-specific routes | Selects the correct wait tier |
| Queue | `notification.delivery.retry.30s.v1` | TTL `30000ms`, then DLX to attempt route `send` | Default first retry wait |
| Queue | `notification.delivery.retry.2m0s.v1` | TTL `120000ms`, then DLX to attempt route `send` | Default second retry wait |
| Queue | `notification.delivery.retry.8m0s.v1` | TTL `480000ms`, then DLX to attempt route `send` | Default third retry wait |
| Direct exchange | `notification.delivery.dlx` | Routes `failed`, `invalid`, `security`, `rejected` | Final-failure routing point |
| Queue | `notification.delivery.dlq.v1` | Bound to all four final routes | Operator investigation and controlled replay |

The exact delayed queue names are generated from configured Go durations. For example, `2m` normalizes to `2m0s` in the queue name.

### Retry Behaviour

Default policy:

```text
Attempt 1 -> transient failure -> wait 30s
Attempt 2 -> transient failure -> wait 2m
Attempt 3 -> transient failure -> wait 8m
Attempt 4 -> transient failure -> DLQ
Terminal failure at any attempt -> DLQ
```

Important reliability rules implemented by the service:

- Retry job contains identifiers and sanitized metadata only; recipient/content/token/OTP fields are rejected.
- Current message is acknowledged only after retry/DLQ publish succeeds and RabbitMQ confirms it.
- Publish uses mandatory routing, so an unroutable job returns an error.
- Attempt claims use a MongoDB processing lease to reduce duplicate sends.
- Invalid or security-sensitive queue payloads are sanitized before DLQ publication.
- RabbitMQ delayed-message plugin is not required.

### Verify RabbitMQ

Base broker health commands are reused from `task1_Dependency.md`.

After service startup, open:

```text
http://localhost:15672
```

Verify:

1. Exchanges `notification.delivery.attempt`, `notification.delivery.retry`, and `notification.delivery.dlx` exist as durable `direct` exchanges.
2. Attempt queue, configured delay-tier queues, and `notification.delivery.dlq.v1` exist.
3. Attempt queue shows a consumer.
4. Delay queues have `x-message-ttl`, `x-dead-letter-exchange`, and `x-dead-letter-routing-key` arguments.
5. DLQ is bound to routes `failed`, `invalid`, `security`, and `rejected`.
6. Retry/DLQ queue depth does not grow unexpectedly.

> Warning: do not manually declare the same queue name with different arguments. RabbitMQ will close the channel with `PRECONDITION_FAILED`.

### External Service Summary

| Service | Required? | Task-specific note |
|---|---:|---|
| RabbitMQ | Yes when enabled | Same broker, new attempt/retry/DLQ topology |
| MongoDB | Yes | Same database, new migration `005` |
| SMTP/Mailpit | Yes for current email end-to-end verification | Stop/restart Mailpit carefully to simulate transient provider outage |
| Prometheus/Grafana | Optional | Add alerts for DLQ depth, old retries, and publish-confirm failures |
| Redis/Kafka/NATS | No | No adapter/config exists |

## 7. Environment Variables

The `.env` location, shell loading method, credential placement, and generic mistakes are already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Environment Variables
Section: Credentials Placement
```

Create/update the Git-ignored file:

```text
backend/services/notification-service/.env
```

The Go service reads process environment with `os.LookupEnv`; it does not automatically parse `.env`:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
```

### Task-Specific Variables

| Variable | Required? | Default | Purpose and validation |
|---|---:|---|---|
| `NOTIFICATION_RETRY_MAX_ATTEMPTS` | Optional | `4` | Total attempts including first attempt; must be `1` to `10` |
| `NOTIFICATION_RETRY_DELAY_1` | Required for retry tier 1; default available | `30s` | Wait after attempt 1 |
| `NOTIFICATION_RETRY_DELAY_2` | Required for retry tier 2; default available | `2m` | Wait after attempt 2 |
| `NOTIFICATION_RETRY_DELAY_3` | Required for retry tier 3; default available | `8m` | Wait after attempt 3 |
| `NOTIFICATION_RETRY_PREFETCH` | Optional | `10` | Max unacknowledged attempt jobs on retry consumer; must be greater than zero |
| `NOTIFICATION_RETRY_PUBLISH_CONFIRM_TIMEOUT` | Optional | `5s` | Max wait for RabbitMQ publish confirmation; must be positive |
| `NOTIFICATION_RETRY_ATTEMPT_LEASE` | Optional | `30s` | Mongo processing claim duration; must be greater than publish-confirm timeout |
| `NOTIFICATION_DELIVERY_ENCRYPTION_KEY` | Required when event consumer enabled | None | Base64 encoding of exactly 32 bytes; protects durable recipient data |

Delay rules:

- Number of delays must equal `MAX_ATTEMPTS - 1`.
- Delays must be positive and strictly increasing.
- If max attempts is greater than `4`, extra values such as `NOTIFICATION_RETRY_DELAY_4` become mandatory.
- Values use Go duration syntax such as `10s`, `2m`, or `1h`.
- Changing delays changes delayed queue names. Review and drain old delay queues before removing them.

### Incremental `.env` Example

Only retry-specific values are shown here:

```env
# Bounded retry and DLQ settings.
NOTIFICATION_RETRY_MAX_ATTEMPTS=4
NOTIFICATION_RETRY_DELAY_1=30s
NOTIFICATION_RETRY_DELAY_2=2m
NOTIFICATION_RETRY_DELAY_3=8m
NOTIFICATION_RETRY_PREFETCH=10
NOTIFICATION_RETRY_PUBLISH_CONFIRM_TIMEOUT=5s
NOTIFICATION_RETRY_ATTEMPT_LEASE=30s
NOTIFICATION_DELIVERY_ENCRYPTION_KEY=<base64-encoded-32-byte-key>
```

Generate a local key:

```bash
openssl rand -base64 32
```

Reused but mandatory activation/runtime values are documented in `task5_Dependency.md`, section `Environment Variables`:

```text
NOTIFICATION_EVENT_CONSUMER_ENABLED=true
NOTIFICATION_RABBITMQ_URL=<valid amqp:// or amqps:// URL>
NOTIFICATION_EMAIL_ENABLED=true
MongoDB and SMTP settings
```

Security notes:

- Encryption key, RabbitMQ password, MongoDB password, and provider credentials Git me commit mat karo.
- Production me `amqps://` and secret-manager injection use karo.
- Encryption key rotation requires a plan for existing encrypted recipient records.
- Retry jobs or logs me recipient, rendered body, provider secret, token, password, or OTP add mat karo.

## 8. Docker Setup

No committed service-specific `Dockerfile` or Docker Compose file was found. Reuse:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: MongoDB With Docker
Section: RabbitMQ Setup
Section: Email Provider Setup With Mailpit
Section: Docker Compose Example
```

Task-specific Docker impact:

| Docker item | Status |
|---|---|
| MongoDB container and volume | Reused |
| RabbitMQ container and volume | Reused; retry queues persist in the same broker volume |
| Mailpit container | Reused for local email acceptance/failure smoke |
| RabbitMQ delayed-message plugin | Not required |
| New container/network/volume | None |
| Service Dockerfile | Not present |
| Committed health checks/restart policies | Not present |

### Ports & Networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Backend gRPC | `9090` | Existing internal API | Reused |
| MongoDB | `27017` | Delivery state, attempts, and failure metadata | Reused |
| RabbitMQ AMQP | `5672` | Attempt, delayed retry, publish confirm, and DLQ traffic | Reused; mandatory when enabled |
| RabbitMQ Management UI | `15672` | Inspect retry topology and DLQ | Reused; recommended |
| Mailpit SMTP | `1025` | Local provider delivery/failure smoke | Reused |
| Mailpit UI | `8025` | Inspect accepted email | Reused |
| Analytics HTTP | `8081` | Later optional metrics/webhooks | No new task port |

No new port is introduced. Host-run service should use `localhost`; a future service container on the same Docker network should use service names such as `mongo`, `rabbitmq`, and `mailpit`.

Port conflict fixes and Docker networking basics are already covered in earlier dependency files.

## 9. Local Development Setup

### Step 1: Read Previous Setup

Follow the previous dependency files listed in section 1. Complete reused Go, MongoDB, RabbitMQ, Mailpit, event-consumer, and provider setup first.

### Step 2: Go To The Runnable Service

```bash
cd backend/services/notification-service
```

### Step 3: Install Existing Dependencies

```bash
GOWORK=off go mod download
```

No new `go get` command is needed.

### Step 4: Start Reused External Services

Start:

1. MongoDB.
2. RabbitMQ with the correct user/vhost.
3. Mailpit or another configured SMTP provider.

Use the commands from `task1_Dependency.md`; they are intentionally not duplicated here.

### Step 5: Add And Load Retry Configuration

Add the incremental values from section 7, keep reused Mongo/RabbitMQ/email values valid, then:

```bash
set -a
. ./.env
set +a
```

### Step 6: Run Migrations Safely

For a fresh database, run all `.up.js` files in order. For an existing database, apply only unapplied migrations and respect the migration-order warning in section 5.

### Step 7: Run Focused Tests

```bash
GOWORK=off go test ./internal/retry
GOWORK=off go test ./internal/config -run 'TestLoadFromEnv.*Retry'
GOWORK=off go test ./internal/repository -run 'TestRetryDelivery'
GOWORK=off go test ./internal/usecase -run 'Test.*DeliveryAttempt|Test.*QueueEvent'
```

### Step 8: Start The Service

```bash
GOWORK=off go run ./cmd/server
```

Expected relevant startup logs:

```text
notification.grpc.started
notification.events.rabbitmq.started
notification.retry.rabbitmq.started
```

Startup fails if MongoDB, RabbitMQ, provider config, retry policy, encryption key, or topology declaration is invalid.

## 10. Running The Project

### Happy-Path Check

Use the valid event-publish example from:

```text
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
Section: Manual End-To-End Event Verification
```

Expected result:

1. Event creates a pending MongoDB delivery.
2. Attempt `1` is published to `notification.delivery.attempt.v1`.
3. Worker claims the attempt.
4. Mailpit accepts the email.
5. MongoDB status becomes `accepted`.
6. Retry queues and DLQ remain empty.

### Transient Failure Retry Check

Use a non-production local environment only:

1. Keep MongoDB and RabbitMQ running.
2. Stop Mailpit or point SMTP to an intentionally unavailable local port.
3. Publish a new valid event with a new `event_id`.
4. Watch service logs for `notification.retry.scheduled`.
5. Verify MongoDB status becomes `retry_scheduled`.
6. Verify a message appears in the first delay queue.
7. Restore Mailpit before the next attempt.
8. Verify the later attempt becomes `accepted`.

Use short test delays only in an isolated broker/vhost. Changing retry delays creates differently named queues and can leave old queues behind.

### DLQ Check

In an isolated local environment, leave the provider unavailable through all configured attempts.

Expected result:

- Attempts stop at `NOTIFICATION_RETRY_MAX_ATTEMPTS`.
- Delivery status becomes `dead_lettered`.
- `last_failure_code` is sanitized, for example `provider_unavailable_exhausted`.
- One metadata-only job appears in `notification.delivery.dlq.v1`.
- No OTP, recipient, rendered body, token, or provider secret appears in the DLQ message.

### Operational Replay Rule

There is no implemented admin replay API or committed replay tool. Do not blindly move DLQ messages back to the attempt exchange.

Before controlled replay:

1. Fix the provider/config/content cause.
2. Check whether the notification is still useful and safe to send.
3. Verify current MongoDB delivery state and attempt count.
4. Preserve the original metadata-only contract.
5. Replay a small selected set.
6. Watch duplicate-send risk, logs, Mongo state, and DLQ depth.

## 11. Common Errors & Fixes

Generic Go, MongoDB, Docker, `.env`, RabbitMQ connection, and provider errors are already covered in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Common Setup Issues

TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
Section: Common Errors & Fixes
```

Task-specific issues:

| Error / symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `NOTIFICATION_RETRY_MAX_ATTEMPTS must be between 1 and 10` | Attempts outside supported bound | Use `1` to `10`; default is `4` | Keep bounded retry policy reviewed |
| `NOTIFICATION_RETRY_DELAY_N is required for configured max attempts` | Attempts increased without adding all delay tiers | Add one positive delay for every retry | Change attempts and delays together |
| `notification retry delays must be strictly increasing` | Later delay is equal to or shorter than earlier delay | Use increasing values such as `10s`, `1m`, `5m` | Validate config in CI |
| `NOTIFICATION_RETRY_PREFETCH must be greater than zero` | Retry prefetch is zero/negative | Use `10` locally | Tune only with queue metrics |
| `NOTIFICATION_RETRY_ATTEMPT_LEASE must exceed publish confirmation timeout` | Lease may expire while publish confirm is still waiting | Increase lease or reduce confirm timeout | Keep a clear safety margin |
| `NOTIFICATION_DELIVERY_ENCRYPTION_KEY must be base64 encoding of 32 bytes` | Missing, malformed, or wrong-length key | Generate with `openssl rand -base64 32` | Store in secret manager and validate before deploy |
| RabbitMQ `PRECONDITION_FAILED` on a retry queue | Existing queue has same name but conflicting TTL/DLX arguments | In local env, drain/remove conflicting queue or restore matching config; plan migration in production | Use versioned topology changes |
| Retry message keeps returning to attempt queue | Provider remains unavailable or processing fails before safe handoff | Check provider, MongoDB, logs, and publish confirms | Alert on old attempt/retry messages |
| DLQ grows continuously | Permanent failures, exhausted provider outage, invalid jobs, or security-blocked payloads | Inspect route/failure codes, fix cause, then controlled replay | Alert on any sustained DLQ growth |
| Invalid job goes to `security` route | Queue JSON contains prohibited content-like fields | Publish only the defined metadata-only retry job | Contract-test all producers |
| Delivery stays `processing` temporarily | Worker crashed while holding a lease | Wait for lease expiry and inspect worker outage | Monitor old processing leases |
| Migration `005` breaks later fields | Migration was re-run after later validator migrations | Restore/apply the latest validator migration and review data | Track applied migrations; never run all migrations repeatedly on an existing DB |
| Retry config changed but old delay queues remain | Queue names derive from duration values | Drain and remove old queues through an approved broker change | Treat retry-delay changes as topology migrations |

## 12. Security & Best Practices

### Task-Specific Security Rules

| Rule | Why |
|---|---|
| Keep retry job metadata-only | Broker messages can be retained, inspected, copied, and replayed |
| Never durable-retry OTP plaintext | OTP secrecy boundary must remain intact |
| Encrypt recipient data at rest | Worker needs recipient later, but plaintext persistence is unsafe |
| Never log encryption key or broker credentials | Logs often reach shared systems |
| Use `amqps://` and least-privilege broker credentials in production | Protects message data and limits broker access |
| Keep publisher confirms enabled | Retry/DLQ handoff should not silently lose work |
| Keep queues/exchanges durable and messages persistent | Broker restart should not erase pending work |
| Keep retries bounded | Infinite retries create cost, duplicate risk, and queue backlog |
| Alert on DLQ and stale processing leases | Final/stuck failures require operator action |
| Review every replay | Old transactional notifications may be harmful or misleading |

### Retry Tuning Best Practices

- Provider timeout should remain shorter than the processing lease.
- Attempt lease should exceed publish-confirm timeout and expected processing time.
- Prefetch badhane se throughput improve ho sakta hai, but concurrent provider load and duplicate-risk window bhi badhta hai.
- Backoff delays ko provider rate limits and incident recovery time ke according tune karo.
- Delay/topology changes ko deployment migration ki tarah treat karo.
- DLQ route and sanitized failure code ko dashboard/alert labels me use karo.

## 13. Missing or Misconfigured Things

| Audit item | Current observation | Recommendation |
|---|---|---|
| Migration history | Migration files exist, but no applied-migration tracker/runner was found | Add a migration runner with ordered, one-time history before production |
| Official local compose | No committed service-specific Compose file was found | Add approved one-command local infra when platform layout is finalized |
| Service health/readiness | No dedicated endpoint reports MongoDB, RabbitMQ worker, or provider readiness | Add readiness covering broker channel/consumer and database access |
| Live RabbitMQ integration tests | Retry logic has unit tests, but no automated container-backed topology/TTL/DLQ integration test was found | Add CI integration tests with isolated RabbitMQ/MongoDB |
| DLQ replay tooling | No admin replay API or committed controlled replay command exists | Build authenticated, audited, selective replay tooling |
| DLQ alerts/dashboard | Task design recommends alerts, but no committed alert rules/dashboard were found | Alert on DLQ depth, oldest retry age, and publish-confirm failures |
| Retry topology migration process | Queue names/arguments depend on retry delays, but no cleanup/versioning automation exists | Document drain/create/remove procedure for config changes |
| Runtime feature coupling | One event-consumer flag starts both event consumer and retry worker | Consider separate explicit worker controls/readiness if operations need independent scaling |
| Broker reconnect | A closed RabbitMQ stream causes service failure; automatic in-process reconnect loop was not found | Use restart policy/orchestrator and consider bounded reconnect logic |
| Full service container | No Dockerfile was found | Add a production build/runtime image before container deployment |
| Local `.env` | File is correctly Git-ignored | Keep secrets local and inject production values from a secret manager |
| Hardcoded credentials | No retry credentials are hardcoded in source; local examples use development credentials | Never reuse example credentials in shared/production environments |
| Full test suite | Task-focused retry tests pass, but the current full suite has an unrelated preference patch test failure | Fix `TestManagePreferencePatchPreservesExplicitFalseAndOmittedFields` before treating full service CI as green |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Fresh Clone Setup | Same repository clone process |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Go Dependency System | Same Go install, module download, build, and generic troubleshooting |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Environment Variables / Credentials Placement | Same `.env` path, loading method, and secret placement |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | MongoDB Setup / MongoDB Migrations | Same server install, Docker, URI, credentials, and `mongosh` pattern |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | RabbitMQ Setup | Same broker installation, Docker volume, user/vhost, ports, and UI |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Email Provider Setup With Mailpit | Same local SMTP provider |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Docker Compose Example | Same optional local infrastructure |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Start The Service / Common Setup Issues | Same run command and generic failures |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | Database Setup | Same delivery collection base |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | Security & Best Practices | Same MongoDB and payload-safety rules |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | Database Setup / New Seed Migration | Same templates rendered during retry |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | Environment Variables / Security & Best Practices | Same provider setup and OTP secrecy rule |
| `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | RabbitMQ / Environment Variables | Same mandatory broker activation, event queues, encryption-key prerequisite, and event-to-attempt handoff |
| `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | Manual End-To-End Event Verification | Same valid event used to create the first delivery attempt |

## 15. Final Checklist

- [ ] Previous dependency documentation reviewed
- [ ] Go `1.26.3` installed and existing module dependencies downloaded
- [ ] MongoDB, RabbitMQ, and Mailpit/SMTP running
- [ ] Existing event-consumer, MongoDB, and email settings valid
- [ ] Retry max attempts is between `1` and `10`
- [ ] Exactly one strictly increasing delay exists for each retry tier
- [ ] Retry prefetch and publish-confirm timeout are positive
- [ ] Attempt lease exceeds publish-confirm timeout
- [ ] Delivery encryption key is Base64 encoding of exactly 32 bytes
- [ ] `.env` loaded into the shell and remains Git-ignored
- [ ] Migration `005` applied in the correct order
- [ ] Migration `005` not re-run over later validator migrations
- [ ] Retry indexes verified
- [ ] Task-focused tests pass
- [ ] Service starts with `notification.retry.rabbitmq.started`
- [ ] Attempt, retry, and dead-letter exchanges verified
- [ ] Attempt, delay-tier, and DLQ queues/arguments verified
- [ ] Happy-path delivery becomes `accepted`
- [ ] Transient provider failure becomes `retry_scheduled`
- [ ] Exhausted/terminal failure becomes `dead_lettered`
- [ ] DLQ message contains metadata only
- [ ] OTP is never persisted or published for durable retry
- [ ] DLQ/retry backlog and old processing leases monitored
- [ ] Replay is controlled, selective, and audited
- [ ] No duplicate shared setup documentation added
