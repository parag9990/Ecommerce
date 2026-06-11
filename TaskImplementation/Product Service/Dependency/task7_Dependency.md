# Project Dependency & Setup Guide

## Variable Values Used

```text
SERVICE_NAME=Product Service
TASK_FILE_NAME=task7.md
INPUT_FILE_PATH=TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}
OUTPUT_FILE_NAME=task7_Dependency.md
OUTPUT_FILE_PATH=TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}
```

Important:

- `INPUT_FILE_PATH` was analyzed only for dependency, setup, environment, database, and DevOps requirements.
- The original `TASK_FILE_NAME` was not modified.
- Shared Go, MongoDB, Docker, RabbitMQ, `.env`, ports, and generic troubleshooting are already documented in earlier dependency files. This file explains only what is new or task-specific for product search event publishing.

## 1. Project Overview

`TASK_FILE_NAME` covers async product search event publishing.

Simple Hinglish:

Jab seller product create, update, publish, unpublish karta hai, ya inventory availability change hoti hai, tab service ek event prepare karke MongoDB outbox me save karta hai. Background relay worker baad me outbox se event uthakar `product.events` RabbitMQ queue me publish karta hai. Search Service is event ko consume karke apna search index update karega.

### Current implementation reality

| Area | Current status | Setup impact |
|---|---|---|
| Product event model | Present in `internal/domain/product_event.go` | Event types, envelope, payload, and validation are implemented |
| Search payload mapper | Present in `internal/mapper/product_search_event.go` | Product document se search-friendly payload ban raha hai |
| Seller write event triggers | Present in `internal/usecase/seller_product_service.go` | Create/update/publish/unpublish flows can queue outbox events |
| Inventory event trigger | Present in `internal/usecase/inventory_service.go` | `ProductInventoryChanged` queues only when in-stock status changes |
| MongoDB outbox repository | Present in `internal/repository/mongo_product_event_outbox_repository.go` | Requires `product_event_outbox` collection |
| MongoDB migration | New `0007_product_event_outbox.up.js` | Must be applied before DB-backed event testing |
| RabbitMQ publisher | Present in `internal/events/rabbitmq_publisher.go` | Publishes persistent JSON messages to durable queue `product.events` |
| Outbox relay worker | Present in `internal/events/outbox_relay.go` | Requires app startup to call `StartBackgroundWorkers` |
| Kafka publisher | Not built in | `PRODUCT_EVENT_BROKER=kafka` needs injected custom publisher |
| Server entrypoint | Still not present | Do not expect `go run .` to start HTTP/gRPC or workers |
| Dockerfile / compose | Still not present for this module | Reuse manual local infra setup from previous docs |

## 2. Tech Stack

Common technology explanation is already documented here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
3. Project Tech Stack Analysis
4. Language-Specific Dependency System: Go
5. Database Analysis: MongoDB
7. External Services Analysis
```

### Task-specific tech stack summary

| Technology / component | Required for `TASK_FILE_NAME`? | Why used | New setup? |
|---|---:|---|---:|
| Go | Yes | Event domain, mapper, usecases, outbox relay, and tests are Go code | No |
| Go modules | Yes | `go.mod` and `go.sum` manage dependencies | No |
| MongoDB | Yes for reliable outbox storage | Stores `product_event_outbox` events and product writes | Partly, new `0007` collection |
| MongoDB Go Driver v2 | Yes | MongoDB repositories, transactions, and outbox updates | No |
| RabbitMQ | Yes when local event publishing is enabled | Broker for `product.events` queue | Setup reused, queue behavior is task-specific |
| RabbitMQ AMQP client | Yes with RabbitMQ broker | `github.com/rabbitmq/amqp091-go` publishes JSON events | No new install, already in module |
| Docker | Optional but recommended | Easy MongoDB/RabbitMQ local containers | No |
| `mongosh` | Required for manual migration | Runs `0007_product_event_outbox.up.js` | No |
| Kafka | Optional/custom only | Config accepts it, but app does not auto-wire Kafka publisher | No local setup recommended |
| Typesense/Search Service | Consumer-side dependency only | Product service only publishes events; it does not run search indexing | No |
| Redis | No | Not used by current event implementation | No |

Hinglish note:

RabbitMQ ek message broker hai. Yahan `SERVICE_NAME` event bhejta hai, aur Search Service future/current consumer side pe us event ko process karega. Is service ko Typesense install karne ki zarurat nahi hai.

## 3. Required Software

Do not reinstall already documented tools. Follow:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
4. Language-Specific Dependency System: Go
5. Database Analysis: MongoDB
7. External Services Analysis -> RabbitMQ
9. Docker and DevOps Setup
10. Migration Setup
```

For `TASK_FILE_NAME` verification, make sure these are available:

| Software | Purpose | Status |
|---|---|---|
| Go matching `go.mod` | Run tests/build | Reused |
| MongoDB | Store products and `product_event_outbox` | Reused plus `0007` migration |
| `mongosh` | Apply and verify migration | Reused |
| RabbitMQ | Publish events when outbox worker is enabled | Reused |
| Docker | Optional local MongoDB/RabbitMQ setup | Reused |

## 4. Dependency Management

Go module setup, `go mod download`, `go mod tidy`, `go test`, `go build`, and common Go module errors are already explained here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
4. Language-Specific Dependency System: Go
```

### What changed for `TASK_FILE_NAME`?

No new third-party dependency needs to be installed manually for this documentation step. The current module already contains the required direct dependencies:

| Dependency | Version | Why it matters |
|---|---:|---|
| `go.mongodb.org/mongo-driver/v2` | `v2.6.0` | MongoDB outbox collection, transactions, relay state updates |
| `github.com/rabbitmq/amqp091-go` | `v1.10.0` | RabbitMQ AMQP publishing for `product.events` |

Task-specific practical commands:

```bash
cd backend/services/product-service
go test ./internal/usecase -run 'ProductEvent|CreateProductQueuesProductEvent|InventoryChangedEvent'
go test ./internal/events
go test ./...
go build ./...
```

Beginner note:

Do not run `go get` just because events exist. Dependencies are already declared. Run `go mod download` only if your local module cache is missing packages.

## 5. Database Setup

MongoDB installation, Docker setup, connection string format, credentials, and generic troubleshooting are reused:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
5. Database Analysis: MongoDB
10. Migration Setup
```

### Database used

| Item | Value |
|---|---|
| Database | `product_db` |
| Connection env | `PRODUCT_MONGO_URI` |
| Database env | `PRODUCT_MONGO_DATABASE` |
| Default MongoDB port | `27017` |

Warning:

Migration scripts use `db.getSiblingDB("product_db")`. Agar local `.env` me `PRODUCT_MONGO_DATABASE` change karte ho, app aur migration alag DB target kar sakte hain. Beginner setup ke liye `product_db` hi rakho.

### Collections used by `TASK_FILE_NAME`

| Collection | Required? | Status | Why |
|---|---:|---|---|
| `products` | Yes | Reused from `0003` and later migrations | Product data se search event payload build hota hai |
| `product_event_outbox` | Yes for DB-backed events | New in `0007` | Pending/published/dead-lettered product events store hote hain |
| `inventory_reservations` | Conditional | Reused from `0006` | Inventory flow can trigger `ProductInventoryChanged` |
| `inventory_snapshots` | Conditional | Reused from `0003` | Inventory audit records, not event-specific |

### Task-specific migration

New migration for `TASK_FILE_NAME`:

```text
backend/services/product-service/migrations/mongo/0007_product_event_outbox.up.js
```

Rollback file:

```text
backend/services/product-service/migrations/mongo/0007_product_event_outbox.down.js
```

What `0007` adds:

| Item | Purpose |
|---|---|
| `product_event_outbox` collection | Stores events before publishing |
| JSON schema validator | Allows only supported event types/status values |
| `idx_product_event_outbox_status_next_occurred` | Relay worker quickly finds due pending/publishing events |
| `idx_product_event_outbox_type_occurred` | Debug/audit by event type and time |
| `ttl_product_event_outbox_published` | Cleans published events after 30 days |

### Run migrations for event verification

Use previous docs for full MongoDB start/install steps. After MongoDB is running:

```bash
cd backend/services/product-service
mongosh "mongodb://localhost:27017" migrations/mongo/0003_create_product_collections.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0004_seller_product_workflow.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0005_product_read_indexes.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0006_inventory_reservations.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0007_product_event_outbox.up.js
```

If testing only seller create/update/publish/unpublish events and not inventory events, `0006` is not strictly required. For full current service verification, run migrations in order through `0007`.

### Verify Task 7 collection and indexes

```bash
mongosh "mongodb://localhost:27017/product_db" --eval "db.getCollectionNames()"
mongosh "mongodb://localhost:27017/product_db" --eval "db.product_event_outbox.getIndexes().map(i => i.name)"
```

Expected `product_event_outbox` indexes:

```text
_id_
idx_product_event_outbox_status_next_occurred
idx_product_event_outbox_type_occurred
ttl_product_event_outbox_published
```

### MongoDB transaction warning

Seller product writes with outbox use MongoDB transactions. This is already explained here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md

Section:
8. Docker Setup -> MongoDB transaction warning
```

Important for `TASK_FILE_NAME`:

| Local mode | Recommendation |
|---|---|
| Testing pure unit tests | No MongoDB transaction setup needed |
| Testing Mongo-backed seller write plus outbox | Use MongoDB replica set mode |
| Testing without event outbox | Set `PRODUCT_EVENTS_ENABLED=false` |

Common transaction error:

```text
Transaction numbers are only allowed on a replica set member or mongos
```

Fix:

Use the replica-set setup from `task4_Dependency.md`, or disable events for local tests that do not need outbox behavior.

## 6. Redis / Queue / External Services

### Redis

Redis is not used by current `TASK_FILE_NAME` implementation. No Redis install, port, credentials, or container is newly required.

### Typesense / Search Service

Typesense and Search Service are consumers of the event, not runtime dependencies of this service task.

| Item | Status |
|---|---|
| `SERVICE_NAME` publishes `product.events` | Yes |
| Search Service consumes and indexes | Separate service scope |
| Typesense installation needed here | No |
| Full reindex job needed here | No |

### RabbitMQ

RabbitMQ setup is reused:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
7. External Services Analysis -> RabbitMQ
```

Task-specific RabbitMQ behavior:

| Item | Value |
|---|---|
| Default broker | `rabbitmq` |
| Queue/topic name | `product.events` |
| Queue declaration | Durable queue declared automatically by publisher |
| Exchange | Empty default exchange |
| Routing key | Same as topic/queue name |
| Message format | JSON event envelope |
| Delivery mode | Persistent |
| Content type | `application/json` |
| Headers | `event_id`, `request_id`, `trace_id`, `source`, `version` |

RabbitMQ is required only when all of these are true:

| Condition | Meaning |
|---|---|
| `PRODUCT_EVENTS_ENABLED=true` | Event recorder is enabled |
| `PRODUCT_OUTBOX_WORKER_ENABLED=true` | Relay worker should publish queued events |
| `PRODUCT_EVENT_BROKER=rabbitmq` | RabbitMQ publisher is selected |
| Runnable app entrypoint calls `StartBackgroundWorkers` | Current module does not yet provide this server entrypoint |

Verify queue after a running app publishes at least one event:

```bash
docker exec ecommerce-product-rabbitmq rabbitmqctl list_queues name messages_ready messages_unacknowledged
```

You should see:

```text
product.events
```

### Kafka

Kafka is not wired by default.

Current behavior:

| Config | Result |
|---|---|
| `PRODUCT_EVENT_BROKER=rabbitmq` | App can auto-create RabbitMQ publisher |
| `PRODUCT_EVENT_BROKER=kafka` | App requires injected `ProductEventPublisher` dependency |

Beginner recommendation:

```env
PRODUCT_EVENT_BROKER=rabbitmq
```

Do not start Kafka for this task unless you are also implementing and injecting a Kafka publisher.

## 7. Environment Variables

Complete `.env` creation, loading, and full variable table are already documented here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
6. Environment Variables
```

Where `.env` should be created:

```text
backend/services/product-service/.env
```

Reminder:

Current config uses `os.LookupEnv`. Sirf `.env` file create karne se values auto-load nahi hoti. Shell me `source .env` ya dotenv/direnv tool use karo.

### Task-specific event variables

For local event/outbox verification:

```env
PRODUCT_EVENTS_ENABLED=true
PRODUCT_EVENTS_TOPIC=product.events
PRODUCT_EVENT_BROKER=rabbitmq
RABBITMQ_URL=amqp://ecommerce:ecommerce@localhost:5672/

PRODUCT_OUTBOX_WORKER_ENABLED=true
PRODUCT_OUTBOX_POLL_INTERVAL_MS=1000
PRODUCT_OUTBOX_BATCH_SIZE=100
PRODUCT_OUTBOX_MAX_ATTEMPTS=5
PRODUCT_OUTBOX_PUBLISH_TIMEOUT_MS=10000
```

For isolated tests where RabbitMQ is not needed:

```env
PRODUCT_EVENTS_ENABLED=false
PRODUCT_OUTBOX_WORKER_ENABLED=false
```

### Variable explanation

| Variable | Required? | Example | Purpose | Security note |
|---|---:|---|---|---|
| `PRODUCT_EVENTS_ENABLED` | Optional, default `true` | `true` | Enables event recorder/outbox creation | Not secret |
| `PRODUCT_EVENTS_TOPIC` | Required when events enabled | `product.events` | Queue/topic used for product events | Not secret |
| `PRODUCT_EVENT_BROKER` | Required when events enabled | `rabbitmq` | Selects publisher type | Not secret |
| `RABBITMQ_URL` | Required for RabbitMQ broker | `amqp://ecommerce:ecommerce@localhost:5672/` | RabbitMQ connection string | Secret outside local |
| `PRODUCT_OUTBOX_WORKER_ENABLED` | Optional, default `true` | `true` | Starts relay worker when app entrypoint calls background workers | Not secret |
| `PRODUCT_OUTBOX_POLL_INTERVAL_MS` | Optional | `1000` | How often relay polls due events | Not secret |
| `PRODUCT_OUTBOX_BATCH_SIZE` | Optional | `100` | Max events per relay cycle | Not secret |
| `PRODUCT_OUTBOX_MAX_ATTEMPTS` | Optional | `5` | Attempts before dead-lettering | Not secret |
| `PRODUCT_OUTBOX_PUBLISH_TIMEOUT_MS` | Optional | `10000` | Timeout for one publish call | Not secret |

Validation rules:

| Rule | Error if broken |
|---|---|
| topic cannot be empty when events are enabled | `PRODUCT_EVENTS_TOPIC is required when PRODUCT_EVENTS_ENABLED is true` |
| broker must be `rabbitmq` or `kafka` | `PRODUCT_EVENT_BROKER must be rabbitmq or kafka` |
| RabbitMQ URL required for RabbitMQ broker | `RABBITMQ_URL is required when PRODUCT_EVENT_BROKER is rabbitmq` |
| poll interval must be greater than `0` | `PRODUCT_OUTBOX_POLL_INTERVAL_MS must be greater than zero` |
| batch size must be greater than `0` | `PRODUCT_OUTBOX_BATCH_SIZE must be greater than zero` |
| max attempts must be greater than `0` | `PRODUCT_OUTBOX_MAX_ATTEMPTS must be greater than zero` |
| publish timeout must be greater than `0` | `PRODUCT_OUTBOX_PUBLISH_TIMEOUT_MS must be greater than zero` |

## 8. Docker Setup

Docker basics are reused:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
5. Database Analysis: MongoDB -> H. Docker setup
5. Database Analysis: MongoDB -> I. Docker Compose example
7. External Services Analysis -> RabbitMQ
9. Docker and DevOps Setup
```

Task-specific Docker impact:

| Docker item | New for `TASK_FILE_NAME`? | Note |
|---|---:|---|
| MongoDB container | No | Reuse existing local MongoDB setup |
| MongoDB replica set | Conditional | Needed for Mongo-backed product write plus outbox transaction tests |
| RabbitMQ container | No | Reuse previous RabbitMQ setup |
| `product.events` queue | Task-specific runtime artifact | Publisher declares it automatically |
| Product service container | No | No Dockerfile exists yet |
| Docker Compose file | No | No repo-owned compose file exists yet |
| New volume | No | Reuse MongoDB/RabbitMQ volumes from previous docs |
| New network | No | No task-specific Docker network added |

If using Docker MongoDB, run migration `0007` against the same MongoDB instance that the app uses. Agar migration host MongoDB me run hui aur app Docker MongoDB se connect kar raha hai, `product_event_outbox` missing lagega.

## 9. Local Development Setup

### Step 1: Read previous dependency docs first

Follow these first because they already explain shared setup:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
TaskImplementation/{SERVICE_NAME}/task6_Dependency.md
```

Main reused topics:

- Go installation and module commands
- MongoDB installation or Docker setup
- MongoDB transaction/replica-set warning
- `.env` location and loading
- RabbitMQ setup and local credentials
- Generic Docker/MongoDB/RabbitMQ troubleshooting
- Existing migrations through `0006`

### Step 2: Go to project directory

```bash
cd backend/services/product-service
```

### Step 3: Install dependencies

No task-specific package install is needed. If dependencies are not downloaded:

```bash
go mod download
```

### Step 4: Setup database/services

For isolated event domain/unit tests:

1. No MongoDB/RabbitMQ is required.
2. Run Go unit tests only.

For Mongo-backed outbox verification:

1. Start MongoDB.
2. Use replica-set mode if testing seller product write plus outbox transaction.
3. Run migrations through `0007`.

For publish-to-RabbitMQ verification:

1. Start MongoDB.
2. Start RabbitMQ.
3. Run migrations through `0007`.
4. Enable event/outbox env vars.
5. Start future app entrypoint that wires `app.New(...)` and calls `StartBackgroundWorkers(...)`.

### Step 5: Add or verify event env values

Use:

```text
backend/services/product-service/.env
```

Recommended local values for event publishing:

```env
PRODUCT_EVENTS_ENABLED=true
PRODUCT_EVENTS_TOPIC=product.events
PRODUCT_EVENT_BROKER=rabbitmq
RABBITMQ_URL=amqp://ecommerce:ecommerce@localhost:5672/
PRODUCT_OUTBOX_WORKER_ENABLED=true
PRODUCT_OUTBOX_POLL_INTERVAL_MS=1000
PRODUCT_OUTBOX_BATCH_SIZE=100
PRODUCT_OUTBOX_MAX_ATTEMPTS=5
PRODUCT_OUTBOX_PUBLISH_TIMEOUT_MS=10000
```

### Step 6: Run migrations

```bash
mongosh "mongodb://localhost:27017" migrations/mongo/0003_create_product_collections.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0004_seller_product_workflow.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0005_product_read_indexes.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0006_inventory_reservations.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0007_product_event_outbox.up.js
```

### Step 7: Start backend service

Current run reality:

| Command | Status |
|---|---|
| `go test ./...` | Supported |
| `go build ./...` | Supported for packages |
| `go run .` | Not supported because no `main` package exists |
| Start HTTP/gRPC server | Not available from this module yet |
| Start outbox worker through app entrypoint | App support exists, but no server/worker `main` exists yet |

So for now, verify through tests/build and MongoDB/RabbitMQ inspection.

### Step 8: Verify `TASK_FILE_NAME` functionality

```bash
go test ./internal/usecase -run 'ProductEvent|CreateProductQueuesProductEvent|InventoryChangedEvent'
go test ./internal/events
go test ./...
go build ./...
```

Task-specific behavior to verify:

| Flow | Expected result |
|---|---|
| Create product | Queues `ProductCreated` when events are enabled |
| Update product | Queues `ProductUpdated` |
| Publish product direct to live | Queues `ProductPublished` |
| Publish product requiring review | Queues `ProductUpdated` with delete search action |
| Unpublish product | Queues `ProductUnpublished` |
| Inventory in-stock status changes | Queues `ProductInventoryChanged` |
| Relay publish success | Marks outbox event `published` |
| Relay publish failure before max attempts | Marks event back to `pending` with future `next_attempt_at` |
| Relay max attempts exceeded | Marks event `dead_lettered` |

## 10. Running the Project

Full shared run instructions are reused:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
11. Complete Project Run Instructions
```

### Ports and networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Backend API | N/A | HTTP/gRPC API | Reused limitation: no server listener yet |
| MongoDB | 27017 | Product DB and outbox storage | Reused |
| RabbitMQ AMQP | 5672 | Publish `product.events` | Reused |
| RabbitMQ Management UI | 15672 | Inspect queue/messages | Reused |
| Kafka | N/A | Not wired by default | Not recommended |
| Redis | N/A | Not used | Not required |
| Typesense | N/A | Search Service dependency, not this service task | Not required |

Port conflict and Docker network troubleshooting are already documented here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
8. Ports and Networking
```

### Inspect outbox state

After a DB-backed flow queues events:

```bash
mongosh "mongodb://localhost:27017/product_db" --eval 'db.product_event_outbox.find({}, {_id:1,event_type:1,status:1,attempts:1,next_attempt_at:1,last_error:1}).sort({occurred_at:-1}).limit(10).toArray()'
```

Status meaning:

| Status | Meaning |
|---|---|
| `pending` | Event is waiting for relay publish |
| `publishing` | Relay has claimed the event |
| `published` | RabbitMQ publish succeeded |
| `dead_lettered` | Event failed after max attempts or invalid payload |

## 11. Common Errors & Fixes

Generic errors are reused:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
12. Common Errors and Fixes
```

Task-specific errors:

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `product event outbox repository is required` | Events enabled but no outbox repository was wired | Wire `ProductEventOutboxRepository` or disable events | In app wiring, provide full Mongo repository before enabling events |
| `PRODUCT_EVENTS_TOPIC is required...` | Empty topic while events enabled | Set `PRODUCT_EVENTS_TOPIC=product.events` | Copy `.env.example` values |
| `RABBITMQ_URL is required...` | RabbitMQ broker selected but URL missing | Add `RABBITMQ_URL` | Keep local `.env` complete |
| `connect rabbitmq: dial tcp ... connection refused` | RabbitMQ not running or wrong URL/port | Start RabbitMQ and verify port `5672` | Run `rabbitmq-diagnostics ping` first |
| Queue `product.events` not visible | No event was published yet, worker not running, or RabbitMQ not connected | Ensure relay worker is started and events exist in outbox | Check MongoDB outbox before checking queue |
| Outbox stuck in `pending` | Worker not running or publish keeps failing | Start worker-enabled entrypoint and inspect `last_error` | Monitor relay logs and RabbitMQ health |
| Outbox stuck in `publishing` | Worker crashed after claiming event | Relay will retry after lease/next attempt time | Keep worker supervised in future runtime |
| Event becomes `dead_lettered` | Publish failed too many times or payload invalid | Fix broker/payload issue and replay manually after review | Set alerts on dead-letter count |
| `PRODUCT_EVENT_BROKER=kafka requires a ProductEventPublisher dependency` | Kafka selected but no built-in Kafka adapter exists | Use `rabbitmq` or implement/inject Kafka publisher | Do not set Kafka in beginner local setup |
| Mongo transaction replica-set error | Product+outbox transaction used on standalone MongoDB | Use replica-set MongoDB or disable events | Follow transaction warning from `task4_Dependency.md` |
| `product_event_outbox` missing | Migration `0007` not applied or applied to different DB | Run `0007` against the same MongoDB URI | Keep `PRODUCT_MONGO_DATABASE=product_db` locally |

## 12. Security & Best Practices

Task-specific security rules:

- Do not commit real `RABBITMQ_URL` credentials.
- Local `ecommerce/ecommerce` RabbitMQ credentials are only for development.
- Keep `PRODUCT_EVENTS_TOPIC` stable because consumers depend on it.
- Treat event payloads as public-ish integration data: do not add secrets, tokens, or internal credentials to payload.
- Preserve event `event_id`, `request_id`, and `trace_id` for debugging and idempotency.
- Consumers should be idempotent by `event_id` because retries can publish or deliver duplicates.
- Keep `PRODUCT_OUTBOX_MAX_ATTEMPTS` high enough for temporary broker outage, but alert on `dead_lettered` events.
- Do not enable Kafka config in production unless a Kafka publisher is actually wired and tested.

Best practices specific to `TASK_FILE_NAME`:

| Practice | Why |
|---|---|
| Run migration `0007` before enabling events | Outbox insert needs the collection and indexes |
| Use MongoDB transactions for product write plus outbox | Product change and event record stay consistent |
| Keep event schema versioned | Search Service can handle future payload changes safely |
| Publish `DELETE` action for non-published products | Draft/unpublished products should not appear in public search |
| Monitor outbox statuses | `pending`, `publishing`, and `dead_lettered` reveal pipeline health |
| Keep RabbitMQ queue durable | Events should survive broker restart |
| Keep messages persistent | Matches durable queue expectation |

## 13. Missing or Misconfigured Things

| Area | Current observation | Setup risk | Recommended fix |
|---|---|---|---|
| Service entrypoint | No `cmd/server/main.go` exists | Cannot start API or worker with `go run .` | Add server/worker entrypoint and call `StartBackgroundWorkers` |
| Dockerfile | No product service Dockerfile exists | Cannot build app container | Add service Dockerfile when runtime entrypoint exists |
| docker-compose | No repo-owned local compose exists | Beginners must run MongoDB/RabbitMQ manually | Add local compose with MongoDB replica set and RabbitMQ |
| Kafka | Config accepts Kafka but no built-in publisher | Misleading runtime option | Implement Kafka publisher or disable option in config |
| Health checks | No app health endpoint yet | Hard to verify worker/broker readiness | Add health/readiness checks after server entrypoint |
| Migration DB name | Migration scripts hardcode `product_db` | Env DB mismatch possible | Parameterize migration target or document fixed DB |
| Queue DLQ | Code uses outbox `dead_lettered` status, not RabbitMQ DLX | Operators may expect broker DLQ | Document outbox dead-letter workflow or add RabbitMQ DLX later |
| Worker supervision | Relay runs as goroutine when called | No process manager yet | Add graceful shutdown and supervisor in runtime entrypoint |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Project Tech Stack Analysis` | Go, MongoDB, RabbitMQ, Docker basics already explained |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Language-Specific Dependency System: Go` | Go module commands and common dependency errors are same |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. Database Analysis: MongoDB` | MongoDB install, Docker run, connection string, credentials are same |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `6. Environment Variables` | `.env` location/loading and common mistakes are same |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. External Services Analysis -> RabbitMQ` | RabbitMQ install, ports, credentials, and verification are same |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `8. Ports and Networking` | Port conflict and Docker networking guidance is same |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `9. Docker and DevOps Setup` | Generic local container setup is same |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `10. Migration Setup` | General migration workflow is same |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `5. Database Setup` | Base catalog collections from `0003` are reused |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | `8. Docker Setup -> MongoDB transaction warning` | Product write plus outbox needs MongoDB transaction support |
| `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | `6. Redis / Queue / External Services -> Search / Typesense` | Search boundary is already clarified |
| `TaskImplementation/{SERVICE_NAME}/task6_Dependency.md` | `6. Redis / Queue / External Services` and `7. Environment Variables` | Event/outbox toggles and inventory-triggered event setup are reused |

## 15. Final Checklist

- [ ] Previous dependency documentation checked before using this file
- [ ] Go dependencies downloaded if local cache was empty
- [ ] MongoDB running on the expected URI
- [ ] `PRODUCT_MONGO_DATABASE=product_db` verified for beginner setup
- [ ] Migrations `0003` through `0007` applied in order for full verification
- [ ] `product_event_outbox` collection exists
- [ ] Outbox indexes verified
- [ ] MongoDB replica-set mode used if testing product write plus outbox transaction
- [ ] RabbitMQ running if publish relay is enabled
- [ ] `PRODUCT_EVENTS_ENABLED` value chosen intentionally
- [ ] `PRODUCT_EVENTS_TOPIC=product.events` configured when events are enabled
- [ ] `PRODUCT_EVENT_BROKER=rabbitmq` used for beginner local setup
- [ ] `RABBITMQ_URL` configured and kept local-only
- [ ] `PRODUCT_OUTBOX_WORKER_ENABLED` configured intentionally
- [ ] Event usecase tests run
- [ ] Outbox relay tests run
- [ ] Full `go test ./...` run
- [ ] `go build ./...` run
- [ ] RabbitMQ queue inspected if publishing was tested
- [ ] Outbox `pending/published/dead_lettered` statuses inspected if DB-backed flow was tested
- [ ] No duplicate setup documentation added
