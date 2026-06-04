# Project Dependency & Setup Guide

## Variable Values Used

```text
SERVICE_NAME=Product Service
TASK_FILE_NAME=task6.md
INPUT_FILE_PATH=TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}
OUTPUT_FILE_NAME=task6_Dependency.md
OUTPUT_FILE_PATH=TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}
```

Important:

- `INPUT_FILE_PATH` was analyzed only for setup, dependency, environment, and DevOps requirements.
- The original `TASK_FILE_NAME` was not modified.
- Generic Go, MongoDB, Docker, RabbitMQ, and `.env` setup already exists in previous dependency files, so this file reuses those sections and explains only task-specific additions.

## 1. Project Overview

`TASK_FILE_NAME` covers inventory operations:

- `ReserveInventory`
- `ReleaseInventory`
- `CommitInventory`
- expired reservation processing through `ExpireReservations`

Simple Hinglish:

Inventory reservation ka matlab checkout ke time stock ko temporary hold karna. Payment fail/cancel ho to hold release hota hai. Payment success ho to reserved stock final order commit me decrement hota hai.

### Current implementation reality

| Area | Current status | Setup impact |
|---|---|---|
| Inventory usecase | Present in `internal/usecase/inventory_service.go` | Tests/build can verify logic |
| Inventory handler | Present in `internal/transport/inventory` | Internal adapter exists, but no full network server wiring yet |
| MongoDB stock writes | Present in `internal/repository/mongo_inventory_repository.go` | Requires MongoDB collections and indexes |
| Reservation collection | New migration `0006_inventory_reservations.up.js` | Must be applied for DB-backed reservation storage |
| Snapshot collection | Reused from `0003_create_product_collections.up.js` | No new snapshot migration in this task |
| Product event outbox | Conditional, via `ProductInventoryChanged` events | Requires `0007` and RabbitMQ only if events/outbox are enabled |
| Dockerfile / compose | Still not present for this service module | Reuse manual local infra setup from previous docs |
| Server entrypoint | Still not present | Do not expect `go run .` to start API/gRPC server |

## 2. Tech Stack

Full technology explanation is already documented here:

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

| Technology | Required for `TASK_FILE_NAME`? | Why used | New setup? |
|---|---:|---|---:|
| Go | Yes | Inventory usecase, repositories, handlers, and tests are Go code | No |
| Go modules | Yes | Dependencies are managed through `go.mod` and `go.sum` | No |
| MongoDB | Yes for DB-backed verification | Product variant stock, reservations, and audit snapshots live in MongoDB | Partly, new `inventory_reservations` collection |
| MongoDB Go Driver v2 | Yes | Atomic `$inc` updates reserve/release/commit stock | No |
| RabbitMQ | Conditional | Publishes product inventory changed events when events/outbox worker is enabled | No, reused |
| Docker | Optional | Easy local MongoDB/RabbitMQ containers | No |
| `mongosh` | Required for manual migrations | Runs migration JavaScript files | No |
| gRPC contract | Planned/internal | `api/master-api.json` lists internal inventory methods | No runnable listener yet |

Hinglish note:

MongoDB yahan mandatory hai kyunki stock quantity aur reservation state database me persist hoti hai. RabbitMQ mandatory tabhi hai jab product events ko actually publish karna ho.

## 3. Required Software

Do not reinstall already documented tools. Follow:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
3. Project Tech Stack Analysis
4. Language-Specific Dependency System: Go
5. Database Analysis: MongoDB
9. Docker and DevOps Setup
10. Migration Setup
```

For `TASK_FILE_NAME` verification, make sure these are available:

| Software | Purpose | Status |
|---|---|---|
| Go matching `go.mod` | Run tests/build | Reused |
| MongoDB or Docker MongoDB | Store products, reservations, snapshots, outbox | Reused plus Task 6 migration |
| `mongosh` | Apply/verify MongoDB migrations | Reused |
| RabbitMQ | Only if event publishing is enabled | Conditional, reused |

## 4. Dependency Management

No new third-party Go dependency is introduced only for `TASK_FILE_NAME`.

Current module file:

```text
backend/services/product-service/go.mod
```

Current direct dependencies:

| Dependency | Current purpose |
|---|---|
| `go.mongodb.org/mongo-driver/v2` | MongoDB collections, atomic inventory updates, indexes, repositories |
| `github.com/rabbitmq/amqp091-go` | RabbitMQ publisher for product events when enabled |

Go module setup, `go mod tidy`, `go mod download`, `go test`, and common module errors are already explained here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
4. Language-Specific Dependency System: Go
```

Task-specific practical commands:

```bash
cd backend/services/product-service
go test ./internal/usecase -run 'ReserveInventory|ReleaseAndCommitInventory|ExpireReservations'
go test ./...
go build ./...
```

## 5. Database Setup

MongoDB installation, Docker run commands, connection string format, credentials, and generic troubleshooting are reused:

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
| `products` | Yes | Reused from `0003` plus later migrations | Variant stock fields yahin update hote hain |
| `inventory_reservations` | Yes | New in `0006` | Order-wise reservation state, TTL, idempotency |
| `inventory_snapshots` | Yes | Reused from `0003` | Reservation/release/commit audit records |
| `product_event_outbox` | Conditional | `0007` | Needed if inventory changed events are enabled |

### Task-specific migration

New migration for `TASK_FILE_NAME`:

```text
backend/services/product-service/migrations/mongo/0006_inventory_reservations.up.js
```

Rollback file:

```text
backend/services/product-service/migrations/mongo/0006_inventory_reservations.down.js
```

What `0006` adds:

| Item | Purpose |
|---|---|
| `inventory_reservations` collection | Stores reservation id, order id, status, items, expiry, timestamps, reason |
| `uq_inventory_reservations_order` | Makes reserve retry safe by preventing duplicate active/order reservation documents |
| `uq_inventory_reservations_idempotency` | Optional idempotency key uniqueness |
| `idx_inventory_reservations_status_expires` | Fast expiry batch lookup |
| `ttl_inventory_reservations_cleanup` | Cleans reservation documents after expiry plus retention window |

Important:

MongoDB TTL index cleanup does not release stock by itself. It only deletes old documents later. Actual stock release needs the service's `ExpireReservations` logic to run before cleanup. Isliye future scheduler/worker setup still needed.

### Run migrations for inventory verification

Use previous docs for full MongoDB start/install steps. After MongoDB is running:

```bash
cd backend/services/product-service
mongosh "mongodb://localhost:27017" migrations/mongo/0003_create_product_collections.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0004_seller_product_workflow.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0005_product_read_indexes.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0006_inventory_reservations.up.js
```

Run `0007` only if product events/outbox are enabled:

```bash
cd backend/services/product-service
mongosh "mongodb://localhost:27017" migrations/mongo/0007_product_event_outbox.up.js
```

### Verify Task 6 collections and indexes

```bash
mongosh "mongodb://localhost:27017/product_db" --eval "db.getCollectionNames()"
mongosh "mongodb://localhost:27017/product_db" --eval "db.inventory_reservations.getIndexes().map(i => i.name)"
mongosh "mongodb://localhost:27017/product_db" --eval "db.inventory_snapshots.getIndexes().map(i => i.name)"
```

Expected `inventory_reservations` indexes:

```text
_id_
uq_inventory_reservations_order
uq_inventory_reservations_idempotency
idx_inventory_reservations_status_expires
ttl_inventory_reservations_cleanup
```

## 6. Redis / Queue / External Services

### Redis

Redis is not used by current `TASK_FILE_NAME` implementation. No Redis install, port, credentials, or container is newly required.

### Kafka

Kafka is not wired by default. Config validation accepts `PRODUCT_EVENT_BROKER=kafka`, but app wiring returns an error unless a custom `ProductEventPublisher` dependency is provided.

Beginner recommendation:

```env
PRODUCT_EVENT_BROKER=rabbitmq
```

or disable events for isolated inventory tests.

### RabbitMQ

RabbitMQ setup is reused:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
7. External Services Analysis -> RabbitMQ
```

For `TASK_FILE_NAME`, RabbitMQ is required only when all of these are true:

| Condition | Meaning |
|---|---|
| `PRODUCT_EVENTS_ENABLED=true` | Inventory changes can queue product events |
| `PRODUCT_OUTBOX_WORKER_ENABLED=true` | App background worker will try to publish queued events |
| `PRODUCT_EVENT_BROKER=rabbitmq` | RabbitMQ publisher is selected |
| Runnable server/worker entrypoint exists | Current module still lacks this entrypoint |

For isolated inventory tests, simpler local mode:

```env
PRODUCT_EVENTS_ENABLED=false
PRODUCT_OUTBOX_WORKER_ENABLED=false
```

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

Current config uses `os.LookupEnv`. Sirf `.env` file create karne se values auto-load nahi hoti. Shell me `source .env` ya tool-specific dotenv loader use karo.

### Task-specific variables to verify

Do not duplicate the full `.env`. For `TASK_FILE_NAME`, check only these values:

```env
PRODUCT_MONGO_URI=mongodb://localhost:27017
PRODUCT_MONGO_DATABASE=product_db
PRODUCT_MONGO_AUTO_CREATE_COLLECTIONS=false

PRODUCT_INVENTORY_DEFAULT_TTL_SECONDS=900
PRODUCT_INVENTORY_MIN_TTL_SECONDS=30
PRODUCT_INVENTORY_MAX_TTL_SECONDS=3600
PRODUCT_INVENTORY_EXPIRY_BATCH_LIMIT=100

# Isolated local inventory verification
PRODUCT_EVENTS_ENABLED=false
PRODUCT_OUTBOX_WORKER_ENABLED=false
```

If testing inventory changed events:

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

### Variable explanation

| Variable | Required? | Example | Purpose | Security note |
|---|---:|---|---|---|
| `PRODUCT_INVENTORY_DEFAULT_TTL_SECONDS` | Optional | `900` | Default reservation hold time, 15 minutes | Not secret |
| `PRODUCT_INVENTORY_MIN_TTL_SECONDS` | Optional | `30` | Smallest allowed request TTL | Not secret |
| `PRODUCT_INVENTORY_MAX_TTL_SECONDS` | Optional | `3600` | Largest allowed request TTL | Not secret |
| `PRODUCT_INVENTORY_EXPIRY_BATCH_LIMIT` | Optional | `100` | Max expired reservations processed in one batch | Not secret |
| `PRODUCT_EVENTS_ENABLED` | Optional | `false` for isolated tests | Controls event/outbox recording | Not secret |
| `PRODUCT_OUTBOX_WORKER_ENABLED` | Optional | `false` for isolated tests | Controls relay worker when app starts workers | Not secret |

Validation rules:

| Rule | Error if broken |
|---|---|
| min TTL must be greater than `0` | `PRODUCT_INVENTORY_MIN_TTL_SECONDS must be greater than zero` |
| default TTL must be greater than `0` | `PRODUCT_INVENTORY_DEFAULT_TTL_SECONDS must be greater than zero` |
| max TTL must be greater than `0` | `PRODUCT_INVENTORY_MAX_TTL_SECONDS must be greater than zero` |
| min TTL cannot exceed default TTL | `PRODUCT_INVENTORY_MIN_TTL_SECONDS cannot exceed PRODUCT_INVENTORY_DEFAULT_TTL_SECONDS` |
| default TTL cannot exceed max TTL | `PRODUCT_INVENTORY_DEFAULT_TTL_SECONDS cannot exceed PRODUCT_INVENTORY_MAX_TTL_SECONDS` |
| expiry batch limit must be greater than `0` | `PRODUCT_INVENTORY_EXPIRY_BATCH_LIMIT must be greater than zero` |

## 8. Docker Setup

Docker basics are reused:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
5. Database Analysis: MongoDB -> H. Docker setup
5. Database Analysis: MongoDB -> I. Docker Compose example
9. Docker and DevOps Setup
```

Task-specific Docker impact:

| Docker item | New for `TASK_FILE_NAME`? | Note |
|---|---:|---|
| MongoDB container | No | Reuse existing local MongoDB setup |
| RabbitMQ container | No | Conditional only for events/outbox publishing |
| App container | No | No service Dockerfile exists yet |
| Docker Compose file | No | No repo-owned compose file exists yet |
| New volume | No | Reuse MongoDB/RabbitMQ volumes from previous docs |
| New network | No | No task-specific Docker network added |

If MongoDB runs in Docker, run `0006` against the same MongoDB instance that the service uses. Agar migration host MongoDB me run hui aur app Docker MongoDB se connect kar raha hai, collection missing lagegi.

## 9. Local Development Setup

### Step 1: Read previous dependency docs first

Follow these first because they already explain shared setup:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
TaskImplementation/{SERVICE_NAME}/task5_Dependency.md
```

Main reused topics:

- Go installation and module commands
- MongoDB installation or Docker setup
- `.env` location and loading
- RabbitMQ setup if events are enabled
- Generic Docker/MongoDB/RabbitMQ troubleshooting
- Existing migrations through `0005`

### Step 2: Go to project directory

```bash
cd backend/services/product-service
```

### Step 3: Install dependencies

No task-specific dependency install is needed. If dependencies are not downloaded:

```bash
go mod download
```

### Step 4: Setup database/services

For isolated inventory verification:

1. Start MongoDB using previous docs.
2. Run migrations `0003` through `0006`.
3. Keep events disabled unless you are testing event publishing.

For event publishing verification:

1. Start MongoDB.
2. Start RabbitMQ.
3. Run migrations `0003` through `0007`.
4. Set event/outbox env vars to enabled.

### Step 5: Add or verify inventory env values

Use:

```text
backend/services/product-service/.env
```

Recommended isolated local values:

```env
PRODUCT_INVENTORY_DEFAULT_TTL_SECONDS=900
PRODUCT_INVENTORY_MIN_TTL_SECONDS=30
PRODUCT_INVENTORY_MAX_TTL_SECONDS=3600
PRODUCT_INVENTORY_EXPIRY_BATCH_LIMIT=100
PRODUCT_EVENTS_ENABLED=false
PRODUCT_OUTBOX_WORKER_ENABLED=false
```

### Step 6: Run migrations

```bash
mongosh "mongodb://localhost:27017" migrations/mongo/0003_create_product_collections.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0004_seller_product_workflow.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0005_product_read_indexes.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0006_inventory_reservations.up.js
```

Optional event/outbox migration:

```bash
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

So for now, verify through tests/build and MongoDB inspection.

### Step 8: Verify `TASK_FILE_NAME` functionality

```bash
go test ./internal/usecase -run 'ReserveInventory|ReleaseAndCommitInventory|ExpireReservations'
go test ./...
go build ./...
```

Task-specific behavior to verify in tests:

| Flow | Expected result |
|---|---|
| Reserve | Increases `reserved_quantity`, creates reservation, writes snapshot |
| Reserve retry with same order | Returns same reservation instead of double-reserving |
| Partial reserve failure | Rolls back already reserved items |
| Release | Decreases `reserved_quantity`, marks reservation released |
| Commit | Decreases `stock_quantity` and `reserved_quantity`, marks committed |
| Expiry | Releases expired reserved stock and marks reservation expired |
| Inventory event | Queues `ProductInventoryChanged` only when in-stock status changes and event recorder is enabled |

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
| MongoDB | `27017` | Product stock, reservations, snapshots, optional outbox | Reused, required for DB-backed verification |
| RabbitMQ AMQP | `5672` | Product inventory events | Reused, conditional |
| RabbitMQ Management UI | `15672` | Local broker debugging | Reused, optional |
| Backend HTTP API | Not available | No app server entrypoint yet | Missing |
| Product gRPC API | Not available | `ReserveInventory`, `ReleaseInventory`, `CommitInventory` are contract-level internal methods | Missing listener/registration |

How to change ports:

- MongoDB and RabbitMQ port change guidance is already in `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`, section `8. Ports and Networking`.
- For this module, no backend API port can be changed yet because no server entrypoint exists.

## 11. Common Errors & Fixes

Generic setup errors are already documented:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
12. Common Errors and Fixes
```

Task-specific errors:

| Error / symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `inventory_reservations` collection missing | `0006` migration not run | Run `0006_inventory_reservations.up.js` after `0003`-`0005` | Keep migration checklist per task |
| `reservation ttl is outside allowed range` | Request TTL is below min or above max env config | Use TTL between `PRODUCT_INVENTORY_MIN_TTL_SECONDS` and `PRODUCT_INVENTORY_MAX_TTL_SECONDS` | Keep default `900`, min `30`, max `3600` locally |
| Config validation fails for TTL vars | Env values are zero, negative, or ordered incorrectly | Fix `.env` values and reload shell env | Use the validation table in this file |
| Duplicate key on `order_id` | Same order is being reserved twice | Treat same order reserve as idempotent retry or use a new order id | Do not reuse order ids in manual tests |
| Duplicate key on `idempotency_key` | Same idempotency key reused for different order data | Use a unique idempotency key per checkout attempt | Generate request-scoped keys |
| `insufficient inventory available` | `stock_quantity - reserved_quantity - safety_stock` is less than requested quantity | Seed product variant with enough stock or reduce requested quantity | Remember safety stock is not sellable |
| `product or variant is not available` | Product is not `published` or variant is not `active` | Use a published product and active variant | Seed valid product lifecycle data |
| `reservation has expired` during commit | Commit happened after `expires_at` | Reserve again and complete checkout within TTL | Keep payment/checkout timeout below reservation TTL |
| Events enabled but no outbox collection | `PRODUCT_EVENTS_ENABLED=true` but `0007` not run | Run `0007_product_event_outbox.up.js` or disable events | Use isolated local env values unless testing events |
| RabbitMQ connection refused | Outbox worker enabled but RabbitMQ is not running or URL is wrong | Start RabbitMQ or set `PRODUCT_OUTBOX_WORKER_ENABLED=false` | Verify event mode before running app workers |
| `PRODUCT_EVENT_BROKER=kafka requires a ProductEventPublisher dependency` | Kafka selected but no default Kafka publisher is wired | Use RabbitMQ or disable events | Do not select Kafka for beginner local setup |
| `go run .` fails | No `main` package/server entrypoint | Use `go test ./...` and `go build ./...` | Wait for service runtime wiring task |

## 12. Security & Best Practices

General security guidance is reused:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
13. Security and Configuration Audit
14. Best Practices
```

Task-specific best practices:

- Keep `PRODUCT_MONGO_URI` out of committed files if it contains username/password.
- Keep RabbitMQ credentials local-only in `.env`; production should use secrets manager/CI/Kubernetes secrets.
- Use idempotency keys for checkout reserve calls. Retry safe behavior inventory correctness ke liye important hai.
- Keep reservation TTL short enough to prevent stock blocking, but long enough for payment completion.
- Run an expiry worker/scheduler before production traffic. MongoDB TTL cleanup alone stock release nahi karta.
- Monitor expired reservation counts, write conflicts, out-of-stock errors, and event dead letters.
- Do not expose MongoDB `27017` or RabbitMQ `5672` publicly.
- Keep event publishing disabled for isolated local inventory checks unless RabbitMQ/outbox setup is intentionally being tested.

## 13. Missing or Misconfigured Things

| Area | Current observation | Impact | Suggested fix |
|---|---|---|---|
| Server entrypoint | No `cmd/server/main.go` or equivalent exists | Backend API/gRPC process cannot be started | Add config load, Mongo client, handlers, gRPC/HTTP server, graceful shutdown |
| gRPC registration | Contract exists in `api/master-api.json`, but no generated proto/server registration seen in this module | Order Service cannot call inventory methods over network yet | Add protobuf definitions/generated code and register inventory handler |
| Expiry scheduler | `ExpireReservations` exists, but no background scheduler entrypoint is wired | Expired reservations may not release stock automatically | Add worker/ticker or queue-based expiry job |
| Dockerfile | Not present | App container cannot be built | Add Dockerfile after server entrypoint exists |
| docker-compose | Not present | Local infra remains manual | Add compose for MongoDB, optional RabbitMQ, and future app service |
| Migration runner | Migrations are plain `mongosh` scripts | Beginners must manually run scripts in order | Add Makefile/task runner or migration tool |
| Health checks | No app readiness endpoint yet | Mongo/RabbitMQ readiness cannot be checked through service | Add `/healthz` or gRPC health with Mongo ping and optional broker check |
| DB name in migrations | Scripts hardcode `product_db` | Env DB mismatch can confuse local setup | Keep local `PRODUCT_MONGO_DATABASE=product_db` or parameterize migrations |
| Event defaults | `.env.example` enables events and outbox worker | Local run may require RabbitMQ/outbox sooner than expected | Disable events for isolated tests or start RabbitMQ and run `0007` |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Project Tech Stack Analysis` | Same Go, MongoDB, Docker, RabbitMQ technology explanation already exists |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Language-Specific Dependency System: Go` | Same Go module commands and troubleshooting apply |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. Database Analysis: MongoDB` | MongoDB install, Docker run, connection string, credentials, and verification are already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `6. Environment Variables` | Full `.env` behavior and variable table already includes inventory/event variables |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. External Services Analysis` | RabbitMQ/Kafka/Redis status already explained |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `8. Ports and Networking` | Same MongoDB/RabbitMQ port guidance applies |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `9. Docker and DevOps Setup` | Dockerfile/compose limitations and local Docker commands are reused |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `10. Migration Setup` | Full current migration order, including `0006` and `0007`, is already listed |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `11. Complete Project Run Instructions` | Same onboarding flow applies, with this file adding inventory-specific verification |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `12. Common Errors and Fixes` | Generic Go/MongoDB/Docker/env/RabbitMQ errors are reused |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | MongoDB choice and `PRODUCT_MONGO_*` naming guidance is reused |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `5. Database Setup` | Base collections and `inventory_snapshots` setup are prerequisites |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | `7. Environment Variables` | Event local-mode caveats are reused when full app wiring is tested |
| `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | `5. Database Setup` | Existing products/read indexes through `0005` are prerequisites before inventory reservation migration |

## 15. Final Checklist

- [ ] Previous dependency documentation checked
- [ ] `INPUT_FILE_PATH` reviewed
- [ ] Original `TASK_FILE_NAME` not modified
- [ ] No duplicate Go/MongoDB/Docker/RabbitMQ setup copied unnecessarily
- [ ] `backend/services/product-service/.env` created and loaded if local config is needed
- [ ] `PRODUCT_MONGO_URI` points to the intended MongoDB instance
- [ ] `PRODUCT_MONGO_DATABASE=product_db` confirmed for local migrations
- [ ] MongoDB running locally or in Docker
- [ ] Migrations `0003`, `0004`, `0005`, and `0006` applied in order
- [ ] `0007_product_event_outbox.up.js` applied only if product events/outbox are enabled
- [ ] `inventory_reservations` indexes verified
- [ ] Inventory TTL env values verified
- [ ] Events disabled for isolated local inventory tests, or RabbitMQ started for event verification
- [ ] `go test ./internal/usecase -run 'ReserveInventory|ReleaseAndCommitInventory|ExpireReservations'` passed
- [ ] `go test ./...` passed
- [ ] `go build ./...` passed
- [ ] Missing server entrypoint/gRPC registration limitation understood
- [ ] Expiry worker/scheduler gap understood before production use
- [ ] Logs checked during inventory tests/debugging
- [ ] No original implementation task file overwritten
