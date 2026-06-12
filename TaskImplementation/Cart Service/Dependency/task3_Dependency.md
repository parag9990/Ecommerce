# ${SERVICE_NAME} ${TASK_FILE_NAME} - Project Dependency & Setup Guide

```text
SERVICE_NAME = "Cart Service"
TASK_FILE_NAME = "task3.md"
INPUT_FILE_PATH = TaskImplementation/${SERVICE_NAME}/${TASK_FILE_NAME}
OUTPUT_FILE_NAME = ${TASK_FILE_NAME without ".md"}_Dependency.md
OUTPUT_FILE_PATH = TaskImplementation/${SERVICE_NAME}/${OUTPUT_FILE_NAME}

PREVIOUS_TASK1_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task1_Dependency.md
PREVIOUS_TASK2_DEPENDENCY_FILE = TaskImplementation/${SERVICE_NAME}/task2_Dependency.md
```

## 1. Project Overview

Ye document `${SERVICE_NAME}` ke `${TASK_FILE_NAME}` ke liye dependency, setup, environment, database, Docker, DevOps, aur troubleshooting guide hai.

`${TASK_FILE_NAME}` ka main focus runtime business logic nahi hai. Is task ka real setup impact hai:

- MongoDB database `cart_db` me `carts` collection ready karna.
- `carts` collection ke JSON Schema validator ko apply karna.
- Required indexes create karna, including owner lookup indexes, unique active cart indexes, TTL index, and item lookup index.
- Existing MongoDB plus Redis setup ko reuse karna.
- Beginner developer ko clear batana ki kya new hai aur kya previous dependency docs se follow karna hai.

Important scope:

- `${INPUT_FILE_PATH}` original implementation guide hai. Use modify nahi kiya gaya.
- `${OUTPUT_FILE_PATH}` new dependency/setup guide hai.
- Current repo me `backend/services/cart-service` Go implementation already present hai.
- `${TASK_FILE_NAME}` koi new database engine, queue, Docker service, port, ya Go package introduce nahi karta.

### Files Inspected

| File | Why inspected |
|---|---|
| `INPUT_FILE_PATH` | `${TASK_FILE_NAME}` scope: MongoDB `carts` collection and embedded `items[]` schema |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | Base setup: Go, MongoDB, Redis, `.env`, Docker, run commands |
| `PREVIOUS_TASK2_DEPENDENCY_FILE` | MongoDB source of truth plus Redis cache split |
| `backend/services/cart-service/go.mod` | Go version and direct dependency check |
| `backend/services/cart-service/.env.example` | Actual environment variable names |
| `backend/services/cart-service/internal/config/config.go` | Required env validation and defaults |
| `backend/services/cart-service/internal/domain/collection.go` | Collection names, max items, index spec |
| `backend/services/cart-service/internal/domain/cart.go` | Cart document, item, totals, coupon preview structs |
| `backend/services/cart-service/internal/repository/mongo_schema.go` | Implemented MongoDB validator and index models |
| `backend/services/cart-service/internal/repository/mongo_collection_manager.go` | Runtime schema bootstrap behavior |
| `backend/services/cart-service/migrations/mongo/001_create_carts_collection.up.js` | Manual MongoDB migration |
| `backend/services/cart-service/migrations/mongo/001_create_carts_collection.down.js` | Rollback behavior |
| `backend/services/cart-service/cmd/server/main.go` | Startup dependencies and schema bootstrap option |
| `backend/services/cart-service/cmd/cart-expiry-worker/main.go` | Worker dependency check |
| `backend/services/cart-service/internal/transport/http/handler.go` | Schema, health, readiness, and cart routes |

## 2. Tech Stack

No new technology is added by `${TASK_FILE_NAME}`. Same `${SERVICE_NAME}` runtime stack reused hai.

| Technology | Required? | `${TASK_FILE_NAME}` role | Full setup reference |
|---|---:|---|---|
| Go `1.26.3` | Yes | `${SERVICE_NAME}` code, schema bootstrap API, repository code | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `4. Go Dependency System` |
| Go modules | Yes | MongoDB and Redis libraries manage karta hai | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `4. Go Dependency System` |
| Standard `net/http` | Yes | `/internal/v1/cart/schema`, `/internal/v1/cart/schema/bootstrap`, `/healthz`, `/readyz` | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `3. High-Level Tech Stack` |
| MongoDB | Yes | `cart_db.carts` durable cart collection | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `5.1 MongoDB` |
| MongoDB JSON Schema validator | Yes for `${TASK_FILE_NAME}` | Invalid cart document shapes ko DB level par reject karta hai | Explained in this file |
| MongoDB indexes | Yes for `${TASK_FILE_NAME}` | Fast lookup, duplicate active cart guard, expiry cleanup support | Explained in this file |
| Redis | Yes for current service startup | Cache and worker lock reused; `${TASK_FILE_NAME}` schema does not add new Redis keys | `PREVIOUS_TASK2_DEPENDENCY_FILE`, section `6.2 Redis` |
| Product Service HTTP API | Required by current API startup/config | Add-item product snapshot dependency, not newly added by `${TASK_FILE_NAME}` | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `6.1 Product Service` |
| CMS Service HTTP API | Required by current API startup/config | Coupon preview dependency, not newly added by `${TASK_FILE_NAME}` | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `6.2 CMS Service` |
| Docker | Optional but recommended | Local MongoDB and Redis run karne ke liye | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `9. Docker and DevOps Setup` |

Simple Hinglish:

- MongoDB permanent cart data rakhta hai.
- Redis fast cache hai, source of truth nahi.
- `${TASK_FILE_NAME}` ka new important part hai MongoDB collection shape: `carts` document ke andar `items[]` embedded cart item snapshots.

## 3. Required Software

Base software setup already documented hai. Duplicate install commands yahan repeat nahi kiye gaye.

| Software | Required? | Status for `${TASK_FILE_NAME}` | Refer |
|---|---:|---|---|
| Git | Yes | Reused | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `10. Project Run Instructions` |
| Go `1.26.3` | Yes | Reused | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `4.2 Go Version` |
| MongoDB server | Yes | Reused database, `${TASK_FILE_NAME}` adds schema/index requirements | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `5.1 MongoDB` |
| `mongosh` | Strongly recommended | Manual migration and schema verification ke liye useful | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `5.1 MongoDB` |
| Redis server | Yes for current service startup | Reused | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `5.2 Redis` |
| Docker / Docker Compose | Optional but recommended | Reused for local MongoDB and Redis | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `9. Docker and DevOps Setup` |
| `curl` | Optional but useful | Schema API and health check verify karne ke liye | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `3. High-Level Tech Stack` |

`${TASK_FILE_NAME}`-specific required software reminder:

- Agar schema manually apply karna hai, `mongosh` chahiye.
- Agar schema API se bootstrap karna hai, `${SERVICE_NAME}` HTTP API running honi chahiye and MongoDB reachable honi chahiye.

## 4. Dependency Management

This is a Go project. `${TASK_FILE_NAME}` does not require any new `go get`.

Current direct dependencies in `backend/services/cart-service/go.mod`:

| Dependency | Version | Status | Why |
|---|---:|---|---|
| `go.mongodb.org/mongo-driver` | `v1.17.9` | Reused | MongoDB collection, validator, indexes, and repository |
| `github.com/redis/go-redis/v9` | `v9.19.0` | Reused | Redis cache and expiry worker lock |

Do not install extra packages only because `${TASK_FILE_NAME}` mentions future tools. Example: `github.com/google/uuid` is not used by current implementation. Current code generates cart/item IDs with `crypto/rand`, so no UUID package is needed for `${TASK_FILE_NAME}`.

For Go module commands, refer:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 4.4 Dependency Commands
```

Useful check from service folder:

```bash
cd backend/services/cart-service
go test ./...
```

Beginner note:

`go.mod` dependencies project ke liye declared hote hain. Agar unnecessary `go get` run kar doge, module file dirty ho jayegi aur future developer confuse ho sakta hai.

## 5. Database Setup

MongoDB is mandatory for `${TASK_FILE_NAME}`.

MongoDB ek document database hai. Simple words me, cart ka ek full JSON-like document `carts` collection me store hota hai. Cart ke items separate table/collection me nahi ja rahe; same cart document ke andar `items[]` array me embedded snapshots ke form me rahenge.

Base MongoDB install, Docker run, connection string, and basic verification already explained hai:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 5.1 MongoDB
```

### `${TASK_FILE_NAME}` Database Ownership

| Setting | Value |
|---|---|
| Database | `cart_db` |
| Collection | `carts` |
| Owner service | `${SERVICE_NAME}` |
| Source of truth | MongoDB |
| Cache companion | Redis |

Important:

- `CART_MONGO_DATABASE` must be exactly `cart_db`.
- Current config rejects any other database name.
- Other services should not directly write `cart_db.carts`; they should call `${SERVICE_NAME}` APIs.

### `carts` Document Shape

High-level document:

```json
{
  "_id": "cart_123",
  "user_id": "user_123",
  "guest_session_id": null,
  "status": "active",
  "items": [],
  "coupon_code": null,
  "coupon_preview": null,
  "totals": {},
  "version": 1,
  "created_at": "2026-05-24T00:00:00Z",
  "updated_at": "2026-05-24T00:00:00Z",
  "expires_at": "2026-08-22T00:00:00Z"
}
```

Main fields:

| Field | Required? | Purpose |
|---|---:|---|
| `_id` | Yes | Cart ID, API me `cart_id` banega |
| `user_id` | Conditional | Logged-in shopper owner |
| `guest_session_id` | Conditional | Guest shopper owner |
| `status` | Yes | `active`, `merged`, `checked_out`, `expired`, `abandoned` |
| `items` | Yes | Embedded cart item snapshots |
| `coupon_code` | No | Applied coupon code, if preview exists |
| `coupon_preview` | No | CMS coupon validation snapshot |
| `totals` | Yes | Subtotal, discount, total, currency, counts |
| `version` | Yes | Optimistic concurrency control |
| `created_at` | Yes | Cart create time |
| `updated_at` | Yes | Last cart mutation time |
| `expires_at` | Yes | TTL and expiry cleanup reference |
| `merged_into_cart_id` | No | Guest cart merge target |
| `checked_out_order_id` | No | Checkout reference |

### Embedded `items[]` Shape

Cart item snapshot:

```json
{
  "item_id": "item_1",
  "product_id": "prod_123",
  "variant_id": "var_1",
  "seller_id": "seller_456",
  "sku_snapshot": "SHOE-BLK-9",
  "title_snapshot": "Running Shoes",
  "image_url_snapshot": "https://cdn.example.com/prod_123/main.jpg",
  "variant_snapshot": {
    "size": "9",
    "color": "Black"
  },
  "unit_price": {
    "amount": 299900,
    "currency": "INR"
  },
  "quantity": 2,
  "line_subtotal": {
    "amount": 599800,
    "currency": "INR"
  },
  "price_snapshot_at": "2026-05-24T00:00:00Z",
  "added_at": "2026-05-24T00:00:00Z",
  "updated_at": "2026-05-24T00:00:00Z"
}
```

Task-specific rules:

- Max unique items per cart: `100`.
- Quantity per item: min `1`, max `10`.
- Same `product_id + variant_id` duplicate line nahi banana; quantity update/combine karo.
- Money integer minor units me store hota hai. Example: INR 2999.00 = `299900`.
- Currency uppercase 3-letter ISO code hona chahiye, for example `INR`.

### Required Indexes

Current implementation and migration create these indexes:

| Index | Unique? | TTL? | Purpose |
|---|---:|---:|---|
| `idx_carts_user_status` | No | No | Logged-in active cart lookup |
| `idx_carts_guest_status` | No | No | Guest active cart lookup |
| `idx_carts_expires_at_ttl` | No | Yes | Expired cart cleanup support |
| `uniq_active_cart_per_user` | Yes | No | One active cart per logged-in user |
| `uniq_active_cart_per_guest_session` | Yes | No | One active cart per guest session |
| `idx_carts_item_product_variant` | No | No | Product/variant lookup and debugging |
| `idx_carts_updated_at` | No | No | Cleanup and abandoned-cart scans |

### Migration / Bootstrap Options

Full migration explanation is already available in:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 5.1 MongoDB
Topic: MongoDB Migration
```

`${TASK_FILE_NAME}`-specific options:

Option 1: run manual migration with `mongosh`.

```bash
cd backend/services/cart-service
mongosh "mongodb://ecommerce_root:ecommerce_password@localhost:27017/cart_db?authSource=admin" \
  migrations/mongo/001_create_carts_collection.up.js
```

Option 2: start service and call internal bootstrap endpoint.

```bash
curl -X POST http://localhost:8084/internal/v1/cart/schema/bootstrap
```

Option 3: local-only startup bootstrap.

```env
CART_BOOTSTRAP_COLLECTIONS_ON_STARTUP=true
```

Production note:

Use controlled migration jobs in production. App startup bootstrap is convenient for local development, but production me random startup migration risky hota hai.

### Verify Collection and Validator

Check collection exists:

```bash
mongosh "mongodb://ecommerce_root:ecommerce_password@localhost:27017/cart_db?authSource=admin" \
  --quiet \
  --eval "db.getSiblingDB('cart_db').getCollectionNames()"
```

Check validator:

```bash
mongosh "mongodb://ecommerce_root:ecommerce_password@localhost:27017/cart_db?authSource=admin" \
  --quiet \
  --eval "db.getSiblingDB('cart_db').getCollectionInfos({ name: 'carts' })[0].options.validator"
```

Check indexes:

```bash
mongosh "mongodb://ecommerce_root:ecommerce_password@localhost:27017/cart_db?authSource=admin" \
  --quiet \
  --eval "db.getSiblingDB('cart_db').carts.getIndexes().map(i => i.name)"
```

Expected important names:

```text
idx_carts_user_status
idx_carts_guest_status
idx_carts_expires_at_ttl
uniq_active_cart_per_user
uniq_active_cart_per_guest_session
idx_carts_item_product_variant
idx_carts_updated_at
```

## 6. Redis / Queue / External Services

`${TASK_FILE_NAME}` does not introduce a new external service.

| Service | Required? | Status | Setup reference |
|---|---:|---|---|
| Redis | Yes for current service startup | Reused cache and worker lock | `PREVIOUS_TASK2_DEPENDENCY_FILE`, section `6.2 Redis` |
| Product Service | Required by current API startup/config | Reused product snapshot dependency | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `6.1 Product Service` |
| CMS Service | Required by current API startup/config | Reused coupon preview dependency | `PREVIOUS_TASK1_DEPENDENCY_FILE`, section `6.2 CMS Service` |
| Kafka | No | Not used by `${TASK_FILE_NAME}` | Not required |
| RabbitMQ | No | Not used by `${TASK_FILE_NAME}` | Not required |
| NATS | No | Not used by `${TASK_FILE_NAME}` | Not required |
| MinIO / S3 | No | Not used by `${TASK_FILE_NAME}` | Not required |
| SMTP / Stripe / Twilio / OAuth | No | Not used by `${TASK_FILE_NAME}` | Not required |

Redis reminder:

- MongoDB = truth.
- Redis = speed.
- `${TASK_FILE_NAME}` schema does not add new Redis keys.
- Existing Redis key patterns are already documented in `PREVIOUS_TASK2_DEPENDENCY_FILE`, section `6.2 Redis`.

## 7. Environment Variables

`${TASK_FILE_NAME}` does not add new environment variables. It reuses existing MongoDB and bootstrap-related env vars.

Full `.env` setup is already documented in:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 8. Complete Environment Variables
```

Create `.env` here:

```text
backend/services/cart-service/.env
```

Important beginner reminder:

Current Go code uses `os.Getenv`. There is no automatic `.env` loader. Sirf `.env` file create karna enough nahi hai; command run karne se pehle env source/export karna hoga.

`${TASK_FILE_NAME}`-specific env focus:

```env
CART_MONGO_URI=mongodb://ecommerce_root:ecommerce_password@localhost:27017/cart_db?authSource=admin
CART_MONGO_DATABASE=cart_db
CART_BOOTSTRAP_COLLECTIONS_ON_STARTUP=false
CART_MONGO_CONNECT_TIMEOUT=5s
CART_MONGO_PING_TIMEOUT=5s
```

| Variable | Required? | `${TASK_FILE_NAME}` meaning | Security note |
|---|---:|---|---|
| `CART_MONGO_URI` | Yes | MongoDB connection string for schema/bootstrap and cart repository | Contains DB password; do not commit real secrets |
| `CART_MONGO_DATABASE` | Yes | Must be `cart_db` | Not secret, but must not be changed |
| `CART_BOOTSTRAP_COLLECTIONS_ON_STARTUP` | Optional | Auto-runs schema/index bootstrap at service or worker startup | Keep `false` in production |
| `CART_MONGO_CONNECT_TIMEOUT` | Optional | Mongo connect timeout | Not secret |
| `CART_MONGO_PING_TIMEOUT` | Optional | Mongo health/readiness timeout | Not secret |

Common env mistakes:

- `CART_MONGO_DATABASE` set to `ecommerce` or `cart`. Current config fails; use `cart_db`.
- `.env` file exists but is not sourced.
- Docker container network uses `mongo:27017`, but host-run service should use `localhost:27017`.
- Host machine uses `localhost:27017`, but service inside Compose should use `mongo:27017`.
- Duration values like `5` are invalid for Go duration config; use `5s`, `700ms`, `15m`, `2160h`.

## 8. Docker Setup

Docker setup is reused. `${TASK_FILE_NAME}` does not add a new Docker container or Dockerfile.

Refer:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 9. Docker and DevOps Setup
```

Current repo inspection notes:

- No `backend/services/cart-service/deploy/Dockerfile` is present.
- No actual `infra/compose/docker-compose.local.yml` file is present in the repo at inspection time.
- Earlier dependency docs provide local Docker examples for MongoDB and Redis.
- For this task, Docker ka practical use hai MongoDB and Redis dependencies run karna.

`${TASK_FILE_NAME}` Docker impact:

| Docker area | Status |
|---|---|
| MongoDB container | Reused |
| Redis container | Reused |
| `${SERVICE_NAME}` container | Not added by `${TASK_FILE_NAME}` |
| New volume | Not added by `${TASK_FILE_NAME}` |
| New network | Not added by `${TASK_FILE_NAME}` |
| New exposed port | Not added by `${TASK_FILE_NAME}` |
| New health check | Not added by `${TASK_FILE_NAME}` |

Beginner note:

Docker se MongoDB and Redis run karna easy hai, but `${SERVICE_NAME}` current local flow me direct `go run` se chal sakta hai.

## 9. Local Development Setup

Use this onboarding flow:

### Step 1: Read previous dependency docs first

Follow these in order:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
PREVIOUS_TASK2_DEPENDENCY_FILE
```

Why:

- Task 1 dependency doc full base setup deta hai.
- Task 2 dependency doc MongoDB vs Redis responsibility clear karta hai.
- `${TASK_FILE_NAME}` only schema/index setup detail add karta hai.

### Step 2: Go to service directory

```bash
cd backend/services/cart-service
```

### Step 3: Install dependencies

No new `${TASK_FILE_NAME}` dependency install needed. If dependencies are not downloaded:

```bash
go mod download
```

### Step 4: Start reused databases/services

Start MongoDB and Redis using previous dependency docs.

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 5.1 MongoDB
Section: 5.2 Redis
```

### Step 5: Create and source `.env`

Use `.env.example`, then source env before running Go.

```bash
cp .env.example .env
set -a
. ./.env
set +a
```

### Step 6: Apply `${TASK_FILE_NAME}` schema

Choose one:

- Manual `mongosh` migration.
- `POST /internal/v1/cart/schema/bootstrap`.
- Local-only `CART_BOOTSTRAP_COLLECTIONS_ON_STARTUP=true`.

### Step 7: Start backend service

```bash
go run ./cmd/server
```

### Step 8: Verify `${TASK_FILE_NAME}` schema

```bash
curl http://localhost:8084/internal/v1/cart/schema
curl http://localhost:8084/readyz
```

## 10. Running the Project

Full run instructions are reused from:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 10. Project Run Instructions
```

`${TASK_FILE_NAME}`-specific run/verify commands:

Start service:

```bash
cd backend/services/cart-service
set -a
. ./.env
set +a
go run ./cmd/server
```

Check health:

```bash
curl http://localhost:8084/healthz
```

Expected:

```json
{"status":"ok"}
```

Check readiness:

```bash
curl http://localhost:8084/readyz
```

Expected:

```json
{"status":"ready"}
```

Get schema spec:

```bash
curl http://localhost:8084/internal/v1/cart/schema
```

Bootstrap schema:

```bash
curl -X POST http://localhost:8084/internal/v1/cart/schema/bootstrap
```

### Ports and Networking

| Service | Port | Purpose | Status |
|---|---:|---|---|
| `${SERVICE_NAME}` HTTP API | `8084` | REST API, health, schema endpoints | Reused |
| MongoDB | `27017` | Durable `cart_db.carts` storage | Reused |
| Redis | `6379` | Cache and worker lock | Reused |
| Product Service | `8082` | Product snapshot validation | Reused |
| CMS Service | `8089` | Coupon validation | Reused |
| Kafka / RabbitMQ | N/A | Not used by `${TASK_FILE_NAME}` | Not required |

Change `${SERVICE_NAME}` API port:

```env
CART_HTTP_ADDR=:8094
```

Then verify:

```bash
curl http://localhost:8094/healthz
```

## 11. Common Errors & Fixes

Generic setup errors are already documented in:

```text
PREVIOUS_TASK1_DEPENDENCY_FILE
Section: 12. Common Errors and Fixes

PREVIOUS_TASK2_DEPENDENCY_FILE
Section: 13. Task-Specific Troubleshooting
```

`${TASK_FILE_NAME}`-specific troubleshooting:

| Error / Symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `CART_MONGO_DATABASE must be "cart_db"` | Wrong DB env value | Set `CART_MONGO_DATABASE=cart_db` | Do not rename Cart database |
| `SCHEMA_BOOTSTRAP_FAILED` | MongoDB unreachable or user lacks collection/index permission | Check `CART_MONGO_URI`, credentials, MongoDB logs | Use correct DB user privileges before bootstrap |
| `/readyz` returns `503` | MongoDB ping failed | Start MongoDB and verify URI | Run MongoDB before service |
| `collection carts not found` | Migration/bootstrap not run | Run migration or call bootstrap endpoint | Add schema setup to local onboarding |
| Insert/update fails with Mongo validation error | Document missing required field or invalid type | Compare document with `${TASK_FILE_NAME}` schema | Always write through `${SERVICE_NAME}` domain/repository |
| Duplicate key on `uniq_active_cart_per_user` | Active cart already exists for same user | Load/update existing active cart instead of insert | Keep one active cart per owner rule |
| Duplicate key on `uniq_active_cart_per_guest_session` | Active cart already exists for same guest session | Reuse existing guest cart | Do not create multiple active carts for same guest |
| `idx_carts_expires_at_ttl` missing | Migration partially failed or old DB state | Re-run bootstrap/migration and verify indexes | Check indexes after migration |
| Cart disappears after `expires_at` passes | TTL index deleted expired document | For audit carts, do not set expired `expires_at` if retention is needed | Define retention policy before checkout/archive |
| Bootstrap endpoint works locally but unsafe in prod | Internal endpoint may be exposed without auth/gateway protection | Restrict route via gateway/network policy | Never expose internal schema endpoints publicly |

## 12. Security & Best Practices

`${TASK_FILE_NAME}`-specific security and setup best practices:

- Keep `CART_BOOTSTRAP_COLLECTIONS_ON_STARTUP=false` in production. Controlled migration job better hai.
- Protect `POST /internal/v1/cart/schema/bootstrap`. Ye internal endpoint public internet se accessible nahi hona chahiye.
- Do not commit real MongoDB credentials in `.env`.
- Use MongoDB users with least required permissions. Local root user okay hai, production me scoped DB user better hai.
- Do not manually insert cart documents in MongoDB unless testing schema. Direct writes domain validation bypass kar sakte hain.
- MongoDB validator helpful hai, but business rules ka replacement nahi hai.
- Keep unique active cart indexes enabled. Ye duplicate active carts ko prevent karte hain.
- TTL index ko carefully treat karo. If checked-out carts need audit retention, `expires_at` policy clear honi chahiye.
- Use `cart_db` only for `${SERVICE_NAME}` owned data.
- Use Redis as cache only. Checkout or order creation should trust MongoDB cart state, not Redis-only state.

Beginner best practice:

Schema change karne ke baad hamesha three checks run karo: collection exists, validator exists, indexes exist.

## 13. Missing or Misconfigured Things

These are audit notes from current implementation/repo inspection. Ye `${TASK_FILE_NAME}` ke setup ko samajhne ke liye important hain.

| Area | Current state | Risk | Suggested fix |
|---|---|---|---|
| `${SERVICE_NAME}` Dockerfile | Not present | Service container build flow incomplete | Add `backend/services/cart-service/deploy/Dockerfile` in a future DevOps task |
| Local compose file | `infra/compose/docker-compose.local.yml` not present | Beginners may need previous doc examples manually | Create real compose file for MongoDB/Redis and services |
| `.env` loading | Code uses `os.Getenv`; no `godotenv` | `.env` file alone will not load | Keep docs explicit, or add a dev-only env loader if project chooses |
| Schema bootstrap endpoint | Internal route exists without visible auth in service code | Public exposure can modify DB schema/indexes | Protect at gateway/network level or add internal auth |
| Mongo JSON Schema owner rule | Validator allows `user_id` and `guest_session_id` nullable shape, but cannot fully enforce exactly-one owner | Direct DB inserts may create bad owner state | Rely on domain validation and avoid direct writes |
| TTL index policy | TTL index on `expires_at` can delete any document with past date | Audit carts may disappear if `expires_at` not managed | Define checkout/retention policy before production |
| Product/CMS env required at startup | Even schema endpoints need config validation for Product/CMS base URLs | Local schema-only startup can fail if URLs missing | Keep `.env.example` complete or decouple schema-only command later |
| Migration rollback | Down script drops indexes and disables validator, but does not drop collection | Old data remains after rollback | Document rollback intent and clean test DB manually when needed |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `4. Go Dependency System` | Same Go module, version, and dependency workflow |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `5.1 MongoDB` | Same MongoDB install, Docker run, URI, and base migration flow |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `5.2 Redis` | Same Redis install and Docker setup |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `6. External Services` | Same Product Service and CMS Service dependencies |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `8. Complete Environment Variables` | Same `.env` location and full env variable list |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `9. Docker and DevOps Setup` | Same local dependency container setup |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `10. Project Run Instructions` | Same clone/install/run service flow |
| `PREVIOUS_TASK1_DEPENDENCY_FILE` | `12. Common Errors and Fixes` | Same generic Go/Mongo/Redis/env troubleshooting |
| `PREVIOUS_TASK2_DEPENDENCY_FILE` | `6. Database and Cache Analysis` | Same MongoDB source of truth plus Redis cache responsibility |
| `PREVIOUS_TASK2_DEPENDENCY_FILE` | `7. Environment Variables` | Same DB/cache env name mapping |
| `PREVIOUS_TASK2_DEPENDENCY_FILE` | `11. Migration and Bootstrap Notes` | Same migration and index verification baseline |
| `PREVIOUS_TASK2_DEPENDENCY_FILE` | `13. Task-Specific Troubleshooting` | Same MongoDB/Redis setup failure patterns |

## 15. Final Checklist

- [ ] Previous dependency documentation checked.
- [ ] No duplicate base install/setup content added.
- [ ] Go dependencies verified; no new `go get` required for `${TASK_FILE_NAME}`.
- [ ] MongoDB running and reachable.
- [ ] Redis running and reachable for current service startup.
- [ ] `.env` created in `backend/services/cart-service/.env`.
- [ ] `.env` sourced/exported before `go run`.
- [ ] `CART_MONGO_URI` points to reachable MongoDB.
- [ ] `CART_MONGO_DATABASE=cart_db`.
- [ ] Product and CMS base URLs present in `.env`.
- [ ] `carts` collection created or updated.
- [ ] MongoDB JSON Schema validator applied.
- [ ] Required indexes verified.
- [ ] TTL index behavior understood.
- [ ] `POST /internal/v1/cart/schema/bootstrap` protected outside local development.
- [ ] `${SERVICE_NAME}` HTTP API starts successfully.
- [ ] `/healthz` returns `ok`.
- [ ] `/readyz` returns `ready`.
- [ ] `/internal/v1/cart/schema` returns expected collection spec.
- [ ] No new Docker service, queue, or external integration added for `${TASK_FILE_NAME}`.
- [ ] Missing/misconfigured items reviewed before production use.

Final beginner summary:

For `${TASK_FILE_NAME}`, pehle previous dependency docs follow karo. New kaam sirf itna hai ki MongoDB me `${SERVICE_NAME}` ka `cart_db.carts` schema, validator, and indexes properly apply and verify ho jaye. Redis, Docker, Go setup, Product Service, and CMS Service same previous setup reuse karte hain.
