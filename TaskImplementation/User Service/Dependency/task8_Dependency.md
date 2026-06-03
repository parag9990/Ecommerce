# Project Dependency & Setup Guide

## Variables Used In This Guide

```text
SERVICE_NAME = User Service
TASK_FILE_NAME = task8.md
INPUT_FILE_PATH = TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}
OUTPUT_FILE_NAME = task8_Dependency.md
OUTPUT_FILE_PATH = TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}
SERVICE_CODE_DIR = backend/services/user-service
MIGRATION_DIR = {SERVICE_CODE_DIR}/migrations
PREVIOUS_DEPENDENCY_DIR = TaskImplementation/{SERVICE_NAME}/Dependency
```

Beginner note: Is document me base setup repeat nahi kiya gaya hai. Go, MySQL, Docker MySQL, gRPC, `.env` loading, validation, audit metadata, and generic troubleshooting already previous dependency docs me covered hain. Yahan focus sirf `TASK_FILE_NAME` ke new event/outbox setup par hai.

## 1. Project Overview

This document is the dependency and setup companion for `INPUT_FILE_PATH`.

Current task ka goal important user-domain writes ko reliable events me convert karna hai:

| Event | Trigger in current implementation | Why it matters |
|---|---|---|
| `UserCreated` | `CreateUser` success ke baad | Notification, analytics, recommendation ko signup signal milta hai. |
| `SellerApproved` | Seller status `draft` or `pending_review` se `active` hota hai | Product/CMS, notification, analytics, recommendation ko seller approval signal milta hai. |
| `AddressUpdated` | Address create, update, delete, or default-address change | Downstream projections ko address state change ka signal milta hai. |

Simple Hinglish:

Event ka matlab hai "kuch important ho chuka hai". Current service pehle MySQL me apna write save karta hai, same transaction me `user_outbox_events` row banata hai, and optional worker later RabbitMQ/Kafka par publish karta hai. Agar MQ down ho, event DB me safe pending rahega.

Task-specific implementation files checked:

| File | Why checked |
|---|---|
| `INPUT_FILE_PATH` | Event contracts, outbox pattern, RabbitMQ/Kafka scope |
| `{SERVICE_CODE_DIR}/go.mod` | New MQ dependencies already present |
| `{SERVICE_CODE_DIR}/internal/config/config.go` | Event, outbox worker, RabbitMQ, Kafka env variables |
| `{SERVICE_CODE_DIR}/cmd/server/main.go` | Outbox recorder, unit of work, worker startup |
| `{SERVICE_CODE_DIR}/internal/domain/user_events.go` | Event payload structs and address change types |
| `{SERVICE_CODE_DIR}/internal/events/*.go` | Envelope, recorder, worker, RabbitMQ/Kafka publishers |
| `{SERVICE_CODE_DIR}/internal/repository/mysql_outbox_repository.go` | Outbox insert, lock, mark published/failed queries |
| `{SERVICE_CODE_DIR}/internal/repository/mysql_unit_of_work.go` | Same DB transaction for business write + outbox insert |
| `{SERVICE_CODE_DIR}/internal/usecase/event_helpers.go` | Payload building, PII hashing, event trigger helpers |
| `{MIGRATION_DIR}/003_create_user_outbox_events.up.sql` | New outbox table migration |
| `{MIGRATION_DIR}/003_create_user_outbox_events.down.sql` | Outbox table rollback |

Important current-code difference from `TASK_FILE_NAME` guide:

| Topic | Task guide wording | Current repo implementation |
|---|---|---|
| Outbox migration number | Example says `008_create_user_outbox_events` | Actual migration is `003_create_user_outbox_events` |
| Event id dependency | Guide suggests `github.com/google/uuid` | Current code uses Go standard library `crypto/rand` + `encoding/hex`; no UUID package needed |
| Config key | Guide mentions `USER_EVENTS_MODE` | Current code does not read `USER_EVENTS_MODE`; do not add it |

## 2. Tech Stack

Base stack is reused:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md`

Sections:
`1. Project Tech Stack Analysis`
`2. Language-Specific Dependency System: Go`
`3. Database Analysis`
`5. External Services Analysis`
```

`TASK_FILE_NAME`-specific stack additions:

| Technology / Library | Required? | What it is | Current task usage |
|---|---:|---|---|
| Outbox pattern | Yes | DB table me pending event store karne ka reliable pattern. | Business write and event enqueue same MySQL transaction me hote hain. |
| MySQL `JSON` column | Yes | MySQL column type jo JSON payload store karta hai. | Event envelope `payload` column me store hota hai. |
| `github.com/rabbitmq/amqp091-go` | Conditional | RabbitMQ Go client. | Default provider `rabbitmq` choose karne par worker publish karta hai. |
| `github.com/segmentio/kafka-go` | Conditional | Kafka Go client. | Alternate provider `kafka` choose karne par worker publish karta hai. |
| RabbitMQ | Conditional | Message broker. Fanout exchange se events consumers tak ja sakte hain. | Default local/event provider when `OUTBOX_WORKER_ENABLED=true`. |
| Kafka | Conditional | High-throughput event streaming platform. | Alternate provider for analytics/replay-heavy event pipelines. |
| Go `crypto/rand` | Yes | Secure random bytes generator. | `evt_...` event ids generate karta hai. |
| Go `encoding/json` | Yes | JSON marshal/unmarshal package. | Event envelope ko DB/MQ payload banata hai. |
| gRPC metadata | Reused | Request/trace metadata carry karta hai. | `x-request-id`, `x-trace-id`, and `traceparent` event envelope me propagate hote hain. |

No Redis, MongoDB, SMTP, Stripe, Firebase, MinIO, Typesense, or Kubernetes dependency is introduced by `TASK_FILE_NAME`.

## 3. Required Software

Base software setup is reused:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md`

Sections:
`3. Database Analysis`
`5. External Services Analysis`
`7. Docker and DevOps Setup`
`8. Project Run Instructions`
```

`TASK_FILE_NAME`-specific requirements:

| Software | Required? | Why |
|---|---:|---|
| MySQL 8.x | Yes | Existing service DB plus new `user_outbox_events` table. MySQL 8 is important for `JSON`, `TIMESTAMP(6)`, and `FOR UPDATE SKIP LOCKED`. |
| MySQL client or migration runner | Yes | Migration `003_create_user_outbox_events.up.sql` apply karne ke liye. |
| RabbitMQ server | Conditional | Required only when `OUTBOX_WORKER_ENABLED=true` and `USER_EVENTS_PROVIDER=rabbitmq`. |
| Kafka broker | Conditional | Required only when `OUTBOX_WORKER_ENABLED=true` and `USER_EVENTS_PROVIDER=kafka`. |
| Docker | Optional | RabbitMQ local container and reused MySQL container ke liye useful. |
| RabbitMQ Management UI | Optional | Local RabbitMQ exchange inspect karne ke liye. |

Runtime mode decision:

| Mode | Required services | Recommended for |
|---|---|---|
| Events disabled | MySQL only | Debugging older tasks or isolating DB/gRPC issues. |
| Outbox enqueue only | MySQL with migration `003` | Verify events are recorded without MQ setup. |
| Full publish with RabbitMQ | MySQL + RabbitMQ | Default `TASK_FILE_NAME` local event publishing. |
| Full publish with Kafka | MySQL + Kafka | Alternate high-throughput/replay workflow. |

## 4. Dependency Management

Go dependency setup is already explained:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md`

Section:
`2. Language-Specific Dependency System: Go`
```

Current `go.mod` already contains the `TASK_FILE_NAME` dependencies:

| Dependency | Status | Purpose |
|---|---|---|
| `github.com/rabbitmq/amqp091-go v1.11.0` | Already present | RabbitMQ publisher |
| `github.com/segmentio/kafka-go v0.4.51` | Already present | Kafka publisher |
| `github.com/go-sql-driver/mysql v1.9.3` | Reused | MySQL/outbox repository |
| `google.golang.org/grpc` | Reused | Metadata and service runtime |

Fresh clone command:

```bash
SERVICE_CODE_DIR=backend/services/user-service
cd "$SERVICE_CODE_DIR"
go mod download
```

Do not run these for current code unless dependencies were removed:

```bash
go get github.com/rabbitmq/amqp091-go
go get github.com/segmentio/kafka-go
```

`github.com/google/uuid` is not required by current implementation. Event ids are generated with standard Go packages, so adding a UUID dependency would be unnecessary setup churn.

Use `go mod tidy` only if Go reports module errors or dependency files actually changed:

```bash
SERVICE_CODE_DIR=backend/services/user-service
cd "$SERVICE_CODE_DIR"
go mod tidy
```

## 5. Database Setup

### Reused MySQL Setup

MySQL installation, local Docker MySQL, DSN format, base credentials, and generic DB troubleshooting are already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md`

Sections:
`3. Database Analysis`
`7. Docker and DevOps Setup`
`9. Common Errors and Fixes`
```

Base schema and migration order are already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/task2_Dependency.md`

Section:
`5. Database Setup`
```

Audit migration prerequisite is already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/task7_Dependency.md`

Section:
`5. Database Setup`
```

### `TASK_FILE_NAME` New Migration

`TASK_FILE_NAME` adds this new table:

```text
{MIGRATION_DIR}/003_create_user_outbox_events.up.sql
{MIGRATION_DIR}/003_create_user_outbox_events.down.sql
```

Migration order for full current service:

| Order | File | Required? | Purpose |
|---:|---|---:|---|
| 1 | `001_create_user_tables.up.sql` | Yes | Base users, addresses, seller profiles, KYC tables |
| 2 | `002_add_user_audit_fields.up.sql` | Yes | Audit fields used by event payloads |
| 3 | `003_create_user_outbox_events.up.sql` | Yes for events | Durable event outbox table |

Important:

- `USER_EVENTS_ENABLED` defaults to `true`.
- If migration `003` is missing and a write triggers an event, inserts can fail with `Table 'user_db.user_outbox_events' doesn't exist`.
- For `TASK_FILE_NAME`, prefer applying migration `003` instead of disabling events.

### Outbox Table

| Column | Purpose |
|---|---|
| `event_id` | Unique event id, used for dedupe and debugging |
| `event_type` | `UserCreated`, `SellerApproved`, or `AddressUpdated` |
| `event_version` | Schema version; current value is `1` |
| `topic` | Default `user.events` |
| `aggregate_type` | `user`, `seller`, or `address` |
| `aggregate_id` | Main entity id for ordering/debugging |
| `payload` | Full JSON envelope sent to MQ |
| `request_id`, `trace_id` | Correlation metadata |
| `status` | `pending`, `published`, `failed`, or `dead` |
| `attempts` | Publish retry count |
| `next_attempt_at` | Retry/backoff schedule |
| `locked_until` | Multi-worker duplicate pickup protection |
| `last_error` | Last publish failure message |
| `occurred_at` | Business event time in UTC |
| `published_at` | MQ publish success time |

### Apply Migrations

Fresh local DB:

```bash
SERVICE_CODE_DIR=backend/services/user-service
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_ADMIN_USER=root

mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_ADMIN_USER" -p \
  < "$SERVICE_CODE_DIR/migrations/001_create_user_tables.up.sql"

mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_ADMIN_USER" -p \
  < "$SERVICE_CODE_DIR/migrations/002_add_user_audit_fields.up.sql"

mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_ADMIN_USER" -p \
  < "$SERVICE_CODE_DIR/migrations/003_create_user_outbox_events.up.sql"
```

Existing DB where `001` and `002` are already applied:

```bash
SERVICE_CODE_DIR=backend/services/user-service
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_ADMIN_USER=root

mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_ADMIN_USER" -p \
  < "$SERVICE_CODE_DIR/migrations/003_create_user_outbox_events.up.sql"
```

### Verify Migration

```bash
mysql -h 127.0.0.1 -P 3306 -u ecommerce_user -p user_db \
  -e "SHOW TABLES LIKE 'user_outbox_events';"

mysql -h 127.0.0.1 -P 3306 -u ecommerce_user -p user_db \
  -e "DESCRIBE user_outbox_events;"
```

After triggering a write:

```bash
mysql -h 127.0.0.1 -P 3306 -u ecommerce_user -p user_db \
  -e "SELECT event_id, event_type, topic, status, attempts, created_at FROM user_outbox_events ORDER BY id DESC LIMIT 10;"
```

Expected:

| Runtime mode | Expected status |
|---|---|
| Outbox enqueue only, worker off | Rows remain `pending` |
| Worker on and MQ reachable | Rows become `published` |
| Worker on and MQ down | Rows become `failed`, retry later |
| Max attempts reached | Rows become `dead`, DLQ publish attempted |

## 6. Redis / Queue / External Services

### Redis

Redis is not used by `TASK_FILE_NAME`.

### RabbitMQ

RabbitMQ ek message broker hai. Simple words me: producer exchange me message publish karta hai, consumer queues us exchange se bind hoke message receive karte hain.

Current implementation:

| Item | Value |
|---|---|
| Provider name | `rabbitmq` |
| Default provider | Yes |
| Main exchange | `user.events` |
| DLQ exchange | `user.events.dlq` |
| Exchange type | `fanout` |
| Message content type | `application/json` |
| Delivery mode | Persistent |
| Required env when worker enabled | `RABBITMQ_URL` |

Local RabbitMQ with Docker:

```bash
docker volume create ecommerce_rabbitmq_data

docker run -d --name ecommerce-rabbitmq \
  -e RABBITMQ_DEFAULT_USER=ecommerce \
  -e RABBITMQ_DEFAULT_PASS=change_me \
  -p 5672:5672 \
  -p 15672:15672 \
  -v ecommerce_rabbitmq_data:/var/lib/rabbitmq \
  rabbitmq:3-management
```

Verify:

```bash
docker logs ecommerce-rabbitmq
docker exec ecommerce-rabbitmq rabbitmqctl status
docker exec ecommerce-rabbitmq rabbitmqctl list_exchanges name type durable
```

Management UI:

```text
URL: http://localhost:15672
Username: ecommerce
Password: change_me
```

Beginner note:

The service declares exchanges, not consumer queues. Agar UI me queue empty/not visible dikhe, iska matlab publisher broken zaruri nahi hai. Consumer service ko apni queue create karke `user.events` exchange se bind karni hogi.

### Kafka

Kafka high-throughput event streaming ke liye alternate provider hai. Is repo me Kafka publisher code exists, but checked-in Kafka Docker Compose ya topic creation script nahi mila.

Current implementation:

| Item | Value |
|---|---|
| Provider name | `kafka` |
| Required env when worker enabled | `KAFKA_BROKERS` |
| Main topic | `user.events` |
| DLQ topic | `user.events.dlq` |
| Message key | `aggregate_id` from event envelope |
| Header | `content-type=application/json` |

Kafka local rule:

- Use the team/platform Kafka setup if available.
- If using Docker, run any Kafka-compatible broker reachable at `127.0.0.1:9092`.
- Pre-create `user.events` and `user.events.dlq` in production.
- Do not enable `USER_EVENTS_PROVIDER=kafka` until `KAFKA_BROKERS` points to a reachable broker.

Verification commands if Kafka CLI is available:

```bash
kafka-topics --bootstrap-server 127.0.0.1:9092 --list
kafka-console-consumer --bootstrap-server 127.0.0.1:9092 --topic user.events --from-beginning
```

## 7. Environment Variables

### Reused Base `.env`

Base `.env` path, loading process, DB DSN, gRPC address, reflection, DB pool, validation phone region, and logging setup are reused:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md`

Section:
`4. Environment Variables`
```

Audit/validation-related env notes are reused:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/task6_Dependency.md`
`TaskImplementation/{SERVICE_NAME}/Dependency/task7_Dependency.md`
```

Important: current code does not automatically load `.env`. Source it before `go run`:

```bash
SERVICE_CODE_DIR=backend/services/user-service
cd "$SERVICE_CODE_DIR"
set -a
. ./.env
set +a
```

### `TASK_FILE_NAME` New `.env` Delta

Add only these event/outbox values to `{SERVICE_CODE_DIR}/.env`.

Outbox enqueue enabled, worker disabled:

```env
# TASK_FILE_NAME event/outbox config.
USER_EVENTS_ENABLED=true
USER_EVENTS_TOPIC=user.events
USER_EVENTS_DLQ=user.events.dlq
USER_EVENTS_PROVIDER=rabbitmq

# Keep worker false when you only want DB outbox rows.
OUTBOX_WORKER_ENABLED=false
OUTBOX_WORKER_BATCH_SIZE=100
OUTBOX_WORKER_POLL_INTERVAL=2s
OUTBOX_WORKER_LOCK_TTL=30s
OUTBOX_MAX_ATTEMPTS=10
OUTBOX_PUBLISH_TIMEOUT=10s
```

Full RabbitMQ publishing:

```env
USER_EVENTS_ENABLED=true
USER_EVENTS_TOPIC=user.events
USER_EVENTS_DLQ=user.events.dlq
USER_EVENTS_PROVIDER=rabbitmq
OUTBOX_WORKER_ENABLED=true
RABBITMQ_URL=amqp://ecommerce:change_me@127.0.0.1:5672/
```

Full Kafka publishing:

```env
USER_EVENTS_ENABLED=true
USER_EVENTS_TOPIC=user.events
USER_EVENTS_DLQ=user.events.dlq
USER_EVENTS_PROVIDER=kafka
OUTBOX_WORKER_ENABLED=true
KAFKA_BROKERS=127.0.0.1:9092
```

Do not add this key for current code:

```env
# Not read by current config.go.
USER_EVENTS_MODE=outbox
```

### Variable Table

| Variable | Required? | Default | Purpose | Security note |
|---|---:|---|---|---|
| `USER_EVENTS_ENABLED` | No | `true` | Enables outbox row creation for write events. | Not secret; decide intentionally per environment. |
| `USER_EVENTS_PROVIDER` | Conditional | `rabbitmq` | Selects publisher when worker is enabled. Allowed: `rabbitmq`, `kafka`. | Not secret. |
| `USER_EVENTS_TOPIC` | Required when events enabled | `user.events` | Main event exchange/topic. | Not secret. Keep stable for consumers. |
| `USER_EVENTS_DLQ` | No | `user.events.dlq` | Dead-letter exchange/topic used after max attempts. | Not secret. Monitor it. |
| `OUTBOX_WORKER_ENABLED` | No | `false` | Starts background publisher loop. | Not secret. Enable only with MQ ready. |
| `OUTBOX_WORKER_BATCH_SIZE` | No | `100` | Number of rows locked per worker poll. | Not secret. Avoid huge values locally. |
| `OUTBOX_WORKER_POLL_INTERVAL` | No | `2s` | Worker poll frequency. | Not secret. Too low can create DB noise. |
| `OUTBOX_WORKER_LOCK_TTL` | No | `30s` | Row lock expiry for stuck workers. | Not secret. Keep above publish timeout. |
| `OUTBOX_MAX_ATTEMPTS` | No | `10` | Retry limit before marking dead. | Not secret. |
| `OUTBOX_PUBLISH_TIMEOUT` | No | `10s` | Per-message publish timeout. | Not secret. |
| `RABBITMQ_URL` | Required for RabbitMQ worker | none | AMQP connection string. | Contains username/password; do not commit real secrets. |
| `KAFKA_BROKERS` | Required for Kafka worker | none | Comma-separated broker list. | Usually not secret, but keep internal. |

## 8. Docker Setup

### Reused Docker Setup

MySQL Docker setup is already explained:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md`

Sections:
`3. Database Analysis -> E. Docker Setup`
`7. Docker and DevOps Setup`
```

### `TASK_FILE_NAME` Docker Changes

| Docker item | Status |
|---|---|
| New `{SERVICE_NAME}` Dockerfile | Not found |
| Project-level Docker Compose | Not found |
| New MySQL volume | Not needed beyond reused MySQL setup |
| New RabbitMQ container | Conditional, needed for RabbitMQ publishing |
| New Kafka container | Conditional, only if choosing Kafka provider |
| New Docker network | Not checked in |
| Health checks | No checked-in app/MQ healthcheck config |

Recommended beginner local setup:

1. Run MySQL using previous docs.
2. Apply migrations `001`, `002`, `003`.
3. Start RabbitMQ with Docker only if `OUTBOX_WORKER_ENABLED=true`.
4. Run `{SERVICE_NAME}` directly with `go run ./cmd/server`.

If everything runs directly on host, use:

```env
RABBITMQ_URL=amqp://ecommerce:change_me@127.0.0.1:5672/
KAFKA_BROKERS=127.0.0.1:9092
```

If later a compose network is added and the app runs inside a container, use service DNS:

```env
RABBITMQ_URL=amqp://ecommerce:change_me@rabbitmq:5672/
KAFKA_BROKERS=kafka:9092
```

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation

Follow these first:

| Order | File | Reuse |
|---:|---|---|
| 1 | `TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md` | Go, MySQL, Docker MySQL, base `.env`, ports, grpcurl |
| 2 | `TaskImplementation/{SERVICE_NAME}/Dependency/task2_Dependency.md` | Base MySQL schema and migration |
| 3 | `TaskImplementation/{SERVICE_NAME}/Dependency/task4_Dependency.md` | gRPC startup and metadata |
| 4 | `TaskImplementation/{SERVICE_NAME}/Dependency/task6_Dependency.md` | Validation env and test setup |
| 5 | `TaskImplementation/{SERVICE_NAME}/Dependency/task7_Dependency.md` | Audit migration and actor metadata |

### Step 2: Go To Project Directory

```bash
SERVICE_CODE_DIR=backend/services/user-service
cd "$SERVICE_CODE_DIR"
```

### Step 3: Install Only Current Dependencies

No new `go get` required. Download existing modules:

```bash
go mod download
```

### Step 4: Setup Databases And Services

Use previous MySQL setup, then apply `TASK_FILE_NAME` migration:

```bash
cd <repo-root>
SERVICE_CODE_DIR=backend/services/user-service

mysql -h 127.0.0.1 -P 3306 -u root -p \
  < "$SERVICE_CODE_DIR/migrations/003_create_user_outbox_events.up.sql"
```

Start RabbitMQ only if full publishing is needed:

```bash
docker run -d --name ecommerce-rabbitmq \
  -e RABBITMQ_DEFAULT_USER=ecommerce \
  -e RABBITMQ_DEFAULT_PASS=change_me \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3-management
```

### Step 5: Add New Or Changed Environment Variables

For outbox-only verification:

```bash
cd backend/services/user-service
set -a
. ./.env
set +a

export USER_EVENTS_ENABLED=true
export USER_EVENTS_TOPIC=user.events
export USER_EVENTS_DLQ=user.events.dlq
export OUTBOX_WORKER_ENABLED=false
```

For RabbitMQ publishing:

```bash
export USER_EVENTS_PROVIDER=rabbitmq
export OUTBOX_WORKER_ENABLED=true
export RABBITMQ_URL='amqp://ecommerce:change_me@127.0.0.1:5672/'
```

### Step 6: Run Migrations If Needed

Full current runtime needs:

```text
001_create_user_tables.up.sql
002_add_user_audit_fields.up.sql
003_create_user_outbox_events.up.sql
```

If you only applied `001` and `002` earlier, apply `003` now.

### Step 7: Start Backend Service

```bash
SERVICE_CODE_DIR=backend/services/user-service
cd "$SERVICE_CODE_DIR"
go run ./cmd/server
```

Expected logs:

```text
user_service_starting
user_service_grpc_listening
```

If worker is enabled, also expect:

```text
user_outbox_worker_started
```

### Step 8: Verify Functionality Related To `TASK_FILE_NAME`

Run tests:

```bash
SERVICE_CODE_DIR=backend/services/user-service
cd "$SERVICE_CODE_DIR"
go test ./internal/events ./internal/repository ./internal/usecase
```

Verify outbox table:

```bash
mysql -h 127.0.0.1 -P 3306 -u ecommerce_user -p user_db \
  -e "SELECT event_type, status, COUNT(*) total FROM user_outbox_events GROUP BY event_type, status;"
```

Trigger a write through existing gRPC/Gateway flow, then check:

```bash
mysql -h 127.0.0.1 -P 3306 -u ecommerce_user -p user_db \
  -e "SELECT event_id, event_type, aggregate_type, aggregate_id, status FROM user_outbox_events ORDER BY id DESC LIMIT 5;"
```

## 10. Running the Project

Ports and networking:

| Service | Port | Purpose | Status |
|---|---:|---|---|
| `{SERVICE_NAME}` gRPC | `50052` | Internal profile/address/seller APIs | Reused |
| MySQL | `3306` | User DB and outbox table | Reused |
| API Gateway HTTP | `8080` | REST-to-gRPC path if using Gateway | Reused |
| RabbitMQ AMQP | `5672` | Event publish connection | New conditional |
| RabbitMQ Management UI | `15672` | Local RabbitMQ inspection | New optional |
| Kafka | `9092` | Alternate event broker | New conditional |

Change ports:

| Need | Change |
|---|---|
| gRPC port conflict | `USER_SERVICE_GRPC_ADDRESS=:50053` |
| MySQL host/port conflict | Update `USER_SERVICE_DATABASE_DSN` |
| RabbitMQ host/port conflict | Update `RABBITMQ_URL` |
| Kafka broker port conflict | Update `KAFKA_BROKERS` |

Recommended local run modes:

| Mode | Env | Result |
|---|---|---|
| Disable events | `USER_EVENTS_ENABLED=false` | No outbox rows created |
| Enqueue only | `USER_EVENTS_ENABLED=true`, `OUTBOX_WORKER_ENABLED=false` | Events stay `pending` in DB |
| Publish RabbitMQ | `USER_EVENTS_PROVIDER=rabbitmq`, `OUTBOX_WORKER_ENABLED=true`, `RABBITMQ_URL=...` | Worker publishes to RabbitMQ exchange |
| Publish Kafka | `USER_EVENTS_PROVIDER=kafka`, `OUTBOX_WORKER_ENABLED=true`, `KAFKA_BROKERS=...` | Worker publishes to Kafka topic |

## 11. Common Errors & Fixes

Generic Go/MySQL/Docker/gRPC errors are already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md`
`TaskImplementation/{SERVICE_NAME}/Dependency/task2_Dependency.md`
`TaskImplementation/{SERVICE_NAME}/Dependency/task4_Dependency.md`
```

`TASK_FILE_NAME`-specific errors:

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `Table 'user_db.user_outbox_events' doesn't exist` | Migration `003` missing while `USER_EVENTS_ENABLED=true` | Apply `{MIGRATION_DIR}/003_create_user_outbox_events.up.sql` or temporarily set `USER_EVENTS_ENABLED=false` | Apply migrations in order before write testing |
| `USER_EVENTS_TOPIC is required when events are enabled` | Topic env set blank | Set `USER_EVENTS_TOPIC=user.events` | Do not blank default env values |
| `OUTBOX_WORKER_ENABLED requires USER_EVENTS_ENABLED` | Worker on but event recording off | Set `USER_EVENTS_ENABLED=true` or worker false | Keep worker/env mode consistent |
| `RABBITMQ_URL is required when RabbitMQ outbox worker is enabled` | Worker enabled with default provider but no URL | Start RabbitMQ and set `RABBITMQ_URL` | Keep worker off until broker config is ready |
| `KAFKA_BROKERS is required when Kafka outbox worker is enabled` | Provider is Kafka but brokers missing | Set `KAFKA_BROKERS=127.0.0.1:9092` or real brokers | Pre-create Kafka config per env |
| `unsupported USER_EVENTS_PROVIDER` | Provider typo like `rabbit` | Use `rabbitmq` or `kafka` | Copy allowed values from this guide |
| `dial rabbitmq: connection refused` | RabbitMQ stopped or URL points to wrong host/port | Start RabbitMQ; verify port `5672`; fix URL | Check `docker ps` and `rabbitmqctl status` |
| Events stay `pending` | Worker disabled or stopped | Set `OUTBOX_WORKER_ENABLED=true` and restart service | Decide whether you want enqueue-only or full publish mode |
| Events repeatedly become `failed` | Broker reachable issue, bad credentials, topic/exchange issue | Check service logs, MQ logs, URL/brokers, permissions | Add MQ smoke check before enabling worker |
| Events become `dead` | Publish failed until max attempts | Inspect `last_error`, DLQ, broker config, then replay manually if needed | Monitor `failed/dead` counts |
| `FOR UPDATE SKIP LOCKED` SQL error | MySQL too old or incompatible | Use MySQL 8-compatible server | Keep local DB version aligned with repo |
| `USER_EVENTS_MODE=outbox` seems ignored | Current config does not read this key | Remove it; use `USER_EVENTS_ENABLED` and `OUTBOX_WORKER_ENABLED` | Document only env vars from `config.go` |
| RabbitMQ UI shows exchange but no queue messages | Publisher declares fanout exchange; no consumer queue bound | Create/bind consumer queue in consumer service | Consumers own their queue declarations |

## 12. Security & Best Practices

Reuse base security notes:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md`

Sections:
`10. Security and Configuration Audit`
`11. Best Practices`
```

`TASK_FILE_NAME`-specific best practices:

- Do not publish raw password, OTP, JWT, refresh token, or auth headers in events.
- Avoid raw email/phone/full address in broad `user.events`; current code hashes email/phone for `UserCreated`.
- Keep `RABBITMQ_URL` and broker passwords out of commits.
- Use separate MQ credentials for producer and consumers.
- Give `{SERVICE_NAME}` publish permission only; consumers should not publish to `user.events`.
- Keep consumers idempotent by `event_id`, because MQ delivery can be at least once.
- Monitor pending, failed, and dead outbox rows.
- Keep `USER_EVENTS_TOPIC` stable. Renaming topic/exchange breaks consumers.
- Use TLS/private networking for production RabbitMQ/Kafka.
- Pre-create Kafka topics and configure retention/replication in production.
- Keep RabbitMQ queues durable and bind consumer queues explicitly.
- Do not expose `{SERVICE_NAME}` gRPC or MQ ports to untrusted public traffic.

## 13. Missing or Misconfigured Things

| Finding | Current status | Risk | Suggested fix |
|---|---|---|---|
| No checked-in project Docker Compose | Not found | Beginners must manually start MySQL/RabbitMQ/Kafka | Add local compose with MySQL, RabbitMQ or Kafka, and healthchecks |
| No `{SERVICE_NAME}` Dockerfile | Not found | App cannot run as container without extra work | Add service Dockerfile in future DevOps task |
| No `.env.example` found | Not found | New env vars can be missed | Add sanitized `.env.example` with event/outbox settings |
| Local `.env` lacks `TASK_FILE_NAME` event vars | Current `.env` only has DB/gRPC/logging | Defaults work partly, but worker publishing needs explicit config | Add `TASK_FILE_NAME` delta from section `7` |
| Migration runner not standardized | Raw SQL only | Manual re-runs and version confusion | Use `golang-migrate`, `goose`, or platform migration runner |
| Task guide migration number differs | Guide says `008`, repo has `003` | Beginner may search wrong file | Use actual `{MIGRATION_DIR}/003_create_user_outbox_events.up.sql` |
| No RabbitMQ queue binding manifest | Publisher declares exchanges only | Consumers may not receive anything until they bind queues | Consumer services should declare durable queues and bindings |
| No Kafka topic creation script | Not found | Kafka startup can fail if auto-create disabled | Add topic provisioning script/infra |
| No MQ health/readiness check in service | Not found | Worker issues appear only in logs/failed rows | Add health/metrics later |
| No metrics exporter found for outbox | Not found | Harder to alert on stuck events | Add counters/gauges for pending/failed/dead rows |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md` | `1. Project Tech Stack Analysis` | Base Go/MySQL/gRPC stack already explained |
| `TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md` | `2. Language-Specific Dependency System: Go` | Go modules, `go.mod`, `go.sum`, `go.work`, commands, cache/proxy issues already covered |
| `TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md` | `3. Database Analysis` | MySQL install, Docker MySQL, DSN, credentials, default ports already documented |
| `TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md` | `4. Environment Variables` | Base `.env` loading and DB/gRPC/logging env already documented |
| `TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md` | `7. Docker and DevOps Setup` | Reused local MySQL Docker setup and missing app Dockerfile caveat |
| `TaskImplementation/{SERVICE_NAME}/Dependency/task1_Dependency.md` | `9. Common Errors and Fixes` | Generic Docker/MySQL/Go/gRPC errors already covered |
| `TaskImplementation/{SERVICE_NAME}/Dependency/task2_Dependency.md` | `5. Database Setup` | Base schema and migration order reused before outbox migration |
| `TaskImplementation/{SERVICE_NAME}/Dependency/task2_Dependency.md` | `11. Common Errors & Fixes` | Migration/table/schema troubleshooting reused |
| `TaskImplementation/{SERVICE_NAME}/Dependency/task3_Dependency.md` | `7. Database Setup Required For Task 3` | Repository layer depends on migrated MySQL schema |
| `TaskImplementation/{SERVICE_NAME}/Dependency/task4_Dependency.md` | `4. Dependency Management` | gRPC/proto/runtime dependency flow reused |
| `TaskImplementation/{SERVICE_NAME}/Dependency/task4_Dependency.md` | `7. Environment Variables` | gRPC metadata and service startup caveats reused |
| `TaskImplementation/{SERVICE_NAME}/Dependency/task6_Dependency.md` | `7. Environment Variables` | Validation phone region and shared validation setup reused |
| `TaskImplementation/{SERVICE_NAME}/Dependency/task7_Dependency.md` | `5. Database Setup` | Audit migration `002` is prerequisite before `TASK_FILE_NAME` |
| `TaskImplementation/{SERVICE_NAME}/Dependency/task7_Dependency.md` | `7. Environment Variables` | Prior event-disable caveat now becomes full `TASK_FILE_NAME` event setup |

## 15. Final Checklist

- [ ] Previous dependency documentation checked before using this guide.
- [ ] No duplicate Go/MySQL/Docker setup copied unnecessarily.
- [ ] `go mod download` completed in `{SERVICE_CODE_DIR}` for fresh clone.
- [ ] MySQL is running and reachable through `USER_SERVICE_DATABASE_DSN`.
- [ ] Migrations `001`, `002`, and `003` applied in order for full current runtime.
- [ ] `user_outbox_events` table exists in `user_db`.
- [ ] `USER_EVENTS_ENABLED` intentionally set for the current run mode.
- [ ] `OUTBOX_WORKER_ENABLED=false` for outbox-only verification.
- [ ] RabbitMQ running and `RABBITMQ_URL` set if RabbitMQ publish mode is enabled.
- [ ] Kafka broker reachable and `KAFKA_BROKERS` set if Kafka publish mode is enabled.
- [ ] `USER_EVENTS_PROVIDER` is either `rabbitmq` or `kafka`.
- [ ] `USER_EVENTS_MODE` not used because current code does not read it.
- [ ] Backend service starts and logs `user_service_grpc_listening`.
- [ ] If worker enabled, logs show `user_outbox_worker_started`.
- [ ] Task-specific writes create `UserCreated`, `SellerApproved`, or `AddressUpdated` outbox rows.
- [ ] Published/failed/dead event rows checked in MySQL.
- [ ] RabbitMQ exchange or Kafka topic verified if full publish mode is enabled.
- [ ] No real MQ/database passwords committed to git.
- [ ] Missing compose, `.env.example`, health checks, and topic/queue provisioning gaps noted for future DevOps work.
