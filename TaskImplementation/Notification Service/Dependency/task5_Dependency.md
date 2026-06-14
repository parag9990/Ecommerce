# Project Dependency & Setup Guide

## 1. Project Overview

### Variables Used

```text
SERVICE_NAME = "Notification Service"
TASK_FILE_NAME = "task5.md"
INPUT_FILE_PATH = TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}
OUTPUT_FILE_NAME = {TASK_FILE_NAME without ".md"}_Dependency.md
OUTPUT_FILE_PATH = TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}
```

This file documents the dependency, environment, database, RabbitMQ, and local run setup needed for the event-consumer implementation described by `{TASK_FILE_NAME}`.

Simple Hinglish summary: Order, Payment, aur User services RabbitMQ par events publish karte hain. `{SERVICE_NAME}` un events ko consume karta hai, MongoDB me duplicate-safe delivery intent store karta hai, aur configured email provider ke through notification bhejne ke liye durable attempt queue me job publish karta hai.

### Current Implementation Scope

The current repository contains the runnable implementation in:

```text
backend/services/notification-service/internal/events/
backend/services/notification-service/internal/domain/event_trigger.go
backend/services/notification-service/internal/usecase/queue_event.go
backend/services/notification-service/internal/repository/mongo_notification_repository.go
backend/services/notification-service/internal/retry/
backend/services/notification-service/internal/config/config.go
backend/services/notification-service/cmd/server/main.go
backend/services/notification-service/migrations/004_add_event_consumer_support.up.js
```

Important runtime fact: event consumption and durable delivery attempts are wired together in the current server. Enabling the event consumer also starts the retry/attempt worker. Isliye RabbitMQ URL, email provider settings, retry configuration, and a delivery encryption key must all be valid.

### Read Previous Dependency Docs First

Do not repeat the unchanged setup. Start with:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
```

The most important reused sections are:

| Previous file | Reused topic |
|---|---|
| `task1_Dependency.md` | Fresh clone, Go modules, `.env` loading, MongoDB, RabbitMQ installation, Mailpit, Docker example, service startup |
| `task2_Dependency.md` | Mongo collections, indexes, safe delivery payload, database troubleshooting |
| `task3_Dependency.md` | Template renderer and seeded order/payment templates |
| `task4_Dependency.md` | Email provider runtime settings, migration order, security patterns |

## 2. Tech Stack

| Technology | Required? | Why it is used | Beginner-friendly explanation |
|---|---:|---|---|
| Go `1.26.3` | Yes | Service runtime, consumer, validation, and worker code | Go compiled backend language hai. Same Go setup is already explained in `task1_Dependency.md`, section `Go Dependency System`. |
| RabbitMQ | Yes when event consumer is enabled | Receives business events and carries durable delivery-attempt jobs | RabbitMQ ek message broker hai. Producer aur consumer ko directly wait karne ki zarurat nahi hoti. |
| `github.com/rabbitmq/amqp091-go` `v1.10.0` | Yes | Go AMQP connection, exchange/queue declaration, consume, ack, reject, and publish APIs | Ye Go client RabbitMQ se programmatic connection banata hai. |
| MongoDB | Yes | Stores event delivery intent, idempotency key, source-event metadata, templates, and attempt state | MongoDB document database hai. Same installation/setup is reused from earlier dependency docs. |
| MongoDB Go Driver `v2.6.0` | Yes | Reads templates and stores duplicate-safe delivery records | Existing Task 2 dependency; no new install step. |
| SMTP provider or Mailpit | Yes for current event delivery flow | All supported event rules currently produce email notifications | Local testing me Mailpit safe fake inbox hai; production me secured SMTP provider use karo. |
| Go `log/slog` | Yes, standard library | Structured consumer and queue logs | Separate package install nahi chahiye. |
| Kafka | No | Architecture alternate only; no Kafka adapter exists in current implementation | RabbitMQ aur Kafka ko ek saath enable mat karo; current source RabbitMQ-specific hai. |
| Redis | No | Not used by this event-consumer implementation | Redis start karne ki zarurat nahi. |

### Supported Event Contract

| Exchange | Expected producer | Routing keys consumed | Queue |
|---|---|---|---|
| `order.events` | `order-service` | `OrderCreated`, `OrderPaid`, `OrderCancelled`, `OrderDelivered` | `notification.order.events.v1` |
| `payment.events` | `payment-service` | `PaymentSucceeded`, `PaymentFailed` | `notification.payment.events.v1` |
| `user.events` | `user-service` | `UserCreated`, `SellerApproved`, `AddressUpdated` | `notification.user.events.v1` |

All exchanges are durable RabbitMQ `topic` exchanges. All three consumer queues are durable and use manual acknowledgements.

## 3. Required Software

| Software | Required? | Setup source |
|---|---:|---|
| Git | Yes for fresh clone | `task1_Dependency.md`, section `Fresh Clone Setup` |
| Go `1.26.3` | Yes | `task1_Dependency.md`, section `Go Dependency System` |
| MongoDB server | Yes | `task1_Dependency.md`, section `MongoDB Setup` |
| `mongosh` | Yes for migrations and verification | `task1_Dependency.md`, section `MongoDB Migrations` |
| RabbitMQ server | Yes when consumer is enabled | `task1_Dependency.md`, section `RabbitMQ Setup` |
| RabbitMQ Management UI | Recommended | Inspect exchanges, queues, bindings, messages, and queue depth |
| Mailpit or a working SMTP provider | Yes for local end-to-end email verification | `task1_Dependency.md`, section `Email Provider Setup With Mailpit` |
| OpenSSL | Recommended | Generates the required 32-byte delivery encryption key |
| Docker | Optional | Easiest local way to run MongoDB, RabbitMQ, and Mailpit |

No new native OS package is required if the reused Docker setup is already working.

## 4. Dependency Management

Go module setup, `go mod download`, `go mod tidy`, version mismatch, proxy/cache troubleshooting, build, and generic test commands are already explained in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Go Dependency System
```

### Task-Specific Go Dependency

The RabbitMQ client is already present in `backend/services/notification-service/go.mod`:

```text
github.com/rabbitmq/amqp091-go v1.10.0
```

Fresh clone command:

```bash
cd backend/services/notification-service
GOWORK=off go mod download
```

Do not run `go get` just for this task. Dependency versions are already locked by `go.mod` and `go.sum`.

Task-focused tests:

```bash
cd backend/services/notification-service
GOWORK=off go test ./internal/events
GOWORK=off go test ./internal/domain -run 'TestEvent'
GOWORK=off go test ./internal/usecase -run 'TestQueueEvent'
GOWORK=off go test ./internal/config -run 'TestLoadFromEnv.*EventConsumer'
GOWORK=off go test ./internal/repository -run 'TestEventDelivery'
```

Why `GOWORK=off`? It keeps verification scoped to this module if a developer's checkout has an incomplete `backend/go.work` workspace.

## 5. Database Setup

### Reused MongoDB Setup

MongoDB installation, Docker command, connection URI, health verification, credential placement, and generic migration execution are already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: MongoDB Setup
Section: MongoDB Migrations
```

Collection design and safe payload rules are already documented in:

```text
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
Section: Database Setup
Section: Security & Best Practices
```

### New Database Change: Migration `004`

Task-specific migration:

```text
backend/services/notification-service/migrations/004_add_event_consumer_support.up.js
```

Rollback:

```text
backend/services/notification-service/migrations/004_add_event_consumer_support.down.js
```

Migration `004` adds:

| Change | Purpose |
|---|---|
| `idempotency_key` | Prevents the same event/template/channel notification from being created twice |
| `source_event_id` | Traces delivery back to producer event |
| `source_event_type` | Records the consumed event type |
| `trace_id` | Correlates logs across services |
| Partial unique index `uniq_event_notification_key` | Enforces event-notification idempotency in MongoDB |
| `welcome_user` email template | Supports `UserCreated` |
| `seller_approved` email template | Supports `SellerApproved` |
| `address_updated_security_notice` email template | Supports `AddressUpdated` |

The order and payment event rules reuse templates seeded by migration `002`.

### Run Migrations

For an incremental database that already has Tasks 1-4 migrations:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
mongosh "$NOTIFICATION_MONGO_URI" migrations/004_add_event_consumer_support.up.js
```

For a fresh database or the current full service, run every migration in order:

```bash
cd backend/services/notification-service
set -a
. ./.env
set +a
for file in migrations/*.up.js; do
  mongosh "$NOTIFICATION_MONGO_URI" "$file"
done
```

Current-runtime note: enabling the consumer also starts the durable retry worker. Migration `005` and later current-service migrations should therefore be applied in a real current-repo run. Running all `.up.js` files is the safest onboarding path.

### Verify Migration `004`

Verify the unique event index:

```bash
mongosh "$NOTIFICATION_MONGO_URI" --quiet --eval '
const d = db.getSiblingDB(process.env.NOTIFICATION_MONGO_DATABASE || "notification_db");
const name = process.env.NOTIFICATION_MONGO_DELIVERIES_COLLECTION || "notification_deliveries";
printjson(d.getCollection(name).getIndexes().filter(i => i.name === "uniq_event_notification_key"));
'
```

Expected result: one unique partial index on `idempotency_key`.

Verify user-event templates:

```bash
mongosh "$NOTIFICATION_MONGO_URI" --quiet --eval '
const d = db.getSiblingDB(process.env.NOTIFICATION_MONGO_DATABASE || "notification_db");
const name = process.env.NOTIFICATION_MONGO_TEMPLATES_COLLECTION || "notification_templates";
printjson(d.getCollection(name).find(
  { template_key: { $in: ["welcome_user", "seller_approved", "address_updated_security_notice"] }, channel: "email" },
  { _id: 1, template_key: 1, channel: 1, status: 1, version: 1 }
).sort({ template_key: 1 }).toArray());
'
```

Expected result: three active version-1 email templates.

## 6. Redis / Queue / External Services

### RabbitMQ Is Newly Mandatory For This Task

RabbitMQ installation, Docker container, credentials, persistent volume, management UI, and base health commands are already explained in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: RabbitMQ Setup
```

For previous tasks RabbitMQ could stay disabled. For the event-consumer runtime it is mandatory.

### Task-Specific RabbitMQ Topology

The service declares the following event topology during startup:

| Type | Name | Durable? | Purpose |
|---|---|---:|---|
| Topic exchange | `order.events` | Yes | Order domain events |
| Topic exchange | `payment.events` | Yes | Payment domain events |
| Topic exchange | `user.events` | Yes | User domain events |
| Queue | `notification.order.events.v1` | Yes | Order events consumed by this service |
| Queue | `notification.payment.events.v1` | Yes | Payment events consumed by this service |
| Queue | `notification.user.events.v1` | Yes | User events consumed by this service |

Current server wiring also declares retry/attempt exchanges and queues from the later durable delivery pipeline. Do not manually create them with different durability, type, or arguments; RabbitMQ returns a `PRECONDITION_FAILED` error when declarations conflict.

### Ack, Reject, And Requeue Behaviour

| Consumer result | RabbitMQ action |
|---|---|
| Valid supported event accepted for delivery | `Ack` |
| Duplicate event | `Ack` |
| Valid but unsupported event type | `Ack` |
| Malformed envelope, wrong producer, invalid payload, or permanent failure | `Reject(requeue=false)` |
| Temporary repository/broker/provider pipeline error | `Nack(requeue=true)` |

Warning: current event queues do not declare a dead-letter exchange. A rejected malformed/permanent event is discarded by RabbitMQ. This is listed again in the configuration audit.

### Verify RabbitMQ

Base Docker and health commands are reused from `task1_Dependency.md`, section `RabbitMQ Setup`.

After the service starts, open:

```text
http://localhost:15672
```

Verify:

1. Exchanges `order.events`, `payment.events`, and `user.events` exist.
2. The three `notification.*.events.v1` queues exist.
3. Each queue has the expected event-type bindings.
4. Consumers show as connected.
5. Queue depth does not continuously grow during normal local testing.

### Other External Services

| Service | Required? | Notes |
|---|---:|---|
| SMTP/Mailpit | Yes for local end-to-end event email | Setup reused from `task1_Dependency.md` |
| Kafka | No | No adapter or config exists |
| Redis | No | Not used |
| NATS/RabbitMQ alternative | No | Do not add a second broker without a separate adapter |
| Prometheus metrics HTTP | Optional | Later analytics functionality; not required to consume an event |

## 7. Environment Variables

The `.env` path, loading method, credential placement, and generic mistakes are already explained in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Environment Variables
Section: Credentials Placement
```

Create or update:

```text
backend/services/notification-service/.env
```

The Go service uses `os.LookupEnv`; it does not automatically parse `.env`. Load it before `go run`:

```bash
set -a
. ./.env
set +a
```

### New Task-Specific Variables

| Variable | Required? | Default | Purpose |
|---|---:|---|---|
| `NOTIFICATION_EVENT_CONSUMER_ENABLED` | Yes to enable this task | `false` | Starts event consumer and durable delivery worker |
| `NOTIFICATION_RABBITMQ_URL` | Yes when enabled | None | AMQP/AMQPS broker URL |
| `NOTIFICATION_ORDER_EVENTS_QUEUE` | Optional | `notification.order.events.v1` | Order-event queue name |
| `NOTIFICATION_PAYMENT_EVENTS_QUEUE` | Optional | `notification.payment.events.v1` | Payment-event queue name |
| `NOTIFICATION_USER_EVENTS_QUEUE` | Optional | `notification.user.events.v1` | User-event queue name |
| `NOTIFICATION_CONSUMER_PREFETCH` | Optional | `10` | Maximum unacknowledged event messages per channel |

Queue names must be non-empty and distinct. Prefetch must be greater than zero.

### Required Runtime Coupling

The current event rules only create email notifications, and the current server queues encrypted delivery attempts. These reused/coupled values must also be configured:

| Variable | Required when consumer enabled? | Why |
|---|---:|---|
| `NOTIFICATION_EMAIL_ENABLED=true` | Yes | Config rejects event consumer without email |
| `NOTIFICATION_EMAIL_PROVIDER` and SMTP values | Yes | Event delivery worker needs a real configured provider |
| `NOTIFICATION_DELIVERY_ENCRYPTION_KEY` | Yes | Encrypts recipient before durable storage |
| `NOTIFICATION_RETRY_*` | Defaults available | Current server starts retry/attempt worker with event consumer |
| `NOTIFICATION_MONGO_URI` | Yes | Templates, idempotency, and delivery intent |

Email and Mongo values are reused from earlier files and are not repeated in full here.

### Task-Specific `.env` Example

```env
# New event-consumer settings.
NOTIFICATION_EVENT_CONSUMER_ENABLED=true
NOTIFICATION_RABBITMQ_URL=amqp://ecommerce:ecommerce_password@localhost:5672/ecommerce
NOTIFICATION_ORDER_EVENTS_QUEUE=notification.order.events.v1
NOTIFICATION_PAYMENT_EVENTS_QUEUE=notification.payment.events.v1
NOTIFICATION_USER_EVENTS_QUEUE=notification.user.events.v1
NOTIFICATION_CONSUMER_PREFETCH=10

# Reused but mandatory with the current consumer wiring.
NOTIFICATION_EMAIL_ENABLED=true
NOTIFICATION_DELIVERY_ENCRYPTION_KEY=<base64-encoded-32-byte-key>
```

Generate a local encryption key:

```bash
openssl rand -base64 32
```

Security notes:

- Never commit the generated key, RabbitMQ password, MongoDB password, or SMTP credentials.
- Use `amqps://` and managed secrets in production.
- Rotating `NOTIFICATION_DELIVERY_ENCRYPTION_KEY` requires a plan for already queued encrypted recipients.
- Local development credentials shown above are not production-safe.

## 8. Docker Setup

No committed service-specific Docker Compose file or Dockerfile was found. Reuse:

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
| RabbitMQ container and volume | Reused setup, now required when consumer is enabled |
| Mailpit container | Reused, required for easiest local email smoke |
| New service container | None |
| New network | None committed |
| New health check | None committed |

### Ports & Networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Backend gRPC | `9090` | Existing internal API | Reused |
| MongoDB | `27017` | Templates, event idempotency, delivery state | Reused |
| RabbitMQ AMQP | `5672` | Event consume and attempt/retry publish/consume | Required for this task |
| RabbitMQ Management UI | `15672` | Inspect topology and publish local test messages | Required only for easy manual debugging |
| Mailpit SMTP | `1025` | Local event email delivery | Reused |
| Mailpit UI | `8025` | View captured email | Reused |
| Analytics HTTP | `8081` | Optional metrics/webhooks | Not required for this task |

When the Go service runs on the host, use `localhost` in Mongo/RabbitMQ/SMTP URLs. When the service runs inside the same Docker network, use container/service names such as `mongo`, `rabbitmq`, and `mailpit`.

## 9. Local Development Setup

### Step 1: Read Previous Setup

Follow the four previous dependency files listed in section 1. In particular, complete the reused Go, MongoDB, RabbitMQ, and Mailpit setup.

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

Start:

1. MongoDB.
2. RabbitMQ.
3. Mailpit or another configured SMTP provider.

Use the existing commands from `task1_Dependency.md`; they are not duplicated here.

### Step 5: Configure And Load Environment

Add the task-specific values from section 7, keep the reused Mongo/email settings valid, then:

```bash
set -a
. ./.env
set +a
```

### Step 6: Run Migrations

Recommended current-repo command:

```bash
for file in migrations/*.up.js; do
  mongosh "$NOTIFICATION_MONGO_URI" "$file"
done
```

### Step 7: Run Focused Tests

```bash
GOWORK=off go test ./internal/events
GOWORK=off go test ./internal/domain -run 'TestEvent'
GOWORK=off go test ./internal/usecase -run 'TestQueueEvent'
GOWORK=off go test ./internal/config -run 'TestLoadFromEnv.*EventConsumer'
GOWORK=off go test ./internal/repository -run 'TestEventDelivery'
```

### Step 8: Start The Service

```bash
GOWORK=off go run ./cmd/server
```

Expected startup logs include:

```text
notification.grpc.started
notification.events.rabbitmq.started
notification.retry.rabbitmq.started
```

If RabbitMQ connection, MongoDB connection, encryption key, or email configuration is invalid, startup fails before the event consumer becomes ready.

## 10. Running The Project

### Manual End-To-End Event Verification

Use RabbitMQ Management UI:

```text
http://localhost:15672
```

1. Open exchange `order.events`.
2. Choose `Publish message`.
3. Set routing key to `OrderPaid`.
4. Publish this JSON payload:

```json
{
  "event_id": "evt_local_order_paid_001",
  "event_type": "OrderPaid",
  "version": 1,
  "occurred_at": "2026-06-06T12:00:00Z",
  "producer": "order-service",
  "trace_id": "trace_local_001",
  "payload": {
    "order_id": "order_local_001",
    "user_id": "user_local_001",
    "email": "buyer@example.com",
    "name": "Riya"
  }
}
```

Expected high-level flow:

1. `notification.order.events.v1` receives the event.
2. Consumer validates envelope, producer, and payload.
3. MongoDB stores a pending delivery with an idempotency key.
4. Consumer publishes attempt `1`.
5. Delivery worker renders the reused order-status template.
6. SMTP/Mailpit accepts the email.
7. Event queue message is acknowledged.

Expected log names include:

```text
notification.event.attempt_enqueued
notification.event.handled
notification.rabbitmq.ack
```

Open Mailpit at `http://localhost:8025` and verify one order-status email.

### Verify Mongo Delivery

```bash
mongosh "$NOTIFICATION_MONGO_URI" --quiet --eval '
const d = db.getSiblingDB(process.env.NOTIFICATION_MONGO_DATABASE || "notification_db");
const name = process.env.NOTIFICATION_MONGO_DELIVERIES_COLLECTION || "notification_deliveries";
printjson(d.getCollection(name).find(
  { source_event_id: "evt_local_order_paid_001" },
  { _id: 1, status: 1, idempotency_key: 1, source_event_id: 1, source_event_type: 1, trace_id: 1, attempts: 1 }
).toArray());
'
```

Publish the exact same event again. Expected result: no second delivery document and no duplicate user email. The event should be safely handled as duplicate/idempotent work.

### Producer Contract Rules

Every published message must:

- Use JSON envelope version `1`.
- Include non-empty `event_id`, `event_type`, `producer`, `occurred_at`, and `payload`.
- Use the exact expected producer for that event family.
- Use RFC3339 timestamp format for `occurred_at`.
- Include only expected payload fields; unknown envelope or payload fields are rejected.
- Provide a valid email because current event rules are email-only.

## 11. Common Errors & Fixes

Generic Go, MongoDB, Docker, `.env`, SMTP, and credential errors are already covered in:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
Section: Common Setup Issues
```

Task-specific issues:

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `NOTIFICATION_RABBITMQ_URL must be an absolute amqp or amqps URL` | Missing URL or wrong scheme such as `http://` | Use a valid `amqp://...` local URL or `amqps://...` production URL | Validate `.env` before deployment |
| `notification event consumer requires the email channel to be enabled` | Consumer enabled while email is disabled | Configure SMTP/Mailpit and set `NOTIFICATION_EMAIL_ENABLED=true` | Treat email as a consumer prerequisite |
| `NOTIFICATION_DELIVERY_ENCRYPTION_KEY must be base64 encoding of 32 bytes` | Key missing, invalid Base64, or wrong decoded length | Generate with `openssl rand -base64 32` | Inject key from a secret manager |
| Queue names must be distinct | Two queue env vars have the same value | Restore defaults or choose three different names | Avoid unnecessary queue-name overrides |
| `NOTIFICATION_CONSUMER_PREFETCH must be greater than zero` | Prefetch is `0`, negative, or invalid | Use `10` for local default | Tune only after observing throughput |
| RabbitMQ `PRECONDITION_FAILED` during startup | Existing exchange/queue was declared with conflicting type/durability/arguments | Remove the conflicting local object or align producer/infra declarations | Share one topology contract |
| Consumer connects but no messages arrive | Producer publishes to wrong exchange/routing key/vhost | Verify exchange, routing key, vhost, and binding in Management UI | Contract-test producer routing |
| Event is rejected immediately | Invalid JSON, missing required field, wrong producer, version not `1`, unknown field, or invalid email | Compare message with section 10 example | Validate event schema in producer |
| Event is acknowledged but no notification is created | Event type is valid JSON but unsupported | Publish one of the supported routing keys | Keep producer and consumer event catalogue synchronized |
| Duplicate event does not send again | Expected idempotency behaviour | Use a new `event_id` for a genuinely new business event | Never reuse event IDs for different events |
| Event queue depth keeps growing | Consumer down, blocked, or slower than producers | Check service logs, RabbitMQ consumer count, MongoDB, SMTP, and retry queues | Alert on queue depth and consumer count |
| Rejected event disappears | Current event queue has no dead-letter exchange | Inspect producer logs and fix payload; add event DLQ before production | Do not rely on requeue for malformed messages |
| Full `go test ./...` fails in preference test | Current unrelated later-task test failure | Use focused task commands for this task and fix later-task test separately | Keep task-focused and full-suite CI jobs |
| Docker commands show daemon permission denied | User cannot access Docker daemon/socket | Start Docker and use the approved local Docker permission setup | Verify `docker ps` before onboarding |

## 12. Security & Best Practices

### Task-Specific Security Rules

| Rule | Why |
|---|---|
| Use `amqps://` in production | Protects broker credentials and event data in transit |
| Use a least-privilege RabbitMQ user/vhost | Limits exchange, queue, publish, and consume access |
| Never log raw event payloads | Payload contains email and business data |
| Keep `event_id` globally unique | Idempotency depends on stable unique event IDs |
| Encrypt durable recipient data | Current queue pipeline stores encrypted recipient ciphertext |
| Do not put secrets in event payloads | Events can be retained, inspected, retried, and copied |
| Keep Mongo unique index enabled | Application-only duplicate checks are not enough under concurrency |
| Keep event schema strict and versioned | Prevents silent contract drift |
| Monitor rejected and requeued messages | Repeated bad messages or outages otherwise stay hidden |
| Keep provider credentials outside Git | `.env` is local only; production secrets belong in a secret manager |

### Best Practices For Producers

- Publish persistent messages to durable exchanges.
- Publish only after the producer's business transaction is durable; an outbox pattern is recommended.
- Use the exact routing key and `producer` value expected by the consumer.
- Never change the meaning of version `1`; introduce a new version for breaking schema changes.
- Reuse the same `event_id` only when retrying the same event.
- Include `trace_id` so logs can be correlated across services.

## 13. Missing Or Misconfigured Things

| Audit item | Current observation | Recommendation |
|---|---|---|
| Event queue DLQ | Event queues have no dead-letter exchange; `Reject(false)` discards messages | Add a dedicated invalid-event DLQ and alerting before production |
| Broker health/readiness endpoint | No dedicated service readiness endpoint was found | Add readiness that reports MongoDB, RabbitMQ consumer, and provider readiness |
| Official local compose | No committed service-specific Compose file was found | Add an approved local compose file for one-command onboarding |
| Broker integration tests | Handler/routing unit tests exist, but no automated live-RabbitMQ integration test was found | Add container-backed topology, ack/reject, and duplicate integration tests |
| Event schema artifact | Contract is implemented in Go/JSON but no shared JSON Schema/protobuf contract was found | Publish a versioned producer-consumer contract artifact |
| Event channel choice | Current event trigger validation supports email only | Keep email enabled; add explicit rules/tests before supporting SMS/push |
| Current runtime coupling | Enabling event consumption also enables retry worker and requires encryption/retry config | Document and deploy `{TASK_FILE_NAME}` and durable delivery runtime together |
| Local `.env` | File is Git-ignored and contains development-style credentials/placeholders | Keep it local; use managed secrets in production |
| Full test suite | Current full suite has an unrelated later-task preference test failure | Fix the later-task test before treating full service CI as green |
| Docker access in current verification environment | Docker CLI exists but daemon socket access was denied | Verify containers and live topology in an environment with Docker daemon access |

## 14. References To Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Fresh Clone Setup | Same repository clone process |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Go Dependency System | Same Go module, download, build, and generic troubleshooting |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Environment Variables | Same `.env` path and shell loading method |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | MongoDB Setup | Same MongoDB install, Docker, credentials, and URI |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | MongoDB Migrations | Same `mongosh` migration execution pattern |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | RabbitMQ Setup | Same broker installation, Docker volume, credentials, ports, and UI |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Email Provider Setup With Mailpit | Same local SMTP provider used by event emails |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Docker Compose Example | Same optional local infrastructure pattern |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Start The Service | Same service startup command |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Common Setup Issues | Same generic Go/Mongo/Docker/provider errors |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | Database Setup | Same template and delivery collections |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | Security & Best Practices | Same safe delivery payload and Mongo security rules |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | Database Setup / New Seed Migration | Same renderer and order/payment templates |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | Environment Variables / Email OTP Smoke Values | Same SMTP provider configuration pattern |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | Run Required Migrations | Same current-repo recommendation to run migrations in order |

## 15. Final Checklist

- [ ] Previous dependency documentation reviewed
- [ ] Go `1.26.3` installed and module dependencies downloaded
- [ ] MongoDB running and reused Mongo settings loaded
- [ ] RabbitMQ running with correct vhost and credentials
- [ ] Mailpit/SMTP provider running
- [ ] `.env` loaded into the shell before startup
- [ ] `NOTIFICATION_EVENT_CONSUMER_ENABLED=true`
- [ ] RabbitMQ URL uses `amqp://` locally or `amqps://` in production
- [ ] Three queue names are non-empty and distinct
- [ ] Consumer prefetch is greater than zero
- [ ] Email channel/provider is enabled and valid
- [ ] Delivery encryption key is a Base64-encoded 32-byte value
- [ ] Migration `004` applied
- [ ] All current-repo migrations applied for full runtime
- [ ] `uniq_event_notification_key` index verified
- [ ] Three user-event templates verified
- [ ] Task-focused tests pass
- [ ] Service starts with event and retry consumer logs
- [ ] RabbitMQ exchanges, queues, bindings, and consumers verified
- [ ] Valid event creates one email and one Mongo delivery
- [ ] Re-publishing the same event does not create a duplicate notification
- [ ] Queue depth and rejected/requeued messages checked
- [ ] No secrets or raw payloads exposed in Git or logs
- [ ] No duplicate setup documentation added
