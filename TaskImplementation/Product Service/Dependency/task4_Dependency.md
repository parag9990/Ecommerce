# Project Dependency & Setup Guide

## Variable Values Used

```text
SERVICE_NAME=Product Service
TASK_FILE_NAME=task4.md
INPUT_FILE_PATH=TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}
OUTPUT_FILE_NAME=task4_Dependency.md
OUTPUT_FILE_PATH=TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}
```

This file is generated for `INPUT_FILE_PATH` and saved as `OUTPUT_FILE_PATH`.

Beginner note:

- `TASK_FILE_NAME` explains seller product CRUD: create draft, update, publish, and unpublish lifecycle.
- Shared Go, MongoDB, Docker, RabbitMQ, ports, and generic `.env` setup already exist in previous dependency docs.
- This guide explains only the new or task-specific setup impact for seller CRUD, CMS moderation policy, and the `0004` MongoDB migration.
- The original `TASK_FILE_NAME` file is not changed.

## 1. Project Overview

`TASK_FILE_NAME` ka goal seller write workflow ko ready karna hai:

| Seller action | Result |
|---|---|
| Create product | New product `draft` status me save hota hai |
| Update product | Seller apne editable product ko update kar sakta hai |
| Publish product | Product validation + CMS moderation rule ke baad `published` ya `submitted` hota hai |
| Unpublish product | Live product `unpublished` status me move hota hai |

Simple Hinglish explanation:

Seller CRUD ka matlab yahan public product browse nahi hai. Iska focus seller ke product write actions par hai. Seller product create karega, data validate hoga, category/SKU check hoga, CMS policy decide karegi ki product direct publish hoga ya review me jayega.

### Current implementation reality

| Area | Current file / behavior | Setup impact |
|---|---|---|
| Seller usecase | `backend/services/product-service/internal/usecase/seller_product_service.go` | No new Go install, but CMS policy and validation env values matter |
| CMS dependency | `backend/services/product-service/internal/client/cms.go` | Currently static policy client; no real CMS server address is configured yet |
| DTO/handler layer | `backend/services/product-service/internal/transport/dto/seller_product.go`, `internal/transport/sellerproduct/handler.go` | In-process handler exists; no HTTP/gRPC server entrypoint exists yet |
| MongoDB repository | `internal/repository/mongo_product_repository.go` | Product writes need `products` and `categories` collections |
| Migration | `migrations/mongo/0004_seller_product_workflow.up.js` | Must run after `0003_create_product_collections.up.js` |
| Events/outbox | Optional current app wiring | If enabled with Mongo transactions, local MongoDB may need replica set mode |
| Service startup | No `cmd/server/main.go` | Use tests/build for verification until server entrypoint is added |

## 2. Tech Stack

Full technology explanation is already available here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
3. Project Tech Stack Analysis
4. Language-Specific Dependency System: Go
5. Database Analysis: MongoDB
7. External Services Analysis
```

Task-specific summary:

| Technology | Required? | Why it matters for `TASK_FILE_NAME` | New explanation needed? |
|---|---:|---|---|
| Go `1.26.3` | Yes | Seller CRUD usecase, validation, repository, and handlers are written in Go | No, reused |
| Go modules | Yes | `go.mod` / `go.sum` manage dependencies | No, reused |
| MongoDB | Yes for DB-backed verification | `products` collection stores seller-owned product documents | Partial, `0004` migration details below |
| MongoDB Go Driver v2 | Yes | Repository uses official driver for inserts, updates, queries, and transactions | No, reused |
| `mongosh` | Yes for manual migrations | Runs `0003` and `0004` migration scripts | No, reused |
| CMS policy client | Yes for publish decision | Current code uses static CMS policy env flags | Yes, task-specific behavior below |
| RabbitMQ | Conditional | Needed only if product events/outbox worker is enabled | Setup reused, `TASK_FILE_NAME` caveat below |
| Docker | Optional | Easiest way to run MongoDB/RabbitMQ locally | Setup reused |
| Redis/Kafka/Typesense | No direct requirement | Not needed for seller CRUD setup in `TASK_FILE_NAME` | No |

Important:

No Gin, Fiber, Echo, HTTP server, generated proto server, or runnable service binary is currently present in this module. API routes are documented in `api/master-api.json`, but a real server entrypoint still has to be added later.

## 3. Required Software

No brand-new software is introduced only by `TASK_FILE_NAME`.

Follow this previous setup first:

```text
This setup is already explained in:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
4. Language-Specific Dependency System: Go
5. Database Analysis: MongoDB
6. Environment Variables
8. Ports and Networking
9. Docker and DevOps Setup
10. Migration Setup
```

For `TASK_FILE_NAME` verification, a beginner needs:

| Software | Purpose | Status |
|---|---|---|
| Git | Clone repository | Reused |
| Go `1.26.3` or compatible installed toolchain | Run tests/build | Reused |
| MongoDB | Store `product_db.products` and `product_db.categories` | Reused, mandatory for DB verification |
| `mongosh` | Run migration scripts | Reused |
| Docker | Optional local MongoDB/RabbitMQ containers | Reused |
| RabbitMQ | Only if product events/outbox worker is enabled | Reused, conditional |

## 4. Dependency Management

Go dependency management is already documented here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
4. Language-Specific Dependency System: Go
```

### What changed in `TASK_FILE_NAME`?

No new Go module dependency is required only for seller CRUD.

Current direct dependencies are still:

| Dependency | Version | Used for |
|---|---:|---|
| `go.mongodb.org/mongo-driver/v2` | `v2.6.0` | MongoDB repository, collection access, transactions |
| `github.com/rabbitmq/amqp091-go` | `v1.10.0` | Product event publisher when outbox/RabbitMQ is enabled |

Do not run `go get` just for `TASK_FILE_NAME`. Agar dependencies missing hain, use:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
4. Language-Specific Dependency System: Go
```

Task-specific verification commands:

```bash
cd backend/services/product-service
go test ./internal/usecase -run 'CreateProduct|UpdateProduct|PublishProduct|UnpublishProduct'
go test ./...
go build ./...
```

## 5. Database Setup

MongoDB installation, Docker setup, connection string format, and generic troubleshooting are already explained here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
5. Database Analysis: MongoDB
```

### Database used

```text
product_db
```

MongoDB simple English:

MongoDB ek document database hai. Is service me product ka data nested format me store hota hai, jaise variants, images, price, category attributes, and lifecycle status.

### Why MongoDB is required for `TASK_FILE_NAME`

Seller CRUD writes these collections:

| Collection | Required? | Why |
|---|---:|---|
| `products` | Yes | Seller product document create/update/publish hota hai |
| `categories` | Yes | Category active hai ya nahi, attribute schema kya hai, ye validate hota hai |
| `product_event_outbox` | Conditional | Only if product events are enabled in current app wiring |

### Migration dependency

Run base collection migration first:

```text
backend/services/product-service/migrations/mongo/0003_create_product_collections.up.js
```

Then run the `TASK_FILE_NAME` migration:

```text
backend/services/product-service/migrations/mongo/0004_seller_product_workflow.up.js
```

`0004` does not create MongoDB from zero. It expects `products` collection to already exist. Agar `0003` pehle nahi chala, migration error dega:

```text
products collection must exist before applying seller product workflow migration
```

### What is new in `0004`

| Change | Why it matters |
|---|---|
| Adds `rejected` to allowed product statuses | CMS rejection ke baad seller product fix karke dobara publish kar sake |
| Requires `image_id` inside image documents | Service can track product images reliably |
| Adds image metadata fields | `alt_text`, `variant_ids`, `width`, `height` support better seller catalog management |

### Run migrations locally

Use this only after MongoDB is already running.

```bash
cd backend/services/product-service
mongosh "mongodb://localhost:27017" migrations/mongo/0003_create_product_collections.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0004_seller_product_workflow.up.js
```

If using the previous Docker container:

```bash
cd backend/services/product-service
docker exec -i ecommerce-product-mongo mongosh < migrations/mongo/0003_create_product_collections.up.js
docker exec -i ecommerce-product-mongo mongosh < migrations/mongo/0004_seller_product_workflow.up.js
```

Rollback only the `TASK_FILE_NAME` migration:

```bash
cd backend/services/product-service
mongosh "mongodb://localhost:27017" migrations/mongo/0004_seller_product_workflow.down.js
```

Warning:

The migration scripts use `db.getSiblingDB("product_db")`. Agar `PRODUCT_MONGO_DATABASE` ko change karte ho, migration target bhi review karna padega. Otherwise app ek DB use karega aur migration dusre DB me run ho sakti hai.

### Verify `TASK_FILE_NAME` migration

```bash
mongosh "mongodb://localhost:27017/product_db" --eval 'db.getCollectionInfos({name:"products"})[0].options.validator.$jsonSchema.properties.status.enum'
```

Expected status list should include:

```text
rejected
```

Check image required fields:

```bash
mongosh "mongodb://localhost:27017/product_db" --eval 'db.getCollectionInfos({name:"products"})[0].options.validator.$jsonSchema.properties.images.items.required'
```

Expected list should include:

```text
image_id
```

## 6. Redis / Queue / External Services

Generic external service setup is already explained here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
7. External Services Analysis
```

### CMS Service / CMS policy

`TASK_FILE_NAME` depends on CMS behavior, but current code does not call a real remote CMS service yet.

Current implementation:

```text
backend/services/product-service/internal/client/cms.go
```

Current behavior:

| CMS behavior | Controlled by env | Meaning |
|---|---|---|
| Seller catalog allowed? | `PRODUCT_CMS_CATALOG_MANAGEMENT_ALLOWED` | `false` blocks seller writes |
| Publish decision | `PRODUCT_CMS_MODERATION_DECISION` | `auto_publish` or `review_required` |
| Draft writes when CMS unavailable | `PRODUCT_CMS_ALLOW_DRAFT_WRITES_WHEN_UNAVAILABLE` | Draft create/update may continue if CMS check fails |

Simple Hinglish:

CMS Service abhi actual network dependency ke form me wired nahi hai. Code ek static policy client use karta hai. Ye local development ke liye easy hai, but production me real CMS gRPC/HTTP client chahiye hoga.

### RabbitMQ and outbox

RabbitMQ setup is reused from previous docs.

Task-specific caveat:

Seller create/update/publish can queue product events if `PRODUCT_EVENTS_ENABLED=true` and app wiring provides a product event recorder. Current `.env.example` keeps product events enabled. In that mode:

- `product_event_outbox` collection is needed.
- RabbitMQ is needed if the outbox worker runs.
- MongoDB transactions may be used for product + outbox writes.

For isolated `TASK_FILE_NAME` seller CRUD testing, simplest local setup is:

```env
PRODUCT_EVENTS_ENABLED=false
PRODUCT_OUTBOX_WORKER_ENABLED=false
```

Use events only when RabbitMQ and outbox setup are intentionally being tested.

### Kafka / Redis / Typesense

Not required for `TASK_FILE_NAME`.

If `PRODUCT_EVENT_BROKER=kafka`, current app wiring requires an injected `ProductEventPublisher`. There is no default Kafka publisher in this module right now, so beginners should keep:

```env
PRODUCT_EVENT_BROKER=rabbitmq
```

or disable events for `TASK_FILE_NAME`-only verification.

## 7. Environment Variables

Complete `.env` behavior and full variable table are already documented here:

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

Current code reads process environment variables using `os.LookupEnv`. Sirf `.env` file bana dene se values automatically load nahi hoti. Previous dependency doc ke `source .env` instructions follow karo.

### `TASK_FILE_NAME` values to check

Do not duplicate the full `.env`. Add or verify only these `TASK_FILE_NAME`-relevant values:

```env
# Seller CRUD validation and CMS policy
PRODUCT_CMS_CATALOG_MANAGEMENT_ALLOWED=true
PRODUCT_CMS_MODERATION_DECISION=auto_publish
PRODUCT_CMS_ALLOW_DRAFT_WRITES_WHEN_UNAVAILABLE=true

# Publish validation behavior
PRODUCT_REQUIRE_PRIMARY_IMAGE_FOR_PUBLISH=false

# Local TASK_FILE_NAME-only testing shortcut if RabbitMQ/outbox is not being tested
PRODUCT_EVENTS_ENABLED=false
PRODUCT_OUTBOX_WORKER_ENABLED=false
```

If you want to test moderation queue behavior instead of direct publish:

```env
PRODUCT_CMS_MODERATION_DECISION=review_required
```

Then `PublishProduct` should move product status to:

```text
submitted
```

### Variable explanation

| Variable | Required? | Example | Purpose | Security note |
|---|---:|---|---|---|
| `PRODUCT_CMS_CATALOG_MANAGEMENT_ALLOWED` | Yes for current static CMS policy | `true` | Allows seller catalog create/update/publish | Not secret, but risky if wrong in production |
| `PRODUCT_CMS_MODERATION_DECISION` | Yes for publish flow | `auto_publish` | Decides `published` vs `submitted` | Not secret, but controls moderation behavior |
| `PRODUCT_CMS_ALLOW_DRAFT_WRITES_WHEN_UNAVAILABLE` | Optional | `true` | Allows draft writes when CMS check errors | Production should review carefully |
| `PRODUCT_REQUIRE_PRIMARY_IMAGE_FOR_PUBLISH` | Optional | `false` | If `true`, publish requires an active primary image | Not secret |
| `PRODUCT_EVENTS_ENABLED` | Optional | `false` for isolated `TASK_FILE_NAME` tests | Enables product event recording/outbox | Not secret |
| `PRODUCT_OUTBOX_WORKER_ENABLED` | Optional | `false` for isolated `TASK_FILE_NAME` tests | Enables outbox relay worker when app entrypoint runs it | Not secret |

Important allowed values:

```text
PRODUCT_CMS_MODERATION_DECISION=auto_publish
PRODUCT_CMS_MODERATION_DECISION=review_required
```

Any other value fails config validation.

## 8. Docker Setup

Docker basics are already explained here:

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
| MongoDB container | No | Reuse previous MongoDB setup |
| RabbitMQ container | Conditional | Needed only if events/outbox worker is enabled |
| Product service container | No | No Dockerfile exists yet |
| Compose network | No | No repo-owned compose file exists yet |
| MongoDB replica set | Conditional | Needed for transaction-backed product+outbox writes |

### MongoDB transaction warning

Current repository can use MongoDB transactions when product writes are saved with outbox events. Standalone MongoDB often fails transactions with an error like:

```text
Transaction numbers are only allowed on a replica set member or mongos
```

Beginner-friendly choices:

| Choice | When to use |
|---|---|
| Set `PRODUCT_EVENTS_ENABLED=false` | Best for `TASK_FILE_NAME`-only local seller CRUD tests |
| Run MongoDB as a replica set | Use when testing product+outbox transaction behavior |

Minimal local replica-set example:

```bash
docker volume create product_mongo_rs_data
docker run -d \
  --name ecommerce-product-mongo-rs \
  -p 27017:27017 \
  -v product_mongo_rs_data:/data/db \
  mongo:8 --replSet rs0 --bind_ip_all

docker exec ecommerce-product-mongo-rs mongosh --eval 'rs.initiate({_id:"rs0",members:[{_id:0,host:"localhost:27017"}]})'
```

Note:

If the app runs inside another container, the replica-set host should be reachable from that container, not only from the host machine. For beginners, disabling events during `TASK_FILE_NAME`-only tests is simpler.

## 9. Local Development Setup

This flow avoids duplicate setup and focuses only on `TASK_FILE_NAME`.

### Step 1: Read previous dependency documentation

Read these first:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
```

Most important reused sections:

- Go dependency setup
- MongoDB install/Docker setup
- `.env` loading
- base `product_db` setup
- generic Docker/MongoDB/RabbitMQ troubleshooting

### Step 2: Go to project directory

```bash
cd backend/services/product-service
```

### Step 3: Install only new dependencies if any

No new dependency is added by `TASK_FILE_NAME`.

If module cache is empty:

```bash
go mod download
```

Use previous dependency docs for detailed Go module troubleshooting.

### Step 4: Setup only new databases/services if any

No new database type is added.

Make sure MongoDB is running, then run `0003` and `0004` migrations in order.

### Step 5: Add only new or changed environment variables

Create or update:

```text
backend/services/product-service/.env
```

For `TASK_FILE_NAME`-only local verification, use:

```env
PRODUCT_MONGO_URI=mongodb://localhost:27017
PRODUCT_MONGO_DATABASE=product_db
PRODUCT_MONGO_AUTO_CREATE_COLLECTIONS=false
PRODUCT_CMS_CATALOG_MANAGEMENT_ALLOWED=true
PRODUCT_CMS_MODERATION_DECISION=auto_publish
PRODUCT_CMS_ALLOW_DRAFT_WRITES_WHEN_UNAVAILABLE=true
PRODUCT_EVENTS_ENABLED=false
PRODUCT_OUTBOX_WORKER_ENABLED=false
```

### Step 6: Run migrations

```bash
mongosh "mongodb://localhost:27017" migrations/mongo/0003_create_product_collections.up.js
mongosh "mongodb://localhost:27017" migrations/mongo/0004_seller_product_workflow.up.js
```

### Step 7: Start backend service

Current state:

There is no runnable server entrypoint yet.

This is expected to fail:

```bash
go run .
```

Use these checks instead:

```bash
go test ./...
go build ./...
```

### Step 8: Verify functionality related to `TASK_FILE_NAME`

Run seller CRUD unit tests:

```bash
go test ./internal/usecase -run 'CreateProduct|UpdateProduct|PublishProduct|UnpublishProduct'
```

Verify migration:

```bash
mongosh "mongodb://localhost:27017/product_db" --eval 'db.getCollectionInfos({name:"products"})[0].options.validator.$jsonSchema.properties.status.enum'
```

Expected:

- `rejected` status is present.
- Product image schema requires `image_id`.

## 10. Running the Project

### Current run reality

| Command | Expected result |
|---|---|
| `go test ./...` | Valid package-level verification |
| `go build ./...` | Valid package build verification |
| `go run .` | Not valid yet, because no `main` package exists |
| Seller REST API call | Not available directly from this module yet |

### Planned API surface from `api/master-api.json`

| Action | Method | Path | gRPC method | Auth |
|---|---|---|---|---|
| Create seller product | `POST` | `/api/v1/seller/products` | `ProductService.CreateProduct` | seller |
| Update seller product | `PATCH` | `/api/v1/seller/products/{product_id}` | `ProductService.UpdateProduct` | seller |
| Publish seller product | `POST` | `/api/v1/seller/products/{product_id}/publish` | `ProductService.PublishProduct` | seller |

Current module has DTO and handler methods, but not an actual HTTP or gRPC listener.

### Ports and networking

Detailed port guidance is already documented here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
8. Ports and Networking
```

Task-specific port table:

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Backend HTTP API | `8080` | Planned gateway/service HTTP port from architecture docs | Reused / not active in this module |
| Backend gRPC | `9090` | Planned internal gRPC port from architecture docs | Reused / not active in this module |
| MongoDB | `27017` | `product_db` storage | Reused, required for DB verification |
| RabbitMQ AMQP | `5672` | Product event publishing | Reused, conditional |
| RabbitMQ UI | `15672` | Local broker management UI | Reused, optional |
| CMS Service | Not configured | Current implementation uses static policy client | No new port |

No new port is introduced by `TASK_FILE_NAME`.

## 11. Common Errors & Fixes

Generic setup errors are already documented here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Section:
12. Common Errors and Fixes
```

Task-specific errors:

| Error / confusion | Cause | Fix | Prevention |
|---|---|---|---|
| `products collection must exist before applying seller product workflow migration` | `0004` was run before `0003` | Run `0003_create_product_collections.up.js`, then rerun `0004` | Always run migrations in numeric order |
| `PRODUCT_CMS_MODERATION_DECISION must be one of auto_publish or review_required` | Env value is invalid | Set `auto_publish` or `review_required` | Copy exact values from this guide |
| Product stays `submitted` after publish | CMS moderation decision is `review_required` | This is expected; CMS approval flow must publish later | Use `auto_publish` only for local direct-publish tests |
| `SELLER_CATALOG_DISABLED` | `PRODUCT_CMS_CATALOG_MANAGEMENT_ALLOWED=false` | Set it to `true` locally | Keep static CMS policy explicit per environment |
| `CMS_UNAVAILABLE` during publish | CMS client returned error or policy check unavailable | Fix CMS wiring/policy; do not force publish blindly | In production fail closed for publish |
| `PRODUCT_OWNERSHIP_MISMATCH` | Actor seller id does not match product seller id | Use correct authenticated seller context | Never pass seller id from request body as source of truth |
| `DUPLICATE_SKU` | Variant SKU already exists in another product | Use a unique SKU or clean test data | Generate unique SKUs for tests/seeds |
| `PRODUCT_NOT_EDITABLE` | Seller tried to edit `published` or non-editable status | Unpublish first or use allowed lifecycle | Follow status transition rules |
| `INVALID_STATUS_TRANSITION` | Publish/unpublish called from invalid status | Move product to draft/rejected/unpublished before publish | Validate current status before command |
| `Transaction numbers are only allowed on a replica set member or mongos` | Events/outbox path used Mongo transaction on standalone MongoDB | Disable events for `TASK_FILE_NAME`-only tests or run MongoDB replica set | Document local mode clearly in `.env` |
| RabbitMQ connection/publish errors | Outbox worker enabled but RabbitMQ not running or `RABBITMQ_URL` wrong | Start RabbitMQ or set `PRODUCT_EVENTS_ENABLED=false` for `TASK_FILE_NAME`-only testing | Keep outbox disabled until broker setup is ready |
| `go run .` fails | No `main` package/server entrypoint exists | Use `go test ./...` and `go build ./...` | Add server entrypoint in future service work |
| Direct Mongo insert fails because `image_id` is missing | `0004` requires image IDs | Insert through service code or include `image_id` manually | Avoid raw DB inserts for product documents |

## 12. Security & Best Practices

General security guidance is already documented here:

```text
Refer:
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md

Sections:
13. Security and Configuration Audit
14. Best Practices
```

Task-specific guidance:

- Do not trust `seller_id` from request body. Seller identity should come from authenticated actor context.
- Keep `PRODUCT_CMS_MODERATION_DECISION=review_required` for cautious staging/production setups until real CMS moderation is wired.
- Use `auto_publish` only when business has approved direct publish behavior.
- Keep `PRODUCT_CMS_ALLOW_DRAFT_WRITES_WHEN_UNAVAILABLE=true` only if draft writes are safe during CMS outage. Publish should fail closed.
- Keep `PRODUCT_EVENTS_ENABLED=false` for isolated `TASK_FILE_NAME` local tests if RabbitMQ/outbox is not part of the test.
- If product events are enabled, use MongoDB replica set mode for transaction-backed product+outbox writes.
- Seed valid active categories before testing create/publish, because category validation can fail if category is missing or inactive.
- Use unique SKU values in local tests to avoid duplicate key confusion.
- Do not expose MongoDB `27017` or RabbitMQ `5672` publicly.
- Keep real `PRODUCT_MONGO_URI` and `RABBITMQ_URL` credentials out of committed files. `.env.example` is fine; `.env` should stay ignored.

## 13. Missing or Misconfigured Things

| Area | Current observation | Impact | Suggested fix |
|---|---|---|---|
| Real CMS integration | Current code uses `StaticCMSPolicyClient`; no `CMS_GRPC_ADDR` or real CMS client exists | CMS dependency is simulated by env flags | Add real CMS client and env vars when CMS Service is ready |
| Server entrypoint | No `cmd/server/main.go` or equivalent exists | No backend API process can be started yet | Add config load, Mongo connection, route/gRPC server, health checks |
| Proto/gRPC transport | Handler layer exists, but generated service server is not present | API gateway cannot call this module directly yet | Add proto definitions/generated handlers or gateway adapter |
| Dockerfile | Not present | App container cannot be built | Add Dockerfile after server entrypoint exists |
| docker-compose | No repo-owned compose file for this service | Local infra setup remains manual | Add local compose for MongoDB, optional replica set, and RabbitMQ |
| Migration runner | Migrations are plain `mongosh` scripts | Beginners must run scripts manually and in order | Add Makefile target or small migration runner |
| DB name in migrations | Scripts hard-code `product_db` | Env DB mismatch can confuse setup | Keep local `PRODUCT_MONGO_DATABASE=product_db` or parameterize migrations |
| Default events vs local Mongo | `.env.example` enables product events, but standalone MongoDB can fail transaction path | Local seller writes may fail if outbox is enabled | Disable events for `TASK_FILE_NAME`-only tests or document replica-set setup |
| Seed data | No beginner seed script for categories/brands | Create/publish tests need valid categories | Add seed script for active category and sample brand |
| Health checks | No health/readiness endpoint exists | Dependency status is hard to verify | Add `/healthz` or gRPC health after server implementation |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Project Tech Stack Analysis` | Same Go, MongoDB, RabbitMQ, Docker technology explanation already exists |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Language-Specific Dependency System: Go` | Same Go module setup applies |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. Database Analysis: MongoDB` | MongoDB install, Docker, connection string, and verification are already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `6. Environment Variables` | Complete `.env` behavior and full variable table already exist |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. External Services Analysis` | RabbitMQ/Kafka/Redis status is already explained |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `8. Ports and Networking` | Same MongoDB/RabbitMQ/general port guidance applies |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `9. Docker and DevOps Setup` | Same missing Dockerfile/compose notes apply |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `10. Migration Setup` | Base migration workflow is reused |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `12. Common Errors and Fixes` | Generic Go/MongoDB/Docker/env errors are already documented |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | MongoDB choice and `PRODUCT_MONGO_*` naming notes are reused |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `7. Environment Variables` | DB env naming mismatch guidance is reused |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `5. Database Setup` | Base `products` and `categories` collection setup is required before `TASK_FILE_NAME` |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `9. Local Development Setup` | Migration verification flow is reused and extended with `0004` |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `13. Missing or Misconfigured Things` | Server entrypoint, Dockerfile, compose, and health-check gaps are unchanged |

## 15. Final Checklist

- [ ] Previous dependency documentation checked
- [ ] No duplicate Go/MongoDB/Docker setup repeated
- [ ] `0003_create_product_collections.up.js` run before `0004`
- [ ] `0004_seller_product_workflow.up.js` applied successfully
- [ ] `products` validator includes `rejected` status
- [ ] `products.images` validator includes required `image_id`
- [ ] `.env` created at `backend/services/product-service/.env`
- [ ] `PRODUCT_CMS_CATALOG_MANAGEMENT_ALLOWED` configured
- [ ] `PRODUCT_CMS_MODERATION_DECISION` set to `auto_publish` or `review_required`
- [ ] `TASK_FILE_NAME`-only local tests either disable events or use transaction-capable MongoDB
- [ ] RabbitMQ running only if events/outbox worker is enabled
- [ ] Seller CRUD unit tests executed
- [ ] `go test ./...` executed
- [ ] `go build ./...` executed
- [ ] No real secrets committed
- [ ] No original implementation task file overwritten
