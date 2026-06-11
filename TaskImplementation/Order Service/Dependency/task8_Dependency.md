# Project Dependency & Setup Guide

## Variables Used In This Document

```text
SERVICE_NAME=Order Service
TASK_FILE_NAME=task8.md
INPUT_FILE_PATH=TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}
OUTPUT_FILE_NAME=${TASK_FILE_NAME without ".md"}_Dependency.md
OUTPUT_FILE_PATH=TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}
SERVICE_CODE_PATH=backend/services/order-service
TASK_2_BASE_MIGRATION=${SERVICE_CODE_PATH}/migrations/001_create_order_tables.up.sql
TASK_4_PAYMENT_MIGRATION=${SERVICE_CODE_PATH}/migrations/002_add_payment_coordination.up.sql
TASK_5_QUERY_INDEX_MIGRATION=${SERVICE_CODE_PATH}/migrations/003_add_grpc_query_indexes.up.sql
TASK_8_OUTBOX_MIGRATION=${SERVICE_CODE_PATH}/migrations/004_create_order_outbox_events.up.sql
PREVIOUS_DEPENDENCY_FILE_TASK_1=TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
PREVIOUS_DEPENDENCY_FILE_TASK_2=TaskImplementation/${SERVICE_NAME}/task2_Dependency.md
PREVIOUS_DEPENDENCY_FILE_TASK_3=TaskImplementation/${SERVICE_NAME}/task3_Dependency.md
PREVIOUS_DEPENDENCY_FILE_TASK_4=TaskImplementation/${SERVICE_NAME}/task4_Dependency.md
PREVIOUS_DEPENDENCY_FILE_TASK_5=TaskImplementation/${SERVICE_NAME}/task5_Dependency.md
PREVIOUS_DEPENDENCY_FILE_TASK_6=TaskImplementation/${SERVICE_NAME}/task6_Dependency.md
PREVIOUS_DEPENDENCY_FILE_TASK_7=TaskImplementation/${SERVICE_NAME}/task7_Dependency.md
```

## Previous Dependency Reuse

Is guide ka goal duplicate setup docs banana nahi hai. Common Go install, MySQL install, Docker Compose basics, Kafka local setup, `.env` loading, gRPC port, and generic troubleshooting pehle dependency guides me already explained hain.

Read these first:

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Project Tech Stack Analysis, Go Dependency System, complete `.env`, MySQL/Kafka/Docker setup, ports, generic errors | Same module and same base local infrastructure |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` | MySQL base schema and migration order | `${TASK_FILE_NAME}` outbox FK depends on `order_status_history` from base schema |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_3}` | Cart-to-order runtime dependencies | `OrderCreated` is emitted from the same checkout/order creation flow |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_4}` | Payment coordination and captured-result outbox caveat | `OrderPaid` is created from verified payment result transition |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_5}` | gRPC metadata, auth boundary, port `9094` | Event-producing commands still enter through the same service boundary when server exists |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_6}` | Seller fulfillment and delivered-event caveat | `OrderDelivered` can be created when fulfillment aggregates to delivered |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_7}` | Idempotent checkout and current migration alignment | Duplicate checkout protection reduces duplicate `OrderCreated` business facts |

Important: Is file me base installation steps repeat nahi kiye gaye. Sirf `${TASK_FILE_NAME}` ke new or task-specific setup points explain kiye gaye hain.

## 1. Project Overview

`${INPUT_FILE_PATH}` ka focus order lifecycle events emit karna hai:

| Event | Trigger |
|---|---|
| `OrderCreated` | Order creation transaction successfully commit hoti hai |
| `OrderPaid` | Trusted payment result `paid` transition commit karta hai |
| `OrderCancelled` | Valid cancellation transition commit hoti hai |
| `OrderDelivered` | Parent order final delivered state commit hota hai |

Current code reality:

| Area | Current Implementation |
|---|---|
| Event names, envelope, payload builders | `${SERVICE_CODE_PATH}/internal/orderevents/event.go` |
| Outbox domain model | `${SERVICE_CODE_PATH}/internal/domain/outbox.go` |
| MySQL outbox table | `${TASK_8_OUTBOX_MIGRATION}` |
| Order create + outbox same transaction | `${SERVICE_CODE_PATH}/internal/repository/mysql_order_repository.go` |
| Status transition + outbox same transaction | `${SERVICE_CODE_PATH}/internal/repository/mysql_order_repository.go` |
| Seller fulfillment delivered outbox path | `${SERVICE_CODE_PATH}/internal/repository/mysql_seller_order_repository.go` |
| Outbox polling/locking repository | `${SERVICE_CODE_PATH}/internal/repository/mysql_outbox_repository.go` |
| Kafka publisher | `${SERVICE_CODE_PATH}/internal/events/kafka_publisher.go` |
| Outbox worker retry/dead-letter loop | `${SERVICE_CODE_PATH}/internal/events/outbox_worker.go` |
| Event/outbox env validation | `${SERVICE_CODE_PATH}/internal/config/config.go` |

Beginner note: Outbox ka simple meaning hai, "pehle DB me event safely save karo, phir background worker queue me publish karega." Isse DB commit successful but Kafka publish lost wali dual-write problem avoid hoti hai.

## 2. Tech Stack

### Reused Technologies

Refer:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Sections:
Project Tech Stack Analysis
Go Dependency System
External Services Analysis -> Kafka
Docker and DevOps Setup
```

| Technology | Status For `${TASK_FILE_NAME}` | Setup Impact |
|---|---|---|
| Go | Reused | Same `${SERVICE_CODE_PATH}/go.mod` module |
| Go modules | Reused | Dependencies already declared in `go.mod` |
| MySQL 8+ / InnoDB | Mandatory for real runtime | `order_outbox_events` table stores durable events |
| Kafka | Mandatory only when event publishing is enabled | Worker publishes pending outbox rows to `order.events` |
| `github.com/segmentio/kafka-go` | Already present in `go.mod` | Kafka producer adapter uses it |
| gRPC / Protobuf | Reused | Existing command/API boundary when server bootstrap exists |
| Docker | Optional local infra | Reuse previous MySQL/Kafka container docs |

### Task-Specific Concepts

| Concept | Required? | Why Used | Beginner-Friendly Explanation |
|---|---:|---|---|
| Transactional outbox | Yes | Order state and event row commit together | Agar order save hua to event bhi DB me safe hai. Kafka down ho to bhi baad me publish ho sakta hai. |
| At-least-once delivery | Yes | Worker retries failed publish attempts | Same event duplicate aa sakta hai, so consumer `event_id` se dedupe karega. |
| Kafka message key = `order_id` | Yes for Kafka path | Same order ke events same partition me ordered rahen | Ek order ka created, paid, delivered order random partitions me scatter nahi hoga. |
| Deduplication key | Yes | Same business transition ka duplicate row prevent karta hai | Repeated callback/retry se same fact ka second outbox row block hota hai. |
| Dead-letter status | Yes | Max retry ke baad manual action visible hota hai | Failed event chupke delete nahi hota, ops team inspect/replay kar sakti hai. |
| RabbitMQ adapter | No in current code | Only documented alternative | Current implementation Kafka client use karti hai. RabbitMQ choose karna ho to alag adapter and dependency add karni hogi. |
| Redis | No | Not used for this task | Outbox correctness MySQL transaction and locks se aati hai. |

## 3. Required Software

| Use Case | Required Software | Notes |
|---|---|---|
| Sirf `${TASK_FILE_NAME}` read karna | Markdown viewer/editor | Runtime setup nahi chahiye |
| Unit tests for event/builders/worker/repository mocks | Go matching `${SERVICE_CODE_PATH}/go.mod` | Live MySQL/Kafka usually not needed |
| Real DB outbox verification | MySQL 8+ and MySQL client | Migration `001` to `004` apply honi chahiye |
| Event publishing verification | Kafka broker and Kafka CLI tools | Only if `ORDER_EVENTS_ENABLED=true` |
| Local infra through containers | Docker | Reuse previous Docker setup |
| Long-running backend service | Go, MySQL, Kafka if enabled, and server bootstrap | Current repo me no `${SERVICE_CODE_PATH}/cmd/server/main.go` found |

Installation steps for Go, MySQL, Docker, and Kafka are already documented in:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
```

## 4. Dependency Management

`${SERVICE_CODE_PATH}` is a Go module. Dependency system same hai as previous tasks.

| Dependency | Present? | Purpose For `${TASK_FILE_NAME}` | Action Needed |
|---|---:|---|---|
| `github.com/segmentio/kafka-go` | Yes | Kafka writer for `order.events` | No `go get` needed |
| `github.com/go-sql-driver/mysql` | Yes | MySQL outbox table read/write | No new action |
| `github.com/DATA-DOG/go-sqlmock` | Yes | Repository tests without live DB | No new action |
| `google.golang.org/grpc` | Yes | Existing service boundary | Reused |
| `google.golang.org/protobuf` | Yes | Existing generated contracts | Reused |
| `github.com/rabbitmq/amqp091-go` | No | Only if platform switches to RabbitMQ | Do not install for current Kafka path |

Reuse dependency commands from:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Section:
Go Dependency System
```

Task-specific useful commands:

```bash
cd ${SERVICE_CODE_PATH}
env GOWORK=off go mod download
env GOWORK=off go test ./internal/domain ./internal/orderevents ./internal/events ./internal/repository ./internal/usecase
```

Beginner note: `go.mod` already includes Kafka client. Agar `go get github.com/segmentio/kafka-go` dobara run karoge to version churn ho sakta hai, isliye avoid karo unless upgrade intentionally required ho.

## 5. Database Setup

### Database Used: MySQL 8+

Full MySQL install, Docker MySQL, credentials, and base schema setup are already documented in:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Sections:
Database Analysis
Docker and DevOps Setup

${PREVIOUS_DEPENDENCY_FILE_TASK_2}
Sections:
Database Analysis
Migration Order
```

### What Is New For `${TASK_FILE_NAME}`

`${TASK_FILE_NAME}` needs the durable outbox table:

```text
${TASK_8_OUTBOX_MIGRATION}
```

Important correction for beginners: `${INPUT_FILE_PATH}` contains implementation-ready example names like `002_create_order_outbox_events`, but current workspace ka real migration file `004_create_order_outbox_events.up.sql` hai. Real local setup me `${TASK_8_OUTBOX_MIGRATION}` use karo.

### Outbox Table Purpose

| Table | Required? | Purpose |
|---|---:|---|
| `order_outbox_events` | Yes for real event-producing runtime | Pending, processing, published, failed, and dead-letter event rows store karta hai |
| `order_status_history` | Yes | Outbox event `source_history_id` FK is table ko reference karta hai |
| `orders` | Yes | Aggregate source for order lifecycle facts |

### Key Columns

| Column | Why Important |
|---|---|
| `event_id` | Consumer dedupe identity |
| `event_type` | `OrderCreated`, `OrderPaid`, `OrderCancelled`, `OrderDelivered` |
| `aggregate_id` | `order_id`, also Kafka ordering key |
| `routing_key` | Broker-friendly route like `order.created` |
| `deduplication_key` | Same business transition duplicate event prevent karta hai |
| `source_history_id` | Event ko status history audit row se link karta hai |
| `payload` | Complete JSON envelope to publish |
| `status`, `attempts`, `next_attempt_at` | Worker retry state |
| `locked_at`, `locked_by` | Multiple workers same row publish na karein |
| `last_error` | Debugging ke liye latest safe failure text |

### Migration Order

Docs-only review ke liye DB migration zaroori nahi. Real DB verification or full current runtime ke liye repository root se migrations order me apply karo:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 < backend/services/order-service/migrations/001_create_order_tables.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/002_add_payment_coordination.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/003_add_grpc_query_indexes.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/004_create_order_outbox_events.up.sql
```

### Verify Outbox Schema

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW TABLES LIKE 'order_outbox_events';"
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SHOW INDEX FROM order_outbox_events;"
```

Expected beginner check:

| Check | Why |
|---|---|
| Table exists | Event-producing repository paths will not fail |
| `uk_order_outbox_deduplication` exists | Duplicate business facts blocked |
| FK to `order_status_history` exists | Event audit traceability available |
| `idx_order_outbox_pending` exists | Worker can find due rows efficiently |

### Credentials Placement

No new DB credentials are introduced by `${TASK_FILE_NAME}`.

Use the existing variable:

```text
ORDER_MYSQL_DSN
```

Refer:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Section:
Environment Variables
```

Security note: DSN me DB password hota hai. `.env`, screenshots, logs, shell history, and Git commits me leak mat karo.

## 6. Redis / Queue / External Services

### Kafka

Kafka current implemented broker path hai.

| Item | Value |
|---|---|
| Topic | `order.events` |
| Producer | `order-service` in event envelope |
| Kafka message key | `aggregate_id` / `order_id` |
| Delivery style | At-least-once |
| Go adapter | `${SERVICE_CODE_PATH}/internal/events/kafka_publisher.go` |

Kafka installation and local Docker setup are reused from:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Sections:
External Services Analysis -> Kafka
Docker and DevOps Setup
```

Task-specific topic verification, only when Kafka is running:

```bash
docker exec -it order-kafka kafka-topics.sh --bootstrap-server localhost:9092 --create --if-not-exists --topic order.events --partitions 3 --replication-factor 1
docker exec -it order-kafka kafka-topics.sh --bootstrap-server localhost:9092 --describe --topic order.events
```

If your Kafka container name is different, replace `order-kafka` with the actual container name.

### RabbitMQ

RabbitMQ is allowed as an architecture alternative in `${INPUT_FILE_PATH}`, but current code does not implement RabbitMQ.

| Question | Answer |
|---|---|
| Is RabbitMQ required now? | No |
| Is `amqp091-go` in `go.mod`? | No |
| Should beginners install RabbitMQ for current setup? | No |
| When to add it? | Only if platform officially switches from Kafka to RabbitMQ |

### Redis, NATS, MinIO, Elasticsearch, SMTP, Stripe, Twilio, Firebase, OAuth

`${TASK_FILE_NAME}` introduces none of these services. Do not start or configure them only for order event emission.

### Docker

Docker remains optional but useful for MySQL and Kafka local dependencies. No new checked-in Dockerfile, Docker Compose service, volume, network, health check, or restart policy was added by `${TASK_FILE_NAME}`.

## 7. Environment Variables

Full `.env` setup is already documented in:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Section:
Environment Variables
```

`${TASK_FILE_NAME}` does not introduce brand-new variable names beyond the event/outbox variables already documented earlier. It makes these variables task-critical:

| Variable | New? | Required? | Default / Example | Task-Specific Purpose |
|---|---:|---:|---|---|
| `ORDER_EVENTS_ENABLED` | No | Optional, default `true` | `true` or `false` | Controls Kafka/outbox publishing setup validation |
| `ORDER_EVENTS_TOPIC` | No | Required if events enabled | `order.events` | Topic worker publishes to |
| `ORDER_KAFKA_BROKERS` | No | Required if events enabled | `localhost:9092` | Kafka broker addresses |
| `ORDER_KAFKA_WRITE_TIMEOUT` | No | Optional | `10s` | Kafka write timeout |
| `ORDER_OUTBOX_BATCH_SIZE` | No | Optional | `100` | Rows worker claims per poll |
| `ORDER_OUTBOX_INTERVAL` | No | Optional | `1s` | Worker polling interval |
| `ORDER_OUTBOX_MAX_ATTEMPTS` | No | Optional | `8` | Retry count before dead-letter |
| `ORDER_OUTBOX_INITIAL_BACKOFF` | No | Optional | `5s` | First retry delay |
| `ORDER_OUTBOX_MAX_BACKOFF` | No | Optional | `10m` | Retry backoff cap |
| `ORDER_OUTBOX_STALE_LOCK_TIMEOUT` | No | Optional | `5m` | Reclaim stuck `processing` rows after worker crash |

Task-specific local event-only delta:

```bash
export ORDER_EVENTS_ENABLED=true
export ORDER_EVENTS_TOPIC=order.events
export ORDER_KAFKA_BROKERS=localhost:9092
export ORDER_KAFKA_WRITE_TIMEOUT=10s
export ORDER_OUTBOX_BATCH_SIZE=100
export ORDER_OUTBOX_INTERVAL=1s
export ORDER_OUTBOX_MAX_ATTEMPTS=8
export ORDER_OUTBOX_INITIAL_BACKOFF=5s
export ORDER_OUTBOX_MAX_BACKOFF=10m
export ORDER_OUTBOX_STALE_LOCK_TIMEOUT=5m
```

Quick local setup without Kafka publishing:

```bash
export ORDER_EVENTS_ENABLED=false
```

Warning: `ORDER_EVENTS_ENABLED=false` avoids Kafka broker validation, but it does not remove the DB table need. Event-producing repository paths can still insert into `order_outbox_events`, so migration `004` is still needed for real DB flows.

How `.env` loads:

```text
Current config uses os.Getenv.
`.env` file is not auto-loaded by the current config package.
```

Refer:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Section:
How environment loading works
```

## 8. Docker Setup

`${TASK_FILE_NAME}` adds no new Docker assets.

Reuse previous Docker setup:

| Container / Service | Status | Reference |
|---|---|---|
| MySQL | Reused | `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` and `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` |
| Kafka | Reused when events enabled | `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` |
| Backend service container | Not available yet | No Dockerfile and no `cmd/server/main.go` currently found |
| RabbitMQ | Not used | No setup required |
| Redis | Not used | No setup required |

### Ports And Networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| MySQL | `3306` | `order_db`, order tables, outbox table | Reused |
| Kafka broker | `9092` | Publish `order.events` when enabled | Reused |
| Kafka controller | `9093` | Local KRaft controller in earlier Docker example | Reused |
| gRPC API | `9094` | Existing service API port when server bootstrap exists | Reused |
| RabbitMQ AMQP | `5672` | Not used by current implementation | Not used |
| Redis | `6379` | Not used by `${TASK_FILE_NAME}` | Not used |

No new host port is introduced by the outbox worker. Worker is a background goroutine/process component, not a separate HTTP/gRPC server.

Docker networking reminder is already explained in previous docs:

```text
Host machine usually uses localhost:3306 and localhost:9092.
Container-to-container usually uses mysql:3306 and kafka:9092.
```

Task-specific port change rule:

| Port Change | Where To Change |
|---|---|
| MySQL host port changed | Update host/port inside `ORDER_MYSQL_DSN` |
| Kafka host port changed | Update `ORDER_KAFKA_BROKERS` |
| gRPC port changed after server bootstrap exists | Update `ORDER_GRPC_ADDR` |

Detailed firewall, Docker network, and generic port-conflict debugging is reused from `${PREVIOUS_DEPENDENCY_FILE_TASK_1}`.

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation First

Follow these before this guide:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
${PREVIOUS_DEPENDENCY_FILE_TASK_2}
${PREVIOUS_DEPENDENCY_FILE_TASK_4}
${PREVIOUS_DEPENDENCY_FILE_TASK_6}
${PREVIOUS_DEPENDENCY_FILE_TASK_7}
```

They cover base install, DB setup, Kafka setup, payment/fulfillment/idempotency caveats, and generic troubleshooting.

Clone repository flow is already documented in:

```text
${PREVIOUS_DEPENDENCY_FILE_TASK_1}
Section:
Project Run Instructions -> Step 1: Clone repository
```

### Step 2: Go To Project Directory

```bash
cd ${SERVICE_CODE_PATH}
```

### Step 3: Install Only New Dependencies If Needed

No new dependency install is required because Kafka client is already in `go.mod`.

```bash
env GOWORK=off go mod download
```

### Step 4: Setup Only New Databases/Services If Needed

Docs-only or unit-test flow:

```text
No MySQL or Kafka service required.
```

Real outbox DB verification:

```text
Start MySQL and apply migrations 001 to 004.
```

Event publish verification:

```text
Start Kafka and create/verify topic order.events.
```

### Step 5: Add Only New Or Changed Environment Variables

No brand-new env names. For event-enabled runtime, ensure these existing variables are configured:

```bash
export ORDER_EVENTS_ENABLED=true
export ORDER_EVENTS_TOPIC=order.events
export ORDER_KAFKA_BROKERS=localhost:9092
```

For local no-Kafka package work:

```bash
export ORDER_EVENTS_ENABLED=false
```

### Step 6: Run Migrations If Needed

For current real runtime, repository root se:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 < backend/services/order-service/migrations/001_create_order_tables.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/002_add_payment_coordination.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/003_add_grpc_query_indexes.up.sql
mysql -u root -p -h 127.0.0.1 -P 3306 order_db < backend/services/order-service/migrations/004_create_order_outbox_events.up.sql
```

### Step 7: Start Backend Service

Current limitation:

```text
${SERVICE_CODE_PATH}/cmd/server/main.go is not present.
```

So beginners should not expect this to work today:

```bash
go run ./cmd/server
```

When bootstrap is added, startup must wire:

| Component | Why |
|---|---|
| Config loader | Reads `ORDER_EVENTS_*`, `ORDER_KAFKA_BROKERS`, and outbox worker settings |
| MySQL connection | Business repositories and outbox repository use same DB |
| Kafka publisher | Sends envelopes to `order.events` |
| Outbox worker | Polls pending/failed rows and marks published/failed/dead-letter |
| Graceful shutdown | Closes Kafka writer and stops worker cleanly |

### Step 8: Verify APIs Or Functionality Related To `${TASK_FILE_NAME}`

Unit/package verification:

```bash
cd ${SERVICE_CODE_PATH}
env GOWORK=off go test ./internal/domain ./internal/orderevents ./internal/events ./internal/repository ./internal/usecase
```

DB verification:

```bash
mysql -u root -p -h 127.0.0.1 -P 3306 order_db -e "SELECT status, COUNT(*) FROM order_outbox_events GROUP BY status;"
```

Kafka verification, only after worker publishing is wired and running:

```bash
docker exec -it order-kafka kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic order.events --from-beginning
```

Expected event behavior:

| Scenario | Expected Outbox/Event Result |
|---|---|
| New checkout creates order | One pending `OrderCreated` row |
| Verified captured payment | One pending `OrderPaid` row |
| Valid cancellation | One pending `OrderCancelled` row |
| Final delivered parent order | One pending `OrderDelivered` row |
| Duplicate callback with no state change | No new outbox row |
| Worker publish success | Row becomes `published` |
| Broker temporary failure | Row becomes `failed` with `next_attempt_at` |
| Attempts exhausted | Row becomes `dead_letter` |

## 10. Running The Project

### Documentation-Only Flow

```bash
sed -n '1,120p' "${INPUT_FILE_PATH}"
sed -n '1,120p' "${OUTPUT_FILE_PATH}"
```

### Task-Specific Test Flow

```bash
cd ${SERVICE_CODE_PATH}
env GOWORK=off go test ./internal/orderevents ./internal/events
```

### Repository/Usecase Test Flow

```bash
cd ${SERVICE_CODE_PATH}
env GOWORK=off go test ./internal/repository ./internal/usecase
```

### Real DB Flow

1. Start MySQL using previous dependency guide.
2. Apply migrations `001`, `002`, `003`, and `004` in order.
3. Set `ORDER_MYSQL_DSN`.
4. Use repository/integration wiring or future server bootstrap to create lifecycle transitions.
5. Check `order_outbox_events`.

### Event-Enabled Flow

1. Start Kafka using previous dependency guide.
2. Create or verify topic `order.events`.
3. Set `ORDER_EVENTS_ENABLED=true`.
4. Set `ORDER_KAFKA_BROKERS=localhost:9092`.
5. Start future service bootstrap with outbox worker wired.
6. Consume `order.events` and verify JSON envelope.

## 11. Common Errors & Fixes

Generic Go/MySQL/Docker/Kafka/`.env` issues are already documented in previous dependency files. This section lists only `${TASK_FILE_NAME}`-specific or newly important errors.

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `Table 'order_outbox_events' doesn't exist` | Migration `004` missing | Apply `${TASK_8_OUTBOX_MIGRATION}` | Apply full current migration set before real runtime |
| `Cannot add foreign key constraint` during migration `004` | Base `order_status_history` table missing | Apply `${TASK_2_BASE_MIGRATION}` first | Always run migrations in order |
| `ORDER_KAFKA_BROKERS is required when order events are enabled` | `ORDER_EVENTS_ENABLED` defaults true and broker list empty | Set `ORDER_KAFKA_BROKERS` or `ORDER_EVENTS_ENABLED=false` locally | Keep local `.env` explicit |
| `ORDER_EVENTS_TOPIC is required when order events are enabled` | Topic env set blank | Use `order.events` | Do not blank out default topic |
| `ORDER_OUTBOX_BATCH_SIZE must be greater than zero` | Invalid numeric env | Use positive integer like `100` | Avoid `0`, negative, or non-numeric values |
| `ORDER_OUTBOX_MAX_BACKOFF must be greater than or equal to ORDER_OUTBOX_INITIAL_BACKOFF` | Bad retry env ordering | Increase max backoff or reduce initial backoff | Keep default `5s` and `10m` unless tuning |
| `Error 1062 duplicate entry ... uk_order_outbox_deduplication` | Same business transition tried to create duplicate event | Treat as idempotency/retry signal, inspect caller flow | Use stable deduplication keys and no duplicate state transition |
| Outbox rows stay `failed` | Kafka down, wrong broker address, or topic issue | Check broker, topic, advertised listener, and worker logs | Add health checks and alerts |
| Outbox rows become `dead_letter` | Max publish attempts exhausted | Fix broker/root cause and replay manually with controlled tooling | Alert on dead-letter count |
| Events consumed out of order for same order | Kafka key not `order_id` or custom publisher changed key | Use `AggregateID` as Kafka key | Do not use random event ID as Kafka partition key |
| Consumer processed duplicate side effect | Consumer did not dedupe `event_id` | Store processed `event_id` per consumer | Consumer idempotency is mandatory |
| `FOR UPDATE SKIP LOCKED` SQL error | MySQL version too old or unsupported mode | Use MySQL 8+ | Keep local/prod MySQL version aligned |

## 12. Security & Best Practices

### Task-Specific Security Rules

| Rule | Why |
|---|---|
| Do not include full address, phone, tokens, card data, or raw headers in event payload | Events fan out to many consumers, so PII/secret exposure risk badhta hai |
| Use env/secret manager for Kafka and DB config | Broker and DB credentials source code me nahi hone chahiye |
| Keep `event_id` stable across publish retries | Consumer dedupe depends on it |
| Keep `deduplication_key` unique per business fact | Duplicate lifecycle facts prevent hote hain |
| Use Kafka TLS/SASL before production if broker is secured | Current config does not expose those env vars yet |
| Alert on `dead_letter` rows | Silent integration loss avoid hota hai |
| Do not set `ORDER_EVENTS_ENABLED=false` in production without explicit product/platform approval | Downstream services events miss kar sakte hain |

### Beginner-Friendly Best Practices

- Apply migrations in order before testing real repository flows.
- Local unit tests ke liye Kafka start karna zaroori nahi.
- Integration/staging me events enabled rakho so event flow actually verify ho.
- Consumer services ko `event_id` dedupe table maintain karni chahiye.
- `last_error` useful hai, but secrets log mat karo.
- Outbox backlog monitor karo. Pending rows ka growing count means downstream event delay.
- RabbitMQ switch karna ho to one approved broker path choose karo, Kafka and RabbitMQ dono me same event blindly publish mat karo.

## 13. Missing Or Misconfigured Things

| Area | Current Gap | Setup Impact | Suggested Fix |
|---|---|---|---|
| Server bootstrap | No `${SERVICE_CODE_PATH}/cmd/server/main.go` found | Long-running service and worker cannot be started with `go run ./cmd/server` | Add bootstrap wiring config, DB, repos, usecases, gRPC server, Kafka publisher, outbox worker |
| Docker assets | No checked-in Dockerfile or compose file found | Beginners must rely on docs/manual infra commands | Add repo-level compose for MySQL/Kafka and Dockerfile after server entrypoint exists |
| `.env.example` | No service `.env.example` found | Beginners may miss event/outbox variables | Add sanitized example with no real secrets |
| Kafka security config | Current config has broker list/topic/timeouts, not SASL/TLS env vars | Production secured Kafka cannot be configured cleanly | Add TLS/SASL config before production |
| Worker metrics | Logs exist, metrics implementation not found | Backlog/dead-letter alerting is manual | Add metrics for pending, published, failed, dead-letter, latency |
| Dead-letter replay tooling | Dead-letter status exists, replay command not found | Manual DB edits may become risky | Add controlled admin replay command |
| RabbitMQ alternative | Documented alternative, no adapter/dependency in current code | RabbitMQ cannot be used without code changes | Add `amqp091-go` adapter only if platform selects RabbitMQ |
| Consumer idempotency | Producer contract documented, consumers not implemented here | Downstream duplicate side effects possible | Each consumer stores processed `event_id` |

No hardcoded DB password, Kafka password, or payment secret was found in the `${TASK_FILE_NAME}`-specific files reviewed. Keep checking during future bootstrap work because secrets often accidentally enter config examples.

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Project Tech Stack Analysis | Go, gRPC, MySQL, Kafka, Docker basics already explained |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Go Dependency System | Same module, `go.mod`, `go.sum`, workspace, and common Go issues |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Environment Variables | Full `.env` sample and env loading behavior already documented |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | External Services Analysis -> Kafka | Kafka install, local broker setup, topic basics reused |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Docker and DevOps Setup | MySQL/Kafka compose examples and port conflict guidance reused |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_1}` | Common Errors and Fixes | Generic Go/MySQL/Docker/Kafka errors reused |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_2}` | Database Analysis and Migration Order | Base `order_db`, `orders`, and `order_status_history` required before outbox FK |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_3}` | Cart/Product checkout setup | `OrderCreated` is produced by same checkout/order creation path |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_4}` | Payment coordination and outbox caveat | `OrderPaid` requires payment migration and outbox migration for captured result |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_5}` | gRPC metadata and ports | Event-producing API commands reuse existing trusted caller boundary and port `9094` |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_6}` | Seller fulfillment setup | `OrderDelivered` can be produced by delivered aggregation flow |
| `${PREVIOUS_DEPENDENCY_FILE_TASK_7}` | Idempotency and full runtime migration alignment | Duplicate checkout protection and migration `001` to `004` guidance reused |

## 15. Final Checklist

- [ ] Previous dependency documentation checked first
- [ ] `${INPUT_FILE_PATH}` reviewed
- [ ] `${OUTPUT_FILE_PATH}` created without modifying original implementation guide
- [ ] No duplicate Go/MySQL/Docker/Kafka installation docs added
- [ ] `go.mod` checked and `github.com/segmentio/kafka-go` already present
- [ ] No RabbitMQ dependency installed for current Kafka path
- [ ] MySQL 8+ available for real DB verification
- [ ] Migrations `001`, `002`, `003`, and `004` applied in order for current runtime
- [ ] `order_outbox_events` table exists
- [ ] `uk_order_outbox_deduplication` unique key exists
- [ ] `ORDER_MYSQL_DSN` points to `order_db`
- [ ] `ORDER_EVENTS_ENABLED` explicitly set for local environment
- [ ] `ORDER_EVENTS_TOPIC=order.events` when events enabled
- [ ] `ORDER_KAFKA_BROKERS` set when events enabled
- [ ] Kafka topic `order.events` created/verified when event publishing is tested
- [ ] Outbox worker settings are positive durations/integers
- [ ] Unit tests for `orderevents`, `events`, repository, and usecase packages run
- [ ] `OrderCreated`, `OrderPaid`, `OrderCancelled`, and `OrderDelivered` rows verified in outbox for matching flows
- [ ] Consumers dedupe by `event_id`
- [ ] Dead-letter rows monitored and not deleted silently
- [ ] Missing `cmd/server/main.go` understood before trying to start service
- [ ] Missing Dockerfile/compose and `.env.example` tracked as setup gaps
- [ ] No secrets committed in `.env` or docs
